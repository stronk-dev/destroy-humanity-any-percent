import Decimal from "break_infinity.js";
import { describe, expect, it } from "vitest";

import fixture from "../../testdata/numeric-lower-domain-v1.json";
import { canonicalString, classify, isStateValue, parseCanonical, quantize } from "../src/numeric";

describe("shared raw lower-domain arithmetic diagnostics", () => {
  it("retains the complete literal population", () => {
    expect(fixture.version).toBe(1);
    expect(fixture.cases).toHaveLength(12);
  });
  for (const row of fixture.cases) {
    it(row.name, () => {
      const a = new Decimal(row.a), b = new Decimal(row.b);
      expect(isStateValue(a) && isStateValue(b)).toBe(true);
      const before = [a.mantissa, a.exponent, b.mantissa, b.exponent];
      let got: Decimal;
      switch (row.op) {
        case "mul": got = a.mul(b); break;
        case "div": got = a.div(b); break;
        case "pow": got = a.pow(Number(row.b)); break;
        default: throw new Error(`unknown operation ${row.op}`);
      }
      // Library display rounds these diagnostics to "0"; raw fields are the
      // oracle. Exponents are strings in the fixture to avoid JSON precision loss.
      expect(got.mantissa).toBe(row.mantissa);
      expect(BigInt(got.exponent).toString()).toBe(row.exponent);
      expect(classify(got)).toBe("finite");
      expect(isStateValue(got)).toBe(row.state);
      if (row.state) {
        expect(parseCanonical(canonicalString(got)).eq(quantize(got))).toBe(true);
      } else {
        expect(got.eq(0)).toBe(false);
        expect(isStateValue(quantize(got))).toBe(false);
        expect(() => canonicalString(got)).toThrow(RangeError);
      }
      expect([a.mantissa, a.exponent, b.mantissa, b.exponent]).toEqual(before);
    });
  }
});
