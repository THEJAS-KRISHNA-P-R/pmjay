package tiering

import (
	"testing"

	"github.com/pmjay-advocate/backend/internal/extract"
	"github.com/pmjay-advocate/backend/internal/hbp"
)

func testDataset(t *testing.T) *hbp.Dataset {
	t.Helper()
	ds, err := hbp.Load()
	if err != nil {
		t.Fatalf("failed to load embedded dataset: %v", err)
	}
	return ds
}

// ---------------------------------------------------------------------
// Control arm: clean, unambiguous cases (H3's first arm). These should
// never require a human to squint at the output.
// ---------------------------------------------------------------------

func TestDecide_ControlArm_CleanGreen(t *testing.T) {
	// Mirrors worked example 26.1 / 29.1: gallbladder removal, confirmed
	// stones, flat "not covered" claim with no pending language.
	ds := testDataset(t)
	desc := "My mother needs her gallbladder removed, doctor confirmed stones on scan, hospital billing desk just told us PMJAY won't cover it and we need to pay before they'll schedule the surgery."
	result := extract.Result{
		ExtractedSituationSummary: "Family reports gallbladder removal for confirmed gallstones; hospital demanding upfront cash payment.",
		Candidates: []extract.CandidateMatch{
			{PackageCode: "SEED-GS-001", ConfidenceP: 93, Reasoning: "clear match to laparoscopic cholecystectomy"},
			{PackageCode: "UNSPECIFIED", ConfidenceP: 5, Reasoning: "not needed, named package fits well"},
		},
		Pending: extract.SignalNotApplicable,
	}

	d := Decide(ds, desc, result)

	if d.Outcome != OutcomeGreen {
		t.Fatalf("expected OutcomeGreen, got %v (reasons: %v)", d.Outcome, d.AmberReasons)
	}
	if d.MatchedPackage == nil || d.MatchedPackage.PackageCode != "SEED-GS-001" {
		t.Errorf("expected matched package SEED-GS-001, got %+v", d.MatchedPackage)
	}
}

func TestDecide_ControlArm_CleanRed(t *testing.T) {
	// Mirrors worked example 26.3 / 29.3: cosmetic procedure, clean
	// exclusion, no competing covered-package signal.
	ds := testDataset(t)
	desc := "We want a cosmetic nose job for my daughter before her wedding, hospital says PMJAY won't pay."
	result := extract.Result{
		ExtractedSituationSummary: "Family requesting cosmetic rhinoplasty, hospital says not covered.",
		Candidates: []extract.CandidateMatch{
			{PackageCode: "UNSPECIFIED", ConfidenceP: 10, Reasoning: "no real procedure package fits a cosmetic request"},
		},
		ExclusionMatches: []extract.ExclusionMatch{
			{Category: "cosmetic", ConfidenceP: 95, Reasoning: "explicitly described as cosmetic, appearance-motivated"},
		},
		Pending: extract.SignalNotApplicable,
	}

	d := Decide(ds, desc, result)

	if d.Outcome != OutcomeRed {
		t.Fatalf("expected OutcomeRed, got %v", d.Outcome)
	}
	if d.MatchedExclusion == nil || d.MatchedExclusion.Category != "cosmetic" {
		t.Errorf("expected matched exclusion 'cosmetic', got %+v", d.MatchedExclusion)
	}
}

// ---------------------------------------------------------------------
// Mechanical arm: obvious-by-construction cases that still need to reach
// the right mechanical answer without being accidentally miscategorised.
// ---------------------------------------------------------------------

func TestDecide_MechanicalArm_CardiacGreenWithPartialCoverageLanguage(t *testing.T) {
	// Extended test bank #1: "hospital says card won't cover the stent
	// itself only the room" — the procedure itself is a clear match; a
	// separately-billed room dispute shouldn't itself create ambiguity
	// about the *procedure's* coverage.
	ds := testDataset(t)
	desc := "Father collapsed, doctors say it's his heart, need an urgent stent, hospital says card won't cover the stent itself only the room."
	result := extract.Result{
		ExtractedSituationSummary: "Urgent cardiac stent; hospital disputes coverage of a room charge, not clearly the stent procedure itself.",
		Candidates: []extract.CandidateMatch{
			{PackageCode: "SEED-CARD-001", ConfidenceP: 88, Reasoning: "clear stent/PTCA match"},
		},
		Pending: extract.SignalNotApplicable,
	}
	d := Decide(ds, desc, result)
	if d.Outcome != OutcomeGreen {
		t.Fatalf("expected OutcomeGreen for the stent procedure itself, got %v", d.Outcome)
	}
}

