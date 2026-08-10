import { StandardNotation } from "@antimatter-dimensions/notations";

import { parseCanonical, quantize } from "../numeric";

export const AMOUNT_MANTISSA_DIGITS = 3;
export const AMOUNT_UNDER_1000_PLACES = 0;

const standard = new StandardNotation();

export function formatAmount(value: string): string {
  const parsed = parseCanonical(value);
  const negative = parsed.lt(0);
  const magnitude = negative ? parsed.neg() : parsed;
  const rounded = quantize(magnitude, AMOUNT_MANTISSA_DIGITS);
  const groupDigits = magnitude.lt(1_000) ? 1 : ((rounded.exponent % 3) + 3) % 3 + 1;
  const mantissaPlaces = AMOUNT_MANTISSA_DIGITS - groupDigits;
  const formatted = standard.format(rounded.toString(), mantissaPlaces, AMOUNT_UNDER_1000_PLACES).trim();
  return negative && formatted !== "0" ? `−${formatted}` : formatted;
}
