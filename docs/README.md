# `docs`

Deep-dive documentation that didn't belong in a folder-level `README.md` because it's cross-cutting — about the *product* and its *safety/data guarantees*, not about any one folder's code. If a folder-level README references "see `docs/X.md`," this is where it's pointing.

Two related files live at the repository root instead of here, deliberately: `../README.md` (the entry point — what this is, quickstart, top-level status) and `../ARCHITECTURE.md` (the system-shape diagram and the handful of build-environment-driven engineering decisions that shaped everything downstream — zero Go dependencies, no database, self-hosted fonts). Read those two first if you haven't; everything in this folder assumes that context.

## The eight documents here

| Document | What it's the answer to |
|---|---|
| `SAFETY_DESIGN.md` | How the care-first guarantee and the tiering logic's conservative bias are actually enforced in code, not just stated as intent. Read this before touching `backend/internal/response` or `backend/internal/tiering`. |
| `DATA_SOURCES.md` | Which package records are real and verified against a government source, which are still placeholder, and exactly how each was checked. **Read this before trusting any specific rupee figure this tool would cite**, and before removing the `verified` field from anywhere in `backend/internal/hbp`. Also the canonical history of every data-extraction session, including the walls hit and the rate-currency question found but not yet resolved. |
| `API.md` | Full request/response shapes and status codes for the four real endpoints. `backend/internal/api/README.md` covers the *why*; this covers the *wire format*. |
| `TESTING.md` | What's actually automated (268 backend tests, 74 frontend tests, exact coverage per package) versus what's a field-study-level question no amount of unit testing can answer. Read this before claiming a change is "tested" — and before assuming a passing suite means more than it actually does. |
| `DEPLOYMENT.md` | Both real deployment paths (colocated on existing infrastructure, or fully standalone), with concrete steps — Fly.io/Oracle Cloud for the backend, Vercel for the frontend, and the same-origin-via-Caddy setup that makes `NEXT_PUBLIC_API_BASE_URL`'s default just work. |
| `OPEN_QUESTIONS.md` | The single most important document if you're picking this project up cold: an honest, prioritized list of what's genuinely unfinished, written the same way each build session left it for the next one. Read this and `../HANDOVER.md` together before starting new work. |
| `VALIDATION_INTERVIEW_GUIDE.md` | A structured guide for the one piece of validation no code change can substitute for — an actual conversation with a family or ASHA worker about whether they'd use this before a crisis, not during one. Still the highest-priority open item on the whole project as of the most recent `../HANDOVER.md`. |
| `MARKETING_AND_GO_VERIFICATION.md` | The two concrete pieces of work the 26 August 2026 restructure (see `../HANDOVER.md`'s addendum of that date) started but didn't finish: compiling and live-testing that session's two backend changes with no Go toolchain available to do it there, and closing the gap between the marketing site and the product brief it was redesigned against — what's a real gap (technical SEO, a missing Languages page) versus what's already in decent shape and just needs a targeted audit. |

## Why documentation is split this way — folder READMEs vs. this folder

A folder-level `README.md` (see `../backend/README.md`, `../frontend/README.md`, and every subfolder beneath them) answers "what does this code do and why is it built this way" — scoped to that folder, written to be read alongside the code it describes. A document in this folder answers a question that cuts across many folders at once, or that isn't really about code at all (data provenance, a validation methodology, a deployment runbook). If you're about to add a new `.md` file and aren't sure which category it's in: does it make sense read standalone, away from any specific folder of code? If yes, it belongs here. If it's really explaining *this specific folder's* code and decisions, it belongs in that folder instead — see the repository root `README.md`'s "How this codebase documents itself" section for the full convention.
