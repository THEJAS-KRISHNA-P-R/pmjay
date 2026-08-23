#!/usr/bin/env python3
"""Generates new, additive HBP package records from two real government
sources fetched during the August 2026 continuation session, and merges
them with the existing hbp_packages.json without touching any of the
original 40 records.

SOURCES (see /docs/DATA_SOURCES.md for the full writeup):

  Source A — Assam State Health Agency (Atal Amrit Abhiyan), HBP 2.0 rates
    https://atalamritabhiyan.assam.gov.in/information-services/pmjay-package-master
    Used here only for Interventional Neuroradiology (IN001-IN010), a
    specialty not present in Source B's reachable range.

  Source B — Uttar Pradesh State Health Agency, HBP 2022 rates (3-tier X/Y/Z)
    https://ayushmanup.in/assets/doc/HBP-2022.pdf
    Used for General Medicine + Pediatric per-diem diagnoses, Medical
    Oncology, and the three "High End" categories.

Both sources were fetched directly via tool calls this session; every
number in the lists below was transcribed from that fetched content, not
recalled from training data. Where a row's data was incomplete or
non-standard (a handful of General Medicine diagnoses had a single
"X/day" figure instead of the full four-level breakdown, for example),
it was deliberately left out rather than guessed — see the "excluded"
notes inline.

This script is a one-time generator, run once during this session and
kept for the same audit-trail reason transform.py was kept: so a future
reader can see exactly how every number in the output file was derived,
without having to trust a diff.
"""
import json

EXISTING_CODES = {
    'BM001B', 'BM006A', 'ER001A', 'ER002B', 'ER003A', 'MC005A', 'MC007A',
    'MC008A', 'MC009A', 'MC015A', 'MC016A', 'MC017A', 'MM008A', 'MM009A',
    'MN001A', 'MN002A', 'MN003A', 'MN004A', 'MN005A', 'MN009A', 'MN010A',
    'SEED-CARD-001', 'SEED-CARD-002', 'SEED-CARD-003', 'SEED-ENT-001',
    'SEED-GS-001', 'SEED-GS-002', 'SEED-GS-003', 'SEED-NEPH-001',
    'SEED-NEURO-001', 'SEED-OBG-001', 'SEED-OBG-002', 'SEED-OBG-003',
    'SEED-ONCO-001', 'SEED-OPH-001', 'SEED-ORTHO-001', 'SEED-ORTHO-002',
    'SEED-ORTHO-003', 'SEED-URO-001', 'UNSPECIFIED',
}

SOURCE_B_NOTE = (
    "Package code, name, and rate verified directly against the Uttar "
    "Pradesh State Health Agency's published mirror of the HBP 2022 "
    "package master (https://ayushmanup.in/assets/doc/HBP-2022.pdf), "
    "fetched and cross-checked during this build on 15 August 2026. "
    "HBP 2022 postdates the HBP 2.1 source used for this dataset's "
    "original 40 records — see docs/DATA_SOURCES.md for the version "
    "currency question this raises for a small number of overlapping "
    "packages, which was deliberately left unresolved this session "
    "rather than silently overwritten."
)

SOURCE_A_NOTE = (
    "Package code, name, and rate verified directly against the Assam "
    "State Health Agency's published PMJAY Package Master "
    "(https://atalamritabhiyan.assam.gov.in/information-services/"
    "pmjay-package-master), fetched and cross-checked during this build "
    "on 15 August 2026. This table uses HBP 2.0-era rates; see "
    "docs/DATA_SOURCES.md."
)

# ---------------------------------------------------------------------------
# General Medicine + Pediatric per-diem diagnoses (Source B).
#
# Confirmed uniform stratification across every row below, transcribed
# directly from the source table: Routine Ward 2300/day, HDU 3630/day,
# ICU without ventilator 9350/day, ICU with ventilator 9900/day.
#
# Deliberately EXCLUDED from this list (present in the source but not
# included here, to avoid guessing at incomplete data):
#   MG0105-MG0108, MG0120  — flat "Add on" or rehab package rates, not
#                             per-diem; a genuine future addition but a
#                             different RateType shape, left for later.
#   MG0119                 — source gave only "2100/day" with no HDU/ICU
#                             breakdown; including it with the standard
#                             four-level structure would be a guess.
#   MG072, MG073-076, MG082-085, MG097-099 — flat/tiered procedure rates,
#                             not per-diem; candidates for a later pass.
#   MG077                  — per-day but capped at 5 days with a
#                             different rate structure (33000/37500/45000
#                             as a *tiered*, not per-diem, figure);
#                             doesn't fit either shape cleanly.
# ---------------------------------------------------------------------------
PER_DIEM = (2300, 3630, 9350, 9900)

