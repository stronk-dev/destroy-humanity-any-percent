import { describe, expect, it } from "vitest";

import fixtureJson from "../../testdata/combat/arithmetic-vectors.json";
import { chart, damage, saturateInt32, type ChartResult, type Temperament } from "../src/combat/arithmetic";
import { battleSeed, substream } from "../src/combat/rng";

const fixture = fixtureJson as {
  version: number;
  damage: readonly { name: string; base_power: number; attacker_atk: number; chart: ChartResult; critical: boolean; expected: number }[];
  rng: { match_seed: string; battle_seed: string; substreams: Readonly<Record<string, string>> };
};
const temperaments: readonly Temperament[] = ["lazy", "playful", "curious", "sassy", "shy", "chaotic"];

describe("combat shared arithmetic", () => {
  it("uses vector schema 1", () => expect(fixture.version).toBe(1));
  it.each(fixture.damage)("$name", (vector) => {
    expect(damage(vector.base_power, vector.attacker_atk, vector.chart, vector.critical)).toBe(vector.expected);
  });
  it("saturates int32 stores", () => {
    expect(saturateInt32(9_223_372_036_854_775_807n)).toBe(2_147_483_647);
    expect(saturateInt32(-9_223_372_036_854_775_808n)).toBe(-2_147_483_648);
  });
  it("gives every Temperament two wins and two losses without a 2x path", () => {
    for (const attacker of temperaments) {
      const results = temperaments.map((defender) => chart(attacker, defender));
      expect(results.filter((value) => value === 1)).toHaveLength(2);
      expect(results.filter((value) => value === -1)).toHaveLength(2);
      expect(results.every((value) => value >= -1 && value <= 1)).toBe(true);
    }
  });
  it("isolates labeled random substreams", () => {
    const battle = battleSeed(BigInt(fixture.rng.match_seed));
    expect(battle.toString()).toBe(fixture.rng.battle_seed);
    for (const [label, expected] of Object.entries(fixture.rng.substreams)) expect(substream(battle, label).next().toString()).toBe(expected);
    const before = substream(battle, "crit").next(); substream(battle, "new_consumer").next();
    expect(substream(battle, "crit").next()).toBe(before);
  });
});
