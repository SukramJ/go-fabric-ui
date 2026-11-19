// Test support: the contract's golden fixtures as a fetch, a fabric-ui
// context on it, and a translator that echoes keys.
import fs from "node:fs";
import path from "node:path";
import { vi } from "vitest";
import { createClient, type FabricClient } from "../lib/api/client";
import { createFabricUI, type FabricUI, type Confirm } from "../lib/context";
import type { Translate } from "../lib/i18n";

// Tests run from ui/; under happy-dom import.meta.url is not a file URL.
export const FIXTURE_DIR = path.resolve(process.cwd(), "../contract/fixtures") + path.sep;

export type Fixture = {
  route: string;
  request: { method: string; path: string; headers?: Record<string, string>; body?: unknown };
  response: { status: number; body?: unknown };
};

/** One fixture of contract/fixtures by name (without .json). */
export function fixture(name: string): Fixture {
  return JSON.parse(fs.readFileSync(FIXTURE_DIR + name + ".json", "utf8")) as Fixture;
}

/** A copy of a fixture's response body, typed as the caller says. */
export function body<T>(name: string): T {
  return structuredClone(fixture(name).response.body) as T;
}

export type Sent = { method: string; path: string; headers: Record<string, string>; body?: unknown };

/**
 * A fetch answering from fixtures: a request is matched by method and path
 * against the fixtures named, later names winning. `sent` records every
 * request.
 */
export function fixtureFetch(names: string[], base = "/api/v1") {
  const routes = new Map<string, Fixture>();
  for (const n of names) {
    const f = fixture(n);
    routes.set(`${f.request.method} ${base}${f.request.path}`, f);
  }
  const sent: Sent[] = [];
  const fetchImpl = vi.fn(async (input: string, init?: RequestInit) => {
    const method = init?.method ?? "GET";
    const headers = (init?.headers ?? {}) as Record<string, string>;
    sent.push({ method, path: input, headers, body: init?.body ? JSON.parse(String(init.body)) : undefined });
    const f = routes.get(`${method} ${input}`);
    if (!f) return new Response(JSON.stringify({ code: "not_found", title: "no fixture", status: 404, type: "x" }), { status: 404 });
    const text = f.response.body === undefined ? null : JSON.stringify(f.response.body);
    return new Response(text, { status: f.response.status });
  });
  return { fetch: fetchImpl, sent, routes };
}

/** Echoes the key, and the variables as JSON — what loom's page tests assert on. */
export const echoT: Translate = (key, vars) => (vars ? `${key}:${JSON.stringify(vars)}` : key);

export type TestUI = FabricUI & {
  toastSpy: { info: ReturnType<typeof vi.fn>; success: ReturnType<typeof vi.fn>; error: ReturnType<typeof vi.fn> };
  confirmSpy: ReturnType<typeof vi.fn>;
  sent: Sent[];
};

/**
 * A client whose streams open on the global EventSource, which a test
 * stubs; the shipped client reads them through fetch (FabricClient.openStream).
 */
export function createStreamClient(baseUrl: string, fetchImpl: Parameters<typeof createClient>[1]): FabricClient {
  return { ...createClient(baseUrl, fetchImpl), openStream: (u) => new EventSource(u) };
}

/** A fabric-ui context on a client answering from the fixtures. */
export function testUI(
  opts: {
    fixtures?: string[];
    /** Answers by "METHOD /path" (under /api/v1) that replace a fixture's. */
    answers?: Record<string, { status: number; body?: unknown }>;
    client?: FabricClient;
    confirm?: boolean;
    t?: Translate;
    subscribeLog?: FabricUI["subscribeLog"];
  } = {},
): TestUI {
  const ff = fixtureFetch(opts.fixtures ?? []);
  for (const [key, res] of Object.entries(opts.answers ?? {})) {
    const [method, p] = key.split(" ");
    ff.routes.set(`${method} /api/v1${p}`, { route: key, request: { method, path: p }, response: res });
  }
  const client = opts.client ?? createClient("/api/v1", ff.fetch);
  client.setToken("test-token");
  const toastSpy = { info: vi.fn(), success: vi.fn(), error: vi.fn() };
  const confirmSpy = vi.fn(async () => opts.confirm ?? true);
  const ui = createFabricUI({
    client,
    t: opts.t ?? echoT,
    toast: toastSpy,
    confirm: confirmSpy as unknown as Confirm,
    subscribe: () => () => {},
    subscribeController: () => () => {},
    subscribeLog: opts.subscribeLog ?? (() => () => {}),
  });
  return Object.assign(ui, { toastSpy, confirmSpy, sent: ff.sent });
}

/**
 * Collects what Svelte throws while it renders or updates (an
 * each_key_duplicate, say), from the window and from the process; stop()
 * removes the listeners.
 */
export function uncaught(): { errors: string[]; stop: () => void } {
  const errors: string[] = [];
  const first = (v: unknown) => String((v as Error)?.message ?? v).split("\n")[0];
  const onErr = (e: ErrorEvent) => {
    errors.push(first(e.error ?? e.message));
    e.preventDefault();
  };
  const onRej = (e: PromiseRejectionEvent) => {
    errors.push(first(e.reason));
    e.preventDefault();
  };
  const onProc = (e: unknown) => errors.push(first(e));
  window.addEventListener("error", onErr);
  window.addEventListener("unhandledrejection", onRej);
  process.on("uncaughtException", onProc);
  process.on("unhandledRejection", onProc);
  return {
    errors,
    stop: () => {
      window.removeEventListener("error", onErr);
      window.removeEventListener("unhandledrejection", onRej);
      process.off("uncaughtException", onProc);
      process.off("unhandledRejection", onProc);
    },
  };
}
