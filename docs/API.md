# API Reference

Base path: `/api/v1`. All request and response bodies are JSON. All examples on this page are **real, actually-generated output** — captured by running the live handlers against `extract.FakeClient` (the same deterministic test double the automated tests use), not hand-typed. DTOs are defined in `backend/internal/api/dto.go`; this document mirrors them.

Seven routes exist. Only one costs money to call.

| Method | Path | Costs an LLM call? | Rate-limited? |
|---|---|---|---|
| `POST` | `/api/v1/cases` | Yes — the one paid call per family interaction | Yes (intake limits) |
| `GET` | `/api/v1/cases/{id}` | No | Yes (general limits) |
| `DELETE` | `/api/v1/cases/{id}` | No | Yes (general limits) |
| `GET` | `/api/v1/cases/{id}/document` | No | Yes (general limits) |
| `POST` | `/api/v1/cases/{id}/evidence` | No | Yes (general limits) |
| `GET` | `/api/v1/packages` | No — plain dataset search | Yes (general limits) |
| `GET` | `/api/v1/health` | No (no store access either) | No |

Middleware order, applied to every route: panic recovery → structured logging → security headers → `Cache-Control: no-store` → CORS → rate limiting (intake only) → timeout and concurrency limiting (intake only). See `internal/api/router.go`.

Two separate rate limiters exist. The **intake** limiter guards the one endpoint that spends money on a model call. The **general** limiter is a deliberately more generous flat allowance shared by every other store-touching endpoint; it exists to bound abuse, not to ration a family's use of their own case. `/api/v1/health` is outside both.

---

## `POST /api/v1/cases`

Runs Section 7's Steps 1–4 in one request: retrieval pre-filter, the one LLM extraction call, deterministic tiering, and templated response construction. This is the request that triggers a paid call to whichever provider `LLM_PROVIDER` selects (see `docs/DEPLOYMENT.md`), which is why it's the only endpoint on the intake limiter (1 request/minute and 20/hour per client IP by default — see `RATE_LIMIT_PER_MINUTE` and `RATE_LIMIT_PER_HOUR` in `backend/.env.example`). The whole request is bounded by `INTAKE_TIMEOUT_SECONDS` (45 s default), which is shorter than the server's write timeout so a slow provider produces the care-first error below rather than a dropped connection.

**Request**

| Field | Type | Rule |
|---|---|---|
| `description` | string | 5–4000 characters. No language requirement — Malayalam, English, or code-mixed input all reach the same extraction step. |

```json
{
  "description": "My mother needs her gallbladder removed, doctor confirmed stones on scan, hospital billing desk just told us PMJAY won't cover it and we need to pay before they'll schedule the surgery."
}
```

**Response — `201 Created`**

```json
{
  "id": "80e5d440-ec7b-4be1-8886-0b752413a9e9",
  "outcome": "green",
  "citation": "Laparoscopic Cholecystectomy",
  "description": "Gallbladder removal for confirmed stones; hospital demanding upfront payment.",
  "persisted": true,
  "care_first_message": "Get treatment first. Dispute the money after. Always. If you can pay now and settle the dispute later, or move to a different hospital, do that — do not let this disagreement delay or stop care.",
  "tier_message": "Based on what you've described — gallbladder removal for confirmed stones; hospital demanding upfront payment. — this matches a package that is listed as covered under PMJAY: Laparoscopic Cholecystectomy (General Surgery), a listed PMJAY package. This hospital should not be asking for payment for this procedure if they are empanelled for General Surgery.",
  "action_steps": [
    "Ask the billing desk, calmly: \"Can you please give this denial to us in writing, with the reason stated?\" Hospitals are required to be able to justify a denial, and a verbal refusal that won't be put in writing is itself worth noting.",
    "If they insist on payment before proceeding and treatment is needed soon, you can pay and dispute the charge afterward — do not let this delay care if it's urgent.",
    "Note down which staff member told you this, and the approximate time."
  ],
  "complaint_text": "--- Draft complaint for CGRMS (submit via the Ayushman App) ---\nDate prepared: 14 August 2026\nSubject: Denial of covered service: Laparoscopic Cholecystectomy\n\nI am writing to report that I was denied a covered service, or asked for payment for a covered service, under my PMJAY card.\n\nSituation described: Gallbladder removal for confirmed stones; hospital demanding upfront payment.\n\nMatched PMJAY package: Laparoscopic Cholecystectomy (General Surgery)\n\nI am requesting review of this denial and, if confirmed incorrect, appropriate action against the hospital involved.\n\n[Add: your PMJAY card number, the hospital name, the staff member's name if known, and the approximate time of the incident, from the evidence you noted below, before submitting.]\n--- End of draft — review before submitting ---",
  "hospital_script": "Ask the billing desk, calmly: \"Can you please give this denial to us in writing, with the reason stated?\" Hospitals are required to be able to justify a denial, and a verbal refusal that won't be put in writing is itself worth noting.",
  "evidence_prompt": "Before you leave this conversation, note down three things while they're still easy to get: (1) the name of the staff member you spoke to, (2) the approximate time, and (3) get the denial in writing if you can — a photo of a written note is enough. This is easy to lose track of later, especially before a shift change."
}
```

