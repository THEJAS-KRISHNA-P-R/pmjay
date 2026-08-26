# `frontend/app/components`

Every React component this frontend has, all client components (`"use client"` where they need interactivity; the presentational ones don't need the directive but are still only ever rendered from client or server components) except `icons.tsx`, which is not a component at all — see its own section below. Ten components live directly in this folder; a further seven, the marketing/content sections used across the five content pages, live in `components/landing/` (its own `README.md`) since they're a genuinely different cluster: no client-fetched data, no test files (see that README for why), reused across multiple pages rather than owned by one. None of the seventeen reach into global state or a shared store; everything is props in, JSX out, which is what makes each one independently unit-testable (see `docs/TESTING.md`).

## The ten components in this folder, grouped by what they're for

### Site chrome & Layout

- **`Header.tsx`** — the public site navbar: brand link, desktop nav links to the four content pages (with `usePathname`-based active-state styling), a mobile hamburger menu (`useState` + `Escape`-to-close + auto-close on route change), and the PMJAY helpline (`tel:14555`).
- **`AppShell.tsx`** — the authenticated-feeling product shell for dashboard, case management, and settings: locked 60px header, fixed desktop sidebar (`md:pl-60 lg:pl-64`), and mobile bottom navigation tab bar.
- **`IntakeForm.tsx`** — the actual form, rendered on the home page and `/cases/new`.

### Workspace & Case Management

- **`CaseCard.tsx`** — compact preview card for listing saved cases on the dashboard, displaying tier badge, date, hospital dispute summary, and status.
- **`ComplaintStatusTracker.tsx`** — local interactive state picker for tracking CGRMS formal grievances (Draft, Submitted, Under Review, Resolved).
- **`DisclaimerNote.tsx`** — standardized muted legal disclaimer banner rendered directly below the care-first banner on case reports.
- **`CareFirstBanner.tsx`** — the frontend's structural mirror of the backend's unconditional care-first guarantee.
- **`TierBadge.tsx`** — one outcome → an icon, a label, and a short description.
- **`TierPanel.tsx`** — composes `TierBadge` with the tier message and verified citation.
- **`ActionSteps.tsx`** — the numbered "what to do right now" list.
- **`CopyableTextBox.tsx`** — one-click copyable text box for dialogue scripts and complaint letters.
- **`HandoffPanel.tsx`** — rendered only for the handoff outcome, surfacing NALSA's free legal aid hotline (15100).
- **`CaseDocumentPanel.tsx`** — links to `GET /api/v1/cases/{id}/document` for the official downloadable/printable PDF.
- **`EvidenceForm.tsx`** — the staff-name/time/written denial note capture form.

## `icons.tsx`: the one shared, non-component file in this folder

Every icon anywhere in this frontend — landing page, tier badges, buttons, the FAQ chevron — is a small hand-drawn inline SVG exported from this one file, not an icon package. That's deliberate, matching the rest of the app: zero third-party UI dependencies (see `package.json` — there was no icon library installed before this file existed, and adding one just to render checkmarks and arrows would be the first UI dependency in the project for very little gain). Each export is a pure function of `{ className }` — no internal state, no conditional rendering, no accessibility semantics of its own. That last point is the reason `icons.tsx` doesn't have a matching `icons.test.tsx` alongside the ten components' tests below: accessibility (whether an icon is `aria-hidden` or carries a label) is entirely the calling component's responsibility and is already covered by that component's own tests (e.g. `TierBadge.test.tsx` asserting its icon wrapper is `aria-hidden`) — a test file for `icons.tsx` itself would only be asserting that each of ~20 near-identical wrapper functions returns an `<svg>`, which isn't meaningfully different from a type error the compiler already catches.

If you need a new icon, add it here rather than inline in a component — even a "one-off" icon tends not to stay one-off in a five-outcome, multi-page app, and a second hand-rolled checkmark elsewhere is exactly the kind of drift this file exists to prevent.



Data fetching happens exactly once, in `app/case/[id]/page.tsx` (client-side, via `lib/api.ts`'s `getCase`) or in `IntakeForm.tsx`'s own submit handler (via `createCase`). Every component in this folder is handed already-fetched data as props and renders it — no component here calls `fetch` itself except the two that own a real user action (`IntakeForm` submitting, `EvidenceForm` saving). This is what makes every component in this folder testable with React Testing Library alone (render with props, assert on output) rather than needing a mocked network layer in every single test file — the mocking only has to happen at the two real network-touching boundaries, and at the page level for the one genuine integration test (see `frontend/app/case/[id]/README.md`).

## If you're adding a new component

- Keep the "plain props in, JSX out" shape unless there's a specific reason not to (a genuine new network-touching action, following `IntakeForm`/`EvidenceForm`'s pattern).
- Give it a same-named `ComponentName.test.tsx` in this same folder — every one of the ten components here has one; a component without a test is the exception that needs justifying, not the default. (`icons.tsx` is the one deliberate exception in this folder, and it isn't a component — see its own section above for why. `components/landing/` is a different folder with its own, different, testing rationale — see its README.)
- If it renders an outcome-dependent color or icon, add to `TierBadge.tsx`'s `TIER_STYLES`, don't invent a second styling map elsewhere.
- If it needs an icon, add it to `icons.tsx` and import it — don't inline a new SVG or reach for an icon package.
- If it needs a status/category tag, use the `.badge` class (`app/globals.css`) plus color utilities — never `rounded-full` for that purpose. See `app/README.md`'s "Surfaces" section for the reasoning.
- This file is part of the codebase's documentation convention — see the repo root `README.md`. Keep it in sync with the code, in the same change.
