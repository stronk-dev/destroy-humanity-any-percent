import { existsSync, readFileSync } from "node:fs";
import { execFileSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import path from "node:path";

const repositoryRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");
const source = readFileSync(new URL("../../kernel/VERSION", import.meta.url), "utf8").trim();
const go = readFileSync(new URL("../../server/kernel/version.go", import.meta.url), "utf8");
const client = readFileSync(new URL("../src/kernel/version.ts", import.meta.url), "utf8");
const guardPath = "kernel/affecting-paths.json";
const correctionsPath = "kernel/history-corrections.json";
const guard = JSON.parse(readFileSync(new URL("../../kernel/affecting-paths.json", import.meta.url), "utf8"));
const goVersion = go.match(/const Version = "([^"]+)"/)?.[1];
const clientVersion = client.match(/KERNEL_VERSION = "([^"]+)"/)?.[1];

if (!/^0\.[0-9]+\.[0-9]+$/.test(source) || source !== goVersion || source !== clientVersion) {
  throw new Error(`kernel version drift: source=${source} go=${goVersion} client=${clientVersion}`);
}

const git = (...args) => execFileSync("git", args, {
  cwd: repositoryRoot,
  encoding: "utf8",
  stdio: ["ignore", "pipe", "pipe"],
}).trim();
const gitStatus = () => execFileSync("git", ["status", "--porcelain=v1", "--untracked-files=all"], { cwd: repositoryRoot, encoding: "utf8" }).trimEnd();
const validateGuard = (value, label) => {
  if (value.schema_version !== 1 || !Array.isArray(value.paths) || value.paths.length === 0 ||
      value.paths.some((entry) => typeof entry !== "string" || entry === "" || entry.startsWith("/") || entry.includes("..")) ||
      value.paths.some((entry, index) => index > 0 && value.paths[index - 1] >= entry)) {
    throw new Error(`${label} has an invalid kernel-affecting path registry`);
  }
  return value;
};
const parseGuard = (data, label) => validateGuard(JSON.parse(data), label);
const validateCorrections = (value, label) => {
  if (value.schema_version !== 1 || !Array.isArray(value.corrections)) throw new Error(`${label} has invalid history corrections`);
  let previous = "";
  const result = new Map();
  for (const entry of value.corrections) {
    if (typeof entry !== "object" || entry === null || Object.keys(entry).sort().join("\0") !== "corrected_in_version\0offending_commit\0reason\0review_log" ||
        !/^[0-9a-f]{40}$/.test(entry.offending_commit) || entry.offending_commit <= previous || !/^0\.[0-9]+\.[0-9]+$/.test(entry.corrected_in_version) ||
        typeof entry.reason !== "string" || entry.reason.length < 20 || typeof entry.review_log !== "string" || !entry.review_log.startsWith("planning/") || !entry.review_log.endsWith("/log.md")) {
      throw new Error(`${label} has invalid history correction entry`);
    }
    previous = entry.offending_commit;
    result.set(entry.offending_commit, entry);
  }
  return result;
};
const emptyCorrections = () => new Map();
const parseCorrections = (data, label) => validateCorrections(JSON.parse(data), label);
const readCorrectionsFile = (filename, label) => existsSync(filename) ? parseCorrections(readFileSync(filename, "utf8"), label) : emptyCorrections();
const affecting = (filename, paths) => !filename.endsWith("_test.go") && paths.some((entry) => entry.endsWith("/") ? filename.startsWith(entry) : filename === entry);
const versionTuple = (value, label) => {
  const match = value.match(/^0\.([0-9]+)\.([0-9]+)$/);
  if (!match) throw new Error(`${label} has invalid kernel version ${value}`);
  return [Number(match[1]), Number(match[2])];
};
const tupleGreater = (left, right) => left[0] > right[0] || left[0] === right[0] && left[1] > right[1];
const assertBump = (label, commit, files, paths, before, after, corrections, headVersion) => {
  const changed = files.filter((filename) => affecting(filename, paths));
  if (changed.length === 0) return;
  const left = versionTuple(before, `${label} parent`);
  const right = versionTuple(after, label);
  if (!files.includes("kernel/VERSION") || right[0] < left[0] || right[0] === left[0] && right[1] <= left[1]) {
    const correction = commit === null ? undefined : corrections.get(commit);
    if (correction !== undefined) {
      const corrected = versionTuple(correction.corrected_in_version, `${label} correction`);
      const head = versionTuple(headVersion, "current corrected version");
      if (tupleGreater(corrected, left) && (head[0] > corrected[0] || head[0] === corrected[0] && head[1] >= corrected[1]) && existsSync(path.join(repositoryRoot, correction.review_log))) return;
    }
    throw new Error(`${label} changes kernel semantics without a real kernel/VERSION bump: ${changed.join(", ")}`);
  }
};
validateGuard(guard, "worktree");
const worktreeCorrections = readCorrectionsFile(path.join(repositoryRoot, correctionsPath), "worktree");
if (git("rev-parse", "--is-shallow-repository") === "true") {
  throw new Error("kernel history guard requires complete Git history");
}

