import { readFileSync } from "node:fs";
import { verifyCITopology } from "./ci-topology.mjs";

const root = new URL("../../", import.meta.url);
const ci = readFileSync(new URL(".github/workflows/ci.yml", root), "utf8");
const maintenance = readFileSync(new URL(".github/workflows/maintenance.yml", root), "utf8");

const mutations = [
  ["exhaustive push gate", ci.replace("make verify-harness-fast", "make verify-harness"), maintenance],
  ["maintenance trigger in blocking CI", ci.replace("  pull_request:\n", "  pull_request:\n  schedule:\n"), maintenance],
  ["missing blocking job", ci.replace(/^  schema:[\s\S]*$/m, ""), maintenance],
  ["unbounded exhaustive command", ci, maintenance.replace("timeout --signal=INT --kill-after=30s 50m make harness-observe", "make harness-observe")],
  ["success-only artifact", ci, maintenance.replace("        if: always()", "        if: success()")],
  ["missing observation validation", ci, maintenance.replace("make harness-observation-check", "make harness-observation-skip")],
  ["wrong observation upload", ci, maintenance.replace("          path: harness-observation.json", "          path: other.json")],
  ["blocking build-cache restoration", ci.replace("          cache: false", "          cache: true"), maintenance],
  ["maintenance build-cache restoration", ci, maintenance.replaceAll("          cache: false", "          cache: true")],
  ["push-triggered maintenance", ci, maintenance.replace("  workflow_dispatch:\n", "  workflow_dispatch:\n  push:\n")],
];

for (const [name, mutatedCI, mutatedMaintenance] of mutations) {
  let rejected = false;
  try {
    verifyCITopology(mutatedCI, mutatedMaintenance);
  } catch {
    rejected = true;
  }
  if (!rejected) throw new Error(`CI topology accepted mutation: ${name}`);
}

console.log(`CI topology negative controls rejected: ${mutations.length}`);
