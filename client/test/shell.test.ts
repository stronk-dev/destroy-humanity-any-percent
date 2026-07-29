import Decimal from "break_infinity.js";
import { describe, expect, it, vi } from "vitest";
import policySource from "../../balance/client-shell/phase0.json";
import { validateReceipt, validateSnapshot, type AuthoritativeSnapshot, type IntentReceipt, type SnapshotStream } from "../src/shell/contracts";
import { ShellController } from "../src/shell/controller";
import { DisplayCounter } from "../src/shell/display";
import { IntentDispatcher } from "../src/shell/intents";
import { bindLifecycle, selectReturnStory } from "../src/shell/lifecycle";
import { parseClientShellPolicy } from "../src/shell/policy";
import { PredictionMachine } from "../src/shell/prediction";
import { reconcileContinuous, reconcileDiscrete } from "../src/shell/reconciliation";
import { ShellRuntime, type PredictionPort } from "../src/shell/runtime";
import type { WorkerOutput } from "../src/shell/worker-protocol";

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
    expect(() => validateSnapshot({ ...snapshot(), surprise: true } as AuthoritativeSnapshot)).toThrow(/fields are not exact/);
    expect(validateReceipt({ revision: 2, intentId: "018f6b7c-9abc-7def-8abc-111111111111", status: "rejected", rejectionCode: "insufficient_resource" }).status).toBe("rejected");
    expect(() => validateReceipt({ revision: 2, intentId: "018f6b7c-9abc-7def-8abc-111111111111", status: "rejected" } as IntentReceipt)).toThrow();
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
  it("is invariant to pulse partitioning because prediction uses the shared closed form", () => {
    const partitioned = new PredictionMachine(policy); partitioned.initialize(snapshot("9.87256122677e5", "1.23456789012e3"), 0); partitioned.pulse(50); partitioned.pulse(100);
    const single = new PredictionMachine(policy); single.initialize(snapshot("9.87256122677e5", "1.23456789012e3"), 0); single.pulse(100);
    expect(partitioned.snapshot(100)).toEqual(single.snapshot(100));
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
  it("uses discrete authoritative steps and receipts for reduced motion", () => {
    const counter = new DisplayCounter({ amount: "1e2", ratePerSecond: "1e0" }, policy, true);
    counter.applyPrediction({ mantissa: 1.005, exponent: 2 });
    expect(counter.applyAuthoritative({ amount: "1e2", ratePerSecond: "1e0" }, 0)).toEqual({ mode: "interpolate", durationMs: 400 });
    expect(counter.view(0)).toMatchObject({ value: "100", pulse: false });
    const receipt = { revision: 2, intentId: "018f6b7c-9abc-7def-8abc-111111111111", status: "rejected", rejectionCode: "unaffordable" } as const;
    expect(counter.applyAuthoritative({ amount: "1e2", ratePerSecond: "1e0" }, 1, receipt)).toEqual({ mode: "rebase", explanation: "unaffordable" });
    expect(counter.view(500)).toMatchObject({ explanation: "unaffordable", pulse: false });
  });
});

it("always dispatches intents to the authoritative adapter", async () => {
  const request = vi.fn(async () => {}); const dispatcher = new IntentDispatcher({ request });
  const intent = { intentId: "018f6b7c-9abc-7def-8abc-111111111111", kind: "buy_generator", expectedRevision: 1, generatorId: "generator.example", count: { mode: "exact", value: 1 } } as const;
  await dispatcher.send(intent); expect(request).toHaveBeenCalledExactlyOnceWith(intent);
  expect(() => dispatcher.send({ ...intent, predictedAffordable: false } as typeof intent)).toThrow(/fields are not exact/);
  const wind = { intentId: "018f6b7c-9abc-7def-8abc-222222222222", kind: "wind_down", expectedRevision: 4, expectedFounderRevision: 2 } as const;
  await dispatcher.send(wind); expect(request).toHaveBeenLastCalledWith(wind);
  expect(() => dispatcher.send({ ...wind, expectedFounderRevision: 0 })).toThrow(/positive integer/);
});

