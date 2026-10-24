package config

import (
	"os"
	"testing"
	"time"
)

// clearEnv removes every env var this package reads, so tests don't leak
// into each other or depend on the ambient environment they run in.
func clearEnv(t *testing.T) {
	t.Helper()
	keys := []string{"PORT", "LLM_PROVIDER", "ANTHROPIC_API_KEY", "CLAUDE_MODEL", "GROQ_API_KEY", "GROQ_MODEL", "GEMINI_API_KEY", "GEMINI_MODEL", "DATA_FILE_PATH", "ALLOWED_ORIGINS", "RATE_LIMIT_PER_MINUTE", "RATE_LIMIT_PER_HOUR", "MAX_CONCURRENT_LLM", "GENERAL_RATE_LIMIT_PER_MINUTE", "GENERAL_RATE_LIMIT_PER_HOUR", "TRUST_PROXY_HEADERS", "CASE_RETENTION_DAYS", "INTAKE_TIMEOUT_SECONDS"}
	for _, k := range keys {
		old, existed := os.LookupEnv(k)
		os.Unsetenv(k)
		k, old, existed := k, old, existed // capture for closure
		t.Cleanup(func() {
			if existed {
				os.Setenv(k, old)
			}
		})
	}
}

func TestLoad_Defaults(t *testing.T) {
	clearEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() with no env vars set should succeed with defaults, got error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %q", cfg.Port)
	}
	if cfg.ClaudeModel == "" {
		t.Error("expected a non-empty default model")
	}
	if cfg.RateLimitPerMinute <= 0 {
		t.Error("expected a positive default rate limit")
	}
	if cfg.RateLimitPerHour <= 0 {
		t.Error("expected a positive default hourly rate limit")
	}
	if cfg.RateLimitPerHour < cfg.RateLimitPerMinute {
		t.Errorf("expected the hourly limit (%d) to be at least the per-minute limit (%d) — a sane hourly backstop shouldn't bind tighter than a single minute already would",
			cfg.RateLimitPerHour, cfg.RateLimitPerMinute)
	}
	if cfg.MaxConcurrentLLM <= 0 {
		t.Error("expected a positive default max concurrent LLM calls")
	}
	if cfg.GeneralRateLimitPerMinute <= 0 {
		t.Error("expected a positive default general rate limit")
	}
	if cfg.GeneralRateLimitPerHour <= 0 {
		t.Error("expected a positive default general hourly rate limit")
	}
	if cfg.GeneralRateLimitPerMinute <= cfg.RateLimitPerMinute {
		t.Errorf("expected the general limit (%d) to be more generous than intake's cost-control limit (%d) by default", cfg.GeneralRateLimitPerMinute, cfg.RateLimitPerMinute)
	}
	if len(cfg.AllowedOrigins) == 0 {
		t.Error("expected a non-empty default allowed origins list")
	}
}

func TestLoad_MissingAPIKeyDoesNotFailStartup(t *testing.T) {
	// Deliberate: the server should start without a key so the rest of
	// the system (data loading, health checks) is independently
	// verifiable. Every extraction call will fail clearly at request
	// time instead — see internal/extract.ClaudeClient.Extract.
	clearEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected Load() to succeed without an API key, got: %v", err)
	}
	if cfg.AnthropicAPIKey != "" {
		t.Errorf("expected empty API key when unset, got %q", cfg.AnthropicAPIKey)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	clearEnv(t)
	os.Setenv("PORT", "9090")
	os.Setenv("ANTHROPIC_API_KEY", "test-key-123")
	os.Setenv("CLAUDE_MODEL", "claude-custom-model")
	os.Setenv("ALLOWED_ORIGINS", "https://example.com, https://other.example.com")
	os.Setenv("RATE_LIMIT_PER_MINUTE", "25")
	os.Setenv("RATE_LIMIT_PER_HOUR", "300")
	os.Setenv("MAX_CONCURRENT_LLM", "7")
	os.Setenv("GENERAL_RATE_LIMIT_PER_MINUTE", "40")
	os.Setenv("GENERAL_RATE_LIMIT_PER_HOUR", "400")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.Port != "9090" {
		t.Errorf("expected overridden port 9090, got %q", cfg.Port)
	}
	if cfg.AnthropicAPIKey != "test-key-123" {
		t.Errorf("expected overridden API key, got %q", cfg.AnthropicAPIKey)
	}
	if cfg.ClaudeModel != "claude-custom-model" {
		t.Errorf("expected overridden model, got %q", cfg.ClaudeModel)
	}
	if len(cfg.AllowedOrigins) != 2 || cfg.AllowedOrigins[0] != "https://example.com" || cfg.AllowedOrigins[1] != "https://other.example.com" {
		t.Errorf("expected two trimmed allowed origins, got %v", cfg.AllowedOrigins)
	}
	if cfg.RateLimitPerMinute != 25 {
		t.Errorf("expected overridden rate limit 25, got %d", cfg.RateLimitPerMinute)
	}
	if cfg.RateLimitPerHour != 300 {
		t.Errorf("expected overridden hourly rate limit 300, got %d", cfg.RateLimitPerHour)
	}
	if cfg.MaxConcurrentLLM != 7 {
		t.Errorf("expected overridden max concurrent LLM calls 7, got %d", cfg.MaxConcurrentLLM)
	}
	if cfg.GeneralRateLimitPerMinute != 40 {
		t.Errorf("expected overridden general rate limit 40, got %d", cfg.GeneralRateLimitPerMinute)
	}
	if cfg.GeneralRateLimitPerHour != 400 {
		t.Errorf("expected overridden general hourly rate limit 400, got %d", cfg.GeneralRateLimitPerHour)
	}
}

