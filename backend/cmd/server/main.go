// Command server is the entrypoint for the PMJAY Point-of-Denial Advocate
// backend. It reads configuration from the environment, loads the
// embedded HBP dataset, constructs the real Claude-backed extractor and
// the file-backed case store, and serves the HTTP API until it receives
// a shutdown signal.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pmjay-advocate/backend/internal/api"
	"github.com/pmjay-advocate/backend/internal/config"
	"github.com/pmjay-advocate/backend/internal/extract"
	"github.com/pmjay-advocate/backend/internal/hbp"
	"github.com/pmjay-advocate/backend/internal/store"
)

// shutdownTimeout bounds how long graceful shutdown waits for in-flight
// requests before giving up. A package var rather than a hardcoded
// literal so a test can shrink it to make the "Shutdown itself times
// out" branch deterministically reachable — real deployments always get
// the 15s default; nothing about production behavior changes.
var shutdownTimeout = 15 * time.Second

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, logger); err != nil {
		logger.Error("fatal startup error", "error", err)
		os.Exit(1)
	}
}

// run holds every bit of orchestration logic that main.go exists for:
// load config, load the dataset, wire the store and extractor, serve
// until ctx is cancelled, then shut down gracefully. It takes ctx rather
// than listening for OS signals itself so it's directly unit-testable —
// a test can cancel ctx deterministically instead of needing to send a
// real signal to the whole test process. main is the only caller that
// needs a real signal.NotifyContext; everything below this line is
// signal-agnostic on purpose.
func run(ctx context.Context, logger *slog.Logger) error {
	config.LoadDotEnv()
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if cfg.ActiveAPIKey() == "" {
		logger.Warn("no API key set for the configured LLM provider — the server will start, but every case intake request will fail until it is configured", "provider", cfg.LLMProvider, "expected_env_var", cfg.ActiveAPIKeyEnvVar())
	}

	dataset, err := hbp.Load()
	if err != nil {
		return err
	}
	logger.Info("dataset loaded", "packages", len(dataset.Packages), "exclusions", len(dataset.Exclusions))

	fileStore, err := store.NewFileStore(cfg.DataFilePath)
	if err != nil {
		return err
	}
	logger.Info("case store ready", "path", cfg.DataFilePath)

	extractor, err := newExtractor(cfg)
	if err != nil {
		return err
	}
	logger.Info("extractor ready", "provider", cfg.LLMProvider)

	server := &api.Server{
		Dataset:   dataset,
		Extractor: extractor,
		Store:     fileStore,
		Logger:    logger,
	}
	handler := api.NewRouter(server, cfg.AllowedOrigins, cfg.RateLimitPerMinute, cfg.RateLimitPerHour, cfg.MaxConcurrentLLM, cfg.GeneralRateLimitPerMinute, cfg.GeneralRateLimitPerHour)

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("listening", "port", cfg.Port, "model", cfg.ClaudeModel)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return err
	}
	logger.Info("shutdown complete")
	return nil
}

// newExtractor builds the Extractor for whichever provider cfg.LLMProvider
// selects. config.Load already validated LLMProvider is one of the three
// recognized values (see validLLMProviders) — the default case below
// existing anyway, rather than assuming that guarantee holds forever, is
// the same "fail loud, don't silently do the wrong thing" principle
// config.Load itself follows, applied one level up.
func newExtractor(cfg config.Config) (extract.Extractor, error) {
	switch cfg.LLMProvider {
	case "groq":
		return extract.NewGroqClient(cfg.GroqAPIKey, cfg.GroqModel), nil
	case "gemini":
		return extract.NewGeminiClient(cfg.GeminiAPIKey, cfg.GeminiModel), nil
	case "anthropic":
		return extract.NewClaudeClient(cfg.AnthropicAPIKey, cfg.ClaudeModel), nil
	default:
		return nil, fmt.Errorf("newExtractor: unrecognized LLM_PROVIDER %q — this should have been caught by config.Load", cfg.LLMProvider)
	}
}