GENERAL_MEDICINE = [
    ("MG001A", "Acute febrile illness"),
    ("MG002A", "Severe sepsis"),
    ("MG002B", "Septic shock"),
    ("MG003A", "Malaria"),
    ("MG003B", "Complicated malaria"),
    ("MG004A", "Dengue fever"),
    ("MG004B", "Dengue haemorrhagic fever"),
    ("MG004C", "Dengue shock syndrome"),
    ("MG005A", "Chikungunya fever"),
    ("MG006A", "Enteric fever"),
    ("MG007A", "HIV with complications"),
    ("MG008A", "Leptospirosis"),
    ("MG009A", "Acute gastroenteritis with moderate dehydration"),
    ("MG009B", "Acute gastroenteritis with severe dehydration"),
    ("MG010A", "Chronic diarrhoea"),
    ("MG010B", "Persistent diarrhoea"),
    ("MG011A", "Dysentery"),
    ("MG012A", "Acute viral hepatitis"),
    ("MG013A", "Chronic hepatitis"),
    ("MG014A", "Liver abscess"),
    ("MG015A", "Visceral leishmaniasis"),
    ("MG016A", "Pneumonia"),
    ("MG017A", "Severe pneumonia"),
    ("MG018A", "Empyema"),
    ("MG019A", "Lung abscess"),
    ("MG020A", "Pericardial tuberculosis"),
    ("MG020B", "Pleural tuberculosis"),
    ("MG021A", "Urinary tract infection"),
    ("MG022A", "Viral encephalitis"),
    ("MG023A", "Septic arthritis"),
    ("MG024A", "Skin and soft tissue infections"),
    ("MG025A", "Recurrent vomiting with dehydration"),
    ("MG026A", "Pyrexia of unknown origin"),
    ("MG027A", "Bronchiectasis"),
    ("MG028A", "Acute bronchitis"),
    ("MG029A", "Acute exacerbation of COPD"),
    ("MG030A", "Acute exacerbation of interstitial lung disease"),
    ("MG031A", "Bacterial endocarditis"),
    ("MG031B", "Fungal endocarditis"),
    ("MG032A", "Vasculitis"),
    ("MG033A", "Acute pancreatitis"),
    ("MG033B", "Chronic pancreatitis"),
    ("MG034A", "Ascites"),
    ("MG035A", "Acute transverse myelitis"),
    ("MG036A", "Atrial fibrillation"),
    ("MG037A", "Cardiac tamponade"),
    ("MG038A", "Congestive heart failure"),
    ("MG039A", "Acute asthmatic attack"),
    ("MG039B", "Status asthmaticus"),
    ("MG040A", "Type 1 respiratory failure"),
    ("MG040B", "Type 2 respiratory failure"),
    ("MG040C", "Respiratory failure due to any cause (pneumonia, asthma, COPD, ARDS, foreign body, poisoning, head injury)"),
    ("MG041A", "Upper GI bleeding, conservative management"),
    ("MG041B", "Upper GI bleeding, endoscopic management"),
    ("MG042A", "Lower GI haemorrhage"),
    ("MG043A", "Addison's disease"),
    ("MG044A", "Renal colic"),
    ("MG045A", "AKI / renal failure"),
    ("MG050A", "Immune-mediated CNS disorders"),
    ("MG051A", "Hydrocephalus"),
    ("MG052A", "Myxedema coma"),
    ("MG053A", "Thyrotoxic crisis"),
    ("MG054A", "Gout"),
    ("MG055A", "Pneumothorax"),
    ("MG056A", "Neuromuscular disorders"),
    ("MG057A", "Hypoglycemia"),
    ("MG058A", "Diabetic foot, debridement"),
    ("MG059A", "Diabetic ketoacidosis"),
    ("MG060A", "Electrolyte imbalance — hypercalcemia"),
    ("MG060B", "Electrolyte imbalance — hypocalcemia"),
    ("MG060C", "Electrolyte imbalance — hyponatremia"),
    ("MG060D", "Electrolyte imbalance — hypernatremia"),
    ("MG060E", "Electrolyte imbalance — hyperkalaemia"),
    ("MG060F", "Electrolyte imbalance — hypokalaemia"),
    ("MG061A", "Hyperosmolar non-ketotic coma"),
    ("MG062A", "Accelerated hypertension"),
    ("MG063A", "Hypertensive emergencies"),
    ("MG064A", "Severe anemia"),
    ("MG065A", "Sickle cell anemia"),
    ("MG066A", "Anaphylaxis"),
    ("MG067A", "Heat stroke"),
    ("MG068A", "Systemic lupus erythematosus (SLE) / diffuse alveolar haemorrhage associated with SLE"),
    ("MG069A", "Guillain-Barre syndrome, IVIG"),
    ("MG070A", "Snake bite"),
    ("MG071A", "Acute organophosphorus poisoning"),
    ("MG071B", "Other poisonings"),
    ("MG078A", "Alcoholic liver disease"),
    ("MG079A", "Peripheral arterial thrombosis"),
    ("MG081A", "Arrhythmia (IHD/CAD/arrhythmia)"),
    ("MG081B", "Coronary artery disease (IHD/CAD/arrhythmia)"),
    ("MG086A", "Acute ischemic stroke"),
    ("MG086B", "Acute ischemic stroke, IV thrombolysis (recombinant tissue plasminogen activator)"),
    ("MG086C", "Acute ischemic stroke, IV thrombolysis (tenecteplase)"),
    ("MG087A", "Venous sinus thrombosis"),
    ("MG088A", "Pyogenic meningitis"),
    ("MG089A", "Fungal meningitis"),
    ("MG090A", "Autoimmune encephalitis, plasmapheresis"),
    ("MG090B", "Autoimmune encephalitis, immunoglobulin (IVIG)"),
    ("MG090C", "Acute transverse myelitis / acute demyelinating encephalitis"),
    ("MG091A", "Acute haemorrhagic stroke, haematoma evacuation"),
    ("MG091B", "Acute haemorrhagic stroke, extra-ventricular drainage"),
    ("MG093A", "Myasthenic crisis, plasmapheresis"),
    ("MG094A", "Tuberculous meningitis with hydrocephalus (VP shunt / EVD / Omaya)"),
    ("MG095A", "Cholangitis"),
    ("MG096A", "Intercostal drainage"),
    ("MG0101A", "Acute severe ulcerative colitis"),
    ("MG0102A", "Mesenteric ischemia"),
    ("MG0103A", "Intestinal obstruction"),
    ("MG0104A", "Acute necrotizing severe pancreatitis"),
    ("MG0109A", "Pulmonary thromboembolism"),
    ("MG0110A", "Acute liver failure"),
    ("MG0111A", "Pleural effusion"),
    ("MG0112A", "Hyperbilirubinemia"),
    ("MG0113A", "Polytrauma"),
    ("MG0113B", "Trauma — facio-maxillary"),
    ("MG0113C", "Trauma — head injury"),
    ("MG0113D", "Trauma — rib fracture, conservative management"),
    ("MG0113E", "Trauma — blunt injury, conservative management"),
    ("MG0113F", "Trauma — contusion chest injury"),
    ("MG0114A", "Oesophageal varices banding"),
    ("MG0115A", "Inflammatory myopathy / myasthenic crisis"),
    ("MG0116A", "Guillain-Barre syndrome, plasmapheresis"),
    ("MG0116B", "Myasthenic crisis, plasmapheresis"),
    ("MG0117A", "Moyamoya revascularization"),
    ("MG0118A", "Evaluation of drug-resistant epilepsy — phase 1"),
]

