package store

import (
	"context"
	"fmt"
	"sync"
)

// MemStore is a thread-safe, pure in-memory Store. State does not survive
// a process restart — used directly in tests, and used internally by
// FileStore as the fast-path cache that every read goes through.
type MemStore struct {
	mu    sync.RWMutex
	cases map[string]CaseRecord
}

// NewMemStore builds an empty in-memory store.
func NewMemStore() *MemStore {
	return &MemStore{cases: make(map[string]CaseRecord)}
}

func (m *MemStore) Create(_ context.Context, c CaseRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.cases[c.ID]; exists {
		return fmt.Errorf("store: case %q already exists", c.ID)
	}
	m.cases[c.ID] = c
	return nil
}

func (m *MemStore) Get(_ context.Context, id string) (CaseRecord, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.cases[id]
	return c, ok, nil
}

func (m *MemStore) Update(_ context.Context, c CaseRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.cases[c.ID]; !exists {
		return fmt.Errorf("store: cannot update case %q: %w", c.ID, ErrNotFound)
	}
	m.cases[c.ID] = c
	return nil
}

func (m *MemStore) AppendEvidence(_ context.Context, id string, e EvidenceEntry) (CaseRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, exists := m.cases[id]
	if !exists {
		return CaseRecord{}, fmt.Errorf("store: cannot append evidence to case %q: %w", id, ErrNotFound)
	}
	c.Evidence = append(c.Evidence, e)
	m.cases[id] = c
	return c, nil
}

// snapshot returns a copy of every case currently held, used by FileStore
// to serialize the full store to disk. Not part of the Store interface —
// callers needing every case are a persistence concern, not an API
// concern (an admin/debug listing endpoint would be a deliberate,
// separate addition, not implied by this).
func (m *MemStore) snapshot() map[string]CaseRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]CaseRecord, len(m.cases))
	for k, v := range m.cases {
		out[k] = v
	}
	return out
}

// loadAll replaces the store's entire contents — used once, by FileStore,
// when restoring from disk at startup.
func (m *MemStore) loadAll(cases map[string]CaseRecord) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cases = cases
}
