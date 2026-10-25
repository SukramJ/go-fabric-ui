<!-- Ported from openccu-loom assets/ui/src/lib/components/ui//PageShell.svelte:1-44 at b40b289a. -->
<script lang="ts">
  import type { Snippet } from "svelte";
  import { cn } from "./utils";

  type Width = "default" | "wide" | "narrow";
  interface Props { width?: Width; class?: string; children: Snippet }

  let { width = "default", class: cls = "", children }: Props = $props();

  const widths: Record<Width, string> = { default: "max-w-6xl", wide: "", narrow: "max-w-4xl" };
</script>

<section class={cn("mx-auto w-full px-4 py-6 sm:px-6", widths[width], cls)}>
  {@render children()}
</section>
