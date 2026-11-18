// @vitest-environment happy-dom
// startScan releases the camera on every way out (UR-P1-conc-105).
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { startScan } from "./index";

let track: { stop: ReturnType<typeof vi.fn> };
let release: (s: MediaStream) => void;

beforeEach(() => {
  track = { stop: vi.fn() };
  const stream = { getTracks: () => [track] } as unknown as MediaStream;
  Object.defineProperty(navigator, "mediaDevices", {
    configurable: true,
    value: {
      getUserMedia: vi.fn(
        () =>
          new Promise<MediaStream>((r) => {
            release = r;
            // Answer at once unless a test holds the prompt open.
            queueMicrotask(() => held || r(stream));
          }),
      ),
    },
  });
  held = false;
  streamOf = stream;
});
afterEach(() => vi.restoreAllMocks());

let held = false;
let streamOf: MediaStream;

// happy-dom's video refuses a fake MediaStream: a stand-in with what startScan touches.
function video(play: () => Promise<void>) {
  return { srcObject: null, setAttribute: () => {}, play, videoWidth: 0, videoHeight: 0 } as unknown as HTMLVideoElement;
}

describe("UR-P1-conc-105: the camera is released when a scan fails to start or is cancelled while starting", () => {
  it("stops every track when play() rejects", async () => {
    await expect(startScan(video(() => Promise.reject(new Error("paused"))), () => {})).rejects.toThrow("paused");
    expect(track.stop).toHaveBeenCalled();
  });

  it("stops every track when aborted while the permission prompt is open, and rejects", async () => {
    held = true;
    const ctrl = new AbortController();
    const p = startScan(video(() => Promise.resolve()), () => {}, ctrl.signal);
    await Promise.resolve();
    ctrl.abort();
    release(streamOf);
    await expect(p).rejects.toThrow();
    expect(track.stop).toHaveBeenCalled();
  });

  it("stops a running scan when aborted after it started", async () => {
    const ctrl = new AbortController();
    const onCode = vi.fn();
    await startScan(video(() => Promise.resolve()), onCode, ctrl.signal);
    expect(track.stop).not.toHaveBeenCalled();
    ctrl.abort();
    expect(track.stop).toHaveBeenCalled();
  });
});
