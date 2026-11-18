// Every key a page or kit component uses exists in both catalogues, and
// the catalogues hold the same keys. Mirrors openccu-loom
// assets/ui/src/lib/i18n.catalog-coverage.test.ts at b40b289a.
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { catalogKeys, catalogs, createTranslator, detectLocale, format } from "./index";
import type {
  AclEntry,
  ConnectionState,
  ControllerCapabilities,
  Credential,
  DiagnosticEvent,
  IcdState,
  MdnsFinding,
  PairingStep,
  PendingUpdate,
} from "../api/client";

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
    // Ecosystem.ecosystem has no enum in the contract; go-fabric's values.
    for (const v of ["apple", "google", "amazon", "smartthings", "aqara", "home_assistant", "unknown"]) expect(en.has(`node.diag.ecosystem.${v}`)).toBe(true);
    for (const v of ["status", "pair", "fabrics", "selection", "endpoints", "diagnostics"]) expect(en.has(`node.tab.${v}`)).toBe(true);
    for (const v of ["interviewed", "enabled", "disabled_done", "removed"]) expect(en.has(`controller.nodes.${v}`)).toBe(true);
  });

  // Each contract enum as a Record over its type: a value the contract
  // adds fails the type check here (npm run check) until the record, and
  // with it the catalogue, names it (UR-P1-test-173).
  it("name every value of the contract's enumerations the pages translate", () => {
    const kinds: Record<DiagnosticEvent["kind"], true> = { pairing: true, session: true, discovery: true, reachability: true, icd: true, ota: true, camera: true };
    const severities: Record<MdnsFinding["severity"], true> = { error: true, warning: true };
    const connections: Record<ConnectionState, true> = { connected: true, disconnected: true, reconnecting: true, waiting_for_discovery: true };
    const privileges: Record<AclEntry["privilege"], true> = { view: true, proxy_view: true, operate: true, manage: true, administer: true };
    const modes: Record<AclEntry["auth_mode"], true> = { pase: true, case: true, group: true };
    const icdModes: Record<NonNullable<IcdState["operating_mode"]>, true> = { sit: true, lit: true };
    const updates: Record<PendingUpdate["status"], true> = { queued: true, in_progress: true, stalled: true };
    const credentials: Record<Credential["kind"], true> = { wifi: true, thread: true };
    const caps: Record<keyof ControllerCapabilities, true> = { ota: true, credentials: true, ble: true, diagnostics: true };
    const phases: Record<PairingStep["phase"], true> = { started: true, succeeded: true, failed: true, skipped: true };
    const keys = [
      ...Object.keys(kinds).map((v) => `node.diag.kind_${v}`),
      ...Object.keys(severities).map((v) => `node.diag.severity.${v}`),
      ...Object.keys(connections).map((v) => `controller.connection.${v}`),
      ...Object.keys(privileges).map((v) => `controller.acl.privilege.${v}`),
      ...Object.keys(modes).map((v) => `controller.acl.mode.${v}`),
      ...Object.keys(icdModes).map((v) => `controller.icd.mode.${v}`),
      ...Object.keys(updates).map((v) => `controller.updates.status.${v}`),
      ...Object.keys(credentials).map((v) => `controller.credential.kind_${v}`),
      ...Object.keys(caps).map((v) => `controller.settings.cap.${v}`),
      ...Object.keys(phases).map((v) => `controller.commission.step.${v}`),
    ];
    expect(keys.filter((k) => !en.has(k) || !de.has(k))).toEqual([]);
  });

  it("use the same placeholders in English and German", () => {
    const names = (s: string) => [...s.matchAll(/\{(\w+)\}/g)].map((m) => m[1]).sort().join(",");
    const differ = Object.keys(catalogs.en).filter((k) => k in catalogs.de && names(catalogs.en[k]) !== names(catalogs.de[k]));
    expect(differ).toEqual([]);
  });

  it("name every tab of the controller's node page and every node action", () => {
    for (const v of ["overview", "endpoints", "attributes", "acl", "bindings", "window", "icd", "fabrics", "updates"]) expect(en.has(`controller.node.tab.${v}`)).toBe(true);
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
