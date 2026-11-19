// @vitest-environment happy-dom
// The streams over fetch (0.10.0, decision 45): fetchEventSource parses the
// event stream and sends the client's headers; a refused stream (403) ends
// without a retry; the stream limit (503 with Retry-After) waits for it.
import { afterEach, describe, expect, it, vi } from "vitest";
import { createClient, REQUEST_HEADER, TOKEN_HEADER } from "./client";
import { fetchEventSource, subscribeControllerStream, subscribeStream, type EventSourceLike } from "./stream";

afterEach(() => {
  vi.useRealTimers();
});

/** A 200 text/event-stream answer whose body the test writes chunk by chunk. */
function sseResponse() {
  let ctrl!: ReadableStreamDefaultController<Uint8Array>;
  const body = new ReadableStream<Uint8Array>({ start: (c) => (ctrl = c) });
  const enc = new TextEncoder();
  return {
    response: new Response(body, { status: 200, headers: { "Content-Type": "text/event-stream" } }),
    write: (s: string) => ctrl.enqueue(enc.encode(s)),
    end: () => ctrl.close(),
  };
}

const flush = () => new Promise((r) => setTimeout(r, 0));

describe("fetchEventSource", () => {
  it("parses events, comments and multi-line data, across chunk borders and CRLF", async () => {
    const sse = sseResponse();
    const seen: string[] = [];
    const src = fetchEventSource({ fetch: async () => sse.response })("/s");
    src.addEventListener("open", () => seen.push("open"));
    src.addEventListener("a", (m) => seen.push(`a:${m.data}`));
    src.addEventListener("message", (m) => seen.push(`message:${m.data}`));
    src.addEventListener("error", () => seen.push("error"));
    await flush();
    sse.write(": hello\n\nevent: a\ndata: {\"x\":1}\n\n");
    sse.write("event: a\r\ndata: one\r");
    sse.write("\ndata: two\r\n\r\n");
    sse.write("data:three\n\n");
    await flush();
    expect(seen).toEqual(["open", 'a:{"x":1}', "a:one\ntwo", "message:three"]);
    expect(src.readyState).toBe(1);
    sse.end();
    await flush();
    expect(seen.at(-1)).toBe("error");
    expect(src.readyState).toBe(2);
  });

  it("reports the status, Retry-After and problem code of a refused answer", async () => {
    const src = fetchEventSource({
      fetch: async () =>
        new Response(JSON.stringify({ code: "too_many_streams", title: "T" }), {
          status: 503,
          headers: { "Content-Type": "application/problem+json", "Retry-After": "5" },
        }),
    })("/s");
    const errors = vi.fn();
    src.addEventListener("error", errors);
    await flush();
    expect(errors).toHaveBeenCalledTimes(1);
    expect([src.readyState, src.status, src.retryAfter, src.problemCode]).toEqual([2, 503, 5, "too_many_streams"]);
  });

  it("aborts the request on close and fires nothing after it", async () => {
    const sse = sseResponse();
    let signal: AbortSignal | undefined;
    const src = fetchEventSource({
      fetch: async (_u, init) => {
        signal = init?.signal ?? undefined;
        return sse.response;
      },
    })("/s");
    const any = vi.fn();
    src.addEventListener("a", any);
    await flush();
    src.close();
    expect(signal?.aborted).toBe(true);
    expect(any).not.toHaveBeenCalled();
  });
});

describe("FabricClient.openStream", () => {
  it("sends the request header and the token on the stream, as on every request (UR-P3-ctr-04)", async () => {
    const sent: { url: string; headers: Record<string, string> }[] = [];
    const client = createClient("/api/v1", async (url, init) => {
      sent.push({ url, headers: (init?.headers ?? {}) as Record<string, string> });
      return sseResponse().response;
    });
    client.setToken("tok");
    const close = subscribeControllerStream(client.controllerStreamUrl(), () => {}, client.openStream, { pauseWhenHidden: false });
    await flush();
    close();
    expect(sent[0].url).toBe("/api/v1/controller/stream");
    expect(sent[0].headers[REQUEST_HEADER]).toBe("1");
    expect(sent[0].headers[TOKEN_HEADER]).toBe("tok");
    expect(sent[0].headers.Accept).toBe("text/event-stream");
  });
});

class Refused implements EventSourceLike {
  static all: Refused[] = [];
  readyState = 0;
  listeners = new Map<string, ((ev: MessageEvent) => void)[]>();
  constructor(
    readonly status: number,
    readonly retryAfter?: number,
    readonly problemCode?: string,
  ) {
    Refused.all.push(this);
  }
  addEventListener(type: string, fn: (ev: MessageEvent) => void) {
    this.listeners.set(type, [...(this.listeners.get(type) ?? []), fn]);
  }
  close() {
    this.readyState = 2;
  }
  fail() {
    this.readyState = 2;
    for (const fn of this.listeners.get("error") ?? []) fn({} as MessageEvent);
  }
}

describe("a refused stream", () => {
  it("ends on 403 without a retry and says why", async () => {
    vi.useFakeTimers();
    Refused.all = [];
    const onDenied = vi.fn();
    subscribeStream("/s", () => {}, () => new Refused(403, undefined, "origin_not_allowed"), { onDenied, pauseWhenHidden: false });
    Refused.all[0].fail();
    await vi.advanceTimersByTimeAsync(60_000);
    expect(Refused.all).toHaveLength(1);
    expect(onDenied).toHaveBeenCalledWith("origin_not_allowed");
  });

  it("waits for Retry-After on 503 before it opens again", async () => {
    vi.useFakeTimers();
    Refused.all = [];
    subscribeStream("/s", () => {}, () => new Refused(503, 5, "too_many_streams"), { pauseWhenHidden: false });
    Refused.all[0].fail();
    await vi.advanceTimersByTimeAsync(4_900);
    expect(Refused.all).toHaveLength(1);
    await vi.advanceTimersByTimeAsync(200);
    expect(Refused.all).toHaveLength(2);
  });
});
