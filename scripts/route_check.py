#!/usr/bin/env python3
"""route_check.py — ENFORCING route ↔ FR gate (the hard sibling of spec_drift.py).

Fails (exit 1) when either:
  1. an active FR declares an endpoint that is NOT registered in routes.go
     (the FR promises a route the code doesn't have), or
  2. a registered route is covered by NO active FR endpoint and is NOT waivered
     (a new endpoint shipped without a contract entry).

Auth handshake routes (/api/v1/auth/*) are exempt (browser OAuth, not a data API).
Legacy routes live in docs/fr/_route_waivers.txt so the baseline starts green;
add an FR entry for new routes rather than a waiver.

Run: python3 scripts/route_check.py     (make route-check)
"""
import sys
from pathlib import Path

# Reuse the extraction logic from the advisory checker (same dir).
sys.path.insert(0, str(Path(__file__).resolve().parent))
from spec_drift import registered_routes, declared, norm  # noqa: E402

REPO = Path(__file__).resolve().parent.parent
WAIVERS = REPO / "docs/fr/_route_waivers.txt"


def load_waivers() -> set[tuple[str, str]]:
    out: set[tuple[str, str]] = set()
    if not WAIVERS.exists():
        return out
    for line in WAIVERS.read_text().splitlines():
        line = line.split("#", 1)[0].strip()
        if not line:
            continue
        parts = line.split(None, 1)
        if len(parts) != 2:
            continue
        method, path = parts
        out.add((method.upper(), norm(path)))
    return out


def main() -> int:
    routes = registered_routes()
    endpoints, _ = declared()
    waivers = load_waivers()

    covered = {(m, norm(p)) for _, m, p in endpoints}
    missing_ep = [(f, m, p) for f, m, p in endpoints if (m, norm(p)) not in routes]
    orphans = sorted(
        r for r in routes
        if r not in covered and "/api/v1/auth/" not in r[1] and r not in waivers
    )
    # A waiver that no longer matches any route is dead — flag it so the list stays honest.
    stale_waivers = sorted(w for w in waivers if w not in routes)

    print(f"Routes: {len(routes)} | FR endpoints: {len(endpoints)} | waivers: {len(waivers)}\n")

    ok = True
    if missing_ep:
        ok = False
        print("✗ FR declares endpoints NOT registered in routes.go:")
        for f, m, p in missing_ep:
            print(f"    {m:6} {p}   ({f})")
    if orphans:
        ok = False
        print("✗ registered routes with NO FR endpoint and NO waiver:")
        for m, p in orphans:
            print(f"    {m:6} {p}   → add an endpoints[] entry to the owning FR, or waive it")
    if stale_waivers:
        ok = False
        print("✗ stale waivers (no matching route — remove from _route_waivers.txt):")
        for m, p in stale_waivers:
            print(f"    {m:6} {p}")

    if ok:
        print("route-check: PASS — every route is specced or waivered")
        return 0
    print("\nroute-check: FAIL")
    return 1


if __name__ == "__main__":
    sys.exit(main())
