# `frontend/app/about`

The `/about` route. One file, `page.tsx`.

## What's here

Mission framing, two grounding cards (data honesty, independence from government), the shared `SafetyPledge` component (also still used as-is, unchanged), and an explicit "what this is not" list. The independence and "what this is not" framing matter specifically because this tool is asking families to describe a real, often distressing situation — it should be unambiguous, in writing, that this isn't an official government channel and doesn't submit anything on a family's behalf.

## If you're changing this page

- `SafetyPledge` is a shared component (`app/components/landing/SafetyPledge.tsx`) — edit it there if the pledges themselves change; this page just renders it alongside additional About-specific context.
- If you add a new claim about data sources or verification counts, keep it consistent with `components/landing/Features.tsx`'s HBP package numbers — both should always state the same total/verified counts.
- This file is part of the codebase's documentation convention — see the repo root `README.md`. Keep it in sync with the code, in the same change.
