// The page families. Each page reads the fabric-ui context
// (provideFabricUI) and renders nothing host-specific.
export { default as NodeStatus } from "./node/Status.svelte";
export { default as NodePair } from "./node/Pair.svelte";
export { default as NodeFabrics } from "./node/Fabrics.svelte";
export { default as NodeSelection } from "./node/Selection.svelte";
export { default as NodeEndpoints } from "./node/Endpoints.svelte";
// NodeEndpointDetail is one endpoint's page; it takes the endpoint number
// as `endpoint` (brief U4a).
export { default as NodeEndpointDetail } from "./node/Endpoint.svelte";
export { default as NodeSurface } from "./node/Surface.svelte";
export { default as NodeDiagnostics } from "./node/Diagnostics.svelte";
export { default as NodeTopology } from "./node/Topology.svelte";
// The controller family (wave U2). ControllerNode takes the node's
// 16-digit hex id as `nodeId` and, optionally, the first `tab`;
// ControllerNodes takes the `tab` list or network, and ControllerNetwork
// is the network tab alone (brief U3b).
export { default as ControllerNodes } from "./controller/Nodes.svelte";
export { default as ControllerNetwork } from "./controller/Network.svelte";
export { default as ControllerCommission } from "./controller/Commission.svelte";
export { default as ControllerNode } from "./controller/Node.svelte";
export { default as ControllerSettings } from "./controller/Settings.svelte";
// The daemon's operation, beside both families (wave U5a): log level, log
// view and version.
export { default as Operation } from "./operation/Operation.svelte";
// The pairing hints per ecosystem, below NodePair's flow (wave U5a).
export { default as NodeEcosystemHints } from "./node/EcosystemHints.svelte";
