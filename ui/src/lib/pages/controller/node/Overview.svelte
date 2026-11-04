<!-- A node's record, properties and connection, and the ping (admin,
     contract 0.5.0): a Matter read of DataModelRevision, with its outcome
     per address and the sentence that says what was measured. -->
<script lang="ts">
  import { useFabricUI } from "../../../context";
  import { errorText } from "../../../store.svelte";
  import type { ControllerNodeDetail, PingResult } from "../../../api/client";
  import Badge from "../../../kit/Badge.svelte";
  import Button from "../../../kit/Button.svelte";
  import Card from "../../../kit/Card.svelte";
  import { connectionLabel, connectionVariant, vendorLabel } from "../format";

  let { detail }: { detail: ControllerNodeDetail } = $props();
  const { t, toast, client } = useFabricUI();

  const n = $derived(detail.node);
  const r = $derived(detail.record);
  const p = $derived(detail.properties);

  const flags = $derived(
    (
      [
        ["supports_thread", p.supports_thread],
        ["supports_wifi", p.supports_wifi],
        ["supports_ethernet", p.supports_ethernet],
        ["mains_powered", p.mains_powered],
        ["battery_powered", p.battery_powered],
        ["intermittently_connected", p.intermittently_connected],
        ["thread_sleepy_end_device", p.thread_sleepy_end_device ?? false],
      ] as const
    ).filter(([, on]) => on).map(([k]) => k),
  );

  let pinging = $state(false);
  let ping = $state<PingResult | null>(null);
  async function runPing() {
    pinging = true;
    try {
      ping = await client.pingNode(n.node_id_hex);
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      pinging = false;
    }
  }

  let busy = $state(false);
  async function setAutoSubscribe(on: boolean) {
    busy = true;
    try {
      await client.setNodeAutoSubscribe(n.node_id_hex, on);
      toast.success(t("controller.overview.saved"));
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      busy = false;
    }
  }
</script>

<div class="grid gap-4 md:grid-cols-2">
  <Card class="p-4" data-testid="node-overview">
    <h3 class="mb-2 font-medium text-[var(--ha-primary-text-color)]">{t("controller.overview.node")}</h3>
    <dl class="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm">
      <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.overview.node_id")}</dt><dd class="font-mono">0x{n.node_id_hex}</dd>
      <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.overview.vendor")}</dt><dd>{vendorLabel(n.vendor_name, n.vendor_id)}</dd>
      <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.overview.product")}</dt><dd>{n.product_name ?? ""}</dd>
      <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.overview.software")}</dt><dd>{p.software_version ?? ""}</dd>
      <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.overview.serial")}</dt><dd>{p.serial_number ?? ""}</dd>
      <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.overview.connection")}</dt>
      <dd><Badge variant={connectionVariant(n.connection)}>{connectionLabel(t, n.connection)}</Badge>
        {#if n.subscription_active}<Badge>{t("controller.nodes.subscribed")}</Badge>{/if}
        {#if n.disabled}<Badge variant="warning">{t("controller.nodes.disabled")}</Badge>{/if}</dd>
      <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.overview.addresses")}</dt><dd class="font-mono text-xs">{r.addresses.join(", ")}</dd>
      <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.overview.fabric_index")}</dt><dd>{r.fabric_index_on_peer ?? "–"}</dd>
      <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.overview.commissioned_at")}</dt><dd>{r.commissioned_at ? new Date(r.commissioned_at).toLocaleString() : "–"}</dd>
    </dl>
    <div class="mt-3 flex flex-wrap gap-2">
      <Button size="sm" variant="outline" disabled={busy} onclick={() => void setAutoSubscribe(!n.subscription_active)}>
        {n.subscription_active ? t("controller.overview.unsubscribe") : t("controller.overview.subscribe")}
      </Button>
      <Button size="sm" variant="outline" disabled={pinging} onclick={() => void runPing()}>
        {pinging ? t("controller.ping.running") : t("controller.ping.button")}
      </Button>
    </div>
    {#if ping}
      <div class="mt-3 text-sm" data-testid="ping-result">
        <p class="text-[var(--ha-secondary-text-color)]">{ping.per_address ? t("controller.ping.method_per_address") : t("controller.ping.method_session")}</p>
        <p class="mt-1">
          <Badge variant={ping.reachable ? "success" : "danger"}>{ping.reachable ? t("controller.ping.reachable") : t("controller.ping.unreachable")}</Badge>
          {#if ping.rtt_ms !== undefined}<span class="ml-2">{t("controller.ping.rtt", { ms: ping.rtt_ms })}</span>{/if}
          {#if ping.error}<span class="ml-2 text-[var(--ha-error-color)]">{ping.error}</span>{/if}
        </p>
        {#if ping.addresses.length > 0}
          <ul class="mt-1 space-y-0.5 font-mono text-xs">
            {#each ping.addresses as a (a.ip)}
              <li data-testid="ping-address">
                {a.ip} · {a.reachable ? t("controller.ping.reachable") : t("controller.ping.unreachable")}{#if a.rtt_ms !== undefined}&nbsp;· {t("controller.ping.rtt", { ms: a.rtt_ms })}{/if}
              </li>
            {/each}
          </ul>
        {/if}
      </div>
    {/if}
  </Card>
  <Card class="p-4">
    <h3 class="mb-2 font-medium text-[var(--ha-primary-text-color)]">{t("controller.overview.properties")}</h3>
    <div class="flex flex-wrap gap-1">
      {#each flags as f (f)}<Badge variant="muted">{t(`controller.overview.flag.${f}`)}</Badge>{/each}
    </div>
    {#if p.idle_mode_duration_seconds}
      <p class="mt-2 text-sm text-[var(--ha-secondary-text-color)]">{t("controller.overview.idle_mode", { seconds: p.idle_mode_duration_seconds })}</p>
    {/if}
    {#if p.specification_version}
      <p class="mt-1 text-sm text-[var(--ha-secondary-text-color)]">{t("controller.overview.spec_version", { version: `0x${p.specification_version.toString(16).toUpperCase().padStart(8, "0")}` })}</p>
    {/if}
  </Card>
</div>
