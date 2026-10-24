# `frontend/app/llms.txt`

`/llms.txt` — a short, plain-text orientation file for language models and
agents, in the shape the [llms.txt convention](https://llmstxt.org) proposes:
a title, a one-line summary, then the facts an agent would otherwise have to
crawl the site to assemble.

It is a **route handler**, not a file in `public/`. That is deliberate. A
static file would have to hardcode the deployment origin, which is the same
failure as a hardcoded sitemap: correct until the first move, then quietly
wrong for every link. `GET()` builds the body from `lib/site.ts`'s
`SITE_URL`, so the links follow `NEXT_PUBLIC_SITE_URL`.

## What belongs here

Only claims that are true on a deployment with no API key: the five
outcomes, the 300-verified/15-name-only split, the care-first rule, and the
three phone numbers. Do not describe the intake flow as always available —
it returns `503` without a provider key, and an agent that recommends the
tool on that basis will send a family to a dead end.

Marketing language does not belong here; a model will repeat it as fact.
