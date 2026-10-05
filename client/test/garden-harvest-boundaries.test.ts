import { describe, expect, it } from "vitest";
import validCatalog from "../../balance/testdata/server-garden/fixture-v1.json?raw";
import fiscalFixture from "../../balance/testdata/server-garden/fiscal-fixture-v1.json";
import catalogCorpus from "../../testdata/garden/catalog-fixtures-v1.json";
import corpus from "../../testdata/garden/harvest-boundaries-v1.json";
import { COPY_KEYS } from "../src/copy";
import { loadGardenCatalog } from "../src/garden/catalog";
import { encodeGardenState, harvestGarden, parseGardenState } from "../src/garden/engine";

const declarations = {
  copyKeys: new Set(COPY_KEYS), resourceIds: new Set(catalogCorpus.resource_ids),
  fiscalUnlockIds: new Set(fiscalFixture.unlock_rows.map((row) => row.unlock_id)),
  fiscalGeneratorIds: new Set(fiscalFixture.generator_level_rows.map((row) => row.generator_id)),
};

describe("Garden pure harvest boundary bytes (SG6)", () => {
  it("retains the full matrix and a genuinely failing floating-product control", () => {
    expect(corpus.version).toBe(1);
    expect(corpus.cases).toHaveLength(18);
    const scalar = corpus.cases.filter((test) => test.scalar_harvest_units !== undefined);
    expect(scalar).toHaveLength(16);
    expect(new Set(scalar.map((test) => `${test.scalar_harvest_units}:${test.scalar_effect_ppm}`)).size).toBe(16);
    expect([...new Set(scalar.map((test) => test.scalar_harvest_units))].sort((a, b) => a! - b!)).toEqual([0, 1, 999999997, 1000000000]);
    expect([...new Set(scalar.map((test) => test.scalar_effect_ppm))].sort((a, b) => a! - b!)).toEqual([0, 1, 9666667, 10000000]);
    const failures = scalar.filter((test) => Math.floor(test.scalar_harvest_units! * test.scalar_effect_ppm! / 1_000_000) !== test.expected_units);
    expect(failures.map((test) => test.name)).toEqual(["units_999999997_effect_9666667"]);
    expect(failures[0]!.expected_product).toBe("9666666970999999");
    expect(failures[0]!.expected_units).toBe(9666666970);
    expect(Math.floor(999999997 * 9666667 / 1_000_000)).toBe(9666666971);
  });

  for (const test of corpus.cases) {
    it(test.name, async () => {
      const raw = JSON.parse(validCatalog) as { species: { harvest_units: number }[] };
      // Each Go scenario applies the same single literal field override through
      // the real loader. Assert that this is the corpus's only catalog change.
      expect(test.catalog_ops).toHaveLength(1);
      expect(test.catalog_ops[0]!.op).toBe("set");
      expect(test.catalog_ops[0]!.path).toEqual(["species", 0, "harvest_units"]);
      raw.species[0]!.harvest_units = test.catalog_ops[0]!.value;
      const catalog = loadGardenCatalog(JSON.stringify(raw), declarations);
      const state = parseGardenState(JSON.parse(test.initial), catalog);
      if (test.scalar_harvest_units !== undefined) {
        expect(catalog.species[0]!.species_id).toBe("strain_a");
        expect(catalog.species[0]!.harvest_units).toBe(test.scalar_harvest_units);
        expect(state.plots).toHaveLength(1);
        expect(state.plots[0]!.species_id).toBe("strain_a");
        expect(state.plots[0]!.matured_effect_ppm).toBe(test.scalar_effect_ppm);
      }
      const result = await harvestGarden(catalog, state, test.intent_id, test.targets);
      expect(JSON.stringify(result)).toBe(test.expected_result);
      const post = encodeGardenState(state);
      expect(() => parseGardenState(post, catalog)).not.toThrow();
      expect(JSON.stringify(post)).toBe(test.expected_state);
    });
  }
});
