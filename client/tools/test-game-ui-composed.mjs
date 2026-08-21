import { spawn } from "node:child_process";
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
const gameserver = spawn("go", ["run", "./cmd/gameserver"], {
  cwd: path.join(repositoryRoot, "server"),
  detached: true,
  env: {
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
  },
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
      if (response.ok) return;
    } catch {
      // The listener is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 100));
  }
  throw new Error("composed gameserver did not become ready");
}

async function stopGameserver() {
  if (gameserverStopped()) return;
  signalGameserver("SIGTERM");
  await Promise.race([
    gameserverStopped() ? Promise.resolve() : new Promise((resolve) => gameserver.once("exit", resolve)),
    new Promise((resolve) => setTimeout(resolve, 10_000)),
  ]);
  if (!gameserverStopped()) signalGameserver("SIGKILL");
}

function gameserverStopped() {
  return gameserver.exitCode !== null || gameserver.signalCode !== null;
}

function signalGameserver(signal) {
  try {
    process.kill(-gameserver.pid, signal);
  } catch (error) {
    if (error?.code === "ESRCH") return;
    if (error?.code !== "EPERM") throw error;
    if (!gameserver.kill(signal) && !gameserverStopped()) throw error;
  }
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
  let captureRecoverySnapshot = false;
  let resolveRecoverySnapshot;
  const recoverySnapshot = new Promise((resolve) => { resolveRecoverySnapshot = resolve; });
  page.on("pageerror", (error) => pageErrors.push(error));
  page.on("websocket", (socket) => socket.on("framesent", (event) => websocketFrames.push(String(event.payload))));
  page.on("response", async (response) => {
    if (!captureRecoverySnapshot || new URL(response.url()).pathname !== "/api/v1/founder/state") return;
    captureRecoverySnapshot = false;
    resolveRecoverySnapshot({ body: await response.json(), status: response.status() });
  });
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
  if (liveSnapshot.status !== 200 || liveSnapshot.body?.schema_version !== 2 || !Number.isSafeInteger(liveSnapshot.body?.founder_revision) || liveSnapshot.body.founder_revision < 1) {
    throw new Error("composed live Game UI v2 snapshot round trip failed");
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
  captureRecoverySnapshot = true;
  const parsedCredentials = JSON.parse(stored.credentials);
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
  const refreshed = await Promise.race([
    recoverySnapshot,
    new Promise((_, reject) => setTimeout(() => reject(new Error("missed receipt did not trigger authoritative browser refresh")), 30_000)),
  ]);
  if (refreshed.status !== 200 || refreshed.body?.revision !== missedReceipt.new_revision) {
    throw new Error("recovered receipt did not land the browser on the committed revision");
  }
  if (pageErrors.length > 0) throw new AggregateError(pageErrors, "composed browser path emitted page errors");
  console.log("composed Game UI bootstrap + snapshot + WebSocket missed-receipt recovery: PASS");
  await page.goto("about:blank");
  await new Promise((resolve) => setTimeout(resolve, 100));
} finally {
  await browser?.close();
  await vite?.close();
  await stopGameserver();
}