func TestDecide_MechanicalArm_OrthopaedicMixedCoveredAndCosmetic(t *testing.T) {
	// Extended test bank #3: knee replacement (covered) bundled with a
	// cosmetic scar-reduction add-on (excluded) billed together.
	ds := testDataset(t)
	desc := "Doctor wants to do a knee replacement but also mentioned a cosmetic scar-reduction procedure at the same time, hospital is billing both as PMJAY-covered."
	result := extract.Result{
		ExtractedSituationSummary: "Knee replacement bundled with an unrelated cosmetic scar-reduction procedure in the same bill.",
		Candidates: []extract.CandidateMatch{
			{PackageCode: "SEED-ORTHO-001", ConfidenceP: 90, Reasoning: "clear total knee replacement match"},
		},
		ExclusionMatches: []extract.ExclusionMatch{
			{Category: "cosmetic", ConfidenceP: 82, Reasoning: "scar-reduction add-on is cosmetic, unrelated to the knee procedure"},
		},
		Pending: extract.SignalNotApplicable,
	}
	d := Decide(ds, desc, result)
	if d.Outcome != OutcomeMixed {
		t.Fatalf("expected OutcomeMixed (split bill), got %v", d.Outcome)
	}
	if d.MatchedPackage == nil || d.MatchedPackage.PackageCode != "SEED-ORTHO-001" {
		t.Errorf("expected matched package SEED-ORTHO-001 alongside the exclusion, got %+v", d.MatchedPackage)
	}
	if d.MatchedExclusion == nil || d.MatchedExclusion.Category != "cosmetic" {
		t.Errorf("expected matched exclusion 'cosmetic' alongside the package, got %+v", d.MatchedExclusion)
	}
}

func TestDecide_MechanicalArm_OrganTransplantIsAmberNotRed(t *testing.T) {
	// Section 6.6: organ transplant is only partially/state-dependently
	// covered — a flat red here would itself be an overclaim.
	ds := testDataset(t)
	desc := "Doctor says my father needs a kidney transplant, hospital says the scheme doesn't cover transplants."
	result := extract.Result{
		ExtractedSituationSummary: "Family describes a kidney transplant, hospital claims no coverage.",
		Candidates: []extract.CandidateMatch{
			{PackageCode: "UNSPECIFIED", ConfidenceP: 20, Reasoning: "no specific named package for transplant in this seed dataset"},
		},
		ExclusionMatches: []extract.ExclusionMatch{
			{Category: "organ_transplant_partial", ConfidenceP: 90, Reasoning: "explicitly a kidney transplant"},
		},
		Pending: extract.SignalNotApplicable,
	}
	d := Decide(ds, desc, result)
	if d.Outcome != OutcomeAmber {
		t.Fatalf("expected OutcomeAmber for the nuanced organ-transplant exclusion, got %v", d.Outcome)
	}
	found := false
	for _, r := range d.AmberReasons {
		if r == ReasonPartialExclusion {
			found = true
		}
	}
	if !found {
		t.Errorf("expected ReasonPartialExclusion among amber reasons, got %v", d.AmberReasons)
	}
}

// ---------------------------------------------------------------------
// Ambiguous arm (H3's third arm, and its explicit safety kill condition):
// these must NEVER resolve to a confident Green or Red.
// ---------------------------------------------------------------------

func TestDecide_AmbiguousArm_PendingPreauthMirrorsWorkedExample(t *testing.T) {
	// Worked example 26.2 / 29.2.
	ds := testDataset(t)
	desc := "The hospital did the operation already, but now they're saying the insurance side hasn't 'cleared' it and we have to pay."
	result := extract.Result{
		ExtractedSituationSummary: "Procedure already performed; hospital citing an uncleared insurance/preauth status as the reason for demanding payment.",
		Candidates: []extract.CandidateMatch{
			{PackageCode: "SEED-CARD-001", ConfidenceP: 80, Reasoning: "plausible cardiac procedure given prior conversation context"},
		},
		Pending: extract.SignalPendingLikely,
	}
	d := Decide(ds, desc, result)
	if d.Outcome != OutcomeAmber {
		t.Fatalf("expected OutcomeAmber for a genuinely pending preauth situation, got %v", d.Outcome)
	}
}

