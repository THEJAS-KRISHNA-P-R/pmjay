#!/usr/bin/env python3
"""Resolves the HBP 2.1-vs-2022 rate currency question the previous
session deliberately left open (see docs/DATA_SOURCES.md, "An open
question this session found but deliberately did not resolve"), and adds
the General Medicine records that session's Source B extraction reached
but could not use because they didn't fit the standard four-level
per-diem pattern.

WHY THIS IS SAFE TO DO NOW, WHERE THE PREVIOUS SESSION CORRECTLY DIDN'T:

1. Version currency: confirmed via three independent, mutually
   corroborating sources — an NHA-affiliated official document catalog
   (snomedct.abdm.gov.in) listing "National Master Health Benefit
   Packages 2.1" and "HBP 2022 package master and OM" as two separate,
   sequential publications; a PIB press release (6 Oct 2021) describing
   an intermediate "HBP 2.2" revision with none of HBP 2022's
   distinguishing features; and independent news coverage (Tribune
   India, ourindia.com) of NHA's own 7-8 April 2022 Mahabalipuram launch
   event, which both independently describe HBP 2022 as introducing
   "differential pricing... for the first time" by city-tier — the exact,
   distinguishing structural signature Source B uses. This is no longer
   circumstantial.

2. Criteria stability: the previous session's caution was specifically
   about clinical *criteria* changing alongside prices (the BM001B Burns
   Management TBSA-threshold example). This session re-fetched Source B
   directly and confirmed all seven affected Cardiology codes are simple
   single-criterion procedure packages in HBP 2022 — no TBSA-style
   eligibility bands, no criteria text to compare at all. The
   criteria-shift risk that blocked the previous session genuinely does
   not apply to these seven records specifically.

WHAT THIS SCRIPT DOES:

  A. Updates MC005A, MC007A, MC008A, MC009A, MC015A, MC016A, MC017A in
     place — the seven Cardiology records flagged in DATA_SOURCES.md.
     Each becomes a "tiered" record (see hbp/types.go's RateType doc
     comment): indicative_rate_inr = Tier3(Z) total, rate_max_inr =
     Tier1(X) total, where "total" = HBP 2022's procedure fee plus its
     separately-listed implant/device cost, matching how the *existing*
     HBP 2.1 figures were already structured (one bundled total — see
     each record's original confidence_notes). The old HBP 2.1 figure is
     preserved in confidence_notes, not discarded, per DATA_SOURCES.md's
     own instruction for how to do this.

  B. Adds 25 new General Medicine / interventional records this
     session's fetch reached in full: MG072B-D, MG073A, MG074A-B,
     MG075A, MG076A, MG077A, MG082A-085A, MG097A, MG099A, MG0105A-0108A,
     MG0120A-G. All are genuinely tiered (Tier3 < Tier1) in Source B.

WHAT THIS SCRIPT DELIBERATELY DOES NOT DO:

  MG098A (PET scan) and MG0119A (Drug resistant epilepsy) were both
  reached in this session's fetch but are NOT included, on purpose:

  - MG098A's ₹0/₹0/₹0 tiered columns are a red herring — the real prices
    are itemized by SCAN TYPE, not city-tier ("Whole body 20500, Brain
    14650, Gallium 68 Peptide 15000, Cardiac 1500"). Representing this
    with rate_type="tiered" would make packageCitation() (see
    response/templates.go) tell a family this varies "depending on the
    treating hospital's city-tier classification" — which is simply
    false for this package, and false in the specific way this whole
    tool exists to avoid: a confident-sounding wrong explanation.

  - MG0119A's note is a bare "2100/day" with no ward/HDU/ICU breakdown.
    rate_type="per_diem" requires a full PerDiemRates struct (see
    hbp/types.go) — inventing the missing HDU/ICU figures by assuming
    they match the standard PER_DIEM pattern every other General
    Medicine diagnosis uses would be exactly the "guessed, not verified"
    move this dataset's whole discipline exists to avoid. Leaving
    rate_type unset and citing ₹2,100 as a flat "indicative rate" would
    be worse: it would silently drop the "/day" and present a daily rate
    as if it were the total cost of care.

  Both are real, correctly-scoped gaps for a future session with time to
  either find the missing breakdown or make a considered schema decision
  — not something to paper over here. See docs/DATA_SOURCES.md.

Run from backend/internal/hbp/data/: python3 reconcile_2022_rates.py
"""
import json

SOURCE_B_URL = "https://ayushmanup.in/assets/doc/HBP-2022.pdf"

