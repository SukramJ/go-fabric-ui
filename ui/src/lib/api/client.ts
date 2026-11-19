// The client of the node and controller routes, typed from contract/openapi.yaml
// (schema.gen.ts). It is the TypeScript half of the contract: the tests in
// client.test.ts read the same golden fixtures as the Go handler tests.
// Mirrors the matter* functions of openccu-loom
// assets/ui/src/lib/api/client.ts:1806-1875 at b40b289a, on the node routes.
import type { components } from "./schema.gen";
import { parseJson, stringifyJson } from "./json";
import { fetchEventSource, type EventSourceFactory, type EventSourceLike } from "./stream";

type S = components["schemas"];
export type Problem = S["Problem"];
export type Status = S["Status"];
export type Warning = S["Warning"];
export type Counter = S["Counter"];
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
export type TopologyGraph = S["TopologyGraph"];
export type TopologyNode = S["TopologyNode"];
export type TopologyEdge = S["TopologyEdge"];
export type TopologyRef = S["TopologyRef"];
export type TopologyDetail = S["TopologyDetail"];
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
export type SessionEvent = S["SessionEvent"];
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
// The inventory (contract 0.4.0).
export type EndpointDetail = S["EndpointDetail"];
export type EndpointDeviceType = S["EndpointDeviceType"];
export type SemanticTag = S["SemanticTag"];
export type EndpointLabel = S["EndpointLabel"];
export type ConformanceVerdict = S["ConformanceVerdict"];
export type ConformanceViolation = S["ConformanceViolation"];
export type ClusterDetail = S["ClusterDetail"];
export type ClusterFeature = S["ClusterFeature"];
export type ClusterAttribute = S["ClusterAttribute"];
export type ClusterElement = S["ClusterElement"];
// The events of both families and the node's endpoint detail (contract 0.4.1).
export type EventRecord = S["EventRecord"];
export type EventRecordList = S["EventRecordList"];
export type AttributeChangedEvent = S["AttributeChangedEvent"];
/**
 * What an event route selects: the lowest event number (inclusive) and how
 * many of the newest. A string since (an EventRecord's number_str) is sent
 * as since_str, exact beyond 2^53 (0.10.0).
 */
export type EventWindow = { since?: number | string; limit?: number };
export type NodeWindow = S["NodeWindow"];
export type AttributePath = S["AttributePath"];
export type AttributeValue = S["AttributeValue"];
export type AttributeValueList = S["AttributeValueList"];
export type WriteRequest = S["WriteRequest"];
export type WriteResult = S["WriteResult"];
export type InvokeRequest = S["InvokeRequest"];
export type InvokeResult = S["InvokeResult"];
export type AclEntry = S["AclEntry"];
/** An entry as PUT /controller/nodes/{id}/acl takes it (0.10.0): no `unknown`. */
export type AclEntryWrite = S["AclEntryWrite"];
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
// The controller's diagnostics (contract 0.9.0).
export type ControllerDiagnostics = S["ControllerDiagnostics"];
export type ControllerPeer = S["ControllerPeer"];
export type ControllerMdns = S["ControllerMdns"];
// Operation, the node selector, ping and the stream's health and trace events (contract 0.5.0).
export type NodeSummary = S["NodeSummary"];
export type NodeSummaryList = S["NodeSummaryList"];
export type LogLevelName = S["LogLevel"]["level"];
export type LogLevel = S["LogLevel"];
export type LogRecord = S["LogRecord"];
export type LogRecordList = S["LogRecordList"];
export type OperationInfo = S["OperationInfo"];
export type PingResult = S["PingResult"];
export type PingAddress = S["PingAddress"];
export type HealthSampleEvent = S["HealthSampleEvent"];
// The node registry (contract 0.6.0).
export type NodesChangedEvent = S["NodesChangedEvent"];
export type NodeConfig = S["NodeConfig"];
// Completeness against go-fabric v0.18.0 (contract 0.7.0).
export type SessionParameters = S["SessionParameters"];
export type TermsAcknowledgement = S["TermsAcknowledgement"];
export type PairingStep = S["PairingStep"];
// The commissioning's BLE decision (contract 0.7.1).
export type BLEDecision = S["BLEDecision"];
/** What GET /operation/log selects: the lowest sequence number (inclusive) and how many of the newest. */
export type LogWindow = { since?: number; limit?: number };
/** What GET /controller/discovery selects by. */
export type DiscoverFilter = { filter?: "any" | "long_discriminator" | "short_discriminator" | "vendor" | "device_type"; value?: number; timeout?: number };

