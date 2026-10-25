<!-- Ported from openccu-loom assets/ui/src/lib/components/ui//PageHeader.svelte:1-72 at b40b289a: title and subtitle on the
     left, actions on the right, wrapping by the rendered width (works inside
     a narrow frame on a wide viewport). -->
<script lang="ts">
  import type { Snippet } from "svelte";
  import { cn } from "./utils";

  interface Props {
    title: string;
    subtitle?: string;
    actions?: Snippet;
    children?: Snippet;
    class?: string;
  }

  let { title, subtitle, actions, children, class: cls = "" }: Props = $props();
</script>

<header class={cn("mb-4 flex flex-wrap items-start justify-between gap-3", cls)}>
  <div>
    <h1 class="text-balance text-2xl font-bold tracking-tight text-[var(--ha-primary-text-color)]">{title}</h1>
    {#if subtitle}
      <p class="text-sm text-[var(--ha-secondary-text-color)]">{subtitle}</p>
    {/if}
    {#if children}
      {@render children()}
    {/if}
  </div>
  {#if actions}
    <div class="flex flex-wrap items-center gap-2">{@render actions()}</div>
  {/if}
</header>
