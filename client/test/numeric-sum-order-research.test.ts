import Decimal from "break_infinity.js";
import { expect, it } from "vitest";
import fixture from "../../testdata/numeric-sum-order-research-v1.json";
import canonicalCarry from "../../testdata/numeric-canonical-group-carry-v1.json";
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

it("retains a canonical partial-term group-carry counterexample without claiming temporal reachability", () => {
  expect(canonicalCarry.version).toBe(1);
  expect(canonicalCarry.repeat_count).toBe(65535);
  const prefix = Array.from({ length: canonicalCarry.repeat_count }, () => parseCanonical(canonicalCarry.repeat_value));
  prefix.push(parseCanonical(canonicalCarry.fixed_last));
  const sum = (variable: string) => {
    const values = [...prefix, parseCanonical(variable)];
    const before = values.map((value) => [value.mantissa, value.exponent]);
    const result = sumDeterministic(values);
    expect(isStateValue(result)).toBe(true);
    expect(values.map((value) => [value.mantissa, value.exponent])).toEqual(before);
    const reversed = sumDeterministic([...values].reverse());
    expect([reversed.mantissa, reversed.exponent]).toEqual([result.mantissa, result.exponent]);
    return result;
  };
  expect(parseCanonical(canonicalCarry.variable_upper).gt(parseCanonical(canonicalCarry.variable_lower))).toBe(true);
  const lower = sum(canonicalCarry.variable_lower), upper = sum(canonicalCarry.variable_upper);
  expect(lower.gt(upper)).toBe(true);
  expect(canonicalString(lower)).toBe(canonicalCarry.lower_payout);
  expect(canonicalString(upper)).toBe(canonicalCarry.upper_payout);
  expect(parseCanonical(canonicalString(lower)).gt(parseCanonical(canonicalString(upper)))).toBe(true);
  // Exact decimal-literal reference at a fixed 1e-11 scale, independent of Decimal.
  const scaled = (source: string): bigint => {
    const [coefficient, exponent] = source.split("e");
    const decimals = coefficient!.split(".")[1]?.length ?? 0;
    const shift = 11 + Number(exponent) - decimals;
    if (!Number.isSafeInteger(shift) || shift < 0 || shift > 16) throw new Error("unsupported exact reference scale");
    return BigInt(coefficient!.replace(".", "")) * 10n ** BigInt(shift);
  };
  const prefixLiteral = BigInt(canonicalCarry.repeat_count) * scaled(canonicalCarry.repeat_value) + scaled(canonicalCarry.fixed_last);
  const lowerLiteral = prefixLiteral + scaled(canonicalCarry.variable_lower);
  expect(lowerLiteral.toString()).toBe(canonicalCarry.literal_lower_numerator);
  const upperLiteral = prefixLiteral + scaled(canonicalCarry.variable_upper);
  expect(upperLiteral - lowerLiteral).toBe(1n);
  expect(upperLiteral.toString()).toBe(canonicalCarry.literal_upper_numerator);
  expect(canonicalCarry.literal_sum_denominator).toBe("100000000000");
  expect(sum("1e0").lt(sum("2e0"))).toBe(true);
  expect(sum(canonicalCarry.variable_upper).eq(upper)).toBe(true);
});
