import fixtureJSON from "../../testdata/cosmetic/catalog-fixtures-v1.json";
import pinnedFixture from "../../balance/testdata/cosmetics/fixture-v1.json?raw";
import { describe, expect, it } from "vitest";

import { loadCosmeticCatalog, validateCosmeticTransition } from "../src/cosmetic/catalog";

const fixture = fixtureJSON as { schema_version: number; cases: { name: string; valid: boolean; artifact: string }[] };

describe("cosmetics catalog (Cosmetic Shop v1 §2, AC1)", () => {
  it("decides every shared Go-authored case identically", () => {
    expect(fixture.schema_version).toBe(1);
    expect(fixture.cases.length).toBeGreaterThan(20);
    for (const testCase of fixture.cases) {
      if (testCase.valid) expect(() => loadCosmeticCatalog(testCase.artifact), testCase.name).not.toThrow();
      else expect(() => loadCosmeticCatalog(testCase.artifact), testCase.name).toThrow();
    }
  });

  it("loads the pinned fixture artifact with exactly Horse Armor at tier 1", () => {
    const catalog = loadCosmeticCatalog(pinnedFixture);
    expect(catalog.items).toEqual([{ cosmetic_id: "horse_armor", slot: "pet", unlock: { kind: "active_company_tier_at_least", tier: 1 } }]);
  });

  it("keeps ids permanent across epochs and allows an unlock retune (AC2)", () => {
    const current = loadCosmeticCatalog(pinnedFixture);
    const retuned = loadCosmeticCatalog(pinnedFixture.replace("\"tier\": 1", "\"tier\": 2"));
    const dropped = loadCosmeticCatalog(pinnedFixture.replace("horse_armor", "zebra_armor"));
    expect(() => validateCosmeticTransition(current, retuned)).not.toThrow();
    expect(() => validateCosmeticTransition(current, dropped)).toThrow(/dropped/u);
    expect(() => validateCosmeticTransition(current, undefined)).toThrow(/disappear/u);
    expect(() => validateCosmeticTransition(undefined, current)).not.toThrow();
  });
});
