import assert from "node:assert/strict";
import { spawn, spawnSync } from "node:child_process";
import { createHash } from "node:crypto";
import { copyFileSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { createServer, request as httpRequest } from "node:http";
import { createServer as createTCPServer, connect } from "node:net";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { chromium } from "playwright";
import { productionClientFiles, productionClientProof } from "./production-client-proof.mjs";
import { navigationAPIProof, pipeObservedAPIResponse } from "./navigation-api-proof.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");
const uiURL = "http://localhost:5173";
const serverURL = "http://127.0.0.1:18083";
assert(process.argv.length === 2 || process.argv.length === 3 && ["--session-diagnostic", "--harvest-fixture"].includes(process.argv[2]), "unknown Garden driver argument");
const sessionDiagnostic = process.argv[2] === "--session-diagnostic";
const harvestFixture = process.argv[2] === "--harvest-fixture";
const fixtureRoot = mkdtempSync(path.join(os.tmpdir(), "cloud-clicker-garden-composed-"));
const started = Date.now();
const sockets = new Set();
const errors = [];
const apiBoundaries = [];
const apiProof = navigationAPIProof(), proxyBoundaries = apiProof.boundaries;
let server, assets, browser, heartbeat;
const pause = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
const copy = JSON.parse(readFileSync(path.join(root, "client/src/copy/generated/catalog.json"), "utf8"));
function text(key, params = {}) {
  const row = copy.entries.find((entry) => entry.key === key);
  assert(row && row.era_variants === null, `expected era-independent fixture key ${key}`);
  assert.deepEqual(row.params.map((param) => param.name).sort(), Object.keys(params).sort());
  return row.text.replace(/\{([a-z_]+)\}/gu, (_, name) => String(params[name]));
}
function write(relative, data) {
  const destination = path.join(fixtureRoot, relative);
  mkdirSync(path.dirname(destination), { recursive: true }); writeFileSync(destination, data);
}
function fixture() {
  const grown = JSON.parse(readFileSync(path.join(root, "testdata/replay/garden-v1.json"), "utf8")).bundles.grown;
  const digest = createHash("sha256"), names = Object.keys(grown.artifacts).sort();
  for (const name of names) {
    const nameBytes = Buffer.from(name), data = Buffer.from(grown.artifacts[name]), frame = Buffer.alloc(8);
    frame.writeBigUInt64BE(BigInt(nameBytes.length)); digest.update(frame).update(nameBytes);
    frame.writeBigUInt64BE(BigInt(data.length)); digest.update(frame).update(data);
    write(`balance/composed/${name}.json`, data);
  }
  assert.equal(`sha256:${digest.digest("hex")}`, grown.constants_hash, "exact existing Garden bundle identity");
  write("balance/epochs/phase0.json", JSON.stringify({ schema_version: 1, current_epoch_id: 1,
    artifacts: names.map((name) => ({ name, path: `balance/composed/${name}.json` })),
    epochs: [{ epoch_id: 1, name: "Composed Garden fixture only", changelog_ref: "changelog/epoch-1.md", accepted_hashes: [grown.constants_hash] }],
  }));
  write("changelog/epoch-1.md", "# Garden test fixture; not a production mint\n");
  for (const relative of ["balance/api/phase0.json", "balance/transport/phase0.json", "moderation/guild-names.txt"]) {
    const destination = path.join(fixtureRoot, relative); mkdirSync(path.dirname(destination), { recursive: true }); copyFileSync(path.join(root, relative), destination);
  }
  return { hash: grown.constants_hash, garden: JSON.parse(grown.artifacts.server_garden) };
}
function sql(statement) {
  const result = spawnSync("docker", ["compose", "-f", "compose.game-ui-test.yml", "exec", "-T", "game-ui-postgres", "psql",
    "-v", "ON_ERROR_STOP=1", "-At", "-U", "cloud_clicker", "-d", "cloud_clicker_game_ui_test", "-c", statement], { cwd: root, encoding: "utf8" });
  assert.equal(result.status, 0, `declared test DB command: ${result.stderr}`); return result.stdout.trim();
}
function head(founderID) {
  assert.match(founderID, /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/u);
  return JSON.parse(sql(`WITH latest AS (
    SELECT DISTINCT ON (s.scope) s.scope,r.state,r.revision,r.version FROM save_streams s JOIN save_revisions r ON r.stream_id=s.id
    WHERE s.owner_kind='founder' AND s.owner_id='${founderID}' AND s.archived_at IS NULL ORDER BY s.scope,r.revision DESC
  ) SELECT jsonb_object_agg(scope,jsonb_build_object('state',state,'revision',revision,'version',version)) FROM latest;`));
}
function matureFixturePlants(founderID, planted) {
  // Explicit test state, not elapsed time or a production growth result. Keep
  // the real salt, anchor, unlock, Company and every other Founder field intact.
  const garden = planted.founder.state.server_garden;
  assert.deepEqual(garden.plots.map(({ row, col, species_id, age_ticks, matured_effect_ppm }) =>
    ({ row, col, species_id, age_ticks, matured_effect_ppm })), [
    { row: 0, col: 0, species_id: "strain_a", age_ticks: 0, matured_effect_ppm: null },
    { row: 0, col: 1, species_id: "strain_a", age_ticks: 0, matured_effect_ppm: null },
    { row: 1, col: 1, species_id: "strain_a", age_ticks: 0, matured_effect_ppm: null },
  ]);
  const plots = garden.plots.map((plot) => ({ ...plot, age_ticks: 3, matured_effect_ppm: 1_000_000 }));
  assert.equal(sql(`WITH current AS (
    SELECT r.* FROM save_revisions r JOIN save_streams s ON s.id=r.stream_id
    WHERE s.owner_kind='founder' AND s.owner_id='${founderID}' AND s.scope='founder' AND s.archived_at IS NULL
    ORDER BY r.revision DESC LIMIT 1
  ), seeded AS (INSERT INTO save_revisions(stream_id,revision,version,state,constants_hash)
    SELECT stream_id,revision+1,version,jsonb_set(state,'{server_garden,plots}','${JSON.stringify(plots)}'::jsonb,false),constants_hash
    FROM current WHERE revision=${planted.founder.revision} RETURNING revision)
    SELECT count(*) FROM seeded;`), "1", "append exactly one controlled mature-state revision");
  const expected = structuredClone(planted);
  expected.founder.revision++;
  expected.founder.state.server_garden.plots = plots;
  assert.deepEqual(head(founderID), expected, "fixture setup may change only plot maturity and its revision");
  console.log("Garden harvest fixture: three native-planted strains explicitly marked mature in disposable DB; no clock, catalog, cash or payout grant; not natural growth/replay proof");
}
async function free(port) {
  await new Promise((resolve, reject) => {
    const probe = createTCPServer(); probe.once("error", reject); probe.listen(port, "127.0.0.1", () => probe.close(resolve));
  });
}
async function ready() {
  const deadline = Date.now() + 60_000;
  while (Date.now() < deadline) {
    assert.equal(errors.length, 0, errors.map(String).join("\n")); assert.equal(server.exitCode, null, "gameserver exited before ready");
    try { if ((await fetch(`${serverURL}/readyz`)).ok) return; } catch { /* listener still starting */ }
    await pause(100);
  }
  throw new Error("Garden gameserver readiness objective not reached");
}
async function serve(observeAPI) {
  const dist = path.join(root, "client/dist");
  const listener = createServer((request, response) => {
    const pathname = new URL(request.url, uiURL).pathname;
    if (pathname.startsWith("/api/")) {
      const boundary = apiProof.observe(response, pathname, request.method, request.headers.referer);
      const upstream = httpRequest(new URL(request.url, serverURL), { method: request.method, headers: { ...request.headers, host: new URL(serverURL).host } },
        (received) => {
          boundary.status = received.statusCode;
          const chunks = [];
          received.on("data", (chunk) => chunks.push(chunk));
          received.on("end", () => {
            boundary.upstream_ended = true;
            try { observeAPI(pathname, received.statusCode, Buffer.concat(chunks), boundary); }
            catch (error) { errors.push(new Error(`Garden proxy observation ${pathname}: ${error.message}`, { cause: error })); }
          });
          response.writeHead(received.statusCode ?? 502, received.headers); pipeObservedAPIResponse(received, response);
        });
      upstream.on("error", (error) => { errors.push(error); response.writeHead(502); response.end(); }); request.pipe(upstream); return;
    }
    const target = path.resolve(dist, pathname === "/" ? "index.html" : pathname.slice(1));
    if (!target.startsWith(`${dist}${path.sep}`) || !(pathname === "/" || pathname === "/favicon.ico" || pathname.startsWith("/assets/"))) { response.writeHead(404); response.end(); return; }
    try {
      const body = readFileSync(target), type = target.endsWith(".js") ? "text/javascript" : target.endsWith(".css") ? "text/css" : target.endsWith(".html") ? "text/html" : "application/octet-stream";
      response.writeHead(200, { "content-type": type }); response.end(body);
    } catch { response.writeHead(404); response.end(); }
  });
  listener.on("upgrade", (request, socket, data) => {
    if (new URL(request.url, uiURL).pathname !== "/connection/websocket") { socket.destroy(); return; }
    const upstream = connect(18083, "127.0.0.1"); sockets.add(socket); sockets.add(upstream);
    const close = () => { socket.destroy(); upstream.destroy(); sockets.delete(socket); sockets.delete(upstream); };
    socket.on("error", close); upstream.on("error", close); socket.on("close", close); upstream.on("close", close);
    upstream.on("connect", () => {
      const headers = [];
      for (let i = 0; i < request.rawHeaders.length; i += 2) headers.push(`${request.rawHeaders[i]}: ${request.rawHeaders[i].toLowerCase() === "host" ? "127.0.0.1:18083" : request.rawHeaders[i + 1]}`);
      upstream.write(`${request.method} ${request.url} HTTP/1.1\r\n${headers.join("\r\n")}\r\n\r\n`); if (data.length) upstream.write(data); socket.pipe(upstream).pipe(socket);
    });
  });
  await new Promise((resolve, reject) => { listener.once("error", reject); listener.listen(5173, "127.0.0.1", resolve); }); return listener;
}
async function snapshot(page) {
  return page.evaluate(async () => {
    const credentials = JSON.parse(localStorage.getItem("cloud-clicker.credentials.v1"));
    const response = await fetch("/api/v1/founder/state", { headers: { Authorization: `Bearer ${credentials.accessToken}` } });
    if (response.status !== 200) throw new Error(`Garden observation snapshot: ${response.status}`); return response.json();
  });
}
function noSalt(data, salt) {
  const encoded = JSON.stringify(data); if (salt) assert(!encoded.includes(salt), "actual hidden Garden salt leaked");
  function visit(value) {
    if (value && typeof value === "object") for (const [key, child] of Object.entries(value)) {
      assert(!["salt_hex", "garden_salt_hex", "garden_base", "rng_s", "draw", "draws"].includes(key), `hidden Garden key ${key}`); visit(child);
    }
  }
  visit(data);
}

