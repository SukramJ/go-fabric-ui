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
import AttributeTable from "./AttributeTable.svelte";
import ClusterView from "./ClusterView.svelte";
import TagChips from "./TagChips.svelte";
import EventList from "./EventList.svelte";
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

describe("TagChips", () => {
  it("shows a tag's label, else namespace:tag, with the numbers in the title", () => {
    render(TagChips, { tags: [{ namespace: 7, tag: 1, label: "One" }, { namespace: 8, tag: 2, mfg_code: 0xfff1 }] });
    expect(screen.getByText("One")).toHaveAttribute("title", "namespace 7 · tag 1");
    expect(screen.getByText("8:2")).toHaveAttribute("title", "namespace 8 · tag 2 · mfg 0xFFF1");
  });
  it("shows the empty text only when given", () => {
    const { container } = render(TagChips, { tags: [] });
    expect(container.textContent?.trim()).toBe("");
    cleanup();
    render(TagChips, { tags: [], empty: "No tags" });
    expect(screen.getByText("No tags")).toBeInTheDocument();
  });
});

describe("ClusterView and AttributeTable", () => {
  it("shows a feature without a title by its code and an attribute without a value or with a status", () => {
    render(ClusterView, {
      cluster: {
        id: 0xfc00, id_hex: "0xFC00", revision: 1, feature_map_hex: "0x00000002", features: [{ code: "X", bit: 1 }],
        attributes: [{ id: 0, id_hex: "0x0000" }, { id: 1, id_hex: "0x0001", name: "Level", value: { a: [1, 2] }, status: 0x86 }],
        accepted: [], generated: [{ id: 3, id_hex: "0x0003" }], events: [],
      },
    });
    expect(screen.getByTestId("feature-chips")).toHaveTextContent(/^X$/);
    expect(screen.getByTestId("attribute-0x0000")).toHaveTextContent("no value");
    expect(screen.getByTestId("attribute-0x0001")).toHaveTextContent("0x86");
    expect(screen.getByTestId("attribute-0x0001").querySelector("pre")?.textContent).toBe(JSON.stringify({ a: [1, 2] }, null, 2));
    expect(screen.getByText("0x0003")).toBeInTheDocument();
  });
  it("names a status of the Interaction Model with its code, an unknown one by its code alone", () => {
    render(AttributeTable, {
      attributes: [
        { id: 0, id_hex: "0x0000", name: "OnOff", status: 0x86 },
        { id: 1, id_hex: "0x0001", status: 0x8f },
        { id: 2, id_hex: "0x0002", status: 0x42 },
        { id: 3, id_hex: "0x0003", value: true },
      ],
    });
    const cell = (hex: string) => screen.getByTestId(`attribute-${hex}`).querySelector("td:last-child")!;
    expect(cell("0x0000")).toHaveTextContent("Unsupported attribute 0x86");
    expect(cell("0x0000").querySelector("[title]")?.getAttribute("title")).toBe("UNSUPPORTED_ATTRIBUTE (0x86)");
    expect(cell("0x0001")).toHaveTextContent("Not readable 0x8F");
    expect(cell("0x0002").textContent?.trim()).toBe("0x42");
    expect(cell("0x0002").querySelector("[title]")).toBeNull();
    expect(cell("0x0003").textContent?.trim()).toBe("");
    expect(screen.getAllByTestId("attribute-status")).toHaveLength(3);
  });
  it("renders an empty attribute table", () => {
    render(AttributeTable, { attributes: [] });
    expect(screen.getByTestId("attribute-table").querySelectorAll("tbody tr")).toHaveLength(0);
  });
});

describe("EventList", () => {
  it("lists the events newest first: number, priority, name or id, endpoint, time and data", () => {
    // events is a Svelte mount option too: the props go under props.
    render(EventList, {
      props: {
        empty: "none",
        events: [
          { number: 3, priority: "debug", endpoint: 0, cluster_hex: "0xFC00", event_hex: "0x0001" },
          { number: 5, priority: "critical", endpoint: 1, cluster_hex: "0x0028", event_hex: "0x0000", name: "StartUp", epoch_ms: 1792929480000, data: { softwareVersion: 1 } },
        ],
      },
    });
    const items = [...screen.getByTestId("event-list").querySelectorAll("li")];
    expect(items.map((li) => li.getAttribute("data-testid"))).toEqual(["event-5", "event-3"]);
    expect(items[0]).toHaveTextContent("StartUp");
    expect(items[0]).toHaveTextContent("Critical");
    expect(items[0].querySelector("time")).toHaveAttribute("datetime", new Date(1792929480000).toISOString());
    expect(items[0].querySelector("pre")?.textContent).toBe(JSON.stringify({ softwareVersion: 1 }, null, 2));
    expect(items[1]).toHaveTextContent("0x0001");
    expect(items[1].querySelector("time")).toBeNull();
    expect(items[1].querySelector("pre")).toBeNull();
  });
  it("shows the empty text without events", () => {
    render(EventList, { props: { events: [], empty: "No events" } });
    expect(screen.getByTestId("event-list-empty")).toHaveTextContent("No events");
  });
});
