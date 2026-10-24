# Triage rules

These rules are a deterministic translation of the decision layer used by the
open-source PMJAY Advocate implementation. Use them after scoring candidates
and exclusions. Do not improvise an outcome that contradicts them.

## Inputs

1. `candidates`: package matches, each with `package_code` and
   `confidence_percent` (0–100), ranked most plausible first. A candidate
   from an unverified record has **no** `package_code`; do not supply one
   and do not lower your confidence because of the gap — the gap is the
   honest state, not missing data you should fill in.
2. `exclusion_matches`: every confirmed exclusion category, each with
   `category` and `confidence_percent`.
3. `pending_signal`: `pending_likely`, `denied_final_likely`, `unclear`, or
   `not_applicable`.
4. `multiple_distinct_issues_detected`: boolean.
5. `family_distress_signal`: boolean.

## Thresholds

| Constant | Value | Purpose |
|---|---:|---|
| `GREEN_CONFIDENCE` | 75 | Minimum top score for a confident package match. |
| `GAP_THRESHOLD` | 15 | If the top two package scores are closer than this, force Amber. |
| `EXCLUSION_CONFIDENCE` | 65 | Minimum score for a real exclusion match. |
| `UNSPECIFIED_CONFIDENCE` | 80 | Higher bar for the discretionary catch-all. It is never a default for "unsure." |

A high top score does not override the gap rule. 91% confidence with the
second candidate at 88% is a disambiguation failure, not a green light.

## Decision order

Apply these in order and stop at the first match, except where a rule says to
continue combining.

### 1. Multiple distinct issues → Handoff

If the description bundles more than one separable problem — for example a
billing dispute, a staff-conduct complaint, and an Aadhaar/ID mismatch — route
to Handoff. Do not try to resolve all three in one answer.

### 2. Confident exclusion with no confident package → Red

If the top exclusion is at least 65 and the package side is not Green or Mixed,
the denial may be correct. Check the nuance rules below before calling it Red.

- The exclusion must exist in `references/exclusions.json`. Never cite an
  exclusion category you cannot trace to the dataset.
- If the exclusion has a `nuance` field, follow the next rule.

### 3. Partial / state-dependent exclusion → Amber

`organ_transplant_partial` is never a flat Red. Coverage depends on the state
and the specific organ/procedure. Route to Amber and recommend confirmation
with 14555 or a Para Legal Volunteer.

### 4. Confident package + confident exclusion → Mixed

If the package side is Green **and** a top exclusion is at least 65, report
both. The bill has two parts and must be split. Do not silently drop either
side.

### 5. Close top-two package candidates → Amber

If the top two package scores are within 15 points of each other, force Amber
regardless of the top score. Ask for the exact procedure name from the doctor
or discharge papers.

### 6. Unspecified procedure handling

`UNSPECIFIED` is PMJAY's discretionary category for a real procedure that does
not map to a named package.

- Score at least 80 → Green, with the discretionary explanation and the
  scheme's per-case cap.
- Score below 80 → Amber, unspecified-unclear. Do not present it as coverage.

Never use `UNSPECIFIED` as a default when you are simply unsure. "Unsure about
a named package" is low confidence, which is a different and more honest
answer.

### 7. Low package confidence → Amber

Top package below 75 → Amber. Say what the closest match is, if any, and say
clearly that you are not confident enough to answer either way.

### 8. Pending pre-authorisation → Amber

If the matched package requires pre-authorisation and either the extraction
signal or the family's own wording points to pending, route to Amber with the
pending-preauth script. See `SKILL.md` for the exact desk question.

Either signal pointing to pending is enough. This is intentionally
recall-biased: a false "pending" costs one extra question; a false "denied"
can send a family into the wrong fight.

### 9. Otherwise → Green

A top package at least 75, a gap of at least 15, no qualifying exclusion, and
no pending signal is Green. Cite the package and the rate only if the record is
in `hbp-packages.json` (verified). If the best record is in the unverified
file, stay Amber.

### 10. Distress interaction

A distress signal does not, by itself, force Handoff. A frightened family with
a clear case should get the clear answer immediately. Distress escalates an
*already unclear* case to Handoff instead of Amber.

## Amber reasons, and what to ask

Each Amber has a specific next question. Do not give a generic "contact the
helpline" answer when a targeted question is available.

| Reason | What to say |
|---|---|
| `low_confidence` | "The closest match I can find is X, but I am not confident enough to say it is definitely this. What is the exact procedure name the doctor wrote?" |
| `close_candidates` | "There are two packages this could genuinely be. Which one applies changes the answer. Do you have the discharge papers or the doctor's procedure name?" |
| `pending_preauth` | "Has the pre-authorisation been submitted, and is it approved, pending, or denied?" |
| `partial_exclusion` | "This depends on your state and the exact procedure. It is not a blanket yes or no. Call 14555 or ask a Para Legal Volunteer to check your state's rule." |
| `unspecified_unclear` | "I cannot find a named package that fits well enough to say. Ask the hospital what procedure name or code they are billing this under." |

## Handoff reasons

| Reason | What the handoff summary must contain |
|---|---|
| `multiple_distinct_issues` | Each separate issue, in the family's own terms, without merging them. |
| `distress_with_unresolved_ambiguity` | The unresolved question, what is already known, and what the family is most worried about. |

A handoff is not a failure state. It is the correct answer when the case is
genuinely beyond a guided flow.

## Red: when the denial is correct

Red is a real outcome, not a lesser feature. Say it plainly:

- Explain the category in plain language.
- Do not imply the hospital is acting in bad faith.
- Do not encourage a grievance that will fail.
- Mention the genuine alternatives: paying, a state scheme, a payment plan, or
  asking the doctor whether a functional medical reason changes the category.
- If the family believes the procedure is medically necessary for a functional
  reason, tell them to ask the treating doctor to document that reason.

## What never changes the decision

- How upset the family is (except as described in rule 10).
- How urgent the treatment is. Urgency changes the *care* advice, not the
  coverage confidence.
- The hospital's tone.
- A guess about the hospital's motive.
- The family's ability to pay.
- Whether a rate "sounds about right."

## One-sentence summary

Green when a verified package is clearly covered; Amber when the honest answer
is "we need one more fact," including every pending pre-auth; Red when the
exclusion is correct; Mixed when the bill has two parts; Handoff when a human
should take over.
