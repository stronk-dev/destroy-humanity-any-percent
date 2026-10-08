// MA2/MA3 real player seam. Default CI population is bounded cancellation;
// the explicit full selector waits the minted five-minute duration unchanged.
import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { readFileSync } from "node:fs";
import path from "node:path";

export async function witnessSoulRecovery(page, repositoryRoot, founderID, complete = false) {
  assert.match(founderID, /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/u);
  const catalog = JSON.parse(readFileSync(path.join(repositoryRoot, "balance/soul/first-content.json"), "utf8"));
  const activity = catalog.recovery_activities.find((row) => row.activity_id === "repot");
  assert.equal(activity.duration_attended_ms, 300_000);
  assert.equal(activity.recovery_amount, 5);
  const beatMS = Math.floor(catalog.policy.recovery_beat_ceiling_ms / 3);
  const copy = JSON.parse(readFileSync(path.join(repositoryRoot, "client/src/copy/generated/catalog.json"), "utf8"));
  const text = (key) => {
    const row = copy.entries.find((entry) => entry.key === key);
    assert(row && row.era_variants === null && row.params.length === 0);
    return row.text;
  };
  function sql(statement) {
    const result = spawnSync("docker", ["compose", "-f", "compose.game-ui-test.yml", "exec", "-T", "game-ui-postgres", "psql",
      "-v", "ON_ERROR_STOP=1", "-At", "-U", "cloud_clicker", "-d", "cloud_clicker_game_ui_test", "-c", statement], { cwd: repositoryRoot, encoding: "utf8" });
    assert.equal(result.status, 0, result.stderr); return result.stdout.trim();
  }
  const heads = () => JSON.parse(sql(`WITH current AS (
    SELECT DISTINCT ON(s.scope) s.scope,s.id,r.revision,r.state FROM save_streams s
    JOIN save_revisions r ON r.stream_id=s.id WHERE s.owner_kind='founder' AND s.owner_id='${founderID}'
    AND s.archived_at IS NULL ORDER BY s.scope,r.revision DESC
  ) SELECT jsonb_object_agg(scope,jsonb_build_object('stream_id',id,'revision',revision,'state',state)) FROM current;`));
  const stored = (sessionID) => {
    assert.match(sessionID, /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/u);
    return JSON.parse(sql(`SELECT row_to_json(r) FROM soul_recovery_sessions r WHERE founder_id='${founderID}' AND session_id='${sessionID}';`));
  };
  const requests = [], pending = [], errors = [];
  const onRequest = (request) => {
    if (new URL(request.url()).pathname.startsWith("/api/v1/soul-recovery/")) requests.push(request);
  };
  const onResponse = (response) => {
    if (!new URL(response.url()).pathname.startsWith("/api/v1/soul-recovery/")) return;
    pending.push(response.body().then((bytes) => {
      assert.equal(response.status(), 200, "native recovery request failed");
      JSON.parse(bytes.toString("utf8"));
    }).catch(() => errors.push(new Error("native recovery response invalid/unsuccessful (body withheld)"))));
  };
  page.on("request", onRequest); page.on("response", onResponse);
  const control = (key) => page.getByRole("button", { name: text(key), exact: true });
  const surface = () => page.locator('main[data-surface="soul_recovery"]');
  const native = async (button) => {
    // Trial actionability waits for the real control without activating it.
    await button.click({ trial: true, timeout: 30_000 });
    await button.focus(); await button.press("Enter");
  };
  async function producingCompany() {
    await native(page.getByRole("button", { name: "The Desk", exact: true }));
    await page.locator('main[data-surface="desk"]').waitFor({ state: "visible", timeout: 30_000 });
    const snapshot = () => page.evaluate(async () => {
      const credentials = JSON.parse(localStorage.getItem("cloud-clicker.credentials.v1"));
      const response = await fetch("/api/v1/founder/state", { headers: { Authorization: `Bearer ${credentials.accessToken}` } });
      if (response.status !== 200) throw new Error(`Recovery observation snapshot ${response.status}`);
      return response.json();
    });
    const purchase = page.locator('section[aria-labelledby="generators-heading"] article')
      .filter({ has: page.getByRole("heading", { name: text("generator.beige_tower.title"), exact: true }) })
      .getByRole("button", { name: text("desk.buy_one"), exact: true });
    async function gameplay(button, kind, fields, observed) {
      await button.click({ trial: true, timeout: 30_000 });
      await page.waitForFunction(() => document.querySelector("main")?.getAttribute("aria-busy") === "false");
      const before = observed ?? await snapshot();
      const reply = page.waitForResponse((response) => new URL(response.url()).pathname === "/api/v1/intents" &&
        response.request().postDataJSON()?.kind === kind, { timeout: 30_000 })
        .then((response) => ({ response }), (error) => ({ error }));
      await native(button);
      const result = await reply;
      if (result.error) throw result.error;
      const response = result.response, body = response.request().postDataJSON(), receipt = await response.json();
      const { intent_id, ...intent } = body;
      if (kind === "decline_exit_offer") {
        assert.match(intent.offer_id, /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/u);
        const { offer_id, ...remaining } = intent;
        assert.deepEqual(remaining, { kind, expected_revision: before.revision });
      } else assert.deepEqual(intent, { kind, expected_revision: before.revision, ...fields });
      assert.equal(response.status(), 200); assert.equal(receipt.outcome, "applied");
      assert.equal(receipt.intent_id, intent_id);
      await page.waitForFunction(() => document.querySelector("main")?.getAttribute("aria-busy") === "false");
      return before;
    }
    let manualActions = 0, offers = 0;
    async function afford() {
      for (let actions = 0; actions < 40; actions += 1) {
        if (await page.locator('main[data-surface="offer_sheet"]').count()) {
          assert(++offers <= 3, "recovery producer repeatedly preempted by Exit offers");
          await gameplay(page.getByRole("button", { name: "Decline", exact: true }), "decline_exit_offer", {});
          await native(page.getByRole("button", { name: "The Desk", exact: true }));
          await page.locator('main[data-surface="desk"]').waitFor();
        }
        const state = await snapshot();
        if (state.generators.find((row) => row.generator_id === "generator.beige_tower").max_affordable >= 1) return;
        await gameplay(page.getByRole("button", { name: "Fix Computer", exact: true }), "perform_manual_batch",
          { action_id: state.manual_action.action_id, count: 1, window_ms: 1 }, state);
        manualActions += 1;
        // Setup at 1 native action/s: one observer read plus the command and
        // host refresh stay below the actual 300/min account refill. No retry,
        // limiter override or production/cadence acceleration.
        await page.waitForTimeout(1_000);
      }
      throw new Error("recovery producer not affordable within bounded native manual actions");
    }
    await afford();
    await gameplay(purchase, "buy_generator", { generator_id: "generator.beige_tower", count: { mode: "exact", value: 1 } });
    const cash = (state) => state.resources.find((row) => row.resource_id === "company.cash");
    const producing = cash(await snapshot());
    // Only compare this small, known first-generator population. This is a
    // positive-production control, not replacement gameplay Decimal arithmetic.
    assert(Number.isFinite(Number(producing.rate_per_second)) && Number(producing.rate_per_second) > 0, "recovery population needs a nonzero idle producer");
    await afford();
    await page.waitForTimeout(beatMS);
    const beforeSecond = await gameplay(purchase, "buy_generator", { generator_id: "generator.beige_tower", count: { mode: "exact", value: 1 } });
    const cost = beforeSecond.generators.find((row) => row.generator_id === "generator.beige_tower").next_cost;
    assert(Number(cash(await snapshot()).amount) + Number(cost) > Number(cash(beforeSecond).amount),
      "normal purchase must accrue real idle production before recovery; not merely own an inert generator");
    console.log(`Soul Recovery production control: ${manualActions} native manual actions, two native purchases and positive stored accrual net of exact purchase cost: PASS`);
  }
  async function command(button, operation, body) {
    const count = requests.length;
    const response = page.waitForResponse((reply) => new URL(reply.url()).pathname === `/api/v1/soul-recovery/${operation}`, { timeout: 30_000 });
    await native(button);
    const reply = await response;
    assert.equal(reply.status(), 200, `native recovery ${operation}`);
    const emitted = requests.slice(count).filter((request) => new URL(request.url()).pathname === `/api/v1/soul-recovery/${operation}`);
    assert.equal(emitted.length, 1, `one native recovery ${operation} request`);
    assert.deepEqual(emitted[0].postDataJSON(), body);
    return reply.json();
  }
  async function open() {
    await native(control("soul.recovery_surface.title"));
    await surface().waitFor({ state: "visible", timeout: 30_000 });
    const row = surface().locator("li").filter({ has: page.getByRole("heading", { name: text(activity.title_copy_key), exact: true }) });
    assert.equal(await row.getByText(text(activity.disclosure_copy_key), { exact: true }).count(), 1);
    return row.getByRole("button", { name: text("soul.recovery_surface.begin"), exact: true });
  }
  async function waitBeat(sessionID) {
    const reply = await page.waitForResponse(async (response) => new URL(response.url()).pathname === "/api/v1/soul-recovery/progress" &&
      response.status() === 200 && (await response.json()).attended_progress_ms > 0, { timeout: 30_000 });
    const progress = await reply.json();
    assert.equal(progress.session_id, sessionID);
    assert.equal(progress.required_duration_attended_ms, activity.duration_attended_ms);
    await page.waitForFunction((value) => document.querySelector(".recovery progress")?.value >= value, progress.attended_progress_ms);
    return progress;
  }
  async function stopped() {
    // An already-returning request may settle; then two normal beat intervals
    // must elapse with no new dispatch. No fake time or adjusted cadence.
    await page.waitForTimeout(100);
    const count = requests.length;
    await page.waitForTimeout(beatMS * 2);
    assert.equal(requests.length, count, "unmounted/terminal recovery continued sending beats");
  }
  function terminalStored(receipt, action, before) {
    const row = stored(receipt.session_id), after = heads();
    assert.equal(row.status, action === "cancel" ? "cancelled" : "resolved");
    assert.deepEqual(row.terminal_receipt, receipt);
    assert.equal(receipt.outcome, "applied"); assert.equal(receipt.action, action);
    assert.equal(receipt.soul_before, before.founder.state.soul);
    assert.equal(receipt.soul_after, action === "cancel" ? receipt.soul_before : Math.min(catalog.policy.soul_max, receipt.soul_before + activity.recovery_amount));
    assert.equal(after.founder.state.soul, receipt.soul_after);
    assert.equal(after.founder.revision, receipt.founder_revision);
    assert.equal(after.company.revision, receipt.company_revision);
    assert.deepEqual(after.company.state.balances, before.company.state.balances, "recovery must not produce resources");
    const kind = action === "cancel" ? "soul_recovery_cancelled.v1" : "soul_recovered.v1";
    assert.equal(sql(`SELECT count(*) FROM events WHERE stream_id='${after.founder.stream_id}' AND intent_id='${receipt.session_id}' AND kind='${kind}';`), "1");
    assert.equal(sql(`SELECT count(*) FROM founder_log WHERE founder_stream_id='${after.founder.stream_id}' AND intent_id='${receipt.session_id}';`), "1");
    assert.equal(sql(`SELECT count(*) FROM events WHERE stream_id='${after.founder.stream_id}' AND intent_id='${receipt.session_id}' AND kind='soul_recovery_started.v1';`), "1");
    return after;
  }
  try {
    await producingCompany();
    const before = heads();
    const started = await command(await open(), "start", { activity_id: activity.activity_id });
    assert.equal(started.required_duration_attended_ms, activity.duration_attended_ms);
    assert(stored(started.session_id).progress_token === started.progress_token, "stored/start capability mismatch (tokens withheld)");
    await waitBeat(started.session_id);
    assert.equal(await control("soul.recovery_surface.finish").count(), 0, "early Finish must not appear");
    await native(page.getByRole("button", { name: "The Desk", exact: true }));
    await page.locator('main[data-surface="desk"]').waitFor();
    await stopped();
    const paused = stored(started.session_id);
    assert.equal(paused.status, "active");
    const reconnected = await command(await open(), "start", { activity_id: activity.activity_id });
    assert.equal(reconnected.session_id, started.session_id);
    assert(reconnected.progress_token !== started.progress_token, "native reconnect must rotate the capability (tokens withheld)");
    assert.equal(reconnected.attended_progress_ms, paused.attended_progress_ms);
    assert(stored(started.session_id).progress_token === reconnected.progress_token, "stored/reconnect capability mismatch (tokens withheld)");
    await waitBeat(started.session_id);
    const cancelled = await command(control("soul.recovery_surface.cancel"), "cancel", { session_id: started.session_id });
    assert.equal(cancelled.cancelled_by, "player");
    terminalStored(cancelled, "cancel", before);
    await page.waitForFunction(() => document.activeElement?.id === "recovery-heading");
    assert.equal(await surface().getByText(text("soul.recovery_surface.cancelled"), { exact: true }).count(), 1);
    await stopped();
    console.log("Soul Recovery native Begin → real attended beat → leave/stopped scheduler → same-session token rotation → Stop early → exact SQL receipt/event/unchanged Soul: PASS");

    if (complete) {
      const original = heads();
      // Explicit depletion setup, not a minted debit, natural earned state or
      // replay proof. Only Soul/revision change; catalog and time stay real.
      assert.equal(sql(`WITH current AS (SELECT r.* FROM save_revisions r WHERE stream_id='${original.founder.stream_id}' ORDER BY revision DESC LIMIT 1),
        seeded AS (INSERT INTO save_revisions(stream_id,revision,version,state,constants_hash)
        SELECT stream_id,revision+1,version,jsonb_set(state,'{soul}','40'::jsonb,false),constants_hash FROM current RETURNING revision)
        SELECT revision FROM seeded;`), String(original.founder.revision + 1));
      const depleted = heads();
      assert.equal(depleted.founder.state.soul, 40);
      assert.equal(JSON.stringify({ ...depleted.founder.state, soul: original.founder.state.soul }), JSON.stringify(original.founder.state));
      assert.equal(JSON.stringify(depleted.company), JSON.stringify(original.company));
      await page.reload({ waitUntil: "networkidle" });
      const beganMS = Date.now();
      const active = await command(await open(), "start", { activity_id: activity.activity_id });
      assert.equal(active.required_duration_attended_ms, 300_000);
      console.log("Soul Recovery full population: explicit Soul40 test setup; minted300000ms/5Soul unchanged; waiting for real browser heartbeats, no clock/progress grant");
      const heartbeat = setInterval(() => console.log(`Soul Recovery real-time observation: ${Math.floor((Date.now() - beganMS) / 1000)}s elapsed`), 30_000);
      try { await control("soul.recovery_surface.finish").waitFor({ state: "visible", timeout: activity.duration_attended_ms + 60_000 }); }
      finally { clearInterval(heartbeat); }
      assert(Date.now() - beganMS >= activity.duration_attended_ms, "natural activity completed before its real duration");
      const ready = stored(active.session_id);
      assert(ready.attended_progress_ms >= activity.duration_attended_ms);
      assert.equal(heads().founder.state.soul, 40, "Soul granted before native Finish");
      const resolved = await command(control("soul.recovery_surface.finish"), "resolve", { session_id: active.session_id });
      terminalStored(resolved, "resolve", depleted);
      assert.equal(resolved.soul_after, 45);
      await page.waitForFunction(() => document.activeElement?.id === "recovery-heading");
      assert.match(await surface().textContent(), /Soul went from 40 to 45/u);
      await stopped();
      await page.reload({ waitUntil: "networkidle" });
      assert.equal(heads().founder.state.soul, 45);
      await open();
      assert.equal(await surface().locator("progress").count(), 0, "reload retained a terminal activity as active");
      console.log("Soul Recovery real five-minute native Begin → natural heartbeat eligibility → Finish → stored/event/rendered Soul40→45 → reload: PASS; depletion setup, not natural debit/history replay or 15/45-minute activity proof");
    }
    await Promise.all(pending);
    assert.equal(errors.length, 0, errors.map(String).join("\n"));
  } finally {
    page.off("request", onRequest); page.off("response", onResponse);
  }
}
