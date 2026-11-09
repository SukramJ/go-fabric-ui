<!-- The controller's network as a graph (GET /controller/topology): the
     controller with a session to each commissioned node, the Thread
     routers and end devices and the Wi-Fi access points the nodes report,
     and the links between them by strength. A commissioned node links to
     its page. The graph reloads when the stream reports topology_changed
     (the controller store's graph epoch, one move per burst); "Read
     diagnostics again" asks the controller to re-read the nodes' network
     diagnostics first (refresh=true, an admin route). A host without a
     topology service answers 501, shown as an empty state. New in brief
     U3b. The Thread border routers go-fabric's registry knows are nodes of
     their own (border_router, contract 0.8.0), and a legend under the
     graph names each kind it draws by its shape. -->
<script lang="ts">
  import { onMount, untrack } from "svelte";
  import { refHref, useFabricUI } from "../../context";
  import { errorText } from "../../store.svelte";
  import { ApiError, type TopologyGraph, type TopologyNode } from "../../api/client";
  import Button from "../../kit/Button.svelte";
  import Card from "../../kit/Card.svelte";
  import EmptyState from "../../kit/EmptyState.svelte";
  import ErrorState from "../../kit/ErrorState.svelte";
  import Graph from "../../kit/Graph.svelte";
  import LoadingState from "../../kit/LoadingState.svelte";
  import { KIND_SHAPE, TOPOLOGY_KINDS } from "../../kit/graph-layout";

  type Props = { follow?: (href: string) => void };
  let { follow }: Props = $props();

  const { t, toast, client, controller, routes } = useFabricUI();

  let loading = $state(true);
  let refreshing = $state(false);
  let error = $state<string | null>(null);
  let notServed = $state(false);
  let graph = $state<TopologyGraph | null>(null);

  async function load() {
    // A reload keeps the drawn graph until the new one arrives.
    if (!graph) loading = true;
    error = null;
    try {
      graph = await client.controllerTopology();
      notServed = false;
    } catch (err) {
      if (err instanceof ApiError && err.status === 501) notServed = true;
      else error = errorText(err);
    } finally {
      loading = false;
    }
  }

  // refresh=true re-reads the nodes' diagnostics; a refusal (403 without
  // admin rights, 502 when the read fails) keeps the drawn graph.
  async function refresh() {
    refreshing = true;
    try {
      graph = await client.controllerTopology(true);
      notServed = false;
      error = null;
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      refreshing = false;
    }
  }

  onMount(load);

  const startEpoch = untrack(() => controller.graphEpoch);
  $effect(() => {
    if (controller.graphEpoch !== startEpoch) untrack(() => void load());
  });

  // Nothing but the controller: no node commissioned.
  const empty = $derived(!!graph && !graph.nodes.some((n) => n.kind !== "node"));
  // The kinds the graph draws, in the contract's order, for the legend.
  const kinds = $derived(graph ? TOPOLOGY_KINDS.filter((k) => graph!.nodes.some((n) => n.kind === k)) : []);
  const hrefOf = (n: TopologyNode) => (n.kind === "peer" ? refHref(n.ref, routes) : undefined);
</script>

<Card class="p-4">
  <div class="flex items-center justify-between gap-2">
    <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("controller.network.title")}</h3>
    {#if !notServed}
      <Button variant="outline" size="sm" onclick={() => void refresh()} disabled={loading || refreshing}>{t("controller.network.refresh")}</Button>
    {/if}
  </div>
  <p class="mt-1 text-sm text-[var(--ha-secondary-text-color)]">{t("controller.network.hint")}</p>
  <div class="mt-3">
    {#if loading}
      <LoadingState message={t("common.loading")} />
    {:else if notServed}
      <EmptyState message={t("controller.network.not_served")} />
    {:else if error}
      <ErrorState message={error} onRetry={() => void load()} />
    {:else if graph && empty}
      <EmptyState message={t("controller.network.empty")} />
    {:else if graph}
      <Graph {graph} {hrefOf} {follow} label={t("controller.network.graph_label")} />
      <ul class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-[var(--ha-secondary-text-color)]" data-testid="network-legend" aria-label={t("controller.network.legend")}>
        {#each kinds as k (k)}
          <li class="inline-flex items-center gap-1.5" data-kind={k}>
            <svg width="22" height="14" viewBox="0 0 22 14" aria-hidden="true">
              <rect x="1" y="1" width="20" height="12" rx={Math.min(KIND_SHAPE[k].rx, 6)} fill="none" stroke="currentColor" stroke-width="1.5" stroke-dasharray={KIND_SHAPE[k].dash} />
            </svg>
            {t(`graph.kind.${k}`)}
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</Card>
