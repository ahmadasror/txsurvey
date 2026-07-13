import { useMutation, useQuery } from "@tanstack/react-query";
import { api } from "@/api/client";
import type { AnswerValue, PublicForm } from "@/types/forms";

/** usePublicForm loads a published form for the runner (no auth). */
export function usePublicForm(slug: string) {
  return useQuery<PublicForm>({
    queryKey: ["public-form", slug],
    queryFn: () => api<PublicForm>(`/public/forms/${slug}`),
    retry: false,
  });
}

export interface SubmitAnswer {
  question_id: string;
  value: AnswerValue;
}

export interface SubmitPayload {
  answers: SubmitAnswer[];
  /** responseId, when present, echoes the paradata session opened by
   *  startResponseSession so the backend finalizes that in-progress row
   *  instead of inserting a new completed one (see internal/dto/submit.go). */
  responseId?: string | null;
}

/** useSubmitResponse posts a completed submission. */
export function useSubmitResponse(slug: string) {
  return useMutation({
    mutationFn: ({ answers, responseId }: SubmitPayload) =>
      api<{ response_id: string }>(`/public/forms/${slug}/responses`, {
        method: "POST",
        body: JSON.stringify({ answers, response_id: responseId || undefined }),
      }),
  });
}

/** startResponseSession opens an in-progress response for paradata capture
 *  (drop-off tracking). Best-effort: callers must swallow rejections — a
 *  failure here must never block the respondent from filling the form. */
export function startResponseSession(slug: string) {
  return api<{ response_id: string }>(`/public/forms/${slug}/start`, { method: "POST" });
}

/** pingProgress advances an in-progress response's furthest-reached question
 *  position. Fire-and-forget: callers must swallow rejections. */
export function pingProgress(slug: string, responseId: string, position: number) {
  return api<{ ok: boolean }>(`/public/forms/${slug}/progress`, {
    method: "POST",
    body: JSON.stringify({ response_id: responseId, position }),
  });
}
