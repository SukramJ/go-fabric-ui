<!-- One endpoint the node serves (contract 0.4.1, GET /node/endpoints/{n}):
     its device types with their revisions, reachability, scope, parent and
     parts, its tags and labels, every server cluster with its values
     (ClusterView, the node flavour), the host's surface decisions on it and
     the events the node's event log holds for it. The page refetches the
     endpoint on its attribute_changed (ids only, no values on the stream)
     and on endpoints_changed or surface_changed; the events after the last
     one it shows come with each refetch, and the stream's cluster_event of
     this endpoint joins the list at once (contract 0.7.0). -->
<script lang="ts">
  import DeniedState from "../../kit/DeniedState.svelte";
  import { untrack } from "svelte";
  import { useFabricUI } from "../../context";
  import { errorText } from "../../store.svelte";
  import type { EndpointDetail, EventRecord, SurfaceDecision } from "../../api/client";
  import Badge from "../../kit/Badge.svelte";
  import Button from "../../kit/Button.svelte";
  import Card from "../../kit/Card.svelte";
  import ClusterView from "../../kit/ClusterView.svelte";
  import ErrorState from "../../kit/ErrorState.svelte";
  import EventList from "../../kit/EventList.svelte";
  import { keyedRows } from "../../kit/data-table";
  import { eventNumberText, mergeEvents, sinceAfter } from "../../kit/events";
  import LoadingState from "../../kit/LoadingState.svelte";
  import TagChips from "../../kit/TagChips.svelte";
  import { hexId } from "./surface";

  let { endpoint }: { endpoint: number } = $props();
  const { t, client, node, routes } = useFabricUI();

  // BridgedDeviceBasicInformation: an endpoint that serves it is bridged.
  const BRIDGED_BASIC_INFORMATION = 0x0039;
  // How many events the page keeps; the route's own limit is 100.
  const MAX_EVENTS = 500;
  // An endpoint with more clusters than this starts with them collapsed.
  const OPEN_CLUSTERS = 6;
  // A cluster panel's open state is set once, when it is created: a reload
  // leaves what the operator opened or closed (UR-P2-ui-191).
  function initiallyOpen(el: HTMLDetailsElement) {
    el.open = untrack(() => (detail?.clusters.length ?? 0) <= OPEN_CLUSTERS);
  }

  let detail = $state<EndpointDetail | null>(null);
  let error = $state<string | null>(null);
  let events = $state<EventRecord[]>([]);
  let eventsError = $state<string | null>(null);
  let allDecisions = $state(false);
  // A newer load wins: an older answer that arrives last is dropped.
  let detailSeq = 0;
  let eventSeq = 0;

  async function loadDetail() {
    const mine = ++detailSeq;
    try {
      const d = await client.nodeEndpointDetail(endpoint);
      if (mine !== detailSeq) return;
      detail = d;
      error = null;
    } catch (err) {
      if (mine === detailSeq) error = errorText(err);
    }
  }

  // fresh reads the log from its start; otherwise only the events after
  // the last one shown (since = its number + 1).
  async function loadEvents(fresh: boolean) {
    const mine = ++eventSeq;
    // The cursor travels as since_str, exact beyond 2^53 (0.10.0).
    const w = fresh ? {} : sinceAfter(events);
    try {
      const l = await client.nodeEndpointEvents(endpoint, w);
      if (mine !== eventSeq) return;
      // A cluster_event the stream added while the read was in flight
      // comes again in the answer: merge by number, in order.
      events = mergeEvents(w.since === undefined ? [] : events, l.events).slice(-MAX_EVENTS);
      eventsError = null;
    } catch (err) {
      if (mine === eventSeq) eventsError = errorText(err);
    }
  }

  function refresh() {
    void loadDetail();
    void loadEvents(true);
  }

  // The endpoint's attribute_changed and the node's structure changes move
  // the store's epochs; each move after the page mounted (or after the
  // endpoint prop changed) refetches.
  let startAttr = untrack(() => node.attributeEpoch(endpoint));
  let startTopo = untrack(() => node.topologyEpoch);

  // A new endpoint on a mounted page: nothing of the previous one stays.
  let shown = untrack(() => endpoint);
  $effect(() => {
    const ep = endpoint;
    untrack(() => {
      if (ep !== shown) {
        shown = ep;
        detail = null;
        error = null;
        events = [];
        eventsError = null;
        startAttr = node.attributeEpoch(ep);
        startTopo = node.topologyEpoch;
      }
      refresh();
    });
  });

  $effect(() => {
    const moved = node.attributeEpoch(endpoint) !== startAttr || node.topologyEpoch !== startTopo;
    if (moved) {
      untrack(() => {
        void loadDetail();
        void loadEvents(false);
      });
    }
  });

  // A cluster_event of this endpoint joins the list unless a refetch
  // brought it already; one of another endpoint is not this page's.
  const startEvent = untrack(() => node.clusterEvent);
  $effect(() => {
    const e = node.clusterEvent;
    if (!e || e === startEvent || e.endpoint !== endpoint) return;
    untrack(() => {
      if (events.some((x) => eventNumberText(x) === eventNumberText(e))) return;
      events = [...events, e].slice(-MAX_EVENTS);
    });
  });

  const isBridged = (d: EndpointDetail) => d.clusters.some((c) => c.id === BRIDGED_BASIC_INFORMATION);

  // The decisions as the surface page shows them: by default without the
  // silent rows, a keep without a reason.
  const isSilent = (d: SurfaceDecision) => d.verdict === "keep" && d.reason === "";
  const decisions = $derived((detail?.decisions ?? []).filter((d) => allDecisions || !isSilent(d)));
  const silentCount = $derived((detail?.decisions ?? []).filter(isSilent).length);
  const verdictVariant = (v: string): "muted" | "warning" | "danger" => (v === "refuse" ? "danger" : v === "warn" || v === "remove" ? "warning" : "muted");
  const clusterLabel = (id: string) => {
    const name = detail?.clusters.find((c) => c.id === Number(id))?.name;
    return name ? `${hexId(id)} ${name}` : hexId(id);
  };
  const deviceTypeLabel = (id: string) => detail?.device_types.find((d) => d.id === Number(id))?.name ?? hexId(id);
  const element = (d: SurfaceDecision) => {
    const own = `${t(`node.surface.kind.${d.kind}`)} ${d.kind === "device-type" ? deviceTypeLabel(d.id) : d.kind === "cluster" ? clusterLabel(d.id) : hexId(d.id)}`;
    return d.kind === "device-type" || d.kind === "cluster" ? own : `${own} · ${clusterLabel(d.cluster)}`;
  };
