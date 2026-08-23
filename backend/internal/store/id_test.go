package store

import (
	"regexp"
	"testing"
)

var uuidV4Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestNewCaseID_MatchesUUIDv4Format(t *testing.T) {
	for i := 0; i < 20; i++ {
		id := NewCaseID()
		if !uuidV4Pattern.MatchString(id) {
			t.Errorf("NewCaseID() = %q, does not match RFC 4122 UUID v4 format", id)
		}
	}
}

func TestNewCaseID_NoCollisionsInVolume(t *testing.T) {
	seen := make(map[string]bool)
	const n = 10000
	for i := 0; i < n; i++ {
		id := NewCaseID()
		if seen[id] {
			t.Fatalf("collision detected after %d IDs generated: %q", i, id)
		}
		seen[id] = true
	}
}
