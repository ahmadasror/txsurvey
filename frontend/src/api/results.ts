import { useQuery } from "@tanstack/react-query";
import { api } from "@/api/client";
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

/** csvUrl is the same-origin export endpoint; a plain anchor download carries
 *  the session cookie automatically. */
export const csvUrl = (formId: string) =>
  `${import.meta.env.VITE_API_BASE_URL ?? "/api/v1"}/forms/${formId}/export.csv`;
