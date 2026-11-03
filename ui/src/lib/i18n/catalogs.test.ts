// The catalogues cover what the pages ask for: English and German name the
// same keys, every literal key a page, the kit or the shell translates is
// in the English catalogue, and the topology graph's kinds, states and the
// detail keys GET /node/topology sends have their names.
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { body } from "../../test/support";
import type { TopologyGraph } from "../api/client";
import { TOPOLOGY_KINDS, TOPOLOGY_STATES } from "../kit/graph-layout";
import { catalogKeys } from ".";

const SRC = fileURLToPath(new URL("../../", import.meta.url));

function sources(dir: string): string[] {
  const out: string[] = [];
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) out.push(...sources(p));
    else if (/\.(svelte|ts)$/.test(e.name) && !e.name.endsWith(".test.ts") && !e.name.endsWith(".gen.ts")) out.push(p);
  }
  return out;
}

describe("i18n catalogues", () => {
  const en = new Set(catalogKeys("en"));

  it("name the same keys in English and German", () => {
    expect([...catalogKeys("de")].sort()).toEqual([...en].sort());
  });

  it("hold every literal key the sources translate", () => {
    const missing = new Set<string>();
    for (const f of [...sources(path.join(SRC, "lib")), ...sources(path.join(SRC, "shell"))]) {
      for (const m of fs.readFileSync(f, "utf8").matchAll(/\bt\(\s*"([a-z_]+(?:\.[a-z0-9_]+)+)"/g)) {
        if (!en.has(m[1])) missing.add(`${path.relative(SRC, f)}: ${m[1]}`);
      }
    }
    expect([...missing]).toEqual([]);
  });

  it("name every topology kind and state and the node graph's detail keys", () => {
    const keys = [...TOPOLOGY_KINDS.map((k) => `graph.kind.${k}`), ...TOPOLOGY_STATES.map((s) => `graph.state.${s}`)];
    for (const n of body<TopologyGraph>("node-topology.200").nodes) for (const d of n.details) keys.push(`graph.detail.${d.key}`);
    expect(keys.filter((k) => !en.has(k))).toEqual([]);
  });
});
