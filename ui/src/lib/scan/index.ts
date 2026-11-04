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
 * throws when the camera cannot be opened (no permission, no camera).
 */
export async function startScan(video: HTMLVideoElement, onCode: (payload: string) => void): Promise<() => void> {
  const stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: "environment" }, audio: false });
  video.srcObject = stream;
  video.setAttribute("playsinline", "true");
  await video.play();
  const detector = await nativeDetector();
  const canvas = document.createElement("canvas");
  let stopped = false;
  let timer: ReturnType<typeof setTimeout> | null = null;
  const stop = () => {
    stopped = true;
    if (timer) clearTimeout(timer);
    for (const track of stream.getTracks()) track.stop();
    video.srcObject = null;
  };
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