`care_first_message` is present on **every** outcome, unconditionally — see `docs/SAFETY_DESIGN.md` §1. `citation`, `action_steps`, `complaint_text`, and `hospital_script` are omitted (via `omitempty`) on outcomes where they don't apply — Amber never produces `complaint_text` (nothing to complain about yet), Red and pure-Handoff never produce either (see `docs/SAFETY_DESIGN.md` and `internal/response/builder_test.go`'s `TestBuild_Amber_NeverProducesAComplaint` / `TestBuild_Red_NeverProducesAComplaintOrActionSteps`). `outcome` is one of `green`, `amber`, `red`, `mixed`, `handoff`.

`persisted` is `true` unless the answer was computed but could not be written to the case store. In that case the guidance is still returned and still valid, but the frontend must not offer follow-up that needs a server record (evidence, deletion, reopening the link). See "A storage failure" below.

**Error responses**

| Status | Cause | Body |
|---|---|---|
| `400` | Malformed JSON | `{"error": "invalid JSON body"}` |
| `400` | Description under 5 characters | `{"error": "description is too short to be meaningful"}` |
| `400` | Description over 4000 characters | `{"error": "description is too long"}` |
| `413` | Request body over the size limit | `{"error": "request body too large"}` |
| `429` | Intake rate limit exceeded for this IP | `{"error": "too many requests, please wait a moment and try again", "fallback_guidance": "If this is urgent, call the PMJAY helpline directly at 14555."}` |
| `502` | The LLM extraction call failed (network, auth, provider error) | `{"error": "the system could not process this request right now", "fallback_guidance": "<care-first text> In the meantime, call the PMJAY helpline directly at 14555 — they can help without needing this tool to be working."}` |
| `503` | This deployment has no provider key configured, so intake cannot run at all | `{"error": "the guided case analysis is not available on this server right now", "fallback_guidance": "<care-first text> Call the PMJAY helpline directly at 14555 — they can help without needing this tool to be working."}` |
| `504` | The provider took longer than `INTAKE_TIMEOUT_SECONDS` | `{"error": "the analysis took too long and was stopped", "fallback_guidance": "<care-first text> Please try once more, or call the PMJAY helpline directly at 14555 — they can help without needing this tool to be working."}` |

All three upstream failures are deliberate, not oversights: Section 10's care-first rule is written as an absolute, so an infrastructure failure is not treated as an exception to it. `fallback_guidance` carries the same care-first text plus a human fallback that doesn't depend on this system working at all — see `internal/api/handlers.go`'s `handleIntake` and `lib/api.ts`'s `ApiError` on the frontend side, which exists specifically to surface this field rather than discard it on a thrown error.

The distinction matters to the person retrying: `503` is a deployment missing a key, so retrying now is pointless and `14555` is the real next step. `504` means the provider was slow, so one retry is reasonable. `502` means the provider failed, so retry, then use the helpline. `GET /api/v1/health` reports `extraction_ready: false` for the `503` case without failing the liveness check.

A storage failure after a successful extraction does **not** fail the request — the computed response is still returned to the family with `persisted: false`; only the ability to re-fetch, add evidence to, or delete that case later is affected. This is logged server-side (with the case ID redacted) rather than surfaced as an error that would discard useful guidance (see the comment on that branch in `handleIntake`).

Real captured example, `400` (description `"hi"`, 2 characters):

```json
{
  "error": "description is too short to be meaningful"
}
```

---

## `GET /api/v1/cases/{id}`

Fetches a previously created case. Same response shape as the `POST` above.

**Response — `200 OK`** (real captured output, same case as above, before any evidence was added):

```json
{
  "id": "80e5d440-ec7b-4be1-8886-0b752413a9e9",
  "outcome": "green",
  "citation": "Laparoscopic Cholecystectomy",
  "care_first_message": "Get treatment first. Dispute the money after. Always. If you can pay now and settle the dispute later, or move to a different hospital, do that — do not let this disagreement delay or stop care.",
  "tier_message": "Based on what you've described — gallbladder removal for confirmed stones; hospital demanding upfront payment. — this matches a package that is listed as covered under PMJAY: Laparoscopic Cholecystectomy (General Surgery), a listed PMJAY package. This hospital should not be asking for payment for this procedure if they are empanelled for General Surgery.",
  "action_steps": ["...same three steps as above..."],
  "complaint_text": "...same draft as above...",
  "hospital_script": "Ask the billing desk, calmly: \"Can you please give this denial to us in writing, with the reason stated?\" Hospitals are required to be able to justify a denial, and a verbal refusal that won't be put in writing is itself worth noting.",
  "evidence_prompt": "Before you leave this conversation, note down three things while they're still easy to get: (1) the name of the staff member you spoke to, (2) the approximate time, and (3) get the denial in writing if you can — a photo of a written note is enough. This is easy to lose track of later, especially before a shift change."
}
```

