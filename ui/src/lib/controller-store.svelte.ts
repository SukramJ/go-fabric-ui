// The controller family's state: the controller's status, the node list,
// the credentials and the update queue, and the progress of a
// commissioning, kept current by the controller stream. One instance per
// fabric-ui context, as the node store.
import {
  ApiError,
  type BLEDecision,
  type ControllerNodeSummary,
  type ControllerStatus,
  type Credential,
  type FabricClient,
  type PairingStep,
  type UpdateQueue,
} from "./api/client";
import { subscribeControllerStream, type ControllerStreamEvent } from "./api/stream";
import type { Toast } from "./store.svelte";
import type { Translate } from "./i18n";

export type CommissionPhase = "idle" | "started" | "succeeded" | "failed";

export interface CommissionProgress {
  phase: CommissionPhase;
  /** The failure's kind (no_device, timeout, …) of a failed commissioning. */
  detail: string | null;
  /** The node a succeeded commissioning added. */
  nodeIdHex: string | null;
  /** The steps the controller reported, in order, each as it started and ended (0.7.0). */
  steps: PairingStep[];
  /** Whether the commissioning searches over Bluetooth LE and, if not, why (0.7.1); null until reported. */
  ble: BLEDecision | null;
}

export interface ControllerStoreDeps {
  t: Translate;
  toast: Toast;
  /** Subscribes to the controller stream; the default opens an EventSource on the client's URL. */
  subscribe?: (onEvent: (ev: ControllerStreamEvent) => void) => () => void;
}

const IDLE: CommissionProgress = { phase: "idle", detail: null, nodeIdHex: null, steps: [], ble: null };

/** How long the store waits before reloading after an event, so a burst reloads once. */
const RELOAD_DELAY = 250;

