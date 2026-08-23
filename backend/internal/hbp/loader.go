package hbp

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed data/hbp_packages.json data/exclusions.json
var dataFS embed.FS

// Load parses the embedded HBP package list and exclusion list into a
// Dataset, validating structural invariants along the way.
//
// Validation is deliberately strict and fails fast at startup rather than
// letting a malformed record surface later as a wrong or uncited answer to
// a family — a bad dataset should be a deploy-time failure, not a
// production incident. This is the "no untraceable record" guarantee the
// spec's Explainability axis (Section 11) and Appendix F depend on: every
// record must carry a non-empty SourceNote, because "where does that come
// from" must always have a real answer.
func Load() (*Dataset, error) {
	pkgBytes, err := dataFS.ReadFile("data/hbp_packages.json")
	if err != nil {
		return nil, fmt.Errorf("hbp: reading embedded package data: %w", err)
	}
	var packages []Package
	if err := json.Unmarshal(pkgBytes, &packages); err != nil {
		return nil, fmt.Errorf("hbp: parsing embedded package data: %w", err)
	}

	exclBytes, err := dataFS.ReadFile("data/exclusions.json")
	if err != nil {
		return nil, fmt.Errorf("hbp: reading embedded exclusion data: %w", err)
	}
	var exclusions []Exclusion
	if err := json.Unmarshal(exclBytes, &exclusions); err != nil {
		return nil, fmt.Errorf("hbp: parsing embedded exclusion data: %w", err)
	}

	ds := &Dataset{Packages: packages, Exclusions: exclusions}
	if err := validate(ds); err != nil {
		return nil, fmt.Errorf("hbp: dataset failed validation: %w", err)
	}
	return ds, nil
}

// MustLoad is Load, panicking on error. Used only at process start in
// main.go, where a bad embedded dataset is a build defect, not a runtime
// condition to recover from.
func MustLoad() *Dataset {
	ds, err := Load()
	if err != nil {
		panic(err)
	}
	return ds
}

func validate(ds *Dataset) error {
	if len(ds.Packages) == 0 {
		return fmt.Errorf("zero packages loaded")
	}
	if len(ds.Exclusions) == 0 {
		return fmt.Errorf("zero exclusions loaded")
	}

	seenCodes := make(map[string]bool, len(ds.Packages))
	for i, p := range ds.Packages {
		if p.PackageCode == "" {
			return fmt.Errorf("package[%d] (%q): empty package_code", i, p.PackageName)
		}
		if seenCodes[p.PackageCode] {
			return fmt.Errorf("package[%d]: duplicate package_code %q", i, p.PackageCode)
		}
		seenCodes[p.PackageCode] = true

		if p.PackageName == "" {
			return fmt.Errorf("package[%d] (%s): empty package_name", i, p.PackageCode)
		}
		if p.Specialty == "" {
			return fmt.Errorf("package %q: empty specialty", p.PackageCode)
		}
		if p.IndicativeRateINR <= 0 {
			return fmt.Errorf("package %q: indicative_rate_inr must be positive, got %d", p.PackageCode, p.IndicativeRateINR)
		}
		if p.SourceNote == "" {
			return fmt.Errorf("package %q: empty source_note (every record must be traceable — Appendix F)", p.PackageCode)
		}

		switch p.RateType {
		case "":
			// Flat single-rate package — no additional invariants.
		case "tiered":
			if p.RateMaxINR <= p.IndicativeRateINR {
				return fmt.Errorf("package %q: rate_type is \"tiered\" but rate_max_inr (%d) is not greater than indicative_rate_inr (%d)", p.PackageCode, p.RateMaxINR, p.IndicativeRateINR)
			}
		case "per_diem":
			if p.PerDiemRates == nil {
				return fmt.Errorf("package %q: rate_type is \"per_diem\" but per_diem_rates is missing", p.PackageCode)
			}
			pd := *p.PerDiemRates
			if pd.RoutineWardINR != p.IndicativeRateINR {
				return fmt.Errorf("package %q: per_diem_rates.routine_ward_inr (%d) must match indicative_rate_inr (%d) so non-per-diem-aware code sees a consistent figure", p.PackageCode, pd.RoutineWardINR, p.IndicativeRateINR)
			}
			if !(pd.RoutineWardINR < pd.HDUINR && pd.HDUINR < pd.ICUNoVentINR && pd.ICUNoVentINR < pd.ICUVentINR) {
				return fmt.Errorf("package %q: per_diem_rates levels must strictly increase (routine < hdu < icu_no_vent < icu_vent), got %d/%d/%d/%d", p.PackageCode, pd.RoutineWardINR, pd.HDUINR, pd.ICUNoVentINR, pd.ICUVentINR)
			}
		default:
			return fmt.Errorf("package %q: unrecognised rate_type %q (want \"\", \"tiered\", or \"per_diem\")", p.PackageCode, p.RateType)
		}
	}

	seenCat := make(map[string]bool, len(ds.Exclusions))
	for i, e := range ds.Exclusions {
		if e.Category == "" {
			return fmt.Errorf("exclusion[%d]: empty category", i)
		}
		if seenCat[e.Category] {
			return fmt.Errorf("exclusion[%d]: duplicate category %q", i, e.Category)
		}
		seenCat[e.Category] = true

		if e.DisplayName == "" {
			return fmt.Errorf("exclusion %q: empty display_name", e.Category)
		}
		if e.Description == "" {
			return fmt.Errorf("exclusion %q: empty description", e.Category)
		}
		if e.SourceNote == "" {
			return fmt.Errorf("exclusion %q: empty source_note (every record must be traceable — Appendix F)", e.Category)
		}
	}

	return nil
}
