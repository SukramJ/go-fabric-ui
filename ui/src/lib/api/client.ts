// The client of the node and controller routes, typed from contract/openapi.yaml
// (schema.gen.ts). It is the TypeScript half of the contract: the tests in
// client.test.ts read the same golden fixtures as the Go handler tests.
// Mirrors the matter* functions of openccu-loom
// assets/ui/src/lib/api/client.ts:1806-1875 at b40b289a, on the node routes.
import type { components } from "./schema.gen";

type S = components["schemas"];
export type Problem = S["Problem"];
export type Status = S["Status"];
export type Warning = S["Warning"];
export type Fabric = S["Fabric"];
export type FabricList = S["FabricList"];
export type Pairing = S["Pairing"];
export type Window = S["Window"];
export type Endpoint = S["Endpoint"];
export type EndpointList = S["EndpointList"];
export type Cluster = S["Cluster"];
export type Session = S["Session"];
export type Occupancy = S["Occupancy"];
export type SessionList = S["SessionList"];
export type Mdns = S["Mdns"];
export type MdnsService = S["MdnsService"];
export type MdnsFinding = S["MdnsFinding"];
export type Compatibility = S["Compatibility"];
export type Ecosystem = S["Ecosystem"];
export type CompatFinding = S["CompatFinding"];
export type DiagnosticEvent = S["DiagnosticEvent"];
export type DiagnosticEventList = S["DiagnosticEventList"];
export type ExposureSchema = S["ExposureSchema"];
export type Column = S["Column"];
export type Filter = S["Filter"];
export type Item = S["Item"];
export type ItemList = S["ItemList"];
export type Verdict = S["Verdict"];
export type VerdictState = Verdict["state"];
export type Update = S["Update"];
export type Applied = S["Applied"];
export type FabricEvent = S["FabricEvent"];
export type WindowEvent = S["WindowEvent"];
export type ExposureEvent = S["ExposureEvent"];
export type SurfaceChangedEvent = S["SurfaceChangedEvent"];
export type SurfaceReport = S["SurfaceReport"];
export type SurfaceEndpoint = S["SurfaceEndpoint"];
export type SurfaceCluster = S["SurfaceCluster"];
export type SurfaceDecision = S["SurfaceDecision"];
export type SurfaceElement = S["SurfaceElement"];
export type SurfaceChange = S["SurfaceChange"];
export type SurfaceElementChange = S["SurfaceElementChange"];

// The controller family (contract 0.2.0).
export type ControllerStatus = S["ControllerStatus"];
export type ControllerCapabilities = S["ControllerCapabilities"];
export type ConnectionState = S["ConnectionState"];
export type ControllerNodeSummary = S["ControllerNodeSummary"];
export type ControllerNodeList = S["ControllerNodeList"];
export type ControllerNodeDetail = S["ControllerNodeDetail"];
export type NodeRecord = S["NodeRecord"];
export type NodeProperties = S["NodeProperties"];
export type NodeEndpoint = S["NodeEndpoint"];
export type NodeCluster = S["NodeCluster"];
export type NodeDeviceType = S["NodeDeviceType"];
export type NodeWindow = S["NodeWindow"];
export type AttributePath = S["AttributePath"];
export type AttributeValue = S["AttributeValue"];
export type AttributeValueList = S["AttributeValueList"];
export type WriteRequest = S["WriteRequest"];
export type WriteResult = S["WriteResult"];
export type InvokeRequest = S["InvokeRequest"];
export type InvokeResult = S["InvokeResult"];
export type AclEntry = S["AclEntry"];
export type AclTarget = S["AclTarget"];
export type BindingTarget = S["BindingTarget"];
export type EndpointBindings = S["EndpointBindings"];
export type BindingList = S["BindingList"];
export type IcdState = S["IcdState"];
export type NodeFabric = S["NodeFabric"];
export type NodeFabricList = S["NodeFabricList"];
export type CommissionRequest = S["CommissionRequest"];
export type Discovered = S["Discovered"];
export type DiscoveredList = S["DiscoveredList"];
export type Credential = S["Credential"];
export type CredentialList = S["CredentialList"];
export type CredentialRequest = S["CredentialRequest"];
export type PendingUpdate = S["PendingUpdate"];
export type UpdateQueue = S["UpdateQueue"];
export type UpdateConsentRequest = S["UpdateConsentRequest"];
export type AvailableUpdate = S["AvailableUpdate"];
export type AvailableUpdateList = S["AvailableUpdateList"];
export type ControllerEvent = S["ControllerEvent"];
/** What GET /controller/discovery selects by. */
export type DiscoverFilter = { filter?: "any" | "long_discriminator" | "short_discriminator" | "vendor" | "device_type"; value?: number; timeout?: number };

