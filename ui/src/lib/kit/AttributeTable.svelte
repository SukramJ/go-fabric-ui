<!-- A cluster's attributes: id in hex, name, value as formatted JSON and
     the status it was answered with — by name for a status of the
     Interaction Model's (IM_STATUS), with its code; the code alone
     otherwise. Written for both families. -->
<script lang="ts">
  import { cn } from "./utils";
  import Badge from "./Badge.svelte";
  import { IM_STATUS, imStatusName, statusHex } from "./im-status";
  import type { ClusterAttribute } from "../api/client";
  import { useT } from "../context";

  type Props = { attributes: ClusterAttribute[]; class?: string };
  let { attributes, class: className }: Props = $props();
  const t = useT();

  const format = (v: unknown) => JSON.stringify(v, null, 2);
</script>

<div class={cn("overflow-x-auto", className)}>
  <table class="w-full text-left text-sm" data-testid="attribute-table">
    <thead class="text-xs text-[var(--ha-secondary-text-color)]">
      <tr>
        <th class="py-1 pr-3 font-medium">{t("inventory.column.id")}</th>
        <th class="py-1 pr-3 font-medium">{t("inventory.column.name")}</th>
        <th class="py-1 pr-3 font-medium">{t("inventory.column.value")}</th>
        <th class="py-1 font-medium">{t("inventory.column.status")}</th>
      </tr>
    </thead>
    <tbody>
      {#each attributes as a (a.id)}
        <tr class="border-t border-[var(--ha-divider-color)] align-top" data-testid={`attribute-${a.id_hex}`}>
          <td class="py-1 pr-3 font-mono text-xs">{a.id_hex}</td>
          <td class="py-1 pr-3">{a.name ?? ""}</td>
          <td class="py-1 pr-3">
            {#if a.value !== undefined}
              <pre class="whitespace-pre-wrap break-all font-mono text-xs">{format(a.value)}</pre>
            {:else}
              <span class="text-xs text-[var(--ha-secondary-text-color)]">{t("inventory.no_value")}</span>
            {/if}
          </td>
          <td class="py-1 text-xs" data-testid={a.status !== undefined ? "attribute-status" : undefined}>
            {#if a.status !== undefined}
              {#if IM_STATUS[a.status]}
                <Badge variant="warning" title={`${imStatusName(a.status)} (${statusHex(a.status)})`}>{t(`inventory.status.${IM_STATUS[a.status]}`)}</Badge>
              {/if}
              <span class="font-mono">{statusHex(a.status)}</span>
            {/if}
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>
