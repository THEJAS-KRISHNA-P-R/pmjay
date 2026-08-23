// Package extract is the one place in this system where a language model
// is actually necessary — turning an unstructured, possibly Malayalam/
// English code-mixed family description into a structured, ranked set of
// candidate HBP package matches with an honest confidence signal attached.
//
// Everything downstream of this package (internal/tiering,
// internal/response) is deliberately plain, deterministic Go with no model
// calls, per the spec's own design principle (Section 58): "AI where
// genuine language ambiguity exists, deterministic auditable logic
// everywhere after." Keeping the boundary exactly here — not one layer
// earlier, not one layer later — is what makes the tier logic testable
// without a live API key and cheap to run at scale.
package extract

import "context"

// PendingSignal is the extraction step's read on whether the family's own
// description sounds like a pre-authorisation that is still pending
// (Section 6.8) rather than a final denial. This is a *signal*, not a
// verdict — internal/tiering combines it with an independent deterministic
// pattern check (internal/tiering/preauth_pattern.go) before deciding
// anything, exactly because a safety-relevant distinction like this one
// should never rest on a single, unverified source.
type PendingSignal string

const (
	SignalPendingLikely PendingSignal = "pending_likely"
	SignalDeniedLikely  PendingSignal = "denied_final_likely"
	SignalUnclear       PendingSignal = "unclear"
	SignalNotApplicable PendingSignal = "not_applicable" // description has no preauth angle at all
)

// CandidateMatch is one possible HBP package the extraction step believes
// could correspond to the family's description, with an explicit
// confidence score. The spec (Appendix Z) requires checking not just a top
// score but the *gap* between the top two candidates — so this type is a
// slice for a reason; collapsing it to a single best guess earlier than
// necessary would throw away exactly the signal the amber tier depends on.
type CandidateMatch struct {
	PackageCode string `json:"package_code"`
	ConfidenceP int    `json:"confidence_percent"` // 0-100
	Reasoning   string `json:"reasoning"`          // short, human-checkable, never shown verbatim to the family
}

// ExclusionMatch is the extraction step's read on whether the family's
// description matches one of the confirmed, published exclusion
// categories (Section 6.6) — evaluated with the same ranked-confidence
// rigor as CandidateMatch, deliberately. The red tier is "not optional,
// and not a lesser feature" (Section 9): it earns the same reasoning
// quality as a covered-package match, not a cheaper keyword-only path.
type ExclusionMatch struct {
	Category    string `json:"category"`
	ConfidenceP int    `json:"confidence_percent"` // 0-100
	Reasoning   string `json:"reasoning"`
}

// Result is the full structured output of one extraction call.
type Result struct {
	// ExtractedSituationSummary restates what the model understood from
	// the family's own words, in plain English, regardless of the input
	// language mix. Used for the confirmation-loop (Section 60, Test 2)
	// and carried into the NALSA handoff summary (Section 12) so nothing
	// has to be re-explained from scratch.
	ExtractedSituationSummary string `json:"extracted_situation_summary"`

	// Candidates is ranked most-plausible-first, from the shortlist
	// internal/retrieval handed in. Never empty when Result is valid —
	// the shortlist always includes the Unspecified catch-all, so even a
	// "no real match" case should surface it as a low-confidence
	// candidate rather than an empty list.
	Candidates []CandidateMatch `json:"candidates"`

	// ExclusionMatches is ranked most-plausible-first, evaluated against
	// the shortlist internal/retrieval drew from the exclusion reference
	// list. May legitimately be empty — most descriptions don't match any
	// exclusion category at all, and an empty slice here is the honest
	// result for those, not a missing field.
	ExclusionMatches []ExclusionMatch `json:"exclusion_matches"`

	// PendingSignal is this call's independent read on the pending-vs-
	// denied question. Combined with the deterministic pattern check in
	// internal/tiering, never trusted alone.
	Pending PendingSignal `json:"pending_signal"`

	// MultipleDistinctIssuesDetected flags a description that bundles more
	// than one separable problem (Section 29.4's handoff example: a
	// billing dispute, a staff-conduct complaint, and an ID-mismatch
	// question, all at once). This is one of the strongest signals that a
	// case should route to human handoff rather than any single tier.
	MultipleDistinctIssuesDetected bool `json:"multiple_distinct_issues_detected"`

	// FamilyDistressSignal is a light-touch, non-diagnostic read on
	// whether the family's own description reads as highly distressed or
	// difficult to follow (Section 12, Section 29.4) — used only to bias
	// toward human handoff sooner, never to characterise the family in
	// anything shown back to them.
	FamilyDistressSignal bool `json:"family_distress_signal"`
}

// Extractor is satisfied by both the real Claude-backed client
// (claude_client.go) and the deterministic test fake (fake_client.go).
// internal/api depends only on this interface, never on the concrete
// client, so the entire HTTP layer is testable without network access or
// an API key.
type Extractor interface {
	Extract(ctx context.Context, familyDescription string, candidates []CandidatePackageInfo, exclusionCandidates []CandidateExclusionInfo) (Result, error)
}

// CandidatePackageInfo is the trimmed-down package information actually
// sent to the model — deliberately not the full hbp.Package struct, so
// that internal fields (ConfidenceNotes meant for engineers, Verified
// flags, SourceNote) never leak into a model prompt or, from there, into
// anything shown to a family.
type CandidatePackageInfo struct {
	PackageCode string   `json:"package_code"`
	PackageName string   `json:"package_name"`
	Specialty   string   `json:"specialty"`
	Keywords    []string `json:"keywords"`
}

// CandidateExclusionInfo is the trimmed-down exclusion information sent
// to the model, mirroring CandidatePackageInfo's reasoning for existing.
type CandidateExclusionInfo struct {
	Category    string   `json:"category"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}
