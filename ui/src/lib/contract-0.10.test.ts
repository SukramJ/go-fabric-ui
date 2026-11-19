// @vitest-environment happy-dom
// The pages on contract 0.10.0's fields: the window's remaining seconds, an
// absent fabric count, a selection stored but not rebuilt, an endpoint's
// refused bindings read, the auto-subscribe setting, event numbers beyond
// 2^53, the commissioning id, the ACL add form, the passcodes and the
// fabric label the server refuses, and a refused log.
import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import Harness from "../test/Harness.svelte";
import NodeHarness from "../test/NodeHarness.svelte";
import { body, fixture, testUI } from "../test/support";
import type { ControllerEvent, EventRecord, Status } from "./api/client";
import { isValidPasscode } from "./api/client";
import { lastNumber, mergeEvents, sinceAfter } from "./kit/events";
import { windowSecondsLeft } from "./countdown";
import Status_ from "./pages/node/Status.svelte";
import Selection from "./pages/node/Selection.svelte";
import Commission from "./pages/controller/Commission.svelte";
import Settings from "./pages/controller/Settings.svelte";
import Operation from "./pages/operation/Operation.svelte";

afterEach(() => {
  cleanup();
  vi.useRealTimers();
});

const L = "0000000000000011";
const status = (patch: Partial<Status>) => ({ status: 200, body: { ...body<Status>("node-status.200"), ...patch } });

describe("the commissioning window's remaining seconds (UR-P2-ui-06)", () => {
  it("counts from window_remaining_seconds and the time the status was read, not from the deadline", () => {
    const st = { window_remaining_seconds: 100, window_deadline: "2000-01-01T00:00:00Z" };
    expect(windowSecondsLeft(st, 1_000_000, 1_010_500)).toBe(90);
    expect(windowSecondsLeft(st, 1_000_000, 2_000_000)).toBe(0);
    expect(windowSecondsLeft({ window_deadline: new Date(5_000).toISOString() }, 0, 0)).toBe(5);
    expect(windowSecondsLeft(null, 0)).toBeNull();
  });

  it("the status page shows it", async () => {
    const ui = testUI({ answers: { "GET /node/status": status({ commissioning_window_open: true, window_remaining_seconds: 42, window_deadline: undefined }) } });
    render(Harness, { ui, page: Status_ });
    expect(await screen.findByTestId("window-left")).toHaveTextContent('"n":42');
  });
});

describe("Status.fabric_count is optional (UR-P1-err-07)", () => {
  it("says the count is unknown when the fabric read failed", async () => {
    const st = { ...body<Status>("node-status.200") };
    delete (st as Partial<Status>).fabric_count;
    render(Harness, { ui: testUI({ answers: { "GET /node/status": { status: 200, body: st } } }), page: Status_ });
    expect(await screen.findByTestId("status-fabrics")).toHaveTextContent("node.status.fabrics_unknown");
  });
});

describe("an exposure apply whose rebuild failed (UR-P2-ui-35)", () => {
  it("says the selection is stored but not in effect", async () => {
    const ui = testUI({
      fixtures: ["node-status.200", "node-exposure-schema.200", "node-exposure.200"],
      answers: { "PUT /node/exposure": { status: 200, body: { applied: 1, reassembled: false } } },
    });
    await ui.node.loadStatus();
    render(Harness, { ui, page: Selection });
    await waitFor(() => expect(ui.node.items.length).toBeGreaterThan(0));
    const item = ui.node.items[0];
    ui.node.markDirty(item, { enabled: !item.enabled });
    await ui.node.saveBulk();
    expect(ui.node.exposureNotInEffect).toBe(true);
    expect(await screen.findByTestId("selection-not-in-effect")).toHaveTextContent("node.selection.not_in_effect");
  });
});

describe("the controller node's tabs on 0.10.0 fields", () => {
  it("an endpoint whose bindings read the node refused shows its status and offers no write (UR-P1-bound-37)", async () => {
    const ui = testUI({
      fixtures: ["controller-node.200"],
      answers: { [`GET /controller/nodes/${L}/bindings`]: { status: 200, body: { endpoints: [{ endpoint: 1, bindings: [], status: 0x7e }] } } },
    });
    render(NodeHarness, { ui, nodeId: L, tab: "bindings" });
    expect(await screen.findByTestId("bindings-status-1")).toHaveTextContent("controller.bindings.read_status");
    expect(screen.queryByText("controller.bindings.add")).toBeNull();
  });

  it("the overview's toggle follows the auto-subscribe setting, not the live subscription (UR-P1-ctr-168)", async () => {
    const d = body<{ node: Record<string, unknown> }>("controller-node.200");
    const ui = testUI({ answers: { [`GET /controller/nodes/${L}`]: { status: 200, body: { ...d, node: { ...d.node, subscription_active: false, auto_subscribe: true } } } } });
    render(NodeHarness, { ui, nodeId: L, tab: "overview" });
    expect(await screen.findByText("controller.overview.unsubscribe")).toBeInTheDocument();
  });

  it("the ACL add form offers CASE and Group only, and an unknown entry is not written back (UR-P1-ui-179)", async () => {
    const ui = testUI({
      fixtures: ["controller-node.200"],
      answers: {
        [`GET /controller/nodes/${L}/acl`]: {
          status: 200,
          body: { entries: [{ privilege: "administer", auth_mode: "case", subjects_hex: [], targets: [] }, { privilege: "unknown", auth_mode: "pase", subjects_hex: [], targets: [] }] },
        },
      },
    });
    render(NodeHarness, { ui, nodeId: L, tab: "acl" });
    const list = await screen.findByTestId("acl-entries");
    // Display keeps PASE and names an unknown privilege.
    expect(list).toHaveTextContent("controller.acl.mode.pase");
    expect(list).toHaveTextContent("controller.acl.privilege.unknown");
    const mode = screen.getByLabelText("controller.acl.mode_label");
    const options = [...mode.querySelectorAll("option")].map((o) => o.value);
    expect(options).toEqual(["case", "group"]);
    await fireEvent.input(screen.getByLabelText("controller.acl.subjects"), { target: { value: "1" } });
    await fireEvent.click(screen.getByText("controller.acl.add_entry"));
    await fireEvent.click(screen.getByText("common.save"));
    await waitFor(() => expect(ui.toastSpy.error).toHaveBeenCalledWith("controller.acl.unknown_entry"));
    expect(ui.sent.some((s) => s.method === "PUT")).toBe(false);
  });
});

