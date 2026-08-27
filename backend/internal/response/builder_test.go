package response

import (
	"strings"
	"testing"

	"github.com/pmjay-advocate/backend/internal/hbp"
	"github.com/pmjay-advocate/backend/internal/tiering"
)

func testDS(t *testing.T) *hbp.Dataset {
	t.Helper()
	ds, err := hbp.Load()
	if err != nil {
		t.Fatalf("failed to load dataset: %v", err)
	}
	return ds
}

func pkgOrFatal(t *testing.T, ds *hbp.Dataset, code string) hbp.Package {
	t.Helper()
	p, ok := ds.PackageByCode(code)
	if !ok {
		t.Fatalf("test setup: package %q not found", code)
	}
	return p
}

func exclOrFatal(t *testing.T, ds *hbp.Dataset, category string) hbp.Exclusion {
	t.Helper()
	e, ok := ds.ExclusionByCategory(category)
	if !ok {
		t.Fatalf("test setup: exclusion %q not found", category)
	}
	return e
}

// ---------------------------------------------------------------------
// THE most important test in this entire codebase: the care-first
// message must be present, verbatim, on every single outcome, with no
// exceptions. This directly implements Appendix S's checklist item and
// Appendix T's definition of done for the safety rule ("a dedicated
// adversarial test... fails to produce" a delay-encouraging response).
// ---------------------------------------------------------------------

func TestBuild_CareFirstMessageIsAlwaysPresent_EveryOutcome(t *testing.T) {
	ds := testDS(t)
	gsPkg := pkgOrFatal(t, ds, "SEED-GS-001")
	orthoPkg := pkgOrFatal(t, ds, "SEED-ORTHO-001")
	cosmeticExcl := exclOrFatal(t, ds, "cosmetic")
	transplantExcl := exclOrFatal(t, ds, "organ_transplant_partial")

	decisions := map[tiering.Outcome]tiering.Decision{
		tiering.OutcomeGreen: {
			Outcome: tiering.OutcomeGreen, MatchedPackage: &gsPkg, MatchedPackageConfidence: 90,
			ExtractedSituationSummary: "test",
		},
		tiering.OutcomeAmber: {
			Outcome: tiering.OutcomeAmber, MatchedPackage: &gsPkg, AmberReasons: []tiering.AmberReason{tiering.ReasonLowConfidence},
			ExtractedSituationSummary: "test",
		},
		tiering.OutcomeRed: {
			Outcome: tiering.OutcomeRed, MatchedExclusion: &cosmeticExcl, MatchedExclusionConfidence: 90,
			ExtractedSituationSummary: "test",
		},
		tiering.OutcomeMixed: {
			Outcome: tiering.OutcomeMixed, MatchedPackage: &orthoPkg, MatchedExclusion: &cosmeticExcl,
			ExtractedSituationSummary: "test",
		},
		tiering.OutcomeHandoff: {
			Outcome: tiering.OutcomeHandoff, HandoffReasons: []tiering.HandoffReason{tiering.HandoffMultipleIssues},
			ExtractedSituationSummary: "test",
		},
	}
	// Also cover the amber-partial-exclusion path specifically, since it
	// takes a different branch through amberMessage.
	decisions["amber_partial_exclusion_variant"] = tiering.Decision{
		Outcome: tiering.OutcomeAmber, MatchedExclusion: &transplantExcl,
		AmberReasons: []tiering.AmberReason{tiering.ReasonPartialExclusion}, ExtractedSituationSummary: "test",
	}

	for name, d := range decisions {
		r := Build(d)
		if r.CareFirstMessage() != CareFirstText {
			t.Errorf("outcome %v: CareFirstMessage() = %q, want exact CareFirstText", name, r.CareFirstMessage())
		}
		if r.CareFirstMessage() == "" {
			t.Errorf("outcome %v: CareFirstMessage() is empty — this must never happen", name)
		}
		if r.Disclaimer() != DisclaimerText {
			t.Errorf("outcome %v: Disclaimer() = %q, want exact DisclaimerText", name, r.Disclaimer())
		}
		if r.Disclaimer() == "" {
			t.Errorf("outcome %v: Disclaimer() is empty — this must never happen", name)
		}
	}
}

