import { describe, expect, it } from "vitest";
import fixture from "../../balance/testdata/pitch-v1.json";
import { parsePitchCatalog } from "../src/pitch/catalog";

describe("Pitch catalog price admission", () => {
  for (const [index, hack] of fixture.growth_hacks.entries()) {
    it(`refuses null price/${hack.hack_id}`, () => {
      const candidate = structuredClone(fixture);
      Object.assign(candidate.growth_hacks[index]!, { price: null });
      expect(() => parsePitchCatalog(candidate)).toThrow();
    });
    it(`preserves explicit zero price/${hack.hack_id}`, () => {
      const candidate = structuredClone(fixture);
      candidate.growth_hacks[index]!.price = 0;
      expect(parsePitchCatalog(candidate).growth_hacks[index]!.price).toBe(0);
    });
  }
});
