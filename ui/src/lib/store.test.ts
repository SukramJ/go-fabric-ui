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

  it("reloads the status once per burst of health_sample and moves the diagnostic epoch once per burst of diagnostic_event", async () => {
    const ui = testUI({ fixtures: ["node-status.200"] });
    const sample = { healthy: false, note: "listener closed", sampled_at: "2026-10-25T12:00:00Z" };
    const trace = { at: "2026-10-25T12:00:00Z", kind: "session" as const, severity: "info" as const, message: "m" };
    ui.node.applyEvent({ type: "health_sample", data: sample });
    ui.node.applyEvent({ type: "health_sample", data: sample });
    ui.node.applyEvent({ type: "diagnostic_event", data: trace });
    ui.node.applyEvent({ type: "diagnostic_event", data: trace });
    expect(ui.node.diagnosticEpoch).toBe(0);
    expect(paths(ui)).toEqual([]);
    await vi.advanceTimersByTimeAsync(400);
    expect(ui.node.diagnosticEpoch).toBe(1);
    expect(paths(ui)).toEqual(["GET /node/status"]);
    ui.node.applyEvent({ type: "diagnostic_event", data: trace });
    ui.node.close();
    await vi.advanceTimersByTimeAsync(400);
    expect(ui.node.diagnosticEpoch).toBe(1);
  });

  it("moves an endpoint's attribute epoch once per burst of its attribute_changed", async () => {
    const ui = testUI({});
    const change = (endpoint: number) => ui.node.applyEvent({ type: "attribute_changed", data: { endpoint, cluster_hex: "0x0006", attributes: ["0x0000"] } });
    change(2);
    change(2);
    change(3);
    expect(ui.node.attributeEpoch(2)).toBe(0);
    await vi.advanceTimersByTimeAsync(300);
    expect([ui.node.attributeEpoch(2), ui.node.attributeEpoch(3), ui.node.attributeEpoch(4)]).toEqual([1, 1, 0]);
    change(2);
    ui.node.close();
    await vi.advanceTimersByTimeAsync(300);
    expect(ui.node.attributeEpoch(2)).toBe(1);
  });

  it("moves the graph epoch once per burst of session, fabric and endpoint events", async () => {
    const ui = testUI({ fixtures: ["node-status.200", "node-fabrics.200"] });
    const s = { session_id: 17, fabric_index: 1 } as never;
    ui.node.applyEvent({ type: "session_opened", data: s });
    ui.node.applyEvent({ type: "subscriptions_changed", data: s });
    ui.node.applyEvent({ type: "fabric_added", data: { fabric_index: 1 } });
    ui.node.applyEvent({ type: "window_opened", data: {} });
    expect(ui.node.graphEpoch).toBe(0);
    await vi.advanceTimersByTimeAsync(400);
    expect(ui.node.graphEpoch).toBe(1);
    ui.node.applyEvent({ type: "session_closed", data: s });
    ui.node.applyEvent({ type: "endpoints_changed", data: {} });
    ui.node.applyEvent({ type: "fabric_removed", data: { fabric_index: 1 } });
    await vi.advanceTimersByTimeAsync(400);
    expect(ui.node.graphEpoch).toBe(2);
    ui.node.applyEvent({ type: "exposure_changed", data: { keys: [] } });
    ui.node.applyEvent({ type: "session_opened", data: s });
    ui.node.close();
    await vi.advanceTimersByTimeAsync(400);
    expect(ui.node.graphEpoch).toBe(2);
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

  it("reloads fabrics on fabric_updated; on commissioning_complete toasts and reloads status and fabrics (0.7.0)", async () => {
    const ui = testUI({ fixtures: ["node-fabrics.200", "node-status.200"] });
    ui.node.applyEvent({ type: "fabric_updated", data: { fabric_index: 1, label: "Home" } });
    await vi.runOnlyPendingTimersAsync();
    expect(paths(ui)).toEqual(["GET /node/fabrics"]);
    ui.node.applyEvent({ type: "commissioning_complete", data: { fabric_index: 3 } });
    await vi.runOnlyPendingTimersAsync();
    expect(ui.toastSpy.success).toHaveBeenCalledWith('node.commissioning.complete:{"index":3}');
    expect(paths(ui)).toEqual(expect.arrayContaining(["GET /node/status"]));
  });

  it("keeps the last cluster_event for the endpoint page (0.7.0)", () => {
    const ui = testUI();
    const ev = { number: 4, priority: "info" as const, endpoint: 2, cluster_hex: "0x0039", event_hex: "0x0003" };
    ui.node.applyEvent({ type: "cluster_event", data: ev });
    expect(ui.node.clusterEvent).toEqual(ev);
    expect(ui.sent).toEqual([]);
  });
});