func TestBuild_UnknownOutcomeFailsSafeToHandoffWithCareFirstIntact(t *testing.T) {
	// Simulates a future Outcome value the switch doesn't yet know about.
	d := tiering.Decision{Outcome: tiering.Outcome("some_future_outcome_nobody_added_a_case_for")}
	r := Build(d)
	if r.CareFirstMessage() != CareFirstText {
		t.Fatal("even the unknown-outcome fail-safe path must carry the care-first message")
	}
	if r.Disclaimer() != DisclaimerText {
		t.Fatal("even the unknown-outcome fail-safe path must carry the disclaimer")
	}
	if r.Outcome() != tiering.OutcomeHandoff {
		t.Errorf("expected unknown outcome to fail safe to Handoff, got %v", r.Outcome())
	}
}

// ---------------------------------------------------------------------
// H4's adversarial sweep: scan every generated text field, across a wide
// spread of decisions, for language that could be read as encouraging a
// delay or refusal of treatment. Per the spec's own kill condition for
// H4, a single instance here is a hard stop, not a statistic.
// ---------------------------------------------------------------------

var delayEncouragingPhrases = []string{
	"do not proceed", "don't proceed", "wait until", "refuse treatment",
	"refuse care", "delay the surgery", "delay treatment", "delay care",
	"hold off on treatment", "don't let them treat", "do not allow treatment",
	"stop the procedure", "cancel the surgery", "cancel treatment",
	"don't pay and don't proceed", "leave without treatment",
}

func allTextFields(r FamilyResponse) []string {
	fields := []string{
		r.CareFirstMessage(), r.Disclaimer(), r.TierMessage(), r.Citation(),
		r.ComplaintText(), r.HospitalScript(), r.HandoffSummary(), r.EvidencePrompt(),
	}
	fields = append(fields, r.ActionSteps()...)
	return fields
}

func TestBuild_H4_NoDelayEncouragingLanguageAcrossAllOutcomes(t *testing.T) {
	ds := testDS(t)
	gsPkg := pkgOrFatal(t, ds, "SEED-GS-001")
	obgPkg := pkgOrFatal(t, ds, "SEED-OBG-001") // requires_preauth: false
	cosmeticExcl := exclOrFatal(t, ds, "cosmetic")
	transplantExcl := exclOrFatal(t, ds, "organ_transplant_partial")

	adversarialDecisions := []tiering.Decision{
		{Outcome: tiering.OutcomeGreen, MatchedPackage: &gsPkg, ExtractedSituationSummary: "surgery needed urgently, hospital demanding cash before proceeding"},
		{Outcome: tiering.OutcomeAmber, MatchedPackage: &gsPkg, AmberReasons: []tiering.AmberReason{tiering.ReasonPendingPreauth}, ExtractedSituationSummary: "already on the operating table, payment dispute ongoing"},
		{Outcome: tiering.OutcomeAmber, MatchedPackage: &obgPkg, AmberReasons: []tiering.AmberReason{tiering.ReasonCloseCandidates}, ExtractedSituationSummary: "in labour right now, unclear which package applies"},
		{Outcome: tiering.OutcomeAmber, MatchedExclusion: &transplantExcl, AmberReasons: []tiering.AmberReason{tiering.ReasonPartialExclusion}, ExtractedSituationSummary: "transplant scheduled for tomorrow morning"},
		{Outcome: tiering.OutcomeRed, MatchedExclusion: &cosmeticExcl, ExtractedSituationSummary: "cosmetic procedure, hospital wants payment now"},
		{Outcome: tiering.OutcomeMixed, MatchedPackage: &gsPkg, MatchedExclusion: &cosmeticExcl, ExtractedSituationSummary: "mixed bill, procedure happening today"},
		{Outcome: tiering.OutcomeHandoff, HandoffReasons: []tiering.HandoffReason{tiering.HandoffMultipleIssues}, ExtractedSituationSummary: "confused, multiple staff giving conflicting instructions mid-treatment"},
	}

	for i, d := range adversarialDecisions {
		r := Build(d)
		for _, field := range allTextFields(r) {
			for _, phrase := range delayEncouragingPhrases {
				if idx := findUnnegatedPhrase(field, phrase); idx != -1 {
					t.Errorf("case %d (outcome %v): generated text contains an UNNEGATED delay-encouraging phrase %q in field: %q", i, d.Outcome, phrase, field)
				}
			}
		}
		// Additionally: the care-first message itself must always be the
		// FIRST thing a consumer of this response would see it as -
		// verify it's exactly the constant, not a paraphrase that could
		// have quietly drifted.
		if r.CareFirstMessage() != CareFirstText {
			t.Errorf("case %d: care-first message drifted from the canonical text", i)
		}
	}
}

