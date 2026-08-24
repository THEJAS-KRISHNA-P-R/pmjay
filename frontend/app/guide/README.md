# `frontend/app/guide`

The `/guide` route ("Your Rights" in the navbar). One file, `page.tsx`.

## What this page is, and isn't

This is the one page in the app whose content didn't already exist elsewhere as a component — it's a from-scratch, plain-language explainer covering what PMJAY covers, a family's core rights at an empanelled hospital, what hospitals aren't allowed to do, common denial patterns, an ordered "what to do at the desk" sequence, and when to escalate to NALSA. Every specific claim in it (cashless treatment, the Ayushman Mitra kiosk requirement, the no-advance-deposit rule) is one already established and used elsewhere in this codebase — `components/landing/FaqSection.tsx` and `components/landing/ScenarioGrid.tsx` state the same facts — reorganized into a single reference page rather than scattered across FAQ answers and scenario cards. Nothing on this page is a new legal claim invented for the redesign; it's existing, already-reviewed content given a proper home.

This page is explicitly framed (see its own intro paragraph) as general guidance, not a restatement of the official scheme rules — it links out to NALSA and the official portals (via `Footer`) rather than positioning itself as the final word.

## The sticky table of contents

`lg:grid-cols-[220px_1fr]` — a sticky section-jump sidebar, but only from the `lg` breakpoint up (`hidden lg:block`). Below that, it's not rendered at all; the page is just a long, linear scroll. This is a deliberate mobile-first call: a family reading this on a phone at a hospital doesn't need a secondary navigation UI competing for a small screen, but the same content on a laptop/desktop benefits from being able to jump straight to "What hospitals can't do" without scrolling past everything else. Every section has `id` + `scroll-mt-24` specifically so the anchor links land below the sticky header instead of underneath it.

## If you're changing this page

- Keep every specific factual claim consistent with `FaqSection.tsx` and `ScenarioGrid.tsx` — if you update a claim here, check whether it needs updating in those two places as well (and vice versa).
- Adding a new `<section>`: give it an `id`, `scroll-mt-24`, and a matching entry in the `SECTIONS` array so the sidebar TOC stays accurate — an untracked section is a real gap on desktop, not just a cosmetic miss.
- This file is part of the codebase's documentation convention — see the repo root `README.md`. Keep it in sync with the code, in the same change.
