#!/usr/bin/env python3
"""Rebuild the standalone skill's reference data from the app's dataset.

Run from the repository root:

    python skills/tools/build_data.py

The app keeps one combined JSON file. The skill deliberately splits it:

* ``hbp-packages.json`` contains only records whose code and rate were
  checked against a published government source.
* ``hbp-packages-unverified.json`` contains the remaining records with
  every rupee figure *and* the internal placeholder identifier removed.
  An agent can still recognise the procedure name, but it cannot quote a
  number or a code that was never verified.

That split is a safety property, not a formatting preference. Keep it.
"""

from __future__ import annotations

import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
SRC = ROOT / "backend" / "internal" / "hbp" / "data"
DEST = ROOT / "skills" / "pmjay-denial-advocate" / "references"

VERIFIED_FIELDS = (
    "package_code",
    "package_name",
    "specialty",
    "indicative_rate_inr",
    "rate_type",
    "rate_max_inr",
    "per_diem_rates",
    "requires_preauth",
    "common_description_keywords",
    "confidence_notes",
    "source_note",
    "known_exclusion_category",
)

UNVERIFIED_FIELDS = (
    "package_name",
    "specialty",
    "requires_preauth",
    "common_description_keywords",
    "confidence_notes",
    "source_note",
)


def main() -> None:
    packages = json.loads((SRC / "hbp_packages.json").read_text(encoding="utf-8"))
    exclusions = json.loads((SRC / "exclusions.json").read_text(encoding="utf-8"))

    verified = [
        {key: p[key] for key in VERIFIED_FIELDS if key in p}
        for p in packages
        if p.get("verified") is True
    ]
    unverified = [
        {key: p[key] for key in UNVERIFIED_FIELDS if key in p}
        for p in packages
        if p.get("verified") is not True
    ]

    # A verified record must carry the real government code. The app's
    # loader enforces this at startup, and repeating it here means the
    # skill cannot be rebuilt from a dataset that predates that check.
    placeholders = [
        p["package_code"]
        for p in packages
        if p.get("verified") is True
        and str(p.get("package_code", "")).startswith("SEED-")
    ]
    if placeholders:
        raise SystemExit(
            "verified record(s) still carry a placeholder code: "
            f"{placeholders}; fix the dataset before building the skill"
        )

    # An unverified record's identifier is a placeholder, not a citation,
    # so it is deliberately not copied into the bundle (see
    # UNVERIFIED_FIELDS). This assertion catches a future edit that adds
    # it back.
    leaked_codes = [p.get("package_code") for p in unverified if "package_code" in p]
    if leaked_codes:
        raise SystemExit(
            f"unverified records leaked placeholder codes: {leaked_codes}"
        )

    DEST.mkdir(parents=True, exist_ok=True)

    (DEST / "hbp-packages.json").write_text(
        json.dumps(verified, ensure_ascii=False, indent=1) + "\n",
        encoding="utf-8",
    )
    (DEST / "hbp-packages-unverified.json").write_text(
        json.dumps(unverified, ensure_ascii=False, indent=1) + "\n",
        encoding="utf-8",
    )
    (DEST / "exclusions.json").write_text(
        json.dumps(exclusions, ensure_ascii=False, indent=1) + "\n",
        encoding="utf-8",
    )

    print(f"verified packages:   {len(verified)}")
    print(f"unverified packages: {len(unverified)}")
    print(f"exclusions:          {len(exclusions)}")
    print(f"written to:          {DEST}")


if __name__ == "__main__":
    main()
