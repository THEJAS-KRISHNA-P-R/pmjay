# `internal/hbp`

The data layer. This package owns the shape of a "PMJAY package" and an "exclusion category," loads them from the JSON files in `data/`, and validates every record at load time so a malformed or dishonest record fails the build instead of reaching a family. Nothing downstream — retrieval, extraction, tiering, response — talks to the JSON files directly; everything goes through the `Dataset` this package produces.

## Files

| File | What it does |
|---|---|
| `types.go` | Defines `Package`, `PerDiemRates`, `Exclusion`, and `Dataset` — the schema. No behavior, just shape and (extensively) the *reasoning* behind each field, written as doc comments because this is the schema a future session is most likely to extend, and the reasoning matters more than the field names. |
| `loader.go` | `Load()` / `MustLoad()` — parses the embedded JSON, then runs `validate()` against every record before returning. A dataset that fails validation panics at startup (`MustLoad`, used by `cmd/server`), not partway through serving a family. |
| `loader_test.go` | Tests for the loader and the original validation rules (non-empty fields, unique codes, a sane package-count ceiling). |
| `loader_ratetype_test.go` | Tests specifically for the `rate_type`/`per_diem_rates`/`rate_max_inr` validation added in the August 2026 session — kept in its own file rather than merged into `loader_test.go` so the two additions to this package (original schema, schema extension) stay separately reviewable. |
| `data/` | The actual JSON files and the Python scripts that produced them. See `data/README.md` — deliberately a separate, longer document, because *what's in the data* and *how confident you should be in any given number* is a bigger subject than *how the Go code loads it*. |

## The schema, and why it looks the way it does

### `Package`

The fields that exist because a **family's safety** depends on them, not because the data happened to have them:

- **`Verified bool`** — the single most important field in this codebase. `internal/response`'s `packageCitation()` refuses to state a specific rupee figure to a family unless this is `true`. Every placeholder/seed record has this `false` and a `source_note` saying so explicitly. This is not a convention that relies on every future contributor remembering to check it — it's enforced by `backend/internal/response/builder_test.go`'s `TestBuild_UnverifiedRateNeverStatedAsFact`, and by `loader.go` refusing to load a record with an empty `source_note` at all (see `validate()`).
- **`RequiresPreauth bool`** — feeds `internal/tiering`'s pending-vs-denied logic. A package that doesn't require pre-authorization can't be legitimately denied *for lack of pre-authorization*, which is one of the more common wrongful-denial patterns this whole tool exists to catch.
- **`CommonDescriptionKeywords []string`** — the vocabulary `internal/retrieval` matches against. Hand-tuned for the original 40 records (real phrases a family might actually say — "baby on ventilator," not "neonatal respiratory support"); mechanically derived for the 250 records added in August 2026 (see `data/README.md`'s note on why that's a real, acknowledged quality gap, not an oversight).

### `RateType`, `RateMaxINR`, `PerDiemRates` — added August 2026, additive

The original schema assumed every package has exactly one number: `IndicativeRateINR`. That's true for a flat surgical package, but false for two real shapes HBP actually uses:

1. **A rate that depends on the treating hospital's city-tier classification** (HBP 2022's Tier1/X, Tier2/Y, Tier3/Z stratification — metro vs. mid-size vs. small-town). `RateType: "tiered"` records store the floor in `IndicativeRateINR` and the ceiling in `RateMaxINR`. The response layer renders both, as a range — it never picks one tier to state as *the* number, because this tool has no way to know which tier a specific family's hospital falls under, and guessing would be a fabricated fact wearing a real number's clothes.
2. **A rate charged per day of admission, stratified by ward level**, not as one total (General Medicine and Pediatric admission packages — fevers, infections, strokes managed medically rather than surgically). `RateType: "per_diem"` records store the Routine Ward figure in `IndicativeRateINR` (so old code that only knows about a single flat rate still sees a sane number) and the full four-level breakdown in `PerDiemRates`. **This is the one field in this whole codebase where getting the citation wrong is easiest and most dangerous**: collapsing a per-day rate into a sentence that reads like a total admission cost is exactly the kind of confidently-wrong number this tool exists to prevent a family from being told. `internal/response/templates_test.go`'s `TestPackageCitation_PerDiemNeverImpliesATotal` is the regression test for this specific property — if you touch `packageCitation()`, run that test first.

Both extensions are fully additive: every record from the original 40 has `RateType == ""` (the zero value), and `validate()`'s `switch p.RateType` treats `""` as "flat rate, no additional invariants" — nothing about how those 40 records load, validate, or render changed when this was added.

### `Exclusion`

Separate from `Package` because exclusions are categories of *reason a denial might be legitimate* (cosmetic procedures, pre-existing conditions outside the waiting period, etc.), not things a family is trying to match against. `internal/tiering` checks a case's described situation against these before deciding a denial looks wrongful — see `docs/SAFETY_DESIGN.md` for how that check fits into the overall decision logic, and `docs/DATA_SOURCES.md` for exactly which exclusion categories are verified against the primary Rajya Sabha source versus the source spec's own citation of it (still an open item — see that doc's priority list).

## Why validation happens here, not in the JSON

`validate()` in `loader.go` enforces, at `go:embed`-load time (i.e. compiled into the binary, checked once at process startup, not per-request):

- Every package/exclusion has non-empty required fields, including `source_note` — a record with no explanation of where its number came from cannot load at all.
- Package codes are unique (a duplicate would make retrieval's matching ambiguous in a way that's hard to debug from the outside).
- `rate_type`'s three shapes (`""`, `"tiered"`, `"per_diem"`) each satisfy their own internal-consistency rules (tiered max genuinely exceeds the floor; per-diem levels strictly increase; per-diem's routine-ward figure matches the top-level `IndicativeRateINR` so the two never silently drift apart).
- A package-count ceiling (currently 400) that exists specifically so the dataset can't silently balloon without the growth being reviewed — see the comment on `TestLoad_PackagesNonEmptyAndReasonableSize` for the full reasoning and history of that number.

The alternative — trusting the JSON and letting a malformed record surface as a bug in whatever family happens to trigger that code path — is exactly backwards for a tool whose entire purpose is being trustworthy about specific numbers. A validation failure here is loud, at startup, before any traffic is served.

## If you're extending this package

- **Adding a genuinely new rate shape** (a third kind, beyond flat/tiered/per-diem): extend the `switch p.RateType` in `validate()` first, with a test for both the "well-formed" and "malformed" case, before touching `data/*.json`. `packageCitation()` in `internal/response/templates.go` needs the matching branch, or the new shape will silently fall through to flat-rate rendering — see that file's own `switch p.RateType` and keep the two in sync.
- **Adding new package records**: see `data/README.md` first. There's a real, working convention (a written, reviewable transform script, not hand-edited JSON) that this session established specifically because hand-editing 250 records invites transcription errors in exactly the numbers a family would be told.
- **This file is part of the codebase's documentation convention** — see the repo root `README.md`'s "How this codebase documents itself" section. If you add a new file to this package that changes what a reader needs to know, update this file in the same change, not as a follow-up.
