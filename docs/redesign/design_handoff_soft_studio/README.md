# Handoff: txsurvey — "Soft Studio" UX & Visual Redesign

## Overview
A redesign of **txsurvey** (a Typeform-like survey builder) toward a calm, warm, editorial
direction called **Soft Studio**. It covers every core surface — **Login, Dashboard, Builder,
Design panel, Runner (respondent), Results, Templates, and Legal (Terms/Privacy)** — plus a
coherent 5-theme color system that replaces the old random presets.

The goal: a survey tool that feels like a conversation — one question per screen, generous
whitespace, one confident accent color, serif display type — without "AI-slop" tropes
(no gradient washes, no random emoji, no Inter/Roboto, no fake metrics).

## About the Design Files
The files in this bundle are **design references authored in HTML** (`*.dc.html` streaming
components + `support.js` runtime). They are **prototypes that demonstrate intended look and
behavior — not production code to copy verbatim.**

The task is to **recreate these designs in the existing txsurvey codebase** using its
established stack and patterns:
- **Frontend:** React SPA — Vite + TypeScript + Tailwind + shadcn/ui + TanStack Query.
- **Backend:** Go (Gin + pgx + PostgreSQL), API-first; SPA embedded in a single binary.
- Theme is applied via CSS variables on the runner container (see `lib/themes.ts` →
  `themeStyle`). Keep that mechanism; just swap the palettes and visual treatment.

To preview a reference: open any `txsurvey *.dc.html` in a browser (they load `support.js`
from the same folder). Recreate the result with shadcn/Tailwind components, not by porting the
HTML.

## Fidelity
**High-fidelity.** Final colors, typography, spacing, radii, copy, and interactions are all
specified below and present in the prototypes. Recreate the UI faithfully using the codebase's
existing component library (shadcn/ui) and Tailwind tokens. Where a value isn't stated, read it
from the corresponding `.dc.html` file's inline styles / logic class.

---

## Design Tokens

### Color — one system, five themes
Each theme shares the same structure: warm **paper** background + one confident **primary** +
dark **ink** text + one **accent**. Replace the old `THEME_PRESETS`
(Corporate/Fun/Comical/Girl/Boy) in `lib/themes.ts` with these. Default = **Pine**
(evolves the heritage teal-on-cream).

Full token set per theme (CSS-variable names used on the runner/container):

| token | Pine | Sand | Grape | Coral | Ink |
|---|---|---|---|---|---|
| `--paper` (bg) | `#F4F1EA` | `#FBF6EC` | `#F7F4FF` | `#FFF4F0` | `#F3F4F6` |
| surface (card) | `#FFFFFF` | `#FFFFFF` | `#FFFFFF` | `#FFFFFF` | `#FFFFFF` |
| `--ink` (text) | `#1D2624` | `#2B2317` | `#1A1530` | `#2A1A16` | `#11141B` |
| body text | `#46504C` | `#5C5240` | `#4A4360` | `#5A4339` | `#3A4150` |
| `--muted` | `#6B736E` | `#8A795A` | `#7A7393` | `#9A7A70` | `#6B7280` |
| `--border`/line | `#E3DDD0` | `#E9DDC6` | `#E3DCF7` | `#F3D8CF` | `#DFE2E7` |
| `--primary` | `#2F6F5E` | `#C2772E` | `#6C4CE0` | `#E5533D` | `#1F2430` |
| primary-soft | `#E7EFE9` | `#F6EAD5` | `#ECE6FC` | `#FBE2DB` | `#E6E8EC` |
| primary-fg | `#FFFFFF` | `#FFFFFF` | `#FFFFFF` | `#FFFFFF` | `#FFFFFF` |
| accent | `#D98E5A` | `#6B8F71` | `#FF6B5C` | `#2BB3A3` | `#6C4CE0` |
| ring (focus) | `rgba(47,111,94,.16)` | `rgba(194,119,46,.16)` | `rgba(108,76,224,.16)` | `rgba(229,83,61,.16)` | `rgba(31,36,48,.14)` |

The existing app stores tokens as HSL triples (`--primary: 180 50% 33%` etc.) and converts hex
via `lib/theme.ts → hexToHslTriple`. Keep that pipeline: define each theme's hexes, convert to
HSL triples, apply on the form/runner container. App chrome (Dashboard/Builder/Results) follows
the current form's theme too.

