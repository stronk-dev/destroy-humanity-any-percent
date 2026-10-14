import { describe, expect, it } from "vitest";

import vectors from "../../testdata/combat/substream-sequences.json";
import { battleSeed, substream } from "../src/combat/rng";

type Stream = { next(): bigint; bound(bound: bigint): bigint };
type Factory = (seed: bigint, label: string) => Stream;

function observe(seed: bigint, pattern: string, factory: Factory): Record<string, string[]> {
  const streams = new Map(vectors.labels.map(label => [label, factory(seed, label)]));
  const result = Object.fromEntries(vectors.labels.map(label => [label, [] as string[]]));
  const growth = pattern === "plain" ? undefined : factory(seed, "new_consumer");
  const consume = (label: string) => {
    result[label]!.push(streams.get(label)!.next().toString());
    if (growth) for (const bound of [1n, 7n, (1n << 63n) + 1n]) growth.bound(bound);
  };
  const order = pattern === "reverse" ? [...vectors.labels].reverse() : vectors.labels;
  if (pattern === "chunked") {
    for (const label of order) for (let draw = 0; draw < vectors.draw_count; draw += 1) consume(label);
  } else {
    for (let draw = 0; draw < vectors.draw_count; draw += 1) for (const label of order) consume(label);
  }
  return result;
}

describe("combat labeled stream sequence isolation", () => {
  it("retains the complete shared seed/label/draw population", () => {
    expect(vectors.version).toBe(1);
    expect(vectors.labels).toEqual(["crit", "obedience", "bot_policy"]);
    expect(vectors.draw_count).toBe(8);
    expect(vectors.seeds.map(row => row.match_seed)).toEqual(["0", "42", "18446744073709551615"]);
    for (const row of vectors.seeds) {
      expect(Object.keys(row.streams)).toEqual(vectors.labels);
      for (const values of Object.values(row.streams)) expect(values).toHaveLength(vectors.draw_count);
    }
  });
  it.each(vectors.seeds.flatMap(row => ["plain", "growth", "reverse", "chunked"].map(pattern => ({ ...row, pattern }))))(
    "preserves all eight draws at seed $match_seed with $pattern consumption", row => {
      const seed = battleSeed(BigInt(row.match_seed));
      expect(seed.toString()).toBe(row.battle_seed);
      expect(observe(seed, row.pattern, substream)).toEqual(row.streams);
    },
  );
  it.each(vectors.seeds)("rejects shared resettable state even when seed $match_seed first-draw checks pass", row => {
    const seed = BigInt(row.battle_seed);
    let current: Stream;
    const shared: Stream = { next: () => current.next(), bound: bound => current.bound(bound) };
    const badFactory: Factory = (seed, label) => { current = substream(seed, label); return shared; };
    for (const label of vectors.labels) {
      expect(badFactory(seed, label).next().toString()).toBe(row.streams[label as keyof typeof row.streams][0]);
    }
    const before = badFactory(seed, "crit").next();
    badFactory(seed, "new_consumer").next();
    expect(badFactory(seed, "crit").next()).toBe(before);
    expect(before.toString()).toBe(row.streams.crit[0]);
    expect(observe(seed, "growth", badFactory)).not.toEqual(row.streams);
  });
});
