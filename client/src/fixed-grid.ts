import { MAX_EXACT_INTEGER } from "./numeric";

export interface FixedGridResult {
  readonly whole: bigint;
  readonly remainder: number;
}

export function integrateFixedGrid(units: number, rate: number, remainder: number, divisor: number): FixedGridResult {
  if (!Number.isSafeInteger(units) || units < 0 || units > MAX_EXACT_INTEGER ||
    !Number.isSafeInteger(rate) || rate < 0 || rate > MAX_EXACT_INTEGER ||
    !Number.isSafeInteger(remainder) || remainder < 0 ||
    !Number.isSafeInteger(divisor) || divisor < 1 || divisor > MAX_EXACT_INTEGER || remainder >= divisor) {
    throw new RangeError("invalid fixed-grid input");
  }
  const numerator = BigInt(units) * BigInt(rate) + BigInt(remainder);
  return { whole: numerator / BigInt(divisor), remainder: Number(numerator % BigInt(divisor)) };
}
