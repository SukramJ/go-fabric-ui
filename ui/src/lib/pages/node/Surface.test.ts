// @vitest-environment happy-dom
// The Surface page on the node-surface fixtures: go-fabric's surface.Report
// of the test node — the root endpoint and two bridged endpoints, a Warn, a
// Remove, a mandatory-kept Keep with a reason and two silent Keeps (one on
// the root), one withheld element and one announced change with two element
// changes.
import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { testUI } from "../../../test/support";
import Surface from "./Surface.svelte";

afterEach(cleanup);

const renderSurface = (fixtures = ["node-surface.200", "node-endpoints.200"]) => {
  const ui = testUI({ fixtures });
  render(Harness, { ui, page: Surface });
  return ui;
};

describe("Surface", () => {
  it("shows the head: level with its origin, profiles, extensions, assembled, the short digest", async () => {
    renderSurface();
    const head = await screen.findByTestId("surface-head");
    expect(head).toHaveTextContent('node.status.surface_level:{"level":"1.5.1","origin":"node.status.level_origin_explicit"}');
    expect(head).toHaveTextContent('node.status.surface_profiles:{"profiles":"apple-home, home-assistant"}');
    expect(head).toHaveTextContent("node.surface.assembled");
    const digest = within(head).getByText('node.surface.policies:{"digest":"4865d61a5fc2"}');
    expect(digest).toHaveAttribute("title", "4865d61a5fc2857785b6cd1f65dc06c3343840275b4ff4fdc49e9521fad2d342");
  });

  it("renders the two endpoints with device type and cluster names joined from the endpoints route", async () => {
    renderSurface();
    const eps = await screen.findAllByTestId("surface-endpoint");
    expect(eps).toHaveLength(2);
    expect(eps[0]).toHaveTextContent("lamp-1");
    expect(eps[0]).toHaveTextContent("On/Off Light");
    expect(eps[0]).toHaveTextContent("0x0006 OnOff");
    expect(eps[1]).toHaveTextContent("sensor-2");
  });

  it("renders the root endpoint first, titled as the root, with its cluster table and lists", async () => {
    renderSurface(["node-surface.200"]);
    const root = await screen.findByTestId("surface-root");
    const [lamp] = screen.getAllByTestId("surface-endpoint");
    expect(root.compareDocumentPosition(lamp) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(within(root).getAllByText("node.surface.root")[0]).toBeInTheDocument();
    expect(root).toHaveTextContent("0x0016");
    expect(within(root).getAllByRole("row")).toHaveLength(2);
    expect(root).toHaveTextContent("0x001D");
    expect(within(root).getByText("node.surface.lists")).toBeInTheDocument();
  });

  it("falls back to the ids when the endpoints route fails", async () => {
    renderSurface(["node-surface.200"]);
    const eps = await screen.findAllByTestId("surface-endpoint");
    expect(eps[0]).toHaveTextContent("0x0100");
    expect(eps[0]).not.toHaveTextContent("On/Off Light");
  });

  it("hides the silent keeps by default and shows the warn, the remove and the reasoned keep", async () => {
    renderSurface();
    const card = await screen.findByTestId("surface-decisions");
    const rows = within(card).getAllByRole("row").slice(1);
    expect(rows).toHaveLength(3);
    expect(rows[0]).toHaveTextContent("node.surface.verdict.warn");
    expect(rows[1]).toHaveTextContent("node.surface.verdict.remove");
    expect(rows[2]).toHaveTextContent(/kept: mandatory at pin/);
    expect(within(card).queryByText("level")).toBeNull();
  });

  it("filters the decisions by verdict and by endpoint", async () => {
    renderSurface();
    const card = await screen.findByTestId("surface-decisions");
    const verdict = within(card).getByLabelText("node.surface.filter_verdict");
    const endpoint = within(card).getByLabelText("node.surface.filter_endpoint");
    const rows = () => within(card).getAllByRole("row").slice(1);
    await fireEvent.change(verdict, { target: { value: "all" } });
    await waitFor(() => expect(rows()).toHaveLength(5));
    await fireEvent.change(verdict, { target: { value: "warn" } });
    await waitFor(() => expect(rows()).toHaveLength(1));
    await fireEvent.change(verdict, { target: { value: "all" } });
    await fireEvent.change(endpoint, { target: { value: "=sensor-2" } });
    await waitFor(() => expect(rows()).toHaveLength(1));
    expect(rows()[0]).toHaveTextContent("node.surface.verdict.remove");
    // A decision on the root's key "" is listed under the root's title.
    await fireEvent.change(endpoint, { target: { value: "=" } });
    await waitFor(() => expect(rows()).toHaveLength(1));
    expect(rows()[0]).toHaveTextContent("node.surface.root");
    expect(rows()[0]).toHaveTextContent("node.surface.verdict.keep");
  });

  it("lists the withheld element and the announced change with its element changes", async () => {
    renderSurface();
    expect(await screen.findByTestId("surface-withheld")).toHaveTextContent("air-1");
    const change = screen.getByTestId("surface-change");
    expect(change).toHaveTextContent("lamp-1");
    expect(change).toHaveTextContent("1a2b3c4d5e6f → 9a8b7c6d5e4f");
    const items = within(change).getAllByTestId("surface-change-item");
    expect(items).toHaveLength(2);
    expect(items[0]).toHaveTextContent("node.surface.change_changed");
    expect(items[0]).toHaveTextContent("revision 5 → revision 6");
    expect(items[1]).toHaveTextContent("node.surface.change_added");
  });

  it("says nothing is withheld and no change is announced for an empty report", async () => {
    const ui = testUI({
      answers: {
        "GET /node/surface": {
          status: 200,
          body: {
            assembled: false, changed: [], decisions: [], effective: { endpoints: [], root: { clusters: [], deviceTypes: [], key: "" } },
            extensions: [], level: "1.6.1", level_origin: "pin", policies: "00", profiles: [], withheld: [],
          },
        },
      },
    });
    render(Harness, { ui, page: Surface });
    expect(await screen.findByText("node.surface.withheld_empty")).toBeInTheDocument();
    expect(screen.getByText("node.surface.changes_empty")).toBeInTheDocument();
    expect(screen.getByText("node.surface.not_assembled")).toBeInTheDocument();
  });

  it("shows the unready state the other pages show without a node", async () => {
    renderSurface(["node-surface.503"]);
    expect(await screen.findByText(/common\.error/)).toBeInTheDocument();
    expect(screen.queryByTestId("surface-head")).toBeNull();
  });

  it("refetches when the store's topology epoch moves", async () => {
    const ui = renderSurface(["node-surface.200", "node-endpoints.200", "node-status.200"]);
    await screen.findByTestId("surface-head");
    const count = () => ui.sent.filter((s) => s.path === "/api/v1/node/surface").length;
    expect(count()).toBe(1);
    ui.node.applyEvent({ type: "endpoints_changed", data: {} });
    await vi.waitFor(() => expect(count()).toBe(2));
  });
});
