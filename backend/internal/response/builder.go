package response

import (
	"time"

	"github.com/pmjay-advocate/backend/internal/tiering"
)

// Build is the ONLY way to construct a FamilyResponse. Read this
// function top to bottom once: the care-first message and the
// not-legal-advice disclaimer are both set on the very first lines,
// unconditionally, before the outcome switch even begins. Every return
// path below shares that same struct literal's already-set fields —
// there is no branch, including future ones a later engineer might add,
// that skips past the first lines to return early. This is what
// Appendix S's checklist item ("no flow ever instructs a family to
// refuse or delay treatment") and Appendix AA's principle ("keep the
// care-first framing in the system's core instructions... enforced at
// the level the model cannot route around") actually mean in code, not
// just in a document — and the same mechanism is why the disclaimer
// can't be silently dropped from some future outcome either.
func Build(d tiering.Decision) FamilyResponse {
	r := FamilyResponse{
		careFirstMessage: CareFirstText,
		disclaimer:       DisclaimerText,
		outcome:          d.Outcome,
	}

	switch d.Outcome {
	case tiering.OutcomeGreen:
		if d.MatchedPackage != nil && d.MatchedPackage.PackageCode == "UNSPECIFIED" {
			r.tierMessage = unspecifiedGreenMessage(d)
			r.citation = "PMJAY Unspecified Procedure discretionary category"
		} else {
			r.tierMessage = greenMessage(d)
			r.citation = d.MatchedPackage.PackageName
		}
		r.actionSteps = []string{
			hospitalScript(d),
			"If they insist on payment before proceeding and treatment is needed soon, you can pay and dispute the charge afterward — do not let this delay care if it's urgent.",
			"Note down which staff member told you this, and the approximate time.",
		}
		r.hospitalScript = hospitalScript(d)
		r.complaintText = complaintText(d, time.Now())
		r.evidencePrompt = evidencePrompt(d)

	case tiering.OutcomeAmber:
		r.tierMessage = amberMessage(d)
		if d.MatchedExclusion != nil {
			r.citation = d.MatchedExclusion.DisplayName
		} else if d.MatchedPackage != nil {
			r.citation = d.MatchedPackage.PackageName
		} else {
			r.citation = "No specific package matched with enough confidence to cite"
		}
		r.actionSteps = []string{amberNextStep(d)}
		r.evidencePrompt = evidencePrompt(d)
		// No complaint text at amber: Section 9, the system never renders
		// its own final verdict on a genuinely disputed case while it is
		// still genuinely disputed.

	case tiering.OutcomeRed:
		r.tierMessage = redMessage(d)
		r.citation = d.MatchedExclusion.DisplayName
		// No action steps, no complaint, no evidence prompt: Section 7
		// Step 5, "it does not manufacture a grievance where none exists."

	case tiering.OutcomeMixed:
		r.tierMessage = mixedMessage(d)
		r.citation = d.MatchedPackage.PackageName + " / " + d.MatchedExclusion.DisplayName
		r.actionSteps = []string{
			"Ask the billing desk to split the bill into the covered part and the excluded part explicitly, in writing.",
			hospitalScript(d),
		}
		r.hospitalScript = hospitalScript(d)
		r.complaintText = complaintText(d, time.Now())
		r.evidencePrompt = evidencePrompt(d)

	case tiering.OutcomeHandoff:
		r.tierMessage = handoffMessage(d)
		r.handoffSummary = handoffSummary(d)
		// Deliberately no citation: a handoff makes no coverage claim of
		// its own to cite (Section 12) — the Para Legal Volunteer works
		// that out with the family directly.

	default:
		// Exhaustiveness guard: if a new Outcome value is ever added to
		// internal/tiering without updating this switch, fail safely
		// toward Handoff rather than toward silence or a guess. See
		// builder_test.go's TestBuild_UnknownOutcomeFailsSafeToHandoff.
		r.outcome = tiering.OutcomeHandoff
		r.tierMessage = "This situation needs a closer look than I can safely give it on my own."
		r.handoffSummary = handoffSummary(d)
	}

	return r
}

func amberNextStep(d tiering.Decision) string {
	for _, reason := range d.AmberReasons {
		if reason == tiering.ReasonPendingPreauth {
			return "Ask the billing desk this exact question: \"Has the pre-authorisation actually been submitted for this case, and what is its current status — approved, pending, or denied?\""
		}
	}
	return "Ask the hospital directly for the exact procedure name or code they are billing this under, in writing if possible — that's the single most useful thing to get next."
}
