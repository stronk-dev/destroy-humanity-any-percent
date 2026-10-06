import { spawn, execFileSync } from "node:child_process";
import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { createInterface } from "node:readline";
import { fileURLToPath } from "node:url";

const root = fileURLToPath(new URL("../../", import.meta.url));
const packageName = "cloud-clicker/server/account";
export const parents = Object.freeze([
  "TestRefreshWireRotationAndReuseIntegration",
  "TestRefreshWireNewFounderBindingIntegration",
  "TestRefreshWireUnknownExpiredAndFaultIntegration",
  "TestRefreshWireLimiterNonMutationIntegration",
]);
export const leaves = Object.freeze([
  `${parents[0]}/access-expired=false`, `${parents[0]}/access-expired=true`,
  parents[1], `${parents[2]}/unknown-canonical`, `${parents[2]}/expired-refresh`,
  `${parents[2]}/closed-database`, parents[3],
]);
const names = new Set([...parents, ...leaves]);
const testPattern = `^(${parents.join("|")})$`;
// Command-line Make variable values are expanded by Make before its recipe shell.
// Preserve the regex end anchor; a single dollar would consume the closing quote.
const makeTestPattern = testPattern.replaceAll("$", () => "$$");

export function refreshPopulationObserver() {
  const state = new Map();
  const errors = [];
  let eventCount = 0, nonJSONLines = 0, packageStarted = false, packageTerminal;
  const reject = (code) => { errors.push(code); };
  function consumeLine(line) {
    if (!line.trim()) return;
    // Make/Compose status lines are not Go events. Missing events still fail.
    if (!line.trimStart().startsWith("{")) { nonJSONLines++; return; }
    let row;
    try { row = JSON.parse(line); } catch { reject("malformed_json"); return; }
    eventCount++;
    if (!row || Array.isArray(row) || typeof row !== "object" || row.Package !== packageName || typeof row.Action !== "string" ||
        row.Test !== undefined && typeof row.Test !== "string" || row.Action === "output" && typeof row.Output !== "string") {
      reject("invalid_event_or_package"); return;
    }
    if (row.Action === "output") return; // Never retain private test Output strings.
    if (packageTerminal) { reject("event_after_package_terminal"); return; }
    if (row.Action === "start" && row.Test === undefined) {
      if (packageStarted) reject("duplicate_package_start");
      packageStarted = true; return;
    }
    if (!packageStarted) { reject("event_before_package_start"); return; }
    if (row.Test !== undefined) {
      if (!names.has(row.Test)) { reject("unexpected_test"); return; }
      const previous = state.get(row.Test);
      if (row.Action === "run") {
        if (previous) reject("duplicate_test_run");
        else state.set(row.Test, "running");
      } else if (row.Action === "pass" || row.Action === "skip" || row.Action === "fail") {
        if (previous !== "running") reject("invalid_test_terminal_order");
        if (row.Action !== "pass") reject(`test_${row.Action}`);
        if (row.Action === "pass" && parents.includes(row.Test)) {
          const children = leaves.filter((name) => name.startsWith(`${row.Test}/`));
          if (children.some((name) => state.get(name) !== "pass")) reject("parent_pass_without_children");
        }
        state.set(row.Test, row.Action);
      } else if (row.Action !== "pause" && row.Action !== "cont") reject("unexpected_test_action");
      return;
    }
    if (row.Action === "pass" || row.Action === "skip" || row.Action === "fail") {
      packageTerminal = row.Action;
      if (row.Action !== "pass") reject(`package_${row.Action}`);
      if ([...names].some((name) => state.get(name) !== "pass")) reject("package_terminal_without_population");
    } else reject("unexpected_package_action");
  }
  function finish(exitCode, sourceStable = true) {
    if (exitCode !== 0) reject("child_nonzero_or_signal");
    if (!sourceStable) reject("source_changed");
    if (!packageStarted || packageTerminal !== "pass") reject("missing_successful_package");
    const missing = [...names].filter((name) => state.get(name) !== "pass");
    if (missing.length) reject("incomplete_population");
    const skipped = [...state].filter(([, value]) => value === "skip").map(([name]) => name);
    const failed = [...state].filter(([, value]) => value === "fail").map(([name]) => name);
    return {
      valid: errors.length === 0, population_complete: missing.length === 0 && errors.length === 0,
      expected_leaves: leaves.length, completed_leaves: leaves.filter((name) => state.get(name) === "pass").length,
      expected_parents: parents.length, completed_parents: parents.filter((name) => state.get(name) === "pass").length,
      missing, skipped, failed, errors, event_count: eventCount, non_json_lines: nonJSONLines,
      child_exit_code: exitCode, package_terminal: packageTerminal ?? null,
      capture_truncated: false, source_stable: sourceStable,
    };
  }
  return { consumeLine, finish };
}