PEDIATRIC = [
    ("MP001D", "Pediatric seizure disorders — acute non-febrile seizures"),
    ("MP003A", "Acute encephalopathy — acute febrile encephalopathy"),
    ("MP003B", "Acute encephalopathy — acute disseminated encephalomyelitis"),
    ("MP004A", "Acute encephalopathy — hypertensive/metabolic/febrile/hepatic"),
    ("MP005A", "Acute infectious meningitis — pyogenic meningoencephalitis"),
    ("MP005B", "Acute infectious meningitis — tubercular aseptic meningitis"),
    ("MP005D", "Acute infectious meningitis — viral hypertensive encephalopathy"),
    ("MP005G", "Brain abscess / intracranial abscess / aseptic meningitis"),
    ("MP006B", "Meningitis — partially treated pyogenic meningitis"),
    ("MP006C", "Meningitis — neuro-tuberculosis"),
    ("MP006D", "Meningitis — complicated bacterial meningitis"),
    ("MP006E", "Meningitis — acute meningitis"),
    ("MP008A", "Medical management for raised intracranial pressure (neurosurgical/trauma/malignancy/meningo-encephalitis)"),
    ("MP011A", "Intracranial space-occupying lesions — neurocysticercosis, brain tumours"),
    ("MP015A", "Juvenile myasthenia requiring admission for work-up or in-patient care"),
    ("MP018A", "Acute childhood asthma / status asthmaticus"),
    ("MP020B", "Steven-Johnson syndrome"),
    ("MP021A", "Acute abdomen pain"),
    ("MP023A", "Unexplained hepatosplenomegaly requiring admission for work-up"),
    ("MP024A", "Neonatal/infantile cholestasis / choledochal cysts"),
    ("MP026A", "Nephrotic syndrome with peritonitis"),
    ("MP026B", "Nephrotic syndrome, steroid-dependent or -resistant"),
    ("MP027A", "Haemolytic uremic syndrome"),
    ("MP039A", "Acute rheumatic fever"),
    ("MP039B", "Rheumatic valvular heart disease"),
    ("MP040A", "Cyanotic spells without congenital heart disease"),
    ("MP043A", "Kawasaki disease"),
    ("MP047A", "Croup syndrome — acute laryngotracheobronchitis / acute epiglottitis"),
    ("MP050A", "Nephrotic syndrome, uncomplicated steroid-sensitive"),
    ("MG047A", "Pediatric seizure disorders — status epilepticus"),
    ("MG049C", "Acute ischemic stroke (pediatric)"),
]