/** The header the admin token travels in (token mode); sent on every request to a same-origin base. */
export const TOKEN_HEADER = "X-Fabric-UI-Token";

/** The header the client sends on every request, the streams included, with REQUEST_HEADER_VALUE (0.10.0). */
export const REQUEST_HEADER = "X-Fabric-UI";

/** The value of REQUEST_HEADER. */
export const REQUEST_HEADER_VALUE = "1";

/** The 403 problem codes: the request lacks admin rights, or the server refuses its host or origin (0.10.0). */
export const DENIED_CODES = ["forbidden", "host_not_allowed", "origin_not_allowed"] as const;
export type DeniedCode = (typeof DENIED_CODES)[number];

/**
 * The 403 problem code of e, when e is a refusal (a 403 ApiError);
 * "forbidden" for a 403 without a known code. Undefined for anything else.
 */
export function deniedCode(e: unknown): DeniedCode | undefined {
  if (!(e instanceof ApiError) || e.status !== 403) return undefined;
  const c = e.problem?.code;
  return (DENIED_CODES as readonly string[]).includes(c ?? "") ? (c as DeniedCode) : "forbidden";
}

/** The confirmation POST /node/maintenance/factory-reset requires. */
export const FACTORY_RESET_CONFIRMATION = "remove-all-fabrics";

/** The confirmation POST /node/maintenance/erase requires (0.7.0). */
export const ERASE_CONFIRMATION = "erase-node";

/** The contract's NodeKey pattern (a node of GET /nodes). */
export const NODE_KEY_PATTERN = /^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$/;

/** The passcodes a device may not use (CommissionRequest.passcode, 0.10.0): the server answers them 400. */
export const INVALID_PASSCODES: readonly number[] = [
  11111111, 22222222, 33333333, 44444444, 55555555, 66666666, 77777777, 88888888, 12345678, 87654321,
];

/** Whether pin is a passcode the commissioning takes: 1 to 99999998 and none of INVALID_PASSCODES. */
export function isValidPasscode(pin: number): boolean {
  return Number.isInteger(pin) && pin >= 1 && pin <= 99999998 && !INVALID_PASSCODES.includes(pin);
}

/** The longest fabric label in bytes of UTF-8 (FabricLabelRequest.label, 0.10.0). */
export const FABRIC_LABEL_MAX_BYTES = 32;

/** The length of s in bytes of UTF-8. */
export function utf8Length(s: string): number {
  return new TextEncoder().encode(s).length;
}

/** A controller node id as the routes take it (the contract's NodeID): 1 to 16 hex digits, case and leading zeros free. */
export const NODE_ID_PATTERN = /^[0-9A-Fa-f]{1,16}$/;

