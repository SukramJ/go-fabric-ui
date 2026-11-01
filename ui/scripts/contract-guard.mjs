// The contract guard: compares contract/openapi.yaml with the copy of the
// last release (contract/previous/openapi.yaml) and fails on a removed
// path, method, response status, schema or response field, and on a
// request field that became required. A minor version only adds. A
// planned operation keeps its path and method; its placeholder responses
// are no promise. It also fails on a response schema that forbids unknown
// fields (closedResponseSchemas): a client generated from such a schema
// rejects the whole answer once a field is added, which a minor may do.
//
//   node scripts/contract-guard.mjs <previous.yaml> <current.yaml>
import fs from "node:fs";
import { fileURLToPath } from "node:url";
import { load as loadYAML } from "js-yaml";

const METHODS = ["get", "put", "post", "delete", "patch", "head", "options", "trace"];

/** The breaking differences from `prev` to `cur`, as sentences. */
export function diffContracts(prev, cur) {
  const out = [];
  const pPaths = prev.paths ?? {};
  const cPaths = cur.paths ?? {};
  for (const [path, item] of Object.entries(pPaths)) {
    const now = cPaths[path];
    if (!now) {
      out.push(`path ${path} removed`);
      continue;
    }
    for (const m of METHODS) {
      if (!item[m]) continue;
      if (!now[m]) {
        out.push(`${m.toUpperCase()} ${path} removed`);
        continue;
      }
      // A planned operation (`x-status: planned`) promised no answer yet:
      // its placeholder responses may go when it is served.
      if (item[m]["x-status"] === "planned") continue;
      for (const status of Object.keys(item[m].responses ?? {})) {
        if (!(now[m].responses ?? {})[status]) out.push(`${m.toUpperCase()} ${path}: response ${status} removed`);
      }
    }
  }
  const pSchemas = prev.components?.schemas ?? {};
  const cSchemas = cur.components?.schemas ?? {};
  for (const [name, schema] of Object.entries(pSchemas)) {
    const now = cSchemas[name];
    if (!now) {
      out.push(`schema ${name} removed`);
      continue;
    }
    for (const prop of Object.keys(schema.properties ?? {})) {
      if (!(now.properties ?? {})[prop]) out.push(`schema ${name}: field ${prop} removed`);
    }
    const wasRequired = new Set(schema.required ?? []);
    for (const req of now.required ?? []) {
      // A response field that becomes required is a promise kept; a
      // request field that becomes required breaks an old client. The
      // guard cannot tell the two apart, so it flags every new requirement
      // on a schema that existed — a reviewer accepts the response case.
      if (!wasRequired.has(req) && (schema.properties ?? {})[req]) out.push(`schema ${name}: field ${req} became required`);
    }
  }
  return out;
}

/** The node a `$ref` names (`#/components/...`), or the node itself. */
function resolve(doc, node) {
  while (node && typeof node === "object" && node.$ref) {
    node = node.$ref.replace(/^#\//, "").split("/").reduce((o, k) => o?.[k], doc);
  }
  return node;
}

/** The body schema of one operation's response; undefined without a body. */
function responseSchema(doc, pathItem, method, status) {
  const content = resolve(doc, pathItem?.[method]?.responses?.[status])?.content ?? {};
  return (content["application/json"] ?? Object.values(content)[0])?.schema;
}

/**
 * The schemas a client decodes a response through that set
 * `additionalProperties: false`, sorted; a named schema by its name, an
 * inline one by the operation that answers it. A generated client turns the
 * keyword into a strict model (pydantic's extra="forbid", for example), so a
 * field added to the response — additive by this guard, a minor by the
 * changelog — makes every installed client reject the whole answer. The
 * strictness is sound only where this daemon decodes, the request body
 * (fabricui decodeJSON rejects unknown fields); issue #27. Reachability is
 * transitive: an array's item type carries the same consequence.
 */
export function closedResponseSchemas(doc) {
  const out = new Set();
  const seen = new Set();
  const walk = (node, via) => {
    if (!node || typeof node !== "object") return;
    if (node.$ref) via = `schema ${node.$ref.split("/").pop()}`;
    const s = resolve(doc, node);
    if (!s || seen.has(s)) return;
    seen.add(s);
    if (s.additionalProperties === false) out.add(via);
    for (const p of Object.values(s.properties ?? {})) walk(p, via);
    walk(s.items, via);
    if (typeof s.additionalProperties === "object") walk(s.additionalProperties, via);
  };
  for (const [path, item] of Object.entries(doc.paths ?? {})) {
    for (const m of METHODS) {
      for (const status of Object.keys(item[m]?.responses ?? {})) {
        walk(responseSchema(doc, item, m, status), `${m.toUpperCase()} ${path}: response ${status}`);
      }
    }
  }
  return [...out].sort();
}

/**
 * The fields of `value` that `schema` does not name, as sentences — the
 * drift check between an encoder and the contract, run on the recorded
 * fixtures at test time instead of on every client's wire. A schema without
 * `properties` (a free object) and one with an `additionalProperties`
 * schema judge nothing; the contract composes no allOf/anyOf/oneOf.
 */
export function unknownFields(doc, schema, value, at = "body") {
  const out = [];
  const walk = (node, v, at) => {
    const s = resolve(doc, node);
    if (!s || v === null || typeof v !== "object") return;
    if (Array.isArray(v)) {
      v.forEach((item, i) => walk(s.items, item, `${at}[${i}]`));
      return;
    }
    const props = s.properties;
    for (const [k, item] of Object.entries(v)) {
      if (props?.[k]) walk(props[k], item, `${at}.${k}`);
      else if (typeof s.additionalProperties === "object") walk(s.additionalProperties, item, `${at}.${k}`);
      else if (props && s.additionalProperties === undefined) out.push(`${at}: field ${k} is not in the contract`);
    }
  };
  walk(schema, value, at);
  return out;
}

/** unknownFields for one contract fixture's response against the route's schema. */
export function fixtureUnknownFields(doc, fixture) {
  const [method, path] = fixture.route.split(" ");
  const schema = responseSchema(doc, doc.paths?.[path], method.toLowerCase(), fixture.response.status);
  return schema && fixture.response.body !== undefined ? unknownFields(doc, schema, fixture.response.body) : [];
}

export function load(file) {
  return loadYAML(fs.readFileSync(file, "utf8"));
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const [prev, cur] = process.argv.slice(2);
  if (!prev || !cur) {
    console.error("usage: contract-guard.mjs <previous.yaml> <current.yaml>");
    process.exit(2);
  }
  const current = load(cur);
  const breaks = diffContracts(load(prev), current);
  if (breaks.length) {
    console.error(`contract guard: ${breaks.length} breaking change(s) against ${prev}:`);
    for (const b of breaks) console.error(`  - ${b}`);
    process.exit(1);
  }
  const closed = closedResponseSchemas(current);
  if (closed.length) {
    console.error(`contract guard: ${closed.length} response schema(s) in ${cur} set additionalProperties: false — a generated client decodes them strictly and the next added field breaks it; keep the keyword for request bodies only:`);
    for (const c of closed) console.error(`  - ${c}`);
    process.exit(1);
  }
  console.log(`contract guard: ${cur} keeps every path, method, response and field of ${prev}; every response schema tolerates unknown fields`);
}