func TestDecide_AmbiguousArm_CloseCandidatesForceAmberRegardlessOfTopScore(t *testing.T) {
	// Appendix Z's central calibration rule: even a top score that reads
	// as "confident" in isolation must go amber if a close second exists.
	ds := testDataset(t)
	result := extract.Result{
		ExtractedSituationSummary: "Ambiguous between two plausible cardiac packages.",
		Candidates: []extract.CandidateMatch{
			{PackageCode: "SEED-CARD-001", ConfidenceP: 82, Reasoning: "stent"},
			{PackageCode: "SEED-CARD-002", ConfidenceP: 74, Reasoning: "bypass — genuinely plausible too"},
		},
		Pending: extract.SignalNotApplicable,
	}
	d := Decide(ds, "vague cardiac description", result)
	if d.Outcome != OutcomeAmber {
		t.Fatalf("expected OutcomeAmber when top two candidates are within GapThreshold, got %v (top=82, second=74, gap=8 < %d)", d.Outcome, GapThreshold)
	}
	foundReason := false
	for _, r := range d.AmberReasons {
		if r == ReasonCloseCandidates {
			foundReason = true
		}
	}
	if !foundReason {
		t.Errorf("expected ReasonCloseCandidates, got %v", d.AmberReasons)
	}
}

func TestDecide_AmbiguousArm_WideGapStillGreen(t *testing.T) {
	// Sanity check on the other side of the same rule: a wide gap between
	// top two candidates should NOT force amber.
	ds := testDataset(t)
	result := extract.Result{
		Candidates: []extract.CandidateMatch{
			{PackageCode: "SEED-GS-001", ConfidenceP: 91, Reasoning: "clear"},
			{PackageCode: "SEED-GS-002", ConfidenceP: 30, Reasoning: "unlikely"},
		},
		Pending: extract.SignalNotApplicable,
	}
	d := Decide(ds, "clear gallbladder description", result)
	if d.Outcome != OutcomeGreen {
		t.Fatalf("expected OutcomeGreen when gap is wide (91 vs 30), got %v", d.Outcome)
	}
}

func TestDecide_AmbiguousArm_LowConfidenceNeverAssertedAsGreenOrRed(t *testing.T) {
	// H3's kill condition, directly: sweep a batch of deliberately
	// low/ambiguous-confidence cases and assert NONE resolve to Green or
	// Red. A confidently-wrong answer on this arm is defined by the spec
	// itself as a safety failure, not an accuracy statistic to average
	// away — so this test has no tolerance for a single failure.
	ds := testDataset(t)
	ambiguousResults := []extract.Result{
		{Candidates: []extract.CandidateMatch{{PackageCode: "SEED-GS-001", ConfidenceP: 40, Reasoning: "weak"}}, Pending: extract.SignalNotApplicable},
		{Candidates: []extract.CandidateMatch{{PackageCode: "SEED-ORTHO-001", ConfidenceP: 55, Reasoning: "unsure"}}, Pending: extract.SignalUnclear},
		{Candidates: []extract.CandidateMatch{{PackageCode: "UNSPECIFIED", ConfidenceP: 45, Reasoning: "doesn't fit anything named well"}}, Pending: extract.SignalNotApplicable},
		{Candidates: []extract.CandidateMatch{{PackageCode: "SEED-OBG-001", ConfidenceP: 60, Reasoning: "guess"}, {PackageCode: "SEED-OBG-002", ConfidenceP: 58, Reasoning: "also plausible"}}, Pending: extract.SignalNotApplicable},
	}
	for i, r := range ambiguousResults {
		d := Decide(ds, "deliberately ambiguous input", r)
		if d.Outcome == OutcomeGreen || d.Outcome == OutcomeRed {
			t.Errorf("case %d: ambiguous input resolved to confident %v — this is exactly H3's safety-failure kill condition", i, d.Outcome)
		}
	}
}

