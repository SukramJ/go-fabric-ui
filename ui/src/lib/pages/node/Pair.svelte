<!-- Pairing: open a commissioning window, show its QR code, manual code and
     countdown, close it, and report the controller that paired. Ported from
     openccu-loom assets/ui/src/routes/matter/MatterPair.svelte:1-279 at
     b40b289a; the host-readiness gate becomes the node's own "no endpoint
     exposed yet" hint, and the colours are the theme tokens. -->
<script lang="ts">
  import { onMount } from "svelte";
  import { useFabricUI } from "../../context";
  import { errorText } from "../../store.svelte";
  import { renderQrSvg } from "../../qr";
  import Button from "../../kit/Button.svelte";
  import Card from "../../kit/Card.svelte";
  import Select from "../../kit/Select.svelte";

  const { t, toast, node } = useFabricUI();

  let hydrateLoading = $state(true);
  let hydrateError = $state<string | null>(null);
  let opening = $state(false);
  let closing = $state(false);
  let selectedDuration = $state(300);

  // A refresh, a tab switch or a window opened elsewhere shows the codes
  // and the countdown instead of the empty form.
  onMount(async () => {
    try {
      await node.hydrateCommissioning();
    } catch (err) {
      hydrateError = errorText(err);
    } finally {
      hydrateLoading = false;
    }
  });

  // $derived, not a const: the labels follow a live locale switch.
  const durationOptions = $derived([
    { label: `5 ${t("node.pair.minutes")}`, value: "300" },
    { label: `10 ${t("node.pair.minutes")}`, value: "600" },
    { label: `15 ${t("node.pair.minutes")}`, value: "900" },
  ]);

  const phase = $derived(node.commissioning.phase);
  const win = $derived(node.commissioning.window);
  const remaining = $derived(node.commissioning.remaining);
  const addedLabel = $derived(node.commissioning.addedFabricLabel);

  // Opening a window on a node controllers already hold adds a controller;
  // the card says which of the two the operator is doing.
  const fabricCount = $derived(node.status?.fabric_count ?? 0);
  const alreadyPaired = $derived(fabricCount > 0);
  const noEndpoint = $derived(node.status !== null && node.status.enabled && node.status.endpoint_count === 0);

  // The clipboard API is unavailable on a page served over plain HTTP (an
  // Ingress, a bare LAN address), so a refusal has to surface.
  async function copy(value: string) {
    try {
      await navigator.clipboard.writeText(value);
      toast.success(t("node.pair.copied"));
    } catch {
      toast.error(t("node.pair.copy_failed"));
    }
  }

  async function openWindow() {
    opening = true;
    try {
      await node.openWindow(selectedDuration);
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      opening = false;
    }
  }

  async function closeWindow() {
    closing = true;
    try {
      await node.closeWindow();
    } catch (err) {
      toast.error(errorText(err));
    } finally {
      closing = false;
    }
  }

  const RING_R = 54;
  const RING_C = 2 * Math.PI * RING_R;
  const ringOffset = $derived(!remaining || !win || !win.duration_seconds ? 0 : RING_C * (1 - remaining / win.duration_seconds));
  const qrSvg = $derived(win?.qr_code ? renderQrSvg(win.qr_code) : null);
</script>

