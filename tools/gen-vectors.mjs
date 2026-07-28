import { writeFile } from "node:fs/promises";
import { createRequire } from "node:module";
import { fileURLToPath } from "node:url";

const requireFromClient = createRequire(new URL("../client/package.json", import.meta.url));
const Decimal = requireFromClient("break_infinity.js");

const outputPath = fileURLToPath(new URL("../testdata/decimal-vectors.json", import.meta.url));
const seed = 0x00c10dc1;
const significantDigits = 12;
const maxExponent = 8_999_999_999_999_999;
let randomState = seed;

function random() {
  randomState = (randomState + 0x6d2b79f5) | 0;
  let value = randomState;
  value = Math.imul(value ^ (value >>> 15), value | 1);
  value ^= value + Math.imul(value ^ (value >>> 7), value | 61);
  return ((value ^ (value >>> 14)) >>> 0) / 4294967296;
}

function randomInt(min, max) {
  return Math.floor(random() * (max - min + 1)) + min;
}

function roundHalfEven(value) {
  const floor = Math.floor(value);
  const fraction = value - floor;
  if (fraction < 0.5) return floor;
  if (fraction > 0.5) return floor + 1;
  return floor % 2 === 0 ? floor : floor + 1;
}

function isStateValue(value) {
  if (
    !Number.isFinite(value.mantissa) ||
    !Number.isSafeInteger(value.exponent) ||
    Math.abs(value.exponent) > maxExponent
  ) return false;
  if (value.mantissa === 0) return value.exponent === 0;
  return Math.abs(value.mantissa) >= 1 && Math.abs(value.mantissa) < 10;
}

function quantize(value) {
  value = new Decimal(value).normalize();
  if (!isStateValue(value) || value.eq(0)) return value;
  const factor = 10 ** (significantDigits - 1);
  let coefficient = roundHalfEven(Math.abs(value.mantissa) * factor) / factor;
  let exponent = value.exponent;
  if (coefficient >= 10) {
    coefficient = 1;
    exponent += 1;
  }
  if (Math.abs(exponent) > maxExponent) return new Decimal(Number.NaN);
  return Decimal.fromMantissaExponent(Math.sign(value.mantissa) * coefficient, exponent);
}

function canonicalString(value) {
  value = quantize(value);
  if (!isStateValue(value)) throw new RangeError("non-finite canonical value");
  if (value.eq(0)) return "0";
  const coefficient = value.mantissa
    .toFixed(significantDigits - 1)
    .replace(/(?:\.0+|(?<=[0-9])0+)$/, "")
    .replace(/\.$/, "");
  return `${coefficient}e${value.exponent}`;
}

function classify(value) {
  if (Number.isNaN(value.mantissa) || Number.isNaN(value.exponent)) return "nan";
  if (value.mantissa !== 0 && value.exponent >= 9e15) {
    return value.mantissa < 0 ? "negative-infinity" : "positive-infinity";
  }
  if (!Number.isFinite(value.mantissa) || !Number.isFinite(value.exponent)) {
    return value.mantissa < 0 ? "negative-infinity" : "positive-infinity";
  }
  return "finite";
}

const boundaryExponents = [
  -8_999_999_999_999_000,
  -4_500_000_000_000_000,
  -1_000_000,
  -301,
  -300,
  -1,
  0,
  1,
  300,
  301,
  1_000_000,
  4_500_000_000_000_000,
  8_999_999_999_999_000,
];

const diagnosticInputs = ["0", "1e0", "-1e0", "Infinity", "-Infinity", "NaN"];

function randomExponent(limit = maxExponent) {
  const roll = random();
  if (roll < 0.2) return randomInt(-Math.min(limit, 300), Math.min(limit, 300));
  if (roll < 0.4) return randomInt(-Math.min(limit, 1_000_000), Math.min(limit, 1_000_000));
  if (roll < 0.5) {
    const candidate = boundaryExponents[randomInt(0, boundaryExponents.length - 1)];
    return Math.max(-limit, Math.min(limit, candidate));
  }
  const sign = random() < 0.5 ? -1 : 1;
  return sign * Math.floor(random() * (limit + 1));
}

function randomDecimal({ positive = false, exponentLimit = maxExponent } = {}) {
  const sign = positive || random() >= 0.45 ? "" : "-";
  const exponent = randomExponent(exponentLimit);
  const mantissa = Number((1 + random() * 8.999999999999).toPrecision(12));
  return canonicalString(Decimal.fromMantissaExponent(Number(`${sign}${mantissa}`), exponent));
}

function supported(result) {
  return classify(result) !== "finite" || isStateValue(result);
}