func TestDecide_AmbiguousArm_EmpanelmentEdgeCaseIsNotFalselyConfident(t *testing.T) {
	// Extended test bank #4: prior card use for cardiology at a hospital
	// is not evidence the same hospital is empanelled for ophthalmology.
	// A naive system might over-trust "this hospital used the card
	// before" — the extraction/tiering layers must not do that; low
	// confidence on the eye-surgery package here is the honest answer,
	// which this test encodes as amber, not a confident guess either way.
	ds := testDataset(t)
	desc := "This hospital did my father's cardiology procedure before under the card, now says they can't do the follow-up eye surgery under it."
	result := extract.Result{
		ExtractedSituationSummary: "Prior cardiology procedure at this hospital under the card; hospital now says a follow-up eye surgery isn't covered there.",
		Candidates: []extract.CandidateMatch{
			{PackageCode: "SEED-OPH-001", ConfidenceP: 55, Reasoning: "cataract/eye package plausible but empanelment for this specialty at this specific hospital is unconfirmed"},
		},
		Pending: extract.SignalUnclear,
	}
	d := Decide(ds, desc, result)
	if d.Outcome == OutcomeGreen || d.Outcome == OutcomeRed {
		t.Fatalf("expected honest amber for an unconfirmed empanelment situation, got confident %v", d.Outcome)
	}
}

// ---------------------------------------------------------------------
// Handoff routing
// ---------------------------------------------------------------------

func TestDecide_Handoff_MultipleDistinctIssues(t *testing.T) {
	// Worked example 26.4 / 29.4.
	ds := testDataset(t)
	desc := "They already operated, then said something about a form not being right, then a different person said something about my husband's ID not matching my name on the card, and now they want us to leave and come back tomorrow."
	result := extract.Result{
		ExtractedSituationSummary:      "Multiple separate issues: a form issue, a name/ID mismatch question, and a request to return the next day, after the procedure was already completed.",
		MultipleDistinctIssuesDetected: true,
		FamilyDistressSignal:           true,
		Candidates:                     []extract.CandidateMatch{{PackageCode: "UNSPECIFIED", ConfidenceP: 10, Reasoning: "too tangled to match cleanly"}},
	}
	d := Decide(ds, desc, result)
	if d.Outcome != OutcomeHandoff {
		t.Fatalf("expected OutcomeHandoff for a genuinely multi-issue case, got %v", d.Outcome)
	}
	found := false
	for _, r := range d.HandoffReasons {
		if r == HandoffMultipleIssues {
			found = true
		}
	}
	if !found {
		t.Errorf("expected HandoffMultipleIssues reason, got %v", d.HandoffReasons)
	}
}

func TestDecide_Handoff_ContradictoryDetailsFromDifferentStaff(t *testing.T) {
	// Extended test bank #5.
	ds := testDataset(t)
	desc := "One staff member said it was denied for one reason, another staff member gave a completely different reason for the same denial, and we don't know which is true."
	result := extract.Result{
		ExtractedSituationSummary:      "Two hospital staff members gave contradictory reasons for the same denial.",
		MultipleDistinctIssuesDetected: true,
		Candidates:                     []extract.CandidateMatch{{PackageCode: "UNSPECIFIED", ConfidenceP: 15, Reasoning: "contradictory account, cannot safely resolve to one package"}},
	}
	d := Decide(ds, desc, result)
	if d.Outcome != OutcomeHandoff {
		t.Fatalf("expected OutcomeHandoff for contradictory staff accounts, got %v", d.Outcome)
	}
}

func TestDecide_Handoff_DistressPlusUnresolvedAmbiguityEscalatesFromAmber(t *testing.T) {
	ds := testDataset(t)
	result := extract.Result{
		FamilyDistressSignal: true,
		Candidates:           []extract.CandidateMatch{{PackageCode: "SEED-GS-001", ConfidenceP: 45, Reasoning: "weak match"}},
		Pending:              extract.SignalUnclear,
	}
	d := Decide(ds, "distressed and unclear description", result)
	if d.Outcome != OutcomeHandoff {
		t.Fatalf("expected distress + low confidence to escalate to Handoff, got %v", d.Outcome)
	}
}

func TestDecide_NoHandoff_DistressAloneOnACleanGreenCase(t *testing.T) {
	// The deliberate design decision documented in Decide's doc comment:
	// distress by itself, on an otherwise clean and confident case,
	// should NOT force a handoff. A distressed family with an
	// unambiguous situation is best served by a fast, clear, calm
	// answer — exactly what worked example 29.1 itself shows.
	ds := testDataset(t)
	result := extract.Result{
		FamilyDistressSignal: true,
		Candidates:           []extract.CandidateMatch{{PackageCode: "SEED-GS-001", ConfidenceP: 92, Reasoning: "clear despite distressed phrasing"}},
		Pending:              extract.SignalNotApplicable,
	}
	d := Decide(ds, "panicked but clearly describes gallbladder situation", result)
	if d.Outcome != OutcomeGreen {
		t.Fatalf("expected distress alone on a clean case to still resolve Green, got %v", d.Outcome)
	}
}

