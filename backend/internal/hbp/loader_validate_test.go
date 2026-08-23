package hbp

import "testing"

// These tests cover validate()'s base field-level checks — the ones every
// record goes through regardless of RateType (see loader_ratetype_test.go
// for the RateType-specific ones). Before this file, these branches were
// only exercised incidentally by loader_test.go's TestLoad_* tests running
// against the real embedded dataset, which — being valid — never actually
// triggers validate()'s error paths. Testing "does the real dataset avoid
// this problem" and "does validate() correctly catch this problem when it
// occurs" are different claims; this file tests the second one directly,
// the way loader_ratetype_test.go already does for RateType.

func validDataset() *Dataset {
	return &Dataset{
		Packages: []Package{{
			PackageCode: "P1", PackageName: "Test Package", Specialty: "Test",
			IndicativeRateINR: 100, CommonDescriptionKeywords: []string{"x"},
			SourceNote: "verified test fixture",
		}},
		Exclusions: []Exclusion{{
			Category: "E1", DisplayName: "Test Exclusion", Description: "x", SourceNote: "x",
		}},
	}
}

func TestValidate_WellFormedDatasetPasses(t *testing.T) {
	// The baseline this whole file's mutations are checked against —
	// if this ever fails, every other test below is meaningless, since
	// they all assume validDataset() itself passes.
	if err := validate(validDataset()); err != nil {
		t.Fatalf("expected the baseline valid dataset to pass, got: %v", err)
	}
}

func TestValidate_ZeroPackagesRejected(t *testing.T) {
	ds := validDataset()
	ds.Packages = nil
	if err := validate(ds); err == nil {
		t.Error("expected validation error for zero packages")
	}
}

func TestValidate_ZeroExclusionsRejected(t *testing.T) {
	ds := validDataset()
	ds.Exclusions = nil
	if err := validate(ds); err == nil {
		t.Error("expected validation error for zero exclusions")
	}
}

func TestValidate_EmptyPackageCodeRejected(t *testing.T) {
	ds := validDataset()
	ds.Packages[0].PackageCode = ""
	if err := validate(ds); err == nil {
		t.Error("expected validation error for empty package_code")
	}
}

func TestValidate_DuplicatePackageCodeRejected(t *testing.T) {
	ds := validDataset()
	dup := ds.Packages[0]
	dup.PackageCode = ds.Packages[0].PackageCode // same code, different record
	ds.Packages = append(ds.Packages, dup)
	if err := validate(ds); err == nil {
		t.Error("expected validation error for a duplicate package_code")
	}
}

func TestValidate_EmptyPackageNameRejected(t *testing.T) {
	ds := validDataset()
	ds.Packages[0].PackageName = ""
	if err := validate(ds); err == nil {
		t.Error("expected validation error for empty package_name")
	}
}

func TestValidate_EmptySpecialtyRejected(t *testing.T) {
	ds := validDataset()
	ds.Packages[0].Specialty = ""
	if err := validate(ds); err == nil {
		t.Error("expected validation error for empty specialty")
	}
}

func TestValidate_NonPositiveIndicativeRateRejected(t *testing.T) {
	for _, rate := range []int{0, -1, -100} {
		ds := validDataset()
		ds.Packages[0].IndicativeRateINR = rate
		if err := validate(ds); err == nil {
			t.Errorf("expected validation error for indicative_rate_inr = %d", rate)
		}
	}
}

func TestValidate_EmptyPackageSourceNoteRejected(t *testing.T) {
	// The specific check Appendix F depends on — every priced record
	// must be traceable to somewhere. Worth its own named test given how
	// much weight this session's own docs put on that guarantee.
	ds := validDataset()
	ds.Packages[0].SourceNote = ""
	if err := validate(ds); err == nil {
		t.Error("expected validation error for empty package source_note (Appendix F traceability)")
	}
}

func TestValidate_EmptyExclusionCategoryRejected(t *testing.T) {
	ds := validDataset()
	ds.Exclusions[0].Category = ""
	if err := validate(ds); err == nil {
		t.Error("expected validation error for empty exclusion category")
	}
}

func TestValidate_DuplicateExclusionCategoryRejected(t *testing.T) {
	ds := validDataset()
	dup := ds.Exclusions[0]
	dup.Category = ds.Exclusions[0].Category
	ds.Exclusions = append(ds.Exclusions, dup)
	if err := validate(ds); err == nil {
		t.Error("expected validation error for a duplicate exclusion category")
	}
}

func TestValidate_EmptyExclusionDisplayNameRejected(t *testing.T) {
	ds := validDataset()
	ds.Exclusions[0].DisplayName = ""
	if err := validate(ds); err == nil {
		t.Error("expected validation error for empty exclusion display_name")
	}
}

func TestValidate_EmptyExclusionDescriptionRejected(t *testing.T) {
	ds := validDataset()
	ds.Exclusions[0].Description = ""
	if err := validate(ds); err == nil {
		t.Error("expected validation error for empty exclusion description")
	}
}

func TestValidate_EmptyExclusionSourceNoteRejected(t *testing.T) {
	ds := validDataset()
	ds.Exclusions[0].SourceNote = ""
	if err := validate(ds); err == nil {
		t.Error("expected validation error for empty exclusion source_note (Appendix F traceability)")
	}
}

func TestDataset_PackageByCode_FoundAndNotFound(t *testing.T) {
	ds := validDataset()
	code := ds.Packages[0].PackageCode

	got, found := ds.PackageByCode(code)
	if !found {
		t.Fatalf("expected PackageByCode(%q) to find the package", code)
	}
	if got.PackageCode != code {
		t.Errorf("expected the returned package's code to be %q, got %q", code, got.PackageCode)
	}

	_, found = ds.PackageByCode("NONEXISTENT-CODE")
	if found {
		t.Error("expected PackageByCode to report not-found for a code that isn't in the dataset")
	}
}

func TestDataset_ExclusionByCategory_FoundAndNotFound(t *testing.T) {
	ds := validDataset()
	category := ds.Exclusions[0].Category

	got, found := ds.ExclusionByCategory(category)
	if !found {
		t.Fatalf("expected ExclusionByCategory(%q) to find the exclusion", category)
	}
	if got.Category != category {
		t.Errorf("expected the returned exclusion's category to be %q, got %q", category, got.Category)
	}

	_, found = ds.ExclusionByCategory("NONEXISTENT-CATEGORY")
	if found {
		t.Error("expected ExclusionByCategory to report not-found for a category that isn't in the dataset")
	}
}

func TestMustLoad_ReturnsRealDataset(t *testing.T) {
	// MustLoad's own panic branch is Load()'s error path one level up —
	// see loader.go's doc comment on why that specific branch is not
	// exercised here: it depends on the embedded filesystem failing,
	// which isn't something a test can trigger without editing the
	// actual embedded source files and rebuilding, not a legitimate
	// per-test technique. This test covers MustLoad's actual job (the
	// success path) directly rather than leaving it implicit.
	ds := MustLoad()
	if ds == nil {
		t.Fatal("expected MustLoad to return a non-nil dataset")
	}
	if len(ds.Packages) == 0 {
		t.Error("expected MustLoad's dataset to have packages")
	}
}
