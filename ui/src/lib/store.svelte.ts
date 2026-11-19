// The node family's state: status, fabrics, the selection with its pending
// changes, and the pairing flow, kept current by the event stream. Ported
// from openccu-loom assets/ui/src/lib/stores/matter.svelte.ts:23-381 at
// b40b289a: one instance per fabric-ui context rather than a module
// singleton, the five-part exposure key replaced by the provider's opaque
// item key, the WebSocket topics by the stream's event types, the
// host-readiness gate dropped.
import {
  ApiError,
  deniedCode,
  type EventRecord,
  type FabricClient,
  type FabricList,
  type ExposureSchema,
  type Item,
  type Status,
  type Update,
  type Window,
} from "./api/client";
import { subscribeStream, type StreamEvent } from "./api/stream";
import { secondsUntil } from "./countdown";
import type { Translate } from "./i18n";

export interface Toast {
  info(message: string): void;
  success(message: string): void;
  error(message: string): void;
}

export type CommissioningPhase = "idle" | "open" | "success";

export interface Commissioning {
  phase: CommissioningPhase;
  window: Window | null;
  remaining: number | null;
  addedFabricLabel: string | null;
  /**
   * The open window does not admit the codes the pairing answered
   * (Pairing.window_match false, 0.10.0): it was opened elsewhere, on
   * other credentials (UR-P1-err-04, decision 28).
   */
  foreignWindow: boolean;
}

export interface NodeStoreDeps {
  t: Translate;
  toast: Toast;
  /**
   * Subscribes to the event stream; the default opens an EventSource on
   * the client's stream URL. A subscription that may have missed events
   * (a reconnect) calls onResync, and the store reloads.
   */
  subscribe?: (onEvent: (ev: StreamEvent) => void, onResync?: () => void) => () => void;
  /** Called on the stream's nodes_changed (0.6.0) with the keys it names; the context reloads its node list. */
  onNodesChanged?: (keys: string[]) => void;
}

/** The stream events that change the node's graph (GET /node/topology). */
const GRAPH_EVENTS = new Set<StreamEvent["type"]>([
  "session_opened",
  "session_closed",
  "subscriptions_changed",
  "endpoints_changed",
  "fabric_added",
  "fabric_removed",
  "fabric_updated",
]);

/** How long a burst of attribute_changed is coalesced before an endpoint's epoch moves. */
const ATTRIBUTE_DELAY = 250;

/** How long a burst of health_sample or diagnostic_event is coalesced before the reload. */
const LIVE_DELAY = 300;

const IDLE: Commissioning = { phase: "idle", window: null, remaining: null, addedFabricLabel: null, foreignWindow: false };

/** Whether a status grants admin rights: the token of token mode, or the host's verdict. */
export function isAdminStatus(s: Pick<Status, "csrf_token" | "admin"> | null | undefined): boolean {
  return !!s?.csrf_token || s?.admin === true;
}

