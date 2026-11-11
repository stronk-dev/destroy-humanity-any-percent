import assert from "node:assert/strict";
import { spawn, spawnSync } from "node:child_process";
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { setTimeout as delay } from "node:timers/promises";
import test from "node:test";

const root = fileURLToPath(new URL("../../", import.meta.url));
function fixture(failure = "") {
  const directory = mkdtempSync(path.join(os.tmpdir(), "composed-ci-"));
  const trace = path.join(directory, "trace.jsonl");
  if (failure === "uuid") writeFileSync(path.join(directory, "node"), "#!/bin/sh\nexit 37\n", { mode: 0o755 });
  writeFileSync(path.join(directory, "docker"), `#!${process.execPath}
const fs = require("node:fs");
const args = process.argv.slice(2);
const phase = args.find(value => ["build", "run", "down"].includes(value));
fs.appendFileSync(process.env.COMPOSE_CI_TRACE, JSON.stringify({phase, args,
  project: process.env.COMPOSE_PROJECT_NAME, root: process.env.COMPOSE_REPOSITORY_ROOT}) + "\\n");
if (phase === "run" && process.env.COMPOSE_CI_FAILURE === "signal") setInterval(() => {}, 1000);
else if (process.env.COMPOSE_CI_FAILURE.split(",").includes(phase)) {
  console.error("seeded " + phase + " failure"); process.exit(37);
}
`, { mode: 0o755 });
  return {
    directory,
    env: { ...process.env, PATH: `${directory}${path.delimiter}${process.env.PATH}`,
      COMPOSE_CI_TRACE: trace, COMPOSE_CI_FAILURE: failure },
    rows() { try { return readFileSync(trace, "utf8").trim().split("\n").filter(Boolean).map(JSON.parse); }
      catch (error) { if (error.code === "ENOENT") return []; throw error; } },
    close() { rmSync(directory, { recursive: true, force: true }); },
  };
}
function invoke(failure) {
  const subject = fixture(failure);
  try {
    const result = spawnSync("make", ["test-game-ui-composed-ci"], { cwd: root,
      env: subject.env, encoding: "utf8", timeout: 10_000 });
    assert.ifError(result.error);
    return { ...result, rows: subject.rows() };
  } finally { subject.close(); }
}

test("real Make owns one unique project, builds then runs full lane and cleans it", () => {
  const result = invoke("");
  assert.equal(result.status, 0, result.stderr);
  assert.deepEqual(result.rows.map(row => row.phase), ["build", "run", "down"]);
  const [first] = result.rows;
  assert.match(first.project, /^cloud-clicker-composed-ci-[0-9a-f-]{36}$/u);
  for (const row of result.rows) {
    assert.equal(row.project, first.project);
    assert.equal(row.root, path.resolve(root));
    assert.deepEqual(row.args.slice(0, 5), ["compose", "-f", "compose.game-ui-test.yml", "-f", "compose.game-ui-ci.yml"]);
  }
  assert.deepEqual(result.rows[1].args.slice(5), ["run", "--rm", "composed-linux"]);
  assert.deepEqual(result.rows[2].args.slice(5), ["down", "--remove-orphans"]);
  assert.notEqual(invoke("").rows[0].project, first.project);
});

test("failed UUID generation refuses before any Docker call", () => {
  const result = invoke("uuid");
  assert.equal(result.status, 2);
  assert.deepEqual(result.rows, []);
});

for (const [failure, phases] of [["build", ["build", "down"]], ["run", ["build", "run", "down"]],
  ["down", ["build", "run", "down"]], ["run,down", ["build", "run", "down"]]]) {
  test(`real Make propagates ${failure} failure and attempts owned cleanup`, () => {
    const result = invoke(failure);
    assert.equal(result.status, 2);
    assert.deepEqual(result.rows.map(row => row.phase), phases);
    for (const phase of failure.split(",")) assert.match(result.stderr, new RegExp(`seeded ${phase} failure`, "u"));
  });
}

test("terminating the actual Make process group cleans only its owned project", async () => {
  const subject = fixture("signal");
  const child = spawn("make", ["test-game-ui-composed-ci"], { cwd: root, env: subject.env,
    detached: true, stdio: "ignore" });
  const exited = new Promise((resolve, reject) => { child.once("exit", (code, signal) => resolve({ code, signal })); child.once("error", reject); });
  async function until(predicate) {
    const deadline = Date.now() + 5_000;
    while (!predicate()) {
      assert.ok(Date.now() < deadline, `process control did not complete: ${JSON.stringify(subject.rows())}`);
      await delay(10);
    }
  }
  try {
    await until(() => subject.rows().some(row => row.phase === "run"));
    process.kill(-child.pid, "SIGTERM");
    const result = await exited;
    assert.notEqual(result.code, 0);
    await until(() => subject.rows().some(row => row.phase === "down"));
    const rows = subject.rows();
    assert.deepEqual(rows.map(row => row.phase), ["build", "run", "down"]);
    assert.ok(rows.every(row => row.project === rows[0].project));
  } finally {
    try { process.kill(-child.pid, "SIGKILL"); } catch (error) { if (error.code !== "ESRCH") throw error; }
    subject.close();
  }
});

test("container entry refuses selectors before launching any setup or test", () => {
  const subject = fixture("");
  try {
    const result = spawnSync("bash", ["client/tools/test-game-ui-ci.sh", "--skip-preflight"], {
      cwd: root, env: subject.env, encoding: "utf8" });
    assert.equal(result.status, 2);
    assert.match(result.stderr, /expected no arguments/u);
    assert.deepEqual(subject.rows(), []);
  } finally { subject.close(); }
});
