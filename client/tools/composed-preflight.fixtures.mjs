import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";
import { leaves, parents } from "./observe-refresh-population.mjs";
import { runComposedPreflight } from "./test-game-ui-preflight.mjs";

const repositoryRoot = fileURLToPath(new URL("../../", import.meta.url));
const persistedParents = [
  "TestSessionRequestSchemaAdmissionIntegration", "TestSoulRecoveryErrorRegistryIntegration",
  "TestFiscalProjectionMatchesPersistedHarvestIntegration", "TestFirstContentEpochPersistedBoundaryIntegration",
  "TestResolveMinigameSessionIntegrationAtomicReplayAndFaults", "TestPublicBoardRankingAndPagesIntegration",
  "TestPublicRunEvidenceIntegration", "TestComposedGameserverPostgresSocketClearingAndGCIntegration",
  "TestComposedMinigameAPILifecycleUsesPinnedTenantResolverIntegration", "TestComposedGameserverExitVerificationAndBoardIntegration",
];
const witnesses = [
  "public SDK HTTP downloads → TypeScript replay: verified + receipt/event/gap/constants refusal; DB-pinned catalogs supplied: PASS",
  "historical verified HTTP evidence survives distinct active epoch/hash; DB-pinned replay and corruption refusal: PASS",
  "stored live Guild inputs → TypeScript replay: complete state/receipt/events + positive idle/zero-producer + omission/duplicate controls: PASS",
];
function refreshOutput(missing) {
  const rows = [];
  const event = (Action, Test) => rows.push({ Action, Package: "cloud-clicker/server/account", ...(Test ? { Test } : {}) });
  event("start");
  for (const parent of parents) {
    event("run", parent);
    for (const leaf of leaves.filter((name) => name.startsWith(`${parent}/`))) {
      if (leaf !== missing) { event("run", leaf); event("pass", leaf); }
    }
    event("pass", parent);
  }
  rows.push({ Action: "output", Package: "cloud-clicker/server/account", Output: "private-output-never-log" });
  event("pass");
  return rows.map((row) => JSON.stringify(row)).join("\n");
}
const persistedOutput = () => [...persistedParents.map((name) => `--- PASS: ${name} (0.01s)`), ...witnesses].join("\n");
function fixture({ missingRefresh, persisted = persistedOutput(), failAt, signalAt, resetFailure, launchFailure } = {}) {
  const calls = [], out = [], err = [], info = [];
  let makeCalls = 0, resets = 0;
  const run = (command, args, options) => {
    calls.push({ command, args, options });
    if (command === "docker") {
      resets++;
      return { status: resets === resetFailure ? 37 : 0, stdout: "", stderr: "reset diagnostic" };
    }
    makeCalls++;
    if (makeCalls === launchFailure) return { error: new Error("spawn failed"), status: null };
    return { status: makeCalls === signalAt ? null : makeCalls === failAt ? 37 : 0,
      signal: makeCalls === signalAt ? "SIGTERM" : null,
      stdout: makeCalls === 1 ? refreshOutput(missingRefresh) : persisted, stderr: "Go diagnostic" };
  };
  return { calls, out, err, info, execute: () => runComposedPreflight("/fixture", {
    run, out: (value) => out.push(value), err: (value) => err.push(value), info: (value) => info.push(value),
  }) };
}

test("preparation keeps both complete Go populations, cold flags, replay witnesses and schema resets", () => {
  const f = fixture(); f.execute();
  assert.deepEqual(f.calls.map((row) => row.command), ["docker", "make", "docker", "make", "docker"]);
  const [refresh, persisted] = f.calls.filter((row) => row.command === "make");
  assert.equal(refresh.args[1], "GO_PACKAGES=./account");
  assert.equal(persisted.args[1], "GO_PACKAGES=./account ./gameui ./production ./leaderboard ./gameserver");
  for (const row of [refresh, persisted]) {
    assert.match(row.args[2], /-count=1/u); assert.ok(row.args[2].endsWith(")$$'"));
    assert.equal(row.options.env.TEST_DATABASE_URL, "postgres://cloud_clicker:cloud_clicker_game_ui_test@127.0.0.1:55433/cloud_clicker_game_ui_test?sslmode=disable");
  }
  for (const parent of parents) assert.ok(refresh.args[2].includes(parent));
  for (const parent of persistedParents) assert.ok(persisted.args[2].includes(parent));
  assert.equal(persisted.options.env.TEST_PUBLIC_REPLAY_CLIENT, "1");
  assert.equal(f.info.length, 1); assert.match(f.info[0], /8\/8/u);
  assert.equal(f.out.join("").includes("private-output-never-log"), false);
  assert.equal(f.out.join(""), persistedOutput());
  assert.deepEqual(f.err, ["Go diagnostic"]);
});

