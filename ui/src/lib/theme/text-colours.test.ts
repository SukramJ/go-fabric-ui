// Small text in a status colour reads --fui-<status>-text, never the
// --ha-<status>-color, which stays for areas, borders and icons (decision
// 31): under the HA skin Home Assistant's error colour is 4.29:1 on white
// (UR-P4-ui-290). The disabled text colour is no colour for text to read
// either (decision 77, UR-P4-ui-300). A class naming one of them as text
// colour is allowed only on an icon, or on a glyph hidden from assistive
// technology.
import fs from "node:fs";
import path from "node:path";
import { describe, expect, it } from "vitest";

const SRC = path.resolve(process.cwd(), "src");
// The disabled colour too (UR-P4-ui-300): 1.88:1 on the HA card. A
// `disabled:` variant (a control that is disabled) is not text to read.
const HA_TEXT = /(?<![\w:-])text-\[var\(--ha-(?:(?:error|warning|success|primary|info)-color|disabled-text-color)\)\]/;
// The line holds an icon, a table of icon colours, or a glyph hidden from
// assistive technology.
const ICON_LINE = /<Icon\b|\bicon:\s*"|aria-hidden="true"/;

function svelteFiles(dir: string): string[] {
  return fs.readdirSync(dir, { withFileTypes: true }).flatMap((e) => {
    const p = path.join(dir, e.name);
    return e.isDirectory() ? svelteFiles(p) : e.name.endsWith(".svelte") ? [p] : [];
  });
}

describe("status text colours (decision 31, UR-P4-ui-290)", () => {
  it("no element with text takes an HA status colour as its text colour", () => {
    const found: string[] = [];
    for (const f of svelteFiles(SRC)) {
      fs.readFileSync(f, "utf8")
        .split("\n")
        .forEach((line, i) => {
          if (HA_TEXT.test(line) && !ICON_LINE.test(line)) found.push(`${path.relative(SRC, f)}:${i + 1}`);
        });
    }
    expect(found).toEqual([]);
  });
});
