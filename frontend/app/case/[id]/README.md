# `frontend/app/case/[id]`

`page.tsx` here is a one-line redirect to `/cases/:id` — nothing else.
This used to be the real case-detail page; see
`frontend/app/cases/[id]/README.md` for where that composition, its
tests, and this file's previous (much longer) documentation actually
live now. Kept as a redirect, not deleted, because a case's URL is the
only way anyone reaches it (no login — `../../../ARCHITECTURE.md`), so
an old bookmarked or shared `/case/:id` link still has to work.
