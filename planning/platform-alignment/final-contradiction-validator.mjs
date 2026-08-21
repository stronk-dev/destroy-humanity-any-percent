import { execFileSync } from "node:child_process";
import { readdirSync, readFileSync } from "node:fs";

const repositoryRoot = new URL("../../", import.meta.url);
const rootPath = decodeURIComponent(repositoryRoot.pathname);
const read = (relative) => readFileSync(new URL(relative, repositoryRoot), "utf8");
const readTSV = (relative) => {
  const lines = read(relative).trimEnd().split("\n");
  return { header: lines[0].split("\t"), rows: lines.slice(1).map((line) => line.split("\t")) };
};
const runGit = (...args) => execFileSync("git", args, { cwd: rootPath, encoding: "utf8" }).trim();
const assert = (condition, message) => { if (!condition) throw new Error(message); };
const countFiles = (relative, predicate = () => true) => {
  const walk = (url) => readdirSync(url, { withFileTypes: true }).reduce((count, entry) => {
    const child = new URL(`${entry.name}${entry.isDirectory() ? "/" : ""}`, url);
    return count + (entry.isDirectory() ? walk(child) : Number(predicate(entry.name)));
  }, 0);
  return walk(new URL(relative.endsWith("/") ? relative : `${relative}/`, repositoryRoot));
};
const lineCount = (value) => value.length === 0 ? 0 : value.split("\n").length;
const countColumn = (rows, column) => rows.reduce((counts, row) => {
  counts[row[column]] = (counts[row[column]] ?? 0) + 1;
  return counts;
}, {});
const exactCounts = (actual, expected, label) => {
  const normalize = (value) => Object.fromEntries(Object.entries(value).sort(([left], [right]) => left.localeCompare(right)));
  assert(JSON.stringify(normalize(actual)) === JSON.stringify(normalize(expected)), `${label} count drift: ${JSON.stringify(actual)}`);
};

const capability = readTSV("planning/platform-alignment/capability-outcome-ledger.tsv");
exactCounts(countColumn(capability.rows, 11), {
  proven_integration: 3,
  proven_bounded_primitive: 41,
  partial_integration: 55,
  backend_or_data_only: 134,
  client_or_fixture_only: 7,
  claimed_only: 3,
  absent: 188,
  blocked: 2,
}, "capability");

const content = readTSV("planning/platform-alignment/gameplay-content-row-ledger.tsv");
exactCounts(countColumn(content.rows, 12), {
  partial_mounted: 173,
  backend_active: 180,
  backend_registered_dormant: 141,
  measurement_only: 55,
  zero_or_empty_placeholder: 21,
  contradicted: 9,
}, "gameplay content");

const oracles = readTSV("planning/platform-alignment/test-oracle-row-ledger.tsv");
exactCounts(countColumn(oracles.rows, 12), {
  dependency_conditional: 51,
  positive_only: 533,
  bounded_discriminating: 171,
  fixture_or_mock_only: 43,
  invalid_or_guarded: 1,
  helper_not_oracle: 2,
  non_discriminating: 1,
}, "test oracle");

const acceptance = readTSV("planning/platform-alignment/active-acceptance-ledger.tsv");
const acceptanceFamilies = { draft: 0, mechanical: 0, proven: 0, partial: 0, contradicted: 0, withdrawn: 0 };
for (const row of acceptance.rows) {
  const state = row[3];
  if (state.startsWith("draft")) acceptanceFamilies.draft += 1;
  else if (state.startsWith("withdrawn")) acceptanceFamilies.withdrawn += 1;
  else if (state.startsWith("cold witness green")) acceptanceFamilies.mechanical += 1;
  else if (state.startsWith("partial") || state.startsWith("unmet")) acceptanceFamilies.partial += 1;
  else if (state.startsWith("contradicted") || state.startsWith("failed")) acceptanceFamilies.contradicted += 1;
  else if (state.startsWith("proven") || state.startsWith("historically proven") || state.startsWith("behavior/review proven")) acceptanceFamilies.proven += 1;
  else throw new Error(`unclassified acceptance state: ${state}`);
}
exactCounts(acceptanceFamilies, { draft: 39, mechanical: 5, proven: 20, partial: 33, contradicted: 10, withdrawn: 4 }, "active acceptance");

