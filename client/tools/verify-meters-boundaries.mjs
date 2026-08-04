import { readFileSync, readdirSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");
const metersDirectory = path.join(root, "server/meters");

function assertBoundary(source, label) {
  if (/"cloud-clicker\/server\/(?:economy|production)(?:"|\/)/u.test(source)) {
    throw new Error(`${label}: meters may not import economy/production state owners`);
  }
}

for (const fixture of [
  'import "cloud-clicker/server/economy"',
  'import ledger "cloud-clicker/server/economy/ledger"',
  'import "cloud-clicker/server/production"',
]) {
  let rejected = false;
  try { assertBoundary(fixture, "fixture"); } catch { rejected = true; }
  if (!rejected) throw new Error(`meters boundary fixture unexpectedly passed: ${fixture}`);
}
assertBoundary('import "cloud-clicker/server/multiplier"', "neutral fixture");

function scan(directory) {
  for (const entry of readdirSync(directory, { withFileTypes: true }).sort((left, right) => left.name.localeCompare(right.name))) {
    const filename = path.join(directory, entry.name);
    if (entry.isDirectory()) { scan(filename); continue; }
    if (!entry.isFile() || !entry.name.endsWith(".go") || entry.name.endsWith("_test.go")) continue;
    assertBoundary(readFileSync(filename, "utf8"), path.relative(root, filename));
  }
}
scan(metersDirectory);

console.log("meters package boundary ok");
