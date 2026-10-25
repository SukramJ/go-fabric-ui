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
    expect(screen.getByText("300s")).toBeInTheDocument();
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
