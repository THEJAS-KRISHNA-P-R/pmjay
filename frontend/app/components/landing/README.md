# `frontend/app/components/landing`

Seven components: `Features.tsx`, `HowItWorks.tsx`, `Comparison.tsx`, `ScenarioGrid.tsx`, `SafetyPledge.tsx`, `FaqSection.tsx`, `Footer.tsx`. A different cluster from the ten in `app/components/` proper — see that folder's `README.md` for how the two relate.

## Why these live separately from `app/components/`, and why none of them has a test file

Everything in `app/components/` either takes case data as props (the result-page components) or owns a real network action (`IntakeForm`, `EvidenceForm`) — that's the whole reason that folder's convention is "every component gets a test." Nothing here does either: every component in this folder is static marketing/informational content — hardcoded arrays of strings rendered into JSX, `useState` at most for local UI state (an open FAQ accordion, an open feature modal), never a prop, never a fetch. A test file for `Features.tsx` would only be asserting that six hardcoded strings appear in six hardcoded places — no meaningfully different from the compiler already catching a typo in the array. That's the actual reason there's no `Features.test.tsx` etc., not an oversight — but it's also why this folder didn't have a `README.md` of its own for a while, since the "why no tests" question needed answering somewhere and it hadn't been written down. This file is that answer.

## Which page renders which component

None of these is single-page-owned anymore — they're shared across the five content pages:

| Component | Rendered on |
|---|---|
| `Features.tsx` | `app/page.tsx` (condensed grid + click-through modal) |
| `HowItWorks.tsx` | `app/how-it-works/page.tsx` |
| `Comparison.tsx` | `app/how-it-works/page.tsx` |
| `ScenarioGrid.tsx` | `app/how-it-works/page.tsx` |
| `SafetyPledge.tsx` | `app/about/page.tsx` |
| `FaqSection.tsx` | `app/faq/page.tsx` |
| `Footer.tsx` | every page except none — it's the one component here rendered on all six routes |

This used to be simpler and worse: all seven rendered on `app/page.tsx`, in this same order, one after another, making the home page an ever-growing single scroll with no way to link directly to, say, just the FAQ. Splitting them across dedicated pages didn't require touching any of these components' internals — it was purely an `app/page.tsx` composition change (see `app/README.md`).

## Content accuracy: three numbers that have to stay in sync

`Features.tsx` states "315" HBP packages and "300 independently verified." The same two numbers are restated in `app/about/page.tsx` and in `app/guide/page.tsx`. If the backend's `internal/hbp/data/hbp_packages.json` dataset ever changes size or verified-count, all three places need updating together — there's no single source of truth these pull from at build time, so this is a manual-consistency responsibility, not something TypeScript will catch for you.

## `Footer.tsx`: `SHELL_WIDTH`, `SITE_LINKS`, and `LEGAL_LINKS`

Organized into a 4-column responsive grid linking site exploration pages (`SITE_LINKS`), compliance and privacy policies (`LEGAL_LINKS`: Privacy Policy, Terms of Service, Legal/Medical Disclaimers, Data Settings), and emergency statutory contact numbers (14555, 15100, 112). Features a persistent bottom legal notice and direct links to `llms.txt`.

## If you're adding a new section here

- Only add a component to this folder if it's genuinely static/informational content with no props and no network call — anything else belongs in `app/components/` instead, with a matching test file.
- Use `.card` and `.badge` (see `app/README.md`'s "Surfaces" section) — no `rounded-full` status tags, no glassmorphism/`backdrop-blur`.
- If it needs an icon, it's in `../icons.tsx` already or belongs there — don't inline a new SVG.
- This file is part of the codebase's documentation convention — see the repo root `README.md`. Keep it in sync with the code, in the same change.
