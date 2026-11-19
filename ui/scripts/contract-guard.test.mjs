import fs from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { execFileSync } from "node:child_process";
import {
  baselineProblem,
  closedResponseSchemas,
  diffContracts,
  fixtureUnknownFields,
  judgeBreaks,
  load,
  narrowed,
  unknownFields,
} from "./contract-guard.mjs";

const contract = (p) => fileURLToPath(new URL(`../../contract/${p}`, import.meta.url));

const base = {
  paths: { "/a": { get: { responses: { 200: {}, 503: {} } }, post: { responses: { 204: {} } } } },
  components: { schemas: { A: { required: ["x"], properties: { x: {}, y: {} } } } },
};

describe("contract guard", () => {
  it("passes the current contract against the previous release, with its accepted breaks", () => {
    const cur = load(contract("openapi.yaml"));
    const breaks = diffContracts(load(contract("previous/openapi.yaml")), cur);
    expect(judgeBreaks(breaks, cur.info["x-accepted-breaks"])).toEqual({ unaccepted: [], stale: [] });
  });

  it("holds the contract of the newest release tag in previous/ (UR-P3-test-02)", () => {
    const git = (...a) => execFileSync("git", a, { cwd: contract("."), encoding: "utf8" });
    const tag = git("describe", "--tags", "--abbrev=0").trim();
    const tagged = git("show", `${tag}:contract/openapi.yaml`);
    const read = (p) => fs.readFileSync(contract(p), "utf8");
    expect(baselineProblem({ previous: read("previous/openapi.yaml"), current: read("openapi.yaml"), tagged, tag })).toBeUndefined();
  });

  it("judges the baseline: the tag's copy, or an unmoved contract", () => {
    expect(baselineProblem({ previous: "a", current: "b", tagged: "a", tag: "v1" })).toBeUndefined();
    expect(baselineProblem({ previous: "z", current: "a", tagged: "a", tag: "v1" })).toBeUndefined();
    expect(baselineProblem({ previous: "z", current: "b", tagged: "a", tag: "v1" })).toMatch(/not the contract of v1/);
  });

  it("reports a removed released route against the 0.9.0 baseline (UR-P3-test-02)", () => {
    const cur = load(contract("previous/openapi.yaml"));
    delete cur.paths["/controller/diagnostics"];
    expect(diffContracts(load(contract("previous/openapi.yaml")), cur)).toContain("path /controller/diagnostics removed");
  });

  it("splits breaks into unaccepted and stale entries", () => {
    expect(judgeBreaks(["a", "b"], ["b", "c"])).toEqual({ unaccepted: ["a"], stale: ["c"] });
    expect(judgeBreaks([], undefined)).toEqual({ unaccepted: [], stale: [] });
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

  it("reports a response field no longer required, a changed type and a narrower request (UR-P3-test-03)", () => {
    const req = (s) => ({ content: { "application/json": { schema: { $ref: `#/components/schemas/${s}` } } } });
    const prev = {
      paths: {
        "/a/{id}": {
          parameters: [{ $ref: "#/components/parameters/ID" }],
          get: { parameters: [{ name: "n", in: "query", schema: { type: "integer" } }], responses: { 200: req("Out") } },
          post: { requestBody: req("In"), responses: { 204: {} } },
        },
      },
      components: {
        parameters: { ID: { name: "id", in: "path", schema: { type: "string" } } },
        schemas: {
          Out: { required: ["c", "m"], properties: { c: { type: "integer" }, m: { type: "string" }, id: { type: "integer" } } },
          In: { properties: { ep: { type: "integer", minimum: 0 }, kind: { type: "string", enum: ["a", "b"] } } },
        },
      },
    };
    const cur = structuredClone(prev);
    cur.components.schemas.Out.required = ["m"];
    cur.components.schemas.Out.properties.id = { type: "string" };
    cur.components.parameters.ID.schema.pattern = "^[0-9A-F]{1,16}$";
    cur.paths["/a/{id}"].get.parameters[0].schema.maximum = 9;
    cur.components.schemas.In.properties.ep.maximum = 65534;
    cur.components.schemas.In.properties.kind.enum = ["a"];
    expect(diffContracts(prev, cur)).toEqual([
      "parameter ID: narrower (pattern)",
      "GET /a/{id}: parameter n: narrower (maximum)",
      "schema Out: field c no longer required",
      "schema Out: field id: type changed from integer to string",
      "schema In: field ep: narrower (maximum)",
      "schema In: field kind: narrower (enum)",
    ]);
    // A response's range may narrow, a request's may widen, a response field may add an enum value.
    const ok = structuredClone(prev);
    ok.components.schemas.Out.properties.m.pattern = "^x$";
    ok.components.schemas.In.properties.ep.minimum = undefined;
    ok.components.schemas.In.properties.kind.enum = ["a", "b", "c"];
    expect(diffContracts(prev, ok)).toEqual([]);
  });

  it("names the keywords that narrow a value", () => {
    expect(narrowed({ maximum: 10 }, { maximum: 9, minLength: 1 })).toEqual(["maximum", "minLength"]);
    expect(narrowed({ maximum: 10, pattern: "x" }, { maximum: 11, pattern: "x" })).toEqual([]);
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
