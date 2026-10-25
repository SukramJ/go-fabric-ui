// @vitest-environment happy-dom
// Ported from openccu-loom assets/ui/src/routes/matter/MatterFabrics.test.ts
// and MatterMaintenance.test.ts at b40b289a, on the contract fixtures.
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { body, testUI } from "../../../test/support";
import type { FabricList } from "../../api/client";
import Fabrics from "./Fabrics.svelte";

afterEach(cleanup);

function renderWith(fabrics: FabricList["fabrics"] | null, extra: Parameters<typeof testUI>[0] = {}) {
  const ui = testUI({
    fixtures: ["node-fabrics.200", "node-status.200"],
    ...extra,
    answers: { ...(fabrics ? { "GET /node/fabrics": { status: 200, body: { fabrics } } } : {}), ...extra.answers },
  });
  render(Harness, { ui, page: Fabrics });
  return ui;
}

const fixtureFabrics = () => body<FabricList>("node-fabrics.200").fabrics;

describe("Fabrics — node id column", () => {
  it("renders the node id as padded hex, not a decimal behind a 0x", async () => {
    renderWith(null);
    expect(await screen.findByText("0xB2C3D4E5F6071829")).toBeInTheDocument();
    expect(screen.getByText("0x0000000000000042")).toBeInTheDocument();
  });

  it("marks a node id that arrived rounded because only the numeric field was served", async () => {
    const [f] = fixtureFabrics();
    renderWith([{ ...f, node_id_hex: "", node_id: 2 ** 60 }]);
    expect(await screen.findByText("node.fabrics.node_id_rounded")).toBeInTheDocument();
  });

  it("does not mark an id that fits the transported precision", async () => {
    const [f] = fixtureFabrics();
    renderWith([{ ...f, node_id_hex: "", node_id: 66 }]);
    expect(await screen.findByText("0x0000000000000042")).toBeInTheDocument();
    expect(screen.queryByText("node.fabrics.node_id_rounded")).toBeNull();
  });

  it("renders a large id from the hex field and does not flag it", async () => {
    renderWith(null);
    await screen.findByText("0xB2C3D4E5F6071829");
    expect(screen.queryByText("node.fabrics.node_id_rounded")).toBeNull();
  });
});

describe("Fabrics — vendor column", () => {
  it("renders the name the server supplies", async () => {
    renderWith(null);
    const names = fixtureFabrics().map((f) => f.vendor_name);
    for (const n of names) expect(await screen.findByText(n)).toBeInTheDocument();
  });

  it("falls back to the raw id when the server sends no name", async () => {
    const [f] = fixtureFabrics();
    renderWith([{ ...f, vendor_name: "", vendor_id: 0xfff1 }]);
    expect(await screen.findByText("0xFFF1")).toBeInTheDocument();
  });
});

describe("Fabrics — adding a controller", () => {
  it("sends the operator to the pairing page instead of opening a second window here", async () => {
    renderWith(null);
    expect(await screen.findByTestId("share-link")).toHaveAttribute("href", "#/node/pair");
  });
});

describe("Fabrics — removing one", () => {
  it("asks, removes it through the client and reloads", async () => {
    const ui = renderWith(null, { fixtures: ["node-fabrics.200", "node-status.200", "node-fabric-remove.204"] });
    await screen.findByText("0x0000000000000042");
    const buttons = screen.getAllByRole("button", { name: "common.remove" });
    await fireEvent.click(buttons[buttons.length - 1]);
    await waitFor(() => expect(ui.sent.some((s) => s.method === "DELETE")).toBe(true));
    expect(ui.confirmSpy).toHaveBeenCalledWith(expect.objectContaining({ destructive: true }));
    expect(ui.toastSpy.success).toHaveBeenCalledWith("node.fabric.unpaired");
  });
});

describe("Fabrics — maintenance actions", () => {
  it("reassembles without asking, because nothing is lost by it", async () => {
    const ui = renderWith(null, { fixtures: ["node-fabrics.200", "node-status.200", "node-maintenance-reassemble.204"] });
    await fireEvent.click(await screen.findByText("node.maint.reassemble"));
    await waitFor(() => expect(ui.toastSpy.success).toHaveBeenCalledWith("node.maint.reassemble_done"));
    expect(ui.confirmSpy).not.toHaveBeenCalled();
  });

  it("surfaces a failed reassembly instead of reporting success", async () => {
    const ui = renderWith(null, { fixtures: ["node-fabrics.200", "node-status.200", "node-maintenance-reassemble.503"] });
    await fireEvent.click(await screen.findByText("node.maint.reassemble"));
    await waitFor(() => expect(ui.toastSpy.error).toHaveBeenCalled());
    expect(ui.toastSpy.success).not.toHaveBeenCalled();
  });

  it("does not remove any pairing when the confirm dialog is declined", async () => {
    const ui = renderWith(null, { confirm: false });
    await fireEvent.click(await screen.findByText("node.maint.reset"));
    await waitFor(() => expect(ui.confirmSpy).toHaveBeenCalled());
    expect(ui.sent.some((s) => s.path.endsWith("/factory-reset"))).toBe(false);
  });

  it("asks destructively and removes every pairing once confirmed", async () => {
    const ui = renderWith(null, { fixtures: ["node-fabrics.200", "node-status.200", "node-maintenance-factory-reset.204"] });
    await fireEvent.click(await screen.findByText("node.maint.reset"));
    await waitFor(() => expect(ui.toastSpy.success).toHaveBeenCalledWith("node.maint.reset_done"));
    expect(ui.confirmSpy).toHaveBeenCalledWith(expect.objectContaining({ destructive: true, title: "node.maint.reset_confirm" }));
    expect(ui.sent.find((s) => s.path.endsWith("/factory-reset"))?.body).toEqual({ confirm: "remove-all-fabrics" });
  });

  it("surfaces a partial reset, which leaves the node paired to something", async () => {
    const ui = renderWith(null, {
      answers: {
        "POST /node/maintenance/factory-reset": {
          status: 500,
          body: { type: "x", title: "Factory reset incomplete: removed 1 of 2 fabrics", status: 500, code: "internal" },
        },
      },
    });
    await fireEvent.click(await screen.findByText("node.maint.reset"));
    await waitFor(() => expect(ui.toastSpy.error).toHaveBeenCalledWith("Factory reset incomplete: removed 1 of 2 fabrics"));
  });
});
