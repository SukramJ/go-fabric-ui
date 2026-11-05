// @vitest-environment happy-dom
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import NodeHarness from "../../../test/NodeHarness.svelte";
import { body, fixtureFetch, testUI } from "../../../test/support";
import { createClient, type ControllerNodeDetail, type EndpointDetail } from "../../api/client";

afterEach(cleanup);

const L = "0000000000000011";
const S = "0000000000000012";

function renderNode(nodeId: string, tab: string, fixtures: string[], extra: Parameters<typeof testUI>[0] = {}) {
  const ui = testUI({ fixtures: [nodeId === L ? "controller-node.200" : "controller-node.200-icd", ...fixtures], ...extra });
  render(NodeHarness, { ui, nodeId, tab });
  return ui;
}

describe("Controller Node", () => {
  it("shows the overview: name, hex id, connection, properties", async () => {
    renderNode(L, "overview", []);
    expect(await screen.findByTestId("node-title")).toHaveTextContent("Kitchen");
    const card = screen.getByTestId("node-overview");
    expect(card).toHaveTextContent("0x0000000000000011");
    expect(card).toHaveTextContent("controller.connection.connected");
    expect(screen.getByText("controller.overview.flag.supports_wifi")).toBeInTheDocument();
  });

  it("pings the node and says it measured a Matter read, per address", async () => {
    const ui = renderNode(L, "overview", ["controller-node-ping.200"]);
    await screen.findByTestId("node-overview");
    await fireEvent.click(screen.getByText("controller.ping.button"));
    const res = await screen.findByTestId("ping-result");
    expect(res).toHaveTextContent("controller.ping.method_per_address");
    expect(res).toHaveTextContent("controller.ping.reachable");
    expect(res).toHaveTextContent('controller.ping.rtt:{"ms":12.5}');
    const rows = screen.getAllByTestId("ping-address");
    expect(rows).toHaveLength(2);
    expect(rows[0]).toHaveTextContent("[fd00::11]:5540");
    expect(rows[0]).toHaveTextContent('controller.ping.rtt:{"ms":12.5}');
    expect(rows[0]).not.toHaveTextContent("controller.ping.unreachable");
    expect(rows[1]).toHaveTextContent("192.168.1.11:5540");
    expect(rows[1]).toHaveTextContent("controller.ping.unreachable");
    expect(rows[1]).toHaveTextContent("client: address unreachable: 192.168.1.11:5540");
    expect(ui.sent.find((r) => r.method === "POST")?.headers["X-Fabric-UI-Token"]).toBe("test-token");
  });

  it("lists the endpoints with device types and clusters", async () => {
    renderNode(L, "endpoints", []);
    const ep1 = await screen.findByTestId("endpoint-1");
    expect(ep1).toHaveTextContent("OnOffLight");
    expect(ep1).toHaveTextContent("OnOff");
    expect(ep1).toHaveTextContent("Binding");
  });

  it("reads an attribute of a chosen cluster", async () => {
    const ui = renderNode(L, "attributes", [], {
      answers: { [`POST /controller/nodes/${L}/attributes`]: { status: 200, body: { values: [{ endpoint: 0, cluster_hex: "0x0028", attribute_hex: "0x0005", value: "Kitchen" }] } } },
    });
    await screen.findByTestId("node-title");
    await fireEvent.change(screen.getByLabelText("controller.attributes.cluster"), { target: { value: "0x0028" } });
    await fireEvent.input(screen.getByLabelText("controller.attributes.attribute"), { target: { value: "5" } });
    await fireEvent.click(screen.getByText("controller.attributes.read"));
    expect(await screen.findByTestId("attribute-result")).toHaveTextContent('"Kitchen"');
    expect(ui.sent.find((s) => s.method === "POST")?.body).toEqual({ paths: [{ endpoint: 0, cluster_hex: "0x0028", attribute_hex: "0x0005" }] });
  });

  it("shows the ACL and asks before a list that locks this controller out", async () => {
    const ui = renderNode(L, "acl", ["controller-node-acl.200", "controller-node-acl-set.200"], { confirm: false });
    const list = await screen.findByTestId("acl-entries");
    expect(list).toHaveTextContent("controller.acl.privilege.administer");
    expect(list).toHaveTextContent("0x000000000001B669");
    await fireEvent.click(screen.getByText("common.remove"));
    await fireEvent.click(screen.getByText("common.save"));
    await waitFor(() => expect(ui.confirmSpy).toHaveBeenCalled());
    expect(ui.sent.some((s) => s.method === "PUT")).toBe(false);
  });

  it("adds an ACL entry and writes the list", async () => {
    const ui = renderNode(L, "acl", ["controller-node-acl.200", "controller-node-acl-set.200"]);
    await screen.findByTestId("acl-entries");
    await fireEvent.input(screen.getByLabelText("controller.acl.subjects"), { target: { value: "99" } });
    await fireEvent.input(screen.getByLabelText("controller.acl.target_cluster"), { target: { value: "6" } });
    await fireEvent.input(screen.getByLabelText("controller.acl.target_endpoint"), { target: { value: "1" } });
    await fireEvent.click(screen.getByText("controller.acl.add_entry"));
    await fireEvent.click(screen.getByText("common.save"));
    await waitFor(() => expect(ui.toastSpy.success).toHaveBeenCalledWith("controller.acl.saved"));
    const put = ui.sent.find((s) => s.method === "PUT");
    expect((put?.body as { entries: unknown[] }).entries[1]).toEqual({ privilege: "operate", auth_mode: "case", subjects_hex: ["0000000000000099"], targets: [{ cluster_hex: "0x0006", endpoint: 1 }] });
  });

  it("shows the bindings of the endpoint with a Binding cluster", async () => {
    renderNode(L, "bindings", ["controller-node-bindings.200"]);
    expect(await screen.findByTestId("bindings-1")).toHaveTextContent("0x0000000000000012");
  });

  it("opens a window and shows its codes and QR code", async () => {
    renderNode(L, "window", ["controller-node-window.200"]);
    await screen.findByTestId("node-title");
    await fireEvent.change(screen.getByLabelText("controller.window.timeout"), { target: { value: "300" } });
    await fireEvent.click(screen.getByText("controller.window.open"));
    const win = await screen.findByTestId("node-window");
    expect(win).toHaveTextContent("34970112332");
    expect(win.querySelector("svg")).not.toBeNull();
  });

  it("shows the ICD state of a sleepy node", async () => {
    renderNode(S, "icd", ["controller-node-icd.200"]);
    const card = await screen.findByTestId("node-icd");
    expect(card).toHaveTextContent("controller.icd.registered");
    expect(card).toHaveTextContent("controller.icd.mode.lit");
  });

  it("lists the node's fabrics, ours without a remove button", async () => {
    renderNode(L, "fabrics", ["controller-node-fabrics.200"]);
    expect(await screen.findByText("controller.fabrics.own")).toBeInTheDocument();
    expect(screen.getAllByText("common.remove")).toHaveLength(1);
  });

  it("shows the node's queued update", async () => {
    renderNode(L, "updates", ["controller-updates.200"]);
    const card = await screen.findByTestId("node-updates");
    await waitFor(() => expect(card).toHaveTextContent("controller.updates.status.queued"));
  });

  // The inventory answers endpoint 1 and OnOff from the fixtures, the root
  // endpoint from the same fixture renumbered.
  function renderInventory(extra: Parameters<typeof testUI>[0] = {}) {
    const root: EndpointDetail = { ...body<EndpointDetail>("controller-node-endpoint.200"), endpoint_id: 0, clusters: [], tags: [] };
    return renderNode(L, "inventory", ["controller-node-endpoint.200", "controller-node-cluster.200"], {
      ...extra,
      answers: { [`GET /controller/nodes/${L}/endpoints/0`]: { status: 200, body: root }, ...extra.answers },
    });
  }
  const clusterReads = (ui: ReturnType<typeof testUI>) => ui.sent.filter((s) => s.path.includes("/clusters/")).length;

  it("shows the inventory from the controller's image and loads a cluster when it is opened", async () => {
    const ui = renderInventory();
    const ep1 = await screen.findByTestId("inventory-endpoint-1");
    await waitFor(() => expect(ep1).toHaveTextContent("OnOffLight r3"));
    expect(ep1).toHaveTextContent("One");
    expect(ep1).toHaveTextContent("controller.inventory.conformance.clean");
    expect(ep1).toHaveTextContent("OnOff");
    expect(clusterReads(ui)).toBe(0);
    await fireEvent.click(screen.getByTestId("cluster-1-0x0006"));
    const view = await screen.findByTestId("cluster-view-0x0006");
    expect(view).toHaveTextContent("LT · Lighting");
    expect(view).toHaveTextContent("0x00000001");
    expect(view).toHaveTextContent('inventory.data_version:{"version":7}');
    expect(screen.getByTestId("attribute-0x0000")).toHaveTextContent("true");
    expect(view).toHaveTextContent("Toggle");
    expect(clusterReads(ui)).toBe(1);
    // Closed and opened again: one more request, never one for a closed cluster.
    await fireEvent.click(screen.getByTestId("cluster-1-0x0006"));
    expect(screen.queryByTestId("cluster-view-0x0006")).toBeNull();
    await fireEvent.click(screen.getByTestId("cluster-1-0x0006"));
    await screen.findByTestId("cluster-view-0x0006");
    expect(clusterReads(ui)).toBe(2);
  });

  it("reloads the inventory and the open cluster when the stream names the node", async () => {
    const ui = renderInventory();
    await fireEvent.click(await screen.findByTestId("cluster-1-0x0006"));
    await screen.findByTestId("cluster-view-0x0006");
    ui.controller.applyEvent({ type: "node_updated", data: { node_id_hex: L } });
    await waitFor(() => expect(clusterReads(ui)).toBe(2));
  });

  it("shows the error state for an endpoint or a cluster the image lacks", async () => {
    renderInventory({
      answers: {
        [`GET /controller/nodes/${L}/endpoints/0`]: { status: 404, body: body("controller-node-endpoint.404") },
        [`GET /controller/nodes/${L}/endpoints/1/clusters/0x0003`]: { status: 404, body: body("controller-node-cluster.404") },
      },
    });
    const ep0 = await screen.findByTestId("inventory-endpoint-0");
    await waitFor(() => expect(ep0).toHaveTextContent("common.error"));
    await fireEvent.click(await screen.findByTestId("cluster-1-0x0003"));
    const ep1 = screen.getByTestId("inventory-endpoint-1");
    await waitFor(() => expect(ep1).toHaveTextContent("common.error"));
    expect(ep1).toHaveTextContent("no such cluster");
  });

  it("keeps the newest endpoint detail when an older reload answers last", async () => {
    const client = createClient("/api/v1", fixtureFetch(["controller-node.200"]).fetch);
    const answers: ((d: EndpointDetail) => void)[] = [];
    const ep = body<EndpointDetail>("controller-node-endpoint.200");
    client.controllerEndpointDetail = (_id, n) => (n === 1 ? new Promise((resolve) => answers.push(resolve)) : Promise.resolve({ ...ep, endpoint_id: 0 }));
    const ui = renderNode(L, "inventory", [], { client });
    const ep1 = await screen.findByTestId("inventory-endpoint-1");
    await waitFor(() => expect(answers).toHaveLength(1));
    ui.controller.applyEvent({ type: "node_updated", data: { node_id_hex: L } });
    await waitFor(() => expect(answers).toHaveLength(2));
    answers[1]({ ...ep, name: "Newer" });
    await waitFor(() => expect(ep1).toHaveTextContent("Newer"));
    answers[0]({ ...ep, name: "Older" });
    await new Promise((r) => setTimeout(r, 20));
    expect(ep1).toHaveTextContent("Newer");
    expect(ep1).not.toHaveTextContent("Older");
  });

  it("shows Reachable on a bridged endpoint and opens the inventory at an endpoint", async () => {
    const node = body<ControllerNodeDetail>("controller-node.200");
    node.endpoints[1].reachable = false;
    renderNode(L, "endpoints", ["controller-node-endpoint.200"], { answers: { [`GET /controller/nodes/${L}`]: { status: 200, body: node } } });
    const ep1 = await screen.findByTestId("endpoint-1");
    expect(ep1).toHaveTextContent("controller.endpoints.unreachable");
    expect(screen.getByTestId("endpoint-0")).not.toHaveTextContent("controller.endpoints.reachable");
    await fireEvent.click(screen.getByTestId("inventory-link-1"));
    expect(await screen.findByTestId("inventory-endpoint-1")).toBeInTheDocument();
  });

  it("lists the events the controller received, newest first, and fetches the newer ones on node_updated", async () => {
    const ui = renderNode(L, "events", ["controller-node-events.200"], {
      answers: {
        [`GET /controller/nodes/${L}/events?since=43`]: {
          status: 200,
          body: { events: [{ number: 44, priority: "info", endpoint: 1, cluster_hex: "0x0039", event_hex: "0x0003", name: "ReachableChanged", data: { reachableNewValue: false } }] },
        },
      },
    });
    const list = await screen.findByTestId("event-list");
    expect([...list.querySelectorAll("li")].map((li) => li.getAttribute("data-testid"))).toEqual(["event-42", "event-41"]);
    expect(screen.getByTestId("event-42")).toHaveTextContent("StartUp");
    expect(screen.getByTestId("event-42")).toHaveTextContent('"softwareVersion": 1');
    expect(screen.getByTestId("event-41")).toHaveTextContent("events.priority.critical");
    ui.controller.applyEvent({ type: "node_updated", data: { node_id_hex: L } });
    expect(await screen.findByTestId("event-44")).toHaveTextContent("ReachableChanged");
    expect(ui.sent.map((s) => s.path)).toContain(`/api/v1/controller/nodes/${L}/events?since=43`);
  });

  it("switches tabs", async () => {
    renderNode(L, "overview", []);
    await screen.findByTestId("node-overview");
    await fireEvent.click(screen.getByText("controller.node.tab.endpoints"));
    expect(await screen.findByTestId("endpoint-0")).toBeInTheDocument();
  });
});
