# `frontend/app`

Next.js 16 App Router root. Four top-level files plus six subfolders — five of those are page routes (`/`, `/how-it-works`, `/guide`, `/faq`, `/about`, `/case/[id]`), the sixth (`components/`) holds the shared UI they're all built from.

## Top-level files

| File | What it does |
|---|---|
| `layout.tsx` | Root layout — loads the two self-hosted fonts, sets page metadata, wraps every page. |
| `page.tsx` | The home page (`/`) — headline, `IntakeForm`, trust stats, a condensed `Features` grid, and an "Explore" section linking out to the other four pages. |
| `global-error.tsx` | Next.js's special top-level error boundary — catches an error that escapes everything else. |
| `globals.css` | Tailwind v4 import + every design token (colors, fonts) as CSS custom properties, plus the shared `.card`/`.badge`/`.field` component classes. |

## Subfolders

| Folder | What it is |
|---|---|
| `components/` | Every React component — see `components/README.md`. |
| `how-it-works/` | The `/how-it-works` page — see its own `README.md`. |
| `guide/` | The `/guide` ("Your Rights") page — see its own `README.md`. |
| `faq/` | The `/faq` page — see its own `README.md`. |
| `about/` | The `/about` page — see its own `README.md`. |
| `dashboard/` | The `/dashboard` page for managing locally saved cases — see `dashboard/README.md`. |
| `cases/` | Case workspaces (`/cases/[id]`, `/cases/new`) — see `cases/README.md`. |
| `settings/` | User preferences & accessibility toggle page — see `settings/README.md`. |
| `privacy/` | Privacy Policy (DPDP Act 2023 compliant) — see `privacy/README.md`. |
| `terms/` | Terms of Service & emergency care clause — see `terms/README.md`. |
| `disclaimer/` | Legal & Medical Disclaimers — see `disclaimer/README.md`. |
| `fonts/` | Self-hosted font files — see `fonts/README.md`. |

## Why five pages instead of one long scroll

This used to be a single-page site: `HowItWorks`, `Comparison`, `ScenarioGrid`, `SafetyPledge`, and `FaqSection` all rendered directly on `page.tsx`, one after another. They're unchanged as components (still all live in `components/landing/`), but each now has a dedicated page — `page.tsx` itself is now deliberately short: the hero, the intake form (the actual point of the whole app), a compact trust-stats bar, a condensed `Features` grid, and an "Explore" card grid linking to the rest. A family arriving with an urgent question shouldn't have to scroll past five sections of marketing copy to find the form; someone who wants the fuller explanation, the rights guide, or the FAQ now has a real URL to land on or be sent directly.

`Header.tsx` exports `SHELL_WIDTH` (`max-w-6xl`) — every page's `<main>`, `Header` itself, and `Footer` all use it, so the page chrome lines up on the same left/right edge everywhere. `case/[id]/` is the one deliberate exception, using its own narrower `max-w-2xl` reading column — see that folder's own README for why.

## `layout.tsx`: why fonts are loaded from local files, not `next/font/google`

Same network-sandbox story as the Go backend's zero third-party dependencies (see `../../ARCHITECTURE.md`): this build environment doesn't reach `fonts.googleapis.com`. The actual OFL-licensed font files were fetched directly from Google's own font repository on GitHub and are checked into `fonts/`, loaded via `next/font/local`. Functionally identical to what `next/font/google` would produce — self-hosted, zero runtime CDN dependency, zero layout shift — this just does the fetch at build-prep time instead of expecting Next.js to do it at first build. If this project's build environment ever gets network access to Google Fonts, switching to `next/font/google` would be a drop-in replacement; there's no reason to make that switch unless the constraint that caused this changes, since the current approach has strictly fewer runtime dependencies.

**Atkinson Hyperlegible for all functional text** is a deliberate, functional choice, stated directly in the source comments in both `layout.tsx` and `globals.css`: it's a typeface published by the Braille Institute of America specifically for readers who find text difficult to read accurately — a real requirement given this product's named user, not an aesthetic preference. Fraunces (a display serif) is used only for the wordmark/hero title; every word a family actually has to read to understand their situation stays in the hyperlegible face. If you're adding new UI text anywhere in this app, it inherits this by default (`font-sans` maps to Atkinson via the CSS variable in `globals.css`) — you'd have to deliberately opt out to use Fraunces or a system font instead, which should be a conscious, justified choice, not a default.

## `globals.css`: why the tier colors are muted, not a typical status-light red/amber/green

