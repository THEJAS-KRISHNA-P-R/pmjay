# Contributing

Thanks for considering it. This is a small public-interest project with one
non-negotiable rule and a strong preference for honest, checkable work over
volume.

## The rule

**Get treatment first. Dispute the money after. Always.**

Any change that could encourage someone to delay, refuse, or cancel care is
rejected regardless of how good the code is. The backend enforces this
structurally in `internal/response` and `internal/tiering`; the frontend
mirrors it. Do not weaken those tests.

## Before you start

- **Never include patient data.** No names, case links, Aadhaar numbers, UHIDs,
  hospital records, or screenshots containing them — in issues, pull requests,
  tests, or fixtures.
- **A rate correction is a data change, not a code change.** The dataset's
  source of truth is `backend/internal/hbp/data/hbp_packages.json`, generated
  by the scripts in that directory. Hand-editing the generated JSON or the
  skill's reference files will be rejected; a packaging check treats them as
  derived artefacts.
- **A missing package is not a bug fix you can guess at.** If you cannot cite
  the government schedule (version and date) for a code or rate, the record
  stays `verified: false` with no rate and no published code. That is a valid
  state, not a failure.

## Data corrections

1. Find the record in `backend/internal/hbp/data/hbp_packages.json`.
2. Confirm the value against a published National Health Authority / State
   Health Agency HBP schedule. Record the source in `docs/DATA_SOURCES.md`.
3. Update the generation script or its input, not the generated file by hand,
   so the provenance survives.
4. Regenerate the skill data:

   ```bash
   python skills/tools/build_data.py
   python skills/package.py --check
   ```

5. Run `go test ./...` — `internal/hbp` validates that no unverified record
   carries a rate.

If you only have a lead (a PDF that might be the right schedule, a state
circular), open an issue with the link instead of changing data.

## Code

### Backend

- Go 1.22+, standard library only. Adding a third-party module needs a strong
  reason; the zero-dependency constraint is deliberate (see `ARCHITECTURE.md`).
- Keep the deterministic layer deterministic. `internal/tiering` and
  `internal/response` must not call a model.
- New HTTP behaviour needs a test in `internal/api` that asserts the status
  code, headers, and body shape — including the negative case (wrong input
  rejected, unverified data not leaked).
- Run:

  ```bash
  cd backend
  go build ./...
  go vet ./...
  gofmt -l .          # must print nothing
  go test ./... -race -cover
  ```

### Frontend

- Next.js 16 App Router, TypeScript strict, Tailwind v4, no third-party UI
  libraries.
- `tsc`, ESLint, tests, and a production build must all pass:

  ```bash
  cd frontend
  npm run lint
  npx tsc --noEmit
  npm test
  npm run build
  ```

- Truthfulness is a review criterion. Do not add copy that implies guaranteed
  coverage, guaranteed approval, "100% cashless", or locally-only storage.
  Describe server storage, AI processing, and the 30-day retention window as
  they actually are.

### Skill

- `skills/pmjay-denial-advocate/` must contain no frontend code, no build
  artefacts, no secrets, and no absolute local paths.
- `SKILL.md` must keep the care-first rule and the pending-vs-denied
  distinction.
- Validate before committing:

  ```bash
  python skills/package.py --check
  ```

## Documentation convention

Every folder in this repository has a `README.md` explaining what it does and
why it is built that way. A new folder gets one in the same change. A change
that meaningfully alters a folder's behaviour updates its README in the same
change. Stale docs are worse than missing ones.

## Commits and pull requests

- Keep a commit to one concern. Explain *why* in the message, not just what
  changed.
- In the pull request, state what you ran and what you could not verify. "Tests
  pass" without the command is not a verification claim.
- If a behaviour change affects privacy, retention, or the care-first rule,
  call it out explicitly so it gets a careful read.

## License

By contributing, you agree your work is licensed under the repository's
[MIT license](LICENSE).
