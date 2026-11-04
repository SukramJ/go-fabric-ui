<!-- What each controller ecosystem needs before it pairs a node: static
     hints with the links matter.js docs/ECOSYSTEMS.md gives (ecosystems.ts),
     beside the node's own compatibility (GET /node/compatibility): the
     ecosystems already holding a fabric are marked, with the findings
     against them. A node without a compatibility answer shows the hints
     alone. -->
<script lang="ts">
  import { onMount } from "svelte";
  import { useFabricUI } from "../../context";
  import type { Compatibility } from "../../api/client";
  import Badge from "../../kit/Badge.svelte";
  import Card from "../../kit/Card.svelte";
  import { ECOSYSTEM_HINTS } from "./ecosystems";

  const { t, client } = useFabricUI();

  let compat = $state<Compatibility | null>(null);

  onMount(async () => {
    try {
      compat = await client.compatibility();
    } catch {
      compat = null;
    }
  });

  const paired = $derived(new Set((compat?.ecosystems ?? []).map((e) => e.ecosystem)));
  const findings = (eco: string) => (compat?.findings ?? []).filter((f) => f.ecosystem === eco);
</script>

<Card class="mt-4 p-6" data-testid="ecosystem-hints">
  <h2 class="mb-1 text-base font-semibold text-[var(--ha-primary-text-color)]">{t("node.pair.hints.title")}</h2>
  <p class="mb-3 text-sm text-[var(--ha-secondary-text-color)]">{t("node.pair.hints.intro")}</p>
  <div class="space-y-2">
    {#each ECOSYSTEM_HINTS as hint (hint.ecosystem)}
      <details class="rounded border border-[var(--ha-divider-color)] px-3 py-2" data-testid="ecosystem-hint" open={paired.has(hint.ecosystem)}>
        <summary class="flex cursor-pointer flex-wrap items-center gap-2 text-sm font-medium text-[var(--ha-primary-text-color)]">
          {t(`node.diag.ecosystem.${hint.ecosystem}`)}
          {#if paired.has(hint.ecosystem)}
            <Badge variant="success">{t("node.pair.hints.paired")}</Badge>
          {/if}
        </summary>
        <p class="mt-2 text-sm text-[var(--ha-secondary-text-color)]">{t(`node.pair.hints.${hint.ecosystem}`)}</p>
        {#if hint.links.length > 0}
          <ul class="mt-1 flex flex-wrap gap-x-4 text-sm">
            {#each hint.links as link (link.key)}
              <li>
                <a class="text-[var(--ha-primary-color)] underline" href={link.href} target="_blank" rel="noopener noreferrer">{t(`node.pair.hints.${hint.ecosystem}.${link.key}`)}</a>
              </li>
            {/each}
          </ul>
        {/if}
        {#each findings(hint.ecosystem) as finding (finding.code + (finding.device_type ?? ""))}
          <p class="mt-2 flex flex-wrap items-start gap-2 text-sm">
            <Badge variant="warning">{t("node.pair.hints.finding")}</Badge>
            <span class="min-w-[12rem] flex-1 text-[var(--ha-primary-text-color)]">{finding.message}</span>
          </p>
        {/each}
      </details>
    {/each}
  </div>
</Card>
