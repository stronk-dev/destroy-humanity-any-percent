import { spawn, spawnSync } from "node:child_process";
import { createServer as createTCPServer } from "node:net";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { chromium } from "playwright";
import { createServer } from "vite";

const clientRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const repositoryRoot = path.resolve(clientRoot, "..");
const gameserverURL = "http://127.0.0.1:18081";
const uiURL = "http://localhost:5173";

const key = Buffer.alloc(32, 7).toString("base64");
const processErrors = [];

await new Promise((resolve, reject) => {
  const probe = createTCPServer();
  probe.once("error", (error) => reject(new Error(`composed gameserver port 18081 is not exclusively available: ${error.message}`)));
  probe.listen(18081, "127.0.0.1", () => probe.close(resolve));
});

const gameserverBinary = path.join(repositoryRoot, ".cache", "game-ui-gameserver");
const gameserverEnvironment = {
  ...process.env,
  CLOUD_CLICKER_ACTIVITY_BRACKET: "activity.standard",
  CLOUD_CLICKER_BOOTSTRAP_KEY: key,
  CLOUD_CLICKER_BOOTSTRAP_KEY_ID: "browser-fixture",
  CLOUD_CLICKER_JWT_KEY: key,
  CLOUD_CLICKER_REPOSITORY_ROOT: repositoryRoot,
  CLOUD_CLICKER_SERVER_ID: "01986666-b001-4000-8000-000000000001",
  DATABASE_URL: "postgres://cloud_clicker:cloud_clicker_game_ui_test@127.0.0.1:55433/cloud_clicker_game_ui_test?sslmode=disable",
  GOCACHE: path.join(repositoryRoot, ".cache", "go-build"),
  LISTEN_ADDR: "127.0.0.1:18081",
};
const buildGameserver = spawnSync("go", ["build", "-o", gameserverBinary, "./cmd/gameserver"], {
  cwd: path.join(repositoryRoot, "server"),
  env: gameserverEnvironment,
  encoding: "utf8",
});
if (buildGameserver.status !== 0) {
  throw new Error(`composed gameserver build failed (${buildGameserver.status}): ${buildGameserver.stderr || buildGameserver.stdout}`);
}
const gameserver = spawn(gameserverBinary, [], {
  cwd: repositoryRoot,
  env: gameserverEnvironment,
  stdio: ["ignore", "pipe", "pipe"],
});
gameserver.stdout.on("data", (value) => process.stdout.write(value));
gameserver.stderr.on("data", (value) => process.stderr.write(value));
gameserver.on("error", (error) => processErrors.push(error));

let vite;
let browser;

