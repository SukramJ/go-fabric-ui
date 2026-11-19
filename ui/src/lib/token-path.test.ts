// @vitest-environment happy-dom
// The token path of the web client (review phase 3, pre-merge): a status
// answered without the token is repeated with it (UR-P3-web-190), the ACL
// page asks when it does not know its own node id, a stream refused with a
// stale token reloads it and reopens once (UR-P3-web-191), and the token
// goes only to a URL the page's origin resolves to (UR-P3-web-194).
import { afterEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import NodeHarness from "../test/NodeHarness.svelte";
import { body, testUI } from "../test/support";
import { createClient, REQUEST_HEADER, TOKEN_HEADER, type ControllerStatus, type Status } from "./api/client";
import { createControllerStore } from "./controller-store.svelte";
import { createNodeStore, type NodeStoreDeps } from "./store.svelte";

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

const L = "0000000000000011";
const FULL = body<ControllerStatus>("controller-status.200");
// The controller status as the server answers it without admin rights (decision 57).
const REDUCED: ControllerStatus = {
  ...FULL,
  compressed_fabric_id_hex: "",
  fabric_id_hex: "",
  fabric_index: 0,
  label: "",
  node_count: 0,
  node_id_hex: "",
  vendor_id: 0,
  vendor_name: "",
};
const NODE = { ...body<Status>("node-status.200-config"), csrf_token: "t1" };

const json = (v: unknown, status = 200) => new Response(JSON.stringify(v), { status, headers: { "Content-Type": "application/json" } });
const forbidden = () => json({ code: "forbidden", title: "Admin rights required", status: 403, type: "x" }, 403);
const tokenOf = (init?: RequestInit) => ((init?.headers ?? {}) as Record<string, string>)[TOKEN_HEADER];

/** A token-mode server: the full status with the current token, the reduced one without. */
function tokenServer(token = "t1") {
  const seen: string[] = [];
  const fetchImpl = async (url: string, init?: RequestInit) => {
    const sent = tokenOf(init);
    seen.push(`${url.replace("/api/v1", "")} ${sent ?? "-"}`);
    const admin = sent === token;
    if (url.endsWith("/controller/status")) return json({ ...(admin ? FULL : REDUCED), csrf_token: token });
    if (url.endsWith("/node/status")) {
      const s = structuredClone(NODE);
      s.csrf_token = token;
      if (admin && s.config) s.config.store = "bridge.db";
      return json(s);
    }
    return json({ code: "not_found", title: "no route", status: 404, type: "x" }, 404);
  };
  return { fetchImpl, seen };
}

describe("UR-P3-web-190: a status answered without the token is repeated with it", () => {
  it("controllerStatus and status return the admin's answer", async () => {
    const s = tokenServer();
    const client = createClient("/api/v1", s.fetchImpl);
    expect((await client.controllerStatus()).node_id_hex).toBe(FULL.node_id_hex);
    expect((await client.status()).config?.store).toBe("bridge.db");
    expect(s.seen).toEqual(["/controller/status -", "/controller/status t1", "/node/status t1"]);
  });

  it("a stale token is replaced and the status repeated once", async () => {
    const s = tokenServer("t2");
    const client = createClient("/api/v1", s.fetchImpl);
    client.setToken("t1");
    expect((await client.status()).config?.store).toBe("bridge.db");
    expect(s.seen).toEqual(["/node/status t1", "/node/status t2"]);
  });

  it("the controller store holds the admin's status after its first load", async () => {
    const s = tokenServer();
    const store = createControllerStore(createClient("/api/v1", s.fetchImpl), {
      t: (k) => k,
      toast: { info: vi.fn(), success: vi.fn(), error: vi.fn() },
      subscribe: () => () => {},
    });
    await store.loadStatus();
    expect(store.status?.node_id_hex).toBe(FULL.node_id_hex);
    expect(store.status?.node_count).toBe(FULL.node_count);
  });
});

describe("UR-P3-web-190: the ACL page without its own node id", () => {
  async function saveAddedEntry(status: ControllerStatus) {
    const ui = testUI({
      fixtures: ["controller-node.200", "controller-node-acl.200", "controller-node-acl-set.200"],
      answers: { "GET /controller/status": { status: 200, body: status } },
    });
    await ui.controller.loadStatus();
    render(NodeHarness, { ui, nodeId: L, tab: "acl" });
    await screen.findByTestId("acl-entries");
    await fireEvent.input(screen.getByLabelText("controller.acl.subjects"), { target: { value: "5" } });
    await fireEvent.click(screen.getByText("controller.acl.add_entry"));
    await fireEvent.click(screen.getByText("common.save"));
    await waitFor(() => expect(ui.sent.some((r) => r.method === "PUT")).toBe(true));
    return ui;
  }

  it("asks before the write when the own id is unknown", async () => {
    const ui = await saveAddedEntry(REDUCED);
    expect(ui.confirmSpy).toHaveBeenCalledTimes(1);
  });

  it("does not ask when the list keeps the own id", async () => {
    const ui = await saveAddedEntry(FULL);
    expect(ui.confirmSpy).not.toHaveBeenCalled();
  });
});

/** A 200 text/event-stream answer that stays open. */
function sse() {
  return new Response(new ReadableStream<Uint8Array>(), { status: 200, headers: { "Content-Type": "text/event-stream" } });
}
const flush = () => new Promise((r) => setTimeout(r, 0));

describe("UR-P3-web-191: a stream refused with a stale token", () => {
  function server(streamAdmits: (token: string | undefined) => boolean) {
    const seen: string[] = [];
    const fetchImpl = async (url: string, init?: RequestInit) => {
      const sent = tokenOf(init);
      seen.push(`${url.replace("/api/v1", "")} ${sent ?? "-"}`);
      if (url.endsWith("/node/status")) return json({ enabled: true, csrf_token: "new" });
      return streamAdmits(sent) ? sse() : forbidden();
    };
    return { fetchImpl, seen };
  }

  it("reloads the token once and reopens with it", async () => {
    const s = server((t) => t === "new");
    const client = createClient("/api/v1", s.fetchImpl);
    client.setToken("old");
    const events: string[] = [];
    const src = client.openStream(client.controllerStreamUrl());
    src.addEventListener("open", () => events.push("open"));
    src.addEventListener("error", () => events.push("error"));
    await vi.waitFor(() => expect(events).toEqual(["open"]));
    expect(src.readyState).toBe(1);
    expect(s.seen).toEqual(["/controller/stream old", "/node/status old", "/controller/stream new"]);
    src.close();
  });

  it("a second refusal ends the source", async () => {
    const s = server(() => false);
    const client = createClient("/api/v1", s.fetchImpl);
    client.setToken("old");
    const events: string[] = [];
    const src = client.openStream(client.controllerStreamUrl());
    src.addEventListener("error", () => events.push("error"));
    await vi.waitFor(() => expect(events).toEqual(["error"]));
    await flush();
    expect(src.readyState).toBe(2);
    expect(src.status).toBe(403);
    expect(s.seen).toEqual(["/controller/stream old", "/node/status old", "/controller/stream new"]);
  });
});

// The node stream is an open route: a stale token is not refused, the
// stream is answered as one without admin rights and stays filtered. The
// store reopens it when a status read brings another token (UR-P3-web-191,
// the rest carried to phase 4).
describe("UR-P3-web-191: the node stream after a token change", () => {
  function server() {
    const state = { token: "t1" };
    const seen: string[] = [];
    const fetchImpl = async (url: string, init?: RequestInit) => {
      seen.push(`${url.replace("/api/v1", "")} ${tokenOf(init) ?? "-"}`);
      if (url.endsWith("/node/status")) return json({ ...NODE, csrf_token: state.token });
      if (url.endsWith("/node/stream")) return sse();
      return json({ code: "not_found", title: "no route", status: 404, type: "x" }, 404);
    };
    return { state, fetchImpl, seen };
  }
  const deps: NodeStoreDeps = { t: (k) => k, toast: { success: () => undefined, error: () => undefined, info: () => undefined } };
  const streams = (seen: string[]) => seen.filter((l) => l.startsWith("/node/stream"));

  it("reopens with the new token when a status read brings one", async () => {
    const s = server();
    const store = createNodeStore(createClient("/api/v1", s.fetchImpl), deps);
    await store.loadStatus();
    store.connect();
    await flush();
    expect(streams(s.seen)).toEqual(["/node/stream t1"]);
    // The daemon restarted: a new token per process.
    s.state.token = "t2";
    await store.loadStatus();
    await vi.waitFor(() => expect(streams(s.seen)).toEqual(["/node/stream t1", "/node/stream t2"]));
    // The reopened stream resyncs once; the same token does not reopen again.
    await flush();
    await store.loadStatus();
    await flush();
    expect(streams(s.seen)).toEqual(["/node/stream t1", "/node/stream t2"]);
    store.close();
  });

  it("reopens a stream opened before the first status brought the token", async () => {
    const s = server();
    const store = createNodeStore(createClient("/api/v1", s.fetchImpl), deps);
    store.connect();
    await flush();
    await store.loadStatus();
    await vi.waitFor(() => expect(streams(s.seen)).toEqual(["/node/stream -", "/node/stream t1"]));
    store.close();
  });

  it("leaves a closed store's stream closed", async () => {
    const s = server();
    const store = createNodeStore(createClient("/api/v1", s.fetchImpl), deps);
    await store.loadStatus();
    s.state.token = "t2";
    await store.loadStatus();
    await flush();
    expect(streams(s.seen)).toEqual([]);
  });
});

describe("UR-P3-web-194: the token only to the page's origin, as the URL resolves", () => {
  it("sends no token to a base that resolves to another host", async () => {
    vi.stubGlobal("location", new URL("http://ui.test/app/"));
    const seen: Record<string, string>[] = [];
    const fetchImpl = async (_u: string, init?: RequestInit) => {
      seen.push((init?.headers ?? {}) as Record<string, string>);
      return new Response(null, { status: 204 });
    };
    for (const base of ["\\\\other.example/api", "/\\other.example", "\\/other.example", " //other.example", "./api/v1"]) {
      const c = createClient(base, fetchImpl);
      c.setToken("secret");
      await c.closeWindow();
    }
    expect(seen.map((h) => h[REQUEST_HEADER])).toEqual(["1", "1", "1", "1", "1"]);
    expect(seen.map((h) => h[TOKEN_HEADER])).toEqual([undefined, undefined, undefined, undefined, "secret"]);
  });

  it("openStream sends the token only to a URL of the page's origin", async () => {
    vi.stubGlobal("location", new URL("http://ui.test/app/"));
    const seen: (string | undefined)[] = [];
    const client = createClient("/api/v1", async (_u, init) => {
      seen.push(tokenOf(init));
      return sse();
    });
    client.setToken("secret");
    const a = client.openStream("https://elsewhere.example/stream");
    const b = client.openStream(client.streamUrl());
    await flush();
    a.close();
    b.close();
    expect(seen).toEqual([undefined, "secret"]);
  });
});
