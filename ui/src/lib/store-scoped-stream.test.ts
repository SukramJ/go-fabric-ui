// @vitest-environment happy-dom
// A node's own store (NodeScope: loadStatus(),
// then connect() at once) opens its stream with the token the shared
// client already holds, and must not open it a second time when its
// first status brings that same token.
import { describe, expect, it } from "vitest";
import { createClient, TOKEN_HEADER } from "./api/client";
import { createNodeStore, type NodeStoreDeps } from "./store.svelte";

const json = (v: unknown) => new Response(JSON.stringify(v), { status: 200, headers: { "Content-Type": "application/json" } });
const sse = () => new Response(new ReadableStream<Uint8Array>(), { status: 200, headers: { "Content-Type": "text/event-stream" } });
const flush = () => new Promise((r) => setTimeout(r, 20));
const deps: NodeStoreDeps = { t: (k) => k, toast: { success: () => undefined, error: () => undefined, info: () => undefined } };

describe("a scoped node's stream (UR-P4-web-195)", () => {
  it("opens once when the client already holds the token", async () => {
    const seen: string[] = [];
    const root = createClient("/api/v1", async (url: string, init?: RequestInit) => {
      seen.push(`${url.replace("/api/v1", "")} ${((init?.headers ?? {}) as Record<string, string>)[TOKEN_HEADER] ?? "-"}`);
      if (url.endsWith("/status")) return json({ enabled: true, csrf_token: "t1" });
      if (url.endsWith("/stream")) return sse();
      return json({});
    });
    await root.status(); // the shell's probe: the shared token is t1
    const store = createNodeStore(root.forNode("second"), deps);
    void store.loadStatus(); // NodeScope.svelte:32-33
    store.connect();
    await flush();
    await flush();
    expect(seen.filter((l) => l.includes("/stream"))).toEqual(["/nodes/second/stream t1"]);
    store.close();
  });
});
