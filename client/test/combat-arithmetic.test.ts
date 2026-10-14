import { describe, expect, it } from "vitest";

import fixtureJson from "../../testdata/combat/arithmetic-vectors.json";
import { chart, clamp, damage, saturateInt32, type ChartResult, type Temperament } from "../src/combat/arithmetic";
import { battleSeed, substream } from "../src/combat/rng";

const fixture = fixtureJson as {
  version: number;
  damage: readonly { name: string; base_power: number; attacker_atk: number; chart: ChartResult; critical: boolean; expected: number }[];
  saturation: readonly { value: string; expected: number }[];
  clamp: readonly { name: string; value: string; minimum: number; maximum: number; expected: number }[];
  invalid_clamp: readonly { name: string; value: string; minimum: number; maximum: number }[];
  invalid_damage: readonly { name: string; base_power: number; attacker_atk: number; chart: ChartResult; critical: boolean }[];
  chart: { temperaments: readonly Temperament[]; rows: readonly (readonly ChartResult[])[] };
  rng: {
    match_seed: string;
    battle_seed: string;
    substreams: Readonly<Record<string, string>>;
    bounded: readonly { label: string; bound: string; expected: string; draws: number }[];
  };
};
const temperaments: readonly Temperament[] = ["lazy", "playful", "curious", "sassy", "shy", "chaotic"];

describe("combat shared arithmetic", () => {
  const invalidIntegers = [
    { name: "native integer", value: 1 },
    { name: "fraction", value: 0.5 },
    { name: "NaN", value: Number.NaN },
    { name: "null", value: null },
    { name: "undefined", value: undefined },
    { name: "numeric string", value: "1" },
    { name: "boxed bigint", value: Object(1n) },
    { name: "boolean", value: true },
  ];
  const stores = [
    { name: "saturation", store: (value: unknown) => saturateInt32(value as bigint) },
    { name: "clamp", store: (value: unknown) => clamp(value as bigint, 0, 100) },
  ];
  it.each(stores.flatMap(store => invalidIntegers.map(value => ({ ...value, store: store.store, operation: store.name }))))(
    "rejects $name at the $operation integer boundary", ({ value, store }) => {
      expect(() => store(value)).toThrow(RangeError);
    },
  );
  it("does not invoke caller coercion at either integer store boundary", () => {
    let coercions = 0;
    const value = { valueOf() { coercions += 1; return 1n; } };
    for (const { store } of stores) expect(() => store(value)).toThrow(RangeError);
    expect(coercions).toBe(0);
  });
  const invalidCritical = [
    { name: "zero", value: 0 }, { name: "one", value: 1 },
    { name: "false string", value: "false" }, { name: "true string", value: "true" },
    { name: "null", value: null }, { name: "undefined", value: undefined },
    { name: "boxed boolean", value: Object(false) },
  ];
  it.each([0, 64].flatMap(base => invalidCritical.map(entry => ({ ...entry, base }))))(
    "rejects non-boolean critical $name even with base power $base", ({ value, base }) => {
      expect(() => damage(base, 64, 0, value as boolean)).toThrow(RangeError);
    },
  );
  it("uses vector schema 1", () => expect(fixture.version).toBe(1));
  it.each(fixture.damage)("$name", (vector) => {
    expect(damage(vector.base_power, vector.attacker_atk, vector.chart, vector.critical)).toBe(vector.expected);
  });
  it("retains the complete shared boundary population", () => {
    expect(fixture.damage).toHaveLength(13);
    expect(fixture.saturation).toHaveLength(9);
    expect(fixture.clamp).toHaveLength(13);
    expect(fixture.invalid_clamp).toHaveLength(5);
    expect(fixture.invalid_damage).toHaveLength(6);
  });
  it.each(fixture.saturation)("stores int64 $value without wrapping", (vector) => {
    expect(saturateInt32(BigInt(vector.value))).toBe(vector.expected);
  });
  it.each(fixture.clamp)("clamps $name", (vector) => {
    expect(clamp(BigInt(vector.value), vector.minimum, vector.maximum)).toBe(vector.expected);
  });
  it.each(fixture.invalid_clamp)("rejects clamp $name", (vector) => {
    expect(() => clamp(BigInt(vector.value), vector.minimum, vector.maximum)).toThrow(RangeError);
  });
  it.each(fixture.invalid_damage)("rejects damage $name", (vector) => {
    expect(() => damage(vector.base_power, vector.attacker_atk, vector.chart, vector.critical)).toThrow(RangeError);
  });
  it("matches every ruled cycle edge, including self and opposite neutrality", () => {
    expect(fixture.chart.temperaments).toEqual(temperaments);
    expect(fixture.chart.rows).toHaveLength(6);
    for (const [row, attacker] of fixture.chart.temperaments.entries()) {
      expect(fixture.chart.rows[row]).toHaveLength(6);
      for (const [column, defender] of fixture.chart.temperaments.entries()) {
        expect(chart(attacker, defender)).toBe(fixture.chart.rows[row]![column]);
      }
    }
    for (const [attacker, defender] of [["unknown", "lazy"], ["lazy", "unknown"], ["", ""]]) {
      expect(() => chart(attacker as Temperament, defender as Temperament)).toThrow(RangeError);
    }
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
  it.each(fixture.rng.bounded)("bounds $label to $bound", (vector) => {
    const battle = battleSeed(BigInt(fixture.rng.match_seed));
    const random = substream(battle, vector.label);
    expect(random.bound(BigInt(vector.bound)).toString()).toBe(vector.expected);

    const replay = substream(battle, vector.label);
    const bound = BigInt(vector.bound);
    const threshold = ((1n << 64n) - bound) % bound;
    let draws = 0;
    do {
      draws++;
    } while (replay.next() < threshold);
    expect(draws).toBe(vector.draws);
  });
});
