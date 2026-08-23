# PMJAY Point-of-Denial Advocate

A tool that helps a family holding a valid Ayushman Bharat (PMJAY) card figure out, in the moment, whether a hospital's coverage denial is correct — and what to do next, without ever telling them to delay treatment while they figure it out.

This is a working implementation of the concept in [`PMJAY_Startathon_Submission.md`](../PMJAY_Startathon_Submission.md) (the original spec this repo builds from). Section references throughout this codebase point back to that document.

## What this actually is

Not a PMJAY eligibility checker. Not a hospital-empanelment lookup. The specific, narrow thing this does — because it's the thing existing tools don't (see the spec's Section 8, "Why this is not just a verifier") — is take an unstructured, often distressed, often Malayalam/English code-mixed description of what a family was just told at a billing desk, and turn it into one of five honest outcomes:

- **Green** — this looks like a real, covered package; here's the citation and what to say.
- **Amber** — genuinely unclear (close-call match, or "pending" vs. "denied" isn't clear yet); here's the one question to ask before anything else.
- **Red** — correctly not covered; a straight, calm answer, not a manufactured grievance.
- **Mixed** — part of the bill is covered, part isn't; they need to be split, not treated as one yes/no.
- **Handoff** — genuinely too tangled or the family needs more than a guided flow; routed to a free NALSA Para Legal Volunteer (15100) with full context already prepared, so nothing has to be re-explained.

Every single one of those five outcomes leads with the same non-negotiable line: **get treatment first, dispute the money after, always.** See [`docs/SAFETY_DESIGN.md`](docs/SAFETY_DESIGN.md) for how that's enforced structurally, not just written down as a rule.

Every outcome also comes with something to actually take away, not just read on screen: a downloadable, printable PDF of the full case — care-first message, the outcome and why, action steps, hospital script, draft complaint, evidence log — a family can hand to hospital staff, attach to a CGRMS complaint, or bring to a NALSA Para Legal Volunteer. See [`backend/internal/document/README.md`](backend/internal/document/README.md).

## Before you read the code: the one thing to know

This repo uses a **seed/placeholder HBP dataset** — real, published PMJAY package *names* and *specialties* throughout, and for 300 of the 315 package records, real package *codes* and *rates* too, checked directly against government-published HBP rate schedules (HBP 2.1 via the National Health Authority's own master list, HBP 2022 and HBP 2.0 via two state health agency sources, and a 19 August reconciliation pass that resolved a version-currency question the previous session had deliberately left open). The other 15 are still placeholders, clearly flagged (`verified: false`) on every record that hasn't been independently checked against an actual government source. See [`docs/DATA_SOURCES.md`](docs/DATA_SOURCES.md) for exactly which is which, why, and for a specific, still-open lead on where the remaining seven specialties might actually be found. Reaching full coverage of the real ~1,900-procedure master list is still the single most important thing to do before this touches a real family; seven specialties (General Surgery, Orthopaedics, Obstetrics & Gynaecology, Ophthalmology, ENT, Urology, Neurosurgery) remain unreached after four separate extraction attempts hit the same kind of wall — real, separate, unglamorous work, not a rounding error.

The other thing worth knowing before anything else: the original spec's own Section 20.1 says the single largest open question — whether a family would actually set this up *before* a crisis, not during one — needs one real conversation with someone who's dealt with a PMJAY denial, before serious build time goes in. That conversation hadn't happened when this codebase was built. The engine here is worth having built either way (that question is about distribution, not about whether the matching/tiering logic works), but it's not a substitute for asking. A ready-to-use guide for that exact conversation is at [`docs/VALIDATION_INTERVIEW_GUIDE.md`](docs/VALIDATION_INTERVIEW_GUIDE.md); see [`docs/OPEN_QUESTIONS.md`](docs/OPEN_QUESTIONS.md) for the fuller picture.

## Quickstart

### Backend

```bash
cd backend
cp .env.example .env   # then set an API key for your chosen LLM_PROVIDER
go run ./cmd/server
```

Requires Go 1.22+. Zero third-party dependencies — see [`ARCHITECTURE.md`](ARCHITECTURE.md) for why, and how that keeps this cheap to run.

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Requires Node 20+. Runs on `http://localhost:3000`, expects the backend reachable at `/api` (see [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md) for the reverse-proxy setup, or set `NEXT_PUBLIC_API_BASE_URL` directly).

### Running the tests

```bash
cd backend
make test           # go test ./... -cover
make test-race       # same, with the race detector

cd ../frontend
npm test             # vitest run — 74 tests across 12 files
```

268 backend test cases as of this build, spanning every worked example and extended test-case scenario in the original spec, plus the H3 (tier-accuracy) and H4 (never-delay-care) adversarial safety sweeps described in Appendix R. See [`docs/TESTING.md`](docs/TESTING.md) for what's actually automated here versus what still needs a real field study — on both the backend and the frontend.

## Repository layout

```
backend/            Go 1.22, standard library only. See ARCHITECTURE.md, backend/README.md.
  cmd/server/        Entrypoint. See cmd/README.md, cmd/server/README.md.
  internal/hbp/       HBP reference dataset + loader. See internal/README.md, internal/hbp/README.md.
    data/              The actual JSON + the scripts that built it. See internal/hbp/data/README.md.
  internal/retrieval/ Cheap keyword pre-filter (zero API cost). See internal/retrieval/README.md.
  internal/extract/   The one LLM call — structured extraction & matching. See internal/extract/README.md.
  internal/tiering/   Deterministic decision logic (green/amber/red/mixed/handoff). See internal/tiering/README.md.
  internal/response/  Templated, safety-guaranteed response text. See internal/response/README.md.
  internal/store/     Case persistence (file-backed, zero external DB). See internal/store/README.md.
  internal/document/  Renders a case as a downloadable/printable PDF. See internal/document/README.md.
  internal/config/    Env-var configuration. See internal/config/README.md.
  internal/api/       HTTP layer. See internal/api/README.md.
frontend/           Next.js 16, App Router, TypeScript, Tailwind v4. See frontend/README.md.
  app/                See frontend/app/README.md.
    components/        Every React component. See frontend/app/components/README.md.
    case/[id]/          The result page. See frontend/app/case/[id]/README.md.
    fonts/              Self-hosted font files. See frontend/app/fonts/README.md.
  lib/                Types + API client. See frontend/lib/README.md.
docs/               Architecture, safety design, API reference, data provenance, deployment, testing, open questions, validation interview guide. See docs/README.md.
HANDOVER.md         Session-to-session handover notes — what the most recent session did, and what's genuinely still open, in priority order. Read this first if you're picking this project up cold.
```

**Every folder and subfolder above has its own `README.md`.** The layout diagram tells you what's where; the READMEs tell you why it's built the way it is. See the next section.

## How this codebase documents itself

Every folder and subfolder in this repository — not just the top level — has a `README.md` explaining three things: what the folder does, what's actually in it, and the specific architecture decisions and reasoning behind the code there. This isn't the same content as `ARCHITECTURE.md` or `docs/`: those cover system-wide and cross-cutting decisions, while a folder's own `README.md` covers what a reader needs to know *standing in that folder*, looking at *that code* — the kind of context that's expensive to reconstruct from git history or from asking around, and cheap to write down once, at the point when the reasoning is freshest (usually: whoever just wrote the code).

**This is a standing convention for this project, not a one-time cleanup.** If you are an AI agent or a human engineer continuing work on this codebase:

- **Any new folder you create gets a `README.md` in the same change that creates it**, not as a follow-up, not "later." A folder that exists for more than one commit without a `README.md` is an oversight to fix immediately, not a backlog item.
- **Any existing folder's `README.md` gets updated in the same change that meaningfully changes what that folder does** — a new file added, a design decision reversed, a dependency added or removed. A stale folder README is worse than a missing one: it actively misleads the next reader instead of just leaving a gap. (This project has direct experience with stale docs causing real confusion — see `docs/OPEN_QUESTIONS.md`'s own history of a doc that fell out of sync with reality across sessions; don't repeat that at the folder-README level.)
- **Write for the next reader, not for yourself right now.** State *why*, not just *what* — a comment that says "uses a token bucket" is worth much less than one that says "a token bucket, not a distributed rate limiter, because this system's actual scale doesn't need one and the cost this specific limiter exists to control is the LLM bill, not abuse in the abstract." Every folder README already written in this repository follows that standard; match it, don't undercut it.
- **If you're an AI agent picking up this project in a new session**: this instruction is not a suggestion to weigh against other priorities — it's part of how this codebase is meant to be worked on, the same way running the test suite before claiming something works is. Continue the convention without being asked again.

## License and attribution

Font files under `frontend/app/fonts/` are Atkinson Hyperlegible (Braille Institute of America) and Fraunces (The Fraunces Project Authors), both SIL Open Font License 1.1 — license text alongside the font files.
