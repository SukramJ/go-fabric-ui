<!-- A node's endpoints as a tree: parts under their parent, each with the
     device types of its DeviceTypeList, Reachable where it is bridged, its
     server clusters and a link into the inventory tab. -->
<script lang="ts">
  import { useFabricUI } from "../../../context";
  import type { ControllerNodeDetail, NodeEndpoint } from "../../../api/client";
  import Badge from "../../../kit/Badge.svelte";
  import Card from "../../../kit/Card.svelte";
  import EmptyState from "../../../kit/EmptyState.svelte";

  let { detail, onInventory }: { detail: ControllerNodeDetail; onInventory?: (endpoint: number) => void } = $props();
  const { t } = useFabricUI();

  // Depth first, parts under their parent, each level in number order.
  const ordered = $derived.by(() => {
    const out: { ep: NodeEndpoint; depth: number }[] = [];
    const children = (parent: number | undefined) => detail.endpoints.filter((e) => e.parent === parent);
    const walk = (ep: NodeEndpoint, depth: number) => {
      out.push({ ep, depth });
      for (const c of children(ep.endpoint)) walk(c, depth + 1);
    };
    for (const top of children(undefined)) walk(top, 0);
    return out;
  });
</script>

{#if detail.endpoints.length === 0}
  <EmptyState message={t("controller.endpoints.empty")} />
{:else}
  <div class="space-y-3">
    {#each ordered as { ep, depth } (ep.endpoint)}
      <Card class="p-4" style={`margin-left: ${depth * 1.5}rem`} data-testid={`endpoint-${ep.endpoint}`}>
        <div class="flex flex-wrap items-center gap-2">
          <span class="font-medium text-[var(--ha-primary-text-color)]">{t("controller.endpoints.endpoint", { number: ep.endpoint })}</span>
          {#if ep.name}<span class="text-sm text-[var(--ha-secondary-text-color)]">{ep.name}</span>{/if}
          {#each ep.device_types as dt (dt.id)}
            <Badge title={dt.id_hex}>{dt.name || dt.id_hex}{dt.revision ? ` r${dt.revision}` : ""}</Badge>
          {/each}
          {#if ep.reachable !== undefined}
            <Badge variant={ep.reachable ? "success" : "danger"}>{t(ep.reachable ? "controller.endpoints.reachable" : "controller.endpoints.unreachable")}</Badge>
          {/if}
          {#if onInventory}
            <button
              type="button"
              class="ml-auto text-sm text-[var(--ha-primary-color)] hover:underline"
              data-testid={`inventory-link-${ep.endpoint}`}
              onclick={() => onInventory(ep.endpoint)}>{t("controller.endpoints.inventory")}</button
            >
          {/if}
        </div>
        {#if ep.clusters.length > 0}
          <ul class="mt-2 flex flex-wrap gap-1">
            {#each ep.clusters as c (c.id)}
              <li><Badge variant="muted" title={`${c.id_hex} r${c.revision}${c.features.length ? " · " + c.features.join(", ") : ""}`}>{c.name || c.id_hex}</Badge></li>
            {/each}
          </ul>
        {/if}
      </Card>
    {/each}
  </div>
{/if}
