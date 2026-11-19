// @vitest-environment happy-dom
// Six rules of the client that no other test held (UR-P4-test-322): each
// case fails when its rule is dropped.
import { afterEach, describe, expect, it, vi } from "vitest";
import { createClient, type FabricClient } from "./api/client";
import { fetchEventSource } from "./api/stream";
import { createNodeStore } from "./store.svelte";
import { createControllerStore } from "./controller-store.svelte";

afterEach(() => {
  vi.useRealTimers();
});

const t = (k: string) => k;
const toast = { info: vi.fn(), success: vi.fn(), error: vi.fn() };
const json = (v: unknown, status = 200) => new Response(JSON.stringify(v), { status, headers: { "Content-Type": "application/json" } });
const forbidden = (code = "forbidden") => json({ code, title: "Refused", status: 403, type: "x" }, 403);
const flush = () => new Promise((r) => setTimeout(r, 0));

describe("UR-P4-test-322: the client's rules", () => {
  // Rule 1: cookies go to the page's own origin only.
  it("a call is sent with credentials same-origin", async () => {
    const seen: (RequestCredentials | undefined)[] = [];
    const client = createClient("/api/v1", async (_u, init) => {
      seen.push(init?.credentials);
      return json({ fabrics: [] });
    });
    await client.fabrics();
    await client.closeWindow().catch(() => undefined);
    expect(seen).toEqual(["same-origin", "same-origin"]);
  });

  // Rule 3: a call is repeated once at most, even when every reload brings
  // another token (the bound UR-P4-err-231's fix leans on).
  it("a call refused 403 forbidden is repeated once, not again on a second refusal", async () => {
    let n = 0;
    const posts: string[] = [];
    const client = createClient("/api/v1", async (url, init) => {
      // A new token on each of the first three reloads; the cap keeps a
      // client without the rule from looping for good.
      if (url.endsWith("/status")) return json({ enabled: true, csrf_token: `t${Math.min(++n, 3)}` });
      posts.push(((init?.headers ?? {}) as Record<string, string>)["X-Fabric-UI-Token"] ?? "-");
      return forbidden();
    });
    client.setToken("t0");
    await expect(client.closeWindow()).rejects.toMatchObject({ status: 403 });
    expect(posts).toEqual(["t0", "t1"]);
  });

  // Rule 2: only the token's refusal (code forbidden) reloads the token.
  it("a 403 of another code is not repeated and reloads no token", async () => {
    const seen: string[] = [];
    const client = createClient("/api/v1", async (url, init) => {
      seen.push(`${init?.method ?? "GET"} ${url.replace("/api/v1", "")}`);
      if (url.endsWith("/status")) return json({ enabled: true, csrf_token: "t2" });
      return forbidden("origin_not_allowed");
    });
    client.setToken("t1");
    await expect(client.closeWindow()).rejects.toMatchObject({ status: 403 });
    expect(seen).toEqual(["POST /node/pairing/window/close"]);
  });

  // Rule 4: an answer of another media type is not a stream.
  it("a 200 stream answer that is not text/event-stream is an error, not an open", async () => {
    const src = fetchEventSource({ fetch: async () => new Response("data: x\n\n", { status: 200, headers: { "Content-Type": "text/plain" } }) })("/s");
    const seen: string[] = [];
    src.addEventListener("open", () => seen.push("open"));
    src.addEventListener("message", () => seen.push("message"));
    src.addEventListener("error", () => seen.push("error"));
    await flush();
    await flush();
    expect(seen).toEqual(["error"]);
    expect([src.readyState, src.status]).toEqual([2, 200]);
  });

  // Rule 5: the fabrics are an admin's; a resync without admin rights does not read them.
  it("the node store's resync reads no fabrics without admin rights", async () => {
    const paths: string[] = [];
    const client = createClient("/api/v1", async (url) => {
      paths.push(url.replace("/api/v1", ""));
      if (url.endsWith("/node/status")) return json({ enabled: true, commissioning_window_open: false, fabric_count: 1, endpoint_count: 1 });
      return json({ fabrics: [] });
    });
    const store = createNodeStore(client, { t, toast, subscribe: () => () => {} });
    await store.loadStatus();
    expect(store.isAdmin).toBe(false);
    paths.length = 0;
    store.resync();
    await flush();
    expect(paths).toEqual(["/node/status"]);
    store.close();
  });

  // Rule 6: the controller stream is an admin's; without admin rights none is opened.
  it("the controller store opens no stream without admin rights", async () => {
    const base = createClient("/api/v1", async (url) => {
      if (url.endsWith("/controller/status")) return json({ enabled: true, admin: false });
      return json({ nodes: [] });
    });
    const openStream = vi.fn(base.openStream);
    const client: FabricClient = { ...base, openStream };
    const store = createControllerStore(client, { t, toast });
    await store.loadStatus();
    expect(store.isAdmin).toBe(false);
    store.connect();
    await flush();
    expect(openStream).not.toHaveBeenCalled();
    store.close();
  });
});
