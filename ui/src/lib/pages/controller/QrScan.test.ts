// @vitest-environment happy-dom
// The scan's Cancel and the page's end while the camera starts
// (UR-P1-conc-105 with -160).
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import Harness from "../../../test/Harness.svelte";
import { testUI } from "../../../test/support";
import QrScan from "./QrScan.svelte";

const scan = vi.hoisted(() => ({ startScan: vi.fn() }));
vi.mock("../../scan", () => scan);

let resolveStart: (stop: () => void) => void;
let started: { onCode: (p: string) => void; signal?: AbortSignal } | null;

beforeEach(() => {
  Object.defineProperty(window, "isSecureContext", { configurable: true, value: true });
  Object.defineProperty(navigator, "mediaDevices", { configurable: true, value: { getUserMedia: vi.fn() } });
  started = null;
  scan.startScan.mockReset().mockImplementation(
    (_video: HTMLVideoElement, onCode: (p: string) => void, signal?: AbortSignal) =>
      new Promise<() => void>((r) => {
        started = { onCode, signal };
        resolveStart = r;
      }),
  );
});
afterEach(cleanup);

async function begin() {
  const onCode = vi.fn();
  const ui = testUI();
  const view = render(Harness, { ui, page: QrScan, pageProps: { onCode } });
  await fireEvent.click(screen.getByText("controller.commission.scan"));
  await vi.waitFor(() => expect(started).not.toBeNull());
  return { ui, onCode, view };
}

describe("UR-P1-conc-105: Cancel or leaving while the camera starts", () => {
  it("Cancel while starting stops the scan once it resolves, and nothing reaches the page", async () => {
    const { ui, onCode } = await begin();
    await fireEvent.click(screen.getByText("common.cancel"));
    expect(started?.signal?.aborted).toBe(true);
    const stop = vi.fn();
    resolveStart(stop);
    await vi.waitFor(() => expect(stop).toHaveBeenCalledTimes(1));
    started?.onCode("MT:Y.K9042C00KA0648G00");
    expect(onCode).not.toHaveBeenCalled();
    expect(ui.toastSpy.success).not.toHaveBeenCalled();
    expect(ui.toastSpy.error).not.toHaveBeenCalled();
  });

  it("leaving the page while starting stops the scan once it resolves", async () => {
    const { onCode, view } = await begin();
    view.unmount();
    const stop = vi.fn();
    resolveStart(stop);
    await vi.waitFor(() => expect(stop).toHaveBeenCalledTimes(1));
    started?.onCode("MT:Y.K9042C00KA0648G00");
    expect(onCode).not.toHaveBeenCalled();
  });
});
