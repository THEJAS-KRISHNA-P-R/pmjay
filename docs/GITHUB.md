# Making the repository public

A checklist for publishing this repo. Nothing here runs automatically; it is
the set of settings and copy that make the project discoverable and safe to
open.

## Repository metadata

**Name:** `pmjay` (or `pmjay-advocate`)

**Description** (GitHub's 350-character field, used by search and social cards):

> Free, independent help for Indian families told a hospital will not cover treatment under Ayushman Bharat (PM-JAY). Search verified HBP package codes and rates, get a care-first next step, and never delay treatment for a billing dispute. Go + Next.js, MIT.

**Website:** the deployed frontend origin, once one exists. Leave blank until
then rather than pointing at `localhost`.

**Topics** (GitHub allows 20; these match how people actually search):

```
ayushman-bharat  pmjay  health-insurance  india  patient-rights
health-benefit-package  claims  denial-management  legal-aid
public-interest  healthcare  government-schemes  nextjs  golang
agent-skill  civic-tech
```

**Social preview:** upload a 1280×640 image (the logo on a plain background is
enough) under Settings → General → Social preview. Without one, GitHub shows a
generic card when the link is shared.

## Visibility and access

1. Settings → General → Danger Zone → **Change visibility → Public**.
2. Confirm that `backend/.env`, `frontend/.env.local`, and `dist/*.zip` are not
   tracked (`git ls-files | findstr env` on Windows, `git ls-files | grep env`
   elsewhere). `.gitignore` covers them, but check the index, not just the
   ignore file.
3. Settings → Branches → protect `main`: require a pull request and a passing
   CI check. The workflow in `.github/workflows/ci.yml` is the check.
4. Settings → Code security → enable **Private vulnerability reporting** and
   **Dependabot alerts**. `SECURITY.md` tells reporters to use the private
   advisory flow.

## Releases

Tag the first public version and attach the skill archive:

```bash
git tag -a v1.0.0 -m "PMJAY Denial Advocate v1.0.0"
git push origin v1.0.0

python skills/package.py            # builds dist/pmjay-denial-advocate-1.0.0.zip
```

Create a GitHub Release for `v1.0.0`, paste the changelog section from
`skills/agensi-listing.md`, and attach the ZIP. `dist/` is git-ignored on
purpose — release assets are the distribution channel, not the tree.

## Search and agent discovery

The app generates these from `NEXT_PUBLIC_SITE_URL`; they only exist on a
deployment, not in the repo:

- `/robots.txt` — allows public pages, blocks `/cases/`, `/dashboard`,
  `/settings`, `/api/`.
- `/sitemap.xml` — public routes only.
- `/llms.txt` and `/llms-full.txt` — factual orientation for agents.

After deploying, submit the sitemap in Google Search Console and Bing Webmaster
Tools. Being indexed is not automatic.

## What not to publish

- No `.env` file, API key, or Stripe secret. The app runs without a key; the
  key is a deployment secret.
- No `cases.json` or any real case record. `backend/data/` is git-ignored.
- No patient data in issues, releases, screenshots, or commit messages.
- No claim of NHA affiliation. The README, footer, and `/about` already state
  independence; keep it that way.

## After publishing

- Watch the first few issues for accidental case links and redact them
  promptly.
- Keep `docs/DATA_SOURCES.md` current when a rate is corrected; the provenance
  is the project's main defence against being just another unverified tariff
  table.
- Re-run `python skills/package.py --check` before every skill release, and
  bump the version in `skills/package.py` and `skills/agensi-listing.md`
  together.