func TestLoad_InvalidRateLimitFailsLoudly(t *testing.T) {
	clearEnv(t)
	os.Setenv("RATE_LIMIT_PER_MINUTE", "not-a-number")
	if _, err := Load(); err == nil {
		t.Fatal("expected Load() to fail on a non-numeric RATE_LIMIT_PER_MINUTE")
	}
}

func TestLoad_NegativeRateLimitFailsLoudly(t *testing.T) {
	clearEnv(t)
	os.Setenv("RATE_LIMIT_PER_MINUTE", "-5")
	if _, err := Load(); err == nil {
		t.Fatal("expected Load() to fail on a non-positive RATE_LIMIT_PER_MINUTE")
	}
}

func TestLoad_InvalidRateLimitPerHourFailsLoudly(t *testing.T) {
	clearEnv(t)
	os.Setenv("RATE_LIMIT_PER_HOUR", "not-a-number")
	if _, err := Load(); err == nil {
		t.Fatal("expected Load() to fail on a non-numeric RATE_LIMIT_PER_HOUR")
	}
}

func TestLoad_NegativeRateLimitPerHourFailsLoudly(t *testing.T) {
	clearEnv(t)
	os.Setenv("RATE_LIMIT_PER_HOUR", "0")
	if _, err := Load(); err == nil {
		t.Fatal("expected Load() to fail on a non-positive RATE_LIMIT_PER_HOUR")
	}
}

func TestLoad_InvalidMaxConcurrentLLMFailsLoudly(t *testing.T) {
	clearEnv(t)
	os.Setenv("MAX_CONCURRENT_LLM", "not-a-number")
	if _, err := Load(); err == nil {
		t.Fatal("expected Load() to fail on a non-numeric MAX_CONCURRENT_LLM")
	}
}

func TestLoad_NegativeMaxConcurrentLLMFailsLoudly(t *testing.T) {
	clearEnv(t)
	os.Setenv("MAX_CONCURRENT_LLM", "-1")
	if _, err := Load(); err == nil {
		t.Fatal("expected Load() to fail on a non-positive MAX_CONCURRENT_LLM")
	}
}

func TestLoad_InvalidGeneralRateLimitPerMinuteFailsLoudly(t *testing.T) {
	clearEnv(t)
	os.Setenv("GENERAL_RATE_LIMIT_PER_MINUTE", "not-a-number")
	if _, err := Load(); err == nil {
		t.Fatal("expected Load() to fail on a non-numeric GENERAL_RATE_LIMIT_PER_MINUTE")
	}
}

func TestLoad_NegativeGeneralRateLimitPerMinuteFailsLoudly(t *testing.T) {
	clearEnv(t)
	os.Setenv("GENERAL_RATE_LIMIT_PER_MINUTE", "0")
	if _, err := Load(); err == nil {
		t.Fatal("expected Load() to fail on a non-positive GENERAL_RATE_LIMIT_PER_MINUTE")
	}
}

func TestLoad_InvalidGeneralRateLimitPerHourFailsLoudly(t *testing.T) {
	clearEnv(t)
	os.Setenv("GENERAL_RATE_LIMIT_PER_HOUR", "not-a-number")
	if _, err := Load(); err == nil {
		t.Fatal("expected Load() to fail on a non-numeric GENERAL_RATE_LIMIT_PER_HOUR")
	}
}

