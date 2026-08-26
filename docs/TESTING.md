# Testing

This document does two things: maps what's actually automated to the spec's own validation protocol (Appendix R, five hypotheses H1–H5), and says plainly which parts of that protocol code cannot answer at all, no matter how well it's tested. Conflating those two categories — implying a field-study question has been "handled" because a related unit test passes — would be exactly the kind of overclaiming this project has otherwise tried hard to avoid.

## Reproducing the numbers below

```bash
cd backend
go build ./...              # compiles clean
go vet ./...                 # no vet warnings
go test ./... -cover -race   # 268 tests, 10 packages, race detector on

cd ../frontend
npm test                     # 77 tests, 12 files (see "Frontend" section below)
```

No live `ANTHROPIC_API_KEY` is required for any of this — see "Why this is testable offline at all" below.

## What's automated: 268 backend tests, 10 packages

| Package | What it covers | Tests | Coverage |
|---|---|---:|---:|
| `internal/hbp` | Dataset schema, `go:embed` loader, `validate()`'s full invariant set, `rate_type` extension | 32 | 91.7% |
| `internal/retrieval` | Zero-API-cost keyword pre-filter, package and exclusion scoring | 15 | 100.0% |
| `internal/extract` | LLM extraction across all 3 providers (`ClaudeClient`, `GroqClient`, `GeminiClient`), raw `net/http`, no SDKs, plus bounded retry/backoff for transient failures on each | 49 | 82.3% |
| `internal/tiering` | Deterministic green/amber/red/mixed/handoff decision logic | 24 | 94.5% |
| `internal/response` | Templated response text, the care-first and disclaimer structural guarantees, tiered/per-diem citation rendering | 21 | 95.8% |
| `internal/store` | File-backed case persistence, atomic writes | 17 | 86.8% |
| `internal/document` | PDF rendering — encoding/layout engine, all 5 outcomes, structural PDF validity, disclaimer placement | 51 | 99.6% |
| `internal/config` | Env-based configuration, LLM provider selection and validation | 13 | 96.9% |
| `internal/api` | HTTP handlers, middleware, routing, request-size limits, IP-spoofing-resistant rate limiting | 38 | 91.9% |
| `cmd/server` | Process orchestration — config→dataset→store→extractor→server wiring, graceful shutdown, every failure path including `os.Exit(1)` via subprocess re-exec | 8 | 95.6% |

Coverage is highest exactly where it matters most — `tiering` and `response`, the two packages that decide and phrase what a family is told, sit at 94–96%, `retrieval` sits at a genuine 100%, and `document`, `config`, and `hbp` all sit at 91–99.6%. `extract` at 82.3% is the one package where a straight percentage reads lower than its actual test rigor suggests: adding two full provider clients (`GroqClient`, `GeminiClient`) roughly doubled the package's line count, and each new client's untested surface is the same shape as `ClaudeClient`'s always was (`WithGroqHTTPClient`/`WithGeminiHTTPClient` — options that exist for tests to use, not code paths tests exercise themselves) — not new, undertested business logic. Every package's remaining gap beyond that is a specific, named, principled exception rather than an untested surface — not "we didn't get to it," but "here's exactly why this one line can't be exercised without making the production code worse to satisfy a test":

- **`hbp.Load()`'s and `cmd/server`'s `run()`'s error branches for `hbp.Load()` failing**: this reads a `//go:embed`-compiled dataset — the files are baked into the binary at build time, so `Load()` returning an error is a build-time defect, not a runtime condition any test can trigger without hand-editing the embedded JSON and rebuilding, which isn't a legitimate per-test technique. `validate()` itself — the actual logic that would catch a malformed record — is fully tested directly against synthetic bad fixtures (see the `TestValidate_*` tests); it's specifically the "what if the embedded file read itself fails" branch that's structurally unreachable.
- **`document`'s defensive `panic`** in the xref-table writer: unreachable given the package's own fixed, gap-free object-numbering scheme — see `internal/document/serialize.go`.
- **`api`'s `handleGetCaseDocument` error branch**: `document.BuildCase` has no failing code path today given a valid `CaseRecord` — see that function's own doc comment. The branch stays in the code as a not-yet-needed safety net, not deleted for coverage's sake.

