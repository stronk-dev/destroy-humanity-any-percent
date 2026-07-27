import Decimal from "break_infinity.js";
import { describe, expect, it } from "vitest";

import fixtureJson from "../../testdata/decimal-vectors.json";
import { affordGeometricSeries, sumGeometricSeries } from "../src/economy";
import { canonicalString, classify, isStateValue, parseCanonical, quantize } from "../src/numeric";

interface Vector {
  assert: "exact" | "approx" | "decision";
  a: string;
  b: string;
  op: string;
  ratio?: string;
  owned?: string;
  expect: string;
  expectClass?: "finite" | "nan" | "positive-infinity" | "negative-infinity";
  edge?: string;
}

interface Coverage {
  operations: Record<string, number>;
  classifications: Record<string, number>;
  edges: string[];
}

const fixture = fixtureJson as { version: number; seed: number; coverage: Coverage; vectors: Vector[] };

const requiredEdges = [
  "div-zero",
  "exp-finite",
  "exp-float-underflow",
  "exp-infinity",
  "infinity-cancellation",
  "infinity-times-zero",
  "ln-negative",
  "ln-zero",
  "log10-negative",
  "log10-zero",
  "negative-infinity-input",
  "positive-infinity-input",
  "pow-negative-fractional",
  "pow-negative-integer",
  "pow-zero-negative",
  "pow-zero-positive",
  "pow-zero-zero",
  "quantize-max-carry",
  "quantize-min-carry",
  "zero-div-zero",
];

function countsBy(values: string[]): Record<string, number> {
  const counts: Record<string, number> = {};
  for (const value of values) counts[value] = (counts[value] ?? 0) + 1;
  return counts;
}

function evaluateDecimal(vector: Vector): Decimal {
  const a = new Decimal(vector.a);
  const b = vector.b === "" ? null : new Decimal(vector.b);
  switch (vector.op) {
    case "add":
    case "sub":
    case "mul":
    case "div":
      return a[vector.op](b!);
    case "commit-add":
      return a.add(b!);
    case "commit-mul":
      return a.mul(b!);
    case "pow":
      return a.pow(Number(vector.b));
    case "log10":
    case "ln":
      return new Decimal(a[vector.op]());
    case "exp":
    case "floor":
      return a[vector.op]();
    case "quantize":
      return quantize(vector.a);
    case "sum":
      return sumGeometricSeries(Number(vector.a), vector.b, vector.ratio!, Number(vector.owned));
    default:
      throw new Error(`operation does not return Decimal: ${vector.op}`);
  }
}

function evaluateExact(vector: Vector): string {
  switch (vector.op) {
    case "canonical":
      return canonicalString(vector.a);
    case "parse-valid":
      try {
        parseCanonical(vector.a);
        return "true";
      } catch {
        return "false";
      }
    case "state-valid":
      return String(isStateValue(new Decimal(vector.a)));
    case "cmp":
      return String(new Decimal(vector.a).cmp(new Decimal(vector.b)));
    case "commit-add":
    case "commit-mul":
      return canonicalString(evaluateDecimal(vector));
    default:
      throw new Error(`unknown exact operation: ${vector.op}`);
  }
}

function expectApprox(vector: Vector): void {
  const got = evaluateDecimal(vector);
  expect(classify(got)).toBe(vector.expectClass);
  if (vector.expectClass !== "finite") return;

  const wanted = new Decimal(vector.expect);
  if (wanted.eq(0)) {
    expect(got.eq(0)).toBe(true);
    return;
  }
  const difference = got.sub(wanted).abs();
  const scale = Decimal.max(got.abs(), wanted.abs());
  expect(difference.div(scale).lte(1e-12)).toBe(true);
}

function expectDecision(vector: Vector): void {
  const owned = Number(vector.owned);
  const got = affordGeometricSeries(vector.a, vector.b, vector.ratio!, owned);
  expect(got).toBe(Number(vector.expect));

  const cash = new Decimal(vector.a);
  expect(sumGeometricSeries(got, vector.b, vector.ratio!, owned).lte(cash)).toBe(true);
  if (got < Number.MAX_SAFE_INTEGER) {
    expect(sumGeometricSeries(got + 1, vector.b, vector.ratio!, owned).gt(cash)).toBe(true);
  }
}

describe("shared decimal golden vectors", () => {
  it("contains the RFC minimum categorized coverage", () => {
    expect(fixture.version).toBe(3);
    expect(fixture.vectors.length).toBeGreaterThanOrEqual(5_000);
    expect(new Set(fixture.vectors.map((vector) => vector.assert))).toEqual(
      new Set(["exact", "approx", "decision"]),
    );
    expect(fixture.coverage.edges).toEqual(requiredEdges);
    expect(fixture.coverage.edges).toEqual(
      fixture.vectors.filter((vector) => vector.edge).map((vector) => vector.edge!).sort(),
    );
    expect(fixture.coverage.operations).toEqual(
      countsBy(fixture.vectors.map((vector) => vector.op)),
    );
    expect(fixture.coverage.classifications).toEqual(
      countsBy(
        fixture.vectors
          .filter((vector) => vector.expectClass)
          .map((vector) => vector.expectClass!),
      ),
    );
    for (const classification of ["nan", "positive-infinity", "negative-infinity"]) {
      expect(fixture.coverage.classifications[classification]).toBeGreaterThan(0);
    }
  });

  it.each(fixture.vectors)("$assert/$op($a, $b)", (vector) => {
    if (vector.assert === "exact") {
      expect(evaluateExact(vector)).toBe(vector.expect);
    } else if (vector.assert === "approx") {
      expectApprox(vector);
    } else {
      expectDecision(vector);
    }
  });
});

describe("numeric-core properties", () => {
  it("round-trips every canonical fixture input idempotently", () => {
    for (const vector of fixture.vectors) {
      for (const source of [vector.a, vector.b, vector.ratio]) {
        if (!source) continue;
        try {
          const parsed = parseCanonical(source);
          expect(canonicalString(parsed)).toBe(source);
        } catch {
          // Operation parameters and diagnostic values are intentionally not all wire values.
        }
      }
    }
  });

  it("rejects non-finite gameplay state", () => {
    for (const source of ["NaN", "Infinity", "-Infinity"]) {
      const value = new Decimal(source);
      expect(isStateValue(value)).toBe(false);
      expect(classify(value)).toBe(source.startsWith("-") ? "negative-infinity" : source === "NaN" ? "nan" : "positive-infinity");
      expect(() => parseCanonical(source)).toThrow();
    }
  });

  it("normalizes equivalent scientific coefficients before quantizing", () => {
    for (const source of ["12.345e2", "0.12345e4", "1.2345e3"]) {
      expect(canonicalString(source)).toBe("1.2345e3");
    }
  });

  it("rejects unsafe representation as state but canonicalizes a normalized clone", () => {
    const unsafe = Decimal.fromMantissaExponent_noNormalize(100, 0);
    expect(isStateValue(unsafe)).toBe(false);
    expect(canonicalString(unsafe)).toBe("1e2");
    expect(unsafe.mantissa).toBe(100);
    expect(unsafe.exponent).toBe(0);
  });

  it("requires canonical zero representation", () => {
    expect(isStateValue(Decimal.fromMantissaExponent_noNormalize(0, 7))).toBe(false);
    expect(isStateValue(new Decimal(0))).toBe(true);
  });
});
