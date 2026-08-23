# `internal/config`

The smallest package in this backend, and one of the highest-covered (96.9%). One file, one job: read environment variables into a `Config` struct, with sane defaults and startup-time validation.

## File

`config.go` — that's the whole package.

## The settings, what each one actually controls, and why

| Env var | Field | Default | What it actually gates |
|---|---|---|---|
| `PORT` | `Port` | `8080` | HTTP listen port. |
| `LLM_PROVIDER` | `LLMProvider` | `anthropic` | Which of the three real `Extractor` implementations `cmd/server/main.go`'s `newExtractor` constructs — `anthropic`, `groq`, or `gemini`. Normalized to lowercase; anything else fails `Load()` outright, not silently defaulting. See `internal/extract/README.md`'s "multi-provider extraction" section. |
| `ANTHROPIC_API_KEY` | `AnthropicAPIKey` | *(empty)* | Auth for `internal/extract`'s Claude API calls — read regardless of `LLM_PROVIDER`, only *used* when it's `anthropic`, so switching providers and back doesn't lose a previously-set key. |
| `CLAUDE_MODEL` | `ClaudeModel` | `claude-haiku-4-5-20251001` | Which Claude model does extraction — see `internal/extract/README.md` on why a small model is the deliberate choice, not a cost-cutting compromise. |
| `GROQ_API_KEY` | `GroqAPIKey` | *(empty)* | Auth for Groq — used only when `LLM_PROVIDER=groq`. |
| `GROQ_MODEL` | `GroqModel` | `openai/gpt-oss-120b` | Which Groq-hosted model does extraction. |
| `GEMINI_API_KEY` | `GeminiAPIKey` | *(empty)* | Auth for Gemini/AI Studio — used only when `LLM_PROVIDER=gemini`. |
| `GEMINI_MODEL` | `GeminiModel` | `gemini-2.5-flash-lite` | Which Gemini model does extraction. |
| `DATA_FILE_PATH` | `DataFilePath` | `./data/cases.json` | Where `store.FileStore` persists case records. |
| `ALLOWED_ORIGINS` | `AllowedOrigins` | `http://localhost:3000` | CORS allowlist — comma-separated, parsed by `splitAndTrim`. |
| `RATE_LIMIT_PER_MINUTE` | `RateLimitPerMinute` | `10` | Per-IP intake rate limit — see below. |

`.env.example` documents exactly these, no more, no less — kept that way deliberately (checked directly against this file as part of the "config files up to date" review this codebase has gone through repeatedly across sessions; see `../../../HANDOVER.md`). If you add a new setting here, add it to `.env.example` in the same change, or a real deployer configuring this from the example file alone will silently get the hardcoded default instead of whatever they meant to set.

Only the API key/model pair matching the active `LLM_PROVIDER` needs a real value — the other two providers' settings can stay entirely empty with no effect on anything. `ActiveAPIKey()` and `ActiveAPIKeyEnvVar()` are the two methods that map "which provider is active" to "which key/env-var-name matters right now" — `cmd/server/main.go`'s startup warning and any future caller both go through these rather than each re-implementing the same three-way switch.

## Three kinds of "wrong," three different responses, all deliberate

- **Missing API key** (for the active provider): the server still starts. This is specifically so the rest of the system — dataset loading, the `/api/v1/health` endpoint, everything that doesn't need to call out to an LLM — can be verified independently of having a real key configured. Every `POST /api/v1/cases` request will fail with a clear, specific error naming exactly which env var is missing (see each client's own check in `internal/extract`) until the key is set, but that failure is scoped to exactly the requests that need it, not the whole process.
- **Malformed `RATE_LIMIT_PER_MINUTE`** (non-numeric, zero, or negative) **or an unrecognized `LLM_PROVIDER`**: `Load()` returns an error and the caller (`cmd/server/main.go`) fails to start at all. The reasoning for the rate limit: a bad value isn't a "some features degrade gracefully" situation — it's silently either an unenforced limit (dangerous: this system's rate limiter is the main defense against a runaway LLM bill, see below) or an unusably strict one, and either failure mode is better caught at deploy time by a loud crash than discovered mid-incident. The reasoning for an unrecognized provider is the same shape: a typo like `LLM_PROVIDER=groc` should fail loudly at startup, not fall through to some undefined zero-value behavior that only surfaces as a confusing failure the first time a real request comes in.
- **A recognized `LLM_PROVIDER` whose own key happens to be unset**: falls under the first case above — the *provider selection itself* is validated strictly (must be a real, spelled-correctly choice), but that specific provider's *credentials* follow the same "start anyway, fail scoped-and-clear at request time" rule as `ANTHROPIC_API_KEY` always has.

This is the general pattern worth following if you add a new setting: ask whether an unset or invalid value is *safe to run with in a degraded way* (missing API key: yes, scoped failure) or *actively dangerous or ambiguous to run with silently* (bad rate limit, unrecognized provider: no) — and fail loudly at `Load()` time for the second kind, not downstream where the cause is harder to trace.

## `RateLimitPerMinute`: why this specific setting is called out as a cost control, not just an abuse control

Every `POST /api/v1/cases` request triggers exactly one paid call to whichever LLM provider is configured (`internal/extract`). Most rate limiters exist to protect availability; this one's doc comment is explicit that it exists to protect the hosting bill first — a bug that causes retry storms, or genuinely abusive traffic, translates directly and immediately into API spend without this limit in place. Worth keeping even on Groq's or Gemini's free tiers, which still have a rate ceiling somewhere this control is what keeps a bug from silently burning through. See `internal/api/middleware.go` (and `internal/api/README.md`) for where this setting is actually enforced.

## If you're extending this package

- **Adding a new env var**: add the field to `Config`, read it in `Load()` (via `getEnv` for a defaulted string, or the `RATE_LIMIT_PER_MINUTE` pattern for something requiring parsing/validation), and add it to `.env.example` with the same explanation-of-defaults comment style that file already uses — in the same change, not a follow-up.
- **Adding a fourth LLM provider**: add it to `validLLMProviders`, add its key/model fields to `Config` and `Load()` (and to `ActiveAPIKey`/`ActiveAPIKeyEnvVar`'s switches), add its defaults to `.env.example` — then see `internal/extract/README.md`'s own "adding a fourth provider" note for the client-side half of this.
- **This file is part of the codebase's documentation convention** — see the repo root `README.md`. Keep it in sync with the code, in the same change.
