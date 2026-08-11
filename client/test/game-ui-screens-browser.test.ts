import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import GameUIApp from "../src/game-ui/GameUIApp.svelte";
import RunEndSurface from "../src/game-ui/RunEndSurface.svelte";
import type { ExitOfferSpawnedEvent, GateCrossedEvent, RunEndedEvent } from "../src/game-ui/events";
import type { GameUIRuntime, GameUIRuntimeMessage } from "../src/game-ui/runtime";
import type { GameUISnapshot } from "../src/api/generated/types";
import type { ParsedGameUISnapshot } from "../src/game-ui/contracts";
import { canonicalString } from "../src/numeric";
import { GAME_UI_PERFORMANCE_BUDGET, validatePerformanceObservation } from "../src/game-ui/performance";
import { amountRenderScheduler } from "../src/ui/render-scheduler";

const snapshot: GameUISnapshot = {
  constants_hash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  evaluated_through_ms: 1_800_000_000_000,
  facts: [{ fact_id: "bootstrap.needed", value: false }, { fact_id: "gate.t0_to_t1", value: false }],
  founder_revision: 1,
  generators: [{ generator_id: "generator.beige_tower", max_affordable: 2, next_cost: "1e1", next_cost_resource_id: "company.cash", owned: 1, provisioned: 0, rate_contribution: "1e0" }],
  manual_action: { action_id: "manual.click", bucket_cap_milli: 50_000, refill_milli_per_ms: 25, refilled_at_ms: 1_800_000_000_000, tokens_milli: 50_000 },
  progress: [{ current: "5e-1", stage_id: "progress.tier", target: "1e0" }],
  resources: [{ amount: "1e2", cap: { amount: "1e1000", reason_key: "resource.company_cash.cap.phase0" }, rate_per_second: "1e0", resource_id: "company.cash" }],
  revision: 1,
  run: { category: "any_percent", exit_count: 0, founder_id: "01985555-1111-7111-8111-111111111111", run_seq: 1, run_started_at_ms: 1_799_999_000_000, tier: 0 },
  schema_version: 2, server_now_ms: 1_800_000_000_000,
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
  constructor(private authenticated = false) {}
  hasCredentials(): boolean { return this.authenticated; }
  async bootstrap(): Promise<GameUISnapshot> { this.authenticated = true; return snapshot; }
  async snapshot(): Promise<GameUISnapshot> { return snapshot; }
  async intent(body: Readonly<Record<string, unknown>>): Promise<void> { this.requests.push(body); }
  subscribe(_founderID: string, listener: (message: GameUIRuntimeMessage) => void): () => void { this.listener = listener; return () => { this.listener = undefined; }; }
}

interface AppExports {
  fixtureSnapshot(value: ParsedGameUISnapshot): void;
  fixtureSurface(value: "desk" | "offer_sheet" | "run_end" | "settings" | "vision_slide"): void;
  fixtureOffer(value: ExitOfferSpawnedEvent): void;
  fixtureRunEnd(value: RunEndedEvent): void;
  fixtureSystem(value: "drain" | "resync"): void;
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
  runtime.listener?.({ kind: "event", revision: 2, scope: "company", value: crossed }); flushSync();
  expect(target.textContent).toContain("Garage");
  runtime.listener?.({ kind: "event", revision: 3, scope: "company", value: ended }); flushSync();
  expect(target.querySelector("main")?.dataset.surface).toBe("run_end");
  expect([...values.values()][0]).toContain('"rta_ms":750');
  await unmount(app); target.remove();
});

it.skipIf(typeof document === "undefined")("replays a v1 bootstrap receipt fail-closed, then accepts from a v2 Founder coordinate", async () => {
  const runtime = new FixtureRuntime(true);
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime } }) as unknown as AppExports;
  await new Promise((resolve) => setTimeout(resolve, 0));
  const { founder_revision: _founderRevision, ...legacy } = snapshot;
  app.fixtureSnapshot({ ...legacy, schema_version: 1 }); app.fixtureOffer(offer); flushSync();
  const buttons = [...target.querySelectorAll("button")];
  const sign = buttons.find((button) => button.textContent === "Sign")!;
  const decline = buttons.find((button) => button.textContent === "Decline")!;
  expect(sign.disabled).toBe(true);
  app.fixtureSnapshot(snapshot); flushSync();
  expect(sign.disabled).toBe(false);
  sign.click(); await new Promise((resolve) => setTimeout(resolve, 0)); flushSync();
  expect(runtime.requests[0]).toMatchObject({ kind: "accept_exit_offer", expected_founder_revision: 1, offer_id: offer.payload.offer_id });
  decline.click(); await new Promise((resolve) => setTimeout(resolve, 0)); flushSync();
  expect(runtime.requests[1]).toMatchObject({ kind: "decline_exit_offer", offer_id: offer.payload.offer_id });
  expect(runtime.requests[1]).not.toHaveProperty("expected_founder_revision");
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
  expect(target.textContent).toContain("Route Knowledge +2");
  expect(target.textContent).not.toContain("clout.reach.preserved");
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

it.skipIf(typeof document === "undefined")("feeds authoritative Game UI snapshots through the archived 20 Hz shell worker", async () => {
  const target = document.createElement("div"); document.body.append(target);
  const app = mount(GameUIApp, { target, props: { runtime: new FixtureRuntime(true) } }) as unknown as AppExports;
  await new Promise((resolve) => setTimeout(resolve, 0));
  app.fixtureSnapshot({ ...snapshot, resources: [{ ...snapshot.resources[0], rate_per_second: "1e3" }] }); app.fixtureSurface("desk"); flushSync();
  const output = target.querySelector(".cc-amount output")!;
  const before = output.textContent;
  await expect.poll(() => output.textContent, { interval: 50, timeout: 5_000 }).not.toBe(before);
  await unmount(app); target.remove();
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
