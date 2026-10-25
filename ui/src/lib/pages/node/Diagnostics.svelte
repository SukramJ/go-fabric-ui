<!-- Diagnostics: discovery (the live DNS-SD advertisement and its
     findings), the connected controllers and the session-id occupancy,
     ecosystem compatibility, and the recent events. Ported from openccu-loom
     assets/ui/src/routes/matter/MatterDiagnostics.svelte:1-352 at b40b289a;
     the endpoint inspector moved to the Endpoints page. -->
<script lang="ts">
  import { onMount } from "svelte";
  import { useFabricUI } from "../../context";
  import { errorText } from "../../store.svelte";
  import type { Compatibility, DiagnosticEvent, Mdns, Occupancy, Session } from "../../api/client";
  import type { DataColumn } from "../../kit/data-table";
  import Badge from "../../kit/Badge.svelte";
  import Button from "../../kit/Button.svelte";
  import Card from "../../kit/Card.svelte";
  import DataTable from "../../kit/DataTable.svelte";
  import EmptyState from "../../kit/EmptyState.svelte";
  import ErrorState from "../../kit/ErrorState.svelte";
  import LoadingState from "../../kit/LoadingState.svelte";

  const { t, client } = useFabricUI();

  let loading = $state(true);
  let error = $state<string | null>(null);
  let sessions = $state<Session[]>([]);
  let occupancy = $state<Occupancy | null>(null);
  let mdns = $state<Mdns | null>(null);
  let compat = $state<Compatibility | null>(null);
  let events = $state<DiagnosticEvent[]>([]);

  async function load() {
    loading = true;
    error = null;
    try {
      const [s, m, c, ev] = await Promise.all([client.sessions(), client.mdns(), client.compatibility(), client.events()]);
      sessions = s.sessions;
      occupancy = s.occupancy;
      mdns = m;
      compat = c;
      events = ev.events;
    } catch (err) {
      error = errorText(err);
    } finally {
      loading = false;
    }
  }

  onMount(load);

  function age(seconds: number): string {
    if (seconds < 60) return t("node.diag.age_seconds", { n: seconds });
    if (seconds < 3600) return t("node.diag.age_minutes", { n: Math.floor(seconds / 60) });
    return t("node.diag.age_hours", { n: Math.floor(seconds / 3600) });
  }

  // A controller that stopped talking leaves its session open, so the
  // peer-side age separates a live controller from a departed one; under
  // five minutes is ordinary quiet (loom MatterDiagnostics.svelte:60-70).
  function peerVariant(seconds: number): "success" | "warning" | "danger" {
    if (seconds < 300) return "success";
    if (seconds < 1800) return "warning";
    return "danger";
  }

  const errorFindings = $derived((mdns?.findings ?? []).filter((f) => f.severity === "error"));
  const warningFindings = $derived((mdns?.findings ?? []).filter((f) => f.severity === "warning"));

  const sessionColumns: DataColumn<Session>[] = $derived([
    { key: "session_id", label: t("node.diag.col_session"), sortable: true, numeric: true, align: "left", get: (s) => s.session_id },
    { key: "fabric_index", label: t("node.diag.col_fabric"), sortable: true, numeric: true, get: (s) => s.fabric_index },
    // Sorted on the seconds, not the rendered age.
    { key: "peer_idle", label: t("node.diag.col_peer_idle"), sortable: true, get: (s) => s.peer_idle_seconds },
    { key: "subscriptions", label: t("node.diag.col_subscriptions"), sortable: true, numeric: true, get: (s) => s.subscriptions },
  ]);

  const eventClass = (sev: string) =>
    sev === "error"
      ? "bg-[color-mix(in_srgb,var(--ha-error-color)_15%,transparent)] text-[var(--ha-error-color)]"
      : sev === "warning"
        ? "bg-[color-mix(in_srgb,var(--ha-warning-color)_15%,transparent)] text-[var(--ha-warning-color)]"
        : "bg-[var(--ha-secondary-background-color)] text-[var(--ha-secondary-text-color)]";
</script>