### Typography
- **Display / headings:** `Newsreader` (serif), weights 400/500/600 — used at weight **500**.
- **UI / body:** `Hanken Grotesk`, weights 400–800.
- **Per-form font choice** (set in the Design panel, applied in the Runner):
  - `Editorial` → Newsreader + Hanken (default)
  - `Modern` → `Bricolage Grotesque` + Hanken
  - `Soft` → Hanken only
  - `Serif` → `Spectral` + Hanken
- Load via Google Fonts in `index.html` / `index.css`. **Do not** use Inter/Roboto.
- Scale: question title 27–32px, welcome H1 34–42px, body 15–17px, labels 11.5–13px
  (uppercase, letter-spacing .05em for field labels), meta 12.5–13px.

### Shape, elevation, motion
- Radius: inputs/choices **11–16px**, cards **14–18px**, modals **18–20px**, pills **20px**.
- Border: `1px solid {line}`. Shadow (cards/modals): `0 14px 38px rgba(29,38,36,.12)`;
  hover lift `0 12px 26px rgba(29,38,36,.10)` + `translateY(-3px)`.
- Focus ring on selected choice: `0 0 0 3px {ring}`.
- Motion: question enter `opacity 0→1` + `translateY(22px→0)`, **~280–300ms**,
  `cubic-bezier(.2,.7,.3,1)`. Button hover `transform/box-shadow` ~150ms. Respect
  `prefers-reduced-motion`.
- Hit targets ≥ **44px**.

---

## Screens / Views

### 1. Login  (`features/auth/LoginPage.tsx`)
- **Purpose:** sign in with Google (sign-in only; invited test users).
- **Layout:** full-height split, `display:flex; flex-wrap:wrap`. Left panel `flex:1.1 1 420px`,
  right `flex:.9 1 380px`. Stacks on narrow.
- **Left (brand panel):** background `var(--primary)`, text white/cream. Wordmark top;
  eyebrow (uppercase, accent); serif H1 ~46px ("Bikin survei yang benar-benar diisi orang.");
  intro paragraph; 4 value rows each with a check-tick chip (`rgba(255,255,255,.16)` square +
  CSS checkmark) and 15.5px text; bottom: avatar stack + social-proof line. A few floating
  geometric shapes (`@keyframes floaty`, ~5s) and large faint circles for depth.
- **Right (card):** centered, max-width 380. Logo mark (52px rounded square, primary, inset
  rotated square w/ accent top border); H2 "Masuk ke txsurvey" (serif); subtitle; **Google
  button** (full width, 52px, `1px solid line`, white, 4-color Google "G" SVG + "Lanjut dengan
  Google"); divider "akun yang diundang"; info box (primary-soft) "mode uji — hanya akun yang
  diundang"; fine print with Ketentuan & Kebijakan Privasi links.
- **States:** Google button hover = lift + shadow + border darkens; on click → spinner +
  "Menghubungkan…" then redirect to dashboard.

### 2. Dashboard  (`features/dashboard/FormsListPage.tsx`, `DashboardLayout.tsx`)
- **Purpose:** list/manage surveys.
- **Layout:** top app bar (60px, surface, `1px` bottom border) with brand + nav
  (Surveys / Template) + user + avatar; content max-width ~980, padded 36px.
- **Header row:** serif H1 "Surveimu" + subtitle (count · total responses); primary CTA
  "+ Survei baru" (44px, primary, radius 12, soft shadow).
- **Cards grid:** 2-col, gap 16. Each card (surface, `1px line`, radius 16, padding 20):
  title (17px/700), status badge top-right, meta line, and a **7-bar mini trend** (bars use
  primary at varying opacity, `align-items:flex-end`). Draft card uses dashed border + paper
  bg. Hover: lift + shadow. Delete = subtle **text** button "Hapus" (muted → red on hover),
  not a loud icon; opens a confirm dialog.
- **Status badge:** Published = primary-soft/primary ("● Published"); Draft = neutral gray;
  Closed = accent-tinted.
- **New-survey dialog:** title input + "Buat survei" → creates draft, opens builder.

### 3. Builder  (`features/builder/BuilderPage.tsx`, `QuestionEditor.tsx`, `SortableQuestionList.tsx`, `LogicEditor.tsx`)
- **Purpose:** edit a survey's questions & settings.
- **Layout:** header bar (back, editable title input, status badge, right actions:
  **Design · Results · Preview · Share link · Publish**). Body grid `280px 1fr`, gap 22.
  Mobile: single pane with a list/editor toggle (existing pattern).
- **Left sidebar:** "+ Tambah pertanyaan" dropdown (types: Pilihan ganda, Ya/Tidak, Rating,
  Teks panjang, Teks singkat). Question list items: number badge + title; selected = primary
  border + primary-soft bg.
- **Editor pane (card, padding 26):** type chip; question title input (serif, paper bg);
  options editor (letter badge + text input + "×" remove, "+ Tambah pilihan"); type-specific
  notes (rating shows "Skala 1–N"; text shows free-answer note; yes/no shows fixed chips);
  footer: required toggle, move up/down, delete.
- **Actions:** Publish toggles status; Share copies the public link (shows "Tersalin ✓");
  Preview opens the Runner; Results navigates to results; Design opens the Design panel.

### 4. Design panel  (extend `features/builder/DesignDialog.tsx`)
- **Purpose:** style the survey (per-form settings).
- **Layout:** modal (max-width 780), grid `300px 1fr`. Left = **live welcome preview**;
  right = controls (scrollable).
- **Preview:** banner block (80px; shows uploaded image as `background:center/cover` or
  primary-soft placeholder), circular logo (54px, overlaps banner via `margin-top:-40px`,
  `box-shadow:0 0 0 3px paper`), eyebrow, serif title, desc, "Mulai →" button. Reflects theme +
  font + copy live.
- **Controls:** Theme (5 swatches; selected = 2px primary border); Font (4 rows, each shows
  its display face as "Aa", selected = primary-soft + primary border); **Logo & banner upload**
  (click or drag-drop image → FileReader/data URL → store on form settings; **no email/URL
  field**); Welcome title input; Subtitle textarea.
- **Persisted settings (per form):** `theme`, `font`, `logoUrl`, `bannerUrl`, `welcomeTitle`,
  `welcomeDesc`. In the real app these map to existing `settings.theme`, `banner_url`,
  `logo_url`, `welcome_title`, `welcome_description` plus a new `font` field.

### 5. Runner (respondent)  (`features/runner/RunnerPage.tsx`, `QuestionScreen.tsx`)
- **Purpose:** answer the survey, one question per screen.
- **Layout:** full-viewport, themed via the form's settings. Thin progress bar on top (primary
  fill). Centered column max-width ~600.
