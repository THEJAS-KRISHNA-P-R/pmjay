# API Reference

Base path: `/api/v1`. All request and response bodies are JSON. All examples on this page are **real, actually-generated output** — captured by running the live handlers against `extract.FakeClient` (the same deterministic test double the automated tests use), not hand-typed. DTOs are defined in `backend/internal/api/dto.go`; this document mirrors them.

Five routes exist. Only one costs money to call.

| Method | Path | Costs an LLM call? | Rate-limited? |
|---|---|---|---|
| `POST` | `/api/v1/cases` | Yes — the one paid call per family interaction | Yes |
| `GET` | `/api/v1/cases/{id}` | No | No |
| `GET` | `/api/v1/cases/{id}/document` | No | No |
| `POST` | `/api/v1/cases/{id}/evidence` | No | No |
| `GET` | `/api/v1/health` | No (no store access either) | No |

Middleware order, applied to every route: panic recovery → structured logging → CORS → rate limiting (intake only). See `internal/api/router.go`.

---

## `POST /api/v1/cases`

Runs Section 7's Steps 1–4 in one request: retrieval pre-filter, the one LLM extraction call, deterministic tiering, and templated response construction. This is the request that triggers a paid call to whichever provider `LLM_PROVIDER` selects (see `docs/DEPLOYMENT.md`), which is why it's the only rate-limited endpoint (10 requests/minute per client IP by default — see `RATE_LIMIT_PER_MINUTE` in `docs/DEPLOYMENT.md`).

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

**Error responses**

| Status | Cause | Body |
|---|---|---|
| `400` | Malformed JSON | `{"error": "invalid JSON body"}` |
| `400` | Description under 5 characters | `{"error": "description is too short to be meaningful"}` |
| `400` | Description over 4000 characters | `{"error": "description is too long"}` |
| `429` | Rate limit exceeded for this IP | `{"error": "too many requests, please wait a moment and try again", "fallback_guidance": "If this is urgent, call the PMJAY helpline directly at 14555."}` |
| `502` | The LLM extraction call itself failed (network, auth, timeout — whichever provider `LLM_PROVIDER` selects) | `{"error": "the system could not process this request right now", "fallback_guidance": "<care-first text> In the meantime, call the PMJAY helpline directly at 14555 — they can help without needing this tool to be working."}` |

The `502` case is deliberate, not an oversight: Section 10's care-first rule is written as an absolute, so an infrastructure failure is not treated as an exception to it. `fallback_guidance` carries the same care-first text plus a human fallback that doesn't depend on this system working at all — see `internal/api/handlers.go`'s `handleIntake` and `lib/api.ts`'s `ApiError` on the frontend side, which exists specifically to surface this field rather than discard it on a thrown error.

A storage failure after a successful extraction does **not** fail the request — the computed response is still returned to the family; only the ability to re-fetch or add evidence to that case later is affected. This is logged server-side, not surfaced as an error to the caller (see the comment on that branch in `handleIntake`).

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

## `GET /api/v1/health`

A dependency-free liveness check — no LLM call, no store access — suitable for a load balancer, uptime monitor, or Docker `HEALTHCHECK` to poll frequently at zero cost.

**Response — `200 OK`** (real captured output against the seed dataset):

```json
{
  "exclusions_loaded": "4",
  "packages_loaded": "315",
  "status": "ok"
}
```

`packages_loaded` / `exclusions_loaded` reflect whatever dataset is embedded at build time — 315 packages (300 verified against sourced government rate schedules, the remainder still placeholder) and 4 exclusion categories today (see `docs/DATA_SOURCES.md`); these numbers will grow further as HBP data coverage continues (see `docs/OPEN_QUESTIONS.md`). This endpoint never returns a non-200 status in the current implementation — a genuinely dead process won't respond at all, which is itself the signal a load balancer needs.

---

## CORS

`internal/api/middleware.go`'s `corsMiddleware` allows only origins in the `ALLOWED_ORIGINS` config list (see `docs/DEPLOYMENT.md` and `.env.example`) — deliberately no wildcard, since the intake endpoint triggers a paid API call per request and an open CORS policy would let any third-party site spend this system's API budget. Preflight `OPTIONS` requests get a bare `204 No Content`.

## A note on what these examples are, and aren't

Every JSON block on this page came from an actual `httptest` round-trip against the real router and handlers, using `extract.FakeClient` in place of a live LLM call (the same fake the automated test suite uses, and the same one regardless of which real provider — Claude, Groq, or Gemini — `LLM_PROVIDER` would otherwise select; see `docs/TESTING.md`). The tier logic, response templates, and JSON shape are exactly what production traffic produces. The one thing that differs from a real request is the extraction step itself: a live call would involve genuine model uncertainty on the input description, where this example was registered against a fixed, known-good extraction result so the output is reproducible for documentation purposes.
