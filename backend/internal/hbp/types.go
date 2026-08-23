// Package hbp holds the structured Health Benefit Package (HBP) reference
// data that every tier decision in this system is checked against.
//
// Schema follows Appendix F of the source spec (PMJAY_Startathon_Submission.md,
// Section 30) deliberately, so that every field the spec asked for has a
// concrete home, and so a reviewer who has read the spec can map this file
// to that section without translation.
package hbp

// Package is one row of the Health Benefit Package master list: a single
// procedure the scheme may cover, with the metadata needed to explain a
// match rather than just assert one.
//
// IMPORTANT — data provenance: see /docs/DATA_SOURCES.md before treating
// any field here as authoritative. Package/Specialty names in the seed
// dataset reflect real, publicly documented PMJAY package categories.
// Code and IndicativeRateINR are seed/placeholder values pending extraction
// from the actual NHA-published HBP master file — each record's Verified
// field says which is true for that record. Shipping this to real users
// without completing that extraction (Section 15.1 / Section 14.2 of the
// spec, "real, unglamorous, easy-to-underestimate engineering work") would
// mean citing numbers nobody has checked. Don't do that.
type Package struct {
	// PackageCode is the official HBP code. Empty/placeholder codes are
	// prefixed "SEED-" so they can never be silently mistaken for a real
	// government code in a citation shown to a family.
	PackageCode string `json:"package_code"`

	// PackageName is the plain package name as published.
	PackageName string `json:"package_name"`

	// Specialty is the clinical specialty, e.g. "General Surgery".
	Specialty string `json:"specialty"`

	// IndicativeRateINR is the published reimbursement rate in INR.
	// Placeholder values are still populated (never zero/omitted) so
	// arithmetic and display logic can be tested honestly, but Verified
	// will be false for any record where this number has not been
	// checked against the real NHA source.
	IndicativeRateINR int `json:"indicative_rate_inr"`

	// RequiresPreauth indicates whether cashless treatment under this
	// package requires pre-authorisation from the State Health Agency
	// before the hospital proceeds. This drives the pending-vs-denied
	// logic in internal/tiering.
	RequiresPreauth bool `json:"requires_preauth"`

	// KnownExclusionCategory is empty for a covered package, or a short
	// tag ("cosmetic", "opd_only", "fertility", "organ_transplant_partial")
	// if this entry exists in the dataset specifically to represent a
	// confirmed exclusion pattern rather than a coverable procedure.
	// Most exclusion handling uses the separate Exclusion list below;
	// this field exists for the rare package-shaped edge case (Section
	// 15.1: "known exclusion flags").
	KnownExclusionCategory string `json:"known_exclusion_category,omitempty"`

	// CommonDescriptionKeywords are plain-language terms a family might
	// actually use, aiding retrieval against informal descriptions before
	// the LLM disambiguation step ever runs (Section 15.1).
	CommonDescriptionKeywords []string `json:"common_description_keywords"`

	// ConfidenceNotes documents any known ambiguity: overlap with a
	// similar package, or proximity to the "unspecified procedure"
	// boundary (Section 8, point 2). Read by humans reviewing a match,
	// not shown verbatim to families.
	ConfidenceNotes string `json:"confidence_notes,omitempty"`

	// SourceNote is the traceability requirement from Appendix F: "Each
	// record in this schema should be traceable back to the specific
	// line or section of the source HBP document it came from." For
	// seed/placeholder records this instead states what needs doing.
	SourceNote string `json:"source_note"`

	// Verified is false unless PackageCode and IndicativeRateINR have
	// actually been checked against the published NHA HBP master file.
	// internal/response refuses to cite an unverified rate figure to a
	// family without a placeholder disclaimer — see response/templates.go.
	Verified bool `json:"verified"`

	// RateType distinguishes how IndicativeRateINR should be interpreted
	// and displayed. The zero value "" is the original, simplest case: a
	// single flat package rate, exactly as all 40 records from the first
	// build round use it. Two additional cases exist:
	//
	//   "tiered" — the true rate depends on the treating hospital's
	//   city-tier classification (HBP 2022 stratifies most packages into
	//   Tier1/X, Tier2/Y, Tier3/Z, roughly metro/mid-size/small-town).
	//   IndicativeRateINR holds the Tier3(Z) floor; RateMaxINR holds the
	//   Tier1(X) ceiling. A family is never told a single confident
	//   number for these — see packageCitation in response/templates.go
	//   — because picking one tier without knowing the hospital's actual
	//   classification would be a guess dressed up as a fact.
	//
	//   "per_diem" — the package is reimbursed per day of admission,
	//   stratified by ward level, not as one total. IndicativeRateINR
	//   holds the Routine Ward figure (so existing arithmetic/validation
	//   that assumes a single positive rate keeps working), and
	//   PerDiemRates holds the full four-level breakdown. This matters
	//   specifically because collapsing a per-day rate into a single
	//   "total cost" number would be the kind of confidently-wrong
	//   number this whole tool exists to avoid handing a family — see
	//   docs/DATA_SOURCES.md for how this was discovered and scoped.
	RateType string `json:"rate_type,omitempty"`

	// RateMaxINR is the Tier1(X), highest-tier end of a tiered rate
	// range. Zero/absent unless RateType is "tiered".
	RateMaxINR int `json:"rate_max_inr,omitempty"`

	// PerDiemRates is the ward-level per-day rate stratification. Nil
	// unless RateType is "per_diem".
	PerDiemRates *PerDiemRates `json:"per_diem_rates,omitempty"`
}

