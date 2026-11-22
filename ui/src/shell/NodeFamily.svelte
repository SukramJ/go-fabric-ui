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
  const OPEN_PAGES = ["status", "pair"];

  // Without admin rights only the status and the pairing are served
  // (0.10.0): the other tabs are not shown.
  const allowed = (p: string) => node.status === null || node.isAdmin || OPEN_PAGES.includes(p);
  const tabs = $derived(
    Object.keys(pages)
      .filter(allowed)
      .filter((p) => p !== "selection" || node.status?.exposure || node.status?.exposure_reason)
      .map((p) => ({ key: p, label: t(`node.tab.${p}`), href: routes[p as keyof typeof routes] })),
  );
  // An endpoint's page sits under the endpoints tab.
  // Object.hasOwn, not in: an inherited name (…/constructor) is no page (UR-P1-ui-112).
  const wanted = $derived(route.page === "endpoint" ? "endpoints" : Object.hasOwn(pages, route.page) ? route.page : "status");
  const page = $derived(allowed(wanted) ? wanted : "status");
  const Current = $derived(pages[page]);
</script>

{#if node.streamDenied}
  <!-- A refused stream ends for good: no event reaches the pages until a status grants the rights again (UR-P4-ui-233). -->
  <p class="mb-3 rounded border border-[var(--ha-warning-color)] p-3 text-sm text-[var(--ha-primary-text-color)]" role="status" data-testid="stream-denied">
    {t("common.stream_denied")}
  </p>
{/if}
<Tabs items={tabs} active={page} />
<div class="mt-4">
  {#if route.endpoint !== undefined && page === "endpoints"}
    {#key route.endpoint}
      <NodeEndpoint endpoint={route.endpoint} />
    {/key}
  {:else}
    <Current />
  {/if}
</div>
