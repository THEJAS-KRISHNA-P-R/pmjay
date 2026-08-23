package hbp

import (
	"strings"
	"testing"
)

func TestLoad_Succeeds(t *testing.T) {
	ds, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if ds == nil {
		t.Fatal("Load() returned nil dataset with nil error")
	}
}

func TestLoad_PackagesNonEmptyAndReasonableSize(t *testing.T) {
	ds := MustLoad()
	if len(ds.Packages) < 10 {
		t.Errorf("expected a representative spread of packages, got only %d", len(ds.Packages))
	}
	// Original ceiling was 200, sized for the first build's ~40-record seed
	// set (see docs/DATA_SOURCES.md history). Raised to 400 during the
	// August 2026 continuation session, which added 250 reviewed records
	// with individual source citations (General Medicine + Pediatric
	// per-diem diagnoses, Medical Oncology, three new "High End" specialty
	// categories, Interventional Neuroradiology) — see
	// transform_2022_additions.py and docs/DATA_SOURCES.md for exactly
	// where every one of those came from. The guard-rail spirit is
	// unchanged: if this count grows again, was the growth reviewed for
	// provenance, the same way this one was?
	if len(ds.Packages) > 400 {
		t.Errorf("package count %d exceeds the documented dataset scope; if this grew, was it reviewed for provenance?", len(ds.Packages))
	}
}

func TestLoad_EveryPackageHasRequiredFields(t *testing.T) {
	ds := MustLoad()
	for _, p := range ds.Packages {
		if p.PackageCode == "" {
			t.Errorf("package %q has empty code", p.PackageName)
		}
		if len(p.CommonDescriptionKeywords) == 0 && p.PackageCode != "UNSPECIFIED" {
			t.Errorf("package %q (%s) has no description keywords — retrieval can never surface it", p.PackageName, p.PackageCode)
		}
		if p.SourceNote == "" {
			t.Errorf("package %q: every record must be traceable (Appendix F) but source_note is empty", p.PackageName)
		}
	}
}

func TestLoad_UnverifiedRecordsAreHonestlyFlagged(t *testing.T) {
	// This test exists to keep the dataset honest over time: it will fail
	// loudly if someone flips verified=true on a record without the
	// source_note actually describing independent verification. It is a
	// deliberately weak heuristic (keyword check), not a substitute for a
	// human actually checking the NHA source — see docs/DATA_SOURCES.md.
	ds := MustLoad()
	for _, p := range ds.Packages {
		if p.Verified {
			lower := strings.ToLower(p.SourceNote)
			if !strings.Contains(lower, "confirm") && !strings.Contains(lower, "verif") {
				t.Errorf("package %q is marked verified=true but source_note doesn't describe verification: %q", p.PackageCode, p.SourceNote)
			}
		}
	}
	for _, e := range ds.Exclusions {
		if e.Verified {
			lower := strings.ToLower(e.SourceNote)
			if !strings.Contains(lower, "confirm") && !strings.Contains(lower, "verif") {
				t.Errorf("exclusion %q is marked verified=true but source_note doesn't describe verification: %q", e.Category, e.SourceNote)
			}
		}
	}
}

func TestLoad_UnspecifiedProcedureCategoryPresent(t *testing.T) {
	// Section 8, point 2: recognising the discretionary "unspecified
	// procedure" catch-all is one of the three reasons this product isn't
	// just a lookup wrapper. If this record goes missing, that capability
	// silently disappears.
	ds := MustLoad()
	found := false
	for _, p := range ds.Packages {
		if p.PackageCode == "UNSPECIFIED" {
			found = true
			if p.IndicativeRateINR != UnspecifiedProcedureCapINR {
				t.Errorf("UNSPECIFIED package rate %d does not match the verified cap constant %d", p.IndicativeRateINR, UnspecifiedProcedureCapINR)
			}
			if !p.Verified {
				t.Error("UNSPECIFIED package cap was independently confirmed during this build and should be marked verified")
			}
		}
	}
	if !found {
		t.Fatal("no package with code UNSPECIFIED found — the discretionary catch-all category is missing")
	}
}

func TestLoad_AllFourConfirmedExclusionCategoriesPresent(t *testing.T) {
	// Section 6.6 names four confirmed exclusion categories specifically.
	// The red tier's credibility (Section 9: "not optional, not a lesser
	// feature") depends on all four actually being in the dataset.
	ds := MustLoad()
	want := map[string]bool{
		"opd_only":                 false,
		"cosmetic":                 false,
		"fertility":                false,
		"organ_transplant_partial": false,
	}
	for _, e := range ds.Exclusions {
		if _, ok := want[e.Category]; ok {
			want[e.Category] = true
		}
	}
	for cat, found := range want {
		if !found {
			t.Errorf("confirmed exclusion category %q from spec Section 6.6 is missing from the dataset", cat)
		}
	}
}

func TestLoad_OrganTransplantHasNuanceNotFlatExclusion(t *testing.T) {
	// Section 6.6: organ transplant is only *partially* covered, state-
	// dependent. A flat red-tier answer here would be an overclaim in the
	// wrong direction — exactly the failure mode Section 10 exists to
	// prevent. The nuance field is what lets internal/tiering route this
	// to amber instead of red.
	ds := MustLoad()
	for _, e := range ds.Exclusions {
		if e.Category == "organ_transplant_partial" {
			if e.Nuance == "" {
				t.Error("organ_transplant_partial exclusion has no nuance text — tiering cannot distinguish it from an unconditional exclusion")
			}
			return
		}
	}
	t.Fatal("organ_transplant_partial exclusion not found")
}

func TestLoad_NoDuplicatePackageCodes(t *testing.T) {
	ds := MustLoad()
	seen := map[string]bool{}
	for _, p := range ds.Packages {
		if seen[p.PackageCode] {
			t.Errorf("duplicate package code: %s", p.PackageCode)
		}
		seen[p.PackageCode] = true
	}
}
