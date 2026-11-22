<!-- Ported from openccu-loom assets/ui/src/lib/components/ui//EmptyState.svelte:1-46 at b40b289a; the description is
     text to read, in the secondary text colour, not the disabled one (decision 77, UR-P4-ui-300). -->
<script lang="ts">
  import type { Snippet } from "svelte";
  import { cn } from "./utils";
  import Icon from "./Icon.svelte";
  import type { IconName } from "./icons";

  type Props = { message: string; description?: string; icon?: IconName; class?: string; action?: Snippet };

  let { message, description, icon = "mdi:text-box-search-outline", class: className, action }: Props = $props();
</script>

<div class={cn("flex flex-col items-center justify-center gap-3 py-12 text-center", className)}>
  <Icon name={icon} size={40} class="text-[var(--ha-disabled-text-color)]" aria-label="" />
  <div class="flex flex-col items-center gap-1">
    <p class="text-sm text-[var(--ha-secondary-text-color)]">{message}</p>
    {#if description}
      <p class="max-w-sm text-balance text-xs text-[var(--ha-secondary-text-color)]">{description}</p>
    {/if}
  </div>
  {#if action}
    {@render action()}
  {/if}
</div>
