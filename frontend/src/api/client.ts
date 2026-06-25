import type { ApiError } from "@/types";
import { apiBase } from "@/lib/paths";

/** ApiRequestError carries the server's error code so callers can branch. */
export class ApiRequestError extends Error {
  status: number;
  code: string;
  constructor(status: number, code: string, message: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

/**
 * api performs a JSON request against the Go API. The session cookie rides
 * along via credentials:"include". On a non-2xx it throws ApiRequestError with
 * the server's {code,message}.
 */
export async function api<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${apiBase}${path}`, {
    credentials: "include",
    headers: { "Content-Type": "application/json", ...(init?.headers ?? {}) },
    ...init,
  });

  const text = await res.text();
  const body = text ? JSON.parse(text) : null;

  if (!res.ok) {
    const err = body as ApiError | null;
    throw new ApiRequestError(
      res.status,
      err?.error?.code ?? "ERROR",
      err?.error?.message ?? res.statusText,
    );
  }
  return (body?.data ?? null) as T;
}

/** loginUrl is the top-level navigation target that starts Google sign-in. */
export const loginUrl = () => `${apiBase}/auth/google/login`;