### Multi-provider LLM support, specifically

`internal/extract` gained `GroqClient` and `GeminiClient` alongside the original `ClaudeClient`, selected at runtime by `LLM_PROVIDER` (see `internal/config/README.md` and `ARCHITECTURE.md`'s "Multi-provider extraction" section for the design rationale). Test coverage mirrors `ClaudeClient`'s existing rigor deliberately, not by accident — the same categories of test for each new client: happy path, missing API key, non-retryable failures (4xx) failing fast without burning the retry budget, retryable failures (429/5xx) succeeding after retry, exhaustion after max attempts, context cancellation, `Retry-After` honored under a cap. Two things worth calling out specifically:

- **`convertToGeminiSchema` (Gemini's schema-format converter, since its structured-output schema spells types in uppercase — a genuine wire-format difference from the standard JSON Schema Claude and Groq both use directly) has its own dedicated unit tests**, independent of the full HTTP round-trip — including one that runs the *actual production schema* through the converter, not a simplified stand-in, specifically because a bug in a simplified test schema could pass while the real schema's actual nesting shape still broke.
- **`internal/config`'s provider-selection logic and `cmd/server`'s `newExtractor` factory both have direct tests** (`TestConfig_ActiveAPIKey_SelectsByProvider`, `TestNewExtractor_SelectsClientByProvider`, etc.) rather than relying on indirect exercise through `run()`'s existing tests — the same "a package's own test suite shouldn't depend on being exercised by a caller elsewhere" principle the cross-package coverage investigation below surfaced, applied going forward rather than only in retrospect.

**What multi-provider support does *not* mean**, worth stating plainly: none of Groq's or Gemini's real API behavior has been validated against this system — every test above runs against a mocked `httptest.Server` matching each provider's *documented* wire format, the same validation depth Claude had before this session (and still has). That gap is real and open regardless of which provider a deployment chooses.

### The coverage investigation, and the security/reliability hardening pass before it

Two more test-writing efforts happened in earlier parts of this same overall body of work, each worth understanding on its own terms:

**The coverage push** (185 → 268 backend tests total, across all sessions) had a methodological finding worth remembering before ever trusting a per-package coverage number again: `go test ./pkg/...`'s coverage is silently incomplete for any package that's more heavily exercised by *callers* than by its own tests. `internal/hbp`'s `PackageByCode`/`ExclusionByCategory` and `internal/retrieval`'s `RetrieveExclusions` all showed as 0% in isolation despite being fully exercised by `internal/tiering`'s and `internal/api`'s test suites — `go test ./... -coverpkg=./...` (cross-package attribution) is what revealed the true picture (91.2% before any of this pass's new tests, not the low-60s/high-80s the siloed numbers implied). Real, dedicated tests were still written for both regardless of that true-but-hidden coverage — a package whose only real exercise comes from a caller elsewhere is one refactor away from silently losing that coverage entirely.

**The security/reliability hardening pass** (185 → 204, entirely from responding to a "make sure this handles real stuff" review, not new business logic):

- **`internal/api` (+7):** the rate limiter's `X-Forwarded-For` handling trusted the wrong (attacker-controlled) end of the header — `TestClientIP_ClientCannotSpoofLeadingForwardedForEntry` proves the fix directly, not just that the function returns *a* value. Two more cover the bucket-eviction fix for the unbounded-memory-growth side of the same bug. Three cover request body size limits (413 on oversized bodies, acceptance right at the limit, evidence-endpoint coverage).
- **`internal/extract` (+11, on `ClaudeClient` specifically, before the other two providers existed):** the Anthropic API call gained bounded retry/backoff for transient failures. Tests cover every path deliberately, not just the happy one: retry-then-succeed (rate-limit and transport-error variants), exhaustion after max attempts, non-retryable failures (4xx, validation errors) failing *fast* rather than burning the retry budget, `Retry-After` honored under a cap and ignored past it, and context cancellation mid-backoff stopping retries immediately.
- **`internal/response` and `internal/document` (+1 each):** the new `DisclaimerText` field is checked through the same exhaustive-across-every-outcome mechanism `CareFirstText` already had (`TestBuild_CareFirstMessageIsAlwaysPresent`, extended rather than duplicated), plus a PDF-level test confirming the disclaimer actually renders in the output bytes, and a negative test confirming it's correctly omitted when absent.

### `internal/document`'s testing approach, specifically

Most of its 50 tests are the same kind of thing every other package here does — unit tests on pure functions, table-driven where that fits. Two things about it are worth calling out because they go beyond what `go test` alone can be trusted to show for a binary output format:

- **`serialize_test.go` verifies PDF structural validity directly**, not just "the code ran without an error" — every cross-reference table offset is independently re-parsed from this package's own output and checked against the actual byte content at that position, rather than trusting `render()`'s internal offset-tracking to have gotten it right.
- **Independent, out-of-Go validation happened during development and is deliberately not part of the committed, automated suite**: generated sample PDFs were parsed with Python's `pypdf` (confirming extracted text matches what was drawn — rupee amounts included, correctly rendered as "Rs." rather than a broken glyph) and rasterized to PNG with `pdftoppm` for direct visual inspection, which is what actually caught this package's most serious bug (see `internal/document/README.md`'s "a real bug this package's own development caught" section — a background box and the heading drawn after it visibly overlapped, something no amount of passing unit tests would have surfaced, because the individual arithmetic operations were each individually correct). Worth redoing by hand after any substantial change to `serialize.go` or `pdf.go`; not repeatable with `go test` alone, and stated here rather than silently only claimed as "tested."

### Why this is testable offline at all

`internal/extract/fake_client.go` is a deterministic test double implementing the same `Extractor` interface the real Claude-backed client does — register an input string, get back a fixed `Result`, no network call, no API key, no nondeterminism. Combined with `store.NewMemStore()` for the API layer's tests, the entire pipeline — HTTP request in, tiered response out — is tested without touching a live model or a real filesystem. This is also what makes the docs/API.md examples reproducible: they were generated by running the real handlers against `FakeClient`, not hand-typed.

The one package that necessarily cannot be tested this way is the real Claude HTTP call itself (`internal/extract/claude_client.go`'s actual `net/http` round-trip to `api.anthropic.com`) — its tests cover request construction, response parsing, and error handling against a local `httptest.Server` standing in for the real endpoint, not the genuine model behavior on genuinely novel input. That gap is real and is one of the reasons H3 (below) ultimately needs cases beyond what a fixed test fixture can provide.

## H3 — tiered logic accuracy, including on cases it wasn't built for

The spec's own three-arm structure (adapted from earlier validation work on a different, cut idea in this same process) is implemented directly as three named groups of tests in `internal/tiering/tier_test.go`:

- **Control arm** — clean, unambiguous cases. `TestDecide_ControlArm_CleanGreen`, `TestDecide_ControlArm_CleanRed`.
- **Mechanical arm** — obvious-by-construction cases that still have to reach the right tier through real logic, not luck: `TestDecide_MechanicalArm_CardiacGreenWithPartialCoverageLanguage`, `TestDecide_MechanicalArm_OrthopaedicMixedCoveredAndCosmetic`, `TestDecide_MechanicalArm_OrganTransplantIsAmberNotRed`.
- **Ambiguous arm** — the arm the spec calls out as the one where a confidently-wrong answer is a safety failure, not just an accuracy shortfall: `TestDecide_AmbiguousArm_PendingPreauthMirrorsWorkedExample`, `TestDecide_AmbiguousArm_CloseCandidatesForceAmberRegardlessOfTopScore`, `TestDecide_AmbiguousArm_WideGapStillGreen`, `TestDecide_AmbiguousArm_LowConfidenceNeverAssertedAsGreenOrRed`, `TestDecide_AmbiguousArm_EmpanelmentEdgeCaseIsNotFalselyConfident`.

**What this does and doesn't establish.** These tests confirm the decision logic is correct and appropriately cautious against the specific scenarios it was written against — including the spec's own worked examples. They do not, and structurally cannot, establish accuracy against the full space of ways real families will actually phrase real situations. H3 as originally scoped is a measurement across real or realistic field cases with a defined accuracy threshold; what exists today is unit-level confirmation that the *logic* behaves correctly on a representative, hand-constructed set, which is necessary but not sufficient evidence for H3's actual claim. Extending the ambiguous arm's case bank is the highest-value next testing investment if this moves toward real deployment — see `docs/OPEN_QUESTIONS.md`.

## H4 — the tool never contributes to a treatment delay

Named by the spec as the single highest-stakes hypothesis in the whole protocol, and the one with the least ambiguity about whether it's code-testable: it is, directly, because it's a claim about generated text, not about real-world behavior change.

`TestBuild_H4_NoDelayEncouragingLanguageAcrossAllOutcomes` (`internal/response/builder_test.go`) implements the spec's own adversarial-scenario design: seven decisions across every outcome tier, each with a summary text engineered to tempt a hedging response, scanned against a 16-phrase delay-encouraging blocklist with negation-aware matching (so "do not let this delay care" doesn't false-positive against itself). Full mechanism described in `docs/SAFETY_DESIGN.md` §2. Per the spec's own kill condition for H4, this test treats a single unnegated match as a hard failure, matching the "hard stop, not a statistic" framing in Appendix R directly.

**What this does and doesn't establish.** This confirms the *generated text* never crosses the line, across every code path that currently exists. It does not confirm a real family, under real stress, reading this text on a real phone screen at a real billing counter, will actually act on it correctly — that's H2's territory, and H2 is not code-testable (below).

## H1, H2, H5 — not code-testable, and presented here as such rather than glossed over

Three of the five hypotheses are, by the spec's own design, questions about the real world that no amount of unit testing can answer:

- **H1** (is a wrongful denial or illegal demand common enough, at the individual-family level, to justify this) requires a real sample of real hospital visits — the spec specifically calls for measurement through partner ASHA workers or a legal-aid clinic, not self-report alone. No code artifact in this repository can produce that number.
- **H2** (does the tool change what a family actually does, not just what they understand) requires comparing real or simulated family behavior with and without the tool. `internal/response`'s tests confirm the tool *says* the right thing; they cannot confirm anyone acts on it.
- **H5** (do families set this up before a crisis, not during one) is, per the spec's own framing, "the one risk that cannot be closed by more design work" — it's the direct formalisation of the open question in Section 20.1 (see `docs/OPEN_QUESTIONS.md`), and the spec explicitly recommends testing it *first*, not last, if this moves forward, since it's both the cheapest of the five to test and the one that changes the most downstream if it fails.

Nothing in this codebase should be read as evidence toward any of these three. Section 33's interview guide is named in the spec as a lightweight first pass specifically at H5, now extracted into a standalone, ready-to-use document at `docs/VALIDATION_INTERVIEW_GUIDE.md`; it hasn't been run as part of this build.

## Frontend: full component coverage plus one page-level integration test

`frontend/package.json`'s `test` script (Vitest + React Testing Library + jsdom) now covers 74 tests across 12 files — up from 35/5:

```bash
cd frontend
npm test
```

What's covered:

- **`CareFirstBanner.test.tsx`** — the frontend half of the care-first guarantee `docs/SAFETY_DESIGN.md` describes on the backend side. Confirms the message renders unmodified, uses `role="alert"` so assistive technology announces it without the user having to find it, and doesn't crash or silently truncate.
- **`TierBadge.test.tsx`** — confirms every one of the 5 outcomes renders an icon, a label, *and* a description together, never color alone (`README.md`'s accessibility claim, Appendix K in the source spec) — plus a pairwise check that no two outcomes accidentally share an icon.
- **`TierPanel.test.tsx`** — the panel that composes `TierBadge` with the tier message and the optional citation; confirms the citation section is absent (not empty) when there's nothing to cite, and that multi-line messages keep their line breaks.
- **`ActionSteps.test.tsx`** — confirms steps render in order, numbered from 1, and that the component renders nothing at all (not an empty section) for a zero-length list — a second, independent guard behind the page-level conditional that already checks this.
- **`HandoffPanel.test.tsx`** — confirms the NALSA framing and the free-of-cost language are both present (a family mid-billing-dispute is exactly the audience likely to wonder if a phone number leads to another bill), that `tel:15100` is correct and distinct from Header's own `tel:14555`, and that an empty summary doesn't crash the panel.
- **`CopyableTextBox.test.tsx`** — the exact copied text matches the displayed text, the "Copied ✓" confirmation state appears and reverts after its timeout window (fake timers, with care taken to only fake timers and not the microtask queue the async clipboard call itself resolves through), and a rejected `navigator.clipboard.writeText` degrades to the plain "Copy" state instead of throwing — matching the source's own documented reasoning for that try/catch.
- **`CaseDocumentPanel.test.tsx`** — the "Download PDF" link points at the backend's `/document` endpoint with the case ID correctly URL-encoded, opens via `target="_blank"` + `rel="noopener"` (a real navigation to the browser's own PDF viewer, not a fetch/blob download — see the component's doc comment for why that distinction matters), and doesn't crash on an edge-case empty ID.
- **`Header.test.tsx`** — the brand link and the `tel:14555` helpline link, the one helpline number reachable before a case even exists.
- **`IntakeForm.test.tsx`** — client-side validation duplicates the backend's exact 5-character minimum (`internal/api/dto.go`) specifically to save a wasted, paid API round-trip on an obviously-too-short description; this suite is what actually catches that duplication silently drifting out of sync. Also covers the `ApiError` + `fallback_guidance` rendering path end to end (docs/API.md's documented 502 contract) and the example-prompt buttons.
- **`EvidenceForm.test.tsx`** — the at-least-one-field rule mirrored from `internal/api/handlers.go`'s evidence handler, plus the success/error/re-render paths.
- **`lib/api.test.ts`** — request construction for all three JSON endpoints, the case a naive implementation gets wrong (a non-JSON error body — e.g. a proxy-level 502 HTML page — falling back to a generic message instead of throwing a confusing parse error on top of the original failure), and `caseDocumentUrl`'s URL-encoding, plus a check that it's a pure URL builder that never itself calls `fetch` (see `lib/api.ts`'s doc comment on why the PDF endpoint is deliberately not wrapped the way the JSON endpoints are).
- **`app/case/[id]/page.test.tsx`** — see below.

### The integration test, and why it's not quite Playwright

This sandbox's `bash_tool` network allowlist (see the system configuration, or just try `npx playwright install` and watch it fail to reach a browser-binary CDN) doesn't extend to downloading a real browser, so a literal Playwright-driven end-to-end test wasn't achievable in this environment — stated plainly rather than claimed and quietly not delivered.

What was achievable, and is a genuine step up from component-level testing: `app/case/[id]/page.test.tsx` renders the **actual, real `CasePage` component** — not a reconstruction of it — with only the network boundary mocked (`@/lib/api`'s `getCase`, `next/navigation`'s `useParams`). This catches the specific class of bug isolated component tests structurally cannot: wiring. A prop passed under the wrong name, a conditional that hides a panel that should be showing, two components that individually render fine but were never actually composed correctly — none of that shows up when each component is tested alone. Concretely, this test:

- confirms the loading → loaded transition never shows both states at once,
- confirms the care-first message renders for **every one of the five outcomes** — the page-level version of `internal/response/builder_test.go`'s `TestBuild_CareFirstMessageIsAlwaysPresent_EveryOutcome`, checking the guarantee survives all the way to what a family's screen actually shows, not just what the backend's JSON contains,
- confirms `HandoffPanel` renders for a handoff outcome and *only* a handoff outcome,
- confirms action steps, the hospital script box, and the complaint-text box each render only when the backend actually sent that field, and
- confirms the error state renders the PMJAY helpline fallback correctly — and this last one caught a real ambiguity while being written: `Header` renders its own `tel:14555` link on every page, so an unscoped query for "the 14555 link" in the error-state test legitimately matched two elements. That's exactly the kind of thing a true integration test is supposed to surface.

Genuinely still missing relative to real Playwright: actual browser rendering (CSS layout, real click/focus/keyboard behavior, viewport-dependent behavior), and testing against a real running backend rather than a mocked API boundary. Worth revisiting if this environment's network constraints change, or from a developer's own machine where a browser binary can actually be installed.
