package response

import (
	"fmt"
	"strings"

	"github.com/pmjay-advocate/backend/internal/hbp"
	"github.com/pmjay-advocate/backend/internal/tiering"
)

// CareFirstText is Section 10's non-negotiable rule, verbatim in intent,
// stated first, every time, in every flow: "Get treatment first. Dispute
// the money after. Always." Nothing in this file may alter, shorten, or
// conditionally suppress this text — see builder.go for how that's
// enforced structurally rather than just by convention.
const CareFirstText = "Get treatment first. Dispute the money after. Always. If you can pay now and settle the dispute later, or move to a different hospital, do that: do not let this disagreement delay or stop care."

// DisclaimerText is shown alongside every response, the same
// unconditional way CareFirstText is (see builder.go). This system
// cites specific package codes and rupee figures and drafts complaint
// text ready to submit — confident enough in form that it must be
// equally clear about what it is not: a legal determination, and not a
// substitute for confirming the current number with the hospital or
// helpline before relying on it. "Verified" (hbp.Package.Verified)
// means checked against the real government source at the time it was
// added to this dataset, not "guaranteed unchanged since."
const DisclaimerText = "This is guidance based on official PMJAY rules, not a legal ruling. Confirm the current rate with the hospital or the PMJAY helpline before relying on any figure here, and remember free legal help from a NALSA Para Legal Volunteer is available if the hospital disagrees."

// packageCitation renders a package's name/specialty as a citation,
// including the specific indicative rate ONLY when the record has
// actually been verified against the real government source. An
// unverified placeholder number is never stated to a family as if it
// were a checked fact — see hbp.Package.Verified's doc comment.
func packageCitation(p hbp.Package) string {
	if !p.Verified {
		return fmt.Sprintf("%s (%s), a listed PMJAY package", p.PackageName, p.Specialty)
	}
	switch p.RateType {
	case "per_diem":
		if pd := p.PerDiemRates; pd != nil {
			return fmt.Sprintf(
				"%s (%s), reimbursed per day of admission rather than as one total (₹%s/day in a general ward, ₹%s/day in HDU, ₹%s/day in ICU without a ventilator, or ₹%s/day in ICU with a ventilator)",
				p.PackageName, p.Specialty,
				formatINR(pd.RoutineWardINR), formatINR(pd.HDUINR), formatINR(pd.ICUNoVentINR), formatINR(pd.ICUVentINR),
			)
		}
	case "tiered":
		if p.RateMaxINR > p.IndicativeRateINR {
			return fmt.Sprintf(
				"%s (%s), listed with an indicative rate of ₹%s to ₹%s depending on the treating hospital's city-tier classification",
				p.PackageName, p.Specialty, formatINR(p.IndicativeRateINR), formatINR(p.RateMaxINR),
			)
		}
	}
	return fmt.Sprintf("%s (%s), listed with an indicative rate of ₹%s", p.PackageName, p.Specialty, formatINR(p.IndicativeRateINR))
}

func formatINR(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	// Indian digit grouping (lakh/crore style: last 3 digits, then groups
	// of 2), matching how the rupee figures are written throughout the
	// source spec (e.g. "₹35,000", "₹1,20,000").
	var out []string
	out = append(out, s[len(s)-3:])
	s = s[:len(s)-3]
	for len(s) > 2 {
		out = append([]string{s[len(s)-2:]}, out...)
		s = s[:len(s)-2]
	}
	if len(s) > 0 {
		out = append([]string{s}, out...)
	}
	return strings.Join(out, ",")
}

func greenMessage(d tiering.Decision) string {
	pkg := *d.MatchedPackage
	return fmt.Sprintf(
		"Based on what you've described (%s), this matches a package that is listed as covered under PMJAY: %s. This hospital should not be asking for payment for this procedure if they are empanelled for %s.",
		lowerFirst(d.ExtractedSituationSummary), packageCitation(pkg), pkg.Specialty,
	)
}

func unspecifiedGreenMessage(d tiering.Decision) string {
	return fmt.Sprintf(
		"Based on what you've described (%s), this looks like a real procedure that may not be one of PMJAY's specifically named packages. The scheme has a separate discretionary category for exactly this situation (\"Unspecified Procedure\"), covering genuine procedures that don't map to a listed package, up to ₹%s per case within your family's overall yearly limit, at the treating hospital's judgement. This is not automatic: it needs the hospital to actually apply it, but it means \"not on the list\" is not the same as \"not covered.\"",
		lowerFirst(d.ExtractedSituationSummary), formatINR(hbp.UnspecifiedProcedureCapINR),
	)
}

