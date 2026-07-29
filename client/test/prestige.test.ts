import { describe, expect, it } from "vitest";

import fixtureJson from "../../testdata/prestige-vectors.json";
import { moralReseed, reputationDelta, reputationLevel } from "../src/prestige";

interface Vector {
  name: string;
  lifetime_value: string;
  current_level: number;
  modifier_ppm: number;
  level: number;
  delta: number;
}

const fixture = fixtureJson as { version: number; threshold: string; vectors: Vector[] };

describe("shared prestige vectors", () => {
  it("uses vector schema 1", () => expect(fixture.version).toBe(1));
  it.each(fixture.vectors)("$name", (vector) => {
    expect(reputationLevel(vector.lifetime_value, fixture.threshold)).toBe(vector.level);
    expect(
      reputationDelta(
        vector.lifetime_value,
        fixture.threshold,
        vector.current_level,
        vector.modifier_ppm,
      ),
    ).toBe(vector.delta);
  });
  it("reseed is exact and bounded", () => {
    expect(moralReseed(0)).toBe(90);
    expect(moralReseed(1)).toBe(90);
    expect(moralReseed(100)).toBe(55);
    expect(moralReseed(Number.MAX_SAFE_INTEGER)).toBe(55);
  });
});
