import assert from "node:assert/strict";
import { once, EventEmitter } from "node:events";
import { createServer, request as httpRequest } from "node:http";
import test from "node:test";
import { chromium } from "playwright";
import { navigationAPIProof } from "./navigation-api-proof.mjs";

test("real navigation cancels a held old-page read; complete upstream bytes remain required", { timeout: 30_000 }, async () => {
  const proof = navigationAPIProof();
  let release, observed;
  const held = new Promise((resolve) => { release = resolve; });
  const complete = new Promise((resolve) => { observed = resolve; });
  const upstream = createServer((_request, response) => {
    release(response);
  });
  await new Promise((resolve) => upstream.listen(0, "127.0.0.1", resolve));
  const upstreamURL = `http://127.0.0.1:${upstream.address().port}`;
  let oldResponse;
  const proxy = createServer((request, response) => {
    if (request.url === "/api/v1/founder/state") {
      oldResponse = response;
      const row = proof.observe(response, request.url, request.method, request.headers.referer);
      const forwarded = httpRequest(upstreamURL, (received) => {
        row.status = received.statusCode;
        const chunks = [];
        received.on("data", (chunk) => chunks.push(chunk));
        received.on("end", () => {
          row.upstream_ended = true;
          observed(Buffer.concat(chunks).toString("utf8"));
        });
        response.writeHead(received.statusCode, received.headers);
        received.pipe(response);
      });
      forwarded.on("error", (error) => observed(error));
      forwarded.end();
      return;
    }
    response.setHeader("content-type", "text/html");
    response.end(request.url === "/" ? '<script>fetch("/api/v1/founder/state").catch(() => {})</script>' : "<p>Reload destination</p>");
  });
  await new Promise((resolve) => proxy.listen(0, "127.0.0.1", resolve));
  let browser;
  try {
    browser = await chromium.launch({ headless: true });
    const page = await browser.newPage();
    const url = `http://127.0.0.1:${proxy.address().port}`;
    await page.goto(url, { waitUntil: "load" });
    const upstreamResponse = await held;
    const bytes = '{"founder_revision":7}';
    await proof.reload(page.url(), async () => {
      const closed = once(oldResponse, "close");
      const navigating = page.goto(`${url}/next`, { waitUntil: "load" });
      await closed;
      assert.equal(proof.boundaries[0].cancelled_by_navigation, 1);
      assert.equal(proof.boundaries[0].browser_finished, false);
      assert.throws(() => proof.finish(), /API proxy observation incomplete/u, "navigation cannot excuse missing upstream JSON");
      upstreamResponse.end(bytes);
      assert.equal(await complete, bytes, "enumerate the complete real upstream body after browser cancellation");
      await navigating;
    });
    assert.deepEqual(proof.finish(), { upstream_responses: 1, navigation_cancelled_reads: 1 });
    // The old guard rejected this actual navigation despite complete upstream data.
    const row = proof.boundaries[0];
    assert.equal(row.status !== null && row.upstream_ended && row.browser_finished, false);
  } finally {
    await browser?.close();
    await Promise.all([proxy, upstream].map((server) => new Promise((resolve) => {
      server.close(resolve); server.closeAllConnections();
    })));
  }
});

for (const [name, method, path, timing] of [
  ["Founder read outside navigation", "GET", "/api/v1/founder/state", "outside"],
  ["Garden read outside navigation", "GET", "/api/v1/garden/current", "outside"],
  ["command during navigation", "POST", "/api/v1/intents", "during"],
  ["unrelated read during navigation", "GET", "/api/v1/unknown", "during"],
  ["new-page Founder read during navigation", "GET", "/api/v1/founder/state", "new"],
  ["old-page read closing after navigation", "GET", "/api/v1/founder/state", "after"],
]) {
  test(`${name} still fails when unfinished`, async () => {
    const proof = navigationAPIProof(), response = new EventEmitter();
    let row;
    const observe = () => { row = proof.observe(response, path, method, timing === "new" ? "http://fixture/new" : "http://fixture/old"); row.status = 200; row.upstream_ended = true; };
    if (timing !== "new") observe();
    if (timing === "outside") response.emit("close");
    else await proof.reload("http://fixture/old", async () => {
      if (timing === "new") observe();
      if (timing !== "after") response.emit("close");
    });
    if (timing === "after") response.emit("close");
    assert.throws(() => proof.finish(), /API proxy observation incomplete/u);
  });
}

test("finished responses pass without a navigation exception", () => {
  const proof = navigationAPIProof(), response = new EventEmitter();
  const row = proof.observe(response, "/api/v1/intents", "POST");
  row.status = 200; row.upstream_ended = true;
  response.emit("finish"); response.emit("close");
  assert.deepEqual(proof.finish(), { upstream_responses: 1, navigation_cancelled_reads: 0 });
});

for (const path of ["/api/v1/founder/state", "/api/v1/garden/current"]) {
  test(`queued old-document ${path} arriving during navigation is explicitly counted`, async () => {
    const proof = navigationAPIProof(), response = new EventEmitter();
    await proof.reload("http://fixture/old", async () => {
      const row = proof.observe(response, path, "GET", "http://fixture/old");
      row.status = 200; row.upstream_ended = true;
      response.emit("close");
    });
    assert.deepEqual(proof.finish(), { upstream_responses: 1, navigation_cancelled_reads: 1 });
  });
}
