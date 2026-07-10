import { useEffect } from "react";

const APP_NAME = "txsurvey";

/**
 * useDocumentTitle sets the browser-tab title as a breadcrumb that always ends
 * in the app name — e.g. useDocumentTitle("Hasil", form?.title) renders
 * "Hasil · Sprint Review PROSPERA · txsurvey". Falsy parts are dropped, so a
 * still-loading title collapses gracefully to just "txsurvey".
 */
export function useDocumentTitle(...parts: Array<string | null | undefined>) {
  const title = [...parts.filter(Boolean), APP_NAME].join(" · ");
  useEffect(() => {
    document.title = title;
  }, [title]);
}
