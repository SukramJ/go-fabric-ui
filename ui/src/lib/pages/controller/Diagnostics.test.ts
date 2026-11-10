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
});
