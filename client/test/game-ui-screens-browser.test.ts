import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it, vi } from "vitest";

import GameUIApp from "../src/game-ui/GameUIApp.svelte";
import RunEndSurface from "../src/game-ui/RunEndSurface.svelte";
import type { ExitOfferSpawnedEvent, GateCrossedEvent, RunEndedEvent } from "../src/game-ui/events";
import type { GameUIRuntime, GameUIRuntimeMessage } from "../src/game-ui/runtime";
import type { IntentOutcome } from "../src/game-ui/intent-outcome";
import type { GameUISnapshot } from "../src/api/generated/types";
import { eraForSnapshot, parseGameUISnapshot, type ParsedGameUISnapshot } from "../src/game-ui/contracts";
import { FEATURES_PRESENTATION } from "../src/game-ui/features-presentation";
import { t, type CopyKey } from "../src/copy";
import { REQUIRED_METER_IDS } from "../src/meters/catalog";
import { canonicalString } from "../src/numeric";
import { GAME_UI_PERFORMANCE_BUDGET, validatePerformanceObservation } from "../src/game-ui/performance";
import { amountRenderScheduler } from "../src/ui/render-scheduler";
import { formatAmount } from "../src/ui/amount-format";
import type { WorkerCommand, WorkerOutput } from "../src/shell/worker-protocol";

const snapshot: GameUISnapshot = {
  constants_hash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  evaluated_through_ms: 1_800_000_000_000,
  facts: [{ fact_id: "bootstrap.needed", value: false }, { fact_id: "gate.t0_to_t1", value: false }],
  founder_revision: 1,
  generators: [{ generator_id: "generator.beige_tower", max_affordable: 2, next_cost: "1e1", next_cost_resource_id: "company.cash", owned: 1, provision_cap: null, provisioned: 0, rate_contribution: "1e0" }],
  manual_action: { action_id: "manual.click", bucket_cap_milli: 50_000, refill_milli_per_ms: 25, refilled_at_ms: 1_800_000_000_000, tokens_milli: 50_000 },
  progress: [{ current: "5e-1", stage_id: "progress.tier", target: "1e0" }],
  resources: [{ amount: "1e2", cap: { amount: "1e1000", reason_key: "resource.company_cash.cap.phase0" }, rate_per_second: "1e0", resource_id: "company.cash" }],
  revision: 1,
  run: { category: "any_percent", exit_count: 0, founder_id: "01985555-1111-7111-8111-111111111111", run_seq: 1, run_started_at_ms: 1_799_999_000_000, tier: 0 },
  features: { achievements: null, active_play: null, fiscal: null, meters: null, minigames: null, pets: null },
  schema_version: 4, server_now_ms: 1_800_000_000_000,
  transitions: { cross_gate: { eligible: true, gate_id: "gate.t0_to_t1", route_id: null }, wind_down: { eligible: false } },
  upgrades: [{ cost_amount: "2e1", cost_resource_id: "company.cash", eligible: true, owned: false, upgrade_id: "upgrade.beige_tower_cache" }],
};

const offer: ExitOfferSpawnedEvent = { cursor: 2, kind: "exit_offer_spawned", occurred_at_ms: 1_800_000_000_000, payload: {
  exit_type: "scripted_first", expires_at_ms: 1_800_000_060_000, offer_id: "01985555-3333-7333-8333-333333333333",
  payout_preview: { clout_reach_note: "clout.reach.preserved", network_slot_unlocks: [], reputation_delta: 1, route_knowledge: 2 },
} };
const ended: RunEndedEvent = { cursor: 3, kind: "run_ended", occurred_at_ms: 1_800_000_001_000, payload: {
  assisted: { advisor: false, commons: false }, attended_ms: 500, ended_at_ms: 1_800_000_001_000, executed_routes: [], exit_type: "scripted_first", faction: null,
  founder_id: snapshot.run.founder_id, gates_crossed: ["gate.t0_to_t1"], generators_purchased_total: 1, ledger_fact_kinds: [], lifetime_value: "1e3",
  payout: { clout_reach_note: "clout.reach.preserved", network_slot_unlocks: [], reputation_delta: 2, route_knowledge: 25 }, pre_timer: false, rta_ms: 1_000,
  run_id: { company_stream_id: "01985555-2222-7222-8222-222222222222", run_seq: 1 }, started_at_ms: 1_800_000_000_000, terminal_seq: 2, tier: 0,
} };
const crossed: GateCrossedEvent = { cursor: 2, kind: "gate_crossed", occurred_at_ms: snapshot.run.run_started_at_ms + 750, payload: {
  founder_id: snapshot.run.founder_id, gate_id: "gate.t0_to_t1", route_id: null,
  run_id: { company_stream_id: ended.payload.run_id.company_stream_id, run_seq: 1 },
} };

class FixtureRuntime implements GameUIRuntime {
  readonly requests: Readonly<Record<string, unknown>>[] = [];
  listener: ((message: GameUIRuntimeMessage) => void) | undefined;
  current: ParsedGameUISnapshot = snapshot;
  snapshotCalls = 0;
  failSnapshot = false;
  intentBlock: Promise<void> | undefined;
  constructor(private authenticated = false, private recoverTransport = true) {}
  hasCredentials(): boolean { return this.authenticated; }
  async bootstrap(): Promise<GameUISnapshot> { this.authenticated = true; return snapshot; }
  async snapshot(): Promise<ParsedGameUISnapshot> { this.snapshotCalls += 1; if (this.failSnapshot) throw new Error("offline"); return this.current; }
  intentOutcome: IntentOutcome = { outcome: "applied", receipt: {} };
  async intent(body: Readonly<Record<string, unknown>>): Promise<IntentOutcome> { this.requests.push(body); await this.intentBlock; return this.intentOutcome; }
  subscribe(_founderID: string, listener: (message: GameUIRuntimeMessage) => void): () => void {
    this.listener = listener;
    if (this.recoverTransport) queueMicrotask(() => { if (this.listener === listener) listener({ kind: "transport_recovered" }); });
    return () => { this.listener = undefined; };
  }
}

interface AppExports {
  fixtureSnapshot(value: ParsedGameUISnapshot): void;
  fixtureSurface(value: "desk" | "offer_sheet" | "run_end" | "settings" | "vision_slide"): void;
  fixtureOffer(value: ExitOfferSpawnedEvent): void;
  fixtureRunEnd(value: RunEndedEvent): void;
  fixtureSystem(value: "drain" | "resync"): void;
  fixtureMonotonicElapsed(value: number): void;
}

