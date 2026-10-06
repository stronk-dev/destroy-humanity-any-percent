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

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");
const uiURL = "http://localhost:5173";
const serverURL = "http://127.0.0.1:18083";
const fixtureRoot = mkdtempSync(path.join(os.tmpdir(), "cloud-clicker-garden-composed-"));
const started = Date.now();
const sockets = new Set();
const errors = [];
let server, assets, browser, heartbeat;
const pause = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
const copy = JSON.parse(readFileSync(path.join(root, "client/src/copy/generated/catalog.json"), "utf8"));
function text(key, params = {}) {
  const row = copy.entries.find((entry) => entry.key === key);
  assert(row && row.era_variants === null, `expected era-independent fixture key ${key}`);
  assert.deepEqual([...row.params].sort(), Object.keys(params).sort());
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
async function serve() {
  const dist = path.join(root, "client/dist");
  const listener = createServer((request, response) => {
    const pathname = new URL(request.url, uiURL).pathname;
    if (pathname.startsWith("/api/")) {
      const upstream = httpRequest(new URL(request.url, serverURL), { method: request.method, headers: { ...request.headers, host: new URL(serverURL).host } },
        (received) => { response.writeHead(received.statusCode ?? 502, received.headers); received.pipe(response); });
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
  await ready(); assets = await serve(); browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });
  const requests = [], views = [], publicData = [], responses = [], transport = { sent: 0, received: 0 };
  page.on("pageerror", (error) => errors.push(error));
  page.on("requestfailed", (request) => errors.push(new Error(`request failed ${new URL(request.url()).pathname}: ${request.failure()?.errorText}`)));
  page.on("request", (request) => {
    const url = new URL(request.url()), origin = new URL(uiURL);
    if (url.host !== origin.host || !["http:", "ws:"].includes(url.protocol) || !(url.pathname === "/" || url.pathname === "/favicon.ico" || /^\/(?:api|assets)\//u.test(url.pathname))) errors.push(new Error(`request outside fixture origin: ${request.url()}`));
    requests.push(request);
  });
  page.on("response", (response) => {
    if (!new URL(response.url()).pathname.startsWith("/api/")) return;
    responses.push((async () => {
      if (response.status() === 204) return;
      const data = await response.json(); publicData.push(data);
      if (new URL(response.url()).pathname === "/api/v1/garden/current") { assert.equal(response.status(), 200); views.push({ data, at: Date.now() }); }
    })().catch((error) => errors.push(error)));
  });
  page.on("websocket", (socket) => {
    assert.equal(socket.url(), `${uiURL.replace("http:", "ws:")}/connection/websocket`);
    socket.on("framesent", () => transport.sent++);
    socket.on("framereceived", ({ payload }) => { transport.received++; for (const line of String(payload).split("\n").filter(Boolean)) { try { publicData.push(JSON.parse(line)); } catch (error) { errors.push(error); } } });
    socket.on("socketerror", (error) => errors.push(new Error(String(error))));
  });
  const control = (key, params) => page.getByRole("button", { name: text(key, params), exact: true });
  const cell = (row, col) => page.locator(".garden button.cell").nth(row * 6 + col);
  const writes = () => requests.filter((request) => request.method() === "POST" && new URL(request.url()).pathname === "/api/v1/intents");
  async function domIntent(button, kind, fields) {
    const before = await snapshot(page), count = writes().length;
    await button.click({ trial: true, timeout: 30_000 });
    const pending = page.waitForResponse((response) => new URL(response.url()).pathname === "/api/v1/intents" && response.request().postDataJSON()?.kind === kind, { timeout: 30_000 });
    await button.click(); const response = await pending, outcome = await response.json();
    const emitted = writes().slice(count); assert.equal(emitted.length, 1, `one DOM ${kind} intent`);
    const { intent_id, ...body } = emitted[0].postDataJSON(); assert.match(intent_id, /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u);
    assert.deepEqual(body, { kind, expected_revision: before.founder_revision, ...fields });
    assert.equal(response.status(), 200); assert.equal(outcome.outcome, "applied", `applied ${kind}: ${JSON.stringify(outcome)}`);
    assert.equal(outcome.founder_revision, before.founder_revision + 1);
    console.log(`Garden DOM ${kind}: applied Founder revision ${outcome.founder_revision}`); return { outcome, intent_id };
  }
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
  await control("surface.fiscal.title").click(); await page.locator('main[data-surface="fiscal"]').waitFor({ timeout: 30_000 });
  const available = await snapshot(page);
  assert(available.features.fiscal.unlocks.some((row) => row.unlock_id === bundle.garden.unlock_id && row.cost === 3 && !row.owned), "actual producer offers Garden unlock");
  const unlockRow = page.locator(".fiscal li").filter({ has: page.getByRole("heading", { name: text("garden.title"), exact: true }) });
  assert.equal(await unlockRow.count(), 1, "Garden unlock producer has no Fiscal DOM consumer");
  await domIntent(unlockRow.getByRole("button"), "spend_fiscal_credit", { target: { kind: "unlock", unlock_id: bundle.garden.unlock_id } });
  await control("garden.title").click(); await cell(0, 0).waitFor({ timeout: 30_000 });
  async function plant(row, col, species) {
    await cell(row, col).click(); await domIntent(control("garden.action.plant_frame", { species: text(`garden.species.${species}.name`) }), "garden_plant", { row, col, species_id: species });
    await page.waitForFunction(({ index }) => document.querySelectorAll(".garden button.cell")[index]?.getAttribute("data-stage") === "growing", { index: row * 6 + col }, { timeout: 30_000 });
  }
  await plant(0, 0, "strain_a"); await plant(0, 1, "strain_a"); await plant(1, 0, "strain_b");
  await cell(1, 0).click(); await domIntent(control("garden.action.uproot"), "garden_uproot", { row: 1, col: 0 });
  const planted = head(founderID), salt = planted.founder.state.server_garden.salt_hex, anchor = planted.founder.state.server_garden.tick_anchor_wall_ms;
  assert.match(salt, /^[0-9a-f]{16}$/u); assert.equal(planted.founder.state.server_garden.tick_seq, 0);
  const substrate = bundle.garden.substrates.find((row) => row.substrate_id === bundle.garden.default_substrate_id);
  const maturity = bundle.garden.species.find((row) => row.species_id === "strain_a").maturation_ticks;
  const due = anchor + substrate.tick_ms * maturity, waitingWrites = writes().length, beforeWaitReads = views.length;
  console.log(`Garden native wall-time objective: ${substrate.tick_ms}ms × ${maturity}; due ${new Date(due).toISOString()}`);
  heartbeat = setInterval(() => console.log(`Garden wait ${Math.round((Date.now() - anchor) / 1000)}s; reads ${views.length}; last observed tick ${views.at(-1)?.data.garden?.tick_seq}`), 30_000);
  await page.waitForFunction(() => {
    const cells = [...document.querySelectorAll(".garden button.cell")].slice(0, 2);
    return cells.length === 2 && cells.every((node) => node.getAttribute("data-stage") === "mature");
  }, undefined, { timeout: Math.max(1, due - Date.now()) + 30_000 });
  clearInterval(heartbeat); heartbeat = undefined;
  assert(Date.now() >= due, "no premature simulated maturity"); assert.equal(writes().length, waitingWrites, "no gameplay write supplied growth");
  assert(views.length >= beforeWaitReads + maturity, "native due timers did not re-read every tick");
  for (const tick of [1, 2, 3]) assert(views.some(({ data }) => {
    const plots = data.garden?.plots.filter((plot) => plot.row === 0 && plot.col < 2) ?? [];
    return data.garden?.tick_seq === tick && plots.length === 2 && plots.every((plot) => plot.age_ticks === tick);
  }), `missing actual advisory tick ${tick}`);
  // Advisory reads must not persist a simulated advance.
  assert.deepEqual(head(founderID).founder, planted.founder, "timed Garden GET persisted state");
  await cell(0, 0).click(); const single = await domIntent(control("garden.action.harvest"), "garden_harvest", { plots: [{ row: 0, col: 0 }] });
  await page.waitForFunction(() => document.querySelector(".garden button.cell")?.getAttribute("data-stage") === "empty", undefined, { timeout: 30_000 });
  const all = await domIntent(control("garden.action.harvest_all"), "garden_harvest", { plots: [{ row: 0, col: 1 }] });
  await domIntent(control("garden.substrate.containerized.name"), "garden_set_substrate", { substrate_id: "containerized" });
  await Promise.all(responses);
  const final = head(founderID);
  assert.equal(final.founder.state.server_garden.substrate_id, "containerized");
  assert(final.founder.state.server_garden.plots.every((plot) => plot.row !== 0 || plot.col > 1));
  assert.equal(final.company.state.balances["company.cash"], "1e1", "two actual five-unit cash credits persisted");
  assert.equal(final.company.revision, initialHead.company.revision + 2, "only two harvests advanced Company");
  const proof = JSON.parse(sql(`SELECT jsonb_build_object(
    'sends',(SELECT COALESCE(sum(quota_used),0) FROM minigame_faucet_window WHERE founder_id='${founderID}' AND minigame_id='server_garden'),
    'company_events',(SELECT jsonb_agg(jsonb_build_object('payload',payload,'intent_id',intent_id,'revision',revision) ORDER BY revision) FROM events WHERE kind='garden_harvest_credited.v1' AND stream_id IN(SELECT id FROM save_streams WHERE owner_id='${founderID}')),
    'founder_events',(SELECT jsonb_agg(jsonb_build_object('payload',payload,'intent_id',intent_id,'revision',revision) ORDER BY revision) FROM events WHERE kind='garden_harvested.v1' AND stream_id IN(SELECT id FROM save_streams WHERE owner_id='${founderID}')),
    'founder_logs',(SELECT jsonb_agg(jsonb_build_object('payload',convert_from(canonical_payload,'UTF8')::jsonb,'inputs',replay_inputs,'receipt',receipt,'revision',applied_revision,'company_stream',source_company_stream_id,'run_seq',source_run_seq,'run_log_seq',source_run_log_seq) ORDER BY seq) FROM founder_log WHERE founder_stream_id IN(SELECT id FROM save_streams WHERE owner_id='${founderID}') AND replay_inputs->'resolved'->>'kind'='garden_harvest_credited'),
    'company_logs',(SELECT jsonb_agg(jsonb_build_object('payload',convert_from(canonical_payload,'UTF8')::jsonb,'inputs',replay_inputs,'receipt',receipt,'revision',applied_revision,'company_stream',company_stream_id,'run_seq',run_seq,'seq',seq) ORDER BY seq) FROM run_log WHERE company_stream_id IN(SELECT id FROM save_streams WHERE owner_id='${founderID}') AND convert_from(canonical_payload,'UTF8')::jsonb->>'kind'='credit_garden_harvest'));`));
  assert.equal(proof.sends, 2); assert.equal(proof.company_events.length, 2); assert.equal(proof.founder_events.length, 2);
  assert.equal(proof.founder_logs.length, 2); assert.equal(proof.company_logs.length, 2);
  for (const [i, harvested] of [single, all].entries()) {
    const { outcome, intent_id } = harvested, companyEvent = proof.company_events[i], founderEvent = proof.founder_events[i];
    const founderLog = proof.founder_logs[i], companyLog = proof.company_logs[i], hash = outcome.harvest.harvest_hash;
    assert.match(hash, /^sha256:[0-9a-f]{64}$/u);
    assert.equal(outcome.credited, "5e0"); assert.equal(outcome.credited_resource_id, "company.cash"); assert.equal(outcome.faucet_applied, true);
    assert.equal(outcome.forfeited_units, 0); assert.equal(outcome.cap_reason_key, null);
    assert.equal(companyEvent.intent_id, intent_id); assert.equal(founderEvent.intent_id, intent_id);
    assert.equal(companyEvent.payload.credited, "5e0"); assert.equal(companyEvent.payload.faucet_applied, true);
    assert.equal(companyEvent.payload.harvest_hash, hash); assert.deepEqual(founderEvent.payload, outcome.harvest);
    assert.equal(companyEvent.revision, outcome.company_revision); assert.equal(founderEvent.revision, outcome.founder_revision);
    assert.equal(founderLog.payload.intent_id, intent_id); assert.equal(companyLog.payload.intent_id, intent_id);
    assert.deepEqual(founderLog.receipt.harvest, outcome.harvest); assert.equal(companyLog.payload.harvest_hash, hash);
    assert.equal(companyLog.inputs.resolved.harvest_hash, hash); assert.equal(companyLog.receipt.harvest_hash, hash);
    assert.equal(companyLog.receipt.credited, "5e0"); assert.equal(companyLog.receipt.faucet_applied, true);
    assert.equal(founderLog.revision, outcome.founder_revision); assert.equal(companyLog.revision, outcome.company_revision);
    assert.equal(founderLog.company_stream, companyLog.company_stream); assert.equal(founderLog.run_seq, companyLog.run_seq); assert.equal(founderLog.run_log_seq, companyLog.seq);
  }
  assert.notEqual(single.intent_id, all.intent_id);
  publicData.forEach((data) => noSalt(data, salt));
  await page.reload({ waitUntil: "networkidle" }); await control("garden.title").click();
  await page.waitForFunction((name) => [...document.querySelectorAll(".garden button")].some((node) => node.textContent.trim() === name && node.getAttribute("aria-pressed") === "true"), text("garden.substrate.containerized.name"), { timeout: 30_000 });
  assert.deepEqual(head(founderID).founder, final.founder, "reload changed persisted Garden");
  await Promise.all(responses); publicData.forEach((data) => noSalt(data, salt)); assert.equal(errors.length, 0, errors.map(String).join("\n"));
  console.log(`Garden composed real wall-clock: DOM bootstrap/unlock/plant/uproot → native three-tick maturation → single/all harvest → substrate/reload; two real cash sends, bound hashes, hidden salt; ${(Date.now() - started) / 1000}s: PASS`);
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
