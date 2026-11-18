// The client's error paths (review phase 1b): an answer that is not JSON,
// an empty one, a problem of the wrong shape.
import { describe, expect, it } from "vitest";
import { ApiError, createClient } from "./client";

const answer = (status: number, text: string | null, type = "application/json") =>
  createClient("/api/v1", async () => new Response(text, { status, headers: { "Content-Type": type } }));

describe("UR-P1-err-106: answers that are not JSON", () => {
  it("rejects a proxy's HTML error page with an ApiError of its status", async () => {
    const err = await answer(502, "<html>502</html>", "text/html").status().catch((e: unknown) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect((err as ApiError).status).toBe(502);
    expect((err as ApiError).message).toBe("GET /node/status: 502");
  });

  it("rejects a 200 whose body is not JSON with an ApiError", async () => {
    const err = await answer(200, "<html>ok</html>", "text/html").fabrics().catch((e: unknown) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect((err as ApiError).status).toBe(200);
  });

  it("rejects a 200 without a body where one is required, not with a TypeError", async () => {
    const err = await answer(200, null).status().catch((e: unknown) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect((err as ApiError).status).toBe(200);
  });

  it("takes a problem only when code, title and detail are strings", async () => {
    const err = (await answer(400, JSON.stringify({ code: "x", title: "T", detail: { a: 1 } }))
      .status()
      .catch((e: unknown) => e)) as ApiError;
    expect(err).toBeInstanceOf(ApiError);
    expect(err.problem).toBeUndefined();
    expect(err.message).toBe("GET /node/status: 400");
    const ok = (await answer(409, JSON.stringify({ code: "busy", title: "Busy", detail: "a window is open", status: 409, type: "x" }))
      .status()
      .catch((e: unknown) => e)) as ApiError;
    expect(ok.problem?.code).toBe("busy");
    expect(ok.message).toBe("a window is open");
  });
});

describe("UR-P1-err-108: a rotated token", () => {
  function server() {
    let token = "a";
    const seen: string[] = [];
    const fetchImpl = async (url: string, init?: RequestInit) => {
      const h = (init?.headers ?? {}) as Record<string, string>;
      seen.push(`${init?.method ?? "GET"} ${url.replace("/api/v1", "")} ${h["X-Fabric-UI-Token"] ?? "-"}`);
      if (url.endsWith("/node/status")) return new Response(JSON.stringify({ enabled: true, csrf_token: token }), { status: 200 });
      if (h["X-Fabric-UI-Token"] !== token) {
        return new Response(JSON.stringify({ code: "forbidden", title: "Admin rights required", status: 403, type: "x" }), { status: 403 });
      }
      return new Response(null, { status: 204 });
    };
    return { fetchImpl, seen, restart: () => (token = "b") };
  }

  it("reloads the status once on a 403 forbidden and repeats the call with the new token", async () => {
    const s = server();
    const client = createClient("/api/v1", s.fetchImpl);
    await client.status();
    s.restart();
    await expect(client.closeWindow()).resolves.toBeUndefined();
    expect(s.seen).toEqual([
      "GET /node/status -",
      "POST /node/pairing/window/close a",
      "GET /node/status -",
      "POST /node/pairing/window/close b",
    ]);
  });

  it("does not repeat when the status brings no other token", async () => {
    const seen: string[] = [];
    const client = createClient("/api/v1", async (url, init) => {
      seen.push(`${init?.method ?? "GET"} ${url}`);
      if (url.endsWith("/node/status")) return new Response(JSON.stringify({ enabled: true, csrf_token: "a" }), { status: 200 });
      return new Response(JSON.stringify({ code: "forbidden", title: "Admin rights required", status: 403, type: "x" }), { status: 403 });
    });
    client.setToken("a");
    await expect(client.closeWindow()).rejects.toMatchObject({ status: 403 });
    expect(seen.filter((l) => l.startsWith("POST"))).toHaveLength(1);
  });
});

describe("UR-P1-web-115: keys and ids are checked where the library takes them", () => {
  it("refuses a node key outside the contract's pattern", () => {
    const client = createClient("/api/v1", async () => new Response(null, { status: 204 }));
    expect(() => client.forNode("..")).toThrow(RangeError);
    expect(() => client.forNode("a/b")).toThrow(RangeError);
    expect(client.forNode("kitchen-2").nodeKey).toBe("kitchen-2");
  });

  it("refuses a controller node id that is a dot segment, before sending", async () => {
    const seen: string[] = [];
    const client = createClient("/api/v1", async (url) => {
      seen.push(url);
      return new Response("{}", { status: 200 });
    });
    await expect(client.controllerNode("..")).rejects.toThrow(RangeError);
    await expect(client.pingNode(".")).rejects.toThrow(RangeError);
    await expect(client.controllerNode("")).rejects.toThrow(RangeError);
    // "../x" is encoded to one segment the server refuses itself.
    await expect(client.controllerNode("../x")).resolves.toEqual({});
    await expect(client.controllerNode("0x11")).resolves.toEqual({});
    expect(seen).toEqual(["/api/v1/controller/nodes/..%2Fx", "/api/v1/controller/nodes/0x11"]);
  });
});