// Observation only: every command reaches the original native Worker unchanged.
function observeNativePrediction() {
  const started = performance.now();
  let firstPredictionMS: number | undefined;
  const commands: { kind: string; rate?: string }[] = [];
  const outputs: WorkerOutput[] = [];
  const workers = new Set<Worker>();
  const lifecycle: { worker: number; phase: string; at_ms: number; detail?: string; terminated?: boolean }[] = [];
  const observers = new Map<Worker, { id: number; terminated: boolean; receive: (event: MessageEvent<WorkerOutput>) => void; error: (event: ErrorEvent) => void }>();
  const receive = (event: MessageEvent<WorkerOutput>) => {
    outputs.push(event.data);
    if (event.data.kind === "predicted_snapshot" && firstPredictionMS === undefined) firstPredictionMS = performance.now() - started;
  };
  const original = Worker.prototype.postMessage;
  const spy = vi.spyOn(Worker.prototype, "postMessage").mockImplementation(function (this: Worker, ...args) {
    if (!workers.has(this)) {
      workers.add(this);
      const owner = { id: workers.size, terminated: false };
      const observeMessage = (event: MessageEvent<WorkerOutput>) => {
        lifecycle.push({ worker: owner.id, phase: "output", at_ms: performance.now() - started, detail: event.data.kind, terminated: owner.terminated });
        receive(event);
      };
      const observeError = (event: ErrorEvent) => {
        lifecycle.push({ worker: owner.id, phase: "error", at_ms: performance.now() - started, detail: event.message || "empty native error message", terminated: owner.terminated });
      };
      observers.set(this, Object.assign(owner, { receive: observeMessage, error: observeError }));
      this.addEventListener("message", observeMessage); this.addEventListener("error", observeError);
    }
    const command = args[0] as WorkerCommand;
    lifecycle.push({ worker: observers.get(this)!.id, phase: "command", at_ms: performance.now() - started, detail: command.kind });
    commands.push({ kind: command.kind, rate: "snapshot" in command ? command.snapshot.resources["company.cash"]?.ratePerSecond : undefined });
    return original.apply(this, args);
  });
  const originalTerminate = Worker.prototype.terminate;
  const terminationSpy = vi.spyOn(Worker.prototype, "terminate").mockImplementation(function (this: Worker) {
    const owner = observers.get(this);
    if (owner) { owner.terminated = true; lifecycle.push({ worker: owner.id, phase: "terminate", at_ms: performance.now() - started }); }
    return originalTerminate.call(this);
  });
  return {
    commands, outputs,
    report(runtime: FixtureRuntime, output: Element | null) {
      const predicted = outputs.filter((row) => row.kind === "predicted_snapshot");
      return { userAgent: navigator.userAgent, visibility: document.visibilityState, firstPredictionMS, snapshotCalls: runtime.snapshotCalls, commands,
        output: output?.textContent, predictions: predicted.length,
        lastPredictions: predicted.slice(-8), gaps: outputs.filter((row) => row.kind === "offline_required"), lifecycle };
    },
    dispose() {
      for (const [worker, owner] of observers) { worker.removeEventListener("message", owner.receive); worker.removeEventListener("error", owner.error); }
      terminationSpy.mockRestore(); spy.mockRestore();
    },
  };
}

async function assertAxe(target: HTMLElement, label: string): Promise<void> {
  const result = await axe.run(target, { runOnly: { type: "tag", values: ["wcag2a", "wcag2aa", "wcag21aa", "wcag22aa"] } });
  expect(result.violations.filter((violation) => violation.impact === "serious" || violation.impact === "critical"), label).toEqual([]);
}

function assertNoMechanicalPresentation(target: HTMLElement): void {
  for (const id of ["generator.beige_tower", "upgrade.beige_tower_cache", "manual.click", "gate.t0_to_t1", "scripted_first"]) {
    expect(target.textContent).not.toContain(id);
  }
}

it.skipIf(typeof document === "undefined")("runs bootstrap and player actions through the mounted Phase-A UI", async () => {
  const { userEvent } = await import("vitest/browser");
  const runtime = new FixtureRuntime(false);
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime } }) as unknown as AppExports;
  flushSync();
  expect(target.querySelector("main")?.dataset.surface).toBe("vision_slide");
  const begin = target.querySelector("button") as HTMLButtonElement;
  begin.focus();
  expect(document.activeElement).toBe(begin);
  expect(getComputedStyle(begin).outlineStyle).not.toBe("none");
  await userEvent.keyboard("{Enter}");
  await tick(); await new Promise((resolve) => setTimeout(resolve, 0)); flushSync();
  expect(target.querySelector("main")?.dataset.surface).toBe("desk");
  expect(target.textContent).toContain("Beige Tower");
  expect(target.textContent).not.toContain("generator.beige_tower");
  const manual = target.querySelector(".manual button") as HTMLButtonElement;
  manual.click(); await new Promise((resolve) => setTimeout(resolve, 0)); flushSync();
  expect(runtime.requests[0]).toMatchObject({ kind: "perform_manual_batch", action_id: "manual.click", count: 1 });
  await unmount(app); target.remove();
});

it.skipIf(typeof document === "undefined")("renders only server-projected transition controls and submits their existing intents", async () => {
  const runtime = new FixtureRuntime(true);
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime } }) as unknown as AppExports;
  await new Promise((resolve) => setTimeout(resolve, 0)); flushSync();
  const crossGate = [...target.querySelectorAll("button")].find((button) => button.textContent === "Move Into the Garage")!;
  const windDown = [...target.querySelectorAll("button")].find((button) => button.textContent === "Wind Down Company")!;
  expect(crossGate.disabled).toBe(false);
  expect(windDown.disabled).toBe(true);
  const tierOne = { ...snapshot, revision: 2, run: { ...snapshot.run, tier: 1 }, transitions: { cross_gate: null, wind_down: { eligible: true } } };
  runtime.current = tierOne;
  runtime.snapshotCalls = 0;
  crossGate.click();
  await new Promise((resolve) => setTimeout(resolve, 0));
  await new Promise((resolve) => setTimeout(resolve, 0));
  flushSync();
  expect(runtime.requests[0]).toMatchObject({ kind: "cross_gate", expected_revision: 1, gate_id: "gate.t0_to_t1", route_id: null });
  expect(runtime.snapshotCalls).toBe(1);
  expect([...target.querySelectorAll("button")].some((button) => button.textContent === "Move Into the Garage")).toBe(false);
  const eligibleWindDown = [...target.querySelectorAll("button")].find((button) => button.textContent === "Wind Down Company")!;
  expect(eligibleWindDown.disabled).toBe(false);
  eligibleWindDown.click(); await new Promise((resolve) => setTimeout(resolve, 0)); flushSync();
  expect(runtime.requests[1]).toMatchObject({ kind: "wind_down", expected_revision: 2, expected_founder_revision: 1 });

  const { transitions: _transitions, ...v2 } = snapshot;
  app.fixtureSnapshot({ ...v2, schema_version: 2 }); flushSync();
  expect(target.textContent).not.toContain("Move Into the Garage");
  expect(target.textContent).not.toContain("Wind Down Company");
  await unmount(app); target.remove();
});

