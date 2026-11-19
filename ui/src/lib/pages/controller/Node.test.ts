// @vitest-environment happy-dom
import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import NodeHarness from "../../../test/NodeHarness.svelte";
import { body, fixtureFetch, testUI } from "../../../test/support";
import { createClient, type ControllerNodeDetail, type EndpointDetail } from "../../api/client";
import { createTranslator } from "../../i18n";

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

  it("shows the operational host when the record has one, and nothing without it", async () => {
    renderNode(L, "overview", []);
    expect(await screen.findByTestId("node-operational-host")).toHaveTextContent("0123456789abcdef.local");
    expect(screen.getByText("controller.overview.operational_host")).toBeInTheDocument();
    cleanup();
    renderNode(S, "overview", []);
    expect(await screen.findByTestId("node-overview")).toBeInTheDocument();
    expect(screen.queryByTestId("node-operational-host")).toBeNull();
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

  // UR-P4-ui-234: the QR encoder refuses an empty payload and one over its
  // capacity; the tab shows the codes without the picture.
  for (const [name, qr] of [
    ["an empty", ""],
    ["an over-long", "x".repeat(3000)],
  ] as const) {
    it(`shows the codes without a picture for ${name} QR payload`, async () => {
      const answer = { ...body<Record<string, unknown>>("controller-node-window.200"), qr_code: qr };
      renderNode(L, "window", ["controller-node-window.200"], {
        answers: { [`POST /controller/nodes/${L}/window`]: { status: 200, body: answer } },
      });
      await screen.findByTestId("node-title");
      await fireEvent.click(screen.getByText("controller.window.open"));
      const win = await screen.findByTestId("node-window");
      expect(win).toHaveTextContent("34970112332");
      expect(win.querySelector("svg")).toBeNull();
    });
  }

  it("shows the ICD state of a sleepy node", async () => {
    renderNode(S, "icd", ["controller-node-icd.200"]);
    const card = await screen.findByTestId("node-icd");
    expect(card).toHaveTextContent("controller.icd.registered");
    expect(card).toHaveTextContent("controller.icd.mode.lit");
  });

  // UR-P4-ui-291, decision 76: only the multi-admin refusal (409
  // icd_multi_admin, 0.12.0) asks; any other 409 is an error toast.
  const ICD_IDLE = { available: true, awake: false, lit_supported: true, operating_mode: "lit", registered: false, supported: true };
  const registerAnswer = (code: string) => ({
    [`GET /controller/nodes/${S}/icd`]: { status: 200, body: ICD_IDLE },
    [`POST /controller/nodes/${S}/icd/register`]: {
      status: 409,
      body: { code, status: 409, title: "Refused by the node", type: "x", detail: `fabricui: ${code}` },
    },
  });

  it("asks the multi-admin question on icd_multi_admin and registers again with allow_multi_admin", async () => {
    const ui = renderNode(S, "icd", [], { answers: registerAnswer("icd_multi_admin"), confirm: true });
    await screen.findByTestId("node-icd");
    await fireEvent.click(screen.getByText("controller.icd.register"));
    await waitFor(() => expect(ui.sent.filter((s) => s.path.endsWith("/icd/register"))).toHaveLength(2));
    expect(ui.confirmSpy).toHaveBeenCalledWith(expect.objectContaining({ title: "controller.icd.multi_admin" }));
    expect(ui.sent.filter((s) => s.path.endsWith("/icd/register"))[1].body).toEqual({ allow_multi_admin: true });
  });

  it("does not ask the multi-admin question for a conflict of the client's own state", async () => {
    const ui = renderNode(S, "icd", [], { answers: registerAnswer("conflict"), confirm: true });
    await screen.findByTestId("node-icd");
    await fireEvent.click(screen.getByText("controller.icd.register"));
    await waitFor(() => expect(ui.toastSpy.error).toHaveBeenCalled());
    expect(ui.confirmSpy).not.toHaveBeenCalledWith(expect.objectContaining({ title: "controller.icd.multi_admin" }));
    expect(ui.sent.filter((s) => s.path.endsWith("/icd/register"))).toHaveLength(1);
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
        [`GET /controller/nodes/${L}/events?since_str=43`]: {
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
    expect(ui.sent.map((s) => s.path)).toContain(`/api/v1/controller/nodes/${L}/events?since_str=43`);
  });

  it("switches tabs", async () => {
    renderNode(L, "overview", []);
    await screen.findByTestId("node-overview");
    await fireEvent.click(screen.getByText("controller.node.tab.endpoints"));
    expect(await screen.findByTestId("endpoint-0")).toBeInTheDocument();
  });
});

describe("Controller Node — a new nodeId and overlapping loads", () => {
  it("drops the previous node's tab state when nodeId changes in place (UR-P1-ui-163)", async () => {
    const acl12 = { entries: [{ auth_mode: "case", fabric_index: 2, privilege: "administer", subjects_hex: ["00000000000000AA"], targets: [] }] };
    const ui = testUI({
      fixtures: ["controller-node.200", "controller-node.200-icd", "controller-node-acl.200"],
      answers: { [`GET /controller/nodes/${S}/acl`]: { status: 200, body: acl12 } },
    });
    const r = render(NodeHarness, { ui, nodeId: L, tab: "acl" });
    expect(await screen.findByTestId("acl-entries")).toHaveTextContent("000000000001B669");
    await r.rerender({ ui, nodeId: S, tab: "acl" });
    expect(screen.queryByText("Kitchen")).toBeNull();
    await waitFor(() => expect(screen.getByTestId("acl-entries")).toHaveTextContent("00000000000000AA"));
    expect(screen.getByTestId("acl-entries")).not.toHaveTextContent("000000000001B669");
    expect(screen.getByTestId("node-title")).toHaveTextContent("Contact Sensor");
  });

  it("shows the newer detail when an older reload answers last (UR-P1-conc-180)", async () => {
    const ui = testUI({ fixtures: ["controller-node.200"] });
    const first = body<ControllerNodeDetail>("controller-node.200");
    let calls = 0;
    let release: (() => void) | undefined;
    ui.client.controllerNode = async () => {
      calls++;
      if (calls === 1) return structuredClone(first);
      if (calls === 2) return new Promise((resolve) => (release = () => resolve({ ...first, node: { ...first.node, node_label: "Old" } })));
      return { ...first, node: { ...first.node, node_label: "New" } };
    };
    render(NodeHarness, { ui, nodeId: L, tab: "overview" });
    await screen.findByTestId("node-title");
    ui.controller.applyEvent({ type: "node_updated", data: { node_id_hex: L } } as never);
    await waitFor(() => expect(release).toBeDefined());
    ui.controller.applyEvent({ type: "node_updated", data: { node_id_hex: L } } as never);
    await waitFor(() => expect(screen.getByTestId("node-title")).toHaveTextContent("New"));
    release!();
    await new Promise((r) => setTimeout(r, 20));
    expect(screen.getByTestId("node-title")).toHaveTextContent("New");
  });
});

describe("Controller Node — ACL and binding input", () => {
  it("refuses an ACL target endpoint that is not a number in range (UR-P1-ui-164)", async () => {
    const ui = renderNode(L, "acl", ["controller-node-acl.200", "controller-node-acl-set.200"]);
    await screen.findByTestId("acl-entries");
    for (const bad of ["abc", "1.5", "-1", "70000"]) {
      await fireEvent.input(screen.getByLabelText("controller.acl.subjects"), { target: { value: "99" } });
      await fireEvent.input(screen.getByLabelText("controller.acl.target_endpoint"), { target: { value: bad } });
      await fireEvent.click(screen.getByText("controller.acl.add_entry"));
    }
    expect(ui.toastSpy.error).toHaveBeenCalledTimes(4);
    expect(ui.toastSpy.error).toHaveBeenCalledWith("controller.acl.invalid_target");
    expect(screen.getByTestId("acl-entries")).not.toHaveTextContent("0000000000000099");
  });

  it("refuses a binding endpoint or group out of range (UR-P1-ui-164)", async () => {
    const ui = renderNode(L, "bindings", ["controller-node-bindings.200", "controller-node-bindings-set.200"]);
    await screen.findByTestId("bindings-1");
    const target = screen.getByLabelText("controller.bindings.target");
    const endpoint = screen.getByLabelText("controller.bindings.endpoint");
    for (const [tg, ep] of [["12", "abc"], ["12", "70000"], ["g:70000", ""]]) {
      await fireEvent.input(target, { target: { value: tg } });
      await fireEvent.input(endpoint, { target: { value: ep } });
      await fireEvent.click(screen.getByText("controller.bindings.add"));
    }
    expect(ui.toastSpy.error).toHaveBeenCalledTimes(3);
    expect(ui.sent.some((s) => s.method === "PUT")).toBe(false);
  });

  it("sends one binding write at a time (UR-P1-conc-165)", async () => {
    const two = {
      endpoints: [
        {
          endpoint: 1,
          bindings: [
            { cluster_hex: "0x0006", endpoint: 1, fabric_index: 1, node_id_hex: "0000000000000012" },
            { cluster_hex: "0x0006", endpoint: 2, fabric_index: 1, node_id_hex: "0000000000000013" },
          ],
        },
      ],
    };
    const ui = renderNode(L, "bindings", ["controller-node-bindings-set.200"], { answers: { [`GET /controller/nodes/${L}/bindings`]: { status: 200, body: two } } });
    await screen.findByTestId("bindings-1");
    const orig = ui.client.setNodeBindings.bind(ui.client);
    let release: (() => void) | undefined;
    let puts = 0;
    ui.client.setNodeBindings = (...args) => {
      puts++;
      return new Promise((resolve) => (release = () => resolve(orig(...args))));
    };
    const removes = screen.getAllByRole("button", { name: "common.remove" });
    await fireEvent.click(removes[0]);
    await fireEvent.click(removes[1]);
    expect(puts).toBe(1);
    release!();
    await waitFor(() => expect(ui.toastSpy.success).toHaveBeenCalledWith("controller.bindings.saved"));
  });
});

describe("Controller Node — the window tab (UR-P1-ui-166)", () => {
  it("offers Close after the tab was left and shown again", async () => {
    const ui = renderNode(L, "window", ["controller-node-window.200", "controller-node-window-close.204"]);
    await screen.findByTestId("node-title");
    await fireEvent.click(screen.getByText("controller.window.open"));
    await screen.findByTestId("node-window");
    await fireEvent.click(screen.getByText("controller.node.tab.overview"));
    await fireEvent.click(screen.getByText("controller.node.tab.window"));
    await fireEvent.click(await screen.findByText("controller.window.close"));
    await waitFor(() => expect(ui.toastSpy.success).toHaveBeenCalledWith("controller.window.closed"));
    expect(ui.sent.some((s) => s.method === "POST" && s.path.endsWith("/window/close"))).toBe(true);
  });

  it("takes the codes off the page when the window's time is up", async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    renderNode(L, "window", ["controller-node-window.200"]);
    await screen.findByTestId("node-title");
    await fireEvent.click(screen.getByText("controller.window.open"));
    expect(await screen.findByTestId("node-window")).toHaveTextContent("34970112332");
    await vi.advanceTimersByTimeAsync(301_000);
    await waitFor(() => expect(screen.queryByTestId("node-window")).toBeNull());
    expect(screen.getByText("controller.window.expired")).toBeInTheDocument();
    vi.useRealTimers();
  });
});

