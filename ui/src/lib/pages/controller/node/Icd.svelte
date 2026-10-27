<!-- The node's ICD icd; registering this controller as its check-in
     client, or unregistering it. -->
<script lang="ts">
  import { onMount } from "svelte";
  import { useFabricUI } from "../../../context";
  import { errorText } from "../../../store.svelte";
  import type { IcdState } from "../../../api/client";
  import Badge from "../../../kit/Badge.svelte";
  import Button from "../../../kit/Button.svelte";
  import Card from "../../../kit/Card.svelte";
  import EmptyState from "../../../kit/EmptyState.svelte";
  import ErrorState from "../../../kit/ErrorState.svelte";
  import LoadingState from "../../../kit/LoadingState.svelte";

  let { nodeId }: { nodeId: string } = $props();
  const { t, toast, confirm, client } = useFabricUI();

  let icd = $state<IcdState | null>(null);
  let error = $state<string | null>(null);
  let busy = $state(false);

  async function load() {
    error = null;
    try {
      icd = await client.nodeIcd(nodeId);
    } catch (err) {
      error = errorText(err);
    }
  }
  onMount(() => void load());

  async function register(allowMultiAdmin = false) {
    busy = true;
    try {
      icd = await client.registerNodeIcd(nodeId, allowMultiAdmin);
      toast.success(t("controller.icd.registered_done"));
    } catch (err) {
      // Other administrators' vendors may not support LIT: ask, then retry.
      if (!allowMultiAdmin && (err as { status?: number }).status === 409) {
        busy = false;
        const ok = await confirm({ title: t("controller.icd.multi_admin"), body: errorText(err), confirmLabel: t("controller.icd.register") });
        if (ok) await register(true);
        return;
      }
      toast.error(errorText(err));
    } finally {
      busy = false;
    }
  }

  async function unregister(force: boolean) {
    busy = true;
    try {
      icd = await client.unregisterNodeIcd(nodeId, force);
      toast.success(t("controller.icd.unregistered_done"));
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      busy = false;
    }
  }
</script>

{#if error}
  <ErrorState message={error} onRetry={() => void load()} />
{:else if !icd}
  <LoadingState message={t("common.loading")} />
{:else if !icd.supported}
  <EmptyState message={t("controller.icd.unsupported")} />
{:else}
  <Card class="p-4" data-testid="node-icd">
    <div class="flex flex-wrap gap-1">
      <Badge variant={icd.registered ? "success" : "muted"}>{icd.registered ? t("controller.icd.registered") : t("controller.icd.unregistered")}</Badge>
      {#if icd.operating_mode}<Badge>{t(`controller.icd.mode.${icd.operating_mode}`)}</Badge>{/if}
      {#if icd.lit_supported}<Badge variant="muted">{t("controller.icd.lit_supported")}</Badge>{/if}
      {#if icd.awake}<Badge variant="success">{t("controller.icd.awake")}</Badge>{/if}
      {#if icd.available === false}<Badge variant="warning">{t("controller.icd.unavailable")}</Badge>{/if}
    </div>
    <dl class="mt-3 grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm">
      <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.icd.next_check_in")}</dt><dd>{icd.next_check_in ? new Date(icd.next_check_in).toLocaleString() : "–"}</dd>
      <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.icd.last_check_in")}</dt><dd>{icd.last_check_in ? new Date(icd.last_check_in).toLocaleString() : "–"}</dd>
    </dl>
    <div class="mt-3 flex flex-wrap gap-2">
      {#if icd.registered}
        <Button variant="outline" disabled={busy} onclick={() => void unregister(false)}>{t("controller.icd.unregister")}</Button>
        <Button variant="outline-destructive" disabled={busy} onclick={() => void unregister(true)}>{t("controller.icd.forget")}</Button>
      {:else}
        <Button disabled={busy} onclick={() => void register()}>{t("controller.icd.register")}</Button>
      {/if}
    </div>
  </Card>
{/if}
