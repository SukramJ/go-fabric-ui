<!-- Renders one page under a fabric-ui context, as a host does. -->
<script lang="ts">
  import type { Component } from "svelte";
  import { provideFabricUI, type FabricUI } from "../lib/context";

  let { ui, page }: { ui: FabricUI; page: Component } = $props();
  // The context is provided once, for the page's lifetime.
  // svelte-ignore state_referenced_locally
  provideFabricUI(ui);
  const Page = $derived(page);
</script>

<Page />
