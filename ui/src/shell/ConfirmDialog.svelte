<script lang="ts">
  import type { ConfirmOptions } from "../lib/context";
  import Button from "../lib/kit/Button.svelte";
  import { trapFocus } from "./focus-trap";
  import { GUARD_MS } from "./confirm.svelte";

  let {
    current,
    answer,
    cancelLabel,
    confirmLabel,
  }: { current: ConfirmOptions | null; answer: (ok: boolean) => void; cancelLabel: string; confirmLabel: string } = $props();

  // The confirm button is armed GUARD_MS after a question is shown: the
  // second click of a double click does not answer the next step of a
  // two-step confirmation (UR-P1-ui-104), and an automated click waits.
  let armed = $state(false);
  $effect(() => {
    if (!current) return;
    armed = false;
    const timer = setTimeout(() => (armed = true), GUARD_MS);
    return () => clearTimeout(timer);
  });
</script>

{#if current}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" role="presentation">
    <div
      class="w-full max-w-md rounded-[var(--ha-radius-card)] bg-[var(--ha-card-background-color)] p-6 shadow-[var(--ha-elevation-modal)]"
      role="alertdialog"
      aria-modal="true"
      aria-labelledby="fui-confirm-title"
      tabindex="-1"
      use:trapFocus={() => answer(false)}
    >
      <h2 id="fui-confirm-title" class="text-base font-semibold text-[var(--ha-primary-text-color)]">{current.title}</h2>
      {#if current.body}
        <p class="mt-2 text-sm text-[var(--ha-secondary-text-color)]">{current.body}</p>
      {/if}
      <div class="mt-6 flex justify-end gap-2">
        <!-- The focus starts on Cancel: Enter on a destructive question does not confirm it. -->
        <Button variant="outline" data-autofocus onclick={() => answer(false)}>{cancelLabel}</Button>
        <Button variant={current.destructive ? "destructive" : "default"} disabled={!armed} onclick={() => answer(true)}>
          {current.confirmLabel ?? confirmLabel}
        </Button>
      </div>
    </div>
  </div>
{/if}
