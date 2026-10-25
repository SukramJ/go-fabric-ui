// The column descriptor of DataTable.svelte. Ported from openccu-loom
// assets/ui/src/lib/components/ui/data-table.ts:1-55 at b40b289a, without
// the per-column filter row and column groups the node pages do not use.

export type DataTableAlign = "left" | "right" | "center";

export type DataColumn<Row> = {
  /** Stable key: sort state, the cell's data-label, the cell snippet's branch. */
  key: string;
  /** Header label. */
  label: string;
  align?: DataTableAlign;
  sortable?: boolean;
  /** Scalar accessor for sorting, the default text and the search. */
  get?: (row: Row) => string | number | null | undefined;
  /** The primary column: on the phone reflow it is the card heading. */
  title?: boolean;
  cellClass?: string;
  headClass?: string;
  /** Tabular figures, right-aligned unless align says otherwise. */
  numeric?: boolean;
};

export type DataTableSort = { key: string; asc: boolean };
