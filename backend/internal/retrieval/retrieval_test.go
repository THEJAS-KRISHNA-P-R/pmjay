package retrieval

import (
	"testing"

	"github.com/pmjay-advocate/backend/internal/hbp"
)

func testDataset(t *testing.T) *hbp.Dataset {
	t.Helper()
	ds, err := hbp.Load()
	if err != nil {
		t.Fatalf("failed to load real embedded dataset for retrieval tests: %v", err)
	}
	return ds
}

func TestRetrieve_ObviousKeywordMatchRanksFirst(t *testing.T) {
	ds := testDataset(t)
	candidates := Retrieve(ds, "doctor says my mother's gallbladder has stones and needs operation")
	if len(candidates) == 0 {
		t.Fatal("expected at least one candidate")
	}
	if candidates[0].Package.PackageCode != "SEED-GS-001" {
		t.Errorf("expected Laparoscopic Cholecystectomy (SEED-GS-001) to rank first, got %s (%s) with score %d",
			candidates[0].Package.PackageName, candidates[0].Package.PackageCode, candidates[0].Score)
	}
}

func TestRetrieve_CardiacStentMatches(t *testing.T) {
	ds := testDataset(t)
	candidates := Retrieve(ds, "father collapsed, doctors say heart, need urgent stent, hospital says card wont cover the stent")
	found := false
	for _, c := range candidates[:min(5, len(candidates))] {
		if c.Package.PackageCode == "SEED-CARD-001" {
			found = true
		}
	}
	if !found {
		t.Error("expected PTCA with stent package in top 5 candidates for a stent-related description")
	}
}

func TestRetrieve_AlwaysIncludesUnspecifiedCatchAll(t *testing.T) {
	ds := testDataset(t)
	descriptions := []string{
		"gallbladder stones operation",
		"completely unrelated gibberish xyz123 blah",
		"",
	}
	for _, desc := range descriptions {
		candidates := Retrieve(ds, desc)
		found := false
		for _, c := range candidates {
			if c.Package.PackageCode == "UNSPECIFIED" {
				found = true
			}
		}
		if !found {
			t.Errorf("description %q: UNSPECIFIED catch-all missing from candidates", desc)
		}
	}
}

func TestRetrieve_NeverReturnsEmptyForNonEmptyDataset(t *testing.T) {
	// The fail-open guarantee: a description that matches nothing must
	// still hand the LLM *something* to reason about, never an empty
	// list that would force a silent "no coverage" outcome.
	ds := testDataset(t)
	candidates := Retrieve(ds, "asdkjhqwe zxcvpoiuy nonsense not medical at all 98765")
	if len(candidates) == 0 {
		t.Fatal("Retrieve returned zero candidates for a non-empty dataset — this violates the fail-open guarantee")
	}
}

func TestRetrieve_EmptyDescriptionDoesNotPanic(t *testing.T) {
	ds := testDataset(t)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Retrieve panicked on empty description: %v", r)
		}
	}()
	candidates := Retrieve(ds, "")
	if len(candidates) == 0 {
		t.Error("expected fail-open candidates even for empty description")
	}
}

func TestRetrieve_RespectsMaxCandidatesBound(t *testing.T) {
	ds := testDataset(t)
	// A description engineered to hit many packages' generic specialty
	// words shouldn't blow past the bound.
	candidates := Retrieve(ds, "surgery operation procedure treatment hospital doctor patient")
	if len(candidates) > MaxCandidates+1 { // +1 tolerance for guaranteed Unspecified append
		t.Errorf("candidate count %d exceeds MaxCandidates=%d bound (with Unspecified tolerance)", len(candidates), MaxCandidates)
	}
}

func TestRetrieve_ZeroSignalFallbackUsesWiderBoundThanNormal(t *testing.T) {
	// Nonsense ASCII tokens that don't match anything real still take the
	// zero-signal path (every package scores 0), same as native-script
	// input does — this just avoids needing non-ASCII literals in this
	// specific assertion. The bound itself should be MaxCandidatesWhenBlind,
	// not the tighter MaxCandidates used when scoring actually found signal.
	ds := testDataset(t)
	candidates := Retrieve(ds, "zzqxjk vvbnmq wwpldf")
	if len(candidates) <= MaxCandidates {
		t.Errorf("expected the zero-signal fallback to use the wider MaxCandidatesWhenBlind=%d bound, got only %d candidates (MaxCandidates=%d)",
			MaxCandidatesWhenBlind, len(candidates), MaxCandidates)
	}
	if len(candidates) > MaxCandidatesWhenBlind+1 { // +1 tolerance for guaranteed Unspecified append
		t.Errorf("candidate count %d exceeds MaxCandidatesWhenBlind=%d bound (with Unspecified tolerance)", len(candidates), MaxCandidatesWhenBlind)
	}
}

func TestRetrieve_NativeScriptDescriptionNowFindsAPreviouslyMissedPackage(t *testing.T) {
	// Real regression test for the gap this fix narrows: MG025A
	// ("Recurrent vomiting with dehydration") sits at index 70 in the
	// embedded dataset — inside the widened MaxCandidatesWhenBlind=80
	// bound but outside the old MaxCandidates=20 one. A description of
	// exactly this situation, in Malayalam with no embedded English or
	// digits, used to tokenize to nothing and silently miss this package;
	// it's included from here on. (The phrase below is illustrative, not
	// a verified professional translation — it exists to be genuine
	// non-ASCII script, not to be exact clinical Malayalam.)
	//
	// This narrows the gap, it does not close it: a package further into
	// the file than MaxCandidatesWhenBlind is still unreachable from a
	// zero-signal description, which on the real ~1,900-code dataset is
	// most of it. See MaxCandidatesWhenBlind's doc comment for what
	// actually closes this.
	ds := testDataset(t)
	description := "എന്റെ കുട്ടിക്ക് ഛർദ്ദിയും വരൾച്ചയും ഉണ്ട്"
	if toks := tokenize(description); len(toks) != 0 {
		t.Fatalf("test assumption violated: expected this description to tokenize to nothing, got %#v", toks)
	}

	candidates := Retrieve(ds, description)
	found := false
	for _, c := range candidates {
		if c.Package.PackageCode == "MG025A" {
			found = true
		}
	}
	if !found {
		t.Error("expected MG025A to be reachable under the widened blind-fallback bound")
	}
}

func TestRetrieve_CaseAndPunctuationInsensitive(t *testing.T) {
	ds := testDataset(t)
	a := Retrieve(ds, "GALLBLADDER stones!!! operation???")
	b := Retrieve(ds, "gallbladder stones operation")
	if len(a) == 0 || len(b) == 0 {
		t.Fatal("expected non-empty candidates for both variants")
	}
	if a[0].Package.PackageCode != b[0].Package.PackageCode {
		t.Errorf("case/punctuation changed top match: %s vs %s", a[0].Package.PackageCode, b[0].Package.PackageCode)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
