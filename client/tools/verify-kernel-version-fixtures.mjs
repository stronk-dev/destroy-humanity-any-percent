import { execFileSync } from "node:child_process";
import { cpSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

const sourceRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");
const run = (root, command, args = []) => execFileSync(command, args, { cwd: root, encoding: "utf8", stdio: "pipe" });
const write = (root, filename, data) => { mkdirSync(path.dirname(path.join(root, filename)), { recursive: true }); writeFileSync(path.join(root, filename), data); };
const commit = (root, message) => { run(root, "git", ["add", "."]); run(root, "git", ["commit", "-m", message]); };
const versionFiles = (root, version) => {
  write(root, "kernel/VERSION", `${version}\n`);
  write(root, "server/kernel/version.go", `package kernel\nconst Version = "${version}"\n`);
  write(root, "client/src/kernel/version.ts", `export const KERNEL_VERSION = "${version}";\n`);
};

const productionGuard = JSON.parse(readFileSync(path.join(sourceRoot, "kernel/affecting-paths.json"), "utf8"));
for (const required of ["client/src/achievements/", "client/src/meters/", "server/achievements/", "server/meters/"]) {
  if (!productionGuard.paths.includes(required)) throw new Error(`kernel registry omits active mechanic package ${required}`);
}

const root = mkdtempSync(path.join(tmpdir(), "cloud-clicker-kernel-guard-"));
try {
  run(root, "git", ["init", "-q"]); run(root, "git", ["config", "user.email", "guard@example.invalid"]); run(root, "git", ["config", "user.name", "Guard Test"]);
  mkdirSync(path.join(root, "client/tools"), { recursive: true });
  cpSync(path.join(sourceRoot, "client/tools/verify-kernel-version.mjs"), path.join(root, "client/tools/verify-kernel-version.mjs"));
  versionFiles(root, "0.1.0"); write(root, "client/src/replay.ts", "export const replay = 1;\n");
  write(root, "kernel/affecting-paths.json", `${JSON.stringify({ schema_version: 1, paths: ["client/src/replay.ts"] }, null, 2)}\n`);
  commit(root, "kernel: introduce guard"); run(root, "node", ["client/tools/verify-kernel-version.mjs"]);

  write(root, "client/src/replay.ts", "export const replay = 2;\n"); versionFiles(root, "0.1.1"); commit(root, "kernel: bump replay");
  run(root, "node", ["client/tools/verify-kernel-version.mjs"]);

  write(root, "client/src/replay.ts", "export const replay = 21;\n"); commit(root, "test: reviewed semantic miss");
  let historicalFailed = false;
  try { run(root, "node", ["client/tools/verify-kernel-version.mjs"]); } catch { historicalFailed = true; }
  if (!historicalFailed) throw new Error("kernel guard accepted an uncorrected historical miss");
  const offending = run(root, "git", ["rev-parse", "HEAD"]).trim();
  write(root, "planning/kernel-fix/log.md", `# Fixture review ledger\n\n## Independent review (\`${offending.slice(0, 7)}^..${offending.slice(0, 7)}\`)\n\n- **Review by:** Fixture Reviewer\n- **Decision:** not approved until the missed bump is corrected.\n`);
  write(root, "kernel/history-corrections.json", `${JSON.stringify({ schema_version: 1, corrections: [{ offending_commit: offending, corrected_in_version: "0.1.2", reason: "Reviewed fixture violation corrected without rewriting cited history.", review_log: "planning/kernel-fix/log.md" }] }, null, 2)}\n`);
  versionFiles(root, "0.1.2"); commit(root, "kernel: fix forward reviewed history");
  run(root, "node", ["client/tools/verify-kernel-version.mjs"]);

  // KRM-F1 adversarial fixtures: correction remaps must be BIJECTIVE and bound to the approved map.
  const reviewSection = (hash, title) => `\n## Independent review (\`${hash}^..${hash}\`) — ${title}\n\n- **Review by:** Fixture Reviewer\n- **Decision:** not approved until the missed bump is corrected.\n`;
  write(root, "client/src/replay.ts", "export const replay = 41;\n"); commit(root, "test: twin miss one");
  const twinOldOne = run(root, "git", ["rev-parse", "HEAD"]).trim();
  write(root, "client/src/replay.ts", "export const replay = 42;\n"); commit(root, "test: twin miss two");
  const twinOldTwo = run(root, "git", ["rev-parse", "HEAD"]).trim();
  const twinLog = readFileSync(path.join(root, "planning/kernel-fix/log.md"), "utf8");
  write(root, "planning/kernel-fix/log.md", twinLog + reviewSection(twinOldOne, "twin one") + reviewSection(twinOldTwo, "twin two"));
  const twinReason = "Reviewed twin fixture misses with identical metadata corrected without rewriting cited history.";
  const twinState = JSON.parse(readFileSync(path.join(root, "kernel/history-corrections.json"), "utf8"));
  twinState.corrections.push(
    { offending_commit: twinOldOne, corrected_in_version: "0.1.4", reason: twinReason, review_log: "planning/kernel-fix/log.md" },
    { offending_commit: twinOldTwo, corrected_in_version: "0.1.4", reason: twinReason, review_log: "planning/kernel-fix/log.md" });
  twinState.corrections.sort((left, right) => left.offending_commit.localeCompare(right.offending_commit));
  write(root, "kernel/history-corrections.json", `${JSON.stringify(twinState, null, 2)}\n`);
  versionFiles(root, "0.1.4"); commit(root, "kernel: fix forward twin misses");
  run(root, "node", ["client/tools/verify-kernel-version.mjs"]);

  execFileSync("git", ["filter-branch", "-f", "--msg-filter", "sed s/^/fixture-/"], { cwd: root, encoding: "utf8", stdio: "pipe", env: { ...process.env, FILTER_BRANCH_SQUELCH_WARNING: "1" } });
  const subjectHash = (subject) => run(root, "git", ["log", "--format=%H %s"]).split("\n").find((line) => line.endsWith(subject)).split(" ")[0];
  const firstNew = subjectHash("fixture-test: reviewed semantic miss");
  const twinNewOne = subjectHash("fixture-test: twin miss one");
  const twinNewTwo = subjectHash("fixture-test: twin miss two");
  const remapLog = readFileSync(path.join(root, "planning/kernel-fix/log.md"), "utf8");
  write(root, "planning/kernel-fix/log.md", remapLog + reviewSection(firstNew, "first remapped") + reviewSection(twinNewOne, "twin one remapped") + reviewSection(twinNewTwo, "twin two remapped"));
  const remapEntries = (pairs) => {
    const state = JSON.parse(readFileSync(path.join(root, "kernel/history-corrections.json"), "utf8"));
    for (const entry of state.corrections) if (pairs[entry.offending_commit]) entry.offending_commit = pairs[entry.offending_commit];
    state.corrections.sort((left, right) => left.offending_commit.localeCompare(right.offending_commit));
    const unique = new Map(state.corrections.map((entry) => [entry.offending_commit, entry]));
    state.corrections = [...unique.values()].sort((left, right) => left.offending_commit.localeCompare(right.offending_commit));
    write(root, "kernel/history-corrections.json", `${JSON.stringify(state, null, 2)}\n`);
  };
  // Fixture 1: two identical-metadata dead corrections must NOT collapse onto one live target.
  write(root, "planning/history-rewrites/fixture.map", `${offending} -> ${firstNew}\n${twinOldOne} -> ${twinNewOne}\n${twinOldTwo} -> ${twinNewOne}\n`);
  remapEntries({ [offending]: firstNew, [twinOldOne]: twinNewOne, [twinOldTwo]: twinNewOne });
  let collapseFailed = false;
  try { run(root, "node", ["client/tools/verify-kernel-version.mjs"]); } catch { collapseFailed = true; }
  if (!collapseFailed) throw new Error("kernel guard accepted a many-to-one correction remap collapse");
  run(root, "git", ["restore", "kernel/history-corrections.json"]);
  // Fixture 2: a dead target with NO approved-map entry must not be replaceable.
  write(root, "planning/history-rewrites/fixture.map", `${offending} -> ${firstNew}\n${twinOldOne} -> ${twinNewOne}\n`);
  remapEntries({ [offending]: firstNew, [twinOldOne]: twinNewOne, [twinOldTwo]: twinNewTwo });
  let unmappedFailed = false;
  try { run(root, "node", ["client/tools/verify-kernel-version.mjs"]); } catch { unmappedFailed = true; }
  if (!unmappedFailed) throw new Error("kernel guard accepted an unmapped correction remap");
  run(root, "git", ["restore", "kernel/history-corrections.json"]);
  // Positive: the complete owner-approved bijection passes, and the committed remap passes the walk.
  write(root, "planning/history-rewrites/fixture.map", `${offending} -> ${firstNew}\n${twinOldOne} -> ${twinNewOne}\n${twinOldTwo} -> ${twinNewTwo}\n`);
  remapEntries({ [offending]: firstNew, [twinOldOne]: twinNewOne, [twinOldTwo]: twinNewTwo });
  run(root, "node", ["client/tools/verify-kernel-version.mjs"]);
  commit(root, "kernel: remap corrections after approved rewrite");
  run(root, "node", ["client/tools/verify-kernel-version.mjs"]);
  const offendingRemapped = firstNew;
  void offendingRemapped;

  const mainAfterCorrection = run(root, "git", ["branch", "--show-current"]).trim();
  run(root, "git", ["switch", "-c", "dangling-correction"]);
  write(root, "client/src/replay.ts", "export const replay = 22;\n");
  commit(root, "test: unreachable semantic miss");
  const dangling = run(root, "git", ["rev-parse", "HEAD"]).trim();
  run(root, "git", ["switch", mainAfterCorrection]);
  run(root, "git", ["branch", "-D", "dangling-correction"]);
  write(root, "planning/dangling/log.md", `# Fixture review ledger\n\n## Independent review (\`${dangling.slice(0, 7)}^..${dangling.slice(0, 7)}\`)\n\n- **Review by:** Fixture Reviewer\n- **Decision:** dangling history must not enter the correction register.\n`);
  const danglingCorrection = JSON.parse(readFileSync(path.join(root, "kernel/history-corrections.json"), "utf8"));
  danglingCorrection.corrections.push({ offending_commit: dangling, corrected_in_version: "0.1.3", reason: "A deleted side-branch semantic miss must not qualify as reachable project history.", review_log: "planning/dangling/log.md" });
  danglingCorrection.corrections.sort((left, right) => left.offending_commit.localeCompare(right.offending_commit));
  write(root, "kernel/history-corrections.json", `${JSON.stringify(danglingCorrection, null, 2)}\n`);
  versionFiles(root, "0.1.3");
  let danglingCorrectionFailed = false;
  try { run(root, "node", ["client/tools/verify-kernel-version.mjs"]); } catch { danglingCorrectionFailed = true; }
  if (!danglingCorrectionFailed) throw new Error("kernel guard accepted a correction outside reachable HEAD history");
  run(root, "git", ["restore", "."]);
  rmSync(path.join(root, "planning/dangling"), { recursive: true, force: true });

  const acceptedCorrection = JSON.parse(readFileSync(path.join(root, "kernel/history-corrections.json"), "utf8"));
  acceptedCorrection.corrections[0].reason = "Mutated history correction should never be accepted by the append-only guard.";
  write(root, "kernel/history-corrections.json", `${JSON.stringify(acceptedCorrection, null, 2)}\n`);
  let correctionMutationFailed = false;
  try { run(root, "node", ["client/tools/verify-kernel-version.mjs"]); } catch { correctionMutationFailed = true; }
  if (!correctionMutationFailed) throw new Error("kernel guard accepted correction mutation");
  run(root, "git", ["restore", "kernel/history-corrections.json"]);

  write(root, "planning/unrelated/log.md", "# Existing but unrelated log\n");
  const reboundCorrection = JSON.parse(readFileSync(path.join(root, "kernel/history-corrections.json"), "utf8"));
  reboundCorrection.corrections[0].review_log = "planning/unrelated/log.md";
  write(root, "kernel/history-corrections.json", `${JSON.stringify(reboundCorrection, null, 2)}\n`);
  let correctionRebindFailed = false;
  try { run(root, "node", ["client/tools/verify-kernel-version.mjs"]); } catch { correctionRebindFailed = true; }
  if (!correctionRebindFailed) throw new Error("kernel guard accepted correction review-log rebinding");
  run(root, "git", ["restore", "kernel/history-corrections.json"]);
  rmSync(path.join(root, "planning/unrelated"), { recursive: true, force: true });

  write(root, "kernel/history-corrections.json", `${JSON.stringify({ schema_version: 1, corrections: [] }, null, 2)}\n`);
  let correctionRemovalFailed = false;
  try { run(root, "node", ["client/tools/verify-kernel-version.mjs"]); } catch { correctionRemovalFailed = true; }
  if (!correctionRemovalFailed) throw new Error("kernel guard accepted correction removal");
  run(root, "git", ["restore", "kernel/history-corrections.json"]);

  const shallow = path.join(root, "shallow-clone");
  run(root, "git", ["clone", "--depth", "1", `file://${root}`, shallow]);
  let shallowFailed = false;
  try { run(shallow, "node", ["client/tools/verify-kernel-version.mjs"]); } catch { shallowFailed = true; }
  rmSync(shallow, { recursive: true, force: true });
  if (!shallowFailed) throw new Error("kernel guard accepted shallow history");

  write(root, "kernel/affecting-paths.json", `${JSON.stringify({ schema_version: 1, paths: [] }, null, 2)}\n`);
  let removalFailed = false;
  try { run(root, "node", ["client/tools/verify-kernel-version.mjs"]); } catch { removalFailed = true; }
  if (!removalFailed) throw new Error("kernel guard accepted path removal");
  run(root, "git", ["restore", "kernel/affecting-paths.json"]);

  write(root, "client/src/replay.ts", "export const replay = 3;\n"); write(root, "kernel/VERSION", " 0.1.1 \n");
  let cosmeticFailed = false;
  try { run(root, "node", ["client/tools/verify-kernel-version.mjs"]); } catch { cosmeticFailed = true; }
  if (!cosmeticFailed) throw new Error("kernel guard accepted a cosmetic VERSION edit");
  run(root, "git", ["restore", "."]);

  write(root, "client/src/new-kernel.ts", "export const semantic = true;\n");
  write(root, "kernel/affecting-paths.json", `${JSON.stringify({ schema_version: 1, paths: ["client/src/new-kernel.ts", "client/src/replay.ts"] }, null, 2)}\n`);
  let growthFailed = false;
  try { run(root, "node", ["client/tools/verify-kernel-version.mjs"]); } catch { growthFailed = true; }
  if (!growthFailed) throw new Error("kernel guard accepted registry growth plus semantics without a bump");

  run(root, "git", ["restore", "."]);
  const mainBranch = run(root, "git", ["branch", "--show-current"]).trim();
  run(root, "git", ["switch", "-c", "guard-merge-side"]);
  write(root, "merge-side.txt", "side\n"); commit(root, "test: side branch");
  run(root, "git", ["switch", mainBranch]);
  write(root, "merge-main.txt", "main\n"); commit(root, "test: main branch");
  run(root, "git", ["merge", "--no-ff", "--no-commit", "guard-merge-side"]);
  write(root, "client/src/replay.ts", "export const replay = 99;\n");
  commit(root, "test: semantic merge resolution without bump");
  let mergeFailed = false;
  try { run(root, "node", ["client/tools/verify-kernel-version.mjs"]); } catch { mergeFailed = true; }
  if (!mergeFailed) throw new Error("kernel guard accepted unversioned merge-resolution semantics");

  console.log("kernel history guard adversarial fixtures ok");
} finally {
  rmSync(root, { recursive: true, force: true });
}
