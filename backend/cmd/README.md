# `cmd`

Go convention: each subdirectory of `cmd/` is a separate `main` package producing one binary. This project has exactly one: `server/`, the backend's entrypoint. See `cmd/server/README.md`.

If this project ever needs a second binary (a one-off migration tool, a CLI for re-running a data transform without invoking Python — see `internal/hbp/data/README.md`), it goes in a new sibling directory here, e.g. `cmd/migrate/`, not inside `server/`. This is why `cmd/` exists as its own directory layer even with only one thing in it today — it's a namespace for entrypoints, kept separate from `internal/` (which is not a namespace for entrypoints, but for everything an entrypoint uses).
