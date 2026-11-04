<!-- A node's inventory from the controller's image (contract 0.4.0): per
     endpoint its device types, tags, labels, reachability and conformance,
     then its server clusters; a cluster loads its values when it is
     expanded, one request each. Nothing is read from the node. The tab
     reloads what it shows when the page's node detail is reloaded (the
     stream's node_updated). -->
<script lang="ts">
  import { untrack } from "svelte";
  import { useFabricUI } from "../../../context";
  import { errorText } from "../../../store.svelte";
  import type { ClusterDetail, ControllerNodeDetail, EndpointDetail } from "../../../api/client";
  import Badge from "../../../kit/Badge.svelte";
  import Card from "../../../kit/Card.svelte";
  import ClusterView from "../../../kit/ClusterView.svelte";
  import EmptyState from "../../../kit/EmptyState.svelte";
  import ErrorState from "../../../kit/ErrorState.svelte";
  import LoadingState from "../../../kit/LoadingState.svelte";
  import TagChips from "../../../kit/TagChips.svelte";

  let { nodeId, detail, focus }: { nodeId: string; detail: ControllerNodeDetail; focus?: number } = $props();
  const { t, client } = useFabricUI();

  // BridgedDeviceBasicInformation: an endpoint that serves it is bridged.
  const BRIDGED_BASIC_INFORMATION = 0x0039;

  type Loaded<T> = { value?: T; error?: string };
  let endpoints = $state<Record<number, Loaded<EndpointDetail>>>({});
  // The expanded clusters by "<endpoint>/<cluster_hex>".
  let clusters = $state<Record<string, Loaded<ClusterDetail>>>({});

  const keyOf = (ep: number, hex: string) => `${ep}/${hex}`;

  // The latest load of each endpoint and cluster: a reload's answer that
  // arrives after a newer one's is dropped, so an older image never lands
  // last.
  const seq: Record<string, number> = {};
  const begin = (key: string) => (seq[key] = (seq[key] ?? 0) + 1);

  async function loadEndpoint(ep: number) {
    const mine = begin(`ep:${ep}`);
    try {
      const value = await client.controllerEndpointDetail(nodeId, ep);
      if (seq[`ep:${ep}`] === mine) endpoints[ep] = { value };
    } catch (err) {
      if (seq[`ep:${ep}`] === mine) endpoints[ep] = { error: errorText(err) };
    }
  }

  async function loadCluster(ep: number, hex: string) {
    const key = keyOf(ep, hex);
    const mine = begin(key);
    try {
      const value = await client.controllerClusterDetail(nodeId, ep, hex);
      if (key in clusters && seq[key] === mine) clusters[key] = { value };
    } catch (err) {
      if (key in clusters && seq[key] === mine) clusters[key] = { error: errorText(err) };
    }
  }

  function toggle(ep: number, hex: string) {
    const key = keyOf(ep, hex);
    if (key in clusters) {
      delete clusters[key];
      return;
    }
    clusters[key] = {};
    void loadCluster(ep, hex);
  }

  // Every endpoint on a new detail (the first, and each reload of the
  // node), and the clusters that are open.
  $effect(() => {
    const eps = detail.endpoints.map((e) => e.endpoint);
    untrack(() => {
      for (const ep of eps) void loadEndpoint(ep);
      for (const key of Object.keys(clusters)) {
        const [ep, hex] = key.split("/");
        void loadCluster(Number(ep), hex);
      }
    });
  });

  // The endpoint the endpoints tab linked to scrolls into view.
  $effect(() => {
    if (focus === undefined || !endpoints[focus]?.value) return;
    untrack(() => document.querySelector(`[data-testid="inventory-endpoint-${focus}"]`)?.scrollIntoView?.({ block: "start" }));
  });

  const isBridged = (e: EndpointDetail) => e.clusters.some((c) => c.id === BRIDGED_BASIC_INFORMATION);
  const conformanceVariant = (s: string) => (s === "clean" ? "success" : s === "violated" ? "danger" : s === "incomplete" ? "warning" : "muted");
</script>