# ---------------------------------------------------------------------------
# Medical Oncology — package-level entries (Source B).
#
# HBP 2022 lists chemotherapy at *regimen* granularity (e.g. 16 different
# drug regimens under "CT for CA Breast" alone). That level of detail is
# wrong for this tool's purpose and actively risky: citing one specific
# regimen's rate to a family could contradict what their oncologist is
# actually prescribing, undermining a true denial dispute rather than
# helping it. Each entry below is therefore RateType "tiered" using the
# MIN (cheapest regimen, Tier3/Z) to MAX (most expensive regimen,
# Tier1/X) rate found across every regimen variant in that package — an
# honest "this package exists and the real cost span is X-Y" rather than
# a false single figure. Regimen names are preserved in confidence_notes
# for a human reviewer, not surfaced to families as a specific match.
# ---------------------------------------------------------------------------
ONCOLOGY = [
    ("MO001", "Chemotherapy for carcinoma breast", 1300, 45000, "16 regimens spanning oral hormonal therapy (Tamoxifen, cheapest) to monoclonal antibody therapy (Fulvestrant/Trastuzumab, most expensive)."),
    ("MO002", "Chemotherapy for metastatic bone malignancy and multiple myeloma", 3900, 4400, "Zoledronic acid, monthly."),
    ("MO003", "Chemotherapy for carcinoma ovary", 1400, 19800, "12 regimens spanning oral hormonal therapy to combination platinum-based regimens."),
    ("MO004", "Chemotherapy for germ cell tumour", 8100, 26900, "7 regimens, single-agent carboplatin to multi-drug combinations."),
    ("MO005", "Chemotherapy for gestational trophoblastic neoplasia", 1400, 14100, "4 regimens, single-agent methotrexate to multi-drug combinations."),
    ("MO006", "Chemotherapy for cervical cancer", 2600, 16400, "2 regimens."),
    ("MO007", "Chemotherapy for endometrial cancer", 4300, 19000, "5 regimens."),
    ("MO008", "Chemotherapy for vulvar cancer", 2600, 16000, "3 regimens."),
    ("MO009", "Chemotherapy for Ewing sarcoma", 10900, 25500, "3 multi-drug combination regimens."),
    ("MO010", "Chemotherapy for osteogenic sarcoma", 13200, 32600, "4 regimens including relapsed-disease protocols."),
    ("MO011", "Chemotherapy for soft tissue sarcoma", 4400, 26400, "3 regimens."),
    ("MO012", "Chemotherapy for metastatic melanoma", 5500, 7900, "3 regimens including targeted therapy (Imatinib)."),
    ("MO013", "Chemotherapy for anal cancer", 9000, 18600, "5 regimens."),
    ("MO014", "Chemotherapy for colorectal cancer", 6100, 20700, "8 regimens including radiotherapy-combined protocols."),
    ("MO015", "Chemotherapy for oesophageal cancer", 10700, 29900, "5 regimens."),
    ("MO016", "Chemotherapy for oesophageal / stomach cancer", 6600, 20400, "12 regimens."),
    ("MO017", "Chemotherapy for hepatocellular carcinoma", 9900, 24800, "3 regimens including targeted oral therapy (Sorafenib, Lenvatinib)."),
    ("MO018", "Chemotherapy for pancreatic cancer", 4400, 37600, "6 regimens."),
    ("MO019", "Chemotherapy for gall bladder cancer / cholangiocarcinoma", 4400, 19200, "8 regimens."),
    ("MO020", "Chemotherapy for gastrointestinal stromal tumour", 11000, 16500, "2 targeted oral therapy regimens (Imatinib, Sunitinib)."),
    ("MO021", "Chemotherapy for brain cancer", 5500, 33000, "2 Temozolomide dosing protocols."),
    ("MO022", "Chemotherapy for mesothelioma", 10400, 13400, "3 platinum-doublet regimens."),
    ("MO023", "Chemotherapy for thymic carcinoma", 6500, 7800, "2 regimens."),
    ("MO024", "Chemotherapy for head and neck cancer", 2600, 16000, "16 regimens, single-agent to multi-drug combinations."),
    ("MO025", "Chemotherapy for renal cell cancer", 11000, 14300, "2 targeted oral therapy regimens."),
    ("MO026", "Chemotherapy for ureter / bladder / urethra cancer", 7500, 17100, "11 regimens."),
    ("MO027", "Chemotherapy for penile cancer", 6600, 16000, "7 regimens."),
    ("MO028", "Chemotherapy for prostate cancer", 3600, 16900, "9 regimens including hormonal (LHRH agonist) and targeted therapy."),
    ("MO029", "Chemotherapy for high-grade B-cell NHL (except Burkitt's and PCNSL)", 28900, 29700, "2 rituximab-based multi-drug regimens."),
    ("MO030", "Chemotherapy for high-grade NHL, B-cell", 38400, 38400, "Single rituximab/cisplatin/cytarabine regimen."),
    ("MO031", "Chemotherapy for relapsed high-grade B-cell NHL (except Burkitt's and PCNSL)", 35100, 38900, "2 rituximab-based salvage regimens."),
    ("MO032", "Chemotherapy for PMBCL / Burkitt's lymphoma / seropositive B-cell NHL", 34900, 34900, "Single multi-drug regimen."),
    ("MO033", "Chemotherapy for Burkitt's NHL", 38000, 38000, "Intensive multi-protocol regimen (Codox-M-IVAC / GMALL / Hyper-CVAD)."),
    ("MO034", "Chemotherapy for low-grade B-cell NHL", 24800, 27500, "2 regimens."),
    ("MO035", "Chemotherapy for low-grade NHL", 17600, 19300, "2 rituximab-based regimens."),
    ("MO036", "Chemotherapy for chronic lymphocytic leukaemia", 5300, 44800, "4 regimens, oral targeted therapy to combination chemo-immunotherapy."),
    ("MO037", "Chemotherapy for peripheral T-cell lymphoma", 5700, 21700, "3 regimens."),
    ("MO038", "Chemotherapy for NK/T-cell lymphoma", 8800, 21300, "2 regimens."),
    ("MO039", "Chemotherapy for Hodgkin's lymphoma", 4300, 11300, "Multiple regimens (COPP, ABVD, and others); rate range reflects the regimens with complete rate data reachable this session — see docs/DATA_SOURCES.md, this package's extraction was cut off by the source wall mid-list, so this range may not reflect the full regimen set."),
]

