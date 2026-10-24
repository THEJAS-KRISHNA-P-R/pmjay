# Skills

A standalone, installable agent skill built from this repository's verified
PMJAY dataset. It has no dependency on the web app or the Go backend: copy the
folder and it works.

```
skills/
├── pmjay-denial-advocate/     the skill itself (this is what ships)
│   ├── SKILL.md
│   ├── references/            300 verified packages, 15 name-only, 4 exclusions
│   └── scripts/lookup.py      dependency-free candidate search
├── tools/build_data.py        regenerates reference data from backend/internal/hbp/data
├── package.py                 validates the skill and builds dist/*.zip
└── agensi-listing.md          marketplace copy, tags, and price guidance
```

## What the skill does

An Indian family is at a hospital billing desk and is told a treatment is not
covered under Ayushman Bharat (PM-JAY). The skill gives the agent one
deterministic job: decide whether that denial is **Green** (looks covered),
**Amber** (one more fact needed, including pending pre-authorisation), **Red**
(the denial is correct), **Mixed** (part covered, part excluded), or **Handoff**
(route to free legal aid), then produce a desk script, a CGRMS complaint draft,
and an evidence checklist.

The care-first rule outranks every other instruction: **get treatment first,
dispute the money after.** The skill refuses to trade care for a billing
argument.

## Install

### Claude Code / Anthropic-style skills directory

```bash
unzip pmjay-denial-advocate-1.0.0.zip -d ~/.claude/skills/
```

The archive already contains the `pmjay-denial-advocate/` folder, so this
creates `~/.claude/skills/pmjay-denial-advocate/SKILL.md`.

### Any agent that reads a skills folder

Copy the `pmjay-denial-advocate` directory anywhere your agent loads skills
from. The only requirement is that `SKILL.md` sits at the folder root and the
agent can execute `python scripts/lookup.py`.

### Requirements

- Python 3.9 or newer.
- No third-party packages. Standard library only.
- No network access. The data ships inside the skill.

### Try it

```bash
cd pmjay-denial-advocate
python scripts/lookup.py "dialysis kidney failure"
python scripts/lookup.py "gallbladder stones operation" --verified-only
python scripts/lookup.py "ivf fertility test tube baby" --exclusions
```

`lookup.py` is a recall-biased shortlist generator. It is not the decision
layer; `references/triage-rules.md` is.

## Rebuilding the data

The reference data is generated, never hand-edited, so every number keeps its
provenance:

```bash
python skills/tools/build_data.py
```

This reads `backend/internal/hbp/data/hbp_packages.json`, keeps only records
marked `verified: true`, and strips every rate field from the unverified ones.
The split is a safety property: an agent may recognise an unverified procedure
name, but it cannot quote a rate that was never checked.

## Packaging

```bash
python skills/package.py --check   # validate only
python skills/package.py           # validate + build dist/pmjay-denial-advocate-1.0.0.zip
```

The validator fails the build if the bundle contains frontend code, build
artifacts, secrets, absolute local paths, invalid JSON, a leaked rate on an
unverified record, or a `SKILL.md` that lost the care-first rule.

## Coverage, stated plainly

- 300 packages verified against a published government HBP schedule.
- 15 procedures known by name only; no code and no rate.
- 4 confirmed exclusion categories.
- Roughly 1,900 procedures exist in the full HBP master list, so this covers
  about one sixth of it. Common general-surgery, orthopaedic, obstetric,
  ophthalmic, ENT, urology, and neurosurgery procedures are the known gap.
- Rates are indicative tariffs, not a guarantee of eligibility, empanelment, or
  the amount a hospital will charge.

This is why the listing guidance recommends launching free or cheap and saying
so in the description. Selling an incomplete tariff table as complete would be
both dishonest and commercially short-lived.

## Licence

MIT, same as the repository. The dataset is an independent compilation of
publicly available National Health Authority material and is not affiliated
with or endorsed by the NHA.