VERSION_CONFIRMATION_NOTE = (
    "HBP 2022 (Source B) confirmed to supersede HBP 2.1 as of this "
    "session: an NHA-affiliated official document catalog "
    "(snomedct.abdm.gov.in/hospital/hbc) lists \"National Master Health "
    "Benefit Packages 2.1\" and \"HBP 2022 package master and OM\" as "
    "separate, sequential publications; NHA's own 7-8 April 2022 "
    "Mahabalipuram launch event (independently reported by Tribune India "
    "and ourindia.com) described HBP 2022 as introducing differential, "
    "city-tier pricing \"for the first time\" — the exact structural "
    "signature Source B uses, distinguishing it from the intermediate "
    "\"HBP 2.2\" (Nov 2021, per PIB press release 6 Oct 2021) which "
    "revised rates by a flat percentage with no tiering. See "
    "docs/DATA_SOURCES.md."
)


def cardiology_source_note(old_rate):
    return (
        f"UPDATED this session: rate revised from the HBP 2.1 figure "
        f"(\u20b9{old_rate:,}, verified in an earlier session) to HBP 2022's "
        f"tiered structure, verified directly against {SOURCE_B_URL} "
        f"(fetched this session). {VERSION_CONFIRMATION_NOTE} This "
        f"record's HBP 2022 procedure fee and implant/device cost are "
        f"listed separately in the source and summed here into one "
        f"bundled total per tier, matching how the superseded HBP 2.1 "
        f"figure was already structured (see this record's prior "
        f"confidence_notes, preserved below)."
    )


def new_record_source_note():
    return (
        f"Package code, name, and rate verified directly against "
        f"{SOURCE_B_URL} (fetched this session, full document reached — "
        f"no extraction wall hit this time, unlike the previous "
        f"session's attempt). {VERSION_CONFIRMATION_NOTE}"
    )


def keywords_from_name(name):
    import re
    words = re.findall(r"[a-zA-Z]+", name.lower())
    seen, out = set(), []
    for w in words:
        if len(w) > 3 and w not in seen and w not in {
            "with", "without", "acute", "chronic", "syndrome", "disease",
        }:
            seen.add(w)
            out.append(w)
        if len(out) >= 8:
            break
    if not out:
        out = [w for w in words if len(w) > 2][:3]
    return out


def make_new_tiered_record(code, name, specialty, low, high, extra_note=""):
    assert high > low, f"{code}: expected a real tiered range, got low={low} high={high}"
    return {
        "package_code": code,
        "package_name": name[:1].upper() + name[1:],
        "specialty": specialty,
        "indicative_rate_inr": low,
        "rate_max_inr": high,
        "rate_type": "tiered",
        "requires_preauth": True,
        "common_description_keywords": keywords_from_name(name),
        "confidence_notes": extra_note,
        "source_note": new_record_source_note(),
        "verified": True,
    }


# ---------------------------------------------------------------------
# A. Cardiology updates — (procedure_low, procedure_high, implant_flat)
#    from HBP 2022 (Source B), fetched and transcribed this session.
# ---------------------------------------------------------------------
CARDIOLOGY_UPDATES = {
    "MC005A": {"proc_low": 44700, "proc_high": 53600, "implant": 55000},
    "MC007A": {"proc_low": 46200, "proc_high": 55400, "implant": 62000},
    "MC008A": {"proc_low": 47400, "proc_high": 56900, "implant": 72000},
    "MC009A": {"proc_low": 40800, "proc_high": 48900, "implant": 30000},
    "MC015A": {"proc_low": 30700, "proc_high": 36800, "implant": 45000},
    "MC016A": {"proc_low": 41300, "proc_high": 49500, "implant": 75000},
    "MC017A": {"proc_low": 43200, "proc_high": 51800, "implant": 21000},
}