// ---------------------------------------------------------------------
// Robustness / defensive edge cases
// ---------------------------------------------------------------------

func TestDecide_EmptyCandidatesDoesNotPanicAndIsHonestAmber(t *testing.T) {
	ds := testDataset(t)
	result := extract.Result{Candidates: []extract.CandidateMatch{}}
	d := Decide(ds, "", result)
	if d.Outcome != OutcomeAmber {
		t.Errorf("expected amber for zero candidates, got %v", d.Outcome)
	}
}

func TestDecide_UnknownPackageCodeFromModelDoesNotCrashOrFalselyAssert(t *testing.T) {
	// A hallucinated package code that doesn't exist in our dataset must
	// never be cited — Explainability (Section 11) requires every claim
	// trace back to a real record.
	ds := testDataset(t)
	result := extract.Result{
		Candidates: []extract.CandidateMatch{{PackageCode: "SOMETHING-THE-MODEL-MADE-UP", ConfidenceP: 95, Reasoning: "hallucinated"}},
	}
	d := Decide(ds, "test", result)
	if d.Outcome == OutcomeGreen {
		t.Fatalf("must never assert Green off a package code that isn't in the dataset")
	}
	if d.MatchedPackage != nil {
		t.Errorf("MatchedPackage should be nil for an unresolvable code, got %+v", d.MatchedPackage)
	}
}

func TestDecide_UnknownExclusionCategoryFromModelDoesNotCrash(t *testing.T) {
	ds := testDataset(t)
	result := extract.Result{
		Candidates:       []extract.CandidateMatch{{PackageCode: "UNSPECIFIED", ConfidenceP: 10, Reasoning: "x"}},
		ExclusionMatches: []extract.ExclusionMatch{{Category: "made_up_category", ConfidenceP: 99, Reasoning: "hallucinated"}},
	}
	d := Decide(ds, "test", result)
	if d.Outcome == OutcomeRed {
		t.Fatalf("must never assert Red off an exclusion category that isn't in the dataset")
	}
}

func TestDecide_RequiresPreauthFalse_PendingLanguageIgnored(t *testing.T) {
	// Normal vaginal delivery (SEED-OBG-002) is seeded with
	// RequiresPreauth: false. Pending-preauth language should be
	// structurally inapplicable to it, not just usually irrelevant.
	ds := testDataset(t)
	pkg, ok := ds.PackageByCode("SEED-OBG-002")
	if !ok {
		t.Fatal("test setup: SEED-OBG-002 not found in dataset")
	}
	if pkg.RequiresPreauth {
		t.Fatal("test setup assumption violated: SEED-OBG-002 must have RequiresPreauth=false for this test to be meaningful")
	}

	result := extract.Result{
		Candidates: []extract.CandidateMatch{{PackageCode: "SEED-OBG-002", ConfidenceP: 90, Reasoning: "clear normal delivery match"}},
		Pending:    extract.SignalPendingLikely, // model (incorrectly, or about something else) flags pending
	}
	d := Decide(ds, "normal delivery, insurance says pending", result)
	if d.Outcome != OutcomeGreen {
		t.Errorf("expected Green — a package that never requires preauth can't have a pending-preauth problem — got %v", d.Outcome)
	}
}

func TestDecide_AlwaysCarriesExtractedSummaryThrough(t *testing.T) {
	// Section 11, Context & continuity: nothing gathered during intake
	// should be dropped anywhere downstream, including into a Decision.
	ds := testDataset(t)
	result := extract.Result{
		ExtractedSituationSummary: "distinctive marker text xyz789",
		Candidates:                []extract.CandidateMatch{{PackageCode: "SEED-GS-001", ConfidenceP: 90, Reasoning: "x"}},
	}
	for _, outcome := range []extract.Result{result} {
		d := Decide(ds, "test", outcome)
		if d.ExtractedSituationSummary != "distinctive marker text xyz789" {
			t.Errorf("extracted summary was not carried through to the Decision: got %q", d.ExtractedSituationSummary)
		}
	}
}
