#!/usr/bin/env python3
"""Build and validate the standalone PMJAY denial-advocate skill bundle.

Run from anywhere:

    python skills/package.py            # validate, then build the zip
    python skills/package.py --check    # validate only, no zip

Output:

    dist/pmjay-denial-advocate-<version>.zip

The zip contains a single top-level folder, which is the layout agent
marketplaces expect: unzip it into a skills directory and the folder
``pmjay-denial-advocate/`` appears with ``SKILL.md`` inside.

The validation is deliberately paranoid. A public skill that leaks a
local path, an API key, or a stale frontend file is worse than a skill
that never ships. Every check here maps to a real failure mode listed in
the error message, so a failure tells you what to fix rather than just
that something is wrong.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import re
import sys
import zipfile
from pathlib import Path

HERE = Path(__file__).resolve().parent
REPO = HERE.parent
SKILL_DIR = HERE / "pmjay-denial-advocate"
DIST = REPO / "dist"

NAME = "pmjay-denial-advocate"
VERSION = "1.0.0"

REQUIRED_FILES = (
    "SKILL.md",
    "references/triage-rules.md",
    "references/escalation-and-rights.md",
    "references/response-templates.md",
    "references/sources.md",
    "references/hbp-packages.json",
    "references/hbp-packages-unverified.json",
    "references/exclusions.json",
    "scripts/lookup.py",
)

ALLOWED_SUFFIXES = {".md", ".json", ".py"}

# Anything here means the bundle picked up application code, build output, or
# editor scratch. The skill is text plus two small scripts on purpose.
FORBIDDEN_NAMES = {
    "node_modules",
    ".next",
    "__pycache__",
    ".pytest_cache",
    ".git",
    ".DS_Store",
    "package.json",
    "package-lock.json",
    "tsconfig.json",
}

FORBIDDEN_SUFFIXES = {
    ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs", ".css", ".scss",
    ".map", ".zip", ".exe", ".dll", ".so", ".dylib", ".pyc", ".log",
    ".env", ".pem", ".key", ".png", ".jpg", ".jpeg", ".gif", ".svg",
    ".woff", ".woff2", ".ttf", ".otf",
}

SECRET_PATTERNS = (
    (re.compile(r"sk-[A-Za-z0-9_\-]{20,}"), "OpenAI-style API key"),
    (re.compile(r"sk-ant-[A-Za-z0-9_\-]{20,}"), "Anthropic API key"),
    (re.compile(r"gh[pousr]_[A-Za-z0-9]{20,}"), "GitHub token"),
    (re.compile(r"AKIA[0-9A-Z]{16}"), "AWS access key"),
    (re.compile(r"-----BEGIN [A-Z ]*PRIVATE KEY-----"), "private key"),
    (re.compile(r"(?i)\b(api[_-]?key|secret|password|passwd|token)\b\s*[:=]\s*[\"'][^\"'\s]{8,}[\"']"), "hardcoded credential"),
    (re.compile(r"xox[baprs]-[A-Za-z0-9-]{10,}"), "Slack token"),
)

ABSOLUTE_PATH_PATTERNS = (
    (re.compile(r"[A-Za-z]:\\Users\\[^\\\s\"']+"), "local Windows path"),
    (re.compile(r"/Users/[^/\s\"']+/"), "local macOS path"),
    (re.compile(r"/home/[^/\s\"']+/"), "local Linux path"),
)


class Failure(Exception):
    pass


def fail(message: str) -> None:
    raise Failure(message)


def read_frontmatter(text: str) -> dict[str, str]:
    if not text.startswith("---\n"):
        fail("SKILL.md must start with a YAML frontmatter block (`---`).")
    end = text.find("\n---\n", 4)
    if end == -1:
        fail("SKILL.md frontmatter is not closed with a `---` line.")
    block = text[4:end]
    fields: dict[str, str] = {}
    for line in block.splitlines():
        if not line.strip() or line.lstrip().startswith("#"):
            continue
        if ":" not in line:
            fail(f"SKILL.md frontmatter line is not `key: value`: {line!r}")
        key, _, value = line.partition(":")
        value = value.strip()
        if value.startswith(('"', "'")) and value.endswith(('"', "'")) and len(value) >= 2:
            value = value[1:-1]
        fields[key.strip()] = value
    return fields


def iter_files() -> list[Path]:
    if not SKILL_DIR.is_dir():
        fail(f"skill folder is missing: {SKILL_DIR}")
    files: list[Path] = []
    skipped = 0
    for path in sorted(SKILL_DIR.rglob("*")):
        if not path.is_file():
            continue
        # Python bytecode is local build noise, not part of the bundle. Skip it
        # rather than failing a packaging run because someone ran py_compile.
        if "__pycache__" in path.parts or path.suffix == ".pyc":
            skipped += 1
            continue
        files.append(path)
    if skipped:
        print(f"  skipped {skipped} local bytecode file(s)")
    return files


def check_layout(files: list[Path]) -> None:
    rel = {p.relative_to(SKILL_DIR).as_posix() for p in files}
    missing = [f for f in REQUIRED_FILES if f not in rel]
    if missing:
        fail("missing required file(s): " + ", ".join(missing))

    for path in files:
        relpath = path.relative_to(SKILL_DIR)
        if path.name in FORBIDDEN_NAMES:
            fail(f"forbidden file or folder in bundle: {relpath.as_posix()}")
        if path.suffix.lower() in FORBIDDEN_SUFFIXES:
            fail(f"forbidden file type in bundle: {relpath.as_posix()}")
        if path.suffix.lower() not in ALLOWED_SUFFIXES:
            fail(f"unexpected file type in bundle: {relpath.as_posix()}")
        if any(part in FORBIDDEN_NAMES for part in relpath.parts):
            fail(f"forbidden folder in bundle: {relpath.as_posix()}")
        if relpath.name.startswith("."):
            fail(f"hidden file in bundle: {relpath.as_posix()}")


def check_skill_md(files: list[Path]) -> str:
    text = (SKILL_DIR / "SKILL.md").read_text(encoding="utf-8")
    meta = read_frontmatter(text)

    name = meta.get("name", "")
    if name != NAME:
        fail(f"SKILL.md `name` must be {NAME!r}, found {name!r}")

    description = meta.get("description", "")
    if not 40 <= len(description) <= 1024:
        fail(
            "SKILL.md `description` must be 40-1024 characters "
            f"(found {len(description)}); marketplaces truncate or reject longer ones"
        )
    lowered = description.lower()
    if not lowered.startswith("use when"):
        fail("SKILL.md `description` should start with 'Use when' so an agent knows when to load it")

    if text.count("---\n") < 2:
        fail("SKILL.md frontmatter delimiters look unbalanced")

    # Structural guardrail: the care-first rule has to survive edits.
    if "Get treatment first" not in text:
        fail("SKILL.md lost the care-first rule; that line is non-negotiable")
    if "14555" not in text:
        fail("SKILL.md lost the PMJAY helpline number")
    return text


def check_json(files: list[Path]) -> None:
    for path in files:
        if path.suffix != ".json":
            continue
        try:
            data = json.loads(path.read_text(encoding="utf-8"))
        except json.JSONDecodeError as exc:
            fail(f"invalid JSON in {path.relative_to(SKILL_DIR).as_posix()}: {exc}")
        if not isinstance(data, list) or not data:
            fail(f"{path.name} must be a non-empty JSON array")

    packages = json.loads((SKILL_DIR / "references/hbp-packages.json").read_text(encoding="utf-8"))
    unverified = json.loads(
        (SKILL_DIR / "references/hbp-packages-unverified.json").read_text(encoding="utf-8")
    )

    rate_fields = ("indicative_rate_inr", "rate_max_inr", "per_diem_rates", "rate_type")
    for record in unverified:
        leaked = [field for field in rate_fields if field in record]
        if leaked:
            fail(
                "unverified record leaks rate field(s) "
                f"{leaked} for {record.get('package_name')!r}; "
                "strip rates in skills/tools/build_data.py and rebuild"
            )
        # A code is a citation. The app dataset keeps a SEED- identifier
        # for these rows so they are addressable, but the skill must not
        # hand an agent an identifier that exists in no government
        # schedule — the agent would repeat it to a family.
        if "package_code" in record:
            fail(
                "unverified record "
                f"{record.get('package_name')!r} carries a package_code "
                f"({record['package_code']!r}); unverified records are "
                "name-only, so drop the key in skills/tools/build_data.py"
            )

    codes = [r.get("package_code") for r in packages if r.get("package_code")]
    duplicates = sorted({c for c in codes if codes.count(c) > 1})
    if duplicates:
        fail(f"duplicate package_code(s) across reference files: {duplicates}")

    # A verified record is a citation. A placeholder identifier under a
    # "verified" badge is worse than an unverified one, because it looks
    # checked. This caught two real records (MC011A, MG072A) that had kept
    # their provisional SEED- codes after their rates were verified.
    placeholder = [
        r.get("package_code")
        for r in packages
        if str(r.get("package_code", "")).startswith("SEED-")
    ]
    if placeholder:
        fail(
            "verified record(s) still carry a placeholder code: "
            f"{placeholder}; use the real HBP code before packaging"
        )

    missing_codes = [r.get("package_name") for r in packages if not r.get("package_code")]
    if missing_codes:
        fail(f"verified record(s) without a package_code: {missing_codes[:5]}")

    if not any(r.get("package_code") == "UNSPECIFIED" for r in packages):
        fail("UNSPECIFIED discretionary record is missing; triage rules depend on it")

    print(f"  reference data: {len(packages)} verified, {len(unverified)} unverified")


def check_text(files: list[Path]) -> None:
    for path in files:
        text = path.read_text(encoding="utf-8", errors="replace")
        rel = path.relative_to(SKILL_DIR).as_posix()

        for pattern, label in SECRET_PATTERNS:
            match = pattern.search(text)
            if match:
                fail(f"possible {label} in {rel}: {match.group(0)[:24]}...")

        for pattern, label in ABSOLUTE_PATH_PATTERNS:
            match = pattern.search(text)
            if match:
                fail(f"{label} in {rel}: {match.group(0)}")

        if "http://localhost" in text or "127.0.0.1" in text:
            fail(f"local development URL in {rel}")

        if "\t" in text and path.suffix == ".py":
            fail(f"tab character in {rel}; Python files must use spaces")

        if text and not text.endswith("\n"):
            fail(f"{rel} does not end with a newline")


def check_encoding(files: list[Path]) -> None:
    for path in files:
        try:
            raw = path.read_bytes()
            raw.decode("utf-8")
        except UnicodeDecodeError as exc:
            fail(f"{path.relative_to(SKILL_DIR).as_posix()} is not valid UTF-8: {exc}")
        if raw.startswith(b"\xef\xbb\xbf"):
            fail(f"{path.relative_to(SKILL_DIR).as_posix()} has a UTF-8 BOM; remove it")


def build(files: list[Path], version: str) -> Path:
    DIST.mkdir(exist_ok=True)
    archive = DIST / f"{NAME}-{version}.zip"
    if archive.exists():
        archive.unlink()

    with zipfile.ZipFile(archive, "w", compression=zipfile.ZIP_DEFLATED, compresslevel=9) as zf:
        for path in files:
            rel = path.relative_to(SKILL_DIR).as_posix()
            info = zipfile.ZipInfo(f"{NAME}/{rel}")
            info.date_time = (1980, 1, 1, 0, 0, 0)
            info.external_attr = 0o644 << 16
            info.compress_type = zipfile.ZIP_DEFLATED
            zf.writestr(info, path.read_bytes())
    return archive


def verify_archive(archive: Path, files: list[Path]) -> None:
    with zipfile.ZipFile(archive) as zf:
        names = zf.namelist()
        expected = [f"{NAME}/{p.relative_to(SKILL_DIR).as_posix()}" for p in files]
        if names != expected:
            fail("archive contents do not match the validated file list")
        if zf.testzip() is not None:
            fail(f"corrupt entry in {archive.name}")
        if f"{NAME}/SKILL.md" not in names:
            fail("archive is missing SKILL.md at its expected location")
        if any(n.endswith("/") for n in names):
            fail("archive contains directory entries; only file entries are expected")


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true", help="validate without building the zip")
    parser.add_argument("--version", default=VERSION, help=f"bundle version (default {VERSION})")
    args = parser.parse_args()

    if not re.fullmatch(r"\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?", args.version):
        print(f"error: version must look like 1.0.0, got {args.version!r}", file=sys.stderr)
        return 2

    print(f"validating {NAME} {args.version}")
    try:
        files = iter_files()
        check_layout(files)
        print(f"  layout: {len(files)} files, no frontend/build artifacts")
        check_skill_md(files)
        print("  SKILL.md: frontmatter, description length, care-first rule OK")
        check_json(files)
        check_encoding(files)
        check_text(files)
        print("  text: no secrets, absolute paths, BOMs, or localhost URLs")
    except Failure as exc:
        print(f"\nFAILED: {exc}", file=sys.stderr)
        return 1

    if args.check:
        print("\ncheck passed; no archive written")
        return 0

    archive = build(files, args.version)
    verify_archive(archive, files)
    digest = hashlib.sha256(archive.read_bytes()).hexdigest()
    size_kb = archive.stat().st_size / 1024
    print("\nbuilt " + str(archive))
    print(f"  size:   {size_kb:.1f} KB")
    print(f"  sha256: {digest}")
    print(f"  layout: {NAME}/ at the archive root")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