export function createControllerStore(client: FabricClient, deps: ControllerStoreDeps) {
  let status = $state<ControllerStatus | null>(null);
  // null until probed; false when the host serves no controller (404).
  let present = $state<boolean | null>(null);
  let statusError = $state<string | null>(null);
  let nodes = $state<ControllerNodeSummary[]>([]);
  let nodesLoading = $state(false);
  let nodesError = $state<string | null>(null);
  let credentials = $state<Credential[]>([]);
  let credentialsError = $state<string | null>(null);
  let updates = $state<UpdateQueue | null>(null);
  let progress = $state<CommissionProgress>({ ...IDLE });
  // Bumped on every event about a node, so an open node page reloads.
  let revision = $state(0);
  let changed = $state<string | null>(null);
  // graphEpoch moves, once per burst, on every topology_changed: the
  // network graph's page (GET /controller/topology) reloads when it does.
  let graphEpoch = $state(0);
  // diagnosticEpoch moves, once per burst, on every
  // controller_diagnostic_event (0.9.0): the diagnostics page reloads.
  let diagnosticEpoch = $state(0);

  let unsub: (() => void) | null = null;
  let reload: ReturnType<typeof setTimeout> | null = null;
  let graphTimer: ReturnType<typeof setTimeout> | null = null;
  let diagnosticTimer: ReturnType<typeof setTimeout> | null = null;

  const message = (err: unknown) => (err instanceof Error ? err.message : String(err));

  /** Reads the status; a 404 marks the family absent, not an error. */
  async function loadStatus() {
    statusError = null;
    try {
      status = await client.controllerStatus();
      present = true;
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) {
        present = false;
        status = null;
        return;
      }
      statusError = message(err);
    }
  }

  async function loadNodes() {
    nodesLoading = true;
    nodesError = null;
    try {
      nodes = (await client.controllerNodes()).nodes;
    } catch (err) {
      nodesError = message(err);
    } finally {
      nodesLoading = false;
    }
  }

  async function loadCredentials() {
    credentialsError = null;
    try {
      credentials = (await client.credentials()).credentials;
    } catch (err) {
      credentialsError = message(err);
    }
  }

  async function loadUpdates() {
    try {
      updates = await client.updates();
    } catch (err) {
      deps.toast.error(message(err));
    }
  }

  function scheduleReload() {
    if (reload) return;
    reload = setTimeout(() => {
      reload = null;
      void loadNodes();
      void loadStatus();
    }, RELOAD_DELAY);
  }

  // A commissioning or a refresh moves the graph in steps; collapse a
  // burst into one move of the graph epoch, as the node store does.
  function scheduleGraphReload() {
    if (graphTimer) clearTimeout(graphTimer);
    graphTimer = setTimeout(() => {
      graphTimer = null;
      graphEpoch++;
    }, 300);
  }

  // A trace entry rarely comes alone (a reconnect records a session and
  // a reachability verdict): one move of the epoch per burst.
  function scheduleDiagnosticReload() {
    if (diagnosticTimer) clearTimeout(diagnosticTimer);
    diagnosticTimer = setTimeout(() => {
      diagnosticTimer = null;
      diagnosticEpoch++;
    }, 300);
  }

  function applyEvent(ev: ControllerStreamEvent) {
    switch (ev.type) {
      case "commissioning_progress": {
        const step = ev.data.step;
        const ble = ev.data.ble;
        if (ble) {
          // The BLE decision is reported under the phase started (0.7.1),
          // before the first step: it is kept, the phase stays.
          progress = { ...progress, phase: progress.phase === "idle" ? "started" : progress.phase, ble };
          break;
        }
        if (step) {
          // A step is reported under the phase started (0.7.0): it joins
          // the list, the commissioning's own phase stays.
          progress = { ...progress, phase: progress.phase === "idle" ? "started" : progress.phase, steps: [...progress.steps, step] };
          break;
        }
        const phase = (ev.data.phase ?? "started") as CommissionPhase;
        // The commissioning's own started begins a new list of steps and
        // clears the last decision.
        const fresh = phase === "started";
        progress = {
          phase,
          detail: ev.data.detail ?? null,
          nodeIdHex: ev.data.node_id_hex ?? null,
          steps: fresh ? [] : progress.steps,
          ble: fresh ? null : progress.ble,
        };
        if (phase === "succeeded") scheduleReload();
        break;
      }
      case "update":
        // The source's update events carry their kind (0.7.0): a finished
        // or failed update is worth a toast; every kind reloads the queue.
        if (ev.data.update_kind === "done") deps.toast.success(deps.t("controller.updates.event_done", { version: ev.data.update_version ?? "" }));
        if (ev.data.update_kind === "failed") deps.toast.error(deps.t("controller.updates.event_failed", { error: ev.data.update_error ?? "" }));
        void loadUpdates();
        break;
      case "topology_changed":
        // The network graph's page reloads on it; no node changed, so the
        // lists stay.
        scheduleGraphReload();
        break;
      case "controller_diagnostic_event":
        // The diagnostics page reloads on it; the lists stay.
        scheduleDiagnosticReload();
        break;
      default:
        changed = ev.data.node_id_hex ?? null;
        revision++;
        scheduleReload();
    }
  }

  function resetProgress() {
    progress = { ...IDLE };
  }

  function connect() {
    if (unsub) return;
    const sub = deps.subscribe ?? ((fn) => subscribeControllerStream(client.controllerStreamUrl(), fn));
    unsub = sub(applyEvent);
  }

  function close() {
    for (const timer of [reload, graphTimer, diagnosticTimer]) if (timer) clearTimeout(timer);
    reload = graphTimer = diagnosticTimer = null;
    unsub?.();
    unsub = null;
  }

  return {
    get status() { return status; },
    get present() { return present; },
    get statusError() { return statusError; },
    get nodes() { return nodes; },
    get nodesLoading() { return nodesLoading; },
    get nodesError() { return nodesError; },
    get credentials() { return credentials; },
    get credentialsError() { return credentialsError; },
    get updates() { return updates; },
    get progress() { return progress; },
    /** Counts the events about nodes; a node page reloads when it moves. */
    get revision() { return revision; },
    /** The node the last event named. */
    get changed() { return changed; },
    /** Moves once per burst of topology_changed; the network page reloads when it does. */
    get graphEpoch() { return graphEpoch; },
    /** Moves once per burst of controller_diagnostic_event (0.9.0); the diagnostics page reloads when it does. */
    get diagnosticEpoch() { return diagnosticEpoch; },
    loadStatus,
    loadNodes,
    loadCredentials,
    loadUpdates,
    applyEvent,
    resetProgress,
    connect,
    close,
  };
}

export type ControllerStore = ReturnType<typeof createControllerStore>;
