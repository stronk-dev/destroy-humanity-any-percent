import { describe, expect, it } from "vitest";
import fixture from "../../testdata/combat/lane-record-vectors.json";
import { decodeLaneRecord } from "../src/combat/lane-record";

describe("Combat C2 structural lane records", () => {
  it("executes the complete shared population", () => {
    expect(fixture.version).toBe(1);
    expect(fixture.cases).toHaveLength(100);
    expect(fixture.cases.filter(row => row.expected !== null)).toHaveLength(8);
  });

  it.each(fixture.cases)("$name", vector => {
    if (vector.expected === null) {
      expect(() => decodeLaneRecord(vector.source)).toThrow(SyntaxError);
    } else {
      const record = decodeLaneRecord(vector.source);
      expect(record).toEqual(vector.expected);
      expect(decodeLaneRecord(JSON.stringify(record))).toEqual(record);
    }
  });

  it("preserves every temperament/target and absent/false/true marker", () => {
    const base = fixture.cases[0]!.expected!;
    let cases = 0;
    for (const temperament of ["lazy", "playful", "curious", "sassy", "shy", "chaotic"]) {
      for (const targets of ["units", "buildings", "any"]) for (const bypass of [undefined, false, true]) {
        const row = { ...base, temperament, targets, ...(bypass === undefined ? {} : { bypass }) };
        const actual = decodeLaneRecord(JSON.stringify(row));
        expect(actual).toEqual(row);
        expect(Object.hasOwn(actual, "bypass")).toBe(bypass !== undefined);
        cases++;
      }
    }
    expect(cases).toBe(54);
  });

  it("returns fresh frozen records/tags without sharing caller storage", () => {
    const vector = fixture.cases.find(row => row.name === "building target and bypass")!;
    const record = decodeLaneRecord(vector.source);
    const next = decodeLaneRecord(vector.source);
    expect(record).not.toBe(next);
    expect(record.tags).not.toBe(next.tags);
    expect(Object.isFrozen(record)).toBe(true);
    expect(Object.isFrozen(record.tags)).toBe(true);
    expect(() => { (record as { hp: number }).hp = 0; }).toThrow(TypeError);
    expect(() => { (record as { bypass: boolean }).bypass = false; }).toThrow(TypeError);
    expect(() => { (record.tags as string[])[0] = "changed"; }).toThrow(TypeError);
    expect(next).toEqual(vector.expected);
    expect(decodeLaneRecord(vector.source)).toEqual(vector.expected);
  });
});
