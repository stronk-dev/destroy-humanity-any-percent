import { describe, expect, it } from "vitest";

import { loadMeterCatalog, REQUIRED_METER_IDS, validateMeterResourceSeparation } from "../src/meters/catalog";

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
