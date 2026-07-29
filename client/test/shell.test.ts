import Decimal from "break_infinity.js";
import { describe, expect, it, vi } from "vitest";
import policySource from "../../balance/client-shell/phase0.json";
import { validateSnapshot, type AuthoritativeSnapshot } from "../src/shell/contracts";
import { DisplayCounter } from "../src/shell/display";
import { IntentDispatcher } from "../src/shell/intents";
import { selectReturnStory } from "../src/shell/lifecycle";
import { parseClientShellPolicy } from "../src/shell/policy";
import { PredictionMachine } from "../src/shell/prediction";
import { reconcileContinuous, reconcileDiscrete } from "../src/shell/reconciliation";

const policy = parseClientShellPolicy(policySource);
const snapshot = (amount = "1e2", rate = "1e0"): AuthoritativeSnapshot => ({
  revision: 1, evaluatedThroughMs: 1_000, constantsHash: `sha256:${"a".repeat(64)}`,
  resources: { "company.cash": { amount, ratePerSecond: rate, cap: { amount: "1e6", reasonKey: "cap.phase0_cash" } } },
  discrete: { "company.tier": 0, "unlock.shop": false }, progress: [{ stageId: "tier.bootstrap", current: "1e0", target: "1e1" }],
});

describe("client shell policy and snapshot boundary", () => {
  it("loads the strict shipped policy", () => expect(policy).toEqual({ tickMs: 50, snapshotMs: 100, catchupCeilingMs: 5000, epsilonLerpPpm: 10000, lerpDurationMs: 400, reconnectStoryThresholdMs: 30000, reducedMotionRenderMs: 500 }));
  it("rejects unknown policy fields", () => expect(() => parseClientShellPolicy({ ...policySource, surprise: true })).toThrow());
  it("accepts canonical state and rejects unexplained caps", () => {
    expect(validateSnapshot(snapshot()).revision).toBe(1);
    expect(() => validateSnapshot({ ...snapshot(), resources: { "company.cash": { amount: "1e2", ratePerSecond: "1e0", cap: { amount: "1e6", reasonKey: "" } } } })).toThrow();
  });
});

describe("fixed-step prediction", () => {
  it("uses wall-clock accumulation and emits at the configured cadence", () => {
    const machine = new PredictionMachine(policy); machine.initialize(snapshot(), 10);
    expect(machine.pulse(60)).toEqual([]);
    const output = machine.pulse(110);
    expect(output[0]?.kind).toBe("predicted_snapshot");
    if (output[0]?.kind === "predicted_snapshot") expect(Decimal.fromMantissaExponent(output[0].snapshot.resources["company.cash"].mantissa, output[0].snapshot.resources["company.cash"].exponent).eq(100.1)).toBe(true);
  });
  it("hands long gaps to offline evaluation without simulating them", () => {
    const machine = new PredictionMachine(policy); machine.initialize(snapshot(), 0);
    expect(machine.pulse(6000)).toEqual([{ kind: "offline_required", gapMs: 6000 }]);
    expect(machine.snapshot(6000).resources["company.cash"]).toEqual({ mantissa: 1, exponent: 2 });
  });
});

describe("reconciliation and display", () => {
  it("covers exact, bend, snap, rejected, and discrete rows", () => {
    expect(reconcileContinuous(new Decimal(100), "1e2", policy)).toEqual({ mode: "none" });
    expect(reconcileContinuous(new Decimal(100.5), "1e2", policy)).toEqual({ mode: "interpolate", durationMs: 400 });
    expect(reconcileContinuous(new Decimal(102), "1e2", policy)).toEqual({ mode: "rebase" });
    expect(reconcileContinuous(new Decimal(100), "1e2", policy, { revision: 2, intentId: "018f6b7c-9abc-7def-8abc-111111111111", status: "rejected", rejectionCode: "insufficient_resource" })).toEqual({ mode: "rebase", explanation: "insufficient_resource" });
    expect(reconcileDiscrete(false, true)).toEqual({ mode: "rebase" });
  });
  it("keeps a producing pathological counter visibly active and explains caps", () => {
    const counter = new DisplayCounter({ amount: "1e20", ratePerSecond: "1e7" }, policy);
    const predicted = new Decimal("1e20").add("1e7"); counter.applyPrediction({ mantissa: predicted.mantissa, exponent: predicted.exponent });
    expect(counter.view(100).activityPpm).toBeGreaterThan(0);
    const capped = new DisplayCounter({ amount: "1e6", ratePerSecond: "1e0", cap: { amount: "1e6", reasonKey: "cap.phase0_cash" } }, policy);
    expect(capped.view(0).capReasonKey).toBe("cap.phase0_cash");
  });
  it("steps rather than interpolates in reduced-motion mode", () => {
    const counter = new DisplayCounter({ amount: "1e2", ratePerSecond: "1e0" }, policy, true);
    counter.applyPrediction({ mantissa: 1.01, exponent: 2 });
    expect(counter.view(0).value).toBe(counter.view(499).value);
    expect(counter.view(500).value).toBe("101");
  });
});

it("always dispatches intents to the authoritative adapter", async () => {
  const request = vi.fn(async () => {}); const dispatcher = new IntentDispatcher({ request });
  const intent = { intentId: "018f6b7c-9abc-7def-8abc-111111111111", kind: "buy_generator", expectedRevision: 1 };
  await dispatcher.send(intent); expect(request).toHaveBeenCalledExactlyOnceWith(intent);
});

it("selects the one-modal return story only beyond thirty seconds", () => {
  expect(selectReturnStory(30_000, true, policy)).toBeUndefined();
  expect(selectReturnStory(600_000, true, policy)).toEqual({ fastForwardMs: 5000, gainsInHeader: true, showRipeQuarter: true, badgesOnly: true });
});
