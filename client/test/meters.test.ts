import { describe, expect, it } from "vitest";
import meterTransitionCorpus from "../../balance/testdata/meters-transition-v1.json";

import { loadMeterCatalog, REQUIRED_METER_IDS, validateMeterResourceSeparation } from "../src/meters/catalog";
import { advanceMeters, contributionKey, MILLIS_PER_HOUR, newMeterState, validateMeterState } from "../src/meters/transition";

function validCatalog(): Record<string, unknown> {
  return {
    schema_version: 1,
    trust_reseed: { base_value: 90, notoriety_numerator: 35, notoriety_denominator: 100, floor_value: 55, ceiling_value: 90 },
    meters: REQUIRED_METER_IDS.map((id, index) => ({
      id, scope: "company", min_value: 0, max_value: 100, initial_value: 50,
      bands: [{ id: "low", floor_value: 0 }, { id: "high", floor_value: 70 }],
      inputs: index === 0 ? [
        { kind: "ledger_fact", fact_kind: "externality.emitted", delta: 3 },
        { kind: "contribution_slot", slot: "upgrades", source_id: "generator.example", delta_per_attended_hour: -2 },
      ] : [],
      decay: { toward_value: 50, rate_per_attended_hour: 2 },
    })),
  };
}

describe("meter catalog", () => {
  it("loads the exact Phase-A set and closed input union", () => {
    const catalog = loadMeterCatalog(JSON.stringify(validCatalog()));
    expect(catalog.meters.map((meter) => meter.id)).toEqual([...REQUIRED_METER_IDS]);
    expect(catalog.meters[0].inputs).toEqual([
      { kind: "ledger_fact", factKind: "externality.emitted", delta: 3 },
      { kind: "contribution_slot", slot: "upgrades", sourceId: "generator.example", deltaPerAttendedHour: -2 },
    ]);
    expect(() => validateMeterResourceSeparation(catalog, ["company.cash"])).not.toThrow();
    expect(() => validateMeterResourceSeparation(catalog, ["trust.users.standing"])).toThrow(/collision/);
  });

  it("rejects missing axes, forbidden IDs, duplicate sources, and decorative fields", () => {
    const cases: Array<(catalog: any) => void> = [
      (catalog) => { catalog.meters.pop(); },
      (catalog) => { catalog.meters[0].id = "externality.total"; },
      (catalog) => { catalog.meters[9].id = "trust.public.grievance"; },
      (catalog) => { catalog.meters[0].inputs.push({ kind: "ledger_fact", fact_kind: "externality.emitted", delta: 1 }); },
      (catalog) => { catalog.meters[0].spendable = false; },
      (catalog) => { catalog.meters[0].bands[1].floor_value = 0; },
    ];
    for (const mutate of cases) {
      const catalog = validCatalog(); mutate(catalog);
      expect(() => loadMeterCatalog(JSON.stringify(catalog))).toThrow();
    }
  });
});

describe("meter transition", () => {
  it("is partition-invariant and offline-stable", () => {
    const catalog = loadMeterCatalog(JSON.stringify(validCatalog()));
    const whole = newMeterState(catalog);
    const split = newMeterState(catalog);
    const active = new Set([contributionKey("upgrades", "generator.example")]);
    advanceMeters(catalog, whole, { attendedMs: MILLIS_PER_HOUR, newFactKinds: new Set(), activeContributions: active });
    for (const attendedMs of [1_234_567, MILLIS_PER_HOUR - 1_234_567]) {
      advanceMeters(catalog, split, { attendedMs, newFactKinds: new Set(), activeContributions: active });
    }
    expect(whole).toEqual(split);
    const offline = newMeterState(catalog);
    const before = structuredClone(offline);
    advanceMeters(catalog, offline, { attendedMs: 0, newFactKinds: new Set(), activeContributions: active });
    expect(offline).toEqual(before);
  });

  it("matches the shared Go/TypeScript transition corpus", () => {
    const corpus = meterTransitionCorpus as {
      version: number;
      cases: Array<{
        name: string;
        initial_value: number;
        initial_decay_remainder: number;
        initial_input_remainder: number;
        steps: Array<{ attended_ms: number; new_fact_kinds: string[]; active_contributions: string[] }>;
        expected_value: number;
        expected_decay_remainder: number;
        expected_input_remainder: number;
        expected_changes: Array<Record<string, unknown>>;
      }>;
    };
    expect(corpus.version).toBe(1);
    const catalog = loadMeterCatalog(JSON.stringify(validCatalog()));
    for (const vector of corpus.cases) {
      const state = newMeterState(catalog);
      state.values["doom.probability"] = vector.initial_value;
      state.decayRemainders["doom.probability"] = vector.initial_decay_remainder;
      state.inputRemainders["doom.probability:1"] = vector.initial_input_remainder;
      const changes = vector.steps.flatMap((step) => advanceMeters(catalog, state, {
        attendedMs: step.attended_ms,
        newFactKinds: new Set(step.new_fact_kinds),
        activeContributions: new Set(step.active_contributions),
      }));
      expect({
        value: state.values["doom.probability"],
        decay: state.decayRemainders["doom.probability"],
        input: state.inputRemainders["doom.probability:1"],
        changes: changes.map((change) => ({
          meter_id: change.meterId,
          from_band: change.fromBand,
          to_band: change.toBand,
          direction: change.direction,
          value_before: change.valueBefore,
          value_after: change.valueAfter,
        })),
      }).toEqual({
        value: vector.expected_value,
        decay: vector.expected_decay_remainder,
        input: vector.expected_input_remainder,
        changes: vector.expected_changes,
      });
    }
  });

  it("aggregates causal inputs and emits only the final band transition", () => {
    const catalog = loadMeterCatalog(JSON.stringify(validCatalog()));
    const state = newMeterState(catalog);
    state.values["doom.probability"] = 69;
    const changes = advanceMeters(catalog, state, { attendedMs: 0, newFactKinds: new Set(["externality.emitted"]), activeContributions: new Set() });
    expect(state.values["doom.probability"]).toBe(72);
    expect(changes).toEqual([{ meterId: "doom.probability", fromBand: "low", toBand: "high", direction: "up", valueBefore: 69, valueAfter: 72 }]);
  });

  it("clears stale decay phase at the target and rejects inexact maps", () => {
    const catalog = loadMeterCatalog(JSON.stringify(validCatalog()));
    const state = newMeterState(catalog);
    state.decayRemainders["doom.probability"] = 42;
    advanceMeters(catalog, state, { attendedMs: 0, newFactKinds: new Set(), activeContributions: new Set() });
    expect(state.decayRemainders["doom.probability"]).toBe(0);
    delete state.values["trust.users.standing"];
    expect(() => validateMeterState(catalog, state)).toThrow(/invalid meter state/);
    const extra = newMeterState(catalog);
    extra.inputRemainders["extra:0"] = 0;
    expect(() => validateMeterState(catalog, extra)).toThrow(/invalid meter state/);
  });
});
