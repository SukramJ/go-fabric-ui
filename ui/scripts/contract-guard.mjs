// The contract guard: compares contract/openapi.yaml with the copy of the
// last release (contract/previous/openapi.yaml) and fails on a removed
// path, method, response status, schema or response field, on a request
// field that became required, on a response field that stopped being
// required, on a changed type, and on a request parameter or field that
// got narrower (pattern, maximum, minimum, length, enum). A minor version
// only adds; a break the release makes on purpose is listed, sentence for
// sentence, in the contract's `info.x-accepted-breaks` and named in
// docs/contract.md — an entry the guard no longer reports fails too, so the
// list empties itself at the next copy to previous/. The baseline is
// checked as well: previous/ is byte for byte the contract of the newest
// release tag, unless the contract has not moved since that tag. A
// planned operation keeps its path and method; its placeholder responses
// are no promise. It also fails on a response schema that forbids unknown
// fields (closedResponseSchemas): a client generated from such a schema
// rejects the whole answer once a field is added, which a minor may do.
//
//   node scripts/contract-guard.mjs <previous.yaml> <current.yaml>
import fs from "node:fs";
import { execFileSync } from "node:child_process";
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
  const responseNamed = reachableSchemas(prev, "response");
  const requestNamed = reachableSchemas(prev, "request");
  for (const [key, { param, at }] of parameterIndex(prev)) {
    const now = parameterIndex(cur).get(key);
    if (now) compareNode(resolve(prev, param.schema), resolve(cur, now.param.schema), at, true, out);
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
    const isRequired = new Set(now.required ?? []);
    if (responseNamed.has(name)) {
      // A client generated from the old contract reads the field as always
      // present (a non-optional type); a response may no longer promise it.
      for (const req of wasRequired) {
        if (!isRequired.has(req) && (now.properties ?? {})[req]) out.push(`schema ${name}: field ${req} no longer required`);
      }
    }
    compareNode(schema, now, `schema ${name}`, requestNamed.has(name), out);
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

/** The type a schema node states: its `type` (a list sorted) or its `$ref`. */
function typeOf(node) {
  if (!node || typeof node !== "object") return undefined;
  if (node.$ref) return node.$ref;
  if (node.type === undefined) return undefined;
  return Array.isArray(node.type) ? [...node.type].sort().join("|") : node.type;
}

/** The keywords by which `now` admits fewer values than `was`. */
export function narrowed(was, now) {
  const out = [];
  const upper = ["maximum", "exclusiveMaximum", "maxLength", "maxItems"];
  const lower = ["minimum", "exclusiveMinimum", "minLength", "minItems"];
  for (const k of upper) if (now[k] !== undefined && (was[k] === undefined || now[k] < was[k])) out.push(k);
  for (const k of lower) if (now[k] !== undefined && (was[k] === undefined || now[k] > was[k])) out.push(k);
  if (now.pattern !== undefined && now.pattern !== was.pattern) out.push("pattern");
  if (now.format !== undefined && now.format !== was.format) out.push("format");
  if (Array.isArray(now.enum) && (!Array.isArray(was.enum) || was.enum.some((v) => !now.enum.includes(v)))) out.push("enum");
  return out;
}

/**
 * A changed type on a node and its inline properties and items, and — where
 * a request reaches it (`request`) — a narrower value range. A `$ref` is
 * judged by its own schema's entry, not followed here.
 */
function compareNode(was, now, at, request, out) {
  if (!was || !now || typeof was !== "object" || typeof now !== "object") return;
  const t0 = typeOf(was);
  const t1 = typeOf(now);
  if (t0 !== undefined && t0 !== t1) {
    out.push(`${at}: type changed from ${t0} to ${t1 ?? "none"}`);
    return;
  }
  if (was.$ref) return;
  if (request) {
    const n = narrowed(was, now);
    if (n.length) out.push(`${at}: narrower (${n.join(", ")})`);
  }
  for (const [prop, p] of Object.entries(was.properties ?? {})) {
    const q = (now.properties ?? {})[prop];
    if (q) compareNode(p, q, `${at}: field ${prop}`, request, out);
  }
  if (was.items && now.items) compareNode(was.items, now.items, `${at}[]`, request, out);
}

/** Every parameter, component and inline, keyed so the next version finds it again. */
function parameterIndex(doc) {
  const out = new Map();
  for (const [name, p] of Object.entries(doc.components?.parameters ?? {})) {
    out.set(`#${name}`, { param: p, at: `parameter ${name}` });
  }
  for (const [path, item] of Object.entries(doc.paths ?? {})) {
    for (const m of METHODS) {
      for (const p of [...(item.parameters ?? []), ...(item[m]?.parameters ?? [])]) {
        if (p.$ref || !item[m]) continue;
        out.set(`${m} ${path} ${p.in} ${p.name}`, { param: p, at: `${m.toUpperCase()} ${path}: parameter ${p.name}` });
      }
    }
  }
  return out;
}

/**
 * The names of the component schemas a response (`side` "response") or a
 * request body or parameter (`side` "request") reaches, transitively.
 */
export function reachableSchemas(doc, side) {
  const names = new Set();
  const seen = new Set();
  const walk = (node) => {
    if (!node || typeof node !== "object") return;
    if (node.$ref?.startsWith("#/components/schemas/")) names.add(node.$ref.split("/").pop());
    const s = resolve(doc, node);
    if (!s || seen.has(s)) return;
    seen.add(s);
    for (const p of Object.values(s.properties ?? {})) walk(p);
    walk(s.items);
    if (typeof s.additionalProperties === "object") walk(s.additionalProperties);
    for (const k of ["allOf", "anyOf", "oneOf"]) for (const x of s[k] ?? []) walk(x);
  };
  for (const item of Object.values(doc.paths ?? {})) {
    for (const m of METHODS) {
      if (!item[m]) continue;
      if (side === "response") {
        for (const status of Object.keys(item[m].responses ?? {})) walk(responseSchema(doc, item, m, status));
        continue;
      }
      for (const c of Object.values(resolve(doc, item[m].requestBody)?.content ?? {})) walk(c.schema);
      for (const p of [...(item.parameters ?? []), ...(item[m].parameters ?? [])]) walk(resolve(doc, p)?.schema);
    }
  }
  return names;
}

/**
 * The guard's verdict on `breaks` given the contract's accepted list: the
 * breaks not accepted, and the accepted entries no break matches (stale).
 */
export function judgeBreaks(breaks, accepted = []) {
  const ok = new Set(accepted);
  const found = new Set(breaks);
  return { unaccepted: breaks.filter((b) => !ok.has(b)), stale: accepted.filter((a) => !found.has(a)) };
}

/**
 * Why previous/ is not the baseline, or undefined: it must be the contract
 * the newest release tag shipped, byte for byte, once the contract has moved
 * past that tag. Before the first change after a release the older copy may
 * stay; the first contract change of the next version copies the tag's.
 */
export function baselineProblem({ previous, current, tagged, tag }) {
  if (current === tagged || previous === tagged) return undefined;
  return `contract/previous/openapi.yaml is not the contract of ${tag}; copy it: git show ${tag}:contract/openapi.yaml > contract/previous/openapi.yaml`;
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
  const top = execFileSync("git", ["rev-parse", "--show-toplevel"], { encoding: "utf8" }).trim();
  const tag = execFileSync("git", ["describe", "--tags", "--abbrev=0"], { cwd: top, encoding: "utf8" }).trim();
  const tagged = execFileSync("git", ["show", `${tag}:contract/openapi.yaml`], { cwd: top, encoding: "utf8" });
  const baseline = baselineProblem({ previous: fs.readFileSync(prev, "utf8"), current: fs.readFileSync(cur, "utf8"), tagged, tag });
  if (baseline) {
    console.error(`contract guard: ${baseline}`);
    process.exit(1);
  }
  const accepted = current.info?.["x-accepted-breaks"] ?? [];
  const { unaccepted, stale } = judgeBreaks(diffContracts(load(prev), current), accepted);
  if (unaccepted.length) {
    console.error(`contract guard: ${unaccepted.length} breaking change(s) against ${prev} not in info.x-accepted-breaks:`);
    for (const b of unaccepted) console.error(`  - ${b}`);
    process.exit(1);
  }
  if (stale.length) {
    console.error(`contract guard: ${stale.length} entr(ies) of info.x-accepted-breaks the guard no longer reports; remove them:`);
    for (const b of stale) console.error(`  - ${b}`);
    process.exit(1);
  }
  const closed = closedResponseSchemas(current);
  if (closed.length) {
    console.error(`contract guard: ${closed.length} response schema(s) in ${cur} set additionalProperties: false — a generated client decodes them strictly and the next added field breaks it; keep the keyword for request bodies only:`);
    for (const c of closed) console.error(`  - ${c}`);
    process.exit(1);
  }
  console.log(`contract guard: ${cur} keeps every path, method, response, field and type of ${prev} (${accepted.length} accepted break(s)); previous/ is the baseline of ${tag}; every response schema tolerates unknown fields`);
}
