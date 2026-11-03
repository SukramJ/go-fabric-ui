// @vitest-environment happy-dom
// The controller's Network page on the controller-topology fixtures: the
// graph of the controller with its nodes, the links of the commissioned
// nodes to their pages, the reload on topology_changed, the admin refresh
// (refused without admin rights) and the host that serves no graph (501).
import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { body, testUI } from "../../../test/support";
import type { TopologyGraph } from "../../api/client";
import Network from "./Network.svelte";
import Nodes from "./Nodes.svelte";

afterEach(() => {
  cleanup();
  vi.useRealTimers();
});

const nodeIds = () => screen.getAllByTestId("graph-node").map((g) => g.getAttribute("data-id"));
const graphNode = (id: string) => screen.getAllByTestId("graph-node").find((g) => g.getAttribute("data-id") === id)!;
const topo = (ui: ReturnType<typeof testUI>) => ui.sent.filter((x) => x.path.startsWith("/api/v1/controller/topology")).map((x) => x.path);

describe("Controller Network", () => {
  it("draws the controller's graph from GET /controller/topology", async () => {
    const ui = testUI({ fixtures: ["controller-topology.200"] });
    render(Harness, { ui, page: Network });
    await screen.findByTestId("graph");
    expect(screen.getByRole("group", { name: "controller.network.graph_label" })).toBeInTheDocument();
    expect(nodeIds()).toEqual(body<TopologyGraph>("controller-topology.200").nodes.map((n) => n.id));
    expect(topo(ui)).toEqual(["/api/v1/controller/topology"]);
  });

  it("links a commissioned node to its page and nothing else", async () => {
    const ui = testUI({ fixtures: ["controller-topology.200"] });
    render(Harness, { ui, page: Network });
    await screen.findByTestId("graph");
    const panel = screen.getByTestId("graph-panel");
    await fireEvent.click(graphNode("peer:0000000000000012"));
    expect(within(panel).getByRole("link")).toHaveAttribute("href", "#/controller/node/0000000000000012");
    for (const id of ["node", "unknown_0011223344556677", "ap_AABBCCDDEEFF"]) {
      await fireEvent.click(graphNode(id));
      expect(within(panel).queryByRole("link")).toBeNull();
    }
  });

  it("reloads once on a burst of topology_changed", async () => {
    vi.useFakeTimers();
    const ui = testUI({ fixtures: ["controller-topology.200"] });
    render(Harness, { ui, page: Network });
    await vi.waitFor(() => expect(screen.getByTestId("graph")).toBeInTheDocument());
    ui.controller.applyEvent({ type: "topology_changed", data: {} });
    ui.controller.applyEvent({ type: "topology_changed", data: {} });
    await vi.advanceTimersByTimeAsync(400);
    expect(topo(ui)).toHaveLength(2);
    expect(screen.getByTestId("graph")).toBeInTheDocument();
  });

  it("asks for a refreshed graph with refresh=true and keeps the graph when it is refused", async () => {
    const ui = testUI({ fixtures: ["controller-topology.200", "controller-topology.403"] });
    render(Harness, { ui, page: Network });
    await screen.findByTestId("graph");
    await fireEvent.click(screen.getByRole("button", { name: "controller.network.refresh" }));
    await waitFor(() => expect(ui.toastSpy.error).toHaveBeenCalled());
    expect(topo(ui)).toEqual(["/api/v1/controller/topology", "/api/v1/controller/topology?refresh=true"]);
    expect(ui.sent[1].headers["X-Fabric-UI-Token"]).toBe("test-token");
    expect(screen.getByTestId("graph")).toBeInTheDocument();
  });

  it("says the host serves no graph on 501, without the refresh", async () => {
    const ui = testUI({ fixtures: ["controller-topology.501"] });
    render(Harness, { ui, page: Network });
    expect(await screen.findByText("controller.network.not_served")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "controller.network.refresh" })).toBeNull();
    expect(screen.queryByTestId("graph")).toBeNull();
  });

  it("says there is nothing to draw without a commissioned node, and shows another failure with a retry", async () => {
    const g = body<TopologyGraph>("controller-topology.200");
    g.nodes = g.nodes.filter((n) => n.kind === "node");
    g.edges = [];
    render(Harness, { ui: testUI({ answers: { "GET /controller/topology": { status: 200, body: g } } }), page: Network });
    expect(await screen.findByText("controller.network.empty")).toBeInTheDocument();
    cleanup();
    const failing = testUI({ answers: { "GET /controller/topology": { status: 500, body: { code: "internal", title: "Internal", status: 500, type: "x", detail: "boom" } } } });
    render(Harness, { ui: failing, page: Network });
    expect(await screen.findByRole("button", { name: "common.reload" })).toBeInTheDocument();
  });
});

describe("Controller Nodes tabs", () => {
  it("switch between the list and the network", async () => {
    const ui = testUI({ fixtures: ["controller-nodes.200", "controller-topology.200"] });
    render(Harness, { ui, page: Nodes });
    await screen.findByText("Kitchen");
    const tabs = screen.getAllByRole("tab");
    expect(tabs.map((t) => [t.textContent?.trim(), t.getAttribute("href")])).toEqual([
      ["controller.nodes.tab.list", "#/controller/nodes"],
      ["controller.nodes.tab.network", "#/controller/nodes?tab=network"],
    ]);
    await fireEvent.click(tabs[1]);
    await screen.findByTestId("graph");
    expect(screen.queryByText("Kitchen")).toBeNull();
    await fireEvent.click(screen.getAllByRole("tab")[0]);
    expect(await screen.findByText("Kitchen")).toBeInTheDocument();
  });
});
