import { readFileSync } from "node:fs";
import { verifyCITopology } from "./ci-topology.mjs";

const root = new URL("../../", import.meta.url);
const ci = readFileSync(new URL(".github/workflows/ci.yml", root), "utf8");
const maintenance = readFileSync(new URL(".github/workflows/maintenance.yml", root), "utf8");
const makefile = readFileSync(new URL("Makefile", root), "utf8");

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
];

for (const [name, mutatedCI, mutatedMaintenance, mutatedMakefile] of mutations) {
  let rejected = false;
  try {
    verifyCITopology(mutatedCI, mutatedMaintenance, mutatedMakefile);
  } catch {
    rejected = true;
  }
  if (!rejected) throw new Error(`CI topology accepted mutation: ${name}`);
}

console.log(`CI topology negative controls rejected: ${mutations.length}`);
