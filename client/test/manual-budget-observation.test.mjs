import { expect, it } from "vitest";
import { composedMode, observedRejectionPair, observeManualBudget, summarizeManualBudget, waitUntilDeadline } from "../tools/observe-manual-budget.mjs";

const sample = () => ({ hz: 2, elapsedMS: 60_000, latenessMS: 20, attempts: 120, activations: 120,
  requests: Array.from({ length: 120 }, (_, index) => ({ manual: true, method: "POST", route: "/api/v1/intents", status: 200, outcome: "applied", at_ms: index * 500 })), pageErrors: 0 });

// Reuse the reviewer's bounded page/clock reproduction to execute the actual
// observer's registered listeners. This is capture-boundary coverage, not native
// input, a real minute-long measurement, or a duplicate response classifier.
async function captureObservation(scenario) {
  const performanceDescriptor = Object.getOwnPropertyDescriptor(globalThis, "performance");
  const originalTimeout = globalThis.setTimeout, originalInfo = console.info;
  let now = 0, readParses = 0;
  const reports = [];
  const encoder = new TextEncoder();
  try {
    Object.defineProperty(globalThis, "performance", { configurable: true, value: { now: () => now } });
    globalThis.setTimeout = (callback, ms) => { now += ms; queueMicrotask(callback); return 0; };
    console.info = (line) => reports.push(JSON.parse(line.slice(line.indexOf("{"))));
    const browser = {
      async newPage() {
        const handlers = new Map();
        let activations = 0;
        const emit = (manual) => {
          const request = {
            url: () => `http://fixture/api/v1/${manual ? "intents" : "founder/state"}`,
            headers: () => ({ authorization: "controlled-test" }),
            method: () => manual ? "POST" : "GET",
            postDataJSON: () => manual ? { kind: "perform_manual_batch" } : undefined,
          };
          handlers.get("request")(request);
          const body = async () => {
            if (!manual) readParses++;
            if (!manual && scenario === "failed-read") throw new Error("synthetic-private-body");
            if (scenario === (manual ? "malformed-manual" : "malformed-read") || !manual && scenario === "malformed-rate-limit") {
              return encoder.encode("{synthetic-private-body");
            }
            if (scenario === (manual ? "invalid-utf8-manual" : "invalid-utf8-read")) {
              return new Uint8Array([...encoder.encode(manual ? '{"outcome":"applied","fixture":"' : '{"fixture":"'), 0xff, ...encoder.encode('"}')]);
            }
            const value = manual ? { outcome: "applied" } : scenario === "rate-limited"
              ? { category: "rate_limited", detail: "account" } : { fixture: "synthetic-private-body" };
            return encoder.encode(`${!manual && scenario === "bom-read" ? "\uFEFF" : ""}${JSON.stringify(value)}`);
          };
          handlers.get("response")({
            request: () => request,
            status: () => !manual && scenario === "unexpected-empty" ? 204
              : !manual && ["rate-limited", "malformed-rate-limit"].includes(scenario) ? 429 : 200,
            body,
            json: async () => JSON.parse(new TextDecoder("utf-8", { ignoreBOM: true }).decode(await body())),
          });
        };
        return {
          on: (event, handler) => handlers.set(event, handler),
          addInitScript: async () => {}, goto: async () => {}, close: async () => {},
          getByRole: () => ({ click: async () => {} }), getByText: () => ({ waitFor: async () => {} }),
          locator: () => ({ focus: async () => {} }), waitForFunction: async () => {},
          evaluate: async (fn) => fn.toString().includes("__manualBudgetActivations") ? activations : undefined,
          keyboard: { press: async () => { activations++; emit(true); if (activations === 1) emit(false); } },
        };
      },
    };
    let failure;
    try { await observeManualBudget(browser, "http://fixture"); }
    catch (error) { failure = error.message; }
    return { reports, readParses, failure };
  } finally {
    Object.defineProperty(globalThis, "performance", performanceDescriptor);
    globalThis.setTimeout = originalTimeout;
    console.info = originalInfo;
  }
}

