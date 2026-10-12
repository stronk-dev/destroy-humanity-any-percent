import { describe, expect, it } from "vitest";
import { canonicalString, isStateValue, parseCanonical, quantize } from "../src/numeric";

// Literal pairs observed from Go's actual Company Evaluate/codec/ledger census.
// This checks client arithmetic only, not a TS Company evaluator or a ledger.
const pairs = [
  {
    id: "floor online refusal",
    before: "1.001e-8999999999999999",
    target: "1.002001e-8999999999999999",
    delta: null,
  },
  {
    id: "near-floor online positive control",
    before: "1.001e-8999999999999996",
    target: "1.002001e-8999999999999996",
    delta: "1.001e-8999999999999999",
  },
  {
    id: "near-floor offline refusal",
    before: "9.009e-8999999999999997",
    target: "9.018009e-8999999999999997",
    delta: null,
  },
  {
    id: "ordinary online positive control",
    before: "1.001e0",
    target: "1.002001e0",
    delta: "1.001e-3",
  },
  {
    id: "ordinary offline positive control",
    before: "9.009e-1",
    target: "9.018009e-1",
    delta: "9.009e-4",
  },
] as const;

describe("R-012 origin delta lower-domain research, not AC6 acceptance", () => {
  it("zero and the minimum positive delta bracket the retained floor target", () => {
    const before = parseCanonical("1.001e-8999999999999999");
    const target = parseCanonical("1.002001e-8999999999999999");
    const minimum = parseCanonical("1e-8999999999999999");
    expect(canonicalString(quantize(before.add(parseCanonical("0"))))).toBe("1.001e-8999999999999999");
    expect(canonicalString(quantize(before.add(minimum)))).toBe("2.001e-8999999999999999");
    expect(quantize(before.add(minimum)).gt(target)).toBe(true);
    expect(() => parseCanonical("1e-9000000000000000")).toThrow(SyntaxError);
    // A one-shot settlement from the original zero balance is representable.
    expect(canonicalString(quantize(parseCanonical("0").add(target)))).toBe("1.002001e-8999999999999999");
  });
  for (const pair of pairs) {
    it(pair.id, () => {
      const before = parseCanonical(pair.before);
      const target = parseCanonical(pair.target);
      expect(isStateValue(before)).toBe(true);
      expect(isStateValue(target)).toBe(true);
      expect(canonicalString(before)).toBe(pair.before);
      expect(canonicalString(target)).toBe(pair.target);
      expect(target.gt(before)).toBe(true);

      const delta = quantize(target.sub(before));
      console.info(pair.id, {
        mantissa: delta.mantissa,
        exponent: delta.exponent,
        admitted: isStateValue(delta),
      });
      if (pair.delta === null) {
        expect(isStateValue(delta)).toBe(false);
        expect(() => canonicalString(delta)).toThrow(RangeError);
      } else {
        expect(isStateValue(delta)).toBe(true);
        expect(canonicalString(delta)).toBe(pair.delta);
        expect(canonicalString(quantize(before.add(delta)))).toBe(pair.target);
      }
    });
  }
});
