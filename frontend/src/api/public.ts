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

/** useSubmitResponse posts a completed submission. */
export function useSubmitResponse(slug: string) {
  return useMutation({
    mutationFn: (answers: SubmitAnswer[]) =>
      api<{ response_id: string }>(`/public/forms/${slug}/responses`, {
        method: "POST",
        body: JSON.stringify({ answers }),
      }),
  });
}
