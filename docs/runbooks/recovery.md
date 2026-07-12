# Runbook — failure recovery

Symptom → cause → fix. Prefer the smallest non-destructive fix; never drop a shared
database to "start clean".

| Symptom | Cause | Fix |
|---|---|---|
| `go test ./tests/` fails at migration bootstrap: `no migration found for version N: file does not exist` | The shared `*_test` DB's `schema_migrations.version` drifted **ahead** of the committed migration files (e.g. a throwaway branch applied a migration to the shared `txsurvey_test`, then the file was reverted). | **Non-destructive** — reset the pointer to the highest committed migration; do **not** drop the shared DB: `docker exec txsurvey-postgres-1 psql -U txsurvey -d txsurvey_test -c "UPDATE schema_migrations SET version=<max_committed>, dirty=false;"`. Leftover additive enum values are harmless. |
| Two new migrations grabbed the **same number** (e.g. two `007_*.up.sql`) — golang-migrate errors on duplicate version at boot. | Concurrent work each took "the next number." | Renumber the **later** one to the next free number (migrations are forward-only; numeric order = build order). A gap is safe; a collision is not. |
| Integration tests refuse to run: `refusing to run TRUNCATE-ing tests against non-test database` | `DATABASE_URL` points at a db whose name lacks `test`. | Point it at a `*_test` DB (e.g. `…/txsurvey_test`). The dev DB `txsurvey` is intentionally rejected — the harness TRUNCATEs. |
| Server won't start locally: address `:8080` already in use. | Port 8080 is occupied by an environment proxy on this host. | Set a free port: `SERVER_PORT=8181 make run`. Tests use `httptest` (no port), so this only affects a manually-run server. |
| `PATCH /forms/:id {slug}` → 422 `SLUG_LOCKED`. | The form is **published**; its slug is frozen so shared `/r/:slug` links keep working. | Expected. Slugs are only editable while a form is a draft. To change a published URL you'd unpublish first (breaks existing links). |
| `PATCH /forms/:id {slug}` → 422 `SLUG_TAKEN`. | Another **live** form already uses that slug (partial-unique index on live rows). | Choose a different slug. |
| Deploy ran but the site is down / stale. | Container didn't come up, nginx cached a stale upstream inode, or the edge is misrouted. | `make verify-deploy` — each red line names the culprit. Check `docker ps` / `docker compose -f docker-compose.prod.yml -p txsurvey logs app`; reload nginx if the container is healthy but the edge 502s. |
| Frontend `npm run test` / `make check` fails with "cannot find module vitest". | Frontend deps not installed in this checkout. | `npm --prefix frontend ci` (or `install`) first; `make check` assumes deps are present. |

## Migration discipline reminder

Migrations are **forward-only** and embedded via `//go:embed`; a new migration must take
the next number. Never edit an already-applied migration in prod. In dev, if you must
re-apply, reset `schema_migrations` as above rather than dropping data. See the dual
logic-engine note in `CLAUDE.md` before touching either engine — parity is now gated
(`make check` runs the shared-fixture parity test on both sides).