it("selects the one-modal return story only beyond thirty seconds", () => {
  expect(selectReturnStory(30_000, true, policy)).toBeUndefined();
  expect(selectReturnStory(600_000, true, policy)).toEqual({ gapMs: 600000, fastForwardMs: 5000, gainsInHeader: true, showOptionalModal: true, badgesOnly: true });
});

it("measures a throttled hidden interval and requests authority on visibility return", () => {
  let now = 0; let visibilityState: DocumentVisibilityState = "visible";
  const documentTarget = new EventTarget(); Object.defineProperty(documentTarget, "visibilityState", { get: () => visibilityState });
  const windowTarget = new EventTarget(); const requestSnapshot = vi.fn(); const flush = vi.fn(); const dispose = vi.fn(); const gaps: number[] = [];
  const stream = { subscribe: () => () => {}, requestSnapshot, flush, dispose } satisfies SnapshotStream;
  const unbind = bindLifecycle(documentTarget as Document, windowTarget as Window, stream, dispose, (gap) => gaps.push(gap), () => now);
  visibilityState = "hidden"; documentTarget.dispatchEvent(new Event("visibilitychange")); now = 600_000;
  visibilityState = "visible"; documentTarget.dispatchEvent(new Event("visibilitychange"));
  windowTarget.dispatchEvent(new Event("pagehide")); documentTarget.dispatchEvent(new Event("freeze"));
  expect(gaps).toEqual([600_000]); expect(requestSnapshot).toHaveBeenCalledOnce(); expect(flush).toHaveBeenCalledOnce(); expect(dispose).toHaveBeenCalledTimes(2);
  unbind();
});

it("wires a ten-minute worker gap to one authoritative request and a return story", () => {
  let consumeSnapshot: ((snapshot: AuthoritativeSnapshot, receipt?: IntentReceipt) => void) | undefined;
  let consumeWorker: ((output: WorkerOutput) => void) | undefined;
  let requests = 0;
  const stream: SnapshotStream = {
    subscribe(consumer) { consumeSnapshot = consumer; return () => { consumeSnapshot = undefined; }; },
    requestSnapshot() { requests++; }, flush() {}, dispose() {},
  };
  const port: PredictionPort = { authoritative: vi.fn(), dispose: vi.fn() };
  const controller = new ShellController(policy);
  const runtime = new ShellRuntime(controller, policy, stream, (_policy, _snapshot, consume) => { consumeWorker = consume; return port; });
  runtime.start(undefined, undefined);
  expect(requests).toBe(1);
  consumeSnapshot?.(snapshot());
  consumeWorker?.({ kind: "offline_required", gapMs: 600_000 });
  expect(requests).toBe(2);
  consumeSnapshot?.({ ...snapshot("7e2"), revision: 2, evaluatedThroughMs: 601_000 });
  const view = controller.view(0);
  expect(view.returnStory?.gapMs).toBe(600_000);
  expect(view.resources["company.cash"]).toMatchObject({ value: "700", pulse: false });
  expect(view.offlineGains).toEqual({ "company.cash": "600" });
  controller.completeReturnFastForward();
  expect(controller.view(0).returnFastForwardComplete).toBe(true);
  runtime.dispose();
});

it("records aggregate-only reconciliation and worker telemetry", () => {
  const controller = new ShellController(policy);
  controller.applyAuthoritative(snapshot(), undefined, 0);
  controller.applyPrediction({ revision: 1, atMonotonicMs: 100, resources: { "company.cash": { mantissa: 1.02, exponent: 2 } } }, 100);
  controller.applyAuthoritative({ ...snapshot(), revision: 2, discrete: { "company.tier": 1, "unlock.shop": true } }, undefined, 100);
  controller.applyAuthoritative({ ...snapshot(), revision: 3 }, { revision: 3, intentId: "018f6b7c-9abc-7def-8abc-111111111111", status: "rejected", rejectionCode: "insufficient_resource" }, 100);
  controller.recordWorker({ kind: "worker_metric", name: "tick_overrun", value: 50 });
  expect(controller.telemetry()).toEqual({ epsilonLerpBreaches: 1, rejectionRebases: { insufficient_resource: 1 }, discreteRebases: 4, workerTickOverruns: 1, workerTickOverrunMs: 50 });
  controller.dispose();
});
