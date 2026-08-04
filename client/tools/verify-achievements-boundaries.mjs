import { readFileSync, readdirSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");
const directory = path.join(root, "server/achievements");

function assertBoundary(source, label) {
  if (/"cloud-clicker\/server\/(?:economy|production|save)(?:"|\/)/u.test(source)) {
    throw new Error(`${label}: achievements may not own ledger spending, production, or persistence`);
  }
  if (/\b(?:CloutLifetime|clout_lifetime|CloutStack)\b/u.test(source)) {
    throw new Error(`${label}: Achievements Foundation may not mint or multiply lifetime Clout`);
  }
}

for (const fixture of [
  'import "cloud-clicker/server/economy"',
  'import "cloud-clicker/server/production"',
  'state.CloutLifetime += grant',
  'const CloutStack = 2',
]) {
  let rejected = false;
  try { assertBoundary(fixture, "fixture"); } catch { rejected = true; }
  if (!rejected) throw new Error(`achievement boundary fixture unexpectedly passed: ${fixture}`);
}
assertBoundary('import "cloud-clicker/server/decimal"', "neutral fixture");

for (const entry of readdirSync(directory, { withFileTypes: true }).sort((left, right) => left.name.localeCompare(right.name))) {
  if (!entry.isFile() || !entry.name.endsWith(".go") || entry.name.endsWith("_test.go")) continue;
  assertBoundary(readFileSync(path.join(directory, entry.name), "utf8"), `server/achievements/${entry.name}`);
}

console.log("achievements package boundary ok");