test("a zero-exit incomplete refresh population stops before later preparation", () => {
  const f = fixture({ missingRefresh: leaves[0] });
  assert.throws(f.execute, /persisted refresh population failed/u);
  assert.equal(f.calls.length, 2); assert.equal(f.info.length, 0);
});

test("every existing persisted parent and replay-consumer witness remains mandatory", () => {
  for (const name of persistedParents) {
    const f = fixture({ persisted: persistedOutput().replace(`--- PASS: ${name} (0.01s)`, `--- SKIP: ${name} (0.01s)`) });
    assert.throws(f.execute, /did not execute and pass/u); assert.equal(f.calls.length, 4);
  }
  for (const witness of witnesses) {
    const f = fixture({ persisted: persistedOutput().replace(witness, "") });
    assert.throws(f.execute, /did not execute and pass/u); assert.equal(f.calls.length, 4);
  }
});

test("Go failures, signals and launch errors stop preparation", () => {
  for (const stage of [1, 2]) for (const kind of ["failAt", "signalAt", "launchFailure"]) {
    const f = fixture({ [kind]: stage });
    assert.throws(f.execute); assert.equal(f.calls.length, stage === 1 ? 2 : 4);
  }
});

test("every failed schema reset stops before the next operation", () => {
  for (const resetFailure of [1, 2, 3]) {
    const f = fixture({ resetFailure });
    assert.throws(f.execute, /composed test DB reset failed/u);
    assert.equal(f.calls.length, resetFailure * 2 - 1);
  }
});

function toolFixture(body) {
  const directory = mkdtempSync(path.join(tmpdir(), "composed-preparation-"));
  const bin = path.join(directory, "bin"), trace = path.join(directory, "trace.jsonl");
  mkdirSync(bin);
  const portModel = path.join(directory, "port-model.mjs");
  writeFileSync(portModel, `import net from 'node:net';
import { EventEmitter } from 'node:events';
import { syncBuiltinESMExports } from 'node:module';
net.createServer = () => {
  const server = new EventEmitter();
  server.listen = (_port, _host, ready) => {
    queueMicrotask(() => process.env.COMPOSED_TEST_PORT_OCCUPIED === '1'
      ? server.emit('error', Object.assign(new Error('address already in use'), { code: 'EADDRINUSE' })) : ready());
    return server;
  };
  server.close = (done) => queueMicrotask(done);
  return server;
};
syncBuiltinESMExports();\n`);
  const source = `#!${process.execPath}\n${body}`;
  for (const tool of ["node", "docker", "go", "vite", "make"]) writeFileSync(path.join(bin, tool), source, { mode: 0o700 });
  return { bin, trace, portModel, env: { ...process.env, PATH: `${bin}:${process.env.PATH}`, COMPOSED_TRACE: trace,
    COMPOSED_TEST_PORT_OCCUPIED: "0" },
    calls: () => { try { return readFileSync(trace, "utf8").trim().split("\n").filter(Boolean).map(JSON.parse); } catch (e) { if (e.code === "ENOENT") return []; throw e; } },
    cleanup: () => rmSync(directory, { recursive: true, force: true }) };
}

test("public standalone modes still stop on backend failure; invalid arguments reach no tools", () => {
  const tools = toolFixture(`const fs=require('node:fs'),path=require('node:path');
const name=path.basename(process.argv[1]);fs.appendFileSync(process.env.COMPOSED_TRACE,JSON.stringify([name,...process.argv.slice(2)])+'\\n');
process.exit(name==='make'?37:0);`);
  try {
    for (const args of [[], ["--tier2-fixture"], ["--soul-recovery-full"], ["--observe-manual-budget"]]) {
      writeFileSync(tools.trace, "");
      const result = spawnSync(process.execPath, ["--import", tools.portModel, "client/tools/test-game-ui-composed.mjs", ...args], { cwd: repositoryRoot, env: tools.env, encoding: "utf8" });
      assert.equal(result.status, 1); assert.match(result.stderr, /persisted refresh population failed/u);
      assert.deepEqual(tools.calls().map((row) => row[0]), ["docker", "make"]);
    }
    for (const script of ["test-game-ui-composed.mjs", "test-game-ui-preflight.mjs"]) {
      writeFileSync(tools.trace, "");
      const result = spawnSync(process.execPath, ["--import", tools.portModel, `client/tools/${script}`, "--unknown"], { cwd: repositoryRoot, env: tools.env, encoding: "utf8" });
      assert.equal(result.status, 1); assert.deepEqual(tools.calls(), []);
    }
  } finally { tools.cleanup(); }
});

