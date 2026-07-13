import { describe, it, expect } from "vitest";
import { conditionMatches, reachablePath } from "./logicEngine";
import type { LogicRule, Question } from "@/types/forms";

// Smoke test proving the harness runs and the TS engine's core operators behave.
// The exhaustive cross-engine check lives in logicEngine.parity.test.ts.
describe("conditionMatches (smoke)", () => {
  it("handles empty/always/is_empty semantics", () => {
    expect(conditionMatches(undefined, "always", null)).toBe(true);
    expect(conditionMatches(undefined, "is_empty", null)).toBe(true);
    expect(conditionMatches("x", "is_not_empty", null)).toBe(true);
    expect(conditionMatches(undefined, "eq", "x")).toBe(false);
  });

  it("compares scalars and matches substrings/membership", () => {
    expect(conditionMatches(5, "gt", 3)).toBe(true);
    expect(conditionMatches("hello world", "contains", "world")).toBe(true);
    expect(conditionMatches(["a", "b"], "contains", "b")).toBe(true);
  });
});

describe("reachablePath (smoke)", () => {
  const q = (id: string, position: number): Question =>
    ({ id, position, type: "short_text", title: id, required: false } as unknown as Question);

  it("walks a linear form in order when there are no rules", () => {
    const questions = [q("a", 0), q("b", 1), q("c", 2)];
    const rules: LogicRule[] = [];
    expect(reachablePath(questions, rules, {})).toEqual(["a", "b", "c"]);
  });
});
