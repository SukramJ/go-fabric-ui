// The standalone palette's text tokens reach WCAG AA (4.5:1) on the card
// background, light and dark (UR-P1-ui-172). Only our own palette is
// checked: the HA skin's literals are Home Assistant's defaults, and inside
// a panel the host's theme rules.
import fs from "node:fs";
import path from "node:path";
import { describe, expect, it } from "vitest";

const css = fs.readFileSync(path.resolve(process.cwd(), "src/lib/theme/tokens.css"), "utf8");

/** The declarations of the first block whose selector list is exactly sel. */
function block(sel: string): Record<string, string> {
  const at = css.indexOf(sel + " {");
  if (at < 0) throw new Error("no block " + sel);
  const body = css.slice(at + sel.length + 2, css.indexOf("}", at));
  const out: Record<string, string> = {};
  for (const m of body.matchAll(/(--[\w-]+):\s*([^;]+);/g)) out[m[1]] = m[2].trim();
  return out;
}

function srgb(value: string): [number, number, number] {
  const hex = /^#([0-9a-f]{6})$/i.exec(value);
  if (hex) return [0, 2, 4].map((i) => parseInt(hex[1].slice(i, i + 2), 16) / 255) as [number, number, number];
  const ok = /^oklch\(([\d.]+)%\s+([\d.]+)\s+([\d.]+)\)$/.exec(value);
  if (!ok) throw new Error("unparsed colour " + value);
  const l = Number(ok[1]) / 100;
  const c = Number(ok[2]);
  const h = (Number(ok[3]) * Math.PI) / 180;
  const a = c * Math.cos(h);
  const b = c * Math.sin(h);
  const l_ = (l + 0.3963377774 * a + 0.2158037573 * b) ** 3;
  const m_ = (l - 0.1055613458 * a - 0.0638541728 * b) ** 3;
  const s_ = (l - 0.0894841775 * a - 1.291485548 * b) ** 3;
  const lin = [
    4.0767416621 * l_ - 3.3077115913 * m_ + 0.2309699292 * s_,
    -1.2684380046 * l_ + 2.6097574011 * m_ - 0.3413193965 * s_,
    -0.0041960863 * l_ - 0.7034186147 * m_ + 1.707614701 * s_,
  ];
  return lin.map((x) => Math.min(1, Math.max(0, x <= 0.0031308 ? 12.92 * x : 1.055 * x ** (1 / 2.4) - 0.055))) as [number, number, number];
}

function luminance(rgb: [number, number, number]): number {
  const [r, g, b] = rgb.map((c) => (c <= 0.04045 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4));
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

function contrast(a: string, b: string): number {
  const [x, y] = [luminance(srgb(a)), luminance(srgb(b))].sort((p, q) => q - p);
  return (x + 0.05) / (y + 0.05);
}

const TEXT = ["--ha-primary-color", "--ha-error-color", "--ha-warning-color", "--ha-success-color", "--ha-info-color", "--ha-secondary-text-color"];

describe("the standalone palette's text contrast (UR-P1-ui-172)", () => {
  for (const [name, sel] of [
    ["light", ":root,\n:host"],
    ["dark", "html.dark,\n:host(.dark)"],
  ] as const) {
    it(`reaches 4.5:1 on the card background, ${name}`, () => {
      const tokens = block(sel);
      const ground = tokens["--ha-card-background-color"];
      const low = TEXT.map((k) => [k, Math.round(contrast(tokens[k], ground) * 100) / 100] as const).filter(([, r]) => r < 4.5);
      expect(low).toEqual([]);
    });
  }
});
