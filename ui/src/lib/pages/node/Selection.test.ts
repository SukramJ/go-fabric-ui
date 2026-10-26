// @vitest-environment happy-dom
// Ported from openccu-loom assets/ui/src/routes/matter/MatterExposureList.test.ts
// at b40b289a, made generic over the provider's schema, plus the preview
// before apply.
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { body, testUI } from "../../../test/support";
import type { ExposureSchema, ItemList } from "../../api/client";
import Selection from "./Selection.svelte";

afterEach(() => {
  cleanup();
  localStorage.clear();
});

const BASE = ["node-exposure-schema.200", "node-exposure.200"];

function renderSel(opts: Parameters<typeof testUI>[0] = {}) {
  const ui = testUI({ ...opts, fixtures: [...BASE, ...(opts.fixtures ?? [])] });
  render(Harness, { ui, page: Selection });
  return ui;
}

describe("Selection — grouping", () => {
  it("prints each group once as a header, not per row", async () => {
    const { container } = render(Harness, { ui: testUI({ fixtures: BASE }), page: Selection });
    await screen.findByText("Kitchen lamp");
    const headers = [...container.querySelectorAll("tbody th")].map((th) => th.textContent?.trim().split(/\s+/)[0]);
    expect(headers).toEqual(["Kitchen", "Hall"]);
  });

  it("does not group when the schema says so", async () => {
    const schema = body<ExposureSchema>("node-exposure-schema.200");
    schema.groups = false;
    const ui = testUI({ fixtures: BASE, answers: { "GET /node/exposure/schema": { status: 200, body: schema } } });
    const { container } = render(Harness, { ui, page: Selection });
    await screen.findByText("Kitchen lamp");
    expect(container.querySelectorAll("tbody")).toHaveLength(1);
  });

  it("renders one select checkbox per item and the host's columns", async () => {
    renderSel();
    await screen.findByText("Kitchen lamp");
    expect(screen.getAllByLabelText("node.selection.select_row")).toHaveLength(3);
    expect(screen.getByRole("columnheader", { name: /Room/ })).toBeInTheDocument();
  });
});

describe("Selection — status icons and legend", () => {
  it("shows the legend with the four state labels and an SVG icon per row", async () => {
    const { container } = render(Harness, { ui: testUI({ fixtures: BASE }), page: Selection });
    await screen.findByText("Kitchen lamp");
    for (const s of ["exposed", "partial", "available", "unmappable"]) expect(screen.getAllByText(`node.selection.state_${s}`).length).toBeGreaterThan(0);
    expect(container.querySelectorAll("tbody td[data-label='node.selection.col_state'] svg")).toHaveLength(3);
  });

  it("gives an unmappable item's checkbox a disabled state and a reason", async () => {
    renderSel();
    await screen.findByText("Hall remote");
    const boxes = screen.getAllByLabelText("node.selection.select_row");
    const disabled = boxes.filter((b) => (b as HTMLInputElement).disabled);
    expect(disabled).toHaveLength(1);
    expect(disabled[0]).toHaveAttribute("title", "node.selection.unmappable_checkbox_title");
  });
});

