<!-- The commissioned nodes: a table with filters by connection, vendor and
     ICD, and the actions an administrator takes on a node — interview,
     open a window for another administrator (on the node's page),
     disable or enable, decommission. A row opens the node's page. -->
<script lang="ts">
  import { onMount } from "svelte";
  import { useFabricUI } from "../../context";
  import { errorText } from "../../store.svelte";
  import type { ControllerNodeSummary } from "../../api/client";
  import type { DataColumn } from "../../kit/data-table";
  import Badge from "../../kit/Badge.svelte";
  import Button from "../../kit/Button.svelte";
  import Card from "../../kit/Card.svelte";
  import DataTable from "../../kit/DataTable.svelte";
  import ErrorState from "../../kit/ErrorState.svelte";
  import LoadingState from "../../kit/LoadingState.svelte";
  import Select from "../../kit/Select.svelte";
  import { connectionLabel, connectionVariant, nodeName, vendorLabel } from "./format";

  const { t, toast, confirm, client, controller, routes } = useFabricUI();

  onMount(() => {
    void controller.loadNodes();
  });

  let connection = $state("");
  let vendor = $state("");
  let icd = $state("");
  let busy = $state<string | null>(null);

  const vendors = $derived([...new Set(controller.nodes.map((n) => vendorLabel(n.vendor_name, n.vendor_id)).filter(Boolean))].sort());

  const rows = $derived(
    controller.nodes.filter(
      (n) =>
        (!connection || n.connection === connection) &&
        (!vendor || vendorLabel(n.vendor_name, n.vendor_id) === vendor) &&
        (!icd || (icd === "yes") === n.icd),
    ),
  );

  const connectionOptions = $derived([
    { value: "", label: t("controller.nodes.filter_all") },
    ...(["connected", "reconnecting", "waiting_for_discovery", "disconnected"] as const).map((s) => ({ value: s, label: connectionLabel(t, s) })),
  ]);
  const vendorOptions = $derived([{ value: "", label: t("controller.nodes.filter_all") }, ...vendors.map((v) => ({ value: v, label: v }))]);
  const icdOptions = $derived([
    { value: "", label: t("controller.nodes.filter_all") },
    { value: "yes", label: t("controller.nodes.icd_yes") },
    { value: "no", label: t("controller.nodes.icd_no") },
  ]);

  function open(n: ControllerNodeSummary) {
    location.hash = routes.controllerNode + n.node_id_hex;
  }

  async function act(n: ControllerNodeSummary, run: () => Promise<void>, done: string) {
    busy = n.node_id_hex;
    try {
      await run();
      toast.success(t(done));
      await controller.loadNodes();
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      busy = null;
    }
  }

  async function decommission(n: ControllerNodeSummary) {
    const ok = await confirm({
      title: t("controller.nodes.remove_confirm"),
      body: t("controller.nodes.remove_confirm_body", { name: nodeName(n) }),
      destructive: true,
      confirmLabel: t("common.remove"),
    });
    if (ok) await act(n, () => client.decommissionNode(n.node_id_hex), "controller.nodes.removed");
  }

  const columns: DataColumn<ControllerNodeSummary>[] = $derived([
    { key: "name", label: t("controller.nodes.col_name"), sortable: true, title: true, get: (n) => nodeName(n) },
    { key: "vendor", label: t("controller.nodes.col_vendor"), sortable: true, get: (n) => vendorLabel(n.vendor_name, n.vendor_id) },
    { key: "node_id", label: t("controller.nodes.col_node_id"), sortable: true, get: (n) => n.node_id_hex },
    { key: "connection", label: t("controller.nodes.col_connection"), sortable: true, get: (n) => n.connection },
    { key: "addresses", label: t("controller.nodes.col_addresses"), get: (n) => n.addresses.join(" ") },
    { key: "action", label: "", align: "right", cellClass: "reflow-actions" },
  ]);
</script>

<div class="space-y-4">
  <div class="flex flex-wrap items-center gap-3">
    <Select class="w-auto" ariaLabel={t("controller.nodes.filter_connection")} value={connection} onValueChange={(v) => (connection = v as typeof connection)} options={connectionOptions} />
    <Select class="w-auto" ariaLabel={t("controller.nodes.filter_vendor")} value={vendor} onValueChange={(v) => (vendor = v as typeof vendor)} options={vendorOptions} />
    <Select class="w-auto" ariaLabel={t("controller.nodes.filter_icd")} value={icd} onValueChange={(v) => (icd = v as typeof icd)} options={icdOptions} />
    <a href={routes.controllerCommission} class="ml-auto text-sm font-medium text-[var(--ha-primary-color)] hover:underline">{t("controller.nodes.commission")}</a>
  </div>
  {#if controller.nodesLoading && controller.nodes.length === 0}
    <LoadingState message={t("common.loading")} />
  {:else if controller.nodesError}
    <ErrorState message={controller.nodesError} onRetry={() => void controller.loadNodes()} />
  {:else}
    <Card class="p-4">
      <DataTable
        {rows}
        {columns}
        rowKey={(n) => n.node_id_hex}
        persistKey="controller-nodes"
        search
        searchPlaceholder={t("common.search")}
        emptyMessage={t("controller.nodes.empty")}
        initialSort={{ key: "name", asc: true }}
        onRowClick={open}
      >
        {#snippet cell(n, col)}
          {#if col.key === "name"}
            <span class="inline-flex flex-wrap items-center gap-1 text-[var(--ha-primary-text-color)]">
              {nodeName(n)}
              {#if n.icd}<Badge variant="muted">{t("controller.nodes.icd")}</Badge>{/if}
              {#if n.disabled}<Badge variant="warning">{t("controller.nodes.disabled")}</Badge>{/if}
            </span>
          {:else if col.key === "vendor"}
            <span class="text-[var(--ha-secondary-text-color)]">{vendorLabel(n.vendor_name, n.vendor_id)}</span>
          {:else if col.key === "node_id"}
            <span class="font-mono text-xs text-[var(--ha-secondary-text-color)]">0x{n.node_id_hex}</span>
          {:else if col.key === "connection"}
            <span class="inline-flex flex-wrap items-center gap-1">
              <Badge variant={connectionVariant(n.connection)}>{connectionLabel(t, n.connection)}</Badge>
              {#if n.subscription_active}<Badge variant="default">{t("controller.nodes.subscribed")}</Badge>{/if}
            </span>
          {:else if col.key === "addresses"}
            <span class="font-mono text-xs text-[var(--ha-secondary-text-color)]">{n.addresses.join(", ")}</span>
          {:else if col.key === "action"}
            <span class="inline-flex flex-wrap justify-end gap-1">
              <Button size="sm" variant="ghost" disabled={busy === n.node_id_hex} onclick={() => void act(n, () => client.interviewNode(n.node_id_hex), "controller.nodes.interviewed")}>
                {t("controller.nodes.interview")}
              </Button>
              <a class="inline-flex items-center px-2 text-sm text-[var(--ha-primary-color)] hover:underline" href={`${routes.controllerNode}${n.node_id_hex}?tab=window`}>{t("controller.nodes.window")}</a>
              {#if n.disabled}
                <Button size="sm" variant="ghost" disabled={busy === n.node_id_hex} onclick={() => void act(n, () => client.enableNode(n.node_id_hex), "controller.nodes.enabled")}>
                  {t("controller.nodes.enable")}
                </Button>
              {:else}
                <Button size="sm" variant="ghost" disabled={busy === n.node_id_hex} onclick={() => void act(n, () => client.disableNode(n.node_id_hex), "controller.nodes.disabled_done")}>
                  {t("controller.nodes.disable")}
                </Button>
              {/if}
              <Button size="sm" variant="ghost" class="text-[var(--ha-error-color)]" disabled={busy === n.node_id_hex} onclick={() => void decommission(n)}>
                {t("common.remove")}
              </Button>
            </span>
          {/if}
        {/snippet}
      </DataTable>
    </Card>
  {/if}
</div>
