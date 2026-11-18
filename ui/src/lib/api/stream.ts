// The event stream helper: GET /node/stream as typed callbacks. A host with
// a hub of its own feeds it from here (see ui/README.md).
import type {
  AttributeChangedEvent,
  ControllerEvent,
  DiagnosticEvent,
  EventRecord,
  ExposureEvent,
  FabricEvent,
  HealthSampleEvent,
  LogRecord,
  NodesChangedEvent,
  SessionEvent,
  SurfaceChangedEvent,
  WindowEvent,
} from "./client";
import { parseJson } from "./json";

/** The event types of the stream, as the contract names them. */
export const STREAM_EVENTS = [
  "fabric_added",
  "fabric_removed",
  "fabric_updated",
  "commissioning_complete",
  "window_opened",
  "window_closed",
  "endpoints_changed",
  "exposure_changed",
  "surface_changed",
  "session_opened",
  "session_closed",
  "subscriptions_changed",
  "attribute_changed",
  "cluster_event",
  "health_sample",
  "diagnostic_event",
  "nodes_changed",
] as const;

export type StreamEventType = (typeof STREAM_EVENTS)[number];

export type StreamEvent =
  | { type: "fabric_added" | "fabric_removed" | "fabric_updated" | "commissioning_complete"; data: FabricEvent }
  | { type: "window_opened" | "window_closed"; data: WindowEvent }
  | { type: "endpoints_changed"; data: Record<string, never> }
  | { type: "exposure_changed"; data: ExposureEvent }
  | { type: "surface_changed"; data: SurfaceChangedEvent }
  | { type: "session_opened" | "session_closed" | "subscriptions_changed"; data: SessionEvent }
  | { type: "attribute_changed"; data: AttributeChangedEvent }
  | { type: "cluster_event"; data: EventRecord }
  | { type: "health_sample"; data: HealthSampleEvent }
  | { type: "diagnostic_event"; data: DiagnosticEvent }
  | { type: "nodes_changed"; data: NodesChangedEvent };

/** The part of EventSource the helper uses, so a test can hand in a fake. */
export interface EventSourceLike {
  /** The contract's event types, and the source's own "open" and "error". */
  addEventListener(type: string, fn: (ev: MessageEvent) => void): void;
  close(): void;
  /** 2 (CLOSED) after an error the browser does not retry: an answer other than 200 text/event-stream. */
  readonly readyState?: number;
}

export type EventSourceFactory = (url: string) => EventSourceLike;

/** What the stream helpers take beside the callbacks. */
export interface StreamOptions {
  /**
   * Called when the stream opens again after a gap — the browser's own
   * reconnect, a reopen after a failed connection, a hidden tab shown
   * again: events may have been missed, so the caller reloads.
   */
  onResync?: () => void;
  /** Whether the stream closes while the page is hidden; true by default where there is a document. */
  pauseWhenHidden?: boolean;
  /**
   * Whether the first open resyncs too: the caller was subscribed before
   * and may have missed events in between (a store connecting again). The
   * reload then follows the server-side subscription, not precedes it.
   */
  resyncOnFirstOpen?: boolean;
}

/** onResync runs at most once per this many ms; opens within it share one trailing call (UR-P1-bound-195). */
export const RESYNC_MIN_MS = 5000;

/** The waits before reopening a stream that failed for good, the last repeated. */
const BACKOFF_MS = [1000, 2000, 5000, 10000, 30000];
const CLOSED = 2;

// Same-origin cookies only, as the client's fetch sends them
// (credentials: "same-origin"; UR-P1-web-116).
const defaultOpen: EventSourceFactory = (u) => new EventSource(u);

/**
 * Opens the stream at url and hands each event of types with its parsed
 * data to deliver. It reopens a stream that failed for good with a
 * backoff, closes it while the page is hidden and reopens it when shown,
 * and calls onResync on every open but the first (the first too when
 * the page was hidden), at most once per RESYNC_MIN_MS. Returns the
 * close function.
 */
