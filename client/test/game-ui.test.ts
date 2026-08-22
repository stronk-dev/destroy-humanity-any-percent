import { describe, expect, it } from "vitest";

import { eraForSnapshot, parseGameUISnapshot, toShellSnapshot } from "../src/game-ui/contracts";
import { decodeGameUIEvent, decodeGameUISystemEvent } from "../src/game-ui/events";
import { GameUINavigation } from "../src/game-ui/navigation";
import { GAME_UI_PERFORMANCE_BUDGET, validatePerformanceObservation } from "../src/game-ui/performance";
import { defaultSurface, GAME_UI_SURFACES } from "../src/game-ui/surface-catalog";
import { requirePresentationConstant } from "../src/game-ui/presentation";
import { renderPrestigeTermRows } from "../src/game-ui/prestige-terms";
import { parseLocalTiming, priorPersonalBest, RTATimer, timingStorageKey, writeLocalRunTiming } from "../src/game-ui/timing";
import { decodeTransportEnvelope } from "../src/transport";

const snapshot = {
  constants_hash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  evaluated_through_ms: 1_800_000_000_000,
  facts: [{ fact_id: "bootstrap.needed", value: false }, { fact_id: "gate.t2_to_t3", value: false }],
  founder_revision: 1,
  generators: [{ generator_id: "generator.beige_tower", max_affordable: 2, next_cost: "1e1", next_cost_resource_id: "company.cash", owned: 1, provisioned: 0, rate_contribution: "1e0" }],
  manual_action: { action_id: "manual.click", bucket_cap_milli: 50_000, refill_milli_per_ms: 25, refilled_at_ms: 1_800_000_000_000, tokens_milli: 50_000 },
  progress: [{ current: "5e-1", stage_id: "progress.tier", target: "1e0" }],
  resources: [{ amount: "1e2", cap: { amount: "1e1000", reason_key: "resource.company_cash.cap.phase0" }, rate_per_second: "1e0", resource_id: "company.cash" }],
  revision: 1,
  run: { category: "any_percent", exit_count: 0, founder_id: "01985555-1111-7111-8111-111111111111", run_seq: 1, run_started_at_ms: 1_799_999_000_000, tier: 0 },
  schema_version: 3,
  server_now_ms: 1_800_000_000_000,
  transitions: { cross_gate: { eligible: true, gate_id: "gate.t0_to_t1", route_id: null }, wind_down: { eligible: false } },
  upgrades: [],
};

describe("Game UI snapshot contract", () => {
  it("requires the two ruled presentation constants without substitution", () => {
    expect(requirePresentationConstant("constant.price_zero")).toBe("$0.00");
    expect(requirePresentationConstant("constant.founder_fallback")).toBe("Founder");
    expect(() => requirePresentationConstant("constant.missing")).toThrow(/missing presentation/);
  });
  it("validates the exact sorted projection and feeds the existing shell snapshot", () => {
    const parsed = parseGameUISnapshot(snapshot);
    const shell = toShellSnapshot(parsed);
    expect(shell).toEqual({
      constantsHash: snapshot.constants_hash,
      discrete: { "bootstrap.needed": false, "gate.t2_to_t3": false },
      evaluatedThroughMs: snapshot.evaluated_through_ms,
      progress: [{ current: "5e-1", stageId: "progress.tier", target: "1e0" }],
      resources: { "company.cash": { amount: "1e2", cap: { amount: "1e1000", reasonKey: "resource.company_cash.cap.phase0" }, ratePerSecond: "1e0" } },
      revision: 1,
    });
    expect(eraForSnapshot(parsed)).toBe("era_1995");
    expect(() => eraForSnapshot(parseGameUISnapshot({ ...snapshot, run: { ...snapshot.run, tier: 2 } }))).toThrow(/no shipped UI era/);
  });

  it("fails closed on unsorted rows, extra save bytes, and cap overflow", () => {
    expect(() => parseGameUISnapshot({ ...snapshot, facts: [...snapshot.facts].reverse() })).toThrow(/sorted/);
    expect(() => parseGameUISnapshot({ ...snapshot, save_state: {} })).toThrow(/exact/);
    expect(() => parseGameUISnapshot({ ...snapshot, resources: [{ ...snapshot.resources[0], amount: "1e1001" }] })).toThrow(/cap/);
  });

  it("accepts stored bootstrap schemas v1/v2 but requires transitions in current schema v3", () => {
    const { founder_revision: _founderRevision, transitions: _transitions, ...legacy } = snapshot;
    expect(parseGameUISnapshot({ ...legacy, schema_version: 1 })).not.toHaveProperty("founder_revision");
    expect(parseGameUISnapshot({ ...legacy, founder_revision: 1, schema_version: 2 })).not.toHaveProperty("transitions");
    expect(() => parseGameUISnapshot({ ...legacy, schema_version: 2 })).toThrow(/exact/);
    expect(() => parseGameUISnapshot({ ...snapshot, schema_version: 1 })).toThrow(/exact/);
    expect(() => parseGameUISnapshot({ ...snapshot, founder_revision: 0 })).toThrow(/safe integer/);
    expect(() => parseGameUISnapshot({ ...snapshot, transitions: { ...snapshot.transitions, cross_gate: { eligible: true, gate_id: "gate.t0_to_t1", route_id: "route.nope" } } })).toThrow(/cross-gate/);
    expect(() => parseGameUISnapshot({ ...legacy, founder_revision: 1, schema_version: 3 })).toThrow(/exact/);
  });
});

