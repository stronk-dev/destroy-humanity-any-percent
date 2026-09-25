import { spawn, spawnSync } from "node:child_process";
import { mkdirSync } from "node:fs";
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
mkdirSync(path.dirname(gameserverBinary), { recursive: true });
const gameserverEnvironment = {
  ...process.env,
  CLOUD_CLICKER_ACTIVITY_BRACKET: "activity.standard",
  CLOUD_CLICKER_BOOTSTRAP_KEY: key,
  CLOUD_CLICKER_BOOTSTRAP_KEY_ID: "browser-fixture",
  CLOUD_CLICKER_CURSOR_KEY: key,
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
  await page.waitForFunction(async (label) => {
    const button = [...document.querySelectorAll("button")]
      .find((candidate) => candidate.textContent?.trim() === label && !candidate.disabled);
    if (!button) return false;
    await new Promise(requestAnimationFrame);
    await new Promise(requestAnimationFrame);
    const style = getComputedStyle(button);
    const bounds = button.getBoundingClientRect();
    return button.isConnected && !button.disabled && button.textContent?.trim() === label &&
      style.display !== "none" && style.visibility !== "hidden" && bounds.width > 0 && bounds.height > 0;
  }, name, { timeout: 30_000 });
  return name;
}

async function clickAppliedIntent(page, buttonLabel, label) {
  return clickAppliedIntentChoice(page, [buttonLabel], label);
}

async function clickAppliedIntentChoice(page, buttonLabels, label) {
  const requestPromise = page.waitForRequest((request) =>
    new URL(request.url()).pathname === "/api/v1/intents", { timeout: 30_000 });
  let request;
  let clickedLabel;
  try {
    [request, clickedLabel] = await Promise.all([requestPromise, page.evaluate(async (expectedLabels) => {
      const deadline = performance.now() + 30_000;
      while (performance.now() < deadline) {
        const main = document.querySelector("main");
        if (main?.getAttribute("aria-busy") === "false") {
          for (const expectedLabel of expectedLabels) {
            const element = [...document.querySelectorAll("button")]
              .find((candidate) => candidate.textContent?.trim() === expectedLabel);
            if (!element || element.disabled) continue;
            const style = getComputedStyle(element);
            const bounds = element.getBoundingClientRect();
            if (style.display !== "none" && style.visibility !== "hidden" && bounds.width > 0 && bounds.height > 0) {
              element.click();
              return expectedLabel;
            }
          }
        }
        await new Promise(requestAnimationFrame);
      }
      throw new Error(`${expectedLabels.join(" or ")} did not become atomically actionable`);
    }, buttonLabels)]);
  } catch (error) {
    const state = await page.evaluate(() => ({
      surface: document.querySelector("main")?.getAttribute("data-surface"),
      ariaBusy: document.querySelector("main")?.getAttribute("aria-busy"),
      buttons: [...document.querySelectorAll("button")].map((candidate) => ({ disabled: candidate.disabled, text: candidate.textContent?.trim() })),
      alerts: [...document.querySelectorAll('[role="alert"]')].map((candidate) => candidate.textContent?.trim()),
    }));
    throw new Error(`${label} emitted no intent request; rendered state=${JSON.stringify(state)}`, { cause: error });
  }
  const response = await request.response();
  if (!response) throw new Error(`${label} visible intent request completed without a response`);
  const body = await response.json();
  if (response.status() !== 200 || body?.outcome !== "applied") {
    throw new Error(`${label} visible intent was not applied (${response.status()}): ${JSON.stringify(body)}`);
  }
  await page.waitForFunction(() => document.querySelector("main")?.getAttribute("aria-busy") === "false", undefined, { timeout: 30_000 });
  return { body, clickedLabel, intent: request.postDataJSON() };
}

function receiptCoordinate(result) {
  return {
    clicked_label: result.clickedLabel,
    intent_kind: result.intent?.kind,
    intent_id: result.body?.intent_id,
    outcome: result.body?.outcome,
    new_revision: result.body?.new_revision,
    founder_revision: result.body?.founder_revision,
  };
}

