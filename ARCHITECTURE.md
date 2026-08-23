# Architecture

## System shape

```
Family types a description (any language mix)
        │
        ▼
┌───────────────────┐     zero API cost — cheap Go-side keyword
│ internal/retrieval │ ←── scoring narrows ~hundreds of packages
└───────────────────┘     down to ~20 plausible candidates
        │
        ▼
┌───────────────────┐     the ONE paid call per request — structured
│  internal/extract  │ ←── output, any of 3 providers (LLM_PROVIDER)
└───────────────────┘
        │
        ▼
┌───────────────────┐     plain, deterministic Go — no model calls,
│  internal/tiering  │ ←── fully offline-testable, ~95% test coverage
└───────────────────┘
        │
        ▼
┌───────────────────┐     the care-first message is set before the
│  internal/response │ ←── tier switch even runs — structurally
└───────────────────┘     impossible to omit (see SAFETY_DESIGN.md)
        │
        ▼
┌───────────────────┐     file-backed JSON, atomic writes,
│   internal/store   │ ←── zero hosted database
└───────────────────┘
        │
        ▼
   internal/api (stdlib net/http, Go 1.22 ServeMux)
        │                          │
        ▼                          ▼  GET .../document
   Next.js frontend        internal/document — the same
                            stored case, rendered as one
                            downloadable/printable PDF
                            instead of JSON (zero third-
                            party deps here too — see
                            that package's own README)
```

This mirrors the source spec's own layering (Section 15): a data layer, an extraction/matching layer where a language model is genuinely necessary, a deterministic response/explanation layer, and an escalation layer — kept as separate packages on purpose, so each one is testable in isolation and the one expensive, non-deterministic component (the LLM call) is the smallest possible surface area. `internal/document` is the one addition to this shape that isn't a sequential pipeline stage — it's a second, on-demand rendering of whatever `internal/store` already has for a case, not something every request passes through.

## The backend has zero third-party Go dependencies. Here's why, and why that turned out to be the right call anyway.

This was discovered, not planned: the build environment's network egress is allowlisted to package registries and GitHub, not to `proxy.golang.org` (Go's default module proxy) or to `golang.org` itself. Testing confirmed that `GOPROXY=direct` can fetch modules hosted directly on `github.com`, but `gin-gonic/gin` pulls in `golang.org/x/*` packages transitively (via `bytedance/sonic`, its fast-JSON dependency), which are unreachable. That ruled out Gin specifically.

Rather than fight the constraint, the backend leans into it:

- **HTTP routing**: Go 1.22's standard-library `http.ServeMux` gained method-and-pattern matching (`mux.HandleFunc("POST /api/v1/cases", ...)`) — genuinely sufficient for an API surface this size, and it means zero router dependency.
- **UUIDs**: a ~15-line `crypto/rand`-based UUID v4 generator (`internal/store/id.go`) instead of `google/uuid`.
- **Persistence**: a file-backed JSON store instead of a SQL driver (more on this below).
- **The Claude API client**: a direct `net/http` call to `api.anthropic.com/v1/messages` instead of the Anthropic SDK — one well-understood HTTP call is easier to audit than a dependency pulled in for a single endpoint, SDK or not.

