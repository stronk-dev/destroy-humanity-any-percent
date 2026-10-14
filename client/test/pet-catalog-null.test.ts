import { describe, expect, it } from "vitest";
import fixture from "../../balance/pets/first-content.json";
import grammarFixture from "../../testdata/pet/catalog-grammar-v1.json";
import { parsePetCatalog, parsePetCatalogGrammar } from "../src/pet/catalog";

function zeroFields(root: typeof fixture) {
  const fields: { path: string; row: object; key: string }[] = [];
  const add = (path: string, row: object, keys: readonly string[]) => {
    for (const key of keys) {
      Object.assign(row, { [key]: 0 });
      fields.push({ path: `${path}/${key}`, row, key });
    }
  };
  add("stat_policy", root.stat_policy, ["diminishing_threshold_ppm", "diminishing_factor_ppm"]);
  root.stat_policy.stats.forEach((row, index) => add(`stat/${index}`, row, ["initial_ppm", "floor_ppm", "decay_ppm_per_grid"]));
  root.actions.forEach((row, index) => add(`action/${index}`, row, ["cooldown_attended_ms", "min_eligible_ppm"]));
  add("trust_policy", root.trust_policy, ["initial_ppm", "neutral_ppm", "floor_ppm", "cap_ppm", "gain_ppm_per_effective_action", "decay_ppm_per_grid"]);
  add("mood/0", root.mood_policy[0]!, ["floor_ppm"]);
  return fields;
}

describe("pet catalog required numeric values", () => {
  const population = zeroFields(structuredClone(fixture));
  it("preserves all explicitly present legal zeros", () => {
    const candidate = structuredClone(fixture);
    expect(zeroFields(candidate)).toHaveLength(31);
    expect(() => parsePetCatalog(candidate)).not.toThrow();
  });
  it("refuses a null floor in the standalone grammar loader", () => {
    expect(() => parsePetCatalogGrammar(grammarFixture)).not.toThrow();
    const candidate = structuredClone(grammarFixture);
    Object.assign(candidate.mood_thresholds[0]!, { floor_ppm: null });
    expect(() => parsePetCatalogGrammar(candidate)).toThrow();
  });
  for (const [index, field] of population.entries()) {
    it(`refuses null/${field.path}`, () => {
      const candidate = structuredClone(fixture);
      const target = zeroFields(candidate)[index]!;
      Object.assign(target.row, { [target.key]: null });
      expect(() => parsePetCatalog(candidate)).toThrow();
    });
  }
});