// PerDiemRates holds the four-level ward-stratified per-day reimbursement
// figures used by General Medicine and Pediatric Medical Management
// admission packages under HBP 2022 — roughly 90 and 30 diagnoses
// respectively, almost all sharing this exact same stratification
// (confirmed directly against the published rate table; see
// docs/DATA_SOURCES.md). A hospital disputing one of these packages will
// often be arguing about which of these four levels applied, not whether
// the package exists at all — so the tool needs to be able to show the
// actual structure, not a single collapsed number.
type PerDiemRates struct {
	// RoutineWardINR is the general-ward per-day rate — the default/
	// lowest tier, and the figure IndicativeRateINR mirrors for records
	// with RateType "per_diem".
	RoutineWardINR int `json:"routine_ward_inr"`

	// HDUINR is the High Dependency Unit per-day rate.
	HDUINR int `json:"hdu_inr"`

	// ICUNoVentINR is the ICU (without ventilator) per-day rate.
	ICUNoVentINR int `json:"icu_no_vent_inr"`

	// ICUVentINR is the ICU (with ventilator) per-day rate — the highest
	// of the four levels.
	ICUVentINR int `json:"icu_vent_inr"`
}

// Exclusion is one row of the smaller, separately maintained exclusion
// reference list (Section 15.1, Section 6.6): confirmed categories PMJAY
// does not cover, used to power the red tier with the same citation
// discipline as the green tier.
type Exclusion struct {
	// Category is a short machine key, e.g. "cosmetic".
	Category string `json:"category"`

	// DisplayName is the human-readable category name.
	DisplayName string `json:"display_name"`

	// Description explains the exclusion in plain language, safe to show
	// directly to a family (Section 29.3's tone).
	Description string `json:"description"`

	// Keywords aid matching an informal description to this exclusion.
	Keywords []string `json:"keywords"`

	// Nuance captures a partial-coverage carve-out, e.g. organ transplant
	// being state-dependent and procedure-specific (Section 6.6). Empty
	// when the exclusion is unconditional.
	Nuance string `json:"nuance,omitempty"`

	// SourceNote documents provenance (Section 6.6: "per a reported
	// government reply in the Rajya Sabha").
	SourceNote string `json:"source_note"`

	// Verified mirrors Package.Verified.
	Verified bool `json:"verified"`
}

// UnspecifiedProcedureCap is the discretionary catch-all ceiling described
// in Section 8, point 2 and confirmed directly against the NHA's own HBP
// 2.2 user guidelines during this build (see /docs/DATA_SOURCES.md): up to
// this many rupees, at the treating institution's judgement, within the
// overall per-family annual limit. This one figure IS independently
// verified, unlike most seed rate data — kept as a named constant rather
// than a JSON record so it can't be silently edited in a data file without
// the accompanying comment being noticed.
const UnspecifiedProcedureCapINR = 100000

// AnnualFamilyLimitINR is the overall per-family, per-year cover ceiling
// (Section 6.1), independently well-documented and stable scheme
// parameter, kept alongside the cap above for the same reason.
const AnnualFamilyLimitINR = 500000

// Dataset is the fully loaded, in-memory HBP reference data: the package
// list and the exclusion list, queried by internal/retrieval and cited by
// internal/response. Loaded once at process start via go:embed — see
// loader.go — and treated as read-only for the life of the process.
type Dataset struct {
	Packages   []Package
	Exclusions []Exclusion
}

// PackageByCode looks up a package by its code, returning (package, true)
// if found. Used by internal/tiering to go from a bare code in an
// extraction result back to the full record needed for citation.
func (d *Dataset) PackageByCode(code string) (Package, bool) {
	for _, p := range d.Packages {
		if p.PackageCode == code {
			return p, true
		}
	}
	return Package{}, false
}

// ExclusionByCategory looks up an exclusion by its category key.
func (d *Dataset) ExclusionByCategory(category string) (Exclusion, bool) {
	for _, e := range d.Exclusions {
		if e.Category == category {
			return e, true
		}
	}
	return Exclusion{}, false
}
