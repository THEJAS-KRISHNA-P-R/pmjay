// Package config loads runtime configuration from environment variables.
// No config file format, no external config-management dependency —
// environment variables are the standard, zero-dependency way to inject
// configuration into a container or a systemd unit, and this system has
// few enough settings that a dedicated format would be overhead, not
// clarity.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config is the full set of runtime settings this system needs.
type Config struct {
	// Port the HTTP server listens on.
	Port string

	// LLMProvider selects which extraction backend Extract calls —
	// "anthropic" (default), "groq", or "gemini". This exists because
	// this system's whole extraction contract (see internal/extract's
	// extractMatchToolSchema) is a schema, not provider-specific prose,
	// so swapping the model behind it is a real, supportable choice, not
	// a hack — see ../../../ARCHITECTURE.md's "multi-provider extraction"
	// section for why a team might genuinely want Groq's free/cheap
	// tier or Gemini's AI Studio free tier over a paid Anthropic key,
	// particularly for a low-volume deployment.
	LLMProvider string

	// AnthropicAPIKey authenticates calls to the real Claude API — read
	// regardless of LLMProvider (so switching back doesn't lose a
	// previously-set key), but only used when LLMProvider is
	// "anthropic". Required in production for that provider; the server
	// will still start without it (so the rest of the system — data
	// loading, health checks — can be verified independently) but every
	// intake request will fail clearly until it's set. Same pattern
	// applies to GroqAPIKey/GeminiAPIKey below for their providers.
	AnthropicAPIKey string

	// ClaudeModel is the model used when LLMProvider is "anthropic".
	// Defaults to a current, cheap model — see ../../../ARCHITECTURE.md on
	// why this system deliberately does not need a large model for this
	// task; the same reasoning sets GroqModel's and GeminiModel's
	// defaults below.
	ClaudeModel string

	// GroqAPIKey authenticates calls to Groq's OpenAI-compatible API —
	// used only when LLMProvider is "groq". See internal/extract's
	// GroqClient for why structured output there uses strict:false.
	GroqAPIKey string

	// GroqModel is the model used when LLMProvider is "groq".
	GroqModel string

	// GeminiAPIKey authenticates calls to Google's Generative Language
	// API — used only when LLMProvider is "gemini". The same key Google
	// AI Studio issues for its free tier works here directly.
	GeminiAPIKey string

	// GeminiModel is the model used when LLMProvider is "gemini".
	GeminiModel string

	// DataFilePath is where FileStore persists case records.
	DataFilePath string

	// AllowedOrigins is the CORS allowlist for the frontend origin(s).
	AllowedOrigins []string

	// RateLimitPerMinute bounds how many intake requests one client IP
	// may make per minute — see internal/api/middleware.go. This is a
	// direct cost control: each intake request triggers one paid LLM
	// call, so this is the backend's main defense against a runaway
	// bill from either a bug or abusive traffic. Worth keeping even on
	// Gemini's or Groq's free tiers — a free tier still has a rate cap
	// somewhere, and this control is what keeps a bug from silently
	// burning through it.
	RateLimitPerMinute int

	// RateLimitPerHour bounds how many intake requests one client IP
	// may make per hour — a secondary backstop so a burst doesn't
	// consume the daily budget.
	RateLimitPerHour int

	// MaxConcurrentLLM bounds simultaneous in-flight LLM calls —
	// prevents thundering herd from overwhelming the provider or
	// exhausting quota in a spike.
	MaxConcurrentLLM int
}

// validLLMProviders is checked by Load — kept as an explicit set rather
// than "anything not empty is fine" so a typo in LLM_PROVIDER (e.g.
// "groc") fails loudly at startup instead of silently falling through to
// whatever the zero-value behavior would be.
var validLLMProviders = map[string]bool{
	"anthropic": true,
	"groq":      true,
	"gemini":    true,
}

