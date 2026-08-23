# `internal/tiering`

The decision layer. Takes one `extract.Result` (the LLM's structured read on a family's situation) and turns it into exactly one `Outcome` — green, amber, red, mixed, or handoff — with zero further model calls. This is the package the spec (Section 58) describes as "close to a decision tree": every threshold is a named constant, every branch is a plain `if`, and the whole thing runs in microseconds against a hand-constructed test case bank rather than needing a live API key. 94.5% test coverage, the highest of any package in this backend, which is deliberate — this is the layer that decides what a family is told, so it's the layer with the least tolerance for an untested branch.

## Files

| File | What it does |
|---|---|
| `tier.go` | `Outcome`, `Decision`, all the tuning constants, and `Decide()` — the actual decision function. |
| `preauth_pattern.go` | `Detect()` — an independent, deterministic, keyword-based second opinion on the pending-vs-denied question, used to cross-check the LLM's own read rather than trust it alone. |

## The five outcomes, and why there are five, not three

The spec's own tiered table (Section 9) names three: green (confidently covered), amber (genuine ambiguity, ask a clarifying question), red (confirmed exclusion). Two more exist because the spec's own worked examples describe situations neither of those three cleanly fits:

- **Mixed** (Section 46.3) — part of what a family described is a covered package, part is a confirmed exclusion, in the same bill (the canonical example: a knee replacement bundled with an unrelated cosmetic add-on). Reporting this as a single yes/no would be wrong regardless of which way it resolved — the honest answer is "yes to this part, no to that part," so `Decision` carries both a `MatchedPackage` and a `MatchedExclusion` at once for this outcome.
- **Handoff** (Section 12) — genuine ambiguity, or a situation description that bundles more than one distinct problem, that a guided flow shouldn't try to resolve on its own. Routes to a human (a NALSA Para Legal Volunteer, via the free 15100 helpline) with a prepared summary, not an API call this system doesn't have access to anyway (see `../../../ARCHITECTURE.md`'s "what was deliberately not built").

## The tuning constants, and the reasoning behind each threshold

All grouped in one block in `tier.go`, specifically so the whole safety-relevant threshold surface is visible in one place rather than scattered through branch logic:

- **`GreenConfidenceThreshold = 75`** — a candidate below this can never be asserted as a confident match, full stop, regardless of how it compares to other candidates.
- **`GapThreshold = 15`** — per the spec's Appendix Z: if the top two candidates' confidence scores are within 15 points of each other, that *alone* forces amber, even if the top score clears the green threshold on its own. A confident-looking top score next to an almost-as-confident second option is exactly the "genuinely close call" this system should surface as a question, not paper over with false certainty.
- **`ExclusionConfidenceThreshold = 65`** — lower than the green package threshold, deliberately: the cost of treating a real exclusion signal as noise (missing a legitimate denial reason) and the cost of treating noise as a real exclusion signal (routing to amber's "partial exclusion" nuance path, or red, when it shouldn't) are not symmetric with the package-match case, and this threshold reflects that.
- **`UnspecifiedConfidenceThreshold = 80`** — deliberately *higher* than the ordinary green threshold. The `UNSPECIFIED` catch-all package exists for genuinely unclassifiable-but-real procedures (Section 8) — it must never become the default answer for "the model wasn't sure," which is a categorically different, and honest, situation (that's low confidence on a named package, i.e. amber, not a confident `UNSPECIFIED` match).
- **`DistressAmbiguityCeiling`** — see the distress-and-handoff interaction below.

**If you're tuning any of these**, the tests in `tier_test.go` (not separately described here — it's exactly the three-arm structure the name suggests: green/amber/red cases hand-constructed from the spec's own worked examples) are what actually validates a threshold change didn't silently flip a previously-correct case. Run them before and after, not just after.

## Two decisions worth understanding deeply, not just reading past

### Why distress doesn't force handoff by itself

`Decide`'s own doc comment states this directly: a family describing a genuinely clear, unambiguous situation — including the spec's own Section 29.1 worked example, itself a family in real distress with a sick relative and money being demanded — still gets a fast, clear answer, not an extra hop to a human. `FamilyDistressSignal` alone never forces `Handoff`. What it *does* is lower the bar for other, already-present ambiguity to resolve toward `Handoff` instead of `Amber` — the reasoning being that a family who both sounds distressed *and* has a genuinely unresolved case is the Section 12 scenario ("needs more support than a guided flow can provide"), while a distressed family with a clear case is best served by clarity delivered quickly, not by being routed away from a direct answer just because they sound upset. Getting this backwards — routing every distressed-sounding description to a human regardless of clarity — would be a worse product for the exact families it's trying hardest to help.

### Why `preauth_pattern.go` exists at all, given the LLM already reports a `PendingSignal`

Because Section 10's absolute care-first rule and the spec's own H3 safety-failure kill condition ("confidently wrong on the ambiguous arm") both argue against letting a single, unverifiable model judgement be the only thing standing between a family and an overclaimed answer on exactly the distinction the spec calls "the single highest-value thing the system does" — whether a claim is still pending (so pushing back too hard could actually hurt the family) or already, finally denied (so pushing back is exactly right). `Detect()` is a second, independent, fully-deterministic read on the same question, using plain keyword matching against the family's *original* text — not a review of what the LLM already concluded, a fresh scan. `decidePackage` only treats a case as a confident, unqualified green when **neither** signal suggests pending *and* the LLM's own signal isn't `unclear` either — any one of the three (LLM says pending, pattern says pending, LLM says unclear) is enough to force amber instead. `Detect`'s own doc comment covers one more specific case worth knowing: if the text contains cues for *both* pending and denied language at once ("they denied it but also said it's still pending review"), `Detect` deliberately resolves to `PatternPendingLikely`, not by picking whichever cue matched first — per the same conservative-bias principle, treating a mixed signal as "possibly still pending" is the safer wrong guess than treating it as a confident final denial.

### Why an unrecognized exclusion category fails toward the package-side outcome, not toward red

If the LLM's `ExclusionMatches` names a category that isn't actually in `ds.Exclusions` (a model referencing something outside the real, confirmed list), `Decide` does not treat that as a red flag to fall back on cautiously — it discards the unverifiable exclusion claim entirely and returns whatever the package-side logic alone concluded. The reasoning is the same explainability principle response citation relies on everywhere else in this codebase (see `internal/response/README.md`): never cite what can't be traced back to a real record. An exclusion this system can't itself verify exists is not evidence of anything, in either direction.

## If you're extending this package

- **Adding a sixth outcome**: `internal/response`'s `packageCitation`/message-generation logic and the frontend's `Outcome` union (`frontend/lib/types.ts`, `TierBadge.tsx`'s `TIER_STYLES`) both need the matching case, or a new outcome will compile here and then render as a blank or default state two layers away. Grep for `OutcomeHandoff` as a template for everywhere a new `Outcome` value needs to be threaded through.
- **Changing a threshold**: see "tuning constants" above — update the constant's own comment with your reasoning, the way every existing one does; a bare number with no comment is a regression waiting to happen the next time someone tunes it without the original context.
- **This file is part of the codebase's documentation convention** — see the repo root `README.md`. Keep it in sync with the code, in the same change.
