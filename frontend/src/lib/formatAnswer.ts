import type { AnswerValue, Question } from "@/types/forms";

/** formatAnswer renders an answer value as readable text, resolving option ids
 *  to their labels (mirrors the server-side CSV formatting). */
export function formatAnswer(q: Question | undefined, value: AnswerValue | undefined): string {
  if (!q || value === undefined || value === null) return "";
  const labelById = new Map((q.metadata.options ?? []).map((o) => [o.id, o.label]));
  switch (q.type) {
    case "multiple_choice":
    case "dropdown":
      return labelById.get(String(value)) ?? String(value);
    case "checkboxes":
      return Array.isArray(value) ? value.map((v) => labelById.get(v) ?? v).join(", ") : String(value);
    case "yes_no":
      return value === true ? "Yes" : value === false ? "No" : String(value);
    default:
      return String(value);
  }
}
