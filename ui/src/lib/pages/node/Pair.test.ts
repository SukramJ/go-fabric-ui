// @vitest-environment happy-dom
// Ported from openccu-loom assets/ui/src/routes/matter/MatterPair.test.ts and
// MatterPair.locale.test.ts at b40b289a, on the contract fixtures.
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { body, testUI } from "../../../test/support";
import { reactiveLocale } from "../../../test/locale.svelte";
import { createTranslator } from "../../i18n";
import type { Status } from "../../api/client";
import Pair from "./Pair.svelte";

const writeText = vi.fn();

beforeEach(() => {
  writeText.mockReset().mockResolvedValue(undefined);
  Object.defineProperty(globalThis.navigator, "clipboard", { value: { writeText }, configurable: true });
});
afterEach(cleanup);

const status = (patch: Partial<Status>) => ({ status: 200, body: { ...body<Status>("node-status.200"), ...patch } });

async function renderPair(st: Partial<Status>, fixtures: string[] = ["node-pairing.200-window"], t?: ReturnType<typeof createTranslator>) {
  const ui = testUI({ fixtures, answers: { "GET /node/status": status(st) }, t });
  await ui.node.loadStatus();
  render(Harness, { ui, page: Pair });
  return ui;
}

describe("Pair — first pairing vs. adding a controller", () => {
  it("frames the action as first-time pairing while no controller holds the node", async () => {
    await renderPair({ fabric_count: 0, commissioning_window_open: false });
    expect(await screen.findByText("node.pair.window_open")).toBeInTheDocument();
    expect(screen.queryByText(/node\.pair\.already_paired/)).toBeNull();
  });

  it("says how many controllers already hold the node and offers to add one", async () => {
    await renderPair({ fabric_count: 3, commissioning_window_open: false });
    expect(await screen.findByText('node.pair.already_paired:{"count":3}')).toBeInTheDocument();
    expect(screen.getByText("node.pair.add_controller")).toBeInTheDocument();
  });
});

describe("Pair — the window", () => {
  it("opens a window and shows its QR code, manual code and countdown", async () => {
    const ui = await renderPair({ fabric_count: 0, commissioning_window_open: false }, ["node-pairing-window.200"]);
    await fireEvent.click(await screen.findByText("node.pair.window_open"));
    expect(await screen.findByText("00020202021")).toBeInTheDocument();
    expect(screen.getByTestId("pair-qr").querySelector("svg")).not.toBeNull();
    expect(screen.getByText('common.seconds_short:{"n":300}')).toBeInTheDocument();
    expect(ui.sent.find((s) => s.path.endsWith("/pairing/window"))?.body).toEqual({ duration_seconds: 300 });
  });

  it("reports a refused open (a window is open already) as an error", async () => {
    const ui = await renderPair({ fabric_count: 0, commissioning_window_open: false }, ["node-pairing-window.409"]);
    await fireEvent.click(await screen.findByText("node.pair.window_open"));
    await waitFor(() => expect(ui.toastSpy.error).toHaveBeenCalledWith("a commissioning window is already open"));
  });

  it("closes the window and returns to the form", async () => {
    await renderPair({}, ["node-pairing.200-window", "node-pairing-window-close.204"]);
    await fireEvent.click(await screen.findByText("node.pair.close_window"));
    expect(await screen.findByText("node.pair.add_controller")).toBeInTheDocument();
  });

  it("shows the paired controller once fabric_added arrives", async () => {
    const ui = await renderPair({}, ["node-pairing.200-window", "node-fabrics.200"]);
    await screen.findByText("00020202021");
    ui.node.applyEvent({ type: "fabric_added", data: { fabric_index: 3, label: "Home" } });
    expect(await screen.findByText("node.pair.success")).toBeInTheDocument();
    expect(screen.getByText("Home")).toBeInTheDocument();
  });
});

