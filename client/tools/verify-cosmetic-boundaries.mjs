// Cosmetic Shop v1 §5 / N3 / N8: the cosmetic packages are mechanically
// isolated. server/cosmetic may depend on the Go standard library only (no
// ledger, multiplier, production, faction, fiscal, achievement, Soul, pet
// care, persistence, or metrics/telemetry package, transitively), and
// client/src/cosmetic may import only its own siblings. Each rule ships with
// fixtures that must be rejected.
import { execFileSync } from "node:child_process";
import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");

export function assertGoDependencies(dependencies, label) {
  const forbidden = dependencies.filter((value) => value !== "cloud-clicker/server/cosmetic" && !value.startsWith("vendor/") && (value.includes(".") || value.startsWith("cloud-clicker/")));
  if (forbidden.length > 0) throw new Error(`${label}: cosmetic may depend on the standard library only, not ${forbidden.join(", ")}`);
}

export function assertTypeScriptImports(source, label) {
  for (const match of source.matchAll(/(?:from|import)\s+["']([^"']+)["']/gu)) {
    if (!match[1].startsWith("./")) throw new Error(`${label}: cosmetic may import only its siblings, not ${match[1]}`);
  }
}

function expectRejected(label, operation) {
  try { operation(); } catch { return; }
  throw new Error(`${label} fixture was not rejected`);
}

for (const fixture of [["cloud-clicker/server/economy"], ["cloud-clicker/server/production"], ["cloud-clicker/server/save"],
  ["github.com/prometheus/client_golang/prometheus"], ["cloud-clicker/server/operations"]]) {
  expectRejected(`Go dependency ${fixture[0]}`, () => assertGoDependencies(["fmt", "cloud-clicker/server/cosmetic", ...fixture], "fixture"));
}
for (const fixture of ['import { applyLogged } from "../replay";', 'import { parseCatalog } from "../economy-kernel";', 'import { metric } from "../game-ui/runtime";', 'import telemetry from "some-telemetry";']) {
  expectRejected(`TS import ${fixture}`, () => assertTypeScriptImports(fixture, "fixture"));
}
assertTypeScriptImports('import { cosmeticItem } from "./catalog";', "neutral fixture");

const goEnvironment = { ...process.env, GOCACHE: process.env.GOCACHE ?? path.join(root, ".cache", "go-build") };
const dependencies = execFileSync("go", ["list", "-deps", "./cosmetic"], { cwd: path.join(root, "server"), encoding: "utf8", env: goEnvironment }).trim().split("\n");
assertGoDependencies(dependencies, "server/cosmetic");
const clientDirectory = path.join(root, "client/src/cosmetic");
let files = 0;
for (const entry of readdirSync(clientDirectory).sort()) {
  if (!entry.endsWith(".ts")) continue;
  assertTypeScriptImports(readFileSync(path.join(clientDirectory, entry), "utf8"), `client/src/cosmetic/${entry}`);
  files += 1;
}
console.log(`cosmetic package boundary ok: server/cosmetic is stdlib-only (${dependencies.length} deps); ${files} client files import siblings only; 9 negative fixtures rejected`);
