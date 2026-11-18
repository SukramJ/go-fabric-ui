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

/**
 * The rows with the key each is rendered under: rowKey's, and for a key
 * that repeats, the key with a count appended until it is unused. The rows of a node
 * or the network may repeat a key, and a repeated key in a keyed each
 * throws (UR-P2-ui-90).
 */
export function keyedRows<Row>(rows: Row[], rowKey: (row: Row) => string): { row: Row; key: string }[] {
  const used = new Set<string>();
  return rows.map((row) => {
    const k = rowKey(row);
    let key = k;
    for (let n = 1; used.has(key); n++) key = `${k}\u0000${n}`;
    used.add(key);
    return { row, key };
  });
}
