// The catalogues cover what the pages ask for: English and German name the
// same keys, every literal key a page, the kit or the shell translates is
// in the English catalogue, and the topology graph's kinds, states, link
// strengths and the detail keys GET /node/topology and GET
// /controller/topology send have their names.
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { body } from "../../test/support";
import type { ConformanceVerdict, TopologyGraph } from "../api/client";
import { TOPOLOGY_KINDS, TOPOLOGY_STATES, TOPOLOGY_STRENGTHS } from "../kit/graph-layout";
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

  it("name every tab of the controller's node page and every conformance status", () => {
    const page = fs.readFileSync(path.join(SRC, "lib/pages/controller/Node.svelte"), "utf8");
    const tabs = [...(page.match(/const TABS = \[([^\]]*)\]/)?.[1] ?? "").matchAll(/"([a-z_]+)"/g)].map((m) => m[1]);
    // A Record over the contract's enum: a new status fails the type check here.
    const statuses: Record<ConformanceVerdict["status"], true> = { unknown: true, incomplete: true, clean: true, violated: true };
    expect(tabs).toContain("inventory");
    const keys = [...tabs.map((k) => `controller.node.tab.${k}`), ...Object.keys(statuses).map((s) => `controller.inventory.conformance.${s}`)];
    expect(keys.filter((k) => !en.has(k))).toEqual([]);
  });

  it("name every topology kind, state and strength and both graphs' detail keys", () => {
    const keys = [
      ...TOPOLOGY_KINDS.map((k) => `graph.kind.${k}`),
      ...TOPOLOGY_STATES.map((s) => `graph.state.${s}`),
      ...TOPOLOGY_STRENGTHS.map((s) => `graph.strength.${s}`),
    ];
    for (const name of ["node-topology.200", "controller-topology.200"]) {
      const g = body<TopologyGraph>(name);
      for (const n of g.nodes) keys.push(`graph.kind.${n.kind}`, `graph.state.${n.state}`, ...n.details.map((d) => `graph.detail.${d.key}`));
      for (const e of g.edges) if (e.strength) keys.push(`graph.strength.${e.strength}`);
    }
    expect(keys.filter((k) => !en.has(k))).toEqual([]);
  });
});