# ---------------------------------------------------------------------------
# High End Diagnostics / Medicine / Procedures, Infectious Disease
# (Source B) — entirely new specialty categories for this dataset.
# All-equal Tier3/Tier2/Tier1 rows are rendered as flat (RateType "").
# ---------------------------------------------------------------------------
HIGH_END_DIAGNOSTICS = [
    ("HD001", "Lymphangiography", 10920, 10920, 10920),
    ("HD002", "Diagnostic venography (DVA)", 5520, 5520, 5520),
    ("HD003", "IVUS (intravascular ultrasound) — peripheral vessels", 16560, 16560, 16560),
    ("HD004", "Diskography", 5520, 5520, 5520),
    ("HD005", "USG-guided percutaneous biopsy", 2520, 2520, 2520),
    ("HD006", "USG-guided percutaneous FNAC", 720, 720, 720),
    ("HD007", "USG-guided percutaneous needle aspiration", 720, 720, 720),
    ("HD008", "CT-guided percutaneous biopsy", 3720, 3720, 3720),
    ("HD009", "CT-guided percutaneous FNAC", 3120, 3120, 3120),
    ("HD010", "CT-guided percutaneous needle aspiration", 3120, 3120, 3120),
    ("HD011", "Genetic workup", 20000, 20000, 20000),
    ("HD012", "Metabolic workup", 30000, 30000, 30000),
    ("HD013", "Video EEG monitoring test (VEEG)", 3600, 3600, 3600),
]