describe("Selection — empty, search and filters", () => {
  it("shows the empty state when there are no items", async () => {
    renderSel({ answers: { "GET /node/exposure": { status: 200, body: { items: [] } } } });
    expect(await screen.findByText("node.selection.empty")).toBeInTheDocument();
  });

  it("filters by the search box over name, group and values", async () => {
    renderSel();
    await screen.findByText("Kitchen lamp");
    await fireEvent.input(screen.getByPlaceholderText("node.selection.search_placeholder"), { target: { value: "remote" } });
    await waitFor(() => expect(screen.queryByText("Kitchen lamp")).toBeNull());
    expect(screen.getByText("Hall remote")).toBeInTheDocument();
  });

  it("offers the schema's filters and a value filter per filterable column", async () => {
    renderSel();
    await screen.findByText("Kitchen lamp");
    const kind = screen.getByLabelText("Kind");
    await fireEvent.change(kind, { target: { value: "sensor" } });
    await waitFor(() => expect(screen.queryByText("Kitchen lamp")).toBeNull());
    expect(screen.getByText("Hall sensor")).toBeInTheDocument();
    const room = screen.getByLabelText("Room");
    expect(within(room).getAllByRole("option").map((o) => o.textContent)).toEqual(["Room: node.selection.filter_all", "Hall", "Kitchen"]);
  });

  it("shows a list-valued column as chips and filters it by membership", async () => {
    const { container } = render(Harness, { ui: testUI({ fixtures: BASE }), page: Selection });
    await screen.findByText("Kitchen lamp");
    const cells = [...container.querySelectorAll("tbody td[data-label='Labels']")];
    expect(cells.map((c) => [...c.querySelectorAll("span span")].map((s) => s.textContent?.trim()))).toEqual([["Evening", "Ceiling"], ["Evening", "Living, dining"], []]);
    const labels = screen.getByLabelText("Labels");
    expect(within(labels).getAllByRole("option").map((o) => o.textContent)).toEqual(["Labels: node.selection.filter_all", "Ceiling", "Evening", "Living, dining"]);
    await fireEvent.change(labels, { target: { value: "Evening" } });
    await waitFor(() => expect(screen.queryByText("Hall remote")).toBeNull());
    expect(screen.getByText("Kitchen lamp")).toBeInTheDocument();
    expect(screen.getByText("Hall sensor")).toBeInTheDocument();
  });

  it("keeps a listed value with a comma one chip and one filter option (issue #10)", async () => {
    const { container } = render(Harness, { ui: testUI({ fixtures: BASE }), page: Selection });
    await screen.findByText("Kitchen lamp");
    const chips = [...container.querySelectorAll("tbody td[data-label='Labels'] span span")].map((s) => s.textContent?.trim());
    expect(chips).toContain("Living, dining");
    expect(chips).not.toContain("Living");
    expect(chips).not.toContain("dining");
    await fireEvent.change(screen.getByLabelText("Labels"), { target: { value: "Living, dining" } });
    await waitFor(() => expect(screen.queryByText("Kitchen lamp")).toBeNull());
    expect(screen.getByText("Hall sensor")).toBeInTheDocument();
    expect(screen.queryByText("Hall remote")).toBeNull();
  });

  it("matches a host filter marked multi by membership, an unmarked one by equality", async () => {
    const schema = body<ExposureSchema>("node-exposure-schema.200");
    schema.filters = [{ field: "labels", label: "Tag", multi: true, options: [{ value: "Ceiling", label: "Ceiling" }] }];
    renderSel({ answers: { "GET /node/exposure/schema": { status: 200, body: schema } } });
    await screen.findByText("Kitchen lamp");
    await fireEvent.change(screen.getByLabelText("Tag"), { target: { value: "Ceiling" } });
    await waitFor(() => expect(screen.queryByText("Hall sensor")).toBeNull());
    expect(screen.getByText("Kitchen lamp")).toBeInTheDocument();
    cleanup();
    schema.filters[0].multi = false;
    renderSel({ answers: { "GET /node/exposure/schema": { status: 200, body: schema } } });
    await screen.findByText("Kitchen lamp");
    await fireEvent.change(screen.getByLabelText("Tag"), { target: { value: "Ceiling" } });
    expect(await screen.findByText("node.selection.empty")).toBeInTheDocument();
  });

  it("shows the host's warnings collapsed above the list", async () => {
    renderSel({ fixtures: ["node-status.200"] });
    const box = await screen.findByTestId("host-warnings");
    expect(box.tagName).toBe("DETAILS");
    expect((box as HTMLDetailsElement).open).toBe(false);
    expect(box).toHaveTextContent('node.warnings.summary:{"count":2}');
    expect(within(box).getByText("rule_unused")).toBeInTheDocument();
  });

  it("says so when the host serves no selection", async () => {
    const ui = testUI({ fixtures: ["node-exposure-schema.503", "node-exposure.200"] });
    render(Harness, { ui, page: Selection });
    expect(await screen.findByText(/common\.error/)).toBeInTheDocument();
    const off = testUI({ fixtures: ["node-status.200-disabled"] });
    await off.node.loadStatus();
    cleanup();
    render(Harness, { ui: off, page: Selection });
    expect(await screen.findByText("node.selection.unavailable")).toBeInTheDocument();
    expect(off.sent.some((s) => s.path.endsWith("/exposure"))).toBe(false);
  });
});

describe("Selection — changes, preview, apply", () => {
  it("collects a bulk change, previews the verdicts, and applies only on Apply", async () => {
    const preview = body<ItemList>("node-exposure-preview.200");
    const ui = renderSel({ fixtures: ["node-exposure-apply.200"], answers: { "POST /node/exposure/preview": { status: 200, body: preview } } });
    await screen.findByText("Kitchen lamp");
    const boxes = screen.getAllByLabelText("node.selection.select_row");
    await fireEvent.click(boxes[0]);
    await fireEvent.click(screen.getByText("node.selection.bulk_expose"));
    expect(await screen.findByTestId("save-bar")).toHaveTextContent('node.selection.pending:{"count":1}');
    expect(screen.getByText("common.modified")).toBeInTheDocument();

    await fireEvent.click(screen.getByText("node.selection.preview"));
    const panel = await screen.findByTestId("preview");
    expect(ui.sent.find((s) => s.path.endsWith("/preview"))?.body).toEqual({ items: [{ key: "lamp-1", enabled: true }] });
    expect(ui.sent.some((s) => s.method === "PUT")).toBe(false);
    expect(within(panel).getByText(preview.items[0].name)).toBeInTheDocument();

    await fireEvent.click(within(panel).getByText("node.selection.apply"));
    await waitFor(() => expect(ui.toastSpy.success).toHaveBeenCalledWith('node.selection.saved_toast:{"count":2}'));
    expect(ui.sent.find((s) => s.method === "PUT")?.body).toEqual({ items: [{ key: "lamp-1", enabled: true }] });
  });

  it("marks a verdict that the change moves", async () => {
    const preview = body<ItemList>("node-exposure-preview.200");
    preview.items[0].verdict.state = "mappable";
    renderSel({ answers: { "POST /node/exposure/preview": { status: 200, body: preview } } });
    await screen.findByText("Hall sensor");
    await fireEvent.click(screen.getAllByLabelText("node.selection.select_row")[1]);
    await fireEvent.click(screen.getByText("node.selection.bulk_hide"));
    await fireEvent.click(await screen.findByText("node.selection.preview"));
    expect(await screen.findByText("node.selection.preview_changed")).toBeInTheDocument();
  });

  it("renames through the drawer and discards on request", async () => {
    const ui = renderSel();
    await fireEvent.click(await screen.findByText("Kitchen lamp"));
    const name = await screen.findByLabelText("node.selection.name");
    await fireEvent.input(name, { target: { value: "Pantry lamp" } });
    await fireEvent.click(screen.getByText("common.save"));
    expect(await screen.findByText("Pantry lamp")).toBeInTheDocument();
    expect(ui.node.pending.get("lamp-1")).toEqual({ key: "lamp-1", enabled: false, name: "Pantry lamp" });
    await fireEvent.click(screen.getByText("node.selection.discard"));
    expect(await screen.findByText("Kitchen lamp")).toBeInTheDocument();
    expect(ui.node.hasDirty).toBe(false);
  });
});
