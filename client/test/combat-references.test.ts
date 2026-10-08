import { describe, expect, it } from "vitest";

import fixture from "../../testdata/combat/reference-vectors.json";
import { decodeDeckReferences, decodeTeamReferences } from "../src/combat/references";

describe("Combat C2 reference envelopes", () => {
  it("uses the complete shared v1 population", () => {
    expect(fixture.version).toBe(1);
    expect(fixture.cases).toHaveLength(32);
  });
  it.each(fixture.cases)("$name", (vector) => {
    const decode = vector.kind === "team" ? decodeTeamReferences : decodeDeckReferences;
    const declarations = new Set(fixture.declarations);
    if (vector.expected === null) expect(() => decode(vector.source, declarations)).toThrow(SyntaxError);
    else expect(decode(vector.source, declarations)).toEqual(vector.expected);
  });
  it("preserves source bytes and isolates the returned slots and declaration set", () => {
    const source = '{"v":1,"ids":["gamma","alpha","beta"]}';
    const declarations = new Set(["alpha", "beta", "gamma"]);
    const team = decodeTeamReferences(source, declarations);
    expect(Object.isFrozen(team)).toBe(true);
    expect(() => { (team as unknown as string[])[0] = "changed"; }).toThrow(TypeError);
    expect(declarations).toEqual(new Set(["alpha", "beta", "gamma"]));
    declarations.delete("alpha");
    expect(() => decodeTeamReferences(source, declarations)).toThrow(SyntaxError);
    expect(() => decodeTeamReferences(source, new Set())).toThrow(SyntaxError);
    expect(team).toEqual(["gamma", "alpha", "beta"]);
    expect(source).toBe('{"v":1,"ids":["gamma","alpha","beta"]}');
  });
});
