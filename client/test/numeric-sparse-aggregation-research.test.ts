import Decimal from "break_infinity.js";
import { expect, it } from "vitest";
import { canonicalString, isStateValue, MAX_EXPONENT } from "../src/numeric";
import fixture from "../../testdata/numeric-sum-order-research-v1.json";

// Unapplied exact-input prototype. No production import or accepted algorithm.
const BASE = 1_000_000_000n;
const SCALE = 1n << 52n;
type Digit = { index: bigint; digit: bigint };
type Piece = { first: bigint; last: bigint; digit: bigint };
type Work = { chunks: number; normalize: number; events: number; runs: number; reads: number };
const bits = (value: number) => {
  const data = new DataView(new ArrayBuffer(8));
  data.setFloat64(0, value);
  return data.getBigUint64(0);
};
const position = (exponent: bigint): [bigint, bigint] => {
  let index = exponent / 9n, remainder = exponent % 9n;
  if (remainder < 0n) { index--; remainder += 9n; }
  return [index, remainder];
};

function normalize(source: Map<bigint, bigint>, work: Work): Digit[] {
  const keys = [...source.keys()].sort((a, b) => a < b ? -1 : a > b ? 1 : 0);
  const result: Digit[] = [];
  let carry = 0n, index = 0n, next = 0;
  while (next < keys.length || carry !== 0n) {
    if (carry === 0n) index = keys[next]!;
    let value = carry;
    if (next < keys.length && keys[next] === index) { value += source.get(index)!; next++; }
    const digit = value % BASE;
    if (digit !== 0n) result.push({ index, digit });
    carry = value / BASE;
    work.normalize++;
    index++;
  }
  return result;
}

function compare(a: readonly Digit[], b: readonly Digit[]): number {
  for (let i = a.length - 1, j = b.length - 1; i >= 0 || j >= 0; i--, j--) {
    if (i < 0) return -1;
    if (j < 0) return 1;
    if (a[i]!.index !== b[j]!.index) return a[i]!.index < b[j]!.index ? -1 : 1;
    if (a[i]!.digit !== b[j]!.digit) return a[i]!.digit < b[j]!.digit ? -1 : 1;
  }
  return 0;
}

function subtract(a: readonly Digit[], b: readonly Digit[], work: Work): Piece[] {
  const pieces: Piece[] = [];
  let borrow = 0n, previous = 0n, i = 0, j = 0;
  while (i < a.length || j < b.length) {
    const index = j === b.length || i < a.length && a[i]!.index < b[j]!.index ? a[i]!.index : b[j]!.index;
    if (work.events > 0 && borrow !== 0n && index > previous + 1n) {
      pieces.push({ first: previous + 1n, last: index - 1n, digit: BASE - 1n });
      work.runs++;
    }
    let value = -borrow;
    if (i < a.length && a[i]!.index === index) value += a[i++]!.digit;
    if (j < b.length && b[j]!.index === index) value -= b[j++]!.digit;
    borrow = 0n;
    if (value < 0n) { value += BASE; borrow = 1n; }
    if (value !== 0n) pieces.push({ first: index, last: index, digit: value });
    previous = index;
    work.events++;
  }
  if (borrow !== 0n) throw new Error("invalid final sparse borrow");
  return pieces;
}

function chunk(pieces: readonly Piece[], index: bigint): bigint {
  let low = 0, high = pieces.length;
  while (low < high) {
    const mid = low + Math.floor((high - low) / 2);
    if (pieces[mid]!.last < index) low = mid + 1; else high = mid;
  }
  return low === pieces.length || pieces[low]!.first > index ? 0n : pieces[low]!.digit;
}

function floorAt(pieces: readonly Piece[], exponent: bigint, work: Work): [bigint, boolean] {
  const [index, remainder] = position(exponent), high = pieces[pieces.length - 1]!.last;
  // Representation-derived invariant, not an arbitrary time/work limit.
  if (high - index > 2n || high < index) throw new Error(`unproved rounding query width ${high - index}`);
  let value = 0n;
  for (let at = high; at >= index; at--) { value = value * BASE + chunk(pieces, at); work.reads++; }
  const scale = 10n ** remainder;
  return [value / scale, pieces[0]!.first < index || chunk(pieces, index) % scale !== 0n];
}

