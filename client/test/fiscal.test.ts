import { describe, expect, it } from "vitest";

import fixture from "../../balance/testdata/fiscal-foundation-v1.json";
import epoch5Economy from "../../balance/testdata/epoch5/economy.json";
import { parseCatalog } from "../src/economy-kernel";
import { fiscalEarlyHarvestDraw, fiscalGeneratorCost, fiscalGeneratorFactor, fiscalHoardFactor, harvestFiscal, loadFiscalCatalog, spendFiscal, sweepFiscal, type FiscalState } from "../src/fiscal";

function economyWithFiscalSources(): ReturnType<typeof parseCatalog> {
  const source = structuredClone(epoch5Economy) as any;
  source.multiplier_sources.push(
    { id: "fiscal.generator.beige_tower", slot: "prestige", target: "generator.beige_tower", provider: "fiscal" },
    { id: "fiscal.hoard", slot: "prestige", target: "all", provider: "fiscal" },
  );
  source.multiplier_sources.sort((left: any, right: any) => left.id < right.id ? -1 : left.id > right.id ? 1 : 0);
  return parseCatalog(source);
}

function mutate(root: any, operation: { op: string; path: Array<string | number>; value?: unknown }): void {
  let current = root; for (const component of operation.path.slice(0, -1)) current = current[component];
  const last = operation.path.at(-1)!; if (operation.op === "delete") delete current[last]; else current[last] = operation.value;
}

describe("fiscal foundation", () => {
  const economy = economyWithFiscalSources();
  it("matches the shared mutation corpus", () => {
    expect(fixture.version).toBe(1); expect(() => loadFiscalCatalog(fixture.baseline, economy)).not.toThrow();
    for (const vector of fixture.invalid_mutations) { const candidate = structuredClone(fixture.baseline); mutate(candidate, vector); expect(() => loadFiscalCatalog(candidate, economy), vector.name).toThrow(); }
  });
  it("matches exact factor and cost vectors", () => {
    const catalog = loadFiscalCatalog(fixture.baseline, economy);
    for (const vector of fixture.factor_vectors) expect(vector.kind === "hoard" ? fiscalHoardFactor(catalog, vector.count) : fiscalGeneratorFactor(catalog, "generator.beige_tower", vector.count)).toBe(vector.expected);
    for (const vector of fixture.cost_vectors) expect(fiscalGeneratorCost(catalog, "generator.beige_tower", vector.current, vector.levels)).toBe(vector.expected);
  });
  it("matches exact seed/substream vectors", async () => {
    for (const vector of fixture.rng_vectors) expect(await fiscalEarlyHarvestDraw(vector.founder_id, vector.sequence)).toBe(vector.expected);
  });
  it("keeps sweep phase and rolls rejected spend back", async () => {
    const catalog = loadFiscalCatalog(fixture.baseline, economy);
    const state: FiscalState = { credit: 998, periodOpenedWallMs: 1_000, periodSequence: 0, generatorLevels: { "generator.beige_tower": 0 }, unlocks: [] };
    expect(sweepFiscal(catalog, state, 1_901)).toMatchObject({ periods: 3, credited: 2, saturated: true, openedAfterMs: 1_900, sequenceAfter: 3 });
    expect(state).toMatchObject({ credit: 1_000, periodOpenedWallMs: 1_900, periodSequence: 3 });
    const rejected: FiscalState = { credit: 0, periodOpenedWallMs: 1_000, periodSequence: 0, generatorLevels: { "generator.beige_tower": 0 }, unlocks: [] };
    expect(() => spendFiscal(catalog, rejected, 1_300, { kind: "unlock", unlockId: "unlock.arcade" })).toThrow("insufficient fiscal credit");
    expect(rejected).toEqual({ credit: 0, periodOpenedWallMs: 1_000, periodSequence: 0, generatorLevels: { "generator.beige_tower": 0 }, unlocks: [] });
    const guaranteed: FiscalState = { credit: 0, periodOpenedWallMs: 1_000, periodSequence: 0, generatorLevels: { "generator.beige_tower": 0 }, unlocks: [] };
    expect(await harvestFiscal(catalog, guaranteed, "founder.one", 1_200)).toMatchObject({ outcome: "guaranteed", creditAfter: 3, sequenceBefore: 0 });
  });
});
