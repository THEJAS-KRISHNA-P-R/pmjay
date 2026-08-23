# `internal/api`

The HTTP layer. Ties every other package together into four real endpoints plus a health check, using Go 1.22's standard-library `http.ServeMux` — no router dependency, no web framework. This is the package with the most test coverage by count (31 tests) because it's where every other package's behavior actually gets exercised end-to-end.

## Files

| File | What it does |
|---|---|
| `router.go` | `NewRouter` — builds the route table and wraps it in the middleware chain. |
| `handlers.go` | `Server` (the dependency container) and the five handler functions. |
| `middleware.go` | Panic recovery, structured logging, CORS, and the per-IP rate limiter. |
| `dto.go` | The wire-format types (`IntakeRequest`, `CaseResponse`, `AddEvidenceRequest`, `ErrorResponse`) and `caseRecordToResponse`, the one function that converts internal storage shape to API shape. |

## The four endpoints

| Method & path | Handler | Costs money? | Notes |
|---|---|---|---|
| `POST /api/v1/cases` | `handleIntake` | **Yes** — one LLM call | The only rate-limited endpoint. |
| `GET /api/v1/cases/{id}` | `handleGetCase` | No | |
| `GET /api/v1/cases/{id}/document` | `handleGetCaseDocument` | No | Same case, rendered as a PDF via `internal/document` instead of JSON. `Content-Disposition: inline` so the browser's own viewer opens it (see the handler's doc comment for why not `attachment`). |
| `POST /api/v1/cases/{id}/evidence` | `handleAddEvidence` | No | Requires at least one of `staff_name`/`approx_time`/`note`. |
| `GET /api/v1/health` | `handleHealth` | No | No LLM call, no store access — safe for a load balancer to poll frequently. |

Full request/response shapes and status codes: `docs/API.md`. This file covers the *why*, not the wire format.

## `Server`: why it's a plain struct of interfaces, not a framework's app object

```go
type Server struct {
    Dataset   *hbp.Dataset
    Extractor extract.Extractor
    Store     store.Store
    Logger    *slog.Logger
}
```

Every field beyond `Dataset` is an interface (`Extractor`, `Store`) or a stdlib type (`*slog.Logger`). `cmd/server/main.go`'s `newExtractor` constructs one of three real `Extractor` implementations depending on `LLM_PROVIDER` (`extract.NewClaudeClient`, `NewGroqClient`, or `NewGeminiClient` — see `internal/config/README.md` and `internal/extract/README.md`), and `store.NewFileStore` for `Store`; every test in this package constructs one with fakes (`extract.NewFakeClient()`, `store.NewMemStore()`) instead, regardless of which real provider is configured. This is what makes the entire HTTP layer — including middleware, routing, and every handler's actual logic — testable without network access or a live API key. If you add a new dependency a handler needs, add it as a field here, typed as an interface if there's any real implementation choice involved, not as a global or a direct concrete-type import.

## `handleIntake`: the one request that costs money, and how it fails

Walk through this handler once if you're going to touch it, because it's where several other packages' contracts meet:

