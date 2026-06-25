#!/usr/bin/env python3
"""
spec_drift.py — advisory FR↔code drift check (lean).

Two cheap, high-value checks against the contract blocks of active FRs:
  1. endpoints[]  — every declared API endpoint must be registered in
     internal/router/routes.go (resolving Gin group prefixes).
  2. db.writes[]  — every declared table must exist in a migration.

Exit code is ALWAYS 0 (advisory, not a gate) — mirrors txhcs `make drift`.

Run: python3 scripts/spec_drift.py
"""
import re
import sys
from pathlib import Path

import yaml

REPO = Path(__file__).resolve().parent.parent
ROUTES = REPO / "internal/router/routes.go"
MIGR = REPO / "internal/database/migrations"
FR_ACTIVE = REPO / "docs/fr/survey/active"

CONTRACT_RE = re.compile(
    r"^##\s+(?:\d+\.\s+)?Contract\s+\(machine-readable\)[^\n]*\n"
    r"(?:[^\n]*\n)*?```yaml\n(.*?)\n```",
    re.IGNORECASE | re.MULTILINE | re.DOTALL,
)
GROUP_RE = re.compile(r'(\w+)\s*:?=\s*(\w+)\.Group\("([^"]*)"\)')
ROUTE_RE = re.compile(r'(\w+)\.(GET|POST|PUT|PATCH|DELETE)\(\s*"([^"]*)"')


def norm(p: str) -> str:
    p = re.sub(r"\{[^}]+\}", "*", p)   # {id} -> *
    p = re.sub(r":[^/]+", "*", p)       # :id  -> *
    return p.rstrip("/") if len(p) > 1 else p


def registered_routes() -> set[tuple[str, str]]:
    src = ROUTES.read_text()
    prefix = {"r": ""}  # the *gin.Engine param in routes.go is named r
    groups = GROUP_RE.findall(src)
    for _ in range(len(groups) + 1):  # resolve nested groups
        for child, parent, arg in groups:
            if parent in prefix and child not in prefix:
                prefix[child] = prefix[parent] + arg
    routes = set()
    for var, method, frag in ROUTE_RE.findall(src):
        routes.add((method, norm(prefix.get(var, "") + frag)))
    return routes


def declared() -> tuple[list, list]:
    endpoints, tables = [], []
    for fr in sorted(FR_ACTIVE.glob("**/*.md")):
        m = CONTRACT_RE.search(fr.read_text())
        if not m:
            continue
        data = yaml.safe_load(m.group(1)) or {}
        for ep in data.get("endpoints", []) or []:
            endpoints.append((fr.name, ep["method"], ep["path"]))
        for w in (data.get("db", {}) or {}).get("writes", []) or []:
            tables.append((fr.name, w["table"]))
    return endpoints, tables


def migration_tables() -> set[str]:
    out = set()
    for sql in MIGR.glob("*.up.sql"):
        out |= set(re.findall(r"CREATE TABLE\s+(?:IF NOT EXISTS\s+)?(\w+)", sql.read_text(), re.I))
    return out


def main() -> int:
    routes = registered_routes()
    endpoints, tables = declared()
    mtables = migration_tables()

    missing_ep = [(f, m, p) for f, m, p in endpoints if (m, norm(p)) not in routes]
    missing_tbl = [(f, t) for f, t in tables if t not in mtables]

    print(f"Routes registered: {len(routes)} | FR endpoints: {len(endpoints)} | FR tables: {len(tables)}\n")

    if missing_ep:
        print("⚠ endpoints declared in an FR but NOT found in routes.go:")
        for f, m, p in missing_ep:
            print(f"    {m:6} {p}   ({f})")
    else:
        print("✓ every FR endpoint resolves to a registered route")

    if missing_tbl:
        print("\n⚠ tables declared in an FR but NOT found in any migration:")
        for f, t in missing_tbl:
            print(f"    {t}   ({f})")
    else:
        print("✓ every FR table exists in a migration")

    covered = {(m, norm(p)) for _, m, p in endpoints}
    orphans = sorted(r for r in routes if r not in covered and "/api/v1/auth/" not in r[1])
    if orphans:
        print(f"\nℹ {len(orphans)} registered route(s) not referenced by any active FR (spec backlog):")
        for m, p in orphans:
            print(f"    {m:6} {p}")

    print("\n(advisory — exit 0)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
