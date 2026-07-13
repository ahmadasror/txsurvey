import { describe, it, expect } from "vitest";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { conditionMatches } from "./logicEngine";
import type { AnswerValue, LogicOperator } from "@/types/forms";

// TS half of the DUAL-ENGINE PARITY GATE. Reads the SAME fixture the Go test reads
// (internal/service/testdata/logic_parity_cases.json) and asserts the TS engine
// agrees with `expected` on every row. If logic_engine.go and logicEngine.ts drift
// apart, one of the two suites goes red. See docs/architecture/adr/001-dual-logic-engine.md.
interface ParityCase {
  name: string;
  answer: AnswerValue | null;
  operator: LogicOperator;
  compareValue: unknown;
  expected: boolean;
}

// frontend/src/lib -> repo root -> internal/service/testdata/...
const fixturePath = fileURLToPath(
  new URL("../../../internal/service/testdata/logic_parity_cases.json", import.meta.url),
);
const fixture = JSON.parse(readFileSync(fixturePath, "utf8")) as { cases: ParityCase[] };

describe("dual-engine parity (TS side of the shared fixture)", () => {
  it("has a non-empty fixture", () => {
    expect(fixture.cases.length).toBeGreaterThan(0);
  });

  for (const c of fixture.cases) {
    it(c.name, () => {
      const answer = (c.answer ?? undefined) as AnswerValue | undefined;
      expect(conditionMatches(answer, c.operator, c.compareValue)).toBe(c.expected);
    });
  }
});
