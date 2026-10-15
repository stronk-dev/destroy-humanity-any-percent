import { describe, expect, it } from "vitest";

import fixtureJson from "../../testdata/production-accrual.json";
import lowerDomainJson from "../../testdata/production-lower-domain-v1.json";
import upperDomainJson from "../../testdata/production-upper-domain-v1.json";
import { canonicalString } from "../src/numeric";
import { accrueConstant } from "../src/production";

interface Vector {
  name: string;
  rates: string[];
  elapsed_ms: number;
  efficiency: string;
  expect?: string;
  expect_error?: boolean;
}

const fixture = fixtureJson as { version: number; vectors: Vector[] };
const lowerDomain = lowerDomainJson as { version: number; vectors: Vector[] };
const upperDomain = upperDomainJson as { version: number; vectors: Vector[] };

describe("shared production accrual vectors", () => {
  it("uses vector schema 1 and complete populations", () => {
    expect(fixture.version).toBe(1);
    expect(lowerDomain.version).toBe(1);
    expect(upperDomain.version).toBe(1);
    expect(fixture.vectors).toHaveLength(16);
    expect(lowerDomain.vectors).toHaveLength(7);
    expect(upperDomain.vectors).toHaveLength(6);
  });

  it.each([...fixture.vectors, ...lowerDomain.vectors, ...upperDomain.vectors])("$name", (vector) => {
    const before = [...vector.rates];
    const evaluate = () => accrueConstant(vector.rates, vector.elapsed_ms, vector.efficiency);
    if (vector.expect_error) expect(evaluate).toThrow();
    else expect(canonicalString(evaluate())).toBe(vector.expect);
    expect(vector.rates).toEqual(before);
  });
});
