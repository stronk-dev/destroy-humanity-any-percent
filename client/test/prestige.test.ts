import { describe, expect, it } from "vitest";

import fixtureJson from "../../testdata/prestige-vectors.json";
import policyJSON from "../../balance/prestige/phase0.json";
import { moralReseed, parsePrestigePolicy, reputationDelta, reputationLevel } from "../src/prestige";

interface Vector {
  name: string;
  threshold?: string;
  lifetime_value: string;
  current_level: number;
  modifier_ppm: number;
  level: number;
  delta: number;
}

const fixture = fixtureJson as { version: number; threshold: string; vectors: Vector[] };

describe("prestige policy", () => {
  it("requires exact names and explicit non-null zero-valued members", () => {
    for (const key of Object.keys(policyJSON)) {
      const source = structuredClone(policyJSON) as unknown as Record<string, unknown>;
      source[key.toUpperCase()] = source[key]; delete source[key];
      expect(() => parsePrestigePolicy(source), key).toThrow(SyntaxError);
    }
    for (const key of ["collapse_route_knowledge", "decline_drift_ppm", "advisor_per_run_ppm", "advisor_cap_ppm"]) {
      for (const missing of [false, true]) {
        const source = structuredClone(policyJSON) as unknown as Record<string, unknown>;
        source[key] = null; if (missing) delete source[key];
        expect(() => parsePrestigePolicy(source), key).toThrow(SyntaxError);
      }
    }
    for (const kind of Object.keys(policyJSON.exit_modifiers_ppm)) {
      const source = structuredClone(policyJSON) as any;
      source.exit_modifiers_ppm[kind] = null;
      expect(() => parsePrestigePolicy(source), kind).toThrow(SyntaxError);
    }
    for (let index = 0; index < 10; index++) {
      const source = structuredClone(policyJSON) as any;
      source.spawn_gate_ppm[index] = null;
      expect(() => parsePrestigePolicy(source)).toThrow(SyntaxError);
    }
    const original = JSON.stringify(policyJSON);
    const partial = `${original.slice(0,-1)},"exit_modifiers_ppm":{"acquihire":900000}}`;
    expect(() => parsePrestigePolicy(JSON.parse(partial))).toThrow(SyntaxError);
  });
  it("refuses the first duration outside the existing Go expiry representation", () => {
    const maximum = Number(((1n << 63n) - 1n) / 1_000_000n);
    expect(parsePrestigePolicy({ ...policyJSON, offer_duration_ms: maximum }).offerDurationMs).toBe(maximum);
    expect(() => parsePrestigePolicy({ ...policyJSON, offer_duration_ms: maximum + 1 })).toThrow(SyntaxError);
  });
});

describe("shared prestige vectors", () => {
  it("uses vector schema 1", () => expect(fixture.version).toBe(1));
  it.each(fixture.vectors)("$name", (vector) => {
    const threshold = vector.threshold ?? fixture.threshold;
    expect(reputationLevel(vector.lifetime_value, threshold)).toBe(vector.level);
    expect(
      reputationDelta(
        vector.lifetime_value,
        threshold,
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
