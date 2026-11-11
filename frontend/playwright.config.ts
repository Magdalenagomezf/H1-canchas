import { defineConfig, devices } from '@playwright/test';

// The app must already be running. Point the tests at it with:
//   E2E_BASE_URL  frontend URL (default http://localhost:5173)
//   E2E_API_URL   backend URL  (default http://localhost:8080)
//   E2E_ADMIN_PHONE / E2E_ADMIN_PASSWORD  an admin account, used to create and clean test data
// CI starts a throwaway stack (see .github/workflows/ci.yml); later the same
// tests run against QA by changing these variables.
export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  fullyParallel: false,
  workers: 1,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [['list'], ['html', { open: 'never' }]] : 'list',
  use: {
    baseURL: process.env.E2E_BASE_URL ?? 'http://localhost:5173',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
  projects: [
    // API integration tests (e2e/api.spec.ts): no browser, run first with `npm run test:api`.
    { name: 'api', testMatch: 'api.spec.ts' },
    // Browser E2E tests: `npm run e2e`.
    { name: 'chromium', testIgnore: 'api.spec.ts', use: { ...devices['Desktop Chrome'] } },
  ],
});
