import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import type { Form, FormListItem, LogicRule, LogicRuleInput, Question, QuestionInput } from "@/types/forms";

const formKey = (id: string) => ["form", id] as const;
const formsKey = ["forms"] as const;

export function useForms() {
  return useQuery<FormListItem[]>({
    queryKey: formsKey,
    queryFn: () => api<FormListItem[]>("/forms"),
  });
}

export function useForm(id: string) {
  return useQuery<Form>({
    queryKey: formKey(id),
    queryFn: () => api<Form>(`/forms/${id}`),
  });
}

export function useCreateForm() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (title: string) => api<Form>("/forms", { method: "POST", body: JSON.stringify({ title }) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: formsKey }),
  });
}

export function useDeleteForm() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api<null>(`/forms/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: formsKey }),
  });
}

export function useUpdateForm(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { title: string; description: string; settings: Form["settings"] }) =>
      api<Form>(`/forms/${id}`, { method: "PATCH", body: JSON.stringify(body) }),
    onSuccess: (form) => {
      qc.setQueryData(formKey(id), form);
      qc.invalidateQueries({ queryKey: formsKey });
    },
  });
}

export function usePublishForm(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (publish: boolean) =>
      api<Form>(`/forms/${id}/${publish ? "publish" : "unpublish"}`, { method: "POST" }),
    onSuccess: (form) => {
      qc.setQueryData(formKey(id), form);
      qc.invalidateQueries({ queryKey: formsKey });
    },
  });
}

// --- Questions (nested) — all invalidate the parent form detail ---

export function useAddQuestion(formId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: QuestionInput) =>
      api<Question>(`/forms/${formId}/questions`, { method: "POST", body: JSON.stringify(input) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: formKey(formId) }),
  });
}

export function useUpdateQuestion(formId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ qid, input }: { qid: string; input: QuestionInput }) =>
      api<Question>(`/forms/${formId}/questions/${qid}`, { method: "PATCH", body: JSON.stringify(input) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: formKey(formId) }),
  });
}

export function useDeleteQuestion(formId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (qid: string) => api<null>(`/forms/${formId}/questions/${qid}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: formKey(formId) }),
  });
}

export function useReorderQuestions(formId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (orderedIds: string[]) =>
      api<null>(`/forms/${formId}/questions/reorder`, {
        method: "PUT",
        body: JSON.stringify({ ordered_ids: orderedIds }),
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: formKey(formId) }),
  });
}

// --- Logic rules (nested) — invalidate the parent form detail ---

export function useAddLogicRule(formId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: LogicRuleInput) =>
      api<LogicRule>(`/forms/${formId}/logic`, { method: "POST", body: JSON.stringify(input) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: formKey(formId) }),
  });
}

export function useUpdateLogicRule(formId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ rid, input }: { rid: string; input: LogicRuleInput }) =>
      api<LogicRule>(`/forms/${formId}/logic/${rid}`, { method: "PATCH", body: JSON.stringify(input) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: formKey(formId) }),
  });
}

export function useDeleteLogicRule(formId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (rid: string) => api<null>(`/forms/${formId}/logic/${rid}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: formKey(formId) }),
  });
}
