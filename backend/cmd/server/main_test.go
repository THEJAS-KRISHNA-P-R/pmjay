package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pmjay-advocate/backend/internal/config"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// setBaseEnv sets exactly the env vars needed for config.Load() to
// succeed, pointed at a scratch data file, on an OS-assigned ephemeral
// port (PORT=0) so this test can never collide with another test or a
// real running instance. t.Setenv unsets automatically at test end.
func setBaseEnv(t *testing.T, dataFilePath string) {
	t.Helper()
	t.Setenv("PORT", "0")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("CLAUDE_MODEL", "")
	t.Setenv("DATA_FILE_PATH", dataFilePath)
	t.Setenv("ALLOWED_ORIGINS", "http://localhost:3000")
	t.Setenv("RATE_LIMIT_PER_MINUTE", "")
}

func TestNewExtractor_SelectsClientByProvider(t *testing.T) {
	cases := []struct {
		provider string
		wantType string
	}{
		{"anthropic", "*extract.ClaudeClient"},
		{"groq", "*extract.GroqClient"},
		{"gemini", "*extract.GeminiClient"},
	}
	for _, tc := range cases {
		t.Run(tc.provider, func(t *testing.T) {
			cfg := config.Config{LLMProvider: tc.provider, ClaudeModel: "m", GroqModel: "m", GeminiModel: "m"}
			got, err := newExtractor(cfg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			gotType := reflect.TypeOf(got).String()
			if gotType != tc.wantType {
				t.Errorf("provider %q: expected %s, got %s", tc.provider, tc.wantType, gotType)
			}
		})
	}
}

func TestNewExtractor_UnrecognizedProviderFailsLoud(t *testing.T) {
	// config.Load() already rejects this before newExtractor ever sees
	// it — this test exists for the same reason newExtractor's default
	// case exists: don't assume that guarantee holds forever just
	// because it holds today.
	cfg := config.Config{LLMProvider: "not-a-real-provider"}
	_, err := newExtractor(cfg)
	if err == nil {
		t.Fatal("expected an error for an unrecognized provider")
	}
}

func TestRun_StartsServesAndShutsDownGracefullyOnContextCancel(t *testing.T) {
	setBaseEnv(t, filepath.Join(t.TempDir(), "cases.json"))

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- run(ctx, discardLogger()) }()

	// Give the goroutine a moment to actually start listening before
	// asking it to stop — run() logs "listening" right before
	// ListenAndServe blocks, so this is generous, not a tight race.
	time.Sleep(150 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected a clean shutdown (nil error) after context cancellation, got: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("run did not return within 5s of context cancellation — graceful shutdown may be hanging")
	}
}

func TestRun_ReturnsConfigError(t *testing.T) {
	setBaseEnv(t, filepath.Join(t.TempDir(), "cases.json"))
	t.Setenv("RATE_LIMIT_PER_MINUTE", "not-a-number") // config.Load's own validation should reject this

	err := run(context.Background(), discardLogger())
	if err == nil {
		t.Fatal("expected run to surface config.Load's validation error")
	}
}

func TestRun_ReturnsStoreCreationError(t *testing.T) {
	// A regular file where a directory needs to go: os.MkdirAll for the
	// data file's parent directory will fail, because "not-a-directory"
	// already exists and is not a directory.
	tmp := t.TempDir()
	blocker := filepath.Join(tmp, "not-a-directory")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	setBaseEnv(t, filepath.Join(blocker, "subdir", "cases.json"))

	err := run(context.Background(), discardLogger())
	if err == nil {
		t.Fatal("expected run to surface the file store's directory-creation error")
	}
}

func TestRun_ReturnsShutdownError(t *testing.T) {
	// A negative timeout makes context.WithTimeout hand back an
	// already-expired context — but Shutdown only consults the context
	// once it finds a non-idle connection to wait on, so an idle server
	// with zero connections returns nil immediately regardless of the
	// context's state. To actually reach Shutdown's error path, this
	// test needs one real, still-open connection at shutdown time: it
	// finds a free port itself (closing its probe listener before
	// run() binds the same port — the standard, low-flake way to learn
	// an address in advance), dials in and deliberately never completes
	// the HTTP request, then triggers shutdown with the timeout already
	// expired.
	orig := shutdownTimeout
	shutdownTimeout = -1 * time.Second
	defer func() { shutdownTimeout = orig }()

	probe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	port := probe.Addr().(*net.TCPAddr).Port
	probe.Close()

	setBaseEnv(t, filepath.Join(t.TempDir(), "cases.json"))
	t.Setenv("PORT", itoa(port))

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- run(ctx, discardLogger()) }()

	addr := "127.0.0.1:" + itoa(port)
	var conn net.Conn
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		conn, err = net.Dial("tcp", addr)
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if conn == nil {
		t.Fatalf("server never started accepting connections on %s: %v", addr, err)
	}
	defer conn.Close()
	// Send a request line with no terminating blank line — the server
	// is left waiting for more of the request, so this connection is
	// genuinely active/non-idle from Shutdown's point of view, not just
	// open.
	conn.Write([]byte("GET /api/v1/health HTTP/1.1\r\nHost: localhost\r\n"))

	cancel()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected run to surface httpServer.Shutdown's error: a non-idle connection was held open past the already-expired shutdownTimeout")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("run did not return within 5s")
	}
}

func TestRun_ReturnsServerStartError(t *testing.T) {
	// Bind a real listener on an OS-assigned port first, then point run()
	// at that exact port — ListenAndServe must fail with "address already
	// in use", exercising the serverErr channel path distinctly from the
	// context-cancellation path.
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	defer l.Close()
	port := l.Addr().(*net.TCPAddr).Port

	setBaseEnv(t, filepath.Join(t.TempDir(), "cases.json"))
	t.Setenv("PORT", itoa(port))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	runErr := run(ctx, discardLogger())
	if runErr == nil {
		t.Fatal("expected run to return an error when the port is already bound")
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func TestMain_ExitsWithCode1WhenRunFails(t *testing.T) {
	// os.Exit terminates the process immediately — it cannot be caught
	// or asserted on from within the same process, so the standard (and
	// only correct) way to test a path that calls it is to re-exec this
	// same test binary as a subprocess and inspect its actual exit code.
	// This is the pattern the Go standard library itself uses for
	// exactly this situation, not a workaround specific to this file.
	if os.Getenv("PMJAY_MAIN_SUBPROCESS") == "1" {
		// We *are* the subprocess: make run() fail deterministically via
		// bad config, then call the real main() and let it really exit.
		os.Setenv("RATE_LIMIT_PER_MINUTE", "not-a-number")
		os.Setenv("PORT", "0")
		os.Setenv("DATA_FILE_PATH", filepath.Join(os.TempDir(), "pmjay-main-subprocess-test-cases.json"))
		os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000")
		main()
		t.Fatal("main() should have called os.Exit(1) and never reached this line")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestMain_ExitsWithCode1WhenRunFails$")
	cmd.Env = append(os.Environ(), "PMJAY_MAIN_SUBPROCESS=1")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected the subprocess to exit with a non-zero code via os.Exit(1), got err=%v, stderr=%s", err, stderr.String())
	}
	if exitErr.ExitCode() != 1 {
		t.Errorf("expected exit code 1 (main.go's os.Exit(1) on a failed run()), got %d. stderr: %s", exitErr.ExitCode(), stderr.String())
	}
}