const sourcePaths = [
  "server/account/refresh_wire_integration_test.go", "server/account/refresh_wire_census_test.go",
  "server/account/bootstrap_integration_test.go", "server/account/account_integration_test.go",
  "server/account/api.go", "server/account/store.go", "server/account/token.go", "server/account/credential.go",
  "server/account/bootstrap.go", "server/go.mod", "server/go.sum", "compose.save-test.yml", "Makefile",
  "client/tools/observe-refresh-population.mjs",
];
function sourceIdentity() {
  return {
    head: execFileSync("git", ["rev-parse", "HEAD"], { cwd: root, encoding: "utf8" }).trim(),
    sha256: Object.fromEntries(sourcePaths.map((path) => [path, createHash("sha256").update(readFileSync(resolve(root, path))).digest("hex")])),
  };
}

async function run() {
  const args = process.argv.slice(2);
  if (args.length > 1 || args.length === 1 && args[0] !== "--missing-db-control") throw new Error("unsupported refresh observer mode");
  const missingDB = args[0] === "--missing-db-control";
  const command = missingDB
    ? ["test-go", "GO_PACKAGES=./account", `GO_TEST_FLAGS=-count=1 -json -run '${makeTestPattern}'`]
    : ["test-save-integration", "SAVE_TEST_PACKAGES=./account", `SAVE_TEST_FLAGS=-json -run '${makeTestPattern}'`, "SAVE_TEST_COUNT=1"];
  const env = { ...process.env };
  if (missingDB) delete env.TEST_DATABASE_URL;
  const before = sourceIdentity(), started = Date.now();
  const observer = refreshPopulationObserver();
  const child = spawn("make", command, { cwd: root, env, stdio: ["ignore", "pipe", "inherit"] });
  const lines = createInterface({ input: child.stdout });
  let captureError = false;
  lines.on("line", (line) => { observer.consumeLine(line); });
  lines.on("error", () => { captureError = true; });
  child.stdout.on("error", () => { captureError = true; });
  const exit = await new Promise((resolveExit, rejectExit) => {
    child.once("error", rejectExit);
    child.once("close", (code, signal) => { resolveExit({ code, signal }); });
  });
  const after = sourceIdentity();
  const report = {
    schema_version: 1, mode: missingDB ? "host_missing_database_control" : "declared_postgres_population",
    started_at: new Date(started).toISOString(), elapsed_ms: Date.now() - started,
    observer_node: process.version, test_pattern: testPattern, command: ["make", ...command], source_before: before, source_after: after,
    ...observer.finish(exit.code, JSON.stringify(before) === JSON.stringify(after)),
    child_signal: exit.signal, capture_error: captureError,
  };
  if (captureError) { report.valid = false; report.population_complete = false; report.errors.push("capture_error"); }
  console.info(JSON.stringify(report, null, 2));
  if (!report.valid) process.exitCode = 1;
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  run().catch(() => {
    console.info(JSON.stringify({ schema_version: 1, valid: false, population_complete: false,
      errors: ["observer_or_child_failure"], capture_error: true, source_stable: false }));
    process.exitCode = 1;
  });
}
