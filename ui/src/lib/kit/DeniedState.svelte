<!-- A notice in place of a page or a section the server refused with 403
     (0.10.0): no admin rights (forbidden), or the server refuses the
     address or the origin the page was loaded from. No retry: the same
     request is refused again. -->
<script lang="ts">
  import { useT } from "../context";
  import EmptyState from "./EmptyState.svelte";

  type Props = { code?: string | null; class?: string };
  let { code, class: className }: Props = $props();
  const t = useT();

  const key = $derived(code === "host_not_allowed" || code === "origin_not_allowed" ? code : "forbidden");
</script>

<div data-testid="denied" data-code={key}>
  <EmptyState icon="mdi:cancel" class={className} message={t(`common.denied.${key}`)} description={t(`common.denied.${key}_hint`)} />
</div>
