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
