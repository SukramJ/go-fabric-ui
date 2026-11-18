// @vitest-environment happy-dom
// The pairing flow against the window the status names (review phase 2):
// the page's own fresh window (UR-P2-ui-05), an erase under an open flow
// (UR-P2-conc-04), the codes after the flow (UR-P2-leak-08).
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { body, fixtureFetch } from "../test/support";
import { createClient, type Status } from "./api/client";
import { createNodeStore } from "./store.svelte";

beforeEach(() => vi.useFakeTimers());
afterEach(() => vi.useRealTimers());

const t = (k: string) => k;
const toast = { info: vi.fn(), success: vi.fn(), error: vi.fn() };

const failed = () => new Response(JSON.stringify({ type: "about:blank", title: "x", status: 500, code: "internal" }), { status: 500 });
const configured = () =>
  new Response(JSON.stringify({ ...body<Record<string, unknown>>("node-pairing.200"), source: "configured", window_open: true }), { status: 200 });

// A node store over a status the test sets and a pairing answer it picks;
// sent records every request.
function harness(opts: { pairing: () => Response; status?: () => Response | null }) {
  let st: Status = { ...body<Status>("node-status.200"), commissioning_window_open: false, window_kind: "none" };
  const ff = fixtureFetch(["node-pairing-window.200", "node-fabrics.200"]);
  const sent: string[] = [];
  const client = createClient("/api/v1", async (url, init) => {
    sent.push(`${init?.method ?? "GET"} ${url.replace("/api/v1", "")}`);
    if (url.endsWith("/node/status")) return opts.status?.() ?? new Response(JSON.stringify(st), { status: 200 });
    if (url.endsWith("/node/pairing")) return opts.pairing();
    return ff.fetch(url, init);
  });
  const store = createNodeStore(client, { t, toast, subscribe: () => () => {} });
  return {
    store,
    reads: () => sent.filter((s) => s === "GET /node/pairing").length,
    set: (patch: Partial<Status>) => (st = { ...st, ...patch }),
  };
}

const fresh = body<{ qr_code: string; manual_code: string }>("node-pairing-window.200");
const basic = (seconds = 300): Partial<Status> => ({
  commissioning_window_open: true,
  window_kind: "basic",
  window_deadline: new Date(Date.now() + seconds * 1000).toISOString(),
  commissioning_window_duration_seconds: seconds,
});

describe("UR-P2-ui-05: the page's own window on fresh credentials", () => {
  // The handler's default: POST /node/pairing/window opens a basic window
  // and the status names it window_kind basic; its window_opened follows
  // on the stream.
  it("its own window_opened keeps the codes the POST answered and reads none", async () => {
    const h = harness({ pairing: failed });
    h.set(basic());
    await h.store.openWindow(300);
    h.store.applyEvent({ type: "window_opened", data: { duration_seconds: 300 } });
    await vi.advanceTimersByTimeAsync(50);
    expect({ phase: h.store.commissioning.phase, qr: h.store.commissioning.window?.qr_code, reads: h.reads() }).toEqual({
      phase: "open",
      qr: fresh.qr_code,
      reads: 0,
    });
  });

  it("a resync keeps the codes of its own window and reads none", async () => {
    const h = harness({ pairing: failed });
    h.set(basic());
    await h.store.openWindow(300);
    await vi.advanceTimersByTimeAsync(10_000);
    h.store.resync();
    await vi.advanceTimersByTimeAsync(50);
    expect({ phase: h.store.commissioning.phase, qr: h.store.commissioning.window?.qr_code, reads: h.reads() }).toEqual({
      phase: "open",
      qr: fresh.qr_code,
      reads: 0,
    });
  });

  it("a window opened while the status was not read is taken as its own at the next status", async () => {
    let down = true;
    const h = harness({ pairing: failed, status: () => (down ? new Response("{}", { status: 503 }) : null) });
    h.set(basic());
    await h.store.openWindow(300);
    down = false;
    h.store.applyEvent({ type: "window_opened", data: { duration_seconds: 300 } });
    await vi.advanceTimersByTimeAsync(50);
    expect({ phase: h.store.commissioning.phase, qr: h.store.commissioning.window?.qr_code, reads: h.reads() }).toEqual({
      phase: "open",
      qr: fresh.qr_code,
      reads: 0,
    });
  });

  it("a later window of another kind still replaces the codes", async () => {
    const h = harness({ pairing: configured });
    h.set(basic());
    await h.store.openWindow(300);
    h.set({ window_kind: "enhanced", window_deadline: new Date(Date.now() + 900_000).toISOString() });
    h.store.resync();
    await vi.advanceTimersByTimeAsync(50);
    expect(h.reads()).toBe(1);
    expect(h.store.commissioning.window?.qr_code).not.toBe(fresh.qr_code);
  });
});
