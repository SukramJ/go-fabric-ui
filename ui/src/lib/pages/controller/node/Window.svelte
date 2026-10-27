<!-- A commissioning window on the node for another administrator: the
     codes, the QR code (renderQrSvg) and the timeout; closing it. The
     passcode stays on this page; it is not stored. -->
<script lang="ts">
  import { useFabricUI } from "../../../context";
  import { errorText } from "../../../store.svelte";
  import { renderQrSvg } from "../../../qr";
  import type { NodeWindow } from "../../../api/client";
  import Button from "../../../kit/Button.svelte";
  import Card from "../../../kit/Card.svelte";
  import Select from "../../../kit/Select.svelte";

  let { nodeId }: { nodeId: string } = $props();
  const { t, toast, client } = useFabricUI();

  let win = $state<NodeWindow | null>(null);
  let timeout = $state("900");
  let busy = $state(false);

  const options = $derived(["180", "300", "600", "900"].map((s) => ({ value: s, label: t("controller.window.minutes", { minutes: Number(s) / 60 }) })));
  const qrSvg = $derived(win ? renderQrSvg(win.qr_code) : null);

  async function open() {
    busy = true;
    try {
      win = await client.openNodeWindow(nodeId, Number(timeout));
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
    <div class="flex flex-wrap items-center gap-2">
      <Select class="w-auto" ariaLabel={t("controller.window.timeout")} value={timeout} onValueChange={(v) => (timeout = v as typeof timeout)} {options} />
      <Button disabled={busy} onclick={() => void open()}>{t("controller.window.open")}</Button>
    </div>
  {:else}
    <div class="flex flex-wrap gap-6" data-testid="node-window">
      <div class="w-48">{@html qrSvg}</div>
      <dl class="grid grid-cols-[auto_1fr] content-start gap-x-4 gap-y-1 text-sm">
        <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.window.manual_code")}</dt>
        <dd class="font-mono">{win.manual_code} <Button size="sm" variant="ghost" onclick={() => void copy(win!.manual_code)}>{t("common.copy")}</Button></dd>
        <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.window.qr_code")}</dt>
        <dd class="font-mono text-xs">{win.qr_code}</dd>
        <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.window.discriminator")}</dt>
        <dd>{win.discriminator}</dd>
        <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.window.timeout")}</dt>
        <dd>{t("controller.window.seconds", { seconds: win.timeout_seconds })}</dd>
      </dl>
    </div>
    <Button class="mt-4" variant="outline-destructive" disabled={busy} onclick={() => void close()}>{t("controller.window.close")}</Button>
  {/if}
</Card>
