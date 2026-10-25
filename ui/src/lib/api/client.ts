// The client of the node routes, typed from contract/openapi.yaml
// (schema.gen.ts). It is the TypeScript half of the contract: the tests in
// client.test.ts read the same golden fixtures as the Go handler tests.
// Mirrors the matter* functions of openccu-loom
// assets/ui/src/lib/api/client.ts:1806-1875 at b40b289a, on the node routes.
import type { components } from "./schema.gen";

type S = components["schemas"];
export type Problem = S["Problem"];
export type Status = S["Status"];
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
  };
}

function isProblem(v: unknown): v is Problem {
  return typeof v === "object" && v !== null && "code" in v && "title" in v;
}
