#!/usr/bin/env python3
"""Definition-of-Done gate — stop behavior changes from landing while their tests
and spec silently rot (the "tech debt discovered at the end" failure mode).

Rule: if the diff (this branch vs its base) touches BEHAVIOR-LAYER Go code —
internal/{handler,service,repository}/*.go, excluding *_test.go — then the same
diff MUST also touch:
  1. a Go test  (any *_test.go, or anything under tests/), and
  2. an FR spec (anything under docs/fr/).

So a new/changed endpoint can't merge with its test suite or its FR left behind.
This is the machine-checked half of Definition of Done: curation happens at the
moment of change, not as a periodic audit that accrues debt.

Escape hatch for a legitimate exception (pure refactor, comment/typo, mechanical
rename — no behavior change): add a trailer line

    DoD-Skip: <reason>

to any commit message in the range, or set DOD_SKIP=1 in the environment. The
reason is echoed so a reviewer sees the waiver.

Base detection: $DOD_BASE, else origin/main, else main. If none resolve, or the
range is empty (e.g. on main with nothing new), the check is a NO-OP — it never
fails spuriously on a shallow clone or a fresh repo.
"""
import os
import subprocess
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

# Behavior-layer prefixes (a change here is assumed to alter observable behavior).
BEHAVIOR_PREFIXES = (
    "internal/handler/",
    "internal/service/",
    "internal/repository/",
)


def git(*args):
    """Run git, returning stdout (stripped) or None on failure."""
    try:
        out = subprocess.run(
            ["git", *args], cwd=ROOT, capture_output=True, text=True
        )
    except FileNotFoundError:
        return None
    if out.returncode != 0:
        return None
    return out.stdout.strip()


def rev_exists(ref):
    return git("rev-parse", "--verify", "--quiet", ref) is not None


def resolve_base():
    env = os.environ.get("DOD_BASE")
    if env:
        return env if rev_exists(env) else None
    for cand in ("origin/main", "main"):
        if rev_exists(cand):
            return cand
    return None


def is_behavior_file(path):
    return (
        path.endswith(".go")
        and not path.endswith("_test.go")
        and path.startswith(BEHAVIOR_PREFIXES)
    )


def is_test_file(path):
    return path.endswith("_test.go") or path.startswith("tests/")


def is_fr_file(path):
    return path.startswith("docs/fr/")


def main():
    if os.environ.get("DOD_SKIP") == "1":
        print("dod-check: SKIP (DOD_SKIP=1 in environment)")
        return 0

    if git("rev-parse", "--is-inside-work-tree") != "true":
        print("dod-check: SKIP — not a git work tree")
        return 0

    base = resolve_base()
    if base is None:
        print("dod-check: SKIP — no base ref (origin/main|main) to diff against")
        return 0

    merge_base = git("merge-base", base, "HEAD")
    if not merge_base:
        print(f"dod-check: SKIP — no merge-base with {base}")
        return 0

    diff = git("diff", "--name-only", merge_base, "HEAD")
    changed = [line for line in (diff or "").splitlines() if line]
    if not changed:
        print(f"dod-check: PASS — no committed changes vs {base}")
        return 0

    behavior = [f for f in changed if is_behavior_file(f)]
    if not behavior:
        print(f"dod-check: PASS — no behavior-layer change vs {base}")
        return 0

    # Waiver via commit-message trailer anywhere in the range.
    bodies = git("log", "--format=%B", f"{merge_base}..HEAD") or ""
    for line in bodies.splitlines():
        if line.strip().lower().startswith("dod-skip:"):
            reason = line.split(":", 1)[1].strip() or "(no reason given)"
            print(f"dod-check: WAIVED via DoD-Skip trailer — {reason}")
            return 0

    has_test = any(is_test_file(f) for f in changed)
    has_fr = any(is_fr_file(f) for f in changed)
    if has_test and has_fr:
        print(f"dod-check: PASS — behavior change ships with test + FR (vs {base})")
        return 0

    print(f"dod-check: FAIL — behavior-layer change without its Definition of Done (vs {base})")
    print("  Behavior files changed:")
    for f in behavior:
        print(f"    - {f}")
    missing = []
    if not has_test:
        missing.append("a Go test (*_test.go or tests/**)")
    if not has_fr:
        missing.append("the owning FR (docs/fr/**)")
    print("  Missing: " + " and ".join(missing))
    print("  Fix: add the test/FR in this change, or waive a genuine refactor with a")
    print("       `DoD-Skip: <reason>` trailer in a commit message (or DOD_SKIP=1).")
    return 1


if __name__ == "__main__":
    sys.exit(main())
