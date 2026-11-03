<!-- The node as a graph (GET /node/topology): the node, its Aggregator,
     the bridged devices and their parts on one side, the fabrics and the
     sessions their controllers hold on the other. A node's link leads to
     the page that already lists it. The graph reloads when the stream
     reports a session, a fabric or the endpoints changing (the store's
     graph epoch, one move per burst). New in brief U3a; no loom page
     precedes it. -->
<script lang="ts">
  import { onMount, untrack } from "svelte";
  import { refHref, useFabricUI } from "../../context";
  import { errorText } from "../../store.svelte";
  import type { TopologyGraph, TopologyNode } from "../../api/client";
  import Button from "../../kit/Button.svelte";
  import Card from "../../kit/Card.svelte";
  import EmptyState from "../../kit/EmptyState.svelte";
  import ErrorState from "../../kit/ErrorState.svelte";
  import Graph from "../../kit/Graph.svelte";
  import LoadingState from "../../kit/LoadingState.svelte";

  type Props = { follow?: (href: string) => void };
  let { follow }: Props = $props();

  const { t, client, node, routes } = useFabricUI();

  let loading = $state(true);
  let error = $state<string | null>(null);
  let graph = $state<TopologyGraph | null>(null);

  async function load() {
    // A reload keeps the drawn graph until the new one arrives.
    if (!graph) loading = true;
    error = null;
    try {
      graph = await client.topology();
    } catch (err) {
      error = errorText(err);
    } finally {
      loading = false;
    }
  }

  onMount(load);

  const startEpoch = untrack(() => node.graphEpoch);
  $effect(() => {
    if (node.graphEpoch !== startEpoch) untrack(() => void load());
  });

  // Nothing but the node and its Aggregator: no device, no controller.
  const empty = $derived(!!graph && !graph.nodes.some((n) => n.kind !== "node" && n.kind !== "aggregator"));
  const hrefOf = (n: TopologyNode) => refHref(n.ref, routes);
</script>

<Card class="p-4">
  <div class="flex items-center justify-between gap-2">
    <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("node.topology.title")}</h3>
    <Button variant="outline" size="sm" onclick={() => void load()} disabled={loading}>{t("common.refresh")}</Button>
  </div>
  <p class="mt-1 text-sm text-[var(--ha-secondary-text-color)]">{t("node.topology.hint")}</p>
  <div class="mt-3">
    {#if loading}
      <LoadingState message={t("common.loading")} />
    {:else if error}
      <ErrorState message={error} onRetry={() => void load()} />
    {:else if graph && empty}
      <EmptyState message={t("node.topology.empty")} />
    {:else if graph}
      <Graph {graph} {hrefOf} {follow} label={t("node.topology.graph_label")} />
    {/if}
  </div>
</Card>
