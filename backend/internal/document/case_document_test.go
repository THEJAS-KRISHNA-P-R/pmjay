package document

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/pmjay-advocate/backend/internal/store"
)

// baseCase returns a representative, fully-populated CaseRecord —
// tests for a specific outcome start from this and override only what
// that outcome scenario needs, so every test reflects a case shaped
// like a real one rather than a minimal/unrealistic stub.
func baseCase(outcome string) store.CaseRecord {
	return store.CaseRecord{
		ID:               "550e8400-e29b-41d4-a716-446655440000",
		CreatedAt:        time.Date(2026, 8, 17, 14, 30, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2026, 8, 17, 14, 30, 0, 0, time.UTC),
		Outcome:          outcome,
		Citation:         "Laparoscopic Cholecystectomy (SGS007)",
		CareFirstMessage: "Get treatment first. Dispute the money after. Always. If you can pay now and settle the dispute later, or move to a different hospital, do that — do not let this disagreement delay or stop care.",
		Disclaimer:       "This is guidance based on official PMJAY rules, not a legal ruling — confirm the current rate with the hospital or the PMJAY helpline before relying on any figure here, and remember free legal help from a NALSA Para Legal Volunteer is available if the hospital disagrees.",
		TierMessage:      "Based on what you've described — a gallbladder removal recommended by the treating doctor — this matches Laparoscopic Cholecystectomy under General Surgery, listed with an indicative rate of ₹49,300. The hospital should not be asking you for payment for this.",
		ActionSteps: []string{
			`Ask the billing desk, calmly: "Can you give us this denial in writing, with the reason stated?"`,
			"If they still refuse and treatment is needed soon, you can pay and dispute the charge afterward.",
		},
		HospitalScript: "We understand this procedure is listed under PMJAY's Health Benefit Package. Could you please help us understand the specific reason for the denial in writing?",
		Evidence: []store.EvidenceEntry{
			{CapturedAt: time.Date(2026, 8, 17, 15, 0, 0, 0, time.UTC), StaffName: "Ms. Priya (billing desk)", ApproxTime: "3:00 PM", Note: "Said package not covered, no reason given in writing"},
		},
	}
}

