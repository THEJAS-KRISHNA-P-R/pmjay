// Package tiering is the deterministic decision layer described in the
// spec (Section 58) as "close to a decision tree" — plain, auditable Go
// logic that turns an extraction result into one of the tiers in Section
// 9, with zero further model calls. Every threshold in this file is a
// named constant specifically so a reviewer can find, question, and tune
// each one without hunting through prose for a magic number.
package tiering

import (
	"sort"

	"github.com/pmjay-advocate/backend/internal/extract"
	"github.com/pmjay-advocate/backend/internal/hbp"
)

// Outcome is the final routing decision. Deliberately five values, not
// three — the spec's own tiered table (Section 9) names three tiers, but
// two more outcomes are necessary to represent cases the spec itself
// describes in worked examples: Mixed (Section 46.3's split bill) and
// Handoff (Section 12).
type Outcome string

const (
	OutcomeGreen   Outcome = "green"
	OutcomeAmber   Outcome = "amber"
	OutcomeRed     Outcome = "red"
	OutcomeMixed   Outcome = "mixed"   // part of the described situation is covered, part is a confirmed exclusion — Section 46.3
	OutcomeHandoff Outcome = "handoff" // Section 12: genuine ambiguity or a family needing more than a guided flow
)

// AmberReason explains *why* a case landed on amber — the response layer
// needs this to generate the right specific question (Section 9's table:
// the amber script is different for a close-candidate call than for a
// pending-preauth call), and it is exactly the kind of "why" a family or
// a judge might reasonably ask for for (Section 11, Explainability).
type AmberReason string

const (
	ReasonLowConfidence      AmberReason = "low_confidence"      // no candidate cleared the green bar
	ReasonCloseCandidates    AmberReason = "close_candidates"    // Appendix Z's top-two-close rule
	ReasonPendingPreauth     AmberReason = "pending_preauth"     // Section 6.8's pending-vs-denied distinction
	ReasonPartialExclusion   AmberReason = "partial_exclusion"   // e.g. organ transplant's state-dependent nuance
	ReasonUnspecifiedUnclear AmberReason = "unspecified_unclear" // Unspecified Procedure candidate present but not confident enough to assert
)

// HandoffReason explains why a case was routed to a human rather than any
// tier at all.
type HandoffReason string

const (
	HandoffMultipleIssues      HandoffReason = "multiple_distinct_issues"
	HandoffDistressWithUnclear HandoffReason = "distress_with_unresolved_ambiguity"
)

// Decision is the full output of Decide — everything internal/response
// needs to generate a cited, tier-appropriate answer without having to
// re-derive any of this reasoning itself.
type Decision struct {
	Outcome Outcome

	// Populated for Green, Amber (candidate-based reasons), and Mixed.
	MatchedPackage           *hbp.Package
	MatchedPackageConfidence int

	// Populated for Amber.
	AmberReasons []AmberReason

	// Populated for Red and Mixed.
	MatchedExclusion           *hbp.Exclusion
	MatchedExclusionConfidence int

	// Populated for Handoff.
	HandoffReasons []HandoffReason

	// Always populated: carried through from extraction for use in
	// responses and, if needed, the NALSA handoff summary (Section 12 —
	// "nothing has to be re-explained from scratch").
	ExtractedSituationSummary string

	// DetectedPattern is this decision's own deterministic pending/denied
	// read, exposed mainly so tests and logs can see why a pending
	// downgrade did or didn't fire, independent of what the LLM said.
	DetectedPattern PreauthPattern
}

// Tuning constants. Named and grouped here, not scattered through the
// decision logic, so the whole safety-relevant threshold surface is
// visible in one place.
const (
	// GreenConfidenceThreshold: a candidate below this can never be
	// asserted as a confident green match.
	GreenConfidenceThreshold = 75

	// GapThreshold: per Appendix Z, if the top two candidates' confidence
	// scores are within this many points of each other, that alone forces
	// amber — regardless of how high the top score looks in isolation.
	GapThreshold = 15

	// ExclusionConfidenceThreshold: the minimum confidence for treating an
	// exclusion match as real rather than noise.
	ExclusionConfidenceThreshold = 65

	// UnspecifiedConfidenceThreshold is deliberately higher than the
	// ordinary green threshold — Section 8 warns this category must never
	// be a default for "I'm not sure", only a genuine judgement that
	// nothing named fits.
	UnspecifiedConfidenceThreshold = 80

	// DistressAmbiguityCeiling: a distress signal only escalates an
	// already-not-clean case to Handoff; see Decide's doc comment for the
	// reasoning on why distress alone, on an otherwise clear case, does
	// not force a handoff.
	DistressAmbiguityCeiling = GreenConfidenceThreshold
)

