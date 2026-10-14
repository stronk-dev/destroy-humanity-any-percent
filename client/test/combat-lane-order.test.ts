import { describe, expect, it } from "vitest";

import vectors from "../../testdata/combat/lane-order-vectors.json";
import { compareLaneOrder, type LaneOrderKey } from "../src/combat/lane-order";

function key(values: readonly number[]): LaneOrderKey {
  return Object.freeze({ tick: values[0]!, player_slot: values[1]!, placement_seq: values[2]! });
}

describe("combat lane input ordering", () => {
  it("uses the complete shared representation corpus", () => {
    expect(vectors.version).toBe(1);
    expect(vectors.comparisons).toHaveLength(10);
  });
  it.each(vectors.comparisons)("$name", row => {
    const left = key(row.left), right = key(row.right);
    expect(compareLaneOrder(left, right)).toBe(row.expected);
    expect(compareLaneOrder(right, left)).toBe(row.expected === 0 ? 0 : -row.expected);
  });
  it("matches independently enumerated ranks and sorts every reversed rotation", () => {
    const keys: LaneOrderKey[] = [];
    for (const tick of [-2_147_483_648, 0, 2_147_483_647]) {
      for (const slot of [0, 1]) {
        for (const seq of [-2_147_483_648, 0, 2_147_483_647]) keys.push(key([tick, slot, seq]));
      }
    }
    expect(keys).toHaveLength(18);
    for (const [i, left] of keys.entries()) {
      for (const [j, right] of keys.entries()) expect(compareLaneOrder(left, right)).toBe(i < j ? -1 : i > j ? 1 : 0);
    }
    for (const shift of keys.keys()) {
      const input = Object.freeze([...keys.slice(shift), ...keys.slice(0, shift)].reverse());
      const before = [...input];
      expect([...input].sort(compareLaneOrder)).toEqual(keys);
      expect(input).toEqual(before);
    }
  });
  const invalid = [undefined, null, [], true, 1, "key",
    ...["tick", "player_slot", "placement_seq"].flatMap(field => [undefined, null, "0", false, 0.5,
      Number.NaN, Number.POSITIVE_INFINITY, -2_147_483_649, 2_147_483_648, Object(0)].map(value => ({ tick: 0, player_slot: 0, placement_seq: 0, [field]: value }))),
  ];
  it.each(invalid.map((value, index) => ({ value, index })))("refuses malformed key $index in both slots before comparison", ({ value }) => {
    const valid = key([2, 0, 0]);
    expect(() => compareLaneOrder(value as LaneOrderKey, valid)).toThrow(RangeError);
    expect(() => compareLaneOrder(valid, value as LaneOrderKey)).toThrow(RangeError);
  });
  it("does not coerce components or reject irrelevant payload fields", () => {
    let coercions = 0;
    const left = { tick: 0, player_slot: 0, placement_seq: { valueOf() { coercions += 1; return 0; } } };
    expect(() => compareLaneOrder(left as unknown as LaneOrderKey, key([1, 0, 0]))).toThrow(RangeError);
    expect(coercions).toBe(0);
    const extended = { tick: 0, player_slot: 0, placement_seq: 0, unrelated: "not ordering authority" };
    expect(compareLaneOrder(extended, key([0, 0, 0]))).toBe(0);
    expect(extended.unrelated).toBe("not ordering authority");
  });
});
