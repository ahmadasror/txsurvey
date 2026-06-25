export type QuestionType =
  | "short_text"
  | "long_text"
  | "multiple_choice"
  | "checkboxes"
  | "dropdown"
  | "rating"
  | "number"
  | "email"
  | "date"
  | "yes_no"
  | "statement";

export type FormStatus = "draft" | "published" | "closed";

export interface Option {
  id: string;
  label: string;
}

export interface QuestionMetadata {
  options?: Option[];
  min?: number;
  max?: number;
  step?: number;
  scale?: number;
  max_length?: number;
  placeholder?: string;
  allow_other?: boolean;
}

export interface Question {
  id: string;
  form_id: string;
  type: QuestionType;
  title: string;
  description: string;
  position: number;
  required: boolean;
  metadata: QuestionMetadata;
  created_at: string;
  updated_at: string;
}

export type LogicOperator =
  | "eq"
  | "neq"
  | "gt"
  | "gte"
  | "lt"
  | "lte"
  | "contains"
  | "not_contains"
  | "is_empty"
  | "is_not_empty";

export type LogicAction = "jump_to" | "show" | "hide" | "end_form";

export interface LogicRule {
  id: string;
  form_id: string;
  source_question_id: string;
  operator: LogicOperator;
  compare_value?: unknown;
  action: LogicAction;
  target_question_id?: string | null;
  priority: number;
  created_at: string;
}

export interface LogicRuleInput {
  source_question_id: string;
  operator: LogicOperator;
  compare_value?: unknown;
  action: LogicAction;
  target_question_id?: string | null;
  priority?: number;
}

export interface ThemeSettings {
  preset?: string;
  accent?: string;
}

export interface FormSettings {
  welcome_title?: string;
  welcome_description?: string;
  thank_you_title?: string;
  thank_you_description?: string;
  redirect_url?: string;
  show_progress: boolean;
  theme: ThemeSettings;
}

export interface Form {
  id: string;
  owner_id: string;
  title: string;
  description: string;
  slug: string;
  status: FormStatus;
  settings: FormSettings;
  published_at?: string;
  created_at: string;
  updated_at: string;
  questions?: Question[];
  logic_rules?: LogicRule[];
}

export interface FormListItem extends Form {
  question_count: number;
  response_count: number;
}

/** QuestionInput is the create/update payload (subset of Question). */
export interface QuestionInput {
  type: QuestionType;
  title: string;
  description?: string;
  required?: boolean;
  metadata?: QuestionMetadata;
}

/** PublicForm is the runner contract returned by GET /public/forms/:slug. */
export interface PublicForm {
  id: string;
  title: string;
  description: string;
  slug: string;
  settings: FormSettings;
  questions: Question[];
  logic_rules: LogicRule[];
}

/** AnswerValue is the per-type answer payload sent on submit. */
export type AnswerValue = string | number | boolean | string[];

export interface Answer {
  id: string;
  response_id: string;
  question_id: string;
  value: AnswerValue;
  created_at: string;
}

export interface ResponseItem {
  id: string;
  form_id: string;
  completed: boolean;
  submitted_at: string;
  answers: Answer[];
}

export interface OptionCount {
  value: string;
  label: string;
  count: number;
}

export interface QuestionSummary {
  question_id: string;
  title: string;
  type: QuestionType;
  answered: number;
  options?: OptionCount[];
  average?: number;
}

export interface FormAnalytics {
  response_count: number;
  completion_rate: number;
  questions: QuestionSummary[];
}
