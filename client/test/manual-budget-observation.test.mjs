import { expect, it } from "vitest";
import { composedMode, observedRejectionPair, summarizeManualBudget, waitUntilDeadline } from "../tools/observe-manual-budget.mjs";

const sample = () => ({ hz: 2, elapsedMS: 60_000, latenessMS: 20, attempts: 120, activations: 120,
  requests: Array.from({ length: 120 }, (_, index) => ({ manual: true, method: "POST", route: "/api/v1/intents", status: 200, outcome: "applied", at_ms: index * 500 })), pageErrors: 0 });
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