describe("Game UI payout presentation", () => {
  it("formats display deltas and withholds unknown governed IDs", () => {
    expect(renderPrestigeTermRows({
      clout_reach_note: "clout.reach.preserved",
      network_slot_unlocks: [],
      reputation_delta: 1,
      route_knowledge: 2,
    }, "era_1995")).toEqual([
      "Clout carries. The personal brand survives the company.",
      "Reputation +1",
      "Route Knowledge +2",
    ]);
    const withheld = renderPrestigeTermRows({
      clout_reach_note: "clout.reach.future",
      network_slot_unlocks: [{ carried_ref: "carried.future", slot: "network.future" }],
      reputation_delta: 2,
      route_knowledge: 3,
    }, "era_1995");
    expect(withheld).toEqual(["Reputation +2", "Route Knowledge +3"]);
    expect(withheld.join(" ")).not.toMatch(/clout\.reach\.future|network\.future|carried\.future/u);
  });
});

describe("Game UI lifecycle and local timing", () => {
  it("pins the five literal surfaces and authoritative default", () => {
    expect(GAME_UI_SURFACES.map((row) => row.surface_id)).toEqual(["desk", "offer_sheet", "run_end", "settings", "vision_slide"]);
    expect(defaultSurface({ "bootstrap.needed": true })).toBe("vision_slide");
    expect(defaultSurface({ "bootstrap.needed": false })).toBe("desk");
  });

  it("preempts tabs in cursor order but defers behind destructive settings", () => {
    const navigation = new GameUINavigation("desk");
    navigation.select("settings"); navigation.settingsConfirmation(true);
    navigation.lifecycle({ cursor: 8, surface: "offer_sheet" });
    navigation.lifecycle({ cursor: 9, surface: "run_end" });
    expect(navigation.active).toBe("settings");
    navigation.settingsConfirmation(false);
    expect(navigation.active).toBe("run_end");
    navigation.lifecycle({ cursor: 7, surface: "offer_sheet" });
    expect(navigation.active).toBe("run_end");
  });

  it("derives RTA only from the frozen server sample and monotonic elapsed", () => {
    const timer = new RTATimer({ serverNowMs: 5_000, runStartedAtMs: 1_000, sampledMonotonicMs: 100 });
    expect(timer.elapsed(350)).toBe(4_250);
    timer.resample({ serverNowMs: 6_000, runStartedAtMs: 1_000, sampledMonotonicMs: 400 });
    expect(timer.elapsed(500)).toBe(5_100);
    timer.terminal(5_050);
    expect(timer.elapsed(900)).toBe(5_050);
  });

  it("falls back to PB dash state on corrupt local-only timing bytes", () => {
    expect(parseLocalTiming("not-json")).toEqual({ schema_version: 1, records: [] });
    expect(parseLocalTiming('{"schema_version":1,"records":[]}')).toEqual({ schema_version: 1, records: [] });
    expect(timingStorageKey("founder-id")).toBe("cloud-clicker.timing.v1.founder-id");
  });

  it("persists byte-sorted local splits and resolves PB only from prior runs", () => {
    const values = new Map<string, string>();
    const storage = { getItem: (key: string) => values.get(key) ?? null, setItem: (key: string, value: string) => { values.set(key, value); } };
    writeLocalRunTiming(storage, { category: "any_percent", founder_id: "founder", pb_rta_ms: 2_000, run_seq: 1, splits: [{ gate_id: "gate.z", rta_ms: 1_500 }, { gate_id: "gate.a", rta_ms: 500 }] });
    writeLocalRunTiming(storage, { category: "any_percent", founder_id: "founder", pb_rta_ms: 1_500, run_seq: 2, splits: [] });
    const document = parseLocalTiming(values.get(timingStorageKey("founder"))!);
    expect(document.records[0].splits.map((row) => row.gate_id)).toEqual(["gate.a", "gate.z"]);
    expect(priorPersonalBest(document, "founder", 2, "any_percent")).toBe(2_000);
    expect(priorPersonalBest(document, "founder", 3, "any_percent")).toBe(1_500);
  });
});

it("pins the executable CI performance literals", () => {
  expect(GAME_UI_PERFORMANCE_BUDGET).toEqual({ cpuThrottle: 4, droppedFrameAllowancePPM: 50_000, durationMS: 60_000, formattedCommitsMaximum: 600, inputCount: 1_200, longTaskCeilingMS: 200, viewport: { height: 720, width: 1280 } });
  expect(() => validatePerformanceObservation({ formattedCommits: 600, inputs: 1_200, longestTaskMS: 200 })).not.toThrow();
  expect(() => validatePerformanceObservation({ formattedCommits: 601, inputs: 1_199, longestTaskMS: 201 })).toThrowError(
    "Game UI performance budget exceeded: inputs=1199, expected=1200; formatted_commits=601, maximum=600; longest_task_ms=201, maximum=200",
  );
});

