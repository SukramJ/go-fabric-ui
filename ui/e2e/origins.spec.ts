// What a page of another origin gets from the shell and its API, end to
// end in the browser (review phase 2, UR-P2-test-95): no frame of the shell
// (frame-ancestors 'self'), no status and no stream (no CORS). The other
// origin is a small server on localhost; the e2e server listens on
// 127.0.0.1, another host name and so another origin.
import { createServer, type Server } from "node:http";
import type { AddressInfo } from "node:net";
import { expect, test } from "@playwright/test";

let other: Server;
let OTHER = "";
const pages = new Map<string, string>();

test.beforeAll(async () => {
  other = createServer((req, res) => {
    res.setHeader("Content-Type", "text/html");
    res.end(pages.get(req.url ?? "") ?? "<!doctype html><title>another origin</title>");
  });
  await new Promise<void>((r) => other.listen(0, "127.0.0.1", r));
  OTHER = `http://localhost:${(other.address() as AddressInfo).port}`;
});
test.afterAll(() => other.close());

test("a page of another origin cannot frame the shell; a same-origin parent can", async ({ page, baseURL }) => {
  const shell = baseURL!;
  const consoleLines: string[] = [];
  page.on("console", (m) => consoleLines.push(m.text()));
  pages.set("/frame.html", `<!doctype html><iframe src="${shell}/#/node/fabrics" width=800 height=600></iframe><iframe src="${shell}/index.html#/controller" width=800 height=600></iframe>`);
  await page.goto(OTHER + "/frame.html");
  await expect.poll(() => consoleLines.filter((l) => l.includes("frame-ancestors")).length).toBeGreaterThanOrEqual(2);
  const frames = page.frames().filter((f) => f !== page.mainFrame());
  expect(frames).toHaveLength(2);
  for (const f of frames) expect(f.url().startsWith(shell)).toBe(false);

  // The same origin is admitted.
  const same = await page.context().newPage();
  await same.route(shell + "/parent.html", (route) => route.fulfill({ contentType: "text/html", body: `<!doctype html><iframe src="/#/node/fabrics" width=800 height=600></iframe>` }));
  await same.goto(shell + "/parent.html");
  await expect(same.frameLocator("iframe").getByText("Home", { exact: true }).first()).toBeVisible();
});

test("a page of another origin reads no status and no stream", async ({ page, baseURL }) => {
  await page.goto(OTHER + "/");
  const got = await page.evaluate(async (base) => {
    const attempt = (p: Promise<Response>) =>
      p.then(
        async (r) => `type ${r.type} status ${r.status} body ${(await r.text()).length}`,
        (e) => "blocked: " + (e as Error).name,
      );
    const out: Record<string, string> = {};
    out.read = await attempt(fetch(base + "/api/v1/node/status"));
    out.readCreds = await attempt(fetch(base + "/api/v1/node/status", { credentials: "include" }));
    out.opaque = await attempt(fetch(base + "/api/v1/node/status", { mode: "no-cors" }));
    out.stream = await new Promise<string>((resolve) => {
      const es = new EventSource(base + "/api/v1/node/stream");
      es.onopen = () => resolve("open");
      es.onerror = () => resolve("error, readyState " + es.readyState);
      setTimeout(() => resolve("timeout"), 4000);
    });
    // The JSON as a script include: nosniff, and JSON is no script.
    out.script = await new Promise<string>((resolve) => {
      const s = document.createElement("script");
      s.src = base + "/api/v1/node/status";
      s.onload = () => resolve("loaded");
      s.onerror = () => resolve("error");
      document.head.appendChild(s);
    });
    return out;
  }, baseURL!);
  expect(got.read).toMatch(/^blocked/);
  expect(got.readCreds).toMatch(/^blocked/);
  expect(got.opaque).toBe("type opaque status 0 body 0");
  expect(got.stream).toMatch(/^error/);
  expect(got.script).toBe("error");
});
