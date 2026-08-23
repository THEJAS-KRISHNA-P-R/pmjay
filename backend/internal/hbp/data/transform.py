import json

PATH = "hbp_packages.json"
REAL_SOURCE = ("Package code, name, specialty group, and total package price verified "
               "directly against the National Health Authority's own published HBP 2.1 "
               "master package list (https://nha.gov.in/img/pmjay-files/HBP-2.1.pdf), "
               "fetched and cross-checked during this build on 14 August 2026. This is "
               "the same official master file an empanelled hospital's own billing "
               "system references. See docs/DATA_SOURCES.md for what this does and does "
               "not establish about the rest of the dataset.")

data = json.load(open(PATH))

# ---------------------------------------------------------------------
# Correct two existing placeholders in place. Package codes are left
# untouched deliberately -- several tests reference "SEED-CARD-001" by
# exact string as a lookup key, and PackageCode is never shown to a
# family (only PackageName/Specialty reach response text -- confirmed
# by grepping internal/response), so updating the DATA behind a stable
# identifier is safe and is the correct fix, not a rename.
# ---------------------------------------------------------------------
by_code = {p["package_code"]: p for p in data}

card = by_code["SEED-CARD-001"]
card["package_name"] = ("Percutaneous Transluminal Coronary Angioplasty (PTCA) with "
                         "Single Bare-Metal Stent, Inclusive of Diagnostic Angiogram")
card["indicative_rate_inr"] = 49300  # real MC011A: procedure 40,600 + 1 bare-metal stent 8,700
card["confidence_notes"] = (card["confidence_notes"] + " This figure is the baseline "
    "single-bare-metal-stent tier (real code MC011A). The same real package scales with "
    "stent count and type: 2 bare-metal stents = 58,000; 3 = 66,700; drug-eluting "
    "variants run 72,200 / 103,800 / 135,400 for 1/2/3 stents respectively. A family "
    "describing a multi-stent or drug-eluting procedure is still a match to this same "
    "package; only the rate differs.")
card["source_note"] = REAL_SOURCE + " Real government package code for this exact procedure: MC011A."
card["verified"] = True

neph = by_code["SEED-NEPH-001"]
neph["specialty"] = "General Medicine"  # correction: real HBP groups haemodialysis under
                                          # General Medicine (MG072), not a separate
                                          # Nephrology empanelment category -- this matters
                                          # because Specialty is what the family-facing
                                          # sentence checks the hospital's empanelment against.
neph["confidence_notes"] = (neph["confidence_notes"] + " Specialty corrected from "
    "'Nephrology' to 'General Medicine' during real-data verification: the actual HBP "
    "master list groups haemodialysis under General Medicine (code MG072), not a "
    "separate Nephrology empanelment category. This matters because the family-facing "
    "message checks the hospital's empanelment against this exact specialty name.")
neph["source_note"] = REAL_SOURCE + " Real government package code for this exact procedure: MG072A."
neph["verified"] = True

# ---------------------------------------------------------------------
# Remove the single vague "Level II NICU, per day" placeholder. Not
# referenced by exact code in any test (confirmed by grep), so this is
# safe to restructure rather than patch -- the real HBP master list
# defines five distinct, genuinely different-priced neonatal severity
# tiers, which is materially more useful than one guessed number for a
# family whose baby is at a specific, describable severity level.
# ---------------------------------------------------------------------
data = [p for p in data if p["package_code"] != "SEED-PAED-001"]

def new_entry(code, name, specialty, rate, preauth, keywords, notes):
    return {
        "package_code": code,
        "package_name": name,
        "specialty": specialty,
        "indicative_rate_inr": rate,
        "requires_preauth": preauth,
        "common_description_keywords": keywords,
        "confidence_notes": notes,
        "source_note": REAL_SOURCE,
        "verified": True,
    }

