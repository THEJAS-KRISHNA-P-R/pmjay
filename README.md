# PMJAY Denial Advocate

**Free, independent help for Indian families told at a hospital billing desk that a treatment is not covered under Ayushman Bharat (PM-JAY).**

Check whether the denial looks correct, find the Health Benefit Package (HBP) code and published rate behind the procedure, and get a plain-language next step — without ever being told to delay care.

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white)](backend/)
[![Next.js](https://img.shields.io/badge/Next.js-16-black?logo=nextdotjs)](frontend/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/tests-329%20Go%20%C2%B7%2077%20frontend-brightgreen)](docs/TESTING.md)

> **Get treatment first. Dispute the money after. Always.**
>
> That rule outranks every other instruction in this project. If you are deciding right now whether to accept care, call the **PM-JAY helpline on 14555** or the **national emergency number 112** first. A billing dispute can wait; a patient cannot.

---

## What this is

A point-of-denial tool, not an eligibility checker and not an empanelment lookup. It is built for one narrow moment: a family is at a counter, holding a valid PM-JAY card, and is told one of:

- a covered procedure is "not covered", and cash is demanded up front;
- "the PM-JAY server is down, pay now and claim later";
- the procedure name is not on the hospital's package list;
- pre-authorisation is "pending", and the family is told that means no;
- discharge papers are being held until the bill is paid.

It turns a distressed, often code-mixed description (English, Hindi, Malayalam, Tamil, and others) into one of five honest outcomes:

| Outcome | Meaning |
|---|---|
| **Green** | Looks like a real, covered package. Shows the package, its HBP code, the published rate where verified, and what to say. |
| **Amber** | Genuinely unclear — a close match, low confidence, or a pending (not denied) pre-authorisation. Asks the one useful question instead of guessing. |
| **Red** | The denial is correct. Says so calmly, explains the exclusion, and points to real alternatives. |
| **Mixed** | Part of the bill is a covered package, part is a real exclusion. Splits the bill instead of answering yes or no. |
| **Handoff** | Detention, coercion, or a situation too tangled for a guided flow. Routes to free legal aid on **15100** with the context already written down. |

Every outcome carries the care-first message, a desk script, an evidence checklist, and — where it applies — a CGRMS complaint draft and a printable PDF of the whole case.

## What makes it different

- **A rate is only shown if someone checked it.** 300 packages carry a code, specialty, and indicative rate verified against a published government HBP schedule. 15 are recognised by name with **every rate field, and the internal placeholder code, absent** — not zero, not a placeholder. The build and packaging steps fail if that boundary is crossed.
- **The tier decision is not made by a model.** One LLM call extracts a structured match; fixed rules in plain Go decide Green/Amber/Red/Mixed/Handoff. The same facts produce the same answer.
- **Search works without a model.** [`/packages`](frontend/app/packages/page.tsx) is a deterministic database lookup. It still works on a deployment with no provider key, which is exactly when someone most needs to look up a code.
- **Missing coverage is stated, not hidden.** The full HBP master list has roughly 1,900 procedures; this index covers 315 names. A missing result means "not in this index", never "not covered by PM-JAY".
- **No account, no Aadhaar number, no phone number.** Cases are reachable by an unguessable link, deleted after 30 days of inactivity, and deletable immediately from the case page. See [the privacy page](frontend/app/privacy/page.tsx).

## Honest limits

This is an independent public-interest project. It is **not** the National Health Authority, a hospital, an insurer, or a lawyer. It cannot:

- check a live pre-authorisation or empanelment status;
- submit a complaint on your behalf;
- change a PM-JAY record;
- guarantee that a rate is what your hospital will accept.

The dataset covers about one sixth of the full procedure list. General surgery, orthopaedics, obstetrics and gynaecology, ophthalmology, ENT, urology, and neurosurgery are the known thin areas. [docs/DATA_SOURCES.md](docs/DATA_SOURCES.md) records exactly what was checked, against which schedule, and when.

## Try it locally

### Backend — Go 1.22+, standard library only

```bash
cd backend
cp .env.example .env    # add ANTHROPIC_API_KEY, GROQ_API_KEY, or GEMINI_API_KEY
go run ./cmd/server
curl localhost:8080/api/v1/health
curl 'localhost:8080/api/v1/packages?q=dialysis'
```

The server starts with no provider key. Health and package search work; intake returns `503` with the helpline number until a key is configured.

### Frontend — Node 20+, Next.js 16

```bash
cd frontend
npm install
npm run dev             # http://localhost:3000
```

It expects the backend at `/api` (same-origin through a proxy), or set `NEXT_PUBLIC_API_BASE_URL`. For a public deployment, set `NEXT_PUBLIC_SITE_URL` too — it drives canonical URLs, the sitemap, robots.txt, and the generated `llms.txt` files. Full instructions: [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md).

### Test it

```bash
cd backend  && go test ./... -race -cover          # 329 tests, zero API key needed
cd ../frontend && npm test && npm run lint && npm run build   # 77 tests
```

## Agent skill

The verified dataset is also packaged as a standalone agent skill with **no frontend code** and **no network dependency** — just Python 3.9+ and the bundled data.

```bash
unzip dist/pmjay-denial-advocate-1.0.0.zip -d ~/.claude/skills/
python ~/.claude/skills/pmjay-denial-advocate/scripts/lookup.py "dialysis kidney failure"
```

The skill follows the same five-outcome protocol and the same care-first rule. See [skills/README.md](skills/README.md) for install options, rebuilding the data, and the validator that refuses to package a leaked rate or a frontend file. Marketplace copy and honest price guidance are in [skills/agensi-listing.md](skills/agensi-listing.md).

## Repository layout

```
backend/            Go 1.22, standard library only — no third-party modules.
  cmd/server/        Entrypoint, config → dataset → store → extractor → server.
  internal/hbp/      Embedded HBP dataset, schema, and validation.
  internal/retrieval/ Keyword pre-filter and package search (zero API cost).
  internal/extract/  The one LLM call, across Anthropic, Groq, and Gemini.
  internal/tiering/  Deterministic Green/Amber/Red/Mixed/Handoff rules.
  internal/response/ Templated, safety-guaranteed response text.
  internal/store/    File-backed case persistence, delete, retention purge.
  internal/document/ From-scratch PDF writer, zero dependencies.
  internal/config/   Environment-variable configuration.
  internal/api/      HTTP layer, middleware, rate limiting, redaction.
frontend/           Next.js 16 App Router, TypeScript, Tailwind v4.
  app/               Pages, package search, case workspace, generated SEO routes.
  lib/               Typed API client, site origin, local case history.
skills/             Standalone agent skill, data build, packaging, listing copy.
docs/               API, deployment, data provenance, safety design, testing.
ARCHITECTURE.md     System shape and the decisions behind it.
HANDOVER.md         Session history and what is genuinely still open.
```

Every folder has its own `README.md` explaining *why* the code is shaped the way it is, not just what is in it. Start with [ARCHITECTURE.md](ARCHITECTURE.md) and [docs/README.md](docs/README.md).

## Machine-readable

- [`/llms.txt`](frontend/app/llms.txt/route.ts) — a short, factual orientation for agents and models.
- [`/llms-full.txt`](frontend/app/llms-full.txt/route.ts) — the longer system context, including what the tool cannot do.
- [`/sitemap.xml`](frontend/app/sitemap.ts), [`/robots.txt`](frontend/app/robots.ts) — generated from the deploy origin.
- [`docs/API.md`](docs/API.md) — all seven endpoints, with real captured request/response examples.

## Contributing and security

Corrections to a package code, specialty, or rate are genuinely valuable, and belong in the dataset — not in a private email or a hand-edited JSON file. Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

Please do **not** put a case link, a patient's name, or medical details in a public issue. See [SECURITY.md](SECURITY.md) for reporting a vulnerability privately.

## License

[MIT](LICENSE). The dataset is an independent compilation of publicly available National Health Authority material and is not affiliated with or endorsed by the NHA, the Ministry of Health and Family Welfare, any State Health Agency, or any hospital.

Fonts under `frontend/app/fonts/` are Atkinson Hyperlegible (Braille Institute of America) and Fraunces (The Fraunces Project Authors), both under the SIL Open Font License 1.1; license text ships alongside the files.
