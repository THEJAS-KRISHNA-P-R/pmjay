package api

import "github.com/pmjay-advocate/backend/internal/store"

// IntakeRequest is the body of POST /api/v1/cases.
type IntakeRequest struct {
	Description string `json:"description"`
}

// CaseResponse is the shape returned for a case, whether just created or
// fetched later — the same shape either way, so the frontend has one
// type to render regardless of which endpoint produced it.
type CaseResponse struct {
	ID          string `json:"id"`
	Outcome     string `json:"outcome"`
	Citation    string `json:"citation,omitempty"`
	Description string `json:"description,omitempty"`

	// Persisted is false only when the answer was computed but writing it
	// to the store failed. The family still gets the guidance — that is
	// the care-first choice — but the UI must not promise that the case
	// can be reopened, annotated, or deleted later, because it can't.
	Persisted bool `json:"persisted"`

	CareFirstMessage string   `json:"care_first_message"`
	Disclaimer       string   `json:"disclaimer"`
	TierMessage      string   `json:"tier_message"`
	ActionSteps      []string `json:"action_steps,omitempty"`
	ComplaintText    string   `json:"complaint_text,omitempty"`
	HospitalScript   string   `json:"hospital_script,omitempty"`
	HandoffSummary   string   `json:"handoff_summary,omitempty"`
	EvidencePrompt   string   `json:"evidence_prompt,omitempty"`

	Evidence []EvidenceResponse `json:"evidence,omitempty"`
}

// EvidenceResponse mirrors store.EvidenceEntry for the API boundary.
type EvidenceResponse struct {
	CapturedAt string `json:"captured_at"`
	StaffName  string `json:"staff_name,omitempty"`
	ApproxTime string `json:"approx_time,omitempty"`
	Note       string `json:"note,omitempty"`
}

// AddEvidenceRequest is the body of POST /api/v1/cases/{id}/evidence.
type AddEvidenceRequest struct {
	StaffName  string `json:"staff_name,omitempty"`
	ApproxTime string `json:"approx_time,omitempty"`
	Note       string `json:"note,omitempty"`
}

// ErrorResponse is the body of every non-2xx JSON response.
//
// FallbackGuidance is populated specifically for infrastructure failures
// (the extraction call itself failing) — Section 10's care-first
// principle is written as an absolute ("Always"), and a family should
// not be left with nothing just because a downstream API call failed.
// See handlers.go's handleIntake for where this gets set.
type ErrorResponse struct {
	Error            string `json:"error"`
	FallbackGuidance string `json:"fallback_guidance,omitempty"`
}

// PackageSearchResponse is the body of GET /api/v1/packages.
//
// Only fields a public search page actually shows are included. The
// dataset's confidence_notes and source_note are engineer-facing prose and
// would triple the payload for no user benefit; the provenance summary
// lives in /api/v1/health and the repository's docs/DATA_SOURCES.md.
type PackageSearchResponse struct {
	Query       string                `json:"query"`
	Specialty   string                `json:"specialty,omitempty"`
	Specialties []string              `json:"specialties"`
	Total       int                   `json:"total"`
	Count       int                   `json:"count"`
	Offset      int                   `json:"offset"`
	Results     []PackageSearchResult `json:"results"`
}

// PackageSearchResult is one row of a package search. Rate fields follow
// the same three shapes as the dataset: flat (indicative_rate_inr),
// tiered (indicative_rate_inr to rate_max_inr), and per-diem
// (per_diem_rates). Verified is included rather than hidden so the UI can
// label an unverified name honestly instead of implying a rate exists.
//
// PackageCode is empty for an unverified record. The dataset keeps a
// placeholder identifier internally so the row is addressable, but a code
// is a citation: publishing "SEED-ORTHO-001" next to "Total Knee
// Replacement" invites a family or an agent to quote an identifier that
// exists in no government schedule. The name, specialty, and pre-auth flag
// are still useful on their own.
type PackageSearchResult struct {
	PackageCode       string        `json:"package_code,omitempty"`
	PackageName       string        `json:"package_name"`
	Specialty         string        `json:"specialty"`
	IndicativeRateINR int           `json:"indicative_rate_inr,omitempty"`
	RateMaxINR        int           `json:"rate_max_inr,omitempty"`
	RateType          string        `json:"rate_type,omitempty"`
	PerDiemRates      *PerDiemRates `json:"per_diem_rates,omitempty"`
	RequiresPreauth   bool          `json:"requires_preauth"`
	Verified          bool          `json:"verified"`
}

// PerDiemRates mirrors hbp.PerDiemRates at the API boundary.
type PerDiemRates struct {
	RoutineWardINR int `json:"routine_ward_inr"`
	HDUINR         int `json:"hdu_inr"`
	ICUNoVentINR   int `json:"icu_no_vent_inr"`
	ICUVentINR     int `json:"icu_vent_inr"`
}

func caseRecordToResponse(c store.CaseRecord) CaseResponse {
	evidence := make([]EvidenceResponse, 0, len(c.Evidence))
	for _, e := range c.Evidence {
		evidence = append(evidence, EvidenceResponse{
			CapturedAt: e.CapturedAt.Format("2006-01-02T15:04:05Z07:00"),
			StaffName:  e.StaffName,
			ApproxTime: e.ApproxTime,
			Note:       e.Note,
		})
	}
	return CaseResponse{
		ID:               c.ID,
		Persisted:        true,
		Outcome:          c.Outcome,
		Citation:         c.Citation,
		Description:      c.FamilyDescriptionRaw,
		CareFirstMessage: c.CareFirstMessage,
		Disclaimer:       c.Disclaimer,
		TierMessage:      c.TierMessage,
		ActionSteps:      c.ActionSteps,
		ComplaintText:    c.ComplaintText,
		HospitalScript:   c.HospitalScript,
		HandoffSummary:   c.HandoffSummary,
		EvidencePrompt:   c.EvidencePrompt,
		Evidence:         evidence,
	}
}
