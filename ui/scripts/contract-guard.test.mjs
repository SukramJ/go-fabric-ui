import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { diffContracts, load } from "./contract-guard.mjs";

const contract = (p) => fileURLToPath(new URL(`../../contract/${p}`, import.meta.url));

const base = {
  paths: { "/a": { get: { responses: { 200: {}, 503: {} } }, post: { responses: { 204: {} } } } },
  components: { schemas: { A: { required: ["x"], properties: { x: {}, y: {} } } } },
};

describe("contract guard", () => {
  it("passes the current contract against the previous release", () => {
    expect(diffContracts(load(contract("previous/openapi.yaml")), load(contract("openapi.yaml")))).toEqual([]);
  });

  it("passes an addition", () => {
    const cur = structuredClone(base);
    cur.paths["/b"] = { get: { responses: { 200: {} } } };
    cur.components.schemas.A.properties.z = {};
    expect(diffContracts(base, cur)).toEqual([]);
  });

  it("lets a planned operation drop its placeholder responses, not its method", () => {
    const prev = structuredClone(base);
    prev.paths["/p"] = { get: { "x-status": "planned", responses: { 501: {} } } };
    const cur = structuredClone(base);
    cur.paths["/p"] = { get: { responses: { 200: {} } } };
    expect(diffContracts(prev, cur)).toEqual([]);
    cur.paths["/p"] = {};
    expect(diffContracts(prev, cur)).toEqual(["GET /p removed"]);
  });

  it("fails on a removed path, method, response, schema and field, and a new requirement", () => {
    const cur = structuredClone(base);
    delete cur.paths["/a"].post;
    delete cur.paths["/a"].get.responses[503];
    delete cur.components.schemas.A.properties.x;
    cur.components.schemas.A.required = ["y"];
    expect(diffContracts(base, cur)).toEqual([
      "GET /a: response 503 removed",
      "POST /a removed",
      "schema A: field x removed",
      "schema A: field y became required",
    ]);
    expect(diffContracts(base, { paths: {}, components: { schemas: {} } })).toEqual(["path /a removed", "schema A removed"]);
  });
});
