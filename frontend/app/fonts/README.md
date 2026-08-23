# `frontend/app/fonts`

Self-hosted font files, loaded by `../layout.tsx` via `next/font/local`. Not fetched at runtime from any CDN — every file here is checked into the repo.

## Files

| File | What it is |
|---|---|
| `AtkinsonHyperlegible-Regular.ttf`, `AtkinsonHyperlegible-Bold.ttf` | The functional-text typeface — used for everything a family actually has to read. |
| `Fraunces-Variable.ttf` | The display typeface — used only for the wordmark/hero title, a single variable-font file covering the full 300–900 weight range. |
| `OFL-AtkinsonHyperlegible.txt`, `OFL-Fraunces.txt` | The SIL Open Font License text for each typeface — kept alongside the font files themselves so licensing terms travel with the assets they cover, not buried in a separate legal folder no one thinks to check. |

## Why these are here at all, instead of `next/font/google`

This build environment's network sandbox doesn't reach `fonts.googleapis.com` — the same category of constraint that pushed the Go backend to zero third-party dependencies (see `../../../ARCHITECTURE.md`). Both fonts are genuinely published on Google Fonts; the actual `.ttf` files were fetched directly from Google's own font repository on GitHub (where Google Fonts' source files are hosted) and checked in here instead of being fetched by Next.js at build time from Google's CDN. The end result — self-hosted, `next/font/local` — is functionally identical to what `next/font/google` would have produced: zero runtime CDN dependency, zero layout shift. This just moved *when* the fetch happens, not *what* gets served.

## Why Atkinson Hyperlegible specifically, not a default system-ui stack

Not a stylistic choice — see `../README.md` and `../globals.css`'s header comment for the fuller reasoning, but the short version: it's a typeface published by the Braille Institute of America, designed specifically to maximize character-recognition accuracy for readers who find text difficult to read correctly. Given this product's named user (a family, often under real stress, trying to understand a dense, high-stakes situation quickly and correctly), that's a functional requirement, not a preference — every word actually explaining a family's situation or their options renders in this face. Fraunces is reserved for the wordmark alone, specifically so it never competes with Atkinson Hyperlegible for the reading a family actually needs to do correctly.

## If you're adding a font

Check the license terms first (an OFL-licensed font, like both of these, explicitly permits redistribution as embedded files — not every font license does), fetch the actual files rather than linking a CDN (given the network-sandbox constraint above still likely applies to wherever this is built), and keep the license text file alongside the font files, matching the existing two.
