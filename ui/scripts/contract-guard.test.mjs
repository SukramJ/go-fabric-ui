import fs from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { closedResponseSchemas, diffContracts, fixtureUnknownFields, load, unknownFields } from "./contract-guard.mjs";

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

  it("keeps every response schema open to unknown fields (issue #27)", () => {
    expect(closedResponseSchemas(load(contract("openapi.yaml")))).toEqual([]);
  });

  it("names a closed response schema, directly or through an item type, and not a closed request body", () => {
    const doc = {
      paths: {
        "/a": {
          post: {
            requestBody: { content: { "application/json": { schema: { $ref: "#/components/schemas/In" } } } },
            responses: {
              200: { content: { "application/json": { schema: { $ref: "#/components/schemas/Out" } } } },
              400: { $ref: "#/components/responses/Problem" },
            },
          },
        },
      },
      components: {
        responses: { Problem: { content: { "application/problem+json": { schema: { type: "object", additionalProperties: false } } } } },
        schemas: {
          In: { type: "object", additionalProperties: false, properties: { x: {} } },
          Out: { type: "object", properties: { items: { type: "array", items: { $ref: "#/components/schemas/Item" } } } },
          Item: { type: "object", additionalProperties: false, properties: { y: {} } },
        },
      },
    };
    expect(closedResponseSchemas(doc)).toEqual(["POST /a: response 400", "schema Item"]);
  });

  it("records every fixture's response with fields the contract names", () => {
    const doc = load(contract("openapi.yaml"));
    const dir = contract("fixtures");
    const names = fs.readdirSync(dir).filter((f) => f.endsWith(".json"));
    expect(names.length).toBeGreaterThan(80);
    for (const f of names) {
      const fixture = JSON.parse(fs.readFileSync(`${dir}/${f}`, "utf8"));
      expect(fixtureUnknownFields(doc, fixture), f).toEqual([]);
    }
  });

  it("reports a fixture field the contract does not name, at its path", () => {
    const doc = load(contract("openapi.yaml"));
    const fixture = JSON.parse(fs.readFileSync(`${contract("fixtures")}/node-surface.200.json`, "utf8"));
    fixture.response.body.wire = 1;
    fixture.response.body.changed[0].changes[0].why = "x";
    expect(fixtureUnknownFields(doc, fixture)).toEqual([
      "body.changed[0].changes[0]: field why is not in the contract",
      "body: field wire is not in the contract",
    ]);
    // A free object (no properties) and a map (additionalProperties schema) judge nothing.
    expect(unknownFields({}, { type: "object" }, { any: 1 })).toEqual([]);
    expect(unknownFields({}, { type: "object", additionalProperties: { type: "string" } }, { any: "1" })).toEqual([]);
  });
});