it.skipIf(typeof document === "undefined")("serializes a distinct transition click that races the preceding action", async () => {
  const runtime = new FixtureRuntime(true);
  let releaseIntent = () => {};
  runtime.intentBlock = new Promise<void>((resolve) => { releaseIntent = resolve; });
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime } }) as unknown as AppExports;
  await new Promise((resolve) => setTimeout(resolve, 0)); flushSync();
  const crossGate = [...target.querySelectorAll("button")].find((button) => button.textContent === "Move Into the Garage")!;
  crossGate.click(); await new Promise((resolve) => setTimeout(resolve, 0)); flushSync();
  expect(runtime.requests).toHaveLength(1);

  runtime.current = { ...snapshot, revision: 2, run: { ...snapshot.run, tier: 1 }, transitions: { cross_gate: null, wind_down: { eligible: true } } };
  app.fixtureSnapshot(runtime.current);
  flushSync();
  const windDown = [...target.querySelectorAll("button")].find((button) => button.textContent === "Wind Down Company")!;
  windDown.dispatchEvent(new MouseEvent("click", { bubbles: true }));
  releaseIntent();
  await new Promise((resolve) => setTimeout(resolve, 0));
  await new Promise((resolve) => setTimeout(resolve, 0));
  flushSync();
  expect(runtime.requests).toHaveLength(2);
  expect(runtime.requests[1]).toMatchObject({ kind: "wind_down", expected_revision: 2 });
  await unmount(app); target.remove();
});

// Real mounted host/native activation; the runtime double controls the order
// of HTTP completion versus a read started before or after the action commits.
for (const [kind, beforeCommit] of [
  ["cross_gate", false], ["cross_gate", true],
  ["decline_exit_offer", false], ["decline_exit_offer", true],
] as const) {
  const name = beforeCommit ? `refreshes ${kind} again when the stream read predates the committed action`
    : `coalesces ${kind} HTTP completion with the pending stream refresh`;
  it.skipIf(typeof document === "undefined")(`${name} (GS0.2)`, async () => {
    const { userEvent } = await import("vitest/browser");
    const releaseReads: (() => void)[] = [];
    class DelayedReadRuntime extends FixtureRuntime {
      override async snapshot(): Promise<ParsedGameUISnapshot> {
        if (this.requests.length === 0) return super.snapshot();
        this.snapshotCalls += 1;
        const value = this.current;
        await new Promise<void>((resolve) => releaseReads.push(resolve));
        return value;
      }
    }
    const runtime = new DelayedReadRuntime(true);
    let releaseIntent = () => {};
    runtime.intentBlock = new Promise<void>((resolve) => { releaseIntent = resolve; });
    const target = document.createElement("div"); document.body.append(target);
    const app = mount(GameUIApp, { target, props: { runtime } }) as unknown as AppExports;
    const settle = async () => {
      for (let step = 0; step < 4; step += 1) { await new Promise((resolve) => setTimeout(resolve, 0)); await tick(); flushSync(); }
    };
    try {
      await settle();
      if (kind === "decline_exit_offer") { app.fixtureOffer(offer); await settle(); }
      const label = kind === "cross_gate" ? "Move Into the Garage" : "Decline";
      const control = [...target.querySelectorAll("button")].find((button) => button.textContent === label)!;
      expect(control.disabled).toBe(false);
      control.focus(); await userEvent.keyboard("{Enter}"); await settle();
      expect(runtime.requests).toHaveLength(1);
      expect(runtime.requests[0]).toMatchObject({ kind, expected_revision: 1 });

      const committed = { ...snapshot, revision: 2, run: { ...snapshot.run, tier: 1 },
        transitions: { cross_gate: null, wind_down: { eligible: true } } };
      runtime.intentOutcome = { outcome: "applied", receipt: { new_revision: 2 } };
      if (!beforeCommit) runtime.current = committed;
      runtime.snapshotCalls = 0;
      if (kind === "cross_gate" && !beforeCommit) runtime.listener?.({ kind: "event", revision: 2, scope: "company", value: crossed });
      else runtime.listener?.({ kind: "receipt" });
      await settle();
      expect(runtime.snapshotCalls).toBe(1);
      runtime.current = committed;
      releaseIntent(); await settle();
      expect(runtime.snapshotCalls, "HTTP completion must reuse the pending authoritative read").toBe(1);
      expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("true");
      releaseReads.splice(0).forEach((release) => release()); await settle();
      if (beforeCommit) {
        expect(runtime.snapshotCalls, "a pre-commit result cannot supply the applied revision").toBe(2);
        expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("true");
        releaseReads.splice(0).forEach((release) => release()); await settle();
      }
      expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("false");
      expect(runtime.requests).toHaveLength(1);
      app.fixtureSurface("desk"); await settle();
      const windDown = [...target.querySelectorAll("button")].find((button) => button.textContent === "Wind Down Company")!;
      expect(windDown.disabled).toBe(false);
      windDown.focus(); await userEvent.keyboard("{Enter}"); await settle();
      expect(runtime.requests[1]).toMatchObject({ kind: "wind_down", expected_revision: 2, expected_founder_revision: 1 });
    } finally {
      releaseIntent(); releaseReads.splice(0).forEach((release) => release());
      await settle(); await unmount(app as never); target.remove();
    }
  });
}

