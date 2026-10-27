// @vitest-environment happy-dom
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/svelte";
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

  it("interviews a node", async () => {
    const ui = testUI({ fixtures: ["controller-nodes.200", "controller-node-interview.204"] });
    render(Harness, { ui, page: Nodes });
    await screen.findByText("Kitchen");
    const buttons = screen.getAllByText("controller.nodes.interview");
    await fireEvent.click(buttons[1]);
    await waitFor(() => expect(ui.toastSpy.success).toHaveBeenCalledWith("controller.nodes.interviewed"));
  });
});
