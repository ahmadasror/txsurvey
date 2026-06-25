import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, ApiRequestError } from "@/api/client";
import type { User } from "@/types";

/** useMe loads the signed-in creator. A 401 resolves to null (not an error). */
export function useMe() {
  return useQuery<User | null>({
    queryKey: ["me"],
    queryFn: async () => {
      try {
        return await api<User>("/auth/me");
      } catch (e) {
        if (e instanceof ApiRequestError && e.status === 401) return null;
        throw e;
      }
    },
    retry: false,
    staleTime: 60_000,
  });
}

/** useLogout clears the session server-side then resets cached identity. */
export function useLogout() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => api<null>("/auth/logout", { method: "POST" }),
    onSuccess: () => {
      qc.setQueryData(["me"], null);
      qc.clear();
    },
  });
}
