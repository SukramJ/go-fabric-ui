// @vitest-environment happy-dom
// The stores' life cycle and ordering (review phase 1b).
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { body, fixtureFetch, testUI } from "../test/support";
import { createClient, type LogRecord, type Status } from "./api/client";
import { createNodeStore } from "./store.svelte";
import { createControllerStore } from "./controller-store.svelte";
import { createFabricUI } from "./context";

beforeEach(() => vi.useFakeTimers());
afterEach(() => vi.useRealTimers());

const t = (k: string) => k;
const toast = { info: vi.fn(), success: vi.fn(), error: vi.fn() };
const paths = (sent: { method: string; path: string }[]) => sent.map((s) => `${s.method} ${s.path.replace("/api/v1", "")}`);

describe("UR-P1-conc-102: a store reloads after a gap in its stream", () => {
  it("node store: the stream's resync reloads status and fabrics, and moves the epochs", async () => {
    const ff = fixtureFetch(["node-status.200", "node-fabrics.200"]);
    let resync: (() => void) | undefined;
    const store = createNodeStore(createClient("/api/v1", ff.fetch), {
      t,
      toast,
      subscribe: (_fn, onResync) => {
        resync = onResync;
        return () => {};
      },
    });
    store.connect();
    expect(ff.sent).toHaveLength(0);
    const epoch = store.graphEpoch;
    resync?.();
    await vi.runOnlyPendingTimersAsync();
    expect(paths(ff.sent)).toEqual(expect.arrayContaining(["GET /node/status", "GET /node/fabrics"]));
    expect(store.graphEpoch).toBe(epoch + 1);
  });

  it("node store: connecting again after disconnect reloads", async () => {
    const ff = fixtureFetch(["node-status.200", "node-fabrics.200"]);
    const store = createNodeStore(createClient("/api/v1", ff.fetch), { t, toast, subscribe: () => () => {} });
    store.connect();
    store.disconnect();
    expect(ff.sent).toHaveLength(0);
    store.connect();
    await vi.runOnlyPendingTimersAsync();
    expect(paths(ff.sent)).toContain("GET /node/status");
  });

  it("node store: an open flow whose window closed during the gap ends", async () => {
    const ui = testUI({
      fixtures: ["node-pairing-window.200", "node-fabrics.200"],
      answers: { "GET /node/status": { status: 200, body: { ...body<Status>("node-status.200"), commissioning_window_open: false } } },
    });
    await ui.node.openWindow(300);
    ui.node.resync();
    await vi.advanceTimersByTimeAsync(10);
    expect(ui.node.commissioning.phase).toBe("idle");
  });

  it("controller store: the stream's resync reloads status and nodes", async () => {
    const ff = fixtureFetch(["controller-status.200", "controller-nodes.200"]);
    let resync: (() => void) | undefined;
    const store = createControllerStore(createClient("/api/v1", ff.fetch), {
      t,
      toast,
      subscribe: (_fn, onResync) => {
        resync = onResync;
        return () => {};
      },
    });
    store.connect();
    const rev = store.revision;
    resync?.();
    await vi.runOnlyPendingTimersAsync();
    expect(paths(ff.sent)).toEqual(expect.arrayContaining(["GET /controller/status", "GET /controller/nodes"]));
    expect(store.revision).toBe(rev + 1);
  });

  it("the default log subscription reads the records logged during a gap", async () => {
    const sources: { listeners: Map<string, ((ev: MessageEvent) => void)[]>; readyState: number }[] = [];
    vi.stubGlobal(
      "EventSource",
      vi.fn(function () {
        const src = {
          listeners: new Map<string, ((ev: MessageEvent) => void)[]>(),
          readyState: 0,
          addEventListener(type: string, fn: (ev: MessageEvent) => void) {
            this.listeners.set(type, [...(this.listeners.get(type) ?? []), fn]);
          },
          close() {},
        };
        sources.push(src);
        return src;
      }),
    );
    const rec = (seq: number): LogRecord => ({ seq, time: "2026-11-18T12:00:00Z", level: "info", msg: `m${seq}`, attrs: {} });
    const seen: string[] = [];
    const client = createClient("/api/v1", async (url) => {
      seen.push(url);
      return new Response(JSON.stringify({ records: [rec(6), rec(7)] }), { status: 200 });
    });
    const ui = createFabricUI({ client });
    const got: number[] = [];
    const close = ui.subscribeLog((r) => got.push(r.seq));
    const emit = (type: string, data = "{}") => sources[0].listeners.get(type)?.forEach((fn) => fn({ data } as MessageEvent));
    emit("open");
    emit("log_record", JSON.stringify(rec(5)));
    sources[0].readyState = 0;
    emit("error");
    emit("open");
    await vi.advanceTimersByTimeAsync(10);
    expect(seen).toEqual(["/api/v1/operation/log?since=6"]);
    expect(got).toEqual([5, 6, 7]);
    close();
    vi.unstubAllGlobals();
  });
});
