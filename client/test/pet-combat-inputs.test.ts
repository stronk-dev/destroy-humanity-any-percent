import { describe, expect, it } from "vitest";
import fixture from "../../testdata/pet/combat-inputs-v1.json";
import { petCombatInputs } from "../src/pet/combat-inputs";

describe("pet care combat input seam", () => {
  it("matches the shared Go/TypeScript vectors", () => {
    expect(fixture.version).toBe(1);
    for (const testCase of fixture.cases) {
      if (testCase.valid) {
        expect(petCombatInputs(testCase.pet_trust_ppm, testCase.soul), testCase.name).toEqual(testCase.expected);
      } else {
        expect(() => petCombatInputs(testCase.pet_trust_ppm, testCase.soul), testCase.name).toThrow(RangeError);
      }
    }
  });
});
