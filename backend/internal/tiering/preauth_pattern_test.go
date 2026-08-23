package tiering

import "testing"

func TestDetect_PendingPhrasings(t *testing.T) {
	cases := []string{
		"the insurance company hasn't cleared it yet",
		"hospital says pre-authorisation is still pending",
		"they said it's still under review",
		"we are waiting for approval from the scheme",
		"nothing has been decided, still in process",
		"the clearance is pending, they told us",
		"it is currently being processed",
		"HOSPITAL SAYS IT'S STILL PENDING", // case insensitivity
	}
	for _, desc := range cases {
		got := Detect(desc)
		if got != PatternPendingLikely {
			t.Errorf("Detect(%q) = %v, want PatternPendingLikely", desc, got)
		}
	}
}

func TestDetect_DeniedFinalPhrasings(t *testing.T) {
	cases := []string{
		"the billing desk told us the card won't cover it",
		"they flatly rejected the claim",
		"hospital refused to accept the card for this",
		"we were told this is not covered, full stop",
		"they declined our request",
		"the hospital said no",
	}
	for _, desc := range cases {
		got := Detect(desc)
		if got != PatternDeniedFinal {
			t.Errorf("Detect(%q) = %v, want PatternDeniedFinal", desc, got)
		}
	}
}

func TestDetect_NoSignalReturnsNone(t *testing.T) {
	cases := []string{
		"my mother needs surgery for her gallbladder",
		"the doctor examined my father today",
		"",
		"we are at the hospital right now with my son",
	}
	for _, desc := range cases {
		got := Detect(desc)
		if got != PatternNone {
			t.Errorf("Detect(%q) = %v, want PatternNone", desc, got)
		}
	}
}

func TestDetect_ConflictingSignalsResolveToPending(t *testing.T) {
	// Appendix AA: bias conservatively. A description containing BOTH
	// pending and denied language is safer treated as "possibly still
	// pending" than as a confident final denial.
	desc := "they denied it at first but then said it's still pending review"
	got := Detect(desc)
	if got != PatternPendingLikely {
		t.Errorf("Detect(%q) = %v, want PatternPendingLikely (conservative bias on conflicting signals)", desc, got)
	}
}

func TestDetect_DoesNotPanicOnWeirdInput(t *testing.T) {
	weird := []string{"", "   ", "🏥💰❓", "a very very very " + string(make([]byte, 5000))}
	for _, desc := range weird {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Detect(%q) panicked: %v", desc, r)
				}
			}()
			Detect(desc)
		}()
	}
}
