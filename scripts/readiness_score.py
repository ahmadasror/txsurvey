#!/usr/bin/env python3
"""Deterministic model-readiness scorer (stack-agnostic ENGINE).

The score becomes a *function of observable facts* instead of a holistic 0-100
judgment: each dimension is a fixed set of BINARY criteria worth fixed points;
score = sum of met points. Same repo state -> identical output (no timestamps, no
randomness), so two runs (or two agents) produce the same number.

Portability split (this is the whole point):
  * GENERIC  — this engine + the *outcome-phrased* criteria (which live in
    readiness.yaml, mirrored/explained in docs/model-readiness-kit.md). They name
    outcomes ("the top invariant has a check that fails on violation"), never tools.
  * PER-REPO — each criterion's `probe` (a shell one-liner; exit 0 = met) or a
    `manual` met/not-met for the genuinely fuzzy ones. Only this layer is
    stack-specific, so the same engine scores a Go repo, a Java repo, etc.

Determinism boundary is reported explicitly: `probe` criteria are machine-verified
and reproducible; `manual` criteria are the evaluator-dependent residue.

Usage:
    python3 scripts/readiness_score.py [readiness.yaml] [--json]

Exit code is always 0 (this is a measurement, not a gate). Probes run with the
readiness.yaml's directory as CWD.
"""
import json
import subprocess
import sys
from pathlib import Path

import yaml


def run_probe(cmd: str, cwd: Path) -> tuple[bool, str]:
    """Run a probe; met == exit 0. Returns (met, short_detail)."""
    try:
        p = subprocess.run(
            cmd, shell=True, cwd=cwd, capture_output=True, text=True, timeout=300
        )
        return p.returncode == 0, f"exit {p.returncode}"
    except subprocess.TimeoutExpired:
        return False, "timeout"
    except Exception as e:  # noqa: BLE001
        return False, f"error: {e}"


def evaluate(doc: dict, base: Path) -> dict:
    dims = []
    for dim in doc.get("dimensions", []):
        crits = []
        for c in dim.get("criteria", []):
            pts = int(c["points"])
            if "probe" in c:
                met, detail = run_probe(c["probe"], base)
                kind = "probe"
            else:
                met = bool(c.get("met", False))
                detail = "manual"
                kind = "manual"
            crits.append({
                "id": c["id"], "points": pts, "met": met, "kind": kind,
                "desc": c.get("desc", ""), "detail": detail,
                "how": c.get("probe", c.get("note", "")),
            })
        earned = sum(c["points"] for c in crits if c["met"])
        weight = sum(c["points"] for c in crits)
        dims.append({
            "id": dim["id"], "name": dim.get("name", dim["id"]),
            "weight": weight, "earned": earned, "criteria": crits,
        })
    total = sum(d["earned"] for d in dims)
    weight_total = sum(d["weight"] for d in dims)
    all_crits = [c for d in dims for c in d["criteria"]]
    probe_pts = sum(c["points"] for c in all_crits if c["kind"] == "probe")
    manual_pts = sum(c["points"] for c in all_crits if c["kind"] == "manual")
    return {
        "repo": doc.get("meta", {}).get("repo", "?"),
        "total": total, "weight_total": weight_total, "dimensions": dims,
        "probe_points": probe_pts, "manual_points": manual_pts,
        "probe_count": sum(1 for c in all_crits if c["kind"] == "probe"),
        "manual_count": sum(1 for c in all_crits if c["kind"] == "manual"),
    }


def render(r: dict) -> str:
    out = []
    out.append(f"Model-Readiness score — {r['repo']}")
    out.append("=" * 58)
    for d in r["dimensions"]:
        out.append(f"\n[{d['earned']:>2}/{d['weight']:>2}]  {d['name']}")
        for c in d["criteria"]:
            mark = "x" if c["met"] else " "
            tag = "probe " if c["kind"] == "probe" else "MANUAL"
            out.append(f"    [{mark}] {c['points']:>2}  ({tag}) {c['id']}: {c['desc']}")
    out.append("\n" + "-" * 58)
    out.append(f"TOTAL: {r['total']} / {r['weight_total']}")
    out.append(
        f"  machine-verified (probe): {r['probe_points']} pts across "
        f"{r['probe_count']} criteria — reproducible"
    )
    out.append(
        f"  evaluator-dependent (manual): {r['manual_points']} pts across "
        f"{r['manual_count']} criteria — the only variance surface"
    )
    return "\n".join(out)


def main() -> int:
    args = [a for a in sys.argv[1:] if a != "--json"]
    as_json = "--json" in sys.argv[1:]
    path = Path(args[0]) if args else Path("readiness.yaml")
    if not path.exists():
        sys.exit(f"readiness_score: {path} not found")
    doc = yaml.safe_load(path.read_text())
    r = evaluate(doc, path.resolve().parent)
    if as_json:
        # Stable JSON (sorted keys) so two runs diff-identical.
        print(json.dumps(r, indent=2, sort_keys=True))
    else:
        print(render(r))
    return 0


if __name__ == "__main__":
    sys.exit(main())
