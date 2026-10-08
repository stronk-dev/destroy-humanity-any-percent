// Reputation R5/R6/R9: actual player controls and persistence, not earned-budget
// pacing, replay verification of the seeded history, content mint or full AC15.
import assert from "node:assert/strict";
import { productionClientFiles, productionClientProof } from "./production-client-proof.mjs";

export async function witnessReputation({ browser, uiURL, clientDist, hash, snapshot, sql, copy }) {
  const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });
  const requests = [], failures = [], assets = [], actions = [];
  const proof = productionClientProof(productionClientFiles(clientDist));
  page.on("pageerror", (error) => failures.push(String(error)));
  page.on("requestfailed", (request) => failures.push(`${new URL(request.url()).pathname}: ${request.failure()?.errorText}`));
  page.on("request", (request) => {
    if (request.method() === "POST" && new URL(request.url()).pathname === "/api/v1/intents") requests.push(request);
  });
  page.context().on("response", (response) => {
    const url = new URL(response.url());
    if (url.origin !== uiURL || url.pathname.startsWith("/api/") || url.pathname === "/favicon.ico") return;
    assets.push(response.body().then((bytes) => proof.response(url.pathname, response.status(), bytes))
      .catch((error) => failures.push(String(error))));
  });
  page.on("worker", (worker) => { try { proof.worker(new URL(worker.url()).pathname); } catch (error) { failures.push(String(error)); } });
  const button = (label) => page.getByRole("button", { name: label, exact: true });
  const ready = () => page.waitForFunction(() => document.querySelector("main")?.getAttribute("aria-busy") === "false", undefined, { timeout: 30_000 });
  const surface = async (name) => { await page.locator(`main[data-surface="${name}"]`).waitFor({ timeout: 30_000 }); await ready(); };
  let founderID;
  const stored = () => JSON.parse(sql(`SELECT json_agg(json_build_object('scope',s.scope,'revision',r.revision,'state',r.state) ORDER BY s.scope)
    FROM save_streams s JOIN LATERAL (SELECT * FROM save_revisions WHERE stream_id=s.id ORDER BY revision DESC LIMIT 1) r ON true
    WHERE s.owner_id='${founderID}' AND s.archived_at IS NULL;`, true).trim());
  const frozen = () => JSON.parse(sql(`SELECT json_agg(json_build_object('run_seq',f.run_seq,'source_id',f.source_id,'slot',f.slot,'target',f.target,'factor',f.factor)
    ORDER BY f.run_seq,f.source_id) FROM run_frozen_contributions f JOIN save_streams s ON s.id=f.company_stream_id WHERE s.owner_id='${founderID}';`, true).trim());
  async function intent(control, kind, fields, founder = false) {
    await ready(); const before = await snapshot(page), count = requests.length;
    await control.click({ trial: true, timeout: 30_000 }); await control.focus();
    const responseTask = page.waitForResponse((response) => new URL(response.url()).pathname === "/api/v1/intents" && response.request().postDataJSON()?.kind === kind,
      { timeout: 30_000 }).then((response) => ({ response }), (error) => ({ error }));
    await control.press("Enter"); const result = await responseTask;
    if (result.error) throw new Error(`native Reputation ${kind}: ${requests.length - count} requests, no matching response`, { cause: result.error });
    assert.equal(requests.length - count, 1, `one native ${kind}`);
    const { intent_id, ...command } = requests.at(-1).postDataJSON();
    assert.match(intent_id, /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u);
    assert.deepEqual(command, { kind, expected_revision: founder ? before.founder_revision : before.revision, ...fields });
    const receipt = await result.response.json();
    assert.equal(result.response.status(), 200); assert.equal(receipt.outcome, "applied", JSON.stringify(receipt));
    assert.equal(receipt.intent_id, intent_id);
    if (founder) assert.equal(receipt.founder_revision, before.founder_revision + 1);
    else assert.equal(receipt.new_revision, before.revision + (kind === "wind_down" ? 2 : 1));
    actions.push({ intent_id, command, receipt }); await ready(); return receipt;
  }
  const row = (arm, id) => page.locator(".reputation .nodes > li").filter({
    has: page.getByRole("heading", { name: copy(arm.nodes.find((node) => node.node_id === id).title_key), exact: true }),
  });
  async function purchase(id, spent, cancelFirst = false) {
    const before = await snapshot(page), arm = before.features.reputation, control = row(arm, id).getByRole("button");
    const head = stored(), contributions = frozen(), count = requests.length;
    // A current snapshot can precede transport subscription acknowledgement.
    // Wait for the actual enabled control before one native activation.
    await control.click({ trial: true, timeout: 30_000 }); await control.focus(); await control.press("Enter");
    const confirm = row(arm, id).getByRole("button", { name: copy("reputation_tree.action.confirm"), exact: true });
    await confirm.waitFor(); assert.equal(await confirm.evaluate((node) => document.activeElement === node), true);
    if (cancelFirst) {
      await confirm.press("Escape");
      assert.equal(await control.count(), 1); assert.equal(await control.evaluate((node) => document.activeElement === node), true);
      assert.equal(requests.length, count, "Escape must not spend"); assert.deepEqual(stored(), head, "Escape preserves both stored heads");
      await control.click({ trial: true, timeout: 30_000 }); await control.focus(); await control.press("Space"); await confirm.waitFor();
    }
    const receipt = await intent(confirm, "purchase_reputation_node", { node_id: id }, true);
    assert.equal(receipt.effective_from, "next_run"); assert.equal(receipt.resolved_cost, arm.nodes.find((node) => node.node_id === id).cost);
    assert.equal(receipt.reputation_available_after, 6 - spent); assert.equal(receipt.reputation_unlock_ppm_after, 50_000);
    const after = await snapshot(page);
    assert.equal(after.features.reputation.level, 6); assert.equal(after.features.reputation.spent, spent);
    assert.equal(after.features.reputation.available, 6 - spent);
    assert.equal(after.features.reputation.bonus_factor_this_run, "1e0");
    assert.equal(after.features.reputation.bonus_factor_next_run, "1.003e0");
    assert(Number(before.resources.find((resource) => resource.resource_id === "company.cash").rate_per_second) > 0, "nonzero rate makes next-run-only observable");
    assert.equal(after.resources.find((resource) => resource.resource_id === "company.cash").rate_per_second, before.resources.find((resource) => resource.resource_id === "company.cash").rate_per_second);
    assert.deepEqual(stored().find((head) => head.scope === "company"), head.find((head) => head.scope === "company"));
    assert.deepEqual(frozen(), contributions, "direct purchase must not rewrite any frozen row");
    await page.waitForFunction((title) => [...document.querySelectorAll(".reputation .nodes > li")].some((node) =>
      node.querySelector("h2")?.textContent === title && node.getAttribute("data-state") === "owned"), copy(arm.nodes.find((node) => node.node_id === id).title_key));
    assert.equal(await row(arm, id).getByRole("button").count(), 0, "owned purchase has no control");
    assert.equal(await row(arm, id).evaluate((node) => document.activeElement === node), true, "applied purchase retains row focus");
    console.log(`Reputation native ${id}: exact Founder receipt, persisted ownership, nonzero current rate unchanged and next-run bonus: PASS`);
  }
  try {
    await page.goto(uiURL, { waitUntil: "networkidle" }); await button("BEGIN ATTEMPT").press("Enter"); await surface("desk");
    await page.getByText(/You are visitor #\d+/u).waitFor({ timeout: 30_000 });
    const initial = await snapshot(page); founderID = initial.run.founder_id;
    assert.match(founderID, /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/u); assert.equal(initial.constants_hash, hash);
    assert.equal(initial.facts.find((row) => row.fact_id === "feature.reputation_tree")?.value, true); assert.equal(initial.features.reputation.available, 0);
    // Only earned budget and cash are seeded. No owned nodes, spent,
    // Exit history, generated units, bonus rows, plans, receipts or events.
    assert.equal(sql(`WITH heads AS (SELECT s.scope,r.* FROM save_streams s JOIN LATERAL
      (SELECT * FROM save_revisions WHERE stream_id=s.id ORDER BY revision DESC LIMIT 1) r ON true
      WHERE s.owner_id='${founderID}' AND s.archived_at IS NULL), seeded AS (
      INSERT INTO save_revisions(stream_id,revision,version,state,constants_hash)
      SELECT stream_id,revision+1,version,CASE WHEN scope='founder' THEN jsonb_set(state,'{reputation_level}','6'::jsonb,false)
        ELSE jsonb_set(state,'{balances,company.cash}','"2e5"'::jsonb,false) END,constants_hash FROM heads RETURNING stream_id)
      SELECT count(*) FROM seeded;`, true).trim(), "2");
    console.log("Reputation fixture: initial earned budget6/cash2e5; not earned-budget/pacing or seeded-history replay proof");
    await page.reload({ waitUntil: "networkidle" }); await surface("desk");
    await intent(button("Move Into the Garage"), "cross_gate", { gate_id: "gate.t0_to_t1", route_id: null });
    await intent(button("Wind Down Company"), "wind_down", { expected_founder_revision: (await snapshot(page)).founder_revision });
    await surface("run_end"); await button("Start the Next Company").press("Enter"); await surface("desk");
    const second = await snapshot(page); assert.equal(second.run.run_seq, 2);
    // Curriculum starter assignment varies. Buy a real producing unit instead
    // of assuming every branch already contains one or accepting a vacuous 0=0.
    assert.equal(sql(`WITH current AS (SELECT r.* FROM save_revisions r JOIN save_streams s ON s.id=r.stream_id
      WHERE s.owner_id='${founderID}' AND s.scope='company' AND s.archived_at IS NULL ORDER BY revision DESC LIMIT 1), seeded AS (
      INSERT INTO save_revisions(stream_id,revision,version,state,constants_hash)
      SELECT stream_id,revision+1,version,jsonb_set(state,'{balances,company.cash}','"2e5"'::jsonb,false),constants_hash FROM current RETURNING revision)
      SELECT count(*) FROM seeded;`, true).trim(), "1");
    console.log("Reputation fixture: explicit run2 cash2e5 for native production/gate controls, not earned cash");
    await page.reload({ waitUntil: "networkidle" }); await surface("desk");
    const generatorIndex = (await snapshot(page)).generators.findIndex((row) => row.generator_id === "generator.beige_tower");
    assert(generatorIndex >= 0, "producing generator is present in the actual projected catalog");
    await intent(page.locator('section[aria-labelledby="generators-heading"] article').nth(generatorIndex).getByRole("button", { name: "Buy 1", exact: true }),
      "buy_generator", { generator_id: "generator.beige_tower", count: { mode: "exact", value: 1 } });
    const gate = await intent(button("Move Into the Garage"), "cross_gate", { gate_id: "gate.t0_to_t1", route_id: null });
    if (gate.snapshot.offer_state !== null) {
      await surface("offer_sheet");
      await intent(button("Decline"), "decline_exit_offer", { offer_id: gate.snapshot.offer_state.offer_id });
      await button("The Desk").press("Enter"); await surface("desk");
    }
    await button(copy("reputation_tree.title")).press("Enter"); await surface("reputation_tree");
    await purchase("reputation.unlock.p05", 1, true);
    await purchase("reputation.starter.cash_small", 3);
    await page.reload({ waitUntil: "networkidle" }); await surface("desk");
    const restored = await snapshot(page); assert.equal(restored.features.reputation.spent, 3);
    assert.equal(restored.features.reputation.bonus_factor_this_run, "1e0"); assert.equal(restored.features.reputation.bonus_factor_next_run, "1.003e0");
    await page.locator(".plan summary").press("Enter");
    const plannedID = "reputation.starter.generated_beige_tower", planned = restored.features.reputation.nodes.find((node) => node.node_id === plannedID);
    const checkbox = page.locator(".plan label").filter({ hasText: copy(planned.title_key) }).getByRole("checkbox");
    assert.equal(await checkbox.isEnabled(), true); await checkbox.press("Space"); assert.equal(await checkbox.isChecked(), true);
    const receipt = await intent(button("Wind Down Company"), "wind_down", { expected_founder_revision: restored.founder_revision, reputation_plan: [plannedID] });
    await surface("run_end"); await button("Start the Next Company").press("Enter"); await surface("desk");
    const third = await snapshot(page), arm = third.features.reputation;
    assert.equal(third.run.run_seq, 3); assert.equal(arm.level, 6); assert.equal(arm.spent, 6); assert.equal(arm.available, 0);
    assert.equal(arm.bonus_factor_this_run, "1.003e0"); assert.equal(arm.bonus_factor_next_run, "1.003e0");
    const heads = stored(), company = heads.find((head) => head.scope === "company").state, founder = heads.find((head) => head.scope === "founder").state;
    assert.deepEqual(founder.reputation_nodes_owned, ["reputation.starter.cash_small", plannedID, "reputation.unlock.p05"]);
    assert.equal(company.generators_provisioned["generator.beige_tower"], 5); assert.equal(company.generators["generator.beige_tower"] ?? 0, 0);
    assert.equal(company.balances["company.cash"], "1e3"); assert.equal(heads.find((head) => head.scope === "company").revision, receipt.new_revision);
    assert.deepEqual(frozen().filter((row) => row.source_id === "reputation.founder_bonus").map(({ run_seq, factor }) => [run_seq, factor]), [[1, "1e0"], [2, "1e0"], [3, "1.003e0"]]);
    const events = JSON.parse(sql(`SELECT json_agg(payload ORDER BY revision,event_id) FROM events
      WHERE stream_id=(SELECT id FROM save_streams WHERE owner_id='${founderID}' AND scope='founder' AND archived_at IS NULL)
      AND kind='reputation_node_purchased.v1';`, true).trim());
    assert.deepEqual(events.map((event) => [event.node_id,event.cost,event.reputation_spent_after,event.source]),
      [["reputation.unlock.p05",1,1,"direct"],["reputation.starter.cash_small",2,3,"direct"],[plannedID,3,6,"exit_plan"]]);
    const logs = JSON.parse(sql(`SELECT json_agg(json_build_object('intent_id',intent_id,'receipt',receipt,'inputs',replay_inputs,'revision',applied_revision) ORDER BY seq) FROM founder_log
      WHERE founder_stream_id=(SELECT id FROM save_streams WHERE owner_id='${founderID}' AND scope='founder' AND archived_at IS NULL);`, true).trim());
    for (const action of actions.filter((action) => action.command.kind === "purchase_reputation_node" || action.command.reputation_plan)) {
      const rows = logs.filter((row) => row.intent_id === action.intent_id); assert.equal(rows.length, 1);
      if (action.command.kind === "purchase_reputation_node") assert.deepEqual(rows[0].receipt, action.receipt);
      else {
        // Exit has linked, distinct Company and Founder receipts. Bind the
        // public Company receipt exactly and inspect the Founder plan arm;
        // do not pretend both streams store the same receipt shape.
        const companyReceipts = JSON.parse(sql(`SELECT json_agg(receipt) FROM run_log WHERE intent_id='${action.intent_id}'
          AND company_stream_id=(SELECT id FROM save_streams WHERE owner_id='${founderID}' AND scope='company' AND archived_at IS NULL);`, true).trim());
        assert.deepEqual(companyReceipts, [action.receipt]);
        assert.equal(rows[0].receipt.intent_id, action.intent_id); assert.equal(rows[0].receipt.outcome, "applied");
        assert.equal(rows[0].revision, action.receipt.founder_revision);
        assert.equal(rows[0].inputs.resolved.kind, "exit.v2");
        assert.deepEqual(rows[0].inputs.resolved.reputation_purchases, [{ node_id: plannedID, resolved_cost: 3 }]);
      }
    }
    await page.reload({ waitUntil: "networkidle" }); await surface("desk");
    const reloaded = await snapshot(page); assert.equal(reloaded.run.run_seq, 3); assert.equal(reloaded.features.reputation.spent, 6);
    assert.equal(reloaded.features.reputation.bonus_factor_this_run, "1.003e0");
    await button(copy("reputation_tree.title")).press("Enter"); await surface("reputation_tree");
    for (const node of reloaded.features.reputation.nodes) {
      const nodeRow = row(reloaded.features.reputation, node.node_id);
      assert.equal(await nodeRow.getAttribute("data-state"), node.state); assert.equal(await nodeRow.getByRole("button").count(), 0);
    }
    await Promise.all(assets); assert.deepEqual(failures, []);
    console.log(`Reputation exact production-client bytes/started Worker: ${JSON.stringify(proof.finish())}: PASS`);
    console.log("Reputation native Escape/purchase → frozen-current/next-only bonus → reload → Exit plan → persisted starters/bonus/receipts/events → run3 reload: PASS; controlled budget, not mint/pacing/full career verification");
  } finally { await page.close(); }
}