Read the file's own header comment before changing any tier color — it's explicit that this is a UX decision downstream of the product's own care-first framing, not a generic aesthetic choice: a family reading this page may already be frightened, and the red (exclusion-confirmed) tier especially should read as "clear and calm" rather than "alarm klaxon." Deep teal as the primary color and warm marigold as the accent were chosen to evoke trust and public-health calm without copying any specific government scheme's actual branding — both were deliberately pulled toward lower saturation and a deeper mid-tone than a typical bright marketing palette, closer to what a serious fintech/healthtech brand would use than a consumer wellness app, for the same "handling real legal/financial content" reason. If you're adjusting the palette, keep the muted-tier-colors reasoning in mind — it's connected to `docs/SAFETY_DESIGN.md`'s broader design philosophy, not just this file's own taste.

Every one of the five `--color-tier-*` groups (green/amber/red/mixed/handoff) ships four tokens — `-bg`, `-border`, `-icon`, `-text` — and `TierBadge.tsx`/`TierPanel.tsx` (the actual on-screen outcome display) consume these directly rather than a stock Tailwind color. That wasn't always true — an earlier pass had the tier tokens defined here but the badge itself rendering plain Tailwind `emerald`/`amber`/`rose`, so the one component doing the most emotionally-loaded work on the page wasn't actually using the calm palette this file argues for. If you add a sixth outcome or touch the tier colors, grep for `tier-` across `app/components/` first — every usage should resolve to one of these four suffixes, never a raw palette color standing in for one.

## Surfaces: `.card` is the only surface, and there is no glassmorphism

`globals.css` defines exactly one surface treatment: `.card` — solid white, hairline `sand-200` border, soft shadow, `rounded-2xl`. There used to be a second, `.glass` (translucent + `backdrop-blur`), reserved for the header and the home page's hero intake card. It was removed deliberately, not just simplified away: no blur-and-translucency layering anywhere in this app is a cross-project house rule (matches the same call made on other production apps this codebase's author maintains), because it reads as a trendy template rather than a serious tool handling real coverage decisions — and translucency measurably hurts text contrast the moment two blurred layers overlap, which is easy to do by accident. Where the old `.glass` treatment gave the header and hero card visual weight, they now get it from `.card` plus a stronger border/shadow instead. If you're tempted to bring blur back for a new "signature moment," don't — reach for size, shadow, or a `card`-with-accent-border treatment instead.

**Status and category tags use `.badge`, never `rounded-full`.** Same cross-project rule: a status indicator is a small, precise, `rounded-md`-ish tag (4px-ish radius, bold uppercase text, tinted border) — never a pill, and never a bare colored dot standing in for a label. `.badge` in `globals.css` is the structural half (radius, padding, uppercase, tracking); callers add the semantic color with ordinary `bg-*`/`border-*`/`text-*` utilities alongside it, e.g. `className="badge bg-teal-50 border-teal-200 text-teal-800"`. `rounded-full` still shows up a few places in this codebase — circular icon buttons (a modal close button), numbered step circles, and the browser-mockup traffic-light dots on the home page — all genuinely different patterns from a status tag, not exceptions to this rule so much as things the rule was never about.

## `global-error.tsx`: the one page that renders even when everything else fails

Next.js's reserved top-level error boundary — catches an error that escapes every other error handling in the app (a render crash in `layout.tsx` itself, for instance, which no page-level error boundary could catch). Even here, at the absolute last line of defense, the care-first framing holds: the fallback UI still surfaces the PMJAY helpline (`tel:14555`) alongside the generic "something went wrong" message and a "Try again" button — a family hitting the worst-case frontend failure still isn't left with nothing actionable.

## If you're extending this folder

- **A new top-level route**: a new folder here with its own `page.tsx` and its own `README.md` in the same change (see the five existing page folders for the shape: `Header`/`Footer` wrap a `<main>` using `SHELL_WIDTH`, `px-4 sm:px-6 lg:px-8` horizontal padding, a `metadata` export for the page title). Add it to `Header.tsx`'s `NAV_LINKS` and `Footer.tsx`'s `SITE_LINKS` if it belongs in site navigation — a page that exists but isn't linked from anywhere is effectively dead.
- **A new design token**: add it to `globals.css`'s `@theme` block, with a comment explaining the reasoning if it's not obvious from the token name alone — matching the standard every existing token in that file sets.
- **A new status/category tag**: use `.badge` plus color utilities, not `rounded-full`.
- **This file is part of the codebase's documentation convention** — see the repo root `README.md`. Keep it in sync with the code, in the same change.
