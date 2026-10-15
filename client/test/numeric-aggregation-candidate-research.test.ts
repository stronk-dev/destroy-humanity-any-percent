import Decimal from "break_infinity.js";
import { expect, it } from "vitest";
import fixture from "../../testdata/numeric-sum-order-research-v1.json";
import accrual from "../../testdata/production-accrual.json";
import { canonicalString, isStateValue, quantize, sumDeterministic } from "../src/numeric";

// Unapplied research counterpart, not a production helper or adopted algorithm.
function alignedCompensated(values: readonly Decimal[]): Decimal {
  if (values.some((value) => !isStateValue(value))) return new Decimal(NaN);
  const nonzero = values.filter((value) => value.mantissa !== 0);
  if (!nonzero.length) return new Decimal(0);
  const exponent = Math.max(...nonzero.map((value) => value.exponent));
  const ordered = [...values].sort((a, b) => a.exponent - b.exponent || Math.abs(a.mantissa) - Math.abs(b.mantissa) || a.mantissa - b.mantissa);
  let sum = 0, correction = 0;
  for (const value of ordered) {
    const x = value.mantissa * 10 ** (value.exponent - exponent);
    const next = sum + x;
    correction += Math.abs(sum) >= Math.abs(x) ? (sum - next) + x : (x - next) + sum;
    sum = next;
  }
  const result = Decimal.fromMantissaExponent(sum + correction, exponent);
  return isStateValue(result) ? result : new Decimal(NaN);
}

it("compares an unapplied aligned candidate with the retained midpoint pair", () => {
  const fixed = fixture.fixed.map((row) => new Decimal(row.rate));
  for (const variable of [fixture.variable.lower_rate, fixture.variable.upper_rate]) {
    const values = [...fixed, new Decimal(variable)];
    const before = values.map((value) => [value.mantissa, value.exponent]);
    const result = alignedCompensated(values);
    expect(canonicalString(result)).toBe("1.01245957117e1");
    expect(canonicalString(alignedCompensated([...values].reverse()))).toBe(canonicalString(result));
    expect(values.map((value) => [value.mantissa, value.exponent])).toEqual(before);
  }
  const boundary = new Decimal("9e8999999999999999");
  expect(alignedCompensated([boundary, boundary, boundary.neg(), boundary.neg()]).eq(0)).toBe(true);
  expect(isStateValue(alignedCompensated([boundary, boundary]))).toBe(false);
  expect(isStateValue(alignedCompensated([new Decimal(NaN)]))).toBe(false);
});

it("checks existing valid accrual fixture formula results without claiming a client caller migration", () => {
  let checked = 0;
  for (const row of accrual.vectors) {
    if (row.expect_error) continue; // This experiment substitutes only the sum, not input admission.
    const rates = row.rates.map((source) => new Decimal(source));
    const result = quantize(alignedCompensated(rates).mul(new Decimal(row.elapsed_ms).div(1000)).mul(new Decimal(row.efficiency)));
    expect(canonicalString(result), row.name).toBe(row.expect);
    checked++;
  }
  expect(checked).toBe(11);
});

it("completes the declared seeded positive carry population, reporting rather than assuming monotonicity", () => {
  let seed = 663n;
  const mask = (1n << 64n) - 1n;
  const next = () => {
    seed = (seed ^ (seed << 13n)) & mask;
    seed ^= seed >> 7n;
    seed = (seed ^ (seed << 17n)) & mask;
    return seed;
  };
  const counts = [{ raw: 0, committed: 0 }, { raw: 0, committed: 0 }];
  for (let index = 0; index < 100_000; index++) {
    const fixed = Array.from({ length: 4 }, () => Decimal.fromMantissaExponent(Number(100_000_000_000_000n + next() % 900_000_000_000_000n) / 1e14, Number(next() % 14n) - 5));
    const lower = [...fixed, new Decimal("9.999999999999998e0")];
    const upper = [...fixed, new Decimal("1e1")];
    [sumDeterministic, alignedCompensated].forEach((sum, arm) => {
      const a = sum(lower), b = sum(upper);
      if (!isStateValue(a) || !isStateValue(b)) throw new Error(`incomplete measurement at ${index}/${arm}`);
      if (a.gt(b)) counts[arm]!.raw++;
      if (quantize(a).gt(quantize(b))) counts[arm]!.committed++;
    });
  }
  // Raw census is runtime-specific: raw continuous results have a tolerance in
  // RFC-0001. Do not turn observed Go counts into an exact JS contract.
  console.info("aggregation-candidate-research-v1", JSON.stringify({ population: 100_000, seed: 663, counts }));
  expect(counts[0]!.raw).toBeGreaterThan(0);
  expect(counts[0]!.committed).toBe(1);
  expect(counts[1]).toEqual({ raw: 0, committed: 0 });
});

it("rejects the aligned candidate because exact large-term cancellation erases a retained tiny term", () => {
  const tiny = new Decimal("1e-300"), large = new Decimal("1e300");
  const values = [large, large.neg(), tiny];
  expect(canonicalString(sumDeterministic(values))).toBe("1e-300");
  expect(canonicalString(alignedCompensated(values))).toBe("0");
});