/** The highest endpoint number a path takes (the contract's EndpointNumber); 65535 is no endpoint. */
export const ENDPOINT_MAX = 65534;

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
  /**
   * Sets the admin token sent on every request to a same-origin base;
   * status() and controllerStatus() set it from csrf_token, and repeat
   * their request once when the answer's token is not the one they sent.
   */
  setToken(token: string | undefined): void;
  /**
   * The event source the client's streams open with (0.10.0): fetch with
   * the client's request headers, as every other request (fetchEventSource);
   * the token only to a URL of the page's origin. A 403 forbidden to an
   * open that carried the token reloads the token and reopens once; a
   * second refusal ends the source as any other. Pass it to
   * subscribeStream, subscribeControllerStream and subscribeLogStream.
   */
  readonly openStream: EventSourceFactory;
  status(): Promise<Status>;
  fabrics(): Promise<FabricList>;
  removeFabric(fabricIndex: number): Promise<void>;
  pairing(): Promise<Pairing>;
  openWindow(durationSeconds: number): Promise<Window>;
  closeWindow(): Promise<void>;
  endpoints(): Promise<EndpointList>;
  /** One endpoint the node serves, with its attribute values, names and surface decisions (the root and the Aggregator too). */
  nodeEndpointDetail(endpoint: number): Promise<EndpointDetail>;
  /** The events the node's event log holds for one endpoint, oldest first. */
  nodeEndpointEvents(endpoint: number, window?: EventWindow): Promise<EventRecordList>;
  /** The node's surface report, in go-fabric's encoding of surface.Report. */
  surface(): Promise<SurfaceReport>;
  sessions(): Promise<SessionList>;
  /** The node as a graph: its bridge, its fabrics and the sessions on them. */
  topology(): Promise<TopologyGraph>;
  mdns(): Promise<Mdns>;
  /** Asks the node to announce its DNS-SD records now (admin; 0.7.0). */
  reannounce(): Promise<void>;
  compatibility(): Promise<Compatibility>;
  events(): Promise<DiagnosticEventList>;
  exposureSchema(): Promise<ExposureSchema>;
  exposure(): Promise<ItemList>;
  /** Applies the updates; 0 for an empty batch (204). */
  applyExposure(items: Update[]): Promise<number>;
  /** Applies the updates and answers what the server did: reassembled false is stored but not in effect (0.10.0); { applied: 0 } for an empty batch. */
  applyExposureResult(items: Update[]): Promise<Applied>;
  previewExposure(items: Update[]): Promise<ItemList>;
  reassemble(): Promise<void>;
  factoryReset(): Promise<void>;
  /** Erases the node — its fabrics, store and endpoint records — and ends it (admin; 0.7.0): 501 when the host wires no erase. */
  erase(): Promise<void>;
  /** The URL of the event stream. */
  streamUrl(): string;
  /**
   * The same client with the node routes on the node `key` lists
   * (/nodes/{key}/…); without a key, on the default node (/node/…). The
   * admin token is shared.
   */
  forNode(key?: string): FabricClient;
  /** The key of the node the node routes address; undefined for the default node. */
  readonly nodeKey: string | undefined;
  /** The device nodes the host serves, the default first (0.5.0). */
  nodes(): Promise<NodeSummaryList>;
  /** Creates a node in the host's node registry (admin; 0.6.0): 409 node_limit when the host serves no further node. */
  createNode(label: string): Promise<NodeSummary>;
  /** Renames a managed node (admin; 0.6.0): 409 node_is_default for the default node. */
  renameNode(key: string, label: string): Promise<void>;
  /** Deletes a managed node with its fabrics, store and selection (admin; 0.6.0): 409 node_is_default for the default node. */
  deleteNode(key: string): Promise<void>;

  // Operation (0.5.0): 503 when the host attached no operations port.
  operationInfo(): Promise<OperationInfo>;
  logLevel(): Promise<LogLevel>;
  setLogLevel(level: LogLevelName): Promise<LogLevel>;
  /** The log ring's records, oldest first. */
  logRecords(window?: LogWindow): Promise<LogRecordList>;
  /** The URL of the log stream (log_record events). */
  logStreamUrl(): string;

  // The controller family. Every node id is the 16-digit hex of node_id_hex.
  /** The controller's status; it sets the admin token from csrf_token, as status() does. 404 when the host serves no controller. */
  controllerStatus(): Promise<ControllerStatus>;
  controllerNodes(): Promise<ControllerNodeList>;
  controllerNode(id: string): Promise<ControllerNodeDetail>;
  /** One endpoint of the node from the controller's image: its structure, no values; nothing is read from the node. */
  controllerEndpointDetail(id: string, endpoint: number): Promise<EndpointDetail>;
  /** One cluster of an endpoint with the values the controller's image holds; nothing is read from the node. */
  controllerClusterDetail(id: string, endpoint: number, clusterHex: string): Promise<ClusterDetail>;
  /** The events the controller received from the node, oldest first; nothing is read from the node. */
  controllerNodeEvents(id: string, window?: EventWindow): Promise<EventRecordList>;
  decommissionNode(id: string): Promise<void>;
  interviewNode(id: string): Promise<void>;
  /** Whether the node answers a Matter read of DataModelRevision, and how fast (admin; 0.5.0). */
  pingNode(id: string): Promise<PingResult>;
  enableNode(id: string): Promise<void>;
  disableNode(id: string): Promise<void>;
  setNodeAutoSubscribe(id: string, enabled: boolean): Promise<void>;
  openNodeWindow(id: string, timeoutSeconds?: number): Promise<NodeWindow>;
  closeNodeWindow(id: string): Promise<void>;
  /** The values as JSON; an integer beyond 2^53 in them is kept as its source text (json.ts), not rounded. */
  readAttributes(id: string, paths: AttributePath[], fabricFiltered?: boolean): Promise<AttributeValueList>;
  /** The value is written as given: a kept integer or a jsonText verbatim; a JS number beyond 2^53 is refused (RangeError). */
  writeAttribute(id: string, req: WriteRequest): Promise<WriteResult>;
  /** The payload as writeAttribute's value; the response's integers as readAttributes'. */
  invokeCommand(id: string, req: InvokeRequest): Promise<InvokeResult>;
  nodeAcl(id: string): Promise<{ entries: AclEntry[] }>;
  setNodeAcl(id: string, entries: AclEntryWrite[]): Promise<WriteResult>;
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
  /** The controller's network as a graph; refresh re-reads the nodes' diagnostics first (admin). 501 when the host serves no graph. */
  controllerTopology(refresh?: boolean): Promise<TopologyGraph>;
  /** The controller's sessions and peers (0.9.0); 501 when the source serves no diagnostics. */
  controllerDiagnostics(): Promise<ControllerDiagnostics>;
  /** What keeps the controller's DNS-SD from working, errors first (0.9.0); 501 as above. */
  controllerMdns(): Promise<ControllerMdns>;
  /** The controller's diagnostic trace, newest first (0.9.0); 501 as above. */
  controllerEvents(): Promise<DiagnosticEventList>;
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
  // One token for every node's client: the handler issues one per process.
  const auth: { token?: string } = {};
  // The token goes only to a base of the page's own origin (UR-P1-web-197).
  const sameOrigin = isSameOrigin(base);

  // The headers of every request, the streams included (UR-P3-ctr-07);
  // a stream's URL is tested on its own (UR-P3-web-194).
  function requestHeaders(url?: string): Record<string, string> {
    const h: Record<string, string> = { [REQUEST_HEADER]: REQUEST_HEADER_VALUE };
    if (auth.token && (url === undefined ? sameOrigin : isSameOrigin(url))) h[TOKEN_HEADER] = auth.token;
    return h;
  }

  // A stream over fetch whose open is refused with 403 forbidden while it
  // carried the token: the token is reloaded once and the stream reopened
  // once (UR-P3-web-191). Any other end reaches the listeners as it is.
  const openStream: EventSourceFactory = (url) => {
    const listeners = new Map<string, ((ev: MessageEvent) => void)[]>();
    let closed = false;
    let retried = false;
    let src: EventSourceLike;
    const dispatch = (type: string, ev: MessageEvent) => {
      for (const fn of listeners.get(type) ?? []) fn(ev);
    };
    const forward = (s: EventSourceLike, type: string) => s.addEventListener(type, (ev) => src === s && dispatch(type, ev));
    const attach = () => {
      const headers = requestHeaders(url);
      const token = headers[TOKEN_HEADER];
      const s = fetchEventSource({ headers: () => headers, fetch: fetchImpl })(url);
      src = s;
      s.addEventListener("error", (ev) => {
        if (src !== s || closed) return;
        if (s.status === 403 && s.problemCode === "forbidden" && token && !retried) {
          retried = true;
          void refreshToken(token).then((changed) => {
            if (closed || src !== s) return;
            if (changed) attach();
            else dispatch("error", ev);
          });
          return;
        }
        dispatch("error", ev);
      });
      for (const type of listeners.keys()) if (type !== "error") forward(s, type);
    };
    attach();
    return {
      addEventListener(type, fn) {
        const l = listeners.get(type);
        if (l) {
          l.push(fn);
          return;
        }
        listeners.set(type, [fn]);
        if (type !== "error") forward(src, type);
      },
      close() {
        closed = true;
        src.close();
      },
      get readyState() {
        return src.readyState;
      },
      get status() {
        return src.status;
      },
      get retryAfter() {
        return src.retryAfter;
      },
      get problemCode() {
        return src.problemCode;
      },
    };
  };

  // A status route: it carries the token, and the request that fetched it
  // without (or with a stale one) was answered as one without admin
  // rights; it is repeated once with the token (UR-P3-web-190).
  async function statusWithToken<T extends { csrf_token?: string }>(path: string): Promise<T> {
    const sent = requestHeaders()[TOKEN_HEADER];
    const s = await call<T>("GET", path);
    if (!s.csrf_token || s.csrf_token === sent) return s;
    auth.token = s.csrf_token;
    return sameOrigin ? call<T>("GET", path) : s;
  }

  // keep names the members whose integers beyond 2^53 are kept exact
  // (json.ts): the generic values of attributes, events and commands.
  type CallOpts = { body?: unknown; keep?: readonly string[]; retried?: boolean };
  async function call<T>(method: string, path: string, opts: CallOpts = {}): Promise<T> {
    // A controller node id outside the contract's NodeID pattern (a dot
    // segment the browser resolves to another route among them) is
    // refused before anything is sent.
    const id = /^\/controller\/nodes\/([^/?]*)/.exec(path)?.[1];
    if (id !== undefined && !NODE_ID_PATTERN.test(id)) throw new RangeError(`not a controller node id: "${id}"`);
    // An endpoint in a path is 0 to ENDPOINT_MAX, as the contract's EndpointNumber.
    const ep = /\/endpoints\/([^/?]*)/.exec(path)?.[1];
    if (ep !== undefined && !(/^\d{1,5}$/.test(ep) && Number(ep) <= ENDPOINT_MAX)) throw new RangeError(`not an endpoint number: "${ep}"`);
    const headers: Record<string, string> = { Accept: "application/json", ...requestHeaders() };
    if (opts.body !== undefined) headers["Content-Type"] = "application/json";
    const token = headers[TOKEN_HEADER];
    // Serialised before the request: a rounded integer is refused unsent.
    const body = opts.body === undefined ? undefined : stringifyJson(opts.body);
    const res = await fetchImpl(base + path, {
      method,
      headers,
      body,
      credentials: "same-origin",
    });
    if (res.status === 204) return undefined as T;
    const text = await res.text();
    // An answer that is not JSON (a proxy's error page) or is empty where
    // a body is due keeps its status: the callers test it (404, 503).
    let data: unknown;
    let parsed = text !== "";
    if (parsed) {
      try {
        data = parseJson(text, opts.keep);
      } catch {
        parsed = false;
      }
    }
    if (!res.ok) {
      const p = parsed && isProblem(data) ? data : undefined;
      // The handler draws a new token per process: after a restart the
      // page holds a stale one. Reload it once and repeat the call once.
      if (res.status === 403 && token && !opts.retried && p?.code === "forbidden" && (await refreshToken(token))) {
        return call<T>(method, path, { ...opts, retried: true });
      }
      throw new ApiError(res.status, p ? p.detail || p.title : `${method} ${path}: ${res.status}`, p);
    }
    if (!parsed) throw new ApiError(res.status, `${method} ${path}: ${res.status}, ${text ? "not JSON" : "no body"}`);
    return data as T;
  }

  // One reload of the token at a time, shared by the calls a restart
  // refused together. It answers whether the token changed.
  let refreshing: Promise<boolean> | null = null;
  function refreshToken(stale: string): Promise<boolean> {
    refreshing ??= (async () => {
      // The token is the process's: the default node's status carries it,
      // or the controller's on a host without the node family.
      for (const p of ["/node/status", "/controller/status"]) {
        try {
          const s = await call<{ csrf_token?: string }>("GET", p);
          if (s.csrf_token) auth.token = s.csrf_token;
          break;
        } catch {
          // the next source
        }
      }
      refreshing = null;
      return auth.token !== undefined && auth.token !== stale;
    })();
    return refreshing;
  }

  const clients = new Map<string, FabricClient>();
  const forNode = (key?: string): FabricClient => {
    const k = key || "";
    if (k && !NODE_KEY_PATTERN.test(k)) throw new RangeError(`not a node key: ${k}`);
    let c = clients.get(k);
    if (!c) {
      c = build(k || undefined);
      clients.set(k, c);
    }
    return c;
  };

  function build(nodeKey: string | undefined): FabricClient {
    // np is the node routes' prefix: the default node's or the keyed node's.
    const np = nodeKey === undefined ? "/node" : `/nodes/${encodeURIComponent(nodeKey)}`;
    const applyExposureResult = async (items: Update[]): Promise<Applied> =>
      (await call<Applied | undefined>("PUT", np + "/exposure", { body: { items } })) ?? { applied: 0 };
    return {
      baseUrl: base,
      nodeKey,
      forNode,
      openStream,
      setToken(t) {
        auth.token = t || undefined;
      },
      status: () => statusWithToken<Status>(np + "/status"),
      fabrics: () => call("GET", np + "/fabrics"),
      removeFabric: (i) => call("DELETE", `${np}/fabrics/${encodeURIComponent(String(i))}`),
      pairing: () => call("GET", np + "/pairing"),
      openWindow: (d) => call("POST", np + "/pairing/window", { body: { duration_seconds: d } }),
      closeWindow: () => call("POST", np + "/pairing/window/close"),
      endpoints: () => call("GET", np + "/endpoints"),
      nodeEndpointDetail: (endpoint) => call("GET", `${np}/endpoints/${endpoint}`, { keep: VALUE_KEYS }),
      nodeEndpointEvents: (endpoint, w) => call("GET", `${np}/endpoints/${endpoint}/events${eventQuery(w)}`, { keep: VALUE_KEYS }),
      surface: () => call("GET", np + "/surface"),
      sessions: () => call("GET", np + "/sessions"),
      topology: () => call("GET", np + "/topology"),
      mdns: () => call("GET", np + "/mdns"),
      reannounce: () => call("POST", np + "/mdns/reannounce"),
      compatibility: () => call("GET", np + "/compatibility"),
      events: () => call("GET", np + "/events"),
      exposureSchema: () => call("GET", np + "/exposure/schema"),
      exposure: () => call("GET", np + "/exposure"),
      applyExposureResult,
      async applyExposure(items) {
        return (await applyExposureResult(items)).applied;
      },
      previewExposure: (items) => call("POST", np + "/exposure/preview", { body: { items } }),
      reassemble: () => call("POST", np + "/maintenance/reassemble"),
      factoryReset: () =>
        call("POST", np + "/maintenance/factory-reset", { body: { confirm: FACTORY_RESET_CONFIRMATION } }),
      erase: () => call("POST", np + "/maintenance/erase", { body: { confirm: ERASE_CONFIRMATION } }),
      streamUrl: () => base + np + "/stream",
      nodes: () => call("GET", "/nodes"),
      createNode: (label) => call("POST", "/nodes", { body: { label } }),
      renameNode: (key, label) => call("PUT", `/nodes/${encodeURIComponent(key)}/label`, { body: { label } }),
      deleteNode: (key) => call("DELETE", `/nodes/${encodeURIComponent(key)}`),
      operationInfo: () => call("GET", "/operation/info"),
      logLevel: () => call("GET", "/operation/log/level"),
      setLogLevel: (level) => call("PUT", "/operation/log/level", { body: { level } }),
      logRecords: (w) => call("GET", `/operation/log${eventQuery(w)}`),
      logStreamUrl: () => base + "/operation/log/stream",

      controllerStatus: () => statusWithToken<ControllerStatus>("/controller/status"),
      controllerNodes: () => call("GET", "/controller/nodes"),
      controllerNode: (id) => call("GET", node(id)),
      controllerEndpointDetail: (id, endpoint) => call("GET", `${node(id)}/endpoints/${endpoint}`, { keep: VALUE_KEYS }),
      controllerClusterDetail: (id, endpoint, clusterHex) =>
        call("GET", `${node(id)}/endpoints/${endpoint}/clusters/${encodeURIComponent(clusterHex)}`, { keep: VALUE_KEYS }),
      controllerNodeEvents: (id, w) => call("GET", `${node(id)}/events${eventQuery(w)}`, { keep: VALUE_KEYS }),
      decommissionNode: (id) => call("DELETE", node(id)),
      interviewNode: (id) => call("POST", node(id) + "/interview"),
      pingNode: (id) => call("POST", node(id) + "/ping"),
      enableNode: (id) => call("POST", node(id) + "/enable"),
      disableNode: (id) => call("POST", node(id) + "/disable"),
      setNodeAutoSubscribe: (id, enabled) => call("POST", node(id) + "/auto-subscribe", { body: { enabled } }),
      openNodeWindow: (id, timeoutSeconds) =>
        call("POST", node(id) + "/window", { body: timeoutSeconds === undefined ? {} : { timeout_seconds: timeoutSeconds } }),
      closeNodeWindow: (id) => call("POST", node(id) + "/window/close"),
      readAttributes: (id, paths, fabricFiltered) =>
        call("POST", node(id) + "/attributes", {
          body: fabricFiltered === undefined ? { paths } : { paths, fabric_filtered: fabricFiltered },
          keep: VALUE_KEYS,
        }),
      writeAttribute: (id, req) => call("PUT", node(id) + "/attributes", { body: req }),
      invokeCommand: (id, req) => call("POST", node(id) + "/commands", { body: req, keep: VALUE_KEYS }),
      nodeAcl: (id) => call("GET", node(id) + "/acl"),
      setNodeAcl: (id, entries) => call("PUT", node(id) + "/acl", { body: { entries } }),
      nodeBindings: (id) => call("GET", node(id) + "/bindings"),
      setNodeBindings: (id, endpoint, bindings) => call("PUT", node(id) + "/bindings", { body: { endpoint, bindings } }),
      nodeIcd: (id) => call("GET", node(id) + "/icd"),
      registerNodeIcd: (id, allowMultiAdmin) =>
        call("POST", node(id) + "/icd/register", { body: allowMultiAdmin === undefined ? {} : { allow_multi_admin: allowMultiAdmin } }),
      unregisterNodeIcd: (id, force) =>
        call("POST", node(id) + "/icd/unregister", { body: force === undefined ? {} : { force } }),
      nodeFabrics: (id) => call("GET", node(id) + "/fabrics"),
      removeNodeFabric: (id, i) => call("DELETE", `${node(id)}/fabrics/${encodeURIComponent(String(i))}`),
      commission: (req) => call("POST", "/controller/commission", { body: req }),
      discover(f = {}) {
        const q = new URLSearchParams();
        if (f.filter) q.set("filter", f.filter);
        if (f.value !== undefined) q.set("value", String(f.value));
        if (f.timeout !== undefined) q.set("timeout", String(f.timeout));
        const qs = q.toString();
        return call("GET", "/controller/discovery" + (qs ? "?" + qs : ""));
      },
      credentials: () => call("GET", "/controller/credentials"),
      setCredential: (req) => call("PUT", "/controller/credentials", { body: req }),
      removeCredential: (kind, id) =>
        call("DELETE", `/controller/credentials?${new URLSearchParams({ kind, id }).toString()}`),
      setFabricLabel: (label) => call("PUT", "/controller/fabric/label", { body: { label } }),
      updates: () => call("GET", "/controller/updates"),
      consentUpdate: (req) => call("POST", "/controller/updates/consent", { body: req }),
      checkUpdates: () => call("POST", "/controller/updates/check"),
      controllerTopology: (refresh) =>
        refresh ? call("GET", "/controller/topology?refresh=true") : call("GET", "/controller/topology"),
      controllerDiagnostics: () => call("GET", "/controller/diagnostics"),
      controllerMdns: () => call("GET", "/controller/mdns"),
      controllerEvents: () => call("GET", "/controller/events"),
      controllerStreamUrl: () => base + "/controller/stream",
    };
  }

  return forNode();
}

