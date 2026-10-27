<!-- Every fabric on the node, ours marked; removing another
     administrator's. Ours goes with decommissioning. -->
<script lang="ts">
  import { onMount } from "svelte";
  import { useFabricUI } from "../../../context";
  import { errorText } from "../../../store.svelte";
  import type { NodeFabric } from "../../../api/client";
  import type { DataColumn } from "../../../kit/data-table";
  import Badge from "../../../kit/Badge.svelte";
  import Button from "../../../kit/Button.svelte";
  import Card from "../../../kit/Card.svelte";
  import DataTable from "../../../kit/DataTable.svelte";
  import ErrorState from "../../../kit/ErrorState.svelte";
  import LoadingState from "../../../kit/LoadingState.svelte";
  import { vendorLabel } from "../format";

  let { nodeId }: { nodeId: string } = $props();
  const { t, toast, confirm, client } = useFabricUI();

  let fabrics = $state<NodeFabric[] | null>(null);
  let error = $state<string | null>(null);

  async function load() {
    error = null;
    try {
      fabrics = (await client.nodeFabrics(nodeId)).fabrics;
    } catch (err) {
      error = errorText(err);
    }
  }
  onMount(() => void load());

  async function remove(f: NodeFabric) {
    const ok = await confirm({
      title: t("controller.fabrics.remove_confirm"),
      body: f.label || vendorLabel(f.vendor_name, f.vendor_id),
      destructive: true,
      confirmLabel: t("common.remove"),
    });
    if (!ok) return;
    try {
      await client.removeNodeFabric(nodeId, f.fabric_index);
      toast.success(t("controller.fabrics.removed"));
      await load();
    } catch (err) {
      toast.error(errorText(err));
    }
  }

  const columns: DataColumn<NodeFabric>[] = $derived([
    { key: "vendor", label: t("node.fabrics.col_vendor"), title: true, sortable: true, get: (f) => vendorLabel(f.vendor_name, f.vendor_id) },
    { key: "label", label: t("node.fabrics.col_label"), sortable: true, get: (f) => f.label },
    { key: "fabric", label: t("node.fabrics.col_fabric"), sortable: true, get: (f) => f.fabric_index },
    { key: "fabric_id", label: t("controller.fabrics.col_fabric_id"), get: (f) => f.fabric_id_hex },
    { key: "action", label: "", align: "right", cellClass: "reflow-actions" },
  ]);
</script>

{#if error}
  <ErrorState message={error} onRetry={() => void load()} />
{:else if !fabrics}
  <LoadingState message={t("common.loading")} />
{:else}
  <Card class="p-4">
    <DataTable rows={fabrics} {columns} rowKey={(f) => String(f.fabric_index)} emptyMessage={t("node.fabrics.empty")} initialSort={{ key: "fabric", asc: true }}>
      {#snippet cell(f, col)}
        {#if col.key === "vendor"}
          <span class="inline-flex items-center gap-1">{vendorLabel(f.vendor_name, f.vendor_id)}{#if f.own}<Badge>{t("controller.fabrics.own")}</Badge>{/if}</span>
        {:else if col.key === "label"}
          {f.label}
        {:else if col.key === "fabric"}
          {f.fabric_index}
        {:else if col.key === "fabric_id"}
          <span class="font-mono text-xs">0x{f.fabric_id_hex}</span>
        {:else if col.key === "action"}
          {#if !f.own}
            <Button size="sm" variant="ghost" class="text-[var(--ha-error-color)]" onclick={() => void remove(f)}>{t("common.remove")}</Button>
          {/if}
        {/if}
      {/snippet}
    </DataTable>
  </Card>
{/if}