describe("event numbers beyond 2^53 (UR-P1-ctr-113, -177)", () => {
  const ev = (n: string): EventRecord => ({ number: Number(n), number_str: n, priority: "info", endpoint: 1, cluster_hex: "0x0028", event_hex: "0x0000" }) as EventRecord;

  it("are merged and ordered by number_str, exactly, and the next read starts after the highest", () => {
    const a = ev("18446744073709551614");
    const b = ev("18446744073709551615");
    // Both numbers round to the same JS number.
    expect(a.number).toBe(b.number);
    const merged = mergeEvents([b], [a, b]);
    expect(merged.map((e) => e.number_str)).toEqual(["18446744073709551614", "18446744073709551615"]);
    expect(lastNumber(merged)).toBe(18446744073709551615n);
    expect(sinceAfter(merged)).toEqual({ since: "18446744073709551616" });
    expect(sinceAfter([])).toEqual({});
  });
});

describe("the commissioning (UR-P1-api-18, -ui-178)", () => {
  it("refuses the passcodes the contract refuses", () => {
    for (const p of [11111111, 22222222, 88888888, 12345678, 87654321, 0, 99999999, 1.5]) expect(isValidPasscode(p), String(p)).toBe(false);
    for (const p of [20202021, 1, 99999998]) expect(isValidPasscode(p), String(p)).toBe(true);
  });

  it("refuses a discovered device's invalid passcode before any request", async () => {
    const ui = testUI({ answers: { "GET /controller/discovery?timeout=5": fixture("controller-discovery.200").response } });
    render(Harness, { ui, page: Commission });
    await fireEvent.click(screen.getByText("controller.commission.search"));
    await screen.findByText("Plug");
    await fireEvent.input(screen.getByLabelText("controller.commission.passcode"), { target: { value: "12345678" } });
    await fireEvent.click(screen.getByText("controller.commission.pair"));
    expect(ui.toastSpy.error).toHaveBeenCalledWith("controller.commission.passcode_invalid");
    expect(ui.sent.some((s) => s.method === "POST")).toBe(false);
  });

  it("shows the progress of its own run only, matched by commissioning_id (decision 59)", () => {
    const ui = testUI();
    const id = ui.controller.beginCommissioning();
    expect(id).toMatch(/^[A-Za-z0-9_-]{1,64}$/);
    const progress = (data: Partial<ControllerEvent>) => ui.controller.applyEvent({ type: "commissioning_progress", data: data as ControllerEvent });
    progress({ phase: "failed", detail: "timeout", commissioning_id: "another" });
    expect(ui.controller.progress.phase).toBe("idle");
    progress({ phase: "started", commissioning_id: id });
    expect(ui.controller.progress.phase).toBe("started");
    // A server before 0.10.0 sends no id: its events are shown.
    progress({ phase: "succeeded", node_id_hex: L });
    expect(ui.controller.progress.phase).toBe("succeeded");
    expect(ui.controller.beginCommissioning()).not.toBe(id);
  });
});

describe("the fabric label (UR-P1-adpt-40)", () => {
  it("refuses a label over 32 bytes of UTF-8, though under 32 characters", async () => {
    const ui = testUI({ fixtures: ["controller-status.200", "controller-credentials.200"] });
    render(Harness, { ui, page: Settings });
    await screen.findByText("HomeNet");
    // 11 characters, 33 bytes.
    await fireEvent.input(screen.getByLabelText("controller.settings.label"), { target: { value: "€".repeat(11) } });
    expect(await screen.findByTestId("label-too-long")).toBeInTheDocument();
    expect(screen.getAllByText("common.save")[0]).toBeDisabled();
    await fireEvent.input(screen.getByLabelText("controller.settings.label"), { target: { value: "€".repeat(10) } });
    await waitFor(() => expect(screen.queryByTestId("label-too-long")).toBeNull());
  });
});

describe("the log without admin rights (decision 20)", () => {
  it("says the log needs admin rights and closes its stream, the info and level stay", async () => {
    const close = vi.fn();
    const ui = testUI({
      fixtures: ["operation-info.200", "operation-log-level.200"],
      answers: { "GET /operation/log?limit=2000": { status: 403, body: { code: "forbidden", title: "Admin rights required", status: 403, type: "x" } } },
      subscribeLog: () => close,
    });
    render(Harness, { ui, page: Operation });
    expect(await screen.findByTestId("operation-info")).toBeInTheDocument();
    expect(await screen.findByTestId("denied")).toHaveAttribute("data-code", "forbidden");
    expect(close).toHaveBeenCalled();
  });
});
