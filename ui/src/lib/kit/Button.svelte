<!-- Ported from openccu-loom assets/ui/src/lib/components/ui//Button.svelte:1-72 at b40b289a; the disabled fills use the
     tokens instead of the slate ramp, and the label on the primary fill uses --ha-on-primary-color (Home Assistant's
     --text-primary-color under the HA skin). -->
<script lang="ts">
  import type { Snippet } from "svelte";
  import type { HTMLButtonAttributes } from "svelte/elements";
  import { cn } from "./utils";

  type Variant = "default" | "outline" | "ghost" | "destructive" | "outline-destructive";
  type Size = "sm" | "md" | "lg" | "icon";

  type Props = HTMLButtonAttributes & { variant?: Variant; size?: Size; children?: Snippet };

  let { variant = "default", size = "md", class: className, children, ...rest }: Props = $props();

  // A filled variant greys out explicitly when disabled: at 50% opacity a
  // saturated fill still reads as active on a dark card.
  const disabledFill =
    "disabled:bg-[var(--ha-secondary-background-color)] disabled:text-[var(--ha-disabled-text-color)] disabled:shadow-none";
  const disabledMuted = "disabled:opacity-50 disabled:text-[var(--ha-disabled-text-color)]";
  const variants: Record<Variant, string> = {
    default: `bg-[var(--ha-primary-color)] text-[var(--ha-on-primary-color)] hover:bg-[var(--ha-primary-color-hover)] shadow-sm ${disabledFill}`,
    outline: `border border-[var(--ha-divider-color)] bg-[var(--ha-card-background-color)] text-[var(--ha-primary-text-color)] hover:bg-[var(--ha-secondary-background-color)] ${disabledMuted}`,
    ghost: `text-[var(--ha-primary-text-color)] hover:bg-[var(--ha-secondary-background-color)] ${disabledMuted}`,
    destructive: `bg-[var(--ha-error-color)] text-white hover:brightness-95 shadow-sm ${disabledFill}`,
    // Outline weight, red border and text; it fills red on hover, at the
    // point of commitment.
    "outline-destructive": `border border-[color-mix(in_srgb,var(--ha-error-color)_40%,transparent)] bg-[var(--ha-card-background-color)] text-[var(--ha-error-color)] hover:border-[var(--ha-error-color)] hover:bg-[var(--ha-error-color)] hover:text-white ${disabledMuted}`,
  };
  // Touch-first heights: even sm clears ~36px.
  const sizes: Record<Size, string> = {
    sm: "h-9 px-3 text-sm",
    md: "h-10 px-4 text-sm",
    lg: "h-11 px-5 text-base",
    icon: "h-10 w-10",
  };
</script>

<button
  type="button"
  class={cn(
    "inline-flex items-center justify-center gap-2 rounded-md font-medium transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--ha-primary-color)] disabled:cursor-not-allowed",
    variants[variant],
    sizes[size],
    className,
  )}
  {...rest}
>
  {@render children?.()}
</button>
