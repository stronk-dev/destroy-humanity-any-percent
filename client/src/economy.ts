import Decimal from "break_infinity.js";

import { isStateValue, MAX_EXACT_INTEGER, type DecimalSource } from "./numeric";

function validCount(value: number): boolean {
  return Number.isSafeInteger(value) && value >= 0 && value <= MAX_EXACT_INTEGER;
}

function validateInputs(
  cash: Decimal,
  base: Decimal,
  ratio: Decimal,
  owned: number,
): void {
  if (
    !validCount(owned) ||
    !isStateValue(cash) ||
    !isStateValue(base) ||
    !isStateValue(ratio) ||
    cash.lt(0) ||
    !base.gt(0) ||
    ratio.lt(1)
  ) {
    throw new RangeError("invalid geometric-series input");
  }
}

export function sumGeometricSeries(
  count: number,
  baseSource: DecimalSource,
  ratioSource: DecimalSource,
  owned: number,
): Decimal {
  const base = new Decimal(baseSource);
  const ratio = new Decimal(ratioSource);
  if (!validCount(count) || !validCount(owned) || !isStateValue(base) || !isStateValue(ratio) || base.lt(0) || ratio.lt(1)) {
    return new Decimal(Number.NaN);
  }
  if (count === 0 || base.eq(0)) return new Decimal(0);
  if (ratio.eq(1)) return base.mul(count);

  const start = base.mul(ratio.pow(owned));
  return start.mul(Decimal.sub(1, ratio.pow(count))).div(Decimal.sub(1, ratio));
}

function affordable(
  count: number,
  cash: Decimal,
  base: Decimal,
  ratio: Decimal,
  owned: number,
): boolean {
  const cost = sumGeometricSeries(count, base, ratio, owned);
  return isStateValue(cost) && cost.lte(cash);
}

function binarySearchAffordable(
  cash: Decimal,
  base: Decimal,
  ratio: Decimal,
  owned: number,
): number {
  if (affordable(MAX_EXACT_INTEGER, cash, base, ratio, owned)) return MAX_EXACT_INTEGER;
  let low = 0;
  let high = MAX_EXACT_INTEGER;
  while (low < high) {
    const mid = low + Math.floor((high - low + 1) / 2);
    if (affordable(mid, cash, base, ratio, owned)) low = mid;
    else high = mid - 1;
  }
  return low;
}

export function affordGeometricSeries(
  cashSource: DecimalSource,
  baseSource: DecimalSource,
  ratioSource: DecimalSource,
  owned: number,
): number {
  const cash = new Decimal(cashSource);
  const base = new Decimal(baseSource);
  const ratio = new Decimal(ratioSource);
  validateInputs(cash, base, ratio, owned);

  if (!affordable(1, cash, base, ratio, owned)) return 0;
  if (ratio.eq(1)) {
    return Math.min(Math.floor(cash.div(base).toNumber()), MAX_EXACT_INTEGER);
  }

  const start = base.mul(ratio.pow(owned));
  const estimate = Math.floor(cash
    .div(start)
    .mul(ratio.sub(1))
    .add(1)
    .log10()
    / ratio.log10());
  if (!Number.isSafeInteger(estimate) || estimate < 0) {
    return binarySearchAffordable(cash, base, ratio, owned);
  }

  let candidate = Math.min(estimate, MAX_EXACT_INTEGER);
  for (let correction = 0; correction < 8 && candidate > 0 && !affordable(candidate, cash, base, ratio, owned); correction++) {
    candidate -= 1;
  }
  for (
    let correction = 0;
    correction < 8 && candidate < MAX_EXACT_INTEGER && affordable(candidate + 1, cash, base, ratio, owned);
    correction++
  ) {
    candidate += 1;
  }

  if (!affordable(candidate, cash, base, ratio, owned) ||
      (candidate < MAX_EXACT_INTEGER && affordable(candidate + 1, cash, base, ratio, owned))) {
    return binarySearchAffordable(cash, base, ratio, owned);
  }
  return candidate;
}
