# `frontend/lib`

The frontend's entire API boundary and type layer — two files, both deliberately hand-written rather than generated.

## Files

| File | What it does |
|---|---|
| `types.ts` | TypeScript types mirroring `backend/internal/api/dto.go` by hand. |
| `api.ts` | Three fetch wrappers (`createCase`, `getCase`, `addEvidence`), one plain URL builder (`caseDocumentUrl`), and `ApiError`. |

## Why these types are hand-written, not code-generated from the Go structs

Stated directly in `types.ts`'s own header comment: this API surface is small (four request/response shapes, one enum) and stable enough that a codegen step (an OpenAPI spec, a Go→TS type generator, a build-time sync script) would be more moving parts than it's worth — another dependency, another build step, another thing that can silently drift out of sync in a different way (a codegen tool with a stale cache is just as capable of lying as a hand-maintained file, just less visibly). This matches the whole project's general bias, on both frontend and backend, toward fewer moving parts over more automation for a surface this size — see `../../ARCHITECTURE.md`.

**The tradeoff this creates, honestly**: `types.ts` can drift from `dto.go` if one changes without the other. There's no compiler or CI check that would catch a field renamed in Go but not renamed here — it would surface as a runtime `undefined` in the UI, not a build failure. `IntakeForm.tsx`'s duplicated 5-character minimum-length check (mirroring `backend/internal/api/handlers.go`'s `minDescriptionLength`) is the same category of risk, acknowledged the same way in that file's own comments. If this project ever grows a CI pipeline, a script that fails the build when `dto.go` and `types.ts` disagree (even a crude one — diff the field name lists) would close this gap cheaply; it doesn't exist today.

## `Outcome`

```ts
export type Outcome = "green" | "amber" | "red" | "mixed" | "handoff";
```

Must stay in exact sync with `backend/internal/tiering.Outcome`'s five constants. If a sixth outcome is ever added on the backend (see `backend/internal/tiering/README.md`), this union needs the matching string literal, and `TierBadge.tsx`'s `TIER_STYLES` map (see `frontend/app/components/README.md`) needs a matching entry, or a new outcome will compile fine on both ends and then render as a blank or default state on real traffic — this is the single most likely "compiles but is wrong" bug this codebase's frontend/backend split can produce, and there's no automated check for it today (see above).

## `api.ts`: `ApiError` exists specifically to carry `fallback_guidance` through

This is the one piece of frontend plumbing most directly downstream of the backend's care-first guarantee. When `internal/extract`'s LLM call fails, `handleIntake` (see `backend/internal/api/README.md`) returns a 502 whose body includes `fallback_guidance` — the care-first text plus the PMJAY helpline number — specifically so a family isn't left with nothing just because of an infrastructure failure. A plain `throw new Error(body.error)` here would silently discard that field. `ApiError` is a small subclass carrying `fallbackGuidance` and `status` alongside the message, specifically so `IntakeForm.tsx`'s catch block can render it (see that component's own error-display JSX, which conditionally shows `fallbackGuidance` right under the main error message).

`handleResponse`'s `try`/`catch` around `res.json()` on the error path exists for a specific, real failure mode: if something between the browser and the actual Go backend returns a non-JSON error body (a reverse proxy's own 502 HTML page, for instance, rather than the backend's own JSON `ErrorResponse`), naively parsing it as JSON would throw a confusing "unexpected token `<`" parse error that buries the original, more useful failure. Falling back to a generic message instead keeps the error a family (or a developer looking at logs) actually sees meaningful.

## `caseDocumentUrl`: not a fetch wrapper, deliberately

The odd one out in this file. Every other exported function here calls `fetch` and returns parsed JSON; `caseDocumentUrl(id)` just builds and returns the URL string for `backend/internal/api`'s `GET /api/v1/cases/{id}/document` (`frontend/app/components/CaseDocumentPanel.tsx`'s "Download PDF" link uses it directly as an `<a href>`). It's deliberately *not* a fetch wrapper: the caller wants the browser to navigate to that URL itself, so the browser's native PDF viewer handles the response, not a `Blob` this code would have to receive, hold in memory, and turn into a downloadable link by hand. That also means a script-initiated `fetch`/`XHR` to a different-origin `API_BASE_URL` would need CORS configured for it; a plain browser navigation (what `caseDocumentUrl` is for) never does, since CORS only governs script-initiated requests — see the function's own doc comment.

## `API_BASE_URL`: the one environment variable this frontend reads

`process.env.NEXT_PUBLIC_API_BASE_URL ?? "/api"` — defaults to same-origin, so local development behind a reverse proxy (or a production same-origin Caddy setup, see `docs/DEPLOYMENT.md`) works with zero configuration. There's no `frontend/.env.example` for this — intentionally; see `frontend/README.md`'s note on why, and `docs/DEPLOYMENT.md` for exactly how this gets set (`vercel env add`, not a checked-in file) when the frontend and backend are deployed to different origins.

## If you're extending this folder

- **Adding a new backend field**: update `types.ts` first (it's the contract everything else in the frontend codes against), then wire it through wherever it needs to render — see `frontend/app/case/[id]/README.md`.
- **Adding a new JSON endpoint**: follow the existing three functions' pattern exactly — a thin `fetch` wrapper, `handleResponse<T>` for the error-handling boilerplate, a specific return type from `types.ts`.
- **Adding a new endpoint whose response isn't JSON the frontend needs to parse** (a file download, a redirect, anything meant for direct browser navigation): follow `caseDocumentUrl`'s pattern instead — a plain URL builder, not a `fetch` wrapper. See its own section above for why that distinction is deliberate, not an inconsistency to "fix" by wrapping it in `fetch` to match the other three.
- **This file is part of the codebase's documentation convention** — see the repo root `README.md`. Keep it in sync with the code, in the same change.