// negationTriggers are words that, if found immediately before a matched
// phrase, flip its meaning — "do not let this delay care" is a safety
// instruction, not a violation, even though the raw substring "delay
// care" appears in it. A naive substring blocklist can't tell these
// apart; checking the preceding window for a negation cue is a more
// honest test of intent than a bare Contains() check would be.
var negationTriggers = []string{
	"not ", "n't ", "never ", "avoid ", "without ", "against ",
}

const negationWindow = 40

// findUnnegatedPhrase returns the index of phrase within field if it
// appears WITHOUT a negation trigger word in the preceding window, or -1
// if either the phrase isn't present or it's present but negated.
func findUnnegatedPhrase(field, phrase string) int {
	lowerField := strings.ToLower(field)
	lowerPhrase := strings.ToLower(phrase)

	searchFrom := 0
	for {
		idx := strings.Index(lowerField[searchFrom:], lowerPhrase)
		if idx == -1 {
			return -1
		}
		absIdx := searchFrom + idx

		windowStart := absIdx - negationWindow
		if windowStart < 0 {
			windowStart = 0
		}
		preceding := lowerField[windowStart:absIdx]

		negated := false
		for _, trig := range negationTriggers {
			if strings.Contains(preceding, trig) {
				negated = true
				break
			}
		}
		if !negated {
			return absIdx
		}
		searchFrom = absIdx + len(lowerPhrase)
		if searchFrom >= len(lowerField) {
			return -1
		}
	}
}

func TestFindUnnegatedPhrase_DirectlyTestsTheNegationLogicItself(t *testing.T) {
	// The negation-detection helper above is itself safety-critical
	// test infrastructure, so it gets its own direct tests rather than
	// only being exercised indirectly through the sweep above.
	cases := []struct {
		field, phrase string
		wantFound     bool
	}{
		{"do not let this delay care if it's urgent", "delay care", false},
		{"you should delay care until tomorrow", "delay care", true},
		{"never delay treatment for this", "delay treatment", false},
		{"please delay treatment for now", "delay treatment", true},
		{"we recommend you wait until the dispute is resolved", "wait until", true},
		{"", "delay care", false},
	}
	for _, c := range cases {
		idx := findUnnegatedPhrase(c.field, c.phrase)
		found := idx != -1
		if found != c.wantFound {
			t.Errorf("findUnnegatedPhrase(%q, %q): found=%v, want=%v", c.field, c.phrase, found, c.wantFound)
		}
	}
}

func TestBuild_CareFirstMessage_PositivelyInstructsProceedingWithCare(t *testing.T) {
	// Not just "absence of bad phrases" — assert the presence of the
	// actual affirmative instruction, since a passing blocklist scan
	// alone could hide a message that says nothing useful at all.
	r := Build(tiering.Decision{Outcome: tiering.OutcomeGreen, MatchedPackage: &hbp.Package{PackageCode: "X", PackageName: "X", Specialty: "X", Verified: false}})
	msg := strings.ToLower(r.CareFirstMessage())
	if !strings.Contains(msg, "treatment first") {
		t.Error("expected care-first message to explicitly say treatment comes first")
	}
	if !strings.Contains(msg, "dispute") {
		t.Error("expected care-first message to explicitly mention disputing after")
	}
}

// ---------------------------------------------------------------------
// Citation discipline (Explainability, Section 11): every coverage claim
// must cite something specific.
// ---------------------------------------------------------------------

func TestBuild_CitationNeverEmptyForCoverageClaimingOutcomes(t *testing.T) {
	ds := testDS(t)
	gsPkg := pkgOrFatal(t, ds, "SEED-GS-001")
	cosmeticExcl := exclOrFatal(t, ds, "cosmetic")

	cases := []tiering.Decision{
		{Outcome: tiering.OutcomeGreen, MatchedPackage: &gsPkg},
		{Outcome: tiering.OutcomeAmber, MatchedPackage: &gsPkg, AmberReasons: []tiering.AmberReason{tiering.ReasonLowConfidence}},
		{Outcome: tiering.OutcomeRed, MatchedExclusion: &cosmeticExcl},
		{Outcome: tiering.OutcomeMixed, MatchedPackage: &gsPkg, MatchedExclusion: &cosmeticExcl},
	}
	for _, d := range cases {
		r := Build(d)
		if r.Citation() == "" {
			t.Errorf("outcome %v: Citation() is empty — every coverage-relevant claim must cite something (Section 11)", d.Outcome)
		}
	}
}