**Error responses**

| Status | Cause | Body (real captured output) |
|---|---|---|
| `404` | No case with this ID | `{"error": "case not found"}` |
| `500` | Store read failed | `{"error": "could not retrieve case"}` |

---

## `DELETE /api/v1/cases/{id}`

Permanently removes a case and its evidence from the store. This is the backend half of the DPDP erasure right described on `/privacy`: there is no soft delete, no archive, and no copy retained afterwards.

**Response — `204 No Content`**

No body. A repeat call (or any unknown ID) returns `404` rather than `204`, so a client can distinguish "deleted now" from "was never there". The frontend's `deleteCase` treats both as success, because the family's goal — that the record is gone — is met either way.

**Error responses**

| Status | Cause | Body |
|---|---|---|
| `404` | No case with this ID | `{"error": "case not found"}` |
| `500` | Store write failed | `{"error": "could not delete case"}` |

Cases are also deleted automatically after `CASE_RETENTION_DAYS` (30 by default) of inactivity, by an hourly purge at startup and on a timer.

---

## `GET /api/v1/cases/{id}/document`

The same case `GET /api/v1/cases/{id}` returns as JSON, rendered instead as a single PDF via `internal/document` — the care-first message, the outcome explanation, action steps, hospital script, draft complaint, and evidence log, formatted as one document a family can download, print, or hand directly to hospital staff. See `internal/document/README.md` for how it's built (a from-scratch PDF writer, zero third-party dependencies, matching the rest of this backend).

**Response — `200 OK`** (real captured headers, same case as the examples above):

```
Content-Type: application/pdf
Content-Disposition: inline; filename="pmjay-case-57f66755-56fd-4c73-b60d-c341ef08f97b.pdf"
Content-Length: 6885
```

Body is the raw PDF bytes (starting `%PDF-1.4`), not JSON — this is the one endpoint on this page that isn't.

`Content-Disposition` is deliberately `inline`, not `attachment`: this opens the PDF directly in the browser's own viewer (desktop and mobile both), which itself provides print/save/share controls — see `handleGetCaseDocument`'s doc comment in `internal/api/handlers.go` for the full reasoning, including why that matters more on the kind of device this project's users are likely to actually be on.

