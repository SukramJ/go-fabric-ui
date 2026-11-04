// One smoke per page: the page loads from the committed bundle, renders
// what the e2e server's doubles serve (the fixtures' data), and the
// browser logs no error on the way. Full flows stay with the vitest page
// tests and the Go end-to-end tests (concept U5 E, "Playwright nur Smoke").
import { expect, test, type Page } from "@playwright/test";

type Smoke = { name: string; hash: string; shows: (page: Page) => ReturnType<Page["getByTestId"]> };

const L = "0000000000000011";

const SMOKES: Smoke[] = [
  { name: "node status", hash: "#/node/status", shows: (p) => p.getByTestId("node-status") },
  { name: "node pair", hash: "#/node/pair", shows: (p) => p.getByTestId("ecosystem-hints") },
  { name: "node fabrics", hash: "#/node/fabrics", shows: (p) => p.getByText("Home", { exact: true }).first() },
  { name: "node selection", hash: "#/node/selection", shows: (p) => p.getByText("Kitchen lamp").first() },
  { name: "node endpoints", hash: "#/node/endpoints", shows: (p) => p.getByTestId("endpoint").first() },
  { name: "node endpoint", hash: "#/node/endpoint/2", shows: (p) => p.getByTestId("endpoint-title") },
  { name: "node surface", hash: "#/node/surface", shows: (p) => p.getByTestId("surface-head") },
  { name: "node diagnostics", hash: "#/node/diagnostics", shows: (p) => p.getByTestId("occupancy") },
  { name: "node topology", hash: "#/node/topology", shows: (p) => p.getByText("Kitchen lamp").first() },
  { name: "second node status", hash: "#/node/second/status", shows: (p) => p.getByTestId("node-status").filter({ hasText: "[::]:5541" }) },
  { name: "second node selection", hash: "#/node/second/selection", shows: (p) => p.getByTestId("exposure-reason") },
  { name: "controller nodes", hash: "#/controller/nodes", shows: (p) => p.getByText("Kitchen").first() },
  { name: "controller network", hash: "#/controller/nodes?tab=network", shows: (p) => p.getByRole("button", { name: "Peer node Light, OK" }) },
  { name: "controller commission", hash: "#/controller/commission", shows: (p) => p.getByRole("button", { name: "Scan QR code" }) },
  { name: "controller settings", hash: "#/controller/settings", shows: (p) => p.getByTestId("controller-info") },
  ...["overview", "endpoints", "inventory", "events", "attributes", "acl", "bindings", "window", "icd", "fabrics", "updates"].map((tab) => ({
    name: `controller node ${tab}`,
    hash: `#/controller/node/${L}?tab=${tab}`,
    shows: (p: Page) => p.getByTestId("node-title"),
  })),
  { name: "operation", hash: "#/operation", shows: (p) => p.getByTestId("log-record").first() },
];

for (const s of SMOKES) {
  test(s.name, async ({ page }) => {
    const errors: string[] = [];
    page.on("console", (msg) => {
      if (msg.type() === "error") errors.push(msg.text());
    });
    page.on("pageerror", (err) => errors.push(err.message));
    await page.goto("/" + s.hash);
    await expect(s.shows(page)).toBeVisible();
    await page.waitForLoadState("networkidle");
    expect(errors).toEqual([]);
  });
}

test("the node selector lists both nodes and switches to the second", async ({ page }) => {
  await page.goto("/#/node/status");
  const selector = page.getByLabel("Node", { exact: true });
  await expect(selector).toBeVisible();
  await selector.selectOption("second");
  await expect(page).toHaveURL(/#\/node\/second\/status$/);
  await expect(page.getByTestId("node-status")).toContainText("[::]:5541");
});
