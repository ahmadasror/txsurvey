import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/api/client";
import { apiBase } from "@/lib/paths";
import type { FormAnalytics, ResponseItem } from "@/types/forms";

export function useResponses(formId: string) {
  return useQuery<ResponseItem[]>({
    queryKey: ["responses", formId],
    queryFn: () => api<ResponseItem[]>(`/forms/${formId}/responses?per_page=200`),
  });
}

export function useAnalytics(formId: string) {
  return useQuery<FormAnalytics>({
    queryKey: ["analytics", formId],
    queryFn: () => api<FormAnalytics>(`/forms/${formId}/analytics`),
  });
}

/** useDeleteResponses clears all collected responses for a form, then refreshes
 *  the analytics + responses views. The form and its questions are untouched. */
export function useDeleteResponses(formId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => api<{ deleted: number }>(`/forms/${formId}/responses`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["analytics", formId] });
      qc.invalidateQueries({ queryKey: ["responses", formId] });
    },
  });
}

/** csvUrl is the same-origin export endpoint; a plain anchor download carries
 *  the session cookie automatically. */
export const csvUrl = (formId: string) => `${apiBase}/forms/${formId}/export.csv`;
