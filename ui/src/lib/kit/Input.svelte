<!-- Ported from openccu-loom assets/ui/src/lib/components/ui//Input.svelte:1-24 at b40b289a. -->
<script lang="ts">
  import type { HTMLInputAttributes } from "svelte/elements";
  import { cn } from "./utils";

  type Props = HTMLInputAttributes & { value?: string | number | null };

  let { class: className, value = $bindable(), ...rest }: Props = $props();
</script>

<input
  bind:value
  class={cn(
    // 16px on phones suppresses iOS Safari's focus zoom; h-10 is a 40px touch row.
    "h-10 w-full rounded-md border border-[var(--ha-divider-color)] bg-[var(--ha-card-background-color)] px-3 py-1 text-base text-[var(--ha-primary-text-color)] shadow-sm transition sm:text-sm",
    "placeholder:text-[var(--ha-secondary-text-color)]",
    "focus-visible:border-[var(--ha-primary-color)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--ha-primary-color)]",
    "disabled:cursor-not-allowed disabled:opacity-50",
    className,
  )}
  {...rest}
/>
