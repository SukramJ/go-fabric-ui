// @vitest-environment happy-dom
// The table's persisted preferences are read with a shape check
// (UR-P1-web-114): localStorage is shared by everything of the origin.
import { afterEach, describe, expect, it } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import type { Component } from "svelte";
import DataTableComponent from "./DataTable.svelte";
import type { DataColumn } from "./data-table";

afterEach(() => {
  cleanup();
  localStorage.clear();
});

type Row = { id: string; name: string };
const rows: Row[] = [
  { id: "1", name: "Charlie" },
  { id: "2", name: "alpha" },
];
const columns: DataColumn<Row>[] = [{ key: "name", label: "Name", sortable: true, get: (r) => r.name }];
const DataTable = DataTableComponent as unknown as Component<Record<string, unknown>>;
const names = (c: HTMLElement) => [...c.querySelectorAll("tbody tr td:first-child")].map((td) => td.textContent?.trim());

describe("UR-P1-web-114: stored table preferences of the wrong shape", () => {
  it("renders the rows when the stored value has the wrong types", () => {
    localStorage.setItem("fabric-ui.table:t1", '{"query":5,"sortKey":{},"sortAsc":"yes"}');
    const { container } = render(DataTable, { rows, columns, rowKey: (r: Row) => r.id, emptyMessage: "none", persistKey: "t1", search: true });
    expect(names(container)).toEqual(["Charlie", "alpha"]);
  });

  it("does not take a stored search, nor keep one", async () => {
    localStorage.setItem("fabric-ui.table:t1", '{"query":"(a+)+$","sortKey":"name","sortAsc":true}');
    const { container } = render(DataTable, { rows, columns, rowKey: (r: Row) => r.id, emptyMessage: "none", persistKey: "t1", search: true });
    // The sort is taken, the search is not.
    expect(names(container)).toEqual(["alpha", "Charlie"]);
    expect((screen.getByRole("searchbox") as HTMLInputElement).value).toBe("");
    await fireEvent.input(screen.getByRole("searchbox"), { target: { value: "alp" } });
    expect(JSON.parse(localStorage.getItem("fabric-ui.table:t1") ?? "{}")).not.toHaveProperty("query");
  });

  it("drops a stored sort key no column has", () => {
    localStorage.setItem("fabric-ui.table:t1", '{"sortKey":"gone","sortAsc":true}');
    render(DataTable, { rows, columns, rowKey: (r: Row) => r.id, emptyMessage: "none", persistKey: "t1" });
    expect(JSON.parse(localStorage.getItem("fabric-ui.table:t1") ?? "{}")).toMatchObject({ sortKey: "" });
  });
});
