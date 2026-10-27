<!-- Renders the controller's node page under a fabric-ui context with its props. -->
<script lang="ts">
  import { provideFabricUI, type FabricUI } from "../lib/context";
  import Node from "../lib/pages/controller/Node.svelte";

  let { ui, nodeId, tab }: { ui: FabricUI; nodeId: string; tab: string } = $props();
  // svelte-ignore state_referenced_locally
  provideFabricUI(ui);
</script>

<Node {nodeId} {tab} />
