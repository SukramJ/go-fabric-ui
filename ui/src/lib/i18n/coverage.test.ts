// Every key a page or kit component uses exists in both catalogues, and
// the catalogues hold the same keys. Mirrors openccu-loom
// assets/ui/src/lib/i18n.catalog-coverage.test.ts at b40b289a.
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { catalogKeys, catalogs, createTranslator, detectLocale, format } from "./index";
import {
  DENIED_CODES,
  type AclEntry,
  type BLEDecision,
  type ConformanceVerdict,
  type ConnectionState,
  type ControllerCapabilities,
  type Credential,
  type DiagnosticEvent,
  type EventRecord,
  type IcdState,
  type MdnsFinding,
  type PairingStep,
  type PendingUpdate,
  type SurfaceDecision,
  type SurfaceElement,
} from "../api/client";
import { TOPOLOGY_KINDS, TOPOLOGY_STATES, TOPOLOGY_STRENGTHS } from "../kit/graph-layout";
import { IM_STATUS } from "../kit/im-status";
import { LEVELS } from "../pages/operation/log";
import { ECOSYSTEM_HINTS } from "../pages/node/ecosystems";

/** The values of a Record over a contract type: a value the contract adds fails the type check. */
const of = <K extends string>(r: Record<K, true>): string[] => Object.keys(r);

// Every prefix a template-literal t() call uses, with the values its
// expression can take (UR-P3-test-33).
const TEMPLATE_KEYS: Record<string, readonly (string | number)[]> = {
  "common.denied.": DENIED_CODES.flatMap((c) => [c, `${c}_hint`]),
  "controller.acl.mode.": of<AclEntry["auth_mode"]>({ pase: true, case: true, group: true, unknown: true }),
  "controller.acl.privilege.": of<AclEntry["privilege"]>({ view: true, proxy_view: true, operate: true, manage: true, administer: true, unknown: true }),
  "controller.commission.ble_skipped.": of<NonNullable<BLEDecision["reason"]>>({ no_central: true, network_only: true, code_without_ble: true, no_network: true }),
  "controller.commission.step.": of<PairingStep["phase"]>({ started: true, succeeded: true, failed: true, skipped: true }),
  "controller.connection.": of<ConnectionState>({ connected: true, disconnected: true, reconnecting: true, waiting_for_discovery: true }),
  "controller.credential.kind_": of<Credential["kind"]>({ wifi: true, thread: true }),
  "controller.icd.mode.": of<NonNullable<IcdState["operating_mode"]>>({ sit: true, lit: true }),
  "controller.inventory.conformance.": of<ConformanceVerdict["status"]>({ unknown: true, incomplete: true, clean: true, violated: true }),
  // Nodes.svelte TABS, Node.svelte TABS, Overview.svelte flags.
  "controller.nodes.tab.": ["list", "network"],
  "controller.node.tab.": ["overview", "endpoints", "inventory", "events", "attributes", "acl", "bindings", "window", "icd", "fabrics", "updates"],
  "controller.overview.flag.": [
    "supports_thread",
    "supports_wifi",
    "supports_ethernet",
    "mains_powered",
    "battery_powered",
    "intermittently_connected",
    "thread_sleepy_end_device",
  ],
  "controller.settings.cap.": of<keyof ControllerCapabilities>({ ota: true, credentials: true, ble: true, diagnostics: true }),
  "controller.updates.status.": of<PendingUpdate["status"]>({ queued: true, in_progress: true, stalled: true }),
  "events.priority.": of<EventRecord["priority"]>({ debug: true, info: true, critical: true }),
  "graph.kind.": TOPOLOGY_KINDS,
  "graph.state.": TOPOLOGY_STATES,
  "graph.strength.": TOPOLOGY_STRENGTHS,
  "inventory.status.": Object.values(IM_STATUS),
  "node.diag.ecosystem.": ["apple", "google", "amazon", "smartthings", "aqara", "home_assistant", "unknown"],
  "node.diag.kind_": of<DiagnosticEvent["kind"]>({ pairing: true, session: true, discovery: true, reachability: true, icd: true, ota: true, camera: true }),
  "node.diag.severity.": of<MdnsFinding["severity"]>({ error: true, warning: true }),
  "node.pair.hints.": ECOSYSTEM_HINTS.map((h) => h.ecosystem),
  // GeneralDiagnostics BootReasonEnum 0–6, as the status page shows them.
  "node.status.boot_reason.": [0, 1, 2, 3, 4, 5, 6],
  // Surface.svelte: the cluster lists' column labels, changeKind's answers.
  "node.surface.": ["col_attributes", "col_accepted", "col_generated", "col_events"],
  "node.surface.change_": ["added", "removed", "changed"],
  "node.surface.kind.": of<SurfaceElement["kind"]>({ "device-type": true, cluster: true, feature: true, attribute: true, command: true, event: true }),
  "node.surface.verdict.": of<SurfaceDecision["verdict"]>({ keep: true, warn: true, remove: true, refuse: true }),
  "node.tab.": ["status", "pair", "fabrics", "selection", "endpoints", "surface", "diagnostics", "topology"],
  "operation.level.": LEVELS,
};

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

  // A template-literal t() call builds its key at run time: every prefix
  // the code uses has an enumeration here — a Record over the contract
  // type, or the constant list the page builds from — and every key it
  // enumerates is in both catalogues. A prefix without one fails, so a
  // new call cannot pass on any key that merely starts with it
  // (UR-P3-test-33).
  it("define every key a template-literal t() call can build", () => {
    const prefixes = new Set([...code.matchAll(/\bt\(\s*`([a-z0-9_.]+)\$\{/g)].map((m) => m[1]));
    expect(prefixes.size).toBeGreaterThan(0);
    const missing = [...prefixes].filter((p) => !(p in TEMPLATE_KEYS));
    expect(missing, "a template-literal prefix without an enumeration in TEMPLATE_KEYS").toEqual([]);
    const keys = Object.entries(TEMPLATE_KEYS).flatMap(([p, vs]) => vs.map((v) => p + v));
    expect(keys.filter((k) => !en.has(k) || !de.has(k))).toEqual([]);
  });

  // Each contract enum as a Record over its type: a value the contract
  // adds fails the type check here (npm run check) until the record, and
  // with it the catalogue, names it (UR-P1-test-173).
  it("name every value of the contract's enumerations the pages translate", () => {
    const kinds: Record<DiagnosticEvent["kind"], true> = { pairing: true, session: true, discovery: true, reachability: true, icd: true, ota: true, camera: true };
    const severities: Record<MdnsFinding["severity"], true> = { error: true, warning: true };
    const connections: Record<ConnectionState, true> = { connected: true, disconnected: true, reconnecting: true, waiting_for_discovery: true };
    const privileges: Record<AclEntry["privilege"], true> = { view: true, proxy_view: true, operate: true, manage: true, administer: true, unknown: true };
    const modes: Record<AclEntry["auth_mode"], true> = { pase: true, case: true, group: true, unknown: true };
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
