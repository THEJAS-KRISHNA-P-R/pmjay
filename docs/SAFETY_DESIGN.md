# Safety Design

This system sits between a family and a hospital billing desk at the exact moment they're being told to pay or wait. A wrong or badly-worded answer here doesn't just cost accuracy points — it can talk a family into delaying treatment, or refusing to pay for something they genuinely owe. Everything in this document exists because of that, not as a compliance exercise.

Three things are enforced below: the care-first message can never be silently omitted, an LLM's confidence never gets asserted as certainty, and a pending pre-authorisation never gets treated as a denial. Each is enforced in code, not just in a prompt, and each has a test that tries to break it rather than one that confirms it works.

## 1. The care-first guarantee is structural, not procedural

The single most important line in this codebase is the first one `Build()` executes, in `backend/internal/response/builder.go`:

```go
func Build(d tiering.Decision) FamilyResponse {
	r := FamilyResponse{careFirstMessage: CareFirstText}
	// ... tier-specific branching happens after this line, never before it
```

`CareFirstText` is set before the function looks at `d.Outcome` at all. It is not one branch's responsibility to remember to include it — every branch already has it, because `r` was constructed with it. There is no code path through `Build()` that skips this assignment, including the `default` case of the outcome switch (see §4 below).

The second half of the guarantee is in `backend/internal/response/types.go`: every field on `FamilyResponse`, including `careFirstMessage`, is unexported. Code outside the `response` package cannot construct a `FamilyResponse` by hand and cannot set or clear `careFirstMessage` directly — the only way to get an instance is through `Build()`, and the only way to read the message back out is through the `CareFirstMessage()` accessor. This turns "the API handler forgot to attach the safety message" from a bug class that unit tests have to keep catching into a bug class the type system doesn't allow to exist. A future contributor adding a sixth outcome tier cannot accidentally ship a response without this message — the struct literal they'd need to bypass it with isn't legal Go from outside the package.

This directly implements Appendix S's pre-submission checklist item on the care-first message and Appendix T's definition of done, which specifically calls for "a dedicated adversarial test" that fails to produce a delay-encouraging response, not merely a manual read-through.

## 2. Testing the guarantee: exhaustive presence, then adversarial absence

Two different tests carry this, deliberately asking two different questions.

**`TestBuild_CareFirstMessageIsAlwaysPresent_EveryOutcome`** asks *is it always there?* It constructs one `Decision` per outcome — Green, Amber (both the low-confidence path and the partial-exclusion path, since they run through different branches of `amberMessage`), Red, Mixed, Handoff — and asserts `CareFirstMessage()` equals `CareFirstText` exactly, on all six. This is a presence check across every known branch.

**`TestBuild_H4_NoDelayEncouragingLanguageAcrossAllOutcomes`** asks a harder question: *even granting the message is present, has anything else in the response quietly undercut it?* It builds seven adversarial decisions, each with an `ExtractedSituationSummary` designed to tempt a template into hedging ("already on the operating table, payment dispute ongoing"; "transplant scheduled for tomorrow morning"), then scans all seven generated text fields on the response — `CareFirstMessage()`, `TierMessage()`, `Citation()`, `ComplaintText()`, `HospitalScript()`, `HandoffSummary()`, `EvidencePrompt()`, plus every string in `ActionSteps()` — against a 16-phrase blocklist (`"do not proceed"`, `"wait until"`, `"delay treatment"`, `"cancel the surgery"`, and so on). Per the spec's own kill condition for hypothesis H4, one unnegated match anywhere is a hard test failure, not a statistic to track.

The blocklist check is negation-aware on purpose. A naive substring match on `"delay care"` would flag the sentence *"do not let this delay care if it's urgent"* — which is the opposite of a violation, it's the safety instruction itself. `findUnnegatedPhrase()` checks the 40 characters immediately before a match for a negation cue (`"not "`, `"n't "`, `"never "`, `"avoid "`, `"without "`, `"against "`); only an unnegated match counts as a failure. This logic is safety-critical test infrastructure in its own right, so it has its own direct unit test, `TestFindUnnegatedPhrase_DirectlyTestsTheNegationLogicItself`, independent of the sweep that uses it.

Worth being direct about, per this project's own working style: **the first draft of this sweep had a false positive** — an early version of a template's wording tripped the naive blocklist despite being a safety instruction rather than a violation, which is exactly what motivated adding negation-awareness rather than just special-casing the one sentence. That's the rigor bar the rest of this document is trying to describe, not just assert.

## 3. Confidence is never rounded up to certainty

