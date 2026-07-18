import { useEffect } from "react";
import { homePath } from "@/lib/paths";

const DEFAULT_DESCRIPTION =
  "Buat survei yang terasa seperti percakapan, lengkap dengan logika bercabang, tema hangat, dan analitik respons.";

export interface PageMetadata {
  title?: string;
  description?: string;
  robots?: "index, follow" | "noindex, nofollow";
  path?: string;
  image?: string;
}

const upsertMeta = (attribute: "name" | "property", key: string, content: string) => {
  let element = document.head.querySelector<HTMLMetaElement>(`meta[${attribute}="${key}"]`);
  if (!element) {
    element = document.createElement("meta");
    element.setAttribute(attribute, key);
    element.dataset.txsurveyMetadata = "true";
    document.head.appendChild(element);
  }
  element.content = content;
};

const upsertCanonical = (href: string) => {
  let element = document.head.querySelector<HTMLLinkElement>('link[rel="canonical"]');
  if (!element) {
    element = document.createElement("link");
    element.rel = "canonical";
    element.dataset.txsurveyMetadata = "true";
    document.head.appendChild(element);
  }
  element.href = href;
};

const absoluteUrl = (path = "") => {
  const configuredOrigin = import.meta.env.VITE_SITE_URL?.replace(/\/$/, "") || window.location.origin;
  const relative = `${homePath}${path.replace(/^\//, "")}`;
  return new URL(relative, `${configuredOrigin}/`).toString();
};

/**
 * Publishes route-specific search and share metadata. Every routed page calls
 * this hook so client navigation replaces, rather than leaks, the previous
 * page's robots/canonical/Open Graph policy.
 */
export function usePageMetadata({
  title,
  description = DEFAULT_DESCRIPTION,
  robots = "noindex, nofollow",
  path,
  image,
}: PageMetadata) {
  useEffect(() => {
    const fullTitle = title ? `${title} · txsurvey` : "txsurvey";
    const canonical = path === undefined ? undefined : absoluteUrl(path);
    const shareImage = image ? absoluteUrl(image) : undefined;

    if (title) document.title = fullTitle;
    upsertMeta("name", "description", description);
    upsertMeta("name", "robots", robots);
    upsertMeta("property", "og:site_name", "txsurvey");
    upsertMeta("property", "og:type", "website");
    upsertMeta("property", "og:title", fullTitle);
    upsertMeta("property", "og:description", description);
    if (canonical) upsertMeta("property", "og:url", canonical);
    else document.head.querySelector('meta[property="og:url"]')?.remove();
    upsertMeta("name", "twitter:card", shareImage ? "summary_large_image" : "summary");
    upsertMeta("name", "twitter:title", fullTitle);
    upsertMeta("name", "twitter:description", description);
    if (shareImage) {
      upsertMeta("property", "og:image", shareImage);
      upsertMeta("name", "twitter:image", shareImage);
    } else {
      document.head.querySelectorAll('meta[property="og:image"], meta[name="twitter:image"]').forEach((element) => element.remove());
    }
    if (canonical) upsertCanonical(canonical);
    else document.head.querySelector('link[rel="canonical"]')?.remove();
  }, [description, image, path, robots, title]);
}