**Error responses** — identical to `GET /api/v1/cases/{id}` above (`404` if the case doesn't exist, `500` if the store read fails), plus:

| Status | Cause | Body |
|---|---|---|
| `500` | Document generation failed | `{"error": "could not generate document"}` |

That last case isn't currently reachable in practice — `document.BuildCase` has no failing code path today (see its doc comment) — but the handler still checks and returns a clear `500` rather than ever risking a truncated or corrupt PDF reaching a family mid-crisis.

---

## `POST /api/v1/cases/{id}/evidence`

Section 7 Step 6 — append one evidence entry (staff name, approximate time, and/or a free-text note) to an existing case, timestamped server-side. At least one of the three fields must be non-empty.

**Request** (real captured example):

```json
{
  "staff_name": "Billing desk clerk (name not given)",
  "approx_time": "around 4pm",
  "note": "Said PMJAY card would not be accepted for this procedure"
}
```

All three fields are individually optional — `{"note": "..."}` alone is valid — but the request is rejected if all three are empty.

**Response — `200 OK`**: the full case again, in the same shape as above, now with an `evidence` array appended (real captured output):

```json
{
  "...": "...(same case fields as above)...",
  "evidence": [
    {
      "captured_at": "2026-08-14T00:49:16Z",
      "staff_name": "Billing desk clerk (name not given)",
      "approx_time": "around 4pm",
      "note": "Said PMJAY card would not be accepted for this procedure"
    }
  ]
}
```

`captured_at` is set server-side (RFC 3339, UTC) at the moment the request is received — the client cannot backdate it. Repeat calls append further entries; nothing is overwritten.

**Error responses**

| Status | Cause | Body |
|---|---|---|
| `400` | Malformed JSON | `{"error": "invalid JSON body"}` |
| `400` | All three fields empty | `{"error": "at least one of staff_name, approx_time, or note is required"}` |
| `404` | No case with this ID | `{"error": "case not found"}` |
| `500` | Store write failed | `{"error": "could not save evidence"}` |

---

## `GET /api/v1/packages`

Deterministic search over the embedded Health Benefit Package index. **No model call** — this is the endpoint the `/packages` page uses, and it keeps working on a deployment with no provider key configured.

**Query parameters** (all optional)

| Name | Type | Rule |
|---|---|---|
| `q` | string | Free text matched against package name, code, specialty, and keywords. Empty returns the whole index. |
| `specialty` | string | Case-insensitive exact specialty filter (e.g. `Cardiology`). |
| `limit` | int | 1–100, default 25. |
| `offset` | int | 0–10000, default 0. |

**Response — `200 OK`**

```json
{
  "query": "dialysis",
  "specialties": ["Cardiology", "General Surgery", "Nephrology", "..."],
  "total": 3,
  "count": 3,
  "offset": 0,
  "results": [
    {
      "package_code": "SR-001",
      "package_name": "Haemodialysis",
      "specialty": "Nephrology",
      "indicative_rate_inr": 1500,
      "rate_type": "per_day",
      "per_diem_rates": {
        "routine_ward_inr": 1500,
        "hdu_inr": 2500,
        "icu_no_vent_inr": 4000,
        "icu_vent_inr": 5000
      },
      "requires_preauth": true,
      "verified": true
    }
  ]
}
```

`specialties` is always the full sorted list from the dataset, regardless of the current filter, so the client can render a stable filter menu. `total` is the number of matches before pagination; `count` is the number returned in this page.

**Rate stripping is structural.** `packageToSearchResult` is the only path from an `hbp.Package` to this API, and it returns early for a record with `verified: false`. An unverified record therefore carries its name, specialty, and pre-auth flag but **no** `package_code`, `indicative_rate_inr`, `rate_max_inr`, `rate_type`, or `per_diem_rates` — not zero, not a placeholder, absent. A packaging test and a handler test both assert this, because a leaked placeholder would be read as a citation.

`package_code` is withheld for the same reason a rate is: the dataset's internal `SEED-` identifiers are development placeholders, not codes that appear in any government schedule. Searching `hernia` returns `Inguinal Hernia Repair (Unilateral)` with `"verified": false` and no `package_code` key at all, rather than an identifier someone might quote at a billing desk.

**Error responses**

| Status | Cause | Body |
|---|---|---|
| `400` | `limit` outside 1–100 or not an integer | `{"error": "limit must be an integer between 1 and 100"}` |
| `400` | `offset` outside 0–10000 or not an integer | `{"error": "offset must be an integer between 0 and 10000"}` |

---

## `GET /api/v1/health`

A dependency-free liveness check — no LLM call, no store access — suitable for a load balancer, uptime monitor, or Docker `HEALTHCHECK` to poll frequently at zero cost.

**Response — `200 OK`** (real captured output against the seed dataset):

```json
{
  "status": "ok",
  "packages_loaded": 315,
  "exclusions_loaded": 4,
  "extraction_ready": true
}
```

`packages_loaded` / `exclusions_loaded` reflect whatever dataset is embedded at build time — 315 packages (300 verified against sourced government rate schedules, 15 recognised by name only) and 4 exclusion categories today (see `docs/DATA_SOURCES.md`); these numbers will grow as HBP data coverage continues (see `docs/OPEN_QUESTIONS.md`).

`extraction_ready` is deliberately separate from `status`. "The process is up" and "intake can actually work" are different facts, and conflating them is how a deployment with a missing API key passes its health check while every real request returns `503`. `status` stays `ok` so a liveness probe does not restart a healthy process over a configuration problem a restart cannot fix. Readiness is determined by the extractor implementing `extract.Availability`; a fake used in tests can report either value.

This endpoint never returns a non-200 status in the current implementation — a genuinely dead process won't respond at all, which is itself the signal a load balancer needs.

---

## CORS

`internal/api/middleware.go`'s `corsMiddleware` allows only origins in the `ALLOWED_ORIGINS` config list (see `docs/DEPLOYMENT.md` and `.env.example`) — deliberately no wildcard, since the intake endpoint triggers a paid API call per request and an open CORS policy would let any third-party site spend this system's API budget. Preflight `OPTIONS` requests get a bare `204 No Content`.

## A note on what these examples are, and aren't

Every JSON block on this page came from an actual `httptest` round-trip against the real router and handlers, using `extract.FakeClient` in place of a live LLM call (the same fake the automated test suite uses, and the same one regardless of which real provider — Claude, Groq, or Gemini — `LLM_PROVIDER` would otherwise select; see `docs/TESTING.md`). The tier logic, response templates, and JSON shape are exactly what production traffic produces. The one thing that differs from a real request is the extraction step itself: a live call would involve genuine model uncertainty on the input description, where this example was registered against a fixed, known-good extraction result so the output is reproducible for documentation purposes.
