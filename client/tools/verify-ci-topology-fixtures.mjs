import { mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { spawnSync } from "node:child_process";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { fileURLToPath } from "node:url";
import { verifyCITopology } from "./ci-topology.mjs";

const root = new URL("../../", import.meta.url);
const ci = readFileSync(new URL(".github/workflows/ci.yml", root), "utf8");
const maintenance = readFileSync(new URL(".github/workflows/maintenance.yml", root), "utf8");
const makefile = readFileSync(new URL("Makefile", root), "utf8");
const browser = readFileSync(new URL("compose.browser-test.yml", root), "utf8");
const browserEntrypoint = readFileSync(new URL("client/tools/test-browser-ci.sh", root), "utf8");

const mutations = [
  ["exhaustive push gate", ci.replace("make verify-harness-fast", "make verify-harness"), maintenance, makefile],
  ["maintenance trigger in blocking CI", ci.replace("  pull_request:\n", "  pull_request:\n  schedule:\n"), maintenance, makefile],
  ["missing blocking job", ci.replace(/^  schema:[\s\S]*$/m, ""), maintenance, makefile],
  ["wrong workflow leaf", ci.replace("make verify-server-core", "make verify-server"), maintenance, makefile],
  ["missing local push leaf", ci, maintenance, makefile.replace("verify-push: verify-server-core verify-harness-fast verify-client test-browser test-game-ui-composed verify-schema", "verify-push: verify-server-core verify-harness-fast verify-client test-browser verify-schema")],
  ["extra local push leaf", ci, maintenance, makefile.replace("verify-push: verify-server-core verify-harness-fast verify-client test-browser test-game-ui-composed verify-schema", "verify-push: verify-server-core verify-harness-fast verify-client test-browser test-game-ui-composed verify-schema verify-server")],
  ["unbounded exhaustive command", ci, maintenance.replace("timeout --signal=INT --kill-after=30s 50m make harness-observe", "make harness-observe"), makefile],
  ["success-only artifact", ci, maintenance.replace("        if: always()", "        if: success()"), makefile],
  ["missing observation validation", ci, maintenance.replace("make harness-observation-check", "make harness-observation-skip"), makefile],
  ["wrong observation upload", ci, maintenance.replace("          path: harness-observation.json", "          path: other.json"), makefile],
  ["blocking build-cache restoration", ci.replace("          cache: false", "          cache: true"), maintenance, makefile],
  ["maintenance build-cache restoration", ci, maintenance.replaceAll("          cache: false", "          cache: true"), makefile],
  ["push-triggered maintenance", ci, maintenance.replace("  workflow_dispatch:\n", "  workflow_dispatch:\n  push:\n"), makefile],
  ["missing native functional phase", ci, maintenance, makefile.replace(/^\tcd client .*vitest\.browser\.config\.ts test\/typer-table-browser\.test\.ts .*\n/m, "")],
  ["missing broad functional phase", ci, maintenance, makefile.replace(/^\tcd client .*--exclude test\/typer-table-browser\.test\.ts .*\n/m, "")],
  ["missing isolated performance", ci, maintenance, makefile.replace("\t$(MAKE) test-game-ui-performance\n", "")],
];

const coldGoCommands = [
  ["test-go", "\tcd server && go test -p 1 -count=1 $(GO_TEST_FLAGS) $(GO_PACKAGES)"],
  ["replay-fixture-check", "\tcd server && go test ./production -run '^TestApplyLoggedCrossRuntimeFixture$$' -count=1"],
  ["pitch-corpus-check", "\tcd server && go test ./pitch -run '^TestPitchContentGate$$' -count=1"],
  ["typer-corpus-check", "\tcd server && go test ./typer -run '^TestTyperContentGate$$' -count=1"],
];
const reputationCommand = "\tcd server && CLOUD_CLICKER_REPUTATION_EXHAUSTIVE=1 go test -p 1 -count=1 -timeout 60m ./harness -run '^(TestReputationTreeRelevance|TestReputationCareerStartersShortenRunThree)$$' -v";
for (const [name, replacement] of [
  ["missing Reputation opt-in", reputationCommand.replace("CLOUD_CLICKER_REPUTATION_EXHAUSTIVE=1 ", "")],
  ["missing Reputation study", reputationCommand.replace("TestReputationTreeRelevance|", "")],
  ["empty Reputation selector", reputationCommand.replace("TestReputationTreeRelevance|TestReputationCareerStartersShortenRunThree", "TestNonexistentReputationPopulation")],
  ["cached Reputation evidence", reputationCommand.replace(" -count=1", "")],
  ["unbounded Reputation evidence", reputationCommand.replace(" -timeout 60m", "")],
]) {
  const changed = makefile.replace(reputationCommand, replacement);
  if (changed === makefile) throw new Error(`Reputation mutation did not fire: ${name}`);
  mutations.push([name, ci, maintenance, changed]);
}
const missingReputationTarget = makefile.replace("reputation-harness-check:", "other-harness-check:");
if (missingReputationTarget === makefile) throw new Error("Reputation target mutation did not fire");
mutations.push(["missing Reputation target", ci, maintenance, missingReputationTarget]);
for (const [target, command] of coldGoCommands) {
  const changed = makefile.replace(command, command.replace(" -count=1", ""));
  if (changed === makefile) throw new Error(`cold Go mutation did not fire: ${target}`);
  mutations.push([`cached ${target}`, ci, maintenance, changed]);
}

for (const [name, mutatedCI, mutatedMaintenance, mutatedMakefile] of mutations) {
  let rejected = false;
  try {
    verifyCITopology(mutatedCI, mutatedMaintenance, mutatedMakefile, browser, browserEntrypoint);
  } catch {
    rejected = true;
  }
  if (!rejected) throw new Error(`CI topology accepted mutation: ${name}`);
}

console.log(`CI topology negative controls rejected: ${mutations.length}`);

// A local Linux run must execute the shared target, not a copied or narrowed lane.
for (const command of [
  "cd client && node_modules/.bin/vitest run --config vitest.browser.config.ts",
  "make test-browser-functional",
  "make test-browser BROWSER_TEST_FLAGS='--project chromium'",
  "make test-browser-focused",
]) {
  let rejected = false;
  try {
    verifyCITopology(ci, maintenance, makefile, browser.replace('command: ["test-browser"]', `command: [${JSON.stringify(command)}]`), browserEntrypoint);
  } catch {
    rejected = true;
  }
  if (!rejected) throw new Error(`CI topology accepted local browser drift: ${command}`);
}
verifyCITopology(ci, maintenance, makefile, browser, browserEntrypoint);
console.log("Linux browser container rejects copied, incomplete and narrowed lanes");

for (const [name, changedMake, changedBrowser, changedSetup] of [
  ["full command narrowed", makefile.replace("run --rm browser\n", "run --rm browser test-browser-focused\n"), browser, browserEntrypoint],
  ["containment removed", makefile, browser.replace("      - 10m\n", ""), browserEntrypoint],
  ["unfrozen install", makefile, browser, browserEntrypoint.replace(" --frozen-lockfile", "")],
  ["setup ignores failure", makefile, browser, browserEntrypoint.replace("set -euo pipefail", "set -uo pipefail")],
  ["setup narrows selector", makefile, browser, browserEntrypoint.replace('exec make "$@"', 'exec make "$@" BROWSER_TEST_FLAGS=--project=chromium')],
]) {
  if (changedMake === makefile && changedBrowser === browser && changedSetup === browserEntrypoint) throw new Error(`setup mutation did not fire: ${name}`);
  let rejected = false;
  try { verifyCITopology(ci, maintenance, changedMake, changedBrowser, changedSetup); } catch { rejected = true; }
  if (!rejected) throw new Error(`CI topology accepted setup drift: ${name}`);
}

// Execute the actual shared script with disposable tool doubles: verify argument
// boundaries and failure propagation without apt/pnpm touching the host.
const setupRoot = mkdtempSync(join(tmpdir(), "cloud-clicker-browser-setup-"));
try {
  const bin = join(setupRoot, "bin"), trace = join(setupRoot, "trace.jsonl");
  mkdirSync(bin);
  const tool = `#!/usr/bin/env node
const fs = require('node:fs'), path = require('node:path');
const command = path.basename(process.argv[1]);
fs.appendFileSync(process.env.BROWSER_SETUP_TRACE, JSON.stringify([command, ...process.argv.slice(2)]) + '\\n');
process.exit(command === process.env.BROWSER_SETUP_FAIL ? 37 : 0);
`;
  for (const name of ["apt-get", "corepack", "pnpm", "make"]) writeFileSync(join(bin, name), tool, { mode: 0o700 });
  const setup = (args, fail = "") => {
    writeFileSync(trace, "");
    const result = spawnSync("bash", [fileURLToPath(new URL("client/tools/test-browser-ci.sh", root)), ...args], {
      cwd: root, encoding: "utf8", env: { ...process.env, PATH: `${bin}:${process.env.PATH}`,
        BROWSER_SETUP_TRACE: trace, BROWSER_SETUP_FAIL: fail },
    });
    const calls = readFileSync(trace, "utf8").trim().split("\n").filter(Boolean).map((line) => JSON.parse(line));
    return { ...result, calls };
  };
  const prelude = [["apt-get", "update"], ["apt-get", "install", "-y", "--no-install-recommends", "make"],
    ["corepack", "enable"], ["corepack", "prepare", "pnpm@11.15.1", "--activate"],
    ["pnpm", "--dir", "client", "install", "--frozen-lockfile"]];
  const selector = "BROWSER_TEST_FLAGS=test/numeric-monotonicity.test.ts -t 'positive strict'";
  for (const args of [["test-browser"], ["test-browser-focused", selector]]) {
    const result = setup(args);
    if (result.status !== 0 || JSON.stringify(result.calls) !== JSON.stringify([...prelude, ["make", ...args]])) {
      throw new Error(`browser setup changed commands/argument boundaries: ${JSON.stringify(result)}`);
    }
  }
  for (const args of [[], ["unknown-target"]]) {
    const result = setup(args);
    if (result.status !== 2 || result.calls.length !== 0) throw new Error("invalid browser setup target reached host tools");
  }
  for (const fail of ["apt-get", "make"]) {
    const result = setup(["test-browser-focused", selector], fail);
    if (result.status !== 37 || result.calls.at(-1)?.[0] !== fail || (fail === "apt-get" && result.calls.length !== 1)) {
      throw new Error("browser setup hid failure or continued after failed setup");
    }
  }
} finally {
  rmSync(setupRoot, { recursive: true, force: true });
}
console.log("Shared Linux browser setup preserves commands/arguments and propagates setup/test failures");

// Exercise actual Make expansion, including focused selection and report safety.
function dryRun(target, ...flags) {
  return spawnSync("make", ["-n", target, ...flags], { cwd: root, encoding: "utf8" });
}
const goDefault = dryRun("test-go", "GO_PACKAGES=./production", "GO_TEST_FLAGS=-run TestApplyLoggedCrossRuntimeFixture");
if (goDefault.status !== 0 || !goDefault.stdout.includes("go test -p 1 -count=1 -run TestApplyLoggedCrossRuntimeFixture ./production")) {
  throw new Error("focused Go flags erased the cold test default or package selector");
}
const goRepeated = dryRun("test-go", "GO_PACKAGES=./production", "GO_TEST_FLAGS=-run TestApplyLoggedCrossRuntimeFixture -count=2");
if (goRepeated.status !== 0 || !goRepeated.stdout.includes("go test -p 1 -count=1 -run TestApplyLoggedCrossRuntimeFixture -count=2 ./production")) {
  throw new Error("explicit developer repetition was overridden by the cold Go default");
}
console.log("Go Make cold default, focused selection and explicit repetition controls passed");
const focused = dryRun("test-browser-focused", "BROWSER_TEST_FLAGS=test/amount-format.test.ts --project chromium");
if (focused.status !== 0 || !focused.stdout.includes("test/amount-format.test.ts --project chromium") ||
    focused.stdout.includes("typer-table-browser.test.ts")) throw new Error("focused browser selector was expanded into a full/native population");
const linuxFocused = dryRun("test-browser-ci-focused", "BROWSER_TEST_FLAGS=test/numeric-monotonicity.test.ts --project firefox");
if (linuxFocused.status !== 0 || !linuxFocused.stdout.includes('browser test-browser-focused "BROWSER_TEST_FLAGS=test/numeric-monotonicity.test.ts --project firefox"')) {
  throw new Error("focused Linux browser selector lost its single Make assignment");
}
const full = dryRun("test-browser-functional", "BROWSER_TEST_FLAGS=--project chromium --project webkit", "BROWSER_REPORT_DIR=.cache/browser-fixture");
if (full.status !== 0 || !full.stdout.includes("--exclude test/typer-table-browser.test.ts") ||
    !full.stdout.includes("vitest.browser.config.ts test/typer-table-browser.test.ts") ||
    !full.stdout.includes("functional-broad.json") || !full.stdout.includes("functional-native.json") ||
    (full.stdout.match(/--project chromium --project webkit/g) ?? []).length !== 2) throw new Error("full functional phase/report expansion is incomplete");
const aggregate = dryRun("test-browser", "BROWSER_TEST_FLAGS=--project chromium --project webkit");
if (aggregate.status !== 0 || aggregate.stdout.indexOf("--exclude test/typer-table-browser.test.ts") < 0 ||
    aggregate.stdout.indexOf("vitest.browser.config.ts test/typer-table-browser.test.ts") < 0 ||
    aggregate.stdout.indexOf("VITE_GAME_UI_PERFORMANCE=1") <= aggregate.stdout.indexOf("vitest.browser.config.ts test/typer-table-browser.test.ts")) {
  throw new Error("full browser Make target did not preserve functional-before-performance ordering");
}
const overwrite = dryRun("test-browser-functional", "BROWSER_TEST_FLAGS=--outputFile.json=shared.json");
if (overwrite.status === 0 || !overwrite.stderr.includes("use BROWSER_REPORT_DIR")) throw new Error("full functional phases accepted an overwritten shared report");
console.log("Browser Make selectors and separate report controls passed");
