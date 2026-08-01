import { readFileSync } from "node:fs";
import { execFileSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import path from "node:path";

const repositoryRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");
const source = readFileSync(new URL("../../kernel/VERSION", import.meta.url), "utf8").trim();
const go = readFileSync(new URL("../../server/kernel/version.go", import.meta.url), "utf8");
const client = readFileSync(new URL("../src/kernel/version.ts", import.meta.url), "utf8");
const guard = JSON.parse(readFileSync(new URL("../../kernel/affecting-paths.json", import.meta.url), "utf8"));
const goVersion = go.match(/const Version = "([^"]+)"/)?.[1];
const clientVersion = client.match(/KERNEL_VERSION = "([^"]+)"/)?.[1];

if (!/^0\.[0-9]+\.[0-9]+$/.test(source) || source !== goVersion || source !== clientVersion) {
  throw new Error(`kernel version drift: source=${source} go=${goVersion} client=${clientVersion}`);
}

if (guard.schema_version !== 1 || !Array.isArray(guard.paths) || guard.paths.length === 0 ||
    guard.paths.some((value) => typeof value !== "string" || value === "" || value.startsWith("/") || value.includes("..")) ||
    guard.paths.some((value, index) => index > 0 && guard.paths[index - 1] >= value)) {
  throw new Error("invalid kernel-affecting path registry");
}

const git = (...args) => execFileSync("git", args, { cwd: repositoryRoot, encoding: "utf8" }).trim();
const isKernelAffecting = (filename) => !filename.endsWith("_test.go") && guard.paths.some((entry) => entry.endsWith("/") ? filename.startsWith(entry) : filename === entry);
const assertVersioned = (label, files) => {
  if (files.some(isKernelAffecting) && !files.includes("kernel/VERSION")) {
    throw new Error(`${label} changes kernel semantics without kernel/VERSION: ${files.filter(isKernelAffecting).join(", ")}`);
  }
};

const guardIntroduction = git("log", "--diff-filter=A", "--format=%H", "--", "kernel/affecting-paths.json").split("\n").filter(Boolean).at(-1);
if (guardIntroduction) {
  const commits = git("rev-list", "--reverse", `${guardIntroduction}^..HEAD`).split("\n").filter(Boolean);
  for (const commit of commits) {
    const files = git("diff-tree", "--no-commit-id", "--name-only", "-r", "--root", commit).split("\n").filter(Boolean);
    assertVersioned(`commit ${commit}`, files);
  }
}

const worktreeFiles = git("status", "--porcelain=v1", "--untracked-files=all").split("\n").filter(Boolean).flatMap((line) => {
  const filename = line.slice(3);
  return filename.includes(" -> ") ? filename.split(" -> ") : [filename];
});
assertVersioned("worktree", worktreeFiles);

console.log(`kernel version parity and history guard ok: ${source}`);
