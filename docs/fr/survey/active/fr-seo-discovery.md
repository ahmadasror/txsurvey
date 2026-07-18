# FR — SEO Discovery & Public Content

> **Module**: survey
> **Status**: Implemented
> **Date**: 2026-07-18
> **ADRs**: `docs/architecture/adr/005-server-enriched-spa-seo.md`
> **Sisters**: `docs/fr/survey/active/fr-forms.md`, `docs/fr/survey/active/fr-runner.md`

---

## 1. Background

txsurvey is a client-rendered application with creator-only workspaces and public
survey links. Search engines should discover useful product education without
indexing private application surfaces or respondent-facing surveys by default.

---

## 2. Functional Requirements

### FR-SEO-001 — Explicit indexing policy

- AC-001-1: login, creator dashboard, templates workspace, builder, and results pages emit `noindex, nofollow`.
- AC-001-2: respondent survey pages emit `noindex, nofollow` by default to protect user-generated and potentially sensitive content.
- AC-001-3: legal and product-education pages emit `index, follow` with a canonical URL.

### FR-SEO-002 — Route-specific metadata

- AC-002-1: every public product page has a unique title and description.
- AC-002-2: indexable pages publish canonical, Open Graph, and Twitter metadata.
- AC-002-3: metadata is replaced on client-side navigation so tags from a previous route cannot leak into the next page.

### FR-SEO-003 — Public discovery content

- AC-003-1: public template previews cover employee pulse, onboarding feedback, and IT service satisfaction.
- AC-003-2: public feature pages explain conditional logic and anonymous surveys.
- AC-003-3: a public guide answers common setup, sharing, response, and privacy questions.
- AC-003-4: all public pages are connected with ordinary crawlable links and a consistent public navigation/footer.

### FR-SEO-004 — Performance-aware routing

- AC-004-1: feature pages are loaded through route-level dynamic imports.
- AC-004-2: authenticated builder/results code is not part of the initial public-page route chunk.

### FR-SEO-005 — Crawlable first response

- AC-005-1: `/` is a public product landing page; the creator workspace lives at `/app`.
- AC-005-2: production HTML responses contain route-specific title, description, robots, canonical, Open Graph, and Twitter metadata before JavaScript runs.
- AC-005-3: indexable routes contain a concise server-rendered content fallback with an H1 and crawlable internal links.
- AC-005-4: unknown non-application paths return HTTP 404 with `noindex, nofollow`.

### FR-SEO-006 — Discovery files and canonical host

- AC-006-1: `/sitemap.xml` returns XML containing only canonical, indexable public pages.
- AC-006-2: `/robots.txt` allows crawling and advertises the sitemap URL.
- AC-006-3: social preview and favicon assets are served from stable public paths.
- AC-006-4: edge configuration redirects the legacy host to `https://brainzap.net` while preserving the path and query string.

---

## 3. Route & Indexing Matrix

| Route | Audience | Robots |
|---|---|---|
| `/login` | creator sign-in | `noindex, nofollow` |
| `/` | public product landing | `index, follow` |
| `/app`, `/templates`, `/forms/*` | creator workspace | `noindex, nofollow` |
| `/r/:slug` | respondents | `noindex, nofollow` |
| `/legal` | public trust | `index, follow` |
| `/template-*` | public discovery | `index, follow` |
| `/fitur/*` | public discovery | `index, follow` |
| `/panduan` | public discovery | `index, follow` |

---

## 4. Contract (machine-readable)

> Drift-detector source. Schema: `docs/fr/_contract-schema.json`.

```yaml
fr_file: docs/fr/survey/active/fr-seo-discovery.md
covers:
  - FR-SEO-001
  - FR-SEO-002
  - FR-SEO-003
  - FR-SEO-004
  - FR-SEO-005
  - FR-SEO-006

cross_links:
  adr_refs:
    - docs/architecture/adr/005-server-enriched-spa-seo.md
  sisters:
    - docs/fr/survey/active/fr-forms.md
    - docs/fr/survey/active/fr-runner.md
```
