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

  // Arrow keys, Home and End move the focus between the button tabs.
  function onKeydown(e: KeyboardEvent) {
    const tabs = [...(e.currentTarget as HTMLElement).querySelectorAll<HTMLElement>('[role="tab"]')];
    const at = tabs.indexOf(document.activeElement as HTMLElement);
    if (at < 0) return;
    let next = -1;
    if (e.key === "ArrowRight") next = (at + 1) % tabs.length;
    else if (e.key === "ArrowLeft") next = (at - 1 + tabs.length) % tabs.length;
    else if (e.key === "Home") next = 0;
    else if (e.key === "End") next = tabs.length - 1;
    if (next < 0) return;
    e.preventDefault();
    tabs[next].focus();
  }

  function itemClass(isActive: boolean): string {
    if (variant === "underline") {
      return cn(
        "inline-flex items-center justify-center gap-2 whitespace-nowrap border-b-2 px-4 py-3 text-center text-sm font-medium transition",
        fill && "flex-1",
        isActive
          ? "border-[var(--ha-primary-color)] text-[var(--fui-primary-text)]"
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

<!-- Route tabs (every item a link) are a navigation: links that navigate,
     the current one marked aria-current. Button tabs follow the tab
     pattern: one Tab stop (the active tab), the arrow keys, Home and End
     move between them. -->
{#snippet content(item: TabItem)}
  {#if item.icon}<Icon name={item.icon} size={16} />{/if}
  {item.label}
  {#if item.badge !== undefined}<Badge variant="muted">{item.badge}</Badge>{/if}
{/snippet}

{#if items.length > 0 && items.every((i) => i.href)}
  <nav class={listClass} aria-label={ariaLabel}>
    {#each items as item (item.key)}
      {@const isActive = item.key === active}
      <a href={item.href} aria-current={isActive ? "page" : undefined} class={itemClass(isActive)} onclick={() => onSelect?.(item.key)}>
        {@render content(item)}
      </a>
    {/each}
  </nav>
{:else}
  <div class={listClass} role="tablist" aria-label={ariaLabel} tabindex="-1" onkeydown={onKeydown}>
    {#each items as item (item.key)}
      {@const isActive = item.key === active}
      <button
        type="button"
        role="tab"
        aria-selected={isActive}
        tabindex={isActive || (!items.some((i) => i.key === active) && item === items[0]) ? 0 : -1}
        class={itemClass(isActive)}
        onclick={() => onSelect?.(item.key)}
      >
        {@render content(item)}
      </button>
    {/each}
  </div>
{/if}
