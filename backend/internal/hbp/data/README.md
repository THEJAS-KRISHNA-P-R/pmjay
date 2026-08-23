# `internal/hbp/data`

The dataset itself, and the scripts that built it. `loader.go` reads `hbp_packages.json` and `exclusions.json` via `go:embed` — this folder's contents are compiled directly into the binary at build time, not read from disk at runtime. There is no database and no runtime data-loading step; if you want to change what packages this tool knows about, you edit the files in this folder and rebuild.

## Files

| File | What it is |
|---|---|
| `hbp_packages.json` | The actual dataset: 315 package records as of this writing, 300 of them `verified: true` against a real government source. This is the file that matters at runtime. |
| `exclusions.json` | Categories of legitimate denial reason (cosmetic procedures, waiting-period exclusions, etc.) — much smaller, hasn't changed since the original build. |
| `transform.py` | The **first** data-population script (14 August 2026 session). Corrected two placeholder records to real verified data, removed one placeholder, and added 20 new real records (Neonatal Care's five-tier structure, Burns, Emergency Medicine, Cardiology device-closure procedures, Mental Health/ECT). Kept in the repo, not deleted after running, so a specific number can be traced back to exactly where it came from without trusting a diff. |
| `transform_2022_additions.py` | The **second** data-population script (August 2026 continuation session). Added 250 new records — General Medicine and Pediatric per-diem diagnoses, Medical Oncology package-level entries, three new "High End" specialty categories, Interventional Neuroradiology — and retired one superseded placeholder. Same audit-trail reasoning as `transform.py`. |
| `reconcile_2022_rates.py` | The **third** data script (19 August 2026 session) — different in kind from the first two, not just in date: it *updates* seven existing Cardiology records in place (resolving the HBP 2.1-vs-2022 version question, see `docs/DATA_SOURCES.md`) as well as adding 26 new General Medicine/interventional records. Its docstring is the canonical record of why two related records reached the same session (`MG098A`, `MG0119A`) were deliberately left out. |
| `hbp_packages.json.bak` | A frozen snapshot of the dataset from *before* `transform.py` ran (20 placeholder records). Not regenerated, not read by any code — kept purely as a historical reference point for "what did this look like before any real verification happened." |

## Why two transform scripts, not one, and not hand-edited JSON

Every one of the 296 records these three scripts added, updated, or corrected represents a specific rupee figure that could end up in a sentence read by a family disputing a medical bill. Hand-editing a 315-record JSON file directly — even carefully — has no audit trail: a reviewer (human or a future session) looking at the final JSON has no way to tell *where a number came from* without re-deriving it from scratch. A script that transcribes from a cited source, with the source URL and fetch date in a constant at the top of the file, gives a different property: the diff between "no such record" and "record with rate ₹X" is a few lines of Python that name their source, not an opaque JSON edit.

Three scripts, not one merged file, because they represent three genuinely separate research sessions with different sources and, for the third, a different *kind* of change (in-place updates as well as additions) — see `docs/DATA_SOURCES.md` for the full story of each, including the version-currency question the second script's own additions surfaced about the first script's numbers, and how the third resolved it. Merging them into one file would blur which specific fetch a given number traces back to.

**If you add a fourth wave of real data**, follow the same pattern: a new `transform_YYYYMMDD_description.py` (or similarly named) file, not an edit to any existing script and not a hand-edit of the JSON. Put your source URL, fetch date, and exactly what you did and didn't include — including what you deliberately left out because the source data was incomplete or ambiguous, the way all three existing scripts do — in a comment block at the top. This is now the established convention for this folder; keep it going.

## What "verified" actually means here, concretely

`docs/DATA_SOURCES.md` is the canonical, detailed answer — read it before trusting any specific number, and definitely before removing the `verified` field from anywhere in the codebase. The short version, specific to what's physically in this folder:

- A `verified: true` record's `package_code`, `package_name`, `specialty`, and rate figure(s) were checked against an actual published government rate schedule during a build session, with the exact source URL and fetch date recorded in that record's `source_note`.
- A `verified: false` (`SEED-`-prefixed code) record has a real, published PMJAY package *name* — not invented — but a placeholder code and rate, clearly marked as such.
- 15 placeholder records remain (`SEED-CARD-002`, `SEED-CARD-003`, `SEED-ENT-001`, `SEED-GS-001/002/003`, `SEED-NEURO-001`, `SEED-OBG-001/002/003`, `SEED-OPH-001`, `SEED-ORTHO-001/002/003`, `SEED-URO-001`) — General Surgery, Orthopaedics, Obstetrics & Gynaecology, Ophthalmology, ENT, Urology, and Neurosurgery, the seven specialties that four separate real-data extraction attempts across three sessions have each hit the same kind of hard wall trying to reach. That's not a gap from lack of trying — see `docs/DATA_SOURCES.md`'s wall-finding sections for exactly what was tried, including the 19 August session's specific, still-open lead (the wall in one source now sits much further in than before — right after Medical Oncology — and whether these seven specialties exist just past it was never actually checked).

## A known, explicitly-scoped gap: keyword quality on the 250 August-2022-additions records

The original 40 records (and the 20 `transform.py` added) have hand-tuned `common_description_keywords` — real phrases a Kerala family might actually type ("baby on ventilator," "leg artery blockage"). The 250 records `transform_2022_additions.py` added use `keywords_from_name()`, a cheap, deterministic word-extraction function (lowercase words over 3 characters, common words filtered out, capped at 8) rather than hand-tuned phrases, for the practical reason that hand-tuning 250 records' worth of retrieval keywords in the time available would have meant either much shallower data coverage or much slower, more error-prone review of the actual rate figures — and the rate figures are the part a mistake in is dangerous. This is a real, acknowledged quality gap in retrieval recall for those 250 records, not an oversight — see `transform_2022_additions.py`'s own `keywords_from_name()` docstring, and treat "hand-tune retrieval keywords for the August 2026 additions" as legitimate, scoped future work if you're looking for something concrete to improve.

## Regenerating the dataset from scratch

You generally shouldn't need to — the checked-in `hbp_packages.json` already reflects both scripts having been run. If you do need to (e.g. testing a change to one of the transform scripts), the scripts mutate `hbp_packages.json` in place, so restore from a known-good copy first:

```bash
# transform.py assumes a 20-placeholder starting point; there is no
# separate "before" snapshot for transform_2022_additions.py checked in
# — hbp_packages.json.bak only covers the state before transform.py.
# If you need to re-run transform_2022_additions.py specifically, back
# up the current hbp_packages.json yourself first.
cp hbp_packages.json /tmp/hbp_packages_backup.json
python3 transform_2022_additions.py
# On failure or to start over:
cp /tmp/hbp_packages_backup.json hbp_packages.json
```

Both scripts include hard collision/duplication checks (`assert code not in EXISTING_CODES`) that abort with `SystemExit` before writing anything if a new record's code would collide with an existing one — this is what makes it safe to re-run against an unexpected starting state without silently corrupting the dataset.
