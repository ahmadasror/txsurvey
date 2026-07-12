#!/usr/bin/env python3
"""Coverage ratchet — fail if a tracked package's statement coverage regresses
below its recorded floor. Baselines = today's state (no retroactive debt); the
gate only catches *new* regressions. Raise a floor with `--update`.

Only DB-free unit-test packages are tracked, so this runs in the fast `make check`
without Postgres. The integration tier (tests/) is covered by `make test` in CI.

    python3 scripts/cover_check.py            # check (exit 1 on regression)
    python3 scripts/cover_check.py --update   # re-bless floors to current
"""
import json
import os
import re
import subprocess
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
BASELINE = os.path.join(ROOT, "scripts", "coverage-baseline.json")
# Float noise guard: a package must not drop more than this below its floor.
EPSILON = 0.1

COV_RE = re.compile(r"^ok\s+(\S+)\s+.*coverage:\s+([0-9.]+)% of statements", re.M)
NOCOV_RE = re.compile(r"^ok\s+(\S+)\s+.*coverage:\s+\[no statements\]", re.M)


def measure(packages):
    """Run `go test -short -cover` and return {package: percent}."""
    cmd = ["go", "test", "-short", "-cover"] + packages
    proc = subprocess.run(cmd, cwd=ROOT, capture_output=True, text=True)
    out = proc.stdout + proc.stderr
    if proc.returncode != 0:
        sys.stderr.write(out)
        sys.exit("cover_check: `go test` failed — fix tests before ratcheting")
    cov = {pkg: float(pct) for pkg, pct in COV_RE.findall(out)}
    for pkg in NOCOV_RE.findall(out):
        cov[pkg] = 0.0
    return cov


def main():
    update = "--update" in sys.argv[1:]
    with open(BASELINE) as f:
        base = json.load(f)
    packages = sorted(base["packages"].keys())
    current = measure(packages)

    if update:
        base["packages"] = {p: current.get(p, 0.0) for p in packages}
        with open(BASELINE, "w") as f:
            json.dump(base, f, indent=2, sort_keys=True)
            f.write("\n")
        print("cover-update: baseline re-blessed to current coverage")
        for p in packages:
            print(f"  {current.get(p, 0.0):5.1f}%  {p}")
        return 0

    failures = []
    print("Coverage ratchet (tracked unit-test packages):")
    for p in packages:
        floor = base["packages"][p]
        cur = current.get(p)
        if cur is None:
            failures.append(f"  MISSING  {p} — no coverage reported (package renamed/removed?)")
            continue
        mark = "ok " if cur + EPSILON >= floor else "RED"
        print(f"  [{mark}] {cur:5.1f}%  (floor {floor:5.1f}%)  {p}")
        if cur + EPSILON < floor:
            failures.append(f"  {p}: {cur:.1f}% < floor {floor:.1f}% — add a test or `make cover-update`")

    if failures:
        print("\ncover-check: FAIL — coverage regressed:")
        print("\n".join(failures))
        return 1
    print("\ncover-check: PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