describe("Pair — copying the codes", () => {
  it("copies the manual code, which is otherwise transcribed by hand", async () => {
    const ui = await renderPair({});
    await fireEvent.click(await screen.findByLabelText("node.pair.copy_manual_code"));
    await waitFor(() => expect(writeText).toHaveBeenCalledWith("00020202021"));
    expect(ui.toastSpy.success).toHaveBeenCalledWith("node.pair.copied");
  });

  it("copies the QR payload, which is the fallback when the code cannot be scanned", async () => {
    await renderPair({});
    await fireEvent.click(await screen.findByLabelText("node.pair.copy_qr_payload"));
    await waitFor(() => expect(writeText).toHaveBeenCalledWith("MT:FRESH1234"));
  });

  it("reports a refused clipboard instead of claiming the code was copied", async () => {
    writeText.mockRejectedValue(new Error("denied"));
    const ui = await renderPair({});
    await fireEvent.click(await screen.findByLabelText("node.pair.copy_manual_code"));
    await waitFor(() => expect(ui.toastSpy.error).toHaveBeenCalledWith("node.pair.copy_failed"));
    expect(ui.toastSpy.success).not.toHaveBeenCalled();
  });
});

describe("Pair — duration picker follows a live locale switch", () => {
  it("re-renders the option labels when the language changes", async () => {
    const loc = reactiveLocale("en");
    const t = createTranslator(loc.get);
    await renderPair({ commissioning_window_open: false }, [], t);
    expect(await screen.findByText("5 min")).toBeInTheDocument();
    loc.set("de");
    expect(await screen.findByText("5 Min.")).toBeInTheDocument();
  });
});

describe("Pair — ecosystem hints", () => {
  it("shows a hint per ecosystem with its links, and marks the ecosystems already paired with their findings", async () => {
    const c = body<{ ecosystems: unknown[]; findings: unknown[] }>("node-compatibility.200");
    c.findings = [{ ecosystem: "apple", code: "x", message: "Apple shows this device type as a switch." }];
    const ui = testUI({
      fixtures: ["node-pairing.200-window", "node-compatibility.200"],
      answers: { "GET /node/status": status({ commissioning_window_open: false }), "GET /node/compatibility": { status: 200, body: c } },
    });
    await ui.node.loadStatus();
    render(Harness, { ui, page: Pair });
    const hints = await screen.findAllByTestId("ecosystem-hint");
    expect(hints.map((h) => h.querySelector("summary")?.textContent?.trim().split(/\s+/)[0])).toEqual([
      "node.diag.ecosystem.apple",
      "node.diag.ecosystem.google",
      "node.diag.ecosystem.amazon",
      "node.diag.ecosystem.smartthings",
      "node.diag.ecosystem.home_assistant",
    ]);
    await waitFor(() => expect(hints[0]).toHaveTextContent("node.pair.hints.paired"));
    expect(hints[1]).toHaveTextContent("node.pair.hints.paired");
    expect(hints[2]).not.toHaveTextContent("node.pair.hints.paired");
    expect(hints[2]).toHaveTextContent("node.pair.hints.amazon");
    expect(hints[0]).toHaveTextContent("Apple shows this device type as a switch.");
    const link = screen.getByText("node.pair.hints.apple.hub").closest("a");
    expect(link).toHaveAttribute("href", "https://support.apple.com/en-us/HT207057");
    expect(link).toHaveAttribute("rel", "noopener noreferrer");
  });

  it("shows the hints alone when the node reports no compatibility", async () => {
    await renderPair({ commissioning_window_open: false });
    expect(await screen.findAllByTestId("ecosystem-hint")).toHaveLength(5);
    expect(screen.queryByText("node.pair.hints.paired")).toBeNull();
  });
});