// Decide turns one extraction Result into a routing Decision. ds is used
// to resolve package/exclusion codes back to full records for citation,
// and rawDescription is the family's own original text, re-scanned here
// independently of whatever the LLM already concluded about it (see
// preauth_pattern.go's doc comment for why that independence matters).
//
// On the distress-and-handoff interaction specifically: a family
// describing a genuinely clear, unambiguous situation — Section 29.1's
// worked example is itself a family in real distress, a sick relative,
// money being demanded — still gets a fast, clear, calm answer, not an
// extra hop to a human. Distress alone does not force Handoff. What it
// does is lower the bar for *other* ambiguity signals already present to
// resolve toward Handoff instead of Amber, on the reasoning that a
// family who both sounds distressed and has a genuinely unresolved case
// is exactly the Section 12 scenario ("needs more support than a guided
// flow can provide"), whereas a distressed family with a clear case is
// best served by clarity delivered quickly.
func Decide(ds *hbp.Dataset, rawDescription string, result extract.Result) Decision {
	pattern := Detect(rawDescription)

	if result.MultipleDistinctIssuesDetected {
		return Decision{
			Outcome:                   OutcomeHandoff,
			HandoffReasons:            []HandoffReason{HandoffMultipleIssues},
			ExtractedSituationSummary: result.ExtractedSituationSummary,
			DetectedPattern:           pattern,
		}
	}

	sortedCandidates := sortedCopy(result.Candidates)
	sortedExclusions := sortedExclusionCopy(result.ExclusionMatches)

	var topExclusion *extract.ExclusionMatch
	if len(sortedExclusions) > 0 && sortedExclusions[0].ConfidenceP >= ExclusionConfidenceThreshold {
		topExclusion = &sortedExclusions[0]
	}

	packageOutcome, packageDecision := decidePackage(ds, sortedCandidates, pattern, result.Pending)

	// No real exclusion signal: pure package-side outcome.
	if topExclusion == nil {
		if result.FamilyDistressSignal && packageOutcome != OutcomeGreen && packageOutcome != OutcomeMixed {
			packageDecision.Outcome = OutcomeHandoff
			packageDecision.HandoffReasons = []HandoffReason{HandoffDistressWithUnclear}
		}
		packageDecision.ExtractedSituationSummary = result.ExtractedSituationSummary
		packageDecision.DetectedPattern = pattern
		return packageDecision
	}

	excl, found := ds.ExclusionByCategory(topExclusion.Category)
	if !found {
		// Model referenced a category that isn't in our dataset. Fail
		// toward the package-side outcome rather than citing a category
		// we cannot ourselves verify exists — never cite what can't be
		// traced back to a real record (Explainability, Section 11).
		if result.FamilyDistressSignal && packageOutcome != OutcomeGreen && packageOutcome != OutcomeMixed {
			packageDecision.Outcome = OutcomeHandoff
			packageDecision.HandoffReasons = []HandoffReason{HandoffDistressWithUnclear}
		}
		packageDecision.ExtractedSituationSummary = result.ExtractedSituationSummary
		packageDecision.DetectedPattern = pattern
		return packageDecision
	}

	// Partial/nuanced exclusion (organ transplant): never a flat red,
	// always routes to amber with the specific nuance attached. See
	// Section 6.6 and hbp.Exclusion.Nuance's doc comment.
	if excl.Nuance != "" {
		return Decision{
			Outcome:                    OutcomeAmber,
			AmberReasons:               []AmberReason{ReasonPartialExclusion},
			MatchedExclusion:           &excl,
			MatchedExclusionConfidence: topExclusion.ConfidenceP,
			ExtractedSituationSummary:  result.ExtractedSituationSummary,
			DetectedPattern:            pattern,
		}
	}

	// A confident exclusion AND a confident, genuinely covered package at
	// once: the mixed case (Section 46.3 — knee replacement bundled with
	// a cosmetic add-on). Report both; never silently drop one for the
	// other, and never treat the whole bill as a single yes/no unit.
	if packageOutcome == OutcomeGreen {
		return Decision{
			Outcome:                    OutcomeMixed,
			MatchedPackage:             packageDecision.MatchedPackage,
			MatchedPackageConfidence:   packageDecision.MatchedPackageConfidence,
			MatchedExclusion:           &excl,
			MatchedExclusionConfidence: topExclusion.ConfidenceP,
			ExtractedSituationSummary:  result.ExtractedSituationSummary,
			DetectedPattern:            pattern,
		}
	}

	// Confident exclusion, no competing confident package match: a clean
	// red. This is deliberately not treated as a lesser-effort path —
	// Section 9: "not optional, and not a lesser feature."
	return Decision{
		Outcome:                    OutcomeRed,
		MatchedExclusion:           &excl,
		MatchedExclusionConfidence: topExclusion.ConfidenceP,
		ExtractedSituationSummary:  result.ExtractedSituationSummary,
		DetectedPattern:            pattern,
	}
}

