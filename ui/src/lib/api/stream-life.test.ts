// @vitest-environment happy-dom
// The stream's life cycle (review phase 1b): a failed connection is
// reopened, a reopen says events may have been missed, a hidden tab rests.
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { subscribeLogStream, subscribeStream, type EventSourceLike } from "./stream";

class FakeSource implements EventSourceLike {
  static all: FakeSource[] = [];
  listeners = new Map<string, ((ev: MessageEvent) => void)[]>();
  closed = false;
  readyState = 0;
  constructor(readonly url: string, readonly init?: EventSourceInit) {
    FakeSource.all.push(this);
  }
  addEventListener(type: string, fn: (ev: MessageEvent) => void) {
    this.listeners.set(type, [...(this.listeners.get(type) ?? []), fn]);
  }
  close() {
    this.closed = true;
    this.readyState = 2;
  }
  emit(type: string, data = "{}") {
    if (type === "open") this.readyState = 1;
    for (const fn of this.listeners.get(type) ?? []) fn({ data } as MessageEvent);
  }
  fail() {
    this.readyState = 2;
    this.emit("error");
  }
}

const open = (u: string) => new FakeSource(u);
let hidden = false;

beforeEach(() => {
  vi.useFakeTimers();
  FakeSource.all = [];
  hidden = false;
  Object.defineProperty(document, "hidden", { configurable: true, get: () => hidden });
});
afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

const setHidden = (h: boolean) => {
  hidden = h;
  document.dispatchEvent(new Event("visibilitychange"));
};

describe("UR-P1-conc-102: a stream that fails is reopened, and a reopen resyncs", () => {
  it("opens a second source after an error in state closed, and resyncs on its open", async () => {
    const onResync = vi.fn();
    const close = subscribeStream("/s", () => {}, open, { onResync });
    const first = FakeSource.all[0];
    first.emit("open");
    expect(onResync).not.toHaveBeenCalled();
    first.fail();
    expect(first.closed).toBe(true);
    await vi.advanceTimersByTimeAsync(1000);
    expect(FakeSource.all).toHaveLength(2);
    FakeSource.all[1].emit("open");
    expect(onResync).toHaveBeenCalledTimes(1);
    close();
    expect(FakeSource.all[1].closed).toBe(true);
  });

  it("resyncs when the browser's own reconnect opens the source again", () => {
    const onResync = vi.fn();
    const close = subscribeStream("/s", () => {}, open, { onResync });
    const src = FakeSource.all[0];
    src.emit("open");
    src.readyState = 0;
    src.emit("error");
    src.emit("open");
    expect(FakeSource.all).toHaveLength(1);
    expect(onResync).toHaveBeenCalledTimes(1);
    close();
  });

  it("backs off, and stops reopening once closed", async () => {
    const close = subscribeStream("/s", () => {}, open);
    FakeSource.all[0].fail();
    await vi.advanceTimersByTimeAsync(1000);
    FakeSource.all[1].fail();
    await vi.advanceTimersByTimeAsync(1000);
    expect(FakeSource.all).toHaveLength(2);
    await vi.advanceTimersByTimeAsync(1000);
    expect(FakeSource.all).toHaveLength(3);
    close();
    FakeSource.all[2].fail();
    await vi.advanceTimersByTimeAsync(60000);
    expect(FakeSource.all).toHaveLength(3);
  });
});

describe("UR-P1-bound-103: a hidden tab holds no stream", () => {
  it("closes the stream while hidden and reopens it, resyncing, when shown", () => {
    const onResync = vi.fn();
    const close = subscribeLogStream("/l", () => {}, open, { onResync });
    FakeSource.all[0].emit("open");
    setHidden(true);
    expect(FakeSource.all[0].closed).toBe(true);
    setHidden(false);
    expect(FakeSource.all).toHaveLength(2);
    FakeSource.all[1].emit("open");
    expect(onResync).toHaveBeenCalledTimes(1);
    close();
    setHidden(true);
    setHidden(false);
    expect(FakeSource.all).toHaveLength(2);
  });
});

describe("UR-P1-web-116: the streams send credentials as the calls do", () => {
  it("opens the default EventSource without withCredentials", () => {
    const ctor = vi.fn(function (this: unknown, url: string, init?: EventSourceInit) {
      return new FakeSource(url, init);
    });
    vi.stubGlobal("EventSource", ctor);
    const close = subscribeStream("/s", () => {});
    expect(ctor).toHaveBeenCalledTimes(1);
    expect(ctor.mock.calls[0][1]?.withCredentials).toBeUndefined();
    close();
  });
});

describe("UR-P1-conc-190: a page loaded hidden resyncs on its first open", () => {
  it("opens no stream while hidden, and the open when shown resyncs", () => {
    hidden = true;
    const onResync = vi.fn();
    const close = subscribeStream("/s", () => {}, open, { onResync });
    expect(FakeSource.all).toHaveLength(0);
    setHidden(false);
    FakeSource.all[0].emit("open");
    expect(onResync).toHaveBeenCalledTimes(1);
    close();
  });
});

describe("UR-P1-bound-195: a stream that opens and drops repeatedly reloads at a bounded rate", () => {
  it("coalesces the resyncs of opens within RESYNC_MIN_MS into one trailing call", async () => {
    const { RESYNC_MIN_MS } = await import("./stream");
    const onResync = vi.fn();
    const close = subscribeStream("/s", () => {}, open, { onResync });
    const src = FakeSource.all[0];
    src.emit("open");
    for (let i = 0; i < 5; i++) {
      src.readyState = 0;
      src.emit("error");
      src.emit("open");
      await vi.advanceTimersByTimeAsync(100);
    }
    expect(onResync).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(RESYNC_MIN_MS);
    expect(onResync).toHaveBeenCalledTimes(2);
    close();
  });
});
