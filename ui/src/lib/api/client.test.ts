// The TypeScript half of the contract: every route of the client against
// the golden fixtures the Go handler tests check the server with.
import { describe, expect, it } from "vitest";
import { ApiError, createClient, TOKEN_HEADER, type FabricClient, type Update } from "./client";
import { fixture, fixtureFetch } from "../../test/support";

type Call = (c: FabricClient, body: unknown) => Promise<unknown>;

const items = (b: unknown) => (b as { items: Update[] }).items;
const label = (b: unknown) => (b as { label: string }).label;

// Each fixture with the client call that sends its request.
const cases: [string, Call][] = [
  ["node-status.200", (c) => c.status()],
  ["node-status.200-disabled", (c) => c.status()],
  ["node-fabrics.200", (c) => c.fabrics()],
  ["node-fabrics.503", (c) => c.fabrics()],
  ["node-fabric-remove.204", (c) => c.removeFabric(2)],
  ["node-fabric-remove.400", (c) => c.removeFabric(0)],
  ["node-fabric-remove.403", (c) => c.removeFabric(2)],
  ["node-pairing.200", (c) => c.pairing()],
  ["node-pairing.200-window", (c) => c.pairing()],
  ["node-pairing.403", (c) => c.pairing()],
  ["node-pairing.503", (c) => c.pairing()],
  ["node-pairing-window.200", (c, b) => c.openWindow((b as { duration_seconds: number }).duration_seconds)],
  ["node-pairing-window.400", (c, b) => c.openWindow((b as { duration_seconds: number }).duration_seconds)],
  ["node-pairing-window.409", (c, b) => c.openWindow((b as { duration_seconds: number }).duration_seconds)],
  ["node-pairing-window.503", (c, b) => c.openWindow((b as { duration_seconds: number }).duration_seconds)],
  ["node-pairing-window-close.204", (c) => c.closeWindow()],
  ["node-endpoints.200", (c) => c.endpoints()],
  ["node-endpoint.200", (c) => c.nodeEndpointDetail(2)],
  ["node-endpoint.404", (c) => c.nodeEndpointDetail(9)],
  ["node-endpoint-events.200", (c) => c.nodeEndpointEvents(2)],
  ["node-surface.200", (c) => c.surface()],
  ["node-surface.503", (c) => c.surface()],
  ["node-sessions.200", (c) => c.sessions()],
  ["node-sessions.503", (c) => c.sessions()],
  ["node-topology.200", (c) => c.topology()],
  ["node-topology.503", (c) => c.topology()],
  ["node-mdns.200", (c) => c.mdns()],
  ["node-compatibility.200", (c) => c.compatibility()],
  ["node-events.200", (c) => c.events()],
  ["node-exposure-schema.200", (c) => c.exposureSchema()],
  ["node-exposure-schema.503", (c) => c.exposureSchema()],
  ["node-exposure.200", (c) => c.exposure()],
  ["node-exposure-apply.200", (c, b) => c.applyExposure(items(b))],
  ["node-exposure-apply.204", (c, b) => c.applyExposure(items(b))],
  ["node-exposure-apply.400", (c, b) => c.applyExposure(items(b))],
  ["node-exposure-apply.403", (c, b) => c.applyExposure(items(b))],
  ["node-exposure-preview.200", (c, b) => c.previewExposure(items(b))],
  ["node-exposure-preview.400", (c, b) => c.previewExposure(items(b))],
  ["node-maintenance-reassemble.204", (c) => c.reassemble()],
  ["node-maintenance-reassemble.503", (c) => c.reassemble()],
  ["node-maintenance-factory-reset.204", (c) => c.factoryReset()],
  ["node-maintenance-factory-reset.400", (c) => c.factoryReset()],
  // Completeness against go-fabric v0.18.0 (0.7.0).
  ["node-status.200-attestation", (c) => c.status()],
  ["node-mdns-reannounce.204", (c) => c.reannounce()],
  ["node-mdns-reannounce.503", (c) => c.reannounce()],
  ["nodes-mdns-reannounce.404", (c) => c.forNode("unknown").reannounce()],
  ["node-maintenance-erase.204", (c) => c.erase()],
  ["node-maintenance-erase.400", (c) => c.erase()],
  ["node-maintenance-erase.501", (c) => c.erase()],
  ["nodes-maintenance-erase.404", (c) => c.forNode("unknown").erase()],
  // The node selector and a keyed node's routes (0.5.0).
  ["nodes.200", (c) => c.nodes()],
  ["nodes-status.200", (c) => c.forNode("second").status()],
  ["nodes-status.404", (c) => c.forNode("unknown").status()],
  // The node registry and the configuration block (0.6.0).
  ["nodes.200-registry", (c) => c.nodes()],
  ["nodes-status.200-registry", (c) => c.forNode("kitchen").status()],
  ["node-status.200-config", (c) => c.status()],
  ["nodes-create.201", (c, b) => c.createNode(label(b))],
  ["nodes-create.400", (c, b) => c.createNode(label(b))],
  ["nodes-create.403", (c, b) => c.createNode(label(b))],
  ["nodes-create.409", (c, b) => c.createNode(label(b))],
  ["nodes-create.503", (c, b) => c.createNode(label(b))],
  ["nodes-label.204", (c, b) => c.renameNode("kitchen", label(b))],
  ["nodes-label.400", (c, b) => c.renameNode("kitchen", label(b))],
  ["nodes-label.404", (c, b) => c.renameNode("attic", label(b))],
  ["nodes-label.409", (c, b) => c.renameNode("default", label(b))],
  ["nodes-delete.204", (c) => c.deleteNode("living-room")],
  ["nodes-delete.404", (c) => c.deleteNode("attic")],
  ["nodes-delete.409", (c) => c.deleteNode("default")],
  // Operation (0.5.0).
  ["operation-info.200", (c) => c.operationInfo()],
  ["operation-info.503", (c) => c.operationInfo()],
  ["operation-log.200", (c) => c.logRecords({ since: 2 })],
  ["operation-log.400", (c) => c.logRecords({ limit: 0 })],
  ["operation-log-level.200", (c) => c.logLevel()],
  ["operation-log-level-set.200", (c, b) => c.setLogLevel((b as { level: "debug" }).level)],
  ["operation-log-level-set.400", (c, b) => c.setLogLevel((b as { level: "debug" }).level)],
  ["operation-log-level-set.403", (c, b) => c.setLogLevel((b as { level: "debug" }).level)],
];

