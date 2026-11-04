// The browser smoke tests (make e2e-web): Chromium only, one smoke per
// page against the Go server of fabricui's TestServeE2EWeb — the committed
// bundle and the API on the handler tests' doubles. make e2e-web builds the
// test binary and names it in FABRIC_UI_E2E_BIN; without it the server is
// started with go test.
import { defineConfig, devices } from "@playwright/test";

const port = Number(process.env.FABRIC_UI_E2E_PORT ?? 18089);
const addr = `127.0.0.1:${port}`;
const bin = process.env.FABRIC_UI_E2E_BIN;

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  workers: 1,
  forbidOnly: !!process.env.CI,
  retries: 0,
  reporter: process.env.CI ? [["list"], ["github"]] : "list",
  use: { baseURL: `http://${addr}`, locale: "en-US", trace: "off" },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
  webServer: {
    command: bin ? `"${bin}" -test.run "^TestServeE2EWeb$" -test.timeout 0` : `go test -count=1 -run "^TestServeE2EWeb$" -timeout 0 ./fabricui`,
    cwd: "..",
    env: { FABRIC_UI_E2E_ADDR: addr, CGO_ENABLED: "0" },
    url: `http://${addr}/api/v1/node/status`,
    reuseExistingServer: false,
    timeout: 180_000,
  },
});
