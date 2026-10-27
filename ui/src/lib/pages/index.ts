// The page families. Each page reads the fabric-ui context
// (provideFabricUI) and renders nothing host-specific.
export { default as NodeStatus } from "./node/Status.svelte";
export { default as NodePair } from "./node/Pair.svelte";
export { default as NodeFabrics } from "./node/Fabrics.svelte";
export { default as NodeSelection } from "./node/Selection.svelte";
export { default as NodeEndpoints } from "./node/Endpoints.svelte";
export { default as NodeDiagnostics } from "./node/Diagnostics.svelte";
// The controller family (wave U2). ControllerNode takes the node's
// 16-digit hex id as `nodeId` and, optionally, the first `tab`.
export { default as ControllerNodes } from "./controller/Nodes.svelte";
export { default as ControllerCommission } from "./controller/Commission.svelte";
export { default as ControllerNode } from "./controller/Node.svelte";
export { default as ControllerSettings } from "./controller/Settings.svelte";