const copy = readTSV("planning/platform-alignment/copy-key-consumption-inventory.tsv");
exactCounts(countColumn(copy.rows, 13), {
  mounted_player_copy: 128,
  shipped_backend_or_data_only: 63,
  shipped_unmounted_surface_copy: 1,
  fixture_or_tool_only: 8,
  unreferenced_candidate: 8,
}, "copy consumption");
assert(readTSV("planning/platform-alignment/dependency-resource-ledger.tsv").rows.length === 30, "dependency-resource denominator drift");
const ready = readTSV("planning/platform-alignment/ready-batch-manifest.tsv");
assert(ready.rows.length === 3 && ready.rows.filter((row) => row[1].startsWith("READY")).length === 3, "READY manifest drift");

const backlogRows = read("planning/platform-alignment/backlog.md").split("\n").filter((line) => /^\| RP-\d{3} \|/.test(line));
assert(backlogRows.length === 110, "RP denominator drift");
backlogRows.forEach((line, index) => {
  const fields = line.split("|").map((value) => value.trim());
  assert(fields[1] === `RP-${String(index + 1).padStart(3, "0")}`, `RP sequence drift at ${fields[1]}`);
  assert(fields[2] && fields[3] && fields[4], `RP route/field missing at ${fields[1]}`);
});

assert(countFiles("rfc", (name) => name.endsWith(".md")) === 70, "total RFC Markdown population drift");
assert(countFiles("rfc/archive", (name) => name.endsWith(".md")) === 46, "archived RFC population drift");
assert(countFiles("docs") === 38, "canonical docs population drift");
assert(countFiles("server", (name) => name.endsWith(".go")) === 340, "Go population drift");
assert(countFiles("client/src") === 82, "client source population drift");
assert(readdirSync(new URL("planning/", repositoryRoot), { withFileTypes: true }).filter((entry) => entry.isDirectory()).length === 23, "planning thread population drift");

const sharedPlanning = runGit("ls-files", "--cached", "--others", "--exclude-standard", "planning");
const sharedPlatform = runGit("ls-files", "--cached", "--others", "--exclude-standard", "planning/platform-alignment");
const ignoredPlanning = runGit("ls-files", "--others", "--ignored", "--exclude-standard", "planning");
const ignoredPlatform = runGit("ls-files", "--others", "--ignored", "--exclude-standard", "planning/platform-alignment");
assert(lineCount(sharedPlanning) === 238 && lineCount(sharedPlatform) === 84, "planning shared-file count drift");
assert(lineCount(ignoredPlanning) === 25 && lineCount(ignoredPlatform) === 1, "planning ignored-file count drift");

const changedPaths = runGit("diff", "--name-only", "190a4fa..HEAD").split("\n").filter(Boolean);
const allowed = /^(README\.md|planning\/(?:CURRENT-STATE\.md|README\.md|platform-alignment\/.*)|rfc\/README\.md)$/;
assert(changedPaths.every((path) => allowed.test(path)), `product/authority path entered audit range: ${changedPaths.filter((path) => !allowed.test(path)).join(",")}`);

const requiredSummaryFragments = [
  ["planning/platform-alignment/plan.md", "zero integrated, 171 bounded, 533 positive-only"],
  ["planning/platform-alignment/capability-map.md", "171 are bounded discriminators, and 533 are positive-only"],
  ["planning/platform-alignment/reality-audit.md", "There are 171 bounded discriminating primitives, 533 positive-only"],
  ["planning/platform-alignment/review-handoff.md", "171 bounded, 533 positive-only"],
];
for (const [file, fragment] of requiredSummaryFragments) assert(read(file).includes(fragment), `summary drift: ${file}`);

console.log(JSON.stringify({
  productCoordinate: "190a4fa",
  currentHead: runGit("rev-parse", "HEAD"),
  auditCommitsAtExecution: lineCount(runGit("log", "--format=%H", "190a4fa..HEAD")),
  changedPathsAtExecution: changedPaths.length,
  populations: { acceptance: acceptance.rows.length, capability: capability.rows.length, copy: 208, content: content.rows.length, oracles: oracles.rows.length, backlog: backlogRows.length },
  acceptanceFamilies,
  forbiddenChangedPaths: [],
}));
