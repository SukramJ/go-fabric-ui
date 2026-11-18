// @vitest-environment happy-dom
// The node's endpoint page on the node-endpoint fixtures: the lamp's
// device types, scope, tags and labels, its clusters with their values,
// the surface decisions on it and its events newest first; the error
// state of an endpoint the node does not serve; the refetch on the
// endpoint's attribute_changed, which fetches only the newer events.
import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { body, fixtureFetch, testUI } from "../../../test/support";
import { createClient } from "../../api/client";
import Endpoint from "./Endpoint.svelte";

afterEach(() => {
  cleanup();
  vi.useRealTimers();
});

function renderEndpoint(endpoint: number, extra: Parameters<typeof testUI>[0] = {}) {
  const ui = testUI({ fixtures: ["node-endpoint.200", "node-endpoint-events.200"], ...extra });
  render(Harness, { ui, page: Endpoint, pageProps: { endpoint } });
  return ui;
}

describe("Node Endpoint", () => {
  it("shows the endpoint with its device types, scope, tags, labels and values", async () => {
    renderEndpoint(2);
    expect(await screen.findByTestId("endpoint-title")).toHaveTextContent("Kitchen lamp");
    const head = screen.getByTestId("endpoint-head");
    expect(head).toHaveTextContent("OnOffLight r3");
    expect(head).toHaveTextContent("BridgedNode r3");
    expect(head).toHaveTextContent("node.endpoints.reachable");
    expect(head).toHaveTextContent('node.endpoint.scope:{"scope":"default"}');
    expect(head).toHaveTextContent("One");
    expect(screen.getByTestId("endpoint-labels")).toHaveTextContent("room: Kitchen");
    expect(head.querySelector('a[href="#/node/endpoint/1"]')).not.toBeNull();
    const onOff = screen.getByTestId("cluster-view-0x0006");
    expect(onOff).toHaveTextContent("LT · Lighting");
    expect(onOff).toHaveTextContent('inventory.data_version:{"version":41}');
    expect(screen.getAllByTestId("attribute-0x0000")[0]).toHaveTextContent("OnOff");
    expect(screen.getAllByTestId("attribute-0x0000")[0]).toHaveTextContent("true");
    expect(screen.getByTestId("cluster-view-0x0039")).toHaveTextContent("ReachableChanged");
  });

  it("shows a failed attribute with its status by name and no value", async () => {
    renderEndpoint(2);
    const view = await screen.findByTestId("cluster-view-0x0039");
    const row = view.querySelector('[data-testid="attribute-0x0001"]')!;
    expect(row).toHaveTextContent("VendorName");
    expect(row).toHaveTextContent("inventory.no_value");
    expect(row).toHaveTextContent("inventory.status.failure");
    expect(row).toHaveTextContent("0x01");
    expect(view.querySelector('[data-testid="attribute-0x0005"]')!.querySelector('[data-testid="attribute-status"]')).toBeNull();
  });

  it("shows the decisions on the endpoint, the silent keeps on request", async () => {
    renderEndpoint(2);
    await screen.findByTestId("endpoint-title");
    expect(screen.getAllByTestId("endpoint-decision")).toHaveLength(2);
    expect(screen.getByTestId("endpoint-decisions")).toHaveTextContent("node.surface.verdict.warn");
    await fireEvent.click(screen.getByText('node.endpoint.decisions_show_all:{"count":1}'));
    expect(screen.getAllByTestId("endpoint-decision")).toHaveLength(3);
  });

  it("lists the endpoint's events newest first with their data", async () => {
    renderEndpoint(2);
    const list = await screen.findByTestId("event-list");
    const items = [...list.querySelectorAll("li")].map((li) => li.getAttribute("data-testid"));
    expect(items).toEqual(["event-9", "event-7"]);
    expect(screen.getByTestId("event-9")).toHaveTextContent("ReachableChanged");
    expect(screen.getByTestId("event-9")).toHaveTextContent("events.priority.info");
    expect(screen.getByTestId("event-9")).toHaveTextContent('"reachableNewValue": true');
  });

  it("shows the error state for an endpoint the node does not serve", async () => {
    renderEndpoint(9, { fixtures: ["node-endpoint.404"], answers: { "GET /node/endpoints/9/events": { status: 200, body: { events: [] } } } });
    expect(await screen.findByText(/no such endpoint: 9/)).toBeInTheDocument();
    expect(await screen.findByTestId("event-list-empty")).toBeInTheDocument();
  });

  it("refetches on its attribute_changed, the events after the last one only", async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    const ui = renderEndpoint(2, {
      answers: {
        "GET /node/endpoints/2/events?since=10": {
          status: 200,
          body: { events: [{ number: 12, priority: "critical", endpoint: 2, cluster_hex: "0x0039", event_hex: "0x0003", name: "ReachableChanged" }] },
        },
      },
    });
    await screen.findByTestId("event-9");
    const detailReads = () => ui.sent.filter((s) => s.path === "/api/v1/node/endpoints/2").length;
    expect(detailReads()).toBe(1);
    ui.node.applyEvent({ type: "attribute_changed", data: { endpoint: 3, cluster_hex: "0x0006", attributes: ["0x0000"] } });
    await vi.advanceTimersByTimeAsync(300);
    expect(detailReads()).toBe(1);
    ui.node.applyEvent({ type: "attribute_changed", data: { endpoint: 2, cluster_hex: "0x0006", attributes: ["0x0000"] } });
    await vi.advanceTimersByTimeAsync(300);
    await waitFor(() => expect(detailReads()).toBe(2));
    expect(await screen.findByTestId("event-12")).toBeInTheDocument();
    expect(screen.getByTestId("event-7")).toBeInTheDocument();
    expect(ui.sent.map((s) => s.path)).toContain("/api/v1/node/endpoints/2/events?since=10");
  });

  it("shows the reason an endpoint has no tags", async () => {
    const d = body<Record<string, unknown>>("node-endpoint.200");
    renderEndpoint(2, { answers: { "GET /node/endpoints/2": { status: 200, body: { ...d, tags: [], scope: undefined } } } });
    expect(await screen.findByText("node.endpoint.no_tags")).toBeInTheDocument();
  });

  it("adds the stream's cluster_event of its endpoint at once, and no other (0.7.0)", async () => {
    const ui = renderEndpoint(2);
    await screen.findByTestId("event-9");
    const ev = { number: 12, priority: "info" as const, endpoint: 2, cluster_hex: "0x0039", event_hex: "0x0003", name: "ReachableChanged" };
    ui.node.applyEvent({ type: "cluster_event", data: { ...ev, number: 11, endpoint: 3 } });
    ui.node.applyEvent({ type: "cluster_event", data: ev });
    const list = await screen.findByTestId("event-list");
    await waitFor(() => expect([...list.querySelectorAll("li")].map((li) => li.getAttribute("data-testid"))).toEqual(["event-12", "event-9", "event-7"]));
  });

  it("lists a cluster_event that lands during an incremental read once (UR-P1-ui-131)", async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    const ui = testUI({ fixtures: ["node-endpoint.200", "node-endpoint-events.200"] });
    const ev = { number: 12, priority: "info" as const, endpoint: 2, cluster_hex: "0x0039", event_hex: "0x0003", name: "ReachableChanged" };
    const orig = ui.client.nodeEndpointEvents.bind(ui.client);
    let release: (() => void) | undefined;
    ui.client.nodeEndpointEvents = (ep, q) =>
      q?.since !== undefined ? new Promise((resolve) => (release = () => resolve({ events: [ev] }))) : orig(ep, q);
    render(Harness, { ui, page: Endpoint, pageProps: { endpoint: 2 } });
    await screen.findByTestId("event-9");
    ui.node.applyEvent({ type: "attribute_changed", data: { endpoint: 2, cluster_hex: "0x0006", attributes: ["0x0000"] } });
    await vi.advanceTimersByTimeAsync(300);
    await waitFor(() => expect(release).toBeDefined());
    ui.node.applyEvent({ type: "cluster_event", data: ev });
    expect(await screen.findByTestId("event-12")).toBeInTheDocument();
    release!();
    await vi.advanceTimersByTimeAsync(50);
    const list = screen.getByTestId("event-list");
    expect([...list.querySelectorAll("li")].map((li) => li.getAttribute("data-testid"))).toEqual(["event-12", "event-9", "event-7"]);
  });

  it("drops the previous endpoint and its baselines when the prop changes (UR-P1-ui-144)", async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    const ui = testUI({ fixtures: ["node-endpoint.200", "node-endpoint-events.200"] });
    const d3 = { ...body<Record<string, unknown>>("node-endpoint.200"), endpoint_id: 3 };
    const orig = ui.client.nodeEndpointDetail.bind(ui.client);
    let release: (() => void) | undefined;
    let reads3 = 0;
    ui.client.nodeEndpointDetail = (ep) => {
      if (ep !== 3) return orig(ep);
      reads3++;
      return new Promise((resolve) => (release = () => resolve(d3 as never)));
    };
    ui.client.nodeEndpointEvents = async () => ({ events: [] });
    const r = render(Harness, { ui, page: Endpoint, pageProps: { endpoint: 2 } });
    await screen.findByTestId("endpoint-head");
    ui.node.applyEvent({ type: "attribute_changed", data: { endpoint: 3, cluster_hex: "0x0006", attributes: ["0x0000"] } });
    await vi.advanceTimersByTimeAsync(300);
    await r.rerender({ ui, page: Endpoint, pageProps: { endpoint: 3 } });
    await vi.advanceTimersByTimeAsync(300);
    expect(screen.queryByTestId("endpoint-head")).toBeNull();
    expect(reads3).toBe(1);
    release!();
    await vi.advanceTimersByTimeAsync(50);
    expect(await screen.findByTestId("endpoint-head")).toBeInTheDocument();
  });
});