/**
 * The members of the contract's generic values (AttributeValue.value,
 * ClusterAttribute.value, EventRecord.data, InvokeResult.response): their
 * integers beyond 2^53 are kept as source text, so a node or fabric id in
 * them stays exact (json.ts).
 */
const VALUE_KEYS = ["value", "data", "response"] as const;

/** The query of an event route; empty without since and limit. */
function eventQuery(w: EventWindow | LogWindow = {}): string {
  const q = new URLSearchParams();
  if (typeof w.since === "string") q.set("since_str", w.since);
  else if (w.since !== undefined) q.set("since", String(w.since));
  if (w.limit !== undefined) q.set("limit", String(w.limit));
  const qs = q.toString();
  return qs ? "?" + qs : "";
}

/**
 * Whether base addresses the page's own origin, decided as fetch resolves
 * it: against the document's base (a host page's <base href> may name
 * another origin, UR-P4-web-04), else the page's URL, the result's origin
 * compared with the page's (UR-P3-web-194). Without a page (no location)
 * against a placeholder origin, so only a base that stays on it — a path —
 * does.
 */
function isSameOrigin(base: string): boolean {
  let here = typeof location !== "undefined" ? location.href : undefined;
  if (!here || here === "about:blank") here = "http://fabric-ui.invalid/";
  let resolveAt = here;
  if (typeof document !== "undefined" && document.querySelector?.("base[href]") && document.baseURI) resolveAt = document.baseURI;
  try {
    return new URL(base, resolveAt).origin === new URL(here).origin;
  } catch {
    return false;
  }
}

/** The path of a controller node. */
function node(id: string): string {
  return `/controller/nodes/${encodeURIComponent(id)}`;
}

/** A problem as the contract has it: code and title strings, detail a string when present. */
function isProblem(v: unknown): v is Problem {
  if (typeof v !== "object" || v === null) return false;
  const p = v as Record<string, unknown>;
  return typeof p.code === "string" && typeof p.title === "string" && (p.detail === undefined || typeof p.detail === "string");
}
