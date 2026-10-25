<!-- The commissioned fabrics, removing one, and the maintenance actions
     (reassemble, remove all pairings). Ported from openccu-loom
     assets/ui/src/routes/matter/MatterFabrics.svelte:1-209 at b40b289a:
     the vendor name from the server, the node id from its exact hex twin,
     the confirmations before every destructive action. -->
<script lang="ts">
  import { onMount } from "svelte";
  import { useFabricUI } from "../../context";
  import { errorText } from "../../store.svelte";
  import type { Fabric } from "../../api/client";
  import type { DataColumn } from "../../kit/data-table";
  import Badge from "../../kit/Badge.svelte";
  import Button from "../../kit/Button.svelte";
  import Card from "../../kit/Card.svelte";
  import DataTable from "../../kit/DataTable.svelte";
  import ErrorState from "../../kit/ErrorState.svelte";
  import LoadingState from "../../kit/LoadingState.svelte";

  const { t, toast, confirm, client, node, routes } = useFabricUI();

  onMount(() => {
    void node.loadFabrics();
  });

  // The server owns the vendor table, so one list cannot drift from the
  // other; the raw id is the fallback.
  function vendorName(f: Fabric): string {
    return f.vendor_name || `0x${f.vendor_id.toString(16).toUpperCase().padStart(4, "0")}`;
  }

  // Controllers print node ids as 16-digit hex; the exact value travels in
  // node_id_hex, the number is only a fallback.
  function nodeIdHex(f: Fabric): string {
    if (f.node_id_hex) return `0x${f.node_id_hex.toUpperCase()}`;
    return `0x${f.node_id.toString(16).toUpperCase().padStart(16, "0")}`;
  }

  // A JSON number carries 53 bits: without the hex twin an id above that
  // was rounded before it reached the page.
  function nodeIdExact(f: Fabric): boolean {
    return Boolean(f.node_id_hex) || Number.isSafeInteger(f.node_id);
  }

  let busy = $state(false);

  async function reassemble() {
    busy = true;
    try {
      await client.reassemble();
      toast.success(t("node.maint.reassemble_done"));
      await node.loadStatus();
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      busy = false;
    }
  }

  async function factoryReset() {
    const ok = await confirm({
      title: t("node.maint.reset_confirm"),
      body: t("node.maint.reset_confirm_body"),
      destructive: true,
      confirmLabel: t("node.maint.reset_confirm_label"),
    });
    if (!ok) return;
    busy = true;
    try {
      await client.factoryReset();
      toast.success(t("node.maint.reset_done"));
      await node.loadFabrics();
      await node.loadStatus();
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      busy = false;
    }
  }

  async function unpair(f: Fabric) {
    const ok = await confirm({
      title: t("node.fabric.unpair_confirm"),
      body: f.label || t("node.fabric.label_unknown"),
      destructive: true,
      confirmLabel: t("common.remove"),
    });
    if (!ok) return;
    try {
      await client.removeFabric(f.fabric_index);
      toast.success(t("node.fabric.unpaired"));
      await node.loadFabrics();
      await node.loadStatus();
    } catch (err) {
      toast.error(errorText(err));
    }
  }

  const columns: DataColumn<Fabric>[] = $derived([
    { key: "vendor", label: t("node.fabrics.col_vendor"), sortable: true, title: true, get: (f) => vendorName(f) },
    { key: "label", label: t("node.fabrics.col_label"), sortable: true, get: (f) => f.label || t("node.fabric.label_unknown") },
    { key: "fabric", label: t("node.fabrics.col_fabric"), sortable: true, get: (f) => f.fabric_index },
    // Sorted on the zero-padded hex: fixed width makes lexical order numeric.
    { key: "node_id", label: t("node.fabrics.col_node_id"), sortable: true, get: (f) => nodeIdHex(f) },
    { key: "action", label: "", align: "right", cellClass: "reflow-actions" },
  ]);
</script>

<div class="space-y-6">
  {#if node.fabricsLoading && node.fabrics.length === 0}
    <LoadingState message={t("common.loading")} />
  {:else if node.fabricsError}
    <ErrorState message={node.fabricsError} onRetry={() => void node.loadFabrics()} />
  {:else}
    <Card class="p-4">
      <DataTable
        rows={node.fabrics}
        {columns}
        rowKey={(f) => String(f.fabric_index)}
        persistKey="node-fabrics"
        emptyMessage={t("node.fabrics.empty")}
        emptyIcon="mdi:link"
        initialSort={{ key: "vendor", asc: true }}
      >
        {#snippet cell(fabric, col)}
          {#if col.key === "vendor"}
            <span class="text-[var(--ha-primary-text-color)]">{vendorName(fabric)}</span>
          {:else if col.key === "label"}
            <span class={fabric.label ? "text-[var(--ha-primary-text-color)]" : "text-[var(--ha-secondary-text-color)]"}>
              {fabric.label || t("node.fabric.label_unknown")}
            </span>
          {:else if col.key === "fabric"}
            <span class="text-[var(--ha-secondary-text-color)]">{fabric.fabric_index}</span>
          {:else if col.key === "node_id"}
            <span class="inline-flex flex-wrap items-center gap-1">
              <span class="font-mono text-xs text-[var(--ha-secondary-text-color)]">{nodeIdHex(fabric)}</span>
              {#if !nodeIdExact(fabric)}
                <Badge variant="warning" title={t("node.fabrics.node_id_rounded_hint")}>{t("node.fabrics.node_id_rounded")}</Badge>
              {/if}
            </span>
          {:else if col.key === "action"}
            <Button size="sm" variant="ghost" class="text-[var(--ha-error-color)]" onclick={() => void unpair(fabric)}>
              {t("common.remove")}
            </Button>
          {/if}
        {/snippet}
      </DataTable>
    </Card>
  {/if}

  <Card class="p-4">
    <h3 class="mb-2 font-medium text-[var(--ha-primary-text-color)]">{t("node.maint.title")}</h3>
    <div class="flex flex-wrap items-center gap-3">
      <Button variant="outline" onclick={() => void reassemble()} disabled={busy}>{t("node.maint.reassemble")}</Button>
      <span class="text-xs text-[var(--ha-secondary-text-color)]">{t("node.maint.reassemble_hint")}</span>
    </div>
    <div class="mt-3 flex flex-wrap items-center gap-3">
      <Button variant="outline-destructive" onclick={() => void factoryReset()} disabled={busy}>{t("node.maint.reset")}</Button>
      <span class="text-xs text-[var(--ha-secondary-text-color)]">{t("node.maint.reset_hint")}</span>
    </div>
  </Card>

  <div class="rounded-[var(--ha-radius-card)] border border-[var(--ha-divider-color)] bg-[var(--ha-secondary-background-color)] p-4">
    <h3 class="mb-2 font-medium text-[var(--ha-primary-text-color)]">{t("node.fabric.share")}</h3>
    <p class="text-sm text-[var(--ha-secondary-text-color)]">{t("node.fabric.share_hint")}</p>
    <a href={routes.pair} data-testid="share-link" class="mt-3 inline-flex items-center text-sm font-medium text-[var(--ha-primary-color)] hover:underline">
      {t("node.fabric.share_go")}
    </a>
  </div>
</div>
