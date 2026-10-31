// @vitest-environment happy-dom
// The node store's reaction to the stream, ported from the event handling
// of openccu-loom assets/ui/src/lib/stores/matter.svelte.ts:92-160 at b40b289a.
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { testUI } from "../test/support";
import { createNodeStore } from "./store.svelte";

beforeEach(() => vi.useFakeTimers());
afterEach(() => vi.useRealTimers());

const paths = (ui: ReturnType<typeof testUI>) => ui.sent.map((s) => `${s.method} ${s.path.replace("/api/v1", "")}`);

describe("node store", () => {
  it("reloads fabrics and status on fabric_added and ends the pairing flow in success", async () => {
    const ui = testUI({ fixtures: ["node-pairing-window.200", "node-fabrics.200", "node-status.200"] });
    await ui.node.openWindow(300);
    expect(ui.node.commissioning.phase).toBe("open");
    ui.node.applyEvent({ type: "fabric_added", data: { fabric_index: 3, vendor_name: "Apple" } });
    expect(ui.node.commissioning).toMatchObject({ phase: "success", addedFabricLabel: "Apple" });
    await vi.runOnlyPendingTimersAsync();
    expect(paths(ui)).toEqual(expect.arrayContaining(["GET /node/fabrics", "GET /node/status"]));
  });

  it("ends an open flow on a close no fabric follows, with a toast", async () => {
    const ui = testUI({ fixtures: ["node-pairing-window.200", "node-status.200"] });
    await ui.node.openWindow(300);
    ui.node.applyEvent({ type: "window_closed", data: {} });
    expect(ui.node.commissioning.phase).toBe("open");
    await vi.advanceTimersByTimeAsync(800);
    expect(ui.node.commissioning.phase).toBe("idle");
    expect(ui.toastSpy.info).toHaveBeenCalledWith("node.commissioning.closed");
  });

  it("keeps the success when the window closes just before the fabric is reported", async () => {
    const ui = testUI({ fixtures: ["node-pairing-window.200", "node-status.200", "node-fabrics.200"] });
    await ui.node.openWindow(300);
    ui.node.applyEvent({ type: "window_closed", data: {} });
    ui.node.applyEvent({ type: "fabric_added", data: { fabric_index: 1, label: "Home" } });
    await vi.advanceTimersByTimeAsync(800);
    expect(ui.node.commissioning.phase).toBe("success");
    expect(ui.toastSpy.info).not.toHaveBeenCalled();
  });

  it("moves the topology epoch on endpoints_changed and surface_changed and reloads the status", async () => {
    const ui = testUI({ fixtures: ["node-status.200"] });
    expect(ui.node.topologyEpoch).toBe(0);
    ui.node.applyEvent({ type: "endpoints_changed", data: {} });
    ui.node.applyEvent({ type: "surface_changed", data: { endpoints: ["lamp-1"] } });
    ui.node.applyEvent({ type: "window_opened", data: {} });
    expect(ui.node.topologyEpoch).toBe(2);
    await vi.runOnlyPendingTimersAsync();
    expect(paths(ui).filter((p) => p === "GET /node/status")).toHaveLength(3);
  });

  it("collapses a burst of exposure_changed into one reload", async () => {
    const ui = testUI({ fixtures: ["node-exposure-schema.200", "node-exposure.200"] });
    for (let i = 0; i < 5; i++) ui.node.applyEvent({ type: "exposure_changed", data: { keys: [`k${i}`] } });
    await vi.advanceTimersByTimeAsync(400);
    expect(paths(ui).filter((p) => p === "GET /node/exposure")).toHaveLength(1);
  });

  it("counts the window down and returns to idle when it runs out", async () => {
    const ui = testUI({ fixtures: ["node-pairing-window.200"] });
    await ui.node.openWindow(300);
    await vi.advanceTimersByTimeAsync(2000);
    expect(ui.node.commissioning.remaining).toBe(298);
    await vi.advanceTimersByTimeAsync(300_000);
    expect(ui.node.commissioning.phase).toBe("idle");
  });

  it("hydrates an open window with what is left until its deadline", async () => {
    vi.setSystemTime(new Date("2026-10-25T12:02:00Z"));
    const ui = testUI({ fixtures: ["node-status.200", "node-pairing.200"] });
    await ui.node.hydrateCommissioning();
    expect(ui.node.commissioning).toMatchObject({ phase: "open", remaining: 180 });
    await vi.advanceTimersByTimeAsync(3000);
    expect(ui.node.commissioning.remaining).toBe(177);
  });

  it("drops a pending change that returns to the item's state", () => {
    const ui = testUI();
    const item = { key: "a", name: "A", group: "", fields: {}, enabled: false, source: "", verdict: { state: "mappable" as const, device_type: 0, device_type_name: "", clusters: [], reason: "" } };
    ui.node.markDirty(item, { enabled: true });
    expect(ui.node.hasDirty).toBe(true);
    ui.node.markDirty(item, { enabled: false });
    expect(ui.node.hasDirty).toBe(false);
  });

  it("subscribes once and closes the stream", () => {
    const close = vi.fn();
    const subscribe = vi.fn(() => close);
    const ui = testUI();
    const store = createNodeStore(ui.client, { t: ui.t, toast: ui.toast, subscribe });
    store.connect();
    store.connect();
    expect(subscribe).toHaveBeenCalledOnce();
    store.close();
    expect(close).toHaveBeenCalledOnce();
  });
});
