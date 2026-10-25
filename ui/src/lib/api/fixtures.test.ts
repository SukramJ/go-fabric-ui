// Every contract fixture is read by a TypeScript test; the Go suite holds
// the same check for its side (fabricui TestEveryFixtureIsUsedByTheGoTests).
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { FIXTURE_DIR } from "../../test/support";

const SRC = fileURLToPath(new URL("../../", import.meta.url));

function testSources(dir: string): string {
  let out = "";
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) out += testSources(p);
    else if (e.name.endsWith(".test.ts")) out += fs.readFileSync(p, "utf8");
  }
  return out;
}

describe("contract fixtures", () => {
  it("are each used by a TypeScript test", () => {
    const names = fs.readdirSync(FIXTURE_DIR).filter((f) => f.endsWith(".json")).map((f) => f.slice(0, -5));
    expect(names.length).toBeGreaterThan(20);
    const src = testSources(SRC);
    const unused = names.filter((n) => !src.includes(`"${n}"`));
    expect(unused).toEqual([]);
  });
});
