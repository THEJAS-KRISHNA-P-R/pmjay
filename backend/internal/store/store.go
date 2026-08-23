// Package store persists case records across the lifetime of a family's
// interaction with the system — from initial intake through evidence
// capture through (if applicable) handoff — so that nothing has to be
// re-entered or re-explained at a later step (Section 11, "Context &
// continuity").
//
// Deliberately NOT a wrapper around a database engine. See
// ../../../ARCHITECTURE.md for the full reasoning; in short, this system's
// write volume (one case per family interaction, evidence appended a
// handful of times per case) does not need a hosted database service to
// run reliably, and running one anyway would be the single largest
// avoidable recurring cost in this system's entire hosting bill. Store
// is an interface specifically so that judgement can be revisited later
// without touching any other package — see MemStore and FileStore for
// the two implementations, and PostgresStore's absence for what a first
// PR adding one would need to satisfy.
package store

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned by Update and AppendEvidence when the given
// case ID does not exist. Callers use errors.Is(err, ErrNotFound) to
// distinguish "case doesn't exist" (a normal, expected condition worth a
// 404) from a real storage failure (worth a 500) — see
// internal/api/handlers.go.
var ErrNotFound = errors.New("store: case not found")

// EvidenceEntry is one piece of evidence a family captured during their
// interaction — Section 7, Step 6.
type EvidenceEntry struct {
	CapturedAt time.Time `json:"captured_at"`
	StaffName  string    `json:"staff_name,omitempty"`
	ApproxTime string    `json:"approx_time,omitempty"`
	Note       string    `json:"note,omitempty"`
}

// CaseRecord is the full persisted state of one family's case. Flat and
// JSON-serializable by construction — deliberately not a direct
// persistence of internal/tiering.Decision or internal/response's
// unexported-field type, both of which are shaped for their own
// packages' concerns, not for storage or for direct API serialization.
type CaseRecord struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	FamilyDescriptionRaw string `json:"family_description_raw"`

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

	Evidence []EvidenceEntry `json:"evidence,omitempty"`
}

// Store is satisfied by MemStore and FileStore. internal/api depends
// only on this interface.
type Store interface {
	// Create persists a new case record. Returns an error if the ID
	// already exists — IDs are minted by the caller (see NewCaseID) and
	// a collision almost certainly indicates a bug upstream, not a
	// condition to silently overwrite.
	Create(ctx context.Context, c CaseRecord) error

	// Get retrieves a case by ID. The bool return is false, with a nil
	// error, when the ID simply doesn't exist — a normal, expected
	// outcome (e.g. a stale link), not treated as a failure.
	Get(ctx context.Context, id string) (CaseRecord, bool, error)

	// Update overwrites an existing case record. Returns an error if the
	// ID does not already exist — use Create for new records.
	Update(ctx context.Context, c CaseRecord) error

	// AppendEvidence adds one evidence entry to an existing case and
	// returns the updated record, or an error if the case doesn't exist.
	AppendEvidence(ctx context.Context, id string, e EvidenceEntry) (CaseRecord, error)
}
