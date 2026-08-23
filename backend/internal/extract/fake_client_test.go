package extract

import (
	"context"
	"errors"
	"testing"
)

func TestFakeClient_ReturnsRegisteredResponse(t *testing.T) {
	f := NewFakeClient()
	want := Result{ExtractedSituationSummary: "test summary"}
	f.Register("some description", want)

	got, err := f.Extract(context.Background(), "some description", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ExtractedSituationSummary != want.ExtractedSituationSummary {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestFakeClient_UnregisteredInputFailsLoudly(t *testing.T) {
	f := NewFakeClient()
	_, err := f.Extract(context.Background(), "never registered", nil, nil)
	if err == nil {
		t.Fatal("expected error for unregistered description — silent fallback would mask real test gaps")
	}
}

func TestFakeClient_RegisteredErrorIsReturned(t *testing.T) {
	f := NewFakeClient()
	wantErr := errors.New("simulated failure")
	f.RegisterError("bad input", wantErr)

	_, err := f.Extract(context.Background(), "bad input", nil, nil)
	if !errors.Is(err, wantErr) {
		t.Errorf("expected registered error to be returned, got %v", err)
	}
}

func TestFakeClient_CallLogRecordsInputs(t *testing.T) {
	f := NewFakeClient()
	f.Register("a", Result{})
	f.Register("b", Result{})

	f.Extract(context.Background(), "a", nil, nil)
	f.Extract(context.Background(), "b", nil, nil)

	if len(f.CallLog) != 2 || f.CallLog[0] != "a" || f.CallLog[1] != "b" {
		t.Errorf("expected call log [a, b], got %v", f.CallLog)
	}
}

func TestFakeClient_SatisfiesExtractorInterface(t *testing.T) {
	var _ Extractor = NewFakeClient() // compile-time check
}
