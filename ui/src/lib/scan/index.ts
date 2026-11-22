// The camera QR scan of the commissioning page, loaded on demand as a chunk
// of its own (dynamic import from QrScan.svelte): the browser's native
// BarcodeDetector where it reads QR codes, else jsQR (Apache-2.0) on the
// video's frames. It reports the first onboarding payload it reads — the
// text a device's QR code carries, "MT:" and the Base38 payload — and
// stops. Only a secure context (HTTPS or localhost) reaches the camera.
import jsQR from "jsqr";

/** Whether text is an onboarding payload a commissioning accepts as its code. */
export function isPayload(text: string): boolean {
  return /^MT:[0-9A-Z.-]+$/.test(text.trim());
}

type Detector = { detect(source: CanvasImageSource): Promise<{ rawValue: string }[]> };
type DetectorCtor = { new (opts: { formats: string[] }): Detector; getSupportedFormats?: () => Promise<string[]> };

/** The native detector when the browser has one that reads QR codes. */
async function nativeDetector(): Promise<Detector | null> {
  const ctor = (globalThis as { BarcodeDetector?: DetectorCtor }).BarcodeDetector;
  if (!ctor) return null;
  try {
    const formats = (await ctor.getSupportedFormats?.()) ?? ["qr_code"];
    return formats.includes("qr_code") ? new ctor({ formats: ["qr_code"] }) : null;
  } catch {
    return null;
  }
}

/** Reads one frame of video with jsQR; null when it finds no code. */
function decodeFrame(video: HTMLVideoElement, canvas: HTMLCanvasElement): string | null {
  const w = video.videoWidth;
  const h = video.videoHeight;
  if (!w || !h) return null;
  canvas.width = w;
  canvas.height = h;
  const ctx = canvas.getContext("2d", { willReadFrequently: true });
  if (!ctx) return null;
  ctx.drawImage(video, 0, 0, w, h);
  const img = ctx.getImageData(0, 0, w, h);
  return jsQR(img.data, w, h, { inversionAttempts: "dontInvert" })?.data ?? null;
}

/**
 * Starts the rear camera on video and scans until a payload is read, which
 * goes to onCode once; returns the function that stops the camera. It
 * throws when the camera cannot be opened (no permission, no camera) or
 * the video does not play, and when `signal` aborts before the scan runs;
 * the camera is released on every such way out, and an abort after the
 * start stops the scan (UR-P1-conc-105).
 */
export async function startScan(video: HTMLVideoElement, onCode: (payload: string) => void, signal?: AbortSignal): Promise<() => void> {
  signal?.throwIfAborted();
  const stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: "environment" }, audio: false });
  let stopped = false;
  let timer: ReturnType<typeof setTimeout> | null = null;
  const stop = () => {
    stopped = true;
    if (timer) clearTimeout(timer);
    for (const track of stream.getTracks()) track.stop();
    video.srcObject = null;
  };
  let detector: Detector | null;
  try {
    signal?.throwIfAborted();
    video.srcObject = stream;
    video.setAttribute("playsinline", "true");
    await video.play();
    detector = await nativeDetector();
    signal?.throwIfAborted();
  } catch (err) {
    stop();
    throw err;
  }
  signal?.addEventListener("abort", stop, { once: true });
  const canvas = document.createElement("canvas");
  const tick = async () => {
    if (stopped) return;
    let text: string | null = null;
    try {
      text = detector ? ((await detector.detect(video))[0]?.rawValue ?? null) : decodeFrame(video, canvas);
    } catch {
      text = null;
    }
    if (stopped) return;
    if (text && isPayload(text)) {
      stop();
      onCode(text.trim());
      return;
    }
    timer = setTimeout(() => void tick(), 200);
  };
  void tick();
  return stop;
}

/** The longer side a photograph is scaled to before jsQR reads it. */
const PHOTO_MAX_SIDE = 1600;

/**
 * Reads an onboarding payload from a photograph (the inventory's attach
 * dialog, contract 0.13.0): the native detector when the browser has one,
 * else jsQR on a canvas of the image scaled to at most 1600 px on its
 * longer side, normal and inverted. The image is decoded in the browser
 * and uploaded nowhere. It needs no secure context: where BarcodeDetector
 * is absent (plain HTTP), jsQR reads it. null when the image holds no
 * onboarding payload or cannot be decoded.
 */
export async function decodeImage(file: Blob): Promise<string | null> {
  let bitmap: ImageBitmap;
  try {
    bitmap = await createImageBitmap(file);
  } catch {
    return null;
  }
  try {
    const detector = await nativeDetector();
    if (detector) {
      try {
        const found = (await detector.detect(bitmap)).map((r) => r.rawValue).find((v) => isPayload(v));
        if (found) return found.trim();
      } catch {
        // jsQR below
      }
    }
    const scale = Math.min(1, PHOTO_MAX_SIDE / Math.max(bitmap.width, bitmap.height));
    const w = Math.max(1, Math.round(bitmap.width * scale));
    const h = Math.max(1, Math.round(bitmap.height * scale));
    const canvas = document.createElement("canvas");
    canvas.width = w;
    canvas.height = h;
    const ctx = canvas.getContext("2d", { willReadFrequently: true });
    if (!ctx) return null;
    ctx.drawImage(bitmap, 0, 0, w, h);
    const text = jsQR(ctx.getImageData(0, 0, w, h).data, w, h, { inversionAttempts: "attemptBoth" })?.data ?? null;
    return text && isPayload(text) ? text.trim() : null;
  } finally {
    bitmap.close?.();
  }
}
