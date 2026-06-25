import { defineConfig, devices } from "@playwright/test";

// E2E runs against the running app. The simplest, zero-build setup uses the
// Vite dev server (root base) proxying /api to a dev-mode Go server:
//
//   # terminal 1 — API (dev mode mounts /auth/dev-login; pick a free port)
//   APP_ENV=development \
//   DATABASE_URL=postgres://txsurvey:txsurvey@localhost:5432/txsurvey_test?sslmode=disable \
//   JWT_SECRET=0123456789012345678901234567890123456789 \
//   SERVER_PORT=8083 ./bin/server      # or: go run ./cmd/server
//
//   # terminal 2 — Vite dev (proxies /api to the API above)
//   cd frontend && VITE_API_PROXY=http://localhost:8083 npm run dev
//
//   # terminal 3 — the test (one-time: npm run test:e2e:install)
//   cd frontend && npm run test:e2e
//
// Auth uses the dev-only /auth/dev-login route, so the API MUST run with
// APP_ENV != production (in prod that route is not mounted).
export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: 1,
  reporter: process.env.CI ? "list" : [["list"], ["html", { open: "never" }]],
  use: {
    baseURL: process.env.E2E_BASE_URL ?? "http://localhost:5173",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
});
