import { describe, expect, it, vi } from "vitest";
import { STREAM_EVENTS, subscribeStream, type EventSourceLike, type StreamEvent } from "./stream";

class FakeSource implements EventSourceLike {
  listeners = new Map<string, (ev: MessageEvent) => void>();
  closed = false;
  addEventListener(type: string, fn: (ev: MessageEvent) => void) {
    this.listeners.set(type, fn);
  }
  close() {
    this.closed = true;
  }
  emit(type: string, data: string) {
    this.listeners.get(type)?.({ data } as MessageEvent);
  }
}

describe("subscribeStream", () => {
  it("delivers the contract's event types with their JSON payloads and closes", () => {
    const src = new FakeSource();
    const got: StreamEvent[] = [];
    const open = vi.fn(() => src);
    const close = subscribeStream("/api/v1/node/stream", (ev) => got.push(ev), open);
    expect(open).toHaveBeenCalledWith("/api/v1/node/stream");
    expect([...src.listeners.keys()]).toEqual([...STREAM_EVENTS]);
    src.emit("fabric_added", '{"fabric_index":3,"label":"Home"}');
    src.emit("exposure_changed", "not json");
    src.emit("window_closed", "{}");
    expect(got).toEqual([
      { type: "fabric_added", data: { fabric_index: 3, label: "Home" } },
      { type: "window_closed", data: {} },
    ]);
    close();
    expect(src.closed).toBe(true);
  });
});