const guardIntroduction = git("log", "--diff-filter=A", "--format=%H", "--", "kernel/affecting-paths.json").split("\n").filter(Boolean).at(-1);
const correctionsAt = (ref, label) => {
  try { return parseCorrections(git("show", `${ref}:${correctionsPath}`), label); } catch { return emptyCorrections(); }
};
if (guardIntroduction) {
  const descendants = git("rev-list", "--reverse", `${guardIntroduction}..HEAD`).split("\n").filter(Boolean);
  const commits = [guardIntroduction, ...descendants];
  for (const commit of commits) {
    const parents = git("rev-list", "--parents", "-n", "1", commit).split(" ").slice(1);
    const childGuard = parseGuard(git("show", `${commit}:${guardPath}`), `commit ${commit}`);
    for (const parent of parents) {
      const files = git("diff", "--name-only", parent, commit).split("\n").filter(Boolean);
      const parentGuard = commit === guardIntroduction ? childGuard : parseGuard(git("show", `${parent}:${guardPath}`), `parent ${parent} of ${commit}`);
      const removed = parentGuard.paths.filter((entry) => !childGuard.paths.includes(entry));
      if (removed.length !== 0) throw new Error(`commit ${commit} removes kernel-affecting paths from parent ${parent}: ${removed.join(", ")}`);
      const before = git("show", `${parent}:kernel/VERSION`).trim();
      const after = git("show", `${commit}:kernel/VERSION`).trim();
      const parentCorrections = correctionsAt(parent, `parent corrections of ${commit}`);
      const childCorrections = correctionsAt(commit, `corrections at ${commit}`);
      for (const correctedCommit of parentCorrections.keys()) if (!childCorrections.has(correctedCommit)) throw new Error(`commit ${commit} removes kernel history correction ${correctedCommit}`);
      const correctionAdditions = [...childCorrections.entries()].filter(([correctedCommit]) => !parentCorrections.has(correctedCommit));
      if (correctionAdditions.length !== 0 && (!files.includes(correctionsPath) || !files.includes("kernel/VERSION") || correctionAdditions.some(([, entry]) => entry.corrected_in_version !== after))) {
        throw new Error(`commit ${commit} adds history corrections without its correcting VERSION bump`);
      }
      assertBump(`commit ${commit} against parent ${parent}`, commit, files, [...new Set([...parentGuard.paths, ...childGuard.paths])], before, after, worktreeCorrections, source);
    }
  }
}

const worktreeFiles = gitStatus().split("\n").filter(Boolean).flatMap((line) => {
  const filename = line.slice(3);
  return filename.includes(" -> ") ? filename.split(" -> ") : [filename];
});
const headGuard = parseGuard(git("show", `HEAD:${guardPath}`), "HEAD");
const removedFromWorktree = headGuard.paths.filter((entry) => !guard.paths.includes(entry));
if (removedFromWorktree.length !== 0) throw new Error(`worktree removes kernel-affecting paths: ${removedFromWorktree.join(", ")}`);
const headCorrections = correctionsAt("HEAD", "HEAD");
for (const commit of headCorrections.keys()) if (!worktreeCorrections.has(commit)) throw new Error(`worktree removes kernel history correction ${commit}`);
const additions = [...worktreeCorrections.entries()].filter(([commit]) => !headCorrections.has(commit));
if (additions.length !== 0 && (!worktreeFiles.includes(correctionsPath) || !worktreeFiles.includes("kernel/VERSION") || additions.some(([, entry]) => entry.corrected_in_version !== source))) {
  throw new Error("new kernel history corrections require the correcting VERSION bump in the same commit");
}
assertBump("worktree", null, worktreeFiles, [...new Set([...headGuard.paths, ...guard.paths])], git("show", "HEAD:kernel/VERSION").trim(), source, worktreeCorrections, source);

console.log(`kernel version parity and history guard ok: ${source}`);
