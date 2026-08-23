package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFileStore_CreateAndReload_SurvivesRestart(t *testing.T) {
	// The core durability promise: create a store, write a case, throw
	// the FileStore object away entirely (simulating a process restart),
	// open a fresh one at the same path, and confirm the data is still
	// there.
	dir := t.TempDir()
	path := filepath.Join(dir, "cases.json")

	fs1, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	ctx := context.Background()
	c := sampleCase("durable-case")
	if err := fs1.Create(ctx, c); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Simulate a full process restart: discard fs1, open fresh.
	fs2, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("re-opening FileStore failed: %v", err)
	}
	got, found, err := fs2.Get(ctx, "durable-case")
	if err != nil {
		t.Fatalf("Get after reload failed: %v", err)
	}
	if !found {
		t.Fatal("case did not survive a simulated restart")
	}
	if got.FamilyDescriptionRaw != c.FamilyDescriptionRaw {
		t.Errorf("reloaded case data mismatch: got %+v", got)
	}
}

func TestFileStore_FirstRun_NoExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "does-not-exist-yet.json")

	fs, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("expected no error on first run with no existing file, got: %v", err)
	}
	_, found, _ := fs.Get(context.Background(), "anything")
	if found {
		t.Error("expected empty store on first run")
	}
}

func TestFileStore_CorruptFileFailsLoudlyAtStartup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "corrupt.json")
	if err := os.WriteFile(path, []byte("{not valid json"), 0o644); err != nil {
		t.Fatalf("test setup: writing corrupt file failed: %v", err)
	}

	_, err := NewFileStore(path)
	if err == nil {
		t.Fatal("expected NewFileStore to fail loudly on a corrupt data file rather than silently start empty")
	}
}

func TestFileStore_WritesAreActuallyOnDisk(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cases.json")
	fs, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	if err := fs.Create(context.Background(), sampleCase("on-disk-check")); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected data file to exist on disk after Create: %v", err)
	}
	if len(data) == 0 {
		t.Error("data file is empty after a successful Create")
	}
}

func TestFileStore_NoLeftoverTempFileAfterSuccessfulWrite(t *testing.T) {
	// The atomic write uses a .tmp file that should be renamed away, never
	// left behind, after a successful write.
	dir := t.TempDir()
	path := filepath.Join(dir, "cases.json")
	fs, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	fs.Create(context.Background(), sampleCase("tmp-check"))

	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Error("expected the .tmp file to be gone (renamed) after a successful write")
	}
}

func TestFileStore_UpdateAndAppendEvidencePersist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cases.json")
	fs, _ := NewFileStore(path)
	ctx := context.Background()

	c := sampleCase("evolving-case")
	fs.Create(ctx, c)

	c.TierMessage = "updated"
	if err := fs.Update(ctx, c); err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if _, err := fs.AppendEvidence(ctx, "evolving-case", EvidenceEntry{Note: "captured"}); err != nil {
		t.Fatalf("AppendEvidence failed: %v", err)
	}

	fs2, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	got, found, _ := fs2.Get(ctx, "evolving-case")
	if !found {
		t.Fatal("case missing after reload")
	}
	if got.TierMessage != "updated" {
		t.Errorf("update did not persist: got tier message %q", got.TierMessage)
	}
	if len(got.Evidence) != 1 || got.Evidence[0].Note != "captured" {
		t.Errorf("evidence did not persist: got %+v", got.Evidence)
	}
}

func TestFileStore_SatisfiesStoreInterface(t *testing.T) {
	dir := t.TempDir()
	fs, err := NewFileStore(filepath.Join(dir, "cases.json"))
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	var _ Store = fs // compile-time check
}