func amberMessage(d tiering.Decision) string {
	hasReason := func(r tiering.AmberReason) bool {
		for _, x := range d.AmberReasons {
			if x == r {
				return true
			}
		}
		return false
	}

	switch {
	case hasReason(tiering.ReasonPendingPreauth):
		return fmt.Sprintf(
			"This sounds like it might be about something called pre-authorisation: before some procedures, the hospital has to get approval from the health scheme's regional office. What you're describing could mean one of two different things, and they're not the same:\n\n"+
				"- The approval is genuinely still pending (meaning it hasn't been decided yet, not refused).\n"+
				"- The approval was actually denied, and the hospital may not be explaining exactly why.\n\n"+
				"These need different responses, so before anything else, ask the billing desk this exact question: \"Has the pre-authorisation actually been submitted for this case, and what is its current status: approved, pending, or denied?\" (This relates to %s.)",
			packageCitation(*d.MatchedPackage),
		)

	case hasReason(tiering.ReasonCloseCandidates):
		return fmt.Sprintf(
			"Based on what you've described (%s), there's more than one PMJAY package this could genuinely be, and I don't want to guess which one applies when it actually matters which. The closest match I can point to is %s, but I'm not confident enough to say that's definitely the right one over another similar package. Before anything else, it would help to know the exact name of the procedure from the doctor or the discharge papers, if you have them.",
			lowerFirst(d.ExtractedSituationSummary), packageCitation(*d.MatchedPackage),
		)

	case hasReason(tiering.ReasonPartialExclusion):
		excl := *d.MatchedExclusion
		return fmt.Sprintf(
			"This is genuinely more complicated than a simple yes or no. %s %s Whether this specific case is covered depends on which state you're in and the exact procedure. It isn't a blanket rule either way, so I don't want to tell you confidently that you're covered or that you're not. This is a good case to check directly with the helpline (14555) or a Para Legal Volunteer, who can look at the specific state rules that apply to you.",
			excl.Description, excl.Nuance,
		)

	case hasReason(tiering.ReasonUnspecifiedUnclear):
		return fmt.Sprintf(
			"Based on what you've described (%s), I can't find a clear, specifically-named PMJAY package that matches well. There is a discretionary category for procedures that don't fit a named package, but I'm not confident enough in this match to tell you it applies here (that would be guessing, not helping). The most useful next step is asking the hospital directly what specific procedure name or code they're billing this under, so we have something concrete to check.",
			lowerFirst(d.ExtractedSituationSummary),
		)

	default: // ReasonLowConfidence, or a matched package with genuinely low confidence
		if d.MatchedPackage != nil {
			return fmt.Sprintf(
				"Based on what you've described (%s), the closest match I can find is %s, but I'm genuinely not confident enough in that match to tell you either way whether this is covered. Guessing here could do more harm than good. It would help a lot to know the exact procedure name from the doctor, or to ask the hospital directly what they're billing this as.",
				lowerFirst(d.ExtractedSituationSummary), packageCitation(*d.MatchedPackage),
			)
		}
		return fmt.Sprintf(
			"Based on what you've described (%s), I can't confidently match this to a specific PMJAY package with the information I have. Rather than guess, the most useful next step is asking the hospital directly for the exact procedure name or code they're billing this under.",
			lowerFirst(d.ExtractedSituationSummary),
		)
	}
}

func redMessage(d tiering.Decision) string {
	excl := *d.MatchedExclusion
	nuanceNote := ""
	if excl.Category == "cosmetic" {
		nuanceNote = " If there's a medical reason behind the procedure beyond appearance (for example, if a doctor has documented a functional health issue, not just a cosmetic preference), that could potentially change how this is classified, and it would be worth asking your doctor directly whether that applies here."
	}
	return fmt.Sprintf(
		"Yes, that's correct, and I want to be straightforward with you about this rather than suggest there's a dispute to be had. %s This isn't this particular hospital being difficult: it's a general rule of the scheme.%s As a %s, this would not be something to file a grievance about, since the hospital is correctly applying the rule.",
		excl.Description, nuanceNote, strings.ToLower(excl.DisplayName),
	)
}

func mixedMessage(d tiering.Decision) string {
	pkg := *d.MatchedPackage
	excl := *d.MatchedExclusion
	return fmt.Sprintf(
		"What you've described has two separate parts, and they need to be handled separately (not accepted or rejected as one bill).\n\n"+
			"The part that matches %s looks like a genuinely covered PMJAY package. The hospital should not be charging you for this part if they're empanelled for %s.\n\n"+
			"The part that matches %s is a confirmed exclusion (%s) and is correctly not covered, regardless of the rest of the bill.\n\n"+
			"Ask the billing desk to split the bill into these two parts explicitly, and only dispute the first part.",
		packageCitation(pkg), pkg.Specialty, strings.ToLower(excl.DisplayName), lowerFirst(excl.Description),
	)
}

func handoffMessage(d tiering.Decision) string {
	return "This sounds like it involves more than one thing at once, and it also sounds like a lot to sort through alone right now. I'm going to connect you with a Para Legal Volunteer: someone whose specific job is helping families through exactly this kind of situation, for free, and who is not employed by this hospital."
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	// Only lower-case if the string looks like a normal sentence starting
	// with an ordinary capital letter, so proper nouns/acronyms already
	// in the extracted summary aren't mangled.
	r := []rune(s)
	if len(r) > 1 && r[0] >= 'A' && r[0] <= 'Z' && !(r[1] >= 'A' && r[1] <= 'Z') {
		r[0] = r[0] + ('a' - 'A')
	}
	return string(r)
}
