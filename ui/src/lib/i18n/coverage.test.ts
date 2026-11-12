// Every key a page or kit component uses exists in both catalogues, and
// the catalogues hold the same keys. Mirrors openccu-loom
// assets/ui/src/lib/i18n.catalog-coverage.test.ts at b40b289a.
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { catalogKeys, createTranslator, detectLocale, format } from "./index";

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

const code = sources(SRC).map((f) => fs.readFileSync(f, "utf8")).join("\n");
const en = new Set(catalogKeys("en"));
const de = new Set(catalogKeys("de"));

describe("catalogues", () => {
  it("hold the same keys in English and German", () => {
    expect([...en].filter((k) => !de.has(k))).toEqual([]);
    expect([...de].filter((k) => !en.has(k))).toEqual([]);
  });

  it("define every literal key the code passes to t()", () => {
    const used = [...code.matchAll(/\bt\(\s*"([a-z0-9_.]+)"/g)].map((m) => m[1]);
    expect(used.length).toBeGreaterThan(100);
    expect([...new Set(used)].filter((k) => !en.has(k))).toEqual([]);
  });

  it("define every key a template-literal t() call can build", () => {
    const prefixes = [...code.matchAll(/\bt\(\s*`([a-z0-9_.]+)\$\{/g)].map((m) => m[1]);
    expect(prefixes.length).toBeGreaterThan(0);
    for (const p of new Set(prefixes)) expect([...en].some((k) => k.startsWith(p)), p).toBe(true);
    // The enumerations the server sends, each with a label.
    for (const v of ["apple", "google", "amazon", "smartthings", "aqara", "home_assistant", "unknown"]) expect(en.has(`node.diag.ecosystem.${v}`)).toBe(true);
    for (const v of ["pairing", "session", "discovery"]) expect(en.has(`node.diag.kind_${v}`)).toBe(true);
    for (const v of ["error", "warning"]) expect(en.has(`node.diag.severity.${v}`)).toBe(true);
    for (const v of ["status", "pair", "fabrics", "selection", "endpoints", "diagnostics"]) expect(en.has(`node.tab.${v}`)).toBe(true);
    // The controller family's enumerations, as the contract names them.
    for (const v of ["connected", "disconnected", "reconnecting", "waiting_for_discovery"]) expect(en.has(`controller.connection.${v}`)).toBe(true);
    for (const v of ["view", "proxy_view", "operate", "manage", "administer"]) expect(en.has(`controller.acl.privilege.${v}`)).toBe(true);
    for (const v of ["pase", "case", "group"]) expect(en.has(`controller.acl.mode.${v}`)).toBe(true);
    for (const v of ["sit", "lit"]) expect(en.has(`controller.icd.mode.${v}`)).toBe(true);
    for (const v of ["queued", "in_progress", "stalled"]) expect(en.has(`controller.updates.status.${v}`)).toBe(true);
    for (const v of ["wifi", "thread"]) expect(en.has(`controller.credential.kind_${v}`)).toBe(true);
    for (const v of ["ota", "credentials", "ble"]) expect(en.has(`controller.settings.cap.${v}`)).toBe(true);
    for (const v of ["overview", "endpoints", "attributes", "acl", "bindings", "window", "icd", "fabrics", "updates"]) expect(en.has(`controller.node.tab.${v}`)).toBe(true);
    for (const v of ["interviewed", "enabled", "disabled_done", "removed"]) expect(en.has(`controller.nodes.${v}`)).toBe(true);
  });
});

describe("translator", () => {
  it("falls back from the locale to English to the key, and formats placeholders", () => {
    let loc = "de";
    const t = createTranslator(() => loc, { de: { "host.only": "nur hier" } });
    expect(t("node.tab.pair")).toBe("Koppeln");
    expect(t("host.only")).toBe("nur hier");
    expect(t("no.such.key")).toBe("no.such.key");
    loc = "fr";
    expect(t("node.tab.pair")).toBe("Pair");
    expect(t("node.status.fabrics", { count: 2 })).toBe("2 fabrics");
    expect(format("{a} {b}", { a: 1 })).toBe("1 {b}");
  });

  it("detects German from the first browser language only", () => {
    expect(detectLocale(["de-DE", "en"])).toBe("de");
    expect(detectLocale(["en-US", "de"])).toBe("en");
    expect(detectLocale([])).toBe("en");
  });
});
