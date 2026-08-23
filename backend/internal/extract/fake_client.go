package extract

import (
	"context"
	"fmt"
)

// FakeClient is a deterministic, offline Extractor used throughout the
// test suite (internal/tiering, internal/response, internal/api) so the
// entire pipeline downstream of extraction is testable without a live API
// key or network access. Registered by exact input string on purpose —
// a fuzzy fake would risk masking a real wiring bug (e.g. the API handler
// accidentally passing the wrong string through) behind a fake that
// "figures out" the right answer anyway.
type FakeClient struct {
	responses map[string]Result
	errors    map[string]error
	// CallLog records every description this fake was asked to extract,
	// in order — tests use this to assert the API layer actually called
	// the extractor with the expected input, not just that it returned
	// the right final response by coincidence.
	CallLog []string
}

// NewFakeClient builds an empty fake. Use Register to add canned
// responses, or NewFakeClientFromFixtures for the full spec scenario set.
func NewFakeClient() *FakeClient {
	return &FakeClient{
		responses: make(map[string]Result),
		errors:    make(map[string]error),
	}
}

// Register adds a canned response for an exact family description string.
func (f *FakeClient) Register(familyDescription string, result Result) {
	f.responses[familyDescription] = result
}

// RegisterError makes the fake return an error for an exact family
// description string — used to test error handling in internal/api
// without needing to break a real network call to exercise that path.
func (f *FakeClient) RegisterError(familyDescription string, err error) {
	f.errors[familyDescription] = err
}

// Extract implements Extractor. Exclusion/package candidate arguments are
// accepted (to satisfy the interface and so callers pass real retrieval
// output through the same code path as production) but intentionally
// ignored by the fake — the fake's whole point is that the *extraction*
// step's judgement is pre-recorded, not recomputed.
func (f *FakeClient) Extract(_ context.Context, familyDescription string, _ []CandidatePackageInfo, _ []CandidateExclusionInfo) (Result, error) {
	f.CallLog = append(f.CallLog, familyDescription)

	if err, ok := f.errors[familyDescription]; ok {
		return Result{}, err
	}
	if result, ok := f.responses[familyDescription]; ok {
		return result, nil
	}
	return Result{}, fmt.Errorf("extract: FakeClient has no registered response for description %q — register one with Register(), or this is a real test gap, not something to work around", familyDescription)
}