describe("controller store", () => {
  it("moves the graph epoch once per burst of topology_changed, and not after close", async () => {
    const ui = testUI();
    ui.controller.applyEvent({ type: "topology_changed", data: {} });
    ui.controller.applyEvent({ type: "topology_changed", data: {} });
    expect(ui.controller.graphEpoch).toBe(0);
    await vi.advanceTimersByTimeAsync(400);
    expect(ui.controller.graphEpoch).toBe(1);
    expect(ui.controller.revision).toBe(0);
    ui.controller.applyEvent({ type: "topology_changed", data: {} });
    ui.controller.close();
    await vi.advanceTimersByTimeAsync(400);
    expect(ui.controller.graphEpoch).toBe(1);
    expect(ui.sent).toEqual([]);
  });

  it("reports an update's end with a toast and reloads the queue on every kind (0.7.0)", async () => {
    const ui = testUI({ fixtures: ["controller-updates.200"] });
    ui.controller.applyEvent({ type: "update", data: { node_id_hex: "0000000000000011", update_kind: "queued", update_version: 2 } });
    expect(ui.toastSpy.success).not.toHaveBeenCalled();
    ui.controller.applyEvent({ type: "update", data: { node_id_hex: "0000000000000011", update_kind: "done", update_version: 2 } });
    ui.controller.applyEvent({ type: "update", data: { node_id_hex: "0000000000000011", update_kind: "failed", update_error: "cancelled" } });
    expect(ui.toastSpy.success).toHaveBeenCalledWith('controller.updates.event_done:{"version":2}');
    expect(ui.toastSpy.error).toHaveBeenCalledWith('controller.updates.event_failed:{"error":"cancelled"}');
    await vi.runOnlyPendingTimersAsync();
    expect(ui.sent.filter((s) => s.path === "/api/v1/controller/updates")).toHaveLength(3);
  });

  it("starts a new step list with each commissioning", () => {
    const ui = testUI();
    const step = { number: 2, sub: 2, name: "PASE", phase: "started" as const };
    ui.controller.applyEvent({ type: "commissioning_progress", data: { phase: "started" } });
    ui.controller.applyEvent({ type: "commissioning_progress", data: { phase: "started", step } });
    ui.controller.applyEvent({ type: "commissioning_progress", data: { phase: "succeeded", node_id_hex: "13" } });
    expect(ui.controller.progress).toMatchObject({ phase: "succeeded", steps: [step] });
    ui.controller.applyEvent({ type: "commissioning_progress", data: { phase: "started" } });
    expect(ui.controller.progress.steps).toEqual([]);
  });

  it("keeps the BLE decision under the phase started, cleared by the next commissioning (0.7.1)", () => {
    const ui = testUI();
    const ble = { used: false, reason: "no_network" as const };
    ui.controller.applyEvent({ type: "commissioning_progress", data: { phase: "started" } });
    ui.controller.applyEvent({ type: "commissioning_progress", data: { phase: "started", ble } });
    expect(ui.controller.progress).toMatchObject({ phase: "started", ble, steps: [] });
    ui.controller.applyEvent({ type: "commissioning_progress", data: { phase: "succeeded", node_id_hex: "13" } });
    expect(ui.controller.progress.ble).toEqual(ble);
    ui.controller.applyEvent({ type: "commissioning_progress", data: { phase: "started" } });
    expect(ui.controller.progress.ble).toBeNull();
  });
});
