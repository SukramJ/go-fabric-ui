// The print rule of tokens.css is scoped (U12-R-01, U12-R2-02): the file
// is a public export, so a rule inside `@media print` that is not under
// `html.fui-print-labels` — the class the inventory page sets only for
// the Print button's print — would change every host's print. Outside
// print, `.print-labels` is never shown.
import fs from "node:fs";
import path from "node:path";
import { describe, expect, it } from "vitest";

const css = fs.readFileSync(path.resolve(process.cwd(), "src/lib/theme/tokens.css"), "utf8").replace(/\/\*[\s\S]*?\*\//g, "");

/** The bodies of every `@media print { … }` block, by brace matching. */
function printBlocks(): string[] {
  const out: string[] = [];
  const re = /@media\s+print\s*\{/g;
  while (re.exec(css) !== null) {
    let depth = 1;
    let i = re.lastIndex;
    for (; i < css.length && depth > 0; i++) {
      if (css[i] === "{") depth++;
      else if (css[i] === "}") depth--;
    }
    out.push(css.slice(re.lastIndex, i - 1));
  }
  return out;
}

/** The selectors of a block's top-level rules. */
function selectors(body: string): string[] {
  return [...body.matchAll(/([^{}]+)\{[^{}]*\}/g)].flatMap((r) => r[1].split(",").map((s) => s.trim())).filter(Boolean);
}

describe("tokens.css print rules", () => {
  it("scope every selector inside @media print to html.fui-print-labels", () => {
    const blocks = printBlocks();
    expect(blocks.length).toBeGreaterThan(0);
    const sels = blocks.flatMap(selectors);
    expect(sels.length).toBeGreaterThan(0);
    expect(sels.filter((s) => !s.startsWith("html.fui-print-labels"))).toEqual([]);
  });

  it("hide the label block outside print", () => {
    const outside = printBlocks().reduce((c, b) => c.replace(b, ""), css);
    expect(outside).toMatch(/\.print-labels\s*\{\s*display:\s*none;\s*\}/);
  });
});
