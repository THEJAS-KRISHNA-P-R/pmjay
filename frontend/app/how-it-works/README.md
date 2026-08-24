# `frontend/app/how-it-works`

The `/how-it-works` route. One file, `page.tsx` — a static server component, no client-side state.

## Why this is its own page, not just a home page section

It used to be: the `HowItWorks`, `Comparison`, and `ScenarioGrid` landing components all rendered directly on `app/page.tsx`, alongside everything else, making the home page an ever-growing single scroll. They're unchanged as components — this page just gives them a dedicated home with room to breathe, and a URL someone can be sent directly (a support message, a WhatsApp forward) without having to scroll past the intake form to find them.

## The "behind the scenes" section

The one piece of content that's genuinely new here (not moved from elsewhere): a plain-language explanation of the actual extraction → tiering pipeline (LLM extracts facts, deterministic Go logic decides the outcome — see `backend/internal/response/README.md` for the real implementation). This exists because "an AI tool told me this" is a reasonable thing for a stressed family to be skeptical of, and the honest answer — the AI never makes the actual coverage decision, fixed rules do — is reassuring precisely because it's specific and checkable, not just a vague "trust us."

## If you're changing this page

- `HowItWorks`, `Comparison`, `ScenarioGrid` are shared components (`app/components/landing/`) — also still importable elsewhere if a future page needs them. Don't fork their content here; edit the component.
- This file is part of the codebase's documentation convention — see the repo root `README.md`. Keep it in sync with the code, in the same change.
