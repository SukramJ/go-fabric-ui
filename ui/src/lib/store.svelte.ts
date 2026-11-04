// The node family's state: status, fabrics, the selection with its pending
// changes, and the pairing flow, kept current by the event stream. Ported
// from openccu-loom assets/ui/src/lib/stores/matter.svelte.ts:23-381 at
// b40b289a: one instance per fabric-ui context rather than a module
// singleton, the five-part exposure key replaced by the provider's opaque
// item key, the WebSocket topics by the stream's event types, the
// host-readiness gate dropped.
import { ApiError, type FabricClient, type FabricList, type ExposureSchema, type Item, type Status, type Update, type Window } from "./api/client";
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
}

export interface NodeStoreDeps {
  t: Translate;
  toast: Toast;
  /** Subscribes to the event stream; the default opens an EventSource on the client's stream URL. */
  subscribe?: (onEvent: (ev: StreamEvent) => void) => () => void;
}

/** The stream events that change the node's graph (GET /node/topology). */
const GRAPH_EVENTS = new Set<StreamEvent["type"]>([
  "session_opened",
  "session_closed",
  "subscriptions_changed",
  "endpoints_changed",
  "fabric_added",
  "fabric_removed",
]);

/** How long a burst of attribute_changed is coalesced before an endpoint's epoch moves. */
const ATTRIBUTE_DELAY = 250;

/** How long a burst of health_sample or diagnostic_event is coalesced before the reload. */
const LIVE_DELAY = 300;

const IDLE: Commissioning = { phase: "idle", window: null, remaining: null, addedFabricLabel: null };

export function createNodeStore(client: FabricClient, deps: NodeStoreDeps) {
  let status = $state<Status | null>(null);
  let statusLoading = $state(false);
  let statusError = $state<string | null>(null);
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

  let unsub: (() => void) | null = null;
  let countdown: ReturnType<typeof setInterval> | null = null;
  let exposureReload: ReturnType<typeof setTimeout> | null = null;
  let closedTimer: ReturnType<typeof setTimeout> | null = null;
  let graphTimer: ReturnType<typeof setTimeout> | null = null;
  let attributeTimer: ReturnType<typeof setTimeout> | null = null;
  let healthTimer: ReturnType<typeof setTimeout> | null = null;
  let diagnosticTimer: ReturnType<typeof setTimeout> | null = null;

  const message = (err: unknown) => (err instanceof Error ? err.message : String(err));

  async function loadStatus() {
    statusLoading = true;
    statusError = null;
    try {
      status = await client.status();
    } catch (err) {
      status = null;
      statusError = message(err);
    } finally {
      statusLoading = false;
    }
  }

  async function loadFabrics() {
    fabricsLoading = true;
    fabricsError = null;
    try {
      fabrics = (await client.fabrics()).fabrics;
    } catch (err) {
      fabricsError = message(err);
    } finally {
      fabricsLoading = false;
    }
  }

  async function loadExposure() {
    exposureLoading = true;
    exposureError = null;
    try {
      const [s, list] = await Promise.all([client.exposureSchema(), client.exposure()]);
      schema = s;
      items = list.items;
    } catch (err) {
      exposureError = message(err);
    } finally {
      exposureLoading = false;
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

  async function saveBulk(): Promise<number> {
    const updates = Array.from(pending.values());
    if (updates.length === 0) return 0;
    const applied = await client.applyExposure(updates);
    pending = new Map();
    await loadExposure();
    return applied;
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
    const win = await client.openWindow(durationSeconds);
    commissioning = { phase: "open", window: win, remaining: win.duration_seconds, addedFabricLabel: null };
    startCountdown(win.duration_seconds);
  }

  // hydrateCommissioning re-syncs the pairing flow from the node, so a
  // reload or a window opened elsewhere shows the codes and the countdown
  // instead of the empty form (loom matter.svelte.ts:282-318).
  async function hydrateCommissioning() {
    if (commissioning.phase !== "idle") return;
    let st = status;
    if (st === null) {
      st = await client.status();
      status = st;
    }
    if (!st.enabled || !st.commissioning_window_open) return;
    const p = await client.pairing();
    const duration = p.duration_seconds ?? st.commissioning_window_duration_seconds ?? 0;
    // The deadline says what is left; without one (a server before 0.1.2)
    // the full duration is the upper bound.
    const left = secondsUntil(st.window_deadline) ?? (duration > 0 ? duration : null);
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
    };
    if (left !== null && left > 0) startCountdown(left);
  }

  async function closeWindow() {
    await client.closeWindow();
    stopCountdown();
    commissioning = { ...IDLE };
  }

  function resetCommissioning() {
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
        commissioning = { ...commissioning, phase: "success", addedFabricLabel: ev.data.label || ev.data.vendor_name || null };
        void loadFabrics();
        void loadStatus();
        break;
      case "fabric_removed":
        void loadFabrics();
        void loadStatus();
        break;
      case "window_opened":
        void loadStatus();
        break;
      case "endpoints_changed":
      case "surface_changed":
        void loadStatus();
        topologyEpoch++;
        break;
      case "window_closed":
        void loadStatus();
        // A commissioning closes the window just before the fabric is
        // reported; only a close that no fabric follows ends the flow.
        if (commissioning.phase === "open" && !closedTimer) {
          closedTimer = setTimeout(() => {
            closedTimer = null;
            if (commissioning.phase === "open") {
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
    }
  }

  function connect() {
    if (unsub) return;
    const sub = deps.subscribe ?? ((fn) => subscribeStream(client.streamUrl(), fn));
    unsub = sub(applyEvent);
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
    get topologyEpoch() { return topologyEpoch; },
    get graphEpoch() { return graphEpoch; },
    /** Moves once per burst of diagnostic_event; the diagnostics page refetches when it does. */
    get diagnosticEpoch() { return diagnosticEpoch; },
    /** Moves once per burst of attribute_changed for the endpoint; its page refetches when it does. */
    attributeEpoch(endpoint: number) { return attributeEpochs[endpoint] ?? 0; },
    loadStatus,
    loadFabrics,
    loadExposure,
    markDirty,
    discardDirty,
    preview,
    saveBulk,
    openWindow,
    hydrateCommissioning,
    closeWindow,
    resetCommissioning,
    applyEvent,
    connect,
    close,
  };
}

export type NodeStore = ReturnType<typeof createNodeStore>;

/** The message of an error a page shows. */
export function errorText(err: unknown): string {
  return err instanceof ApiError ? err.message : err instanceof Error ? err.message : String(err);
}
