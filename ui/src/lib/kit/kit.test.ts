// @vitest-environment happy-dom
// Ported from openccu-loom assets/ui/src/lib/components/ui/
// state-components.test.ts, Button.theme.test.ts, DataTable.test.ts and
// DataTable.group.test.ts at b40b289a.
import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import { createRawSnippet, type Component } from "svelte";
import Button from "./Button.svelte";
import Card from "./Card.svelte";
import DataTableComponent from "./DataTable.svelte";
import EmptyState from "./EmptyState.svelte";
import ErrorState from "./ErrorState.svelte";
import LoadingState from "./LoadingState.svelte";
import Select from "./Select.svelte";
import type { DataColumn } from "./data-table";
import { makeTextMatcher } from "./utils";

afterEach(() => {
  cleanup();
  localStorage.clear();
});

describe("LoadingState", () => {
  it("renders the default loading label and a status spinner", () => {
    render(LoadingState);
    expect(screen.getByRole("status")).toHaveTextContent("Loading…");
  });
  it("renders a custom label when provided", () => {
    render(LoadingState, { label: "Fetching" });
    expect(screen.getByRole("status")).toHaveTextContent("Fetching");
  });
});

describe("EmptyState", () => {
  it("renders the message and the optional description", () => {
    render(EmptyState, { message: "Nothing", description: "Add one" });
    expect(screen.getByText("Nothing")).toBeInTheDocument();
    expect(screen.getByText("Add one")).toBeInTheDocument();
  });
  it("omits the description line when not provided", () => {
    const { container } = render(EmptyState, { message: "Nothing" });
    expect(container.querySelectorAll("p")).toHaveLength(1);
  });
});

describe("ErrorState", () => {
  it("renders the localized error prefix plus the message, and no retry without onRetry", () => {
    render(ErrorState, { message: "boom" });
    expect(screen.getByText("Error: boom")).toBeInTheDocument();
    expect(screen.queryByRole("button")).toBeNull();
  });
  it("renders a retry button that invokes onRetry", async () => {
    const onRetry = vi.fn();
    render(ErrorState, { message: "boom", onRetry });
    await fireEvent.click(screen.getByRole("button", { name: "Reload" }));
    expect(onRetry).toHaveBeenCalledOnce();
  });
});

describe("theme tokens on the kit", () => {
  it("renders Button variants and the focus ring with --ha-* tokens, not hard-coded colours", () => {
    render(Button, { children: createRawSnippet(() => ({ render: () => "<span>Go</span>" })) });
    const cls = screen.getByRole("button").className;
    expect(cls).toContain("bg-[var(--ha-primary-color)]");
    expect(cls).toContain("focus-visible:ring-[var(--ha-primary-color)]");
    expect(cls).not.toMatch(/\b(bg|text)-(slate|teal|brand)-\d/);
  });
  it("renders Card with token-backed border and background", () => {
    const { container } = render(Card);
    const cls = (container.firstElementChild as HTMLElement).className;
    expect(cls).toContain("border-[var(--ha-divider-color)]");
    expect(cls).toContain("bg-[var(--ha-card-background-color)]");
  });
});

describe("Select", () => {
  it("reports the chosen value", async () => {
    const onValueChange = vi.fn();
    render(Select, { options: [{ value: "a", label: "A" }, { value: "b", label: "B" }], value: "a", onValueChange, ariaLabel: "pick" });
    await fireEvent.change(screen.getByLabelText("pick"), { target: { value: "b" } });
    expect(onValueChange).toHaveBeenCalledWith("b");
  });
});

type Row = { id: string; name: string; n: number; g: string };
const rows: Row[] = [
  { id: "1", name: "Charlie", n: 3, g: "x" },
  { id: "2", name: "alpha", n: 1, g: "y" },
  { id: "3", name: "Bravo", n: 2, g: "x" },
];
const columns: DataColumn<Row>[] = [
  { key: "name", label: "Name", sortable: true, get: (r) => r.name },
  { key: "n", label: "N", sortable: true, numeric: true, get: (r) => r.n },
];
// The generic component, its row type erased for the props bag.
const DataTable = DataTableComponent as unknown as Component<Record<string, unknown>>;
const names = (c: HTMLElement) => [...c.querySelectorAll("tbody tr td:first-child")].map((td) => td.textContent?.trim());

