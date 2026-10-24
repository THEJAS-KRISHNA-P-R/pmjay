# Data provenance and limits

Read this before quoting any number.

## What is in the skill

| File | Records | What is verified |
|---|---:|---|
| `hbp-packages.json` | 300 | Package code, name, specialty, and rate checked against a published government HBP schedule. |
| `hbp-packages-unverified.json` | 15 | Real, published PMJAY procedure names only. Codes and rates are unknown. Every rate field has been removed. |
| `exclusions.json` | 4 | Four confirmed exclusion categories, with source notes. |

The full HBP master list contains roughly **1,900 procedures**. This skill
covers 315 names, of which 300 have a verified code and rate. It is a useful
reference, not a complete tariff schedule.

## What "verified" means

A verified record's package code, name, specialty, and rate were checked
against a published government source during the dataset build, with the
source recorded in that record's `source_note`.

"Verified" does **not** mean:

- the rate is current today;
- the rate applies to every hospital or every city tier;
- the patient is eligible;
- the hospital is empanelled for that specialty;
- the scheme will approve the pre-authorisation.

Always confirm the current rate with the hospital or **14555** before relying
on it.

## Rate shapes

The dataset uses three shapes. Do not flatten them into one number.

1. **Flat** — one indicative package rate.
2. **Tiered** — a range from the lowest city tier to the highest. The family's
   hospital tier decides the actual figure. Do not quote the top of the range
   as "the rate."
3. **Per-diem** — a per-day rate stratified by ward level: routine ward, HDU,
   ICU without ventilator, ICU with ventilator. Do not present the routine-ward
   figure as the total cost of an admission.

## Version currency

PMJAY has published HBP 2.1, HBP 2022, and later revisions. Rates and even
eligibility criteria can change between versions. The dataset resolved a
specific HBP 2.1-versus-2022 question for a set of Cardiology records, and the
source notes for individual records say which schedule they were checked
against. If a family is disputing a current bill, ask the hospital which HBP
version and city tier it is applying. Do not assume the version in this
dataset is the one in force at that hospital today.

## Coverage gaps

Fifteen records remain unverified, concentrated in seven specialties:

- General Surgery
- Orthopaedics
- Obstetrics & Gynaecology
- Ophthalmology
- ENT
- Urology
- Neurosurgery

These include common procedures such as gallbladder removal, appendectomy,
hernia repair, caesarean section, cataract surgery, and knee replacement. The
skill can recognise the names, but it cannot quote a verified code or rate for
them. Treat those matches as **Amber**, not Green.

The remaining 300 verified records are concentrated in General Medicine,
Medical Oncology, Pediatric Medical Management, High End Medicine,
Interventional Neuroradiology, High End Diagnostics, High End Procedures,
Cardiology, Neonatal Care, Emergency Medicine, Burns, Mental Disorders, and
Infectious Diseases.

## Where the data came from

The dataset was compiled from publicly available National Health Authority
(NHA) and State Health Agency materials, including published HBP master lists
and rate schedules. Individual records carry a `source_note` naming the
specific source and, where applicable, the fetch date and the version used.

The exclusion categories are carried from public government replies and NHA
materials as described in each record's `source_note`, and were not all
independently re-verified from the primary record.

This skill is an independent compilation. It is **not affiliated with,
endorsed by, or operated by** the National Health Authority, the Ministry of
Health and Family Welfare, any State Health Agency, or any hospital.

## Legal and medical status

This is guidance, not a legal ruling and not a medical opinion. It does not
create an advocate-client relationship. A NALSA Para Legal Volunteer (**15100**)
or the PMJAY helpline (**14555**) can confirm the specific case, and a doctor
decides what treatment is medically necessary.

The care-first rule is absolute: do not delay, cancel, or refuse treatment to
argue about a bill.

## Data corrections

If a rate or package code is wrong, correct it at the source and regenerate
the reference files rather than editing the JSON by hand. The repository's
`skills/tools/build_data.py` rebuilds the skill data from the application's
dataset. A hand-edited number has no provenance and should not be trusted.

## Last updated

Dataset last materially updated: **August 2026**. Skill packaging:
**October 2026**. Check the repository's `docs/DATA_SOURCES.md` for the
current, detailed provenance trail before treating any figure as authoritative.
