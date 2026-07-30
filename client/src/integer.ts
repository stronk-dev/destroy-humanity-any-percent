export function idiv(value: bigint, divisor: bigint): bigint {
  if (value < 0n || divisor <= 0n) throw new RangeError("idiv requires non-negative value and positive divisor");
  return value / divisor;
}
