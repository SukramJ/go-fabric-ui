// @vitest-environment happy-dom
// The Topology page on the node-topology fixtures: the graph of a bridge
// with two devices (one with a part), two fabrics and three sessions, the
// links into the pages that list each node, the reload on the stream's
// graph events and the error and empty states.
import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { body, testUI } from "../../../test/support";
import type { TopologyGraph } from "../../api/client";
import Topology from "./Topology.svelte";

afterEach(() => {
  cleanup();
  vi.useRealTimers();
});

const nodeIds = () => screen.getAllByTestId("graph-node").map((g) => g.getAttribute("data-id"));
const graphNode = (id: string) => screen.getAllByTestId("graph-node").find((g) => g.getAttribute("data-id") === id)!;

describe("Topology", () => {
  it("draws the node's graph from GET /node/topology", async () => {
    const ui = testUI({ fixtures: ["node-topology.200"] });
    render(Harness, { ui, page: Topology });
    await screen.findByTestId("graph");
    expect(screen.getByRole("group", { name: "node.topology.graph_label" })).toBeInTheDocument();
    expect(nodeIds()).toEqual(["node", "ep:1", "ep:2", "ep:3", "ep:4", "fabric:1", "fabric:2", "session:17", "session:19", "session:18"]);
    expect(ui.sent.map((s) => s.path)).toEqual(["/api/v1/node/topology"]);
  });

  it("links a node to the page that lists it, on the context's routes", async () => {
    const ui = testUI({ fixtures: ["node-topology.200"] });
    render(Harness, { ui, page: Topology });
    await screen.findByTestId("graph");
    const panel = screen.getByTestId("graph-panel");
    for (const [id, href] of [
      ["ep:4", "#/node/endpoint/4"],
      ["fabric:2", "#/node/fabrics"],
      ["session:18", "#/node/diagnostics"],
      ["node", "#/node/status"],
    ]) {
      await fireEvent.click(graphNode(id));
      expect(within(panel).getByRole("link")).toHaveAttribute("href", href);
    }
    // A detail key the translator does not name shows as written.
    await fireEvent.click(graphNode("session:18"));
    expect(panel).toHaveTextContent("session_type");
    expect(panel).toHaveTextContent("PASE");
  });

  it("reloads once on a burst of the stream's session, fabric and endpoint events", async () => {
    vi.useFakeTimers();
    const ui = testUI({ fixtures: ["node-topology.200"] });
    render(Harness, { ui, page: Topology });
    await vi.waitFor(() => expect(screen.getByTestId("graph")).toBeInTheDocument());
    const s = body<{ sessions: never[] }>("node-sessions.200").sessions[0];
    ui.node.applyEvent({ type: "session_opened", data: s });
    ui.node.applyEvent({ type: "subscriptions_changed", data: s });
    ui.node.applyEvent({ type: "fabric_added", data: { fabric_index: 1 } });
    ui.node.applyEvent({ type: "endpoints_changed", data: {} });
    await vi.advanceTimersByTimeAsync(400);
    const topo = () => ui.sent.filter((x) => x.path === "/api/v1/node/topology");
    expect(topo()).toHaveLength(2);
    ui.node.applyEvent({ type: "session_closed", data: s });
    await vi.advanceTimersByTimeAsync(400);
    expect(topo()).toHaveLength(3);
    expect(screen.getByTestId("graph")).toBeInTheDocument();
  });

  it("shows the server's problem with a retry while the node is unavailable", async () => {
    const ui = testUI({ fixtures: ["node-topology.503"] });
    render(Harness, { ui, page: Topology });
    expect(await screen.findByText(/common\.error the host serves no node/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "common.reload" })).toBeInTheDocument();
    expect(screen.queryByTestId("graph")).toBeNull();
  });

  it("says there is nothing to draw for a node without devices and controllers", async () => {
    const g = body<TopologyGraph>("node-topology.200");
    g.nodes = g.nodes.filter((n) => n.kind === "node" || n.kind === "aggregator");
    g.edges = g.edges.filter((e) => e.source === "node" && e.target === "ep:1");
    const ui = testUI({ answers: { "GET /node/topology": { status: 200, body: g } } });
    render(Harness, { ui, page: Topology });
    expect(await screen.findByText("node.topology.empty")).toBeInTheDocument();
    expect(screen.queryByTestId("graph")).toBeNull();
  });

  it("refreshes on the button", async () => {
    const ui = testUI({ fixtures: ["node-topology.200"] });
    render(Harness, { ui, page: Topology });
    await screen.findByTestId("graph");
    await fireEvent.click(screen.getByRole("button", { name: "common.refresh" }));
    await waitFor(() => expect(ui.sent.filter((x) => x.path === "/api/v1/node/topology")).toHaveLength(2));
  });
});
