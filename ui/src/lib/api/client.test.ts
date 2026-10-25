// The TypeScript half of the contract: every route of the client against
// the golden fixtures the Go handler tests check the server with.
import { describe, expect, it } from "vitest";
import { ApiError, createClient, TOKEN_HEADER, type FabricClient, type Update } from "./client";
import { fixture, fixtureFetch } from "../../test/support";

type Call = (c: FabricClient, body: unknown) => Promise<unknown>;

const items = (b: unknown) => (b as { items: Update[] }).items;

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
  ["node-sessions.200", (c) => c.sessions()],
  ["node-sessions.503", (c) => c.sessions()],
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
    // The factory-reset refusal is the server's answer to a body this
    // client never sends; every other request body is the fixture's.
    if (name !== "node-maintenance-factory-reset.400") expect(sent.body).toEqual(f.request.body);
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
