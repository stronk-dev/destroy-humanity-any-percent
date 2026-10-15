import Decimal from "break_infinity.js";
import { describe, expect, it } from "vitest";

import fixture from "../../testdata/numeric-monotonicity-v1.json";
import { canonicalString, isStateValue, quantize, sumDeterministic } from "../src/numeric";

// No tolerance: raw reversals cannot be hidden by canonical serialization.
function orderedPair(lower: Decimal, upper: Decimal, strict: boolean): boolean {
  return isStateValue(lower) && isStateValue(upper) &&
    (strict ? lower.lt(upper) : lower.lte(upper));
}

describe("ordered arithmetic transforms", () => {
  it("retains the shared literal population", () => {
    expect(fixture.version).toBe(1);
    expect(fixture.cases).toHaveLength(18);
  });

  for (const row of fixture.cases) {
    const strict = row.strict === true;
    for (const name of ["add", "multiply", "divide", "quantize", "sum"] as const) {
      it(`${row.name}/${name}`, () => {
        const lower = new Decimal(row.lower), upper = new Decimal(row.upper);
        const addend = new Decimal(row.addend), scale = new Decimal(row.scale);
        const terms = row.terms.map((source) => new Decimal(source));
        const inputs = [lower, upper, addend, scale, ...terms];
        const before = inputs.map((value) => [value.mantissa, value.exponent]);
        expect(inputs.every(isStateValue)).toBe(true);
        expect(lower.cmp(upper)).toBe(row.equal ? 0 : -1);
        expect(scale.gt(0)).toBe(true);
        expect(Boolean(row.equal && strict)).toBe(false);
        const run = (value: Decimal): Decimal => {
          switch (name) {
            case "add": return value.add(addend);
            case "multiply": return value.mul(scale);
            case "divide": return value.div(scale);
            case "quantize": return quantize(value);
            case "sum": return sumDeterministic([value, ...terms]);
          }
        };
        const a = run(lower), b = run(upper);
        expect(orderedPair(a, b, strict), `raw (${a.mantissa},e${a.exponent}) / (${b.mantissa},e${b.exponent})`).toBe(true);
        expect(orderedPair(quantize(a), quantize(b), strict), "committed order").toBe(true);
        if (row.equal) expect(a.eq(b)).toBe(true);
        expect(inputs.map((value) => [value.mantissa, value.exponent])).toEqual(before);
      });
    }
  }

  it.each([
    ["increasing", new Decimal(1), new Decimal(2), true, true],
    ["rounding-plateau", new Decimal(1), new Decimal(1), false, true],
    ["reversed", new Decimal(2), new Decimal(1), false, false],
    ["reversed-below-canonical-resolution", new Decimal("1.000000000000001"), new Decimal(1), false, false],
    ["constant-shortcut", new Decimal(0), new Decimal(0), true, false],
    ["nan", new Decimal(Number.NaN), new Decimal(1), false, false],
    ["infinity", new Decimal(1), new Decimal(Number.POSITIVE_INFINITY), false, false],
    ["unnormalized", Decimal.fromMantissaExponent_noNormalize(10, 0), new Decimal(20), false, false],
  ] as const)("order control: %s", (_name, lower, upper, strict, pass) => {
    expect(orderedPair(lower, upper, strict)).toBe(pass);
  });

  it("rejects the original out-of-domain division case rather than calling it saved state", () => {
    const lower = new Decimal("1e-8999999999999999").div(2);
    const upper = new Decimal("2e-8999999999999999").div(2);
    expect([lower.mantissa, lower.exponent]).toEqual([5, -9_000_000_000_000_000]);
    expect(isStateValue(lower)).toBe(false);
    expect(() => canonicalString(lower)).toThrow(RangeError);
    expect(orderedPair(lower, upper, false)).toBe(false);
    expect(canonicalString(upper)).toBe("1e-8999999999999999");
  });
});
