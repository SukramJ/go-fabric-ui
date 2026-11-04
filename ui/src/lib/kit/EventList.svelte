<!-- A list of Matter events (contract 0.4.1, EventRecord), newest first:
     the event number, its time, the priority as a badge, the endpoint, the
     cluster and the event's name, and its data as formatted JSON. The
     routes answer oldest first; the list turns them round. Written for
     both families (the node's endpoint page, the controller's events tab). -->
<script lang="ts">
  import { cn } from "./utils";
  import type { EventRecord } from "../api/client";
  import { useT } from "../context";
  import Badge from "./Badge.svelte";

  // empty is the text without events.
  type Props = { events: EventRecord[]; empty: string; class?: string };
  let { events, empty, class: className }: Props = $props();
  const t = useT();

  const newestFirst = $derived([...events].reverse());
  const variant = (p: EventRecord["priority"]) => (p === "critical" ? "danger" : p === "info" ? "default" : "muted");
  const time = (ms: number) => new Date(ms).toLocaleString();
  const format = (v: unknown) => JSON.stringify(v, null, 2);
</script>

{#if events.length === 0}
  <p class={cn("text-xs text-[var(--ha-secondary-text-color)]", className)} data-testid="event-list-empty">{empty}</p>
{:else}
  <ul class={cn("divide-y divide-[var(--ha-divider-color)]", className)} data-testid="event-list">
    {#each newestFirst as e (e.number)}
      <li class="py-2 text-sm" data-testid={`event-${e.number}`}>
        <div class="flex flex-wrap items-center gap-2">
          <span class="font-mono text-xs text-[var(--ha-secondary-text-color)]">#{e.number}</span>
          <Badge variant={variant(e.priority)}>{t(`events.priority.${e.priority}`)}</Badge>
          <span class="font-medium text-[var(--ha-primary-text-color)]" title={`${e.cluster_hex} · ${e.event_hex}`}>{e.name ?? e.event_hex}</span>
          <span class="font-mono text-xs text-[var(--ha-secondary-text-color)]">{e.cluster_hex} · {e.event_hex}</span>
          <span class="text-xs text-[var(--ha-secondary-text-color)]">{t("events.endpoint", { endpoint: e.endpoint })}</span>
          {#if e.epoch_ms !== undefined}
            <time class="text-xs text-[var(--ha-secondary-text-color)]" datetime={new Date(e.epoch_ms).toISOString()}>{time(e.epoch_ms)}</time>
          {/if}
        </div>
        {#if e.data !== undefined}
          <pre class="mt-1 whitespace-pre-wrap break-all font-mono text-xs">{format(e.data)}</pre>
        {/if}
      </li>
    {/each}
  </ul>
{/if}
