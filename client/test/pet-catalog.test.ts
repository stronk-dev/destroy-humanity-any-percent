import { describe, expect, it } from "vitest";
import fixture from "../../testdata/pet/catalog-grammar-v1.json";
import { parsePetCatalogGrammar } from "../src/pet/catalog";

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
