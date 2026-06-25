import type { QuestionType, QuestionMetadata } from "@/types/forms";

export interface QuestionTypeDef {
  type: QuestionType;
  label: string;
  /** default metadata when a question of this type is first created */
  defaultMetadata: QuestionMetadata;
  /** does this type carry a fixed option set? */
  isChoice: boolean;
  /** display-only (no answer) */
  isStatement: boolean;
}

export const QUESTION_TYPES: QuestionTypeDef[] = [
  { type: "short_text", label: "Short text", defaultMetadata: {}, isChoice: false, isStatement: false },
  { type: "long_text", label: "Long text", defaultMetadata: {}, isChoice: false, isStatement: false },
  { type: "email", label: "Email", defaultMetadata: {}, isChoice: false, isStatement: false },
  { type: "number", label: "Number", defaultMetadata: {}, isChoice: false, isStatement: false },
  { type: "date", label: "Date", defaultMetadata: {}, isChoice: false, isStatement: false },
  {
    type: "multiple_choice",
    label: "Multiple choice",
    defaultMetadata: { options: [{ id: "", label: "Option 1" }] },
    isChoice: true,
    isStatement: false,
  },
  {
    type: "checkboxes",
    label: "Checkboxes",
    defaultMetadata: { options: [{ id: "", label: "Option 1" }] },
    isChoice: true,
    isStatement: false,
  },
  {
    type: "dropdown",
    label: "Dropdown",
    defaultMetadata: { options: [{ id: "", label: "Option 1" }] },
    isChoice: true,
    isStatement: false,
  },
  { type: "rating", label: "Rating", defaultMetadata: { scale: 5 }, isChoice: false, isStatement: false },
  { type: "yes_no", label: "Yes / No", defaultMetadata: {}, isChoice: false, isStatement: false },
  { type: "statement", label: "Statement", defaultMetadata: {}, isChoice: false, isStatement: true },
];

export const typeDef = (t: QuestionType): QuestionTypeDef =>
  QUESTION_TYPES.find((d) => d.type === t) ?? QUESTION_TYPES[0];

export const typeLabel = (t: QuestionType): string => typeDef(t).label;