// Load reads configuration from environment variables, applying the
// defaults documented in .env.example. It does not fail merely because
// the selected provider's API key is unset — see AnthropicAPIKey's field
// doc comment — but does fail on structurally invalid values (e.g. a
// non-numeric rate limit, or an LLM_PROVIDER value that isn't one of the
// three supported providers) so a deployment misconfiguration is caught
// at startup, not mid-incident.
func Load() (Config, error) {
	cfg := Config{
		Port:                getEnv("PORT", "8080"),
		LLMProvider:         strings.ToLower(getEnv("LLM_PROVIDER", "anthropic")),
		AnthropicAPIKey:     os.Getenv("ANTHROPIC_API_KEY"),
		ClaudeModel:         getEnv("CLAUDE_MODEL", "claude-haiku-4-5-20251001"),
		GroqAPIKey:          os.Getenv("GROQ_API_KEY"),
		GroqModel:           getEnv("GROQ_MODEL", "openai/gpt-oss-120b"),
		GeminiAPIKey:        os.Getenv("GEMINI_API_KEY"),
		GeminiModel:         getEnv("GEMINI_MODEL", "gemini-2.5-flash-lite"),
		DataFilePath:        getEnv("DATA_FILE_PATH", "./data/cases.json"),
		AllowedOrigins:      splitAndTrim(getEnv("ALLOWED_ORIGINS", "http://localhost:3000")),
		RateLimitPerMinute:  1,
		RateLimitPerHour:    20,
		MaxConcurrentLLM:    3,
	}

	if !validLLMProviders[cfg.LLMProvider] {
		return Config{}, fmt.Errorf("config: LLM_PROVIDER must be one of anthropic, groq, gemini — got %q", cfg.LLMProvider)
	}

	if raw := os.Getenv("RATE_LIMIT_PER_MINUTE"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, fmt.Errorf("config: RATE_LIMIT_PER_MINUTE must be an integer, got %q: %w", raw, err)
		}
		if n <= 0 {
			return Config{}, fmt.Errorf("config: RATE_LIMIT_PER_MINUTE must be positive, got %d", n)
		}
		cfg.RateLimitPerMinute = n
	}

	if raw := os.Getenv("RATE_LIMIT_PER_HOUR"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, fmt.Errorf("config: RATE_LIMIT_PER_HOUR must be an integer, got %q: %w", raw, err)
		}
		if n <= 0 {
			return Config{}, fmt.Errorf("config: RATE_LIMIT_PER_HOUR must be positive, got %d", n)
		}
		cfg.RateLimitPerHour = n
	}

	if raw := os.Getenv("MAX_CONCURRENT_LLM"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, fmt.Errorf("config: MAX_CONCURRENT_LLM must be an integer, got %q: %w", raw, err)
		}
		if n <= 0 {
			return Config{}, fmt.Errorf("config: MAX_CONCURRENT_LLM must be positive, got %d", n)
		}
		cfg.MaxConcurrentLLM = n
	}

	if cfg.Port == "" {
		return Config{}, fmt.Errorf("config: PORT must not be empty")
	}
	if len(cfg.AllowedOrigins) == 0 {
		return Config{}, fmt.Errorf("config: ALLOWED_ORIGINS must not be empty")
	}

	return cfg, nil
}

// ActiveAPIKey returns whichever provider-specific key LLMProvider
// selects — the one place that mapping lives, so main.go's startup
// warning and any future caller don't each need their own copy of the
// same three-way switch.
func (c Config) ActiveAPIKey() string {
	switch c.LLMProvider {
	case "groq":
		return c.GroqAPIKey
	case "gemini":
		return c.GeminiAPIKey
	default:
		return c.AnthropicAPIKey
	}
}

// ActiveAPIKeyEnvVar names the environment variable ActiveAPIKey reads
// from, for error/warning messages that need to tell an operator exactly
// which one to set.
func (c Config) ActiveAPIKeyEnvVar() string {
	switch c.LLMProvider {
	case "groq":
		return "GROQ_API_KEY"
	case "gemini":
		return "GEMINI_API_KEY"
	default:
		return "ANTHROPIC_API_KEY"
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// LoadDotEnv loads key=value pairs from a local .env file if present,
// without overriding any environment variables that are already set.
func LoadDotEnv() {
	paths := []string{".env", "../.env", "../../.env"}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				k := strings.TrimSpace(parts[0])
				v := strings.TrimSpace(parts[1])
				v = strings.Trim(v, `"'`)
				if _, set := os.LookupEnv(k); !set {
					os.Setenv(k, v)
				}
			}
		}
		break
	}
}