`internal/tiering/tier.go` enforces a confidence-gap rule (Appendix Z): if the top two candidate packages returned by extraction are within a configured margin of each other, the case is forced to Amber regardless of how high the top score is on its own. A 91%-confidence top match sitting 2 points above the second-best candidate is not treated the same as a 91%-confidence match sitting 40 points clear — the first is a real disambiguation failure wearing a high score, and Amber is the honest tier for it. `TestDecide_AmbiguousArm_CloseCandidatesForceAmberRegardlessOfTopScore` and `TestDecide_AmbiguousArm_WideGapStillGreen` pin both sides of this: the gap forces Amber when it's narrow, and does *not* force Amber just because a gap exists at all.

The same file also enforces that low confidence is never asserted as either a clear match *or* a clear exclusion — `TestDecide_AmbiguousArm_LowConfidenceNeverAssertedAsGreenOrRed` specifically checks that a low-confidence exclusion match routes to Amber, not to Red, because per the spec a confidently-wrong Red (telling a family they're excluded when they're not) is a worse failure mode than an over-cautious Amber.

Unverified data gets the same treatment at the response layer. `packageCitation()` in `internal/response/templates.go` only states a specific rupee figure when the underlying package record has `Verified: true` — a seed/placeholder package (see `docs/DATA_SOURCES.md`) is cited by name and specialty only, never with an invented number attached. `TestBuild_UnverifiedRateNeverStatedAsFact` enforces this directly; `TestBuild_VerifiedRateIsStated` confirms the opposite doesn't accidentally get suppressed too.

## 4. Unknown outcomes fail safe to a human, with the guarantee still intact

`TestBuild_UnknownOutcomeFailsSafeToHandoffWithCareFirstIntact` simulates an `Outcome` value that doesn't exist yet — the scenario a future contributor creates the moment they add a sixth tier to `tiering` without updating `response`'s switch statement. The `default` case of that switch routes to `OutcomeHandoff` rather than panicking, returning a zero-value citation, or silently defaulting to Green. Combined with §1, this means the single riskiest kind of future bug in this codebase — an unhandled case falling through — degrades to "a human reviews this case," not to "a family receives an unsafe or blank answer."

## 5. Pending is not denied — the single highest-value distinction this system makes

Section 6.8 and Section 8's third point both name this as the thing that separates a real reasoning system from a lookup table: a genuinely pending pre-authorisation is easily and honestly mistaken for a flat denial, by a stressed family and sometimes by imprecisely worded hospital staff. Getting this wrong in either direction is costly — telling a family a pending case is denied invites them to escalate a normal administrative wait as if it were a rejection; telling them a real denial is just "pending" leaves them waiting on something that was never coming.

This is handled by two independent signals, combined conservatively rather than trusted individually:

1. **The LLM's own read**, returned as a `Pending` field (`internal/extract/types.go`) on the extraction result — one of `SignalYes`, `SignalNo`, or `SignalNotApplicable`.
2. **A deterministic, model-free keyword-pattern detector**, `internal/tiering/preauth_pattern.go`, which scans the family's own description for phrasing that indicates a pending state independent of what the LLM concluded.

`tiering.Decide()` combines these conservatively: either signal alone pointing to "pending" is enough to route the case to Amber with `ReasonPendingPreauth`, rather than requiring both to agree. This is a deliberate recall-biased choice mirroring `internal/retrieval`'s own "fail open" design — the cost of over-flagging a case as Amber-pending (a family gets an honest "this needs a specific question answered" instead of a green light) is much lower than the cost of missing a real pending case and asserting a denial that hasn't actually happened. `TestDecide_AmbiguousArm_PendingPreauthMirrorsWorkedExample` pins this against the spec's own worked example so the behavior stays anchored to a concrete, named scenario rather than an abstract rule.

## 6. Exclusion nuance: state-dependent exclusions don't get flattened to Red

Not every exclusion is a clean, unconditional Red. Organ transplant coverage is state-dependent under PMJAY — sometimes covered, sometimes not, depending on specifics the system cannot always resolve from a short description. `tier.go` routes this case to Amber (`ReasonPartialExclusion`) rather than Red, and `TestDecide_MechanicalArm_OrganTransplantIsAmberNotRed` locks this in. Collapsing a genuinely conditional exclusion into a flat Red would be exactly the "confidently wrong" failure mode §3 is written to avoid, just approached from the exclusion side instead of the package-match side.

## 7. What "safety-critical" scope does and doesn't cover here

This document covers correctness and honesty of the system's own output — it does not resolve whether the system reaches the right family at the right moment, whether a family trusts an app-based tool at a hospital billing desk under stress, or whether the underlying HBP data is complete enough to be relied on unsupervised. Those are open, unresolved questions, tracked honestly in `docs/OPEN_QUESTIONS.md` rather than folded into this document to make the safety story look more finished than it is. Automated coverage of what's described above is mapped test-by-test in `docs/TESTING.md`, including which of the spec's own validation hypotheses (Appendix R) are and are not code-testable at all.