try {
  const bundle = fixture();
  // Shared ephemeral DB is serial: detect another composed driver before reset.
  for (const port of [18081, 18082, 18083, 5173, 5174]) await free(port);
  sql("SET client_min_messages TO WARNING; DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;");
  const key = Buffer.alloc(32, 7).toString("base64");
  server = spawn(path.join(root, ".cache/bin/gameserver"), [], { cwd: root, stdio: ["ignore", "pipe", "pipe"], env: { ...process.env,
    CLOUD_CLICKER_ACTIVITY_BRACKET: "activity.standard", CLOUD_CLICKER_BOOTSTRAP_KEY: key, CLOUD_CLICKER_BOOTSTRAP_KEY_ID: "garden-browser-fixture",
    CLOUD_CLICKER_CURSOR_KEY: key, CLOUD_CLICKER_JWT_KEY: key, CLOUD_CLICKER_REPOSITORY_ROOT: fixtureRoot,
    CLOUD_CLICKER_SERVER_ID: "01986666-b001-4000-8000-000000000003", LISTEN_ADDR: "127.0.0.1:18083",
    DATABASE_URL: "postgres://cloud_clicker:cloud_clicker_game_ui_test@127.0.0.1:55433/cloud_clicker_game_ui_test?sslmode=disable",
  } });
  server.on("error", (error) => errors.push(error)); server.stdout.on("data", (data) => process.stdout.write(data)); server.stderr.on("data", (data) => process.stderr.write(data));
  await ready();
  const clientProof = productionClientProof(productionClientFiles(path.join(root, "client/dist")));
  browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });
  const requests = [], views = [], publicData = [], responses = [], transport = { sent: 0, received: 0, closed: 0 };
  let expectExpiredGarden = false;
  page.on("pageerror", (error) => errors.push(error));
  page.on("requestfailed", (request) => errors.push(new Error(`request failed ${new URL(request.url()).pathname}: ${request.failure()?.errorText}`)));
  page.on("request", (request) => {
    const url = new URL(request.url()), origin = new URL(uiURL);
    if (url.host !== origin.host || !["http:", "ws:"].includes(url.protocol) || !(url.pathname === "/" || url.pathname === "/favicon.ico" || /^\/(?:api|assets)\//u.test(url.pathname))) errors.push(new Error(`request outside fixture origin: ${request.url()}`));
    requests.push(request);
  });
  page.context().on("response", (response) => {
    const url = new URL(response.url()), pathname = url.pathname;
    if (url.origin !== uiURL || pathname.startsWith("/api/") || pathname === "/favicon.ico") return;
    responses.push((response.status() === 304 ? Promise.resolve(undefined) : response.body())
      .then((body) => clientProof.response(pathname, response.status(), body))
      .catch((error) => errors.push(error)));
  });
  page.on("worker", (worker) => {
    try { clientProof.worker(new URL(worker.url()).pathname); }
    catch (error) { errors.push(error); }
  });
  let navigationSequence = 0;
  async function reload() {
    // Actual API bytes are observed in the proxy, not Chromium's navigation-
    // scoped response cache. Native command receipts are also read immediately
    // in domIntent, and the real consumer must render the resulting state.
    await Promise.all(responses);
    assert.equal(errors.length, 0, errors.map(String).join("\n"));
    // A distinct document URL lets the proxy bind the browser's real Referer:
    // an old-page request may arrive after navigation STARTS, but a cancelled
    // new-page read must never receive that old-page exception. No fetch shim.
    const next = new URL(uiURL);
    next.searchParams.set("garden_navigation", String(++navigationSequence));
    await apiProof.reload(page.url(), () => page.goto(next.href, { waitUntil: "networkidle" }));
  }
  page.on("websocket", (socket) => {
    assert.equal(socket.url(), `${uiURL.replace("http:", "ws:")}/connection/websocket`);
    socket.on("framesent", () => transport.sent++);
    socket.on("framereceived", ({ payload }) => { transport.received++; for (const line of String(payload).split("\n").filter(Boolean)) { try { publicData.push(JSON.parse(line)); } catch (error) { errors.push(error); } } });
    socket.on("socketerror", (error) => errors.push(new Error(String(error))));
    socket.on("close", () => transport.closed++);
  });
  const control = (key, params) => page.getByRole("button", { name: text(key, params), exact: true });
  const cell = (row, col) => page.locator(".garden button.cell").nth(row * 6 + col);
  const writes = () => requests.filter((request) => request.method() === "POST" && new URL(request.url()).pathname === "/api/v1/intents");
  async function domIntent(button, kind, fields, key = "Enter") {
    const before = await snapshot(page), count = writes().length;
    await button.click({ trial: true, timeout: 30_000 });
    const pending = page.waitForResponse((response) => new URL(response.url()).pathname === "/api/v1/intents" && response.request().postDataJSON()?.kind === kind, { timeout: 30_000 })
      .then((value) => ({ value }), (error) => ({ error }));
    if (harvestFixture) { await button.focus(); await page.keyboard.press(key); }
    else await button.click();
    const received = await pending;
    if (received.error) throw new Error(`Garden DOM ${kind}: no response; ${writes().slice(count).filter((request) => request.postDataJSON()?.kind === kind).length} matching intents emitted`, { cause: received.error });
    const response = received.value, receiptBytes = await response.text(), outcome = JSON.parse(receiptBytes);
    const emitted = writes().slice(count); assert.equal(emitted.length, 1, `one DOM ${kind} intent`);
    const { intent_id, ...body } = emitted[0].postDataJSON(); assert.match(intent_id, /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u);
    assert.deepEqual(body, { kind, expected_revision: before.founder_revision, ...fields });
    assert.equal(response.status(), 200); assert.equal(outcome.outcome, "applied", `applied ${kind}: ${JSON.stringify(outcome)}`);
    assert.equal(outcome.intent_id, intent_id, "receipt belongs to the native activation");
    assert.equal(outcome.founder_revision, before.founder_revision + 1);
    await page.waitForFunction(() => document.querySelector("main.game-ui")?.getAttribute("aria-busy") === "false", undefined, { timeout: 30_000 });
    console.log(`Garden DOM ${kind}: applied Founder revision ${outcome.founder_revision}`); return { outcome, intent_id, body: emitted[0].postDataJSON(), receiptBytes };
  }
  assets = await serve((pathname, status, body, boundary) => {
    apiBoundaries.push({ path: pathname, status });
    if (status === 204) return;
    const data = JSON.parse(body.toString("utf8")); publicData.push(data);
    if (pathname === "/api/v1/garden/current") {
      if (expectExpiredGarden && status === 401) assert.deepEqual(data, { category: "unauthorized", detail: "access_token" });
      else {
        assert.equal(status, 200, `Garden GET returned ${status}`);
        views.push({ data, at: Date.now(), boundary });
      }
    }
  });
  await page.goto(uiURL, { waitUntil: "networkidle" }); await page.getByRole("button", { name: "BEGIN ATTEMPT", exact: true }).click();
  await page.locator('main[data-surface="desk"]').waitFor({ state: "visible", timeout: 30_000 });
  try {
    await page.getByText(/You are visitor #\d+/u).waitFor({ state: "visible", timeout: 30_000 });
  } catch (error) {
    throw new Error(`Garden live presence failed: ${JSON.stringify({ transport, errors: errors.map(String), main: (await page.locator("main").innerText()).slice(0, 500) })}`, { cause: error });
  }
  assert(transport.sent > 0 && transport.received > 0, "real subscribed WebSocket");
  const initial = await snapshot(page), founderID = initial.run.founder_id, initialHead = head(founderID);
  assert.equal(initial.constants_hash, bundle.hash); assert.equal(initialHead.founder.version, 25);
  assert.equal(initialHead.founder.state.fiscal_credit, 0); assert.deepEqual(initialHead.founder.state.server_garden.plots, []);
  assert.deepEqual(initialHead.founder.state.server_garden.seed_collection, ["strain_a", "strain_b"]);
  assert.equal(initialHead.company.state.balances["company.cash"], "0");
  assert(Object.values(initialHead.company.state.generators).every((count) => count === 0), "no passive generator credit in cash oracle");
  // Compile every end-of-journey SQL column now, before the real 15-minute wait.
  sql("SELECT quota_used FROM minigame_faucet_window LIMIT 0; SELECT revision,payload,intent_id FROM events LIMIT 0; SELECT canonical_payload,replay_inputs,receipt,applied_revision,source_company_stream_id,source_run_seq,source_run_log_seq FROM founder_log LIMIT 0; SELECT company_stream_id,run_seq,seq,canonical_payload,replay_inputs,receipt,applied_revision FROM run_log LIMIT 0;");
  await control("garden.title").click(); await page.getByText(text("garden.state.locked"), { exact: true }).waitFor({ timeout: 30_000 });
  await page.locator('nav button[data-nav-surface="fiscal"]').click(); await page.locator('main[data-surface="fiscal"]').waitFor({ timeout: 30_000 });
  const available = await snapshot(page);
  assert(available.features.fiscal.unlocks.some((row) => row.unlock_id === bundle.garden.unlock_id && row.cost === 3 && !row.owned), "actual producer offers Garden unlock");
  const unlockRow = page.locator(".fiscal li").filter({ has: page.getByRole("heading", { name: text("garden.title"), exact: true }) });
  assert.equal(await unlockRow.count(), 1, "Garden unlock producer has no Fiscal DOM consumer");
  // The initial displayed preview can precede the first completed period.
  // Collect it by the real player control, never force-refresh or grant credit.
  await page.locator('.fiscal[data-phase="guaranteed"]').waitFor({ timeout: 30_000 });
  await domIntent(page.locator(".fiscal").getByRole("button", { name: text("fiscal.harvest"), exact: true }), "harvest_fiscal_period", {});
  await domIntent(unlockRow.getByRole("button"), "spend_fiscal_credit", { target: { kind: "unlock", unlock_id: bundle.garden.unlock_id } });
  await control("garden.title").click(); await cell(0, 0).waitFor({ timeout: 30_000 });
  if (sessionDiagnostic) {
    // Explicit controlled access-row expiry, NOT a production clock/TTL change,
    // natural JWT expiry, or an automatically renewing runtime. No token output.
    const before = head(founderID), beforeWrites = writes().length;
    const tokenID = await page.evaluate(() => {
      const value = JSON.parse(localStorage.getItem("cloud-clicker.credentials.v1"));
      return JSON.parse(atob(value.accessToken.split(".")[1].replace(/-/gu, "+").replace(/_/gu, "/"))).jti;
    });
    assert.match(tokenID, /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u);
    const familyCounts = () => JSON.parse(sql(`SELECT jsonb_build_object(
      'sessions',(SELECT count(*) FROM sessions WHERE family_id=issued.family_id),
      'consumed',(SELECT count(*) FROM sessions WHERE family_id=issued.family_id AND consumed_at IS NOT NULL),
      'revoked',(SELECT count(*) FROM session_families WHERE family_id=issued.family_id AND revoked_at IS NOT NULL),
      'access',(SELECT count(*) FROM access_tokens WHERE family_id=issued.family_id))
      FROM access_tokens issued WHERE issued.jti='${tokenID}';`));
    assert.deepEqual(familyCounts(), { sessions: 1, consumed: 0, revoked: 0, access: 1 });
    await control("surface.desk.title").click();
    expectExpiredGarden = true;
    assert.equal(sql(`WITH expired AS (UPDATE access_tokens SET expires_at=clock_timestamp()-interval '1 millisecond'
      WHERE jti='${tokenID}' AND founder_id='${founderID}' AND revoked_at IS NULL RETURNING jti)
      SELECT count(*) FROM expired;`), "1", "expire only exact diagnostic access row");
    const expired = page.waitForResponse((response) => new URL(response.url()).pathname === "/api/v1/garden/current", { timeout: 30_000 });
    await control("garden.title").click();
    const expiredResponse = await expired;
    assert.equal(expiredResponse.status(), 401, "actual runtime Garden read must observe controlled expiry");
    assert.deepEqual(await expiredResponse.json(), { category: "unauthorized", detail: "access_token" });
    const socketDeadline = Date.now() + 35_000;
    while (transport.closed === 0 && Date.now() < socketDeadline) await pause(100);
    assert.equal(transport.closed, 1, "original subscribed socket did not close in the existing alive window");
    assert.equal(requests.filter((request) => new URL(request.url()).pathname === "/api/v1/session/refresh").length, 0, "runtime unexpectedly supplied a renewal consumer");
    assert.deepEqual(familyCounts(), { sessions: 1, consumed: 0, revoked: 0, access: 1 });
    assert.deepEqual(head(founderID), before, "expiry changed gameplay heads");
    console.log("Session diagnostic: actual runtime Garden GET 401; subscribed socket closed; stored refresh unconsumed; no automatic renewal: CONFIRMED GAP");
    // Positive control operated by the test, explicitly not the client workflow.
    const rotation = await page.evaluate(async () => {
      const key = "cloud-clicker.credentials.v1", previous = JSON.parse(localStorage.getItem(key));
      const response = await fetch("/api/v1/session/refresh", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ refresh_token: previous.refreshToken }) });
      const pair = await response.json(), keys = Object.keys(pair).sort();
      if (response.status !== 200 || keys.join("\0") !== "access_token\0refresh_token" ||
          typeof pair.access_token !== "string" || typeof pair.refresh_token !== "string" ||
          pair.access_token === previous.accessToken || pair.refresh_token === previous.refreshToken) {
        return { status: response.status, validPair: false };
      }
      localStorage.setItem(key, JSON.stringify({ ...previous, accessToken: pair.access_token, refreshToken: pair.refresh_token }));
      return { status: response.status, validPair: true };
    });
    assert.deepEqual(rotation, { status: 200, validPair: true });
    assert.deepEqual(familyCounts(), { sessions: 2, consumed: 1, revoked: 0, access: 2 });
    expectExpiredGarden = false;
    await reload();
    await page.getByText(/You are visitor #\d+/u).waitFor({ state: "visible", timeout: 30_000 });
    await control("garden.title").click(); await cell(0, 0).waitFor({ timeout: 30_000 });
    assert.equal((await snapshot(page)).run.founder_id, founderID, "manual rotation replaced the Founder");
    assert.deepEqual(head(founderID), before, "rotation/reload changed gameplay heads");
    assert.equal(writes().length, beforeWrites, "diagnostic/reload emitted gameplay intents");
    assert.equal(requests.filter((request) => new URL(request.url()).pathname === "/api/v1/bootstrap").length, 1, "diagnostic/reload created a replacement account");
    assert.equal(requests.filter((request) => new URL(request.url()).pathname === "/api/v1/session/refresh").length, 1, "more than the test's one rotation was submitted");
    await Promise.all(responses); assert.equal(errors.length, 0, errors.map(String).join("\n"));
    console.log(`Session diagnostic: one test-operated HTTP rotation; same Founder/Garden restored by actual reload; gameplay heads unchanged; ${(Date.now() - started) / 1000}s: OBSERVATION COMPLETE, NOT AUTOMATIC-RENEWAL ACCEPTANCE`);
  } else {
  async function plant(row, col, species) {
    try {
      assert.equal(await cell(row, col).getAttribute("data-stage"), "empty", `plant target ${row},${col}`);
      await cell(row, col).click(); await domIntent(control("garden.action.plant_frame", { species: text(`garden.species.${species}.name`) }), "garden_plant", { row, col, species_id: species });
    } catch (error) {
      throw new Error(`Garden plant ${row},${col} failed: ${JSON.stringify({ view: views.at(-1)?.data, garden: await page.locator(".garden").innerText(), errors: errors.map(String) })}`, { cause: error });
    }
    await page.waitForFunction(({ index }) => document.querySelectorAll(".garden button.cell")[index]?.getAttribute("data-stage") === "growing", { index: row * 6 + col }, { timeout: 30_000 });
  }
  await plant(0, 0, "strain_a"); await plant(0, 1, "strain_a"); await plant(1, 0, "strain_b");
  await cell(1, 0).click(); await domIntent(control("garden.action.uproot"), "garden_uproot", { row: 1, col: 0 });
  if (harvestFixture) await plant(1, 1, "strain_a");
  const planted = head(founderID), salt = planted.founder.state.server_garden.salt_hex, anchor = planted.founder.state.server_garden.tick_anchor_wall_ms;
  assert.match(salt, /^[0-9a-f]{16}$/u); assert.equal(planted.founder.state.server_garden.tick_seq, 0);
  const substrate = bundle.garden.substrates.find((row) => row.substrate_id === bundle.garden.default_substrate_id);
  const maturity = bundle.garden.species.find((row) => row.species_id === "strain_a").maturation_ticks;
  const due = anchor + substrate.tick_ms * maturity, waitingWrites = writes().length, beforeWaitReads = views.filter(({ boundary }) => boundary.browser_finished).length;
  if (harvestFixture) {
    assert.equal(substrate.tick_ms, 300_000); assert.equal(maturity, 3, "fixture does not accelerate growth");
    matureFixturePlants(founderID, planted);
    await reload(); await control("garden.title").click();
    await page.waitForFunction(() => [0, 1, 7].every((index) =>
      document.querySelectorAll(".garden button.cell")[index]?.getAttribute("data-stage") === "mature"), undefined, { timeout: 30_000 });
    assert.equal(writes().length, waitingWrites, "loading controlled maturity must not submit gameplay");
  } else {
  console.log(`Garden native wall-time objective: ${substrate.tick_ms}ms × ${maturity}; due ${new Date(due).toISOString()}`);
  heartbeat = setInterval(() => console.log(`Garden wait ${Math.round((Date.now() - anchor) / 1000)}s; reads ${views.length}; last observed tick ${views.at(-1)?.data.garden?.tick_seq}`), 30_000);
  await page.waitForFunction(() => {
    const cells = [...document.querySelectorAll(".garden button.cell")].slice(0, 2);
    return cells.length === 2 && cells.every((node) => node.getAttribute("data-stage") === "mature");
  }, undefined, { timeout: Math.max(1, due - Date.now()) + 30_000 });
  clearInterval(heartbeat); heartbeat = undefined;
  assert(Date.now() >= due, "no premature simulated maturity"); assert.equal(writes().length, waitingWrites, "no gameplay write supplied growth");
  assert(views.filter(({ boundary }) => boundary.browser_finished).length >= beforeWaitReads + maturity, "native due timers did not re-read every tick");
  for (const tick of [1, 2, 3]) assert(views.some(({ data, boundary }) => {
    const plots = data.garden?.plots.filter((plot) => plot.row === 0 && plot.col < 2) ?? [];
    return boundary.browser_finished && data.garden?.tick_seq === tick && plots.length === 2 && plots.every((plot) => plot.age_ticks === tick);
  }), `missing actual advisory tick ${tick}`);
  // Advisory reads must not persist a simulated advance.
  assert.deepEqual(head(founderID).founder, planted.founder, "timed Garden GET persisted state");
  }
  await cell(0, 0).click(); const single = await domIntent(control("garden.action.harvest"), "garden_harvest", { plots: [{ row: 0, col: 0 }] });
  await page.waitForFunction(() => document.querySelector(".garden button.cell")?.getAttribute("data-stage") === "empty", undefined, { timeout: 30_000 });
  const remainingPlots = harvestFixture ? [{ row: 0, col: 1 }, { row: 1, col: 1 }] : [{ row: 0, col: 1 }];
  const all = await domIntent(control("garden.action.harvest_all"), "garden_harvest", { plots: remainingPlots }, "Space");
  await domIntent(control("garden.substrate.containerized.name"), "garden_set_substrate", { substrate_id: "containerized" });
  await Promise.all(responses);
  const final = head(founderID);
  assert.equal(final.founder.state.server_garden.substrate_id, "containerized");
  assert(final.founder.state.server_garden.plots.every((plot) => plot.row !== 0 || plot.col > 1));
  assert.equal(final.company.state.balances["company.cash"], harvestFixture ? "1.5e1" : "1e1", "only actual harvest cash credits persisted");
  assert.equal(final.company.revision, initialHead.company.revision + 2, "only two harvests advanced Company");
  const proof = JSON.parse(sql(`SELECT jsonb_build_object(
    'sends',(SELECT COALESCE(sum(quota_used),0) FROM minigame_faucet_window WHERE founder_id='${founderID}' AND minigame_id='server_garden'),
    'company_events',(SELECT jsonb_agg(jsonb_build_object('payload',payload,'intent_id',intent_id,'revision',revision) ORDER BY revision) FROM events WHERE kind='garden_harvest_credited.v1' AND stream_id IN(SELECT id FROM save_streams WHERE owner_id='${founderID}')),
    'founder_events',(SELECT jsonb_agg(jsonb_build_object('payload',payload,'intent_id',intent_id,'revision',revision) ORDER BY revision) FROM events WHERE kind='garden_harvested.v1' AND stream_id IN(SELECT id FROM save_streams WHERE owner_id='${founderID}')),
    'founder_logs',(SELECT jsonb_agg(jsonb_build_object('intent_id',intent_id,'payload',convert_from(canonical_payload,'UTF8')::jsonb,'inputs',replay_inputs,'receipt',receipt,'revision',applied_revision,'company_stream',source_company_stream_id,'run_seq',source_run_seq,'run_log_seq',source_run_log_seq) ORDER BY seq) FROM founder_log WHERE founder_stream_id IN(SELECT id FROM save_streams WHERE owner_id='${founderID}') AND replay_inputs->'resolved'->>'kind'='garden_harvest_credited'),
    'company_logs',(SELECT jsonb_agg(jsonb_build_object('payload',convert_from(canonical_payload,'UTF8')::jsonb,'inputs',replay_inputs,'receipt',receipt,'revision',applied_revision,'company_stream',company_stream_id,'run_seq',run_seq,'seq',seq) ORDER BY seq) FROM run_log WHERE company_stream_id IN(SELECT id FROM save_streams WHERE owner_id='${founderID}') AND convert_from(canonical_payload,'UTF8')::jsonb->>'kind'='credit_garden_harvest'));`));
  assert.equal(proof.sends, 2); assert.equal(proof.company_events.length, 2); assert.equal(proof.founder_events.length, 2);
  assert.equal(proof.founder_logs.length, 2); assert.equal(proof.company_logs.length, 2);
  for (const [i, harvested] of [single, all].entries()) {
    const { outcome, intent_id } = harvested, companyEvent = proof.company_events[i], founderEvent = proof.founder_events[i];
    const founderLog = proof.founder_logs[i], companyLog = proof.company_logs[i], hash = outcome.harvest.harvest_hash;
    assert.match(hash, /^sha256:[0-9a-f]{64}$/u);
    const targets = i === 0 ? [{ row: 0, col: 0 }] : remainingPlots;
    const expectedPlots = targets.map(({ row, col }) => ({ col, row, species_id: "strain_a", units: 5 }));
    const units = 5 * targets.length, credited = units === 10 ? "1e1" : "5e0";
    const expectedHash = `sha256:${createHash("sha256").update(JSON.stringify({ intent_id, plots: expectedPlots, total_units: units })).digest("hex")}`;
    assert.deepEqual(outcome.harvest, { plots: expectedPlots, total_units: units, seeds_discovered: [], harvest_hash: expectedHash });
    assert.equal(outcome.credited, credited); assert.equal(outcome.credited_resource_id, "company.cash"); assert.equal(outcome.faucet_applied, true);
    assert.equal(outcome.forfeited_units, 0); assert.equal(outcome.cap_reason_key, null);
    assert.equal(companyEvent.intent_id, intent_id); assert.equal(founderEvent.intent_id, intent_id);
    assert.equal(companyEvent.payload.credited, credited); assert.equal(companyEvent.payload.faucet_applied, true);
    assert.equal(companyEvent.payload.harvest_hash, hash); assert.deepEqual(founderEvent.payload, outcome.harvest);
    assert.equal(companyEvent.revision, outcome.company_revision); assert.equal(founderEvent.revision, outcome.founder_revision);
    assert.equal(founderLog.intent_id, intent_id); assert.equal(companyLog.payload.intent_id, intent_id);
    assert.deepEqual(founderLog.payload, { kind: "garden_harvest", plots: targets, expected_revision: harvested.body.expected_revision });
    assert.deepEqual(founderLog.receipt.harvest, outcome.harvest); assert.equal(companyLog.payload.harvest_hash, hash);
    assert.equal(companyLog.inputs.resolved.harvest_hash, hash); assert.equal(companyLog.receipt.harvest_hash, hash);
    assert.equal(companyLog.receipt.credited, credited); assert.equal(companyLog.receipt.faucet_applied, true);
    assert.equal(founderLog.revision, outcome.founder_revision); assert.equal(companyLog.revision, outcome.company_revision);
    assert.equal(founderLog.company_stream, companyLog.company_stream); assert.equal(founderLog.run_seq, companyLog.run_seq); assert.equal(founderLog.run_log_seq, companyLog.seq);
  }
  assert.notEqual(single.intent_id, all.intent_id);
  if (harvestFixture) {
    // Diagnostic HTTP retries of the native command, not extra gameplay or
    // substitutes for the native happy path. Neither may credit cash twice.
    const beforeRetry = head(founderID);
    const sends = () => sql(`SELECT row_to_json(w) FROM minigame_faucet_window w WHERE founder_id='${founderID}' AND minigame_id='server_garden';`);
    const rowCounts = () => sql(`SELECT jsonb_build_object(
      'events',(SELECT count(*) FROM events WHERE stream_id IN(SELECT id FROM save_streams WHERE owner_id='${founderID}')),
      'intents',(SELECT count(*) FROM intent_records WHERE stream_id IN(SELECT id FROM save_streams WHERE owner_id='${founderID}')),
      'founder_logs',(SELECT count(*) FROM founder_log WHERE founder_stream_id IN(SELECT id FROM save_streams WHERE owner_id='${founderID}')),
      'company_logs',(SELECT count(*) FROM run_log WHERE company_stream_id IN(SELECT id FROM save_streams WHERE owner_id='${founderID}')));`);
    const beforeWindow = sends();
    const beforeRows = rowCounts();
    const retry = async (body) => page.evaluate(async (payload) => {
      const credentials = JSON.parse(localStorage.getItem("cloud-clicker.credentials.v1"));
      const response = await fetch("/api/v1/intents", { method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${credentials.accessToken}` }, body: JSON.stringify(payload) });
      return { status: response.status, bytes: await response.text() };
    }, body);
    const repeated = await retry(all.body);
    assert.equal(repeated.status, 200); assert.equal(repeated.bytes, all.receiptBytes, "retry returns exact stored receipt bytes");
    const conflict = await retry({ ...all.body, plots: [{ row: 0, col: 0 }] });
    const refused = JSON.parse(conflict.bytes);
    assert.equal(conflict.status, 200); assert.equal(refused.outcome, "rejected");
    assert.equal(refused.intent_id, all.intent_id);
    assert.deepEqual(refused.rejection, { category: "idempotency_conflict", detail: all.intent_id }, "changed-body retry must refuse");
    assert.deepEqual(head(founderID), beforeRetry, "retry/refusal may not change either saved head");
    assert.equal(sends(), beforeWindow, "retry/refusal may not consume or modify the faucet window");
    assert.equal(rowCounts(), beforeRows, "retry/refusal may not append events, logs or intent records");
    console.log("Garden actual HTTP retry: identical receipt; changed-body refusal; saved heads, faucet and event/log/intent counts unchanged: PASS");
  }
  publicData.forEach((data) => noSalt(data, salt));
  await reload(); await control("garden.title").click();
  await page.waitForFunction((name) => [...document.querySelectorAll(".garden button")].some((node) => node.textContent.trim() === name && node.getAttribute("aria-pressed") === "true"), text("garden.substrate.containerized.name"), { timeout: 30_000 });
  assert.deepEqual(head(founderID).founder, final.founder, "reload changed persisted Garden");
  if (harvestFixture) {
    assert.equal(await cell(0, 0).getAttribute("data-stage"), "empty");
    assert.equal(await cell(0, 1).getAttribute("data-stage"), "empty");
    assert.equal(await cell(1, 1).getAttribute("data-stage"), "empty");
    assert.equal(await control("garden.action.harvest_all").count(), 0, "reload must not offer harvested plants again");
    const restored = await snapshot(page), cashIndex = restored.resources.findIndex((row) => row.resource_id === "company.cash");
    assert(cashIndex >= 0); assert.equal(restored.resources[cashIndex].amount, "1.5e1", "reload reads actual credited cash");
    await control("surface.desk.title").click();
    await page.waitForFunction((index) => document.querySelectorAll('section[aria-labelledby="resources-heading"] output')[index]?.textContent === "15", cashIndex, { timeout: 30_000 });
    console.log("Garden reload: harvested plots empty, substrate retained, stored/server/rendered cash 15: PASS");
  }
  await Promise.all(responses); publicData.forEach((data) => noSalt(data, salt)); assert.equal(errors.length, 0, errors.map(String).join("\n"));
  console.log(`Garden loaded exact built HTML/JS/CSS and bundled prediction Worker: ${JSON.stringify(clientProof.finish())}: PASS`);
  console.log(`Garden composed ${harvestFixture ? "controlled-maturity fixture" : "real wall-clock"}: DOM bootstrap/unlock/plant/uproot → ${harvestFixture ? "explicit mature-state setup" : "native three-tick maturation"} → native single/all harvest → substrate/reload; two real cash sends, bound hashes, hidden salt; ${(Date.now() - started) / 1000}s: PASS`);
  }
  console.log(`Garden API observation: ${JSON.stringify(apiProof.finish())}; cancelled old-page reads are not consumed-read proof; all complete JSON still enumerated`);
} catch (error) {
  throw new Error(`Garden composed objective failed; boundary errors: ${JSON.stringify(errors.map(String))}; HTTP statuses: ${JSON.stringify(apiBoundaries)}; recent proxy boundaries: ${JSON.stringify(proxyBoundaries.slice(-12))}`, { cause: error });
} finally {
  if (heartbeat) clearInterval(heartbeat); await browser?.close(); for (const socket of sockets) socket.destroy();
  if (assets) await new Promise((resolve, reject) => assets.close((error) => error ? reject(error) : resolve()));
  if (server && server.exitCode === null && server.signalCode === null) {
    const exited = new Promise((resolve) => server.once("exit", resolve)); server.kill("SIGTERM");
    const guard = setTimeout(() => { if (server.exitCode === null && server.signalCode === null) server.kill("SIGKILL"); }, 10_000);
    await exited; clearTimeout(guard);
  }
  rmSync(fixtureRoot, { recursive: true, force: true });
}
