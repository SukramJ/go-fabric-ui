<!-- The commissioning page's camera scan (contract unchanged: the payload
     fills CommissionRequest.code). The decoder is a chunk of its own
     (lib/scan), loaded on the first click. A page outside a secure context
     (plain HTTP on a LAN address, an Ingress without TLS) has no camera:
     it explains that instead of offering the button, as the clipboard
     refusal on the pairing page does. -->
<script lang="ts">
  import { onDestroy, tick } from "svelte";
  import { useFabricUI } from "../../context";
  import { errorText } from "../../store.svelte";
  import Button from "../../kit/Button.svelte";

  let { onCode, disabled = false }: { onCode: (payload: string) => void; disabled?: boolean } = $props();
  const { t, toast } = useFabricUI();

  const canScan = typeof window !== "undefined" && window.isSecureContext === true && !!navigator.mediaDevices?.getUserMedia;

  let open = $state(false);
  let starting = $state(false);
  let video = $state<HTMLVideoElement | null>(null);
  let stop: (() => void) | null = null;

  async function start() {
    open = true;
    starting = true;
    try {
      const scan = await import("../../scan");
      // The video element renders once open is true.
      await tick();
      if (!video) throw new Error("no video element");
      stop = await scan.startScan(video, (payload) => {
        stop = null;
        open = false;
        onCode(payload);
        toast.success(t("controller.commission.scan_read"));
      });
    } catch (err) {
      open = false;
      toast.error(t("controller.commission.scan_failed", { error: errorText(err) }));
    } finally {
      starting = false;
    }
  }

  function cancel() {
    stop?.();
    stop = null;
    open = false;
  }

  onDestroy(cancel);
</script>

{#if !canScan}
  <p class="text-xs text-[var(--ha-secondary-text-color)]" data-testid="scan-insecure">{t("controller.commission.scan_insecure")}</p>
{:else if open}
  <div class="w-full space-y-2" data-testid="scan-view">
    <!-- svelte-ignore a11y_media_has_caption -->
    <video bind:this={video} class="max-h-72 w-full max-w-sm rounded border border-[var(--ha-divider-color)] bg-black" muted></video>
    <div class="flex items-center gap-2">
      <span class="text-xs text-[var(--ha-secondary-text-color)]">{starting ? t("controller.commission.scan_starting") : t("controller.commission.scan_hint")}</span>
      <Button type="button" variant="outline" size="sm" onclick={cancel}>{t("common.cancel")}</Button>
    </div>
  </div>
{:else}
  <Button type="button" variant="outline" {disabled} onclick={() => void start()}>{t("controller.commission.scan")}</Button>
{/if}
