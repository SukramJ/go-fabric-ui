<!-- The host's own counters (Status.counters, contract 0.8.0): a small
     table under the host's warnings on the status page — code, value and
     the host's one-line meaning, shown verbatim; nothing when the host
     reports none. -->
<script lang="ts">
  import { useFabricUI } from "../../context";
  import type { Counter } from "../../api/client";
  import Card from "../../kit/Card.svelte";

  type Props = { counters: Counter[] };
  let { counters }: Props = $props();

  const { t } = useFabricUI();
</script>

{#if counters.length > 0}
  <Card class="mb-4 p-4" data-testid="host-counters">
    <h3 class="mb-2 font-medium text-[var(--ha-primary-text-color)]">{t("node.counters.title")}</h3>
    <table class="w-full text-sm">
      <tbody>
        {#each counters as c, i (i)}
          <tr class="border-t border-[var(--ha-divider-color)] first:border-t-0">
            <td class="py-1 pr-3 font-mono text-xs text-[var(--ha-secondary-text-color)]">{c.code}</td>
            <td class="py-1 pr-3 text-right font-mono tabular-nums text-[var(--ha-primary-text-color)]">{c.value}</td>
            <td class="py-1 text-[var(--ha-primary-text-color)]">{c.message}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </Card>
{/if}