function receivedLifecycleKinds(frames) {
  const kinds = ["exit_offer_spawned", "exit_offer_resolved", "run_ended", "run_started", "receipt"];
  return frames.flatMap((frame) => kinds.filter((kind) => frame.includes(`\\"kind\\":\\"${kind}\\"`) || frame.includes(`"kind":"${kind}"`)));
}

function receivedPlayerCoordinates(frames, channel) {
  const coordinates = [];
  const record = (publication) => {
    if (!publication || typeof publication !== "object") return;
    let envelope = publication.data;
    if (typeof envelope === "string") {
      try { envelope = JSON.parse(envelope); } catch { return; }
    }
    if (envelope?.ch !== channel) return;
    coordinates.push({ envelope_kind: envelope.kind, event_kind: envelope.payload?.kind, offset: publication.offset, revision: envelope.rev });
  };
  for (const frame of frames) {
    for (const line of String(frame).split("\n").filter(Boolean)) {
      let value;
      try { value = JSON.parse(line); } catch { continue; }
      if (value?.push?.channel === channel) record(value.push.pub);
      if (value?.subscribe && Array.isArray(value.subscribe.publications)) value.subscribe.publications.forEach(record);
    }
  }
  return coordinates;
}

function uuidV7() {
  const bytes = crypto.getRandomValues(new Uint8Array(16));
  let timestamp = Date.now();
  for (let index = 5; index >= 0; index -= 1) { bytes[index] = timestamp % 256; timestamp = Math.floor(timestamp / 256); }
  bytes[6] = (bytes[6] & 0x0f) | 0x70;
  bytes[8] = (bytes[8] & 0x3f) | 0x80;
  const hex = [...bytes].map((value) => value.toString(16).padStart(2, "0")).join("");
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}

async function founderIntent(accessToken, body) {
  const response = await fetch(`${gameserverURL}/api/v1/intents`, {
    method: "POST",
    headers: { Authorization: `Bearer ${accessToken}`, "Content-Type": "application/json" },
    body: JSON.stringify({ intent_id: uuidV7(), ...body }),
  });
  return { status: response.status, body: await response.json() };
}

async function founderState(accessToken) {
  const response = await fetch(`${gameserverURL}/api/v1/founder/state`, { headers: { Authorization: `Bearer ${accessToken}` } });
  if (response.status !== 200) throw new Error(`founder state read failed (${response.status})`);
  return response.json();
}

async function unlockPitchWithFiscalIntents(accessToken) {
  // The composed Fiscal clock guarantees a harvest 200 ms after the period
  // opens; wait past it, then spend the credit on the Pitch unlock.
  await new Promise((resolve) => setTimeout(resolve, 400));
  let state = await founderState(accessToken);
  const harvest = await founderIntent(accessToken, { kind: "harvest_fiscal_period", expected_revision: state.founder_revision });
  if (harvest.status !== 200 || harvest.body?.outcome !== "applied") throw new Error(`fiscal harvest was not applied (${harvest.status}): ${JSON.stringify(harvest.body)}`);
  state = await founderState(accessToken);
  const spend = await founderIntent(accessToken, { kind: "spend_fiscal_credit", expected_revision: state.founder_revision, target: { kind: "unlock", unlock_id: "minigame.pitch" } });
  if (spend.status !== 200 || spend.body?.outcome !== "applied") throw new Error(`fiscal Pitch unlock was not applied (${spend.status}): ${JSON.stringify(spend.body)}`);
}

