import { describe, expect, it } from "vitest";

import arithmetic from "../../testdata/combat/arithmetic-vectors.json";
import vectors from "../../testdata/combat/identity-vectors.json";
import { chartIdentities, identity, type CombatIdentity } from "../src/combat/identity";

describe("combat runtime identities", () => {
  it("admits exactly the specified 36 identities without runtime chart interactions", () => {
    expect(vectors.version).toBe(1);
    expect(vectors.runtimes).toEqual(["container", "vm", "bare_metal", "serverless", "edge", "mainframe"]);
    expect(arithmetic.chart.temperaments).toHaveLength(6);
    const identities = arithmetic.chart.temperaments.flatMap((temperament) => vectors.runtimes.map((runtime) => identity(temperament, runtime)));
    expect(identities).toHaveLength(36);
    for (const [left, attacker] of identities.entries()) {
      const counts = [0, 0, 0];
      for (const [right, defender] of identities.entries()) {
        const result = chartIdentities(attacker, defender);
        expect(result).toBe(arithmetic.chart.rows[Math.floor(left / 6)]![Math.floor(right / 6)]);
        counts[result + 1]!++;
      }
      expect(counts).toEqual([12, 12, 12]);
    }
  });

  it.each(vectors.invalid)("refuses $name in constructors and both caller-built chart slots", (vector) => {
    expect(() => identity(vector.temperament, vector.runtime)).toThrow(RangeError);
    const invalid = { temperament: vector.temperament, runtime: vector.runtime } as CombatIdentity;
    const valid = identity("lazy", "container");
    expect(() => chartIdentities(invalid, valid)).toThrow(RangeError);
    expect(() => chartIdentities(valid, invalid)).toThrow(RangeError);
  });

  it("refuses non-string axes instead of coercing them", () => {
    for (const value of [null, undefined, 0, true, [], {}, new String("lazy")]) {
      expect(() => identity(value, "container")).toThrow(RangeError);
      expect(() => identity("lazy", value)).toThrow(RangeError);
    }
  });

  it("owns immutable identity storage and never silently repairs caller values", () => {
    expect(vectors.invalid).toHaveLength(12);
    const first = identity("lazy", "vm"), second = identity("lazy", "vm");
    expect(first).toEqual({ temperament: "lazy", runtime: "vm" });
    expect(first).not.toBe(second);
    expect(Object.isFrozen(first)).toBe(true);
    expect(() => { (first as { runtime: string }).runtime = "edge"; }).toThrow(TypeError);
    const mutable: CombatIdentity = { temperament: "lazy", runtime: "vm" };
    expect(chartIdentities(mutable, second)).toBe(0);
    expect(mutable).toEqual({ temperament: "lazy", runtime: "vm" });
  });
});
