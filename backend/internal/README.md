# `internal`

Everything the `cmd/server` binary is built from. Go's `internal/` convention makes every package here unimportable from outside this module — nothing in this directory is a public API for anyone else to depend on, which is exactly right for an application (as opposed to a library): these packages exist to serve `cmd/server`, not to be reused elsewhere.

## The nine packages, in dependency order

```
hbp          — data & schema, depends on nothing else here
retrieval    — depends on: hbp
extract      — depends on: (nothing else here — deliberately isolated, see below)
tiering      — depends on: hbp, extract
response     — depends on: tiering
store        — depends on: (nothing else here)
document     — depends on: store
config       — depends on: (nothing else here)
api          — depends on: hbp, extract, response, retrieval, store, tiering, document
                (the only package that touches all the others)
```

This mirrors `../../ARCHITECTURE.md`'s system-shape diagram exactly, and that's not a coincidence — the package boundaries in this codebase *are* the architecture diagram, not a separate thing that happens to agree with it. A request flows `retrieval → extract → tiering → response → store`, and the Go import graph enforces that exact ordering: `tiering` can import `extract`'s types, but `extract` cannot import anything from `tiering` even by mistake, because `tiering` isn't a dependency `extract`'s `go.mod`-equivalent (there's only one module, but the import direction is what matters) would ever need. If you find yourself wanting an earlier-stage package to import a later-stage one, that's usually a sign the logic belongs in a different package, not a sign to add the import.

`document` sits after `store` for the same reason: it renders a `store.CaseRecord` into a PDF, so it depends on that type, but nothing upstream of `store` needs to know `document` exists — a request can be extracted, tiered, and answered without anyone ever asking for the PDF form of the result.

## Why `extract` is deliberately isolated

It's the *only* package here that calls a language model, and the only one with a real, non-Go dependency (the Anthropic API, over the network). Every package downstream of it (`tiering`, `response`) depends on `extract`'s *types* (`Result`, `PendingSignal`, etc.) but never on `extract` needing to actually run a network call to be tested — `extract.FakeClient` satisfies the same `Extractor` interface `api.Server` depends on, so the entire rest of the system is testable without a live API key. See `internal/extract/README.md` and `../../ARCHITECTURE.md`'s "why the extraction/tiering boundary is exactly where it is."

## One package per architectural concern, not per Go idiom

Some of these packages are small (`config`, `retrieval` are each a single file). They're still separate packages, not merged into a shared `internal/util` or folded into `api`, because each one represents a distinct question this system has to answer, and keeping them separate is what makes each question independently testable and independently reviewable:

| Package | The one question it answers |
|---|---|
| `hbp` | What does the government actually say about this package, and how confident are we? |
| `retrieval` | Which packages could this description plausibly be about, cheaply? |
| `extract` | What did the family actually mean, understood properly (the one place language ambiguity is resolved)? |
| `tiering` | Given that understanding, what's the honest, safety-checked verdict? |
| `response` | What, exactly, does the family get told — worded how? |
| `store` | How does this case's state survive between requests? |
| `document` | What does the family actually walk away with — something real to hand to a person, not just words on a screen? |
| `config` | What does this specific deployment look like? |
| `api` | How does all of the above become an HTTP request/response? |

## If you're adding a tenth package

Ask which of the questions above it's actually answering — if it's a genuinely new concern (not a variation on an existing one), it probably deserves its own package following this same pattern: a focused doc comment on the package's own primary file explaining *why* it exists and where its boundaries are, not just *what* it contains. Every existing package here follows that convention (look at the top of `retrieval.go`, `tier.go`, `types.go` in `response` and `extract`, `store.go`, or `winansi.go` in `document`) — it's cheap to write once, at the point when the reasoning is freshest, and expensive to reconstruct later from git history.

This directory-level file is part of the codebase's documentation convention — see the repo root `README.md`, and each individual package's own `README.md` for the deep-dive on that package specifically. If you add a tenth package, add a row to the table above and a line to the dependency graph, in the same change that adds the package.