<div class="max-w-lg">
  {#if hydrateLoading}
    <p class="text-sm text-[var(--ha-secondary-text-color)]">{t("node.pair.loading")}</p>
  {:else if hydrateError}
    <p class="text-sm text-[var(--ha-error-color)]">{t("node.pair.load_error")}</p>
  {:else if phase === "idle"}
    <Card class="p-6">
      <h2 class="mb-4 text-base font-semibold text-[var(--ha-primary-text-color)]">{t("node.pair.window_open_duration")}</h2>
      {#if alreadyPaired}
        <p class="mb-4 text-sm text-[var(--ha-secondary-text-color)]">{t("node.pair.already_paired", { count: fabricCount })}</p>
      {/if}
      {#if noEndpoint}
        <p class="mb-4 text-sm font-medium text-[var(--ha-secondary-text-color)]">{t("node.pair.waiting")}</p>
      {/if}
      <div class="mb-4 flex flex-wrap items-center gap-3">
        <Select
          class="w-auto"
          value={String(selectedDuration)}
          onValueChange={(v) => (selectedDuration = Number(v))}
          ariaLabel={t("node.pair.window_open_duration")}
          options={durationOptions}
        />
        <Button class="w-full sm:w-auto" disabled={opening} onclick={() => void openWindow()}>
          {opening ? t("common.saving") : alreadyPaired ? t("node.pair.add_controller") : t("node.pair.window_open")}
        </Button>
      </div>
      <p class="text-xs text-[var(--ha-secondary-text-color)]">{t("node.pair.qr_caption")}</p>
    </Card>
  {:else if phase === "open" && win}
    <Card class="p-6">
      <h2 class="mb-4 text-base font-semibold text-[var(--ha-primary-text-color)]">{t("node.pair.qr_caption")}</h2>
      <div class="flex flex-col items-center gap-4">
        {#if remaining !== null}
          <div class="relative flex items-center justify-center">
            <svg width="120" height="120" viewBox="0 0 120 120" aria-hidden="true">
              <circle cx="60" cy="60" r={RING_R} fill="none" stroke="var(--ha-divider-color)" stroke-width="8" />
              <circle
                cx="60"
                cy="60"
                r={RING_R}
                fill="none"
                stroke="var(--ha-primary-color)"
                stroke-width="8"
                stroke-dasharray={RING_C}
                stroke-dashoffset={ringOffset}
                stroke-linecap="round"
                transform="rotate(-90 60 60)"
                style="transition: stroke-dashoffset 1s linear;"
              />
            </svg>
            <span class="absolute text-xl font-semibold text-[var(--ha-primary-text-color)]">{remaining}s</span>
          </div>
        {/if}
        {#if qrSvg}
          <div class="max-w-full rounded border border-[var(--ha-divider-color)] p-2" aria-label={t("node.pair.qr_caption")} data-testid="pair-qr">
            <!-- The SVG is built locally from the payload; nothing user-supplied is injected. -->
            <div class="h-auto max-w-full">{@html qrSvg}</div>
          </div>
        {/if}
        <div class="text-center">
          <p class="mb-1 text-xs text-[var(--ha-secondary-text-color)]">{t("node.pair.manual_code")}</p>
          <div class="flex items-center justify-center gap-2">
            <p class="overflow-x-auto whitespace-nowrap font-mono text-lg font-semibold tracking-widest text-[var(--ha-primary-text-color)]">
              {win.manual_code}
            </p>
            <Button variant="ghost" size="sm" aria-label={t("node.pair.copy_manual_code")} onclick={() => void copy(win.manual_code)}>
              {t("common.copy")}
            </Button>
          </div>
        </div>
        <details class="w-full">
          <summary class="cursor-pointer text-xs text-[var(--ha-secondary-text-color)]">{t("node.pair.qr_payload")}</summary>
          <div class="mt-1 flex items-start gap-2">
            <p class="break-all font-mono text-xs text-[var(--ha-secondary-text-color)]">{win.qr_code}</p>
            <Button variant="ghost" size="sm" aria-label={t("node.pair.copy_qr_payload")} onclick={() => void copy(win.qr_code)}>
              {t("common.copy")}
            </Button>
          </div>
        </details>
        <Button variant="outline" size="sm" disabled={closing} onclick={() => void closeWindow()}>
          {closing ? t("common.saving") : t("node.pair.close_window")}
        </Button>
      </div>
    </Card>
  {:else if phase === "success"}
    <Card class="p-6 text-center">
      <div class="mb-3 text-4xl text-[var(--ha-success-color)]" aria-hidden="true">✓</div>
      <h2 class="mb-2 text-base font-semibold text-[var(--ha-primary-text-color)]">{t("node.pair.success")}</h2>
      {#if addedLabel}
        <p class="mb-4 text-sm text-[var(--ha-secondary-text-color)]">{addedLabel}</p>
      {/if}
      <Button onclick={() => node.resetCommissioning()}>{t("common.close")}</Button>
    </Card>
  {/if}
</div>
