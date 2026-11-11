// Shared real-Postgres populations. Importing this module does not run them.
import { spawnSync } from "node:child_process";
import path from "node:path";
import { createServer as createTCPServer } from "node:net";
import { fileURLToPath } from "node:url";
import { parents as refreshTests, refreshPopulationObserver } from "./observe-refresh-population.mjs";

const root = path.resolve(fileURLToPath(new URL("../../", import.meta.url)));

export async function assertComposedPortAvailable() {
  await new Promise((resolve, reject) => {
    const probe = createTCPServer();
    probe.once("error", (error) => reject(new Error(`composed gameserver port 18081 is not exclusively available: ${error.message}`)));
    probe.listen(18081, "127.0.0.1", () => probe.close(resolve));
  });
}

export function resetComposedDatabase(repositoryRoot = root, run = spawnSync) {
  const result = run("docker", ["compose", "-f", "compose.game-ui-test.yml", "exec", "-T", "game-ui-postgres",
    "psql", "-v", "ON_ERROR_STOP=1", "-U", "cloud_clicker", "-d", "cloud_clicker_game_ui_test",
    "-c", "SET client_min_messages TO WARNING; DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;"],
  { cwd: repositoryRoot, encoding: "utf8" });
  if (result.error) throw result.error;
  if (result.status !== 0) throw new Error(`composed test DB reset failed: ${result.stderr || result.stdout}`);
}

export function runComposedPreflight(repositoryRoot = root, tools = {}) {
  const { run = spawnSync, out = (text) => process.stdout.write(text), err = (text) => process.stderr.write(text),
    info = (text) => console.info(text) } = tools;
  const testDatabaseURL = "postgres://cloud_clicker:cloud_clicker_game_ui_test@127.0.0.1:55433/cloud_clicker_game_ui_test?sslmode=disable";
  resetComposedDatabase(repositoryRoot, run);
  // Exercise the actual refresh API/repository before starting the browser. The
  // existing observer requires every case to pass and discards private Go output;
  // a package-level PASS with missing or skipped database tests is not success.
  const refreshChecks = run("make", ["test-go", "GO_PACKAGES=./account", `GO_TEST_FLAGS=-count=1 -json -run '^(${refreshTests.join("|")})$$'`], {
    cwd: repositoryRoot,
    env: { ...process.env, TEST_DATABASE_URL: testDatabaseURL },
    encoding: "utf8",
  });
  if (refreshChecks.error) throw refreshChecks.error;
  const refreshObserver = refreshPopulationObserver();
  for (const line of (refreshChecks.stdout ?? "").split("\n")) refreshObserver.consumeLine(line);
  const refreshResult = refreshObserver.finish(refreshChecks.status);
  if (!refreshResult.valid) {
    throw new Error(`persisted refresh population failed: ${JSON.stringify({
      completed: refreshResult.completed_leaves, expected: refreshResult.expected_leaves,
      missing: refreshResult.missing, skipped: refreshResult.skipped, failed: refreshResult.failed,
      errors: refreshResult.errors, exit: refreshChecks.status, signal: refreshChecks.signal,
    })}`);
  }
  info(`persisted refresh API population passed: ${refreshResult.completed_leaves}/${refreshResult.expected_leaves} cases (real Postgres, HTTP over net.Pipe; not browser renewal)`);
  resetComposedDatabase(repositoryRoot, run);
  const persistedTests = [
    "TestSessionRequestSchemaAdmissionIntegration",
    "TestSoulRecoveryErrorRegistryIntegration",
    "TestFiscalProjectionMatchesPersistedHarvestIntegration",
    "TestFirstContentEpochPersistedBoundaryIntegration",
    "TestResolveMinigameSessionIntegrationAtomicReplayAndFaults",
    "TestPublicBoardRankingAndPagesIntegration",
    "TestPublicRunEvidenceIntegration",
    "TestComposedGameserverPostgresSocketClearingAndGCIntegration",
    "TestComposedMinigameAPILifecycleUsesPinnedTenantResolverIntegration",
    "TestComposedGameserverExitVerificationAndBoardIntegration",
  ];
  const persistedChecks = run("make", ["test-go", "GO_PACKAGES=./account ./gameui ./production ./leaderboard ./gameserver", `GO_TEST_FLAGS=-count=1 -v -run '^(${persistedTests.join("|")})$$'`], {
    cwd: repositoryRoot,
    env: { ...process.env, TEST_DATABASE_URL: testDatabaseURL, TEST_PUBLIC_REPLAY_CLIENT: "1" },
    encoding: "utf8",
  });
  out(persistedChecks.stdout ?? "");
  err(persistedChecks.stderr ?? "");
  if (persistedChecks.status !== 0) {
    throw new Error(`persisted integration population failed (${persistedChecks.status})`);
  }
  for (const name of persistedTests) {
    if (!persistedChecks.stdout?.includes(`--- PASS: ${name} (`)) {
      throw new Error(`${name} did not execute and pass (${persistedChecks.status})`);
    }
  }
  if (!persistedChecks.stdout?.includes("public SDK HTTP downloads → TypeScript replay: verified + receipt/event/gap/constants refusal; DB-pinned catalogs supplied: PASS")) {
    throw new Error("public SDK/TypeScript replay integration did not execute and pass");
  }
  if (!persistedChecks.stdout?.includes("historical verified HTTP evidence survives distinct active epoch/hash; DB-pinned replay and corruption refusal: PASS")) {
    throw new Error("historical public evidence rollover integration did not execute and pass");
  }
  if (!persistedChecks.stdout?.includes("stored live Guild inputs → TypeScript replay: complete state/receipt/events + positive idle/zero-producer + omission/duplicate controls: PASS")) {
    throw new Error("stored live Guild/TypeScript replay integration did not execute and pass");
  }
  // Diagnostic streams/catalogs must not become the browser journey's epoch.
  resetComposedDatabase(repositoryRoot, run);
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  if (process.argv.length !== 2) throw new Error("unknown composed preparation arguments");
  await assertComposedPortAvailable();
  runComposedPreflight();
}
