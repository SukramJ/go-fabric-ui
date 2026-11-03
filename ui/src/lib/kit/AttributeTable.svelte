<!-- A cluster's attributes: id in hex, name, value as formatted JSON and
     the status it was answered with. Written for both families. -->
<script lang="ts">
  import { cn } from "./utils";
  import type { ClusterAttribute } from "../api/client";
  import { useT } from "../context";

  type Props = { attributes: ClusterAttribute[]; class?: string };
  let { attributes, class: className }: Props = $props();
  const t = useT();

  const format = (v: unknown) => JSON.stringify(v, null, 2);
  const statusHex = (s: number) => "0x" + s.toString(16).toUpperCase().padStart(2, "0");
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
          <td class="py-1 font-mono text-xs">{a.status !== undefined ? statusHex(a.status) : ""}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>
