<!-- Ported from openccu-loom assets/ui/src/lib/components/ui//Select.svelte:1-80 at b40b289a, on a native <select>
     instead of bits-ui: the same props and look, the platform's own popup
     and keyboard handling, one dependency less. -->
<script lang="ts">
  import { useT } from "../context";
  import { cn } from "./utils";

  type Option = { value: string; label: string };
  type Props = {
    options: Option[];
    value?: string;
    placeholder?: string;
    disabled?: boolean;
    onValueChange?: (value: string) => void;
    ariaLabel?: string;
    class?: string;
  };

  const t = useT();
  let {
    options,
    value = $bindable(""),
    placeholder = t("select.placeholder"),
    disabled = false,
    onValueChange,
    ariaLabel,
    class: className,
  }: Props = $props();
</script>

<select
  bind:value
  {disabled}
  aria-label={ariaLabel}
  onchange={(e) => onValueChange?.((e.currentTarget as HTMLSelectElement).value)}
  class={cn(
    "h-10 w-full rounded-md border border-[var(--ha-divider-color)] bg-[var(--ha-card-background-color)] px-3 py-1 text-base text-[var(--ha-primary-text-color)] shadow-sm sm:text-sm",
    "focus-visible:border-[var(--ha-primary-color)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--ha-primary-color)]",
    "disabled:cursor-not-allowed disabled:opacity-50",
    className,
  )}
>
  {#if !options.some((o) => o.value === value)}
    <option value="" disabled>{placeholder}</option>
  {/if}
  {#each options as option (option.value)}
    <option value={option.value}>{option.label}</option>
  {/each}
</select>
