# `frontend/app`

Next.js 16 App Router root. Three top-level files plus three subfolders.

## Top-level files

| File | What it does |
|---|---|
| `layout.tsx` | Root layout — loads the two self-hosted fonts, sets page metadata, wraps every page. |
| `page.tsx` | The home page (`/`) — headline, `IntakeForm`, the always-visible emergency helpline line. |
| `global-error.tsx` | Next.js's special top-level error boundary — catches an error that escapes everything else. |
| `globals.css` | Tailwind v4 import + every design token (colors, fonts) as CSS custom properties. |

## Subfolders

| Folder | What it is |
|---|---|
| `components/` | Every React component — see `components/README.md`. |
| `case/[id]/` | The result page — see `case/[id]/README.md` (and `case/README.md` for why that intermediate folder exists). |
| `fonts/` | Self-hosted font files — see `fonts/README.md`. |

## `layout.tsx`: why fonts are loaded from local files, not `next/font/google`

Same network-sandbox story as the Go backend's zero third-party dependencies (see `../../ARCHITECTURE.md`): this build environment doesn't reach `fonts.googleapis.com`. The actual OFL-licensed font files were fetched directly from Google's own font repository on GitHub and are checked into `fonts/`, loaded via `next/font/local`. Functionally identical to what `next/font/google` would produce — self-hosted, zero runtime CDN dependency, zero layout shift — this just does the fetch at build-prep time instead of expecting Next.js to do it at first build. If this project's build environment ever gets network access to Google Fonts, switching to `next/font/google` would be a drop-in replacement; there's no reason to make that switch unless the constraint that caused this changes, since the current approach has strictly fewer runtime dependencies.

**Atkinson Hyperlegible for all functional text** is a deliberate, functional choice, stated directly in the source comments in both `layout.tsx` and `globals.css`: it's a typeface published by the Braille Institute of America specifically for readers who find text difficult to read accurately — a real requirement given this product's named user, not an aesthetic preference. Fraunces (a display serif) is used only for the wordmark/hero title; every word a family actually has to read to understand their situation stays in the hyperlegible face. If you're adding new UI text anywhere in this app, it inherits this by default (`font-sans` maps to Atkinson via the CSS variable in `globals.css`) — you'd have to deliberately opt out to use Fraunces or a system font instead, which should be a conscious, justified choice, not a default.

## `globals.css`: why the tier colors are muted, not a typical status-light red/amber/green

Read the file's own header comment before changing any tier color — it's explicit that this is a UX decision downstream of the product's own care-first framing, not a generic aesthetic choice: a family reading this page may already be frightened, and the red (exclusion-confirmed) tier especially should read as "clear and calm" rather than "alarm klaxon." Deep teal as the primary color and warm marigold as the accent were chosen to evoke trust and public-health calm without copying any specific government scheme's actual branding. If you're adjusting the palette, keep the muted-tier-colors reasoning in mind — it's connected to `docs/SAFETY_DESIGN.md`'s broader design philosophy, not just this file's own taste.

## `global-error.tsx`: the one page that renders even when everything else fails

Next.js's reserved top-level error boundary — catches an error that escapes every other error handling in the app (a render crash in `layout.tsx` itself, for instance, which no page-level error boundary could catch). Even here, at the absolute last line of defense, the care-first framing holds: the fallback UI still surfaces the PMJAY helpline (`tel:14555`) alongside the generic "something went wrong" message and a "Try again" button — a family hitting the worst-case frontend failure still isn't left with nothing actionable.

## If you're extending this folder

- **A new top-level route**: a new folder here with its own `page.tsx`, following `case/[id]/`'s pattern if it needs a dynamic segment.
- **A new design token**: add it to `globals.css`'s `@theme` block, with a comment explaining the reasoning if it's not obvious from the token name alone — matching the standard every existing token in that file sets.
- **This file is part of the codebase's documentation convention** — see the repo root `README.md`. Keep it in sync with the code, in the same change.
