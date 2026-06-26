import { test, expect } from "@playwright/test";

// t0 regression baseline: the core creator→respondent happy path.
// Covers the bug where "+ Tambah pertanyaan" silently failed (empty title 422):
// if adding a question regresses, step 4 fails.
//
// Auth uses the dev-only /auth/dev-login route (no Google), so the server under
// test MUST run with APP_ENV != production. See playwright.config.ts for how to
// start the server and run this.

const API = "/api/v1";
const CREATOR = { email: "e2e@txsurvey.test", name: "E2E Creator" };

test.describe("create-survey regression (t0)", () => {
  test("dev-login → create → add question → publish → public submit", async ({ page }) => {
    // 1. Authenticate without Google (sets the session cookie on this context).
    const login = await page.request.post(`${API}/auth/dev-login`, { data: CREATOR });
    expect(login.ok(), "dev-login must succeed (run server with APP_ENV=development)").toBeTruthy();

    // 2. Dashboard.
    await page.goto("/");
    await expect(page.getByRole("heading", { name: "Surveimu" })).toBeVisible();

    // 3. Create a survey.
    const title = `E2E Smoke ${Date.now()}`;
    await page.getByRole("button", { name: /Survei baru/ }).first().click();
    await page.getByPlaceholder(/Feedback pelanggan/).fill(title);
    await page.getByRole("button", { name: /Buat survei/ }).click();
    await expect(page).toHaveURL(/\/forms\/[0-9a-f-]+$/);

    // 4. Add a question — the regression guard.
    await page.getByRole("button", { name: /Tambah pertanyaan/ }).click();
    await page.getByRole("menuitem", { name: "Teks singkat" }).click();

    // A long title to exercise wrapping / the mobile no-overflow guard below.
    const qTitle = "Seberapa puas kamu dengan proses onboarding dan dukungan tim selama minggu pertama bekerja di sini?";
    const titleField = page.getByPlaceholder("Tulis pertanyaan…");
    await expect(titleField).toBeVisible();
    await titleField.fill(qTitle);
    await page.getByRole("button", { name: "Simpan" }).click();
    await expect(page.getByText(qTitle)).toBeVisible(); // appears in the question list

    // Regression guard: on a phone viewport the builder must wrap (not scroll
    // horizontally) even with a long question title.
    await page.setViewportSize({ width: 390, height: 844 });
    const overflow = await page.evaluate(
      () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
    );
    expect(overflow, "builder must not scroll horizontally on mobile").toBeLessThanOrEqual(1);
    await page.setViewportSize({ width: 1280, height: 800 });

    // 5. Publish.
    await page.getByRole("button", { name: "Terbitkan" }).click();
    await expect(page.getByText("published")).toBeVisible();

    // 6. Resolve the public slug via the API, then run the survey as a respondent.
    const formId = page.url().split("/forms/")[1];
    const detail = await page.request.get(`${API}/forms/${formId}`);
    const slug = (await detail.json()).data.slug as string;
    expect(slug, "published form should expose a slug").toBeTruthy();

    await page.goto(`/r/${slug}`);
    await page.getByRole("button", { name: /Mulai/ }).click();
    await expect(page.getByRole("heading", { name: qTitle })).toBeVisible();
    await page.getByPlaceholder("Tulis jawabanmu…").fill("Baik sekali");
    await page.getByRole("button", { name: /Kirim/ }).click();

    // 7. Thank-you screen.
    await expect(page.getByRole("heading", { name: /Makasih/ })).toBeVisible();
  });
});
