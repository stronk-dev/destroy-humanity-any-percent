import { existsSync, readFileSync, readdirSync } from "node:fs";
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
const sameCorrection = (left, right) => left.offending_commit === right.offending_commit &&
  left.corrected_in_version === right.corrected_in_version && left.reason === right.reason && left.review_log === right.review_log;
// Corrections retain their immutable active-planning citation after an RFC archives. Resolve that
// citation to the process-defined archive location only when the original path no longer exists;
// the correction record itself therefore stays append-only across the archival move.
const reviewLogPath = (reviewLog) => {
  const active = path.join(repositoryRoot, reviewLog);
  if (existsSync(active) || !reviewLog.startsWith("planning/") || reviewLog.startsWith("planning/archive/")) return active;
  const archived = path.join(repositoryRoot, "planning/archive", reviewLog.slice("planning/".length));
  return existsSync(archived) ? archived : active;
};
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
      if (tupleGreater(corrected, left) && (head[0] > corrected[0] || head[0] === corrected[0] && head[1] >= corrected[1]) && existsSync(reviewLogPath(correction.review_log))) return;
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
const reviewRecordsCorrection = (correction) => {
  const filename = reviewLogPath(correction.review_log);
  if (!existsSync(filename)) return false;
  const content = readFileSync(filename, "utf8");
  const ranges = /\(`([0-9a-f]{7,40})\^\.\.([0-9a-f]{7,40})`\)/g;
  for (const match of content.matchAll(ranges)) {
    let left;
    let right;
    try {
      left = git("rev-parse", "--verify", `${match[1]}^{commit}`);
      right = git("rev-parse", "--verify", `${match[2]}^{commit}`);
    } catch { continue; }
    if (left !== correction.offending_commit || right !== correction.offending_commit) continue;
    const heading = content.lastIndexOf("\n## ", match.index);
    const sectionStart = heading < 0 ? 0 : heading + 1;
    const nextHeading = content.indexOf("\n## ", match.index + match[0].length);
    const section = content.slice(sectionStart, nextHeading < 0 ? content.length : nextHeading);
    if (section.includes("**Review by:**") && section.includes("**Decision:**")) return true;
  }
  return false;
};
const correctionTargetsMissedBump = (correction) => {
  let parents;
  try {
    git("merge-base", "--is-ancestor", correction.offending_commit, "HEAD");
    parents = git("rev-list", "--parents", "-n", "1", correction.offending_commit).split(" ").slice(1);
  } catch { return false; }
  return parents.length > 0 && parents.some((parent) => {
    const files = git("diff", "--name-only", parent, correction.offending_commit).split("\n").filter(Boolean);
    if (!files.some((filename) => affecting(filename, guard.paths))) return false;
    const before = git("show", `${parent}:kernel/VERSION`).trim();
    const after = git("show", `${correction.offending_commit}:kernel/VERSION`).trim();
    const left = versionTuple(before, `correction target ${correction.offending_commit} parent`);
    const right = versionTuple(after, `correction target ${correction.offending_commit}`);
    return !files.includes("kernel/VERSION") || !tupleGreater(right, left);
  });
};
// History-rewrite remap affordance (2026-08-06, KRM-F1 closure): after an owner-approved rewrite,
// a correction whose target commit is no longer an ancestor of HEAD may be replaced ONLY by the
// exact old→new pair recorded in the tracked approved rewrite map, one-to-one, with all immutable
// correction fields identical and the new target live. All other removal or mutation remains
// forbidden.
const commitIsLiveAncestor = (commit) => {
  try { git("merge-base", "--is-ancestor", commit, "HEAD"); return true; } catch { return false; }
};
const loadApprovedRemaps = () => {
  const remaps = new Map();
  const seenNew = new Set();
  const dir = path.join(repositoryRoot, "planning/history-rewrites");
  if (!existsSync(dir)) return remaps;
  for (const entry of readdirSync(dir)) {
    if (!entry.endsWith(".map")) continue;
    for (const line of readFileSync(path.join(dir, entry), "utf8").split("\n")) {
      const match = /^([0-9a-f]{40}) -> ([0-9a-f]{40})$/.exec(line.trim());
      if (!match) continue;
      // KRM-F1b: the approved manifest must itself be an injection — duplicate old or new hashes
      // across all tracked map files are rejected (a chain A->B then B->C across rewrites is fine;
      // two olds onto one new, or one old twice, is not).
      if (remaps.has(match[1])) throw new Error(`approved rewrite map has duplicate old hash ${match[1]}`);
      if (seenNew.has(match[2])) throw new Error(`approved rewrite map has duplicate new hash ${match[2]}`);
      remaps.set(match[1], match[2]);
      seenNew.add(match[2]);
    }
  }
  return remaps;
};
const approvedRemaps = loadApprovedRemaps();
// A remap target is acceptable when it is live NOW, or when a later approved rewrite maps it onward
// to a hash that is (a chain A->B->C across successive rewrites; historical walk transitions cite
// intermediate targets that later rewrites killed).
const remapTargetReachesLive = (hash, depth = 0) =>
  depth < 64 && (commitIsLiveAncestor(hash) || (approvedRemaps.has(hash) && remapTargetReachesLive(approvedRemaps.get(hash), depth + 1)));
