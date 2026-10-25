<!-- The node's status: whether it runs, listens, is healthy and
     advertised, its counts and the commissioning window. Ported from the
     status card of openccu-loom assets/ui/src/routes/matter/Matter.svelte:47-100
     at b40b289a, without the tabs (the shell or the host navigates). The
     host's configuration warnings come first, in their own card. An open
     window counts down from the deadline the status reports. -->
<script lang="ts">
  import { onMount } from "svelte";
  import { useFabricUI } from "../../context";
  import { secondsUntil } from "../../countdown";
  import Badge from "../../kit/Badge.svelte";
  import Card from "../../kit/Card.svelte";
  import ErrorState from "../../kit/ErrorState.svelte";
  import LoadingState from "../../kit/LoadingState.svelte";
  import HostWarnings from "./HostWarnings.svelte";

  const { t, node } = useFabricUI();

  let now = $state(Date.now());

  onMount(() => {
    void node.loadStatus();
    const tick = setInterval(() => (now = Date.now()), 1000);
    return () => clearInterval(tick);
  });

  const s = $derived(node.status);
  // Seconds left in the open window, from its deadline; the duration it
  // was opened for when the server reports no deadline.
  const windowLeft = $derived(
    s?.commissioning_window_open
      ? (secondsUntil(s.window_deadline, now) ?? ((s.commissioning_window_duration_seconds ?? 0) > 0 ? s.commissioning_window_duration_seconds! : null))
      : null,
  );
</script>

{#if node.statusLoading && !s}
  <LoadingState message={t("common.loading")} />
{:else if node.statusError && !s}
  <ErrorState message={node.statusError} onRetry={() => void node.loadStatus()} />
{:else if s && !s.enabled}
  <HostWarnings warnings={s.warnings ?? []} />
  <Card class="p-4">
    <p class="text-sm font-medium text-[var(--ha-secondary-text-color)]">{t("node.status.disabled")}</p>
  </Card>
{:else if s}
  <HostWarnings warnings={s.warnings ?? []} />
  <Card class="p-4" data-testid="node-status">
    <div class="flex flex-wrap items-center gap-4">
      <span class="font-medium text-[var(--ha-primary-text-color)]">{t("node.status.enabled")}</span>
      <Badge variant={s.listening ? "success" : "muted"}>
        {s.listening ? t("node.status.listening") : t("node.status.not_listening")}
      </Badge>
      {#if s.healthy !== undefined}
        <Badge variant={s.healthy ? "success" : "danger"} title={s.health_note}>
          {s.healthy ? t("node.status.healthy") : t("node.status.unhealthy")}
        </Badge>
      {/if}
      <span class="text-sm text-[var(--ha-secondary-text-color)]">{t("node.status.endpoints", { count: s.endpoint_count })}</span>
      <span class="text-sm text-[var(--ha-secondary-text-color)]">{t("node.status.fabrics", { count: s.fabric_count })}</span>
      {#if s.enabled_count !== undefined}
        <span class="text-sm text-[var(--ha-secondary-text-color)]">{t("node.status.exposed", { count: s.enabled_count })}</span>
      {/if}
      {#if s.advertising}
        <Badge variant="default">{t("node.status.advertising")}</Badge>
      {/if}
      {#if s.commissioning_window_open}
        <Badge variant="warning">
          {t("node.pair.window_open")}
          {#if windowLeft !== null}
            <span data-testid="window-left">({windowLeft}s)</span>
          {/if}
        </Badge>
      {/if}
    </div>
    {#if s.listen_addr}
      <p class="mt-2 text-xs text-[var(--ha-secondary-text-color)]">{t("node.status.listen_addr", { addr: s.listen_addr })}</p>
    {/if}
  </Card>
{/if}