- **Welcome:** logo mark, eyebrow "Survei Internal", serif H1 (welcome title), desc, primary
  "Mulai →" button, meta line ("N pertanyaan · ±2 menit · anonim"). Optional banner/logo.
- **Question:** "NN → dari N" indicator (primary); serif question title; optional description;
  body by type:
  - choice / yes_no → ChoiceButton list (letter badge A/B/C; selected = primary border +
    primary-soft + ring; checkmark when selected)
  - rating → square buttons 1..scale (selected = primary fill + ring)
  - long/short → textarea/input (paper bg, primary focus)
  - Footer: primary "OK →" (or "Kirim" on last), "tekan Enter ↵" hint, "← Kembali".
- **Done:** circle with CSS checkmark (primary), serif "Makasih, sudah terkirim!", desc,
  "↺ Isi lagi" (prototype only). A couple of floating accent shapes.
- **Interactions:** Enter advances (not in textarea); digit keys 1–9 pick option/rating/yes-no;
  picking a choice/rating auto-advances after ~170ms; required-field validation message;
  enter animation per question (see Motion). These mirror the existing keyboard logic in
  `RunnerPage.tsx` — keep it.

### 6. Results  (`features/results/ResultsPage.tsx`)
- **Purpose:** read responses & analytics, export CSV.
- **Layout:** header (back, title, response count, **✎ Edit**, **↓ Download CSV**); tabs
  **Ringkasan / Respons** (active tab = paper-filled tab with top border, sitting on content).
- **Ringkasan:** 3 stat tiles (Respon, Penyelesaian %, Pertanyaan — numbers in serif; the %
  in primary). Then one card per question: title + "N jawaban"; for choice → horizontal bar
  list (track = primary-soft, fill = primary at decreasing opacity per option, tabular-nums
  counts right-aligned); for rating → big serif average; for open-ended → "lihat tab Respons".
- **Respons:** table (header row paper bg; rows separated by `1px line`). Download CSV builds a
  real CSV blob and downloads it.
- **Keep numbers honest** — no decorative/invented stats.

### 7. Templates  (new route; uses `api/forms.ts` createForm + question seeding)
- **Purpose:** start from a ready-made survey.
- **Layout:** same app bar (Template nav active); serif H1 "Mulai dari template"; 3-col card
  grid.
