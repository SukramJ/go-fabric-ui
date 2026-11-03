<!-- One server cluster's inventory: name, id, revision, the feature map in
     hex with its features as chips (code · title, the bit in the title),
     the data version, the attribute table and the command and event lists.
     Written for both families. -->
<script lang="ts">
  import { cn } from "./utils";
  import type { ClusterDetail, ClusterElement } from "../api/client";
  import { useT } from "../context";
  import AttributeTable from "./AttributeTable.svelte";
  import Badge from "./Badge.svelte";

  type Props = { cluster: ClusterDetail; class?: string };
  let { cluster, class: className }: Props = $props();
  const t = useT();

  const lists = $derived<[string, ClusterElement[]][]>([
    ["inventory.accepted", cluster.accepted],
    ["inventory.generated", cluster.generated],
    ["inventory.events", cluster.events],
  ]);
</script>

<div class={cn("space-y-3", className)} data-testid={`cluster-view-${cluster.id_hex}`}>
  <div class="flex flex-wrap items-center gap-2 text-sm">
    <span class="font-medium text-[var(--ha-primary-text-color)]">{cluster.name ?? cluster.id_hex}</span>
    <span class="font-mono text-xs text-[var(--ha-secondary-text-color)]">{cluster.id_hex}</span>
    <Badge variant="muted">{t("inventory.revision", { revision: cluster.revision })}</Badge>
    <Badge variant="muted" title={t("inventory.feature_map")}>{cluster.feature_map_hex}</Badge>
    {#if cluster.data_version !== undefined}
      <Badge variant="muted">{t("inventory.data_version", { version: cluster.data_version })}</Badge>
    {/if}
  </div>
  {#if cluster.features.length > 0}
    <ul class="flex flex-wrap gap-1" data-testid="feature-chips">
      {#each cluster.features as f (f.bit)}
        <li><Badge title={t("inventory.feature_bit", { bit: f.bit })}>{f.title ? `${f.code} · ${f.title}` : f.code}</Badge></li>
      {/each}
    </ul>
  {/if}
  <AttributeTable attributes={cluster.attributes} />
  {#each lists as [key, items] (key)}
    <div class="text-sm">
      <span class="text-xs text-[var(--ha-secondary-text-color)]">{t(key)}:</span>
      {#if items.length === 0}
        <span class="text-xs text-[var(--ha-secondary-text-color)]">{t("inventory.none")}</span>
      {:else}
        {#each items as e (e.id)}
          <Badge variant="muted" class="mr-1" title={e.id_hex}>{e.name ?? e.id_hex}</Badge>
        {/each}
      {/if}
    </div>
  {/each}
</div>