describe("Controller Node — the updates tab (UR-P1-ui-167)", () => {
  it("does not claim an empty queue when the queue could not be read", async () => {
    const ui = renderNode(L, "updates", [], { answers: { "GET /controller/updates": { status: 500, body: { code: "internal", title: "boom", status: 500, type: "x" } } } });
    await screen.findByTestId("node-updates");
    await waitFor(() => expect(ui.toastSpy.error).toHaveBeenCalled());
    expect(await screen.findByText("controller.updates.queue_unknown")).toBeInTheDocument();
    expect(screen.queryByText("controller.updates.queue_empty")).toBeNull();
  });

  // Decision 70: one node's failed check no longer fails the answer; the
  // check lists what was found and, per node, why a check failed.
  const checked = (found: string, failed: string) => ({
    "POST /controller/updates/check": {
      status: 200,
      body: {
        updates: [{ node_id_hex: found, vendor_id: 65521, product_id: 32769, software_version: 2, software_version_string: "1.1.0" }],
        errors: [{ node_id_hex: failed, error: "download failed" }],
      },
    },
  });

  it("offers the update found for this node while another node's check failed (decision 70)", async () => {
    const ui = renderNode(L, "updates", ["controller-updates.200"], { answers: checked(L, S) });
    await screen.findByTestId("node-updates");
    await fireEvent.click(screen.getByText("controller.updates.check"));
    expect(await screen.findByText('controller.updates.consent:{"version":"1.1.0"}')).toBeInTheDocument();
    expect(screen.queryByTestId("updates-check-error")).toBeNull();
    expect(ui.toastSpy.error).not.toHaveBeenCalled();
  });

  it("shows this node's failed check where its updates would be (decision 70)", async () => {
    const ui = renderNode(L, "updates", ["controller-updates.200"], { answers: checked(S, L) });
    await screen.findByTestId("node-updates");
    await fireEvent.click(screen.getByText("controller.updates.check"));
    expect(await screen.findByTestId("updates-check-error")).toHaveTextContent('controller.updates.check_failed:{"error":"download failed"}');
    expect(screen.queryByText('controller.updates.consent:{"version":"1.1.0"}')).toBeNull();
    // Not "no update available": the check did not answer for this node.
    expect(ui.toastSpy.info).not.toHaveBeenCalled();
  });
});

