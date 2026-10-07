// Manual research, not a release/CI acceptance lane. Uses the actual DOM,
// runtime, HTTP limiter and composed server; no gameplay API setup or retries.
export function composedMode(args) {
  if (args.length === 0) return "journey";
  if (args.length === 1 && args[0] === "--observe-manual-budget") return "manual-budget";
  throw new Error("unknown composed arguments (use no arguments or --observe-manual-budget)");
}

// Do not print arbitrary receipt detail or identifiers. Recognize only the
// specific stale-coordinate pair being investigated; expose others as other.
export function observedRejectionPair(body) {
  return body.rejection?.category === "revision_conflict" && body.rejection?.detail === "expected_revision"
    ? "revision_conflict/expected_revision" : "other";
}

export async function waitUntilDeadline(deadline, now = () => performance.now(), pause = (ms) => new Promise((resolve) => setTimeout(resolve, ms))) {
  for (;;) {
    const remaining = deadline - now();
    if (!Number.isFinite(remaining)) throw new Error("invalid observation clock/deadline");
    if (remaining <= 0) return;
    await pause(Math.ceil(remaining));
  }
}

export function summarizeManualBudget({ hz, elapsedMS, latenessMS, attempts, activations, requests, pageErrors }) {
  const scheduled = hz * 60;
  const counts = {};
  for (const row of requests) {
    const key = `${row.method} ${row.route} ${row.status}${row.outcome ? ` ${row.outcome}` : ""}${row.rejectionPair ? `/${row.rejectionPair}` : ""}`;
    counts[key] = (counts[key] ?? 0) + 1;
  }
  const manual = requests.filter((row) => row.manual);
  const firstLimited = requests.find((row) => row.status === 429);
  const errors = [];
  if (![2, 4].includes(hz) || !Number.isFinite(elapsedMS) || !Number.isFinite(latenessMS)) errors.push("invalid profile coordinates");
  if (attempts !== scheduled || elapsedMS < 60_000) errors.push("incomplete input population");
  if (latenessMS > 100) errors.push("input cadence exceeded predeclared 100 ms lateness");
  if (!Number.isInteger(activations) || activations < 1 || activations > attempts || manual.length !== activations) errors.push("native activation/request population mismatch");
  if (requests.some((row) => ![200, 429].includes(row.status) || row.invalidBody ||
      row.manual && row.status === 200 && !["applied", "rejected"].includes(row.outcome) ||
      row.status === 429 && row.outcome !== "rate_limited/account")) errors.push("incomplete/invalid response population");
  if (pageErrors !== 0) errors.push("uncaught browser errors");
  return { hz, scheduled, attempts, activations, nonactivating_inputs: attempts - activations,
    elapsed_ms: elapsedMS, maximum_input_lateness_ms: latenessMS,
    requests: requests.length, counts, first_429: firstLimited ? { at_ms: firstLimited.at_ms, route: firstLimited.route } : null,
    valid: errors.length === 0, errors };
}

export async function observeManualBudget(browser, uiURL) {
  // Separate fresh accounts; serial profiles avoid cross-profile scheduling load.
  for (const hz of [2, 4]) {
    const page = await browser.newPage({ viewport: { width: 1280, height: 720 } });
    const requests = [], replies = [], pageErrors = [];
    const byRequest = new WeakMap();
    let start, attempts = 0, latenessMS = 0;
    page.on("pageerror", (error) => pageErrors.push(error));
    page.on("request", (request) => {
      const route = new URL(request.url()).pathname;
      if (start === undefined || !request.headers().authorization || !route.startsWith("/api/v1/")) return;
      const row = { at_ms: Math.round(performance.now() - start), method: request.method(), route, status: "pending",
        manual: route === "/api/v1/intents" && request.postDataJSON()?.kind === "perform_manual_batch" };
      requests.push(row); byRequest.set(request, row);
    });
    page.on("requestfailed", (request) => { const row = byRequest.get(request); if (row) row.status = "failed"; });
    page.on("response", (response) => {
      const row = byRequest.get(response.request());
      if (!row) return;
      row.status = response.status();
      if (row.manual || row.status === 429) replies.push((async () => {
        try {
          const body = await response.json();
          if (row.status === 200 && ["applied", "rejected"].includes(body.outcome)) {
            row.outcome = body.outcome;
            if (body.outcome === "rejected") row.rejectionPair = observedRejectionPair(body);
          }
          else if (row.status === 429 && body.category === "rate_limited" && body.detail === "account") row.outcome = "rate_limited/account";
          else row.invalidBody = true;
        } catch { row.invalidBody = true; }
      })());
    });
    try {
      await page.addInitScript(() => {
        globalThis.__manualBudgetActivations = 0;
        document.addEventListener("click", (event) => {
          if (event.isTrusted && event.target instanceof Element && event.target.matches(".manual button")) globalThis.__manualBudgetActivations++;
        }, true);
      });
      await page.goto(uiURL, { waitUntil: "networkidle" });
      await page.getByRole("button", { name: "BEGIN ATTEMPT", exact: true }).click();
      await page.getByText(/You are visitor #\d+/u).waitFor({ state: "visible", timeout: 30_000 });
      await page.waitForFunction(() => document.querySelector("main")?.getAttribute("aria-busy") === "false");
      await page.locator(".manual button").focus();
      start = performance.now();
      for (let index = 0; index < hz * 60; index++) {
        const due = start + index * 1_000 / hz;
        await waitUntilDeadline(due);
        latenessMS = Math.max(latenessMS, performance.now() - due);
        attempts++;
        // Native input even while disabled: do not force a click or delay the
        // prescribed cadence until a control becomes enabled.
        await page.locator(".manual button").focus();
        await page.keyboard.press("Enter");
      }
      await waitUntilDeadline(start + 60_000);
      const elapsedMS = performance.now() - start;
      await page.waitForFunction(() => document.querySelector("main")?.getAttribute("aria-busy") === "false", undefined, { timeout: 30_000 });
      await page.evaluate(() => new Promise((resolve) => requestAnimationFrame(() => requestAnimationFrame(resolve))));
      await Promise.all(replies);
      const activations = await page.evaluate(() => globalThis.__manualBudgetActivations);
      const result = summarizeManualBudget({ hz, elapsedMS, latenessMS, attempts, activations, requests, pageErrors: pageErrors.length });
      console.info(`manual account-budget observation (not release acceptance): ${JSON.stringify(result)}`);
      if (!result.valid) throw new Error(`invalid manual account-budget observation: ${result.errors.join("; ")}`);
    } finally { await page.close(); }
  }
}
