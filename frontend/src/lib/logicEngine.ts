// logicEngine.ts MIRRORS internal/service/logic_engine.go — keep them in
// lockstep. The runner uses it for navigation (jump/show/hide/end); the Go side
// re-validates the reachable path on submit, so this is UX only, never trusted.
import type { AnswerValue, LogicOperator, LogicRule, Question } from "@/types/forms";

type Answers = Record<string, AnswerValue | undefined>;

const isEmpty = (v: AnswerValue | undefined): boolean =>
  v === undefined || v === null || v === "" || (Array.isArray(v) && v.length === 0);

const valueEqual = (a: unknown, b: unknown): boolean => {
  if (typeof a === "number" || typeof b === "number") return Number(a) === Number(b);
  return String(a) === String(b);
};

export function conditionMatches(answer: AnswerValue | undefined, op: LogicOperator, compare: unknown): boolean {
  const empty = isEmpty(answer);
  if (op === "is_empty") return empty;
  if (op === "is_not_empty") return !empty;
  if (empty) return false;

  switch (op) {
    case "eq":
      return valueEqual(answer, compare);
    case "neq":
      return !valueEqual(answer, compare);
    case "gt":
      return Number(answer) > Number(compare);
    case "gte":
      return Number(answer) >= Number(compare);
    case "lt":
      return Number(answer) < Number(compare);
    case "lte":
      return Number(answer) <= Number(compare);
    case "contains":
      return Array.isArray(answer)
        ? answer.some((x) => valueEqual(x, compare))
        : String(answer).includes(String(compare));
    case "not_contains":
      return Array.isArray(answer)
        ? !answer.some((x) => valueEqual(x, compare))
        : !String(answer).includes(String(compare));
    default:
      return false;
  }
}

interface Nav {
  ordered: Question[];
  firstId: () => string | null;
  nextFrom: (currentId: string) => string | null;
}

/** makeNav builds the navigation helpers for a given answer state. */
function makeNav(questions: Question[], rules: LogicRule[], answers: Answers): Nav {
  const ordered = [...questions].sort((a, b) => a.position - b.position);
  const posByID = new Map(ordered.map((q, i) => [q.id, i]));

  const bySource = new Map<string, LogicRule[]>();
  const showByTarget = new Map<string, LogicRule[]>();
  const hideByTarget = new Map<string, LogicRule[]>();
  for (const r of rules) {
    if (r.action === "jump_to" || r.action === "end_form") {
      (bySource.get(r.source_question_id) ?? bySource.set(r.source_question_id, []).get(r.source_question_id)!).push(r);
    } else if (r.target_question_id) {
      const m = r.action === "show" ? showByTarget : hideByTarget;
      (m.get(r.target_question_id) ?? m.set(r.target_question_id, []).get(r.target_question_id)!).push(r);
    }
  }
  for (const list of bySource.values()) list.sort((a, b) => a.priority - b.priority);

  const matches = (r: LogicRule) => conditionMatches(answers[r.source_question_id], r.operator, r.compare_value);

  const visible = (qid: string): boolean => {
    const hidden = (hideByTarget.get(qid) ?? []).some(matches);
    const shows = showByTarget.get(qid) ?? [];
    const base = shows.length ? shows.some(matches) : true;
    return base && !hidden;
  };

  const firstVisibleFrom = (idx: number): number => {
    for (let i = idx; i < ordered.length; i++) if (visible(ordered[i].id)) return i;
    return -1;
  };

  const idAt = (i: number) => (i >= 0 ? ordered[i].id : null);

  const nextFrom = (currentId: string): string | null => {
    const curIdx = posByID.get(currentId);
    if (curIdx === undefined) return null;
    for (const r of bySource.get(currentId) ?? []) {
      if (!matches(r)) continue;
      if (r.action === "end_form") return null;
      if (r.target_question_id && posByID.has(r.target_question_id)) {
        return idAt(firstVisibleFrom(posByID.get(r.target_question_id)!));
      }
    }
    return idAt(firstVisibleFrom(curIdx + 1));
  };

  return { ordered, firstId: () => idAt(firstVisibleFrom(0)), nextFrom };
}

/** firstQuestionId returns the first reachable question id (or null). */
export function firstQuestionId(questions: Question[], rules: LogicRule[], answers: Answers): string | null {
  return makeNav(questions, rules, answers).firstId();
}

/** nextQuestionId returns the next reachable question after currentId, or null
 *  (meaning the form ends). */
export function nextQuestionId(
  questions: Question[],
  rules: LogicRule[],
  answers: Answers,
  currentId: string,
): string | null {
  return makeNav(questions, rules, answers).nextFrom(currentId);
}

/** reachablePath replays the full reachable order for the current answers —
 *  used to size the progress bar. */
export function reachablePath(questions: Question[], rules: LogicRule[], answers: Answers): string[] {
  const nav = makeNav(questions, rules, answers);
  const path: string[] = [];
  const seen = new Set<string>();
  let cur = nav.firstId();
  let steps = 0;
  while (cur && !seen.has(cur) && steps <= nav.ordered.length) {
    seen.add(cur);
    path.push(cur);
    cur = nav.nextFrom(cur);
    steps++;
  }
  return path;
}