function approximateVector(op, a, b, result, extra = {}) {
  return {
    assert: "approx",
    a,
    b,
    op,
    expect: result.toString(),
    expectClass: classify(result),
    ...extra,
  };
}

function pushBinary(vectors, op, count) {
  while (count > 0) {
    const a = randomDecimal();
    const b = randomDecimal();
    const result = new Decimal(a)[op](new Decimal(b));
    if (!supported(result)) continue;
    vectors.push(approximateVector(op, a, b, result));
    count--;
  }
}

function pushUnary(vectors, op, count, inputFactory = () => randomDecimal()) {
  while (count > 0) {
    const a = inputFactory();
    const rawResult = new Decimal(a)[op]();
    const result = rawResult instanceof Decimal ? rawResult : new Decimal(rawResult);
    if (!supported(result)) continue;
    vectors.push(approximateVector(op, a, "", result));
    count--;
  }
}

function sumGeometric(count, base, ratio, owned) {
  if (count === 0) return new Decimal(0);
  const start = new Decimal(base).mul(new Decimal(ratio).pow(owned));
  return start.mul(Decimal.sub(1, new Decimal(ratio).pow(count))).div(Decimal.sub(1, ratio));
}

function verifiedAffordable(cash, base, ratio, owned, ceiling = 1_000) {
  cash = new Decimal(cash);
  let count = 0;
  while (count < ceiling && sumGeometric(count + 1, base, ratio, owned).lte(cash)) count++;
  return count;
}

const vectors = [];

function evaluateEdge(op, a, b) {
  const value = new Decimal(a);
  switch (op) {
    case "add":
    case "mul":
    case "div":
      return value[op](new Decimal(b));
    case "pow":
      return value.pow(Number(b));
    case "log10":
    case "ln":
      return new Decimal(value[op]());
    case "exp":
      return value.exp();
    case "quantize":
      return quantize(a);
    default:
      throw new Error(`unknown edge operation: ${op}`);
  }
}

for (const [edge, op, a, b = ""] of [
  ["pow-zero-zero", "pow", "0", "0"],
  ["pow-zero-positive", "pow", "0", "2"],
  ["pow-zero-negative", "pow", "0", "-1"],
  ["pow-negative-integer", "pow", "-2e0", "3"],
  ["pow-negative-fractional", "pow", "-2e0", "0.5"],
  ["div-zero", "div", "1e0", "0"],
  ["zero-div-zero", "div", "0", "0"],
  ["log10-zero", "log10", "0"],
  ["log10-negative", "log10", "-1e0"],
  ["ln-zero", "ln", "0"],
  ["ln-negative", "ln", "-1e0"],
  ["exp-float-underflow", "exp", "-1000e0"],
  ["exp-finite", "exp", "1000e0"],
  ["exp-infinity", "exp", "2.1e16"],
  ["positive-infinity-input", "add", "Infinity", "1e0"],
  ["negative-infinity-input", "mul", "-Infinity", "2e0"],
  ["infinity-cancellation", "add", "Infinity", "-Infinity"],
  ["infinity-times-zero", "mul", "Infinity", "0"],
  ["quantize-max-carry", "quantize", "9.999999999995e8999999999999999"],
  ["quantize-min-carry", "quantize", "9.999999999995e-8999999999999999"],
  ["div-reciprocal-boundary", "div", "1e8999999999999999", "2.5e8999999999999999"],
  ["mul-lower-bound-carry", "mul", "9e-8999999999999999", "2e-1"],
]) {
  vectors.push(approximateVector(op, a, b, evaluateEdge(op, a, b), { edge }));
}

for (const [input, expect] of [
  ["0", "0"],
  ["1", "1e0"],
  ["-1", "-1e0"],
  ["123456789012345678901", "1.23456789012e20"],
  ["1.234567890125e0", "1.23456789012e0"],
  ["1.234567890135e0", "1.23456789013e0"],
  ["-1.234567890125e0", "-1.23456789012e0"],
  ["-1.234567890135e0", "-1.23456789013e0"],
  ["9.999999999995e0", "1e1"],
  ["1e-8999999999999999", "1e-8999999999999999"],
  ["9e8999999999999999", "9e8999999999999999"],
]) {
  vectors.push({ assert: "exact", a: input, b: "", op: "canonical", expect });
}

for (const input of [
  "0",
  "1e0",
  "-4.25e-7",
  "9.87654321012e123456",
  "1.0e0",
  "1e+1",
  "01e0",
  "1E0",
  "NaN",
  "Infinity",
  "1e9000000000000001",
]) {
  const valid = input === "0" || /^-?[1-9](?:\.\d{0,10}[1-9])?e(?:0|-?[1-9]\d*)$/.test(input) &&
    Math.abs(Number(input.slice(input.indexOf("e") + 1))) <= maxExponent;
  vectors.push({ assert: "exact", a: input, b: "", op: "parse-valid", expect: String(valid) });
}

