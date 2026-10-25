<!-- Ported from openccu-loom assets/ui/src/lib/components/ui//Badge.svelte:1-43 at b40b289a. -->
<script lang="ts">
  import type { Snippet } from "svelte";
  import { cn } from "./utils";

  type Variant = "default" | "success" | "warning" | "danger" | "muted";
  // title is a plain HTML passthrough: a badge that carries meaning in its
  // colour needs a text equivalent, and colour alone is not one.
  type Props = { variant?: Variant; class?: string; title?: string; children?: Snippet };

  let { variant = "default", class: className, title, children }: Props = $props();

  // A 15% tint of the matching token as the fill and the token as the text
  // colour, so the badge follows the active skin and flips with .dark.
  const variants: Record<Variant, string> = {
    default: "bg-[color-mix(in_srgb,var(--ha-primary-color)_15%,transparent)] text-[var(--ha-primary-color)]",
    success: "bg-[color-mix(in_srgb,var(--ha-success-color)_15%,transparent)] text-[var(--ha-success-color)]",
    warning: "bg-[color-mix(in_srgb,var(--ha-warning-color)_15%,transparent)] text-[var(--ha-warning-color)]",
    danger: "bg-[color-mix(in_srgb,var(--ha-error-color)_15%,transparent)] text-[var(--ha-error-color)]",
    muted: "bg-[var(--ha-secondary-background-color)] text-[var(--ha-secondary-text-color)]",
  };
</script>

<span class={cn("inline-flex items-center rounded-full px-2 py-0.5 text-[11px] font-medium", variants[variant], className)} {title}>
  {@render children?.()}
</span>