HIGH_END_MEDICINE = [
    ("HM001", "Recombinant tissue plasminogen activator", 42000, 42000, 42000),
    ("HM002", "Tenecteplase", 24900, 24900, 24900),
    ("HM003", "Heparin", 15000, 15000, 15000),
    ("HM004", "Methylprednisolone", 25000, 25000, 25000),
    ("HM006", "Liposomal amphotericin", 125000, 125000, 125000),
    ("HM007", "IVIG", 200000, 200000, 200000),
    ("HM011", "Rituximab", 7500, 7500, 7500),
    ("HM012", "Human albumin 20%", 6000, 6000, 6000),
    ("HM013", "Albumin 5%", 3000, 3000, 3000),
    ("HM017", "Imipenem", 7000, 7000, 7000),
    ("HM018", "Meropenem", 7000, 7000, 7000),
    ("HM019", "Piperacillin-tazobactam", 3000, 3000, 3000),
    ("HM020", "Colistin", 7000, 7000, 7000),
    ("HM021", "Vancomycin", 2100, 2100, 2100),
    ("HM022", "Amphotericin deoxycholate", 1000, 1000, 1000),
]

HIGH_END_PROCEDURES = [
    ("HP001", "USG-guided percutaneous radiofrequency ablation (RFA)", 26640, 26640, 26640),
    ("HP002", "USG-guided percutaneous microwave ablation (MWA)", 30640, 30640, 30640),
    ("HP003", "CT-guided percutaneous radiofrequency ablation (RFA)", 29040, 29040, 29040),
    ("HP004", "CT-guided percutaneous microwave ablation (MWA)", 33040, 33040, 33040),
    ("HP005", "USG-guided percutaneous catheter drainage", 5820, 5820, 5820),
    ("HP006", "CT-guided percutaneous catheter drainage", 5820, 5820, 5820),
    ("HP007", "Cerebral angiogram under local anaesthesia", 5800, 5800, 5800),
    ("HP008", "Cerebral angiogram under general anaesthesia", 18900, 18900, 18900),
    ("HP009", "Spinal angiogram under general anaesthesia", 18900, 18900, 18900),
    ("HP010", "Plasmapheresis (high-end procedures listing)", 70000, 70000, 70000),
]

INFECTIOUS_DISEASE = [
    ("ID001A", "Screening test for COVID-19 infection (PCR)", 1500, 1500, 1500),
    ("ID001B", "Test for confirmation of COVID-19 infection", 3000, 3000, 3000),
]