</script>

<!-- Without admin rights the routes of this page are refused (0.10.0): a notice, not an error. -->
{#if node.status !== null && !node.isAdmin}
  <DeniedState />
{:else}
<div class="space-y-4">
  <div class="flex items-center justify-between gap-2">
    <a href={routes.endpoints} class="text-sm text-[var(--fui-primary-text)] hover:underline">← {t("node.endpoint.back")}</a>
    <Button variant="outline" size="sm" onclick={refresh}>{t("common.refresh")}</Button>
  </div>
  {#if error}
    <ErrorState message={error} onRetry={() => void loadDetail()} />
  {:else if !detail}
    <LoadingState message={t("common.loading")} />
  {:else}
    {@const d = detail}
    <Card class="p-4" data-testid="endpoint-head">
      <div class="flex flex-wrap items-center gap-2">
        <span class="font-mono text-sm text-[var(--ha-secondary-text-color)]">#{d.endpoint_id}</span>
        <h2 class="text-lg font-medium text-[var(--ha-primary-text-color)]" data-testid="endpoint-title">{d.name || t("node.endpoint.title", { number: d.endpoint_id })}</h2>
        {#each d.device_types as dt, i (i)}
          <Badge title={dt.id_hex}>{dt.name ?? dt.id_hex} r{dt.revision}</Badge>
        {/each}
        {#if isBridged(d)}
          <Badge variant={d.reachable ? "success" : "danger"}>{t(d.reachable ? "node.endpoints.reachable" : "node.endpoints.unreachable")}</Badge>
        {/if}
        {#if d.scope}
          <Badge variant="muted" title={t("node.endpoint.scope_hint")}>{t("node.endpoint.scope", { scope: d.scope })}</Badge>
        {/if}
      </div>
      <div class="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-[var(--ha-secondary-text-color)]">
        {#if d.parent_endpoint_id !== undefined}
          <span>{t("node.endpoint.parent")} <a class="text-[var(--fui-primary-text)] hover:underline" href={routes.endpoint + d.parent_endpoint_id}>#{d.parent_endpoint_id}</a></span>
        {/if}
        {#if d.parts.length > 0}
          <span data-testid="endpoint-parts">
            {t("node.endpoint.parts")}
            {#each d.parts as p, pi (pi)}
              <a class="ml-1 text-[var(--fui-primary-text)] hover:underline" href={routes.endpoint + p}>#{p}</a>
            {/each}
          </span>
        {/if}
      </div>
      <TagChips class="mt-2" tags={d.tags} empty={t("node.endpoint.no_tags")} />
      {#if d.labels.length > 0}
        <ul class="mt-2 flex flex-wrap gap-1" data-testid="endpoint-labels">
          {#each d.labels as l, i (i)}
            <li><Badge variant="muted" title={t(l.user ? "controller.inventory.user_label" : "controller.inventory.fixed_label")}>{l.label}: {l.value}</Badge></li>
          {/each}
        </ul>
      {/if}
    </Card>

    <Card class="p-4">
      <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("node.endpoint.clusters")}</h3>
      <p class="mt-1 text-xs text-[var(--ha-secondary-text-color)]">{t("node.endpoint.clusters_hint")}</p>
      <ul class="mt-3 divide-y divide-[var(--ha-divider-color)]">
        <!-- Keyed by cluster, so a panel keeps its open state when a reload
             adds a cluster before it; a repeated id gets a suffix (UR-P2-ui-191). -->
        {#each keyedRows(d.clusters, (x) => x.id_hex) as { row: c, key } (key)}
          <li class="py-1">
            <details {@attach initiallyOpen} data-testid={`endpoint-cluster-${c.id_hex}`}>
              <summary class="flex cursor-pointer flex-wrap items-center gap-2 py-1 text-sm">
                <span class="font-medium">{c.name ?? c.id_hex}</span>
                <span class="font-mono text-xs text-[var(--ha-secondary-text-color)]">{c.id_hex} r{c.revision}</span>
              </summary>
              <ClusterView class="pb-2 pl-4" cluster={c} />
            </details>
          </li>
        {/each}
      </ul>
    </Card>

    <Card class="p-4" data-testid="endpoint-decisions">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("node.surface.decisions")}</h3>
        {#if silentCount > 0}
          <Button variant="outline" size="sm" onclick={() => (allDecisions = !allDecisions)}>
            {allDecisions ? t("node.endpoint.decisions_hide_silent") : t("node.endpoint.decisions_show_all", { count: silentCount })}
          </Button>
        {/if}
      </div>
      {#if decisions.length === 0}
        <p class="mt-2 text-xs text-[var(--ha-secondary-text-color)]">{t("node.surface.decisions_empty")}</p>
      {:else}
        <ul class="mt-2 space-y-1 text-sm">
          {#each decisions as dec, i (i)}
            <li class="flex flex-wrap items-center gap-2" data-testid="endpoint-decision">
              <Badge variant={verdictVariant(dec.verdict)}>{t(`node.surface.verdict.${dec.verdict}`)}</Badge>
              {#if dec.mandatory}<Badge variant="muted">{t("node.surface.mandatory")}</Badge>{/if}
              <span>{element(dec)}</span>
              <span class="font-mono text-xs text-[var(--ha-secondary-text-color)]">{dec.policy}</span>
              {#if dec.reason}<span class="text-xs text-[var(--ha-secondary-text-color)]">{dec.reason}</span>{/if}
            </li>
          {/each}
        </ul>
      {/if}
    </Card>
  {/if}

  <Card class="p-4" data-testid="endpoint-events">
    <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("events.title")}</h3>
    <p class="mt-1 text-xs text-[var(--ha-secondary-text-color)]">{t("node.endpoint.events_hint")}</p>
    {#if eventsError}
      <ErrorState class="mt-2" message={eventsError} onRetry={() => void loadEvents(true)} />
    {:else}
      <EventList class="mt-2" {events} empty={t("events.empty")} />
    {/if}
  </Card>
</div>
{/if}
