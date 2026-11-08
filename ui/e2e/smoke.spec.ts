// One smoke per page: the page loads from the committed bundle, renders
// what the e2e server's doubles serve (the fixtures' data), and the
// browser logs no error on the way. Full flows stay with the vitest page
// tests and the Go end-to-end tests (concept U5 E, "Playwright nur Smoke").
import { expect, test, type Page } from "@playwright/test";

type Smoke = { name: string; hash: string; shows: (page: Page) => ReturnType<Page["getByTestId"]> };

const L = "0000000000000011";

const SMOKES: Smoke[] = [
  { name: "node status", hash: "#/node/status", shows: (p) => p.getByTestId("node-config") },
  { name: "node pair", hash: "#/node/pair", shows: (p) => p.getByTestId("ecosystem-hints") },
  { name: "node fabrics", hash: "#/node/fabrics", shows: (p) => p.getByText("Home", { exact: true }).first() },
  { name: "node selection", hash: "#/node/selection", shows: (p) => p.getByText("Kitchen lamp").first() },
  { name: "node endpoints", hash: "#/node/endpoints", shows: (p) => p.getByTestId("endpoint").first() },
  { name: "node endpoint", hash: "#/node/endpoint/2", shows: (p) => p.getByTestId("endpoint-title") },
  { name: "node surface", hash: "#/node/surface", shows: (p) => p.getByTestId("surface-head") },
  { name: "node diagnostics", hash: "#/node/diagnostics", shows: (p) => p.getByTestId("occupancy") },
  { name: "node topology", hash: "#/node/topology", shows: (p) => p.getByText("Kitchen lamp").first() },
  { name: "second node status", hash: "#/node/second/status", shows: (p) => p.getByTestId("node-status").filter({ hasText: "[::]:5541" }) },
  { name: "second node manage", hash: "#/node/second/status", shows: (p) => p.getByTestId("node-manage-hint") },
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
  // Completeness against go-fabric v0.18.0 (contract 0.7.0).
  { name: "node status attestation", hash: "#/node/status", shows: (p) => p.getByTestId("node-attestation") },
  { name: "node diagnostics reannounce", hash: "#/node/diagnostics", shows: (p) => p.getByTestId("node-reannounce") },
  { name: "node fabrics erase", hash: "#/node/fabrics", shows: (p) => p.getByTestId("node-erase") },
  { name: "controller commission terms", hash: "#/controller/commission", shows: (p) => p.getByTestId("commission-terms") },
  // The BLE decision (contract 0.7.1): the network select, with the default preselected.
  { name: "controller commission ble", hash: "#/controller/commission", shows: (p) => p.getByLabel("Network for the device") },
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

// The node registry (contract 0.6.0): create a node from the selector, find
// it listed and on its selection page, delete it with both confirmations
// from its status page and land on the default node.
test("a node is created from the selector and deleted from its status page", async ({ page }) => {
  const errors: string[] = [];
  page.on("pageerror", (err) => errors.push(err.message));
  await page.goto("/#/node/status");
  const selector = page.getByLabel("Node", { exact: true });
  await expect(selector).toBeVisible();
  const before = await selector.locator("option").count();
  await selector.selectOption({ label: "Create node…" });
  const dialog = page.getByTestId("node-create-dialog");
  await dialog.getByLabel("Name").fill("Alexa");
  await dialog.getByRole("button", { name: "Create" }).click();
  await expect(page).toHaveURL(/#\/node\/alexa\/selection$/);
  await expect(page.getByLabel("Node", { exact: true }).locator("option")).toHaveCount(before + 1);
  await expect(page.getByLabel("Node", { exact: true })).toHaveValue("alexa");

  await page.goto("/#/node/alexa/status");
  await expect(page.getByTestId("node-config-source")).toContainText("allocated");
  await page.getByTestId("node-manage-delete").click();
  await page.getByRole("alertdialog").getByRole("button", { name: "Delete node…" }).click();
  await page.getByRole("alertdialog").getByRole("button", { name: "Delete for good" }).click();
  await expect(page).toHaveURL(/#\/node\/status$/);
  await expect(page.getByLabel("Node", { exact: true }).locator("option")).toHaveCount(before);
  expect(errors).toEqual([]);
});

// Re-announce from the diagnostics page (contract 0.7.0): the admin route
// answers and the page says so.
test("the node re-announces from the diagnostics page", async ({ page }) => {
  const errors: string[] = [];
  page.on("pageerror", (err) => errors.push(err.message));
  await page.goto("/#/node/diagnostics");
  await page.getByTestId("node-reannounce").click();
  await expect(page.getByText("The node announced its records again.")).toBeVisible();
  expect(errors).toEqual([]);
});

// A commissioning with a Terms and Conditions acknowledgement on the memory
// controller (contract 0.7.0): the steps it reports appear under the
// progress.
test("a commissioning with terms lists its steps", async ({ page }) => {
  const errors: string[] = [];
  page.on("pageerror", (err) => errors.push(err.message));
  await page.goto("/#/controller/commission");
  await page.getByLabel("Pairing code").fill("34970112332");
  await page.getByTestId("commission-terms").locator("summary").click();
  await page.getByTestId("commission-terms-version").fill("1");
  await page.getByTestId("commission-terms-acknowledgements").fill("1");
  await page.getByRole("button", { name: "Commission", exact: true }).click();
  await expect(page.getByTestId("commission-steps")).toContainText("Subscribe");
  await expect(page.getByText("Commissioned", { exact: true })).toBeVisible();
  expect(errors).toEqual([]);
});

// The BLE decision (contract 0.7.1): on a host with a central the default
// credential is preselected, so no hint shows, and a commissioning
// without a choice uses it — the progress says Bluetooth LE is used. The
// memory controller commissions its one device once (the terms flow
// above did), so the result is not asserted here: the decision comes
// before it.
test("a commissioning without a choice uses the preselected network", async ({ page }) => {
  const errors: string[] = [];
  page.on("pageerror", (err) => errors.push(err.message));
  await page.goto("/#/controller/commission");
  await expect(page.getByLabel("Network for the device")).toHaveValue("thread:mesh");
  await expect(page.getByTestId("commission-ble-hint")).toHaveCount(0);
  await page.getByLabel("Pairing code").fill("34970112332");
  await page.getByRole("button", { name: "Commission", exact: true }).click();
  await expect(page.getByTestId("commission-ble")).toHaveText("Bluetooth LE: used");
  expect(errors).toEqual([]);
});

// The erase (contract 0.7.0), last: it empties the default node's fabrics.
// Two confirmations, then the toast and the notice while the status is
// polled.
test("the node is erased after two confirmations", async ({ page }) => {
  const errors: string[] = [];
  page.on("pageerror", (err) => errors.push(err.message));
  await page.goto("/#/node/fabrics");
  await page.getByTestId("node-erase").click();
  await page.getByRole("alertdialog").getByRole("button", { name: "Erase node…" }).click();
  await page.getByRole("alertdialog").getByRole("button", { name: "Erase for good" }).click();
  await expect(page.getByText(/erased\.$/)).toBeVisible();
  await expect(page.getByTestId("node-erased")).toBeVisible();
  await expect(page.getByText("Home", { exact: true })).toHaveCount(0, { timeout: 10_000 });
  expect(errors).toEqual([]);
});
