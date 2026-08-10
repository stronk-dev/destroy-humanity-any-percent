import { StandardNotation } from "@antimatter-dimensions/notations";

import { parseCanonical } from "../numeric";

export const AMOUNT_MANTISSA_PLACES = 2;
export const AMOUNT_UNDER_1000_PLACES = 0;

const standard = new StandardNotation();

export function formatAmount(value: string): string {
  const parsed = parseCanonical(value);
  const negative = parsed.lt(0);
  const magnitude = negative ? parsed.neg() : parsed;
  const formatted = standard.format(magnitude.toString(), AMOUNT_MANTISSA_PLACES, AMOUNT_UNDER_1000_PLACES).trim();
  return negative ? `−${formatted}` : formatted;
}
