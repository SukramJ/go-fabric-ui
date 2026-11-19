// @vitest-environment happy-dom
// A stream refused 403 ends for good (UR-P3-web-191). The store says so
// and reloads its status, which turns a page that lost its rights to its
// notice; when a later status grants the rights again, the stream is
// opened again (UR-P4-ui-233).
import { describe, expect, it, vi } from "vitest";
import { createClient } from "./api/client";
import { createNodeStore } from "./store.svelte";
import { createControllerStore } from "./controller-store.svelte";

const t = (k: string) => k;
const toast = { info: vi.fn(), success: vi.fn(), error: vi.fn() };
const json = (v: unknown, status = 200) => new Response(JSON.stringify(v), { status, headers: { "Content-Type": "application/json" } });
const sse = () => new Response(new ReadableStream<Uint8Array>(), { status: 200, headers: { "Content-Type": "text/event-stream" } });
const flush = async () => {
  for (let i = 0; i < 5; i++) await new Promise((r) => setTimeout(r, 0));
};

/** Host mode (Options.Admin, no token): the session's right is `admin`, and the stream is refused while it is gone. */
function host(statusPath: string, extra: Record<string, unknown>) {
  const state = { admin: true, streams: 0, statusReads: 0 };
  const client = createClient("/api/v1", async (url) => {
    if (url.endsWith("/stream")) {
      state.streams++;
      return state.admin ? sse() : json({ code: "forbidden", title: "Admin rights required", status: 403, type: "x" }, 403);
    }
    if (url.endsWith(statusPath)) {
      state.statusReads++;
      return json({ ...extra, admin: state.admin });
    }
    return json({ code: "not_found", title: "x", status: 404, type: "x" }, 404);
  });
  return { client, state };
}

describe("UR-P4-ui-233: a refused stream is shown and reopened with the right", () => {
  it("node store: the refusal reloads the status, and a status with the right back reopens the stream", async () => {
    const { client, state } = host("/node/status", { enabled: true, commissioning_window_open: false, fabric_count: 0, endpoint_count: 1 });
    const store = createNodeStore(client, { t, toast });
    await store.loadStatus();
    state.admin = false; // the session ends; the next open is refused
    store.connect();
    await flush();
    expect(store.streamDenied).toBe("forbidden");
    expect(state.statusReads).toBe(2);
    expect(store.isAdmin).toBe(false);
    state.admin = true;
    await store.loadStatus();
    await flush();
    expect(state.streams).toBe(2);
    expect(store.streamDenied).toBeNull();
    store.close();
  });

  it("controller store: the refusal reloads the status, and a status with the right back reopens the stream", async () => {
    const { client, state } = host("/controller/status", { enabled: true });
    const store = createControllerStore(client, { t, toast });
    await store.loadStatus();
    state.admin = false;
    store.connect();
    await flush();
    expect(store.streamDenied).toBe("forbidden");
    expect(state.statusReads).toBe(2);
    expect(store.isAdmin).toBe(false);
    state.admin = true;
    await store.loadStatus();
    await flush();
    expect(state.streams).toBe(2);
    expect(store.streamDenied).toBeNull();
    store.close();
  });
});