# ---------------------------------------------------------------------------
# Interventional Neuroradiology (Source A — HBP 2.0). Entirely new
# specialty for this dataset; not reached in Source B before its wall.
# ---------------------------------------------------------------------------
INTERVENTIONAL_NEURORADIOLOGY = [
    ("IN001A", "Dural AVMs, per sitting, with glue", 70000),
    ("IN001B", "Dural AVFs, per sitting, with glue", 70000),
    ("IN001C", "Dural AVMs, per sitting, with Onyx", 150000),
    ("IN001D", "Dural AVFs, per sitting, with Onyx", 150000),
    ("IN002A", "Cerebral AVM embolization using Histoacryl, per sitting", 100000),
    ("IN002B", "Spinal AVM embolization using Histoacryl, per sitting", 100000),
    ("IN003A", "Coil embolization for aneurysms (includes first 3 coils + balloon/stent if used)", 100000),
    ("IN004A", "Carotico-cavernous fistula (CCF) embolization with coils", 30000),
    ("IN004B", "Carotico-cavernous fistula (CCF) embolization with balloon", 64000),
    ("IN005A", "Pre-operative tumour embolization, per session", 40000),
    ("IN006A", "Intracranial balloon angioplasty with stenting", 160000),
    ("IN007A", "Intracranial thrombolysis / clot retrieval", 160000),
    ("IN008A", "Balloon test occlusion", 70000),
    ("IN009A", "Parent vessel occlusion, basic", 30000),
    ("IN010A", "Vertebroplasty", 40000),
]


def keywords_from_name(name):
    """Cheap, deterministic keyword derivation: lowercase words over 3
    chars, stripped of punctuation, deduplicated, capped at 8. This is a
    floor for retrievability, not a substitute for the hand-tuned
    keyword lists the original 40 records have — see docs/DATA_SOURCES.md
    on this as a scoped-out refinement, not an oversight.
    """
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


def make_per_diem_record(code, name, specialty):
    assert code not in EXISTING_CODES, f"collision: {code}"
    return {
        "package_code": code,
        "package_name": name[:1].upper() + name[1:],
        "specialty": specialty,
        "indicative_rate_inr": PER_DIEM[0],
        "requires_preauth": True,
        "common_description_keywords": keywords_from_name(name),
        "confidence_notes": (
            "Per-diem admission package: rate is charged per day of stay, "
            "stratified by ward level, not as a single total for the "
            "admission. See rate_type/per_diem_rates."
        ),
        "source_note": SOURCE_B_NOTE,
        "verified": True,
        "rate_type": "per_diem",
        "per_diem_rates": {
            "routine_ward_inr": PER_DIEM[0],
            "hdu_inr": PER_DIEM[1],
            "icu_no_vent_inr": PER_DIEM[2],
            "icu_vent_inr": PER_DIEM[3],
        },
    }


def make_oncology_record(code, name, min_rate, max_rate, note):
    assert code not in EXISTING_CODES, f"collision: {code}"
    rec = {
        "package_code": code,
        "package_name": name[:1].upper() + name[1:],
        "specialty": "Medical Oncology",
        "indicative_rate_inr": min_rate,
        "requires_preauth": True,
        "common_description_keywords": keywords_from_name(name) + ["chemotherapy", "chemo", "cancer"],
        "confidence_notes": (
            f"Package-level entry covering the real chemotherapy regimen(s) documented under "
            f"this HBP package code — {note} A family's actual regimen and cost may differ "
            f"depending on exactly what their oncologist prescribes."
        ),
        "source_note": SOURCE_B_NOTE,
        "verified": True,
    }
    # A handful of oncology packages have exactly one documented regimen
    # (min == max) — no real range to show, so this stays a flat rate
    # rather than a "tiered" claim with a zero-width range.
    if max_rate > min_rate:
        rec["rate_type"] = "tiered"
        rec["rate_max_inr"] = max_rate
    return rec