test("occupied composed port prevents all preparation tool calls", () => {
  const tools = toolFixture(`const fs=require('node:fs'),path=require('node:path');
const name=path.basename(process.argv[1]);fs.appendFileSync(process.env.COMPOSED_TRACE,JSON.stringify([name,...process.argv.slice(2)])+'\\n');
process.exit(name==='make'?37:0);`);
  try {
    for (const [script, args] of [["test-game-ui-composed.mjs", []], ["test-game-ui-composed.mjs", ["--tier2-fixture"]],
      ["test-game-ui-preflight.mjs", []]]) {
      writeFileSync(tools.trace, "");
      const result = spawnSync(process.execPath, ["--import", tools.portModel, `client/tools/${script}`, ...args], {
        cwd: repositoryRoot, env: { ...tools.env, COMPOSED_TEST_PORT_OCCUPIED: "1" }, encoding: "utf8",
      });
      assert.equal(result.status, 1); assert.deepEqual(tools.calls(), []);
      assert.match(result.stderr, /port 18081 is not exclusively available/u);
    }
  } finally { tools.cleanup(); }
});

test("aggregate Make gates all browser journeys behind one preparation, including with -j", () => {
  const tools = toolFixture(`const fs=require('node:fs'),path=require('node:path');
const tool=path.basename(process.argv[1]),args=process.argv.slice(2);const label=tool==='node'?path.basename(args[0]):tool;
fs.appendFileSync(process.env.COMPOSED_TRACE,JSON.stringify([tool,...args])+'\\n');
process.exit(label===process.env.COMPOSED_FAIL_PROGRAM?37:0);`);
  // Use the real Make, including its recursive garden/N5 targets. All external
  // recipe tools are doubles; this is graph/failure evidence, not browser proof.
  rmSync(path.join(tools.bin, "make"));
  try {
    for (const fail of ["", "test-game-ui-preflight.mjs"]) {
      writeFileSync(tools.trace, "");
      const result = spawnSync("make", ["-j4", "test-game-ui-composed", `CLIENT_BIN=${tools.bin}`], {
        cwd: repositoryRoot, env: { ...tools.env, COMPOSED_FAIL_PROGRAM: fail }, encoding: "utf8",
      });
      const nodes = tools.calls().filter((row) => row[0] === "node" && row[1] !== "--test");
      const preparations = nodes.filter((row) => row[1] === "client/tools/test-game-ui-preflight.mjs");
      const journeys = nodes.filter((row) => row[1] === "client/tools/game-ui-composed-journey.mjs");
      const browserScripts = new Set(["game-ui-composed-journey.mjs", "test-cosmetic-composed.mjs", "test-cosmetic-n5.mjs", "test-garden-composed.mjs"]);
      const browsers = nodes.filter((row) => browserScripts.has(path.basename(row[1])));
      assert.equal(preparations.length, 1);
      if (fail) { assert.equal(result.status, 2); assert.equal(browsers.length, 0); }
      else {
        assert.equal(result.status, 0); assert.deepEqual(journeys.map((row) => row.slice(2)), [[], ["--tier2-fixture"]]);
        for (const browser of browsers) assert.ok(nodes.indexOf(preparations[0]) < nodes.indexOf(browser));
        assert.equal(nodes.some((row) => row[1] === "client/tools/test-game-ui-composed.mjs"), false);
        for (const script of ["test-cosmetic-composed.mjs", "test-cosmetic-n5.mjs", "test-garden-composed.mjs"]) {
          assert.ok(nodes.some((row) => row[1].endsWith(script)));
        }
      }
    }
  } finally { tools.cleanup(); }
});