func TestLoad_NegativeGeneralRateLimitPerHourFailsLoudly(t *testing.T) {
	clearEnv(t)
	os.Setenv("GENERAL_RATE_LIMIT_PER_HOUR", "-20")
	if _, err := Load(); err == nil {
		t.Fatal("expected Load() to fail on a non-positive GENERAL_RATE_LIMIT_PER_HOUR")
	}
}

func TestLoad_EmptyAllowedOriginsFailsLoudly(t *testing.T) {
	clearEnv(t)
	os.Setenv("ALLOWED_ORIGINS", "   ,  ,")
	if _, err := Load(); err == nil {
		t.Fatal("expected Load() to fail when ALLOWED_ORIGINS resolves to an empty list")
	}
}

func TestLoad_DefaultProviderIsAnthropic(t *testing.T) {
	clearEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.LLMProvider != "anthropic" {
		t.Errorf("expected default LLM_PROVIDER anthropic (backward compatible with every deployment before this option existed), got %q", cfg.LLMProvider)
	}
}

func TestLoad_ProviderIsCaseInsensitive(t *testing.T) {
	clearEnv(t)
	os.Setenv("LLM_PROVIDER", "GROQ")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.LLMProvider != "groq" {
		t.Errorf("expected LLM_PROVIDER to be normalized to lowercase, got %q", cfg.LLMProvider)
	}
}

func TestLoad_UnrecognizedProviderFailsLoudly(t *testing.T) {
	clearEnv(t)
	os.Setenv("LLM_PROVIDER", "groc") // typo for groq
	if _, err := Load(); err == nil {
		t.Fatal("expected Load() to fail on an unrecognized LLM_PROVIDER value rather than silently doing something unexpected")
	}
}

func TestLoad_GroqAndGeminiDefaults(t *testing.T) {
	clearEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.GroqModel == "" {
		t.Error("expected a non-empty default Groq model even when LLM_PROVIDER isn't groq — switching providers later shouldn't require also setting the model")
	}
	if cfg.GeminiModel == "" {
		t.Error("expected a non-empty default Gemini model for the same reason")
	}
}

func TestLoad_GroqAndGeminiEnvOverrides(t *testing.T) {
	clearEnv(t)
	os.Setenv("GROQ_API_KEY", "groq-test-key")
	os.Setenv("GROQ_MODEL", "custom-groq-model")
	os.Setenv("GEMINI_API_KEY", "gemini-test-key")
	os.Setenv("GEMINI_MODEL", "custom-gemini-model")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.GroqAPIKey != "groq-test-key" || cfg.GroqModel != "custom-groq-model" {
		t.Errorf("expected overridden Groq settings, got key=%q model=%q", cfg.GroqAPIKey, cfg.GroqModel)
	}
	if cfg.GeminiAPIKey != "gemini-test-key" || cfg.GeminiModel != "custom-gemini-model" {
		t.Errorf("expected overridden Gemini settings, got key=%q model=%q", cfg.GeminiAPIKey, cfg.GeminiModel)
	}
}

func TestConfig_ActiveAPIKey_SelectsByProvider(t *testing.T) {
	cases := []struct {
		provider string
		want     string
	}{
		{"anthropic", "anthropic-key"},
		{"groq", "groq-key"},
		{"gemini", "gemini-key"},
	}
	for _, tc := range cases {
		cfg := Config{
			LLMProvider:     tc.provider,
			AnthropicAPIKey: "anthropic-key",
			GroqAPIKey:      "groq-key",
			GeminiAPIKey:    "gemini-key",
		}
		if got := cfg.ActiveAPIKey(); got != tc.want {
			t.Errorf("provider %q: ActiveAPIKey() = %q, want %q", tc.provider, got, tc.want)
		}
	}
}

func TestConfig_ActiveAPIKeyEnvVar_SelectsByProvider(t *testing.T) {
	cases := []struct {
		provider string
		want     string
	}{
		{"anthropic", "ANTHROPIC_API_KEY"},
		{"groq", "GROQ_API_KEY"},
		{"gemini", "GEMINI_API_KEY"},
	}
	for _, tc := range cases {
		cfg := Config{LLMProvider: tc.provider}
		if got := cfg.ActiveAPIKeyEnvVar(); got != tc.want {
			t.Errorf("provider %q: ActiveAPIKeyEnvVar() = %q, want %q", tc.provider, got, tc.want)
		}
	}
}

