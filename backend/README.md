# `backend`

The Go backend — a single static binary, zero third-party dependencies, that serves the PMJAY Point-of-Denial Advocate's API. If you're new to this codebase, read `../ARCHITECTURE.md` first (the system-shape diagram there is the map everything below fits into), then come back here for how the code on disk maps to that architecture.

## Layout

```
backend/
├── cmd/server/       entrypoint — see cmd/README.md, cmd/server/README.md
├── internal/         everything the binary is built from — see internal/README.md
│   ├── hbp/          data & schema layer
│   │   └── data/     the actual JSON dataset + the scripts that built it
│   ├── retrieval/     zero-cost keyword pre-filter
│   ├── extract/       the one package that calls a model
│   ├── tiering/       deterministic decision logic
│   ├── response/      builds the actual words a family reads
│   ├── store/          case persistence, no database
│   ├── document/        renders a case as a downloadable/printable PDF
│   ├── config/         env-var configuration
│   └── api/            HTTP layer tying all of the above together
├── Dockerfile         multi-stage build → distroless runtime image
├── Makefile           build/test/lint targets — see below
├── .env.example       every env var this backend reads, kept in sync with internal/config
└── go.mod             module github.com/pmjay-advocate/backend, go 1.22.2, zero dependencies
```

Every one of the `internal/` subdirectories has its own `README.md` with the actual depth — architecture decisions, why each threshold or design choice is what it is, what to check before extending it. This file is the map; those are the territory.

## Quickstart

```bash
go build ./...              # compiles clean, zero external deps to fetch
go vet ./...
gofmt -l .                   # should print nothing
go test ./... -cover -race   # 268 tests, 10 packages, as of this writing
```

Or via `make check` — see the `Makefile`, which wraps exactly this sequence plus `build`/`build-arm64`/`build-amd64`/`clean` targets. No `ANTHROPIC_API_KEY` is required for any of the above; see `docs/TESTING.md` for why the whole suite runs fully offline.

To actually run the server:

```bash
cp .env.example .env    # then fill in ANTHROPIC_API_KEY if you want real case intake to work
go run ./cmd/server
curl localhost:8080/api/v1/health
```

## Why zero third-party dependencies, briefly

Discovered, not planned — the original build environment's network egress couldn't reliably reach Go's module proxy. Rather than fight it, every dependency that would normally come from a library is a small, well-understood piece of stdlib code instead (routing: Go 1.22's `http.ServeMux`; UUIDs: 15 lines of `crypto/rand`; the Anthropic client: a direct `net/http` call; PDF generation: a from-scratch writer in `internal/document`, feasible specifically because the real content it renders turned out to need only two non-ASCII characters — see that package's own `README.md`). The result — a fully static binary, 5.7 MB stripped for `linux/arm64` (the actual deployment target — see `../ARCHITECTURE.md`), zero supply-chain surface — turned out to be a genuine engineering win, not just a workaround. Full story: `../ARCHITECTURE.md`.

## Testing philosophy in one sentence

Everything downstream of the one LLM call (`internal/extract`) is plain, deterministic Go, which is why `internal/tiering` and `internal/response` — the two packages that decide and phrase what a family is told — sit at 94–96% coverage, and why the entire suite runs without a live API key (`extract.FakeClient` stands in). See `docs/TESTING.md` for the full breakdown, including what's *not* code-testable at all (the field-study-level questions no amount of unit testing can answer).

## This directory is part of the codebase's documentation convention

Every folder and subfolder in this repository has a `README.md` explaining what it does and why it's built the way it is — not just what's in it, but the reasoning a future engineer (human or AI) would otherwise have to reconstruct from git history or from asking around. **If you add a new folder anywhere in this backend, give it a `README.md` in the same change, not as a follow-up.** See the repository root `README.md`'s "How this codebase documents itself" section for the full convention and why it exists.