it.skipIf(typeof document === "undefined")("keeps terminal commands disabled until the ordered event channel is recovered", async () => {
  const runtime = new FixtureRuntime(true, false);
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime } }) as unknown as AppExports;
  await new Promise((resolve) => setTimeout(resolve, 0));
  app.fixtureSnapshot({ ...snapshot, run: { ...snapshot.run, tier: 1 }, transitions: { cross_gate: null, wind_down: { eligible: true } } });
  flushSync();
  const windDown = [...target.querySelectorAll("button")].find((button) => button.textContent === "Wind Down Company")!;
  expect(windDown.disabled).toBe(true);
  runtime.listener?.({ kind: "transport_recovered" }); flushSync();
  expect(windDown.disabled).toBe(false);
  runtime.listener?.({ kind: "transport_closed" }); flushSync();
  expect(windDown.disabled).toBe(true);
  await unmount(app); target.remove();
});

it.skipIf(typeof document === "undefined")("continues from either terminal state only to the exact next Company snapshot", async () => {
  const runtime = new FixtureRuntime(true);
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime } }) as unknown as AppExports;
  await new Promise((resolve) => setTimeout(resolve, 0));
  app.fixtureSnapshot(snapshot); app.fixtureRunEnd(ended); flushSync();
  expect(target.textContent).toContain("Your First Company Failed");
  const next = { ...snapshot, revision: 2, run: { ...snapshot.run, run_seq: 2, tier: 1 }, transitions: { cross_gate: null, wind_down: { eligible: true } } };
  runtime.current = next;
  runtime.snapshotCalls = 0;
  const continuation = [...target.querySelectorAll("button")].find((button) => button.textContent === "Start the Next Company")!;
  continuation.click(); await new Promise((resolve) => setTimeout(resolve, 0)); flushSync();
  expect(runtime.snapshotCalls).toBe(1);
  expect(runtime.requests).toEqual([]);
  expect(target.querySelector("main")?.dataset.surface).toBe("desk");

  const standard = { ...ended, payload: { ...ended.payload, exit_type: "collapse" as const, tier: 1 } };
  app.fixtureRunEnd(standard); runtime.current = snapshot; runtime.snapshotCalls = 0; flushSync();
  expect(target.textContent).toContain("The Company Has Exited");
  const mismatched = [...target.querySelectorAll("button")].find((button) => button.textContent === "Start the Next Company")!;
  mismatched.click(); await new Promise((resolve) => setTimeout(resolve, 0)); flushSync();
  expect(runtime.snapshotCalls).toBe(1);
  expect(target.querySelector("main")?.dataset.surface).toBe("run_end");
  expect(target.textContent).toContain("OFFLINE — progress is parked on this machine until the server picks up again.");
  await unmount(app); target.remove();
});

it.skipIf(typeof document === "undefined")("passes the C11 axe gate on all five Phase-A surfaces and the two system beats", async () => {
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime: new FixtureRuntime(false) } }) as unknown as AppExports;
  flushSync(); assertNoMechanicalPresentation(target); await assertAxe(target, "vision_slide");
  app.fixtureSnapshot(snapshot); app.fixtureSurface("desk"); flushSync(); assertNoMechanicalPresentation(target); await assertAxe(target, "desk");
  app.fixtureOffer(offer); flushSync(); assertNoMechanicalPresentation(target); await assertAxe(target, "offer_sheet");
  app.fixtureRunEnd(ended); flushSync(); assertNoMechanicalPresentation(target); await assertAxe(target, "run_end");
  app.fixtureSurface("settings"); app.fixtureSystem("drain"); app.fixtureSystem("resync"); flushSync(); assertNoMechanicalPresentation(target); await assertAxe(target, "settings/system");
  await unmount(app); target.remove();
});

it.skipIf(typeof document === "undefined")("renders the authoritative cap explanation on the Desk", async () => {
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime: new FixtureRuntime(true) } }) as unknown as AppExports;
  await new Promise((resolve) => setTimeout(resolve, 0));
  app.fixtureSnapshot({ ...snapshot, resources: [{ ...snapshot.resources[0], cap: { ...snapshot.resources[0].cap!, amount: "1e2" } }] });
  app.fixtureSurface("desk"); flushSync();
  expect(target.textContent).toContain("Cash is capped. The cap is a number, the number is visible, and nothing will ever sell you the difference.");
  await unmount(app); target.remove();
});

it.skipIf(typeof document === "undefined")("renders the drain story beat", async () => {
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime: new FixtureRuntime(true) } }) as unknown as AppExports;
  app.fixtureSystem("drain"); flushSync();
  expect(target.textContent).toContain("The server needs a minute. Your equipment keeps working while it's away, and nothing is lost. Back soon!!");
  await unmount(app); target.remove();
});

it.skipIf(typeof document === "undefined")("renders the resync story beat", async () => {
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime: new FixtureRuntime(true) } }) as unknown as AppExports;
  app.fixtureSystem("resync"); flushSync();
  expect(target.textContent).toContain("The two copies of the books disagreed, so everything was recounted. The server's copy wins. It always does. Nothing was lost.");
  await unmount(app); target.remove();
});

it.skipIf(typeof document === "undefined")("renders run-end from the decoded terminal payload without a snapshot input", async () => {
  const target = document.createElement("div"); document.body.append(target);
  const component = mount(RunEndSurface, { target, props: { ended } }); flushSync();
  expect(target.textContent).toContain("First Failure");
  expect(target.textContent).not.toContain(snapshot.run.founder_id);
  await unmount(component); target.remove();
});

it.skipIf(typeof document === "undefined")("records immutable gate timing locally and lets lifecycle events preempt the Desk", async () => {
  const runtime = new FixtureRuntime(true);
  const values = new Map<string, string>();
  const timingStorage = { getItem: (key: string) => values.get(key) ?? null, setItem: (key: string, value: string) => { values.set(key, value); } };
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime, timingStorage } });
  await new Promise((resolve) => setTimeout(resolve, 0)); flushSync();
  runtime.current = { ...snapshot, revision: 2, run: { ...snapshot.run, tier: 1 }, transitions: { cross_gate: null, wind_down: { eligible: true } } };
  runtime.snapshotCalls = 0;
  runtime.listener?.({ kind: "event", revision: 2, scope: "company", value: crossed }); flushSync();
  await new Promise((resolve) => setTimeout(resolve, 0)); flushSync();
  expect(runtime.snapshotCalls).toBe(1);
  expect(target.textContent).toContain("Garage");
  expect([...target.querySelectorAll("button")].find((button) => button.textContent === "Wind Down Company")?.disabled).toBe(false);
  runtime.snapshotCalls = 0;
  // Centrifuge may deliver the ordered terminal event and its trailing receipt
  // synchronously inside one browser message callback, before Svelte flushes.
  runtime.listener?.({ kind: "event", revision: 3, scope: "company", value: ended });
  runtime.current = { ...snapshot, revision: 4, run: { ...snapshot.run, run_seq: 2 } };
  runtime.listener?.({ kind: "receipt" });
  runtime.listener?.({ kind: "event", revision: 4, scope: "company", value: { ...offer, cursor: 4 } });
  await new Promise((resolve) => setTimeout(resolve, 0)); flushSync();
  expect(runtime.snapshotCalls).toBe(0);
  expect(target.querySelector("main")?.dataset.surface).toBe("run_end");
  expect([...values.values()][0]).toContain('"rta_ms":750');
  await unmount(app); target.remove();
});