func TestLoadDotEnv_LoadsUnsetVarsAndPreservesSetVars(t *testing.T) {
	clearEnv(t)
	os.Setenv("PORT", "7070") // already set, should NOT be overwritten

	// Write a temporary .env in a temp dir and test parsing
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(origDir)

	envContent := "# Comment line\n\nPORT=9999\nLLM_PROVIDER=gemini\nGEMINI_API_KEY=\"test-gemini-key\"\n"
	if err := os.WriteFile(".env", []byte(envContent), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	LoadDotEnv()

	if got := os.Getenv("PORT"); got != "7070" {
		t.Errorf("expected PORT to remain 7070 (already set), got %q", got)
	}
	if got := os.Getenv("LLM_PROVIDER"); got != "gemini" {
		t.Errorf("expected LLM_PROVIDER to be loaded as gemini, got %q", got)
	}
	if got := os.Getenv("GEMINI_API_KEY"); got != "test-gemini-key" {
		t.Errorf("expected GEMINI_API_KEY to be loaded with quotes stripped, got %q", got)
	}
}

func TestLoad_PrivacyAndProxyDefaults(t *testing.T) {
	clearEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.TrustProxyHeaders {
		t.Error("TRUST_PROXY_HEADERS must default to false: trusting the header without a proxy lets callers bypass every per-IP limit")
	}
	if cfg.CaseRetentionDays != 30 {
		t.Errorf("expected 30-day default retention, got %d", cfg.CaseRetentionDays)
	}
	if cfg.IntakeTimeout != 45*time.Second {
		t.Errorf("expected 45s default intake timeout, got %v", cfg.IntakeTimeout)
	}
}

func TestLoad_ParsesRetentionProxyAndTimeout(t *testing.T) {
	clearEnv(t)
	t.Setenv("TRUST_PROXY_HEADERS", "true")
	t.Setenv("CASE_RETENTION_DAYS", "7")
	t.Setenv("INTAKE_TIMEOUT_SECONDS", "60")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.TrustProxyHeaders {
		t.Error("expected TRUST_PROXY_HEADERS=true to be honoured")
	}
	if cfg.CaseRetentionDays != 7 {
		t.Errorf("expected 7 retention days, got %d", cfg.CaseRetentionDays)
	}
	if cfg.IntakeTimeout != 60*time.Second {
		t.Errorf("expected 60s intake timeout, got %v", cfg.IntakeTimeout)
	}
}

func TestLoad_ZeroRetentionIsAllowedAndDisablesPurging(t *testing.T) {
	clearEnv(t)
	t.Setenv("CASE_RETENTION_DAYS", "0")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.CaseRetentionDays != 0 {
		t.Errorf("expected 0 to be preserved as the explicit opt-out, got %d", cfg.CaseRetentionDays)
	}
}

func TestLoad_RejectsInvalidRetentionAndTimeout(t *testing.T) {
	cases := []struct{ key, value string }{
		{"CASE_RETENTION_DAYS", "-1"},
		{"CASE_RETENTION_DAYS", "not-a-number"},
		{"INTAKE_TIMEOUT_SECONDS", "4"},
		{"INTAKE_TIMEOUT_SECONDS", "121"},
		{"TRUST_PROXY_HEADERS", "sometimes"},
	}
	for _, tc := range cases {
		t.Run(tc.key+"="+tc.value, func(t *testing.T) {
			clearEnv(t)
			t.Setenv(tc.key, tc.value)
			if _, err := Load(); err == nil {
				t.Fatalf("expected Load to reject %s=%q", tc.key, tc.value)
			}
		})
	}
}

func TestActiveModelMatchesProvider(t *testing.T) {
	cfg := Config{
		LLMProvider: "groq",
		ClaudeModel: "claude-x",
		GroqModel:   "groq-x",
		GeminiModel: "gemini-x",
	}
	if got := cfg.ActiveModel(); got != "groq-x" {
		t.Errorf("expected groq model, got %q", got)
	}
	cfg.LLMProvider = "gemini"
	if got := cfg.ActiveModel(); got != "gemini-x" {
		t.Errorf("expected gemini model, got %q", got)
	}
	cfg.LLMProvider = "anthropic"
	if got := cfg.ActiveModel(); got != "claude-x" {
		t.Errorf("expected anthropic model, got %q", got)
	}
}
