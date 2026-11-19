// @vitest-environment happy-dom
// The standalone shell's probes and streams (review phase 1b).
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/svelte";
import { body as fixtureBody, fixtureFetch } from "../test/support";
import App from "./App.svelte";

type Source = { url: string; closed: boolean; addEventListener: () => void; close: () => void };
let sources: Source[] = [];

beforeEach(() => {
  sources = [];
  location.hash = "";
  vi.stubGlobal(
    "EventSource",
    vi.fn(function (url: string) {
      const src: Source = {
        url,
        closed: false,
        addEventListener: () => {},
        close() {
          this.closed = true;
        },
      };
      sources.push(src);
      return src;
    }),
  );
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

/** A fetch over the fixtures under the shell's base; `override` answers first when it returns a response. */
function serve(names: string[], override: (url: string) => Response | undefined = () => undefined) {
  const ff = fixtureFetch(names, "./api/v1");
  const fetchImpl = vi.fn(async (url: string, init?: RequestInit) => override(url) ?? ff.fetch(url, init));
  vi.stubGlobal("fetch", fetchImpl);
  return fetchImpl;
}


// The streams the shell reads through fetch (0.10.0): each request to a
// stream URL answers an event stream that stays open until it is aborted.
type Stream = { url: string; signal?: AbortSignal; headers: Record<string, string> };
let streams: Stream[] = [];
function streamAnswer(url: string, init?: RequestInit): Response | undefined {
  if (!url.endsWith("/stream")) return undefined;
  streams.push({ url, signal: init?.signal ?? undefined, headers: (init?.headers ?? {}) as Record<string, string> });
  return new Response(new ReadableStream(), { status: 200, headers: { "Content-Type": "text/event-stream" } });
}
const live = (suffix: string) => streams.filter((s) => s.url.endsWith(suffix) && !s.signal?.aborted);

describe("UR-P1-err-107: a failed probe is repeated", () => {
  it("shows the controller family once its status answers after a 502", async () => {
    let calls = 0;
    serve(["controller-status.200", "controller-nodes.200"], (url) => {
      if (url.endsWith("/controller/status") && calls++ === 0) return new Response("<html>502</html>", { status: 502 });
      return undefined;
    });
    render(App);
    // The node probe answers 404 (no fixture), the controller's 502: undecided.
    expect(await screen.findByText("The server did not answer. The page tries again by itself.")).toBeInTheDocument();
    await waitFor(() => expect(screen.getByRole("heading", { name: "Controller" })).toBeInTheDocument(), { timeout: 4000 });
    // The 502, the answer without the token, its repeat with it (UR-P3-web-190).
    expect(calls).toBe(3);
  });
});

describe("UR-P1-bound-103: a family's stream only while its pages are shown", () => {
  it("opens the node stream on the node family and the controller stream on the controller family, not both", async () => {
    streams = [];
    serve(["node-status.200", "nodes.200", "controller-status.200", "controller-nodes.200"]);
    const inner = globalThis.fetch as unknown as ReturnType<typeof vi.fn>;
    const base = inner.getMockImplementation() as (url: string, init?: RequestInit) => Promise<Response>;
    inner.mockImplementation(async (url: string, init?: RequestInit) => streamAnswer(url, init) ?? base(url, init));
    render(App);
    await waitFor(() => expect(live("/node/stream")).toHaveLength(1));
    expect(live("/controller/stream")).toHaveLength(0);
    // Read through fetch with the request header, not EventSource (decision 45).
    expect(live("/node/stream")[0].headers["X-Fabric-UI"]).toBe("1");
    expect(sources).toHaveLength(0);
    location.hash = "#/controller/nodes";
    window.dispatchEvent(new HashChangeEvent("hashchange"));
    await waitFor(() => expect(live("/controller/stream")).toHaveLength(1));
    expect(live("/node/stream")).toHaveLength(0);
  });
});

describe("gated pages (0.10.0): a request without admin rights", () => {
  const json = (status: number, b: unknown) => new Response(JSON.stringify(b), { status, headers: { "Content-Type": "application/json" } });

  it("sees the node family's status and pairing tabs only, and no controller family", async () => {
    serve(["nodes.200"], (url) => {
      if (url.endsWith("/node/status")) return json(200, { ...fixtureBody<object>("node-status.200"), admin: false, csrf_token: undefined });
      if (url.endsWith("/controller/status")) return json(200, { ...fixtureBody<object>("controller-status.200"), admin: false, csrf_token: undefined });
      return undefined;
    });
    location.hash = "#/node/fabrics";
    render(App);
    expect(await screen.findByTestId("node-status")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Status" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Pair" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Fabrics" })).toBeNull();
    expect(screen.queryByRole("link", { name: "Controller" })).toBeNull();
  });

  it("sees the notice on the controller family alone", async () => {
    serve([], (url) => (url.endsWith("/controller/status") ? json(200, { ...fixtureBody<object>("controller-status.200"), admin: false, csrf_token: undefined }) : undefined));
    render(App);
    expect(await screen.findByTestId("denied")).toHaveAttribute("data-code", "forbidden");
    expect(screen.queryByRole("link", { name: "Nodes" })).toBeNull();
  });

  it("a refused address is a notice, and the shell does not probe again", async () => {
    let calls = 0;
    serve([], (url) => {
      if (!url.endsWith("/status")) return undefined;
      calls++;
      return json(403, { code: "host_not_allowed", title: "Host not allowed", status: 403, type: "x" });
    });
    render(App);
    expect(await screen.findByTestId("denied")).toHaveAttribute("data-code", "host_not_allowed");
    const seen = calls;
    await new Promise((r) => setTimeout(r, 1500));
    expect(calls).toBe(seen);
  });
});

describe("UR-P1-ui-112: an inherited name is no page", () => {
  it("#/controller/__proto__ shows the nodes page", async () => {
    serve(["controller-status.200", "controller-nodes.200"]);
    location.hash = "#/controller/__proto__";
    render(App);
    await waitFor(() => expect(screen.getByRole("heading", { name: "Controller" })).toBeInTheDocument());
    await waitFor(() => expect(screen.getByRole("link", { name: "Nodes", current: "page" })).toBeInTheDocument());
    // The nodes page itself renders (its node list).
    expect(await screen.findByText("Kitchen")).toBeInTheDocument();
  });

  it("#/node/<default key>/constructor shows the status page", async () => {
    serve(["node-status.200", "nodes.200"]);
    location.hash = "#/node/default/constructor";
    render(App);
    expect(await screen.findByTestId("node-status")).toBeInTheDocument();
  });
});

describe("UR-P1-conc-199: one probe chain, none after the shell is gone", () => {
  it("a probe in flight when the shell is destroyed schedules no retry", async () => {
    let calls = 0;
    serve([]);
    const fetchImpl = globalThis.fetch as unknown as ReturnType<typeof vi.fn>;
    const inner = fetchImpl.getMockImplementation() as (url: string, init?: RequestInit) => Promise<Response>;
    // The controller probe answers 502 after 100 ms.
    fetchImpl.mockImplementation(async (url: string, init?: RequestInit) => {
      if (url.endsWith("/controller/status")) {
        calls++;
        await new Promise((r) => setTimeout(r, 100));
        return new Response("<html>502</html>", { status: 502 });
      }
      return inner(url, init);
    });
    const view = render(App);
    await waitFor(() => expect(calls).toBe(1));
    view.unmount();
    await new Promise((r) => setTimeout(r, 1500));
    expect(calls).toBe(1);
  });
});

describe("the Operation page waits for the probes", () => {
  it("reads no log while the node status is still unanswered", async () => {
    let release = () => {};
    const statusHeld = new Promise<void>((r) => (release = r));
    const fetchImpl = serve(["node-status.200", "operation-info.200"]);
    const inner = fetchImpl.getMockImplementation() as (url: string, init?: RequestInit) => Promise<Response>;
    fetchImpl.mockImplementation(async (url: string, init?: RequestInit) => {
      if (url.endsWith("/node/status")) await statusHeld;
      return streamAnswer(url, init) ?? inner(url, init);
    });
    location.hash = "#/node/endpoints";
    render(App);
    await waitFor(() => expect(fetchImpl.mock.calls.some(([u]) => String(u).endsWith("/operation/info"))).toBe(true));
    await new Promise((r) => setTimeout(r, 100));
    expect(fetchImpl.mock.calls.filter(([u]) => String(u).includes("/operation/log"))).toEqual([]);
    release();
    expect(await screen.findByRole("heading", { name: "Node" })).toBeInTheDocument();
    expect(fetchImpl.mock.calls.filter(([u]) => String(u).includes("/operation/log"))).toEqual([]);
  });
});
