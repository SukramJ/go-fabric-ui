// @vitest-environment happy-dom
// Ported from openccu-loom assets/ui/src/routes/matter/MatterDiagnostics.test.ts
// at b40b289a, on the contract fixtures (the endpoint cases moved to
// Endpoints.test.ts).
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, render, screen } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { body, testUI } from "../../../test/support";
import type { Compatibility, DiagnosticEventList, Mdns, SessionList } from "../../api/client";
import Diagnostics from "./Diagnostics.svelte";
import Endpoints from "./Endpoints.svelte";

afterEach(cleanup);

const ALL = ["node-sessions.200", "node-mdns.200", "node-compatibility.200", "node-events.200", "node-endpoints.200"];

function renderDiag(answers: Parameters<typeof testUI>[0] extends infer O ? (O extends { answers?: infer A } ? A : never) : never = {}) {
  const ui = testUI({ fixtures: ALL, answers });
  render(Harness, { ui, page: Diagnostics });
  return ui;
}

describe("Diagnostics", () => {
  it("shows a discovery finding with its explanation", async () => {
    const m = body<Mdns>("node-mdns.200");
    m.findings = [{ severity: "error", code: "no_address", message: "No address is advertised." }];
    renderDiag({ "GET /node/mdns": { status: 200, body: m } });
    expect(await screen.findByText("node.diag.severity.error")).toBeInTheDocument();
    expect(screen.getByText("No address is advertised.")).toBeInTheDocument();
  });

  it("says when the node does not advertise at all", async () => {
    renderDiag({ "GET /node/mdns": { status: 200, body: { advertising: false, services: [], findings: [] } } });
    expect(await screen.findByText("node.diag.not_advertising")).toBeInTheDocument();
  });

  it("renders how long each controller has been quiet", async () => {
    renderDiag();
    expect(await screen.findByText('node.diag.age_seconds:{"n":45}')).toBeInTheDocument();
  });

  it("marks a session that carries no subscriptions, but not a pairing session", async () => {
    const s = body<SessionList>("node-sessions.200");
    s.sessions[0].subscriptions = 0;
    renderDiag({ "GET /node/sessions": { status: 200, body: s } });
    expect(await screen.findAllByText("node.diag.no_subscriptions")).toHaveLength(1);
    expect(screen.getByText("node.diag.pase")).toBeInTheDocument();
  });

  it("reports the session-table occupancy including ids reserved by incomplete handshakes", async () => {
    renderDiag();
    expect(await screen.findByTestId("occupancy")).toHaveTextContent('"reserved":1');
  });

  it("reports an ecosystem incompatibility for the paired controllers", async () => {
    const c = body<Compatibility>("node-compatibility.200");
    c.findings = [{ ecosystem: "amazon", code: "endpoint_ceiling", message: "Alexa drops a bridge above 90 endpoints." }];
    renderDiag({ "GET /node/compatibility": { status: 200, body: c } });
    expect(await screen.findByText("Alexa drops a bridge above 90 endpoints.")).toBeInTheDocument();
    expect(screen.getByText("node.diag.ecosystem.apple")).toBeInTheDocument();
  });

  it("shows a recorded pairing failure", async () => {
    renderDiag();
    expect(await screen.findByText("A controller's pairing attempt was refused.")).toBeInTheDocument();
  });

  it("renders two identical events recorded in the same second", async () => {
    const ev = body<DiagnosticEventList>("node-events.200");
    ev.events = [ev.events[0], ev.events[0]];
    renderDiag({ "GET /node/events": { status: 200, body: ev } });
    expect(await screen.findAllByTestId("event")).toHaveLength(2);
  });

  it("does not claim an empty trace when the fetch failed", async () => {
    const ui = testUI({ fixtures: [...ALL, "node-sessions.503"] });
    render(Harness, { ui, page: Diagnostics });
    expect(await screen.findByText(/common\.error/)).toBeInTheDocument();
    expect(screen.queryByText("node.diag.events_empty")).toBeNull();
  });
});

describe("Endpoints", () => {
  it("lists the bridged endpoints with their names, device types and parents", async () => {
    const ui = testUI({ fixtures: ["node-endpoints.200"] });
    render(Harness, { ui, page: Endpoints });
    expect(await screen.findAllByTestId("endpoint")).toHaveLength(3);
    expect(screen.getByText("Kitchen lamp")).toBeInTheDocument();
    expect(screen.getByText("On/Off Light")).toBeInTheDocument();
    expect(screen.getByText('node.endpoints.part_of:{"id":3}')).toBeInTheDocument();
    expect(screen.getByText("node.endpoints.unreachable")).toBeInTheDocument();
  });

  it("says so when nothing is exposed", async () => {
    const ui = testUI({ answers: { "GET /node/endpoints": { status: 200, body: { endpoints: [] } } } });
    render(Harness, { ui, page: Endpoints });
    expect(await screen.findByText("node.endpoints.empty")).toBeInTheDocument();
  });
});
