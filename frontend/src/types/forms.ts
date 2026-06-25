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

export interface ThemeSettings {
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
