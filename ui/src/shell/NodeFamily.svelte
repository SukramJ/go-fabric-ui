<!-- The node family's tabs and current page, on the node of the context
     around it (NodeScope): the selection tab shows when the node serves a
     selection or the host says why it serves none. -->
<script lang="ts">
  import type { Component } from "svelte";
  import { useFabricUI } from "../lib/context";
  import Tabs from "../lib/kit/Tabs.svelte";
  import NodeEndpoint from "../lib/pages/node/Endpoint.svelte";
  import type { Route } from "./route";

  type Pages = Record<string, Component<any>>;
  let { route, pages }: { route: Route; pages: Pages } = $props();

  const { t, node, routes } = useFabricUI();

  const tabs = $derived(
    Object.keys(pages)
      .filter((p) => p !== "selection" || node.status?.exposure || node.status?.exposure_reason)
      .map((p) => ({ key: p, label: t(`node.tab.${p}`), href: routes[p as keyof typeof routes] })),
  );
  // An endpoint's page sits under the endpoints tab.
  const page = $derived(route.page === "endpoint" ? "endpoints" : route.page in pages ? route.page : "status");
  const Current = $derived(pages[page]);
</script>

<Tabs items={tabs} active={page} />
<div class="mt-4">
  {#if route.endpoint !== undefined}
    {#key route.endpoint}
      <NodeEndpoint endpoint={route.endpoint} />
    {/key}
  {:else}
    <Current />
  {/if}
</div>
