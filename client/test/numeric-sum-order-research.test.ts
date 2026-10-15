import Decimal from "break_infinity.js";
import { expect, it } from "vitest";
import fixture from "../../testdata/numeric-sum-order-research-v1.json";
import { canonicalString, isStateValue, parseCanonical, sumDeterministic } from "../src/numeric";

// Retained negative evidence, not approved decreasing-production semantics.
// This exercises the numeric producer prefix, not the full replay evaluator.
it("retains the positive-rate regrouping counterexample with canonical catalog inputs", () => {
  expect(fixture.version).toBe(1);
  expect(fixture.fixed).toHaveLength(4);
  expect(fixture.variable.upper_count - fixture.variable.lower_count).toBe(1);
  const fixed = fixture.fixed.map((row) => {
    const actual = parseCanonical(row.base_rate).mul(new Decimal(row.count));
    expect(isStateValue(actual)).toBe(true);
    expect(actual.eq(new Decimal(row.rate))).toBe(true);
    return actual;
  });
  const payouts = [fixture.variable.lower_count, fixture.variable.upper_count].map((count, index) => {
    expect(Number.isSafeInteger(count)).toBe(true);
    const variable = parseCanonical(fixture.variable.base_rate).mul(new Decimal(count));
    expect(variable.eq(new Decimal(index === 0 ? fixture.variable.lower_rate : fixture.variable.upper_rate))).toBe(true);
    const sources = [...fixed, variable];
    const before = sources.map((value) => [value.mantissa, value.exponent]);
    const result = sumDeterministic(sources);
    expect(canonicalString(result)).toBe(index === 0 ? fixture.lower_payout : fixture.upper_payout);
    expect(sources.map((value) => [value.mantissa, value.exponent])).toEqual(before);
    return parseCanonical(canonicalString(result));
  });
  expect(payouts[0]!.gt(payouts[1]!)).toBe(true);
  // Ordinary positive/equal controls do not all decrease.
  expect(sumDeterministic([new Decimal(1), ...fixed]).lt(sumDeterministic([new Decimal(2), ...fixed]))).toBe(true);
  expect(sumDeterministic([new Decimal(10), ...fixed]).eq(sumDeterministic([...fixed, new Decimal(10)]))).toBe(true);
});
