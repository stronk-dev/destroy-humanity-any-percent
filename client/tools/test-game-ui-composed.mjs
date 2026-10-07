import { spawn, spawnSync } from "node:child_process";
import { mkdirSync } from "node:fs";
import { createServer as createTCPServer } from "node:net";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { chromium } from "playwright";
import { createServer } from "vite";
import { assertOpportunityClaimEffect, assertOpportunityReadStatus } from "./opportunity-claim-proof.mjs";

const clientRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const repositoryRoot = path.resolve(clientRoot, "..");
const gameserverURL = "http://127.0.0.1:18081";
const uiURL = "http://localhost:5173";

const key = Buffer.alloc(32, 7).toString("base64");
const processErrors = [];
let pitchOffersDeclined = 0;

// The composed target runs multiple fixture epochs in one named, ephemeral
// database. A later invocation must not inherit the previous witness's epoch.
function resetTestDatabase() {
  const resetDatabase = spawnSync("docker", ["compose", "-f", "compose.game-ui-test.yml", "exec", "-T", "game-ui-postgres",
    "psql", "-v", "ON_ERROR_STOP=1", "-U", "cloud_clicker", "-d", "cloud_clicker_game_ui_test",
    "-c", "SET client_min_messages TO WARNING; DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;"],
  { cwd: repositoryRoot, encoding: "utf8" });
  if (resetDatabase.status !== 0) throw new Error(`composed test DB reset failed: ${resetDatabase.stderr || resetDatabase.stdout}`);
}

await new Promise((resolve, reject) => {
  const probe = createTCPServer();
  probe.once("error", (error) => reject(new Error(`composed gameserver port 18081 is not exclusively available: ${error.message}`)));
  probe.listen(18081, "127.0.0.1", () => probe.close(resolve));
});

const testDatabaseURL = "postgres://cloud_clicker:cloud_clicker_game_ui_test@127.0.0.1:55433/cloud_clicker_game_ui_test?sslmode=disable";
resetTestDatabase();
const fiscalProjectionTest = "TestFiscalProjectionMatchesPersistedHarvestIntegration";
const fiscalProjection = spawnSync("make", ["test-go", "GO_PACKAGES=./gameui", `GO_TEST_FLAGS=-count=1 -v -run ^${fiscalProjectionTest}$$`], {
  cwd: repositoryRoot,
  env: { ...process.env, TEST_DATABASE_URL: testDatabaseURL },
  encoding: "utf8",
});
process.stdout.write(fiscalProjection.stdout ?? "");
process.stderr.write(fiscalProjection.stderr ?? "");
if (fiscalProjection.status !== 0 || !fiscalProjection.stdout?.includes(`--- PASS: ${fiscalProjectionTest} (`)) {
  throw new Error(`persisted Fiscal projection/harvest did not execute and pass (${fiscalProjection.status})`);
}
// Diagnostic streams/catalogs must not become the browser journey's epoch.
resetTestDatabase();

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
  DATABASE_URL: testDatabaseURL,
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

async function clickReadyPitchControl(page, control) {
  return page.evaluate(async (target) => {
    const deadline = performance.now() + 30_000;
    while (performance.now() < deadline) {
      const main = document.querySelector("main");
      if (main?.getAttribute("data-surface") === "offer_sheet") return "offer";
      let button;
      if (target === "start") button = [...(main?.querySelectorAll("button") ?? [])]
        .find((candidate) => candidate.textContent?.trim() === "Start a pitch");
      else button = [...document.querySelectorAll('section[aria-labelledby="fiscal-unlocks-heading"] li')]
        .find((item) => item.querySelector("h3")?.textContent?.trim() === "The Pitch")
        ?.querySelector("button");
      if (main?.getAttribute("aria-busy") === "false" && button && !button.disabled) {
        button.click();
        return "clicked";
      }
      await new Promise(requestAnimationFrame);
    }
    throw new Error(`Pitch ${target} control never became atomically actionable`);
  }, control);
}

