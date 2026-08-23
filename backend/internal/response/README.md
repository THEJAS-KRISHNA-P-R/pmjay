# `internal/response`

Turns one `tiering.Decision` into the actual words a family reads. This is the last backend layer before the frontend — everything here is plain Go string templating with values interpolated, never freely LLM-generated, so the exact framing a family sees on a given outcome can never drift from one case to the next in a way nobody reviewed.

## Files

| File | What it does |
|---|---|
| `types.go` | `FamilyResponse` — the output type — and why every field on it is unexported. |
| `builder.go` | `Build()`, the *only* way to construct a `FamilyResponse`, and the switch over `tiering.Outcome` that decides what gets populated for each tier. |
| `generators.go` | The per-field generator functions `Build` calls: `hospitalScript`, `complaintText`, `evidencePrompt`, `handoffSummary`. |
| `templates.go` | `CareFirstText` (the one message that's always present) and every tier's message-generation function, plus `packageCitation` — the single place a rupee figure is ever rendered to a family. |

## The most important thing about this package: how the care-first rule is enforced

Read `types.go`'s package-level doc comment and `builder.go`'s doc comment on `Build` before changing anything here — they say this directly, but it's worth restating because it's the one property in this whole codebase that most depends on getting the *mechanism* right, not just the intent:

**Every field of `FamilyResponse` is unexported, and `Build` is the only function anywhere that can construct one.** `Build` sets `careFirstMessage: CareFirstText` on literally the first line, before the `switch d.Outcome` even begins — every one of the five outcome branches, and the `default` exhaustiveness-guard branch, shares that same already-initialized struct literal. There is no code path, including ones a future engineer might add without reading this file carefully, that can return a `FamilyResponse` with an empty or missing care-first message. This isn't a convention enforced by a linter or a code-review checklist — it's enforced by the type system: there is no exported constructor, no exported struct literal syntax available outside this package, and no field a caller from another package could see or zero out. A reviewer doesn't have to trust that every call site remembered the safety rule; the compiler won't allow the alternative to compile. `builder_test.go`'s `TestBuild_CareFirstMessageIsAlwaysPresent_EveryOutcome` is the test that would catch a regression here, and the frontend has its own version of the same test at the page level — see `frontend/app/case/[id]/README.md`.

The `default` case in `Build`'s switch is the same principle applied to a different failure mode: if `internal/tiering` ever adds a sixth `Outcome` value without this switch being updated to handle it, the fallback is `OutcomeHandoff` with a generic "needs a closer look" message — not silence, not a zero-value response, not a guess. Failing toward a human is the safe direction to fail in; failing toward an empty or wrong answer is not.

## What each outcome actually gets, and why the differences are deliberate

| Outcome | Citation | Action steps | Complaint text | Evidence prompt |
|---|---|---|---|---|
| Green | the matched package | hospital script + "pay and dispute later if urgent" + note the staff member | yes | yes |
| Amber | package or exclusion, whichever was matched — or an honest "nothing matched with enough confidence" | one targeted next question | **no** | yes |
| Red | the matched exclusion | none | none | **no** |
| Mixed | both package and exclusion, joined | split-the-bill request + hospital script | yes, scoped to the covered portion only | yes |
| Handoff | none | none (handoff summary instead) | none | none |

Three gaps in that table are load-bearing, not oversights:

- **No complaint text at Amber.** Per Section 9 of the source spec: "the system never renders its own final verdict on a genuinely disputed case while it is still genuinely disputed." Generating a formal complaint before the ambiguity is even resolved would be asserting a conclusion the system hasn't actually reached.
- **No action steps, complaint, or evidence prompt at Red.** Per Section 7 Step 5: "it does not manufacture a grievance where none exists." A confirmed legitimate exclusion is not a wrongful denial, and treating it like one — even by offering the same scaffolding — would be actively misleading.
- **No citation at Handoff.** A handoff makes no coverage claim of its own (Section 12) — there's nothing to cite, because citing something here would imply a conclusion this system deliberately isn't reaching. The Para Legal Volunteer works out the actual answer with the family directly, using `handoffSummary`'s context as a starting point.

## `templates.go`: `packageCitation` is the one place a rupee figure reaches a family

This is the function most likely to matter if you're touching this package. Three render paths, by `Package.RateType`:

- **Unverified** (`Verified == false`): package name and specialty only, *never* a number. This is what makes the seed/placeholder dataset entries (see `internal/hbp/README.md`) safe to have in the dataset at all — a family is told "this is a listed PMJAY package," never handed an invented rate.
- **`"tiered"`**: a range (`₹X to ₹Y depending on the treating hospital's city-tier classification`), never a single picked number — because this tool has no way to know which tier a specific hospital falls under, and picking one would fabricate certainty that doesn't exist. Falls back to plain flat-rate rendering if `RateMaxINR` was somehow left unset (defensive — see `TestPackageCitation_TieredWithoutMaxFallsBackToFlatRendering`).
- **`"per_diem"`**: an explicit four-level "per day" sentence (Routine Ward / HDU / ICU without ventilator / ICU with ventilator), never phrased in a way that could read as a single total. `TestPackageCitation_PerDiemNeverImpliesATotal` in `templates_test.go` is the regression test for this specific property — read it before changing this branch, since the phrasing itself (not just the numbers) is what the test checks.

If you add a fourth `RateType` value in `internal/hbp`, this `switch p.RateType` needs the matching branch, or the new shape silently falls through to the plain flat-rate rendering — which is the *safe* direction to fail (no crash, no blank output), but is still probably not the rendering you actually want, so add the branch rather than relying on the fallback.

## `generators.go`: the smaller pieces `Build` assembles

- **`hospitalScript`** is fixed template text, not per-case generation, on purpose — the exact wording a family is told to say at a billing desk is safety-critical framing (calm, specific, puts the burden of justification on the hospital) that should never vary case to case in a way nobody reviewed.
- **`complaintText`** is explicitly labelled a *draft for the family to review and submit themselves* — this system cannot and does not submit a CGRMS complaint on anyone's behalf (no public API exists for that; see `../../../ARCHITECTURE.md`'s "what was deliberately not built"). The generated text includes an explicit `[Add: ...]` placeholder for details this system doesn't have (card number, hospital name), rather than guessing or leaving them out silently.
- **`handoffSummary`** exists so a family who may already be exhausted from explaining their situation once doesn't have to do it again with a Para Legal Volunteer — the full `ExtractedSituationSummary` is carried through, with an explicit note in the generated text saying as much.

## If you're extending this package

- **Adding a new outcome-specific field to `FamilyResponse`**: add the unexported field, the exported accessor method (following the existing pattern — a plain getter, `ActionSteps()`'s defensive copy via `append([]string(nil), ...)` if the field is a slice, so callers can't mutate internal state through the returned value), and populate it in every relevant branch of `Build`'s switch. Then thread it through `internal/api`'s DTO and the frontend's `lib/types.ts` and the component that renders it — see `internal/api/README.md` and `frontend/lib/README.md`.
- **This file is part of the codebase's documentation convention** — see the repo root `README.md`. Keep it in sync with the code, in the same change.
