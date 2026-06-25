// Single source of truth for path-prefix-aware URLs. Everything derives from
// Vite's BASE_URL (the resolved `base`, e.g. "/" in dev or "/txsurvey/" in the
// subpath production build), so the app works at the root or under a subpath
// without per-call hardcoding.

const BASE = import.meta.env.BASE_URL; // "/" or "/txsurvey/"

/** API base path. An explicit VITE_API_BASE_URL overrides; otherwise it is
 *  derived from the SPA base (e.g. "/txsurvey/api/v1"). */
export const apiBase = import.meta.env.VITE_API_BASE_URL ?? `${BASE}api/v1`;

/** React Router basename — undefined when mounted at the domain root. */
export const routerBasename = BASE === "/" ? undefined : BASE.replace(/\/$/, "");

/** App home (creator dashboard / login), prefix-aware. */
export const homePath = BASE;

/** Public runner path (in-app, prefix-aware) for a form slug. */
export const runnerPath = (slug: string) => `${BASE}r/${slug}`;

/** Absolute public runner URL (for share links). */
export const runnerUrl = (slug: string) => `${window.location.origin}${runnerPath(slug)}`;

/** assetUrl resolves a stored asset path ("uploads/x.png") to a prefix-aware URL. */
export const assetUrl = (p?: string): string | undefined =>
  p ? `${BASE}${p.replace(/^\//, "")}` : undefined;
