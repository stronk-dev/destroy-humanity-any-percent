import Decimal from "break_infinity.js";

export const CANONICAL_SIGNIFICANT_DIGITS = 12;
export const MAX_EXACT_INTEGER = Number.MAX_SAFE_INTEGER;
// break_infinity uses ±9e15 as sentinel/normalization boundaries. Keeping the
// gameplay exponent strictly inside them prevents its "Infinity" parser token
// from becoming indistinguishable from valid authoritative state.
export const MAX_EXPONENT = 8_999_999_999_999_999;

const canonicalPattern = /^(?:0|-?[1-9](?:\.\d{0,10}[1-9])?e(?:0|-?[1-9]\d*))$/;

export type DecimalSource = Decimal | string | number;

function roundHalfEven(value: number): number {
  const floor = Math.floor(value);
  const fraction = value - floor;
  if (fraction < 0.5) return floor;
  if (fraction > 0.5) return floor + 1;
  return floor % 2 === 0 ? floor : floor + 1;
}

export function isStateValue(value: Decimal): boolean {
  return (
    Number.isFinite(value.mantissa) &&
    Number.isSafeInteger(value.exponent) &&
    Math.abs(value.exponent) <= MAX_EXPONENT
  );
}

export function quantize(
  source: DecimalSource,
  significantDigits = CANONICAL_SIGNIFICANT_DIGITS,
): Decimal {
  const value = new Decimal(source);
  if (!isStateValue(value) || value.eq(0)) return value;
  if (!Number.isInteger(significantDigits) || significantDigits < 1 || significantDigits > 15) {
    return new Decimal(Number.NaN);
  }

  const factor = 10 ** (significantDigits - 1);
  const direct =
    typeof source === "string"
      ? source.match(/^(-?[0-9]+(?:\.[0-9]+)?)e(-?[0-9]+)$/)
      : null;
  const inputCoefficient = direct ? Number(direct[1]) : value.mantissa;
  let coefficient = roundHalfEven(Math.abs(inputCoefficient) * factor) / factor;
  let exponent = direct ? Number(direct[2]) : value.exponent;
  if (coefficient >= 10) {
    coefficient = 1;
    exponent += 1;
  }
  if (Math.abs(exponent) > MAX_EXPONENT) return new Decimal(Number.NaN);
  return Decimal.fromMantissaExponent(Math.sign(value.mantissa) * coefficient, exponent);
}

export function canonicalString(source: DecimalSource): string {
  const value = quantize(source);
  if (!isStateValue(value)) {
    throw new RangeError("non-finite or out-of-range Decimal cannot enter gameplay state");
  }
  if (value.eq(0)) return "0";

  const coefficient = value.mantissa
    .toFixed(CANONICAL_SIGNIFICANT_DIGITS - 1)
    .replace(/(?:\.0+|(?<=[0-9])0+)$/, "")
    .replace(/\.$/, "");
  return `${coefficient}e${value.exponent}`;
}

export function parseCanonical(source: string): Decimal {
  if (!canonicalPattern.test(source)) {
    throw new SyntaxError(`invalid canonical Decimal: ${source}`);
  }
  if (source === "0") return new Decimal(0);

  const exponent = Number(source.slice(source.indexOf("e") + 1));
  if (!Number.isSafeInteger(exponent) || Math.abs(exponent) > MAX_EXPONENT) {
    throw new SyntaxError(`canonical Decimal exponent out of range: ${source}`);
  }
  const value = new Decimal(source);
  if (!isStateValue(value) || canonicalString(value) !== source) {
    throw new SyntaxError(`non-canonical Decimal: ${source}`);
  }
  return value;
}

export function classify(value: Decimal): "finite" | "nan" | "positive-infinity" | "negative-infinity" {
  if (Number.isNaN(value.mantissa) || Number.isNaN(value.exponent)) {
    return "nan";
  }
  if (!Number.isFinite(value.mantissa) || !Number.isFinite(value.exponent)) {
    return value.mantissa < 0 ? "negative-infinity" : "positive-infinity";
  }
  return "finite";
}
