<script lang="ts">
  import type { ToastEntry } from "./toast.svelte";

  let { entries, dismiss, closeLabel }: { entries: ToastEntry[]; dismiss: (id: number) => void; closeLabel: string } = $props();

  const tone = (s: ToastEntry["severity"]) =>
    s === "error" ? "border-[var(--ha-error-color)]" : s === "success" ? "border-[var(--ha-success-color)]" : "border-[var(--ha-info-color)]";
</script>

<div class="pointer-events-none fixed bottom-4 right-4 z-50 flex max-w-sm flex-col gap-2" aria-live="polite">
  {#each entries as e (e.id)}
    <div
      class="pointer-events-auto flex items-start gap-3 rounded-[var(--ha-radius-control)] border-l-4 {tone(e.severity)} bg-[var(--ha-card-background-color)] px-4 py-3 text-sm text-[var(--ha-primary-text-color)] shadow-[var(--ha-elevation-modal)]"
      role={e.severity === "error" ? "alert" : "status"}
    >
      <span class="flex-1">{e.message}</span>
      <button type="button" class="text-[var(--ha-secondary-text-color)]" aria-label={closeLabel} onclick={() => dismiss(e.id)}>✕</button>
    </div>
  {/each}
</div>
