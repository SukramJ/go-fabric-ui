// Writes src/lib/api/schema.gen.ts from contract/openapi.yaml through
// openapi-typescript's API; src/lib/api/schema.test.ts runs generate() and
// fails when the committed file differs.
import fs from "node:fs";
import { fileURLToPath } from "node:url";
import openapiTS, { astToString } from "openapi-typescript";

export const HEADER =
  "// Generated from contract/openapi.yaml by `npm run gen:types`. Do not edit;\n" +
  "// src/lib/api/schema.test.ts fails when this file is stale.\n\n";

export const SPEC = new URL("../../contract/openapi.yaml", import.meta.url);
export const OUT = fileURLToPath(new URL("../src/lib/api/schema.gen.ts", import.meta.url));

export async function generate() {
  return HEADER + astToString(await openapiTS(SPEC));
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  fs.writeFileSync(OUT, await generate());
  console.log(`wrote ${OUT}`);
}