The result is a real, if accidental, engineering benefit: **zero supply-chain surface**, a smaller attack surface to audit (directly relevant given this codebase's sibling project has an active security-audit workstream), and — this is the part that matters most for "reduce server bills" — a fully static binary that cross-compiles to any target with one command and no C toolchain:

```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" ./cmd/server
```

Verified during this build: **5.7 MB**, stripped, for `linux/arm64` — directly matching a t4g.micro's Graviton architecture. No cross-compiler, no Docker buildx multi-stage dance required (though a Dockerfile is included for convenience — see `backend/Dockerfile`).

## Why no database

Case volume here is genuinely small: one record per family interaction, a handful of evidence entries appended per case. That does not need a hosted database engine to run reliably, and running one anyway would be the single largest avoidable recurring cost in this system's hosting bill.

`internal/store` is an interface (`Store`) with two implementations:

- `MemStore` — pure in-memory, used in tests.
- `FileStore` — wraps `MemStore` for fast reads, flushes the full snapshot to a single JSON file on every write using atomic write-then-rename (write to `cases.json.tmp`, then `os.Rename` over `cases.json` — either the rename completes fully or not at all, so a crash mid-write can never corrupt the file).

If real scale ever demands it, `Store` is the seam: a `PostgresStore` or a Supabase-backed implementation is a self-contained addition that touches nothing else in the codebase.

## Cost breakdown

Everything below is the *marginal* cost of running this system — assuming it's colocated on infrastructure the admin already runs for a separate project (the existing t4g.micro, already paid for), which is the cheapest realistic path for a hackathon-stage product.

| Component | Cost |
|---|---|
| Backend hosting (colocated on existing t4g.micro) | $0 marginal (binary is 5.7 MB, memory footprint is trivial next to an existing Go service) |
| Case storage | $0 (flat file, no hosted DB) |
| Frontend hosting (Vercel free tier) | $0 |
| LLM inference | Only real variable cost — one call per family interaction to whichever provider `LLM_PROVIDER` selects, with the prompt kept small specifically by the retrieval pre-filter (see below). Groq and Gemini's free tiers can cover this at $0 too — see "Multi-provider extraction" below. |

If run as a fully standalone deployment instead (not colocated), a Fly.io or Oracle Cloud free-tier instance covers the backend at $0 as well — see `docs/DEPLOYMENT.md` for both paths with concrete steps.

### Why the retrieval pre-filter is a cost decision, not just a latency one

`internal/retrieval` does cheap, zero-API-cost keyword scoring in Go to narrow the full package list down to ~20 candidates *before* the LLM ever sees the request. On the real ~1,900-code HBP master list (this repo ships a smaller seed set — see `docs/DATA_SOURCES.md`), sending the entire dataset as context on every single request would make token cost scale with catalog size, not with query complexity. The pre-filter is deliberately recall-biased (it would rather return too many candidates than risk excluding the right one — see the doc comment in `internal/retrieval/retrieval.go` on "failing open") — precision is the LLM's job, recall is retrieval's job, and keeping that boundary sharp is what keeps the one paid call in this whole system cheap and small.

### Why Claude Haiku specifically — and why this is now a real, supported choice, not a hardcoded assumption

The extraction task is bounded classification over a short, pre-filtered candidate list — not open-ended generation. It doesn't need a large model to do well, and this system's entire per-query cost is dominated by that one call. See `internal/extract/claude_client.go`; the model is configurable via `CLAUDE_MODEL` if a different tradeoff is ever wanted.

## Multi-provider extraction: why `Extractor` has three real implementations, not one

Every extraction call was originally hardcoded to Claude. `LLM_PROVIDER` (`internal/config`) now selects between three real backends — `anthropic` (default, unchanged behavior), `groq`, or `gemini` — each a full `Extractor` implementation in `internal/extract` (`ClaudeClient`, `GroqClient`, `GeminiClient`). This section is about *why* that's a real architectural choice worth making, not just a list of what changed; see `internal/extract/README.md` and `internal/config/README.md` for the file-level detail.

**Why this is safe to do at all.** The extraction *contract* — what fields the model must produce, what each one means, `validatePayload`'s rules for rejecting a malformed response — was already fully decoupled from Claude specifically, expressed as a JSON Schema (`extractMatchToolSchema`) and a shared payload type (`toolExtractionPayload`), not as Claude-specific prompt engineering. Adding two more providers didn't require touching that contract at all — each new client either sends the same schema directly (Groq, which accepts standard JSON Schema the same way Anthropic's tool-calling does) or through a mechanical format conversion (Gemini, whose schema format spells types in uppercase — see `internal/extract/gemini_client.go`'s `convertToGeminiSchema`). The system prompt, `systemPrompt`, is sent verbatim to all three — it's an encoding of the source spec's Appendix AA, not model-specific phrasing.

**Why a team might genuinely want this**, beyond "more options is nice": this system's own cost-consciousness (see above) applies just as much to a real deployment as it did to the original hackathon build. Groq's free tier and Gemini's AI Studio free tier both cover this system's actual traffic pattern (one short classification call per family interaction) comfortably for a low-volume deployment, where a paid Anthropic key is a real, if small, ongoing cost. This isn't a claim that Groq or Gemini produce identical extraction quality to Claude — they're different models, and nothing about this change asserts otherwise — it's a claim that the *option* to trade some quality for zero marginal cost is a legitimate one for this system's actual users to make, not something the architecture should foreclose by hardcoding one vendor.

**What's shared across all three clients, and what's deliberately not.** Shared: `buildUserContent` (candidate/exclusion serialization), `toolExtractionPayload` and `validatePayload` (parsing and the real safety-net validation), `parseRetryAfterSeconds`, `truncate` — all provider-agnostic, all living in `claude_client.go` and used by the other two files directly, not duplicated. *Not* shared: each client's retry/backoff loop, written independently per provider despite being structurally similar. That's a deliberate choice following this codebase's existing precedent (`internal/retrieval`'s `scorePackage`/`scoreExclusion`, kept parallel rather than merged into one generic function) — a reader should be able to understand any one client's `Extract` method start to finish without jumping to a shared retry-orchestration file to see what it actually does on a 429.

