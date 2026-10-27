<!-- Ported from openccu-loom assets/ui/src/lib/components/ui//Tabs.svelte:1-107 at b40b289a: the underline and segmented
     weights; the active marker uses the primary token so the skin applies. -->
<script lang="ts">
  import type { TabItem } from "./tabs";
  import { cn } from "./utils";
  import Badge from "./Badge.svelte";
  import Icon from "./Icon.svelte";

  interface Props {
    items: TabItem[];
    active: string;
    variant?: "underline" | "segmented";
    fill?: boolean;
    ariaLabel?: string;
    onSelect?: (key: string) => void;
    class?: string;
  }

  let { items, active, variant = "underline", fill = false, ariaLabel, onSelect, class: cls = "" }: Props = $props();

  const listClass = $derived(
    variant === "underline"
      ? // The divider is an inset shadow, not a border the active marker
        // overlaps with -mb-px: that 1px overflow made the scrollable list
        // show a vertical scrollbar.
        cn("flex gap-1 overflow-x-auto overflow-y-hidden shadow-[inset_0_-1px_0_var(--ha-divider-color)]", cls)
      : cn("inline-flex flex-wrap gap-0.5 rounded-lg bg-[var(--ha-secondary-background-color)] p-0.5", cls),
  );

  function itemClass(isActive: boolean): string {
    if (variant === "underline") {
      return cn(
        "inline-flex items-center justify-center gap-2 whitespace-nowrap border-b-2 px-4 py-3 text-center text-sm font-medium transition",
        fill && "flex-1",
        isActive
          ? "border-[var(--ha-primary-color)] text-[var(--ha-primary-color)]"
          : "border-transparent text-[var(--ha-secondary-text-color)] hover:text-[var(--ha-primary-text-color)]",
      );
    }
    return cn(
      "rounded-md px-3 py-1.5 text-sm font-medium transition",
      isActive
        ? "bg-[var(--ha-card-background-color)] text-[var(--ha-primary-text-color)] shadow-sm"
        : "text-[var(--ha-secondary-text-color)] hover:text-[var(--ha-primary-text-color)]",
    );
  }
</script>

<div class={listClass} role="tablist" aria-label={ariaLabel}>
  {#each items as item (item.key)}
    {@const isActive = item.key === active}
    {#if item.href}
      <a href={item.href} role="tab" aria-selected={isActive} class={itemClass(isActive)} onclick={() => onSelect?.(item.key)}>
        {#if item.icon}<Icon name={item.icon} size={16} />{/if}
        {item.label}
        {#if item.badge !== undefined}<Badge variant="muted">{item.badge}</Badge>{/if}
      </a>
    {:else}
      <button type="button" role="tab" aria-selected={isActive} class={itemClass(isActive)} onclick={() => onSelect?.(item.key)}>
        {#if item.icon}<Icon name={item.icon} size={16} />{/if}
        {item.label}
        {#if item.badge !== undefined}<Badge variant="muted">{item.badge}</Badge>{/if}
      </button>
    {/if}
  {/each}
</div>
