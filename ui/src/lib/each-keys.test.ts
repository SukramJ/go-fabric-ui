// The guard of the class of UR-P1-ui-130, -131, -139 and UR-P2-ui-90: a
// keyed {#each} throws each_key_duplicate on a repeated key, in production
// too, and the tab or page it is on renders nothing. Every keyed each in
// the sources keys by its index, or by a key listed here with the reason
// it cannot repeat. A key taken from data a node, a controller or a host
// reports is not such a reason: the contract makes none of those lists
// unique, so those lists key by position.
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const SRC = fileURLToPath(new URL("../", import.meta.url));

// "file (under src/) :: key" → why the key cannot repeat.
const UNIQUE: Record<string, string> = {
  "lib/pages/operation/Operation.svelte :: rec.seq": "the records are merged by seq (log.ts appendRecord, mergeRecords)",
  "lib/pages/operation/Operation.svelte :: k": "the keys of one object",
  "shell/Toaster.svelte :: e.id": "the toaster's own counter",
  "lib/pages/node/EcosystemHints.svelte :: hint.ecosystem": "ECOSYSTEM_HINTS, a constant of the page",
  "lib/pages/node/EcosystemHints.svelte :: link.key": "ECOSYSTEM_HINTS, a constant of the page",
  "lib/pages/node/Surface.svelte :: label": "the four list names the page writes out",
  "lib/pages/controller/Settings.svelte :: k": "a constant list of the page",
  "lib/pages/controller/node/Overview.svelte :: f": "a constant list of the page, filtered",
  "lib/pages/controller/Network.svelte :: k": "TOPOLOGY_KINDS, filtered",
  "lib/kit/ClusterView.svelte :: key": "the four list names the component writes out",
  "lib/kit/DataTable.svelte :: col.key": "the columns the page defines",
  "lib/kit/DataTable.svelte :: group.key": "the keys of a Map",
  "lib/kit/DataTable.svelte :: r.key": "keyedRows suffixes a repeated row key",
  "lib/kit/Tabs.svelte :: item.key": "the tabs the page defines",
  "lib/kit/Graph.svelte :: e.index": "the edge's position in the input",
  "lib/kit/Graph.svelte :: ln.id": "the layout's nodes, a Map by id",
};

function sources(dir: string): string[] {
  const out: string[] = [];
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) out.push(...sources(p));
    else if (e.name.endsWith(".svelte")) out.push(p);
  }
  return out;
}

/** The text of every {#each …} tag in text, up to its matching brace. */
function eachTags(text: string): string[] {
  const out: string[] = [];
  for (let at = text.indexOf("{#each"); at >= 0; at = text.indexOf("{#each", at + 1)) {
    let depth = 0;
    for (let i = at; i < text.length; i++) {
      if (text[i] === "{") depth++;
      else if (text[i] === "}" && --depth === 0) {
        out.push(text.slice(at + 1, i).replace(/\s+/g, " ").trim());
        break;
      }
    }
  }
  return out;
}

/** The key of a keyed each tag and its index name; null for an unkeyed one. */
function keyOf(tag: string): { key: string; index?: string } | null {
  if (!tag.endsWith(")")) return null;
  // The key is the parenthesised group the tag ends in.
  let depth = 0;
  let open = -1;
  for (let i = tag.length - 1; i >= 0; i--) {
    if (tag[i] === ")") depth++;
    else if (tag[i] === "(" && --depth === 0) {
      open = i;
      break;
    }
  }
  if (open < 0) return null;
  const head = tag.slice(0, open).trimEnd();
  // An unkeyed each may end in a call: the group is a key only after "as <pattern>".
  const as = head.lastIndexOf(" as ");
  if (as < 0 || /[.(]$/.test(head) || !/^ as [[{\w]/.test(head.slice(as))) return null;
  const idx = /,\s*(\w+)$/.exec(head);
  return { key: tag.slice(open + 1, -1).trim(), index: idx?.[1] };
}

describe("keyed each blocks (UR-P2-ui-90)", () => {
  const files = sources(SRC);
  const found = new Set<string>();
  const offenders: string[] = [];
  for (const f of files) {
    const rel = path.relative(SRC, f).split(path.sep).join("/");
    for (const tag of eachTags(fs.readFileSync(f, "utf8"))) {
      const k = keyOf(tag);
      if (!k || (k.index !== undefined && k.key === k.index)) continue;
      const id = `${rel} :: ${k.key}`;
      if (id in UNIQUE) found.add(id);
      else offenders.push(`${id}    (${tag})`);
    }
  }

  it("reads the tags it is meant to read", () => {
    expect(keyOf("#each a as b (b.id)")).toEqual({ key: "b.id", index: undefined });
    expect(keyOf("#each a as b, i (i)")).toEqual({ key: "i", index: "i" });
    expect(keyOf("#each a as { ep, depth }, i (i)")).toEqual({ key: "i", index: "i" });
    expect(keyOf("#each a as [k, v] (k)")).toEqual({ key: "k", index: undefined });
    expect(keyOf("#each (xs as const).filter((k) => ok(k)) as k (k)")).toEqual({ key: "k", index: undefined });
    expect(keyOf("#each a.filter((x) => x) as b")).toBeNull();
    expect(keyOf("#each rows as row (rowKey(row))")).toEqual({ key: "rowKey(row)", index: undefined });
    expect(keyOf('#each fs as f (f.code + (f.type ?? ""))')).toEqual({ key: 'f.code + (f.type ?? "")', index: undefined });
    expect(files.length).toBeGreaterThan(50);
  });

  it("keys every each by its index or by a key that cannot repeat", () => {
    expect(offenders).toEqual([]);
  });

  it("lists no key that is gone from the sources", () => {
    expect(Object.keys(UNIQUE).filter((id) => !found.has(id))).toEqual([]);
  });
});