describe("Game UI decoded event boundary", () => {
  const envelope = (kind: string, payload: unknown) => decodeTransportEnvelope({
    v: 2, ch: "player:founder", kind: "event", rev: 2,
    constants_hash: snapshot.constants_hash, ts: "2026-08-10T12:00:00Z",
    payload: { event_id: "event-2", kind, scope: "company", rev: 2, cursor_effect: "advance", payload },
  })!;

  it("decodes gate time from the immutable envelope and normalizes offer terminal arms", () => {
    const gate = decodeGameUIEvent(envelope("gate_crossed", { founder_id: snapshot.run.founder_id, gate_id: "gate.t0_to_t1", route_id: null, run_id: { company_stream_id: "01985555-2222-7222-8222-222222222222", run_seq: 1 } }));
    expect(gate).toMatchObject({ cursor: 2, kind: "gate_crossed", occurred_at_ms: Date.parse("2026-08-10T12:00:00Z") });
    const resolved = decodeGameUIEvent(envelope("exit_offer_declined", { offer_id: "01985555-3333-7333-8333-333333333333", run_seq: 1 }));
    expect(resolved).toMatchObject({ kind: "exit_offer_resolved", payload: { resolution: "declined", run_seq: 1 } });
    const accepted = decodeGameUIEvent(envelope("exit_offer_resolved", { offer_id: "01985555-3333-7333-8333-333333333333", resolution: "accepted" }));
    expect(accepted).toMatchObject({ kind: "exit_offer_resolved", payload: { resolution: "accepted", run_seq: null } });
    expect(() => decodeGameUIEvent(envelope("exit_offer_resolved", { offer_id: "01985555-3333-7333-8333-333333333333", resolution: "declined" }))).toThrow();
  });

  it("decodes a run-end object without accepting snapshot bytes", () => {
    const ended = decodeGameUIEvent(envelope("run_ended", {
      assisted: { advisor: false, commons: false }, attended_ms: 500, ended_at_ms: 2_000,
      executed_routes: [], exit_type: "scripted_first", faction: null, founder_id: snapshot.run.founder_id,
      gates_crossed: ["gate.t0_to_t1"], generators_purchased_total: 1, ledger_fact_kinds: [], lifetime_value: "1e3",
      payout: { clout_reach_note: "clout.reach.preserved", network_slot_unlocks: [], reputation_delta: 2, route_knowledge: 25 },
      pre_timer: false, rta_ms: 1_000, run_id: { company_stream_id: "01985555-2222-7222-8222-222222222222", run_seq: 1 },
      started_at_ms: 1_000, terminal_seq: 2, tier: 1,
    }));
    expect(ended).toMatchObject({ kind: "run_ended", payload: { rta_ms: 1_000, tier: 1 } });
    expect(() => decodeGameUIEvent(envelope("run_ended", { ...(ended as { payload: object }).payload, snapshot: {} }))).toThrow(/exact/);
    const branched = decodeGameUIEvent(envelope("run_ended", { ...(ended as { payload: object }).payload, branch: "burnout", starter_package: { kind: "generated_generators", generator_id: "generator.beige_tower", count: 10 } }));
    expect(branched).toMatchObject({ kind: "run_ended", payload: { branch: "burnout", starter_package: { count: 10 } } });
    expect(decodeGameUIEvent(envelope("run_ended", { ...(ended as { payload: object }).payload, branch: "burnout", starter_package: { kind: "generated_generators", generator_id: "generator.beige_tower", count: 9 } }))).toMatchObject({ payload: { starter_package: { count: 9 } } });
    expect(() => decodeGameUIEvent(envelope("run_ended", { ...(ended as { payload: object }).payload, branch: "burnout", starter_package: { kind: "generated_generators", generator_id: "generator.beige_tower", count: 0 } }))).toThrow(/integer/);
  });

  it("owns restart and resync system payloads", () => {
    const restart = decodeTransportEnvelope({ v: 2, ch: "world", kind: "system", rev: 0, constants_hash: snapshot.constants_hash, ts: "2026-08-10T12:00:00Z", payload: { code: "server_restarting", resume_after_ms: 5000 } })!;
    const resync = decodeTransportEnvelope({ v: 2, ch: "world", kind: "system", rev: 0, constants_hash: snapshot.constants_hash, ts: "2026-08-10T12:00:00Z", payload: { code: "resync_required" } })!;
    expect(decodeGameUISystemEvent(restart)).toEqual({ kind: "server_restarting", resume_after_ms: 5000 });
    expect(decodeGameUISystemEvent(resync)).toEqual({ kind: "resync_required" });
  });
});