it.skipIf(typeof document === "undefined")("replays a v1 bootstrap receipt fail-closed, then accepts from a v2 Founder coordinate", async () => {
  const runtime = new FixtureRuntime(true);
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime } }) as unknown as AppExports;
  await new Promise((resolve) => setTimeout(resolve, 0));
  const { founder_revision: _founderRevision, transitions: _transitions, ...legacy } = snapshot;
  app.fixtureSnapshot({ ...legacy, schema_version: 1 }); app.fixtureOffer(offer); flushSync();
  const buttons = [...target.querySelectorAll("button")];
  const sign = buttons.find((button) => button.textContent === "Sign")!;
  const decline = buttons.find((button) => button.textContent === "Decline")!;
  expect(sign.disabled).toBe(true);
  app.fixtureSnapshot(snapshot); flushSync();
  expect(sign.disabled).toBe(false);
  sign.click(); await new Promise((resolve) => setTimeout(resolve, 0)); flushSync();
  expect(runtime.requests[0]).toMatchObject({ kind: "accept_exit_offer", expected_founder_revision: 1, offer_id: offer.payload.offer_id });
  runtime.current = { ...snapshot, revision: 2 };
  runtime.snapshotCalls = 0;
  decline.click();
  await new Promise((resolve) => setTimeout(resolve, 0));
  await new Promise((resolve) => setTimeout(resolve, 0));
  flushSync();
  expect(runtime.requests[1]).toMatchObject({ kind: "decline_exit_offer", offer_id: offer.payload.offer_id });
  expect(runtime.requests[1]).not.toHaveProperty("expected_founder_revision");
  expect(runtime.snapshotCalls).toBe(1);
  await unmount(app); target.remove();
});

it.skipIf(typeof document === "undefined")("renders ruled constants and complete decoded payout terms without mechanical substitutes", async () => {
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime: new FixtureRuntime(true) } }) as unknown as AppExports;
  await new Promise((resolve) => setTimeout(resolve, 0));
  app.fixtureSnapshot(snapshot); app.fixtureSurface("desk"); flushSync();
  expect(target.textContent).toContain("$0.00");
  const order = [...target.querySelectorAll("button")].find((button) => button.textContent === "PLACE ORDER")!;
  order.click(); flushSync();
  expect(target.textContent).toContain("Thank you, Founder!!");
  app.fixtureOffer(offer); flushSync();
  expect(target.textContent).toContain("Clout carries. The personal brand survives the company.");
  expect(target.textContent).toContain("Reputation +1");
  expect(target.textContent).not.toContain("Reputation +1e0");
  expect(target.textContent).toContain("Route Knowledge +2");
  expect(target.textContent).not.toContain("clout.reach.preserved");
  await unmount(app); target.remove();
});

it.skipIf(typeof document === "undefined")("max-advances the Founder CAS coordinate across out-of-order events", async () => {
  const runtime = new FixtureRuntime(true);
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime } }) as unknown as AppExports;
  await new Promise((resolve) => setTimeout(resolve, 0));
  app.fixtureSnapshot(snapshot);
  runtime.listener?.({ kind: "event", revision: 4, scope: "founder", value: crossed });
  runtime.listener?.({ kind: "event", revision: 3, scope: "founder", value: crossed });
  app.fixtureOffer(offer); flushSync();
  const sign = [...target.querySelectorAll("button")].find((button) => button.textContent === "Sign")!;
  sign.click(); await new Promise((resolve) => setTimeout(resolve, 0)); flushSync();
  expect(runtime.requests[0]).toMatchObject({ kind: "accept_exit_offer", expected_founder_revision: 4 });
  await unmount(app); target.remove();
});

it.skipIf(typeof document === "undefined")("renders save age from the last authoritative snapshot", async () => {
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime: new FixtureRuntime(true) } }) as unknown as AppExports;
  await new Promise((resolve) => setTimeout(resolve, 0));
  app.fixtureSnapshot(snapshot); app.fixtureMonotonicElapsed(61_000); app.fixtureSurface("settings"); flushSync();
  expect(target.textContent).toContain("Saved to the server 0:01:01 ago.");
  await unmount(app); target.remove();
});

it.skipIf(typeof document === "undefined")("switches the authoritative tier era without replacing persistent focus", async () => {
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime: new FixtureRuntime(true) } }) as unknown as AppExports;
  await new Promise((resolve) => setTimeout(resolve, 0));
  app.fixtureSnapshot(snapshot); app.fixtureSurface("desk"); flushSync();
  const settings = target.querySelector("nav button:last-child") as HTMLButtonElement;
  expect(settings.textContent).toBe("Options");
  settings.focus();
  app.fixtureSnapshot({ ...snapshot, run: { ...snapshot.run, tier: 1 } }); flushSync();
  expect(target.querySelector("main")?.getAttribute("data-era")).toBe("era_2000");
  expect(document.activeElement).toBe(settings);
  await unmount(app); target.remove();
});

it.skipIf(typeof document === "undefined")("renders the Tier-2 era_2010 Desk and Run End under the axe gate (rfc/tier2-content.md §E1)", async () => {
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime: new FixtureRuntime(true) } }) as unknown as AppExports;
  await new Promise((resolve) => setTimeout(resolve, 0));
  app.fixtureSnapshot({ ...snapshot, run: { ...snapshot.run, tier: 2 } }); app.fixtureSurface("desk"); flushSync();
  expect(target.querySelector("main")?.getAttribute("data-era")).toBe("era_2010");
  expect(target.querySelector("main")?.style.getPropertyValue("--cc-color-accent")).toBe("#1864ab");
  assertNoMechanicalPresentation(target); await assertAxe(target, "tier-2 desk");
  app.fixtureRunEnd({ ...ended, payload: { ...ended.payload, tier: 2 } }); flushSync();
  assertNoMechanicalPresentation(target); await assertAxe(target, "tier-2 run_end");
  await unmount(app); target.remove();
});

