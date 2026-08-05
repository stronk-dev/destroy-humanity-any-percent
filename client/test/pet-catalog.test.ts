import { describe, expect, it } from "vitest";
import fixture from "../../testdata/pet/catalog-grammar-v1.json";
import { parsePetCatalog, parsePetCatalogGrammar } from "../src/pet/catalog";

describe("pet catalog grammar", () => {
  it("loads the shared exact-key fixture", () => {
    const grammar = parsePetCatalogGrammar(fixture);
    expect(grammar.mood_thresholds).toEqual(fixture.mood_thresholds);
    expect(grammar.behavior_candidates).toEqual(fixture.behavior_candidates);
  });

  it("rejects ambiguous thresholds and candidates", () => {
    const base = structuredClone(fixture) as Record<string, any>;
    const cases = [
      { ...base, mood_thresholds: base.mood_thresholds.slice(0, 3) },
      { ...base, mood_thresholds: base.mood_thresholds.map((row: any, index: number) => index === 1 ? { ...row, mood_member: "withdrawn" } : row) },
      { ...base, mood_thresholds: base.mood_thresholds.map((row: any, index: number) => index === 2 ? { ...row, floor_ppm: 1 } : row) },
      { ...base, behavior_candidates: [{ ...base.behavior_candidates[0], event: "offline_tick" }] },
      { ...base, behavior_candidates: [{ ...base.behavior_candidates[0], duration_grid_ticks: 0 }] },
      { ...base, behavior_candidates: [base.behavior_candidates[0], { ...base.behavior_candidates[0], duration_grid_ticks: 2 }] },
      { ...base, extra: true },
    ];
    for (const value of cases) expect(() => parsePetCatalogGrammar(value)).toThrow();
  });
});

describe("complete pet artifact", () => {
  it("loads the full C17 policy union", () => {
    const catalog = parsePetCatalog({ schema_version: 1,
      stat_policy: { grid_ms: 60_000, stats: [
        { stat_id: "hunger", initial_ppm: 800_000, floor_ppm: 100_000, decay_ppm_per_grid: 1_000 },
        { stat_id: "energy", initial_ppm: 800_000, floor_ppm: 100_000, decay_ppm_per_grid: 1_000 },
        { stat_id: "cleanliness", initial_ppm: 800_000, floor_ppm: 100_000, decay_ppm_per_grid: 1_000 },
        { stat_id: "affection", initial_ppm: 800_000, floor_ppm: 100_000, decay_ppm_per_grid: 1_000 }], diminishing_threshold_ppm: 700_000, diminishing_factor_ppm: 500_000 },
      actions: [], trust_policy: { initial_ppm: 500_000, neutral_ppm: 500_000, floor_ppm: 100_000, cap_ppm: 1_000_000, gain_ppm_per_effective_action: 1_000, decay_ppm_per_grid: 100 },
      mood_policy: [{ mood_member: "withdrawn", floor_ppm: 0 }, { mood_member: "restless", floor_ppm: 250_000 }, { mood_member: "neutral", floor_ppm: 500_000 }, { mood_member: "engaged", floor_ppm: 750_000 }], behavior_policy: [] });
    expect(catalog.stat_policy.stats).toHaveLength(4);
  });
});
