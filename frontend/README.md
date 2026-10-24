# `frontend`

Next.js 16 (App Router) + TypeScript + Tailwind v4. Talks to the Go backend over a small, typed API surface. Deployed as a static-friendly app — see `../docs/DEPLOYMENT.md` for the Vercel path this was built against.

## Layout

```
frontend/
├── app/
│   ├── layout.tsx, page.tsx, global-error.tsx, globals.css   — see app/README.md
│   ├── components/       every React component — see app/components/README.md
│   ├── packages/         no-LLM package search page (server page + client explorer)
│   ├── case/[id]/         the result page — see app/case/[id]/README.md
│   ├── robots.ts, sitemap.ts        generated from NEXT_PUBLIC_SITE_URL
│   ├── llms.txt/, llms-full.txt/    generated agent-orientation text
│   └── fonts/              self-hosted font files — see app/fonts/README.md
├── lib/
│   ├── types.ts             hand-mirrors backend/internal/api/dto.go
│   ├── api.ts                fetch wrappers + ApiError — see lib/README.md
│   └── site.ts               the one place the public origin is decided
├── vitest.config.mts, vitest.setup.ts    test runner config
├── eslint.config.mjs
├── tsconfig.json
└── package.json
```

Every subfolder has its own `README.md` with the real depth. This file is the map.

## Quickstart

```bash
npm ci
npm run dev      # http://localhost:3000, expects the backend at /api (see below)
npm test          # vitest run — 77 tests across 12 files, as of this writing
npm run lint
npm run build
```

No backend needs to be running for `npm test`/`lint`/`build` — the test suite mocks the network boundary (`@/lib/api`) rather than hitting a real server; see `docs/TESTING.md`.

## Environment variables

Two, both public and both read at build time:

- `NEXT_PUBLIC_API_BASE_URL` — defaults to `/api` (same-origin), so local development behind a reverse proxy works with zero configuration.
- `NEXT_PUBLIC_SITE_URL` — the public origin used for canonical tags, Open Graph URLs, `sitemap.xml`, `robots.txt`, and the generated `llms.txt` files. Falls back to `http://localhost:3000`, which is only correct in development. A production build that forgets it emits localhost URLs in its SEO surface; see `lib/site.ts` and `docs/DEPLOYMENT.md`.

There's no `frontend/.env.example` for these: for two public vars, the convention this project uses is setting them directly at deploy time (`vercel env add`, or the platform-equivalent) rather than a checked-in file.

## Why fonts are self-hosted local files, not `next/font/google`

Short version: this build environment can't reach `fonts.googleapis.com`. Full reasoning, and why the specific typeface choice (Atkinson Hyperlegible) is functional rather than stylistic: `app/README.md` and `app/fonts/README.md`.

## Testing philosophy in one sentence

Every component takes plain props and renders them - the few that own a real user action (`IntakeForm` submitting, `EvidenceForm` saving, `PackagesExplorer` searching) fetch through `lib/api.ts` themselves, with the network boundary mocked in tests. That is what makes every other component testable with React Testing Library alone, and what makes the one genuine integration test (`app/cases/[id]/page.test.tsx`, rendering the actual page against a mocked API boundary) meaningfully different from, and additive to, the component-level tests around it. Full breakdown: `docs/TESTING.md`.

## This directory is part of the codebase's documentation convention

Every folder and subfolder in this repository has a `README.md` explaining what it does and why it's built the way it is. **If you add a new folder anywhere in this frontend, give it a `README.md` in the same change, not as a follow-up.** See the repository root `README.md`'s "How this codebase documents itself" section.
