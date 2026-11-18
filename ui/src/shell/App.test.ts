// @vitest-environment happy-dom
// The standalone shell's probes and streams (review phase 1b).
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/svelte";
import { fixtureFetch } from "../test/support";
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

const open = (suffix: string) => sources.filter((s) => s.url.endsWith(suffix) && !s.closed);

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
    expect(calls).toBe(2);
  });
});

describe("UR-P1-bound-103: a family's stream only while its pages are shown", () => {
  it("opens the node stream on the node family and the controller stream on the controller family, not both", async () => {
    serve(["node-status.200", "nodes.200", "controller-status.200", "controller-nodes.200"]);
    render(App);
    await waitFor(() => expect(open("/node/stream")).toHaveLength(1));
    expect(open("/controller/stream")).toHaveLength(0);
    location.hash = "#/controller/nodes";
    window.dispatchEvent(new HashChangeEvent("hashchange"));
    await waitFor(() => expect(open("/controller/stream")).toHaveLength(1));
    expect(open("/node/stream")).toHaveLength(0);
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
    const inner = fetchImpl.getMockImplementation()!;
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
