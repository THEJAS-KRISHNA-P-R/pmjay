# `internal/retrieval`

One file, one job: turn a family's raw description text into a short, ranked list of plausible HBP packages *before* any paid LLM call happens. This is the cost lever for the whole system — see `../../../ARCHITECTURE.md`'s "why the retrieval pre-filter is a cost decision, not just a latency one." On the real ~1,900-code HBP master list, sending the entire dataset as LLM context on every query would make token cost scale with catalog size, not query complexity; this package is what keeps it scaling with query complexity instead.

## File

`retrieval.go` — that's the whole package. No subpackages, no other files (test coverage lives in `retrieval_test.go`, not listed separately here since it's exactly what you'd expect: cases exercising the scoring and fail-open behavior below).

## The two functions, and why they fail differently on purpose

### `Retrieve(ds, description) []Candidate` — fails open

Scores every package by keyword overlap, returns the top `MaxCandidates` (20), sorted highest-first. The one property worth understanding before touching this function: **if nothing scores above zero, it does not return an empty list.** It falls back to returning a bounded slice of the *entire* dataset instead.

This is deliberate, not a bug waiting to be "fixed." A retrieval layer that returns nothing on a genuinely novel description would silently turn "our keyword matching didn't recognize this phrasing" into "this isn't covered" by the time it reaches the family — exactly the overclaiming failure mode the rest of this codebase works hard to avoid (see `docs/SAFETY_DESIGN.md`). Keeping this layer recall-biased — happy to pass through too many candidates, never willing to decide on its own authority that nothing matches — means a wrong call here is a cost problem (the LLM reasons over a slightly larger list than necessary) rather than a correctness problem (a real match getting silently dropped before the one component that actually understands language ever sees it).

`ensureUnspecifiedIncluded` is the same philosophy applied to one specific case: the `UNSPECIFIED` catch-all package almost never wins on keyword overlap (nothing about its name looks like what a family would type — that's the point of it), so `Retrieve` explicitly appends it if the scoring pass didn't already surface it, rather than relying on it to compete fairly against packages tuned to match real phrasing.

### `RetrieveExclusions(ds, description) []ExclusionCandidate` — does not fail open, on purpose

Scores every exclusion category the same way, but returns the *entire* list every time, sorted by score, rather than truncating. The exclusion list is short (a handful of categories, not hundreds of packages), so there's no cost pressure to truncate it, and truncating it would risk a real applicable exclusion being cut before the LLM step gets to reason about it. Where `Retrieve` trades completeness for cost under a hard candidate cap, `RetrieveExclusions` doesn't have to make that trade at all — so it doesn't.

## Why this layer is "deliberately dumb" (the file's own words)

Plain English-lexicon keyword overlap — lowercased, tokenized on ASCII letters/digits, stopwords stripped, no stemming, no fuzzy matching, no embeddings. It does not attempt to understand Malayalam or English/Malayalam code-mixed phrasing, and that's correct, not a limitation to fix here: the LLM extraction step downstream (`internal/extract`) receives the family's full original text, in whatever language mix they actually used, and does the real language understanding. This package's only job is to not accidentally exclude the right answer from a shortlist — recall, not precision. Precision is `internal/extract`'s job. Keeping that boundary sharp (per the source spec's Section 58: "AI where genuine language ambiguity exists, deterministic logic everywhere else") is what keeps the one expensive, non-deterministic call in this whole pipeline small and cheap.

## Scoring, concretely

- A package's `CommonDescriptionKeywords` (see `internal/hbp/README.md` on why these are hand-tuned for 40 records and mechanically derived for 250) are checked as whole phrases — every token in the keyword phrase must appear in the description — and score `3 × phrase length` when they hit, specifically to reward a full phrase match ("baby on ventilator") over incidental single-word overlap.
- Individual words from the package name and specialty score 1 point each if present and not a stopword — a much weaker signal, there to catch descriptions that don't happen to use a curated phrase but do use the package's own terminology.
- Exclusions score the same way, against `Keywords` and `DisplayName`.

## If you're extending this package

- **Changing `MaxCandidates`**: this is a cost/recall tradeoff, not a magic number — raising it costs more per LLM call (more candidates to reason over) but very marginally reduces the already-small risk of a true match falling just outside the top 20. Document which way you're trading before changing it.
- **Adding a smarter scoring method** (embeddings, fuzzy matching, stemming): fine in principle, but preserve the fail-open property on `Retrieve` and the never-truncate property on `RetrieveExclusions` — those are the actual safety-relevant behaviors here, not the specific scoring formula.
- **This file is part of the codebase's documentation convention** — see the repo root `README.md`. Keep this file in sync with the code, in the same change, not as a follow-up.
