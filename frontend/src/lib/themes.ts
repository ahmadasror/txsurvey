import type { CSSProperties } from "react";
import { hexToHslTriple } from "@/lib/theme";
import type { ThemeSettings } from "@/types/forms";

// ── Soft Studio palettes ─────────────────────────────────────────────────────
// One system, five themes: warm `paper` background + one confident `primary` +
// dark `ink` text + one decorative `accent`. Defined as hexes and converted to
// the "H S% L%" triples Tailwind's CSS variables expect (see lib/theme.ts) so a
// per-form theme can be applied by overriding the variables on a container.

interface ThemeTokens {
  paper: string; // page background
  ink: string; // headings / strong text
  body: string; // paragraph text
  muted: string; // secondary / meta text
  line: string; // borders & input outlines
  primary: string; // the one confident accent
  primarySoft: string; // primary tint (selected bg, tracks)
  accent: string; // decorative secondary accent (eyebrows, shapes)
}

export interface ThemePreset {
  id: string;
  label: string;
  swatch: string; // primary hex, for the picker chip
  accentSwatch: string; // accent hex, for the picker chip ring
  tokens: ThemeTokens;
  /** CSS variables (HSL triples) applied on a themed container. */
  vars: Record<string, string>;
}

const hsl = (hex: string) => hexToHslTriple(hex) ?? "0 0% 0%";

/** buildVars maps Soft Studio tokens onto the shadcn/ui CSS-variable contract
 *  (so every existing component keeps working) plus a few additions
 *  (--primary-soft, --body, --brand-accent). */
function buildVars(t: ThemeTokens): Record<string, string> {
  return {
    "--background": hsl(t.paper),
    "--foreground": hsl(t.ink),
    "--card": "0 0% 100%",
    "--card-foreground": hsl(t.ink),
    "--popover": "0 0% 100%",
    "--popover-foreground": hsl(t.ink),
    "--primary": hsl(t.primary),
    "--primary-foreground": "0 0% 100%",
    "--primary-soft": hsl(t.primarySoft),
    "--secondary": hsl(t.primarySoft),
    "--secondary-foreground": hsl(t.primary),
    // shadcn `--muted` is a *surface* token (track fills, subtle bg); the
    // Soft Studio "muted" is a *text* color → maps to --muted-foreground.
    "--muted": hsl(t.line),
    "--muted-foreground": hsl(t.muted),
    "--accent": hsl(t.primarySoft),
    "--accent-foreground": hsl(t.primary),
    "--brand-accent": hsl(t.accent),
    "--body": hsl(t.body),
    "--border": hsl(t.line),
    "--input": hsl(t.line),
    "--ring": hsl(t.primary),
  };
}

const THEME_DEFS: Array<Omit<ThemePreset, "vars">> = [
  {
    id: "pine",
    label: "Pine",
    swatch: "#2F6F5E",
    accentSwatch: "#D98E5A",
    tokens: {
      paper: "#F4F1EA",
      ink: "#1D2624",
      body: "#46504C",
      muted: "#6B736E",
      line: "#E3DDD0",
      primary: "#2F6F5E",
      primarySoft: "#E7EFE9",
      accent: "#D98E5A",
    },
  },
  {
    id: "sand",
    label: "Sand",
    swatch: "#C2772E",
    accentSwatch: "#6B8F71",
    tokens: {
      paper: "#FBF6EC",
      ink: "#2B2317",
      body: "#5C5240",
      muted: "#8A795A",
      line: "#E9DDC6",
      primary: "#C2772E",
      primarySoft: "#F6EAD5",
      accent: "#6B8F71",
    },
  },
  {
    id: "grape",
    label: "Grape",
    swatch: "#6C4CE0",
    accentSwatch: "#FF6B5C",
    tokens: {
      paper: "#F7F4FF",
      ink: "#1A1530",
      body: "#4A4360",
      muted: "#7A7393",
      line: "#E3DCF7",
      primary: "#6C4CE0",
      primarySoft: "#ECE6FC",
      accent: "#FF6B5C",
    },
  },
  {
    id: "coral",
    label: "Coral",
    swatch: "#E5533D",
    accentSwatch: "#2BB3A3",
    tokens: {
      paper: "#FFF4F0",
      ink: "#2A1A16",
      body: "#5A4339",
      muted: "#9A7A70",
      line: "#F3D8CF",
      primary: "#E5533D",
      primarySoft: "#FBE2DB",
      accent: "#2BB3A3",
    },
  },
  {
    id: "ink",
    label: "Ink",
    swatch: "#1F2430",
    accentSwatch: "#6C4CE0",
    tokens: {
      paper: "#F3F4F6",
      ink: "#11141B",
      body: "#3A4150",
      muted: "#6B7280",
      line: "#DFE2E7",
      primary: "#1F2430",
      primarySoft: "#E6E8EC",
      accent: "#6C4CE0",
    },
  },
];

export const THEME_PRESETS: ThemePreset[] = THEME_DEFS.map((d) => ({
  ...d,
  vars: buildVars(d.tokens),
}));

// Default = Pine (evolves the heritage teal-on-cream).
export const DEFAULT_THEME_ID = "pine";

export const presetById = (id?: string): ThemePreset | undefined =>
  THEME_PRESETS.find((p) => p.id === id);

// ── Per-form display font ─────────────────────────────────────────────────────
// UI/body type is always Hanken Grotesk; the per-form choice only swaps the
// *display* (heading) family, applied via the --font-display CSS variable.

export interface FontPreset {
  id: string;
  label: string;
  display: string; // CSS font-family stack for headings
  sample: string; // glyph shown in the picker
}

export const FONT_PRESETS: FontPreset[] = [
  { id: "editorial", label: "Editorial", display: '"Newsreader", ui-serif, Georgia, serif', sample: "Aa" },
  { id: "modern", label: "Modern", display: '"Bricolage Grotesque", ui-sans-serif, system-ui, sans-serif', sample: "Aa" },
  { id: "soft", label: "Soft", display: '"Hanken Grotesk", ui-sans-serif, system-ui, sans-serif', sample: "Aa" },
  { id: "serif", label: "Serif", display: '"Spectral", ui-serif, Georgia, serif', sample: "Aa" },
];

export const DEFAULT_FONT_ID = "editorial";

export const fontById = (id?: string): FontPreset =>
  FONT_PRESETS.find((f) => f.id === id) ?? FONT_PRESETS[0];

/** themeStyle resolves a form's ThemeSettings + font into the CSS-variable style
 *  a themed container applies. Unknown/legacy presets fall back to the app
 *  default (Pine, defined on :root) by returning only the font override. */
export function themeStyle(theme?: ThemeSettings, font?: string): CSSProperties | undefined {
  const preset = presetById(theme?.preset);
  const fontVars = { "--font-display": fontById(font).display } as Record<string, string>;

  if (preset) return { ...preset.vars, ...fontVars } as CSSProperties;
  if (theme?.accent) {
    const t = hexToHslTriple(theme.accent);
    if (t) return { "--primary": t, "--ring": t, ...fontVars } as CSSProperties;
  }
  return fontVars as CSSProperties;
}
