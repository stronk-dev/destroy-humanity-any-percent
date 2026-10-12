import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import { createServer } from "node:http";
import os from "node:os";
import { fileURLToPath } from "node:url";
import { chromium, firefox, webkit } from "playwright";

// R-011 primitive experiment only. No game client, actual credential or Account API.
const [head, ...selected] = process.argv.slice(2);
const engines = { chromium, firefox, webkit };
assert(/^[0-9a-f]{40}$/u.test(head ?? ""), "expected exact source HEAD");
assert(selected.length > 0 && new Set(selected).size === selected.length &&
  selected.every((name) => Object.hasOwn(engines, name)), "expected unique declared browser names");
const arms = ["healthy", "interrupted", "without-marker", "quota-denied"];
const report = {
  question: "R-011 interrupted dispatch primitives, not renewal acceptance",
  source_head: head,
  instrument_sha256: createHash("sha256").update(readFileSync(fileURLToPath(import.meta.url))).digest("hex"),
  observed_at: new Date().toISOString(),
  platform: { os: os.platform(), arch: os.arch(), release: os.release() },
  selected_engines: selected,
  unexecuted_engines: Object.keys(engines).filter((name) => !selected.includes(name)),
  expected_cases: selected.length * arms.length * 5,
  completed_cases: 0,
  guard_exhausted: false,
  runtime_errors: [],
  observations: [],
  status: "running",
};
let row, browser;
async function within(promise, objective) {
  let timer;
  try {
    return await Promise.race([promise, new Promise((_, reject) => {
      timer = setTimeout(() => {
        report.guard_exhausted = true;
        reject(new Error(`invalid observation: guard exhausted before ${objective}`));
      }, 30_000);
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
const server = createServer((request, response) => {
  if (request.method === "GET" && request.url === "/") {
    response.writeHead(200, { "Content-Type": "text/html", "Cache-Control": "no-store" });
    response.end("<!doctype html><title>R-011 synthetic dispatch observation</title>");
    return;
  }
  if (request.method !== "POST" || request.url !== `/dispatch/${row?.id}`) {
    response.writeHead(404); response.end(); return;
  }
  const current = row;
  let body = "";
  request.on("data", (chunk) => { body += chunk; });
  request.on("end", () => {
    try {
      assert.equal(body, '{"generation":1}', "unexpected synthetic request");
      current.requests++;
      if (current.requests === 1) {
        current.reply = response; // Actual first HTTP reply remains held at the receiver.
      } else {
        response.writeHead(409, { "Content-Type": "application/json" });
        response.end('{"result":"reused"}');
      }
    } catch (error) {
      report.runtime_errors.push(String(error));
      response.writeHead(400); response.end();
    }
  });
});
const generationKey = "r011.synthetic-generation";
const markerKey = "r011.synthetic-dispatch-marker";
// Research envelope only; not an adopted credential document or production marker.
const marker = JSON.stringify({ v: 1, generation: "synthetic-generation-1", phase: "dispatched-or-unknown" });
function start(page, withoutMarker = false) {
  return page.evaluate(async ({ id, generationKey, markerKey, marker, withoutMarker }) => {
    return navigator.locks.request(`r011:${id}`, { mode: "exclusive" }, async () => {
      if (localStorage.getItem(generationKey) === "2") return { outcome: "newer", generation: 2 };
      if (!withoutMarker && localStorage.getItem(markerKey) !== null) {
        return { outcome: "ambiguous", generation: 1 };
      }
      if (!withoutMarker) {
        try { localStorage.setItem(markerKey, marker); }
        catch (error) { return { outcome: "storage-denied", name: error.name }; }
      }
      const response = await fetch(`/dispatch/${id}`, {
        method: "POST", body: JSON.stringify({ generation: 1 }),
        headers: { "Content-Type": "application/json" },
      });
      if (response.status === 409) return { outcome: "reused", generation: 1 };
      if (response.status !== 200) throw new Error("unexpected synthetic dispatch reply");
      const next = await response.json();
      if (next.generation !== 2) throw new Error("invalid synthetic generation");
      localStorage.setItem(generationKey, "2");
      localStorage.removeItem(markerKey);
      return { outcome: "rotated", generation: 2 };
    });
  }, { id: row.id, generationKey, markerKey, marker, withoutMarker })
    .then((result) => ({ completed: true, result }), (error) => ({ completed: false, error: String(error) }));
}
const storage = (page) => page.evaluate(({ generationKey, markerKey }) => ({
  generation: localStorage.getItem(generationKey), marker: localStorage.getItem(markerKey),
}), { generationKey, markerKey });

try {
  await new Promise((resolve, reject) => {
    server.once("error", reject); server.listen(0, "127.0.0.1", resolve);
  });
  const origin = `http://127.0.0.1:${server.address().port}`;
  for (const engine of selected) {
    // Playwright owns launch timeout/cancellation. Racing an uncancelled launch
    // could leave a late browser process outside our cleanup handle.
    browser = await engines[engine].launch({ headless: true, timeout: 30_000 });
    const observation = { engine, version: browser.version(), cases: [] };
    report.observations.push(observation);
    for (const arm of arms) {
      for (let repetition = 0; repetition < 5; repetition++) {
        row = { id: `${engine}-${arm}-${repetition}`, arm, requests: 0, pending_observed: false };
        const began = performance.now();
        const context = await browser.newContext();
        try {
          async function page() {
            const value = await context.newPage();
            value.on("pageerror", (error) => report.runtime_errors.push(String(error)));
            await value.goto(origin, { waitUntil: "load" });
            assert(await value.evaluate(() => isSecureContext && typeof navigator.locks?.query === "function" &&
              typeof navigator.locks?.request === "function"), "native Web Locks unavailable");
            return value;
          }
          let a = await page();
          const b = await page();
          await a.evaluate((key) => localStorage.setItem(key, "1"), generationKey);
          if (arm === "quota-denied") {
            row.quota = await a.evaluate(() => {
              let bytes = 0, index = 0, failures = 0;
              for (const length of [262144, 1024, 32]) {
                for (;;) {
                  if (bytes + 2 * length > 16 * 1024 * 1024) throw new Error("invalid observation: quota-fill guard");
                  try { localStorage.setItem(`fill-${index++}`, "q".repeat(length)); bytes += 2 * length; }
                  catch (error) {
                    if (!(error instanceof DOMException) || error.name !== "QuotaExceededError") throw error;
                    failures++; break;
                  }
                }
              }
              return { bytes, failures };
            });
            assert.equal(row.quota.failures, 3, "native quota did not refuse each fill size");
            row.first = await within(start(a), "quota-denied candidate completion");
            assert.deepEqual(row.first, { completed: true, result: { outcome: "storage-denied", name: "QuotaExceededError" } });
            assert.equal(row.requests, 0, "dispatch preceded durable marker admission");
            assert.deepEqual(await storage(b), { generation: "1", marker: null });
          } else {
            const first = start(a, arm === "without-marker");
            await until(() => row.requests === 1, "first actual HTTP dispatch");
            assert.deepEqual(await storage(b), { generation: "1", marker: arm === "without-marker" ? null : marker });
            const second = start(b, arm === "without-marker");
            await until(async () => {
              row.pending_observed = await b.evaluate(async (name) =>
                (await navigator.locks.query()).pending.some((lock) => lock.name === name), `r011:${row.id}`);
              return row.pending_observed;
            }, "native queued successor");
            if (arm === "healthy") {
              row.reply.writeHead(200, { "Content-Type": "application/json" });
              row.reply.end('{"generation":2}');
              row.first = await within(first, "healthy first completion");
              row.second = await within(second, "healthy queued completion");
              assert.deepEqual(row.first, { completed: true, result: { outcome: "rotated", generation: 2 } });
              assert.deepEqual(row.second, { completed: true, result: { outcome: "newer", generation: 2 } });
              assert.deepEqual(await storage(b), { generation: "2", marker: null });
              assert.equal(row.requests, 1);
            } else {
              await within(a.close(), "actual dispatched owner-page closure");
              assert(a.isClosed());
              const interrupted = await within(first, "closed owner's interrupted callback");
              assert.equal(interrupted.completed, false, "owner callback completed before page closure");
              row.first = { interrupted: true };
              row.second = await within(second, "successor after page closure");
              if (arm === "interrupted") {
                assert.deepEqual(row.second, { completed: true, result: { outcome: "ambiguous", generation: 1 } });
                a = await page(); await a.reload({ waitUntil: "load" });
                row.reloaded = await within(start(a), "replacement-page stop");
                assert.deepEqual(row.reloaded, row.second);
                assert.deepEqual(await storage(a), { generation: "1", marker });
                assert.equal(row.requests, 1, "queued/reloaded successor redispatched");
              } else {
                assert.deepEqual(row.second, { completed: true, result: { outcome: "reused", generation: 1 } });
                assert.equal(row.requests, 2, "marker-omission control did not redispatch");
                assert.deepEqual(await storage(b), { generation: "1", marker: null });
              }
            }
          }
          row.elapsed_ms = Number((performance.now() - began).toFixed(3));
          const { reply, ...retained } = row;
          observation.cases.push(retained);
          report.completed_cases++;
        } finally {
          row.reply?.destroy();
          await context.close();
        }
      }
      console.log(`R-011 ${engine} ${arm}: 5/5 complete`);
    }
    await browser.close(); browser = undefined;
  }
  assert.equal(report.completed_cases, report.expected_cases);
  assert.equal(report.guard_exhausted, false);
  assert.deepEqual(report.runtime_errors, []);
  report.status = "complete-selected-population";
  console.log(JSON.stringify(report));
} catch (error) {
  if (error.name === "TimeoutError" || String(error).includes("quota-fill guard")) report.guard_exhausted = true;
  report.status = "invalid";
  console.error(JSON.stringify({ ...report, failed_case: row && { id: row.id, arm: row.arm, requests: row.requests }, error: String(error) }));
  throw error;
} finally {
  await browser?.close();
  server.closeAllConnections();
  if (server.listening) await new Promise((resolve, reject) => server.close((error) => error ? reject(error) : resolve()));
}
