<!-- The node's surface report (GET /node/surface, go-fabric's
     surface.Report): the policies it is composed under, the effective
     surface per endpoint, the policies' decisions, the elements withheld
     and the surface changes announced in this process. Names come from
     GET /node/endpoints where its ids match; otherwise the id is shown.
     The page idioms are those of Endpoints.svelte and Diagnostics.svelte. -->
<script lang="ts">
  import { onMount, untrack } from "svelte";
  import { useFabricUI } from "../../context";
  import { errorText } from "../../store.svelte";
  import type { Endpoint, SurfaceCluster, SurfaceDecision, SurfaceReport } from "../../api/client";
  import type { DataColumn } from "../../kit/data-table";
  import Badge from "../../kit/Badge.svelte";
  import Button from "../../kit/Button.svelte";
  import Card from "../../kit/Card.svelte";
  import DataTable from "../../kit/DataTable.svelte";
  import EmptyState from "../../kit/EmptyState.svelte";
  import ErrorState from "../../kit/ErrorState.svelte";
  import LoadingState from "../../kit/LoadingState.svelte";
  import Select from "../../kit/Select.svelte";
  import { hexId, levelOrigin, shortDigest } from "./surface";

  const { t, client, node } = useFabricUI();

  let loading = $state(true);
  let error = $state<string | null>(null);
  let report = $state<SurfaceReport | null>(null);
  let endpoints = $state<Endpoint[]>([]);

  async function load() {
    loading = true;
    error = null;
    try {
      // The names are a convenience: a failed endpoints call leaves ids.
      const [r, eps] = await Promise.all([
        client.surface(),
        client.endpoints().then(
          (l) => l.endpoints,
          () => [] as Endpoint[],
        ),
      ]);
      report = r;
      endpoints = eps;
    } catch (err) {
      error = errorText(err);
    } finally {
      loading = false;
    }
  }

  onMount(load);

  // A reassembly or an announced surface change moves the store's epoch.
  const startEpoch = untrack(() => node.topologyEpoch);
  $effect(() => {
    if (node.topologyEpoch !== startEpoch) untrack(() => void load());
  });

  // Names joined from the endpoints route by numeric id: the device type
  // names it carries and the cluster names of every endpoint.
  const deviceTypeNames = $derived(new Map(endpoints.filter((e) => e.device_type_name).map((e) => [e.device_type, e.device_type_name])));
  const clusterNames = $derived(new Map(endpoints.flatMap((e) => e.clusters.filter((c) => c.name).map((c) => [c.id, c.name!] as const))));
  const deviceTypeLabel = (id: string) => deviceTypeNames.get(Number(id)) ?? hexId(id);
  const clusterLabel = (id: string) => {
    const name = clusterNames.get(Number(id));
    return name ? `${hexId(id)} ${name}` : hexId(id);
  };
  const endpointLabel = (key: string) => (key === "" ? t("node.surface.root") : key);

  // The decisions filter: by default everything but the silent rows, a
  // keep without a reason.
  let verdictFilter = $state("default");
  let endpointFilter = $state("");
  const verdictOptions = $derived([
    { value: "default", label: t("node.surface.filter_default") },
    { value: "all", label: t("node.surface.filter_all") },
    ...["keep", "warn", "remove", "refuse"].map((v) => ({ value: v, label: t(`node.surface.verdict.${v}`) })),
  ]);
  const endpointOptions = $derived([
    { value: "", label: t("node.surface.filter_endpoint_all") },
    ...[...new Set((report?.decisions ?? []).map((d) => d.endpoint))].sort().map((k) => ({ value: `=${k}`, label: endpointLabel(k) })),
  ]);
  const isSilent = (d: SurfaceDecision) => d.verdict === "keep" && d.reason === "";
  const decisions = $derived(
    (report?.decisions ?? []).filter(
      (d) =>
        (verdictFilter === "all" || (verdictFilter === "default" ? !isSilent(d) : d.verdict === verdictFilter)) &&
        (endpointFilter === "" || `=${d.endpoint}` === endpointFilter),
    ),
  );

  const verdictVariant = (v: string): "muted" | "warning" | "danger" => (v === "refuse" ? "danger" : v === "warn" || v === "remove" ? "warning" : "muted");
  const element = (e: { kind: string; id: string; cluster: string }) => {
    const own = `${t(`node.surface.kind.${e.kind}`)} ${e.kind === "device-type" ? deviceTypeLabel(e.id) : e.kind === "cluster" ? clusterLabel(e.id) : hexId(e.id)}`;
    return e.kind === "device-type" || e.kind === "cluster" ? own : `${own} · ${clusterLabel(e.cluster)}`;
  };

  const clusterColumns: DataColumn<SurfaceCluster>[] = $derived([
    { key: "id", label: t("node.surface.col_cluster"), title: true, sortable: true, align: "left", get: (c) => Number(c.id) },
    { key: "revision", label: t("node.surface.col_revision"), numeric: true, get: (c) => c.revision },
    { key: "feature_map", label: t("node.surface.col_feature_map"), get: (c) => hexId(c.featureMap) },
    { key: "attributes", label: t("node.surface.col_attributes"), numeric: true, get: (c) => c.attributes.length },
    { key: "accepted", label: t("node.surface.col_accepted"), numeric: true, get: (c) => c.accepted.length },
    { key: "generated", label: t("node.surface.col_generated"), numeric: true, get: (c) => c.generated.length },
    { key: "events", label: t("node.surface.col_events"), numeric: true, get: (c) => c.events.length },
  ]);

  type DecisionRow = SurfaceDecision & { index: number };
  const decisionRows = $derived(decisions.map((d) => ({ ...d, index: (report?.decisions ?? []).indexOf(d) })));
  const decisionColumns: DataColumn<DecisionRow>[] = $derived([
    { key: "endpoint", label: t("node.surface.col_endpoint"), title: true, align: "left", get: (d) => endpointLabel(d.endpoint) },
    { key: "element", label: t("node.surface.col_element"), align: "left", get: (d) => element(d) },
    { key: "policy", label: t("node.surface.col_policy"), align: "left", get: (d) => d.policy },
    { key: "verdict", label: t("node.surface.col_verdict"), align: "left", get: (d) => d.verdict },
    { key: "reason", label: t("node.surface.col_reason"), align: "left", get: (d) => d.reason },
  ]);

  const changeKind = (before: string, after: string) => (before === "" ? "added" : after === "" ? "removed" : "changed");
