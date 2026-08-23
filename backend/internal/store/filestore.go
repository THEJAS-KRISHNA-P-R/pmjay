package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileStore is a durable Store backed by a single JSON file on disk. All
// reads are served from an in-memory MemStore for speed; every write goes
// to memory AND is flushed to disk before the call returns, using an
// atomic write-then-rename so a crash or power loss mid-write can never
// leave a half-written, corrupted file behind — the rename either
// happens completely or not at all, on any POSIX filesystem.
//
// This is the entire persistence story for this system (see
// ../../../ARCHITECTURE.md, "reducing bills"): no database engine, no
// hosted database service, no separate backup job to configure — one
// file, one directory that already needs backing up along with
// everything else on the instance.
type FileStore struct {
	mem     *MemStore
	path    string
	writeMu sync.Mutex // serializes disk writes; MemStore already handles in-memory concurrency
}

// NewFileStore opens (or creates) a FileStore at path, loading any
// existing data found there first.
func NewFileStore(path string) (*FileStore, error) {
	fs := &FileStore{mem: NewMemStore(), path: path}

	// 0o700: this directory holds nothing but real families' case
	// records (health situations, denial details, personal
	// circumstances) — no reason for anyone but the process owner to
	// even see filenames in here, let alone read them.
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("store: creating data directory: %w", err)
	}

	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		var cases map[string]CaseRecord
		if err := json.Unmarshal(data, &cases); err != nil {
			return nil, fmt.Errorf("store: existing data file %q is corrupt: %w", path, err)
		}
		fs.mem.loadAll(cases)
	case os.IsNotExist(err):
		// First run — nothing to load, an empty file will be written on
		// first write.
	default:
		return nil, fmt.Errorf("store: reading data file %q: %w", path, err)
	}

	return fs, nil
}

func (fs *FileStore) Create(ctx context.Context, c CaseRecord) error {
	if err := fs.mem.Create(ctx, c); err != nil {
		return err
	}
	return fs.flush()
}

func (fs *FileStore) Get(ctx context.Context, id string) (CaseRecord, bool, error) {
	return fs.mem.Get(ctx, id)
}

func (fs *FileStore) Update(ctx context.Context, c CaseRecord) error {
	if err := fs.mem.Update(ctx, c); err != nil {
		return err
	}
	return fs.flush()
}

func (fs *FileStore) AppendEvidence(ctx context.Context, id string, e EvidenceEntry) (CaseRecord, error) {
	updated, err := fs.mem.AppendEvidence(ctx, id, e)
	if err != nil {
		return CaseRecord{}, err
	}
	if err := fs.flush(); err != nil {
		return CaseRecord{}, err
	}
	return updated, nil
}

// flush serializes the full in-memory snapshot to disk atomically: write
// to a temp file in the same directory (so the rename is guaranteed to
// be on the same filesystem, which is what makes it atomic), then rename
// over the real path.
func (fs *FileStore) flush() error {
	fs.writeMu.Lock()
	defer fs.writeMu.Unlock()

	snapshot := fs.mem.snapshot()
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("store: marshalling snapshot: %w", err)
	}

	tmpPath := fs.path + ".tmp"
	// 0o600: owner read/write only. This file is the persisted store of
	// every case this instance holds — real families' health situations
	// and denial details — so it must not be world- or group-readable
	// regardless of what the containing directory's permissions are.
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return fmt.Errorf("store: writing temp file: %w", err)
	}
	if err := os.Rename(tmpPath, fs.path); err != nil {
		return fmt.Errorf("store: renaming temp file into place: %w", err)
	}
	return nil
}