new_entries = [
    # --- Neonatal Care: replaces SEED-PAED-001 with the real 5-tier structure ---
    new_entry("MN001A", "Basic Neonatal Care Package (managed at mother's bedside, no SNCU/NICU admission)",
        "Neonatal Care", 500, True,
        ["newborn needs feeding support", "baby mild jaundice", "baby needs monitoring",
         "baby with mother", "newborn checkup"],
        "Lowest-acuity tier: babies managed beside the mother without SNCU/NICU admission. "
        "If the family describes actual NICU admission, this is very likely the wrong tier -- "
        "see MN002A-MN005A."),
    new_entry("MN002A", "Special Neonatal Care Package (SNCU/NICU admission, short-term monitoring)",
        "Neonatal Care", 3000, True,
        ["baby in nicu", "baby admitted to nicu", "premature baby", "baby breathing fast",
         "baby needs phototherapy", "baby jaundice severe", "newborn admitted"],
        "First tier that requires actual SNCU/NICU admission, for short-term/milder conditions "
        "(mild respiratory distress, severe jaundice needing intensive phototherapy, monitoring)."),
    new_entry("MN003A", "Intensive Neonatal Care Package",
        "Neonatal Care", 5000, True,
        ["baby on cpap", "baby needs breathing support", "baby low birth weight",
         "baby has infection", "baby premature under 1800 grams"],
        "For babies needing non-invasive respiratory support (CPAP) or short-duration "
        "mechanical ventilation (under 24 hours), or uncomplicated sepsis/pneumonia."),
    new_entry("MN004A", "Advanced Neonatal Care Package",
        "Neonatal Care", 6000, True,
        ["baby on ventilator", "baby very premature", "baby cooling therapy",
         "baby heart rhythm problem", "baby birth asphyxia severe"],
        "For babies needing invasive ventilation beyond 24 hours, therapeutic hypothermia, "
        "or very low birth weight (1200-1499g)."),
    new_entry("MN005A", "Critical Care Neonatal Package",
        "Neonatal Care", 7000, True,
        ["baby extremely critical", "baby multiple organ failure", "baby under 1200 grams",
         "baby high frequency ventilator", "baby fighting for life"],
        "Highest-acuity tier: extremely low birth weight (under 1200g) or multi-organ support "
        "including high-frequency ventilation. A family describing this severity should not be "
        "matched to a lower, cheaper tier."),
    new_entry("MN009A", "Advanced Surgery for Retinopathy of Prematurity",
        "Neonatal Care", 15000, True,
        ["baby eye surgery", "baby retinopathy", "premature baby eye problem", "baby ROP surgery"],
        "Distinct from the laser-therapy-only ROP package (session-based, not in this seed set); "
        "this is the surgical-intervention tier."),
    new_entry("MN010A", "Ventriculoperitoneal (VP) Shunt Surgery / External Drainage for Hydrocephalus",
        "Neonatal Care", 5000, True,
        ["baby water on brain", "baby hydrocephalus", "baby shunt surgery", "baby brain fluid"],
        "Neonatal-specific hydrocephalus shunt package; an older child or adult with the same "
        "condition would be matched under Neurosurgery instead, not this code."),

    # --- Burns Management ---
    new_entry("BM001B", "Thermal Burns (up to 40% Total Body Surface Area), including skin grafting/flap cover",
        "Burns Management", 40000, True,
        ["burns", "fire burns", "skin grafting", "burn surgery", "burnt in accident"],
        "Real package scales sharply by % TBSA: this is the up-to-40% tier (40,000). The "
        "40-60% tier is 50,000 and the over-60% tier is 80,000 -- confirm severity before "
        "citing a specific figure to a family."),
    new_entry("BM006A", "Post-Burn Contracture Surgery for Functional Improvement",
        "Burns Management", 50000, True,
        ["burn scar surgery", "contracture release", "burn healed but can't move",
         "skin tightness after burns"],
        "For functional-improvement surgery after burns have healed, not acute burn treatment "
        "itself -- a distinct package from BM001-BM005."),

    # --- Emergency Medicine ---
    new_entry("ER001A", "Laceration - Suturing / Dressing",
        "Emergency Medicine", 2000, False,
        ["stitches", "cut needing stitches", "wound suturing", "deep cut"],
        "Low-value, high-frequency package; a hospital demanding payment upfront for this "
        "specific service is a common, easily-verified denial pattern."),
    new_entry("ER002B", "Cardiopulmonary Emergency, Unstable, with Resuscitation",
        "Emergency Medicine", 10000, False,
        ["heart stopped", "not breathing", "emergency resuscitation", "CPR given",
         "collapsed", "cardiac arrest"],
        "The unstable/resuscitation tier (10,000); a stable-status cardiopulmonary emergency "
        "is a separate, lower-rate code (ER002A, 2,000) not included in this seed set."),
    new_entry("ER003A", "Animal Bites (Excluding Snake Bite)",
        "Emergency Medicine", 1700, False,
        ["dog bite", "animal bite", "rabies treatment", "bitten by dog"],
        "Explicitly excludes snake bite, which is handled under a separate General Medicine "
        "package (MG070) not included in this seed set."),

    # --- Cardiology (interventional, beyond the existing PTCA/CABG/angiogram entries) ---
    new_entry("MC005A", "Balloon Mitral Valvotomy",
        "Cardiology", 90700, True,
        ["heart valve procedure", "mitral valve", "balloon valve procedure"],
        "Includes balloon and accessories in the total package price -- a hospital itemising "
        "the balloon separately as an extra charge is worth flagging."),
    new_entry("MC007A", "ASD (Atrial Septal Defect) Device Closure",
        "Cardiology", 98900, True,
        ["hole in heart", "ASD closure", "atrial septal defect", "heart hole surgery"],
        "Congenital heart defect closure; device cost is bundled into the total package price."),
    new_entry("MC008A", "VSD (Ventricular Septal Defect) Device Closure",
        "Cardiology", 109900, True,
        ["hole in heart", "VSD closure", "ventricular septal defect"],
        "Congenital heart defect closure, device bundled into total package price."),
    new_entry("MC009A", "PDA (Patent Ductus Arteriosus) Device Closure",
        "Cardiology", 55000, True,
        ["PDA closure", "patent ductus arteriosus", "baby heart defect closure"],
        "Common in infants/children; device cost bundled into total package price."),
    new_entry("MC015A", "Single Chamber Permanent Pacemaker Implantation",
        "Cardiology, CTVS", 69500, True,
        ["pacemaker", "single chamber pacemaker", "heart rhythm device"],
        "Rate-responsive pacemaker device is bundled into the total package price."),
    new_entry("MC016A", "Double Chamber Permanent Pacemaker Implantation",
        "Cardiology, CTVS", 108000, True,
        ["pacemaker", "dual chamber pacemaker", "double chamber pacemaker"],
        "Rate-responsive pacemaker device is bundled into the total package price."),
    new_entry("MC017A", "Peripheral Angioplasty",
        "Cardiology", 55500, True,
        ["leg artery blockage", "peripheral angioplasty", "leg stent", "blocked leg artery"],
        "Distinct from coronary (heart) angioplasty/PTCA -- this is for peripheral (typically "
        "limb) arteries. Bare-metal peripheral stent bundled into total package price."),

    # --- Mental Health ---
    new_entry("MM008A", "Pre-ECT / Pre-TMS Investigation Package",
        "Mental Disorders", 10000, True,
        ["tests before ECT", "pre-ECT workup", "tests before shock therapy"],
        "Bundle of required pre-procedure investigations (bloodwork, ECG, imaging) -- billed "
        "and matched separately from the ECT/TMS session itself (MM009A/MM010A)."),
    new_entry("MM009A", "Electro Convulsive Therapy (ECT), Per Session",
        "Mental Disorders", 3000, True,
        ["ECT", "shock therapy", "electroconvulsive therapy"],
        "Per-session rate -- a course typically involves multiple sessions, each separately "
        "billable under this same code."),
]

data.extend(new_entries)

json.dump(data, open(PATH, "w"), indent=2, ensure_ascii=False)
print(f"Done. Total packages now: {len(data)}")
print(f"Verified now: {sum(1 for p in data if p['verified'])}")
print(f"Still placeholder: {sum(1 for p in data if not p['verified'])}")