async function playPitchThroughUI(page, accessToken) {
  const minigameResponses = [];
  const snapshotRevisions = [];
  page.on("response", async (response) => {
    const pathname = new URL(response.url()).pathname;
    if (pathname === "/api/v1/founder/state" && response.status() === 200) {
      try { snapshotRevisions.push((await response.json()).revision); } catch { snapshotRevisions.push(undefined); }
      return;
    }
    if (!pathname.startsWith("/api/v1/minigames/")) return;
    try { minigameResponses.push({ pathname, status: response.status(), body: await response.json() }); }
    catch { minigameResponses.push({ pathname, status: response.status(), invalid_json: true }); }
  });
  const nav = page.getByRole("button", { name: "The Pitch", exact: true });
  await nav.click();
  await page.locator('main[data-surface="minigame_session"]').waitFor({ state: "visible", timeout: 30_000 });

  // Locked first: the real server answers 409 not_eligible/fiscal_unlock_required
  // and the surface stays on the launcher with the typed reason.
  const lockedCreate = page.waitForResponse((response) => new URL(response.url()).pathname === "/api/v1/minigames/pitch/sessions", { timeout: 30_000 });
  await page.getByRole("button", { name: "Start a pitch", exact: true }).click();
  const locked = await lockedCreate;
  const lockedBody = await locked.json();
  if (locked.status() !== 409 || lockedBody.category !== "not_eligible" || lockedBody.detail !== "fiscal_unlock_required") {
    throw new Error(`locked Pitch create was not the typed fiscal rejection (${locked.status()}): ${JSON.stringify(lockedBody)}`);
  }
  await page.getByText("Locked. Unlock it with Fiscal credit first.", { exact: true }).waitFor({ state: "visible", timeout: 30_000 });

  await unlockPitchWithFiscalIntents(accessToken);
  const create = page.waitForResponse((response) => new URL(response.url()).pathname === "/api/v1/minigames/pitch/sessions", { timeout: 30_000 });
  await page.getByRole("button", { name: "Start a pitch", exact: true }).click();
  const created = await create;
  if (created.status() !== 200) throw new Error(`unlocked Pitch create failed (${created.status()}): ${JSON.stringify(await created.json())}`);

  let commands = 0;
  const terminal = page.getByText("Credited to the company", { exact: false });
  for (let step = 0; step < 64; step += 1) {
    if (await terminal.isVisible()) break;
    const hand = page.locator("fieldset input[type=checkbox]");
    await page.waitForFunction(() => document.querySelector("fieldset input[type=checkbox]") !== null ||
      [...document.querySelectorAll("button")].some((button) => button.textContent?.trim() === "Close the shop" && !button.disabled) ||
      document.body.textContent?.includes("Credited to the company"), undefined, { timeout: 30_000 });
    if (await terminal.isVisible()) break;
    const commandResponse = page.waitForResponse((response) => /^\/api\/v1\/minigames\/sessions\/[^/]+\/commands$/u.test(new URL(response.url()).pathname), { timeout: 30_000 });
    if (await hand.count() > 0) {
      // Keyboard-only selection: focus each native checkbox and press Space.
      const count = Math.min(4, await hand.count());
      for (let index = 0; index < count; index += 1) {
        await hand.nth(index).focus();
        await page.keyboard.press("Space");
      }
      const selected = await page.locator("fieldset input[type=checkbox]:checked").count();
      if (selected !== count) throw new Error(`keyboard selection checked ${selected} of ${count} cards`);
      await page.getByRole("button", { name: "Pitch these cards", exact: true }).focus();
      await page.keyboard.press("Enter");
    } else {
      await page.getByRole("button", { name: "Close the shop", exact: true }).focus();
      await page.keyboard.press("Enter");
    }
    const response = await commandResponse;
    if (response.status() !== 200) throw new Error(`Pitch command failed (${response.status()}): ${JSON.stringify(await response.json())}`);
    commands += 1;
    await page.waitForFunction(() => document.querySelector("main section[aria-busy]")?.getAttribute("aria-busy") !== "true", undefined, { timeout: 30_000 });
  }
  if (!await terminal.isVisible()) throw new Error(`Pitch did not reach a terminal receipt within the step bound; responses=${JSON.stringify(minigameResponses.map((row) => [row.pathname, row.status]))}`);
  const final = minigameResponses.filter((row) => row.status === 200 && row.body?.status === "resolved").at(-1);
  const receipt = final?.body?.resolution_receipt;
  if (!receipt || receipt.outcome !== "applied" || receipt.credited_resource_id !== "company.cash") throw new Error(`terminal response carried no applied receipt: ${JSON.stringify(final)}`);
  // The host refreshes the authoritative Game UI snapshot exactly once after
  // the terminal receipt; it must reflect the resolution's Company revision.
  const deadline = Date.now() + 30_000;
  while (!snapshotRevisions.some((revision) => Number.isSafeInteger(revision) && revision >= receipt.company_revision) && Date.now() < deadline) {
    await new Promise((resolve) => setTimeout(resolve, 50));
  }
  const refreshedRevision = snapshotRevisions.filter((revision) => Number.isSafeInteger(revision) && revision >= receipt.company_revision).at(0);
  if (refreshedRevision === undefined) {
    const state = await founderState(accessToken);
    throw new Error(`no post-terminal Game UI refresh reached company revision ${receipt.company_revision}; UI snapshots=${JSON.stringify(snapshotRevisions)} server=${state.revision}`);
  }
  const current = await fetch(`${gameserverURL}/api/v1/minigames/sessions/current`, { headers: { Authorization: `Bearer ${accessToken}` } }).then((response) => response.json());
  if (current.kind !== "none") throw new Error(`resolved Pitch session is still current: ${JSON.stringify(current)}`);
  await page.getByRole("button", { name: "Back to the desk", exact: true }).click();
  await page.locator('main[data-surface="desk"]').waitFor({ state: "visible", timeout: 30_000 });
  return { commands, credited: receipt.credited_delta, companyRevision: receipt.company_revision, refreshedRevision };
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
  // The composed binary mounts the unauthenticated public read surface
  // beside the account API (API Foundation C10); prove it through the proxy.
  const publicEpochs = await fetch(`${uiURL}/api/public/v1/epochs?limit=1`);
  const publicEpochsBody = await publicEpochs.json();
  if (publicEpochs.status !== 200 || !publicEpochs.headers.get("x-request-id") || !publicEpochs.headers.get("etag") ||
      !Array.isArray(publicEpochsBody.items) || publicEpochsBody.items.length !== 1 || publicEpochsBody.items[0].ended_at !== null) {
    throw new Error(`composed public epochs failed: ${publicEpochs.status} ${JSON.stringify(publicEpochsBody)}`);
  }
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
  const websocketReceivedFrames = [];
  const snapshotRevisions = [];
  page.on("pageerror", (error) => pageErrors.push(error));
  page.on("response", async (response) => {
    if (new URL(response.url()).pathname !== "/api/v1/founder/state" || response.status() !== 200) return;
    try {
      const body = await response.json();
      snapshotRevisions.push({ founder_revision: body?.founder_revision, revision: body?.revision, run_seq: body?.run?.run_seq });
    } catch { snapshotRevisions.push({ invalid_json: true }); }
  });
  page.on("websocket", (socket) => {
    socket.on("framesent", (event) => websocketFrames.push(String(event.payload)));
    socket.on("framereceived", (event) => websocketReceivedFrames.push(String(event.payload)));
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
  const features = liveSnapshot.body?.features;
  const doom = features?.meters?.meters?.find((row) => row.meter_id === "doom.probability");
  if (!features || features.active_play !== null || features.pets !== null || features.meters?.meters?.length !== 11 || doom?.value !== 50 ||
      !Array.isArray(features.achievements?.rows) || features.achievements.rows.length === 0 || !Number.isSafeInteger(features.fiscal?.credit) ||
      features.minigames?.rows?.find((row) => row.minigame_id === "pitch")?.unlocked !== false ||
      !liveSnapshot.body.facts.some((fact) => fact.fact_id === "feature.fiscal" && fact.value === true)) {
    throw new Error(`composed live Game UI v4 features arms are not the pinned live systems: ${JSON.stringify(features)}`);
  }
  if (liveSnapshot.status !== 200 || liveSnapshot.body?.schema_version !== 4 || !Number.isSafeInteger(liveSnapshot.body?.founder_revision) || liveSnapshot.body.founder_revision < 1 ||
      liveSnapshot.body?.transitions?.cross_gate?.gate_id !== "gate.t0_to_t1" || liveSnapshot.body.transitions.cross_gate.eligible !== false || liveSnapshot.body?.transitions?.wind_down?.eligible !== false) {
    throw new Error("composed live Game UI v4 transition snapshot round trip failed");
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
  const firstExit = await clickAppliedIntent(page, firstWindDown, "first wind-down");
  try {
    await page.locator('main[data-surface="run_end"]').waitFor({ state: "visible", timeout: 30_000 });
  } catch (error) {
    const finalSurface = await page.locator("main").getAttribute("data-surface");
    const playerFrames = websocketReceivedFrames.filter((frame) => frame.includes(`player:${liveSnapshot.body.run.founder_id}`));
    throw new Error(`first terminal did not render; command=${JSON.stringify(receiptCoordinate(firstExit))} surface=${finalSurface} page_errors=${pageErrors.map(String).join(" | ")} received_kinds=${JSON.stringify(receivedLifecycleKinds(playerFrames))} player_coordinates=${JSON.stringify(receivedPlayerCoordinates(websocketReceivedFrames, playerChannel))} snapshots=${JSON.stringify(snapshotRevisions)}`, { cause: error });
  }
  await page.getByText("Your First Company Failed", { exact: true }).waitFor({ state: "visible", timeout: 30_000 });
  await page.getByRole("button", { name: "Start the Next Company", exact: true }).click();
  await page.locator('main[data-surface="desk"]').waitFor({ state: "visible", timeout: 30_000 });

  seedGateRequirement(liveSnapshot.body.run.founder_id);
  await page.reload({ waitUntil: "networkidle" });
  const secondCrossGate = await waitForEnabledButton(page, "Move Into the Garage");
  await clickAppliedIntent(page, secondCrossGate, "second cross-gate");
  const secondExit = await clickAppliedIntentChoice(page, ["Wind Down Company", "Decline"], "second wind-down or offer preemption");
  if (secondExit.clickedLabel === "Decline") {
    await page.evaluate(() => {
      const desk = [...document.querySelectorAll("button")].find((button) => button.textContent?.trim() === "The Desk");
      if (!desk || desk.disabled) throw new Error("Desk navigation unavailable after declining preempting offer");
      desk.click();
    });
    await page.locator('main[data-surface="desk"]').waitFor({ state: "visible", timeout: 30_000 });
    const secondWindDown = await waitForEnabledButton(page, "Wind Down Company");
    await clickAppliedIntent(page, secondWindDown, "second wind-down after declining offer");
  }
  try {
    await page.locator('main[data-surface="run_end"]').waitFor({ state: "visible", timeout: 30_000 });
  } catch (error) {
    const finalSurface = await page.locator("main").getAttribute("data-surface");
    const playerFrames = websocketReceivedFrames.filter((frame) => frame.includes(`player:${liveSnapshot.body.run.founder_id}`));
    throw new Error(`second terminal did not render; command=${JSON.stringify(receiptCoordinate(secondExit))} surface=${finalSurface} page_errors=${pageErrors.map(String).join(" | ")} received_kinds=${JSON.stringify(receivedLifecycleKinds(playerFrames))} player_coordinates=${JSON.stringify(receivedPlayerCoordinates(websocketReceivedFrames, playerChannel))} snapshots=${JSON.stringify(snapshotRevisions)}`, { cause: error });
  }
  await page.getByText("The Company Has Exited", { exact: true }).waitFor({ state: "visible", timeout: 30_000 });
  await page.getByRole("button", { name: "Start the Next Company", exact: true }).click();
  await page.locator('main[data-surface="desk"]').waitFor({ state: "visible", timeout: 30_000 });
  // MA AC5: The Pitch through the Game UI against the composed server. The
  // Fiscal unlock is bought with real player intents over the public intent
  // API (the composed epoch pins minigame.pitch at 3 credit); no Fiscal
  // surface exists yet, and no database row is written for eligibility.
  const pitch = await playPitchThroughUI(page, parsedCredentials.accessToken);
  if (pageErrors.length > 0) throw new AggregateError(pageErrors, "composed browser path emitted page errors");
  console.log(`composed Pitch surface: unlock via Fiscal intents, ${pitch.commands} UI commands, terminal receipt credited ${pitch.credited} at company revision ${pitch.companyRevision}, snapshot refreshed to ${pitch.refreshedRevision}: PASS`);
  console.log("composed Game UI v4 features + transitions + both terminal states + next-run continuation + WebSocket recovery: PASS");
  await page.goto("about:blank");
  await new Promise((resolve) => setTimeout(resolve, 100));
} finally {
  await browser?.close();
  await vite?.close();
  await stopGameserver();
}
