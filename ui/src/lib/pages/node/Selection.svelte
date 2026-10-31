<!-- The host's selection: which items the node exposes. Generic over the
     provider's schema — its columns, its filters, its groups — with search,
     bulk selection, a dirty map, a preview of the verdicts before anything
     is written, and the save/discard bar. Ported from openccu-loom
     assets/ui/src/routes/matter/MatterExposureList.svelte:1-584 at
     b40b289a: the host's fixed columns, kind filter and per-item conflict
     hints become the schema's columns and filters; the class chips become a
     value filter per filterable column; Preview before Apply is new, and so
     are list-valued fields (Column.multi, Filter.multi) and the host's
     warnings above the list. Items whose verdict is unmappable — nothing
     the host could expose now — are hidden until the toggle above the list
     shows them (0.2.5). -->
<script lang="ts">
  import { onMount } from "svelte";
  import { useFabricUI } from "../../context";
  import { errorText } from "../../store.svelte";
  import type { Item, VerdictState } from "../../api/client";
  import type { DataColumn } from "../../kit/data-table";
  import type { IconName } from "../../kit/icons";
  import { makeTextMatcher } from "../../kit/utils";
  import Badge from "../../kit/Badge.svelte";
  import Button from "../../kit/Button.svelte";
  import Card from "../../kit/Card.svelte";
  import DataTable from "../../kit/DataTable.svelte";
  import EmptyState from "../../kit/EmptyState.svelte";
  import ErrorState from "../../kit/ErrorState.svelte";
  import Icon from "../../kit/Icon.svelte";
  import Input from "../../kit/Input.svelte";
  import LoadingState from "../../kit/LoadingState.svelte";
  import Select from "../../kit/Select.svelte";
  import HostWarnings from "./HostWarnings.svelte";
  import { distinctValues, fieldMatches, itemValues } from "./fields";

  const { t, toast, node } = useFabricUI();

  // A host without a selection says so in its status; the page then
  // loads nothing.
  const unavailable = $derived(node.status !== null && !node.status.exposure);

  onMount(() => {
    // The status carries the host's warnings; a host page that mounts the
    // selection alone has not loaded it yet.
    if (node.status === null && !node.statusLoading) void node.loadStatus();
    if (!unavailable) void node.loadExposure();
  });

  const ALL = "__all__";
  let search = $state("");
  let showUnmappable = $state(false);
  let filterValues = $state<Record<string, string>>({});
  let selected = $state<Set<string>>(new Set());
  let drawer = $state<Item | null>(null);
  let drawerName = $state("");
  let drawerEnabled = $state(false);
  let previewItems = $state<Item[] | null>(null);
  let previewing = $state(false);
  let saving = $state(false);

  const schema = $derived(node.schema);

  // The value filters: the host's fixed filters, then one per filterable
  // column with the distinct values the items carry — for a list-valued
  // column the individual values, not their combinations.
  type ValueFilter = { field: string; label: string; multi: boolean; options: { value: string; label: string }[] };
  const valueFilters = $derived.by<ValueFilter[]>(() => {
    if (!schema) return [];
    const out: ValueFilter[] = schema.filters.map((f) => ({ field: f.field, label: f.label, multi: !!f.multi, options: f.options }));
    for (const col of schema.columns) {
      if (!col.filter || out.some((f) => f.field === col.key)) continue;
      const values = distinctValues(node.items, col.key, !!col.multi);
      out.push({ field: col.key, label: col.label, multi: !!col.multi, options: values.map((v) => ({ value: v, label: v })) });
    }
    return out;
  });
  const multiField = $derived(new Set(valueFilters.filter((f) => f.multi).map((f) => f.field)));
  const multiColumn = $derived(new Set((schema?.columns ?? []).filter((c) => c.multi).map((c) => "field:" + c.key)));

  // The item as the page shows it: with its pending change applied.
  function current(item: Item): Item {
    const p = node.pending.get(item.key);
    if (!p) return item;
    return { ...item, enabled: p.enabled, name: p.name ?? item.name };
  }

  // The items nothing could expose now; hidden unless the toggle shows them.
  const unmappableCount = $derived(node.items.filter((i) => i.verdict.state === "unmappable").length);

  const filtered = $derived.by(() => {
    const match = makeTextMatcher(search);
    return node.items.filter((item) => {
      if (!showUnmappable && item.verdict.state === "unmappable") return false;
      for (const [field, value] of Object.entries(filterValues)) {
        if (value !== ALL && !fieldMatches(item, field, value, multiField.has(field))) return false;
      }
      if (!search.trim()) return true;
      const it = current(item);
      return [it.name, it.group, it.verdict.device_type_name, ...Object.values(it.fields), ...Object.values(it.lists ?? {}).flat()].some((v) => v && match(v));
    });
  });

  type StateInfo = { icon: IconName; color: string; label: string };

  function stateInfo(item: Item): StateInfo {
    const it = current(item);
    const s = it.verdict.state;
    if (s === "unmappable") return { icon: "mdi:cancel", color: "text-[var(--ha-error-color)]", label: t("node.selection.state_unmappable") };
    if (it.enabled && s === "mappable") return { icon: "mdi:check-circle", color: "text-[var(--ha-success-color)]", label: t("node.selection.state_exposed") };
    if (it.enabled && s === "partially_mappable") return { icon: "mdi:alert-circle", color: "text-[var(--ha-warning-color)]", label: t("node.selection.state_partial") };
    return { icon: "mdi:circle-outline", color: "text-[var(--ha-disabled-text-color)]", label: t("node.selection.state_available") };
  }

  function verdictHint(s: VerdictState): string {
    if (s === "unmappable") return t("node.selection.unmappable_hint");
    if (s === "partially_mappable") return t("node.selection.partially_mappable_hint");
    return t("node.selection.mappable_hint");
  }

  const legend = $derived<StateInfo[]>([
    { icon: "mdi:check-circle", color: "text-[var(--ha-success-color)]", label: t("node.selection.state_exposed") },
    { icon: "mdi:alert-circle", color: "text-[var(--ha-warning-color)]", label: t("node.selection.state_partial") },
    { icon: "mdi:circle-outline", color: "text-[var(--ha-disabled-text-color)]", label: t("node.selection.state_available") },
    { icon: "mdi:cancel", color: "text-[var(--ha-error-color)]", label: t("node.selection.state_unmappable") },
  ]);

  const bulkable = (item: Item) => item.verdict.state !== "unmappable";

  function toggleSelect(item: Item) {
    if (!bulkable(item)) return;
    const next = new Set(selected);
    if (next.has(item.key)) next.delete(item.key);
    else next.add(item.key);
    selected = next;
  }

  function selectAll() {
    selected = new Set(filtered.filter(bulkable).map((i) => i.key));
  }

  function bulkSet(enabled: boolean) {
    for (const item of filtered) if (selected.has(item.key)) node.markDirty(item, { enabled });
    selected = new Set();
  }

  function openDrawer(item: Item) {
    const it = current(item);
    drawer = item;
    drawerName = it.name;
    drawerEnabled = it.enabled;
  }

  function saveDrawer() {
    if (!drawer) return;
    node.markDirty(drawer, { enabled: drawerEnabled, name: drawerName });
    drawer = null;
  }

  async function preview() {
    previewing = true;
    try {
      previewItems = await node.preview();
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      previewing = false;
    }
  }

  async function apply() {
    saving = true;
    try {
      const applied = await node.saveBulk();
      previewItems = null;
      toast.success(t("node.selection.saved_toast", { count: applied }));
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      saving = false;
    }
  }

  function discard() {
    node.discardDirty();
    previewItems = null;
  }

  const before = (key: string) => node.items.find((i) => i.key === key);
  const verdictMoved = (a: Item | undefined, b: Item) =>
    !a || a.verdict.state !== b.verdict.state || a.verdict.device_type !== b.verdict.device_type || a.verdict.reason !== b.verdict.reason;

  const columns: DataColumn<Item>[] = $derived([
    { key: "select", label: t("node.selection.col_select") },
    { key: "state", label: t("node.selection.col_state"), align: "center", get: (i) => stateInfo(i).label },
    { key: "name", label: t("node.selection.col_name"), sortable: true, title: true, get: (i) => current(i).name },
    ...(schema?.columns ?? []).map((c) => ({
      key: "field:" + c.key,
      label: c.label,
      sortable: true,
      get: (i: Item) => (c.multi ? itemValues(i, c.key).join(", ") : (i.fields[c.key] ?? "")),
    })),
    { key: "device_type", label: t("node.selection.col_device_type"), sortable: true, get: (i) => i.verdict.device_type_name || "—" },
  ]);