func TestBuild_UnverifiedRateNeverStatedAsFact(t *testing.T) {
	ds := testDS(t)
	gsPkg := pkgOrFatal(t, ds, "SEED-GS-001")
	if gsPkg.Verified {
		t.Fatal("test setup assumption violated: SEED-GS-001 should be an unverified seed record")
	}
	d := tiering.Decision{Outcome: tiering.OutcomeGreen, MatchedPackage: &gsPkg, ExtractedSituationSummary: "test"}
	r := Build(d)
	rateStr := "₹" + formatINR(gsPkg.IndicativeRateINR)
	if strings.Contains(r.TierMessage(), rateStr) {
		t.Errorf("tier message states an unverified placeholder rate (%s) as if it were a checked fact: %q", rateStr, r.TierMessage())
	}
}

func TestBuild_VerifiedRateIsStated(t *testing.T) {
	ds := testDS(t)
	unspecified := pkgOrFatal(t, ds, "UNSPECIFIED")
	if !unspecified.Verified {
		t.Fatal("test setup assumption violated: UNSPECIFIED should be verified")
	}
	d := tiering.Decision{Outcome: tiering.OutcomeGreen, MatchedPackage: &unspecified, ExtractedSituationSummary: "test"}
	r := Build(d)
	if !strings.Contains(r.TierMessage(), "1,00,000") {
		t.Errorf("expected the verified Unspecified Procedure cap to be stated explicitly, got: %q", r.TierMessage())
	}
}

// ---------------------------------------------------------------------
// Per-outcome content shape.
// ---------------------------------------------------------------------

func TestBuild_Green_ProducesComplaintAndScript(t *testing.T) {
	ds := testDS(t)
	pkg := pkgOrFatal(t, ds, "SEED-GS-001")
	r := Build(tiering.Decision{Outcome: tiering.OutcomeGreen, MatchedPackage: &pkg, ExtractedSituationSummary: "test"})
	if r.ComplaintText() == "" {
		t.Error("expected non-empty complaint text for Green")
	}
	if r.HospitalScript() == "" {
		t.Error("expected non-empty hospital script for Green")
	}
	if r.EvidencePrompt() == "" {
		t.Error("expected non-empty evidence prompt for Green")
	}
	if len(r.ActionSteps()) == 0 {
		t.Error("expected non-empty action steps for Green")
	}
}

func TestBuild_Amber_NeverProducesAComplaint(t *testing.T) {
	// Section 9: "the system never renders its own final verdict on a
	// genuinely disputed case" — generating a complaint at amber would
	// be exactly that.
	ds := testDS(t)
	pkg := pkgOrFatal(t, ds, "SEED-GS-001")
	r := Build(tiering.Decision{Outcome: tiering.OutcomeAmber, MatchedPackage: &pkg, AmberReasons: []tiering.AmberReason{tiering.ReasonLowConfidence}, ExtractedSituationSummary: "test"})
	if r.ComplaintText() != "" {
		t.Errorf("amber must never produce complaint text, got: %q", r.ComplaintText())
	}
}

func TestBuild_Red_NeverProducesAComplaintOrActionSteps(t *testing.T) {
	// Section 7 Step 5: "it does not manufacture a grievance where none
	// exists."
	ds := testDS(t)
	excl := exclOrFatal(t, ds, "cosmetic")
	r := Build(tiering.Decision{Outcome: tiering.OutcomeRed, MatchedExclusion: &excl, ExtractedSituationSummary: "test"})
	if r.ComplaintText() != "" {
		t.Errorf("red must never produce complaint text, got: %q", r.ComplaintText())
	}
	if len(r.ActionSteps()) != 0 {
		t.Errorf("red should have no action steps (nothing to dispute), got: %v", r.ActionSteps())
	}
	if r.EvidencePrompt() != "" {
		t.Errorf("red should have no evidence prompt (nothing to build a case toward), got: %q", r.EvidencePrompt())
	}
}

func TestBuild_Red_CosmeticIncludesFunctionalReasonCaveat(t *testing.T) {
	// Worked example 29.3's specific nuance: a documented functional
	// medical reason can change the classification.
	ds := testDS(t)
	excl := exclOrFatal(t, ds, "cosmetic")
	r := Build(tiering.Decision{Outcome: tiering.OutcomeRed, MatchedExclusion: &excl, ExtractedSituationSummary: "test"})
	if !strings.Contains(r.TierMessage(), "functional") {
		t.Errorf("expected the cosmetic red-tier message to mention the functional-reason caveat, got: %q", r.TierMessage())
	}
}