async function declinePitchOffer(page, destination, surface) {
  const result = await clickAppliedIntent(page, "Decline", "Pitch offer preemption");
  if (result.intent?.kind !== "decline_exit_offer") throw new Error(`Pitch preemption did not use visible Decline: ${JSON.stringify(result.intent)}`);
  pitchOffersDeclined += 1;
  await page.getByRole("button", { name: destination, exact: true }).click();
  await page.locator(`main[data-surface="${surface}"]`).waitFor({ state: "visible", timeout: 30_000 });
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

// GS5-A4: DOM-only manual clicks on the real server until the Desk region
// projects an opportunity, then a DOM claim; the next authoritative snapshot
// must show the claim's effect (a live buff, or a credited lucky payout).
async function witnessOpportunityClaim(page) {
  const { parseCanonical } = await vite.ssrLoadModule("/src/numeric.ts");
  const expiredClaims = [];
  for (let attempt = 0; attempt < 60; attempt += 1) {
    const claimable = await page.evaluate(() => [...document.querySelectorAll("[data-region='desk.region.opportunity'] button")]
      .some((button) => button.textContent?.trim() === "Claim" && !button.disabled));
    if (!claimable) {
      await clickAppliedIntent(page, "Fix Computer", "GS5 manual click toward an opportunity spawn");
      await new Promise((resolve) => setTimeout(resolve, 250));
      continue;
    }
    let result;
    try { result = await clickAppliedIntent(page, "Claim", "GS5 opportunity claim"); }
    catch (error) {
      if (!/opportunity_expired|opportunity_not_pending/u.test(String(error?.message))) throw error;
      expiredClaims.push(String(error.message).slice(0, 200));
      if (expiredClaims.length >= 3) throw new Error(`GS5 claims kept expiring: ${JSON.stringify(expiredClaims)}`);
      continue;
    }
    const claim = result.body?.receipt?.opportunity;
    if (result.intent?.kind !== "claim_opportunity" || !claim?.effect_row_id) throw new Error(`GS5 claim receipt carried no opportunity evidence: ${JSON.stringify(result.body)}`);
    const successor = await page.evaluate(async () => {
      const parsed = JSON.parse(localStorage.getItem("cloud-clicker.credentials.v1"));
      const response = await fetch("/api/v1/founder/state", { headers: { Authorization: `Bearer ${parsed.accessToken}` } });
      return { status: response.status, snapshot: await response.json() };
    });
    assertOpportunityReadStatus(successor.status);
    const after = successor.snapshot;
    const proofBranch = assertOpportunityClaimEffect(result, after, parseCanonical);
    return { effect_row_id: claim.effect_row_id, proof_branch: proofBranch, revision: after.revision,
      manual_clicks: attempt - expiredClaims.length, expired_claims: expiredClaims.length };
  }
  throw new Error("GS5: no opportunity became claimable within 60 manual clicks (4/s, under the account limiter)");
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
    coordinates.push({ envelope_kind: envelope.kind, event_kind: envelope.payload?.kind, offset: publication.offset, revision: envelope.rev,
      achievement_id: envelope.payload?.payload?.achievement_id, run_seq: envelope.payload?.payload?.run_id?.run_seq });
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

async function founderState(accessToken) {
  const response = await fetch(`${gameserverURL}/api/v1/founder/state`, { headers: { Authorization: `Bearer ${accessToken}` } });
  if (response.status !== 200) throw new Error(`founder state read failed (${response.status})`);
  return response.json();
}

// GS2-A4: earn through a real first purchase, not SQL setup or a supplied
// earned-row fixture. Observe the live event, announcement and refreshed DOM.
async function witnessFirstPurchaseAchievement(page, frames, accessToken) {
  const achievementID = "achievement.generators_purchased_1", generatorID = "generator.beige_tower";
  const before = await founderState(accessToken);
  const row = before.features?.achievements?.rows.find((candidate) => candidate.achievement_id === achievementID);
  const generator = before.generators.find((candidate) => candidate.generator_id === generatorID);
  if (!row || row.earned !== null || !generator || generator.owned !== 0) {
    throw new Error("GS2-A4 requires a locked achievement and genuinely unowned first generator");
  }
  const { t } = await vite.ssrLoadModule("/src/copy/index.ts");
  const { eraForSnapshot } = await vite.ssrLoadModule("/src/game-ui/contracts.ts");
  const { GAME_UI_PRESENTATION } = await vite.ssrLoadModule("/src/game-ui/presentation.ts");
  const era = eraForSnapshot(before), title = t(row.copy_key, {}, era);
  const achievementNav = page.getByRole("button", { name: t("surface.achievements.title", {}, era), exact: true });
  const displayed = page.locator(".achievements li").filter({ has: page.getByRole("heading", { name: title, exact: true }) });
  await achievementNav.click();
  if (await displayed.getAttribute("data-earned") !== "none" ||
      await displayed.locator("strong").innerText() !== t("achievements.state.locked", {}, era)) {
    throw new Error("GS2-A4 first-purchase row was not visibly locked before buying");
  }
  await page.getByRole("button", { name: "The Desk", exact: true }).click();
  const generatorTitle = t(GAME_UI_PRESENTATION.generators.get(generatorID).title_key, {}, era);
  const buy = page.locator('section[aria-labelledby="generators-heading"] article')
    .filter({ has: page.getByRole("heading", { name: generatorTitle, exact: true }) })
    .getByRole("button", { name: t("desk.buy_one", {}, era), exact: true });
  // Use the player's manual action at the same 4/s cadence as GS5, with the
  // existing 30-second action guard. No fixture cash or direct intent call.
  const deadline = Date.now() + 30_000;
  let clicks = 0;
  while (!await buy.isEnabled() && Date.now() < deadline) {
    await clickAppliedIntent(page, "Fix Computer", "GS2 manual work toward first generator");
    clicks += 1;
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  if (!await buy.isEnabled()) throw new Error(`GS2-A4 first generator never became affordable after ${clicks} manual actions`);
  const frameStart = frames.length;
  const purchases = [];
  const recordPurchase = (request) => {
    if (request.method() === "POST" && new URL(request.url()).pathname === "/api/v1/intents" &&
        request.postDataJSON()?.kind === "buy_generator") purchases.push(request);
  };
  page.on("request", recordPurchase);
  const responseTask = page.waitForResponse((response) => response.request().method() === "POST" &&
    new URL(response.url()).pathname === "/api/v1/intents" && response.request().postDataJSON()?.kind === "buy_generator",
  { timeout: 30_000 }).then((value) => ({ value }), (error) => ({ error }));
  await buy.focus();
  await page.keyboard.press("Enter");
  const result = await responseTask;
  if (result.error) throw result.error;
  const response = result.value, intent = response.request().postDataJSON(), receipt = await response.json();
  if (purchases.length !== 1 || response.status() !== 200 || receipt.outcome !== "applied" || intent.generator_id !== generatorID ||
      intent.count?.mode !== "exact" || intent.count.value !== 1) throw new Error("GS2-A4 DOM first purchase was not applied exactly once");
  const announcement = t("achievements.earned_announcement", { achievement: title }, era);
  await page.locator("p.announcement[role=status]").getByText(announcement, { exact: true }).waitFor({ state: "visible", timeout: 30_000 });
  const events = receivedPlayerCoordinates(frames.slice(frameStart), `player:${before.run.founder_id}`)
    .filter((event) => event.event_kind === "achievement_earned.v1" && event.achievement_id === achievementID);
  if (events.length !== 1 || events[0].revision !== receipt.new_revision || events[0].run_seq !== before.run.run_seq) {
    throw new Error(`GS2-A4 achievement event did not bind to the purchased run/revision: ${JSON.stringify(events)}`);
  }
  await page.waitForFunction(() => document.querySelector("main")?.getAttribute("aria-busy") === "false", undefined, { timeout: 30_000 });
  await achievementNav.click();
  await displayed.locator("strong").getByText(t("achievements.state.earned_run", {}, era), { exact: true }).waitFor({ state: "visible", timeout: 30_000 });
  const after = await founderState(accessToken);
  page.off("request", recordPurchase);
  if (purchases.length !== 1 || await displayed.getAttribute("data-earned") !== "run" || after.revision < receipt.new_revision ||
      after.features.achievements.rows.find((candidate) => candidate.achievement_id === achievementID)?.earned !== "run" ||
      after.generators.find((candidate) => candidate.generator_id === generatorID)?.owned !== 1) {
    throw new Error("GS2-A4 refreshed earned DOM did not agree with persisted ownership/achievement");
  }
  await page.getByRole("button", { name: "The Desk", exact: true }).click();
  console.log(`composed GS2 achievement: ${clicks} manual actions → one native Enter purchase → live earned event/announcement → refreshed earned row at Company revision ${receipt.new_revision}: PASS`);
}

// GS0.2/GS1-A4: corrupt only the outgoing DOM-origin request, never supply a
// response or throw from a runtime double. Account/Production/Postgres decide
// the refusal and the real browser runtime must render and recover from it.
async function witnessFiscalServerRefusals(page, accessToken) {
  const { t } = await vite.ssrLoadModule("/src/copy/index.ts");
  const { eraForSnapshot } = await vite.ssrLoadModule("/src/game-ui/contracts.ts");
  const rejectedIDs = [];
  for (const arm of ["invalid", "stale"]) {
    const before = await founderState(accessToken), era = eraForSnapshot(before);
    if (before.founder_revision <= 1) throw new Error("stale Fiscal control requires an older positive revision");
    const requests = [], reads = [], diagnostics = [], routed = [];
    const observeRequest = (request) => {
      const pathname = new URL(request.url()).pathname;
      if (pathname === "/api/v1/intents" && request.method() === "POST") requests.push(request);
      if (pathname === "/api/v1/founder/state") reads.push(request);
    };
    const observeConsole = (message) => {
      if (message.type() === "error" && message.text().startsWith("game UI invariant:")) diagnostics.push(message.text());
    };
    const corruptRequest = async (route) => {
      const request = route.request(), original = request.postDataJSON();
      if (request.method() !== "POST" || original?.kind !== "harvest_fiscal_period") return route.continue();
      const sent = arm === "invalid" ? { ...original, expected_revision: 0 }
        : { ...original, expected_revision: original.expected_revision - 1 };
      routed.push({ original, sent });
      await route.continue({ postData: JSON.stringify(sent) });
    };
    page.on("request", observeRequest); page.on("console", observeConsole);
    await page.route("**/api/v1/intents", corruptRequest);
    try {
      const harvest = page.getByRole("button", { name: t("fiscal.harvest", {}, era), exact: true });
      const responseTask = page.waitForResponse((response) => new URL(response.url()).pathname === "/api/v1/intents" &&
        response.request().postDataJSON()?.kind === "harvest_fiscal_period", { timeout: 30_000 })
        .then((value) => ({ value }), (error) => ({ error }));
      await harvest.focus(); await page.keyboard.press("Enter");
      const result = await responseTask;
      if (result.error) throw result.error;
      const response = result.value, body = await response.json();
      if (routed.length !== 1 || requests.length !== 1 || routed[0].original.expected_revision !== before.founder_revision) {
        throw new Error(`Fiscal ${arm} refusal did not originate once at the current Founder revision`);
      }
      rejectedIDs.push(routed[0].original.intent_id);
      if (arm === "invalid") {
        if (response.status() !== 400 || JSON.stringify(Object.keys(body).sort()) !== '["category","detail"]' ||
            body.category !== "invalid" || body.detail !== "intent") throw new Error(`wrong real invalid-intent response: ${JSON.stringify(body)}`);
      } else if (response.status() !== 200 || body.outcome !== "rejected" || body.current_revision !== before.founder_revision ||
          body.intent_id !== routed[0].original.intent_id || body.rejection?.category !== "revision_conflict" ||
          body.rejection.detail !== "expected_revision") {
        throw new Error(`wrong real stale-revision receipt: ${JSON.stringify(body)}`);
      }
      const notice = t(arm === "invalid" ? "intent.rejection.unknown" : "intent.conflict", {}, era);
      await page.locator(".fiscal .intent-notice[role=status]").getByText(notice, { exact: true }).waitFor({ state: "visible", timeout: 30_000 });
      await page.waitForFunction(() => document.querySelector("main")?.getAttribute("aria-busy") === "false", undefined, { timeout: 30_000 });
      await page.evaluate(() => new Promise((resolve) => requestAnimationFrame(() => requestAnimationFrame(resolve))));
      const after = await founderState(accessToken);
      if (requests.length !== 1 || routed.length !== 1 || reads.length !== (arm === "stale" ? 1 : 0) ||
          JSON.stringify(diagnostics) !== JSON.stringify(arm === "invalid" ? ["game UI invariant: invalid intent response"] : []) ||
          after.revision !== before.revision || after.founder_revision !== before.founder_revision ||
          after.features.fiscal.credit !== before.features.fiscal.credit ||
          JSON.stringify(after.features.fiscal.period) !== JSON.stringify(before.features.fiscal.period)) {
        throw new Error(`Fiscal ${arm} refusal changed saved state, retried, or failed recovery: ${JSON.stringify({ requests: requests.length, reads: reads.length, diagnostics })}`);
      }
      if (await harvest.isDisabled() || await harvest.getAttribute("aria-disabled") === "true" ||
          !await harvest.evaluate((button) => document.activeElement === button)) {
        throw new Error(`Fiscal ${arm} refusal did not preserve an enabled, focused consent control`);
      }
      console.log(`composed Fiscal ${arm}: real HTTP${response.status()}, exact surface notice, ${reads.length} refresh, no automatic retry or persisted change: PASS`);
    } finally {
      await page.unroute("**/api/v1/intents", corruptRequest);
      page.off("request", observeRequest); page.off("console", observeConsole);
    }
  }
  return rejectedIDs;
}

async function witnessFiscalRefusalJourney() {
  // Independent UI-created account: diagnostic reads must not consume the
  // main gameplay population's account bucket or change its command sequence.
  const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });
  const errors = [];
  page.on("pageerror", (error) => errors.push(error));
  try {
    await page.goto(uiURL, { waitUntil: "networkidle" });
    await page.getByRole("button", { name: "BEGIN ATTEMPT", exact: true }).click();
    await page.locator('main[data-surface="desk"]').waitFor({ state: "visible", timeout: 30_000 });
    const accessToken = await page.evaluate(() => JSON.parse(localStorage.getItem("cloud-clicker.credentials.v1")).accessToken);
    await page.getByRole("button", { name: "Earnings Calls", exact: true }).click();
    await page.locator('main[data-surface="fiscal"]').waitFor({ state: "visible", timeout: 30_000 });
    await new Promise((resolve) => setTimeout(resolve, 400));
    await clickAppliedIntent(page, "Hold the earnings call", "Fiscal refusal journey initial harvest");
    const rejectedIDs = await witnessFiscalServerRefusals(page, accessToken);
    const before = await founderState(accessToken);
    const consent = await clickAppliedIntent(page, "Hold the earnings call", "Fiscal fresh consent after server refusals");
    const after = await founderState(accessToken);
    if (consent.intent?.kind !== "harvest_fiscal_period" || rejectedIDs.includes(consent.intent.intent_id) ||
        consent.intent.expected_revision !== before.founder_revision || consent.body.founder_revision !== before.founder_revision + 1 ||
        after.founder_revision !== consent.body.founder_revision || after.features.fiscal.credit !== consent.body.fiscal_credit_after ||
        after.features.fiscal.period.seq !== consent.body.seq_after ||
        after.features.fiscal.period.opened_wall_ms !== consent.body.period_opened_wall_ms || errors.length !== 0) {
      throw new Error(`fresh Fiscal consent did not apply at the current Founder revision: ${JSON.stringify(consent)}; errors=${errors.map(String)}`);
    }
    console.log(`composed Fiscal fresh consent: new intent UUID, current Founder revision, applied receipt and persisted revision${after.founder_revision}: PASS`);
  } finally { await page.close(); }
}

async function unlockPitchThroughFiscalUI(page) {
  await page.getByRole("button", { name: "Earnings Calls", exact: true }).click();
  await page.locator('main[data-surface="fiscal"]').waitFor({ state: "visible", timeout: 30_000 });
  // The composed Fiscal clock guarantees a harvest 200 ms after the period
  // opens. The player uses the rendered controls for both prerequisite intents.
  await new Promise((resolve) => setTimeout(resolve, 400));
  const harvest = await clickAppliedIntent(page, "Hold the earnings call", "Fiscal harvest for Pitch");
  if (harvest.intent?.kind !== "harvest_fiscal_period") throw new Error(`Fiscal harvest control emitted ${JSON.stringify(harvest.intent)}`);
  const intentRequest = page.waitForRequest((request) => new URL(request.url()).pathname === "/api/v1/intents" &&
    request.postDataJSON()?.kind === "spend_fiscal_credit", { timeout: 30_000 });
  let unlockClicked = false;
  for (let offers = 0; offers < 3 && !unlockClicked; offers += 1) {
    const action = await clickReadyPitchControl(page, "unlock");
    if (action === "offer") await declinePitchOffer(page, "Earnings Calls", "fiscal");
    else unlockClicked = true;
  }
  if (!unlockClicked) throw new Error("Pitch Fiscal unlock was repeatedly preempted by Exit offers");
  let request;
  try { request = await intentRequest; }
  catch (error) {
    const state = await page.evaluate(() => ({
      surface: document.querySelector("main")?.getAttribute("data-surface"),
      busy: document.querySelector("main")?.getAttribute("aria-busy"),
      unlock: [...document.querySelectorAll('section[aria-labelledby="fiscal-unlocks-heading"] li')]
        .filter((item) => item.querySelector("h3")?.textContent?.trim() === "The Pitch")
        .map((item) => ({ text: item.textContent?.trim(), disabled: item.querySelector("button")?.disabled })),
      alerts: [...document.querySelectorAll('[role="alert"]')].map((item) => item.textContent?.trim()),
      trace: globalThis.__composedPitchActionTrace,
    }));
    throw new Error(`Pitch Fiscal unlock control emitted no intent request; state=${JSON.stringify(state)}`, { cause: error });
  }
  const intent = request.postDataJSON();
  if (intent?.kind !== "spend_fiscal_credit" || intent.target?.kind !== "unlock" || intent.target?.unlock_id !== "minigame.pitch") {
    throw new Error(`Pitch-specific Fiscal unlock control emitted ${JSON.stringify(intent)}`);
  }
  const response = await request.response();
  if (!response) throw new Error("Pitch Fiscal unlock emitted no response");
  const body = await response.json();
  if (response.status() !== 200 || body?.outcome !== "applied") throw new Error(`Pitch Fiscal unlock was not applied (${response.status()}): ${JSON.stringify(body)}`);
  await page.waitForFunction(() => document.querySelector("main")?.getAttribute("aria-busy") === "false", undefined, { timeout: 30_000 });
  await page.getByRole("button", { name: "The Pitch", exact: true }).click();
  await page.locator('main[data-surface="minigame_session"]').waitFor({ state: "visible", timeout: 30_000 });
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
  const lockedCreate = page.waitForResponse((response) => new URL(response.url()).pathname === "/api/v1/minigames/pitch/sessions", { timeout: 30_000 }).catch((error) => error);
  try {
    let clicked = false;
    for (let offers = 0; offers < 3 && !clicked; offers += 1) {
      const action = await clickReadyPitchControl(page, "start");
      if (action === "offer") await declinePitchOffer(page, "The Pitch", "minigame_session");
      else clicked = true;
    }
    if (!clicked) throw new Error("locked Pitch Start was repeatedly preempted by Exit offers");
  }
  catch (error) {
    throw new Error(`locked Pitch control could not be clicked; surface=${await page.locator("main").getAttribute("data-surface")} page_text=${JSON.stringify((await page.locator("main").innerText()).slice(0, 800))} trace=${JSON.stringify(await page.evaluate(() => globalThis.__composedPitchActionTrace))}`, { cause: error });
  }
  const locked = await lockedCreate;
  if (locked instanceof Error) {
    throw new Error(`locked Pitch create emitted no response; surface=${await page.locator("main").getAttribute("data-surface")} minigame_responses=${JSON.stringify(minigameResponses)} page_text=${JSON.stringify((await page.locator("main").innerText()).slice(0, 800))} trace=${JSON.stringify(await page.evaluate(() => globalThis.__composedPitchActionTrace))}`, { cause: locked });
  }
  const lockedBody = await locked.json();
  if (locked.status() !== 409 || lockedBody.category !== "not_eligible" || lockedBody.detail !== "fiscal_unlock_required") {
    throw new Error(`locked Pitch create was not the typed fiscal rejection (${locked.status()}): ${JSON.stringify(lockedBody)}`);
  }
  await page.getByText("Locked. Unlock it with Fiscal credit first.", { exact: true }).waitFor({ state: "visible", timeout: 30_000 });

  await unlockPitchThroughFiscalUI(page);
  const beforeSession = await founderState(accessToken);
  if (beforeSession?.transitions?.wind_down?.eligible !== true) {
    throw new Error(`Tier-1 Wind Down was not eligible before Pitch: ${JSON.stringify(beforeSession?.transitions)}`);
  }
  const create = page.waitForResponse((response) => new URL(response.url()).pathname === "/api/v1/minigames/pitch/sessions", { timeout: 30_000 });
  let startClicked = false;
  for (let offers = 0; offers < 3 && !startClicked; offers += 1) {
    const action = await clickReadyPitchControl(page, "start");
    if (action === "offer") await declinePitchOffer(page, "The Pitch", "minigame_session");
    else startClicked = true;
  }
  if (!startClicked) throw new Error("unlocked Pitch Start was repeatedly preempted by Exit offers");
  const created = await create;
  if (created.status() !== 200) throw new Error(`unlocked Pitch create failed (${created.status()}): ${JSON.stringify(await created.json())}`);
  const duringSession = await founderState(accessToken);
  if (duringSession?.transitions?.wind_down?.eligible !== false) {
    throw new Error(`active Pitch session offered Wind Down: ${JSON.stringify(duringSession?.transitions)}`);
  }

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
  const afterSession = await founderState(accessToken);
  if (afterSession?.transitions?.wind_down?.eligible !== true) {
    throw new Error(`Tier-1 Wind Down did not return after Pitch: ${JSON.stringify(afterSession?.transitions)}`);
  }
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
  await witnessFiscalRefusalJourney();
  const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });
  await page.addInitScript(() => {
    const BrowserWebSocket = globalThis.WebSocket;
    globalThis.__cloudClickerTestSockets = [];
    globalThis.__composedPitchActionTrace = [];
    for (const type of ["pointerdown", "pointerup", "click"]) document.addEventListener(type, (event) => {
      const button = event.target instanceof Element ? event.target.closest("button") : null;
      const label = button?.textContent?.trim();
      if (label !== "Start a pitch" && !label?.startsWith("Unlock for ")) return;
      globalThis.__composedPitchActionTrace.push({ type, label, disabled: button.disabled,
        busy: document.querySelector("main")?.getAttribute("aria-busy"), time: performance.now() });
      if (globalThis.__composedPitchActionTrace.length > 32) globalThis.__composedPitchActionTrace.shift();
    }, true);
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
  // Clout v1 CV9: the pinned epoch declares no axis_stack, so the composed
  // server withholds the arm and reports the feature fact false (the arm is
  // fixture-first until a content mint pins the v5 economy).
  if (features.axis_stack !== undefined && features.axis_stack !== null ||
      !liveSnapshot.body.facts.some((fact) => fact.fact_id === "feature.axis_stack" && fact.value === false)) {
    throw new Error(`composed live Game UI projected an axis stack the pinned epoch does not declare: ${JSON.stringify(features.axis_stack)}`);
  }
  // GS5: the pinned epoch has the opportunities artifact, so the live arm is
  // projected beside the null-only active_play and the feature fact follows it.
  if (features.opportunity === null || typeof features.opportunity !== "object" || !Array.isArray(features.opportunity.buffs) ||
      !liveSnapshot.body.facts.some((fact) => fact.fact_id === "feature.active_play" && fact.value === true)) {
    throw new Error(`composed live Game UI did not project the GS5 opportunity arm: ${JSON.stringify(features.opportunity)}`);
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

  await witnessFirstPurchaseAchievement(page, websocketReceivedFrames, parsedCredentials.accessToken);

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
  // The Fiscal prerequisite and Pitch play both originate in rendered player
  // controls; the composed epoch pins minigame.pitch at 3 credit.
  const opportunity = await witnessOpportunityClaim(page);
  console.log(`composed GS5 opportunity: ${opportunity.manual_clicks} manual clicks, claimed ${opportunity.effect_row_id} (${opportunity.expired_claims} expired attempts), ${opportunity.proof_branch} effect bound to next snapshot revision ${opportunity.revision}: PASS`);
  seedGateRequirement(liveSnapshot.body.run.founder_id);
  await page.reload({ waitUntil: "networkidle" });
  const pitchTierGate = await waitForEnabledButton(page, "Move Into the Garage");
  await clickAppliedIntent(page, pitchTierGate, "Pitch-run cross-gate");
  const pitch = await playPitchThroughUI(page, parsedCredentials.accessToken);
  if (pageErrors.length > 0) throw new AggregateError(pageErrors, "composed browser path emitted page errors");
  console.log(`composed Fiscal UI harvest + Pitch unlock: ${pitch.commands} Pitch UI commands, ${pitchOffersDeclined} visible offer declines, terminal receipt credited ${pitch.credited} at company revision ${pitch.companyRevision}, snapshot refreshed to ${pitch.refreshedRevision}: PASS`);
  console.log("composed Game UI v4 features + transitions + both terminal states + next-run continuation + WebSocket recovery: PASS");
  await page.goto("about:blank");
  await new Promise((resolve) => setTimeout(resolve, 100));
} finally {
  await browser?.close();
  await vite?.close();
  await stopGameserver();
}
