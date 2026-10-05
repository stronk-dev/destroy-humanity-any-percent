import { spawn, spawnSync } from "node:child_process";
import { createHash } from "node:crypto";
import { copyFileSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { createServer as createHTTPServer, request as httpRequest } from "node:http";
import { createServer as createTCPServer, connect as connectTCP } from "node:net";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { chromium } from "playwright";
import { build } from "vite";

const clientRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const repositoryRoot = path.resolve(clientRoot, "..");
const gameserverURL = "http://127.0.0.1:18082";
const uiURL = "http://localhost:5173";
const databaseURL = "postgres://cloud_clicker:cloud_clicker_game_ui_test@127.0.0.1:55433/cloud_clicker_game_ui_test?sslmode=disable";
const fixtureRoot = mkdtempSync(path.join(os.tmpdir(), "cloud-clicker-cosmetic-ac14-"));
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
  const economy = JSON.parse(artifacts.get("economy").toString("utf8"));
  economy.multiplier_sources.push({ id: "reputation.founder_bonus", slot: "prestige", target: "all", provider: "reputation_tree" });
  artifacts.set("economy", Buffer.from(JSON.stringify(economy)));
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

function testDatabaseSQL(sql) {
  const result = spawnSync("docker", ["compose", "-f", "compose.game-ui-test.yml", "exec", "-T", "game-ui-postgres",
    "psql", "-v", "ON_ERROR_STOP=1", "-U", "cloud_clicker", "-d", "cloud_clicker_game_ui_test", "-c", sql],
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
    SELECT stream_id,revision+1,version,jsonb_set(state,'{balances,company.cash}',to_jsonb('1e5'::text),false),constants_hash FROM current;`);
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
  const dist = path.join(clientRoot, "dist");
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

async function founderDOMIntent(page, requests, control, kind, expectedFields) {
  const before = await snapshot(page);
  const matching = () => requests.filter((request) => request.method() === "POST" &&
    new URL(request.url()).pathname === "/api/v1/intents" && request.postDataJSON()?.kind === kind);
  const priorCount = matching().length;
  await control.click({ trial: true, timeout: 30_000 });
  const responseTask = page.waitForResponse((response) => response.request().method() === "POST" &&
    new URL(response.url()).pathname === "/api/v1/intents" && response.request().postDataJSON()?.kind === kind, { timeout: 30_000 });
  await control.click();
  const response = await responseTask;
  const receipt = await response.json();
  const emitted = matching();
  const body = emitted.at(-1)?.postDataJSON();
  const requiredKeys = ["intent_id", "kind", "expected_revision", ...Object.keys(expectedFields)].sort();
  if (emitted.length !== priorCount + 1 || !body || Object.keys(body).sort().join("\0") !== requiredKeys.join("\0") ||
      body.expected_revision !== before.founder_revision || Object.entries(expectedFields).some(([key, value]) => body[key] !== value) ||
      response.status() !== 200 || receipt.outcome !== "applied" || receipt.founder_revision !== before.founder_revision + 1) {
    throw new Error(`cosmetic G10 ${kind} did not emit one exact Founder-scoped applied intent: ${JSON.stringify({ body, receipt, emitted: emitted.length - priorCount })}`);
  }
  return receipt;
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

async function openPetSurface(page) {
  await page.getByRole("button", { name: plainFixtureCopy("pet.care.panel.title"), exact: true }).click();
  await page.locator('main[data-surface="pet"]').waitFor({ state: "visible", timeout: 30_000 });
}

async function assertLivePetOverlay(page, { present, reducedMotion = false }) {
  const surface = page.locator(".pet-care");
  const overlay = surface.locator('.portrait .overlay[data-render="horse_armor"]');
  if (!present) {
    if (await surface.locator(".overlay").count() !== 0) throw new Error("cosmetic G10 live overlay remained after unequip");
    return;
  }
  if (await overlay.count() !== 1 || !await overlay.isVisible()) throw new Error("cosmetic G10 live pet overlay missing after equip");
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
  await build({ configFile: path.join(clientRoot, "vite.config.ts"), root: clientRoot, logLevel: "error" });
  staticServer = await serveBuiltClient();
  browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });
  const violations = [];
  const directViolations = [];
  const requests = [];
  const websocketEvents = [];
  const pageErrors = [];
  const failedRequests = [];
  page.on("pageerror", (error) => pageErrors.push(error));
  page.on("requestfailed", (request) => failedRequests.push(`${request.url()}: ${request.failure()?.errorText}`));
  page.on("request", (request) => { requests.push(request); try { assertNetwork(request.url(), "request"); } catch (error) { violations.push(error.message); } });
  page.on("websocket", (socket) => {
    websocketEvents.push({ url: socket.url(), sent: 0, received: 0, closed: false, errors: 0 });
    const state = websocketEvents.at(-1);
    socket.on("framesent", () => { state.sent += 1; });
    socket.on("framereceived", () => { state.received += 1; });
    socket.on("close", () => { state.closed = true; });
    socket.on("socketerror", () => { state.errors += 1; });
    try { assertNetwork(socket.url(), "websocket"); } catch (error) { violations.push(error.message); }
  });
  await page.addInitScript(() => {
    const failures = [];
    globalThis.__cosmeticN5Failures = failures;
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
  await page.goto(uiURL, { waitUntil: "networkidle" });
  await page.getByRole("button", { name: "BEGIN ATTEMPT" }).click();
  await page.locator('main[data-surface="desk"]').waitFor({ state: "visible", timeout: 30_000 });
  try {
    await page.getByText(/You are visitor #\d+/u).waitFor({ state: "visible", timeout: 30_000 });
  } catch (error) {
    throw new Error(`cosmetic live transport/presence did not reach the visitor counter: ${JSON.stringify({ websocketEvents, failedRequests, pageErrors: pageErrors.map(String), main: (await page.locator("main").innerText()).slice(0, 500) })}`, { cause: error });
  }
  const initial = await snapshot(page);
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
  const buy = shelf.getByRole("button", { name: /^Buy for/u });
  // A completed server snapshot can precede the Svelte/transport-ready render
  // by a frame. Require the actual control to become actionable, without
  // dispatching a second purchase or treating a disabled button as success.
  try { await buy.click({ trial: true, timeout: 30_000 }); }
  catch (error) { throw new Error(`cosmetic AC14 Buy control did not become actionable: ${await shelf.innerText()}`, { cause: error }); }
  const beforeIntents = requests.filter((request) => request.method() === "POST" && new URL(request.url()).pathname === "/api/v1/intents" && request.postDataJSON()?.kind === "acquire_cosmetic").length;
  const appliedResponse = page.waitForResponse((response) => response.request().method() === "POST" &&
    new URL(response.url()).pathname === "/api/v1/intents" && response.request().postDataJSON()?.kind === "acquire_cosmetic", { timeout: 30_000 });
  await buy.click();
  const response = await appliedResponse;
  const receipt = await response.json();
  if (response.status() !== 200 || receipt.outcome !== "applied" || receipt.event?.kind !== "cosmetic_acquired.v1" || receipt.event?.payload?.cosmetic_id !== "horse_armor" || receipt.event.payload.order_number !== 1) {
    throw new Error(`cosmetic AC14 Buy returned no applied zero-price receipt: ${JSON.stringify(receipt)}`);
  }
  const acquired = requests.filter((request) => request.method() === "POST" && new URL(request.url()).pathname === "/api/v1/intents" && request.postDataJSON()?.kind === "acquire_cosmetic");
  if (acquired.length !== beforeIntents + 1 || acquired.at(-1).postDataJSON()?.cosmetic_id !== "horse_armor") {
    throw new Error(`cosmetic AC14 Buy issued ${acquired.length - beforeIntents} acquire requests`);
  }
  await shelf.locator('[data-cosmetic="horse_armor"][data-state="owned"]').waitFor({ state: "visible", timeout: 30_000 });
  await shelf.locator("p.receipt[role=status]").waitFor({ state: "visible", timeout: 30_000 });
  directViolations.push(...await page.evaluate(() => globalThis.__cosmeticN5Failures));
  await page.reload({ waitUntil: "networkidle" });
  await shelf.locator('[data-cosmetic="horse_armor"][data-state="owned"]').waitFor({ state: "visible", timeout: 30_000 });
  const afterReload = await snapshot(page);
  if (afterReload.features?.cosmetics?.items?.[0]?.owned !== true || await shelf.locator("p.receipt[role=status]").count() !== 0) {
    throw new Error(`cosmetic AC14 reload did not recover server-owned state: ${JSON.stringify(afterReload.features?.cosmetics)}`);
  }

  // G10: real adoption and wearing, not an injected pet or a literal equipped
  // snapshot. All gameplay writes originate from the built client's DOM.
  const availability = afterReload.features?.pet_adoption?.pet_adoption;
  if (!availability || availability.count !== 0) throw new Error("cosmetic G10 bootstrap lacks the empty real adoption producer");
  const nameKey = await page.locator('input[name="pet-name"]:checked').inputValue();
  const adoption = await founderDOMIntent(page, requests,
    page.getByRole("button", { name: plainFixtureCopy("pet.adoption.action.adopt"), exact: true }), "adopt_pet",
    { species_id: availability.starter_species_id, name_key: nameKey });
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
  directViolations.push(...await page.evaluate(() => globalThis.__cosmeticN5Failures));
  await page.reload({ waitUntil: "networkidle" });
  assertPersistedWearer(await snapshot(page), petID, "horse_armor");
  await openPetSurface(page);
  await assertLivePetOverlay(page, { present: true });

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
  directViolations.push(...await page.evaluate(() => globalThis.__cosmeticN5Failures));
  if (violations.length || directViolations.length || pageErrors.length) {
    throw new Error(`cosmetic AC14 network/page violations: ${JSON.stringify({ violations, directViolations, pageErrors: pageErrors.map(String) })}`);
  }
  console.log(`composed Cosmetic AC14/G10: T0 locked → T1 Buy → owned reload → DOM adopt/equip → live annoyed overlay → worn reload/reduced motion → DOM unequip → unworn reload; N5 requests ${requests.length}, no violation; ${(Date.now() - startedAt) / 1000}s: PASS`);
} finally {
  await browser?.close();
  for (const socket of proxySockets) socket.destroy();
  if (staticServer) await new Promise((resolve, reject) => staticServer.close((error) => error ? reject(error) : resolve()));
  await stopGameserver();
  rmSync(fixtureRoot, { recursive: true, force: true });
}