describe("client against the contract fixtures", () => {
  it.each(cases)("%s", async (name, call) => {
    const f = fixture(name);
    const ff = fixtureFetch([name]);
    const client = createClient("/api/v1", ff.fetch);
    const token = f.request.headers?.[TOKEN_HEADER];
    client.setToken(token);
    const result = call(client, f.request.body);
    if (f.response.status >= 400) {
      const err = await result.then(
        () => null,
        (e: unknown) => e,
      );
      expect(err).toBeInstanceOf(ApiError);
      expect((err as ApiError).status).toBe(f.response.status);
      expect((err as ApiError).problem?.code).toBe((f.response.body as { code: string }).code);
    } else if (name === "node-exposure-apply.200") {
      expect(await result).toBe((f.response.body as { applied: number }).applied);
    } else if (f.response.status === 204) {
      expect(await result).toBe(name === "node-exposure-apply.204" ? 0 : undefined);
    } else {
      expect(await result).toEqual(f.response.body);
    }
    expect(ff.sent).toHaveLength(1);
    const sent = ff.sent[0];
    expect(`${sent.method} ${sent.path}`).toBe(`${f.request.method} /api/v1${f.request.path}`);
    expect(sent.headers[TOKEN_HEADER]).toBe(token);
    // The factory-reset and erase refusals are the server's answers to a
    // body this client never sends; every other request body is the
    // fixture's.
    if (name !== "node-maintenance-factory-reset.400" && name !== "node-maintenance-erase.400") expect(sent.body).toEqual(f.request.body);
  });

  it("takes the admin token from the status of token mode", async () => {
    const ff = fixtureFetch(["node-pairing-window-close.204"]);
    ff.routes.set("GET /api/v1/node/status", {
      route: "GET /node/status",
      request: { method: "GET", path: "/node/status" },
      response: { status: 200, body: { ...(fixture("node-status.200").response.body as object), csrf_token: "abc" } },
    });
    const client = createClient("/api/v1/", ff.fetch);
    await client.status();
    await client.closeWindow();
    expect(ff.sent[1].headers[TOKEN_HEADER]).toBe("abc");
    expect(ff.sent[0].headers[TOKEN_HEADER]).toBeUndefined();
    expect(client.streamUrl()).toBe("/api/v1/node/stream");
  });

  it("sends since and limit as the event routes' query", async () => {
    const ff = fixtureFetch([]);
    const client = createClient("/api/v1", ff.fetch);
    await client.nodeEndpointEvents(2, { since: 8, limit: 5 }).catch(() => undefined);
    await client.controllerNodeEvents("0000000000000011", { since: 43 }).catch(() => undefined);
    expect(ff.sent.map((s) => s.path)).toEqual(["/api/v1/node/endpoints/2/events?since=8&limit=5", "/api/v1/controller/nodes/0000000000000011/events?since=43"]);
  });

  it("addresses a keyed node's routes and shares the token with it", async () => {
    const ff = fixtureFetch(["nodes-status.200"]);
    const client = createClient("/api/v1", ff.fetch);
    const second = client.forNode("second");
    expect(second.nodeKey).toBe("second");
    expect(client.nodeKey).toBeUndefined();
    expect(client.forNode("second")).toBe(second);
    expect(second.forNode()).toBe(client);
    expect(client.forNode("")).toBe(client);
    await second.status();
    await second.closeWindow().catch(() => undefined);
    await client.closeWindow().catch(() => undefined);
    expect(ff.sent.map((s) => `${s.method} ${s.path} ${s.headers[TOKEN_HEADER] ?? "-"}`)).toEqual([
      "GET /api/v1/nodes/second/status -",
      "POST /api/v1/nodes/second/pairing/window/close test-token",
      "POST /api/v1/node/pairing/window/close test-token",
    ]);
    expect(second.streamUrl()).toBe("/api/v1/nodes/second/stream");
    expect(client.logStreamUrl()).toBe("/api/v1/operation/log/stream");
  });

  it("never sends the token on a read route", async () => {
    const ff = fixtureFetch(["node-fabrics.200", "node-sessions.200"]);
    const client = createClient("/api/v1", ff.fetch);
    client.setToken("secret");
    await client.fabrics();
    await client.sessions();
    expect(ff.sent.every((s) => s.headers[TOKEN_HEADER] === undefined)).toBe(true);
  });

  it("keeps a 64-bit id exact in its hex twin", async () => {
    const ff = fixtureFetch(["node-sessions.200"]);
    const s = await createClient("/api/v1", ff.fetch).sessions();
    expect(s.sessions[0].local_node_id_hex).toBe("B2C3D4E5F6071829");
    expect(Number.isSafeInteger(s.sessions[0].local_node_id)).toBe(false);
  });
});
