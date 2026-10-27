// The controller family's half of the contract: every controller route of
// the client against the golden fixtures the Go handler tests check the
// server with.
import { describe, expect, it } from "vitest";
import { ApiError, createClient, TOKEN_HEADER, type FabricClient } from "./client";
import { fixture, fixtureFetch } from "../../test/support";

type Call = (c: FabricClient, body: unknown) => Promise<unknown>;

const L = "0000000000000011";
const S = "0000000000000012";
const b = (x: unknown) => x as any;

const cases: [string, Call][] = [
  ["controller-status.200", (c) => c.controllerStatus()],
  ["controller-nodes.200", (c) => c.controllerNodes()],
  ["controller-node.200", (c) => c.controllerNode(L)],
  ["controller-node.200-icd", (c) => c.controllerNode(S)],
  ["controller-node.400", (c) => c.controllerNode("xyz")],
  ["controller-node.404", (c) => c.controllerNode("00000000000000AA")],
  ["controller-node-decommission.204", (c) => c.decommissionNode(S)],
  ["controller-node-decommission.403", (c) => c.decommissionNode(S)],
  ["controller-node-interview.204", (c) => c.interviewNode(L)],
  ["controller-node-enable.204", (c) => c.enableNode(L)],
  ["controller-node-disable.204", (c) => c.disableNode(L)],
  ["controller-node-auto-subscribe.204", (c, x) => c.setNodeAutoSubscribe(L, b(x).enabled)],
  ["controller-node-window.200", (c, x) => c.openNodeWindow(L, b(x).timeout_seconds)],
  ["controller-node-window.400", (c, x) => c.openNodeWindow(L, b(x).timeout_seconds)],
  ["controller-node-window-close.204", (c) => c.closeNodeWindow(L)],
  ["controller-node-attributes.200", (c, x) => c.readAttributes(L, b(x).paths)],
  ["controller-node-attributes.400", (c, x) => c.readAttributes(L, b(x).paths)],
  ["controller-node-attribute-write.200", (c, x) => c.writeAttribute(L, b(x))],
  ["controller-node-attribute-write.403", (c, x) => c.writeAttribute(L, b(x))],
  ["controller-node-command.200", (c, x) => c.invokeCommand(L, b(x))],
  ["controller-node-acl.200", (c) => c.nodeAcl(L)],
  ["controller-node-acl-set.200", (c, x) => c.setNodeAcl(L, b(x).entries)],
  ["controller-node-acl-set.400", (c, x) => c.setNodeAcl(L, b(x).entries)],
  ["controller-node-bindings.200", (c) => c.nodeBindings(L)],
  ["controller-node-bindings-set.200", (c, x) => c.setNodeBindings(L, b(x).endpoint, b(x).bindings)],
  ["controller-node-icd.200", (c) => c.nodeIcd(S)],
  ["controller-node-icd-register.200", (c, x) => c.registerNodeIcd(S, b(x).allow_multi_admin)],
  ["controller-node-icd-register.409", (c) => c.registerNodeIcd(S)],
  ["controller-node-icd-unregister.200", (c, x) => c.unregisterNodeIcd(S, b(x).force)],
  ["controller-node-fabrics.200", (c) => c.nodeFabrics(L)],
  ["controller-node-fabric-remove.204", (c) => c.removeNodeFabric(L, 2)],
  ["controller-node-fabric-remove.400", (c) => c.removeNodeFabric(L, 1)],
  ["controller-commission.200", (c, x) => c.commission(b(x))],
  ["controller-commission.200-discovered", (c, x) => c.commission(b(x))],
  ["controller-commission.400", (c, x) => c.commission(b(x))],
  ["controller-commission.404", (c, x) => c.commission(b(x))],
  ["controller-commission.502", (c, x) => c.commission(b(x))],
  ["controller-discovery.200", (c) => c.discover({ filter: "long_discriminator", value: 3840 })],
  ["controller-discovery.400", (c) => c.discover({ filter: "vendor" })],
  ["controller-credentials.200", (c) => c.credentials()],
  ["controller-credentials-set.204", (c, x) => c.setCredential(b(x))],
  ["controller-credentials-set.400", (c, x) => c.setCredential(b(x))],
  ["controller-credentials-remove.204", (c) => c.removeCredential("thread", "mesh")],
  ["controller-credentials-remove.404", (c) => c.removeCredential("wifi", "none")],
  ["controller-fabric-label.204", (c, x) => c.setFabricLabel(b(x).label)],
  ["controller-fabric-label.400", (c, x) => c.setFabricLabel(b(x).label)],
  ["controller-updates.200", (c) => c.updates()],
  ["controller-updates-consent.204", (c, x) => c.consentUpdate(b(x))],
  ["controller-updates-check.200", (c) => c.checkUpdates()],
];

describe("controller client against the contract fixtures", () => {
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
    } else if (f.response.status === 204) {
      expect(await result).toBeUndefined();
    } else {
      expect(await result).toEqual(f.response.body);
    }
    expect(ff.sent).toHaveLength(1);
    const sent = ff.sent[0];
    expect(`${sent.method} ${sent.path}`).toBe(`${f.request.method} /api/v1${f.request.path}`);
    expect(sent.headers[TOKEN_HEADER]).toBe(token);
    expect(sent.body).toEqual(f.request.body);
  });

  it("takes the admin token from the controller status of token mode", async () => {
    const ff = fixtureFetch(["controller-status.200", "controller-node-interview.204"]);
    const client = createClient("/api/v1", ff.fetch);
    await client.controllerStatus();
    await client.interviewNode(L);
    expect(ff.sent[1].headers[TOKEN_HEADER]).toBe("test-token");
    expect(client.controllerStreamUrl()).toBe("/api/v1/controller/stream");
  });

  it("answers a path the server does not serve with the not_found problem", async () => {
    const f = fixture("controller-unknown.404");
    const ff = fixtureFetch([]);
    ff.routes.set("GET /api/v1/controller/status", { ...f, request: { method: "GET", path: "/controller/status" } });
    const err = await createClient("/api/v1", ff.fetch)
      .controllerStatus()
      .catch((e: unknown) => e);
    expect((err as ApiError).status).toBe(404);
    expect((err as ApiError).problem?.code).toBe("not_found");
  });
});
