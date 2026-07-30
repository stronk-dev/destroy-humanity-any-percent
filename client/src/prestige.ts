import Decimal from "break_infinity.js";

import { MAX_EXACT_INTEGER, parseCanonical } from "./numeric";

export function reputationLevel(lifetimeValueSource: string, thresholdSource: string): number {
  const lifetimeValue = parseCanonical(lifetimeValueSource);
  const threshold = parseCanonical(thresholdSource);
  if (lifetimeValue.lt(0) || !threshold.gt(0)) throw new RangeError("invalid prestige values");
  const ratio = Decimal.fromMantissaExponent(
    lifetimeValue.mantissa / threshold.mantissa,
    lifetimeValue.exponent - threshold.exponent,
  ).normalize();
  if (ratio.lt(1)) return 0;
  let low = 1;
  let high = MAX_EXACT_INTEGER;
  while (low < high) {
    const mid = low + Math.floor((high - low + 1) / 2);
    const candidate = new Decimal(mid);
    const cube = candidate.mul(candidate).mul(candidate);
    if (cube.lte(ratio)) low = mid;
    else high = mid - 1;
  }
  return low;
}

export function reputationDelta(
  lifetimeValue: string,
  threshold: string,
  currentLevel: number,
  modifierPPM: number,
): number {
  if (
    !Number.isSafeInteger(currentLevel) ||
    currentLevel < 0 ||
    !Number.isSafeInteger(modifierPPM) ||
    modifierPPM < 0
  ) {
    throw new RangeError("invalid prestige integers");
  }
  const delta = Math.max(0, reputationLevel(lifetimeValue, threshold) - currentLevel);
  const exact = (BigInt(delta) * BigInt(modifierPPM)) / 1_000_000n;
  const cap = BigInt(MAX_EXACT_INTEGER);
  if (exact >= cap) return MAX_EXACT_INTEGER;
  return Number(exact);
}

export function moralReseed(notoriety: number): number {
  if (!Number.isSafeInteger(notoriety) || notoriety < 0) throw new RangeError("invalid notoriety");
  if (notoriety >= 100) return 55;
  return Math.max(55, Math.min(90, 90 - Math.floor((notoriety * 35) / 100)));
}
