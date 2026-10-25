// Fails when a package of the bundle's vendor chunk is missing from
// THIRD-PARTY-NOTICES.md (name and version) or has no licence text under
// licenses/npm/, or when its licence is not one the project accepts.
//
//   FABRIC_UI_VENDOR_LIST=vendor.json npm run build
//   node scripts/check-notices.mjs vendor.json ../THIRD-PARTY-NOTICES.md ../licenses/npm
import fs from "node:fs";
import path from "node:path";

const ACCEPTED = new Set(["MIT", "ISC", "Apache-2.0", "BSD-2-Clause", "BSD-3-Clause"]);
const [listFile, noticesFile, textDir] = process.argv.slice(2);
const list = JSON.parse(fs.readFileSync(listFile, "utf8"));
const notices = fs.readFileSync(noticesFile, "utf8");
const problems = [];
for (const { name, version, license } of list) {
  if (!ACCEPTED.has(license)) problems.push(`${name}: licence ${license} is not accepted`);
  if (!notices.includes(`| \`${name}\` | ${version} | ${license} |`)) problems.push(`${name}@${version} (${license}) is not in ${noticesFile}`);
  const text = path.join(textDir, name.replace(/^@/, "").replace("/", "-") + ".txt");
  if (!fs.existsSync(text)) problems.push(`${name}: no licence text at ${text}`);
}
if (problems.length) {
  console.error("notices check failed:");
  for (const p of problems) console.error(`  - ${p}`);
  process.exit(1);
}
console.log(`notices check: ${list.length} bundled packages listed with their licence texts`);