export function createNodeStore(client: FabricClient, deps: NodeStoreDeps) {
  let status = $state<Status | null>(null);
  // When the status was read (ms since the epoch): its window_remaining_seconds counts from then.
  let statusAt = $state(0);
  function setStatus(s: Status) {
    status = s;
    statusAt = Date.now();
  }
  let statusLoading = $state(false);
  let statusError = $state<string | null>(null);
  // The 403 problem code of the last status read, or null (0.10.0).
  let statusDenied = $state<string | null>(null);
  // The 403 problem code the stream was refused with, or null (0.10.0).
  let streamDenied = $state<string | null>(null);
  // The last apply stored the selection, but its rebuild failed: not in effect (0.10.0).
  let exposureNotInEffect = $state(false);
  let fabrics = $state<FabricList["fabrics"]>([]);
  let fabricsLoading = $state(false);
  let fabricsError = $state<string | null>(null);
  let schema = $state<ExposureSchema | null>(null);
  let items = $state<Item[]>([]);
  let exposureLoading = $state(false);
  let exposureError = $state<string | null>(null);
  let pending = $state<Map<string, Update>>(new Map());
  let commissioning = $state<Commissioning>({ ...IDLE });
  // topologyEpoch counts the stream's endpoints_changed and
  // surface_changed; a page that shows endpoints or their surface
  // refetches when it moves.
  let topologyEpoch = $state(0);
  // graphEpoch moves, once per burst, on every event that changes the
  // node's graph (GET /node/topology): the sessions, the fabrics and the
  // endpoints.
  let graphEpoch = $state(0);
  // attributeEpochs moves an endpoint's counter once per burst of its
  // attribute_changed: the endpoint page refetches when its counter moves.
  let attributeEpochs = $state<Record<number, number>>({});
  const attributePending = new Set<number>();
  // diagnosticEpoch moves once per burst of diagnostic_event: the
  // diagnostics page refetches its trace when it moves (0.5.0).
  let diagnosticEpoch = $state(0);
  // clusterEvent is the last cluster_event (0.7.0): the endpoint page of
  // its endpoint adds it to the events it shows.
  let clusterEvent = $state<EventRecord | null>(null);

  let unsub: (() => void) | null = null;
  // Whether the stream was connected before: connecting again resyncs.
  let connectedBefore = false;
  // The token the status had when this store opened its stream: the
  // stream's admin test runs once, at its open (UR-P3-web-191).
  let streamToken: string | undefined;
  let countdown: ReturnType<typeof setInterval> | null = null;
  let exposureReload: ReturnType<typeof setTimeout> | null = null;
  let closedTimer: ReturnType<typeof setTimeout> | null = null;
  let graphTimer: ReturnType<typeof setTimeout> | null = null;
  let attributeTimer: ReturnType<typeof setTimeout> | null = null;
  let healthTimer: ReturnType<typeof setTimeout> | null = null;
  let diagnosticTimer: ReturnType<typeof setTimeout> | null = null;

  const message = (err: unknown) => (err instanceof Error ? err.message : String(err));

  // Each loader counts its calls: an answer a newer call overtook is
  // dropped, as the node list's (nodes-store) is (UR-P1-conc-109).
  let statusGen = 0;
  let fabricsGen = 0;
  let exposureGen = 0;

  async function loadStatus() {
    const mine = ++statusGen;
    statusLoading = true;
    statusError = null;
    try {
      const s = await client.status();
      if (mine === statusGen) {
        setStatus(s);
        statusDenied = null;
        reopenOnNewToken(s);
      }
    } catch (err) {
      if (mine === statusGen) {
        status = null;
        statusError = message(err);
        statusDenied = deniedCode(err) ?? null;
      }
    } finally {
      if (mine === statusGen) statusLoading = false;
    }
  }

  async function loadFabrics() {
    const mine = ++fabricsGen;
    fabricsLoading = true;
    fabricsError = null;
    try {
      const list = (await client.fabrics()).fabrics;
      if (mine === fabricsGen) fabrics = list;
    } catch (err) {
      if (mine === fabricsGen) fabricsError = message(err);
    } finally {
      if (mine === fabricsGen) fabricsLoading = false;
    }
  }

  async function loadExposure() {
    const mine = ++exposureGen;
    exposureLoading = true;
    exposureError = null;
    try {
      const [s, list] = await Promise.all([client.exposureSchema(), client.exposure()]);
      if (mine !== exposureGen) return;
      schema = s;
      items = list.items;
    } catch (err) {
      if (mine === exposureGen) exposureError = message(err);
    } finally {
      if (mine === exposureGen) exposureLoading = false;
    }
  }

  // One exposure_changed may follow a burst of others (a host applying
  // hundreds of rows); collapse them into one reload.
  function scheduleExposureReload() {
    if (exposureReload) clearTimeout(exposureReload);
    exposureReload = setTimeout(() => {
      exposureReload = null;
      void loadExposure();
    }, 300);
  }

  // A commissioning opens a session, adds a fabric and its subscriptions
  // in one breath; collapse them into one move of the graph epoch.
  function scheduleGraphReload() {
    if (graphTimer) clearTimeout(graphTimer);
    graphTimer = setTimeout(() => {
      graphTimer = null;
      graphEpoch++;
    }, 300);
  }

  function scheduleAttributeReload(endpoint: number) {
    attributePending.add(endpoint);
    if (attributeTimer) return;
    attributeTimer = setTimeout(() => {
      attributeTimer = null;
      for (const ep of attributePending) attributeEpochs[ep] = (attributeEpochs[ep] ?? 0) + 1;
      attributePending.clear();
    }, ATTRIBUTE_DELAY);
  }

  // A health verdict changes the status; a burst of them is one reload.
  function scheduleHealthReload() {
    if (healthTimer) return;
    healthTimer = setTimeout(() => {
      healthTimer = null;
      void loadStatus();
    }, LIVE_DELAY);
  }

  function scheduleDiagnosticReload() {
    if (diagnosticTimer) return;
    diagnosticTimer = setTimeout(() => {
      diagnosticTimer = null;
      diagnosticEpoch++;
    }, LIVE_DELAY);
  }

  function markDirty(item: Item, patch: { enabled?: boolean; name?: string }) {
    const base: Update = pending.get(item.key) ?? { key: item.key, enabled: item.enabled };
    const next: Update = { ...base, ...patch };
    if (next.name !== undefined && next.name === item.name) delete next.name;
    const map = new Map(pending);
    if (next.enabled === item.enabled && next.name === undefined) map.delete(item.key);
    else map.set(item.key, next);
    pending = map;
  }

  function discardDirty() {
    pending = new Map();
  }

  async function preview(): Promise<Item[]> {
    const updates = Array.from(pending.values());
    if (updates.length === 0) return [];
    return (await client.previewExposure(updates)).items;
  }

  // Rebuilds the node from the stored selection: a success puts a stored
  // selection the last apply could not rebuild into effect (UR-P4-ui-61).
  async function reassemble() {
    await client.reassemble();
    exposureNotInEffect = false;
  }

  async function saveBulk(): Promise<number> {
    const updates = Array.from(pending.values());
    if (updates.length === 0) return 0;
    const res = await client.applyExposureResult(updates);
    pending = new Map();
    exposureNotInEffect = res.reassembled === false;
    await loadExposure();
    return res.applied;
  }

  function stopCountdown() {
    if (countdown !== null) {
      clearInterval(countdown);
      countdown = null;
    }
  }

  function startCountdown(seconds: number) {
    stopCountdown();
    let remaining = seconds;
    commissioning = { ...commissioning, remaining };
    countdown = setInterval(() => {
      remaining -= 1;
      commissioning = { ...commissioning, remaining };
      if (remaining <= 0) {
        stopCountdown();
        if (commissioning.phase === "open") commissioning = { ...IDLE };
      }
    }, 1000);
  }

  async function openWindow(durationSeconds: number) {
    flowGen++;
    const win = await client.openWindow(durationSeconds);
    const mine = ++flowGen;
    // Which kind the window is (basic on fresh credentials, the node's
    // own on configured ones) and its deadline on the node's clock are the
    // status's to say: the first open status after the POST names it
    // (UR-P2-ui-05).
    shown = { adopt: true };
    commissioning = { phase: "open", window: win, remaining: win.duration_seconds, addedFabricLabel: null, foreignWindow: false };
    startCountdown(win.duration_seconds);
    // Read it now, so a later status is judged against it. A failed read
    // leaves the adoption to the next status.
    const g = statusGen;
    let st: Status;
    try {
      st = await client.status();
    } catch {
      return;
    }
    if (g === statusGen) setStatus(st);
    if (mine === flowGen && commissioning.phase === "open") otherWindow(st);
  }

  // The window the open flow shows, as the status named it (kind,
  // deadline in ms on the node's clock): a fresh status naming another
  // one replaces it (UR-P1-err-198). adopt: the flow's own window, not
  // yet named by a status.
  let shown: { kind?: string; deadline?: number; adopt?: boolean } = {};
  // How far a deadline may move and still be the same window (clock and
  // round trip).
  const DEADLINE_SLACK_MS = 5000;
  function windowOf(st: Status): { kind?: string; deadline?: number } {
    return { kind: st.window_kind, deadline: st.window_deadline ? Date.parse(st.window_deadline) : undefined };
  }
  function otherWindow(st: Status): boolean {
    if (shown.adopt) {
      if (st.commissioning_window_open) shown = windowOf(st);
      return false;
    }
    if (st.window_kind && shown.kind && st.window_kind !== shown.kind) return true;
    const d = st.window_deadline ? Date.parse(st.window_deadline) : NaN;
    return !Number.isNaN(d) && shown.deadline !== undefined && Math.abs(d - shown.deadline) > DEADLINE_SLACK_MS;
  }

  // Counts the changes of the pairing flow (an open, a close, a fabric
  // added): a hydrate that one overtook writes nothing (UR-P1-conc-110).
  let flowGen = 0;

  // hydrateCommissioning re-syncs the pairing flow from the node, so a
  // reload or a window opened elsewhere shows the codes and the countdown
  // instead of the empty form (loom matter.svelte.ts:282-318). Without
  // admin rights it reads no codes (GET /pairing is admin-only): the Pair
  // page shows the window from the status.
  async function hydrateCommissioning() {
    if (commissioning.phase !== "idle") return;
    const mine = ++flowGen;
    // The status is read fresh: the pairing's answer is judged against
    // the window it names now (UR-P1-err-04, correction 2).
    const sMine = ++statusGen;
    const st = await client.status();
    if (sMine === statusGen) setStatus(st);
    if (mine !== flowGen || commissioning.phase !== "idle") return;
    if (!st.enabled || !st.commissioning_window_open || !isAdminStatus(st)) return;
    const p = await client.pairing();
    // The flow changed while the codes were read (a fabric added, the
    // window closed or opened here): that change stands.
    if (mine !== flowGen || commissioning.phase !== "idle") return;
    // The window closed between the two reads.
    if (!p.window_open) return;
    const duration = p.duration_seconds ?? st.commissioning_window_duration_seconds ?? 0;
    // What is left, as the node counts it (window_remaining_seconds,
    // 0.10.0; UR-P2-ui-06); from the deadline on a server before it;
    // without either (a server before 0.1.2) the full duration is the
    // upper bound.
    const left =
      p.window_remaining_seconds ?? st.window_remaining_seconds ?? secondsUntil(st.window_deadline) ?? (duration > 0 ? duration : null);
    shown = windowOf(st);
    commissioning = {
      phase: "open",
      window: {
        discriminator: p.discriminator,
        passcode: p.passcode,
        duration_seconds: duration,
        qr_code: p.qr_code,
        manual_code: p.manual_code,
      },
      remaining: left,
      addedFabricLabel: null,
      // The server says whether the open window admits these codes
      // (Pairing.window_match, 0.10.0; decision 28).
      foreignWindow: p.window_match === false,
    };
    if (left !== null && left > 0) startCountdown(left);
  }

  async function closeWindow() {
    flowGen++;
    await client.closeWindow();
    flowGen++;
    stopCountdown();
    commissioning = { ...IDLE };
  }

  function resetCommissioning() {
    flowGen++;
    stopCountdown();
    commissioning = { ...IDLE };
  }

  function applyEvent(ev: StreamEvent) {
    if (GRAPH_EVENTS.has(ev.type)) scheduleGraphReload();
    switch (ev.type) {
      case "fabric_added":
        if (closedTimer) {
          clearTimeout(closedTimer);
          closedTimer = null;
        }
        stopCountdown();
        flowGen++;
        // The codes are done with: the success shows none (UR-P2-leak-08).
        commissioning = { ...commissioning, phase: "success", window: null, remaining: null, addedFabricLabel: ev.data.label || ev.data.vendor_name || null };
        void loadFabrics();
        void loadStatus();
        break;
      case "fabric_removed":
        void loadFabrics();
        void loadStatus();
        break;
      case "fabric_updated":
        void loadFabrics();
        break;
      case "commissioning_complete":
        deps.toast.success(deps.t("node.commissioning.complete", { index: ev.data.fabric_index }));
        void loadFabrics();
        void loadStatus();
        break;
      case "cluster_event":
        clusterEvent = ev.data;
        break;
      case "window_opened":
        // A window opened elsewhere (another tab, the host) shows its
        // codes on an idle flow, as a mount of the Pair page would
        // (UR-P1-ui-132).
        flowGen++;
        // A close just before belongs to the window this one replaced: its
        // timer would end the flow on an open window (UR-P1-err-198).
        if (closedTimer) {
          clearTimeout(closedTimer);
          closedTimer = null;
        }
        if (commissioning.phase === "idle") void hydrateCommissioning().catch(() => undefined);
        else void loadStatus().then(syncCommissioning);
        break;
      case "endpoints_changed":
      case "surface_changed":
        void loadStatus();
        topologyEpoch++;
        break;
      case "window_closed":
        flowGen++;
        void loadStatus();
        // A commissioning closes the window just before the fabric is
        // reported; only a close that no fabric follows ends the flow.
        if (commissioning.phase === "open" && !closedTimer) {
          closedTimer = setTimeout(() => {
            closedTimer = null;
            if (commissioning.phase === "open") {
              flowGen++;
              stopCountdown();
              commissioning = { ...IDLE };
              deps.toast.info(deps.t("node.commissioning.closed"));
            }
          }, 750);
        }
        break;
      case "exposure_changed":
        scheduleExposureReload();
        break;
      case "attribute_changed":
        scheduleAttributeReload(ev.data.endpoint);
        break;
      case "health_sample":
        scheduleHealthReload();
        break;
      case "diagnostic_event":
        scheduleDiagnosticReload();
        break;
      case "nodes_changed":
        deps.onNodesChanged?.(ev.data.keys);
        break;
    }
  }

  // After a gap in the stream the state may be stale: reload what the
  // events would have moved (UR-P1-conc-102).
  function resync() {
    void loadStatus().then(syncCommissioning);
    // The fabrics are an admin's (0.10.0): a request without admin rights is refused.
    if (status === null || isAdminStatus(status)) void loadFabrics();
    if (schema !== null) scheduleExposureReload();
    topologyEpoch++;
    graphEpoch++;
    diagnosticEpoch++;
    deps.onNodesChanged?.([]);
  }

  // An open flow whose window the reloaded status no longer shows open
  // closed while the stream was away.
  function syncCommissioning() {
    if (commissioning.phase === "open" && status !== null && !status.commissioning_window_open) {
      flowGen++;
      stopCountdown();
      commissioning = { ...IDLE };
    } else if (commissioning.phase === "open" && status !== null && otherWindow(status)) {
      // Another window than the one shown: its codes and the notice are
      // judged afresh.
      flowGen++;
      stopCountdown();
      commissioning = { ...IDLE };
      void hydrateCommissioning().catch(() => undefined);
    } else if (commissioning.phase === "idle" && status?.commissioning_window_open) {
      void hydrateCommissioning().catch(() => undefined);
    }
  }

  // The node stream is an open route: opened with a stale token (a daemon
  // restart, a new token per process) or before the first status brought
  // one, it is answered as one without admin rights and filtered for its
  // whole life. A status with another token reopens it; the new open
  // resyncs (UR-P3-web-191).
  function reopenOnNewToken(s: Status) {
    if (!unsub || deps.subscribe || !s.csrf_token || s.csrf_token === streamToken) return;
    disconnect();
    connect();
  }

  function connect() {
    if (unsub) return;
    if (deps.subscribe) {
      // A host's own subscription: reload now when connecting again.
      unsub = deps.subscribe(applyEvent, resync);
      if (connectedBefore) resync();
    } else {
      // Connecting again, the reload waits for the new stream's open, so
      // no event falls between the reload and the subscription
      // (UR-P1-conc-192).
      streamDenied = null;
      // The token the stream opens with is the one the shared client
      // sends now, not this store's last status: a scoped store connects
      // before its first status, with the token the shell's probe set
      // (UR-P4-web-195).
      streamToken = client.currentToken ? client.currentToken() : status?.csrf_token;
      unsub = subscribeStream(client.streamUrl(), applyEvent, client.openStream, {
        onResync: resync,
        resyncOnFirstOpen: connectedBefore,
        onDenied: (code) => (streamDenied = code ?? "forbidden"),
      });
    }
    connectedBefore = true;
  }

  /** Closes the stream only; connect() opens it again and reloads (the shell, while another family is shown). */
  function disconnect() {
    unsub?.();
    unsub = null;
  }

  function close() {
    stopCountdown();
    for (const timer of [exposureReload, closedTimer, graphTimer, attributeTimer, healthTimer, diagnosticTimer]) if (timer) clearTimeout(timer);
    exposureReload = closedTimer = graphTimer = attributeTimer = healthTimer = diagnosticTimer = null;
    attributePending.clear();
    unsub?.();
    unsub = null;
  }

  return {
    get status() { return status; },
    get statusLoading() { return statusLoading; },
    get statusError() { return statusError; },
    /** When the status was read (ms since the epoch); Status.window_remaining_seconds counts from then. */
    get statusAt() { return statusAt; },
    /** The 403 problem code of the last status read (forbidden, host_not_allowed, origin_not_allowed), or null (0.10.0). */
    get statusDenied() { return statusDenied; },
    /** The 403 problem code the stream was refused with, or null; the stream is not reopened then (0.10.0). */
    get streamDenied() { return streamDenied; },
    /** True after an apply whose selection was stored but not rebuilt (Applied.reassembled false, 0.10.0). */
    get exposureNotInEffect() { return exposureNotInEffect; },
    get fabrics() { return fabrics; },
    get fabricsLoading() { return fabricsLoading; },
    get fabricsError() { return fabricsError; },
    get schema() { return schema; },
    get items() { return items; },
    get exposureLoading() { return exposureLoading; },
    get exposureError() { return exposureError; },
    get pending() { return pending; },
    get hasDirty() { return pending.size > 0; },
    get commissioning() { return commissioning; },
    /** Whether the status grants admin rights (the token of token mode, or the host's verdict). */
    get isAdmin() { return isAdminStatus(status); },
    get topologyEpoch() { return topologyEpoch; },
    get graphEpoch() { return graphEpoch; },
    /** Moves once per burst of diagnostic_event; the diagnostics page refetches when it does. */
    get diagnosticEpoch() { return diagnosticEpoch; },
    /** The last cluster_event (0.7.0); the endpoint page of its endpoint adds it to its list. */
    get clusterEvent() { return clusterEvent; },
    /** Moves once per burst of attribute_changed for the endpoint; its page refetches when it does. */
    attributeEpoch(endpoint: number) { return attributeEpochs[endpoint] ?? 0; },
    loadStatus,
    loadFabrics,
    loadExposure,
    markDirty,
    discardDirty,
    preview,
    saveBulk,
    reassemble,
    openWindow,
    hydrateCommissioning,
    closeWindow,
    resetCommissioning,
    applyEvent,
    connect,
    disconnect,
    resync,
    close,
  };
}

export type NodeStore = ReturnType<typeof createNodeStore>;

/** The message of an error a page shows. */
export function errorText(err: unknown): string {
  return err instanceof ApiError ? err.message : err instanceof Error ? err.message : String(err);
}