1. Validate description length (`minDescriptionLength = 5`, `maxDescriptionLength = 4000`) — the lower bound specifically exists to avoid spending an LLM call on an obviously-too-short input; the frontend's `IntakeForm` duplicates this exact check client-side for the same reason (see `frontend/app/components/README.md` — and note the duplication is a real, acknowledged coupling: if this number changes here, it needs to change there too, or the two layers silently disagree about what's valid).
2. `retrieval.Retrieve` / `retrieval.RetrieveExclusions` narrow the full dataset to a short candidate list (zero cost, see `internal/retrieval/README.md`).
3. Both candidate lists get mapped down to the trimmed `CandidatePackageInfo`/`CandidateExclusionInfo` shapes — this is where internal-only fields (`ConfidenceNotes`, `Verified`, `SourceNote`) get left behind before anything reaches a model prompt.
4. `s.Extractor.Extract` — the one paid call. **If this fails**, the handler does not just return a bare error. It returns a 502 with `FallbackGuidance` set to `response.CareFirstText` plus the PMJAY helpline number (14555). This is a direct code-level consequence of Section 10's care-first rule being written as an absolute ("Always") — an infrastructure failure is explicitly not treated as an exception to it. A family whose request failed for a boring technical reason still gets the one thing that's always safe to tell them, plus a human fallback that doesn't depend on this system working at all.
5. `tiering.Decide` → `response.Build` — plain Go, no further model calls, see those packages' own READMEs.
6. The built response gets flattened into a `store.CaseRecord` and persisted. **If persistence fails here, the handler still returns the response successfully** — logged as an error, but the family isn't denied an answer they were already correctly given just because a disk write failed. Only follow-up actions (fetching this case again later, appending evidence) would be affected by that specific failure, not the answer itself.

## `middleware.go`: four concerns, each doing exactly one thing

Applied in this order (see `router.go`): `recoverMiddleware` (outermost) → `loggingMiddleware` → `corsMiddleware` → route-specific `rateLimitMiddleware` (only on the intake endpoint, innermost).

- **`recoverMiddleware`**: a panic in any single handler becomes a clean 500, not a crashed process taking down every other in-flight request. This exists for the same reason the sibling YourFee project needed a dedicated goroutine-panic-recovery fix in production — an unrecovered panic's blast radius should never extend past the one request that caused it.
- **`corsMiddleware`**: allowlist only, no wildcard, ever. The reasoning is explicit in the source comment: the intake endpoint costs real money per call, so an open CORS policy would be a standing invitation for an unrelated website to spend this system's API budget on its visitors' behalf.
- **`rateLimiter`**: a simple in-memory per-IP token bucket, deliberately not a distributed one (no Redis, no external service) — see `internal/config/README.md`'s note on `RateLimitPerMinute` being a cost control first, an abuse control second. `clientIP` prefers the *last* entry in `X-Forwarded-For` — the one a trusted reverse proxy itself appends, not whatever a client sent — and falls back to the raw connection address when the header is absent. This system is designed to run behind exactly one such proxy (Caddy — see `docs/DEPLOYMENT.md`, including the operational requirement that makes this safe: the Go process must never be reachable except through it). Stale buckets are swept periodically so a caller that varies its rate-limit key can't grow this map without bound.
- **`loggingMiddleware`**: one structured `slog` line per request (method, path, status, duration) — plain stdlib, no logging framework, matching this backend's zero-third-party-dependency stance throughout.

## `dto.go`: why API shapes are separate types from storage shapes

`CaseResponse` is not `store.CaseRecord` with JSON tags — it's its own type, built by `caseRecordToResponse`. This is the same "don't let an internal shape leak into a public boundary" principle `internal/response`'s unexported `FamilyResponse` fields enforce one layer earlier (see `internal/response/README.md`). If `store.CaseRecord` ever needs an internal-only field (an audit flag, a processing-version marker), adding it doesn't automatically expose it over the API — `caseRecordToResponse` is the one place that decides what actually crosses that boundary, and it has to be updated deliberately for a new field to reach a client at all.

## If you're extending this package

- **Adding a new endpoint**: register it in `router.go`, decide deliberately whether it needs `rateLimitMiddleware` (does it call the LLM, directly or indirectly?), and add a handler in `handlers.go` following the existing pattern (validate → call downstream packages → `writeJSON`/`writeError`).
- **Adding a field to any response**: it needs to exist on the internal type (`store.CaseRecord` or `response.FamilyResponse`), then get threaded through `caseRecordToResponse` into the matching DTO field, then get mirrored in `frontend/lib/types.ts` (hand-maintained, not codegenerated — see `frontend/lib/README.md`) and whatever component actually renders it.
- **This file is part of the codebase's documentation convention** — see the repo root `README.md`. Keep it in sync with the code, in the same change.
