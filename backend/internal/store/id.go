package store

import (
	"crypto/rand"
	"fmt"
)

// NewCaseID generates a random, unpredictable case ID (RFC 4122 UUID v4
// format). Implemented directly with crypto/rand rather than pulling in
// a UUID dependency — see ../../../ARCHITECTURE.md on why this backend has
// zero external Go dependencies. Fifteen lines of well-understood stdlib
// code is a smaller thing to audit than a dependency, for a need this
// small.
func NewCaseID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand failing indicates a broken system entropy source —
		// a condition serious enough that continuing with a predictable
		// fallback ID would be worse than crashing loudly.
		panic(fmt.Sprintf("store: crypto/rand failed, cannot safely generate a case ID: %v", err))
	}
	// Set version (4) and variant (RFC 4122) bits per the spec.
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
