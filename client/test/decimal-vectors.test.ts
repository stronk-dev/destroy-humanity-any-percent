import Decimal from "break_infinity.js";
import { describe, expect, it } from "vitest";

import fixtureJson from "../../testdata/decimal-vectors.json";
import { affordGeometricSeries, sumGeometricSeries } from "../src/economy";
import { canonicalString, classify, isStateValue, parseCanonical } from "../src/numeric";

interface Vector {
  assert: "exact" | "approx" | "decision";
  a: string;
  b: string;
  op: string;
  ratio?: string;
  owned?: string;
  expect: string;
  expectClass?: "finite" | "nan" | "positive-infinity" | "negative-infinity";
}

const fixture = fixtureJson as { version: number; seed: number; vectors: Vector[] };

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
    expect(fixture.version).toBe(2);
    expect(fixture.vectors.length).toBeGreaterThanOrEqual(5_000);
    expect(new Set(fixture.vectors.map((vector) => vector.assert))).toEqual(
      new Set(["exact", "approx", "decision"]),
    );
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
      expect(isStateValue(new Decimal(source))).toBe(false);
      expect(() => parseCanonical(source)).toThrow();
    }
  });
});
