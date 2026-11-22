// @vitest-environment happy-dom
// The status has three readers (loadStatus, hydrateCommissioning, openWindow):
// an answer one of them overtook must not leave the store loading, without
// a status, or on a stale stream (UR-P4-conc-230).
import { describe, expect, it } from "vitest";
import { createClient, TOKEN_HEADER } from "./api/client";
import { createNodeStore, type NodeStoreDeps } from "./store.svelte";

const json = (v: unknown, status = 200) => new Response(JSON.stringify(v), { status, headers: { "Content-Type": "application/json" } });
const sse = () => new Response(new ReadableStream<Uint8Array>(), { status: 200, headers: { "Content-Type": "text/event-stream" } });
const flush = () => new Promise((r) => setTimeout(r, 20));
const deps: NodeStoreDeps = { t: (k) => k, toast: { success: () => undefined, error: () => undefined, info: () => undefined } };

/** A client whose status answers the test releases one by one. */
function held() {
  const pending: { url: string; answer: (r: Response) => void }[] = [];
  const client = createClient("/api/v1", (url) => new Promise<Response>((res) => pending.push({ url, answer: res })));
  return { client, pending };
}
const ST = { enabled: true, commissioning_window_open: false, fabric_count: 0, endpoint_count: 1 };

describe("UR-P4-conc-230: a hydrate overtakes a loadStatus in flight", () => {
  it("statusLoading ends when both have answered", async () => {
    const { client, pending } = held();
    const store = createNodeStore(client, { ...deps, subscribe: () => () => {} });
    const a = store.loadStatus(); // the shell's probe (App.svelte:198)
    const h = store.hydrateCommissioning(); // the Pair page's mount
    pending[1].answer(json(ST));
    await h;
    pending[0].answer(json(ST));
    await a;
    expect(store.status).not.toBeNull();
    expect(store.statusLoading).toBe(false);
  });

  it("a failed hydrate read does not leave the store without a status and loading for good", async () => {
    const { client, pending } = held();
    const store = createNodeStore(client, { ...deps, subscribe: () => () => {} });
    const a = store.loadStatus();
    const h = store.hydrateCommissioning().catch(() => undefined);
    pending[1].answer(new Response("<html>502</html>", { status: 502 }));
    await h;
    pending[0].answer(json(ST));
    await a;
    // Selection.svelte:57 waits while status === null && statusLoading; Fabrics.svelte:37 loads only when !statusLoading.
    expect({ status: store.status !== null, loading: store.statusLoading, error: store.statusError }).toEqual({ status: true, loading: false, error: null });
  });

  it("the node stream is reopened when the overtaken status brought a new token", async () => {
    let token = "t1";
    const streams: string[] = [];
    const gate: (() => void)[] = [];
    let hold = false;
    const client = createClient("/api/v1", async (url: string, init?: RequestInit) => {
      const sent = ((init?.headers ?? {}) as Record<string, string>)[TOKEN_HEADER] ?? "-";
      if (url.endsWith("/stream")) {
        streams.push(sent);
        return sse();
      }
      if (url.endsWith("/status")) {
        if (hold) await new Promise<void>((r) => gate.push(r));
        return json({ ...ST, csrf_token: token });
      }
      return json({ code: "not_found", title: "x" }, 404);
    });
    const store = createNodeStore(client, deps);
    await store.loadStatus();
    store.connect();
    await flush();
    expect(streams).toEqual(["t1"]);
    // The daemon restarts: a new token. A status reload and a hydrate (a
    // window_opened on an idle flow, or the Pair page's mount) overlap.
    token = "t2";
    hold = true;
    const a = store.loadStatus();
    const h = store.hydrateCommissioning();
    hold = false;
    for (let i = 0; i < 6; i++) {
      gate.splice(0).forEach((r) => r());
      await flush();
    }
    await Promise.all([a, h]);
    await flush();
    expect(client.currentToken?.()).toBe("t2");
    expect(streams).toEqual(["t1", "t2"]);
    store.close();
  });
});