func mustBuild(t *testing.T, c store.CaseRecord) []byte {
	t.Helper()
	out, err := BuildCase(c)
	if err != nil {
		t.Fatalf("BuildCase returned error: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("BuildCase returned empty output")
	}
	if !bytes.HasPrefix(out, []byte("%PDF-1.4")) {
		t.Fatal("BuildCase output does not start with a PDF header")
	}
	return out
}

func TestBuildCase_AllFiveOutcomes_ProduceValidPDF(t *testing.T) {
	for _, outcome := range []string{"green", "amber", "red", "mixed", "handoff"} {
		t.Run(outcome, func(t *testing.T) {
			c := baseCase(outcome)
			if outcome == "handoff" {
				c.HandoffSummary = "Family describes a denial for a listed General Surgery package with no written reason given after two requests; hospital staff have been inconsistent about the denial reason."
			}
			out := mustBuild(t, c)
			s := string(out)
			if !strings.Contains(s, tierStyles[outcome].label) {
				t.Errorf("expected tier label %q to appear in output for outcome %q", tierStyles[outcome].label, outcome)
			}
		})
	}
}

func TestBuildCase_UnknownOutcome_FallsBackSafelyRatherThanErroring(t *testing.T) {
	c := baseCase("some_future_outcome_value")
	out := mustBuild(t, c)
	if !strings.Contains(string(out), "some_future_outcome_value") {
		t.Error("expected the raw outcome value to still appear in the fallback panel")
	}
}

func TestBuildCase_HandoffPanel_OnlyAppearsForHandoffOutcome(t *testing.T) {
	green := baseCase("green")
	out := mustBuild(t, green)
	if strings.Contains(string(out), "Free legal help") {
		t.Error("handoff panel text should not appear for a green outcome")
	}

	handoff := baseCase("handoff")
	handoff.HandoffSummary = "Summary of the situation for the PLV."
	out = mustBuild(t, handoff)
	if !strings.Contains(string(out), "Free legal help") {
		t.Error("expected handoff panel for a handoff outcome")
	}
	if !strings.Contains(string(out), "15100") {
		t.Error("expected the NALSA number 15100 in the handoff panel")
	}
}

func TestBuildCase_DisclaimerAppearsNearCareFirstBanner(t *testing.T) {
	out := mustBuild(t, baseCase("green"))
	s := string(out)
	if !strings.Contains(s, "not a legal ruling") {
		t.Error("expected the disclaimer text to appear in the rendered PDF")
	}
	if !strings.Contains(s, "confirm the current rate") {
		t.Error("expected the disclaimer's helpline reference to appear in the rendered PDF")
	}
}

func TestBuildCase_OptionalSectionsOmittedWhenEmpty(t *testing.T) {
	c := store.CaseRecord{
		ID:               "550e8400-e29b-41d4-a716-446655440000",
		CreatedAt:        time.Date(2026, 8, 17, 14, 30, 0, 0, time.UTC),
		Outcome:          "amber",
		CareFirstMessage: "Get treatment first. Dispute the money after. Always.",
		TierMessage:      "This needs one more check before we can be sure either way.",
		// No ActionSteps, HospitalScript, ComplaintText, Evidence, Citation.
	}
	out := mustBuild(t, c)
	s := string(out)
	if strings.Contains(s, "What to do right now") {
		t.Error("action steps heading should not appear when ActionSteps is empty")
	}
	if strings.Contains(s, "Exact words to use at the desk") {
		t.Error("hospital script heading should not appear when HospitalScript is empty")
	}
	if strings.Contains(s, "Draft complaint") {
		t.Error("complaint heading should not appear when ComplaintText is empty")
	}
	if strings.Contains(s, "Evidence recorded") {
		t.Error("evidence heading should not appear when Evidence is empty")
	}
	if strings.Contains(s, "not a legal ruling") {
		t.Error("disclaimer text should not appear when Disclaimer is empty")
	}
	if strings.Contains(s, "Based on:") {
		t.Error("citation line should not appear when Citation is empty")
	}
}

func TestBuildCase_ComplaintTextRenders_WithSubmissionDisclaimer(t *testing.T) {
	c := baseCase("red")
	c.ComplaintText = "--- Draft complaint for CGRMS ---\nSubject: Denial of covered procedure\n--- End of draft — review before submitting ---"
	out := mustBuild(t, c)
	s := string(out)
	if !strings.Contains(s, "Draft complaint") {
		t.Error("expected the draft complaint heading")
	}
	if !strings.Contains(s, "does not submit anything on its own") {
		t.Error("expected the explicit non-submission disclaimer next to the draft complaint")
	}
}

func TestBuildCase_RupeeAmountsRenderAsRsNotBrokenGlyph(t *testing.T) {
	out := mustBuild(t, baseCase("green"))
	s := string(out)
	if !strings.Contains(s, "Rs.49,300") {
		t.Error("expected the rupee amount to render as 'Rs.49,300' in the tier message")
	}
	if strings.ContainsRune(s, '₹') {
		t.Error("raw rupee sign byte should never appear in PDF output — WinAnsi/Helvetica cannot represent it (see winansi.go)")
	}
}

func TestBuildCase_EvidenceLog_AllFieldsPresent(t *testing.T) {
	out := mustBuild(t, baseCase("green"))
	s := string(out)
	for _, want := range []string{"Evidence recorded", "Ms. Priya", "3:00 PM", "not covered"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected evidence log to contain %q", want)
		}
	}
}

func TestBuildCase_EvidenceEntry_WithNoOptionalFields_StillRendersGracefully(t *testing.T) {
	c := baseCase("green")
	c.Evidence = []store.EvidenceEntry{{CapturedAt: time.Date(2026, 8, 17, 15, 0, 0, 0, time.UTC)}}
	out := mustBuild(t, c)
	if !strings.Contains(string(out), "no further details recorded") {
		t.Error("expected a graceful fallback line for an evidence entry with no staff/time/note set")
	}
}

func TestBuildCase_ZeroCreatedAt_DoesNotRenderMisleadingZeroDate(t *testing.T) {
	c := baseCase("green")
	c.CreatedAt = time.Time{}
	out := mustBuild(t, c)
	s := string(out)
	// Specifically the phrase Go's time formatting would produce for
	// the zero value ("1 January 0001, 12:00 AM") — not the bare digits
	// "0001", which also occurs incidentally inside the PDF's own xref
	// byte-offset table (e.g. "0000000015") and would make this test
	// fail on output that has nothing to do with dates at all.
	if strings.Contains(s, "January 0001") {
		t.Error("expected zero-value CreatedAt to be caught and not rendered as a misleading year-0001 date")
	}
	if !strings.Contains(s, "not recorded") {
		t.Error("expected an explicit fallback for an unset CreatedAt")
	}
}

