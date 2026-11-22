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

// The status text tokens (owner decision 31): --fui-*-text reach 4.5:1 on
// the card, and on a badge's 15% tint of the status colour, in all four
// palettes — standalone light and dark, and the HA skin's fallbacks (Home
// Assistant's defaults) light and dark. The HA colours stay for areas,
// borders and icons.
type RGB = [number, number, number];

/** Splits s at its top-level commas. */
function args(s: string): string[] {
  const out: string[] = [];
  let depth = 0;
  let cur = "";
  for (const ch of s) {
    if (ch === "(") depth++;
    if (ch === ")") depth--;
    if (ch === "," && depth === 0) {
      out.push(cur.trim());
      cur = "";
    } else cur += ch;
  }
  out.push(cur.trim());
  return out;
}

/** Evaluates a colour expression against the declarations in env: var() with fallbacks, color-mix(in srgb, …), hex and oklch literals. */
function evalColor(expr: string, env: Record<string, string>): RGB {
  const e = expr.trim();
  const fn = /^([\w-]+)\((.*)\)$/s.exec(e);
  if (fn && fn[1] === "var") {
    const [name, ...rest] = args(fn[2]);
    if (name in env) return evalColor(env[name], env);
    if (rest.length === 0) throw new Error("unset " + name);
    return evalColor(rest.join(", "), env);
  }
  if (fn && fn[1] === "color-mix") {
    const [space, a, b] = args(fn[2]);
    expect(space).toBe("in srgb");
    const pa = /^(.*)\s+([\d.]+)%$/s.exec(a);
    const p = pa ? Number(pa[2]) / 100 : 0.5;
    const ca = evalColor(pa ? pa[1] : a, env);
    const cb = evalColor(b, env);
    return [0, 1, 2].map((i) => ca[i] * p + cb[i] * (1 - p)) as RGB;
  }
  return srgb(e);
}

function ratio(a: RGB, b: RGB): number {
  const [x, y] = [luminance(a), luminance(b)].sort((p, q) => q - p);
  return (x + 0.05) / (y + 0.05);
}

const ROOT = ":root,\n:host";
const DARK = "html.dark,\n:host(.dark)";
const HA = 'html[data-skin="ha"],\n:host([data-skin="ha"])';
const HA_DARK = 'html[data-skin="ha"].dark,\n:host([data-skin="ha"].dark)';
const STATUS = ["primary", "error", "warning", "success", "info"] as const;

describe("the status text tokens (UR-P1-ui-172, decision 31)", () => {
  for (const [name, sels] of [
    ["standalone light", [ROOT]],
    ["standalone dark", [ROOT, DARK]],
    ["HA skin light", [ROOT, HA]],
    ["HA skin dark", [ROOT, DARK, HA, HA_DARK]],
  ] as const) {
    it(`reach 4.5:1 on the card and on a badge's tint, ${name}`, () => {
      const env: Record<string, string> = Object.assign({}, ...sels.map(block));
      const card = evalColor("var(--ha-card-background-color)", env);
      const low: string[] = [];
      for (const s of STATUS) {
        const text = evalColor(`var(--fui-${s}-text)`, env);
        const tint = evalColor(`color-mix(in srgb, var(--ha-${s}-color) 15%, var(--ha-card-background-color))`, env);
        const onCard = ratio(text, card);
        const onTint = ratio(text, tint);
        if (onCard < 4.5 || onTint < 4.5) low.push(`${s}: ${onCard.toFixed(2)} / ${onTint.toFixed(2)}`);
      }
      expect(low).toEqual([]);
    });
  }
});

// The labels on a fill (decision 77, UR-P4-ui-299): a filled button's label
// reaches 4.5:1 on its fill, at rest and hovered, in all four palettes;
// Button.svelte reads these tokens and no other colour for its fills.
const button = fs.readFileSync(path.resolve(process.cwd(), "src/lib/kit/Button.svelte"), "utf8");

describe("the labels on a fill (decision 77, UR-P4-ui-299)", () => {
  for (const [name, sels] of [
    ["standalone light", [ROOT]],
    ["standalone dark", [ROOT, DARK]],
    ["HA skin light", [ROOT, HA]],
    ["HA skin dark", [ROOT, DARK, HA, HA_DARK]],
  ] as const) {
    it(`reach 4.5:1 on the primary and the error fill, ${name}`, () => {
      const env: Record<string, string> = Object.assign({}, ...sels.map(block));
      const low: string[] = [];
      for (const [label, fill] of [
        ["--fui-on-primary", "--fui-primary-fill"],
        ["--fui-on-primary", "--fui-primary-fill-hover"],
        ["--fui-on-error", "--fui-error-fill"],
        ["--fui-on-error", "--fui-error-fill-hover"],
      ]) {
        const r = ratio(evalColor(`var(${label})`, env), evalColor(`var(${fill})`, env));
        if (r < 4.5) low.push(`${label} on ${fill}: ${r.toFixed(2)}`);
      }
      expect(low).toEqual([]);
    });
  }

  it("the filled button variants read the fill and label tokens", () => {
    for (const cls of [
      "bg-[var(--fui-primary-fill)] text-[var(--fui-on-primary)] hover:bg-[var(--fui-primary-fill-hover)]",
      "bg-[var(--fui-error-fill)] text-[var(--fui-on-error)] hover:bg-[var(--fui-error-fill-hover)]",
      "hover:bg-[var(--fui-error-fill)] hover:text-[var(--fui-on-error)]",
    ])
      expect(button).toContain(cls);
    expect(button).not.toMatch(/text-white|bg-\[var\(--ha-(primary|error)-color\)\]/);
  });
});

describe("secondary text, the empty state's description (decision 77, UR-P4-ui-300)", () => {
  for (const [name, sels] of [
    ["standalone light", [ROOT]],
    ["standalone dark", [ROOT, DARK]],
    ["HA skin light", [ROOT, HA]],
    ["HA skin dark", [ROOT, DARK, HA, HA_DARK]],
  ] as const) {
    it(`reaches 4.5:1 on the card, ${name}`, () => {
      const env: Record<string, string> = Object.assign({}, ...sels.map(block));
      expect(ratio(evalColor("var(--ha-secondary-text-color)", env), evalColor("var(--ha-card-background-color)", env))).toBeGreaterThanOrEqual(4.5);
    });
  }
});
