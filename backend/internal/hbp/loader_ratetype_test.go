package hbp

import "testing"

// These tests cover the "tiered" and "per_diem" RateType extension added
// this session (see docs/DATA_SOURCES.md) — additive to the schema all 40
// original records use, so loader_test.go's existing coverage is untouched
// and still exercises the RateType=="" path implicitly via those records.

func TestValidate_TieredRateRequiresMaxGreaterThanFloor(t *testing.T) {
	ds := &Dataset{
		Packages: []Package{{
			PackageCode: "T1", PackageName: "x", Specialty: "x",
			IndicativeRateINR: 100, RateType: "tiered", RateMaxINR: 100,
			CommonDescriptionKeywords: []string{"x"}, SourceNote: "verified test fixture",
		}},
		Exclusions: []Exclusion{{Category: "x", DisplayName: "x", Description: "x", SourceNote: "x"}},
	}
	if err := validate(ds); err == nil {
		t.Error("expected validation error when rate_max_inr does not exceed indicative_rate_inr for a tiered package")
	}
}

func TestValidate_TieredRateAcceptsValidRange(t *testing.T) {
	ds := &Dataset{
		Packages: []Package{{
			PackageCode: "T2", PackageName: "x", Specialty: "x",
			IndicativeRateINR: 100, RateType: "tiered", RateMaxINR: 150,
			CommonDescriptionKeywords: []string{"x"}, SourceNote: "verified test fixture",
		}},
		Exclusions: []Exclusion{{Category: "x", DisplayName: "x", Description: "x", SourceNote: "x"}},
	}
	if err := validate(ds); err != nil {
		t.Errorf("expected a valid tiered range to pass validation, got: %v", err)
	}
}

func TestValidate_PerDiemRequiresPerDiemRates(t *testing.T) {
	ds := &Dataset{
		Packages: []Package{{
			PackageCode: "P1", PackageName: "x", Specialty: "x",
			IndicativeRateINR: 100, RateType: "per_diem", PerDiemRates: nil,
			CommonDescriptionKeywords: []string{"x"}, SourceNote: "verified test fixture",
		}},
		Exclusions: []Exclusion{{Category: "x", DisplayName: "x", Description: "x", SourceNote: "x"}},
	}
	if err := validate(ds); err == nil {
		t.Error("expected validation error when rate_type is per_diem but per_diem_rates is nil")
	}
}

func TestValidate_PerDiemRoutineWardMustMatchIndicativeRate(t *testing.T) {
	ds := &Dataset{
		Packages: []Package{{
			PackageCode: "P2", PackageName: "x", Specialty: "x",
			IndicativeRateINR: 100, RateType: "per_diem",
			PerDiemRates:              &PerDiemRates{RoutineWardINR: 999, HDUINR: 1000, ICUNoVentINR: 1001, ICUVentINR: 1002},
			CommonDescriptionKeywords: []string{"x"}, SourceNote: "verified test fixture",
		}},
		Exclusions: []Exclusion{{Category: "x", DisplayName: "x", Description: "x", SourceNote: "x"}},
	}
	if err := validate(ds); err == nil {
		t.Error("expected validation error when per_diem_rates.routine_ward_inr does not match indicative_rate_inr")
	}
}

func TestValidate_PerDiemLevelsMustStrictlyIncrease(t *testing.T) {
	cases := []PerDiemRates{
		{RoutineWardINR: 100, HDUINR: 100, ICUNoVentINR: 200, ICUVentINR: 300}, // hdu not > routine
		{RoutineWardINR: 100, HDUINR: 200, ICUNoVentINR: 150, ICUVentINR: 300}, // icu_no_vent not > hdu
		{RoutineWardINR: 100, HDUINR: 200, ICUNoVentINR: 300, ICUVentINR: 250}, // icu_vent not > icu_no_vent
	}
	for i, pd := range cases {
		pd := pd
		ds := &Dataset{
			Packages: []Package{{
				PackageCode: "P3", PackageName: "x", Specialty: "x",
				IndicativeRateINR: pd.RoutineWardINR, RateType: "per_diem", PerDiemRates: &pd,
				CommonDescriptionKeywords: []string{"x"}, SourceNote: "verified test fixture",
			}},
			Exclusions: []Exclusion{{Category: "x", DisplayName: "x", Description: "x", SourceNote: "x"}},
		}
		if err := validate(ds); err == nil {
			t.Errorf("case %d: expected validation error for non-increasing per-diem levels %+v", i, pd)
		}
	}
}

func TestValidate_PerDiemValidRecordPasses(t *testing.T) {
	ds := &Dataset{
		Packages: []Package{{
			PackageCode: "P4", PackageName: "x", Specialty: "x",
			IndicativeRateINR: 2300, RateType: "per_diem",
			PerDiemRates:              &PerDiemRates{RoutineWardINR: 2300, HDUINR: 3630, ICUNoVentINR: 9350, ICUVentINR: 9900},
			CommonDescriptionKeywords: []string{"x"}, SourceNote: "verified test fixture",
		}},
		Exclusions: []Exclusion{{Category: "x", DisplayName: "x", Description: "x", SourceNote: "x"}},
	}
	if err := validate(ds); err != nil {
		t.Errorf("expected a well-formed per-diem record to pass validation, got: %v", err)
	}
}

func TestValidate_UnrecognisedRateTypeRejected(t *testing.T) {
	ds := &Dataset{
		Packages: []Package{{
			PackageCode: "P5", PackageName: "x", Specialty: "x",
			IndicativeRateINR: 100, RateType: "made_up",
			CommonDescriptionKeywords: []string{"x"}, SourceNote: "verified test fixture",
		}},
		Exclusions: []Exclusion{{Category: "x", DisplayName: "x", Description: "x", SourceNote: "x"}},
	}
	if err := validate(ds); err == nil {
		t.Error("expected validation error for an unrecognised rate_type value")
	}
}
