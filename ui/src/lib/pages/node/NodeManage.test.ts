// @vitest-environment happy-dom
// The node registry on the status page (contract 0.6.0): the read-only
// configuration block of every node, and a managed node's rename and
// delete with two confirmations naming its label and fabric count.
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { testUI, type TestUI } from "../../../test/support";
import { createNodeStore } from "../../store.svelte";
import Status from "./Status.svelte";

afterEach(cleanup);

// kitchen is the registry node of the fixtures, on a context of its own as
// the shell's NodeScope builds it.
async function kitchen(ui: TestUI) {
  const client = ui.client.forNode("kitchen");
  const scoped = { ...ui, client, node: createNodeStore(client, { t: ui.t, toast: ui.toast, subscribe: () => () => {} }) };
  await scoped.node.loadStatus();
  await ui.nodes.load();
  render(Harness, { ui: scoped, page: Status });
  return scoped;
}

describe("the configuration block", () => {
  it("shows a configured node's values with the host's note, never a passcode", async () => {
    render(Harness, { ui: testUI({ fixtures: ["node-status.200-config"] }), page: Status });
    const block = await screen.findByTestId("node-config");
    expect(screen.getByTestId("node-config-source")).toHaveTextContent('node.config.from_configuration_note:{"note":"the --bridge-* flags"}');
    for (const text of ["Main bridge", "[::]:5540", "3840", "go-fabric (0xFFF1)", "Bridge (0x8000)", "bridge.db"]) expect(block).toHaveTextContent(text);
    expect(block.textContent).not.toMatch(/passcode/i);
    // The default node is configuration: no manage card.
    expect(screen.queryByTestId("node-manage")).toBeNull();
  });

  it("is absent when the host describes no configuration", async () => {
    render(Harness, { ui: testUI({ fixtures: ["node-status.200"] }), page: Status });
    await screen.findByTestId("node-status");
    expect(screen.queryByTestId("node-config")).toBeNull();
  });
});

describe("a managed node", () => {
  it("says the host allocated its values and renames it", async () => {
    const ui = testUI({ fixtures: ["nodes-status.200-registry", "nodes.200-registry", "nodes-label.204"] });
    await kitchen(ui);
    expect(await screen.findByTestId("node-config-source")).toHaveTextContent("node.config.from_registry");
    const field = (await screen.findByTestId("node-manage-label")) as HTMLInputElement;
    expect(field.value).toBe("Kitchen");
    await fireEvent.input(field, { target: { value: "Kitchen and dining" } });
    await fireEvent.click(screen.getByText("node.manage.rename"));
    await waitFor(() => expect(ui.sent.some((r) => r.method === "PUT")).toBe(true));
    const put = ui.sent.find((r) => r.method === "PUT")!;
    expect(put.path).toBe("/api/v1/nodes/kitchen/label");
    expect(put.body).toEqual({ label: "Kitchen and dining" });
    expect(screen.getByTestId("node-manage-hint")).toHaveTextContent("node.manage.hint.alexa");
  });

  it("deletes only after two confirmations that name the label and the fabric count", async () => {
    const ui = testUI({ fixtures: ["nodes-status.200-registry", "nodes.200-registry"], answers: { "DELETE /nodes/kitchen": { status: 204 } } });
    await kitchen(ui);
    await fireEvent.click(await screen.findByTestId("node-manage-delete"));
    await waitFor(() => expect(ui.sent.some((r) => r.method === "DELETE")).toBe(true));
    expect(ui.confirmSpy).toHaveBeenCalledTimes(2);
    const [first, second] = ui.confirmSpy.mock.calls.map((c) => (c as unknown as [{ title: string; body: string }])[0]);
    expect(first.title).toBe('node.manage.delete_title:{"label":"Kitchen"}');
    expect(first.body).toBe('node.manage.delete_body:{"label":"Kitchen","count":1}');
    expect(second.body).toBe('node.manage.delete_again_body:{"count":1}');
    await waitFor(() => expect(ui.toastSpy.success).toHaveBeenCalledWith('node.manage.deleted:{"label":"Kitchen"}'));
  });

  it("deletes nothing when the second confirmation is declined", async () => {
    const ui = testUI({ fixtures: ["nodes-status.200-registry", "nodes.200-registry"] });
    let n = 0;
    ui.confirmSpy.mockImplementation(async () => ++n === 1);
    await kitchen(ui);
    await fireEvent.click(await screen.findByTestId("node-manage-delete"));
    await waitFor(() => expect(ui.confirmSpy).toHaveBeenCalledTimes(2));
    expect(ui.sent.some((r) => r.method === "DELETE")).toBe(false);
  });
});

describe("the node list", () => {
  it("reloads on the stream's nodes_changed", async () => {
    const ui = testUI({ fixtures: ["nodes.200-registry"] });
    ui.node.applyEvent({ type: "nodes_changed", data: { keys: ["kitchen"] } });
    await waitFor(() => expect(ui.nodes.list.map((n) => n.key)).toEqual(["default", "kitchen"]));
    expect(ui.nodes.find("kitchen")?.managed).toBe(true);
  });

  it("creates a node and lists it before it answers", async () => {
    const ui = testUI({ fixtures: ["nodes-create.201"], answers: { "GET /nodes": { status: 200, body: { nodes: [{ key: "living-room", label: "Living room", enabled: true, managed: true, exposure: true }] } } } });
    const created = await ui.nodes.create("Living room");
    expect(created.key).toBe("living-room");
    expect(ui.nodes.find("living-room")).toBeDefined();
  });
});
