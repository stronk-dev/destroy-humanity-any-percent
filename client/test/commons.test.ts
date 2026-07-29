import { describe, expect, it } from "vitest";
import raw from "../../balance/commons/phase0.json";
import vectors from "../../testdata/commons/formula-vectors.json";
import { commonsModifier, enclosureIndex, parseCommonsCatalog } from "../src/commons";

const catalog = parseCommonsCatalog({ ...raw, source_weights: [
  { source_id: "source.ethical", slot: "doctrine", weight_ppm: 1_000_000, forsworn: false },
  { source_id: "source.dark", slot: "upgrades", weight_ppm: 1_000_000, forsworn: true },
] });
describe("commons formula parity", () => {
  for (const vector of vectors.enclosure) it(vector.name, () => expect(enclosureIndex(catalog, vector.sources.map((source: { source_id: string; factor: string }) => ({ sourceId: source.source_id, factor: source.factor, slot: source.source_id === "source.dark" ? "upgrades" : "doctrine" })))).toBe(vector.expected));
  for (const vector of vectors.modifiers) it(`modifier ${vector.health_ppm}/${vector.solidarity_ppm}`, () => expect(commonsModifier(catalog, vector.health_ppm, vector.solidarity_ppm)).toBe(vector.expected));
});

it("reads the convex exponent from balance data", () => {
  const quadratic = parseCommonsCatalog({ ...raw, collective_exponent_ppm: 2_000_000 });
  expect(commonsModifier(quadratic, 675_000, 0)).toBe("1.75e0");
  expect(() => parseCommonsCatalog({ ...raw, collective_exponent_ppm: 999_999 })).toThrow();
});
