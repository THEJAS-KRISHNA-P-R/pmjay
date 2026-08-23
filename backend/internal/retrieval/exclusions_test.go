package retrieval

import (
	"testing"

	"github.com/pmjay-advocate/backend/internal/hbp"
)

// RetrieveExclusions and scoreExclusion were previously exercised only
// indirectly, via internal/api's handler tests calling this package as a
// library — real coverage, but not coverage this package's own test suite
// showed, and fragile: if api's tests were ever refactored to stop
// touching a real Dataset, this package would silently lose its only
// exercise of these two functions. Testing them directly here, the way
// TestRetrieve_* already does for Retrieve/scorePackage, doesn't just
// change a coverage number — it makes this package's own test suite
// actually stand on its own.

func testExclusionDataset() *hbp.Dataset {
	return &hbp.Dataset{
		Packages: []hbp.Package{{
			PackageCode: "P1", PackageName: "x", Specialty: "x",
			IndicativeRateINR: 100, CommonDescriptionKeywords: []string{"x"}, SourceNote: "x",
		}},
		Exclusions: []hbp.Exclusion{
			{
				Category: "COSMETIC", DisplayName: "Cosmetic Procedures",
				Description: "Purely cosmetic procedures with no functional medical necessity.",
				Keywords:    []string{"cosmetic surgery", "hair transplant", "botox"},
				SourceNote:  "x",
			},
			{
				Category: "WAITING_PERIOD", DisplayName: "Waiting Period Not Met",
				Description: "Pre-existing condition within the scheme's waiting period.",
				Keywords:    []string{"pre-existing", "waiting period"},
				SourceNote:  "x",
			},
		},
	}
}

func TestRetrieveExclusions_KeywordMatchRanksAboveNoMatch(t *testing.T) {
	ds := testExclusionDataset()
	candidates := RetrieveExclusions(ds, "hospital denied our claim, said it's a hair transplant not covered")

	if len(candidates) != len(ds.Exclusions) {
		t.Fatalf("expected every exclusion category to be returned (the LLM step always sees the complete list — see RetrieveExclusions's doc comment), got %d of %d", len(candidates), len(ds.Exclusions))
	}
	if candidates[0].Exclusion.Category != "COSMETIC" {
		t.Errorf("expected the keyword-matching category (COSMETIC, via 'hair transplant') to rank first, got %s", candidates[0].Exclusion.Category)
	}
	if candidates[0].Score <= candidates[1].Score {
		t.Errorf("expected the matching category to score strictly higher than the non-matching one: COSMETIC=%d, WAITING_PERIOD=%d", candidates[0].Score, candidates[1].Score)
	}
}

func TestRetrieveExclusions_ReturnsAllCategoriesEvenWithNoMatch(t *testing.T) {
	// The doc comment on RetrieveExclusions is explicit that this list is
	// never filtered down, only sorted — a description with no exclusion
	// keywords at all must still return every category, all scored zero.
	ds := testExclusionDataset()
	candidates := RetrieveExclusions(ds, "completely unrelated words about nothing exclusion-shaped at all")

	if len(candidates) != len(ds.Exclusions) {
		t.Fatalf("expected all %d exclusion categories regardless of match, got %d", len(ds.Exclusions), len(candidates))
	}
	for _, c := range candidates {
		if c.Score != 0 {
			t.Errorf("expected zero score for category %s with no keyword overlap, got %d", c.Exclusion.Category, c.Score)
		}
	}
}

func TestRetrieveExclusions_EmptyDescriptionDoesNotPanic(t *testing.T) {
	ds := testExclusionDataset()
	candidates := RetrieveExclusions(ds, "")
	if len(candidates) != len(ds.Exclusions) {
		t.Errorf("expected all exclusion categories back even for an empty description, got %d", len(candidates))
	}
}

func TestScoreExclusion_MultiWordKeywordRequiresAllTokensPresent(t *testing.T) {
	e := hbp.Exclusion{Category: "C", DisplayName: "x", Keywords: []string{"hair transplant"}}

	full := scoreExclusion(e, tokenSetFor("i need a hair transplant"))
	partial := scoreExclusion(e, tokenSetFor("i need a hair cut"))

	if full == 0 {
		t.Error("expected a non-zero score when all tokens of a multi-word keyword are present")
	}
	if partial != 0 {
		t.Errorf("expected a zero score when only some tokens of a multi-word keyword are present, got %d", partial)
	}
}

func TestScoreExclusion_EmptyKeywordAfterTokenizationIsSkipped(t *testing.T) {
	// A keyword that tokenizes to nothing (pure punctuation) must not
	// contribute to the score or be treated as trivially "matched" —
	// scoreExclusion's len(kwTokens) == 0 guard exists specifically for
	// this, mirroring scorePackage's identical guard.
	e := hbp.Exclusion{Category: "C", DisplayName: "x", Keywords: []string{"---", "  "}}
	score := scoreExclusion(e, tokenSetFor("some completely unrelated description"))
	if score != 0 {
		t.Errorf("expected an empty-after-tokenization keyword to contribute nothing, got score %d", score)
	}
}

func TestScoreExclusion_DisplayNameOverlapContributesButStopwordsDont(t *testing.T) {
	e := hbp.Exclusion{Category: "C", DisplayName: "Cosmetic Surgery for the Face"}

	withOverlap := scoreExclusion(e, tokenSetFor("this was purely cosmetic"))
	withoutOverlap := scoreExclusion(e, tokenSetFor("totally unrelated medical emergency"))

	if withOverlap <= withoutOverlap {
		t.Errorf("expected display-name word overlap to score higher than no overlap: with=%d without=%d", withOverlap, withoutOverlap)
	}
}

func tokenSetFor(description string) map[string]bool {
	tokens := tokenize(description)
	set := make(map[string]bool, len(tokens))
	for _, t := range tokens {
		set[t] = true
	}
	return set
}

func TestScorePackage_EmptyKeywordAfterTokenizationIsSkipped(t *testing.T) {
	p := hbp.Package{PackageName: "x", CommonDescriptionKeywords: []string{"---", "***"}}
	score := scorePackage(p, tokenSetFor("some unrelated description"))
	if score != 0 {
		t.Errorf("expected an empty-after-tokenization package keyword to contribute nothing, got score %d", score)
	}
}

func TestEnsureUnspecifiedIncluded_DatasetWithoutUnspecifiedPackageIsUnchanged(t *testing.T) {
	// Every real dataset has an UNSPECIFIED catch-all (see
	// TestRetrieve_AlwaysIncludesUnspecifiedCatchAll against the real
	// data) — this test covers the defensive fallback for a dataset that
	// doesn't, so a future dataset change can't silently make this
	// function panic or behave surprisingly.
	ds := &hbp.Dataset{Packages: []hbp.Package{{PackageCode: "ONLY-ONE", PackageName: "x"}}}
	candidates := []Candidate{{Package: ds.Packages[0], Score: 5}}

	got := ensureUnspecifiedIncluded(candidates, ds)
	if len(got) != 1 {
		t.Errorf("expected candidates to pass through unchanged when no UNSPECIFIED package exists in the dataset, got %d entries", len(got))
	}
}
