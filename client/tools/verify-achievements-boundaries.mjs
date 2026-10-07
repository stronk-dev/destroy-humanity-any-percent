import { readFileSync, readdirSync } from "node:fs";
import { execFileSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { parse } from "svelte/compiler";
import { buildCopyArtifact } from "./copy-pipeline.mjs";

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

function scan(current) {
  for (const entry of readdirSync(current, { withFileTypes: true }).sort((left, right) => left.name.localeCompare(right.name))) {
    const filename = path.join(current, entry.name);
    if (entry.isDirectory()) { scan(filename); continue; }
    if (!entry.isFile() || !entry.name.endsWith(".go") || entry.name.endsWith("_test.go")) continue;
    assertBoundary(readFileSync(filename, "utf8"), path.relative(root, filename));
  }
}
scan(directory);

// GS2-A3: score is not the separate Clout ledger. This is a bounded source
// firewall, not a whole-program/dataflow or persisted-player acceptance proof.
function assertTrophyCase(source, label) {
  const ast = parse(source, { modern: true });
  const visit = (node) => {
    if (node === null || typeof node !== "object") return;
    if (node.type === "Comment") return;
    const name = node.type === "Identifier" ? node.name : undefined;
    const text = node.type === "Literal" ? node.value : node.type === "Text" ? node.data
      : node.type === "TemplateElement" ? node.value?.cooked : undefined;
    if ((typeof name === "string" && /clout/iu.test(name)) ||
        (typeof text === "string" && /clout/iu.test(text))) {
      throw new Error(`${label}: GS2-A3 Trophy Case may not bind or label score as Clout`);
    }
    for (const [key, value] of Object.entries(node)) {
      // ESTree comments are metadata, not executable/player-facing bindings.
      if (["leadingComments", "trailingComments", "comments"].includes(key)) continue;
      if (Array.isArray(value)) for (const child of value) visit(child);
      else if (value !== null && typeof value === "object") visit(value);
    }
  };
  visit(ast);
}

function assertAchievementCopy(entries, label) {
  // The source artifact is built by the canonical Copy Pipeline, not read
  // from generated output. Unknown/duplicate/schema-invalid entries fail there.
  const rows = entries.filter((entry) => /^(?:achievements\.|achievement\.|surface\.achievements\.)/u.test(entry.key));
  for (const key of ["achievements.score_frame", "achievements.grant_frame", "surface.achievements.title"]) {
    if (!rows.some((entry) => entry.key === key)) throw new Error(`${label}: GS2-A3 required score presentation ${key} missing`);
  }
  for (const row of rows) {
    if ([row.key, row.text, ...Object.values(row.era_variants ?? {})].some((text) => /clout/iu.test(text))) {
      throw new Error(`${label}: GS2-A3 achievement copy ${row.key} may not label score as Clout`);
    }
  }
}

function expectGS2Rejected(label, check) {
  try { check(); } catch (error) {
    // A malformed fixture is not evidence the semantic boundary works.
    if (!(error instanceof SyntaxError) && /GS2-A3/u.test(error.message)) return;
    throw error;
  }
  throw new Error(`${label}: GS2-A3 fixture unexpectedly passed`);
}
const componentFaults = [
  '<script>const CloutLifetime = arm.score.lifetime;</script><p>{CloutLifetime}</p>',
  '<p>{arm.CloutLifetime}</p>', '<p>{arm["clout_lifetime"]}</p>',
  '<p>{arm[`clout_lifetime`]}</p>', '<p>Clout: {arm.score.lifetime}</p>',
  '<p>{"Clout: " + arm.score.lifetime}</p>',
  '<p>{t("axis.clout", {value: arm.score.lifetime}, era)}</p>',
];
for (const source of componentFaults) expectGS2Rejected("component fixture", () => assertTrophyCase(source, "component fixture"));
assertTrophyCase('<script>// score, not Clout\nconst score = arm.score.lifetime;</script><!-- no Clout binding --><p>{score}</p>', "score/comment control");
const positiveCopy = [
  { key: "achievements.score_frame", text: "Score {run}", era_variants: null },
  { key: "achievements.grant_frame", text: "Worth {score} score", era_variants: null },
  { key: "surface.achievements.title", text: "Trophy Case", era_variants: null },
  { key: "axis.clout", text: "Clout", era_variants: null },
];
assertAchievementCopy(positiveCopy, "score/separate Clout control");
const copyFaults = [
  positiveCopy.map((row) => row.key === "achievements.score_frame" ? { ...row, text: "Clout {run}" } : row),
  positiveCopy.map((row) => row.key === "achievements.grant_frame" ? { ...row, era_variants: { era_2000: "Worth {score} Clout" } } : row),
  [...positiveCopy, { key: "achievements.clout", text: "Score", era_variants: null }],
  positiveCopy.filter((row) => row.key !== "achievements.score_frame"),
];
for (const entries of copyFaults) expectGS2Rejected("copy fixture", () => assertAchievementCopy(entries, "copy fixture"));
assertTrophyCase(readFileSync(path.join(root, "client/src/game-ui/AchievementsSurface.svelte"), "utf8"), "AchievementsSurface.svelte");
assertAchievementCopy(buildCopyArtifact().artifact.entries, "canonical source copy");
console.log(`GS2-A3 source boundary ok: actual Trophy Case + canonical achievement copy/era variants; ${componentFaults.length} component and ${copyFaults.length} copy negatives`);

function assertDependencies(dependencies, label) {
  const forbidden = dependencies.filter((value) => /^cloud-clicker\/server\/(?:economy|production|save)(?:\/|$)/u.test(value));
  if (forbidden.length > 0) throw new Error(`${label}: forbidden transitive dependencies: ${forbidden.join(", ")}`);
}
const goEnvironment = { ...process.env, GOCACHE: process.env.GOCACHE ?? path.join(root, ".cache", "go-build") };
const fixtureDependencies = execFileSync("go", ["list", "-deps", "./internal/boundaryfixtures/achievementsroot"], { cwd: path.join(root, "server"), encoding: "utf8", env: goEnvironment }).trim().split("\n");
let fixtureRejected = false;
try { assertDependencies(fixtureDependencies, "transitive fixture"); } catch { fixtureRejected = true; }
if (!fixtureRejected) throw new Error("achievements transitive dependency fixture unexpectedly passed");
const dependencies = execFileSync("go", ["list", "-deps", "./achievements"], { cwd: path.join(root, "server"), encoding: "utf8", env: goEnvironment }).trim().split("\n");
assertDependencies(dependencies, "server/achievements");

console.log("achievements package boundary ok");
