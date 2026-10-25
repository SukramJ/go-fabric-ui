<!-- Ported from openccu-loom assets/ui/src/lib/components/ui//Card.svelte:1-18 at b40b289a. -->
<script lang="ts">
  import type { Snippet } from "svelte";
  import type { HTMLAttributes } from "svelte/elements";
  import { cn } from "./utils";

  type Props = HTMLAttributes<HTMLDivElement> & { children?: Snippet };
  let { class: className, children, ...rest }: Props = $props();
</script>

<div
  class={cn("rounded-[var(--ha-radius-card)] border border-[var(--ha-divider-color)] bg-[var(--ha-card-background-color)] shadow-[var(--ha-elevation-card)]", className)}
  {...rest}
>
  {@render children?.()}
</div>
