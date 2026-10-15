import Decimal from "break_infinity.js";

import { classify, isStateValue, MAX_EXACT_INTEGER, parseCanonical, quantize, sumDeterministic } from "./numeric";

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

  const rates = rateSources.map((source) => parseCanonical(source));
  for (const rate of rates) {
    if (!isStateValue(rate) || rate.lt(0)) throw new RangeError("rates must be non-negative state values");
  }
  const totalRate = sumDeterministic(rates);
  if (!isStateValue(totalRate)) throw new RangeError("rate sum is outside the Decimal domain");
  if (efficiency.eq(0)) return new Decimal(0);
  const seconds = new Decimal(elapsedMilliseconds).div(1000);
  const accrued = totalRate.mul(seconds);
  // The pinned library can scale its infinity-class sentinel back into range.
  // Go refuses that intermediate; finite lower-domain recovery remains valid.
  if (classify(accrued) !== "finite") throw new RangeError("accrual intermediate is non-finite");
  const delta = quantize(accrued.mul(efficiency));
  if (!isStateValue(delta)) throw new RangeError("accrual result is outside the Decimal domain");
  return delta;
}
