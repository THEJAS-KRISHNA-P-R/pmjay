# `frontend/app/case`

Legacy namespace. `/case` (singular) used to host the whole case-detail
experience; it now only holds two redirect shims to `/cases` (plural),
kept so that any bookmarked or previously-shared `/case/:id` link keeps
working (a case's URL is the only way to reach it — no login, see
`../../ARCHITECTURE.md` — so an old link has to keep resolving).

- `page.tsx` — redirects `/case` → `/cases/new`.
- `[id]/page.tsx` — redirects `/case/:id` → `/cases/:id`.

The actual case workspace, and the note this file used to end on about
a `/case` index page being impossible without a way to look up a
family's past cases — now solved by `frontend/lib/caseHistory.ts`'s
local, no-login case list — both live under `frontend/app/cases/`
instead. See `frontend/app/cases/[id]/README.md`.