it.skipIf(typeof document === "undefined")("submits incorporate from the Tier-2 control and keeps the energy bar inert (rfc/tier2-content.md §E2/§E3)", async () => {
  const runtime = new FixtureRuntime(true);
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime } }) as unknown as AppExports;
  await new Promise((resolve) => setTimeout(resolve, 0));
  const tier2 = { ...snapshot, run: { ...snapshot.run, tier: 2 }, transitions: { cross_gate: null, wind_down: { eligible: false },
    incorporate: { factions: [{ copy_key: "incorporate.bootstrapper", faction_id: "bootstrapper" }, { copy_key: "incorporate.open_source", faction_id: "open_source" }] } } } as ParsedGameUISnapshot;
  runtime.current = tier2;
  app.fixtureSnapshot(tier2); app.fixtureSurface("desk"); flushSync();
  const buttons = [...target.querySelectorAll("fieldset.incorporate button")] as HTMLButtonElement[];
  expect(buttons).toHaveLength(2);
  expect(buttons.map((button) => button.textContent)).not.toContain("open_source");
  const refill = [...target.querySelectorAll("section.energy button")] as HTMLButtonElement[];
  expect(refill).toHaveLength(1);
  refill[0]!.click(); flushSync();
  expect(runtime.requests).toHaveLength(0);
  expect(target.querySelector("section.energy [role=status]")).not.toBeNull();
  expect(target.querySelector("section.energy small")?.textContent).toMatch(/Parody/);
  assertNoMechanicalPresentation(target); await assertAxe(target, "tier-2 incorporate and energy");
  buttons[1]!.click(); await new Promise((resolve) => setTimeout(resolve, 0)); flushSync();
  expect(runtime.requests).toHaveLength(1);
  expect(runtime.requests[0]).toMatchObject({ kind: "incorporate", faction_id: "open_source", expected_revision: snapshot.revision });
  await unmount(app); target.remove();
});

it.skipIf(typeof document === "undefined")("feeds authoritative Game UI snapshots through the archived 20 Hz shell worker", async () => {
  const observation = observeNativePrediction();
  const runtime = new FixtureRuntime(true);
  const fast = { ...snapshot, resources: [{ ...snapshot.resources[0], rate_per_second: "1e3" }] };
  // Refresh and direct publication must describe the same authority.
  runtime.current = fast;
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime } }) as unknown as AppExports;
  try {
    await new Promise((resolve) => setTimeout(resolve, 0));
    app.fixtureSnapshot(fast); app.fixtureSurface("desk"); flushSync();
    const output = target.querySelector(".cc-amount output")!;
    const before = output.textContent;
    await expect.poll(() => output.textContent, { interval: 50, timeout: 5_000 }).not.toBe(before);
    expect(observation.outputs.some((row) => row.kind === "predicted_snapshot")).toBe(true);
    expect(observation.commands.filter((row) => row.rate !== undefined).every((row) => row.rate === "1e3")).toBe(true);
  } finally {
    console.info("R-010 original", JSON.stringify(observation.report(runtime, target.querySelector(".cc-amount output"))));
    try { await unmount(app); } finally {
      console.info("R-010 original cleanup", JSON.stringify(observation.report(runtime, target.querySelector(".cc-amount output"))));
      target.remove(); observation.dispose();
    }
  }
});

it.skipIf(typeof document === "undefined").each(["conflicting", "consistent"] as const)("observes native worker predictions with %s visible-return authority", async (authority) => {
  const fast = { ...snapshot, resources: [{ ...snapshot.resources[0], rate_per_second: "1e3" }] };
  const runtime = new FixtureRuntime(true);
  runtime.current = authority === "consistent" ? fast : snapshot;
  const observation = observeNativePrediction();
  const previousVisibility = Object.getOwnPropertyDescriptor(document, "visibilityState");
  // Controlled visible-return event, not a clock or worker replacement.
  Object.defineProperty(document, "visibilityState", { configurable: true, get: () => "visible" });
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime } }) as unknown as AppExports;
  let refreshTimer: ReturnType<typeof setInterval> | undefined;
  try {
    await new Promise((resolve) => setTimeout(resolve, 0));
    await expect.poll(() => observation.outputs.some((row) => row.kind === "predicted_snapshot"), { interval: 50, timeout: 5_000 }).toBe(true);
    const measurementStart = observation.outputs.length;
    app.fixtureSnapshot(fast); app.fixtureSurface("desk"); flushSync();
    const output = target.querySelector(".cc-amount output")!;
    const before = output.textContent;
    const initialCalls = runtime.snapshotCalls;
    const visibleReturn = () => { document.dispatchEvent(new Event("visibilitychange")); };
    visibleReturn(); refreshTimer = setInterval(visibleReturn, 250);
    await new Promise((resolve) => setTimeout(resolve, 2_000)); flushSync();
    expect(runtime.snapshotCalls).toBeGreaterThan(initialCalls + 3);
    expect(observation.commands.filter((row) => row.rate === (authority === "consistent" ? "1e3" : "1e0")).length).toBeGreaterThan(3);
    expect(observation.outputs.slice(measurementStart).filter((row) => row.kind === "predicted_snapshot").length).toBeGreaterThan(0);
    if (authority === "conflicting") expect(output.textContent).toBe(before);
    else await expect.poll(() => output.textContent, { interval: 50, timeout: 5_000 }).not.toBe(before);
  } finally {
    clearInterval(refreshTimer);
    console.info(`R-010 ${authority}`, JSON.stringify(observation.report(runtime, target.querySelector(".cc-amount output"))));
    try { await unmount(app); } finally {
      console.info(`R-010 ${authority} cleanup`, JSON.stringify(observation.report(runtime, target.querySelector(".cc-amount output"))));
      target.remove(); observation.dispose();
      if (previousVisibility) Object.defineProperty(document, "visibilityState", previousVisibility);
      else Reflect.deleteProperty(document, "visibilityState");
    }
  }
});

