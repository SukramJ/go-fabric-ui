// The committed client types are what `npm run gen:types` writes from the
// contract today.
import fs from "node:fs";
import { describe, expect, it } from "vitest";
// @ts-expect-error — a plain .mjs script without declarations.
import { generate, OUT } from "../../../scripts/gen-types.mjs";

describe("schema.gen.ts", () => {
  it("is not stale against contract/openapi.yaml", async () => {
    expect(fs.readFileSync(OUT as string, "utf8")).toBe(await (generate as () => Promise<string>)());
  });
});
