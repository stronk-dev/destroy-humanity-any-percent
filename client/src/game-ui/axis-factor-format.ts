import { parseCanonical } from "../numeric";

// CV9 displays the supplied factor, not a rounded resource amount. Place the
// decimal point within the canonical coefficient using text only: Decimal's
// toString() can expose binary artifacts (1.005e1 -> 10.049999999999999).
// If expansion would need extra zeros, retain the exact canonical notation.
export function formatAxisFactor(value: string): string {
  const parsed = parseCanonical(value);
  if (value === "0") return value;
  const coefficient = value.slice(0, value.indexOf("e"));
  const negative = coefficient.startsWith("-");
  const digits = (negative ? coefficient.slice(1) : coefficient).replace(".", "");
  const point = parsed.exponent + 1;
  if (point <= 0 || point > digits.length) return value;
  const fraction = digits.slice(point);
  return `${negative ? "-" : ""}${digits.slice(0, point)}${fraction ? `.${fraction}` : ""}`;
}
