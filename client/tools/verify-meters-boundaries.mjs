import { readFileSync, readdirSync } from "node:fs";
import { execFileSync } from "node:child_process";
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

function assertDependencies(dependencies, label) {
  const forbidden = dependencies.filter((value) => /^cloud-clicker\/server\/(?:economy|production)(?:\/|$)/u.test(value));
  if (forbidden.length > 0) throw new Error(`${label}: forbidden transitive dependencies: ${forbidden.join(", ")}`);
}
let fixtureRejected = false;
try { assertDependencies(["cloud-clicker/server/economy"], "transitive fixture"); } catch { fixtureRejected = true; }
if (!fixtureRejected) throw new Error("meters transitive dependency fixture unexpectedly passed");
const dependencies = execFileSync("go", ["list", "-deps", "./meters"], { cwd: path.join(root, "server"), encoding: "utf8", env: { ...process.env, GOCACHE: "/tmp/cloud-clicker-boundary-go-cache" } }).trim().split("\n");
assertDependencies(dependencies, "server/meters");

console.log("meters package boundary ok");