{#snippet sessionCell(s: Session, col: DataColumn<Session>)}
  {#if col.key === "session_id"}
    {s.session_id}
    {#if s.is_pase}<Badge variant="muted">{t("node.diag.pase")}</Badge>{/if}
  {:else if col.key === "fabric_index"}
    {s.fabric_index}
  {:else if col.key === "peer_idle"}
    <Badge variant={peerVariant(s.peer_idle_seconds)}>{age(s.peer_idle_seconds)}</Badge>
  {:else if col.key === "subscriptions"}
    {s.subscriptions}
    {#if s.subscriptions === 0 && !s.is_pase}
      <span class="ml-2 text-xs text-[var(--ha-warning-color)]">{t("node.diag.no_subscriptions")}</span>
    {/if}
  {/if}
{/snippet}

<div class="space-y-6">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-semibold text-[var(--ha-primary-text-color)]">{t("node.diag.title")}</h2>
    <Button variant="outline" onclick={() => void load()} disabled={loading}>{t("common.refresh")}</Button>
  </div>

  {#if loading}
    <LoadingState />
  {:else if error}
    <ErrorState message={error} onRetry={() => void load()} />
  {:else}
    <Card class="p-4">
      <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("node.diag.discovery")}</h3>
      {#if !mdns?.advertising}
        <p class="mt-2 text-sm text-[var(--ha-secondary-text-color)]">{t("node.diag.not_advertising")}</p>
      {:else}
        {#if errorFindings.length === 0 && warningFindings.length === 0}
          <p class="mt-2 text-sm text-[var(--ha-success-color)]">{t("node.diag.discovery_ok")}</p>
        {/if}
        <ul class="mt-2 space-y-2">
          {#each [...errorFindings, ...warningFindings] as finding (finding.code + (finding.service ?? ""))}
            <li class="flex flex-wrap items-start gap-2 text-sm">
              <Badge variant={finding.severity === "error" ? "danger" : "warning"}>{t(`node.diag.severity.${finding.severity}`)}</Badge>
              <span class="min-w-[16rem] flex-1 text-[var(--ha-primary-text-color)]">{finding.message}</span>
            </li>
          {/each}
        </ul>
        <div class="mt-3 space-y-1 text-xs text-[var(--ha-secondary-text-color)]">
          {#each mdns.services as svc (svc.service_type + svc.instance_name)}
            <div>
              <span class="font-mono">{svc.service_type}</span>
              · {t("node.diag.port")}&nbsp;{svc.port} · {svc.addresses.join(", ")}
              {#if svc.subtypes.length > 0}· {svc.subtypes.join(" ")}{/if}
            </div>
          {/each}
        </div>
      {/if}
    </Card>

    <Card class="p-4">
      <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("node.diag.sessions")}</h3>
      <p class="mt-1 text-xs text-[var(--ha-secondary-text-color)]">{t("node.diag.sessions_hint")}</p>
      <!-- An id staked by a handshake that never completed holds its slot
           and appears in no row below; this line makes a filling table visible. -->
      {#if occupancy}
        <p class="mt-1 text-xs text-[var(--ha-secondary-text-color)]" data-testid="occupancy">
          {t("node.diag.sessions_occupancy", { live: occupancy.live, reserved: occupancy.reserved, free: occupancy.free, capacity: occupancy.capacity })}
        </p>
      {/if}
      <div class="mt-3">
        <DataTable
          rows={sessions}
          columns={sessionColumns}
          rowKey={(s) => String(s.session_id)}
          cell={sessionCell}
          persistKey="node-sessions"
          initialSort={{ key: "session_id", asc: true }}
          emptyMessage={t("node.diag.no_sessions")}
        />
      </div>
    </Card>

    {#if compat}
      <Card class="p-4">
        <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("node.diag.compatibility")}</h3>
        <div class="mt-2 flex flex-wrap gap-2">
          {#each compat.ecosystems as eco (eco.fabric_index)}
            <Badge variant="muted">{t(`node.diag.ecosystem.${eco.ecosystem}`)}</Badge>
          {/each}
        </div>
        {#if compat.findings.length === 0}
          <p class="mt-3 text-sm text-[var(--ha-success-color)]">{t("node.diag.compat_ok")}</p>
        {:else}
          <ul class="mt-3 space-y-2">
            {#each compat.findings as finding (finding.ecosystem + finding.code)}
              <li class="flex flex-wrap items-start gap-2 text-sm">
                <Badge variant="warning">{t(`node.diag.ecosystem.${finding.ecosystem}`)}</Badge>
                <span class="min-w-[16rem] flex-1 text-[var(--ha-primary-text-color)]">{finding.message}</span>
              </li>
            {/each}
          </ul>
        {/if}
      </Card>
    {/if}

    <!-- Inside the loaded branch: "nothing recorded" is a claim about the
         node and must not be printed under a failed fetch. -->
    <Card class="p-4">
      <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("node.diag.events")}</h3>
      <p class="mt-1 text-xs text-[var(--ha-secondary-text-color)]">{t("node.diag.events_hint")}</p>
      {#if events.length === 0}
        <EmptyState class="mt-3" message={t("node.diag.events_empty")} />
      {:else}
        <ul class="mt-3 space-y-2">
          <!-- Keyed by position: two events can share second and message. -->
          {#each events as event, i (i)}
            <li class="flex gap-3 text-sm" data-testid="event">
              <span class="shrink-0 font-mono text-xs text-[var(--ha-secondary-text-color)]">{new Date(event.at).toLocaleTimeString()}</span>
              <span class="shrink-0 rounded px-1.5 py-0.5 text-xs {eventClass(event.severity)}">{t(`node.diag.kind_${event.kind}`)}</span>
              <span class="text-[var(--ha-primary-text-color)]">{event.message}</span>
            </li>
          {/each}
        </ul>
      {/if}
    </Card>
  {/if}
</div>