- **Cards:** accent-colored icon mark (its theme's primary), title, description, **type chips**
  (unique question types), footer meta ("N pertanyaan · ±M menit") + "Pakai →".
- **3 seed templates** (each pre-fills questions + theme + font + welcome copy):
  - **Pulse Karyawan** (Pine, Editorial) — tenure (choice), satisfaction (choice), NPS
    (rating 10), open improvement (long), follow-up (yes/no).
  - **Feedback Onboarding** (Sand, Editorial) — role clarity (rating 5), training adequacy
    (choice), mentor helpful (yes/no), 2× open text.
  - **Kepuasan Layanan IT** (Ink, Soft) — service used (choice), resolved (yes/no), speed
    (rating 5), overall satisfaction (choice), suggestions (long).
- "Pakai →" creates a draft form from the template and opens the Builder.

### 8. Legal — Terms & Privacy  (new static route)
- **Purpose:** Ketentuan Layanan & Kebijakan Privasi.
- **Layout:** sticky top bar (brand + "← Kembali ke masuk" → Login). Reading column max-width
  760. Eyebrow "Dokumen hukum", serif H1, "Terakhir diperbarui …", segmented tabs
  (Ketentuan / Privasi), intro, then numbered sections (NN + serif H2 + paragraphs + accent
  bullet lists).
- **Contact = form, not email.** Bottom card: "Ada pertanyaan soal dokumen ini?" with a
  **Nama (opsional)** input + **Pertanyaan** textarea + "Kirim pertanyaan" → success state
  ("Terima kasih, pertanyaanmu terkirim." + "Kirim lagi"). **Never collect an email address
  here.** Disclaimer: example doc, not legal advice.

---

## Interactions & Behavior
- **Navigation flow:** Login → (Google) → Dashboard; Dashboard ⇄ Templates; card → Builder;
  Builder → Results / Design / Preview(Runner); avatar → sign out → Login; Login → Legal → Login.
- **Animations:** question enter slide/fade ~280ms ease-out; modal/overlay fade; button hover
  lift; spinner on sign-in/connect.
- **Hover:** cards lift, buttons lift + stronger shadow, choices highlight border.
- **Loading:** spinner (border-spin) on async actions.
- **Errors:** inline required-field message in destructive color.
- **Validation:** required questions block "OK/Kirim"; contact form requires a non-empty
  message.
- **Responsive:** Login split stacks; Builder collapses to single pane (existing toggle);
  grids drop to 1 col on small screens.

## State Management
- **Dashboard/Studio:** `forms[]` (id, title, status, responseCount, completion, trend,
  questions[], analytics, settings{theme,font,logoUrl,bannerUrl,welcomeTitle,welcomeDesc}),
  `activeId`, `selectedQId`, dialogs (new/delete), `addMenuOpen`, `designOpen`, `resultsTab`,
  `copied`. In production these come from the API via TanStack Query + mutations
  (`api/forms.ts`, `api/results.ts`).
- **Runner:** `phase` (welcome|question|done), `idx`, `answers{}`, `error`. Navigation respects
  the conditional-logic engine (`lib/logicEngine.ts`) — only ask reachable questions; submit
  only reachable answers (already implemented; preserve).
- **Legal:** `tab`, contact `{name,msg,sent}`.

## Design Tokens (quick reference)
See the table above for the 5 palettes. Radius 11–20, border `1px line`, shadow
`0 14px 38px rgba(29,38,36,.12)`, ring `0 0 0 3px {ring}`, motion 150–300ms.

## Assets
- **No raster assets shipped.** Logo/brand mark is CSS (rounded square + inset rotated square
  with accent top border). Google "G" is an inline 4-color SVG (keep Google branding intact on
  the sign-in button). Checkmarks/ticks are CSS borders. Trend bars / charts are divs.
- Logo & banner are **user-uploaded** at runtime (data URL in prototype; map to existing
  `logo_url` / `banner_url` upload in production).

## Files (design references in this bundle)
- `txsurvey Login.dc.html` — Login (split UVP + Google).
- `txsurvey Studio.dc.html` — Dashboard + Builder + Design panel + Results + Templates (one app).
- `txsurvey Runner.dc.html` — respondent flow (all question types, keyboard, transitions).
- `txsurvey Legal.dc.html` — Terms/Privacy + contact form.
- `txsurvey UX Handoff.dc.html` — the human-readable UX guidance doc (printable).
- `txsurvey Redesign.dc.html` — the original 2-direction exploration (Soft Studio chosen).
- `support.js` — runtime needed to open the `.dc.html` files in a browser (reference only).

Open any `.dc.html` in a browser to see the intended result; implement in the React codebase.
