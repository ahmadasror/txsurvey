import type { CSSProperties } from "react";
import { hexToHslTriple } from "@/lib/theme";
import type { ThemeSettings } from "@/types/forms";

export interface ThemePreset {
  id: string;
  label: string;
  emoji: string;
  swatch: string; // primary hex, for the picker chip
  /** CSS variables (HSL triples) applied on the runner container. */
  vars: Record<string, string>;
}

// Default (corporate) mirrors the txhcs palette: teal on warm cream.
export const THEME_PRESETS: ThemePreset[] = [
  {
    id: "corporate",
    label: "Corporate",
    emoji: "🏢",
    swatch: "#2a7f7f",
    vars: {
      "--primary": "180 50% 33%",
      "--primary-foreground": "0 0% 100%",
      "--ring": "180 50% 33%",
      "--background": "30 15% 97%",
      "--foreground": "200 20% 14%",
      "--accent": "180 40% 92%",
      "--accent-foreground": "180 50% 22%",
      "--muted": "30 15% 94%",
      "--muted-foreground": "200 12% 42%",
      "--border": "30 12% 88%",
    },
  },
  {
    id: "fun",
    label: "Fun",
    emoji: "🎉",
    swatch: "#f97316",
    vars: {
      "--primary": "21 90% 52%",
      "--primary-foreground": "0 0% 100%",
      "--ring": "21 90% 52%",
      "--background": "40 100% 98%",
      "--foreground": "24 30% 16%",
      "--accent": "21 90% 93%",
      "--accent-foreground": "21 80% 30%",
      "--muted": "40 50% 94%",
      "--muted-foreground": "24 20% 42%",
      "--border": "30 40% 88%",
    },
  },
  {
    id: "comical",
    label: "Comical",
    emoji: "🤪",
    swatch: "#7c3aed",
    vars: {
      "--primary": "268 70% 52%",
      "--primary-foreground": "0 0% 100%",
      "--ring": "268 70% 52%",
      "--background": "50 100% 96%",
      "--foreground": "268 35% 18%",
      "--accent": "50 100% 84%",
      "--accent-foreground": "268 60% 28%",
      "--muted": "50 60% 92%",
      "--muted-foreground": "268 18% 42%",
      "--border": "50 50% 84%",
    },
  },
  {
    id: "girl",
    label: "Girl",
    emoji: "🌸",
    swatch: "#ec4899",
    vars: {
      "--primary": "330 75% 56%",
      "--primary-foreground": "0 0% 100%",
      "--ring": "330 75% 56%",
      "--background": "340 70% 98%",
      "--foreground": "330 30% 20%",
      "--accent": "330 75% 94%",
      "--accent-foreground": "330 60% 34%",
      "--muted": "340 40% 95%",
      "--muted-foreground": "330 15% 45%",
      "--border": "335 40% 90%",
    },
  },
  {
    id: "boy",
    label: "Boy",
    emoji: "🚀",
    swatch: "#2563eb",
    vars: {
      "--primary": "221 83% 53%",
      "--primary-foreground": "0 0% 100%",
      "--ring": "221 83% 53%",
      "--background": "214 100% 98%",
      "--foreground": "221 35% 18%",
      "--accent": "221 83% 94%",
      "--accent-foreground": "221 70% 30%",
      "--muted": "214 45% 95%",
      "--muted-foreground": "221 15% 45%",
      "--border": "214 40% 90%",
    },
  },
];

export const DEFAULT_THEME_ID = "corporate";

export const presetById = (id?: string): ThemePreset | undefined =>
  THEME_PRESETS.find((p) => p.id === id);

/** themeStyle resolves a form's ThemeSettings into the CSS-variable style the
 *  runner container applies. Falls back to a legacy accent hex, then to the
 *  app default theme (no override). */
export function themeStyle(theme?: ThemeSettings): CSSProperties | undefined {
  const preset = presetById(theme?.preset);
  if (preset) return preset.vars as CSSProperties;
  if (theme?.accent) {
    const t = hexToHslTriple(theme.accent);
    if (t) return { "--primary": t, "--ring": t } as CSSProperties;
  }
  return undefined;
}
