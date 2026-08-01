import { execFileSync } from "node:child_process";
import { cpSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
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

  console.log("kernel history guard adversarial fixtures ok");
} finally {
  rmSync(root, { recursive: true, force: true });
}
