package store

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func sampleCase(id string) CaseRecord {
	return CaseRecord{
		ID:                   id,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		FamilyDescriptionRaw: "test description",
		Outcome:              "green",
		Citation:             "Test Package",
		CareFirstMessage:     "care first text",
		TierMessage:          "tier message",
	}
}

func TestMemStore_CreateAndGet(t *testing.T) {
	s := NewMemStore()
	ctx := context.Background()
	c := sampleCase("case-1")

	if err := s.Create(ctx, c); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	got, found, err := s.Get(ctx, "case-1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if !found {
		t.Fatal("expected case to be found")
	}
	if got.FamilyDescriptionRaw != c.FamilyDescriptionRaw {
		t.Errorf("got %+v, want %+v", got, c)
	}
}

func TestMemStore_GetMissing_ReturnsFalseNotError(t *testing.T) {
	s := NewMemStore()
	_, found, err := s.Get(context.Background(), "does-not-exist")
	if err != nil {
		t.Fatalf("expected nil error for a missing ID, got: %v", err)
	}
	if found {
		t.Fatal("expected found=false for a missing ID")
	}
}

func TestMemStore_CreateDuplicateIDFails(t *testing.T) {
	s := NewMemStore()
	ctx := context.Background()
	c := sampleCase("dup")
	if err := s.Create(ctx, c); err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	if err := s.Create(ctx, c); err == nil {
		t.Fatal("expected error creating a duplicate ID")
	}
}

func TestMemStore_UpdateNonExistentFails(t *testing.T) {
	s := NewMemStore()
	err := s.Update(context.Background(), sampleCase("never-created"))
	if err == nil {
		t.Fatal("expected error updating a case that was never created")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected error to wrap ErrNotFound, got: %v", err)
	}
}

func TestMemStore_UpdatePersistsChanges(t *testing.T) {
	s := NewMemStore()
	ctx := context.Background()
	c := sampleCase("case-2")
	s.Create(ctx, c)

	c.TierMessage = "updated message"
	if err := s.Update(ctx, c); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	got, _, _ := s.Get(ctx, "case-2")
	if got.TierMessage != "updated message" {
		t.Errorf("expected updated tier message, got %q", got.TierMessage)
	}
}

func TestMemStore_AppendEvidenceToNonExistentFails(t *testing.T) {
	s := NewMemStore()
	_, err := s.AppendEvidence(context.Background(), "nope", EvidenceEntry{Note: "x"})
	if err == nil {
		t.Fatal("expected error appending evidence to a nonexistent case")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected error to wrap ErrNotFound, got: %v", err)
	}
}

func TestMemStore_AppendEvidenceAccumulates(t *testing.T) {
	s := NewMemStore()
	ctx := context.Background()
	s.Create(ctx, sampleCase("case-3"))

	s.AppendEvidence(ctx, "case-3", EvidenceEntry{StaffName: "Nurse A", Note: "first"})
	updated, err := s.AppendEvidence(ctx, "case-3", EvidenceEntry{StaffName: "Nurse B", Note: "second"})
	if err != nil {
		t.Fatalf("AppendEvidence failed: %v", err)
	}
	if len(updated.Evidence) != 2 {
		t.Fatalf("expected 2 accumulated evidence entries, got %d", len(updated.Evidence))
	}
	if updated.Evidence[0].StaffName != "Nurse A" || updated.Evidence[1].StaffName != "Nurse B" {
		t.Errorf("evidence entries out of order or wrong content: %+v", updated.Evidence)
	}
}

func TestMemStore_ConcurrentAccessIsSafe(t *testing.T) {
	// Run with -race to actually catch data races; a plain pass here is
	// necessary but not sufficient — see Makefile's `test-race` target.
	s := NewMemStore()
	ctx := context.Background()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			id := NewCaseID()
			c := sampleCase(id)
			if err := s.Create(ctx, c); err != nil {
				t.Errorf("concurrent create %d failed: %v", n, err)
				return
			}
			if _, err := s.AppendEvidence(ctx, id, EvidenceEntry{Note: "concurrent"}); err != nil {
				t.Errorf("concurrent append %d failed: %v", n, err)
			}
		}(i)
	}
	wg.Wait()
}
