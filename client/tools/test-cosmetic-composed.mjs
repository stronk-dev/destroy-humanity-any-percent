import { spawn, spawnSync } from "node:child_process";
import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { copyFileSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { createServer as createHTTPServer, request as httpRequest } from "node:http";
import { createServer as createTCPServer, connect as connectTCP } from "node:net";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { chromium } from "playwright";
import { build } from "vite";
import { activateCosmeticBuy, assertNativeCosmeticBuyTrace } from "./activate-cosmetic-buy.mjs";
import { assertCareSnapshot } from "./care-snapshot-proof.mjs";
import { cosmeticN5Fixture, n5CleanupMarker, n5Faults, n5RejectionPrefix } from "./cosmetic-n5-fixture.mjs";

const args = process.argv.slice(2);
const n5Fault = args[0]?.startsWith("--n5-fault=") ? args[0].slice("--n5-fault=".length) : null;
if (args.length !== 0 && (args.length !== 1 || (args[0] !== "--axis-stack" && !n5Faults.includes(n5Fault)))) {
  throw new Error("supported composed fixture options: --axis-stack, --n5-fault=checkout, --n5-fault=payment-request");
}
const axisFixture = args[0] === "--axis-stack";
const buyKey = axisFixture ? " " : "Enter";
const clientRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const repositoryRoot = path.resolve(clientRoot, "..");
const gameserverURL = "http://127.0.0.1:18082";
const uiURL = "http://localhost:5173";
const databaseURL = "postgres://cloud_clicker:cloud_clicker_game_ui_test@127.0.0.1:55433/cloud_clicker_game_ui_test?sslmode=disable";
const fixtureRoot = mkdtempSync(path.join(os.tmpdir(), "cloud-clicker-cosmetic-ac14-"));
const clientDist = path.join(fixtureRoot, "client-dist");
const startedAt = Date.now();
const processErrors = [];
let gameserver;
let staticServer;
let browser;
const proxySockets = new Set();

function writeFixture(relative, data) {
  const destination = path.join(fixtureRoot, relative);
  mkdirSync(path.dirname(destination), { recursive: true });
  writeFileSync(destination, data);
}

function constantsHash(artifacts) {
  const hash = createHash("sha256");
  for (const name of [...artifacts.keys()].sort()) {
    const data = artifacts.get(name);
    const nameBytes = Buffer.from(name);
    const frame = Buffer.alloc(8);
    frame.writeBigUInt64BE(BigInt(nameBytes.length));
    hash.update(frame).update(nameBytes);
    frame.writeBigUInt64BE(BigInt(data.length));
    hash.update(frame).update(data);
  }
  return `sha256:${hash.digest("hex")}`;
}

function buildFixtureRoot() {
  const seed = JSON.parse(readFileSync(path.join(repositoryRoot, "balance/epochs/phase0.json"), "utf8"));
  const artifacts = new Map(seed.artifacts.map(({ name, path: relative }) => [name, readFileSync(path.join(repositoryRoot, relative))]));
  // Exact test-only chain used by replaycatalog's Reputation/Pet/Cosmetics fixtures.
  // AC11 variant uses the existing declared PR fixture, not adopted epoch data.
  const economy = JSON.parse(axisFixture
    ? readFileSync(path.join(repositoryRoot, "balance/testdata/axis-stack/economy-v5-fixture.json"), "utf8")
    : artifacts.get("economy").toString("utf8"));
  economy.multiplier_sources.push({ id: "reputation.founder_bonus", slot: "prestige", target: "all", provider: "reputation_tree" });
  artifacts.set("economy", Buffer.from(JSON.stringify(economy)));
  // RP-318: make the already-real adoption/feed journey cross a status band.
  // This disposable catalog is not minted content: hunger starts low, the
  // other stats high, so one ordinary feed must produce low → high.
  const pets = JSON.parse(artifacts.get("pets").toString("utf8"));
  for (const stat of pets.stat_policy.stats) stat.initial_ppm = stat.stat_id === "hunger" ? 499_999 : 800_000;
  artifacts.set("pets", Buffer.from(JSON.stringify(pets)));
  for (const [name, relative] of [
    ["reputation_tree", "balance/testdata/reputation-tree/fixture-v1.json"],
    ["pet_species", "balance/testdata/pet-species/fixture-v1.json"],
    ["cosmetics", "balance/testdata/cosmetics/fixture-v1.json"],
  ]) artifacts.set(name, readFileSync(path.join(repositoryRoot, relative)));
  const hash = constantsHash(artifacts);
  const names = [...artifacts.keys()].sort();
  for (const name of names) writeFixture(`balance/cosmetic-ac14/${name}.json`, artifacts.get(name));
  writeFixture("balance/epochs/phase0.json", `${JSON.stringify({
    schema_version: 1, current_epoch_id: 1,
    artifacts: names.map((name) => ({ name, path: `balance/cosmetic-ac14/${name}.json` })),
    epochs: [{ epoch_id: 1, name: "Cosmetic AC14 fixture", changelog_ref: "changelog/epoch-1.md", accepted_hashes: [hash] }],
  }, null, 2)}\n`);
  writeFixture("changelog/epoch-1.md", "# Cosmetic AC14 fixture\n");
  for (const relative of ["moderation/guild-names.txt", "balance/transport/phase0.json", "balance/api/phase0.json"]) {
    const destination = path.join(fixtureRoot, relative);
    mkdirSync(path.dirname(destination), { recursive: true });
    copyFileSync(path.join(repositoryRoot, relative), destination);
  }
  return hash;
}

function testDatabaseSQL(sql, tuplesOnly = false) {
  const result = spawnSync("docker", ["compose", "-f", "compose.game-ui-test.yml", "exec", "-T", "game-ui-postgres",
    "psql", "-v", "ON_ERROR_STOP=1", "-U", "cloud_clicker", "-d", "cloud_clicker_game_ui_test", ...(tuplesOnly ? ["-At"] : []), "-c", sql],
  { cwd: repositoryRoot, encoding: "utf8" });
  if (result.status !== 0) throw new Error(`cosmetic test DB command failed: ${result.stderr || result.stdout}`);
  return result.stdout;
}

function resetTestDatabase() {
  // This is only the named Compose test database. It runs after the existing
  // composed UI driver, before starting this driver, never on an operator DB.
  testDatabaseSQL("DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;");
}

function seedGateRequirement(founderID) {
  if (!/^[0-9a-f-]{36}$/u.test(founderID)) throw new Error("invalid Founder ID in cosmetic test setup");
  const result = testDatabaseSQL(`WITH current AS (
    SELECT revision.stream_id, revision.revision, revision.version, revision.state, revision.constants_hash
    FROM save_revisions revision JOIN save_streams stream ON stream.id=revision.stream_id
    WHERE stream.owner_kind='founder' AND stream.owner_id='${founderID}' AND stream.scope='company' AND stream.archived_at IS NULL
    ORDER BY revision.revision DESC LIMIT 1
  ) INSERT INTO save_revisions(stream_id,revision,version,state,constants_hash)
    SELECT stream_id,revision+1,version,jsonb_set(state,'{balances,company.cash}',to_jsonb('${axisFixture ? "1e8" : "1e5"}'::text),false),constants_hash FROM current;`);
  if (!result.includes("INSERT 0 1")) throw new Error(`cosmetic gate setup inserted no revision: ${result}`);
}

async function waitForPort(port) {
  await new Promise((resolve, reject) => {
    const probe = createTCPServer();
    probe.once("error", (error) => reject(new Error(`cosmetic test port ${port} unavailable: ${error.message}`)));
    probe.listen(port, "127.0.0.1", () => probe.close(resolve));
  });
}

async function waitForReady() {
  const deadline = Date.now() + 60_000;
  while (Date.now() < deadline) {
    if (processErrors.length || gameserver.exitCode !== null) throw processErrors[0] ?? new Error(`cosmetic gameserver exited ${gameserver.exitCode}`);
    try {
      const response = await fetch(`${gameserverURL}/readyz`);
      if (response.ok) return;
    } catch { /* still starting */ }
    await new Promise((resolve) => setTimeout(resolve, 100));
  }
  throw new Error("cosmetic gameserver did not become ready");
}

async function stopGameserver() {
  if (!gameserver || gameserver.exitCode !== null || gameserver.signalCode !== null) return;
  gameserver.kill("SIGTERM");
  const deadline = Date.now() + 10_000;
  while (gameserver.exitCode === null && gameserver.signalCode === null && Date.now() < deadline) {
    await new Promise((resolve) => setTimeout(resolve, 50));
  }
  if (gameserver.exitCode === null && gameserver.signalCode === null) gameserver.kill("SIGKILL");
}

function assertNetwork(url, kind) {
  const parsed = new URL(url, uiURL);
  const origin = new URL(uiURL);
  const sameOrigin = parsed.host === origin.host && ["http:", "ws:"].includes(parsed.protocol);
  const pathAllowed = kind === "websocket"
    ? parsed.pathname === "/connection/websocket"
    : parsed.pathname === "/" || /^\/(?:assets\/|api\/)/u.test(parsed.pathname) || parsed.pathname === "/favicon.ico";
  if (!sameOrigin || !pathAllowed) throw new Error(`cosmetic N5 ${kind} escaped allowlist: ${url}`);
}

async function serveBuiltClient() {
  const dist = clientDist;
  const server = createHTTPServer((request, response) => {
    const pathname = new URL(request.url, uiURL).pathname;
    if (pathname.startsWith("/api/")) {
      const upstream = httpRequest(new URL(request.url, gameserverURL), {
        method: request.method, headers: { ...request.headers, host: new URL(gameserverURL).host },
      }, (received) => { response.writeHead(received.statusCode ?? 502, received.headers); received.pipe(response); });
      upstream.on("error", (error) => { response.writeHead(502); response.end(`API proxy failed: ${error.message}`); });
      request.pipe(upstream);
      return;
    }
    if (pathname !== "/" && pathname !== "/favicon.ico" && !pathname.startsWith("/assets/")) {
      response.writeHead(404); response.end(); return;
    }
    const relative = pathname === "/" ? "index.html" : pathname.slice(1);
    const target = path.resolve(dist, relative);
    if (!target.startsWith(`${dist}${path.sep}`)) { response.writeHead(403); response.end(); return; }
    try {
      const data = readFileSync(target);
      const type = target.endsWith(".js") ? "text/javascript" : target.endsWith(".css") ? "text/css" : target.endsWith(".svg") ? "image/svg+xml" : target.endsWith(".html") ? "text/html" : "application/octet-stream";
      response.writeHead(200, { "content-type": type }); response.end(data);
    } catch { response.writeHead(404); response.end(); }
  });
  server.on("upgrade", (request, socket, head) => {
    if (new URL(request.url, uiURL).pathname !== "/connection/websocket") { socket.destroy(); return; }
    const upstream = connectTCP(18082, "127.0.0.1");
    proxySockets.add(socket); proxySockets.add(upstream);
    const close = () => { socket.destroy(); upstream.destroy(); proxySockets.delete(socket); proxySockets.delete(upstream); };
    socket.on("error", close); upstream.on("error", close);
    socket.on("close", close); upstream.on("close", close);
    upstream.on("connect", () => {
      const headers = [];
      for (let index = 0; index < request.rawHeaders.length; index += 2) {
        const name = request.rawHeaders[index];
        headers.push(`${name}: ${name.toLowerCase() === "host" ? "127.0.0.1:18082" : request.rawHeaders[index + 1]}`);
      }
      upstream.write(`${request.method} ${request.url} HTTP/1.1\r\n${headers.join("\r\n")}\r\n\r\n`);
      if (head.length) upstream.write(head);
      socket.pipe(upstream).pipe(socket);
    });
  });
  await new Promise((resolve, reject) => { server.once("error", reject); server.listen(5173, "127.0.0.1", resolve); });
  return server;
}

async function snapshot(page) {
  return page.evaluate(async () => {
    const raw = localStorage.getItem("cloud-clicker.credentials.v1");
    if (!raw) throw new Error("missing cosmetic test credentials");
    const credentials = JSON.parse(raw);
    const response = await fetch("/api/v1/founder/state", { headers: { Authorization: `Bearer ${credentials.accessToken}` } });
    if (response.status !== 200) throw new Error(`cosmetic snapshot status ${response.status}`);
    return response.json();
  });
}

function plainFixtureCopy(key) {
  const artifact = JSON.parse(readFileSync(path.join(clientRoot, "src/copy/generated/catalog.json"), "utf8"));
  const row = artifact.entries.find((entry) => entry.key === key);
  if (!row || row.params.length !== 0 || row.era_variants !== null) {
    throw new Error(`cosmetic G10 fixture requires an unparameterized, era-independent copy key: ${key}`);
  }
  return row.text;
}

async function companyDOMIntent(page, requests, control, kind, fields) {
  const before = await snapshot(page);
  const matching = () => requests.filter((request) => request.method() === "POST" &&
    new URL(request.url()).pathname === "/api/v1/intents" && request.postDataJSON()?.kind === kind);
  const priorCount = matching().length;
  // Native keyboard activation of the actual mounted control; no direct API
  // writes, runtime calls, forced disabled activation or intent retry.
  await control.click({ trial: true, timeout: 30_000 });
  await control.focus();
  const responseTask = page.waitForResponse((response) => response.request().method() === "POST" &&
    new URL(response.url()).pathname === "/api/v1/intents" && response.request().postDataJSON()?.kind === kind,
  { timeout: 30_000 }).then((value) => ({ value }), (error) => ({ error }));
  await control.press("Enter");
  const result = await responseTask;
  if (result.error) throw result.error;
  const response = result.value;
  const receipt = await response.json();
  const emitted = matching();
  assert.equal(emitted.length, priorCount + 1, `Clout ${kind}: exactly one DOM intent`);
  const body = emitted.at(-1).postDataJSON();
  assert.deepEqual(body, { intent_id: body.intent_id, expected_revision: before.revision, kind, ...fields }, `Clout ${kind}: exact request`);
  assert.equal(response.status(), 200, `Clout ${kind}: HTTP result`);
  assert.equal(receipt.intent_id, body.intent_id, `Clout ${kind}: bound receipt`);
  assert.equal(receipt.outcome, "applied", `Clout ${kind}: authoritative application`);
  assert.equal(receipt.new_revision, before.revision + 1, `Clout ${kind}: Company revision`);
  await page.waitForFunction(() => document.querySelector("main.game-ui")?.getAttribute("aria-busy") === "false",
    undefined, { timeout: 30_000 });
  return receipt;
}

async function assertAxis(page, { input, product, factors, owned = false }) {
  const view = await snapshot(page);
  const arm = view.features?.axis_stack;
  assert.ok(arm, "Clout AC11: real snapshot must carry axis_stack");
  assert.equal(view.facts.find((row) => row.fact_id === "feature.axis_stack")?.value, true);
  assert.equal(arm.input_kind, "achievement_attainment_run");
  assert.equal(arm.input_value, input);
  assert.equal(arm.input_cap, 44);
  assert.equal(arm.saturated, false);
  assert.equal(arm.product, product);
  assert.deepEqual(arm.interns.map((row) => [row.upgrade_id, row.minimum, row.factor_ppm, row.owned, row.factor]), [
    ["upgrade.pr_intern_1", 6, 25_000, owned, factors[0]],
    ["upgrade.pr_intern_2", 10, 20_000, false, factors[1]],
  ]);
  assert.deepEqual(arm.contributions, owned
    ? [{ source_id: "upgrade.pr_intern_1.axis", upgrade_id: "upgrade.pr_intern_1", factor: factors[0] }] : []);
  const panel = page.locator("section.axis");
  await panel.waitFor({ state: "visible", timeout: 30_000 });
  // Fixture factors here are all e0; literal text equality never calls the
  // subject's formatter or computes a factor from the raw achievement count.
  assert.deepEqual(await panel.locator("output").allTextContents(), [product, ...factors].map((value) => value.replace(/e0$/u, "")));
  assert.ok((await panel.innerText()).includes(`Achievement attainment this run: ${input} of 44`));
  assert.ok((await panel.innerText()).includes(plainFixtureCopy("axis_stack.formula.caption")));
  assert.deepEqual(await panel.locator("progress").evaluateAll((bars) => bars.map((bar) => ({ value: bar.value, max: bar.max }))),
    owned ? [{ value: Math.min(input, 10), max: 10 }] : [{ value: Math.min(input, 6), max: 6 }, { value: Math.min(input, 10), max: 10 }]);
  return view;
}

async function witnessAxis(page, requests) {
  await assertAxis(page, { input: 2, product: "1e0", factors: ["1.05e0", "1.04e0"] });
  const help = page.locator("section.axis details");
  // Native summary does not expose a button role in every browser. Keep the
  // native semantics and test its label and actual keyboard behavior instead.
  const opener = help.locator("summary");
  assert.equal(await opener.count(), 1, "built PR Codex must have one native summary");
  assert.equal(await opener.getAttribute("aria-label"), plainFixtureCopy("axis_stack.help.label"));
  assert.equal(await opener.getAttribute("role"), null);
  const intentsBeforeHelp = requests.filter((request) => request.method() === "POST" && new URL(request.url()).pathname === "/api/v1/intents").length;
  assert.equal(await help.evaluate((node) => node.open), false);
  await opener.press("Enter");
  assert.equal(await help.evaluate((node) => node.open), true);
  assert.equal(await help.locator(".codex").isVisible(), true);
  assert.equal(await help.locator(".codex").innerText(), plainFixtureCopy("codex.axis_stack"));
  await opener.press("Space");
  assert.equal(await help.evaluate((node) => node.open), false);
  assert.equal(requests.filter((request) => request.method() === "POST" && new URL(request.url()).pathname === "/api/v1/intents").length, intentsBeforeHelp,
    "opening PR help must not submit gameplay");
  console.info("Clout CV9: built-client native PR Codex open/close passed (owner copy pending)");
  const pr = page.locator('section[aria-labelledby="upgrades-heading"] article').filter({ has: page.getByRole("heading", { name: plainFixtureCopy("upgrade.pr_intern_1.title"), exact: true }) });
  assert.equal(await pr.getByRole("button").isDisabled(), true, "PR purchase below required attainment is disabled");
  const generators = page.locator('section[aria-labelledby="generators-heading"] article').filter({ has: page.getByRole("heading", { name: plainFixtureCopy("generator.beige_tower.title"), exact: true }) });
  const purchase = await companyDOMIntent(page, requests, generators.getByRole("button", { name: plainFixtureCopy("desk.buy_max"), exact: true }),
    "buy_generator", { generator_id: "generator.beige_tower", count: { mode: "max" } });
  const ready = await assertAxis(page, { input: 12, product: "1e0", factors: ["1.3e0", "1.24e0"] });
  assert.ok(ready.revision >= purchase.new_revision);
  assert.deepEqual(ready.features.axis_stack.attained.map((row) => row.achievement_id), [
    "achievement.first_gate", "achievement.generators_owned_100", "achievement.generators_purchased_1", "achievement.generators_purchased_25",
  ]);
  const receipt = await companyDOMIntent(page, requests, pr.getByRole("button", { name: plainFixtureCopy("desk.buy_one"), exact: true }),
    "buy_upgrade", { upgrade_id: "upgrade.pr_intern_1" });
  const after = await assertAxis(page, { input: 12, product: "1.3e0", factors: ["1.3e0", "1.24e0"], owned: true });
  assert.ok(after.revision >= receipt.new_revision);
  assert.equal(after.upgrades.find((row) => row.upgrade_id === "upgrade.pr_intern_1")?.owned, true);
  assert.equal(await pr.getByRole("button").isDisabled(), true, "owned PR cannot be bought again");
  await page.reload({ waitUntil: "networkidle" });
  await page.locator('main[data-surface="desk"]').waitFor({ state: "visible", timeout: 30_000 });
  const restored = await assertAxis(page, { input: 12, product: "1.3e0", factors: ["1.3e0", "1.24e0"], owned: true });
  assert.equal(restored.run.founder_id, after.run.founder_id);
  assert.equal(restored.run.run_seq, after.run.run_seq);
  assert.equal(restored.upgrades.find((row) => row.upgrade_id === "upgrade.pr_intern_1")?.owned, true);
  // Inspect the actual stored head/event, not just another browser read from
  // the still-running service. Setup seeded cash only, never these fields.
  const founderID = restored.run.founder_id;
  assert.match(founderID, /^[0-9a-f-]{36}$/u);
  const persisted = JSON.parse(testDatabaseSQL(`SELECT jsonb_build_object(
    'version', revision.version, 'revision', revision.revision,
    'attainment', revision.state->'attainment_score_run',
    'attained', revision.state->'achievements_attained_run',
    'owned', revision.state->'upgrades_owned' ? 'upgrade.pr_intern_1',
    'purchases', (SELECT count(*) FROM events WHERE stream_id=revision.stream_id
      AND kind='upgrade_purchased' AND payload->>'upgrade_id'='upgrade.pr_intern_1'))
    FROM save_revisions revision JOIN save_streams stream ON stream.id=revision.stream_id
    WHERE stream.owner_kind='founder' AND stream.owner_id='${founderID}' AND stream.scope='company' AND stream.archived_at IS NULL
    ORDER BY revision.revision DESC LIMIT 1;`, true).trim());
  assert.deepEqual(persisted, { version: 19, revision: receipt.new_revision, attainment: 12,
    attained: ready.features.axis_stack.attained.map((row) => row.achievement_id), owned: true, purchases: 1 });
  console.log("composed Clout AC11 fixture: DOM first gate → attainment 2 → native generator purchase → attainment 12 → native PR purchase → bound receipt → product 1.3 → owned reload + exact SQL head/event: PASS");
}

async function founderDOMIntent(page, requests, control, kind, expectedFields) {
  const before = await snapshot(page);
  const matching = () => requests.filter((request) => request.method() === "POST" &&
    new URL(request.url()).pathname === "/api/v1/intents" && request.postDataJSON()?.kind === kind);
  const priorCount = matching().length;
  await page.evaluate(() => { globalThis.__cosmeticActionTrace = []; });
  const responses = [];
  const observeResponse = (response) => {
    const request = response.request();
    if (request.method() === "POST" && new URL(request.url()).pathname === "/api/v1/intents" &&
        request.postDataJSON()?.kind === kind) responses.push({ status: response.status() });
  };
  page.on("response", observeResponse);
  let phase = "actionability";
  try {
    await control.click({ trial: true, timeout: 30_000 });
    const deadline = Date.now() + 30_000;
    const remaining = () => {
      const budget = deadline - Date.now();
      if (budget <= 0) throw new Error(`cosmetic G10 ${kind} exceeded the action deadline`);
      return budget;
    };
    const responseTask = page.waitForResponse((response) => response.request().method() === "POST" &&
      new URL(response.url()).pathname === "/api/v1/intents" && response.request().postDataJSON()?.kind === kind, { timeout: remaining() })
      .then((value) => ({ value }), (error) => ({ error }));
    if (kind === "equip_cosmetic" || kind === "unequip_cosmetic") {
      phase = "guarded-dom-activation";
      // G10 proves the real DOM consumer/persisted result, not physical pointer
      // timing. Wait for ready state and activate once in that
      // same browser task; a completed pointer sequence need not emit a click.
      await page.evaluate(activateCosmeticBuy, { label: await control.innerText(), state: "owned", budgetMs: remaining() });
    } else if (kind === "adopt_pet") {
      phase = "native-keyboard-activation";
      await control.focus();
      await page.keyboard.press(axisFixture ? "Space" : "Enter");
    } else {
      phase = "pointer-activation";
      await control.click({ timeout: remaining() });
    }
    phase = "response";
    const result = await responseTask;
    if (result.error) throw result.error;
    const response = result.value;
    const receipt = await response.json();
    const emitted = matching();
    const body = emitted.at(-1)?.postDataJSON();
    const requiredKeys = ["intent_id", "kind", "expected_revision", ...Object.keys(expectedFields)].sort();
    if (emitted.length !== priorCount + 1 || !body || Object.keys(body).sort().join("\0") !== requiredKeys.join("\0") ||
        body.expected_revision !== before.founder_revision || Object.entries(expectedFields).some(([key, value]) => body[key] !== value) ||
        response.status() !== 200 || receipt.outcome !== "applied" || receipt.intent_id !== body.intent_id ||
        receipt.founder_revision !== before.founder_revision + 1) {
      throw new Error(`cosmetic G10 ${kind} did not emit one exact Founder-scoped applied intent: ${JSON.stringify({ body, receipt, emitted: emitted.length - priorCount })}`);
    }
    // GS0.2: an HTTP receipt is not the mounted host's completed authoritative
    // refresh. Stay within this action's deadline; do not retry the intent or
    // poll for an overlay that a settled but broken consumer never rendered.
    await page.waitForFunction(() => document.querySelector("main.game-ui")?.getAttribute("aria-busy") === "false",
      undefined, { timeout: remaining() });
    return receipt;
  } catch (error) {
    const dom = await page.evaluate(() => ({ events: globalThis.__cosmeticActionTrace,
      dropped_events: globalThis.__cosmeticActionTrace?.dropped ?? 0,
      surface: document.querySelector("main")?.getAttribute("data-surface"),
      busy: document.querySelector("main")?.getAttribute("aria-busy") }));
    const emitted = matching().slice(priorCount);
    throw new Error(`cosmetic G10 ${kind} boundary failed: ${JSON.stringify({ phase, dom,
      emitted: emitted.length, responses,
      requestFailures: emitted.map((request) => request.failure()?.errorText ?? null) })}`, { cause: error });
  } finally {
    page.off("response", observeResponse);
    await page.evaluate(() => { globalThis.__cosmeticActionTrace = null; });
  }
}

function assertPersistedWearer(view, petID, worn) {
  const pets = view.features?.pet_adoption?.pets;
  const arm = view.features?.cosmetics;
  const item = arm?.items?.find((row) => row.cosmetic_id === "horse_armor");
  const expectedWearers = worn ? [petID] : [];
  if (pets?.length !== 1 || pets[0].pet_id !== petID || arm?.active !== true || item?.owned !== true ||
      JSON.stringify(item.worn_by) !== JSON.stringify(expectedWearers) || arm.wearers?.length !== 1 ||
      arm.wearers[0].pet_id !== petID || arm.wearers[0].worn !== worn) {
    throw new Error(`cosmetic G10 persisted pet/wearer projection diverged: ${JSON.stringify({ pets, arm, petID, worn })}`);
  }
}

function assertPersistedCare(view, receipt, hash) {
  const founderID = view.run.founder_id;
  for (const id of [founderID, receipt.pet_id, receipt.intent_id]) assert.match(id, /^[0-9a-f-]{36}$/u);
  // Read only the actual head and this command's log from the disposable test
  // DB. Do not seed care state or project from a receipt-only fixture.
  const persisted = JSON.parse(testDatabaseSQL(`SELECT jsonb_build_object(
    'revision', revision.revision, 'constants_hash', revision.constants_hash,
    'age_ms', revision.state->'age_ms', 'care', revision.state->'pets'->'${receipt.pet_id}',
    'receipt', log.receipt,
    'command_attended_ms', log.replay_inputs->'resolved'->'attendance'->'effective_founder_attended_ms')
    FROM save_revisions revision JOIN save_streams stream ON stream.id=revision.stream_id
    JOIN founder_log log ON log.founder_stream_id=stream.id AND log.intent_id='${receipt.intent_id}'
    WHERE stream.owner_kind='founder' AND stream.owner_id='${founderID}' AND stream.scope='founder' AND stream.archived_at IS NULL
    ORDER BY revision.revision DESC LIMIT 1;`, true).trim());
  const catalog = JSON.parse(readFileSync(path.join(fixtureRoot, "balance/cosmetic-ac14/pets.json"), "utf8"));
  return assertCareSnapshot(view, receipt, persisted, catalog, hash);
}

async function witnessCare(page, requests, petID, hash, statusPublications) {
  const before = await snapshot(page);
  const pet = before.features?.pet_adoption?.pets?.find((row) => row.pet_id === petID);
  if (!pet?.eligible_action_ids.includes("care.feed")) throw new Error("Garage care real adopted pet cannot be fed");
  assert.equal(pet.status_band, "low", "status fixture must start with a genuinely low adopted pet");
  const control = page.locator(".pet-care").getByRole("button", { name: plainFixtureCopy("pet.care.action.feed.title"), exact: true });
  let receipt;
  const hostReads = [];
  const observeRead = (response) => {
    if (response.request().method() === "GET" && new URL(response.url()).pathname === "/api/v1/founder/state") {
      hostReads.push(response);
    }
  };
  page.on("response", observeRead);
  try { receipt = await founderDOMIntent(page, requests, control, "care_action", { pet_id: petID, action_id: "care.feed" }); }
  catch (error) {
    const emitted = requests.some((row) => row.method() === "POST" &&
      new URL(row.url()).pathname === "/api/v1/intents" && row.postDataJSON()?.kind === "care_action");
    if (!emitted) throw new Error("Garage care DOM callback emitted no care intent", { cause: error });
    throw error;
  } finally { page.off("response", observeRead); }
  const request = requests.filter((row) => row.method() === "POST" &&
    new URL(row.url()).pathname === "/api/v1/intents" && row.postDataJSON()?.kind === "care_action").at(-1)?.postDataJSON();
  if (receipt.intent_id !== request?.intent_id || receipt.pet_id !== petID || receipt.action_id !== "care.feed" ||
      !Number.isSafeInteger(receipt.applied_ppm) || receipt.applied_ppm <= 0 ||
      !Number.isSafeInteger(receipt.before_ppm) || !Number.isSafeInteger(receipt.after_ppm) || receipt.after_ppm <= receipt.before_ppm) {
    throw new Error("Garage care DOM action returned no bound positive actual care receipt");
  }
  // Observe the host's own authoritative refresh before issuing any additional
  // diagnostic read. A later test fetch is a different attendance sample and
  // cannot tell us which band the mounted host should have rendered.
  const refreshes = await Promise.all(hostReads.map(async (response) => {
    assert.equal(response.status(), 200, "care host refresh succeeded");
    return response.json();
  }));
  const refreshed = refreshes.filter((view) => view.founder_revision === receipt.founder_revision).at(-1);
  assert.ok(refreshed, "care action must produce an authoritative host refresh at the applied revision");
  const projected = assertPersistedCare(refreshed, receipt, hash);
  await page.locator(".pet-care").getByText(plainFixtureCopy(`pet.care.band.${projected.status_band}`), { exact: true }).waitFor({ state: "visible", timeout: 30_000 });
  await page.waitForFunction(() => document.querySelector('.pet-care button[data-action-id="care.feed"]')?.disabled === true, undefined, { timeout: 30_000 });
  assertPersistedCare(await snapshot(page), receipt, hash);
  const statusEvents = JSON.parse(testDatabaseSQL(`SELECT coalesce(jsonb_agg(jsonb_build_object(
    'event_id',event_id,'revision',revision,'payload',payload)), '[]'::jsonb)
    FROM events WHERE intent_id='${receipt.intent_id}' AND kind='pet_status_changed.v1';`, true).trim());
  assert.equal(statusEvents.length, 1, "feed must store exactly one actual status change");
  const changed = statusEvents[0];
  assert.equal(changed.revision, receipt.founder_revision, "status event binds the care command revision");
  assert.deepEqual(changed.payload, { pet_id: petID, from_status_band: "low", to_status_band: "high" });
  const catalog = JSON.parse(readFileSync(path.join(clientRoot, "src/copy/generated/catalog.json"), "utf8"));
  const copy = catalog.entries.find((entry) => entry.key === "pet.status_changed_announcement");
  assert.deepEqual(copy?.params, [{ name: "band", type: "string" }]);
  assert.equal(copy.era_variants, null, "status fixture copy is era-independent");
  const sentence = copy.text.replace("{band}", plainFixtureCopy("pet.care.band.high"));
  try {
    await page.locator('.announcement[role="status"]').getByText(sentence, { exact: true }).waitFor({ state: "visible", timeout: 30_000 });
  } catch (error) {
    throw new Error(`pet status announcement missing: ${JSON.stringify({ changed, statusPublications,
      host: await page.locator("main").innerText(),
      relatedEvents: JSON.parse(testDatabaseSQL(`SELECT coalesce(jsonb_agg(jsonb_build_object(
        'revision',revision,'kind',kind,'intent_id',intent_id,'payload',payload) ORDER BY revision,occurred_at),'[]'::jsonb)
        FROM events WHERE stream_id=(SELECT stream_id FROM events WHERE event_id='${changed.event_id}');`, true).trim()) })}`, { cause: error });
  }
  const publications = statusPublications.filter((event) => event.payload?.event_id === changed.event_id);
  assert.ok(publications.length > 0, "stored status event must reach the real socket");
  // Transport may replay an event; the host must deduplicate its announcement,
  // not demand exactly-once network delivery. Every observed copy stays bound.
  for (const event of publications) {
    assert.equal(event.payload.scope, "founder");
    assert.equal(event.rev, changed.revision);
    assert.deepEqual(event.payload.payload, changed.payload);
  }
  console.log("composed RP-318: adopted low pet → native feed → exact SQL low/high event → actual WebSocket → visible polite status: PASS");
  console.log(`composed Garage care: DOM care.feed → exact stored receipt/state → actual host read at attendance ${projected.attended_ms} → rendered band ${projected.status_band}/feed ineligible at Founder revision ${receipt.founder_revision}: PASS`);
  return receipt;
}

async function openPetSurface(page) {
  await page.getByRole("button", { name: plainFixtureCopy("pet.care.panel.title"), exact: true }).click();
  await page.locator('main[data-surface="pet"]').waitFor({ state: "visible", timeout: 30_000 });
}

async function assertLivePetOverlay(page, { present, reducedMotion = false }) {
  const surface = page.locator(".pet-care");
  const overlay = surface.locator('.portrait .overlay[data-render="horse_armor"]');
  const overlayState = () => page.evaluate(() => {
    const main = document.querySelector("main");
    const surface = main?.querySelector(".pet-care");
    const overlays = surface?.querySelectorAll('.portrait .overlay[data-render="horse_armor"]');
    return { surface: main?.getAttribute("data-surface") ?? null, busy: main?.getAttribute("aria-busy") ?? null,
      pets: surface?.querySelectorAll("article.pet").length ?? 0, overlays: overlays?.length ?? 0,
      stale: Boolean(surface?.querySelector(".care-stale")), pending: Boolean(surface?.querySelector(".care-pending")) };
  });
  if (!present) {
    if (await surface.locator(".overlay").count() !== 0) throw new Error("cosmetic G10 live overlay remained after unequip");
    return;
  }
  if (await overlay.count() !== 1 || !await overlay.isVisible()) {
    throw new Error(`cosmetic G10 live pet overlay missing after equip: ${JSON.stringify(await overlayState())}`);
  }
  const details = await overlay.evaluate((node) => {
    const walker = document.createTreeWalker(node, NodeFilter.SHOW_TEXT);
    const text = [];
    while (walker.nextNode()) if (walker.currentNode.textContent.trim()) text.push(walker.currentNode.textContent);
    return { reaction: node.dataset.reaction, animate: node.dataset.animate, hidden: node.getAttribute("aria-hidden"), text,
      earsVisible: getComputedStyle(node.querySelector(".ears-back")).display !== "none",
      animation: getComputedStyle(node.querySelector(".flick")).animationName,
      reducedMotion: matchMedia("(prefers-reduced-motion: reduce)").matches };
  });
  if (details.reaction !== "annoyed" || details.hidden !== "true" || details.text.length !== 0 || !details.earsVisible ||
      details.reducedMotion !== reducedMotion ||
      (reducedMotion ? details.animation !== "none" : details.animation === "none")) {
    throw new Error(`cosmetic G10 live pet reaction/presentation invalid: ${JSON.stringify(details)}`);
  }
}

try {
  const hash = buildFixtureRoot();
  resetTestDatabase();
  await waitForPort(18082);
  await waitForPort(5173);
  const binary = path.join(repositoryRoot, ".cache", "cosmetic-ac14-gameserver");
  mkdirSync(path.dirname(binary), { recursive: true });
  const environment = {
    ...process.env,
    CLOUD_CLICKER_ACTIVITY_BRACKET: "activity.standard",
    CLOUD_CLICKER_BOOTSTRAP_KEY: Buffer.alloc(32, 7).toString("base64"),
    CLOUD_CLICKER_BOOTSTRAP_KEY_ID: "cosmetic-fixture",
    CLOUD_CLICKER_CURSOR_KEY: Buffer.alloc(32, 7).toString("base64"),
    CLOUD_CLICKER_JWT_KEY: Buffer.alloc(32, 7).toString("base64"),
    CLOUD_CLICKER_REPOSITORY_ROOT: fixtureRoot,
    CLOUD_CLICKER_SERVER_ID: "01986666-b001-4000-8000-000000000002",
    DATABASE_URL: databaseURL,
    GOCACHE: path.join(repositoryRoot, ".cache", "go-build"),
    LISTEN_ADDR: "127.0.0.1:18082",
  };
  const goBuild = spawnSync("go", ["build", "-o", binary, "./cmd/gameserver"], { cwd: path.join(repositoryRoot, "server"), env: environment, encoding: "utf8" });
  if (goBuild.status !== 0) throw new Error(`cosmetic gameserver build failed: ${goBuild.stderr || goBuild.stdout}`);
  gameserver = spawn(binary, [], { cwd: repositoryRoot, env: environment, stdio: ["ignore", "pipe", "pipe"] });
  gameserver.stdout.on("data", (value) => process.stdout.write(value));
  gameserver.stderr.on("data", (value) => process.stderr.write(value));
  gameserver.on("error", (error) => processErrors.push(error));
  await waitForReady();
  await build({ configFile: path.join(clientRoot, "vite.config.ts"), root: clientRoot, logLevel: "error",
    build: { outDir: clientDist, emptyOutDir: true }, plugins: n5Fault ? [cosmeticN5Fixture(n5Fault)] : [] });
  staticServer = await serveBuiltClient();
  browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });
  const violations = [];
  const directViolations = [];
  const requests = [];
  const websocketEvents = [];
  const statusPublications = [];
  const pageErrors = [];
  const failedRequests = [];
  const safetyBlockedRequests = [];
  const acquireResponses = [];
  const isAcquire = (request) => request.method() === "POST" &&
    new URL(request.url()).pathname === "/api/v1/intents" && request.postDataJSON()?.kind === "acquire_cosmetic";
  page.on("pageerror", (error) => pageErrors.push(error));
  page.on("requestfailed", (request) => failedRequests.push(`${request.url()}: ${request.failure()?.errorText}`));
  page.on("request", (request) => { requests.push(request); try { assertNetwork(request.url(), "request"); } catch (error) { violations.push(error.message); } });
  page.on("response", (response) => {
    if (isAcquire(response.request())) acquireResponses.push({ status: response.status() });
  });
  page.on("websocket", (socket) => {
    websocketEvents.push({ url: socket.url(), sent: 0, received: 0, closed: false, errors: 0 });
    const state = websocketEvents.at(-1);
    socket.on("framesent", () => { state.sent += 1; });
    socket.on("framereceived", (frame) => {
      state.received += 1;
      for (const line of String(frame.payload).trim().split("\n")) {
        if (!line) continue;
        const event = JSON.parse(line).push?.pub?.data;
        if (event?.kind === "event" && event.payload?.kind === "pet_status_changed.v1") statusPublications.push(event);
      }
    });
    socket.on("close", () => { state.closed = true; });
    socket.on("socketerror", () => { state.errors += 1; });
    try { assertNetwork(socket.url(), "websocket"); } catch (error) { violations.push(error.message); }
  });
  if (n5Fault) await page.route("**/*", async (route) => {
    // Independent local-only safety net for deliberately bad fixture builds.
    // If reached, the negative verifier refuses to credit the primary trap.
    try { assertNetwork(route.request().url(), "request"); }
    catch { safetyBlockedRequests.push(route.request().url()); await route.abort("blockedbyclient"); return; }
    await route.continue();
  });
  await page.addInitScript(() => {
    const failures = [];
    globalThis.__cosmeticN5Failures = failures;
    globalThis.__cosmeticN5FixtureExecutions = [];
    window.addEventListener("cosmetic-n5-fixture", (event) => { globalThis.__cosmeticN5FixtureExecutions.push(event.detail); });
    // Passive input trace includes the native keyboard Buy activation below;
    // no retries, network bodies/tokens or gameplay API shortcuts.
    globalThis.__cosmeticBuyTrace = [];
    globalThis.__cosmeticActionTrace = null;
    for (const type of ["pointerdown", "pointerup", "click", "keydown", "keyup"]) document.addEventListener(type, (event) => {
      const traces = [globalThis.__cosmeticBuyTrace, globalThis.__cosmeticActionTrace].filter(Array.isArray);
      if (traces.length === 0) return;
      const target = event.target instanceof Element ? event.target : null;
      const button = target?.closest("button");
      const main = document.querySelector("main");
      const item = button?.closest("[data-cosmetic]");
      const row = { type, time: performance.now(), trusted: event.isTrusted, key: event.key ?? null,
        focused: button ? document.activeElement === button : false,
        target: target?.tagName ?? null, button: button?.textContent ?? null,
        disabled: button?.disabled ?? null, connected: button?.isConnected ?? null,
        cosmetic: item?.getAttribute("data-cosmetic") ?? null, state: item?.getAttribute("data-state") ?? null,
        surface: main?.getAttribute("data-surface") ?? null, busy: main?.getAttribute("aria-busy") ?? null };
      for (const trace of traces) {
        trace.push(row);
        if (trace.length > 32) { trace.shift(); trace.dropped = (trace.dropped ?? 0) + 1; }
      }
    }, { capture: true, passive: true });
    const originalFetch = window.fetch;
    window.fetch = (input, init) => {
      const url = typeof input === "string" ? input : input instanceof URL ? input.href : input.url;
      const parsed = new URL(url, location.href);
      if (parsed.origin !== location.origin || !/^\/api\//u.test(parsed.pathname)) {
        failures.push(`fetch ${url}`);
        return Promise.reject(new TypeError("cosmetic network trap"));
      }
      return originalFetch.call(window, input, init);
    };
    const OriginalSocket = window.WebSocket;
    window.WebSocket = class extends OriginalSocket {
      constructor(url, protocols) {
        const parsed = new URL(String(url), location.href);
        if (parsed.host !== location.host || parsed.pathname !== "/connection/websocket") {
          failures.push(`websocket ${url}`);
          throw new TypeError("cosmetic network trap");
        }
        super(url, protocols);
      }
    };
    window.open = (url) => { failures.push(`window.open ${url}`); return null; };
    window.PaymentRequest = function PaymentRequestTrap() { failures.push("PaymentRequest"); throw new TypeError("payment blocked"); };
    if (navigator.credentials) for (const name of ["get", "create", "store"]) {
      navigator.credentials[name] = () => { failures.push(`navigator.credentials.${name}`); return Promise.reject(new TypeError("credentials blocked")); };
    }
  });
  async function requireCleanN5(phase) {
    const direct = [...directViolations, ...await page.evaluate(() => globalThis.__cosmeticN5Failures)];
    if (violations.length || direct.length || pageErrors.length || safetyBlockedRequests.length) {
      const report = { phase, fixture: n5Fault,
        fixtureExecutions: await page.evaluate(() => globalThis.__cosmeticN5FixtureExecutions),
        violations, directViolations: direct, safetyBlockedRequests, pageErrors: pageErrors.map(String) };
      console.error(`${n5RejectionPrefix}${JSON.stringify(report)}`);
      throw new Error("cosmetic N5 network/payment gate rejected the built client");
    }
  }
  await page.goto(uiURL, { waitUntil: "networkidle" });
  await page.getByRole("button", { name: "BEGIN ATTEMPT" }).click();
  await page.locator('main[data-surface="desk"]').waitFor({ state: "visible", timeout: 30_000 });
  try {
    await page.getByText(/You are visitor #\d+/u).waitFor({ state: "visible", timeout: 30_000 });
  } catch (error) {
    throw new Error(`cosmetic live transport/presence did not reach the visitor counter: ${JSON.stringify({ websocketEvents, failedRequests, pageErrors: pageErrors.map(String), main: (await page.locator("main").innerText()).slice(0, 500) })}`, { cause: error });
  }
  const initial = await snapshot(page);
  if (axisFixture) await assertAxis(page, { input: 0, product: "1e0", factors: ["1e0", "1e0"] });
  const firstArm = initial.features?.cosmetics;
  if (initial.constants_hash !== hash || firstArm?.active !== true || firstArm.items?.[0]?.cosmetic_id !== "horse_armor" || firstArm.items[0].owned !== false || firstArm.items[0].acquirable !== false || await page.locator('[data-testid="cosmetic-shelf"]').count() !== 0) {
    throw new Error(`cosmetic AC14 bootstrap/locked arm invalid: ${JSON.stringify({ hash: initial.constants_hash, arm: firstArm })}`);
  }
  seedGateRequirement(initial.run.founder_id);
  directViolations.push(...await page.evaluate(() => globalThis.__cosmeticN5Failures));
  await page.reload({ waitUntil: "networkidle" });
  const crossGate = page.getByRole("button", { name: "Move Into the Garage", exact: true });
  await crossGate.waitFor({ state: "visible", timeout: 30_000 });
  if (!await crossGate.isEnabled()) throw new Error("cosmetic AC14 T1 cross-gate disabled after server-side setup");
  await crossGate.click();
  const shelf = page.locator('[data-testid="cosmetic-shelf"]');
  await shelf.waitFor({ state: "visible", timeout: 30_000 });
  const beforeBuy = await snapshot(page);
  if (beforeBuy.run.tier !== 1 || beforeBuy.features?.cosmetics?.items?.[0]?.acquirable !== true || beforeBuy.features.cosmetics.items[0].owned !== false) {
    throw new Error(`cosmetic AC14 T1 server snapshot invalid: ${JSON.stringify(beforeBuy.features?.cosmetics)}`);
  }
  if (axisFixture) await witnessAxis(page, requests);
  const buy = shelf.getByRole("button", { name: /^Buy for/u });
  // A completed server snapshot can precede the Svelte/transport-ready render
  // by a frame. Require the actual control to become actionable, without
  // dispatching a second purchase or treating a disabled button as success.
  const beforeIntents = requests.filter(isAcquire).length;
  let phase = "actionability", response, boundary, buyLabel;
  async function buyBoundary() {
    const dom = await page.evaluate(() => {
      const events = globalThis.__cosmeticBuyTrace;
      if (!Array.isArray(events)) throw new Error("cosmetic Buy trace unavailable");
      globalThis.__cosmeticBuyTrace = null;
      const item = document.querySelector('[data-testid="cosmetic-shelf"] [data-cosmetic="horse_armor"]');
      const button = item?.querySelector("button");
      return { events, dropped_events: events.dropped ?? 0, state: item?.getAttribute("data-state") ?? null,
        disabled: button?.disabled ?? null, busy: document.querySelector("main")?.getAttribute("aria-busy") ?? null };
    });
    const emitted = requests.filter(isAcquire).slice(beforeIntents);
    return { phase, dom, emitted: emitted.length, responses: acquireResponses,
      requestFailures: emitted.map((request) => request.failure()?.errorText ?? null),
      pageErrors: pageErrors.map((error) => error.message), websocketEvents };
  }
  try {
    await buy.click({ trial: true, timeout: 30_000 });
    buyLabel = await buy.innerText();
    await buy.focus();
    assert.equal(await buy.evaluate((control) => document.activeElement === control), true, "Buy must receive focus before native input");
    await page.evaluate(() => { globalThis.__cosmeticBuyTrace = []; });
    phase = "native-keyboard-activation";
    // Handle rejection immediately even if click itself fails first. This
    // keeps the original response timeout without an unhandled second error.
    const appliedResponse = page.waitForResponse((value) => isAcquire(value.request()), { timeout: 30_000 })
      .then((value) => ({ value }), (error) => ({ error }));
    // Focus is setup; the browser's native Enter/Space default action must
    // activate the real button. No DOM click fallback or intent retry.
    await page.keyboard.press(axisFixture ? "Space" : "Enter");
    phase = "response";
    const result = await appliedResponse;
    if (result.error) throw result.error;
    response = result.value;
  } catch (error) {
    throw new Error(`cosmetic AC14 Buy boundary failed: ${JSON.stringify(await buyBoundary())}`, { cause: error });
  }
  boundary = await buyBoundary();
  console.log(`cosmetic AC14 Buy boundary: ${JSON.stringify(boundary)}`);
  assertNativeCosmeticBuyTrace(boundary.dom, buyKey, buyLabel);
  const receipt = await response.json();
  if (response.status() !== 200 || receipt.outcome !== "applied" || receipt.event?.kind !== "cosmetic_acquired.v1" || receipt.event?.payload?.cosmetic_id !== "horse_armor" || receipt.event.payload.order_number !== 1) {
    throw new Error(`cosmetic AC14 Buy returned no applied zero-price receipt: ${JSON.stringify(receipt)}`);
  }
  const acquired = requests.filter((request) => request.method() === "POST" && new URL(request.url()).pathname === "/api/v1/intents" && request.postDataJSON()?.kind === "acquire_cosmetic");
  const acquireBody = acquired.at(-1)?.postDataJSON();
  if (acquired.length !== beforeIntents + 1 || acquireBody?.cosmetic_id !== "horse_armor") {
    throw new Error(`cosmetic AC14 Buy issued ${acquired.length - beforeIntents} acquire requests`);
  }
  assert.deepEqual(Object.keys(acquireBody).sort(), ["cosmetic_id", "expected_revision", "intent_id", "kind"]);
  assert.equal(acquireBody.expected_revision, beforeBuy.founder_revision);
  assert.equal(receipt.intent_id, acquireBody.intent_id, "receipt must belong to this native activation");
  assert.equal(receipt.founder_revision, beforeBuy.founder_revision + 1);
  await shelf.locator('[data-cosmetic="horse_armor"][data-state="owned"]').waitFor({ state: "visible", timeout: 30_000 });
  await shelf.locator("p.receipt[role=status]").waitFor({ state: "visible", timeout: 30_000 });
  await shelf.locator(".owned").evaluate((owned) => {
    if (document.activeElement !== owned) throw new Error("native Buy lost focus instead of transferring it to the owned state");
  });
  await requireCleanN5("after-native-buy");
  directViolations.push(...await page.evaluate(() => globalThis.__cosmeticN5Failures));
  await page.reload({ waitUntil: "networkidle" });
  await shelf.locator('[data-cosmetic="horse_armor"][data-state="owned"]').waitFor({ state: "visible", timeout: 30_000 });
  const afterReload = await snapshot(page);
  if (afterReload.features?.cosmetics?.items?.[0]?.owned !== true || await shelf.locator("p.receipt[role=status]").count() !== 0) {
    throw new Error(`cosmetic AC14 reload did not recover server-owned state: ${JSON.stringify(afterReload.features?.cosmetics)}`);
  }
  assert.equal(afterReload.founder_revision, receipt.founder_revision);
  console.log(`composed Cosmetic native ${JSON.stringify(buyKey)}: trusted input → one bound applied receipt → owned focus → persisted reload: PASS`);

  // G10: real adoption and wearing, not an injected pet or a literal equipped
  // snapshot. All gameplay writes originate from the built client's DOM.
  const availability = afterReload.features?.pet_adoption?.pet_adoption;
  if (!availability || availability.count !== 0) throw new Error("cosmetic G10 bootstrap lacks the empty real adoption producer");
  const nameKey = await page.locator('input[name="pet-name"]:checked').inputValue();
  const adoption = await founderDOMIntent(page, requests,
    page.getByRole("button", { name: plainFixtureCopy("pet.adoption.action.adopt"), exact: true }), "adopt_pet",
    { species_id: availability.starter_species_id, name_key: nameKey });
  assert.equal(await page.locator("#pet-adoption-heading").evaluate((heading) => document.activeElement === heading),
    true, "authoritative native adoption must focus the welcome heading");
  const petID = adoption.pet_id;
  if (typeof petID !== "string" || !/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/u.test(petID)) {
    throw new Error("cosmetic G10 adoption returned no real pet identity");
  }
  assertPersistedWearer(await snapshot(page), petID, null);
  const equip = await founderDOMIntent(page, requests, shelf.getByRole("button"), "equip_cosmetic",
    { cosmetic_id: "horse_armor", pet_id: petID });
  if (equip.event?.kind !== "cosmetic_equipped.v1" || equip.event.payload?.pet_id !== petID ||
      equip.event.payload?.cosmetic_id !== "horse_armor" || equip.event.payload?.replaced_cosmetic_id !== null) {
    throw new Error(`cosmetic G10 equip returned no exact wearing event: ${JSON.stringify(equip)}`);
  }
  assertPersistedWearer(await snapshot(page), petID, "horse_armor");
  await openPetSurface(page);
  await assertLivePetOverlay(page, { present: true });
  const care = await witnessCare(page, requests, petID, hash, statusPublications);
  directViolations.push(...await page.evaluate(() => globalThis.__cosmeticN5Failures));
  await page.reload({ waitUntil: "networkidle" });
  assertPersistedWearer(await snapshot(page), petID, "horse_armor");
  assertPersistedCare(await snapshot(page), care, hash);
  await openPetSurface(page);
  await assertLivePetOverlay(page, { present: true });
  const restoredFeed = page.locator(".pet-care").getByRole("button", { name: plainFixtureCopy("pet.care.action.feed.title"), exact: true });
  if (!await restoredFeed.isDisabled()) throw new Error("Garage care reload lost persisted feed ineligibility");
  console.log("composed Garage care: reload retains the actual public care outcome and feed ineligibility: PASS");

  // Exercise the browser preference and the fresh mounted host's propagation,
  // not merely a fixture prop named reducedMotion on the isolated component.
  await page.emulateMedia({ reducedMotion: "reduce" });
  directViolations.push(...await page.evaluate(() => globalThis.__cosmeticN5Failures));
  await page.reload({ waitUntil: "networkidle" });
  assertPersistedWearer(await snapshot(page), petID, "horse_armor");
  await openPetSurface(page);
  await assertLivePetOverlay(page, { present: true, reducedMotion: true });
  await page.locator("nav button").first().click();
  await page.locator('main[data-surface="desk"]').waitFor({ state: "visible", timeout: 30_000 });
  const unequip = await founderDOMIntent(page, requests, shelf.getByRole("button"), "unequip_cosmetic", { pet_id: petID });
  if (unequip.event?.kind !== "cosmetic_unequipped.v1" || unequip.event.payload?.pet_id !== petID || unequip.event.payload?.cosmetic_id !== "horse_armor") {
    throw new Error(`cosmetic G10 unequip returned no exact wearing event: ${JSON.stringify(unequip)}`);
  }
  assertPersistedWearer(await snapshot(page), petID, null);
  await openPetSurface(page);
  await assertLivePetOverlay(page, { present: false });
  directViolations.push(...await page.evaluate(() => globalThis.__cosmeticN5Failures));
  await page.reload({ waitUntil: "networkidle" });
  assertPersistedWearer(await snapshot(page), petID, null);
  await openPetSurface(page);
  await assertLivePetOverlay(page, { present: false });
  await requireCleanN5("complete");
  console.log(`composed Cosmetic AC14/G10: T0 locked → T1 Buy → owned reload → DOM adopt/equip → live annoyed overlay → worn reload/reduced motion → DOM unequip → unworn reload; N5 requests ${requests.length}, no violation; ${(Date.now() - startedAt) / 1000}s: PASS`);
} finally {
  await browser?.close();
  for (const socket of proxySockets) socket.destroy();
  if (staticServer) await new Promise((resolve, reject) => staticServer.close((error) => error ? reject(error) : resolve()));
  await stopGameserver();
  rmSync(fixtureRoot, { recursive: true, force: true });
  if (n5Fault) console.log(n5CleanupMarker);
}