const excusedRemapPairs = (beforeCorrections, afterCorrections) => {
  const pairs = new Map();
  const usedNew = new Set();
  for (const [oldHash, entry] of beforeCorrections) {
    if (afterCorrections.has(oldHash)) continue;
    if (commitIsLiveAncestor(oldHash)) continue;
    const newHash = approvedRemaps.get(oldHash);
    if (newHash === undefined || usedNew.has(newHash)) continue;
    // KRM-F1b: the remap target must be NEWLY ADDED — a dead correction may never collapse onto a
    // correction that already existed before the comparison.
    if (beforeCorrections.has(newHash)) continue;
    const replacement = afterCorrections.get(newHash);
    if (replacement === undefined || !remapTargetReachesLive(newHash)) continue;
    if (replacement.reason !== entry.reason || replacement.review_log !== entry.review_log ||
        replacement.corrected_in_version !== entry.corrected_in_version) continue;
    pairs.set(oldHash, newHash);
    usedNew.add(newHash);
  }
  return pairs;
};
for (const correction of worktreeCorrections.values()) {
  if (!reviewRecordsCorrection(correction)) throw new Error(`history correction ${correction.offending_commit} is not bound to its independent review range`);
  if (!correctionTargetsMissedBump(correction)) throw new Error(`history correction ${correction.offending_commit} does not identify an actual guarded version miss`);
}
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
      for (const [correctedCommit, parentCorrection] of parentCorrections) {
        const childCorrection = childCorrections.get(correctedCommit);
        if (childCorrection === undefined) {
          if (excusedRemapPairs(parentCorrections, childCorrections).has(correctedCommit)) continue;
          throw new Error(`commit ${commit} removes kernel history correction ${correctedCommit}`);
        }
        if (!sameCorrection(parentCorrection, childCorrection)) throw new Error(`commit ${commit} mutates kernel history correction ${correctedCommit}`);
      }
      const correctionAdditions = [...childCorrections.entries()].filter(([correctedCommit]) => !parentCorrections.has(correctedCommit));
      const walkExcusedNew = new Set(excusedRemapPairs(parentCorrections, childCorrections).values());
      const nonRemapCorrectionAdditions = correctionAdditions.filter(([correctedCommit]) => !walkExcusedNew.has(correctedCommit));
      if (nonRemapCorrectionAdditions.length !== 0 && (!files.includes(correctionsPath) || !files.includes("kernel/VERSION") || nonRemapCorrectionAdditions.some(([, entry]) => entry.corrected_in_version !== after))) {
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
for (const [commit, headCorrection] of headCorrections) {
  const worktreeCorrection = worktreeCorrections.get(commit);
  if (worktreeCorrection === undefined) {
    if (excusedRemapPairs(headCorrections, worktreeCorrections).has(commit)) continue;
    throw new Error(`worktree removes kernel history correction ${commit}`);
  }
  if (!sameCorrection(headCorrection, worktreeCorrection)) throw new Error(`worktree mutates kernel history correction ${commit}`);
}
const additions = [...worktreeCorrections.entries()].filter(([commit]) => !headCorrections.has(commit));
const worktreeExcusedNew = new Set(excusedRemapPairs(headCorrections, worktreeCorrections).values());
const nonRemapAdditions = additions.filter(([commit]) => !worktreeExcusedNew.has(commit));
if (nonRemapAdditions.length !== 0 && (!worktreeFiles.includes(correctionsPath) || !worktreeFiles.includes("kernel/VERSION") || nonRemapAdditions.some(([, entry]) => entry.corrected_in_version !== source))) {
  throw new Error("new kernel history corrections require the correcting VERSION bump in the same commit");
}
assertBump("worktree", null, worktreeFiles, [...new Set([...headGuard.paths, ...guard.paths])], git("show", "HEAD:kernel/VERSION").trim(), source, worktreeCorrections, source);

console.log(`kernel version parity and history guard ok: ${source}`);
