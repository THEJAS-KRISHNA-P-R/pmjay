# `internal/api`

The HTTP layer. Ties every other package together into seven endpoints, using Go 1.22's standard-library `http.ServeMux` — no router dependency, no web framework. This is the package with the most test coverage by count (69 tests) because it's where every other package's behavior actually gets exercised end-to-end.

## Files

| File | What it does |
|---|---|
| `router.go` | `NewRouter` — builds the route table and wraps it in the middleware chain. Takes an `Options` struct rather than a long argument list. |
| `handlers.go` | `Server` (the dependency container) and the seven handler functions. |
| `middleware.go` | Panic recovery, structured logging, security headers, `no-store`, CORS, and the two rate limiters. |
| `dto.go` | The wire-format types (`IntakeRequest`, `CaseResponse`, `PackageSearchResult`, `AddEvidenceRequest`, `ErrorResponse`) and the conversions from internal shapes. |

## The seven endpoints

| Method & path | Handler | Costs money? | Notes |
|---|---|---|---|
| `POST /api/v1/cases` | `handleIntake` | **Yes** — one LLM call | The only intake-rate-limited endpoint, and the only one under a request timeout. |
| `GET /api/v1/cases/{id}` | `handleGetCase` | No | General rate limiter only. |
| `DELETE /api/v1/cases/{id}` | `handleDeleteCase` | No | 204 on success, 404 for an unknown ID. The backend half of the DPDP erasure right. |
| `GET /api/v1/cases/{id}/document` | `handleGetCaseDocument` | No | Same case, rendered as a PDF via `internal/document` instead of JSON. `Content-Disposition: inline` so the browser's own viewer opens it (see the handler's doc comment for why not `attachment`). |
| `POST /api/v1/cases/{id}/evidence` | `handleAddEvidence` | No | Requires at least one of `staff_name`/`approx_time`/`note`. |
| `GET /api/v1/packages` | `handlePackages` | No | Deterministic search over the embedded dataset. Works with no provider key configured. |
| `GET /api/v1/health` | `handleHealth` | No | No LLM call, no store access — safe for a load balancer to poll frequently. Reports `extraction_ready` separately from `status`. |

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
4. `s.Extractor.Extract` — the one paid call. **If this fails**, the handler does not just return a bare error. It distinguishes three failures, each with `FallbackGuidance` set to `response.CareFirstText` plus the PMJAY helpline number (14555): `503` when the deployment has no provider key (`extract.ErrNotConfigured`), `504` when the provider exceeded `INTAKE_TIMEOUT_SECONDS`, and `502` otherwise. This is a direct code-level consequence of Section 10's care-first rule being written as an absolute ("Always") — an infrastructure failure is explicitly not treated as an exception to it. A family whose request failed for a boring technical reason still gets the one thing that's always safe to tell them, plus a human fallback that doesn't depend on this system working at all. The status distinction matters for the retry: a missing key will not fix itself, a timeout might.
5. `tiering.Decide` → `response.Build` — plain Go, no further model calls, see those packages' own READMEs. The whole request runs under a context bounded by `Options.IntakeTimeout`, which is intentionally shorter than the server's write timeout.
6. The built response gets flattened into a `store.CaseRecord` and persisted. **If persistence fails here, the handler still returns the response successfully** with `persisted: false` — logged as an error (with the case ID redacted), but the family isn't denied an answer they were already correctly given just because a disk write failed. Only follow-up actions (fetching this case again later, appending evidence, deleting it) would be affected by that specific failure, not the answer itself.

## `handlePackages`: why an unverified record loses both its rate and its code

`packageToSearchResult` is deliberately the only path from an `hbp.Package` to the public API. It returns early on `!p.Verified`, which omits every rate field **and** the package code. A placeholder identifier such as `SEED-ORTHO-001` exists in no government schedule; publishing it next to a real procedure name would invite a family or an agent to quote it at a billing desk as if it were a citation. The name, specialty, and pre-auth flag are still useful, so those remain. If you add a field to `PackageSearchResult`, decide explicitly which side of this line it falls on, and add the negative test (`TestPackages_UnverifiedRecordNeverExposesARate`, `TestPackages_UnverifiedRecordOmitsTheCodeInRawJSON`).

## `middleware.go`: small concerns, each doing exactly one thing

Applied in this order (see `router.go`): `recoverMiddleware` (outermost) → `loggingMiddleware` → `securityHeadersMiddleware` → `noStoreMiddleware` → `corsMiddleware` → route-specific rate limiting and, on intake only, timeout + concurrency limiting (innermost).

- **`recoverMiddleware`**: a panic in any single handler becomes a clean 500, not a crashed process taking down every other in-flight request. An unrecovered panic's blast radius should never extend past the one request that caused it.
- **`securityHeadersMiddleware`** / **`noStoreMiddleware`**: `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer`, and `Cache-Control: no-store` on every response. A case response is health and financial data; no intermediary should keep a copy.
- **`corsMiddleware`**: allowlist only, no wildcard, ever. The reasoning is explicit in the source comment: the intake endpoint costs real money per call, so an open CORS policy would be a standing invitation for an unrelated website to spend this system's API budget on its visitors' behalf.
- **Two rate limiters**: a deliberately tight intake limiter (1/minute, 20/hour by default) around the one paid endpoint, and a flat, more generous general limiter shared by every other store-touching route. Both are in-memory per-IP token buckets, deliberately not distributed (no Redis, no external service) — see `internal/config/README.md`'s note on `RateLimitPerMinute` being a cost control first, an abuse control second. `clientIP` only reads `X-Forwarded-For` / `X-Real-IP` when `TRUST_PROXY_HEADERS=true`; with the default `false` it uses the connection address, because a directly reachable caller can forge a proxy header and mint unlimited rate-limit identities. Behind Caddy (or any proxy that overwrites the header) set it true; see `docs/DEPLOYMENT.md`. Stale buckets are swept periodically so a caller that varies its rate-limit key can't grow this map without bound.
- **`loggingMiddleware`**: one structured `slog` line per request (method, path, status, duration) — plain stdlib, no logging framework, matching this backend's zero-third-party-dependency stance throughout. Case IDs are redacted to a short prefix before logging: the ID is the capability that grants access to a case.

## `dto.go`: why API shapes are separate types from storage shapes

`CaseResponse` is not `store.CaseRecord` with JSON tags — it's its own type, built by `caseRecordToResponse`. This is the same "don't let an internal shape leak into a public boundary" principle `internal/response`'s unexported `FamilyResponse` fields enforce one layer earlier (see `internal/response/README.md`). If `store.CaseRecord` ever needs an internal-only field (an audit flag, a processing-version marker), adding it doesn't automatically expose it over the API — `caseRecordToResponse` is the one place that decides what actually crosses that boundary, and it has to be updated deliberately for a new field to reach a client at all.

## If you're extending this package

- **Adding a new endpoint**: register it in `router.go`, decide deliberately whether it needs `rateLimitMiddleware` (does it call the LLM, directly or indirectly?), and add a handler in `handlers.go` following the existing pattern (validate → call downstream packages → `writeJSON`/`writeError`).
- **Adding a field to any response**: it needs to exist on the internal type (`store.CaseRecord` or `response.FamilyResponse`), then get threaded through `caseRecordToResponse` into the matching DTO field, then get mirrored in `frontend/lib/types.ts` (hand-maintained, not codegenerated — see `frontend/lib/README.md`) and whatever component actually renders it.
- **This file is part of the codebase's documentation convention** — see the repo root `README.md`. Keep it in sync with the code, in the same change.