# ---------------------------------------------------------------------
# B. New records — (name, specialty, tier3_low, tier1_high, note)
# ---------------------------------------------------------------------
NEW_RECORDS = [
    ("MG072B", "Peritoneal dialysis (catheter review)", "General Medicine, Nephrology", 500, 875,
     "Consumables ₹1,000 listed separately in source; not included in this figure."),
    ("MG072C", "Acute haemodialysis", "General Medicine, Nephrology, Organ transplantation", 1500, 1800, ""),
    ("MG072D", "Chronic haemodialysis", "General Medicine, Nephrology, Organ transplantation", 1500, 1800, ""),
    ("MG073A", "Plasmapheresis", "General Medicine, Pediatric Medical Management", 2000, 2300, ""),
    ("MG074A", "Whole blood transfusion", "General Medicine, Pediatric Medical Management", 2000, 2300, ""),
    ("MG074B", "Blood component transfusion (RDP, PC, SDP)", "General Medicine, Pediatric Medical Management", 2000, 2300,
     "Includes platelet transfusion."),
    ("MG075A", "High-end radiological diagnostic (CT, MRI, nuclear imaging)", "General Medicine, Pediatric Medical Management, Cardiology", 5000, 5800, ""),
    ("MG076A", "High-end histopathology and advanced serology investigations", "General Medicine, Pediatric Medical Management, Cardiology", 5000, 5800,
     "Includes biopsies."),
    ("MG077A", "Continuous renal replacement therapy / continuous veno-venous hemofiltration", "Pediatric Medical Management, General Medicine, Nephrology", 33000, 45000,
     "Per-day initiation cost for disposables, maximum 5 days in one admission — not a flat per-admission total."),
    ("MG082A", "Bone marrow aspiration or biopsy", "Interventional radiology, General Medicine, Pain specialists", 1200, 1600, ""),
    ("MG083A", "Lumbar puncture", "Interventional radiology, General Medicine, Pain specialists", 100, 200, ""),
    ("MG084A", "Joint aspiration", "Interventional radiology, General Medicine, Pain specialists", 200, 300, ""),
    ("MG085A", "DVT pneumatic compression stockings (ICU add-on)", "Interventional radiology, General Medicine", 900, 1200,
     "Add-on package for ICU admission, not a standalone treatment."),
    ("MG097A", "Endobronchial ultrasound guided fine needle biopsy", "General Medicine, Pulmonology (interventional)", 15700, 18000, ""),
    ("MG099A", "Platelet pheresis", "General Medicine, Pediatric Medical Management", 11000, 12700, ""),
    ("MG0105A", "Pulmonary thromboembolism (add-on, including thrombolytic therapy)", "General Medicine (high-end drugs)", 25000, 28800,
     "Add-on package — see MG0109A for the base per-diem admission package."),
    ("MG0106A", "Diffuse alveolar hemorrhage associated with SLE / vasculitis / GP syndrome", "General Medicine (high-end drugs)", 136000, 156400, ""),
    ("MG0107A", "Severe or refractory vasculitis", "General Medicine (high-end drugs)", 75000, 86300, ""),
    ("MG0108A", "Acute liver failure / fulminant hepatitis", "General Medicine", 50000, 57500, ""),
    ("MG0120A", "Comprehensive medical rehabilitation for spinal injury, TBI, CVA, or cerebral palsy", "General Medicine, Neurology, Pediatric Medical Management", 25000, 28800,
     "With or without orthosis."),
    ("MG0120B", "Comprehensive medical rehabilitation for complications of a specified disability", "General Medicine, Neurology, Pediatric Medical Management", 35000, 40300,
     "Includes chemodenervation, with or without orthosis."),
    ("MG0120C", "Single-event multiple-level surgery for spasticity management in cerebral palsy", "General Medicine, Neurology, Pediatric Medical Management", 15000, 17300, ""),
    ("MG0120D", "Medical rehabilitation of muscular dystrophy", "General Medicine, Neurology, Pediatric Medical Management", 7000, 8100, ""),
    ("MG0120E", "Medical rehabilitation for intellectual disability", "General Medicine, Neurology, Pediatric Medical Management", 7000, 8100, ""),
    ("MG0120F", "Medical rehabilitation for specific learning disability", "General Medicine, Neurology, Pediatric Medical Management", 7000, 8100, ""),
    ("MG0120G", "Medical rehabilitation for multiple disability", "General Medicine, Neurology, Pediatric Medical Management", 7000, 8100, ""),
]


def main():
    path = "hbp_packages.json"
    with open(path) as f:
        data = json.load(f)

    by_code = {p["package_code"]: p for p in data}

    updated = 0
    for code, r in CARDIOLOGY_UPDATES.items():
        rec = by_code.get(code)
        assert rec is not None, f"expected existing record {code} not found"
        old_rate = rec["indicative_rate_inr"]
        old_notes = rec.get("confidence_notes", "")

        low = r["proc_low"] + r["implant"]
        high = r["proc_high"] + r["implant"]
        assert high > low

        rec["indicative_rate_inr"] = low
        rec["rate_max_inr"] = high
        rec["rate_type"] = "tiered"
        rec["confidence_notes"] = (
            f"{old_notes} Previously listed as a flat ₹{old_rate:,} "
            f"under HBP 2.1 — see source_note for why this changed to a "
            f"tiered ₹{low:,}–₹{high:,} range."
        ).strip()
        rec["source_note"] = cardiology_source_note(old_rate)
        rec["verified"] = True
        updated += 1

    codes_seen = set(by_code.keys())
    added = 0
    for code, name, specialty, low, high, note in NEW_RECORDS:
        assert code not in codes_seen, f"collision: {code} already exists"
        data.append(make_new_tiered_record(code, name, specialty, low, high, note))
        added += 1

    with open(path, "w") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
        f.write("\n")

    print(f"Updated {updated} existing Cardiology records.")
    print(f"Added {added} new records.")
    print(f"Total records now: {len(data)}")


if __name__ == "__main__":
    main()
