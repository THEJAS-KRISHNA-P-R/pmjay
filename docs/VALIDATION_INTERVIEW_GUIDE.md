# Validation Interview Guide

This is not new content — it's Section 33 (Appendix I) and Section 70 (Appendix HH) of the source spec, pulled out of an 1,851-line document into one page that's actually usable: something to open on a phone or print before a real conversation, rather than something that stays a good intention because it's buried at line 1012 of a spec nobody re-opens after the build starts.

**Why this exists:** `docs/OPEN_QUESTIONS.md` and `docs/TESTING.md` both say the same thing — the single highest-priority next step for this project is not more code, it's this conversation. Nothing this build produced (tiering accuracy, test coverage, safety guarantees) is evidence toward the question below, because it's a question about distribution and real-world trust, not about whether the matching logic works. This document is the fastest way to actually go find out, rather than the fastest way to feel like the question was addressed.

**Who to talk to:** an ASHA worker, a NALSA Para Legal Volunteer, or anyone who has personally dealt with a PMJAY coverage dispute. This is a short, informal conversation — not a formal research study, and not something that needs scheduling anxiety to attach to it.

---

## The conversation

**Opening** (say this, or something close to it — the goal is an honest answer, not a polite one):

> "We're building something for people dealing with hospitals about their Ayushman card, and I want your honest opinion, including if you think this is a bad idea."

**Five questions, in this order:**

1. **"Have you seen, or been part of, a situation where a hospital said something wasn't covered by a PMJAY card, and it turned out that wasn't quite right, or wasn't fully explained?"**
   Establishes whether the problem is even recognizable to this person before anything else gets asked.

2. **"In that situation, what did the family actually do?"**
   Get the real behavior, not a hypothetical.

3. **"If something had been available beforehand — like when the card was first issued — that let a family check 'is this really not covered' right there at the hospital, do you think a family would have set that up in advance?"**
   This is the load-bearing question. Everything else is context for this one.

4. **"Who would a family in that situation trust to help them — a family member, a health worker like yourself, a local leader, someone else?"**
   Tells you whether an ASHA-worker distribution channel is realistic, or whether a different channel would actually be trusted.

5. **"What would stop a family from using something like this, even if it existed?"**
   Deliberately invites the strongest possible objection, rather than only collecting positive feedback.

**What to do with the answer:**
- **Clear yes to Q3** → proceed with building this out further, roughly as scoped.
- **Clear no to Q3** → the mechanism doesn't need to change, but *how and when it reaches a family* does — revisit distribution before writing more code.
- **Mixed or uncertain** → still useful. It means the pre-crisis-onboarding question needs testing with a small real prototype in front of real people, not resolving on paper by reasoning about it further.

---

## Notes template

Copy this into a notes app, or print it, before the conversation:

```
Date:
Who was interviewed (role — e.g. ASHA worker, PLV, family member with direct experience):
Location / district:

Q1 response (has seen or been part of a wrongful-denial-adjacent situation):


Q2 response (what the family actually did):


Q3 response (would a family set this up in advance — THE key question):


Q4 response (who would a family actually trust to help):


Q5 response (what would stop a family from using this):


Overall signal (positive / negative / mixed):

If positive — proceed with the build as scoped.
If negative — revisit distribution strategy before further build time.
If mixed — identify what a lightweight prototype test could resolve that this conversation alone couldn't.
```

---

## After the conversation

Whatever the answer, it's worth updating `docs/OPEN_QUESTIONS.md` with what was learned — not to close the question with a checkmark, but so the next person (including a future instance of whoever's reading this) doesn't have to wonder whether this ever happened, or re-derive the same five questions from scratch.
