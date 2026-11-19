<!-- A commissioning window on the node for another administrator: the
     codes, the QR code (renderQrSvg) and the time left; closing it. The
     passcode stays on this page; it is not stored. The codes leave the
     page when the window's time is up, and Close is offered also when the
     page does not know of an open window (one opened before the tab was
     left, or elsewhere): there is no route that reads the window. -->
<script lang="ts">
  import { onDestroy } from "svelte";
  import { useFabricUI } from "../../../context";
  import { errorText } from "../../../store.svelte";
  import { qrSvgOrNull } from "../../../qr";
  import type { NodeWindow } from "../../../api/client";
  import Button from "../../../kit/Button.svelte";
  import Card from "../../../kit/Card.svelte";
  import Select from "../../../kit/Select.svelte";

  let { nodeId }: { nodeId: string } = $props();
  const { t, toast, client } = useFabricUI();

  let win = $state<NodeWindow | null>(null);
  let timeout = $state("900");
  let busy = $state(false);
  // Seconds until the open window's timeout, counted on this page.
  let left = $state(0);
  let expired = $state(false);
  let timer: ReturnType<typeof setInterval> | undefined;

  function stopTimer() {
    if (timer !== undefined) clearInterval(timer);
    timer = undefined;
  }
  onDestroy(stopTimer);

  function startTimer(seconds: number) {
    stopTimer();
    const until = Date.now() + seconds * 1000;
    left = seconds;
    timer = setInterval(() => {
      left = Math.max(0, Math.ceil((until - Date.now()) / 1000));
      if (left === 0) {
        stopTimer();
        win = null;
        expired = true;
      }
    }, 1000);
  }

  const options = $derived(["180", "300", "600", "900"].map((s) => ({ value: s, label: t("controller.window.minutes", { minutes: Number(s) / 60 }) })));
  const qrSvg = $derived(qrSvgOrNull(win?.qr_code));

  async function open() {
    busy = true;
    try {
      win = await client.openNodeWindow(nodeId, Number(timeout));
      expired = false;
      startTimer(win.timeout_seconds);
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      busy = false;
    }
  }

  async function close() {
    busy = true;
    try {
      await client.closeNodeWindow(nodeId);
      stopTimer();
      win = null;
      toast.success(t("controller.window.closed"));
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      busy = false;
    }
  }

  async function copy(v: string) {
    try {
      await navigator.clipboard.writeText(v);
      toast.success(t("node.pair.copied"));
    } catch {
      toast.error(t("node.pair.copy_failed"));
    }
  }
</script>

<Card class="p-4">
  <p class="mb-3 text-sm text-[var(--ha-secondary-text-color)]">{t("controller.window.hint")}</p>
  {#if !win}
    {#if expired}
      <p class="mb-3 text-sm text-[var(--ha-secondary-text-color)]" role="status">{t("controller.window.expired")}</p>
    {/if}
    <div class="flex flex-wrap items-center gap-2">
      <Select class="w-auto" ariaLabel={t("controller.window.timeout")} value={timeout} onValueChange={(v) => (timeout = v as typeof timeout)} {options} />
      <Button disabled={busy} onclick={() => void open()}>{t("controller.window.open")}</Button>
      <Button variant="outline-destructive" disabled={busy} onclick={() => void close()}>{t("controller.window.close")}</Button>
    </div>
    <p class="mt-2 text-xs text-[var(--ha-secondary-text-color)]">{t("controller.window.close_hint")}</p>
  {:else}
    <div class="flex flex-wrap gap-6" data-testid="node-window">
      {#if qrSvg}
        <div class="w-48">{@html qrSvg}</div>
      {/if}
      <dl class="grid grid-cols-[auto_1fr] content-start gap-x-4 gap-y-1 text-sm">
        <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.window.manual_code")}</dt>
        <dd class="font-mono">{win.manual_code} <Button size="sm" variant="ghost" onclick={() => void copy(win!.manual_code)}>{t("common.copy")}</Button></dd>
        <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.window.qr_code")}</dt>
        <dd class="font-mono text-xs">{win.qr_code}</dd>
        <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.window.discriminator")}</dt>
        <dd>{win.discriminator}</dd>
        <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.window.left")}</dt>
        <dd data-testid="window-left">{t("controller.window.seconds", { seconds: left })}</dd>
      </dl>
    </div>
    <Button class="mt-4" variant="outline-destructive" disabled={busy} onclick={() => void close()}>{t("controller.window.close")}</Button>
  {/if}
</Card>