func TestBuild_Mixed_CitesBothPackageAndExclusion(t *testing.T) {
	ds := testDS(t)
	pkg := pkgOrFatal(t, ds, "SEED-ORTHO-001")
	excl := exclOrFatal(t, ds, "cosmetic")
	r := Build(tiering.Decision{Outcome: tiering.OutcomeMixed, MatchedPackage: &pkg, MatchedExclusion: &excl, ExtractedSituationSummary: "test"})
	if !strings.Contains(r.Citation(), pkg.PackageName) {
		t.Errorf("expected mixed citation to include package name, got: %q", r.Citation())
	}
	if !strings.Contains(r.Citation(), excl.DisplayName) {
		t.Errorf("expected mixed citation to include exclusion name, got: %q", r.Citation())
	}
	if !strings.Contains(r.TierMessage(), "separately") && !strings.Contains(r.TierMessage(), "two separate") {
		t.Errorf("expected mixed message to explicitly say the bill needs splitting, got: %q", r.TierMessage())
	}
}

func TestBuild_Handoff_NoCitationNoComplaintButHasSummary(t *testing.T) {
	r := Build(tiering.Decision{
		Outcome:                   tiering.OutcomeHandoff,
		HandoffReasons:            []tiering.HandoffReason{tiering.HandoffMultipleIssues},
		ExtractedSituationSummary: "the family's original situation, word for word marker ABC123",
	})
	if r.Citation() != "" {
		t.Errorf("handoff should cite nothing (Section 12), got: %q", r.Citation())
	}
	if r.ComplaintText() != "" {
		t.Errorf("handoff should not generate a complaint, got: %q", r.ComplaintText())
	}
	if r.HandoffSummary() == "" {
		t.Fatal("handoff summary must not be empty")
	}
	if !strings.Contains(r.HandoffSummary(), "ABC123") {
		t.Error("handoff summary must carry the extracted situation through (Section 12: nothing re-explained from scratch)")
	}
}

func TestBuild_Handoff_DistressReasonGetsItsOwnSpecificMessage(t *testing.T) {
	// A family flagged for handoff specifically because they showed
	// distress needs the volunteer to know that going in — the generic
	// "genuine ambiguity" text a HandoffMultipleIssues case gets doesn't
	// carry that signal, and losing it here would mean the one person
	// positioned to notice never does.
	distressed := Build(tiering.Decision{
		Outcome:                   tiering.OutcomeHandoff,
		HandoffReasons:            []tiering.HandoffReason{tiering.HandoffDistressWithUnclear},
		ExtractedSituationSummary: "test situation",
	})
	if !strings.Contains(distressed.HandoffSummary(), "distress") {
		t.Errorf("expected the distress-specific handoff reason to mention distress, got: %q", distressed.HandoffSummary())
	}

	multiIssue := Build(tiering.Decision{
		Outcome:                   tiering.OutcomeHandoff,
		HandoffReasons:            []tiering.HandoffReason{tiering.HandoffMultipleIssues},
		ExtractedSituationSummary: "test situation",
	})
	if strings.Contains(multiIssue.HandoffSummary(), "distress") {
		t.Errorf("a multiple-issues handoff should not carry the distress-specific wording, got: %q", multiIssue.HandoffSummary())
	}
	if distressed.HandoffSummary() == multiIssue.HandoffSummary() {
		t.Error("expected the two handoff reasons to produce distinguishable summaries, got identical text")
	}
}

func TestBuild_Amber_PendingReason_AsksTheRightQuestion(t *testing.T) {
	ds := testDS(t)
	pkg := pkgOrFatal(t, ds, "SEED-CARD-001")
	r := Build(tiering.Decision{
		Outcome: tiering.OutcomeAmber, MatchedPackage: &pkg,
		AmberReasons: []tiering.AmberReason{tiering.ReasonPendingPreauth}, ExtractedSituationSummary: "test",
	})
	if !strings.Contains(r.TierMessage(), "pending") {
		t.Errorf("expected pending-reason amber message to explain the pending concept, got: %q", r.TierMessage())
	}
	found := false
	for _, step := range r.ActionSteps() {
		if strings.Contains(step, "current status") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected an action step asking the hospital for the preauth status, got: %v", r.ActionSteps())
	}
}
