// @vitest-environment happy-dom
// The controller's Diagnostics page on the controller-diagnostics,
// -mdns and -events fixtures (contract 0.9.0): the discovery finding, the
// peers with their state and their link to the node's page, the sessions
// with transport and resume, the events newest first, the reload on a
// burst of controller_diagnostic_event, and the source that serves no
// diagnostics (capabilities.diagnostics false, or 501).
import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, render, screen, within } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { body, testUI } from "../../../test/support";
import type { ControllerStatus } from "../../api/client";
import Diagnostics from "./Diagnostics.svelte";

afterEach(() => {
  cleanup();
  vi.useRealTimers();
});

const FIXTURES = ["controller-diagnostics.200", "controller-mdns.200", "controller-events.200"];
const diagCalls = (ui: ReturnType<typeof testUI>) => ui.sent.filter((x) => x.path === "/api/v1/controller/diagnostics");

describe("Controller Diagnostics", () => {
  it("shows the finding, the peers, the sessions and the events newest first", async () => {
    const ui = testUI({ fixtures: FIXTURES });
    render(Harness, { ui, page: Diagnostics });
    const finding = await screen.findByTestId("finding");
    expect(finding).toHaveTextContent("None of the 2 commissioned nodes");
    const links = screen.getAllByTestId("peer-link");
    expect(links.map((a) => a.getAttribute("href"))).toEqual(["#/controller/node/0000000000000011", "#/controller/node/0000000000000012"]);
    expect(screen.getByText("controller.connection.connected")).toBeInTheDocument();
    expect(screen.getByText("controller.connection.waiting_for_discovery")).toBeInTheDocument();
    expect(screen.getByText("UDP")).toBeInTheDocument();
    expect(screen.getByText("controller.diag.resumed")).toBeInTheDocument();
    // capacity 0: the controller tracks no occupancy, so no line.
    expect(screen.queryByTestId("occupancy")).toBeNull();
    const events = screen.getAllByTestId("event");
    expect(events).toHaveLength(2);
    expect(within(events[0]).getByText("node.diag.kind_icd")).toBeInTheDocument();
    expect(within(events[1]).getByText("node.diag.kind_session")).toBeInTheDocument();
  });

  it("reloads once on a burst of controller_diagnostic_event", async () => {
    vi.useFakeTimers();
    const ui = testUI({ fixtures: FIXTURES });
    render(Harness, { ui, page: Diagnostics });
    await vi.waitFor(() => expect(screen.getAllByTestId("event")).toHaveLength(2));
    const ev = body<{ events: { at: string; kind: "icd"; severity: "warning"; message: string }[] }>("controller-events.200").events[0];
    ui.controller.applyEvent({ type: "controller_diagnostic_event", data: ev });
    ui.controller.applyEvent({ type: "controller_diagnostic_event", data: ev });
    await vi.advanceTimersByTimeAsync(400);
    expect(diagCalls(ui)).toHaveLength(2);
    expect(screen.getAllByTestId("event")).toHaveLength(2);
  });

  it("shows the not-served state when the routes answer 501", async () => {
    const ui = testUI({ fixtures: ["controller-diagnostics.501", "controller-mdns.200", "controller-events.200"] });
    render(Harness, { ui, page: Diagnostics });
    expect(await screen.findByText("controller.diag.not_served")).toBeInTheDocument();
    expect(ui.toastSpy.error).not.toHaveBeenCalled();
  });

  it("asks no route when the status says the source serves no diagnostics", async () => {
    const status = body<ControllerStatus>("controller-status.200");
    const ui = testUI({ answers: { "GET /controller/status": { status: 200, body: { ...status, capabilities: { ...status.capabilities, diagnostics: false } } } } });
    await ui.controller.loadStatus();
    render(Harness, { ui, page: Diagnostics });
    expect(await screen.findByText("controller.diag.not_served")).toBeInTheDocument();
    expect(diagCalls(ui)).toHaveLength(0);
  });

  it("renders two findings with the same code and no service (UR-P1-ui-130)", async () => {
    const findings = [
      { severity: "warning", code: "interface_not_found", message: "Interface eth9 not found." },
      { severity: "warning", code: "interface_not_found", message: "Interface wlan9 not found." },
    ];
    const ui = testUI({ fixtures: FIXTURES, answers: { "GET /controller/mdns": { status: 200, body: { findings } } } });
    render(Harness, { ui, page: Diagnostics });
    expect(await screen.findByText("Interface eth9 not found.")).toBeInTheDocument();
    expect(screen.getByText("Interface wlan9 not found.")).toBeInTheDocument();
  });
});

describe("UR-P4-ui-301: the peers are named when the page is the first one opened", () => {
  it("loads the node list itself and shows a commissioned peer by its name", async () => {
    const ui = testUI({ fixtures: [...FIXTURES, "controller-nodes.200"] });
    render(Harness, { ui, page: Diagnostics });
    const links = await screen.findAllByTestId("peer-link");
    await vi.waitFor(() => expect(links[0]).toHaveTextContent("Kitchen"));
  });
});

describe("UR-P4-ui-303: a session's peer links only to a node the controller holds", () => {
  it("links a registered peer, and shows a PASE session's and an unknown peer's id as text", async () => {
    const d = body<{ sessions: Record<string, unknown>[] }>("controller-diagnostics.200");
    const s0 = d.sessions[0];
    d.sessions = [s0, { ...s0, session_id: 300, peer_node_id_hex: "0000000000000000", is_pase: true }, { ...s0, session_id: 301, peer_node_id_hex: "00000000000000AB" }];
    const ui = testUI({ fixtures: FIXTURES, answers: { "GET /controller/diagnostics": { status: 200, body: d } } });
    render(Harness, { ui, page: Diagnostics });
    const own = await screen.findByTestId("session-peer-256");
    expect(own.tagName).toBe("A");
    expect(own).toHaveAttribute("href", "#/controller/node/0000000000000011");
    expect(screen.getByTestId("session-peer-300").tagName).not.toBe("A");
    expect(screen.getByTestId("session-peer-301").tagName).not.toBe("A");
    expect(screen.getByTestId("session-peer-301")).toHaveTextContent("0x00000000000000AB");
  });
});
