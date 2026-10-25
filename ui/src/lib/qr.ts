// The QR code of a pairing payload as an inline SVG string, drawn on the
// client from the payload the API returns. Ported from openccu-loom
// assets/ui/src/lib/qr.ts:1-95 at b40b289a (qrcode, MIT): the matrix from
// QRCode.create, one horizontal path segment per run of dark modules.
// Chosen over a second encoder of our own: the library is small, MIT and
// already proven on loom's pairing page.
import { create as qrCreate } from "qrcode";

const MARGIN = 4; // quiet-zone modules (the spec minimum)

function modulesToPath(data: Uint8Array, size: number, margin: number): string {
  let path = "";
  let moveBy = 0;
  let newRow = false;
  let lineLength = 0;
  for (let i = 0; i < data.length; i++) {
    const col = i % size;
    const row = Math.floor(i / size);
    if (col === 0 && !newRow) newRow = true;
    if (data[i]) {
      lineLength++;
      if (!(i > 0 && col > 0 && data[i - 1])) {
        if (newRow) {
          path += `M${col + margin} ${0.5 + row + margin}`;
          newRow = false;
        } else {
          path += `m${moveBy} 0`;
        }
        moveBy = 0;
      }
      if (!(col + 1 < size && data[i + 1])) {
        path += `h${lineLength}`;
        lineLength = 0;
      }
    } else {
      moveBy++;
    }
  }
  return path;
}

/**
 * A self-contained SVG of `payload` as a QR code. Black on a white plate
 * whatever the theme: a scanner needs a light quiet zone and dark modules.
 */
export function renderQrSvg(payload: string, opts?: { size?: number }): string {
  const px = opts?.size ?? 200;
  const qr = qrCreate(payload, { errorCorrectionLevel: "M" });
  const { size, data } = qr.modules;
  const total = size + MARGIN * 2;
  const path = modulesToPath(data, size, MARGIN);
  return (
    `<svg xmlns="http://www.w3.org/2000/svg" width="${px}" height="${px}" viewBox="0 0 ${total} ${total}" shape-rendering="crispEdges">` +
    `<path fill="white" d="M0 0h${total}v${total}H0z"/>` +
    `<path stroke="black" d="${path}"/>` +
    `</svg>`
  );
}
