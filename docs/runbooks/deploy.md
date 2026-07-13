# Runbook — deploy to production

txsurvey has **no CD**: pushing to `main` + green CI does **not** deploy. The live
service is the `txsurvey-app-1` Docker container on this VPS (image
`txsurvey-app:prod`, SPA embedded), behind nginx + Cloudflare at
**https://brainzap.net/txsurvey** and **https://ct.tuxceria.biz.id/txsurvey**.

Deploy is a manual, ordered sequence. Run it from the repo root on the prod host.

## Steps

1. **Gate — land the change green first.**
   ```bash
   make check          # lint + unit tests + coverage + routes + docs + FE tests/build
   ```
   Don't build an image from a red tree.

2. **Build + recreate the container** (rebuilds the embedded-SPA image, brief downtime):
   ```bash
   docker compose -f docker-compose.prod.yml -p txsurvey up -d --build
   ```
   Postgres (`txsurvey-postgres-1`, not in the prod compose) and uploads
   (`./data/uploads` host volume) persist. Migrations run idempotently at boot.
   The `orphan container [txsurvey-postgres-1]` warning is expected — ignore it.

3. **Verify the deploy actually came up:**
   ```bash
   make verify-deploy   # container up -> local /health -> public edge (see verify-deploy.sh)
   ```
   Or manually:
   ```bash
   curl -s -o /dev/null -w "%{http_code}\n" http://172.17.0.1:8092/health          # local app
   curl -s -o /dev/null -w "%{http_code}\n" -L https://brainzap.net/txsurvey/       # public edge
   ```
   Both `200` → done. Anything else → see [recovery.md](recovery.md).

## Gotchas

- **Secrets** come from `deploy/production.env` (gitignored — never commit it).
- **Google OAuth** redirect URIs / test users are Console-only; they can't be changed
  from code.
- `VITE_BASE=/txsurvey/` (in `frontend/.env.production`) bakes the subpath into the SPA;
  nginx strips `/txsurvey/` before proxying, so the Go app serves at root.
- To publish a docs/asset-only change (e.g. the assessment page/infographic), you still
  need a full image rebuild — the SPA is embedded at build time, not served from disk.