describe("Node Endpoint — a cluster panel keeps its state across a reload (UR-P2-ui-191)", () => {
  it("keeps a closed panel closed when a cluster is added in front", async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    const ff = fixtureFetch(["node-endpoint.200", "node-endpoint-events.200"]);
    const ui = testUI({ client: createClient("/api/v1", ff.fetch) });
    render(Harness, { ui, page: Endpoint, pageProps: { endpoint: 2 } });
    const panel = (hex: string) => screen.getByTestId(`endpoint-cluster-${hex}`) as HTMLDetailsElement;
    await screen.findByTestId("endpoint-cluster-0x0006");
    expect(panel("0x0006").open).toBe(true);
    panel("0x0006").open = false;
    await fireEvent(panel("0x0006"), new Event("toggle"));
    // The reload answers one more cluster, in front of the others.
    const detail = body<{ clusters: Record<string, unknown>[] }>("node-endpoint.200");
    detail.clusters = [{ ...detail.clusters[1], id: 3, id_hex: "0x0003", name: "Identify" }, ...detail.clusters];
    ff.routes.set("GET /api/v1/node/endpoints/2", { route: "", request: { method: "GET", path: "" }, response: { status: 200, body: detail } });
    ui.node.applyEvent({ type: "attribute_changed", data: { endpoint: 2, cluster_hex: "0x0006", attributes: ["0x0000"] } });
    await vi.advanceTimersByTimeAsync(300);
    await screen.findByTestId("endpoint-cluster-0x0003");
    expect({ onOff: panel("0x0006").open, added: panel("0x0003").open }).toEqual({ onOff: false, added: true });
  });
});