<p class="mb-3 text-xs text-[var(--ha-secondary-text-color)]">{t("controller.inventory.hint")}</p>
{#if detail.endpoints.length === 0}
  <EmptyState message={t("controller.endpoints.empty")} />
{:else}
  <div class="space-y-3">
    {#each detail.endpoints as ne (ne.endpoint)}
      {@const loaded = endpoints[ne.endpoint]}
      <Card class={["p-4", focus === ne.endpoint ? "ring-2 ring-[var(--ha-primary-color)]" : ""].join(" ")} data-testid={`inventory-endpoint-${ne.endpoint}`}>
        <div class="flex flex-wrap items-center gap-2">
          <span class="font-medium text-[var(--ha-primary-text-color)]">{t("controller.endpoints.endpoint", { number: ne.endpoint })}</span>
          {#if loaded?.value}
            {@const ep = loaded.value}
            <span class="text-sm text-[var(--ha-secondary-text-color)]">{ep.name}</span>
            {#each ep.device_types as dt, i (i)}
              <Badge title={dt.id_hex}>{dt.name ?? dt.id_hex} r{dt.revision}</Badge>
            {/each}
            {#if isBridged(ep)}
              <Badge variant={ep.reachable ? "success" : "danger"}>{t(ep.reachable ? "controller.endpoints.reachable" : "controller.endpoints.unreachable")}</Badge>
            {/if}
            {#if ep.conformance}
              <Badge variant={conformanceVariant(ep.conformance.status)} title={ep.conformance.unjudged.length ? t("controller.inventory.unjudged", { clusters: ep.conformance.unjudged.map((c) => c.name ?? c.id_hex).join(", ") }) : undefined}>
                {t(`controller.inventory.conformance.${ep.conformance.status}`)}
              </Badge>
            {/if}
          {/if}
        </div>
        {#if loaded?.error}
          <ErrorState class="mt-2" message={loaded.error} onRetry={() => void loadEndpoint(ne.endpoint)} />
        {:else if !loaded?.value}
          <LoadingState class="mt-2" label={t("common.loading")} />
        {:else}
          {@const ep = loaded.value}
          <TagChips class="mt-2" tags={ep.tags} />
          {#if ep.labels.length > 0}
            <ul class="mt-2 flex flex-wrap gap-1">
              {#each ep.labels as l, i (i)}
                <li><Badge variant="muted" title={t(l.user ? "controller.inventory.user_label" : "controller.inventory.fixed_label")}>{l.label}: {l.value}</Badge></li>
              {/each}
            </ul>
          {/if}
          {#if ep.parts.length > 0}
            <p class="mt-2 text-xs text-[var(--ha-secondary-text-color)]">{t("controller.inventory.parts", { parts: ep.parts.join(", ") })}</p>
          {/if}
          {#if ep.conformance && ep.conformance.violations.length > 0}
            <ul class="mt-2 space-y-1 text-xs text-[var(--ha-error-color)]" data-testid={`violations-${ne.endpoint}`}>
              {#each ep.conformance.violations as v, i (i)}
                <li>{v.device_type} · {v.requirement}: {v.detail}</li>
              {/each}
            </ul>
          {/if}
          {#if ep.clusters.length === 0}
            <p class="mt-2 text-xs text-[var(--ha-secondary-text-color)]">{t("controller.inventory.no_clusters")}</p>
          {:else}
            <ul class="mt-3 divide-y divide-[var(--ha-divider-color)]">
              {#each ep.clusters as c (c.id)}
                {@const key = keyOf(ne.endpoint, c.id_hex)}
                {@const open = clusters[key]}
                <li class="py-1">
                  <button
                    type="button"
                    class="flex w-full flex-wrap items-center gap-2 py-1 text-left text-sm hover:text-[var(--ha-primary-color)]"
                    aria-expanded={open !== undefined}
                    data-testid={`cluster-${ne.endpoint}-${c.id_hex}`}
                    onclick={() => toggle(ne.endpoint, c.id_hex)}
                  >
                    <span aria-hidden="true">{open !== undefined ? "▾" : "▸"}</span>
                    <span class="font-medium">{c.name ?? c.id_hex}</span>
                    <span class="font-mono text-xs text-[var(--ha-secondary-text-color)]">{c.id_hex} r{c.revision}</span>
                    {#each c.features as f (f.bit)}
                      <Badge variant="muted" title={t("inventory.feature_bit", { bit: f.bit })}>{f.code}</Badge>
                    {/each}
                  </button>
                  {#if open !== undefined}
                    <div class="pb-2 pl-5">
                      {#if open.error}
                        <ErrorState message={open.error} onRetry={() => void loadCluster(ne.endpoint, c.id_hex)} />
                      {:else if open.value}
                        <ClusterView cluster={open.value} />
                      {:else}
                        <LoadingState label={t("common.loading")} />
                      {/if}
                    </div>
                  {/if}
                </li>
              {/each}
            </ul>
          {/if}
        {/if}
      </Card>
    {/each}
  </div>
{/if}
