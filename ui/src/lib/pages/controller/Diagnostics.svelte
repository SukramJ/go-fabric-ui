<!-- The controller's diagnostics (contract 0.9.0, issue #60): discovery
     (GET /controller/mdns, the DNS-SD findings), the open sessions and the
     peers (GET /controller/diagnostics: one row per peer with its
     connection state, address, check-ins and sessions; a commissioned node
     links to its page), and the recent events (GET /controller/events,
     newest first). The node family's Diagnostics page is the model, without
     its compatibility card (node only). A burst of the stream's
     controller_diagnostic_event reloads it in place (the controller store's
     diagnostic epoch). A source that serves no diagnostics
     (capabilities.diagnostics false, the routes 501) shows an empty state. -->
<script lang="ts">
  import { onMount } from "svelte";
  import { useFabricUI } from "../../context";
  import { errorText } from "../../store.svelte";
  import { ApiError, type ControllerMdns, type ControllerPeer, type DiagnosticEvent, type Occupancy, type Session } from "../../api/client";
  import type { DataColumn } from "../../kit/data-table";
  import Badge from "../../kit/Badge.svelte";
  import Button from "../../kit/Button.svelte";
  import Card from "../../kit/Card.svelte";
  import DataTable from "../../kit/DataTable.svelte";
  import EmptyState from "../../kit/EmptyState.svelte";
  import ErrorState from "../../kit/ErrorState.svelte";
  import LoadingState from "../../kit/LoadingState.svelte";
  import { connectionLabel, connectionVariant, nodeName } from "./format";

  const { t, client, controller, routes } = useFabricUI();

  let loading = $state(true);
  let error = $state<string | null>(null);
  let notServed = $state(false);
  let sessions = $state<Session[]>([]);
  let peers = $state<ControllerPeer[]>([]);
  let occupancy = $state<Occupancy | null>(null);
  let mdns = $state<ControllerMdns | null>(null);
  let events = $state<DiagnosticEvent[]>([]);

  // The status says so before any route is asked.
  const served = $derived(controller.status?.capabilities.diagnostics !== false);

  // seq guards against an older load landing after a newer one.
  let seq = 0;

  // A quiet load keeps what the page shows until the answer is there: the
  // stream's reloads must not blank the page.
  async function load(quiet = false) {
    const mine = ++seq;
    if (!served) {
      loading = false;
      return;
    }
    if (!quiet) loading = true;
    error = null;
    try {
      const [d, m, ev] = await Promise.all([client.controllerDiagnostics(), client.controllerMdns(), client.controllerEvents()]);
      if (mine !== seq) return;
      sessions = d.sessions;
      peers = d.peers;
      occupancy = d.occupancy;
      mdns = m;
      events = ev.events;
      notServed = false;
    } catch (err) {
      if (mine !== seq) return;
      if (err instanceof ApiError && err.status === 501) notServed = true;
      else error = errorText(err);
    } finally {
      if (mine === seq) loading = false;
    }
  }

  onMount(() => void load());

  // The trace moved: reload in place. The first run only records the epoch.
  let seenEpoch = controller.diagnosticEpoch;
  $effect(() => {
    const epoch = controller.diagnosticEpoch;
    if (epoch === seenEpoch) return;
    seenEpoch = epoch;
    void load(true);
  });

  function age(seconds: number): string {
    if (seconds < 60) return t("node.diag.age_seconds", { n: seconds });
    if (seconds < 3600) return t("node.diag.age_minutes", { n: Math.floor(seconds / 60) });
    return t("node.diag.age_hours", { n: Math.floor(seconds / 3600) });
  }

  const time = (iso: string | undefined) => (iso ? new Date(iso).toLocaleString() : "—");

  // A peer's name: the commissioned node's label or product, else its id.
  function peerName(p: ControllerPeer): string {
    const n = controller.nodes.find((x) => x.node_id_hex === p.node_id_hex);
    return n ? nodeName(n) : `0x${p.node_id_hex}`;
  }

  const errorFindings = $derived((mdns?.findings ?? []).filter((f) => f.severity === "error"));
  const warningFindings = $derived((mdns?.findings ?? []).filter((f) => f.severity === "warning"));

  const sessionColumns: DataColumn<Session>[] = $derived([
    { key: "session_id", label: t("node.diag.col_session"), sortable: true, numeric: true, align: "left", get: (s) => s.session_id },
    { key: "peer", label: t("controller.diag.col_peer"), sortable: true, get: (s) => s.peer_node_id_hex },
    { key: "transport", label: t("controller.diag.col_transport"), sortable: true, get: (s) => s.transport ?? "" },
    { key: "created", label: t("controller.diag.col_created"), sortable: true, get: (s) => s.created_at ?? "" },
    // Sorted on the seconds, not the rendered age.
    { key: "peer_idle", label: t("controller.diag.col_peer_idle"), sortable: true, get: (s) => s.peer_idle_seconds },
    { key: "flags", label: t("controller.diag.col_flags"), get: (s) => Number(!!s.resumed) + 2 * Number(!!s.peer_lost) },
  ]);

  const peerColumns: DataColumn<ControllerPeer>[] = $derived([
    { key: "node", label: t("controller.diag.col_node"), sortable: true, align: "left", get: (p) => p.node_id_hex },
    { key: "state", label: t("controller.diag.col_state"), sortable: true, get: (p) => p.state ?? "" },
    { key: "reachable", label: t("controller.diag.col_reachable"), sortable: true, get: (p) => Number(p.reachable) },
    { key: "address", label: t("controller.diag.col_address"), get: (p) => p.operational_address ?? "" },
    { key: "offline", label: t("controller.diag.col_offline"), sortable: true, numeric: true, get: (p) => p.time_offline_ms },
    { key: "check_in", label: t("controller.diag.col_check_in"), get: (p) => p.last_check_in_at ?? "" },
    { key: "sessions", label: t("controller.diag.col_sessions"), numeric: true, get: (p) => p.session_ids.length },
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
  {:else if col.key === "peer"}
    <a class="font-mono text-[var(--ha-primary-color)] hover:underline" href={routes.controllerNode + s.peer_node_id_hex}>0x{s.peer_node_id_hex}</a>
  {:else if col.key === "transport"}
    {s.transport ? s.transport.toUpperCase() : "—"}
  {:else if col.key === "created"}
    {time(s.created_at)}
  {:else if col.key === "peer_idle"}
    {age(s.peer_idle_seconds)}
  {:else if col.key === "flags"}
    {#if s.resumed}<Badge variant="muted">{t("controller.diag.resumed")}</Badge>{/if}
    {#if s.peer_lost}<Badge variant="danger">{t("controller.diag.peer_lost")}</Badge>{/if}
  {/if}
{/snippet}

{#snippet peerCell(p: ControllerPeer, col: DataColumn<ControllerPeer>)}
  {#if col.key === "node"}
    {#if p.registered}
      <a class="text-[var(--ha-primary-color)] hover:underline" href={routes.controllerNode + p.node_id_hex} data-testid="peer-link">{peerName(p)}</a>
    {:else}
      <span class="font-mono">0x{p.node_id_hex}</span>
      <Badge variant="muted">{t("controller.diag.unregistered")}</Badge>
    {/if}
    {#if p.load_error}
      <div class="text-xs text-[var(--ha-error-color)]">{p.load_error}</div>
    {/if}
  {:else if col.key === "state"}
    {#if p.state}<Badge variant={connectionVariant(p.state)}>{connectionLabel(t, p.state)}</Badge>{:else}—{/if}
    {#if p.subscription_active}<Badge variant="muted">{t("controller.diag.subscribed")}</Badge>{/if}
  {:else if col.key === "reachable"}
    <Badge variant={p.reachable ? "success" : "warning"}>{p.reachable ? t("controller.diag.yes") : t("controller.diag.no")}</Badge>
  {:else if col.key === "address"}
    <span class="font-mono text-xs">{p.operational_address ?? "—"}</span>
    {#if p.operational_host}<div class="font-mono text-xs text-[var(--ha-secondary-text-color)]">{p.operational_host}</div>{/if}
    {#if p.tcp_unsupported}<div class="text-xs text-[var(--ha-secondary-text-color)]">{t("controller.diag.tcp_unsupported")}</div>{/if}
  {:else if col.key === "offline"}
    {p.time_offline_ms > 0 ? age(Math.floor(p.time_offline_ms / 1000)) : "—"}
  {:else if col.key === "check_in"}
    {#if p.last_check_in_at || p.next_expected_check_in}
      <div class="text-xs">{t("controller.diag.check_in_last", { at: time(p.last_check_in_at) })}</div>
      <div class="text-xs text-[var(--ha-secondary-text-color)]">{t("controller.diag.check_in_next", { at: time(p.next_expected_check_in) })}</div>
    {:else}—{/if}
  {:else if col.key === "sessions"}
    {p.session_ids.length > 0 ? p.session_ids.join(", ") : "—"}
  {/if}
{/snippet}

<div class="space-y-6">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-semibold text-[var(--ha-primary-text-color)]">{t("controller.diag.title")}</h2>
    {#if served && !notServed}
      <Button variant="outline" onclick={() => void load()} disabled={loading}>{t("common.refresh")}</Button>
    {/if}
  </div>

  {#if !served || notServed}
    <EmptyState message={t("controller.diag.not_served")} />
  {:else if loading}
    <LoadingState />
  {:else if error}
    <ErrorState message={error} onRetry={() => void load()} />
  {:else}
    <Card class="p-4">
      <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("node.diag.discovery")}</h3>
      {#if errorFindings.length === 0 && warningFindings.length === 0}
        <p class="mt-2 text-sm text-[var(--ha-success-color)]">{t("controller.diag.discovery_ok")}</p>
      {/if}
      <ul class="mt-2 space-y-2">
        {#each [...errorFindings, ...warningFindings] as finding (finding.code + (finding.service ?? ""))}
          <li class="flex flex-wrap items-start gap-2 text-sm" data-testid="finding">
            <Badge variant={finding.severity === "error" ? "danger" : "warning"}>{t(`node.diag.severity.${finding.severity}`)}</Badge>
            <span class="min-w-[16rem] flex-1 text-[var(--ha-primary-text-color)]">{finding.message}</span>
          </li>
        {/each}
      </ul>
    </Card>

    <Card class="p-4">
      <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("controller.diag.peers")}</h3>
      <p class="mt-1 text-xs text-[var(--ha-secondary-text-color)]">{t("controller.diag.peers_hint")}</p>
      <div class="mt-3">
        <DataTable
          rows={peers}
          columns={peerColumns}
          rowKey={(p) => `${p.fabric_index}:${p.node_id_hex}`}
          cell={peerCell}
          persistKey="controller-peers"
          initialSort={{ key: "node", asc: true }}
          emptyMessage={t("controller.diag.no_peers")}
        />
      </div>
    </Card>

    <Card class="p-4">
      <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("controller.diag.sessions")}</h3>
      <p class="mt-1 text-xs text-[var(--ha-secondary-text-color)]">{t("controller.diag.sessions_hint")}</p>
      {#if occupancy && occupancy.capacity > 0}
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
          persistKey="controller-sessions"
          initialSort={{ key: "session_id", asc: true }}
          emptyMessage={t("controller.diag.no_sessions")}
        />
      </div>
    </Card>

    <!-- Inside the loaded branch: "nothing recorded" is a claim about the
         controller and must not be printed under a failed fetch. -->
    <Card class="p-4">
      <h3 class="font-medium text-[var(--ha-primary-text-color)]">{t("node.diag.events")}</h3>
      <p class="mt-1 text-xs text-[var(--ha-secondary-text-color)]">{t("node.diag.events_hint")}</p>
      {#if events.length === 0}
        <EmptyState class="mt-3" message={t("controller.diag.events_empty")} />
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