// decidePackage handles the package-match side of the decision in
// isolation, before any exclusion interaction is layered on top. Returns
// the outcome (only ever Green, Amber, or Red is never produced here —
// "no real match at all" is represented as Amber/low-confidence, since an
// honest "I can't tell" is always available and is never a dead end) plus
// the Decision fields it's responsible for.
func decidePackage(ds *hbp.Dataset, sortedCandidates []extract.CandidateMatch, pattern PreauthPattern, llmPending extract.PendingSignal) (Outcome, Decision) {
	if len(sortedCandidates) == 0 {
		return OutcomeAmber, Decision{
			Outcome:      OutcomeAmber,
			AmberReasons: []AmberReason{ReasonLowConfidence},
		}
	}

	top := sortedCandidates[0]
	pkg, found := ds.PackageByCode(top.PackageCode)
	if !found {
		return OutcomeAmber, Decision{
			Outcome:      OutcomeAmber,
			AmberReasons: []AmberReason{ReasonLowConfidence},
		}
	}

	// Appendix Z's gap rule: close top-two candidates force amber
	// regardless of the top score's own magnitude.
	if len(sortedCandidates) >= 2 {
		gap := top.ConfidenceP - sortedCandidates[1].ConfidenceP
		if gap < GapThreshold {
			return OutcomeAmber, Decision{
				Outcome:                  OutcomeAmber,
				AmberReasons:             []AmberReason{ReasonCloseCandidates},
				MatchedPackage:           &pkg,
				MatchedPackageConfidence: top.ConfidenceP,
			}
		}
	}

	if pkg.PackageCode == "UNSPECIFIED" {
		if top.ConfidenceP >= UnspecifiedConfidenceThreshold {
			return OutcomeGreen, Decision{
				Outcome:                  OutcomeGreen,
				MatchedPackage:           &pkg,
				MatchedPackageConfidence: top.ConfidenceP,
			}
		}
		return OutcomeAmber, Decision{
			Outcome:                  OutcomeAmber,
			AmberReasons:             []AmberReason{ReasonUnspecifiedUnclear},
			MatchedPackage:           &pkg,
			MatchedPackageConfidence: top.ConfidenceP,
		}
	}

	if top.ConfidenceP < GreenConfidenceThreshold {
		return OutcomeAmber, Decision{
			Outcome:                  OutcomeAmber,
			AmberReasons:             []AmberReason{ReasonLowConfidence},
			MatchedPackage:           &pkg,
			MatchedPackageConfidence: top.ConfidenceP,
		}
	}

	// Confident single-candidate match. Last check: the pending-vs-denied
	// distinction (Section 6.8), only relevant if this package can even
	// require pre-authorisation.
	if pkg.RequiresPreauth {
		llmSaysPending := llmPending == extract.SignalPendingLikely
		llmUnclear := llmPending == extract.SignalUnclear
		patternSaysPending := pattern == PatternPendingLikely

		if llmSaysPending || patternSaysPending || llmUnclear {
			return OutcomeAmber, Decision{
				Outcome:                  OutcomeAmber,
				AmberReasons:             []AmberReason{ReasonPendingPreauth},
				MatchedPackage:           &pkg,
				MatchedPackageConfidence: top.ConfidenceP,
			}
		}
	}

	return OutcomeGreen, Decision{
		Outcome:                  OutcomeGreen,
		MatchedPackage:           &pkg,
		MatchedPackageConfidence: top.ConfidenceP,
	}
}

func sortedCopy(candidates []extract.CandidateMatch) []extract.CandidateMatch {
	out := make([]extract.CandidateMatch, len(candidates))
	copy(out, candidates)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ConfidenceP > out[j].ConfidenceP
	})
	return out
}

func sortedExclusionCopy(matches []extract.ExclusionMatch) []extract.ExclusionMatch {
	out := make([]extract.ExclusionMatch, len(matches))
	copy(out, matches)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ConfidenceP > out[j].ConfidenceP
	})
	return out
}
