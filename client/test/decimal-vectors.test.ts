import { readFileSync } from "node:fs";
import Decimal from "break_eternity.js";
import { describe, expect, it } from "vitest";

import {
  affordGeometricSeries,
  sumGeometricSeries,
} from "../src/economy";

interface Vector {
  a: string;
  b: string;
  op: string;
  ratio?: string;
  owned?: string;
  expect: string;
}

interface VectorFile {
  seed: number;
  vectors: Vector[];
}

const vectorPath = new URL(
  "../../testdata/decimal-vectors.json",
  import.meta.url,
);
const fixture = JSON.parse(readFileSync(vectorPath, "utf8")) as VectorFile;

function evaluate(vector: Vector): string {
  const a = new Decimal(vector.a);
  const b = vector.b === "" ? null : new Decimal(vector.b);

  switch (vector.op) {
    case "add":
    case "sub":
    case "mul":
    case "div":
      return a[vector.op](b!).toString();
    case "pow":
      return a.pow(Number(vector.b)).toString();
    case "log10":
    case "ln":
    case "exp":
    case "floor":
      return a[vector.op]().toString();
    case "cmp":
      return String(a.cmp(b!));
    case "sum":
      return sumGeometricSeries(
        Number(vector.a),
        vector.b,
        vector.ratio!,
        Number(vector.owned),
      ).toString();
    case "afford":
      return affordGeometricSeries(
        vector.a,
        vector.b,
        vector.ratio!,
        Number(vector.owned),
      ).toString();
    default:
      throw new Error(`unknown vector operation: ${vector.op}`);
  }
}

describe("shared decimal golden vectors", () => {
  it("contains the RFC minimum coverage", () => {
    expect(fixture.vectors.length).toBeGreaterThanOrEqual(5_000);
    expect(new Set(fixture.vectors.map((vector) => vector.op))).toEqual(
      new Set([
        "add",
        "sub",
        "mul",
        "div",
        "pow",
        "log10",
        "ln",
        "exp",
        "floor",
        "cmp",
        "sum",
        "afford",
      ]),
    );
  });

  it.each(fixture.vectors)("$op($a, $b)", (vector) => {
    expect(evaluate(vector)).toBe(vector.expect);
  });
});

