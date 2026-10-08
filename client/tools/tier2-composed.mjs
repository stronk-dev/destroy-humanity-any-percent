// Tier-2 A/B/E integration, not pacing, Headcount, an epoch mint or AC10.
import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { createHash } from "node:crypto";
import { copyFileSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { productionClientProof } from "./production-client-proof.mjs";

export function tier2Fixture(repositoryRoot) {
  const root = mkdtempSync(path.join(os.tmpdir(), "cloud-clicker-tier2-composed-"));
  const write = (relative, bytes) => {
    const target = path.join(root, relative);
    mkdirSync(path.dirname(target), { recursive: true }); writeFileSync(target, bytes);
  };
  const seed = JSON.parse(readFileSync(path.join(repositoryRoot, "balance/epochs/phase0.json")));
  const replacements = { economy: "economy", routes: "routes", categories: "categories" };
  const artifacts = seed.artifacts.map(({ name, path: original }) => ({ name, bytes: readFileSync(path.join(repositoryRoot,
    replacements[name] ? `balance/testdata/t2/${replacements[name]}-candidate-v1.json` : original)) })).sort((a, b) => a.name < b.name ? -1 : 1);
  const digest = createHash("sha256"), frame = Buffer.alloc(8);
  for (const { name, bytes } of artifacts) {
    const encoded = Buffer.from(name);
    frame.writeBigUInt64BE(BigInt(encoded.length)); digest.update(frame).update(encoded);
    frame.writeBigUInt64BE(BigInt(bytes.length)); digest.update(frame).update(bytes);
    write(`balance/tier2-fixture/${name}.json`, bytes);
  }
  const hash = `sha256:${digest.digest("hex")}`;
  assert.equal(hash, readFileSync(path.join(repositoryRoot, "balance/testdata/t2/candidate-bundle-hash.txt"), "utf8").trim(), "exact existing unminted candidate bundle");
  write("balance/epochs/phase0.json", JSON.stringify({ schema_version: 1, current_epoch_id: 1,
    artifacts: artifacts.map(({ name }) => ({ name, path: `balance/tier2-fixture/${name}.json` })),
    epochs: [{ epoch_id: 1, name: "Unminted Tier-2 test fixture", changelog_ref: "changelog/epoch-1.md", accepted_hashes: [hash] }],
  }));
  write("changelog/epoch-1.md", "# Controlled Tier-2 test fixture, not epoch 9\n");
  for (const relative of ["moderation/guild-names.txt", "balance/api/phase0.json", "balance/transport/phase0.json"]) {
    const target = path.join(root, relative); mkdirSync(path.dirname(target), { recursive: true }); copyFileSync(path.join(repositoryRoot, relative), target);
  }
  return { root, hash, cleanup: () => rmSync(root, { recursive: true, force: true }) };
}

export async function witnessTier2(browser, uiURL, builtFiles, repositoryRoot, fixture) {
  const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });
  const errors = [], requests = [], assets = [], frames = [];
  const proof = productionClientProof(builtFiles), actions = [];
  page.on("pageerror", (error) => errors.push(String(error)));
  page.on("requestfailed", (request) => errors.push(`${new URL(request.url()).pathname}: ${request.failure()?.errorText}`));
  page.on("request", (request) => { if (request.method() === "POST" && new URL(request.url()).pathname === "/api/v1/intents") requests.push(request); });
  page.context().on("response", (response) => {
    const url = new URL(response.url());
    if (url.origin !== uiURL || url.pathname.startsWith("/api/") || url.pathname === "/favicon.ico") return;
    assets.push((response.status() === 304 ? Promise.resolve(undefined) : response.body())
      .then((bytes) => proof.response(url.pathname, response.status(), bytes)).catch((error) => errors.push(String(error))));
  });
  page.on("worker", (worker) => { try { proof.worker(new URL(worker.url()).pathname); } catch (error) { errors.push(String(error)); } });
  page.on("websocket", (socket) => socket.on("framereceived", ({ payload }) => frames.push(String(payload))));
  function sql(statement) {
    const result = spawnSync("docker", ["compose", "-f", "compose.game-ui-test.yml", "exec", "-T", "game-ui-postgres", "psql",
      "-v", "ON_ERROR_STOP=1", "-At", "-U", "cloud_clicker", "-d", "cloud_clicker_game_ui_test", "-c", statement], { cwd: repositoryRoot, encoding: "utf8" });
    assert.equal(result.status, 0, result.stderr); return result.stdout.trim();
  }
  const snapshot = () => page.evaluate(async () => {
    const credentials = JSON.parse(localStorage.getItem("cloud-clicker.credentials.v1"));
    const response = await fetch("/api/v1/founder/state", { headers: { Authorization: `Bearer ${credentials.accessToken}` } });
    if (response.status !== 200) throw new Error(`Tier-2 snapshot ${response.status}`); return response.json();
  });
  const button = (label) => page.getByRole("button", { name: label, exact: true });
  async function intent(control, kind, fields, key = "Enter") {
    await page.waitForFunction(() => document.querySelector("main")?.getAttribute("aria-busy") === "false", undefined, { timeout: 30_000 });
    const before = await snapshot(), count = requests.length;
    await control.click({ trial: true, timeout: 30_000 });
    const pending = page.waitForResponse((response) => new URL(response.url()).pathname === "/api/v1/intents" && response.request().postDataJSON()?.kind === kind,
      { timeout: 30_000 }).then((response) => ({ response }), (error) => ({ error }));
    await control.focus(); await page.keyboard.press(key);
    const result = await pending;
    if (result.error) throw new Error(`Tier-2 native ${kind} emitted ${requests.length - count} requests and no matching response`, { cause: result.error });
    const receipt = await result.response.json(), emitted = requests.slice(count);
    assert.equal(emitted.length, 1, `one native ${kind}`);
    const { intent_id, ...command } = emitted[0].postDataJSON();
    assert.match(intent_id, /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u);
    assert.deepEqual(command, { kind, expected_revision: before.revision, ...fields });
    assert.equal(result.response.status(), 200); assert.equal(receipt.outcome, "applied", JSON.stringify(receipt));
    assert.equal(receipt.intent_id, intent_id);
    // Prestige commits a terminal revision and the next Company's genesis;
    // ordinary Company commands advance once (production/prestige.go).
    assert.equal(receipt.new_revision, before.revision + (kind === "wind_down" ? 2 : 1));
    if (kind === "wind_down") {
      assert.equal(receipt.founder_revision, before.founder_revision + 1);
      assert.equal(receipt.snapshot.run_seq, before.run.run_seq + 1);
    }
    // Ordinary replay commands store canonical fields separately from the
    // log's intent_id column; compare both, without inventing a payload field.
    actions.push({ intent_id, kind, command, receipt });
    await page.waitForFunction(() => document.querySelector("main")?.getAttribute("aria-busy") === "false", undefined, { timeout: 30_000 });
    return receipt;
  }
  let founderID;
  function seedCash(amount) {
    assert.match(founderID, /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/u);
    assert(["1e5", "9e8"].includes(amount));
    assert.equal(sql(`WITH current AS (
      SELECT r.* FROM save_revisions r JOIN save_streams s ON s.id=r.stream_id
      WHERE s.owner_id='${founderID}' AND s.scope='company' AND s.archived_at IS NULL ORDER BY r.revision DESC LIMIT 1
    ), seeded AS (INSERT INTO save_revisions(stream_id,revision,version,state,constants_hash)
      SELECT stream_id,revision+1,version,jsonb_set(state,'{balances,company.cash}',to_jsonb('${amount}'::text),false),constants_hash FROM current RETURNING revision)
      SELECT count(*) FROM seeded;`), "1");
    console.log(`Tier-2 fixture: explicit disposable cash setup ${amount}; not pacing or earned-cash proof`);
  }
  async function desk() {
    await page.locator('main[data-surface="desk"]').waitFor({ state: "visible", timeout: 30_000 });
    assert.deepEqual(errors, [], "built-client errors");
  }
  async function declineOffer(receipt) {
    if (receipt.snapshot.offer_state !== null) {
      await page.locator('main[data-surface="offer_sheet"]').waitFor({ timeout: 30_000 });
      await intent(button("Decline"), "decline_exit_offer", { offer_id: receipt.snapshot.offer_state.offer_id });
      await button("The Desk").click(); await desk();
    }
  }
  try {
    await page.goto(uiURL, { waitUntil: "networkidle" }); await button("BEGIN ATTEMPT").press("Enter"); await desk();
    await page.getByText(/You are visitor #\d+/u).waitFor({ state: "visible", timeout: 30_000 });
    const initial = await snapshot(); founderID = initial.run.founder_id;
    assert.equal(initial.constants_hash, fixture.hash); assert.equal(initial.run.tier, 0);
    assert.equal(initial.transitions.incorporate, undefined);
    seedCash("1e5"); await page.reload({ waitUntil: "networkidle" }); await desk();
    await intent(button("Move Into the Garage"), "cross_gate", { gate_id: "gate.t0_to_t1", route_id: null });
    // Reach a genuine second Company; do not fabricate Founder Exit history.
    await intent(button("Wind Down Company"), "wind_down", { expected_founder_revision: (await snapshot()).founder_revision });
    await page.locator('main[data-surface="run_end"]').waitFor({ timeout: 30_000 });
    await button("Start the Next Company").press("Enter"); await desk();
    seedCash("9e8"); await page.reload({ waitUntil: "networkidle" }); await desk();
    await declineOffer(await intent(button("Move Into the Garage"), "cross_gate", { gate_id: "gate.t0_to_t1", route_id: null }));
    const tierOne = await snapshot();
    assert.equal(tierOne.run.tier, 1); assert.equal(tierOne.transitions.incorporate, undefined);
    assert.equal(tierOne.transitions.cross_gate.gate_id, "gate.t1_to_t2"); assert.equal(tierOne.transitions.cross_gate.eligible, true);
    await declineOffer(await intent(button("Move Into the Garage"), "cross_gate", { gate_id: "gate.t1_to_t2", route_id: null }));
    const tierTwo = await snapshot(); assert.equal(tierTwo.run.tier, 2); assert.equal(tierTwo.run.run_seq, 2);
    assert.equal(await page.locator("main").getAttribute("data-era"), "era_2010");
    const presentation = JSON.parse(readFileSync(path.join(repositoryRoot, "balance/testdata/t2/presentation-candidate-v3.json")));
    const copy = JSON.parse(readFileSync(path.join(repositoryRoot, "client/src/copy/generated/catalog.json")));
    const title = (key) => { const row = copy.entries.find((entry) => entry.key === key); assert(row && row.era_variants === null); return row.text; };
    const generatorCard = (id) => page.locator('section[aria-labelledby="generators-heading"] article').filter({
      has: page.getByRole("heading", { name: title(presentation.generators.find((row) => row.id === id).title_key), exact: true }),
    });
    for (const id of ["generator.open_plan_floor", "generator.managed_services_contract", "generator.hot_desk_program"]) {
      await intent(generatorCard(id).getByRole("button", { name: "Buy 1", exact: true }), "buy_generator", { generator_id: id, count: { mode: "exact", value: 1 } });
    }
    for (const id of ["upgrade.ping_pong_table", "upgrade.move_fast_break_things", "upgrade.nap_pod"]) {
      await intent(page.locator(`[data-upgrade="${id}"]`).getByRole("button", { name: "Buy 1", exact: true }), "buy_upgrade", { upgrade_id: id }, "Space");
    }
    const faction = tierTwo.transitions.incorporate.factions.find((row) => row.faction_id === "open_source"); assert(faction);
    await intent(page.locator("fieldset.incorporate").getByRole("button", { name: title(faction.copy_key), exact: true }), "incorporate", { faction_id: "open_source" });
    const final = await snapshot(); assert.equal(final.transitions.incorporate, undefined);
    const state = JSON.parse(sql(`SELECT r.state FROM save_revisions r JOIN save_streams s ON s.id=r.stream_id WHERE s.owner_id='${founderID}' AND s.scope='company' AND s.archived_at IS NULL ORDER BY r.revision DESC LIMIT 1;`));
    assert.equal(state.tier, 2); assert.equal(state.faction_id, "open_source"); assert.equal(state.gates_crossed["gate.t1_to_t2"], true);
    for (const id of ["generator.open_plan_floor", "generator.managed_services_contract", "generator.hot_desk_program"]) assert.equal(state.generators[id], 1);
    for (const id of ["upgrade.ping_pong_table", "upgrade.move_fast_break_things", "upgrade.nap_pod"]) assert(state.upgrades_owned.includes(id));
    const history = JSON.parse(sql(`SELECT json_agg(json_build_object('intent_id',intent_id,'command',convert_from(canonical_payload,'UTF8')::jsonb,'receipt',receipt,'revision',applied_revision) ORDER BY seq) FROM run_log WHERE company_stream_id IN (SELECT id FROM save_streams WHERE owner_id='${founderID}' AND scope='company' AND archived_at IS NULL) AND run_seq=${final.run.run_seq};`));
    for (const action of actions.slice(2)) {
      const matching = history.filter((row) => row.intent_id === action.intent_id);
      assert.equal(matching.length, 1, `one durable log for ${action.kind}`);
      assert.deepEqual(matching[0].command, action.command); assert.deepEqual(matching[0].receipt, action.receipt);
      assert.equal(matching[0].revision, action.receipt.new_revision);
    }
    await page.reload({ waitUntil: "networkidle" }); await desk();
    const restored = await snapshot(); assert.equal(restored.run.tier, 2); assert.equal(restored.revision, final.revision); assert.equal(restored.constants_hash, fixture.hash);
    assert.equal(await page.locator("fieldset.incorporate").count(), 0);
    for (const id of ["generator.open_plan_floor", "generator.managed_services_contract", "generator.hot_desk_program"]) {
      assert.equal(restored.generators.find((row) => row.generator_id === id)?.owned, 1);
      assert.equal(await generatorCard(id).getByText("Owned: 1", { exact: true }).count(), 1, "reload renders persisted ownership");
    }
    for (const id of ["upgrade.ping_pong_table", "upgrade.move_fast_break_things", "upgrade.nap_pod"]) assert.equal(await page.locator(`[data-upgrade="${id}"] button`).isDisabled(), true);
    await Promise.all(assets); assert.deepEqual(errors, []); assert(frames.length > 0, "real transport");
    console.log(`Tier-2 exact built HTML/JS/CSS/Worker: ${JSON.stringify(proof.finish())}: PASS`);
    console.log("Tier-2 candidate: native first failure/next Company → both gates → three generators/three upgrades → incorporation → SQL history and reload: PASS; explicit cash fixture, no Headcount/pacing/mint/full-acceptance claim");
  } catch (error) { throw new Error(`Tier-2 composed failed; browser errors=${JSON.stringify(errors)}`, { cause: error }); }
  finally { await page.close(); }
}
