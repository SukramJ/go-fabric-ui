<!-- A Descriptor's TagList as chips: the tag's label where it has one,
     else namespace:tag; the numbers and a manufacturer code in the title.
     Written for both families (the controller's inventory, the node's
     endpoint page). -->
<script lang="ts">
  import { cn } from "./utils";
  import type { SemanticTag } from "../api/client";
  import Badge from "./Badge.svelte";

  // empty is the text an endpoint without tags shows; nothing when absent.
  type Props = { tags: SemanticTag[]; empty?: string; class?: string };
  let { tags, empty, class: className }: Props = $props();

  const hex = (n: number) => "0x" + n.toString(16).toUpperCase().padStart(4, "0");
  const title = (tag: SemanticTag) =>
    `namespace ${tag.namespace} · tag ${tag.tag}` + (tag.mfg_code !== undefined ? ` · mfg ${hex(tag.mfg_code)}` : "");
</script>

{#if tags.length > 0}
  <ul class={cn("flex flex-wrap gap-1", className)} data-testid="tag-chips">
    {#each tags as tag, i (i)}
      <li><Badge variant="muted" title={title(tag)}>{tag.label || `${tag.namespace}:${tag.tag}`}</Badge></li>
    {/each}
  </ul>
{:else if empty}
  <p class={cn("text-xs text-[var(--ha-secondary-text-color)]", className)}>{empty}</p>
{/if}
