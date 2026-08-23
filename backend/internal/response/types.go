// Package response turns a tiering.Decision into the actual words shown
// to a family: templated, cited, and — this is the part that must never
// depend on anyone remembering to do it right — always led by the
// care-first safety rule from Section 10 of the spec, on every single
// outcome, with no exceptions.
//
// The enforcement mechanism is the type system, not a checklist. Every
// field of FamilyResponse is unexported. The only way to obtain a
// FamilyResponse from outside this package is to call Build, and Build
// sets the care-first message unconditionally, before it even looks at
// which tier it's building for. There is no second constructor, no
// exported struct literal, and no field a caller could zero out. A
// reviewer does not have to trust that every call site remembered the
// safety rule — the compiler will not allow a FamilyResponse to exist
// without one.
package response

import "github.com/pmjay-advocate/backend/internal/tiering"

// FamilyResponse is the fully-built, ready-to-display answer for one
// family interaction. Construct only via Build.
type FamilyResponse struct {
	careFirstMessage string
	disclaimer       string
	outcome          tiering.Outcome
	tierMessage      string
	citation         string
	actionSteps      []string
	complaintText    string // empty when not applicable (e.g. red tier)
	hospitalScript   string // empty when not applicable
	handoffSummary   string // populated only for Handoff outcome
	evidencePrompt   string // empty only for Handoff (handoff has its own summary instead)
}

// CareFirstMessage is always non-empty and always exactly CareFirstText —
// see builder_test.go's TestBuild_CareFirstMessageIsAlwaysPresent for the
// exhaustive check across every outcome.
func (r FamilyResponse) CareFirstMessage() string { return r.careFirstMessage }

// Disclaimer is always non-empty and always exactly DisclaimerText — same
// unconditional guarantee as CareFirstMessage, checked by the same
// exhaustive test. See DisclaimerText's doc comment for why this exists
// alongside a system that already withholds unverified figures
// structurally (packageCitation) rather than just flagging them.
func (r FamilyResponse) Disclaimer() string { return r.disclaimer }

// Outcome is the tier/routing this response was built for.
func (r FamilyResponse) Outcome() tiering.Outcome { return r.outcome }

// TierMessage is the main explanatory text for this tier.
func (r FamilyResponse) TierMessage() string { return r.tierMessage }

// Citation names the specific package or exclusion rule this response is
// based on — never empty for Green, Amber, Red, or Mixed (Explainability,
// Section 11). Empty only for Handoff, which cites nothing because it
// makes no coverage claim at all.
func (r FamilyResponse) Citation() string { return r.citation }

// ActionSteps are the concrete "do this right now" items, in order.
func (r FamilyResponse) ActionSteps() []string { return append([]string(nil), r.actionSteps...) }

// ComplaintText is the pre-filled CGRMS complaint, ready to copy into the
// official Ayushman App. Empty when no complaint is warranted (Red) or
// applicable (Handoff, which routes to a person instead).
func (r FamilyResponse) ComplaintText() string { return r.complaintText }

// HospitalScript is the exact words suggested for the family to say at
// the billing desk. Empty when not applicable.
func (r FamilyResponse) HospitalScript() string { return r.hospitalScript }

// HandoffSummary is populated only for Handoff — the full context passed
// to a NALSA Para Legal Volunteer so nothing has to be re-explained
// (Section 12).
func (r FamilyResponse) HandoffSummary() string { return r.handoffSummary }

// EvidencePrompt is the staff-name/time/written-denial capture prompt
// (Section 7, Step 6), shown at the point in the flow where it's still
// actionable. Empty for Handoff and Red (nothing to capture evidence
// toward once no grievance is being made).
func (r FamilyResponse) EvidencePrompt() string { return r.evidencePrompt }