**What this doesn't change.** `internal/api` still depends only on the `Extractor` interface — it has no idea which provider is behind it, and none of `internal/tiering` or `internal/response` changed at all. `FakeClient` (used by every other package's tests) is unaffected. The only new code path in `cmd/server/main.go` is `newExtractor`, a four-way switch on `cfg.LLMProvider` that config.Load() has already validated is one of the three real options.

## Why the extraction/tiering boundary is exactly where it is

Per the source spec's own Section 58 ("AI where genuine language ambiguity exists, deterministic auditable logic everywhere after"), `internal/extract` is the *only* package that calls a model. Everything downstream — the green/amber/red/mixed/handoff decision, the confidence-gap rule, the pending-vs-denied cross-check — is plain Go with no model calls, which is what makes `internal/tiering` reach 94.5% test coverage and be fully testable offline. `internal/extract/fake_client.go` is a deterministic test double that lets the entire pipeline, including the HTTP layer's integration tests, run without a live API key.

## Frontend stack

Next.js 16 (App Router) + TypeScript + Tailwind v4, deployed as a static-friendly app that talks to the Go backend over a small, typed API surface (`lib/types.ts` mirrors `backend/internal/api/dto.go` by hand — the API is small and stable enough that a codegen step would be more moving parts than it's worth).

**Fonts are self-hosted from local files, not `next/font/google`.** Same network-sandbox story as the Go dependencies: this build environment doesn't reach `fonts.googleapis.com`. The actual OFL-licensed font files were fetched directly from Google's own font repository on GitHub and are checked into `frontend/app/fonts/`, loaded via `next/font/local`. Functionally identical to `next/font/google`'s output (self-hosted, zero runtime CDN dependency, zero layout shift) — this just does the fetch at build-prep time instead of first-build time.

**Atkinson Hyperlegible**, not a generic system-ui stack, for all functional text. This is a functional choice, not a stylistic one: it's a typeface published by the Braille Institute of America specifically for readers who find text difficult to read accurately — a real requirement given this product's named user (Section 5 of the source spec), not an aesthetic preference. Fraunces is used only for the wordmark; every word a family actually has to read to understand their situation stays in the hyperlegible face.

## What was deliberately not built

Matching the source spec's own Section 15.5 ("deliberately out of scope") and Section 16 (build-priority order):

- Full ~1,900-code HBP coverage (seed dataset only — see `docs/DATA_SOURCES.md`)
- Live PMJAY pre-authorisation status checking (no public API exists for this)
- Automated CGRMS complaint submission (drafts only, now rendered as a real printable/downloadable PDF via `internal/document` rather than just an on-screen text box — but the family still submits it themselves via the official Ayushman App; this tool still doesn't, and isn't meant to, submit anything on the family's behalf)
- Live NALSA case-routing integration (the handoff panel gives the real, verified toll-free number — 15100 — and a prepared summary, not an API handoff)
- WhatsApp or other no-install distribution channel (web app only for this build; see the source spec's Appendix Q for the roadmap case)
- Deep Malayalam-language UI localization of the generated response *templates* (the system understands Malayalam/English code-mixed *input* natively via the LLM step, but the hard-coded response templates in `internal/response/templates.go` are English-only — see `docs/DATA_SOURCES.md`'s note on why this wasn't machine-translated for a safety-critical, medical/legal-adjacent context without expert review)