describe("UR-P1-err-04: a window opened elsewhere", () => {
  const configured = (patch: Record<string, unknown> = {}) => ({
    status: 200,
    body: { ...body<Record<string, unknown>>("node-pairing.200-window"), source: "configured", window_open: true, ...patch },
  });

  async function renderWith(st: Partial<Status>, pairing: { status: number; body?: unknown }) {
    const ui = testUI({ answers: { "GET /node/status": status(st), "GET /node/pairing": pairing } });
    render(Harness, { ui, page: Pair });
    return ui;
  }

  it("shows the configured codes with a notice that they may not open a window of another kind, and offers Close", async () => {
    await renderWith({ commissioning_window_open: true, window_kind: "enhanced" }, configured());
    expect(await screen.findByTestId("pair-foreign")).toHaveTextContent("node.pair.foreign_window");
    expect(screen.getByText("00020202021")).toBeInTheDocument();
    expect(screen.getByText("node.pair.close_window")).toBeInTheDocument();
  });

  it("shows the codes without a notice for the page's own window, a window answered by its own codes, or a server without window_kind", async () => {
    await renderWith({ commissioning_window_open: true, window_kind: "own" }, configured());
    expect(await screen.findByText("00020202021")).toBeInTheDocument();
    expect(screen.queryByTestId("pair-foreign")).toBeNull();
    cleanup();
    await renderWith({ commissioning_window_open: true, window_kind: "basic" }, configured({ source: "window" }));
    expect(await screen.findByText("00020202021")).toBeInTheDocument();
    expect(screen.queryByTestId("pair-foreign")).toBeNull();
    cleanup();
    await renderWith({ commissioning_window_open: true, window_kind: undefined }, configured());
    expect(await screen.findByText("00020202021")).toBeInTheDocument();
    expect(screen.queryByTestId("pair-foreign")).toBeNull();
  });

  it("stays on the form when the window closed between the status and the pairing read", async () => {
    await renderWith({ commissioning_window_open: true, window_kind: "own" }, configured({ window_open: false }));
    expect(await screen.findByText("node.pair.add_controller")).toBeInTheDocument();
    expect(screen.queryByText("00020202021")).toBeNull();
  });

  it("judges the codes against a fresh status, not the cached one", async () => {
    const ui = testUI({
      answers: { "GET /node/status": status({ commissioning_window_open: true, window_kind: "enhanced" }), "GET /node/pairing": configured() },
    });
    // The cached status still says no window is open.
    ui.node.applyEvent({ type: "window_closed", data: {} });
    render(Harness, { ui, page: Pair });
    expect(await screen.findByTestId("pair-foreign")).toBeInTheDocument();
  });
});

describe("UR-P1-ui-132: a window opened while the page shows the form", () => {
  it("shows the codes on window_opened", async () => {
    let open = false;
    const ui = testUI({ fixtures: ["node-pairing.200-window"] });
    const statusBody = () => ({ ...body<Status>("node-status.200"), commissioning_window_open: open });
    // The status answers what the window is now.
    ui.client.status = async () => statusBody();
    render(Harness, { ui, page: Pair });
    expect(await screen.findByText("node.pair.add_controller")).toBeInTheDocument();
    open = true;
    ui.node.applyEvent({ type: "window_opened", data: {} });
    expect(await screen.findByText("00020202021")).toBeInTheDocument();
  });
});

describe("UR-P1-ui-133: a viewer without admin rights", () => {
  it("reads no codes, shows the open window from the status and says the codes need admin rights", async () => {
    const deadline = new Date(Date.now() + 120_000).toISOString();
    const ui = await renderPair({ admin: false, csrf_token: undefined, commissioning_window_open: true, window_deadline: deadline });
    expect(await screen.findByTestId("pair-viewer")).toHaveTextContent("node.pair.admin_required");
    expect(screen.getByTestId("pair-viewer")).toHaveTextContent("node.pair.window_open_left");
    expect(screen.queryByText("node.pair.window_open")).toBeNull();
    expect(screen.queryByText("node.pair.close_window")).toBeNull();
    expect(screen.queryByText("node.pair.load_error")).toBeNull();
    expect(ui.sent.some((s) => s.path.endsWith("/pairing"))).toBe(false);
  });
});

describe("UR-P1-ui-141: the countdown's unit is translated", () => {
  it("renders the seconds through the catalogue", async () => {
    await renderPair({});
    expect(await screen.findByText(/^common\.seconds_short:/)).toBeInTheDocument();
  });
});