</script>

{#snippet groupHeader(group: string, rows: Item[])}
  <div class="flex flex-wrap items-center gap-2">
    <span class="font-semibold text-[var(--ha-primary-text-color)]">{group || "—"}</span>
    <span
      class="ml-auto rounded-full bg-[var(--ha-card-background-color)] px-2 py-0.5 text-[10px] tabular-nums text-[var(--ha-secondary-text-color)]"
      aria-label={t("node.selection.group_count", { count: rows.length })}>{rows.length}</span
    >
  </div>
{/snippet}

{#snippet itemCell(item: Item, col: DataColumn<Item>)}
  {#if col.key === "select"}
    <label class="flex items-center justify-center">
      <input
        type="checkbox"
        checked={selected.has(item.key)}
        disabled={!bulkable(item)}
        title={bulkable(item) ? undefined : t("node.selection.unmappable_checkbox_title")}
        onclick={(e) => {
          e.stopPropagation();
          toggleSelect(item);
        }}
        class="h-5 w-5 cursor-pointer disabled:cursor-not-allowed disabled:opacity-40"
        aria-label={t("node.selection.select_row")}
      />
    </label>
  {:else if col.key === "state"}
    {@const si = stateInfo(item)}
    <span class="inline-flex justify-center"><Icon name={si.icon} size={18} class={si.color} aria-label={si.label} title={si.label} /></span>
  {:else if col.key === "name"}
    <button type="button" class="text-left hover:underline" onclick={() => openDrawer(item)}>
      <span class="text-[var(--ha-primary-text-color)]">{current(item).name}</span>
      {#if node.pending.has(item.key)}
        <span class="ml-1 text-xs text-[var(--ha-primary-color)]">{t("common.modified")}</span>
      {/if}
    </button>
  {:else if col.key === "device_type"}
    <span class="text-[var(--ha-secondary-text-color)]">{item.verdict.device_type_name || "—"}</span>
  {:else if multiColumn.has(col.key)}
    {@const values = itemValues(item, col.key.slice("field:".length))}
    {#if values.length === 0}
      <span class="text-[var(--ha-secondary-text-color)]">—</span>
    {:else}
      <span class="inline-flex flex-wrap gap-1">
        {#each values as v, i (i)}<Badge variant="muted">{v}</Badge>{/each}
      </span>
    {/if}
  {:else}
    <span class="text-[var(--ha-secondary-text-color)]">{col.get?.(item) || "—"}</span>
  {/if}
{/snippet}

{#if unavailable}
  <EmptyState message={t("node.selection.unavailable")} />
{:else if node.exposureLoading && !schema}
  <LoadingState message={t("common.loading")} />
{:else if node.exposureError}
  <ErrorState message={node.exposureError} onRetry={() => void node.loadExposure()} />
{:else if schema}
  <HostWarnings warnings={node.status?.warnings ?? []} collapsed />
  <div class="mb-3 flex flex-col gap-2">
    <div class="flex flex-wrap items-center gap-2">
      <Input placeholder={t("node.selection.search_placeholder")} bind:value={search} class="w-full sm:w-64" />
      {#each valueFilters as f (f.field)}
        <Select
          class="w-auto"
          value={filterValues[f.field] ?? ALL}
          ariaLabel={f.label}
          onValueChange={(v) => (filterValues = { ...filterValues, [f.field]: v })}
          options={[{ value: ALL, label: `${f.label}: ${t("node.selection.filter_all")}` }, ...f.options]}
        />
      {/each}
      {#if unmappableCount > 0}
        <label class="ml-auto flex cursor-pointer items-center gap-1.5 text-xs text-[var(--ha-secondary-text-color)]">
          <input type="checkbox" bind:checked={showUnmappable} class="h-4 w-4 cursor-pointer" />
          {t("node.selection.show_unmappable", { count: unmappableCount })}
        </label>
      {/if}
    </div>
    <div class="flex flex-wrap items-center gap-2">
      {#if selected.size > 0}
        <Button size="sm" onclick={() => bulkSet(true)}>{t("node.selection.bulk_expose")}</Button>
        <Button size="sm" variant="outline" onclick={() => bulkSet(false)}>{t("node.selection.bulk_hide")}</Button>
        <Button size="sm" variant="ghost" onclick={() => (selected = new Set())}>{t("common.cancel")}</Button>
      {:else}
        <Button size="sm" variant="ghost" onclick={selectAll}>{t("node.selection.select_all")}</Button>
      {/if}
      {#if node.hasDirty}
        <div class="ml-auto flex flex-wrap items-center gap-2" data-testid="save-bar">
          <span class="text-xs text-[var(--ha-secondary-text-color)]">{t("node.selection.pending", { count: node.pending.size })}</span>
          <Button size="sm" disabled={previewing} onclick={() => void preview()}>{t("node.selection.preview")}</Button>
          <Button size="sm" variant="outline" onclick={discard}>{t("node.selection.discard")}</Button>
        </div>
      {/if}
    </div>
  </div>

  {#if previewItems}
    <Card class="mb-4 p-4" data-testid="preview">
      <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("node.selection.preview_title")}</h3>
      <p class="mt-1 text-xs text-[var(--ha-secondary-text-color)]">{t("node.selection.preview_hint")}</p>
      <ul class="mt-3 space-y-2">
        {#each previewItems as after (after.key)}
          {@const was = before(after.key)}
          <li class="flex flex-wrap items-center gap-2 text-sm">
            <span class="font-medium text-[var(--ha-primary-text-color)]">{after.name}</span>
            <Badge variant={after.enabled ? "success" : "muted"}>
              {after.enabled ? t("node.selection.state_exposed") : t("node.selection.state_available")}
            </Badge>
            <span class="text-xs text-[var(--ha-secondary-text-color)]">{after.verdict.device_type_name || "—"} · {verdictHint(after.verdict.state)}</span>
            {#if verdictMoved(was, after)}
              <Badge variant="warning">{t("node.selection.preview_changed")}</Badge>
            {/if}
            {#if after.verdict.reason}
              <span class="w-full text-xs text-[var(--ha-secondary-text-color)]">{after.verdict.reason}</span>
            {/if}
          </li>
        {/each}
      </ul>
      <div class="mt-4 flex gap-2">
        <Button size="sm" disabled={saving} onclick={() => void apply()}>{saving ? t("common.saving") : t("node.selection.apply")}</Button>
        <Button size="sm" variant="outline" onclick={() => (previewItems = null)}>{t("node.selection.back")}</Button>
      </div>
    </Card>
  {/if}

  {#if filtered.length === 0}
    <EmptyState message={t("node.selection.empty")} icon="mdi:list-checks" />
  {:else}
    <Card class="p-4">
      <div class="mb-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-[var(--ha-secondary-text-color)]">
        <span class="font-semibold">{t("node.selection.legend")}:</span>
        {#each legend as li (li.label)}
          <span class="inline-flex items-center gap-1"><Icon name={li.icon} size={14} class={li.color} /><span>{li.label}</span></span>
        {/each}
      </div>
      <DataTable
        rows={filtered}
        {columns}
        rowKey={(i) => i.key}
        cell={itemCell}
        groupBy={schema.groups ? (i) => i.group : undefined}
        groupHeader={schema.groups ? groupHeader : undefined}
        rowClass={(i) => (selected.has(i.key) ? "bg-[var(--ha-secondary-background-color)]" : "")}
        persistKey="node-selection"
        emptyMessage={t("node.selection.empty")}
      />
    </Card>
  {/if}
{:else}
  <EmptyState message={t("node.selection.unavailable")} />
{/if}

{#if drawer}
  {@const item = current(drawer)}
  <div
    class="fixed inset-0 z-40 bg-black/30"
    onclick={() => (drawer = null)}
    onkeydown={(e) => {
      if (e.key === "Escape") drawer = null;
    }}
    role="presentation"
    tabindex="-1"
  ></div>
  <aside
    class="fixed right-0 top-0 z-50 flex h-full w-full max-w-sm flex-col overflow-y-auto border-l border-[var(--ha-divider-color)] bg-[var(--ha-card-background-color)]"
    aria-label={t("node.selection.drawer_aria")}
  >
    <div class="flex items-center justify-between border-b border-[var(--ha-divider-color)] p-4">
      <h2 class="text-base font-semibold text-[var(--ha-primary-text-color)]">{drawer.name}</h2>
      <button type="button" class="rounded-md p-2 hover:bg-[var(--ha-secondary-background-color)]" onclick={() => (drawer = null)} aria-label={t("common.close")}>✕</button>
    </div>
    <div class="flex-1 space-y-4 p-4">
      <div>
        <label for="fui-drawer-name" class="mb-1 block text-xs font-semibold text-[var(--ha-secondary-text-color)]">{t("node.selection.name")}</label>
        <Input id="fui-drawer-name" bind:value={drawerName} placeholder={drawer.name} />
      </div>
      <div class="flex items-center gap-3">
        <input
          type="checkbox"
          id="fui-drawer-enabled"
          bind:checked={drawerEnabled}
          disabled={item.verdict.state === "unmappable"}
          title={item.verdict.state === "unmappable" ? t("node.selection.unmappable_checkbox_title") : undefined}
          class="cursor-pointer disabled:cursor-not-allowed disabled:opacity-40"
        />
        <label for="fui-drawer-enabled" class="text-sm text-[var(--ha-primary-text-color)]">{t("node.selection.exposed")}</label>
      </div>
      <div>
        <p class="mb-1 text-xs font-semibold text-[var(--ha-secondary-text-color)]">{t("node.selection.drawer_state")}</p>
        <p class="text-sm text-[var(--ha-primary-text-color)]">{verdictHint(item.verdict.state)}</p>
        {#if item.verdict.reason}
          <p class="mt-1 text-xs text-[var(--ha-secondary-text-color)]">{item.verdict.reason}</p>
        {/if}
      </div>
      {#if item.source}
        <div>
          <p class="mb-1 text-xs font-semibold text-[var(--ha-secondary-text-color)]">{t("node.selection.drawer_source")}</p>
          <p class="text-sm text-[var(--ha-primary-text-color)]">{item.source}</p>
        </div>
      {/if}
      {#if item.verdict.device_type}
        <div>
          <p class="mb-1 text-xs font-semibold text-[var(--ha-secondary-text-color)]">{t("node.selection.drawer_device_type")}</p>
          <p class="text-sm text-[var(--ha-primary-text-color)]">{item.verdict.device_type_name}</p>
        </div>
      {/if}
      {#if item.verdict.clusters.length > 0}
        <div>
          <p class="mb-1 text-xs font-semibold text-[var(--ha-secondary-text-color)]">{t("node.selection.drawer_clusters")}</p>
          <p class="font-mono text-sm text-[var(--ha-primary-text-color)]">
            {item.verdict.clusters.map((c) => `0x${c.toString(16).toUpperCase().padStart(4, "0")}`).join(" · ")}
          </p>
        </div>
      {/if}
    </div>
    <div class="flex gap-2 border-t border-[var(--ha-divider-color)] p-4">
      <Button size="sm" onclick={saveDrawer}>{t("common.save")}</Button>
      <Button size="sm" variant="outline" onclick={() => (drawer = null)}>{t("common.cancel")}</Button>
    </div>
  </aside>
{/if}