const chromiumPerformanceLane = typeof navigator !== "undefined" && /Chrome/u.test(navigator.userAgent);
it.skipIf(!chromiumPerformanceLane)("holds the observable 20 Hz / 10 Hz screen budget for sixty simulated seconds", async () => {
  expect({ width: window.innerWidth, height: window.innerHeight }).toEqual(GAME_UI_PERFORMANCE_BUDGET.viewport);
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime: new FixtureRuntime(true) } }) as unknown as AppExports;
  app.fixtureSnapshot(snapshot); app.fixtureSurface("desk"); flushSync();
  const amount = target.querySelector(".cc-amount output");
  expect(amount).not.toBeNull();
  let formattedCommits = 0;
  const mutations = new MutationObserver((rows) => { formattedCommits += rows.length; });
  mutations.observe(amount!, { characterData: true, childList: true, subtree: true });
  let longestTaskMS = 0;
  const tasks = typeof PerformanceObserver !== "undefined" && PerformanceObserver.supportedEntryTypes.includes("longtask")
    ? new PerformanceObserver((list) => { for (const entry of list.getEntries()) longestTaskMS = Math.max(longestTaskMS, entry.duration); })
    : undefined;
  tasks?.observe({ entryTypes: ["longtask"] });
  for (let input = 1; input <= GAME_UI_PERFORMANCE_BUDGET.inputCount; input++) {
    app.fixtureSnapshot({ ...snapshot, revision: input + 1, resources: [{ ...snapshot.resources[0], amount: canonicalString(input + 100) }] });
    flushSync();
    if (input % 2 === 0) amountRenderScheduler.flush();
    if (input % 40 === 0) await new Promise((resolve) => setTimeout(resolve, 0));
  }
  await tick();
  mutations.disconnect(); tasks?.disconnect();
  validatePerformanceObservation({ formattedCommits, inputs: GAME_UI_PERFORMANCE_BUDGET.inputCount, longestTaskMS });
  await unmount(app); target.remove();
}, 75_000);

// Garage AC7 supplements, rather than rewrites, the original null-feature guard.
// Lucky is a payout opportunity, not a fourth buff. No fixture gameplay intent.
function populatedGaragePerformanceSnapshot(): ParsedGameUISnapshot {
  return parseGameUISnapshot({
    ...snapshot,
    facts: [...snapshot.facts,
      { fact_id: "feature.achievements", value: true },
      { fact_id: "feature.active_play", value: true },
      { fact_id: "feature.fiscal", value: true },
      { fact_id: "feature.meters", value: true },
    ].sort((a, b) => a.fact_id < b.fact_id ? -1 : a.fact_id > b.fact_id ? 1 : 0),
    generators: [{ ...snapshot.generators[0], provisioned: 3,
      provision_cap: { amount: 3, reason_key: "generator.beige_tower.provisioned_cap" } }],
    upgrades: [{ ...snapshot.upgrades[0], eligible: false, owned: true }],
    features: {
      ...snapshot.features,
      achievements: { rows: [{ achievement_id: "achievement.first_gate", condition_scope: "run",
        copy_key: "achievement.first_gate", earned: "run", proof_kind: "provenance", score_grant: 2 }],
      score: { lifetime: 0, run: 2 } },
      fiscal: { credit: 4, credit_cap: { amount: 1000, reason_key: "cap.fiscal_credit" }, credit_per_period: 3,
        generator_levels: [{ generator_id: "generator.beige_tower", level: 0,
          level_cap: { amount: 100, reason_key: "cap.fiscal_level.beige_tower" }, next_level_cost: 1, ppm_per_level: 10_000 }],
        hoard: { cap_credits: 100, preview_ppm: 40_000, reason_note: "next_run" },
        period: { auto_ms: 300_000, early_ms: 100_000, early_success_ppm: 500_000,
          guaranteed_ms: 200_000, opened_wall_ms: snapshot.server_now_ms, seq: 0 },
        sweep_preview: { credit_after: 4, credited: 0, periods: 0, saturated: false },
        unlocks: [{ cost: 3, owned: false, unlock_id: "minigame.pitch" }] },
      meters: { meters: REQUIRED_METER_IDS.map((meter_id) => ({ meter_id, min: 0, max: 100, value: 50,
        band_id: "low", bands: [{ band_id: "low", floor_value: 0 }, { band_id: "high", floor_value: 70 }] })) },
      opportunity: {
        attended_now_ms: 2_000,
        pending: { effect_row_id: "active.lucky", expires_attended_ms: 5_500,
          opportunity_id: "01986666-0000-7000-8000-000000000001", selected_generator_id: null },
        buffs: [
          { buff_instance_id: "01986666-0000-7000-8000-00000000000b", effect_row_id: "active.building",
            expires_attended_ms: 4_000, selected_target: "generator.beige_tower" },
          { buff_instance_id: "01986666-0000-7000-8000-00000000000c", effect_row_id: "active.click",
            expires_attended_ms: 3_000, selected_target: null },
          { buff_instance_id: "01986666-0000-7000-8000-00000000000d", effect_row_id: "active.production",
            expires_attended_ms: 7_000, selected_target: null },
        ],
        combo: { cap: "1e4", reason_key: "cap.active_combo", saturated: false },
      },
    },
  });
}

