# `frontend/app/packages`

The no-model package search page. `/packages` is the one public surface that
keeps working when the deployment has no provider key — which is exactly the
state a family arrives in when the intake flow answers `503`.

- `page.tsx` — server component. Owns the SEO metadata (title, canonical,
  social card) and the explanatory copy: what "verified" means, what "name
  only" means, and the honest coverage statement (315 names of roughly 1,900
  in the full master list). It is static: no data is fetched at build time.
- `PackagesExplorer.tsx` — client component. Debounced calls to
  `searchPackages` in `lib/api.ts`, a specialty dropdown, pagination, and
  the result rows. It ranks nothing and computes no rates; the back end is
  the only judge of what matches.

## What this page deliberately does not do

It does not interpret the result. A hit means "this index contains this
procedure", not "your denial is wrong"; a miss means "not in this index",
never "not covered by PM-JAY". The copy says both, because a lookup table
with a search box invites exactly the opposite reading.

It also does not print a package code or a rate for an unverified record.
The backend omits both; this page renders the omission as a "Name only"
badge plus a sentence explaining that neither was checked. Do not add a
fallback that "helpfully" reconstructs either one.

## Tests

No dedicated test file. The component's behaviour is covered indirectly by
the backend's package-search tests (rate and code stripping, pagination) and
by `npm run build`, which type-checks the hand-mirrored types in
`lib/types.ts` against their use here. If this file grows real client logic
— filtering, sorting, local caching — add a Vitest file alongside it.
