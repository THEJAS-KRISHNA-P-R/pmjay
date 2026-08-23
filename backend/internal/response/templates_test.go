package response

import (
	"strings"
	"testing"

	"github.com/pmjay-advocate/backend/internal/hbp"
)

func TestFormatINR(t *testing.T) {
	cases := map[int]string{
		0:        "0",
		5:        "5",
		999:      "999",
		1000:     "1,000",
		9500:     "9,500",
		28500:    "28,500",
		100000:   "1,00,000",
		500000:   "5,00,000",
		1500000:  "15,00,000",
		12345678: "1,23,45,678",
	}
	for input, want := range cases {
		got := formatINR(input)
		if got != want {
			t.Errorf("formatINR(%d) = %q, want %q", input, got, want)
		}
	}
}

// TestPackageCitation_FlatRateUnchanged locks in the original rendering for
// RateType=="" so the extension added this session cannot silently change
// how any of the original 40 records are cited.
func TestPackageCitation_FlatRateUnchanged(t *testing.T) {
	p := hbp.Package{PackageName: "X Surgery", Specialty: "General Surgery", IndicativeRateINR: 28500, Verified: true}
	got := packageCitation(p)
	want := "X Surgery (General Surgery), listed with an indicative rate of ₹28,500"
	if got != want {
		t.Errorf("packageCitation() = %q, want %q", got, want)
	}
}

func TestPackageCitation_UnverifiedNeverCitesAnyRate(t *testing.T) {
	for _, p := range []hbp.Package{
		{PackageName: "X", Specialty: "Y", IndicativeRateINR: 100, Verified: false},
		{PackageName: "X", Specialty: "Y", IndicativeRateINR: 100, RateType: "tiered", RateMaxINR: 200, Verified: false},
		{PackageName: "X", Specialty: "Y", IndicativeRateINR: 100, RateType: "per_diem", PerDiemRates: &hbp.PerDiemRates{RoutineWardINR: 100, HDUINR: 200, ICUNoVentINR: 300, ICUVentINR: 400}, Verified: false},
	} {
		got := packageCitation(p)
		if strings.Contains(got, "₹") {
			t.Errorf("unverified package must never cite a rate figure, got: %q", got)
		}
	}
}

// TestPackageCitation_TieredShowsRangeNotSingleNumber is the core safety
// property for "tiered" packages: the citation must contain both the floor
// and the ceiling, not just one number picked without knowing the actual
// hospital's city-tier classification.
func TestPackageCitation_TieredShowsRangeNotSingleNumber(t *testing.T) {
	p := hbp.Package{
		PackageName: "PTCA", Specialty: "Cardiology",
		IndicativeRateINR: 50800, RateType: "tiered", RateMaxINR: 60900,
		Verified: true,
	}
	got := packageCitation(p)
	if !strings.Contains(got, "50,800") || !strings.Contains(got, "60,900") {
		t.Errorf("tiered citation must show both floor and ceiling, got: %q", got)
	}
	if !strings.Contains(got, "city-tier") && !strings.Contains(got, "hospital") {
		t.Errorf("tiered citation should explain why there's a range, got: %q", got)
	}
}

// TestPackageCitation_PerDiemNeverImpliesATotal is the single most
// important safety property added this session: a per-day rate must never
// be phrased in a way a family could read as "the total cost of this
// admission is ₹X" — that would be a confidently wrong number, exactly the
// failure mode Section 10 of the source spec exists to prevent.
func TestPackageCitation_PerDiemNeverImpliesATotal(t *testing.T) {
	p := hbp.Package{
		PackageName: "Acute febrile illness", Specialty: "General Medicine",
		IndicativeRateINR: 2300, RateType: "per_diem",
		PerDiemRates: &hbp.PerDiemRates{RoutineWardINR: 2300, HDUINR: 3630, ICUNoVentINR: 9350, ICUVentINR: 9900},
		Verified:     true,
	}
	got := packageCitation(p)
	for _, want := range []string{"2,300", "3,630", "9,350", "9,900", "per day"} {
		if !strings.Contains(got, want) {
			t.Errorf("per_diem citation missing %q, got: %q", want, got)
		}
	}
	// The exact original single-rate phrase must not appear standalone —
	// that phrasing reads as a flat total, which this rate is not.
	if strings.Contains(got, "listed with an indicative rate of ₹2,300") {
		t.Errorf("per_diem citation must not use the flat-rate phrasing that implies a single total: %q", got)
	}
}

func TestPackageCitation_TieredWithoutMaxFallsBackToFlatRendering(t *testing.T) {
	// Defensive case: a "tiered" record where RateMaxINR was left unset
	// (e.g. a data bug) must not silently render a broken/misleading
	// citation — it should fall back to the plain flat-rate phrasing
	// rather than claim a range that doesn't exist.
	p := hbp.Package{PackageName: "X", Specialty: "Y", IndicativeRateINR: 100, RateType: "tiered", Verified: true}
	got := packageCitation(p)
	want := "X (Y), listed with an indicative rate of ₹100"
	if got != want {
		t.Errorf("packageCitation() = %q, want %q", got, want)
	}
}
