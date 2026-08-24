# `frontend/app/faq`

The `/faq` route. One file, `page.tsx` — thin wrapper around the shared `FaqSection` component plus a closing "call the helpline" card.

## Why this is separate from `components/landing/FaqSection.tsx`

The FAQ content itself lives in the component (see `app/components/README.md`), same as `HowItWorks`/`Comparison`/`ScenarioGrid`. This file is just the dedicated page shell — header, page title, the component, a closing CTA — so `/faq` is a real, linkable, shareable URL instead of only existing as a scroll-to section on the home page.

## If you're changing this page

- FAQ content itself (the questions and answers) belongs in `components/landing/FaqSection.tsx`, not here — this file should stay a thin page shell.
- This file is part of the codebase's documentation convention — see the repo root `README.md`. Keep it in sync with the code, in the same change.
