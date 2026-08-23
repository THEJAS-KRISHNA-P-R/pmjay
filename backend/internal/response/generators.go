package response

import (
	"fmt"
	"time"

	"github.com/pmjay-advocate/backend/internal/tiering"
)

// hospitalScript is the exact words suggested for the billing desk
// (spec Section 7, Step 5; Appendix T's definition of done ties this
// directly to worked example 29.1's tone). Fixed template text with the
// specific situation interpolated — not freely generated per Appendix Z,
// so the safety-critical framing can never drift case to case.
func hospitalScript(d tiering.Decision) string {
	return "Ask the billing desk, calmly: \"Can you please give this denial to us in writing, with the reason stated?\" Hospitals are required to be able to justify a denial, and a verbal refusal that won't be put in writing is itself worth noting."
}

// complaintText generates the CGRMS-ready complaint text (spec Section
// 7, Step 5 and Section 15.3). Deliberately labelled as a draft for the
// family to review and submit themselves through the official app — this
// system does not and cannot submit on their behalf (Section 14.2, "no
// public API for direct submission was found").
func complaintText(d tiering.Decision, generatedAt time.Time) string {
	var subject, body string
	switch d.Outcome {
	case tiering.OutcomeGreen:
		pkg := *d.MatchedPackage
		subject = fmt.Sprintf("Denial of covered service: %s", pkg.PackageName)
		body = fmt.Sprintf(
			"I am writing to report that I was denied a covered service, or asked for payment for a covered service, under my PMJAY card.\n\n"+
				"Situation described: %s\n\n"+
				"Matched PMJAY package: %s (%s)\n\n"+
				"I am requesting review of this denial and, if confirmed incorrect, appropriate action against the hospital involved.",
			d.ExtractedSituationSummary, pkg.PackageName, pkg.Specialty,
		)
	case tiering.OutcomeMixed:
		pkg := *d.MatchedPackage
		subject = fmt.Sprintf("Denial of covered portion of a mixed bill: %s", pkg.PackageName)
		body = fmt.Sprintf(
			"I am writing to report a billing dispute where part of my treatment should be covered under my PMJAY card.\n\n"+
				"Situation described: %s\n\n"+
				"Covered portion (should not be charged): %s (%s)\n\n"+
				"I understand a separate, genuinely non-covered portion of this bill (%s) is correctly excluded and is not part of this complaint.\n\n"+
				"I am requesting review of the covered portion specifically.",
			d.ExtractedSituationSummary, pkg.PackageName, pkg.Specialty, d.MatchedExclusion.DisplayName,
		)
	default:
		// Amber and Red do not generate a complaint: Amber because
		// nothing has been confirmed wrong yet (Section 9: "the system
		// never renders its own final verdict"), Red because none is
		// warranted (Section 7, Step 5: "it does not manufacture a
		// grievance where none exists").
		return ""
	}

	return fmt.Sprintf(
		"--- Draft complaint for CGRMS (submit via the Ayushman App) ---\n"+
			"Date prepared: %s\n"+
			"Subject: %s\n\n"+
			"%s\n\n"+
			"[Add: your PMJAY card number, the hospital name, the staff member's name if known, and the approximate time of the incident, from the evidence you noted below, before submitting.]\n"+
			"--- End of draft — review before submitting ---",
		generatedAt.Format("2 January 2006"), subject, body,
	)
}

// evidencePrompt is Section 7 Step 6: capture evidence while it's still
// capturable. Shown for every outcome where a dispute genuinely might
// exist — Green, Amber, and Mixed — never for Red (nothing to build a
// case toward) or Handoff (the human takes it from here).
func evidencePrompt(d tiering.Decision) string {
	switch d.Outcome {
	case tiering.OutcomeGreen, tiering.OutcomeAmber, tiering.OutcomeMixed:
		return "Before you leave this conversation, note down three things while they're still easy to get: (1) the name of the staff member you spoke to, (2) the approximate time, and (3) get the denial in writing if you can — a photo of a written note is enough. This is easy to lose track of later, especially before a shift change."
	default:
		return ""
	}
}

// handoffSummary builds the full context passed to a NALSA Para Legal
// Volunteer, so — per Section 12 — nothing has to be re-explained from
// scratch by a family who may already be exhausted from explaining it
// once.
func handoffSummary(d tiering.Decision) string {
	reasonText := "This case involves genuine ambiguity or complexity a guided flow can't safely resolve alone."
	for _, r := range d.HandoffReasons {
		if r == tiering.HandoffMultipleIssues {
			reasonText = "This case bundles more than one separate issue at once."
		}
	}
	return fmt.Sprintf(
		"--- Handoff summary for Para Legal Volunteer ---\n\n"+
			"Reason for handoff: %s\n\n"+
			"What the family described, in their own words (summarised): %s\n\n"+
			"This family has not yet been asked to repeat any of this — please use this summary as your starting point rather than asking them to start over.\n"+
			"--- End of handoff summary ---",
		reasonText, d.ExtractedSituationSummary,
	)
}