describe("Controller Node — forgetting an ICD registration (UR-P1-ui-176)", () => {
  it("asks first, and sends nothing when the answer is no", async () => {
    const ui = renderNode(S, "icd", ["controller-node-icd.200", "controller-node-icd-unregister.200"], { confirm: false });
    await screen.findByTestId("node-icd");
    await fireEvent.click(screen.getByText("controller.icd.forget"));
    await waitFor(() => expect(ui.confirmSpy).toHaveBeenCalledWith(expect.objectContaining({ destructive: true, title: "controller.icd.forget_confirm" })));
    expect(ui.sent.some((s) => s.method === "POST" && s.path.endsWith("/icd/unregister"))).toBe(false);
  });
});

describe("Controller Node — dates in the UI's language (UR-P1-ui-182)", () => {
  // Both languages: the process's own default locale (the machine's) is
  // one of them, so on any machine one case fails without the fix.
  it("formats the commissioning time in the UI's language, not the runtime's", async () => {
    renderNode(L, "overview", [], { t: createTranslator(() => "de") });
    expect(await screen.findByTestId("node-overview")).toHaveTextContent(/20\.10\.2026/);
    cleanup();
    renderNode(L, "overview", [], { t: createTranslator(() => "en") });
    expect(await screen.findByTestId("node-overview")).toHaveTextContent(/10\/20\/2026/);
  });
});

describe("Controller Node — the ACL refusal before the lock-out question (UR-P4-ui-66)", () => {
  it("refuses a list with an unknown entry without asking about a lock-out first", async () => {
    const ui = renderNode(L, "acl", [], {
      answers: {
        [`GET /controller/nodes/${L}/acl`]: { status: 200, body: { entries: [{ privilege: "unknown", auth_mode: "pase", subjects_hex: [], targets: [] }] } },
      },
    });
    await screen.findByTestId("acl-entries");
    await fireEvent.input(screen.getByLabelText("controller.acl.subjects"), { target: { value: "1" } });
    await fireEvent.click(screen.getByText("controller.acl.add_entry"));
    await fireEvent.click(screen.getByText("common.save"));
    await waitFor(() => expect(ui.toastSpy.error).toHaveBeenCalledWith("controller.acl.unknown_entry"));
    expect(ui.confirmSpy).not.toHaveBeenCalled();
    expect(ui.sent.some((s) => s.method === "PUT")).toBe(false);
  });
});
