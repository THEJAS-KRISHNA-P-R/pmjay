# `internal/store`

Persists case records — one per family interaction, from intake through evidence capture through, if applicable, handoff — so nothing has to be re-entered or re-explained at a later step. No database. See `../../../ARCHITECTURE.md`'s "why no database" for the full reasoning; the short version is that this system's actual write volume (one record per family, a handful of evidence entries appended per case) doesn't need a hosted database engine to run reliably, and running one anyway would be the single largest avoidable line item in this system's hosting bill.

## Files

| File | What it does |
|---|---|
| `store.go` | The `Store` interface, `CaseRecord`, `EvidenceEntry`, `ErrNotFound`. |
| `memstore.go` | `MemStore` — thread-safe, pure in-memory implementation. |
| `filestore.go` | `FileStore` — wraps `MemStore` for reads, flushes to a single JSON file atomically on every write. |
| `id.go` | `NewCaseID()` — a ~15-line `crypto/rand`-based UUID v4 generator. |

## Why `Store` is an interface with exactly two implementations

`internal/api` depends only on `Store`, never on `*FileStore` directly — this is the seam. `MemStore` is used directly in tests (fast, no filesystem, trivially resettable between test cases) and internally by `FileStore` as its in-memory read path. If real scale ever demands a hosted database, a `PostgresStore` or Supabase-backed implementation is a self-contained addition that satisfies the same four-method interface and touches nothing else in this codebase — no call site outside this package would need to change.

## `FileStore`: how it stays durable without a database

Every write (`Create`, `Update`, `AppendEvidence`) goes to the in-memory `MemStore` first, then immediately flushes the **entire current snapshot** to disk before the call returns — not just the one changed record. `flush()`'s atomicity is the specific thing worth understanding if you touch this file:

1. Marshal the full snapshot to JSON.
2. Write it to `<path>.tmp` — a temp file in the *same directory* as the real path, specifically so the next step's rename is guaranteed to be on the same filesystem (required for the atomicity guarantee below to hold).
3. `os.Rename(tmpPath, fs.path)` — on any POSIX filesystem, a rename is atomic: it either completes fully or doesn't happen at all. There is no intermediate state where the file is half-written. A crash or power loss at any point before the rename leaves the *old* file intact; a crash after leaves the *new* file intact. There is no window where `cases.json` can be observed corrupted or truncated.

`writeMu sync.Mutex` serializes the disk-write side specifically — `MemStore` already handles its own in-memory concurrency (an `sync.RWMutex` guarding the map), so `writeMu` exists only to prevent two concurrent flushes from racing on the same temp-file path, not to protect the map itself.

**Reads never touch disk.** `Get` goes straight to `fs.mem.Get`. This means read latency is unaffected by write volume or disk speed, and it's why `MemStore`'s `snapshot()`/`loadAll()` helpers exist — deliberately *not* part of the public `Store` interface, since "give me every case" is a persistence-layer concern (loading at startup, flushing on write), not something `internal/api` should ever need to ask a generic `Store` for.

## `id.go`: why a hand-rolled UUID generator

Fifteen-ish lines of `crypto/rand` plus the two bit-twiddling lines that set the RFC 4122 version/variant bits, instead of a `google/uuid` dependency — see `../../../ARCHITECTURE.md` on why this backend has zero third-party Go dependencies at all. `crypto/rand.Read` failing is treated as `panic`-worthy, not a soft failure with a predictable fallback ID: a broken system entropy source is a serious enough condition that continuing with a guessable case ID (which, unlike most IDs in this system, is effectively also a bearer credential — anyone who has a case's URL can view it, per `internal/api/README.md`'s note on the intentional absence of a login system) would be worse than crashing loudly at that moment.

## `CaseRecord`: flat and JSON-native by design

Deliberately **not** a direct persistence of `tiering.Decision` or `response.FamilyResponse` — both of those types are shaped for their own packages' internal concerns (`FamilyResponse` in particular has every field unexported specifically so it can only be constructed via `response.Build`, see `internal/response/README.md`). `CaseRecord` is its own flat, exported, plain-JSON-serializable type, built once (in `internal/api/handlers.go`) from a `FamilyResponse`'s public accessor methods at the point a case is actually saved. This keeps `internal/store` from needing to import or understand either of those other packages' internal shapes, and means a change to how `FamilyResponse` is built internally never has a ripple effect on what gets persisted or on the on-disk JSON format.

## If you're extending this package

- **Adding a `PostgresStore` or similar**: satisfy the `Store` interface exactly as written; don't widen it unless every existing implementation (`MemStore`, `FileStore`) can reasonably support the new method too — the whole point of the interface is that `internal/api` and its tests don't need to know or care which implementation is behind it.
- **Adding a field to `CaseRecord`**: this is persisted, user-visible-eventually data — consider whether existing `cases.json` files (or `.tmp` files mid-flush) need a migration path, since `FileStore.NewFileStore` unmarshal-fails loudly (`"existing data file %q is corrupt"`) on a file it can't parse, which an incompatible schema change could trigger.
- **This file is part of the codebase's documentation convention** — see the repo root `README.md`. Keep it in sync with the code, in the same change.
