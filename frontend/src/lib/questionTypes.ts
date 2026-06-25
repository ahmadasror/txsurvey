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
  { type: "short_text", label: "Teks singkat", defaultMetadata: {}, isChoice: false, isStatement: false },
  { type: "long_text", label: "Teks panjang", defaultMetadata: {}, isChoice: false, isStatement: false },
  { type: "email", label: "Email", defaultMetadata: {}, isChoice: false, isStatement: false },
  { type: "number", label: "Angka", defaultMetadata: {}, isChoice: false, isStatement: false },
  { type: "date", label: "Tanggal", defaultMetadata: {}, isChoice: false, isStatement: false },
  {
    type: "multiple_choice",
    label: "Pilihan ganda",
    defaultMetadata: { options: [{ id: "", label: "Pilihan 1" }] },
    isChoice: true,
    isStatement: false,
  },
  {
    type: "checkboxes",
    label: "Kotak centang",
    defaultMetadata: { options: [{ id: "", label: "Pilihan 1" }] },
    isChoice: true,
    isStatement: false,
  },
  {
    type: "dropdown",
    label: "Dropdown",
    defaultMetadata: { options: [{ id: "", label: "Pilihan 1" }] },
    isChoice: true,
    isStatement: false,
  },
  { type: "rating", label: "Rating", defaultMetadata: { scale: 5 }, isChoice: false, isStatement: false },
  { type: "yes_no", label: "Ya / Tidak", defaultMetadata: {}, isChoice: false, isStatement: false },
  { type: "statement", label: "Pernyataan", defaultMetadata: {}, isChoice: false, isStatement: true },
];

export const typeDef = (t: QuestionType): QuestionTypeDef =>
  QUESTION_TYPES.find((d) => d.type === t) ?? QUESTION_TYPES[0];

export const typeLabel = (t: QuestionType): string => typeDef(t).label;
