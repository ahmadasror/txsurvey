import { useEffect } from "react";

const OPTIONAL_FONT_URLS: Record<string, string> = {
  modern: "https://fonts.googleapis.com/css2?family=Bricolage+Grotesque:opsz,wght@12..96,400..700&display=swap",
  serif: "https://fonts.googleapis.com/css2?family=Spectral:wght@400;500;600&display=swap",
};

/** Loads non-default survey fonts only on routes/forms that need them. */
export function useOptionalFonts(...fontIds: Array<string | undefined>) {
  const key = [...new Set(fontIds.filter(Boolean))].sort().join(",");

  useEffect(() => {
    for (const fontId of key.split(",").filter(Boolean)) {
      const href = OPTIONAL_FONT_URLS[fontId];
      if (!href || document.head.querySelector(`link[data-txsurvey-font="${fontId}"]`)) continue;
      const link = document.createElement("link");
      link.rel = "stylesheet";
      link.href = href;
      link.dataset.txsurveyFont = fontId;
      document.head.appendChild(link);
    }
  }, [key]);
}
