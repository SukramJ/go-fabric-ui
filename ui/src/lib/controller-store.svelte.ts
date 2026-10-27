// The controller family's state: the controller's status, the node list,
// the credentials and the update queue, and the progress of a
// commissioning, kept current by the controller stream. One instance per
// fabric-ui context, as the node store.
import {
  ApiError,
  type ControllerNodeSummary,
  type ControllerStatus,
  type Credential,
  type FabricClient,
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
}

export interface ControllerStoreDeps {
  t: Translate;
  toast: Toast;
  /** Subscribes to the controller stream; the default opens an EventSource on the client's URL. */
  subscribe?: (onEvent: (ev: ControllerStreamEvent) => void) => () => void;
}

const IDLE: CommissionProgress = { phase: "idle", detail: null, nodeIdHex: null };

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

  let unsub: (() => void) | null = null;
  let reload: ReturnType<typeof setTimeout> | null = null;

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

  function applyEvent(ev: ControllerStreamEvent) {
    switch (ev.type) {
      case "commissioning_progress": {
        const phase = (ev.data.phase ?? "started") as CommissionPhase;
        progress = { phase, detail: ev.data.detail ?? null, nodeIdHex: ev.data.node_id_hex ?? null };
        if (phase === "succeeded") scheduleReload();
        break;
      }
      case "update":
        void loadUpdates();
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
    if (reload) clearTimeout(reload);
    reload = null;
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
