// The event stream helper: GET /node/stream as typed callbacks. A host with
// a hub of its own feeds it from here (see ui/README.md).
import type { ExposureEvent, FabricEvent, WindowEvent } from "./client";

/** The event types of the stream, as the contract names them. */
export const STREAM_EVENTS = [
  "fabric_added",
  "fabric_removed",
  "window_opened",
  "window_closed",
  "endpoints_changed",
  "exposure_changed",
] as const;

export type StreamEventType = (typeof STREAM_EVENTS)[number];

export type StreamEvent =
  | { type: "fabric_added" | "fabric_removed"; data: FabricEvent }
  | { type: "window_opened" | "window_closed"; data: WindowEvent }
  | { type: "endpoints_changed"; data: Record<string, never> }
  | { type: "exposure_changed"; data: ExposureEvent };

/** The part of EventSource the helper uses, so a test can hand in a fake. */
export interface EventSourceLike {
  addEventListener(type: string, fn: (ev: MessageEvent) => void): void;
  close(): void;
}

export type EventSourceFactory = (url: string) => EventSourceLike;

/**
 * Subscribes to the stream at `url` and calls `onEvent` for every event;
 * an event whose data is not JSON is skipped. The browser's EventSource
 * reconnects on its own. Returns the close function.
 */
export function subscribeStream(
  url: string,
  onEvent: (ev: StreamEvent) => void,
  open: EventSourceFactory = (u) => new EventSource(u, { withCredentials: true }),
): () => void {
  const es = open(url);
  for (const type of STREAM_EVENTS) {
    es.addEventListener(type, (msg: MessageEvent) => {
      let data: unknown;
      try {
        data = JSON.parse(String(msg.data));
      } catch {
        return;
      }
      onEvent({ type, data } as StreamEvent);
    });
  }
  return () => es.close();
}
