import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { readFileSync, writeFileSync } from "node:fs";
import { createServer } from "node:http";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { chromium, firefox, webkit } from "playwright";

// R-011 primitive research only. No production runtime, credentials or API.
const sourceHead = process.argv[2];
assert(process.argv.length === 3 && /^[0-9a-f]{40}$/u.test(sourceHead ?? ""), "expected one source HEAD argument");
const script = fileURLToPath(import.meta.url), root = path.resolve(path.dirname(script), "../..");
const report = {
  schema_version: 1, question: "R-011 primitive wave, not session acceptance", observed_at: new Date().toISOString(),
  source_head: sourceHead, instrument_sha256: createHash("sha256").update(readFileSync(script)).digest("hex"),
  platform: { os: os.platform(), arch: os.arch(), release: os.release() },
  expected: { engines: ["chromium", "firefox", "webkit"], arms: ["exclusive", "bypass", "distinct", "isolated"], repetitions: 10, arm_cases: 120, termination_cases: 30 },
  completed: { engines: 0, arm_cases: 0, termination_cases: 0 },
  status: "running", exclusions: [], guard_exhausted: false, runtime_errors: [], observations: [],
};
let row;
// This is the existing Playwright observation guard, not a measured product budget.
const observationGuardMS = 30_000;
async function within(promise, objective) {
  let timer;
  try {
    return await Promise.race([promise, new Promise((_, reject) => {
      timer = setTimeout(() => { report.guard_exhausted = true; reject(new Error(`invalid observation: guard fired before ${objective}`)); }, observationGuardMS);
    })]);
  } finally { clearTimeout(timer); }
}
async function until(predicate, objective) {
  let stopped = false;
  try {
    await within((async () => {
      while (!stopped && !await predicate()) await new Promise((resolve) => setTimeout(resolve, 10));
    })(), objective);
  } finally { stopped = true; }
}
function note(participant, phase, caseID) {
  assert(row && row.id === caseID, "late/stale case observation");
  assert(["a", "b", "c", "driver"].includes(participant));
  assert(["enter", "exit", "close_started", "close_completed"].includes(phase));
  row.trace.push([participant, phase, Number((performance.now() - row.started).toFixed(3))]);
}
function entered(participant) { return row.trace.some(([who, phase]) => who === participant && phase === "enter"); }
function maximumHolders(trace) {
  const holders = new Set(); let maximum = 0;
  for (const [who, phase] of trace) {
    assert(["enter", "exit"].includes(phase), "non-interval event in exclusion arm");
    if (phase === "enter") { assert(!holders.has(who), "duplicate acquisition"); holders.add(who); }
    else { assert(holders.delete(who), "release without acquisition"); }
    maximum = Math.max(maximum, holders.size);
  }
  assert.equal(holders.size, 0, "incomplete holder population");
  return maximum;
}
const server = createServer((request, response) => {
  if (request.method !== "GET" || request.url !== "/") { response.writeHead(404); response.end(); return; }
  response.writeHead(200, { "Content-Type": "text/html", "Cache-Control": "no-store" });
  response.end("<!doctype html><title>R-011 synthetic primitive observation</title>");
});
let browser;
try {
  await new Promise((resolve, reject) => { server.once("error", reject); server.listen(0, "127.0.0.1", resolve); });
  const origin = `http://127.0.0.1:${server.address().port}`;
  for (const [engine, browserType] of Object.entries({ chromium, firefox, webkit })) {
    browser = await browserType.launch({ headless: true });
    const shared = await browser.newContext(), isolated = await browser.newContext();
    const observation = { engine, version: browser.version(), user_agent: null, capabilities: null, storage: null, cases: [], termination: [] };
    report.observations.push(observation);
    async function page(context, participant) {
      const value = await context.newPage();
      value.on("pageerror", (error) => report.runtime_errors.push({ engine, participant, error: String(error) }));
      await value.exposeFunction("r011Observe", (phase, caseID) => note(participant, phase, caseID));
      await value.goto(origin, { waitUntil: "load" });
      return value;
    }
    let a = await page(shared, "a");
    const b = await page(shared, "b"), c = await page(isolated, "c");
    observation.user_agent = await a.evaluate(() => navigator.userAgent);
    observation.capabilities = await a.evaluate(() => ({ secure_context: isSecureContext, request: typeof navigator.locks?.request === "function", query: typeof navigator.locks?.query === "function" }));
    assert.deepEqual(observation.capabilities, { secure_context: true, request: true, query: true }, `${engine} native API unavailable; no fallback authorized`);
    const storageKey = "r011.synthetic-generation";
    await a.evaluate((key) => localStorage.setItem(key, "generation-1"), storageKey);
    const sharedValue = await b.evaluate((key) => localStorage.getItem(key), storageKey);
    const isolatedValue = await c.evaluate((key) => localStorage.getItem(key), storageKey);
    assert.equal(sharedValue, "generation-1", "same-context storage not shared");
    assert.equal(isolatedValue, null, "isolated context unexpectedly shares storage");
    observation.storage = { same_context: sharedValue, isolated_context: isolatedValue, replacement_reload: null };
    function start(owner, arm, name) {
      return owner.evaluate(async ({ caseID, arm, name }) => {
        const work = async () => {
          await window.r011Observe("enter", caseID);
          await new Promise((resolve) => { window.r011Release = resolve; });
          await window.r011Observe("exit", caseID);
          delete window.r011Release;
        };
        if (arm === "bypass") await work();
        else await navigator.locks.request(name, { mode: "exclusive" }, work);
      }, { caseID: row.id, arm, name }).then(() => ({ completed: true }), (error) => ({ completed: false, error: String(error) }));
    }
    const release = (owner) => within(owner.evaluate(() => {
      if (typeof window.r011Release !== "function") throw new Error("holder not ready for release");
      window.r011Release();
    }), "explicit holder release");
    async function pendingOrEntered(owner, participant, name) {
      let pending = false;
      await until(async () => {
        pending = await owner.evaluate(async (name) => (await navigator.locks.query()).pending.some((lock) => lock.name === name && lock.mode === "exclusive"), name);
        return pending || entered(participant);
      }, "native queued contender or overlapping control");
      return pending;
    }
    for (const arm of report.expected.arms) {
      for (let repetition = 0; repetition < 10; repetition++) {
        const name = `r011:${engine}:${arm}:${repetition}`, contender = arm === "isolated" ? c : b, participant = arm === "isolated" ? "c" : "b";
        row = { id: `${engine}:${arm}:${repetition}`, arm, repetition, started: performance.now(), trace: [], peak_holders: null, pending_observed: false };
        observation.cases.push(row);
        const first = start(a, arm, name);
        await until(() => entered("a"), "first holder acquisition");
        const second = start(contender, arm, arm === "distinct" ? `${name}:other` : name);
        if (arm === "exclusive") {
          row.pending_observed = await pendingOrEntered(b, "b", name);
          assert(!entered("b"), "exclusive lock admitted overlapping holders");
          assert(row.pending_observed, "no native queued contender observed");
          await release(a); await until(() => entered("b"), "queued successor acquisition");
          await release(b);
        } else {
          await until(() => entered(participant), "overlap-control acquisition");
          await Promise.all([release(a), release(contender)]);
        }
        assert.deepEqual(await within(first, "first completed holder"), { completed: true });
        assert.deepEqual(await within(second, "second completed holder"), { completed: true });
        assert.equal(row.trace.length, 4, "exact two-holder interval population");
        row.peak_holders = maximumHolders(row.trace);
        assert.equal(row.peak_holders, arm === "exclusive" ? 1 : 2, `${arm} oracle/control did not discriminate`);
        row.elapsed_ms = Number((performance.now() - row.started).toFixed(3)); delete row.started;
        report.completed.arm_cases++;
      }
      console.log(`R-011 ${engine} ${arm}: 10/10 completed, peak ${arm === "exclusive" ? 1 : 2}`);
    }
    for (let repetition = 0; repetition < 10; repetition++) {
      const name = `r011:${engine}:termination:${repetition}`;
      row = { id: `${engine}:termination:${repetition}`, repetition, started: performance.now(), trace: [], pending_observed: false };
      observation.termination.push(row);
      const first = start(a, "termination", name);
      await until(() => entered("a"), "termination holder acquisition");
      const second = start(b, "termination", name);
      row.pending_observed = await pendingOrEntered(b, "b", name);
      assert(row.pending_observed && !entered("b"), "successor was not excluded before owner closure");
      note("driver", "close_started", row.id);
      await within(a.close(), "actual owner page closure"); note("driver", "close_completed", row.id);
      assert(a.isClosed(), "owner remains open");
      await until(() => entered("b"), "successor after owner page termination");
      assert.equal((await within(first, "terminated owner's pending call")).completed, false, "closed owner completed rather than being interrupted");
      await release(b); assert.deepEqual(await within(second, "termination successor completion"), { completed: true });
      const closedAt = row.trace.findIndex(([who, phase]) => who === "driver" && phase === "close_started");
      const enteredAt = row.trace.findIndex(([who, phase]) => who === "b" && phase === "enter");
      assert(closedAt >= 0 && enteredAt > closedAt, "successor acquired before observed close start");
      assert.equal(row.trace.length, 5, "exact termination trace population");
      row.elapsed_ms = Number((performance.now() - row.started).toFixed(3)); delete row.started;
      report.completed.termination_cases++;
      a = await page(shared, "a");
    }
    await b.evaluate((key) => localStorage.setItem(key, "generation-2"), storageKey);
    await a.reload({ waitUntil: "load" });
    observation.storage.replacement_reload = await a.evaluate((key) => localStorage.getItem(key), storageKey);
    assert.equal(observation.storage.replacement_reload, "generation-2", "replacement/reload lost shared storage");
    console.log(`R-011 ${engine}: 10/10 actual page terminations; storage shared/isolated/reloaded; ${observation.version}`);
    report.completed.engines++;
    await shared.close(); await isolated.close(); await browser.close(); browser = undefined;
  }
  assert.deepEqual(report.completed, { engines: 3, arm_cases: 120, termination_cases: 30 });
  assert.equal(report.runtime_errors.length, 0, JSON.stringify(report.runtime_errors));
  assert.equal(report.guard_exhausted, false);
  report.status = "complete";
  writeFileSync(path.join(root, "planning/platform-alignment/browser-coordination-observation.v1.json"), `${JSON.stringify(report, null, 2)}\n`);
  console.log("R-011 primitive observation complete: 120 interval cases / 30 page terminations; NOT automatic-renewal, process-crash or production timing acceptance");
} catch (error) {
  report.status = "invalid";
  console.error(JSON.stringify({ source_head: report.source_head, instrument_sha256: report.instrument_sha256, observed_at: report.observed_at, status: report.status, completed: report.completed, exclusions: report.exclusions, guard_exhausted: report.guard_exhausted, runtime_errors: report.runtime_errors, last_case: row, error: String(error) }));
  throw error;
} finally {
  await browser?.close(); server.closeAllConnections();
  await new Promise((resolve, reject) => server.close((error) => error ? reject(error) : resolve()));
}