describe("DataTable", () => {
  it("renders a row per data row, the column labels, and get() text without a snippet", () => {
    const { container } = render(DataTable, { rows, columns, rowKey: (r: Row) => r.id, emptyMessage: "none" });
    expect(container.querySelectorAll("tbody tr")).toHaveLength(3);
    expect(screen.getByText("Name")).toBeInTheDocument();
    expect(screen.getByText("Charlie")).toBeInTheDocument();
  });

  it("shows emptyMessage and its description when rows is empty", () => {
    render(DataTable, { rows: [], columns, rowKey: (r: Row) => r.id, emptyMessage: "none", emptyDescription: "later" });
    expect(screen.getByText("none")).toBeInTheDocument();
    expect(screen.getByText("later")).toBeInTheDocument();
  });

  it("sorts ascending on the first header click, descending on the second, and switches columns", async () => {
    const { container } = render(DataTable, { rows, columns, rowKey: (r: Row) => r.id, emptyMessage: "none" });
    await fireEvent.click(screen.getByRole("button", { name: /Name/ }));
    expect(names(container)).toEqual(["alpha", "Bravo", "Charlie"]);
    await fireEvent.click(screen.getByRole("button", { name: /Name/ }));
    expect(names(container)).toEqual(["Charlie", "Bravo", "alpha"]);
    await fireEvent.click(screen.getByRole("button", { name: /^N\s*$/ }));
    expect(names(container)).toEqual(["alpha", "Bravo", "Charlie"]);
  });

  it("filters by the search box, shows emptyMessage without a match, and restores on clear", async () => {
    const { container } = render(DataTable, { rows, columns, rowKey: (r: Row) => r.id, emptyMessage: "none", search: true });
    const box = screen.getByRole("searchbox");
    await fireEvent.input(box, { target: { value: "brav" } });
    expect(names(container)).toEqual(["Bravo"]);
    await fireEvent.input(box, { target: { value: "zzz" } });
    expect(screen.getByText("none")).toBeInTheDocument();
    await fireEvent.input(box, { target: { value: "" } });
    expect(container.querySelectorAll("tbody tr")).toHaveLength(3);
  });

  it("renders one tbody per group, every row under its key once, in sort order, the key as text without a header snippet", async () => {
    const { container } = render(DataTable, { rows, columns, rowKey: (r: Row) => r.id, emptyMessage: "none", groupBy: (r: Row) => r.g });
    const bodies = container.querySelectorAll("tbody");
    expect(bodies).toHaveLength(2);
    expect(bodies[0].querySelector("th")?.textContent?.trim()).toBe("x");
    expect(bodies[0].querySelectorAll("td[data-label='Name']")).toHaveLength(2);
    expect(bodies[1].querySelectorAll("td[data-label='Name']")).toHaveLength(1);
    await fireEvent.click(screen.getByRole("button", { name: /Name/ }));
    expect(container.querySelectorAll("tbody")[0].querySelector("th")?.textContent?.trim()).toBe("y");
  });

  it("persists the sort under persistKey", async () => {
    render(DataTable, { rows, columns, rowKey: (r: Row) => r.id, emptyMessage: "none", persistKey: "t1" });
    await fireEvent.click(screen.getByRole("button", { name: /Name/ }));
    expect(JSON.parse(localStorage.getItem("fabric-ui.table:t1") ?? "{}")).toMatchObject({ sortKey: "name", sortAsc: true });
  });
});

describe("makeTextMatcher", () => {
  it("matches the literal text and, when valid, the regular expression", () => {
    expect(makeTextMatcher("lamp|light")("Ceiling light")).toBe(true);
    expect(makeTextMatcher("Kitchen (left)")("Kitchen (left) lamp")).toBe(true);
    expect(makeTextMatcher("(")("a(b")).toBe(true);
    expect(makeTextMatcher("")("anything")).toBe(true);
  });
});
