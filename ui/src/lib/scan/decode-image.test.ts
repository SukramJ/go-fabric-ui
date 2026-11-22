// @vitest-environment happy-dom
// decodeImage reads an onboarding payload from a photograph (contract
// 0.13.0): jsQR on a canvas where the browser has no detector, the native
// detector where it has one; anything but a payload is null.
import { afterEach, describe, expect, it, vi } from "vitest";
import { create as qrCreate } from "qrcode";
import { decodeImage } from "./index";

const SCALE = 4;
const MARGIN = 4;

/** The RGBA pixels of text's QR code, dark on light, as a photograph's canvas would hold them. */
function qrPixels(text: string): { data: Uint8ClampedArray; side: number } {
  const { size, data } = qrCreate(text, { errorCorrectionLevel: "M" }).modules;
  const side = (size + 2 * MARGIN) * SCALE;
  const px = new Uint8ClampedArray(side * side * 4).fill(255);
  for (let y = 0; y < side; y++) {
    for (let x = 0; x < side; x++) {
      const mx = Math.floor(x / SCALE) - MARGIN;
      const my = Math.floor(y / SCALE) - MARGIN;
      if (mx >= 0 && my >= 0 && mx < size && my < size && data[my * size + mx]) {
        const i = (y * side + x) * 4;
        px[i] = px[i + 1] = px[i + 2] = 0;
      }
    }
  }
  return { data: px, side };
}

/** A photograph of text's QR code: createImageBitmap and the canvas stubbed with its pixels. */
function photo(text: string) {
  const { data, side } = qrPixels(text);
  const close = vi.fn();
  vi.stubGlobal("createImageBitmap", vi.fn(async () => ({ width: side, height: side, close })));
  vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue({
    drawImage: () => {},
    getImageData: (_x: number, _y: number, w: number, h: number) => {
      expect([w, h]).toEqual([side, side]);
      return { data, width: w, height: h };
    },
  } as unknown as CanvasRenderingContext2D);
  return { blob: new Blob(["x"], { type: "image/png" }), close };
}

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

describe("decodeImage", () => {
  it("reads an onboarding payload with jsQR when the browser has no detector, and releases the bitmap", async () => {
    const { blob, close } = photo("MT:Y.K9042C00KA0648G00");
    expect(await decodeImage(blob)).toBe("MT:Y.K9042C00KA0648G00");
    expect(close).toHaveBeenCalled();
  });

  it("returns null for a QR code that is no onboarding payload", async () => {
    const { blob } = photo("https://example.com");
    expect(await decodeImage(blob)).toBeNull();
  });

  it("returns null for a file that is no image", async () => {
    vi.stubGlobal("createImageBitmap", vi.fn(async () => Promise.reject(new Error("not an image"))));
    expect(await decodeImage(new Blob(["x"]))).toBeNull();
  });

  it("takes the native detector's payload, and skips its other codes", async () => {
    const { blob } = photo("https://example.com");
    class Detector {
      static getSupportedFormats = async () => ["qr_code"];
      async detect() {
        return [{ rawValue: "https://example.com" }, { rawValue: "MT:Y.K9042C00KA0648G00" }];
      }
    }
    vi.stubGlobal("BarcodeDetector", Detector);
    expect(await decodeImage(blob)).toBe("MT:Y.K9042C00KA0648G00");
  });
});