it.skipIf(!chromiumPerformanceLane)("holds the observable 20 Hz / 10 Hz screen budget with populated Garage regions and a completed native observation", async ({ annotate }) => {
  const budget = GAME_UI_PERFORMANCE_BUDGET;
  expect({ width: window.innerWidth, height: window.innerHeight }).toEqual(budget.viewport);
  const longTasksSupported = typeof PerformanceObserver !== "undefined" && PerformanceObserver.supportedEntryTypes.includes("longtask");
  if (!longTasksSupported || typeof MutationObserver === "undefined") throw new Error("invalid performance observation: native observers unavailable");
  const populated = populatedGaragePerformanceSnapshot();
  const runtime = new FixtureRuntime(true); runtime.current = populated;
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime } }) as unknown as AppExports;
  let formattedCommits = 0, inputs = 0, regionSamples = 0, longestTaskMS = 0;
  let completed = false;
  const mutations = new MutationObserver((rows) => { formattedCommits += rows.length; });
  const collectTasks = (rows: readonly PerformanceEntry[]) => {
    for (const entry of rows) longestTaskMS = Math.max(longestTaskMS, entry.duration);
  };
  const tasks = new PerformanceObserver((list) => collectTasks(list.getEntries()));
  const frame = () => new Promise<void>((resolve) => requestAnimationFrame(() => resolve()));
  let amount: HTMLOutputElement | null = null;
  try {
    app.fixtureSnapshot(populated); app.fixtureSurface("desk"); flushSync();
    // Allow the actual mount/subscription recovery to finish before observation.
    await new Promise((resolve) => setTimeout(resolve, 0)); await tick(); flushSync();
    amountRenderScheduler.flush(); flushSync(); await frame();
    amount = target.querySelector('section[aria-labelledby="resources-heading"] .cc-amount output');
    expect(amount).not.toBeNull();
    const opportunity = target.querySelector('[data-region="desk.region.opportunity"]');
    expect(opportunity).not.toBeNull();
    const text = (key: CopyKey, slots: Parameters<typeof t>[1] = {}) => t(key, slots, eraForSnapshot(populated));
    const census = () => {
      expect(target.querySelector('section[aria-labelledby="resources-heading"] .cc-amount output')).toBe(amount);
      expect(amount!.isConnected).toBe(true);
      expect(target.querySelector('[data-region="desk.region.opportunity"]')).toBe(opportunity);
      expect(opportunity!.querySelector("button")?.textContent).toBe(text("desk.opportunity.claim"));
      expect(opportunity!.querySelector("h3")?.textContent).toBe(text(FEATURES_PRESENTATION.opportunityEffects.get("active.lucky")!.title_key));
      expect([...opportunity!.querySelectorAll("li span:first-child")].map((row) => row.textContent)).toEqual(
        ["active.building", "active.click", "active.production"].map((id) => text(FEATURES_PRESENTATION.opportunityEffects.get(id)!.title_key)));
      expect(opportunity!.textContent).toContain(text("cap.active_combo"));
      expect(opportunity!.querySelector(".cc-amount output")?.textContent).toBe(formatAmount("1e4"));
      const generator = target.querySelector('section[aria-labelledby="generators-heading"]');
      expect(generator?.textContent).toContain(text("desk.provisioned_frame", { count: 3 }));
      expect(generator?.textContent).toContain(text("generator.beige_tower.provisioned_cap"));
      expect(target.querySelector('section[aria-labelledby="upgrades-heading"] strong')?.textContent).toBe(text("desk.upgrade.owned"));
      const shelf = [...target.querySelectorAll("section.card")].find((row) => row.querySelector("h2")?.textContent === text("cosmetic.horse_armor_free.title"));
      expect(shelf?.textContent).toContain(text("cosmetic.horse_armor_free.disclosure"));
      expect([...target.querySelectorAll("nav button")].map((row) => row.textContent)).toEqual([
        "surface.desk.title", "surface.achievements.title", "surface.fiscal.title", "surface.meters.title", "surface.settings.title",
      ].map((key) => text(key as CopyKey)));
      expect(runtime.requests).toHaveLength(0);
    };
    census();
    mutations.observe(amount!, { characterData: true, childList: true, subtree: true });
    tasks.observe({ entryTypes: ["longtask"] });
    for (let input = 1; input <= budget.inputCount; input++) {
      const current = parseGameUISnapshot({ ...populated, revision: input + 1,
        resources: [{ ...populated.resources[0], amount: canonicalString(input + 100) }] });
      runtime.current = current;
      app.fixtureSnapshot(current); inputs++;
      flushSync();
      if (input % 2 === 0) { amountRenderScheduler.flush(); flushSync(); }
      census(); regionSamples++;
      if (input % 40 === 0) await new Promise((resolve) => setTimeout(resolve, 0));
    }
    await tick();
    // Microtasks alone do not deliver the last task's PerformanceObserver record.
    await frame(); await frame(); await new Promise((resolve) => setTimeout(resolve, 0));
    formattedCommits += mutations.takeRecords().length; collectTasks(tasks.takeRecords());
    mutations.disconnect(); tasks.disconnect();
    expect(regionSamples).toBe(budget.inputCount);
    expect(formattedCommits, "invalid observation: no visible cash rendering").toBeGreaterThan(0);
    expect(amount!.textContent, "terminal cash must reflect the actual completed inputs").toBe(formatAmount(canonicalString(budget.inputCount + 100)));
    census();
    validatePerformanceObservation({ formattedCommits, inputs, longestTaskMS });
    completed = true;
  } finally {
    mutations.disconnect(); tasks.disconnect();
    const report = { population: "garage-gs5-gs6-desk", viewport: budget.viewport, simulatedDurationMS: budget.durationMS,
      inputs, regionSamples, formattedCommits, longestTaskMS, longTasksSupported, completed,
      terminalCash: amount?.textContent ?? null, intents: runtime.requests.length,
      exclusions: ["real-60-second-manual-4x-profile", "Firefox-WebKit", "SQL-default-player", "GS4-pet",
        "Pitch-runtime-port", "later-adoption-cosmetics-axis-stack-reputation-T2-regions", "full-current-release-population"] };
    try { await unmount(app); } finally { target.remove(); }
    await annotate(`Garage populated performance observation: ${JSON.stringify(report)}`, { contentType: "application/json", body: new TextEncoder().encode(JSON.stringify(report)) });
  }
}, 75_000);

it.skipIf(typeof document === "undefined")("renders a rejected intent's reason in the status region instead of going offline (GS0.2, F2)", async () => {
  const runtime = new FixtureRuntime(true);
  runtime.intentOutcome = { outcome: "rejected", category: "unaffordable", detail: "company.cash", currentRevision: 1, sessionExpired: false };
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime } });
  await new Promise((resolve) => setTimeout(resolve, 0)); flushSync();
  const buy = [...target.querySelectorAll("button")].find((button) => button.textContent === "Buy 1")!;
  buy.click();
  await new Promise((resolve) => setTimeout(resolve, 0)); await new Promise((resolve) => setTimeout(resolve, 0)); flushSync();
  expect(runtime.requests[0]).toMatchObject({ kind: "buy_generator", expected_revision: 1 });
  expect(target.querySelector(".intent-notice")?.textContent).toBe("Not enough funds for that.");
  expect(target.querySelector("main")?.getAttribute("aria-busy")).toBe("false");
  runtime.intentOutcome = { outcome: "rejected", category: "revision_conflict", detail: "expected_revision", currentRevision: 2, sessionExpired: false };
  runtime.snapshotCalls = 0;
  buy.click();
  await new Promise((resolve) => setTimeout(resolve, 0)); await new Promise((resolve) => setTimeout(resolve, 0)); flushSync();
  expect(target.querySelector(".intent-notice")?.textContent).toContain("The books changed");
  expect(runtime.snapshotCalls).toBe(1);
  await unmount(app); target.remove();
});