function openStream(
  url: string,
  types: readonly string[],
  deliver: (type: string, msg: MessageEvent) => void,
  open: EventSourceFactory,
  opts: StreamOptions,
): () => void {
  const doc = opts.pauseWhenHidden !== false && typeof document !== "undefined" ? document : null;
  let es: EventSourceLike | null = null;
  // Whether an open has been seen or a gap made one due: the next open
  // resyncs. A page loaded hidden opens its first stream late, after its
  // loads, so that open resyncs too (UR-P1-conc-190).
  let resyncDue = !!doc?.hidden || !!opts.resyncOnFirstOpen;
  let attempt = 0;
  let timer: ReturnType<typeof setTimeout> | null = null;
  let done = false;
  // A stream that opens and drops repeatedly (a proxy cutting idle
  // connections) reloads once per RESYNC_MIN_MS, not on every open.
  let lastResync = -Infinity;
  let resyncTimer: ReturnType<typeof setTimeout> | null = null;
  const resync = () => {
    if (!opts.onResync || resyncTimer) return;
    const run = () => {
      resyncTimer = null;
      if (done) return;
      lastResync = Date.now();
      opts.onResync?.();
    };
    const wait = lastResync + RESYNC_MIN_MS - Date.now();
    if (wait <= 0) run();
    else resyncTimer = setTimeout(run, wait);
  };

  const stop = () => {
    es?.close();
    es = null;
  };
  const start = () => {
    const src = open(url);
    es = src;
    for (const type of types) src.addEventListener(type, (msg) => es === src && deliver(type, msg));
    src.addEventListener("open", () => {
      if (es !== src) return;
      attempt = 0;
      if (resyncDue) resync();
      resyncDue = true;
    });
    src.addEventListener("error", () => {
      if (es !== src || src.readyState !== CLOSED) return; // the browser retries by itself
      stop();
      resyncDue = true;
      const wait = BACKOFF_MS[Math.min(attempt++, BACKOFF_MS.length - 1)];
      timer = setTimeout(() => {
        timer = null;
        if (!done && !doc?.hidden) start();
      }, wait);
    });
  };
  const onVisibility = () => {
    if (done || !doc) return;
    if (doc.hidden) {
      if (timer) clearTimeout(timer);
      timer = null;
      if (es) {
        stop();
        resyncDue = true;
      }
    } else if (!es && !timer) {
      start();
    }
  };

  doc?.addEventListener("visibilitychange", onVisibility);
  if (!doc?.hidden) start();
  return () => {
    done = true;
    if (timer) clearTimeout(timer);
    if (resyncTimer) clearTimeout(resyncTimer);
    doc?.removeEventListener("visibilitychange", onVisibility);
    stop();
  };
}

/** The data of msg as JSON, or undefined when it is not JSON. */
function parseData(msg: MessageEvent, keep?: readonly string[]): { data: unknown } | undefined {
  try {
    return { data: parseJson(String(msg.data), keep) };
  } catch {
    return undefined;
  }
}

/**
 * Subscribes to the stream at `url` and calls `onEvent` for every event;
 * an event whose data is not JSON is skipped. A failed connection is
 * reopened, the stream rests while the page is hidden, and
 * `opts.onResync` says when events may have been missed. Returns the
 * close function.
 */
export function subscribeStream(
  url: string,
  onEvent: (ev: StreamEvent) => void,
  open: EventSourceFactory = defaultOpen,
  opts: StreamOptions = {},
): () => void {
  return openStream(
    url,
    STREAM_EVENTS,
    (type, msg) => {
      // An event's data keeps its integers beyond 2^53 exact, as the
      // client's event routes do.
      const p = parseData(msg, type === "cluster_event" ? ["data"] : undefined);
      if (p) onEvent({ type, data: p.data } as StreamEvent);
    },
    open,
    opts,
  );
}

/** The event types of the controller stream (GET /controller/stream). */
export const CONTROLLER_STREAM_EVENTS = [
  "node_added",
  "node_updated",
  "node_removed",
  "connection",
  "subscription",
  "commissioning_progress",
  "update",
  "topology_changed",
  "controller_diagnostic_event",
] as const;

export type ControllerStreamEventType = (typeof CONTROLLER_STREAM_EVENTS)[number];

/** One controller stream event; controller_diagnostic_event (0.9.0) carries a trace entry, every other a ControllerEvent. */
export type ControllerStreamEvent =
  | { type: Exclude<ControllerStreamEventType, "controller_diagnostic_event">; data: ControllerEvent }
  | { type: "controller_diagnostic_event"; data: DiagnosticEvent };

/**
 * Subscribes to the controller stream at `url`, as subscribeStream does
 * for the node's. Returns the close function.
 */
export function subscribeControllerStream(
  url: string,
  onEvent: (ev: ControllerStreamEvent) => void,
  open: EventSourceFactory = defaultOpen,
  opts: StreamOptions = {},
): () => void {
  return openStream(
    url,
    CONTROLLER_STREAM_EVENTS,
    (type, msg) => {
      const p = parseData(msg);
      if (p) onEvent({ type, data: p.data } as ControllerStreamEvent);
    },
    open,
    opts,
  );
}

/**
 * Subscribes to the log stream at `url` (GET /operation/log/stream) and
 * calls `onRecord` for every log_record. Returns the close function.
 */
export function subscribeLogStream(
  url: string,
  onRecord: (rec: LogRecord) => void,
  open: EventSourceFactory = defaultOpen,
  opts: StreamOptions = {},
): () => void {
  return openStream(
    url,
    ["log_record"],
    (_type, msg) => {
      const p = parseData(msg);
      if (p) onRecord(p.data as LogRecord);
    },
    open,
    opts,
  );
}