async function waitForReady() {
  const deadline = Date.now() + 60_000;
  while (Date.now() < deadline) {
    if (processErrors.length > 0 || gameserver.exitCode !== null) throw processErrors[0] ?? new Error(`gameserver exited ${gameserver.exitCode}`);
    try {
      const response = await fetch(`${gameserverURL}/readyz`);
      if (response.ok) {
        // A stale listener must not let a newly failed child masquerade as ready.
        await new Promise((resolve) => setTimeout(resolve, 100));
        if (processErrors.length > 0 || gameserver.exitCode !== null) throw processErrors[0] ?? new Error(`gameserver exited ${gameserver.exitCode}`);
        return;
      }
    } catch {
      // The listener is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 100));
  }
  throw new Error("composed gameserver did not become ready");
}

async function stopGameserver() {
  if (gameserverStopped()) return;
  gameserver.kill("SIGTERM");
  const deadline = Date.now() + 10_000;
  while (!gameserverStopped() && Date.now() < deadline) {
    await new Promise((resolve) => setTimeout(resolve, 50));
  }
  if (!gameserverStopped()) {
    gameserver.kill("SIGKILL");
    await new Promise((resolve) => gameserver.once("exit", resolve));
  }
}

function gameserverStopped() {
  return gameserver.exitCode !== null || gameserver.signalCode !== null;
}

function seedGateRequirement(founderID) {
  if (!/^[0-9a-f-]{36}$/u.test(founderID)) throw new Error("invalid Founder ID in server-side setup");
  const sql = `WITH current AS (
    SELECT revision.stream_id, revision.revision, revision.version, revision.state, revision.constants_hash
    FROM save_revisions revision
    JOIN save_streams stream ON stream.id = revision.stream_id
    WHERE stream.owner_kind = 'founder' AND stream.owner_id = '${founderID}' AND stream.scope = 'company' AND stream.archived_at IS NULL
    ORDER BY revision.revision DESC LIMIT 1
  ) INSERT INTO save_revisions(stream_id, revision, version, state, constants_hash)
    SELECT stream_id, revision + 1, version,
      jsonb_set(state, '{balances,company.cash}', to_jsonb('1e5'::text), false), constants_hash FROM current;`;
  const result = spawnSync("docker", ["compose", "-f", "compose.game-ui-test.yml", "exec", "-T", "game-ui-postgres",
    "psql", "-v", "ON_ERROR_STOP=1", "-U", "cloud_clicker", "-d", "cloud_clicker_game_ui_test", "-c", sql],
  { cwd: repositoryRoot, encoding: "utf8" });
  if (result.status !== 0 || !result.stdout.includes("INSERT 0 1")) {
    throw new Error(`server-side gate setup failed (${result.status}): ${result.stderr || result.stdout}`);
  }
}

async function waitForEnabledButton(page, name) {
  await page.waitForFunction((label) => [...document.querySelectorAll("button")]
    .some((button) => button.textContent?.trim() === label && !button.disabled), name, { timeout: 30_000 });
  return page.getByRole("button", { name, exact: true });
}

async function clickAppliedIntent(page, button, label) {
  const requestPromise = page.waitForRequest((request) =>
    new URL(request.url()).pathname === "/api/v1/intents", { timeout: 30_000 });
  await button.click();
  const request = await requestPromise;
  const response = await request.response();
  if (!response) throw new Error(`${label} visible intent request completed without a response`);
  const body = await response.json();
  if (response.status() !== 200 || body?.outcome !== "applied") {
    throw new Error(`${label} visible intent was not applied (${response.status()}): ${JSON.stringify(body)}`);
  }
  return body;
}

try {
  await waitForReady();
  vite = await createServer({
    configFile: path.join(clientRoot, "vite.config.ts"),
    root: clientRoot,
    server: {
      host: "127.0.0.1",
      port: 5173,
      strictPort: true,
      proxy: {
        "/api": { target: gameserverURL },
        "/connection": { target: gameserverURL, ws: true },
      },
    },
  });
  await vite.listen();
  browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });
  await page.addInitScript(() => {
    const BrowserWebSocket = globalThis.WebSocket;
    globalThis.__cloudClickerTestSockets = [];
    globalThis.WebSocket = class extends BrowserWebSocket {
      constructor(url, protocols) {
        super(url, protocols);
        globalThis.__cloudClickerTestSockets.push(this);
      }
    };
  });
  const pageErrors = [];
  const websocketFrames = [];
  page.on("pageerror", (error) => pageErrors.push(error));
  page.on("websocket", (socket) => socket.on("framesent", (event) => websocketFrames.push(String(event.payload))));
  await page.goto(uiURL, { waitUntil: "networkidle" });
  await page.getByRole("button", { name: "BEGIN ATTEMPT" }).click();
  await page.locator('main[data-surface="desk"]').waitFor({ state: "visible", timeout: 30_000 });
  await page.getByText(/You are visitor #\d+/u).waitFor({ state: "visible", timeout: 30_000 });
  const stored = await page.evaluate(() => ({
    bootstrap: localStorage.getItem("cloud-clicker.bootstrap-key.v1"),
    credentials: localStorage.getItem("cloud-clicker.credentials.v1"),
  }));
  if (stored.bootstrap !== null || stored.credentials === null) throw new Error("bootstrap credential handoff was not committed before navigation");
  const liveSnapshot = await page.evaluate(async () => {
    const storedCredentials = localStorage.getItem("cloud-clicker.credentials.v1");
    if (!storedCredentials) throw new Error("missing composed credentials");
    const parsed = JSON.parse(storedCredentials);
    const response = await fetch("/api/v1/founder/state", { headers: { Authorization: `Bearer ${parsed.accessToken}` } });
    return { body: await response.json(), status: response.status };
  });
  if (liveSnapshot.status !== 200 || liveSnapshot.body?.schema_version !== 3 || !Number.isSafeInteger(liveSnapshot.body?.founder_revision) || liveSnapshot.body.founder_revision < 1 ||
      liveSnapshot.body?.transitions?.cross_gate?.gate_id !== "gate.t0_to_t1" || liveSnapshot.body.transitions.cross_gate.eligible !== false || liveSnapshot.body?.transitions?.wind_down?.eligible !== false) {
    throw new Error("composed live Game UI v3 transition snapshot round trip failed");
  }
  const recoveryBefore = await page.evaluate(() => {
    const key = Object.keys(localStorage).find((candidate) => candidate.startsWith("cloud-clicker.transport.v1."));
    if (!key) throw new Error("production runtime did not persist transport positions");
    return { key, value: JSON.parse(localStorage.getItem(key)) };
  });
  const playerChannel = `player:${liveSnapshot.body.run.founder_id}`;
  if (!recoveryBefore.value[playerChannel]?.epoch || !Number.isSafeInteger(recoveryBefore.value[playerChannel]?.offset)) {
    throw new Error("production runtime player position is incomplete");
  }
  websocketFrames.length = 0;
  await page.evaluate(async () => {
    const socket = globalThis.__cloudClickerTestSockets.at(-1);
    if (!socket) throw new Error("missing production WebSocket");
    socket.close();
    if (socket.readyState !== WebSocket.CLOSED) await new Promise((resolve) => socket.addEventListener("close", resolve, { once: true }));
  });
  const parsedCredentials = JSON.parse(stored.credentials);
  const expectedRecoveryRevision = liveSnapshot.body.revision + 1;
  const recoverySnapshot = page.waitForResponse(async (response) => {
    if (new URL(response.url()).pathname !== "/api/v1/founder/state" || response.status() !== 200) return false;
    const body = await response.json();
    return body?.revision === expectedRecoveryRevision;
  }, { timeout: 30_000 });
  const missedIntent = await fetch(`${gameserverURL}/api/v1/intents`, {
    method: "POST",
    headers: { Authorization: `Bearer ${parsedCredentials.accessToken}`, "Content-Type": "application/json" },
    body: JSON.stringify({
      intent_id: "01985555-5555-7555-8555-555555555555",
      kind: "perform_manual_batch",
      expected_revision: liveSnapshot.body.revision,
      action_id: liveSnapshot.body.manual_action.action_id,
      count: 1,
      window_ms: 1,
    }),
  });
  const missedReceipt = await missedIntent.json();
  if (missedIntent.status !== 200 || missedReceipt.new_revision !== liveSnapshot.body.revision + 1) {
    throw new Error(`missed intent did not commit exactly once (${missedIntent.status})`);
  }
  await page.waitForFunction(({ key, channel, offset }) => {
    const raw = localStorage.getItem(key);
    if (!raw) return false;
    return JSON.parse(raw)[channel]?.offset > offset;
  }, { key: recoveryBefore.key, channel: playerChannel, offset: recoveryBefore.value[playerChannel].offset }, { timeout: 30_000 });
  const recoveredCommand = websocketFrames.map((frame) => {
    try { return JSON.parse(frame); } catch { return undefined; }
  }).find((frame) => frame?.subscribe?.channel === playerChannel && frame.subscribe.recover === true);
  if (recoveredCommand?.subscribe.epoch !== recoveryBefore.value[playerChannel].epoch || recoveredCommand.subscribe.offset !== recoveryBefore.value[playerChannel].offset) {
    throw new Error("production runtime did not reconnect from the persisted player position");
  }
  const refreshedResponse = await recoverySnapshot;
  const refreshed = await refreshedResponse.json();
  if (refreshed.revision !== missedReceipt.new_revision) {
    throw new Error(`recovered receipt landed revision ${refreshed.revision}, expected ${missedReceipt.new_revision}`);
  }

  // GU-C28 permits ordinary server-side setup so Chromium proves the UI-owned
  // controls without replaying the already-proven two-hour policy or gaining a
  // clock/control endpoint. Every transition below originates from a visible
  // enabled button and reaches the server through runtime.ts.
  seedGateRequirement(liveSnapshot.body.run.founder_id);
  await page.reload({ waitUntil: "networkidle" });
  const crossGate = await waitForEnabledButton(page, "Move Into the Garage");
  await clickAppliedIntent(page, crossGate, "first cross-gate");
  const firstWindDown = await waitForEnabledButton(page, "Wind Down Company");
  await clickAppliedIntent(page, firstWindDown, "first wind-down");
  await page.locator('main[data-surface="run_end"]').waitFor({ state: "visible", timeout: 30_000 });
  await page.getByText("Your First Company Failed", { exact: true }).waitFor({ state: "visible", timeout: 30_000 });
  await page.getByRole("button", { name: "Start the Next Company", exact: true }).click();
  await page.locator('main[data-surface="desk"]').waitFor({ state: "visible", timeout: 30_000 });

  seedGateRequirement(liveSnapshot.body.run.founder_id);
  await page.reload({ waitUntil: "networkidle" });
  const secondCrossGate = await waitForEnabledButton(page, "Move Into the Garage");
  await clickAppliedIntent(page, secondCrossGate, "second cross-gate");
  const secondWindDown = await waitForEnabledButton(page, "Wind Down Company");
  await clickAppliedIntent(page, secondWindDown, "second wind-down");
  await page.locator('main[data-surface="run_end"]').waitFor({ state: "visible", timeout: 30_000 });
  await page.getByText("The Company Has Exited", { exact: true }).waitFor({ state: "visible", timeout: 30_000 });
  await page.getByRole("button", { name: "Start the Next Company", exact: true }).click();
  await page.locator('main[data-surface="desk"]').waitFor({ state: "visible", timeout: 30_000 });
  if (pageErrors.length > 0) throw new AggregateError(pageErrors, "composed browser path emitted page errors");
  console.log("composed Game UI v3 transitions + both terminal states + next-run continuation + WebSocket recovery: PASS");
  await page.goto("about:blank");
  await new Promise((resolve) => setTimeout(resolve, 100));
} finally {
  await browser?.close();
  await vite?.close();
  await stopGameserver();
}
