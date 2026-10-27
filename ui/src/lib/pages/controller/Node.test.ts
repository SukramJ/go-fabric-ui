// @vitest-environment happy-dom
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import NodeHarness from "../../../test/NodeHarness.svelte";
import { testUI } from "../../../test/support";

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

  it("switches tabs", async () => {
    renderNode(L, "overview", []);
    await screen.findByTestId("node-overview");
    await fireEvent.click(screen.getByText("controller.node.tab.endpoints"));
    expect(await screen.findByTestId("endpoint-0")).toBeInTheDocument();
  });
});
