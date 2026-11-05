// Manual, unapplied proposal model. Does not export/replace product arithmetic.
import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { readFileSync, writeFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import Decimal from "../../../client/node_modules/break_infinity.js/dist/break_infinity.common.js";
import { canonicalString, classify, isStateValue } from "../../../client/src/numeric.ts";

const root = new URL("../../../", import.meta.url);
const read = (path) => readFileSync(new URL(path, root));
const digest = (bytes) => createHash("sha256").update(bytes).digest("hex");
const inputs = ["client/src/numeric.ts", "client/src/economy.ts", "client/src/commons.ts",
  "server/decimal/decimal.go", "kernel/VERSION", "testdata/decimal-vectors.json",
  "testdata/numeric-lower-domain-v1.json", "client/node_modules/break_infinity.js/dist/break_infinity.common.js"];
const hashes = Object.fromEntries(inputs.map((path) => [path, digest(read(path))]));
const packageVersion = JSON.parse(read("client/node_modules/break_infinity.js/package.json")).version;
assert.equal(packageVersion, "2.2.0");
const MAX = 8999999999999999;
const counters = { evaluations: 0, overrides: 0, shrinking: 0, growing: 0, unexpected: 0 };
const pair = (value) => [value.mantissa, value.exponent];
const equalPair = (a, b) => Object.is(a[0], b[0]) && Object.is(a[1], b[1]);
function finiteProfile(value) {
  return classify(value) === "finite" && Number.isFinite(value.mantissa) &&
    Number.isFinite(value.exponent) && Number.isInteger(value.exponent) &&
    (value.mantissa === 0 ? value.exponent === 0 : Math.abs(value.mantissa) >= 1 && Math.abs(value.mantissa) < 10);
}
function project(source, power, raw) {
  if (raw.mantissa !== 0 || raw.exponent !== 0 || !finiteProfile(source)) return raw;
  if (power === null) {
    if (source.mantissa < 0) { counters.shrinking++; return new Decimal(Number.NaN); }
    if (source.mantissa > 0) { counters.growing++; return new Decimal(Number.POSITIVE_INFINITY); }
  } else {
    if (source.mantissa === 0 || !Number.isFinite(power) ||
      source.mantissa < 0 && !Number.isInteger(power)) return raw;
    const magnitude = source.abs().cmp(1);
    if (magnitude < 0 && power > 0 || magnitude > 0 && power < 0) {
      counters.shrinking++; return new Decimal(Number.NaN);
    }
    if (magnitude > 0 && power > 0 || magnitude < 0 && power < 0) {
      counters.growing++;
      return new Decimal(source.mantissa < 0 && Math.abs(power % 2) === 1 ? Number.NEGATIVE_INFINITY : Number.POSITIVE_INFINITY);
    }
  }
  counters.unexpected++;
  throw new Error("unexpected provider zero at unit magnitude, power0 or exp0");
}
function evaluate(source, power) {
  counters.evaluations++;
  const before = pair(source);
  const clone = new Decimal(source);
  const raw = power === null ? clone.exp() : clone.pow(power);
  const rawBefore = pair(raw);
  const projected = project(source, power, raw);
  if (projected !== raw) counters.overrides++;
  assert.ok(equalPair(before, pair(source)), "source mutated");
  assert.ok(equalPair(rawBefore, pair(raw)), "provider result mutated");
  return { raw, projected };
}
function observe(value) {
  let wire = null;
  try { wire = canonicalString(value); } catch (error) { assert.ok(error instanceof RangeError); }
  return { class: classify(value), state: isStateValue(value), wire,
    mantissa: String(value.mantissa), exponent: Number.isFinite(value.exponent) && Number.isInteger(value.exponent)
      ? BigInt(value.exponent).toString() : String(value.exponent) };
}
function next(value, up) {
  const view = new DataView(new ArrayBuffer(8));
  view.setFloat64(0, value);
  view.setBigUint64(0, view.getBigUint64(0) + ((value > 0) === up ? 1n : -1n));
  return view.getFloat64(0);
}

const corpus = JSON.parse(read("testdata/decimal-vectors.json"));
assert.equal(corpus.vectors.length, 6296);
const legacy = corpus.vectors.filter((row) => row.op === "pow" || row.op === "exp");
assert.equal(legacy.filter((row) => row.op === "pow").length, 600);
assert.equal(legacy.filter((row) => row.op === "exp").length, 353);
const lower = JSON.parse(read("testdata/numeric-lower-domain-v1.json"));
assert.equal(lower.cases.length, 24);
const lowerRows = lower.cases.filter((row) => row.op === "pow" || row.op === "exp");
assert.equal(lowerRows.filter((row) => row.op === "pow").length, 11);
assert.equal(lowerRows.filter((row) => row.op === "exp").length, 2);
for (const row of [...legacy, ...lowerRows]) {
  const { raw, projected } = evaluate(new Decimal(row.a), row.op === "exp" ? null : Number(row.b));
  assert.ok(equalPair(pair(raw), pair(projected)), `retained result changed: ${row.op}/${row.a}/${row.b}`);
  assert.equal(classify(projected), classify(raw));
}
assert.equal(counters.overrides, 0, "retained corpus reached a new override");

const edge = Number(-(2n ** 63n)) / -MAX;
// name, op, source, power, raw accepted zero?, projected class, projected state, optional wire
const cases = [
  ["lower-pow2048", "pow", `1e-${MAX}`, 2048, false, "finite", false],
  ["lower-exp-1e20", "exp", "-1e20", null, false, "finite", false],
  ["lower-exp-1e308", "exp", "-1e308", null, false, "finite", false],
  ["lower-exp-1e309", "exp", "-1e309", null, true, "nan", false],
  ["lower-finite-power-overflow", "pow", `1e-${MAX}`, 1e308, true, "nan", false],
  ["int64-neighbor-low", "pow", `1e-${MAX}`, next(edge, false), false, "finite", false],
  ["int64-edge", "pow", `1e-${MAX}`, edge, false, "finite", false],
  ["int64-neighbor-high", "pow", `1e-${MAX}`, next(edge, true), false, "finite", false],
  ["lower-negative-odd", "pow", `-1e-${MAX}`, 2049, false, "finite", false],
  ["lower-fractional", "pow", `1e-${MAX}`, 2048.5, false, "finite", false],
  ["upper-finite-power-overflow", "pow", `1e${MAX}`, 1e308, true, "positive-infinity", false],
  ["upper-finite-exp-overflow", "exp", "1e309", null, true, "positive-infinity", false],
  ["upper-negative-base-even-overflow", "pow", `-1e${MAX}`, 1e308, true, "positive-infinity", false],
  ["upper-negative-base-odd2049", "pow", `-1e${MAX}`, 2049, false, "negative-infinity", false],
  ["upper-negative-base-even2048", "pow", `-1e${MAX}`, 2048, false, "positive-infinity", false],
  ["upper-negative-base-safe-odd", "pow", `-1e${MAX}`, Number.MAX_SAFE_INTEGER, false, "negative-infinity", false],
  ["small-base-negative-power-overflow", "pow", `1e-${MAX}`, -1e308, true, "positive-infinity", false],
  ["negative-small-base-negative-overflow", "pow", `-1e-${MAX}`, -1e308, true, "positive-infinity", false],
  ["zero-power-zero", "pow", "0", 0, false, "finite", true, "1e0"],
  ["zero-power-positive", "pow", "0", 2, false, "finite", true, "0"],
  ["zero-power-negative-compatibility", "pow", "0", -1, false, "finite", true, "0"],
  ["coefficient-normalization-rescue", "pow", Decimal.fromMantissaExponent(9.999999999999998, -MAX + 1), 1.0000000000000002, false, "finite", true, `1e-${MAX}`],
  ["tiny-power-recovery", "pow", `1e-${MAX}`, 1e-308, false, "finite", true, "1e0"],
  ["exp-zero", "exp", "0", null, false, "finite", true, "1e0"],
];
assert.equal(cases.length, 24);
const observations = cases.map(([name, op, input, power, rawZero, projectedClass, projectedState, wire]) => {
  const source = new Decimal(input);
  assert.ok(isStateValue(source), `${name}: input not admitted`);
  assert.ok(op === "exp" || Number.isFinite(power), `${name}: power not finite`);
  const { raw, projected } = evaluate(source, power);
  const before = observe(raw), after = observe(projected);
  assert.equal(after.class, projectedClass, name);
  assert.equal(after.state, projectedState, name);
  if (rawZero) {
    assert.equal(before.state, true, `${name}: baseline failure not reproduced`);
    assert.equal(before.wire, "0", `${name}: baseline not valid zero`);
    assert.equal(after.wire, null, `${name}: projected refusal not observed`);
  } else {
    assert.ok(equalPair(pair(raw), pair(projected)), `${name}: unaffected control changed`);
  }
  if (wire !== undefined) assert.equal(after.wire, wire, name);
  return { name, op, power: power === null ? null : String(power), source: pair(source).map(String), raw: before, projected: after };
});
assert.deepEqual(counters, { evaluations: 990, overrides: 7, shrinking: 2, growing: 5, unexpected: 0 });
for (const path of inputs) assert.equal(digest(read(path)), hashes[path], `bound source changed: ${path}`);
const result = { version: 1, projection_only: true, runtime_implemented: false,
  provider_version: packageVersion, kernel_version: String(read("kernel/VERSION")).trim(),
  source_sha256: hashes, model_sha256: digest(readFileSync(fileURLToPath(import.meta.url))),
  populations: { retained_golden: 953, retained_lower: 13, targeted: 24 }, counters,
  excluded_cases: 0, truncated: false, guard_exhausted: false, objective_reached: true, observations };
const output = JSON.stringify(result, null, 2) + "\n";
if (process.argv.length > 3) throw new Error("usage: checked-power-policy.mjs [new-output-path]");
if (process.argv[2]) writeFileSync(process.argv[2], output, { flag: "wx" });
else process.stdout.write(output);
if (process.argv[2]) process.stdout.write(`PASS proposal projection: 966 retained results unchanged; 24 targeted; 7 raw valid zeros refused; no runtime implementation.\n`);
