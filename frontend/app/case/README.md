# `frontend/app/case`

A pure Next.js App Router namespace — this folder itself has no `page.tsx` and renders nothing on its own. `/case` is not a real route; it exists only so `[id]/` can define the dynamic `/case/:id` route. See `frontend/app/case/[id]/README.md` for the actual page.

If a genuine `/case` index page is ever needed (a list of a family's past cases, say — nothing in this build persists a way to look that up without knowing the exact ID, see `backend/internal/store/README.md`'s note on case IDs being effectively bearer credentials), it would live directly in this folder as `page.tsx`, alongside the `[id]/` subfolder, not inside it.
