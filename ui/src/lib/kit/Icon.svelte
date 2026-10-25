<!-- Ported from openccu-loom assets/ui/src/lib/components/ui//Icon.svelte:1-35 at b40b289a. -->
<script lang="ts">
  import { resolveIcon, type IconName } from "./icons";

  type Props = { name: IconName; size?: number; class?: string; "aria-label"?: string; title?: string };

  let { name, size = 18, class: className = "", "aria-label": ariaLabel, title }: Props = $props();

  const Cmp = $derived(resolveIcon(name));
</script>

<Cmp {size} class={className} aria-label={ariaLabel} aria-hidden={ariaLabel ? undefined : "true"} {title} />
