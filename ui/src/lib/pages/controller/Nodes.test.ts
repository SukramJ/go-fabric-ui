// @vitest-environment happy-dom
import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { testUI } from "../../../test/support";
import Nodes from "./Nodes.svelte";

afterEach(cleanup);

describe("Controller Nodes", () => {
  it("lists the nodes with name, vendor, hex id and connection", async () => {
    render(Harness, { ui: testUI({ fixtures: ["controller-nodes.200"] }), page: Nodes });
    expect(await screen.findByText("Kitchen")).toBeInTheDocument();
    expect(screen.getByText("Contact Sensor")).toBeInTheDocument();
    expect(screen.getByText("0x0000000000000011")).toBeInTheDocument();
    expect(screen.getAllByText("controller.connection.connected").length).toBeGreaterThan(0);
    expect(screen.getAllByText("controller.nodes.icd").length).toBe(1);
  });

  it("filters by connection", async () => {
    render(Harness, { ui: testUI({ fixtures: ["controller-nodes.200"] }), page: Nodes });
    await screen.findByText("Kitchen");
    await fireEvent.change(screen.getByLabelText("controller.nodes.filter_connection"), { target: { value: "waiting_for_discovery" } });
    await waitFor(() => expect(screen.queryByText("Kitchen")).toBeNull());
    expect(screen.getByText("Contact Sensor")).toBeInTheDocument();
  });

  it("decommissions after the confirmation and reloads", async () => {
    const ui = testUI({ fixtures: ["controller-nodes.200", "controller-node-decommission.204"] });
    render(Harness, { ui, page: Nodes });
    await screen.findByText("Contact Sensor");
    const remove = screen.getAllByText("common.remove");
    // Rows sort by name: Contact Sensor before Kitchen.
    await fireEvent.click(remove[0]);
    await waitFor(() => expect(ui.toastSpy.success).toHaveBeenCalledWith("controller.nodes.removed"));
    expect(ui.confirmSpy).toHaveBeenCalled();
    expect(ui.sent.some((s) => s.method === "DELETE" && s.path.endsWith("/controller/nodes/0000000000000012"))).toBe(true);
  });

  // A stored onboarding code goes with the node (0.13.0): the confirmation
  // says so only for a node that has one.
  it("warns of the stored onboarding code only for a node that has one", async () => {
    const ui = testUI({ fixtures: ["controller-nodes.200-onboarding"], confirm: false });
    render(Harness, { ui, page: Nodes });
    await screen.findByText("Kitchen");
    const bodies = async (row: string) => {
      ui.confirmSpy.mockClear();
      const tr = screen.getByText(row).closest("tr")!;
      await fireEvent.click(within(tr).getByText("common.remove"));
      return (ui.confirmSpy.mock.calls[0] as unknown as [{ body: string }])[0].body;
    };
    expect(await bodies("Contact Sensor")).toContain("controller.nodes.remove_confirm_code");
    expect(await bodies("Kitchen")).not.toContain("controller.nodes.remove_confirm_code");
    cleanup();
    const plain = testUI({ fixtures: ["controller-nodes.200"], confirm: false });
    render(Harness, { ui: plain, page: Nodes });
    await screen.findByText("Contact Sensor");
    await fireEvent.click(within(screen.getByText("Contact Sensor").closest("tr")!).getByText("common.remove"));
    expect((plain.confirmSpy.mock.calls[0] as unknown as [{ body: string }])[0].body).not.toContain("remove_confirm_code");
  });

  it("interviews a node", async () => {
    const ui = testUI({ fixtures: ["controller-nodes.200", "controller-node-interview.204"] });
    render(Harness, { ui, page: Nodes });
    await screen.findByText("Kitchen");
    const buttons = screen.getAllByText("controller.nodes.interview");
    await fireEvent.click(buttons[1]);
    await waitFor(() => expect(ui.toastSpy.success).toHaveBeenCalledWith("controller.nodes.interviewed"));
  });
});

describe("Controller Nodes — busy per node (UR-P1-ui-169)", () => {
  it("keeps a node's actions disabled while its call runs, whatever else runs", async () => {
    const ui = testUI({ fixtures: ["controller-nodes.200"] });
    const held: Record<string, () => void> = {};
    const calls: string[] = [];
    ui.client.interviewNode = (id: string) => {
      calls.push(id);
      return new Promise<void>((resolve) => (held[id] = resolve));
    };
    render(Harness, { ui, page: Nodes });
    await screen.findByText("Kitchen");
    const buttons = () => screen.getAllByRole("button", { name: "controller.nodes.interview" });
    await fireEvent.click(buttons()[0]);
    await fireEvent.click(buttons()[1]);
    await fireEvent.click(buttons()[0]);
    expect(calls).toHaveLength(2);
    expect(new Set(calls).size).toBe(2);
  });
});

describe("UR-P4-ui-292: a node row opens the node through the host's routes", () => {
  it("follows the host's controllerNode route, not a hash built from it", async () => {
    const ui = testUI({ fixtures: ["controller-nodes.200"] });
    ui.routes = { ...ui.routes, controllerNode: "/app/controller/node/" };
    const follow = vi.fn();
    location.hash = "";
    render(Harness, { ui, page: Nodes, pageProps: { follow } });
    await fireEvent.click(await screen.findByText("Kitchen"));
    expect(location.hash).toBe("");
    expect(follow).toHaveBeenCalledWith("/app/controller/node/0000000000000011");
  });

  it("without a follow, the default routes still lead to the node's hash", async () => {
    location.hash = "";
    render(Harness, { ui: testUI({ fixtures: ["controller-nodes.200"] }), page: Nodes });
    await fireEvent.click(await screen.findByText("Kitchen"));
    await waitFor(() => expect(location.hash).toBe("#/controller/node/0000000000000011"));
  });
});