it("retains complete controlled capture profiles without claiming DTO validation", async () => {
  const result = await captureObservation("healthy");
  expect(result.failure).toBeUndefined();
  expect(result.reports.map((report) => [report.hz, report.valid, report.attempts])).toEqual([[2, true, 120], [4, true, 240]]);
  expect(JSON.stringify(result.reports)).not.toContain("synthetic-private-body");
});
for (const scenario of ["malformed-read", "failed-read", "invalid-utf8-read"]) it(`rejects ${scenario} through actual nonmanual response capture`, async () => {
  const result = await captureObservation(scenario);
  expect(result.readParses).toBe(1);
  expect(result.failure).toMatch(/invalid manual account-budget observation/);
  expect(result.reports).toHaveLength(1);
  expect(result.reports[0].valid).toBe(false);
  expect(result.reports[0].errors).toContain("incomplete/invalid response population");
  expect(JSON.stringify(result.reports)).not.toContain("synthetic-private-body");
});
it("retains malformed manual and rate-limit response refusal without repairing UTF-8", async () => {
  for (const scenario of ["malformed-manual", "malformed-rate-limit", "invalid-utf8-manual"]) {
    const result = await captureObservation(scenario);
    expect(result.failure).toMatch(/invalid manual account-budget observation/);
    expect(result.reports[0].valid).toBe(false);
  }
});
it("matches the existing consumer's BOM tolerance for a valid read", async () => {
  const result = await captureObservation("bom-read");
  expect(result.failure).toBeUndefined();
  expect(result.reports.every((report) => report.valid)).toBe(true);
});
it("retains valid limiter findings and refuses unexpected empty statuses without parsing them", async () => {
  const limited = await captureObservation("rate-limited");
  expect(limited.failure).toBeUndefined();
  expect(limited.reports.every((report) => report.valid && report.first_429?.route === "/api/v1/founder/state")).toBe(true);
  const empty = await captureObservation("unexpected-empty");
  expect(empty.readParses).toBe(0);
  expect(empty.failure).toMatch(/invalid manual account-budget observation/);
});
it("keeps the manual observation separate from the default journey and rejects unknown flags", () => {
  expect(composedMode([])).toBe("journey");
  expect(composedMode(["--observe-manual-budget"])).toBe("manual-budget");
  expect(() => composedMode(["--observe-manual-budge"])).toThrow(/unknown composed/);
  expect(() => composedMode(["--observe-manual-budget", "--extra"])).toThrow(/unknown composed/);
});
it("does not complete when a timer returns before its actual deadline", async () => {
  let now = 0;
  const delays = [];
  await waitUntilDeadline(10, () => now, async (delay) => { delays.push(delay); now += delay === 10 ? 9 : delay; });
  expect(delays).toEqual([10, 1]); expect(now).toBe(10);
  await expect(waitUntilDeadline(10, () => NaN)).rejects.toThrow(/invalid observation clock/);
});
it("reports limiter refusals as measured negative results, not invalid or healthy profiles", () => {
  const input = sample(); input.requests[50].status = 429; input.requests[50].outcome = "rate_limited/account";
  const result = summarizeManualBudget(input);
  expect(result.valid).toBe(true);
  expect(result.first_429).toEqual({ at_ms: 25_000, route: "/api/v1/intents" });
  expect(result.counts["POST /api/v1/intents 429 rate_limited/account"]).toBe(1);
});
it("identifies only the exact stale pair without exposing arbitrary receipt data", () => {
  expect(observedRejectionPair({ rejection: { category: "revision_conflict", detail: "expected_revision" } })).toBe("revision_conflict/expected_revision");
  expect(observedRejectionPair({ rejection: { category: "revision_conflict", detail: "private-data" } })).toBe("other");
  expect(observedRejectionPair({ rejection: { category: "not_eligible", detail: "expected_revision" } })).toBe("other");
  expect(observedRejectionPair({})).toBe("other");
  const input = sample(); input.requests[0].outcome = "rejected"; input.requests[0].rejectionPair = "revision_conflict/expected_revision";
  expect(summarizeManualBudget(input).counts["POST /api/v1/intents 200 rejected/revision_conflict/expected_revision"]).toBe(1);
});
for (const [name, change] of [
  ["short duration", (input) => { input.elapsedMS = 59_999; }],
  ["truncated inputs", (input) => { input.attempts = 119; }],
  ["late cadence", (input) => { input.latenessMS = 101; }],
  ["negative lateness", (input) => { input.latenessMS = -1; }],
  ["silent activation loss", (input) => { input.activations = 121; }],
  ["missing dependency", (input) => { input.activations = 0; input.requests = []; }],
  ["unfinished request", (input) => { input.requests[0].status = "pending"; }],
  ["malformed response", (input) => { input.requests[0].invalidBody = true; }],
  ["nonfinite timing", (input) => { input.elapsedMS = NaN; }],
  ["wrong rate", (input) => { input.hz = 3; }],
  ["unclassified outcome", (input) => { delete input.requests[0].outcome; }],
  ["unexpected auth failure", (input) => { input.requests[0].status = 401; }],
  ["page error", (input) => { input.pageErrors = 1; }],
]) it(`rejects ${name}`, () => {
  const input = sample(); change(input); expect(summarizeManualBudget(input).valid).toBe(false);
});
