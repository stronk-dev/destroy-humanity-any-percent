import { describe, expect, it } from "vitest";
import fixture from "../../balance/guilds/phase0.json";
import { parseGuildCatalog } from "../src/guild";

describe("Guild catalog required literal members", () => {
  it("preserves the declared zero consumption bonus", () => {
    expect(parseGuildCatalog(fixture).consumptionBonusPpmPerUnit).toBe(0);
  });
  for (const key of Object.keys(fixture)) {
    it(`refuses alias/${key}`, () => {
      const row = structuredClone(fixture) as Record<string, unknown>;
      row[key.toUpperCase()] = row[key];
      delete row[key];
      expect(() => parseGuildCatalog(row)).toThrow();
    });
  }
  for (const variant of ["missing", "null"] as const) {
    it(`refuses zero field/${variant}`, () => {
      const row = structuredClone(fixture) as Record<string, unknown>;
      if (variant === "missing") delete row.consumption_bonus_ppm_per_unit;
      else row.consumption_bonus_ppm_per_unit = null;
      expect(() => parseGuildCatalog(row)).toThrow();
    });
  }
});
