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
	ID       string `json:"id"`
	Outcome  string `json:"outcome"`
	Citation string `json:"citation,omitempty"`

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
		Outcome:          c.Outcome,
		Citation:         c.Citation,
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
