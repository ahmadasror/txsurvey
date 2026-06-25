# FR — Google Sign-in + Session

> **Module**: survey
> **Status**: Implemented (backfilled)
> **Date**: 2026-06-25
> **ADRs**: `docs/architecture/adr/002-google-signin-cookie-session.md`
> **Sisters**: `docs/fr/survey/active/fr-forms.md`

---

## 1. Background

Creators authenticate with Google (sign-in only). After the OAuth handshake the app
mints its own HS256 session JWT and stores it in an httpOnly cookie. Respondents are
anonymous (public link) and never hit these endpoints.

---

## 2. Functional Requirements

### FR-AUTH-001 — Start Google sign-in
**Code**: `internal/handler/auth_handler.go` (`GoogleLogin`), `internal/service/auth_service.go` (`AuthURL`)

- AC-001-1: `GET /auth/google/login` 302-redirects to Google's consent URL with `state` + PKCE `code_challenge` + scopes `openid email profile`.
- AC-001-2: 503 `GOOGLE_NOT_CONFIGURED` when client id/secret are unset.

### FR-AUTH-002 — OAuth callback → session
**Code**: `auth_handler.go` (`GoogleCallback`), `auth_service.go` (`HandleCallback`)

- AC-002-1: verifies `state` (CSRF) and exchanges the code with the PKCE verifier.
- AC-002-2: upserts the user by `google_sub`, mints a session JWT, sets the `session` httpOnly cookie (`SameSite=Lax`, path = `COOKIE_PATH`), and 302s to `APP_BASE_URL`.
- AC-002-3: 400 `OAUTH_STATE` on an invalid/expired state.

### FR-AUTH-003 — Current user / logout
**Code**: `auth_handler.go` (`Me`, `Logout`), `internal/middleware/auth.go` (`SessionAuth`)

- AC-003-1: `GET /auth/me` returns the signed-in creator; 401 `UNAUTHORIZED` without a valid cookie.
- AC-003-2: `POST /auth/logout` clears the session cookie (Max-Age 0).

---

## 3. API Surface

| Method | Path | Auth |
|---|---|---|
| GET | `/api/v1/auth/google/login` | public |
| GET | `/api/v1/auth/google/callback` | public |
| GET | `/api/v1/auth/me` | session |
| POST | `/api/v1/auth/logout` | session |

### Error Codes

| Code | HTTP | Trigger |
|---|---|---|
| `GOOGLE_NOT_CONFIGURED` | 503 | OAuth creds unset |
| `OAUTH_STATE` | 400 | Invalid/expired state |
| `UNAUTHORIZED` | 401 | Missing/invalid session cookie |

---

## 4. Data Model Touch Points

| Table | Read | Write | Notes |
|---|:---:|:---:|---|
| `users` | ✓ | ✓ | upsert on `google_sub` |

---

## 5. Cross-links

- ADR: `docs/architecture/adr/002-google-signin-cookie-session.md`

---

## 6. Contract (machine-readable)

> Drift-detector source. Schema: `docs/fr/_contract-schema.json`.

```yaml
fr_file: docs/fr/survey/active/fr-auth.md
covers:
  - FR-AUTH-001
  - FR-AUTH-002
  - FR-AUTH-003

endpoints:
  - id: AC-AUTH-LOGIN
    method: GET
    path: /api/v1/auth/google/login
    auth: { mode: public }
  - id: AC-AUTH-CALLBACK
    method: GET
    path: /api/v1/auth/google/callback
    auth: { mode: public }
  - id: AC-AUTH-ME
    method: GET
    path: /api/v1/auth/me
    auth: { mode: session }
  - id: AC-AUTH-LOGOUT
    method: POST
    path: /api/v1/auth/logout
    auth: { mode: session }

db:
  writes:
    - table: users
      columns_declared: [id, google_sub, email, name, picture_url, created_at, updated_at]

cross_links:
  adr_refs:
    - docs/architecture/adr/002-google-signin-cookie-session.md
  sisters:
    - docs/fr/survey/active/fr-forms.md
```
