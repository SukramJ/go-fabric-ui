<!-- Ported from openccu-loom assets/ui/src/lib/components/ui/DataTable.svelte:1-469
     at b40b289a: sortable headers, an optional search box, sort and search
     persisted under persistKey, the phone reflow (table-reflow in
     tokens.css), groupBy sections and activatable rows. The column filter
     row, column groups and row expansion are left out; the node pages use
     none of them. -->
<script lang="ts" generics="Row">
  import type { Snippet } from "svelte";
  import type { IconName } from "./icons";
  import type { DataColumn, DataTableSort } from "./data-table";
  import { makeTextMatcher } from "./utils";
  import { useT } from "../context";
  import Input from "./Input.svelte";
  import EmptyState from "./EmptyState.svelte";

  let {
    rows,
    columns,
    rowKey,
    search = false,
    searchPlaceholder = "",
    initialSort = null,
    persistKey = "",
    emptyMessage,
    emptyDescription = "",
    emptyIcon = "mdi:format-list-bulleted",
    cell,
    rowClass,
    onRowClick,
    groupBy,
    groupHeader,
  }: {
    rows: Row[];
    columns: DataColumn<Row>[];
    rowKey: (row: Row) => string;
    search?: boolean;
    searchPlaceholder?: string;
    initialSort?: DataTableSort | null;
    /** When set, sort and search persist under this localStorage key. */
    persistKey?: string;
    emptyMessage: string;
    emptyDescription?: string;
    emptyIcon?: IconName;
    cell?: Snippet<[Row, DataColumn<Row>]>;
    rowClass?: (row: Row) => string;
    /** Makes every row activatable; a click inside a control stays the control's. */
    onRowClick?: (row: Row) => void;
    /** Splits the rows into sections, after search and sort, in first-row order. */
    groupBy?: (row: Row) => string;
    groupHeader?: Snippet<[string, Row[]]>;
  } = $props();

  const t = useT();

  type Persisted = { sortKey: string; sortAsc: boolean; query: string };

  // Read initialSort once; a persisted state wins.
  function initialState(): Persisted {
    let base: Persisted = { sortKey: initialSort?.key ?? "", sortAsc: initialSort?.asc ?? true, query: "" };
    if (persistKey) {
      try {
        const raw = localStorage.getItem("fabric-ui.table:" + persistKey);
        if (raw) base = { ...base, ...(JSON.parse(raw) as Partial<Persisted>) };
      } catch {
        // storage unavailable — the initial sort stands.
      }
    }
    return base;
  }

  const init0 = initialState();
  let sortKey = $state(init0.sortKey);
  let sortAsc = $state(init0.sortAsc);
  let query = $state(init0.query);

  $effect(() => {
    if (!persistKey) return;
    const snapshot: Persisted = { sortKey, sortAsc, query };
    try {
      localStorage.setItem("fabric-ui.table:" + persistKey, JSON.stringify(snapshot));
    } catch {
      // storage unavailable — sort and search do not persist.
    }
  });

  function toggleSort(col: DataColumn<Row>) {
    if (!col.sortable) return;
    if (sortKey === col.key) sortAsc = !sortAsc;
    else {
      sortKey = col.key;
      sortAsc = true;
    }
  }

  function alignClass(col: DataColumn<Row>): string {
    const align = col.align ?? (col.numeric ? "right" : "left");
    return align === "right" ? "text-right" : align === "center" ? "text-center" : "text-left";
  }

  function rowClickHandler(row: Row) {
    return (event: MouseEvent) => {
      const target = event.target as HTMLElement | null;
      if (target?.closest("button, a, input, select, textarea, label")) return;
      onRowClick?.(row);
    };
  }

  function rowKeyHandler(row: Row) {
    return (event: KeyboardEvent) => {
      if (event.key !== "Enter" && event.key !== " ") return;
      const target = event.target as HTMLElement | null;
      if (target?.closest("button, a, input, select, textarea")) return;
      event.preventDefault();
      onRowClick?.(row);
    };
  }

  function compare(a: unknown, b: unknown): number {
    const an = a === null || a === undefined;
    const bn = b === null || b === undefined;
    if (an && bn) return 0;
    if (an) return 1; // nulls last
    if (bn) return -1;
    if (typeof a === "number" && typeof b === "number") return a - b;
    return String(a).localeCompare(String(b), undefined, { sensitivity: "base", numeric: true });
  }

  const processed = $derived.by(() => {
    let list = rows;
    if (search && query.trim()) {
      const match = makeTextMatcher(query);
      list = list.filter((r) =>
        columns.some((c) => {
          const v = c.get?.(r);
          return v != null && match(String(v));
        }),
      );
    }
    const col = sortKey ? columns.find((c) => c.key === sortKey) : undefined;
    if (col?.get) {
      const get = col.get;
      list = [...list].sort((a, b) => {
        const cmp = compare(get(a), get(b));
        return sortAsc ? cmp : -cmp;
      });
    }
    return list;
  });

  const groups = $derived.by(() => {
    if (!groupBy) return null;
    const out = new Map<string, Row[]>();
    for (const row of processed) {
      const key = groupBy(row);
      const bucket = out.get(key);
      if (bucket) bucket.push(row);
      else out.set(key, [row]);
    }
    return Array.from(out, ([key, rows]) => ({ key, rows }));
  });

  function ariaSort(col: DataColumn<Row>): "ascending" | "descending" | "none" {
    if (!col.sortable || sortKey !== col.key) return "none";
    return sortAsc ? "ascending" : "descending";
  }