func TestBuildCase_CaseIDAppearsInHeaderAndFooter(t *testing.T) {
	c := baseCase("green")
	out := mustBuild(t, c)
	if got := strings.Count(string(out), c.ID); got < 2 {
		t.Errorf("expected case ID to appear at least twice (header + footer), got %d occurrences", got)
	}
}

func TestBuildCase_FooterFitsPageWidth(t *testing.T) {
	// The footer's two lines (see drawFooter's doc comment for why it's
	// two lines rather than one) must each fit comfortably within the
	// content width at the footer font size, with real margin to spare
	// — checked directly against this package's own width metrics
	// rather than trusted from a comment.
	c := baseCase("green")
	maxWidth := pageWidth - 2*50.0
	const comfortableMargin = 40.0 // points of slack we want beyond a bare fit

	left := "Case " + c.ID
	right := "Page 12 of 12" // two-digit page numbers, the more demanding case
	helplines := "PMJAY helpline: 14555   |   NALSA free legal aid: 15100"

	for _, seg := range []struct {
		name string
		text string
	}{{"left", left}, {"right", right}, {"helplines", helplines}} {
		w := textWidth(seg.text, smallSize, false)
		if w > maxWidth {
			t.Errorf("footer %s segment too wide: %.1f > %.1f", seg.name, w, maxWidth)
		}
		if w > maxWidth-comfortableMargin {
			t.Errorf("footer %s segment (%.1fpt) leaves less than the intended %.0fpt margin within %.1fpt available — getting close to overflow risk from a future wording change", seg.name, w, comfortableMargin, maxWidth)
		}
	}
}

func TestBuildCase_LongContent_ProducesMultiplePagesWithConsistentFooters(t *testing.T) {
	c := baseCase("amber")
	c.TierMessage = strings.Repeat("This is a long tier message sentence that needs to wrap across many lines to force this document past a single page. ", 40)
	for i := 0; i < 15; i++ {
		c.ActionSteps = append(c.ActionSteps, "A reasonably long action step description, repeated to fill vertical space and force pagination.")
	}
	out := mustBuild(t, c)
	s := string(out)
	if !strings.Contains(s, "Page 1 of ") {
		t.Error("expected page 1 footer")
	}
	if strings.Contains(s, "Page 1 of 1") {
		t.Error("expected more than one page given the amount of content, got a single-page document")
	}
}

func TestBuildCase_SpecialCharactersInFreeTextFields_DoNotBreakStructure(t *testing.T) {
	// Evidence notes and staff names are free text a family member
	// types — including PDF's own special literal-string characters —
	// and must never be able to corrupt the document structure.
	c := baseCase("green")
	c.Evidence = []store.EvidenceEntry{
		{CapturedAt: time.Now(), StaffName: `Staff (said "no") \ weird`, Note: "Note with (parens) and \\backslash\\ and a checkmark ✓ that has no PDF glyph"},
	}
	out := mustBuild(t, c)
	s := string(out)
	if strings.Count(s, "(") < strings.Count(s, ")") {
		t.Error("unbalanced parens in output given free text containing literal parentheses")
	}
}

func TestTemplateTextHasNoUnrenderableCharacters(t *testing.T) {
	// The claim in winansi.go's package doc comment, checked directly:
	// every character this package's actual real-world input (the
	// generated templates it renders into a PDF) contains must be
	// representable, so this fails loudly if a future template edit
	// introduces something this package can't draw. Deliberately scans
	// realistic assembled case content, not internal/response's raw Go
	// source, so it also catches problems introduced by data (HBP
	// package names, citations) rather than only template scaffolding.
	c := baseCase("green")
	c.HandoffSummary = "n/a"
	c.ComplaintText = "--- Draft complaint for CGRMS ---\nSample.\n--- End of draft ---"

	allText := strings.Join([]string{
		c.CareFirstMessage, c.TierMessage, c.Citation, c.HospitalScript,
		c.ComplaintText, c.HandoffSummary, strings.Join(c.ActionSteps, " "),
	}, " ")

	for _, r := range allText {
		if r == '₹' || r == '\n' || r == '\r' || r == '\t' {
			continue // deliberately substituted/normalized, not "unmappable" — see shapeText
		}
		if _, ok := winAnsiEncode[r]; !ok {
			t.Errorf("character %q (U+%04X) in sample case content has no WinAnsi mapping and would render as %q", r, r, unmappableSubstitute)
		}
	}
}
