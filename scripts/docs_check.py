#!/usr/bin/env python3
"""
docs_check.py — docs coherence sentinel (lean).

Checks that the spec docs are internally consistent ("aligned"):
  ERRORS (fail the gate):
    - each active FR contract block's `fr_file` matches its real path
    - each FR cross_links.adr_refs[] points to a file that exists
    - each FR cross_links.sisters[] points to an FR file that exists
    - the ADR index (architecture/adr/README.md) links resolve, and every
      NNN-*.md ADR file is listed in the index (no missing / no dangling)
  WARNINGS (informational):
    - active FRs not listed in docs/README.md "Current FRs"
    - ADR files referenced by no FR (orphan decisions)

Run: python3 scripts/docs_check.py    (exit non-zero on any ERROR)
"""
import re
import sys
from pathlib import Path

import yaml

REPO = Path(__file__).resolve().parent.parent
FR_ROOT = REPO / "docs/fr/survey"
ADR_DIR = REPO / "docs/architecture/adr"
DOCS_README = REPO / "docs/README.md"

CONTRACT_RE = re.compile(
    r"^##\s+(?:\d+\.\s+)?Contract\s+\(machine-readable\)[^\n]*\n"
    r"(?:[^\n]*\n)*?```yaml\n(.*?)\n```",
    re.IGNORECASE | re.MULTILINE | re.DOTALL,
)

errors: list[str] = []
warns: list[str] = []


def err(m: str):
    errors.append(m)


def warn(m: str):
    warns.append(m)


def active_frs() -> list[Path]:
    return [f for f in sorted((FR_ROOT / "active").glob("*.md")) if f.name != "README.md"]


def check_frs():
    for fr in active_frs():
        rel = fr.relative_to(REPO).as_posix()
        m = CONTRACT_RE.search(fr.read_text())
        if not m:
            err(f"{rel}: no contract block")
            continue
        try:
            data = yaml.safe_load(m.group(1)) or {}
        except yaml.YAMLError as e:
            err(f"{rel}: YAML error: {e}")
            continue

        if data.get("fr_file") != rel:
            err(f"{rel}: contract fr_file='{data.get('fr_file')}' != actual path '{rel}'")

        links = data.get("cross_links", {}) or {}
        for adr in links.get("adr_refs", []) or []:
            if not (REPO / adr).is_file():
                err(f"{rel}: adr_ref points to missing file '{adr}'")
        for sis in links.get("sisters", []) or []:
            if not (REPO / sis).is_file():
                err(f"{rel}: sister points to missing file '{sis}'")


def check_adr_index():
    if not ADR_DIR.is_dir():
        return
    adr_files = {p.name for p in ADR_DIR.glob("[0-9]*.md")}
    if not DOCS_README and not (ADR_DIR / "README.md").is_file():
        return
    index = (ADR_DIR / "README.md")
    if not index.is_file():
        err("docs/architecture/adr/README.md (ADR index) is missing")
        return
    linked = set(re.findall(r"\(([0-9]{3}-[^)]+\.md)\)", index.read_text()))
    for f in sorted(adr_files - linked):
        err(f"ADR {f} exists but is not listed in the ADR index README")
    for f in sorted(linked - adr_files):
        err(f"ADR index links to {f}, which does not exist")


def check_readme_fr_list():
    if not DOCS_README.is_file():
        return
    text = DOCS_README.read_text()
    for fr in active_frs():
        if fr.name not in text:
            warn(f"active FR {fr.name} not listed in docs/README.md")


def check_orphan_adrs():
    if not ADR_DIR.is_dir():
        return
    referenced = set()
    for fr in active_frs():
        m = CONTRACT_RE.search(fr.read_text())
        if not m:
            continue
        data = yaml.safe_load(m.group(1)) or {}
        for adr in (data.get("cross_links", {}) or {}).get("adr_refs", []) or []:
            referenced.add(Path(adr).name)
    for p in sorted(ADR_DIR.glob("[0-9]*.md")):
        if p.name not in referenced:
            warn(f"ADR {p.name} is referenced by no active FR")


def main() -> int:
    check_frs()
    check_adr_index()
    check_readme_fr_list()
    check_orphan_adrs()

    if errors:
        print("✗ docs coherence — ERRORS:")
        for e in errors:
            print(f"    {e}")
    else:
        print("✓ docs coherence — cross-links + ADR index aligned")

    if warns:
        print("\nℹ warnings:")
        for w in warns:
            print(f"    {w}")

    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())
