import { describe, expect, it } from "vitest";

import fixtureJson from "../../testdata/economy-kernel.json";
import {
  canonicalBulkCost,
  parseCatalog,
  type EconomyCatalog,
} from "../src/economy-kernel";

interface CurveVector {
  generator_id: string;
  owned: number;
  count: number;
  cash: string;
  expect_cost: string;
  expect_affordable: number;
}

interface KernelFixture {
  version: number;
  catalog: Record<string, unknown>;
  curve_vectors: CurveVector[];
  invalid_cases: string[];
}

const fixture = fixtureJson as KernelFixture;

describe("shared economy catalog", () => {
  it("loads strict versioned definitions", () => {
    const catalog = parseCatalog(fixture.catalog);
    expect(fixture.version).toBe(1);
    expect(catalog.resources).toHaveLength(3);
    expect(catalog.generatorClasses).toHaveLength(5);
    expect(catalog.resource("company.cash")?.hardcap?.reasonKey).toBe(
      "resource.company_cash.cap.depleted",
    );
    expect(catalog.resource("company.users")?.hardcap).toBeNull();
    expect(catalog.generatorClass("generator.constant")?.production).toEqual({
      resourceId: "company.users",
      baseRate: "1e0",
    });
    expect(Object.isFrozen(catalog.resources)).toBe(true);
    expect(Object.isFrozen(catalog.resources[0])).toBe(true);
  });

  it.each(fixture.invalid_cases)("rejects %s", (name) => {
    expect(() => parseCatalog(mutateCatalog(fixture.catalog, name))).toThrow(SyntaxError);
  });

  it("keeps catalog v1 readable without production semantics", () => {
    const legacy = structuredClone(fixture.catalog);
    legacy.schema_version = 1;
    for (const generator of legacy.generator_classes as Array<Record<string, unknown>>) {
      delete generator.production;
    }
    const catalog = parseCatalog(legacy);
    expect(catalog.generatorClass("generator.constant")?.production).toBeNull();
  });
});

describe("shared cost-curve vectors", () => {
  const catalog = parseCatalog(fixture.catalog);

  it.each(fixture.curve_vectors)("quotes $generator_id", (vector) => {
    expect(
      canonicalBulkCost(catalog, vector.generator_id, vector.owned, vector.count),
    ).toBe(vector.expect_cost);
    const affordable = catalog.maxAffordable(vector.generator_id, vector.cash, vector.owned);
    expect(affordable).toBe(vector.expect_affordable);
    expect(catalog.bulkCost(vector.generator_id, vector.owned, affordable).lte(vector.cash)).toBe(
      true,
    );
    if (affordable < Number.MAX_SAFE_INTEGER - vector.owned) {
      expect(
        catalog.bulkCost(vector.generator_id, vector.owned, affordable + 1).gt(vector.cash),
      ).toBe(true);
    }
  });

  it("does not contain a global geometric ratio", () => {
    const ratios = catalog.generatorClasses
      .map((definition) => definition.price.curve)
      .filter((curve) => curve.kind === "geometric")
      .map((curve) => curve.ratio);
    expect(new Set(ratios)).toEqual(new Set(["1.1e0", "1.13e0"]));
  });
});

function mutateCatalog(source: Record<string, unknown>, name: string): unknown {
  const root = structuredClone(source);
  const resources = root.resources as Array<Record<string, unknown>>;
  const generators = root.generator_classes as Array<Record<string, unknown>>;
  const resource = resources[0];
  const generator = generators[0];
  const price = generator.price as Record<string, unknown>;
  const production = generator.production as Record<string, unknown>;

  switch (name) {
    case "unsupported-version":
      root.schema_version = 3;
      break;
    case "missing-root-field":
      delete root.resources;
      break;
    case "unknown-root-field":
      root.unexpected = true;
      break;
    case "missing-hardcap":
      delete resource.hardcap;
      break;
    case "unknown-nested-field":
      resource.unexpected = true;
      break;
    case "duplicate-resource-id":
      resources.push(structuredClone(resource));
      break;
    case "duplicate-generator-id":
      generators.push(structuredClone(generator));
      break;
    case "dangling-resource-id":
      price.resource_id = "company.missing";
      break;
    case "invalid-id":
      resource.id = "Company.Cash";
      break;
    case "invalid-scope":
      resource.scope = "universe";
      break;
    case "invalid-numeric-kind":
      resource.numeric_kind = "float64";
      break;
    case "invalid-decimal":
      resource.initial = "100";
      break;
    case "invalid-bounds":
      resource.minimum = "1e0";
      break;
    case "invalid-curve-kind":
      price.curve = { kind: "script" };
      break;
    case "invalid-curve-parameter":
      price.curve = { kind: "geometric", ratio: "9e-1" };
      break;
    case "missing-production":
      delete generator.production;
      break;
    case "dangling-production-resource":
      production.resource_id = "company.missing";
      break;
    case "nonpositive-production-rate":
      production.base_rate = "0";
      break;
    case "cross-scope-production":
      production.resource_id = "founder.reputation";
      break;
    default:
      throw new Error(`unimplemented invalid fixture case: ${name}`);
  }
  return root;
}

// Keep the type imported and checked in browser builds; the class is the read-only client seam.
const _catalogTypeCheck: EconomyCatalog = parseCatalog(fixture.catalog);
void _catalogTypeCheck;
