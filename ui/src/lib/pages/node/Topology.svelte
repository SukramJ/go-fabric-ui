<!-- The node as a graph (GET /node/topology): the node, its Aggregator,
     the bridged devices and their parts on one side, the fabrics and the
     sessions their controllers hold on the other. A node's link leads to
     the page that already lists it. The graph reloads when the stream
     reports a session, a fabric or the endpoints changing (the store's
     graph epoch, one move per burst). New in brief U3a; no loom page
     precedes it. -->
<script lang="ts">
  import DeniedState from "../../kit/DeniedState.svelte";
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

  const { t, toast, client, node, routes } = useFabricUI();

  let loading = $state(true);
  let error = $state<string | null>(null);
  let graph = $state<TopologyGraph | null>(null);

  // A newer load wins: an older answer that arrives last is dropped.
  let seq = 0;

  async function load() {
    const mine = ++seq;
    // A reload keeps the drawn graph until the new one arrives.
    if (!graph) loading = true;
    error = null;
    try {
      const g = await client.topology();
      if (mine === seq) graph = g;
    } catch (err) {
      // With a graph drawn, a failed reload keeps it: the failure is a
      // toast, as on the controller's network (UR-P4-ui-305).
      if (mine !== seq) return;
      if (graph) toast.error(errorText(err));
      else error = errorText(err);
    } finally {
      if (mine === seq) loading = false;
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

<!-- Without admin rights the routes of this page are refused (0.10.0): a notice, not an error. -->
{#if node.status !== null && !node.isAdmin}
  <DeniedState />
{:else}
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
{/if}
