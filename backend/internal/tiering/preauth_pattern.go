package tiering

import "strings"

// PreauthPattern is a deterministic, keyword-pattern read on whether a
// family's own description sounds like a pending pre-authorisation
// (Section 6.8) rather than a final denial. This exists specifically as
// an independent second opinion to the LLM's own PendingSignal — Section
// 10's absolute care-first rule and H3's safety-failure kill condition
// ("confidently wrong on the ambiguous arm") both argue against letting
// a single, unverifiable model judgement be the only thing standing
// between a family and an overclaimed answer on exactly the distinction
// the spec calls "the single highest-value thing the system does."
//
// This is intentionally simple English-keyword matching, not an attempt
// at full language understanding — that job belongs to the LLM step,
// which reads the family's original text directly, in whatever language
// mix they used. This detector is a cheap, fast, always-available,
// fully-offline-testable safety net layered on top, not a replacement.
type PreauthPattern string

const (
	PatternPendingLikely PreauthPattern = "pending_likely"
	PatternDeniedFinal   PreauthPattern = "denied_final"
	PatternNone          PreauthPattern = "none"
)

var pendingPhrases = []string{
	"pending", "not cleared", "hasn't cleared", "has not cleared",
	"not yet approved", "not yet cleared", "still processing",
	"waiting for approval", "waiting period", "yet to be approved",
	"under process", "under review", "still under review",
	"not been decided", "haven't decided", "hasn't decided",
	"not confirmed yet", "in process", "being processed",
	"awaiting approval", "awaiting clearance", "clearance is pending",
	"insurance hasn't cleared", "hasn't cleared it",
}

var deniedFinalPhrases = []string{
	"denied", "rejected", "refused", "won't cover", "wont cover",
	"will not cover", "doesn't cover", "does not cover", "declined",
	"not covered", "turned down", "flatly said no", "said no",
	"card won't cover", "card wont cover",
}

// Detect scans raw description text for pending-vs-denied language cues.
// When both pending and denied cues are present at once (a genuinely
// mixed signal — e.g. "they denied it but also said it's still pending
// review"), Detect deliberately returns PatternPendingLikely rather than
// picking a side: per Appendix AA's conservative-bias principle, treating
// a mixed signal as "possibly still pending" is the safer error to make
// than treating it as a confident final denial.
func Detect(description string) PreauthPattern {
	lower := strings.ToLower(description)

	pendingHit := containsAny(lower, pendingPhrases)
	deniedHit := containsAny(lower, deniedFinalPhrases)

	switch {
	case pendingHit:
		return PatternPendingLikely
	case deniedHit:
		return PatternDeniedFinal
	default:
		return PatternNone
	}
}

func containsAny(haystack string, needles []string) bool {
	for _, n := range needles {
		if strings.Contains(haystack, n) {
			return true
		}
	}
	return false
}