</script>

{#snippet clusterCell(c: SurfaceCluster, col: DataColumn<SurfaceCluster>)}
  {#if col.key === "id"}
    <span class="font-mono">{clusterLabel(c.id)}</span>
  {:else if col.key === "feature_map"}
    <span class="font-mono">{hexId(c.featureMap)}</span>
  {:else}
    {col.get?.(c)}
  {/if}
{/snippet}

{#snippet decisionCell(d: DecisionRow, col: DataColumn<DecisionRow>)}
  {#if col.key === "verdict"}
    <Badge variant={verdictVariant(d.verdict)}>{t(`node.surface.verdict.${d.verdict}`)}</Badge>
    {#if d.mandatory}<Badge variant="muted">{t("node.surface.mandatory")}</Badge>{/if}
  {:else if col.key === "policy"}
    <span class="font-mono text-xs">{d.policy}</span>
  {:else}
    {col.get?.(d)}
  {/if}
{/snippet}

<div class="space-y-6">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-semibold text-[var(--ha-primary-text-color)]">{t("node.surface.title")}</h2>
    <Button variant="outline" onclick={() => void load()} disabled={loading}>{t("common.refresh")}</Button>
  </div>

  {#if loading}
    <LoadingState />
  {:else if error}
    <ErrorState message={error} onRetry={() => void load()} />
  {:else if report}
    {@const r = report}
    <Card class="p-4">
      <p class="text-xs text-[var(--ha-secondary-text-color)]">{t("node.surface.hint")}</p>
      <div class="mt-3 flex flex-wrap items-center gap-2" data-testid="surface-head">
        <Badge variant="default">{t("node.status.surface_level", { level: r.level, origin: levelOrigin(t, r.level_origin) })}</Badge>
        <Badge variant="muted">
          {r.profiles.length > 0 ? t("node.status.surface_profiles", { profiles: r.profiles.join(", ") }) : t("node.status.surface_no_profile")}
        </Badge>
        <Badge variant="muted">
          {r.extensions.length > 0 ? t("node.status.surface_extensions", { extensions: r.extensions.join(", ") }) : t("node.status.surface_no_extension")}
        </Badge>
        <Badge variant={r.assembled ? "success" : "muted"}>{r.assembled ? t("node.surface.assembled") : t("node.surface.not_assembled")}</Badge>
        <span class="font-mono text-xs text-[var(--ha-secondary-text-color)]" title={r.policies}>{t("node.surface.policies", { digest: shortDigest(r.policies) })}</span>
      </div>
    </Card>

    <Card class="p-4">
      <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("node.surface.endpoints")}</h3>
      {#if r.effective.endpoints.length === 0}
        <EmptyState class="mt-3" message={t("node.surface.endpoints_empty")} />
      {:else}
        <div class="mt-3 space-y-2">
          {#each r.effective.endpoints as ep (ep.key)}
            <details class="rounded border border-[var(--ha-divider-color)] p-3" data-testid="surface-endpoint">
              <summary class="flex cursor-pointer flex-wrap items-center gap-2">
                {#if ep.number}<span class="font-mono text-sm text-[var(--ha-secondary-text-color)]">#{ep.number}</span>{/if}
                <span class="font-medium text-[var(--ha-primary-text-color)]">{endpointLabel(ep.key)}</span>
                {#each ep.deviceTypes as dt (dt.id)}
                  <span class="text-xs text-[var(--ha-secondary-text-color)]">{deviceTypeLabel(dt.id)} ({t("node.surface.col_revision")} {dt.revision})</span>
                {/each}
              </summary>
              <div class="mt-3">
                <DataTable
                  rows={ep.clusters}
                  columns={clusterColumns}
                  rowKey={(c) => c.id}
                  cell={clusterCell}
                  emptyMessage={t("node.surface.clusters_empty")}
                />
                {#if ep.clusters.length > 0}
                  <details class="mt-2">
                    <summary class="cursor-pointer text-xs text-[var(--ha-secondary-text-color)]">{t("node.surface.lists")}</summary>
                    <dl class="mt-2 space-y-2 text-xs">
                      {#each ep.clusters as c (c.id)}
                        <div>
                          <dt class="font-medium text-[var(--ha-primary-text-color)]">{clusterLabel(c.id)}</dt>
                          {#each [["col_attributes", c.attributes], ["col_accepted", c.accepted], ["col_generated", c.generated], ["col_events", c.events]] as [label, ids] (label)}
                            <dd class="text-[var(--ha-secondary-text-color)]">
                              {t(`node.surface.${label}`)}: <span class="font-mono">{(ids as string[]).length > 0 ? (ids as string[]).map((id) => hexId(id)).join(" ") : "–"}</span>
                            </dd>
                          {/each}
                        </div>
                      {/each}
                    </dl>
                  </details>
                {/if}
              </div>
            </details>
          {/each}
        </div>
      {/if}
    </Card>

    <Card class="p-4" data-testid="surface-decisions">
      <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("node.surface.decisions")}</h3>
      <p class="mt-1 text-xs text-[var(--ha-secondary-text-color)]">{t("node.surface.decisions_hint")}</p>
      <div class="mt-3 flex flex-wrap gap-2">
        <Select class="w-auto" ariaLabel={t("node.surface.filter_verdict")} value={verdictFilter} onValueChange={(v) => (verdictFilter = v)} options={verdictOptions} />
        <Select class="w-auto" ariaLabel={t("node.surface.filter_endpoint")} value={endpointFilter} onValueChange={(v) => (endpointFilter = v)} options={endpointOptions} />
      </div>
      <div class="mt-3">
        <DataTable
          rows={decisionRows}
          columns={decisionColumns}
          rowKey={(d) => String(d.index)}
          cell={decisionCell}
          emptyMessage={t("node.surface.decisions_empty")}
        />
      </div>
    </Card>

    <Card class="p-4">
      <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("node.surface.withheld")}</h3>
      {#if r.withheld.length === 0}
        <p class="mt-2 text-sm text-[var(--ha-secondary-text-color)]">{t("node.surface.withheld_empty")}</p>
      {:else}
        <ul class="mt-2 space-y-1 text-sm">
          {#each r.withheld as w, i (i)}
            <li class="flex flex-wrap items-center gap-2" data-testid="surface-withheld">
              <span class="font-medium text-[var(--ha-primary-text-color)]">{endpointLabel(w.endpoint)}</span>
              <span class="text-[var(--ha-secondary-text-color)]">{element(w)}</span>
              {#if w.mandatory}<Badge variant="muted">{t("node.surface.mandatory")}</Badge>{/if}
            </li>
          {/each}
        </ul>
      {/if}
    </Card>

    <Card class="p-4">
      <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("node.surface.changes")}</h3>
      {#if r.changed.length === 0}
        <p class="mt-2 text-sm text-[var(--ha-secondary-text-color)]">{t("node.surface.changes_empty")}</p>
      {:else}
        <ul class="mt-2 space-y-3">
          {#each r.changed as c, i (i)}
            <li class="rounded border border-[var(--ha-divider-color)] p-3" data-testid="surface-change">
              <div class="flex flex-wrap items-center gap-2 text-sm">
                <span class="font-medium text-[var(--ha-primary-text-color)]">{endpointLabel(c.endpoint)}</span>
                <span class="text-xs text-[var(--ha-secondary-text-color)]">{new Date(c.at).toLocaleString()}</span>
                <span class="font-mono text-xs text-[var(--ha-secondary-text-color)]" title={`${c.before} → ${c.after}`}>
                  {shortDigest(c.before)} → {shortDigest(c.after)}
                </span>
              </div>
              {#if c.changes.length > 0}
                <ul class="mt-2 space-y-1 text-xs">
                  {#each c.changes as ch, j (j)}
                    <li class="flex flex-wrap items-center gap-2" data-testid="surface-change-item">
                      <Badge variant="muted">{t(`node.surface.change_${changeKind(ch.before, ch.after)}`)}</Badge>
                      <span class="text-[var(--ha-primary-text-color)]">{element(ch)}</span>
                      <span class="text-[var(--ha-secondary-text-color)]">{ch.before || "–"} → {ch.after || "–"}</span>
                    </li>
                  {/each}
                </ul>
              {/if}
            </li>
          {/each}
        </ul>
      {/if}
    </Card>
  {/if}
</div>
