import Decimal from "break_infinity.js";

import { canonicalString, isStateValue, MAX_EXACT_INTEGER, parseCanonical, quantize } from "./numeric";

export function accrueConstant(
  rateSources: readonly string[],
  elapsedMilliseconds: number,
  efficiencySource: string,
): Decimal {
  if (
    !Number.isSafeInteger(elapsedMilliseconds) ||
    elapsedMilliseconds < 0 ||
    elapsedMilliseconds > MAX_EXACT_INTEGER
  ) {
    throw new RangeError("elapsed milliseconds must be an exact non-negative integer");
  }
  const efficiency = parseCanonical(efficiencySource);
  if (efficiency.lt(0)) throw new RangeError("efficiency must be non-negative");

  const ordered = rateSources.map((source) => parseCanonical(source));
  for (const rate of ordered) {
    if (!isStateValue(rate) || rate.lt(0)) throw new RangeError("rates must be non-negative state values");
  }
  ordered.sort((left, right) => {
    if (left.exponent !== right.exponent) return left.exponent - right.exponent;
    return canonicalString(left).localeCompare(canonicalString(right));
  });
  let totalRate = new Decimal(0);
  for (const rate of ordered) {
    totalRate = totalRate.add(rate);
    if (!isStateValue(totalRate)) throw new RangeError("rate sum is outside the Decimal domain");
  }
  const seconds = new Decimal(elapsedMilliseconds).div(1000);
  const delta = quantize(totalRate.mul(seconds).mul(efficiency));
  if (!isStateValue(delta)) throw new RangeError("accrual result is outside the Decimal domain");
  return delta;
}
