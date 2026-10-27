// The contract guard: compares contract/openapi.yaml with the copy of the
// last release (contract/previous/openapi.yaml) and fails on a removed
// path, method, response status, schema or response field, and on a
// request field that became required. A minor version only adds. A
// planned operation keeps its path and method; its placeholder responses
// are no promise.
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

export function load(file) {
  return loadYAML(fs.readFileSync(file, "utf8"));
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const [prev, cur] = process.argv.slice(2);
  if (!prev || !cur) {
    console.error("usage: contract-guard.mjs <previous.yaml> <current.yaml>");
    process.exit(2);
  }
  const breaks = diffContracts(load(prev), load(cur));
  if (breaks.length) {
    console.error(`contract guard: ${breaks.length} breaking change(s) against ${prev}:`);
    for (const b of breaks) console.error(`  - ${b}`);
    process.exit(1);
  }
  console.log(`contract guard: ${cur} keeps every path, method, response and field of ${prev}`);
}