/** The header the admin token travels in (token mode). */
export const TOKEN_HEADER = "X-Fabric-UI-Token";

/** The confirmation POST /node/maintenance/factory-reset requires. */
export const FACTORY_RESET_CONFIRMATION = "remove-all-fabrics";

/** A failed call: the status and, when the server sent one, its problem. */
export class ApiError extends Error {
  readonly status: number;
  readonly problem?: Problem;
  constructor(status: number, message: string, problem?: Problem) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.problem = problem;
  }
}

export interface FabricClient {
  /** The base URL the routes are relative to (the mount prefix). */
  readonly baseUrl: string;
  /** Sets the admin token sent on admin routes; status() sets it from csrf_token. */
  setToken(token: string | undefined): void;
  status(): Promise<Status>;
  fabrics(): Promise<FabricList>;
  removeFabric(fabricIndex: number): Promise<void>;
  pairing(): Promise<Pairing>;
  openWindow(durationSeconds: number): Promise<Window>;
  closeWindow(): Promise<void>;
  endpoints(): Promise<EndpointList>;
  /** The node's surface report, in go-fabric's encoding of surface.Report. */
  surface(): Promise<SurfaceReport>;
  sessions(): Promise<SessionList>;
  mdns(): Promise<Mdns>;
  compatibility(): Promise<Compatibility>;
  events(): Promise<DiagnosticEventList>;
  exposureSchema(): Promise<ExposureSchema>;
  exposure(): Promise<ItemList>;
  /** Applies the updates; 0 for an empty batch (204). */
  applyExposure(items: Update[]): Promise<number>;
  previewExposure(items: Update[]): Promise<ItemList>;
  reassemble(): Promise<void>;
  factoryReset(): Promise<void>;
  /** The URL of the event stream. */
  streamUrl(): string;

  // The controller family. Every node id is the 16-digit hex of node_id_hex.
  /** The controller's status; it sets the admin token from csrf_token, as status() does. 404 when the host serves no controller. */
  controllerStatus(): Promise<ControllerStatus>;
  controllerNodes(): Promise<ControllerNodeList>;
  controllerNode(id: string): Promise<ControllerNodeDetail>;
  decommissionNode(id: string): Promise<void>;
  interviewNode(id: string): Promise<void>;
  enableNode(id: string): Promise<void>;
  disableNode(id: string): Promise<void>;
  setNodeAutoSubscribe(id: string, enabled: boolean): Promise<void>;
  openNodeWindow(id: string, timeoutSeconds?: number): Promise<NodeWindow>;
  closeNodeWindow(id: string): Promise<void>;
  readAttributes(id: string, paths: AttributePath[], fabricFiltered?: boolean): Promise<AttributeValueList>;
  writeAttribute(id: string, req: WriteRequest): Promise<WriteResult>;
  invokeCommand(id: string, req: InvokeRequest): Promise<InvokeResult>;
  nodeAcl(id: string): Promise<{ entries: AclEntry[] }>;
  setNodeAcl(id: string, entries: AclEntry[]): Promise<WriteResult>;
  nodeBindings(id: string): Promise<BindingList>;
  setNodeBindings(id: string, endpoint: number, bindings: BindingTarget[]): Promise<WriteResult>;
  nodeIcd(id: string): Promise<IcdState>;
  registerNodeIcd(id: string, allowMultiAdmin?: boolean): Promise<IcdState>;
  unregisterNodeIcd(id: string, force?: boolean): Promise<IcdState>;
  nodeFabrics(id: string): Promise<NodeFabricList>;
  removeNodeFabric(id: string, fabricIndex: number): Promise<void>;
  commission(req: CommissionRequest): Promise<ControllerNodeDetail>;
  discover(filter?: DiscoverFilter): Promise<DiscoveredList>;
  credentials(): Promise<CredentialList>;
  setCredential(req: CredentialRequest): Promise<void>;
  removeCredential(kind: Credential["kind"], id: string): Promise<void>;
  setFabricLabel(label: string): Promise<void>;
  updates(): Promise<UpdateQueue>;
  consentUpdate(req: UpdateConsentRequest): Promise<void>;
  checkUpdates(): Promise<AvailableUpdateList>;
  /** The URL of the controller's event stream. */
  controllerStreamUrl(): string;
}

