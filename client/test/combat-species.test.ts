import { describe, expect, it } from "vitest";

import fixture from "../../testdata/combat/species-vectors.json";
import identities from "../../testdata/combat/identity-vectors.json";
import { decodeSpeciesRecord } from "../src/combat/species";

describe("Combat C2 species record", () => {
  it("executes the complete shared structural population", () => {
    expect(fixture.version).toBe(1);
    expect(fixture.cases).toHaveLength(75);
  });

  it.each(fixture.cases)("$name", (vector) => {
    const declarations = new Set(fixture.declarations);
    if (vector.expected === null) {
      expect(() => decodeSpeciesRecord(vector.source, declarations)).toThrow(SyntaxError);
    } else {
      const record = decodeSpeciesRecord(vector.source, declarations);
      expect(record).toEqual(vector.expected);
      expect(decodeSpeciesRecord(JSON.stringify(record), declarations)).toEqual(record);
    }
  });

  it("uses every existing temperament/runtime identity without a new chart", () => {
    const base = fixture.cases[0]!.expected!;
    const temperaments = ["lazy", "playful", "curious", "sassy", "shy", "chaotic"];
    let pairs = 0;
    for (const temperament of temperaments) for (const runtime of identities.runtimes) {
      const row = { ...base, temperament, runtime };
      expect(decodeSpeciesRecord(JSON.stringify(row), new Set(fixture.declarations))).toEqual(row);
      pairs++;
    }
    expect(pairs).toBe(36);
  });

  it("owns frozen stats, four slots and tags without changing the declaration set", () => {
    const vector = fixture.cases.find(row => row.name === "ordered repeated references and tags")!;
    const declarations = new Set(fixture.declarations);
    const record = decodeSpeciesRecord(vector.source, declarations);
    expect(Object.isFrozen(record)).toBe(true);
    expect(Object.isFrozen(record.stats)).toBe(true);
    expect(Object.isFrozen(record.moves)).toBe(true);
    expect(Object.isFrozen(record.tags)).toBe(true);
    expect(() => { (record.stats as { hp: number }).hp = 1; }).toThrow(TypeError);
    expect(() => { (record.moves as unknown as string[])[0] = "changed"; }).toThrow(TypeError);
    expect(() => { (record.tags as string[])[0] = "changed"; }).toThrow(TypeError);
    expect(declarations).toEqual(new Set(fixture.declarations));
    declarations.delete("beta");
    expect(() => decodeSpeciesRecord(vector.source, declarations)).toThrow(SyntaxError);
    expect(() => decodeSpeciesRecord(vector.source, new Set())).toThrow(SyntaxError);
    expect(record).toEqual(vector.expected);
  });
});
