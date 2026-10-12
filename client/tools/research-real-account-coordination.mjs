import assert from "node:assert/strict";
import { createServer, request as httpRequest } from "node:http";

// Test-only coordinator and reply boundary. No production credential document,
// automatic renewal, expiry rule or recovery policy is implemented here.
export async function observeRealAccountCoordination({ browser, serverURL, sql }) {
  const credentialKey = "r011.real-account-test-pair";
  const markerKey = "r011.real-account-test-marker";
  const gameTables = ["accounts", "account_emails", "account_founders", "save_streams", "save_revisions",
    "events", "intent_records", "run_genesis", "founder_genesis", "transport_player_outbox"];
  const privateSnapshot = () => gameTables.map((table) => sql(
    `SELECT COALESCE(jsonb_agg(to_jsonb(r) ORDER BY to_jsonb(r)::text),'[]'::jsonb) FROM ${table} r;`)).join("\n");
  const errors = [], observations = [];
  let current;
  async function within(promise, objective) {
    let timer;
    try {
      return await Promise.race([promise, new Promise((_, reject) => {
        timer = setTimeout(() => reject(new Error(`invalid observation: guard exhausted before ${objective}`)), 30_000);
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
  const receiver = createServer((request, response) => {
    if (request.method === "GET" && request.url === "/") {
      response.writeHead(200, { "Content-Type": "text/html", "Cache-Control": "no-store" });
      response.end("<!doctype html><title>R-011 real Account observation</title>"); return;
    }
    if (!["/api/v1/bootstrap", "/api/v1/session/refresh", "/api/v1/founder/state"].includes(request.url)) {
      response.writeHead(404); response.end(); return;
    }
    const owner = current, refresh = request.url === "/api/v1/session/refresh";
    if (refresh) owner.refresh_requests++;
    const upstream = httpRequest(new URL(request.url, serverURL), {
      method: request.method, headers: { ...request.headers, host: new URL(serverURL).host },
    }, (received) => {
      const chunks = [];
      received.on("data", (chunk) => chunks.push(chunk));
      received.on("end", () => {
        const body = Buffer.concat(chunks);
        const deliver = () => { response.writeHead(received.statusCode, received.headers); response.end(body); };
        try {
          if (refresh) {
            owner.upstream_statuses.push(received.statusCode);
            if (owner.upstream_statuses.length === 1) {
              assert.equal(received.statusCode, 200, "first real rotation failed (body withheld)");
              assert.equal(received.headers["cache-control"], "no-store");
              assert.equal(received.headers["content-type"], "application/json");
              const pair = JSON.parse(body);
              assert(Object.keys(pair).sort().join("\0") === "access_token\0refresh_token" &&
                typeof pair.access_token === "string" && pair.access_token.length > 0 &&
                typeof pair.refresh_token === "string" && pair.refresh_token.length > 0,
              "real rotation has an invalid pair (body withheld)");
              owner.held = { deliver, response, pair }; // Private, never serialized into the report.
              return;
            }
          }
          deliver();
        } catch (error) {
          errors.push(error); response.destroy();
        }
      });
    });
    upstream.on("error", (error) => { errors.push(error); response.destroy(); });
    request.pipe(upstream);
  });
  try {
    await new Promise((resolve, reject) => { receiver.once("error", reject); receiver.listen(0, "127.0.0.1", resolve); });
    const origin = `http://127.0.0.1:${receiver.address().port}`;
    for (const arm of ["healthy", "interrupted", "without-marker"]) {
      current = { arm, refresh_requests: 0, upstream_statuses: [] };
      const context = await browser.newContext();
      try {
        async function page() {
          const result = await context.newPage();
          result.on("pageerror", (error) => errors.push(error));
          await result.goto(origin, { waitUntil: "load" });
          assert(await result.evaluate(() => isSecureContext && typeof navigator.locks?.request === "function" &&
            typeof navigator.locks?.query === "function"), "native Web Locks unavailable");
          return result;
        }
        let a = await page();
        const b = await page();
        const identities = await a.evaluate(async (key) => {
          const response = await fetch("/api/v1/bootstrap", {
            method: "POST", headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ idempotency_key: Array.from(crypto.getRandomValues(new Uint8Array(32)),
              (byte) => byte.toString(16).padStart(2, "0")).join("") }),
          });
          if (response.status !== 201) throw new Error(`real test bootstrap refused: status=${response.status} (body withheld)`);
          const data = await response.json();
          if (typeof data.session.access_token !== "string" || typeof data.session.refresh_token !== "string") {
            throw new Error("real test bootstrap lacks a pair (body withheld)");
          }
          localStorage.setItem(key, JSON.stringify({ generation: 1,
            accessToken: data.session.access_token, refreshToken: data.session.refresh_token }));
          return { account: data.account.account_id, founder: data.game_ui_snapshot.run.founder_id };
        }, credentialKey);
        assert.match(identities.account, /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u);
        const familyCounts = () => JSON.parse(sql(`SELECT jsonb_build_array(
          (SELECT count(*) FROM sessions WHERE account_id='${identities.account}'),
          (SELECT count(*) FROM sessions WHERE account_id='${identities.account}' AND consumed_at IS NOT NULL),
          (SELECT count(*) FROM sessions WHERE account_id='${identities.account}' AND revoked_at IS NOT NULL),
          (SELECT count(*) FROM access_tokens WHERE account_id='${identities.account}'),
          (SELECT count(*) FROM access_tokens WHERE account_id='${identities.account}' AND revoked_at IS NOT NULL),
          (SELECT count(*) FROM session_families WHERE account_id='${identities.account}' AND revoked_at IS NOT NULL));`));
        assert.deepEqual(familyCounts(), [1, 0, 0, 1, 0, 0], "fresh real family");
        const before = privateSnapshot();
        const storedPair = (owner) => owner.evaluate((key) => localStorage.getItem(key), credentialKey);
        const initialPair = await storedPair(a);
        function start(owner) {
          return owner.evaluate(async ({ arm, credentialKey, markerKey }) => {
            return navigator.locks.request(`r011-real:${arm}`, { mode: "exclusive" }, async () => {
              const previous = JSON.parse(localStorage.getItem(credentialKey));
              if (previous.generation === 2) return { outcome: "newer" };
              if (arm !== "without-marker" && localStorage.getItem(markerKey) !== null) return { outcome: "ambiguous" };
              if (arm !== "without-marker") localStorage.setItem(markerKey, JSON.stringify({ generation: 1, uncertain: true }));
              const response = await fetch("/api/v1/session/refresh", {
                method: "POST", headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ refresh_token: previous.refreshToken }),
              });
              const pair = await response.json();
              if (response.status === 401 && pair.category === "refresh_reused" && pair.detail === "session_family_revoked") {
                return { outcome: "reused", status: 401 };
              }
              if (response.status !== 200 || Object.keys(pair).sort().join("\0") !== "access_token\0refresh_token" ||
                typeof pair.access_token !== "string" || typeof pair.refresh_token !== "string" ||
                !pair.access_token || !pair.refresh_token || pair.access_token === previous.accessToken || pair.refresh_token === previous.refreshToken) {
                throw new Error("real test rotation failed (pair/body withheld)");
              }
              localStorage.setItem(credentialKey, JSON.stringify({ generation: 2, accessToken: pair.access_token, refreshToken: pair.refresh_token }));
              localStorage.removeItem(markerKey);
              return { outcome: "rotated", status: 200 };
            });
          }, { arm, credentialKey, markerKey }).then((result) => ({ completed: true, result }), () => ({ completed: false }));
        }
        const first = start(a);
        await until(() => current.held !== undefined || errors.length > 0, "actual upstream refresh response");
        assert.equal(errors.length, 0, "upstream refresh observation failed (private output withheld)");
        assert.deepEqual(familyCounts(), [2, 1, 0, 2, 0, 0], "rotation committed before client delivery");
        const second = start(b);
        await until(() => b.evaluate(async (name) => (await navigator.locks.query()).pending.some((lock) => lock.name === name),
          `r011-real:${arm}`), "native queued contender");
        if (arm === "healthy") {
          current.held.deliver();
          assert.deepEqual(await within(first, "healthy owner completion"), { completed: true, result: { outcome: "rotated", status: 200 } });
          assert.deepEqual(await within(second, "healthy successor completion"), { completed: true, result: { outcome: "newer" } });
          assert((await storedPair(a)) === (await storedPair(b)) && (await storedPair(a)) !== initialPair,
            "tabs did not retain the same new pair (credentials withheld)");
          assert.equal(current.refresh_requests, 1);
          assert.deepEqual(current.upstream_statuses, [200]);
        } else {
          await within(a.close(), "actual owner page closure after commit");
          assert(a.isClosed());
          assert.deepEqual(await within(first, "closed owner interruption"), { completed: false });
          if (arm === "interrupted") {
            assert.deepEqual(await within(second, "stopped successor"), { completed: true, result: { outcome: "ambiguous" } });
            a = await page(); await a.reload({ waitUntil: "load" });
            assert.deepEqual(await within(start(a), "stopped reloaded successor"), { completed: true, result: { outcome: "ambiguous" } });
            assert.equal(current.refresh_requests, 1);
            assert.deepEqual(current.upstream_statuses, [200]);
            assert.deepEqual(familyCounts(), [2, 1, 0, 2, 0, 0]);
          } else {
            assert.deepEqual(await within(second, "real reuse control"), { completed: true, result: { outcome: "reused", status: 401 } });
            assert.equal(current.refresh_requests, 2);
            assert.deepEqual(current.upstream_statuses, [200, 401]);
            assert.deepEqual(familyCounts(), [2, 1, 2, 2, 2, 1], "real reuse must revoke every family generation");
            const rotatedAccess = await fetch(new URL("/api/v1/founder/state", serverURL), {
              headers: { Authorization: `Bearer ${current.held.pair.access_token}` },
            });
            assert.equal(rotatedAccess.status, 401, "revoked descendant remained usable");
            assert.deepEqual(await rotatedAccess.json(), { category: "unauthorized", detail: "access_token" });
          }
          assert((await storedPair(b)) === initialPair, "interruption/reuse overwrote the original pair (credentials withheld)");
        }
        const access = await b.evaluate(async (key) => {
          const credentials = JSON.parse(localStorage.getItem(key));
          const response = await fetch("/api/v1/founder/state", { headers: { Authorization: `Bearer ${credentials.accessToken}` } });
          const data = await response.json();
          return response.status === 200 ? { status: 200, founder: data.run.founder_id } : { status: response.status, error: data };
        }, credentialKey);
        if (arm === "without-marker") assert.deepEqual(access, { status: 401, error: { category: "unauthorized", detail: "access_token" } });
        else assert.deepEqual(access, { status: 200, founder: identities.founder });
        assert(privateSnapshot() === before, "refresh changed watched account/game tables (contents withheld)");
        assert.equal(errors.length, 0, "observation has errors (private output withheld)");
        observations.push({ arm, refresh_requests: current.refresh_requests, upstream_statuses: current.upstream_statuses,
          family_counts: familyCounts(), native_queued: true, game_rows_unchanged: true, access_status: access.status });
      } finally {
        current.held?.response.destroy();
        await context.close();
      }
    }
    assert.equal(observations.length, 3);
    console.log(`R-011 real Chromium/Account/Postgres observation: ${JSON.stringify({ observations,
      watched_tables: gameTables, credentials_logged: false, automatic_renewal: false })}`);
  } finally {
    receiver.closeAllConnections();
    if (receiver.listening) await new Promise((resolve, reject) => receiver.close((error) => error ? reject(error) : resolve()));
  }
}