function sparseSum(values: readonly Decimal[]): { value: Decimal; work: Work } {
  const work: Work = { chunks: 0, normalize: 0, events: 0, runs: 0, reads: 0 };
  const invalid = () => ({ value: new Decimal(NaN), work });
  const positive = new Map<bigint, bigint>(), negative = new Map<bigint, bigint>();
  for (const value of values) {
    if (!isStateValue(value)) return invalid();
    if (value.mantissa === 0) continue;
    const raw = bits(value.mantissa), binaryExponent = Number((raw >> 52n) & 2047n) - 1023;
    if (binaryExponent < 0 || binaryExponent > 3) throw new Error("invalid normalized significand");
    let coefficient = ((raw & (SCALE - 1n)) | SCALE) << BigInt(binaryExponent);
    let [index, remainder] = position(BigInt(value.exponent));
    coefficient *= 10n ** remainder;
    const target = value.mantissa < 0 ? negative : positive;
    while (coefficient !== 0n) {
      const digit = coefficient % BASE;
      if (digit !== 0n) target.set(index, (target.get(index) ?? 0n) + digit);
      coefficient /= BASE;
      index++;
      work.chunks++;
    }
  }
  let p = normalize(positive, work), n = normalize(negative, work);
  const sign = compare(p, n);
  if (sign === 0) return { value: new Decimal(0), work };
  if (sign < 0) [p, n] = [n, p];
  const pieces = subtract(p, n, work), last = pieces[pieces.length - 1];
  if (!last) throw new Error("invalid sparse difference");
  let exponent = 9n * last.last + BigInt(last.digit.toString().length - 1) - 15n;
  let [integer] = floorAt(pieces, exponent, work);
  if (integer < SCALE) { exponent--; [integer] = floorAt(pieces, exponent, work); }
  if (exponent < -BigInt(MAX_EXPONENT) || exponent > BigInt(MAX_EXPONENT)) return invalid();
  let binaryExponent = 0;
  while (binaryExponent < 3 && integer >= SCALE << BigInt(binaryExponent + 1)) binaryExponent++;
  const [scaled, sticky] = floorAt(pieces, exponent - 1n, work);
  const divisor = 10n << BigInt(binaryExponent), remainder = scaled % divisor;
  let quotient = scaled / divisor;
  if (2n * remainder > divisor || 2n * remainder === divisor && (sticky || quotient % 2n !== 0n)) quotient++;
  if (quotient < SCALE || quotient > 1n << 53n) throw new Error("unproved significand width");
  const value = Decimal.fromMantissaExponent(sign * Number(quotient) * 2 ** (binaryExponent - 52), Number(exponent));
  return isStateValue(value) ? { value, work } : invalid();
}

it("retains full-range cancellation, compressed borrows, exact ties and domain refusals", () => {
  const u = 2 ** -52, max = MAX_EXPONENT;
  const rows: [string, Decimal[], Decimal][] = [
    ["empty", [], new Decimal(0)],
    ["tiny cancellation", [new Decimal("1e300"), new Decimal("-1e300"), new Decimal("1e-300")], new Decimal("1e-300")],
    ["full-domain cancellation", [new Decimal(`1e${max}`), new Decimal(`-5e${max - 1}`), new Decimal(`-5e${max - 1}`), new Decimal(`1e-${max}`)], new Decimal(`1e-${max}`)],
    ["full-domain borrow", [new Decimal(`1e${max}`), new Decimal(`-1e-${max}`)], new Decimal(`1e${max}`)],
    ["even tie", [new Decimal(1), Decimal.fromMantissaExponent(1 + 5 * u, -1), new Decimal("-1e-1")], new Decimal(1)],
    ["remote above tie", [new Decimal(1), Decimal.fromMantissaExponent(1 + 5 * u, -1), new Decimal("-1e-1"), new Decimal(`1e-${max}`)], new Decimal(1 + u)],
    ["remote below tie", [new Decimal(1), Decimal.fromMantissaExponent(1 + 5 * u, -1), new Decimal("-1e-1"), new Decimal(`-1e-${max}`)], new Decimal(1)],
    ["odd tie", [new Decimal(1 + u), Decimal.fromMantissaExponent(1 + 5 * u, -1), new Decimal("-1e-1")], new Decimal(1 + 2 * u)],
    ["odd below tie", [new Decimal(1 + u), Decimal.fromMantissaExponent(1 + 5 * u, -1), new Decimal("-1e-1"), new Decimal(`-1e-${max}`)], new Decimal(1 + u)],
    ["negative remote tie", [new Decimal(-1), Decimal.fromMantissaExponent(-1 - 5 * u, -1), new Decimal("1e-1"), new Decimal(`-1e-${max}`)], new Decimal(-1 - u)],
    ["recoverable overflow", [new Decimal(`9e${max}`), new Decimal(`9e${max}`), new Decimal(`-9e${max}`)], new Decimal(`9e${max}`)],
    ["net overflow", [new Decimal(`9e${max}`), new Decimal(`9e${max}`)], new Decimal(NaN)],
    ["net underflow", [Decimal.fromMantissaExponent(1 + u, -max), new Decimal(`-1e-${max}`)], new Decimal(NaN)],
    ["invalid", [new Decimal(NaN)], new Decimal(NaN)],
  ];
  for (const [name, values, want] of rows) {
    const { value } = sparseSum(values);
    if (Number.isNaN(want.mantissa)) expect(Number.isNaN(value.mantissa), name).toBe(true);
    else expect([value.mantissa, value.exponent], name).toEqual([want.mantissa, want.exponent]);
  }
});

