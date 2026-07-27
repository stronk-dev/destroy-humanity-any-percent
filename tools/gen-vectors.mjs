import { writeFile } from "node:fs/promises";
import { createRequire } from "node:module";
import { fileURLToPath } from "node:url";

const requireFromClient = createRequire(
  new URL("../client/package.json", import.meta.url),
);
const Decimal = requireFromClient("break_eternity.js");

const outputPath = fileURLToPath(
  new URL("../testdata/decimal-vectors.json", import.meta.url),
);
const seed = 0x00c10dc1;
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

const specialInputs = [
  "0",
  "1",
  "-1",
  "1e-7",
  "1.000000000000001e-7",
  "1e-300",
  "1e300",
  "1e301",
  "1e1000000",
  "1e8999999999999000",
  "1e-8999999999999000",
  "Infinity",
  "-Infinity",
  "NaN",
];

function randomExponent(limit = 4_000_000_000_000_000) {
  const roll = random();
  if (roll < 0.2) return randomInt(-300, 300);
  if (roll < 0.4) return randomInt(-1_000_000, 1_000_000);
  if (roll < 0.5) {
    const candidate = boundaryExponents[randomInt(0, boundaryExponents.length - 1)];
    return Math.max(-limit, Math.min(limit, candidate));
  }
  return Math.trunc((random() * 2 - 1) * limit);
}

function randomDecimal({ positive = false, exponentLimit = 4_000_000_000_000_000 } = {}) {
  if (!positive && random() < 0.025) {
    return specialInputs[randomInt(0, specialInputs.length - 1)];
  }
  const sign = positive || random() >= 0.45 ? "" : "-";
  const exponent = randomExponent(exponentLimit);
  // Arbitrary mantissas are exercised where both runtimes use direct IEEE-754
  // arithmetic. At logarithmic magnitudes, powers of ten avoid treating
  // platform-libm last-bit differences as Decimal algorithm differences.
  const mantissa =
    Math.abs(exponent) <= 15
      ? Number((1 + random() * 8.999999999999).toPrecision(15))
      : 1;
  return new Decimal(`${sign}${mantissa}e${exponent}`).toString();
}

function supported(result) {
  return !Number.isFinite(result.layer) || result.layer <= 1;
}

function pushBinary(vectors, op, count) {
  while (count > 0) {
    const a = randomDecimal();
    const b = randomDecimal();
    const left = new Decimal(a);
    const right = new Decimal(b);
    const result = left[op](right);
    if (!supported(result)) continue;
    vectors.push({ a, b, op, expect: result.toString() });
    count--;
  }
}

function pushUnary(vectors, op, count, inputFactory = () => randomDecimal()) {
  while (count > 0) {
    const a = inputFactory();
    const result = new Decimal(a)[op]();
    if (!supported(result)) continue;
    vectors.push({ a, b: "", op, expect: result.toString() });
    count--;
  }
}

const vectors = [];

for (const [a, b] of [
  ["1.000000000000001", "1"],
  ["100000000000001", "100000000000000"],
  ["-1.000000000000001", "-1"],
  ["1e1000", "1e1000"],
  ["1e-1000", "1e-1000"],
]) {
  vectors.push({ a, b, op: "sub", expect: new Decimal(a).sub(b).toString() });
}

for (const op of ["add", "sub", "mul", "div"]) {
  pushBinary(vectors, op, 800);
}

while (vectors.filter((vector) => vector.op === "pow").length < 600) {
  const a = randomDecimal({ exponentLimit: 2_000_000_000_000_000 });
  const powers = [-3, -2, -1, -0.5, 0, 0.5, 1, 1.5, 2, 3];
  const b = String(powers[randomInt(0, powers.length - 1)]);
  const result = new Decimal(a).pow(Number(b));
  if (!supported(result)) continue;
  vectors.push({ a, b, op: "pow", expect: result.toString() });
}

pushUnary(vectors, "log10", 400, () =>
  random() < 0.08
    ? specialInputs[randomInt(0, specialInputs.length - 1)]
    : randomDecimal({ positive: true, exponentLimit: 8_999_999_999_999_000 }),
);
pushUnary(vectors, "ln", 400, () =>
  random() < 0.08
    ? specialInputs[randomInt(0, specialInputs.length - 1)]
    : randomDecimal({ positive: true, exponentLimit: 3_000_000_000_000_000 }),
);
pushUnary(vectors, "floor", 250);

const expInputs = [
  "-1000",
  "-1",
  "0",
  "1",
  "709.7",
  "710",
  "1000",
  "8999999999999000",
  "Infinity",
  "-Infinity",
  "NaN",
];
pushUnary(vectors, "exp", 350, () => {
  if (random() < 0.15) return expInputs[randomInt(0, expInputs.length - 1)];
  return String(Math.trunc((random() * 2 - 1) * 8_000_000_000_000_000));
});

while (vectors.filter((vector) => vector.op === "cmp").length < 250) {
  const a = randomDecimal();
  const b = randomDecimal();
  vectors.push({
    a,
    b,
    op: "cmp",
    expect: String(new Decimal(a).cmp(new Decimal(b))),
  });
}

for (const ratio of [1.07, 1.13, 1.15]) {
  for (let index = 0; index < 100; index++) {
    const base = randomDecimal({ positive: true, exponentLimit: 1_000_000 });
    const owned = randomInt(0, 500);
    const count = randomInt(0, 500);
    const sum = Decimal.sumGeometricSeries(count, base, ratio, owned);
    vectors.push({
      a: String(count),
      b: base,
      op: "sum",
      ratio: String(ratio),
      owned: String(owned),
      expect: sum.toString(),
    });

    const target = randomInt(0, 500);
    const cash = Decimal.sumGeometricSeries(target, base, ratio, owned);
    vectors.push({
      a: cash.toString(),
      b: base,
      op: "afford",
      ratio: String(ratio),
      owned: String(owned),
      expect: Decimal.affordGeometricSeries(cash, base, ratio, owned).toString(),
    });
  }
}

await writeFile(
  outputPath,
  `${JSON.stringify({ seed, vectors }, null, 2)}\n`,
  "utf8",
);

console.log(`wrote ${vectors.length} vectors to ${outputPath}`);
