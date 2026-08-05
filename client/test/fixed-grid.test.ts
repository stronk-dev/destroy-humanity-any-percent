import { describe, expect, it } from "vitest";
import fixture from "../../testdata/fixed-grid-v1.json";
import { integrateFixedGrid } from "../src/fixed-grid";

describe("fixed-grid integrator", () => {
  it("matches the shared Go/TypeScript vectors", () => {
    for (const testCase of fixture.cases) {
      if (!testCase.valid) {
        expect(() => integrateFixedGrid(testCase.units, testCase.rate, testCase.remainder, testCase.divisor), testCase.name).toThrow(RangeError);
        continue;
      }
      const actual = integrateFixedGrid(testCase.units, testCase.rate, testCase.remainder, testCase.divisor);
      expect({ whole: actual.whole.toString(), next_remainder: actual.remainder }, testCase.name).toEqual({
        whole: testCase.whole,
        next_remainder: testCase.next_remainder,
      });
    }
  });

  it("is partition invariant across a large attended interval", () => {
    const combined = integrateFixedGrid(9_000_000_000_000, 333_333, 777_777, 1_000_003);
    const first = integrateFixedGrid(4_000_000_000_000, 333_333, 777_777, 1_000_003);
    const second = integrateFixedGrid(5_000_000_000_000, 333_333, first.remainder, 1_000_003);
    expect({ whole: first.whole + second.whole, remainder: second.remainder }).toEqual(combined);
  });
});
