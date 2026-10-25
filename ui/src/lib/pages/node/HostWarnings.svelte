<!-- The host's findings about its own configuration (Status.warnings):
     a card on the status page, a collapsed list above the selection. -->
<script lang="ts">
  import { useFabricUI } from "../../context";
  import type { Warning } from "../../api/client";
  import Badge from "../../kit/Badge.svelte";
  import Card from "../../kit/Card.svelte";

  type Props = { warnings: Warning[]; collapsed?: boolean };
  let { warnings, collapsed = false }: Props = $props();

  const { t } = useFabricUI();
</script>

{#snippet list()}
  <ul class="space-y-2">
    {#each warnings as w, i (i)}
      <li class="flex flex-wrap items-center gap-2 text-sm">
        <Badge variant="warning">{w.code}</Badge>
        <span class="text-[var(--ha-primary-text-color)]">{w.message}</span>
        {#if w.since}
          <span class="text-xs text-[var(--ha-secondary-text-color)]">{t("node.warnings.since", { time: new Date(w.since).toLocaleString() })}</span>
        {/if}
      </li>
    {/each}
  </ul>
{/snippet}

{#if warnings.length > 0}
  {#if collapsed}
    <details class="mb-3 rounded-md border border-[var(--ha-divider-color)] px-3 py-2" data-testid="host-warnings">
      <summary class="cursor-pointer text-sm font-medium text-[var(--ha-warning-color)]">{t("node.warnings.summary", { count: warnings.length })}</summary>
      <div class="mt-2">{@render list()}</div>
    </details>
  {:else}
    <Card class="mb-4 p-4" data-testid="host-warnings">
      <h3 class="mb-2 font-medium text-[var(--ha-primary-text-color)]">{t("node.warnings.title")}</h3>
      {@render list()}
    </Card>
  {/if}
{/if}
