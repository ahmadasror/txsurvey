#!/usr/bin/env python3
"""Coverage ratchet — a tracked package's statement coverage may only go UP.

Two tiers (scripts/coverage-baseline.json):
  - "packages"  : load-bearing floors (service layer, auth). STRICT — any drop
                  below the recorded floor fails the build. No epsilon: the whole
                  point of a ratchet is that it cannot quietly decay.
  - "advisory"  : thin layers (handlers are ~pass-through) where a coverage
                  percentage is noise. Reported for visibility and a soft warning
                  on regression, but it does NOT gate. This keeps the teeth on the
                  service layer instead of pretending a 2% handler floor is
                  meaningful.

Baselines = state on the day a floor is blessed (no retroactive debt); the gate
catches *new* regressions. Lock in a gain — or re-bless after moving code between
tiers — with `--update`.

Only DB-free unit-test packages are tracked, so this runs in the fast `make check`
without Postgres. The integration tier (tests/) is covered by `make test` in CI.

    python3 scripts/cover_check.py            # check (exit 1 on a hard regression)
    python3 scripts/cover_check.py --update   # re-bless floors to current
"""
import json
import os
import re
import subprocess
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
BASELINE = os.path.join(ROOT, "scripts", "coverage-baseline.json")

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
    hard = base.get("packages", {})
    soft = base.get("advisory", {})
    all_pkgs = sorted(hard) + sorted(soft)
    current = measure(all_pkgs)

    if update:
        base["packages"] = {p: current.get(p, 0.0) for p in hard}
        base["advisory"] = {p: current.get(p, 0.0) for p in soft}
        with open(BASELINE, "w") as f:
            json.dump(base, f, indent=2, sort_keys=True)
            f.write("\n")
        print("cover-update: baseline re-blessed to current coverage")
        for p in all_pkgs:
            tier = "advisory" if p in soft else "hard"
            print(f"  {current.get(p, 0.0):5.1f}%  [{tier}]  {p}")
        return 0

    failures = []
    print("Coverage ratchet — load-bearing (strict, no epsilon):")
    for p in sorted(hard):
        floor = hard[p]
        cur = current.get(p)
        if cur is None:
            failures.append(f"  MISSING  {p} — no coverage reported (package renamed/removed?)")
            continue
        ok = cur >= floor
        print(f"  [{'ok ' if ok else 'RED'}] {cur:5.1f}%  (floor {floor:5.1f}%)  {p}")
        if not ok:
            failures.append(f"  {p}: {cur:.1f}% < floor {floor:.1f}% — add a test or `make cover-update`")

    if soft:
        print("Advisory — reported only (not gated):")
        for p in sorted(soft):
            floor = soft[p]
            cur = current.get(p)
            if cur is None:
                print(f"  [warn] MISSING  {p}")
                continue
            flag = "warn" if cur < floor else "ok "
            note = "  ↓ regressed (advisory)" if cur < floor else ""
            print(f"  [{flag}] {cur:5.1f}%  (was {floor:5.1f}%)  {p}{note}")

    if failures:
        print("\ncover-check: FAIL — load-bearing coverage regressed:")
        print("\n".join(failures))
        return 1
    print("\ncover-check: PASS")
    return 0


if __name__ == "__main__":
    sys.exit(main())