</script>

{#if search}
  <div class="mb-3 max-w-md">
    <Input type="search" placeholder={searchPlaceholder || t("common.search")} bind:value={query} />
  </div>
{/if}

{#if processed.length === 0}
  <EmptyState message={emptyMessage} description={emptyDescription} icon={emptyIcon} />
{:else}
  <div class="overflow-x-auto">
    <table class="table-reflow w-full text-sm">
      <thead class="border-b border-[var(--ha-divider-color)] text-left text-xs font-semibold uppercase tracking-wide text-[var(--ha-secondary-text-color)]">
        <tr>
          {#each columns as col (col.key)}
            <th class="px-3 py-2 {alignClass(col)} {col.headClass ?? ''}" aria-sort={ariaSort(col)} scope="col">
              {#if col.sortable}
                <button
                  type="button"
                  class="inline-flex items-center gap-1 font-semibold uppercase tracking-wide hover:text-[var(--ha-primary-text-color)]"
                  onclick={() => toggleSort(col)}
                >
                  <span>{col.label}</span>
                  <span class="w-3 text-[10px]" aria-hidden="true">{sortKey === col.key ? (sortAsc ? "▲" : "▼") : ""}</span>
                </button>
              {:else}
                {col.label}
              {/if}
            </th>
          {/each}
        </tr>
      </thead>
      {#snippet bodyRow(row: Row)}
        <tr
          class="border-b border-[var(--ha-divider-color)] last:border-0 hover:bg-[var(--ha-secondary-background-color)] {onRowClick ? 'cursor-pointer' : ''} {rowClass?.(row) ?? ''}"
          tabindex={onRowClick ? 0 : undefined}
          onclick={onRowClick ? rowClickHandler(row) : undefined}
          onkeydown={onRowClick ? rowKeyHandler(row) : undefined}
        >
          {#each columns as col (col.key)}
            <td
              class="px-3 py-2 {alignClass(col)} {col.numeric ? 'tabular-nums' : ''} {col.title ? 'reflow-title' : ''} {col.cellClass ?? ''}"
              data-label={col.label}
            >
              {#if cell}{@render cell(row, col)}{:else}{col.get?.(row) ?? "—"}{/if}
            </td>
          {/each}
        </tr>
      {/snippet}

      {#if groups}
        {#each groups as group (group.key)}
          <tbody>
            <tr class="border-b border-[var(--ha-divider-color)] bg-[var(--ha-secondary-background-color)]">
              <th colspan={columns.length} scope="colgroup" class="px-3 py-2 text-left font-normal">
                {#if groupHeader}{@render groupHeader(group.key, group.rows)}{:else}{group.key}{/if}
              </th>
            </tr>
            {#each group.rows as row (rowKey(row))}
              {@render bodyRow(row)}
            {/each}
          </tbody>
        {/each}
      {:else}
        <tbody>
          {#each processed as row (rowKey(row))}
            {@render bodyRow(row)}
          {/each}
        </tbody>
      {/if}
    </table>
  </div>
{/if}
