<!-- Ported from openccu-loom assets/ui/src/lib/components/ui//LoadingState.svelte:1-30 at b40b289a. -->
<script lang="ts">
  import { useT } from "../context";
  import { cn } from "./utils";
  import Spinner from "./Spinner.svelte";

  type Props = { label?: string; message?: string; class?: string };
  let { label, message, class: className }: Props = $props();
  const t = useT();
</script>

<div class={cn("flex items-center justify-center gap-2 py-8 text-sm text-[var(--ha-secondary-text-color)]", className)} role="status" aria-live="polite">
  <Spinner size={18} />
  <span>{label ?? message ?? t("common.loading")}</span>
</div>
