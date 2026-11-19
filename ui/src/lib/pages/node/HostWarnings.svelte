<!-- The host's findings about its own configuration (Status.warnings):
     a card on the status page, a collapsed list above the selection. -->
<script lang="ts">
  import { useFabricUI } from "../../context";
  import type { Warning } from "../../api/client";
  import Badge from "../../kit/Badge.svelte";
  import Card from "../../kit/Card.svelte";
  import { formatDate } from "../../kit/dates";

  type Props = { warnings: Warning[]; collapsed?: boolean };
  let { warnings, collapsed = false }: Props = $props();

  const { t } = useFabricUI();
  // Without admin rights a warning carries its code only (decision 34):
  // one line says why, instead of a row of bare codes (UR-P4-ui-65).
  const bare = $derived(warnings.length > 0 && warnings.every((w) => !w.message));
</script>

{#snippet list()}
  <ul class="space-y-2">
    {#each warnings as w, i (i)}
      <li class="flex flex-wrap items-center gap-2 text-sm">
        <Badge variant="warning">{w.code}</Badge>
        {#if w.message}<span class="text-[var(--ha-primary-text-color)]">{w.message}</span>{/if}
        {#if w.since}
          <span class="text-xs text-[var(--ha-secondary-text-color)]">{t("node.warnings.since", { time: formatDate(t, w.since) })}</span>
        {/if}
      </li>
    {/each}
  </ul>
  {#if bare}
    <p class="mt-2 text-xs text-[var(--ha-secondary-text-color)]" data-testid="host-warnings-viewer">{t("node.warnings.viewer")}</p>
  {/if}
{/snippet}

{#if warnings.length > 0}
  {#if collapsed}
    <details class="mb-3 rounded-md border border-[var(--ha-divider-color)] px-3 py-2" data-testid="host-warnings">
      <summary class="cursor-pointer text-sm font-medium text-[var(--fui-warning-text)]">{t("node.warnings.summary", { count: warnings.length })}</summary>
      <div class="mt-2">{@render list()}</div>
    </details>
  {:else}
    <Card class="mb-4 p-4" data-testid="host-warnings">
      <h3 class="mb-2 font-medium text-[var(--ha-primary-text-color)]">{t("node.warnings.title")}</h3>
      {@render list()}
    </Card>
  {/if}
{/if}
