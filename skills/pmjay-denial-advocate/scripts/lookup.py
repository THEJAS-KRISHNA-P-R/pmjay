#!/usr/bin/env python3
"""Search the bundled PMJAY HBP reference without any dependencies.

This mirrors the recall-biased keyword scorer used by the PMJAY Advocate
application. It is a shortlist generator, not a decision engine: read the
person's actual situation, score the results yourself, and apply the rules in
``references/triage-rules.md``.

Examples
--------
    python scripts/lookup.py "dialysis kidney failure"
    python scripts/lookup.py "cataract operation" --verified-only
    python scripts/lookup.py "ivf fertility" --exclusions
    python scripts/lookup.py "gallbladder stones" --json

Exit codes
----------
0   at least one candidate scored above zero
1   no candidate scored above zero (not the same as "not covered")
2   bad input or missing reference data
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
REF = HERE.parent / "references"

# Windows consoles default to a legacy code page that cannot encode ₹.
# Agents and pipes generally expect UTF-8; do not crash on a rupee sign.
for stream in (sys.stdout, sys.stderr):
    if hasattr(stream, "reconfigure"):
        stream.reconfigure(encoding="utf-8", errors="replace")

STOPWORDS = {
    "the", "a", "an", "is", "are", "was", "for", "to", "of", "and", "or",
    "in", "on", "with", "at", "it", "he", "she", "his", "her", "my", "our",
    "we", "us", "has", "have", "had", "this", "that",
}


def tokenize(text: str) -> list[str]:
    """Lowercase, split on non-ASCII-alphanumerics, drop stopwords."""
    text = text.lower()
    out: list[str] = []
    current: list[str] = []
    for ch in text:
        if ("a" <= ch <= "z") or ("0" <= ch <= "9"):
            current.append(ch)
        elif current:
            word = "".join(current)
            if word not in STOPWORDS:
                out.append(word)
            current = []
    if current:
        word = "".join(current)
        if word not in STOPWORDS:
            out.append(word)
    return out


def token_match(token: str, other: str) -> bool:
    """Exact match, or a meaningful shared stem (haemodialysis ~ dialysis)."""
    if token == other:
        return True
    shorter, longer = sorted((token, other), key=len)
    return len(shorter) >= 5 and shorter in longer


def phrase_present(keyword_tokens: list[str], query_tokens: set[str]) -> bool:
    return bool(keyword_tokens) and all(
        any(token_match(keyword_token, query_token) for query_token in query_tokens)
        for keyword_token in keyword_tokens
    )


def score_record(record: dict, query_tokens: set[str]) -> int:
    score = 0
    for keyword in record.get("common_description_keywords", []):
        keyword_tokens = tokenize(keyword)
        if phrase_present(keyword_tokens, query_tokens):
            score += 3 * len(keyword_tokens)
    for token in tokenize(record.get("package_name", "")):
        if any(token_match(query_token, token) for query_token in query_tokens):
            score += 2
    for token in tokenize(record.get("specialty", "")):
        if any(token_match(query_token, token) for query_token in query_tokens):
            score += 1
    return score


def score_exclusion(record: dict, query_tokens: set[str]) -> int:
    score = 0
    for keyword in record.get("keywords", []):
        keyword_tokens = tokenize(keyword)
        if phrase_present(keyword_tokens, query_tokens):
            score += 3 * len(keyword_tokens)
    for token in tokenize(record.get("display_name", "")):
        if any(token_match(query_token, token) for query_token in query_tokens):
            score += 2
    return score


def load(path: Path) -> list[dict]:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except FileNotFoundError:
        sys.stderr.write(
            f"missing reference data: {path}\n"
            "Run `python skills/tools/build_data.py` from the repository root.\n"
        )
        raise SystemExit(2)
    except json.JSONDecodeError as exc:
        sys.stderr.write(f"invalid reference data: {path}: {exc}\n")
        raise SystemExit(2)


def format_rate(pkg: dict) -> str:
    rate_type = pkg.get("rate_type") or "flat"
    if rate_type == "per_diem":
        per_diem = pkg.get("per_diem_rates") or {}
        return (
            "per day: routine ward ₹{routine_ward_inr}, HDU ₹{hdu_inr}, "
            "ICU (no vent) ₹{icu_no_vent_inr}, ICU (vent) ₹{icu_vent_inr}"
        ).format(
            routine_ward_inr=per_diem.get("routine_ward_inr", "?"),
            hdu_inr=per_diem.get("hdu_inr", "?"),
            icu_no_vent_inr=per_diem.get("icu_no_vent_inr", "?"),
            icu_vent_inr=per_diem.get("icu_vent_inr", "?"),
        )
    if rate_type == "tiered":
        low = pkg.get("indicative_rate_inr")
        high = pkg.get("rate_max_inr")
        if high and high > (low or 0):
            return f"₹{low:,}–₹{high:,} depending on city tier"
    rate = pkg.get("indicative_rate_inr")
    return f"₹{rate:,}" if isinstance(rate, int) else "rate unavailable"


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("query", help="what the family said, or English clinical terms")
    parser.add_argument(
        "--verified-only",
        action="store_true",
        help="exclude the 15 unverified name-only records",
    )
    parser.add_argument(
        "--exclusions",
        action="store_true",
        help="also score the four confirmed exclusion categories",
    )
    parser.add_argument("--limit", type=int, default=10, help="max results (default 10, max 50)")
    parser.add_argument("--json", action="store_true", help="print machine-readable JSON")
    args = parser.parse_args()

    if not args.query.strip():
        parser.error("query must not be empty")
    args.limit = max(1, min(args.limit, 50))

    verified = load(REF / "hbp-packages.json")
    unverified = [] if args.verified_only else load(REF / "hbp-packages-unverified.json")

    query_tokens = set(tokenize(args.query))
    if not query_tokens:
        sys.stderr.write(
            "No searchable Latin-alphabet terms in the query. Translate the procedure "
            "into English clinical terms first, then run the lookup.\n"
        )
        return 2

    ranked: list[dict] = []
    for pkg in verified:
        score = score_record(pkg, query_tokens)
        if score > 0:
            ranked.append({**pkg, "match_score": score, "status": "verified"})
    for pkg in unverified:
        score = score_record(pkg, query_tokens)
        if score > 0:
            ranked.append({**pkg, "match_score": score, "status": "unverified"})
    ranked.sort(key=lambda p: p["match_score"], reverse=True)
    ranked = ranked[: args.limit]

    exclusion_results: list[dict] = []
    if args.exclusions:
        exclusions = load(REF / "exclusions.json")
        exclusion_results = sorted(
            ({**e, "match_score": score_exclusion(e, query_tokens)} for e in exclusions),
            key=lambda e: e["match_score"],
            reverse=True,
        )

    if args.json:
        print(json.dumps(
            {
                "query": args.query,
                "query_tokens": sorted(query_tokens),
                "candidates": ranked,
                "exclusions": exclusion_results,
                "note": (
                    "Shortlist only. Apply references/triage-rules.md. "
                    "Never quote a rate from a record with status 'unverified', "
                    "and never supply an HBP code for one: unverified records "
                    "deliberately carry no package_code."
                ),
            },
            ensure_ascii=False,
            indent=2,
        ))
        return 0 if ranked else 1

    if not ranked and not any(e["match_score"] > 0 for e in exclusion_results):
        print("No candidate scored above zero.")
        print()
        print("This is NOT a denial and NOT evidence that nothing is covered.")
        print("Ask for the exact procedure name and HBP code, or call 14555.")
        print("Consider PMJAY's discretionary 'Unspecified Procedure' category,")
        print("but only when the procedure genuinely does not map to a named package.")
        return 1

    print(f'Query: "{args.query}"')
    print(f"Tokens: {', '.join(sorted(query_tokens))}")
    print()
    print(f"Top {len(ranked)} package candidate(s):")
    for index, pkg in enumerate(ranked, start=1):
        status = pkg["status"]
        print(f"\n{index}. [{status}] {pkg['package_name']}")
        if status == "verified":
            print(f"   code:       {pkg.get('package_code', 'unknown')}")
        else:
            print("   code:       no verified HBP code — do not supply one")
        print(f"   specialty:  {pkg['specialty']}")
        print(f"   score:      {pkg['match_score']}")
        if status == "verified":
            print(f"   rate:       {format_rate(pkg)}")
            if pkg.get("rate_type") == "per_diem":
                print("   note:       per-day rate, not a total admission cost")
            if pkg.get("rate_type") == "tiered":
                print("   note:       range depends on the hospital's city tier")
            if pkg["package_code"] == "UNSPECIFIED":
                print("   note:       discretionary cap, not a normal package rate")
            if pkg.get("requires_preauth"):
                print("   pre-auth:   usually required")
        else:
            print("   rate:       NOT AVAILABLE — never quote a figure for this record")
            print("   note:       name-only match; treat as Amber and ask the hospital/14555")
        if pkg.get("known_exclusion_category"):
            print(f"   exclusion:  {pkg['known_exclusion_category']}")

    if exclusion_results:
        print("\nExclusion categories (score independently, not instead):")
        for exclusion in exclusion_results:
            marker = "*" if exclusion["match_score"] > 0 else " "
            print(f" {marker} {exclusion['display_name']} ({exclusion['category']}) "
                  f"— score {exclusion['match_score']}")
            if exclusion.get("nuance"):
                print(f"     nuance: {exclusion['nuance']}")

    print("\nShortlist only. Apply references/triage-rules.md.")
    print("A rate is indicative, not a guarantee. Confirm with the hospital or 14555.")
    return 0 if ranked else 1


if __name__ == "__main__":
    raise SystemExit(main())
