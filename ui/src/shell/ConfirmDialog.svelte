<script lang="ts">
  import type { ConfirmOptions } from "../lib/context";
  import Button from "../lib/kit/Button.svelte";

  let {
    current,
    answer,
    cancelLabel,
    confirmLabel,
  }: { current: ConfirmOptions | null; answer: (ok: boolean) => void; cancelLabel: string; confirmLabel: string } = $props();
</script>

{#if current}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" role="presentation">
    <div
      class="w-full max-w-md rounded-[var(--ha-radius-card)] bg-[var(--ha-card-background-color)] p-6 shadow-[var(--ha-elevation-modal)]"
      role="alertdialog"
      aria-modal="true"
      aria-labelledby="fui-confirm-title"
    >
      <h2 id="fui-confirm-title" class="text-base font-semibold text-[var(--ha-primary-text-color)]">{current.title}</h2>
      {#if current.body}
        <p class="mt-2 text-sm text-[var(--ha-secondary-text-color)]">{current.body}</p>
      {/if}
      <div class="mt-6 flex justify-end gap-2">
        <Button variant="outline" onclick={() => answer(false)}>{cancelLabel}</Button>
        <Button variant={current.destructive ? "destructive" : "default"} onclick={() => answer(true)}>
          {current.confirmLabel ?? confirmLabel}
        </Button>
      </div>
    </div>
  </div>
{/if}
