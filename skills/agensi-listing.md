# Agensi listing copy

Draft listing for the standalone skill. Everything below is copy you can paste
into the marketplace form; nothing here is submitted automatically.

## Title

PMJAY Denial Advocate

## Tagline (max ~80 characters)

Turn a PM-JAY denial into a calm, evidence-backed next step.

## Short description (marketplace card, ~200 characters)

An agent skill for Indian families told a hospital will not cover treatment
under Ayushman Bharat. Care-first triage, 300 source-verified HBP packages, and
the exact words to use at the billing desk — never a guess.

## Description

**A denial at the billing desk is not an answer. It is a question with a
deadline.**

PMJAY Denial Advocate is an agent skill for the moment an Indian family is
told that a treatment is not covered under Ayushman Bharat (PM-JAY), that they
must pay cash for a supposedly covered package, or that the pre-authorisation
is "still pending."

It does one job: decide whether that refusal is correct, and say what to do
next in language a frightened person can act on at a hospital counter.

### What it returns

- **Green** — looks like a real covered package. Package name, HBP code, and
  the indicative rate, with the exact sentence to say at the desk.
- **Amber** — genuinely unclear, a close call, or a pending pre-authorisation.
  One targeted question instead of a guess.
- **Red** — the denial is correct. It says so plainly, explains the exclusion,
  and points to real alternatives instead of manufacturing a grievance.
- **Mixed** — part of the bill is covered and part is a confirmed exclusion.
  Split the bill; do not answer as one yes or no.
- **Handoff** — too tangled for a guided flow. Routes to a free NALSA Para
  Legal Volunteer with the context already written down.

Every outcome comes with a short hospital script, a CGRMS complaint draft with
explicit placeholders for facts it does not have, and an evidence checklist.

### The rule that outranks everything else

> Get treatment first. Dispute the money after. Always.

The skill will not tell anyone to delay, cancel, or refuse care while a billing
question is open. If the description sounds like an emergency, it leads with
care and the emergency numbers before any analysis.

### What is inside

| File | Contents |
|---|---|
| `SKILL.md` | The workflow, decision protocol, and prohibitions. |
| `references/hbp-packages.json` | **300** packages whose code, name, specialty, and rate were checked against a published government HBP schedule. |
| `references/hbp-packages-unverified.json` | **15** real PMJAY procedure names with **every rate field removed**. Name recognition only. |
| `references/exclusions.json` | The **4** confirmed exclusion categories, with nuances. |
| `references/triage-rules.md` | The exact thresholds and decision order. |
| `references/escalation-and-rights.md` | CGRMS, 14555, NALSA 15100, evidence checklist, detention and coercion. |
| `references/response-templates.md` | Desk script, complaint draft, handoff summary. |
| `scripts/lookup.py` | Dependency-free candidate search. No API key, no network. |

### Honest coverage

The full HBP master list has roughly **1,900 procedures**. This skill covers
**315 names**, of which **300 carry a verified code and rate**. Seven major
specialties are thin: general surgery, orthopaedics, obstetrics and
gynaecology, ophthalmology, ENT, urology, and neurosurgery. Common procedures
such as gallbladder removal, appendectomy, hernia repair, caesarean section,
cataract surgery, and knee replacement are recognised by name but have no
verified rate in this bundle.

If the best match is unverified, the skill stays Amber and says the code and
rate must be confirmed with the hospital or 14555. It will not invent a rupee
figure.

### What it deliberately does not do

- It does not submit a complaint, check live pre-authorisation status, or book
  legal aid. There is no public API for those, so it drafts and routes instead.
- It does not claim to be the National Health Authority, a hospital, or a
  lawyer.
- It does not ask for an Aadhaar number or full medical records to triage.
- It does not treat a pending pre-authorisation as a denial.

### Works with

Any agent that can load a skills folder and run Python 3.9+. Tested with
Claude Code-style `SKILL.md` layout. No integrations, credentials, or network
calls required.

### Who it is for

Families and the people helping them: community health workers, NGO case
workers, patient-rights volunteers, journalists verifying a claim, and any
agent builder who wants a grounded, non-hallucinating PMJAY reference instead
of a model improvising tariff numbers.

### Support and updates

MIT licensed. Source, provenance notes, and the data build script ship with the
repository. If a rate is wrong, the fix belongs in the dataset so the
provenance stays intact — not in a hand-edited JSON file.

Independent public-interest project. Not affiliated with or endorsed by the
National Health Authority, the Ministry of Health and Family Welfare, any State
Health Agency, or any hospital.

## Tags

`healthcare`, `india`, `ayushman-bharat`, `pmjay`, `health-insurance`,
`patient-rights`, `legal-aid`, `claims`, `denial-management`, `public-interest`,
`government-schemes`, `agent-skill`

## Category

Healthcare / Civic & Public Interest (choose whichever the marketplace offers;
`healthcare` is the primary).

## Price guidance

**Recommended: free, or ₹199–₹499 one-time.**

Reasons to launch low:

1. Coverage is 315 of ~1,900 procedures. Charging a premium for a table that is
   one-sixth complete invites refunds and bad reviews.
2. The value is in the decision protocol and the honesty constraints, which are
   cheap to copy. The defensible asset is the verified dataset, and it is still
   growing.
3. A free or low-cost launch earns the corrections you need. Every wrong rate a
   user reports is worth more than the margin.

Suggested ladder:

| Stage | Price | Trigger to move up |
|---|---|---|
| Launch | Free | First 100 installs; collect corrections. |
| Verified | ₹199–₹499 | Verified coverage above 800 procedures, including the seven missing specialties. |
| Pro | ₹999+ | Adds current-version rate tables per state, a maintained update cadence, and a changelog with source dates. |

Do not price above ₹999 while any common procedure still returns "name only,
no rate." The description above says so explicitly; the price should agree with
it.

## Install instructions (for the listing body)

1. Download `pmjay-denial-advocate-1.0.0.zip`.
2. Unzip it into your agent's skills directory:
   `unzip pmjay-denial-advocate-1.0.0.zip -d ~/.claude/skills/`
3. Confirm `~/.claude/skills/pmjay-denial-advocate/SKILL.md` exists.
4. Ask your agent: *"My father needs dialysis and the hospital says PMJAY will
   not cover it. What do I say at the desk?"*

Python 3.9+ is required for `scripts/lookup.py`. There are no other
dependencies.

## Demo prompts

- "The hospital says gallbladder surgery is not covered under PMJAY and wants
  ₹80,000 in cash. What do I do?"
- "My mother's cataract operation was refused under Ayushman Bharat."
- "They said the pre-authorisation is pending, not rejected. Is that a denial?"
- "Is IVF covered under PMJAY?"
- "The hospital is holding my father's discharge papers until we pay."

## Changelog

### 1.0.0

- First public bundle.
- 300 verified packages, 15 name-only records, 4 exclusion categories.
- Care-first triage protocol, pending-vs-denied distinction, mixed-bill handling.
- Deterministic lookup with no network access.
- Validator refuses to package secrets, local paths, frontend code, or leaked
  rates on unverified records.

## What to say if a buyer asks "is this legal advice?"

No. It is structured guidance based on published PMJAY rules, plus a route to
free legal aid (NALSA, 15100). A Para Legal Volunteer or the PMJAY helpline
confirms the individual case.
