<!-- Ported from openccu-loom assets/ui/src/lib/components/ui//ErrorState.svelte:1-36 at b40b289a. -->
<script lang="ts">
  import { useT } from "../context";
  import { cn } from "./utils";
  import Card from "./Card.svelte";
  import Button from "./Button.svelte";
  import Icon from "./Icon.svelte";

  type Props = { message: string; onRetry?: () => void; class?: string };
  let { message, onRetry, class: className }: Props = $props();
  const t = useT();
</script>

<Card class={cn("flex items-start gap-3 p-3", className)}>
  <Icon name="mdi:alert-circle" size={18} class="mt-0.5 shrink-0 text-[var(--ha-error-color)]" aria-label="" />
  <div class="min-w-0 flex-1">
    <p class="text-sm text-[var(--ha-error-color)]">{t("common.error")} {message}</p>
    {#if onRetry}
      <Button variant="outline" size="sm" class="mt-2" onclick={onRetry}>{t("common.reload")}</Button>
    {/if}
  </div>
</Card>
