<!-- The node family's context for one node the host lists (contract
     0.5.0): without a key the shell's own context (the default node, the
     /node/* routes); with one, the same context on a client of
     /nodes/<key>/…, a store of its own with its stream, and the routes
     under #/node/<key>/…. The shell keys it by the node, so a switch
     builds a fresh store and closes the old stream. The node list
     (ui.nodes, 0.6.0) stays the parent's. -->
<script lang="ts">
  import { onDestroy, onMount, type Snippet } from "svelte";
  import { provideFabricUI, routesForNode, useFabricUI, type FabricUI } from "../lib/context";
  import { createNodeStore } from "../lib/store.svelte";

  let { nodeKey, children }: { nodeKey?: string; children: Snippet } = $props();

  const parent = useFabricUI();
  // The key is fixed for this instance: the shell recreates it on a switch.
  // svelte-ignore state_referenced_locally
  const key = nodeKey;
  let scoped: FabricUI | null = null;
  if (key) {
    const client = parent.client.forNode(key);
    scoped = provideFabricUI({
      ...parent,
      client,
      node: createNodeStore(client, { t: parent.t, toast: parent.toast, onNodesChanged: () => void parent.nodes.load() }),
      routes: routesForNode(parent.routes, key),
    });
  }

  onMount(() => {
    if (!scoped) return;
    void scoped.node.loadStatus();
    scoped.node.connect();
  });
  onDestroy(() => scoped?.node.close());
</script>

{@render children()}