for (const input of diagnosticInputs) {
  const value = new Decimal(input);
  vectors.push({ assert: "exact", a: input, b: "", op: "state-valid", expect: String(isStateValue(value)) });
}

for (const [a, b] of [
  ["1.000000000001e0", "1e0"],
  ["-1e1000", "1e1000"],
  ["1e-1000", "0"],
]) {
  vectors.push({ assert: "exact", a, b, op: "cmp", expect: String(new Decimal(a).cmp(new Decimal(b))) });
}

for (const op of ["add", "sub", "mul", "div"]) pushBinary(vectors, op, 800);

while (vectors.filter((vector) => vector.op === "pow").length < 600) {
  const a = randomDecimal();
  const powers = [-3, -2, -1, -0.5, 0, 0.5, 1, 1.5, 2, 3];
  const b = String(powers[randomInt(0, powers.length - 1)]);
  const result = new Decimal(a).pow(Number(b));
  if (!supported(result)) continue;
  vectors.push(approximateVector("pow", a, b, result));
}

pushUnary(vectors, "log10", 400, () =>
  randomDecimal({ positive: true, exponentLimit: 8_999_999_999_999_000 }),
);
pushUnary(vectors, "ln", 400, () =>
  randomDecimal({ positive: true, exponentLimit: 3_000_000_000_000_000 }),
);
pushUnary(vectors, "floor", 250);

const expInputs = ["-1000e0", "-1e0", "0", "1e0", "7.097e2", "7.1e2", "1e3", "8.999999999999e15"];
pushUnary(vectors, "exp", 350, () => {
  if (random() < 0.15) return expInputs[randomInt(0, expInputs.length - 1)];
  return canonicalString(new Decimal(Math.trunc((random() * 2 - 1) * 8_000_000_000_000_000)));
});

while (vectors.filter((vector) => vector.op === "cmp").length < 250) {
  const a = randomDecimal();
  const b = randomDecimal();
  vectors.push({ assert: "exact", a, b, op: "cmp", expect: String(new Decimal(a).cmp(new Decimal(b))) });
}

for (let index = 0; index < 200; index++) {
  const a = randomDecimal({ exponentLimit: 1_000_000 });
  const b = randomDecimal({ exponentLimit: 1_000_000 });
  const op = index % 2 === 0 ? "commit-add" : "commit-mul";
  const result = op === "commit-add" ? new Decimal(a).add(b) : new Decimal(a).mul(b);
  if (!isStateValue(result)) {
    index--;
    continue;
  }
  vectors.push({ assert: "exact", a, b, op, expect: canonicalString(result) });
}

for (const ratioNumber of [1.07, 1.13, 1.15]) {
  const ratio = canonicalString(new Decimal(ratioNumber));
  for (let index = 0; index < 100; index++) {
    const base = randomDecimal({ positive: true, exponentLimit: 1_000_000 });
    const owned = randomInt(0, 500);
    const count = randomInt(0, 500);
    const sum = sumGeometric(count, base, ratio, owned);
    vectors.push(approximateVector("sum", String(count), base, sum, { ratio, owned: String(owned) }));

    const target = randomInt(0, 500);
    const targetCost = sumGeometric(target, base, ratio, owned);
    const nextPrice = new Decimal(base).mul(new Decimal(ratio).pow(owned + target));
    const cash = canonicalString(targetCost.add(nextPrice.mul(random() * 0.9)));
    const expect = verifiedAffordable(cash, base, ratio, owned);
    vectors.push({
      assert: "decision",
      a: cash,
      b: base,
      op: "afford",
      ratio,
      owned: String(owned),
      expect: String(expect),
    });
  }
}

function countsBy(values) {
  const counts = {};
  for (const value of values) counts[value] = (counts[value] ?? 0) + 1;
  return Object.fromEntries(Object.entries(counts).sort(([a], [b]) => a.localeCompare(b)));
}

const coverage = {
  operations: countsBy(vectors.map((vector) => vector.op)),
  classifications: countsBy(vectors.filter((vector) => vector.expectClass).map((vector) => vector.expectClass)),
  edges: vectors.filter((vector) => vector.edge).map((vector) => vector.edge).sort(),
};

await writeFile(outputPath, `${JSON.stringify({ version: 3, seed, coverage, vectors }, null, 2)}\n`, "utf8");
console.log(`wrote ${vectors.length} categorized vectors to ${outputPath}`);
console.log(JSON.stringify(coverage));
