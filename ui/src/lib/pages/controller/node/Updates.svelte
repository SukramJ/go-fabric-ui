<!-- The node's entries of the update queue, an available update to
     consent to, and the check. -->
<script lang="ts">
  import { onMount } from "svelte";
  import { useFabricUI } from "../../../context";
  import { errorText } from "../../../store.svelte";
  import type { AvailableUpdate } from "../../../api/client";
  import Badge from "../../../kit/Badge.svelte";
  import Button from "../../../kit/Button.svelte";
  import Card from "../../../kit/Card.svelte";
  import EmptyState from "../../../kit/EmptyState.svelte";

  let { nodeId }: { nodeId: string } = $props();
  const { t, toast, client, controller } = useFabricUI();

  let available = $state<AvailableUpdate[]>([]);
  let busy = $state(false);

  onMount(() => void controller.loadUpdates());

  const queue = $derived((controller.updates?.queue ?? []).filter((u) => u.node_id_hex === nodeId));

  async function check() {
    busy = true;
    try {
      available = (await client.checkUpdates()).updates.filter((u) => u.node_id_hex === nodeId);
      if (available.length === 0) toast.info(t("controller.updates.none_available"));
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      busy = false;
    }
  }

  async function consent(u: AvailableUpdate) {
    busy = true;
    try {
      await client.consentUpdate({ node_id_hex: u.node_id_hex, vendor_id: u.vendor_id, product_id: u.product_id, target_software_version: u.software_version });
      toast.success(t("controller.updates.consented"));
      available = available.filter((a) => a !== u);
      await controller.loadUpdates();
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      busy = false;
    }
  }
</script>

{#if controller.updates && !controller.updates.available}
  <EmptyState message={t("controller.updates.unavailable")} />
{:else}
  <Card class="p-4" data-testid="node-updates">
    {#if queue.length === 0}
      <p class="text-sm text-[var(--ha-secondary-text-color)]">{t("controller.updates.queue_empty")}</p>
    {:else}
      <ul class="space-y-1 text-sm">
        {#each queue as u (u.target_software_version)}
          <li class="flex flex-wrap items-center gap-2">
            <span>{t("controller.updates.version", { version: u.target_software_version })}</span>
            <Badge>{t(`controller.updates.status.${u.status}`)}</Badge>
            {#if u.progress_status}<span class="text-xs text-[var(--ha-secondary-text-color)]">{u.progress_status}</span>{/if}
            {#if u.transferred !== undefined && u.length}<span class="text-xs">{Math.round((u.transferred / u.length) * 100)}%</span>{/if}
          </li>
        {/each}
      </ul>
    {/if}
    <div class="mt-3 flex flex-wrap items-center gap-2">
      <Button variant="outline" disabled={busy} onclick={() => void check()}>{t("controller.updates.check")}</Button>
      {#each available as u (u.software_version)}
        <Button disabled={busy} onclick={() => void consent(u)}>{t("controller.updates.consent", { version: u.software_version_string || String(u.software_version) })}</Button>
      {/each}
    </div>
  </Card>
{/if}