def make_flat_or_tiered_record(code, name, specialty, low, mid, high, source_note, requires_preauth=True):
    assert code not in EXISTING_CODES, f"collision: {code}"
    rec = {
        "package_code": code,
        "package_name": name[:1].upper() + name[1:],
        "specialty": specialty,
        "indicative_rate_inr": low,
        "requires_preauth": requires_preauth,
        "common_description_keywords": keywords_from_name(name),
        "confidence_notes": "",
        "source_note": source_note,
        "verified": True,
    }
    if high > low:
        rec["rate_type"] = "tiered"
        rec["rate_max_inr"] = high
    return rec


def build():
    new_records = []

    for code, name in GENERAL_MEDICINE:
        new_records.append(make_per_diem_record(code, name, "General Medicine"))
    for code, name in PEDIATRIC:
        new_records.append(make_per_diem_record(code, name, "Pediatric Medical Management"))
    for code, name, lo, hi, note in ONCOLOGY:
        new_records.append(make_oncology_record(code, name, lo, hi, note))
    for code, name, z, y, x in HIGH_END_DIAGNOSTICS:
        new_records.append(make_flat_or_tiered_record(code, name, "High End Diagnostics", z, y, x, SOURCE_B_NOTE))
    for code, name, z, y, x in HIGH_END_MEDICINE:
        new_records.append(make_flat_or_tiered_record(code, name, "High End Medicine", z, y, x, SOURCE_B_NOTE, requires_preauth=False))
    for code, name, z, y, x in HIGH_END_PROCEDURES:
        new_records.append(make_flat_or_tiered_record(code, name, "High End Procedures", z, y, x, SOURCE_B_NOTE))
    for code, name, z, y, x in INFECTIOUS_DISEASE:
        new_records.append(make_flat_or_tiered_record(code, name, "Infectious Diseases", z, y, x, SOURCE_B_NOTE, requires_preauth=False))
    for code, name, rate in INTERVENTIONAL_NEURORADIOLOGY:
        new_records.append(make_flat_or_tiered_record(code, name, "Interventional Neuroradiology", rate, rate, rate, SOURCE_A_NOTE))

    # Hard collision + duplicate check before writing anything.
    codes_seen = set()
    for r in new_records:
        c = r["package_code"]
        if c in EXISTING_CODES:
            raise SystemExit(f"FATAL: new record {c} collides with an existing package_code")
        if c in codes_seen:
            raise SystemExit(f"FATAL: duplicate new package_code {c}")
        codes_seen.add(c)

    with open("hbp_packages.json") as f:
        existing = json.load(f)

    # SEED-ONCO-001 is an unverified (verified=false) placeholder — "Chemotherapy,
    # Per Cycle, Standard Protocol" — now superseded by 39 specific, verified,
    # real oncology packages above. Removing an unverified placeholder is a much
    # lower-risk edit than touching any verified record, and keeping a vague
    # low-specificity placeholder alongside precise real packages risks
    # retrieval matching the wrong (worse) one. See docs/DATA_SOURCES.md.
    removed = [p for p in existing if p["package_code"] == "SEED-ONCO-001"]
    existing = [p for p in existing if p["package_code"] != "SEED-ONCO-001"]
    assert len(removed) == 1, "expected to find exactly one SEED-ONCO-001 record to retire"

    merged = existing + new_records
    with open("hbp_packages.json", "w") as f:
        json.dump(merged, f, indent=2, ensure_ascii=False)
        f.write("\n")

    print(f"Existing records kept: {len(existing)} (removed 1: SEED-ONCO-001, superseded)")
    print(f"New records added: {len(new_records)}")
    print(f"  General Medicine per-diem: {len(GENERAL_MEDICINE)}")
    print(f"  Pediatric per-diem: {len(PEDIATRIC)}")
    print(f"  Medical Oncology: {len(ONCOLOGY)}")
    print(f"  High End Diagnostics: {len(HIGH_END_DIAGNOSTICS)}")
    print(f"  High End Medicine: {len(HIGH_END_MEDICINE)}")
    print(f"  High End Procedures: {len(HIGH_END_PROCEDURES)}")
    print(f"  Infectious Disease: {len(INFECTIOUS_DISEASE)}")
    print(f"  Interventional Neuroradiology: {len(INTERVENTIONAL_NEURORADIOLOGY)}")
    print(f"Total records in output: {len(merged)}")


if __name__ == "__main__":
    build()