type Fetch = (input: string, init?: RequestInit) => Promise<Response>;

/**
 * A client of the node routes under `baseUrl` (the mount prefix, e.g.
 * "./api/v1" for the bundle the Go module serves). `fetchImpl` is the
 * host's fetch — its authentication, its base path.
 */
export function createClient(baseUrl: string, fetchImpl: Fetch = (i, init) => fetch(i, init)): FabricClient {
  const base = baseUrl.replace(/\/+$/, "");
  let token: string | undefined;

  async function call<T>(method: string, path: string, opts: { body?: unknown; admin?: boolean } = {}): Promise<T> {
    const headers: Record<string, string> = { Accept: "application/json" };
    if (opts.body !== undefined) headers["Content-Type"] = "application/json";
    if (opts.admin && token) headers[TOKEN_HEADER] = token;
    const res = await fetchImpl(base + path, {
      method,
      headers,
      body: opts.body === undefined ? undefined : JSON.stringify(opts.body),
      credentials: "same-origin",
    });
    if (res.status === 204) return undefined as T;
    const text = await res.text();
    const data: unknown = text ? JSON.parse(text) : undefined;
    if (!res.ok) {
      const p = isProblem(data) ? data : undefined;
      throw new ApiError(res.status, p ? p.detail || p.title : `${method} ${path}: ${res.status}`, p);
    }
    return data as T;
  }

  return {
    baseUrl: base,
    setToken(t) {
      token = t || undefined;
    },
    async status() {
      const s = await call<Status>("GET", "/node/status");
      if (s.csrf_token) token = s.csrf_token;
      return s;
    },
    fabrics: () => call("GET", "/node/fabrics"),
    removeFabric: (i) => call("DELETE", `/node/fabrics/${encodeURIComponent(String(i))}`, { admin: true }),
    pairing: () => call("GET", "/node/pairing", { admin: true }),
    openWindow: (d) => call("POST", "/node/pairing/window", { admin: true, body: { duration_seconds: d } }),
    closeWindow: () => call("POST", "/node/pairing/window/close", { admin: true }),
    endpoints: () => call("GET", "/node/endpoints"),
    surface: () => call("GET", "/node/surface"),
    sessions: () => call("GET", "/node/sessions"),
    mdns: () => call("GET", "/node/mdns"),
    compatibility: () => call("GET", "/node/compatibility"),
    events: () => call("GET", "/node/events"),
    exposureSchema: () => call("GET", "/node/exposure/schema"),
    exposure: () => call("GET", "/node/exposure"),
    async applyExposure(items) {
      const res = await call<Applied | undefined>("PUT", "/node/exposure", { admin: true, body: { items } });
      return res?.applied ?? 0;
    },
    previewExposure: (items) => call("POST", "/node/exposure/preview", { admin: true, body: { items } }),
    reassemble: () => call("POST", "/node/maintenance/reassemble", { admin: true }),
    factoryReset: () =>
      call("POST", "/node/maintenance/factory-reset", { admin: true, body: { confirm: FACTORY_RESET_CONFIRMATION } }),
    streamUrl: () => base + "/node/stream",

    async controllerStatus() {
      const s = await call<ControllerStatus>("GET", "/controller/status");
      if (s.csrf_token) token = s.csrf_token;
      return s;
    },
    controllerNodes: () => call("GET", "/controller/nodes"),
    controllerNode: (id) => call("GET", node(id)),
    decommissionNode: (id) => call("DELETE", node(id), { admin: true }),
    interviewNode: (id) => call("POST", node(id) + "/interview", { admin: true }),
    enableNode: (id) => call("POST", node(id) + "/enable", { admin: true }),
    disableNode: (id) => call("POST", node(id) + "/disable", { admin: true }),
    setNodeAutoSubscribe: (id, enabled) => call("POST", node(id) + "/auto-subscribe", { admin: true, body: { enabled } }),
    openNodeWindow: (id, timeoutSeconds) =>
      call("POST", node(id) + "/window", { admin: true, body: timeoutSeconds === undefined ? {} : { timeout_seconds: timeoutSeconds } }),
    closeNodeWindow: (id) => call("POST", node(id) + "/window/close", { admin: true }),
    readAttributes: (id, paths, fabricFiltered) =>
      call("POST", node(id) + "/attributes", { body: fabricFiltered === undefined ? { paths } : { paths, fabric_filtered: fabricFiltered } }),
    writeAttribute: (id, req) => call("PUT", node(id) + "/attributes", { admin: true, body: req }),
    invokeCommand: (id, req) => call("POST", node(id) + "/commands", { admin: true, body: req }),
    nodeAcl: (id) => call("GET", node(id) + "/acl"),
    setNodeAcl: (id, entries) => call("PUT", node(id) + "/acl", { admin: true, body: { entries } }),
    nodeBindings: (id) => call("GET", node(id) + "/bindings"),
    setNodeBindings: (id, endpoint, bindings) => call("PUT", node(id) + "/bindings", { admin: true, body: { endpoint, bindings } }),
    nodeIcd: (id) => call("GET", node(id) + "/icd"),
    registerNodeIcd: (id, allowMultiAdmin) =>
      call("POST", node(id) + "/icd/register", { admin: true, body: allowMultiAdmin === undefined ? {} : { allow_multi_admin: allowMultiAdmin } }),
    unregisterNodeIcd: (id, force) =>
      call("POST", node(id) + "/icd/unregister", { admin: true, body: force === undefined ? {} : { force } }),
    nodeFabrics: (id) => call("GET", node(id) + "/fabrics"),
    removeNodeFabric: (id, i) => call("DELETE", `${node(id)}/fabrics/${encodeURIComponent(String(i))}`, { admin: true }),
    commission: (req) => call("POST", "/controller/commission", { admin: true, body: req }),
    discover(f = {}) {
      const q = new URLSearchParams();
      if (f.filter) q.set("filter", f.filter);
      if (f.value !== undefined) q.set("value", String(f.value));
      if (f.timeout !== undefined) q.set("timeout", String(f.timeout));
      const qs = q.toString();
      return call("GET", "/controller/discovery" + (qs ? "?" + qs : ""), { admin: true });
    },
    credentials: () => call("GET", "/controller/credentials"),
    setCredential: (req) => call("PUT", "/controller/credentials", { admin: true, body: req }),
    removeCredential: (kind, id) =>
      call("DELETE", `/controller/credentials?${new URLSearchParams({ kind, id }).toString()}`, { admin: true }),
    setFabricLabel: (label) => call("PUT", "/controller/fabric/label", { admin: true, body: { label } }),
    updates: () => call("GET", "/controller/updates"),
    consentUpdate: (req) => call("POST", "/controller/updates/consent", { admin: true, body: req }),
    checkUpdates: () => call("POST", "/controller/updates/check", { admin: true }),
    controllerStreamUrl: () => base + "/controller/stream",
  };
}

/** The path of a controller node. */
function node(id: string): string {
  return `/controller/nodes/${encodeURIComponent(id)}`;
}

function isProblem(v: unknown): v is Problem {
  return typeof v === "object" && v !== null && "code" in v && "title" in v;
}
