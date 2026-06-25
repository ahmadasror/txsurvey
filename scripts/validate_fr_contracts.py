#!/usr/bin/env python3
"""
validate_fr_contracts.py — extract the `## Contract (machine-readable)` YAML block
from each FR and validate it against docs/fr/_contract-schema.json.

Run: python3 scripts/validate_fr_contracts.py [--module survey]
Exit 0 on all-pass; non-zero on any schema violation (meant to be a hard gate).

Adapted (lean) from txhcs scripts/validate-fr-contract-blocks.py.
"""
import argparse
import json
import re
import sys
from pathlib import Path

import yaml
from jsonschema import Draft202012Validator

REPO_ROOT = Path(__file__).resolve().parent.parent
IN_SCOPE = ["survey"]
SKIP = ("README.md", "index.md")

CONTRACT_RE = re.compile(
    r"^##\s+(?:\d+\.\s+)?Contract\s+\(machine-readable\)[^\n]*\n"
    r"(?:[^\n]*\n)*?```yaml\n(.*?)\n```",
    re.IGNORECASE | re.MULTILINE | re.DOTALL,
)


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--module", action="append", help="restrict to module(s)")
    args = ap.parse_args()

    schema = json.loads((REPO_ROOT / "docs/fr/_contract-schema.json").read_text())
    validator = Draft202012Validator(schema)

    total = valid = invalid = no_block = 0
    for module in (args.module or IN_SCOPE):
        mod_dir = REPO_ROOT / "docs/fr" / module
        if not mod_dir.is_dir():
            continue
        print(f"[{module}]")
        for fr in sorted(mod_dir.glob("**/*.md")):
            if fr.name in SKIP:
                continue
            total += 1
            rel = fr.relative_to(REPO_ROOT).as_posix()
            m = CONTRACT_RE.search(fr.read_text())
            if not m:
                print(f"  - {rel} — no contract block")
                no_block += 1
                continue
            try:
                data = yaml.safe_load(m.group(1))
            except yaml.YAMLError as e:
                print(f"  ✗ {rel} — YAML error: {e}")
                invalid += 1
                continue
            errs = sorted(validator.iter_errors(data), key=lambda e: list(e.absolute_path))
            if not errs:
                print(f"  ✓ {rel}")
                valid += 1
            else:
                print(f"  ✗ {rel} — {len(errs)} schema error(s):")
                for e in errs[:6]:
                    loc = ".".join(str(p) for p in e.absolute_path) or "<root>"
                    print(f"      {loc}: {e.message}")
                invalid += 1

    print(f"\nSummary: total={total} valid={valid} invalid={invalid} no_block={no_block}")
    return 0 if invalid == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