// Independent bounded dense oracle: multiply an exact common integer by 5^52
// to obtain a finite decimal expansion, then let native Number parse the exact
// normalized text. It reuses no sparse reducer, borrow or binary-rounding code.
function denseOracle(values: readonly Decimal[]): Decimal {
  if (!values.length) return new Decimal(0);
  if (values.some((v) => !isStateValue(v) || Math.abs(v.exponent) > 32)) throw new Error("outside dense oracle population");
  const minimum = Math.min(...values.map((v) => v.exponent));
  let integer = 0n;
  for (const value of values) integer += BigInt(value.mantissa * 2 ** 52) * 10n ** BigInt(value.exponent - minimum);
  if (integer === 0n) return new Decimal(0);
  const sign = integer < 0n ? -1 : 1;
  const text = ((integer < 0n ? -integer : integer) * 5n ** 52n).toString();
  const mantissa = Number(text[0] + "." + text.slice(1));
  return Decimal.fromMantissaExponent(sign * mantissa, text.length - 1 + minimum - 52);
}

it("matches independent dense arithmetic on the declared signed population", async () => {
  let seed = 664n;
  const mask = (1n << 64n) - 1n;
  const next = () => { seed = (seed ^ (seed << 13n)) & mask; seed ^= seed >> 7n; seed = (seed ^ (seed << 17n)) & mask; return seed; };
  const witnesses: string[] = [];
  let maxWork: Work = { chunks: 0, normalize: 0, events: 0, runs: 0, reads: 0 };
  for (let profile = 0; profile < 2048; profile++) {
    const values = Array.from({ length: Number(next() % 17n) }, () => {
      let mantissa = Number(100_000_000_000_000n + next() % 900_000_000_000_000n) / 1e14;
      const exponent = Number(next() % 65n) - 32;
      if ((next() & 1n) !== 0n) mantissa = -mantissa;
      return Decimal.fromMantissaExponent(mantissa, exponent);
    });
    if (profile % 4 === 0 && values.length) values.push(values[0]!.neg());
    const before = values.map((v) => [v.mantissa, v.exponent]);
    const { value, work } = sparseSum(values), want = denseOracle(values);
    expect([value.mantissa, value.exponent], `profile${profile}`).toEqual([want.mantissa, want.exponent]);
    expect(values.map((v) => [v.mantissa, v.exponent])).toEqual(before);
    const reversed = sparseSum([...values].reverse()).value;
    expect([reversed.mantissa, reversed.exponent]).toEqual([value.mantissa, value.exponent]);
    let digits = 0;
    for (let n = values.length; n > 0; n = Math.floor(n / Number(BASE))) digits++;
    if (work.chunks > 3 * values.length || work.normalize > work.chunks * (1 + digits) || work.events > work.normalize || work.runs > work.events || work.reads > 9) throw new Error(`derived work bound violated ${profile}`);
    if (work.normalize > maxWork.normalize) maxWork = work;
    witnesses.push(`${profile}:${values.length}|${values.map((v) => `${bits(v.mantissa).toString(16).padStart(16, "0")}:${v.exponent};`).join("")}=>${bits(value.mantissa).toString(16).padStart(16, "0")}:${value.exponent}\n`);
  }
  const digest = [...new Uint8Array(await crypto.subtle.digest("SHA-256", new TextEncoder().encode(witnesses.join(""))))].map((b) => b.toString(16).padStart(2, "0")).join("");
  console.info("sparse-aggregation-research-v1", JSON.stringify({ profiles: 2048, seed: 664, digest, maxWork }));
  expect(digest).toBe("ef6cdc7626f91ff237961488f84f07a022b5ac746542d53dfb68fb258768047f");
});

it("matches the rational diagnosis of both original positive-rate inputs", () => {
  for (const source of [fixture.variable.lower_rate, fixture.variable.upper_rate]) {
    const values = [new Decimal(source), ...fixture.fixed.map((row) => new Decimal(row.rate))];
    const value = sparseSum(values).value, want = denseOracle(values);
    expect([value.mantissa, value.exponent]).toEqual([want.mantissa, want.exponent]);
    expect(canonicalString(value)).toBe("1.01245957117e1");
  }
});
