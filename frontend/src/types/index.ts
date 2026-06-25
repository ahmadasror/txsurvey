// TS mirrors of the Go DTOs. Kept by hand — the surface is small.

export interface User {
  id: string;
  email: string;
  name: string;
  picture_url: string;
  created_at: string;
  updated_at: string;
}

export interface ApiSuccess<T> {
  success: true;
  data: T;
  message: string;
  meta?: { page: number; per_page: number; total: number };
}

export interface ApiError {
  success: false;
  error: { code: string; message: string; details?: unknown };
}
