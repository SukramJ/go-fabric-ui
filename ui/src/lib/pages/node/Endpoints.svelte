<!-- The bridged endpoints as a controller sees them, without commissioning
     one first. Ported from the endpoint card of openccu-loom
     assets/ui/src/routes/matter/MatterDiagnostics.svelte:229-260 at
     b40b289a: the host's name for the endpoint, its parent, its device
     type and clusters. An endpoint whose surface changed in this process
     (GET /node/surface, changed[]) carries a marker. Each endpoint links to
     its own page (GET /node/endpoints/{n}), the root endpoint and the
     Aggregator included: the route's structural list (go-fabric
     Node.Endpoints, contract 0.5.0). -->
<script lang="ts">
  import { onMount, untrack } from "svelte";
  import { useFabricUI } from "../../context";
  import { errorText } from "../../store.svelte";
  import type { Endpoint, EndpointList, SurfaceReport } from "../../api/client";
  import Badge from "../../kit/Badge.svelte";
  import Button from "../../kit/Button.svelte";
  import Card from "../../kit/Card.svelte";
  import EmptyState from "../../kit/EmptyState.svelte";
  import ErrorState from "../../kit/ErrorState.svelte";
  import LoadingState from "../../kit/LoadingState.svelte";

  const { t, client, node, routes } = useFabricUI();

  let loading = $state(true);
  let error = $state<string | null>(null);
  let endpoints = $state<Endpoint[]>([]);
  let structuralList = $state<NonNullable<EndpointList["structural"]>>([]);
  let surface = $state<SurfaceReport | null>(null);

  async function load() {
    loading = true;
    error = null;
    try {
      // The surface report only marks endpoints: its failure (an older
      // server, a report that does not encode) does not fail the page.
      const [eps, rep] = await Promise.all([client.endpoints(), client.surface().catch(() => null)]);
      endpoints = eps.endpoints;
      structuralList = eps.structural ?? [];
      surface = rep;
    } catch (err) {
      error = errorText(err);
    } finally {
      loading = false;
    }
  }

  onMount(load);

  // endpoints_changed and surface_changed from the stream move the
  // store's epoch; the page refetches on each move after it mounted.
  const startEpoch = untrack(() => node.topologyEpoch);
  $effect(() => {
    if (node.topologyEpoch !== startEpoch) untrack(() => void load());
  });

  // A changed endpoint is named by its surface key, which go-fabric
  // renders from the endpoint's stable key (node/surface_stability.go:62-63,
  // persist.go:90-94) — the source.key this route carries
  // (handlers_diagnostics.go flattenTopology). An endpoint without a key
  // is matched by its number through the report's effective surface,
  // which carries each endpoint's key and number.
  const changedKeys = $derived(new Set((surface?.changed ?? []).map((c) => c.endpoint)));
  const changedNumbers = $derived(
    new Set((surface?.effective.endpoints ?? []).filter((e) => e.number && changedKeys.has(e.key)).map((e) => e.number!)),
  );
  const surfaceChanged = (ep: Endpoint) => (!!ep.source.key && changedKeys.has(ep.source.key)) || changedNumbers.has(ep.endpoint_id);

  const hex = (n: number, w = 4) => `0x${n.toString(16).toUpperCase().padStart(w, "0")}`;
  const topLevel = $derived(new Set(endpoints.map((e) => e.endpoint_id)));
  // The root and the Aggregator as the node's endpoint tree lists them;
  // the root (endpoint 0) is named by the page, the others by the name the
  // route gives them. A server without the list (before 0.5.0) shows the
  // root alone.
  const structural = $derived(
    (structuralList.length > 0 ? structuralList : [{ endpoint_id: 0, name: "" }]).map((s) => ({
      number: s.endpoint_id,
      label: s.endpoint_id === 0 ? t("node.endpoints.root") : s.name,
    })),
  );
</script>

<Card class="p-4">
  <div class="flex items-center justify-between gap-2">
    <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("node.endpoints.title")}</h3>
    <Button variant="outline" size="sm" onclick={() => void load()} disabled={loading}>{t("common.refresh")}</Button>
  </div>
  <p class="mt-1 text-xs text-[var(--ha-secondary-text-color)]">{t("node.endpoints.hint")}</p>
  {#if loading}
    <LoadingState />
  {:else if error}
    <ErrorState class="mt-3" message={error} onRetry={() => void load()} />
  {:else}
    <ul class="mt-3 flex flex-wrap gap-2" data-testid="structural-endpoints">
      {#each structural as s (s.number)}
        <li>
          <a class="rounded border border-[var(--ha-divider-color)] px-3 py-1 text-sm text-[var(--ha-primary-color)] hover:underline" href={routes.endpoint + s.number}>
            <span class="font-mono text-[var(--ha-secondary-text-color)]">#{s.number}</span> {s.label}
          </a>
        </li>
      {/each}
    </ul>
  {/if}
  {#if !loading && !error && endpoints.length === 0}
    <EmptyState class="mt-3" message={t("node.endpoints.empty")} />
  {:else if !loading && !error}
    <ul class="mt-3 space-y-2">
      {#each endpoints as ep (ep.endpoint_id)}
        <li class="rounded border border-[var(--ha-divider-color)] p-3" data-testid="endpoint">
          <div class="flex flex-wrap items-center gap-2">
            <span class="font-mono text-sm text-[var(--ha-secondary-text-color)]">#{ep.endpoint_id}</span>
            <a class="font-medium text-[var(--ha-primary-color)] hover:underline" href={routes.endpoint + ep.endpoint_id}>{ep.name}</a>
            <Badge variant={ep.reachable ? "success" : "muted"}>
              {ep.reachable ? t("node.endpoints.reachable") : t("node.endpoints.unreachable")}
            </Badge>
            {#if surfaceChanged(ep)}
              <Badge variant="warning">{t("node.endpoints.surface_changed")}</Badge>
            {/if}
            <span class="text-xs text-[var(--ha-secondary-text-color)]">
              {ep.device_type_name || hex(ep.device_type)}{ep.device_type_revision ? ` r${ep.device_type_revision}` : ""}
            </span>
            {#if topLevel.has(ep.parent_endpoint_id)}
              <span class="text-xs text-[var(--ha-secondary-text-color)]">{t("node.endpoints.part_of", { id: ep.parent_endpoint_id })}</span>
            {/if}
          </div>
          {#if ep.clusters.length > 0}
            <div class="mt-2 flex flex-wrap gap-1">
              {#each ep.clusters as cluster (cluster.id)}
                <span class="rounded bg-[var(--ha-secondary-background-color)] px-2 py-0.5 text-xs text-[var(--ha-secondary-text-color)]">
                  {cluster.name || hex(cluster.id)}
                </span>
              {/each}
            </div>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
</Card>
