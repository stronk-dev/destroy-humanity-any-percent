import { describe, expect, it } from "vitest";

import phase0Json from "../../balance/catalogs/phase0.json";
import fixtureJson from "../../testdata/economy-kernel.json";
import engineFixtureJson from "../../testdata/production-engine.json";
import {
  canonicalBulkCost,
  parseCatalog,
  subProgressValue,
  type EconomyCatalog,
} from "../src/economy-kernel";
import { canonicalString } from "../src/numeric";

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
  multiplier_catalog_cases: Array<{ name: string; expect_valid: boolean }>;
}

const fixture = fixtureJson as KernelFixture;
const engineFixture = engineFixtureJson as {
  version: number;
  progress_cases: Array<{
    name: string;
    tier: number;
    cash: string;
    generator_count: number;
    expect: string;
  }>;
  resource_log_target_cases: Array<{
    name: string;
    target: string;
    expect_valid: boolean;
  }>;
  resource_log_progress_cases: Array<{
    name: string;
    target: string;
    value: string;
    expect: string;
  }>;
};

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

  it("loads the strict production catalog v3 contract", () => {
    const catalog = parseCatalog(phase0Json);
    expect(catalog.manualActions).toEqual([
      { id: "manual.click", output: { resourceId: "company.cash", amountPerAction: "1e0" } },
    ]);
    expect(catalog.multiplierSources).toEqual([]);
    expect(catalog.progressCoordinates.map((coordinate) => coordinate.tier)).toEqual([0, 1, 2, 3]);
    expect(catalog.manualPolicy).toEqual({ refillMilliPerMs: 25, bucketCapMilli: 50_000 });
    expect(catalog.offlinePolicy).toMatchObject({
      efficiency: "9e-1",
      accrualCapMs: 86_400_000,
      bankRatioNumerator: 1,
      bankRatioDenominator: 2,
      bankCapMs: 259_200_000,
    });
  });

  it.each(fixture.multiplier_catalog_cases)("validates multiplier case $name", (vector) => {
    const parse = () => parseCatalog(mutateMultiplierCatalog(vector.name));
    if (vector.expect_valid) expect(parse).not.toThrow();
    else expect(parse).toThrow(SyntaxError);
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

describe("shared progress coordinates", () => {
  const catalog = parseCatalog(phase0Json);

  it.each(engineFixture.progress_cases)("evaluates $name", (vector) => {
    expect(engineFixture.version).toBe(1);
    expect(canonicalString(subProgressValue(catalog, {
      balances: { "company.cash": vector.cash },
      generatorCounts: { "generator.beige_tower": vector.generator_count },
    }, vector.tier))).toBe(vector.expect);
  });

  it.each(engineFixture.resource_log_target_cases)(
    "validates $name in both positions",
    (vector) => {
      for (const composite of [false, true]) {
        const parse = () => catalogWithResourceLogTarget(vector.target, composite);
        if (vector.expect_valid) expect(parse).not.toThrow();
        else expect(parse).toThrow(SyntaxError);
      }
    },
  );

  it.each(engineFixture.resource_log_progress_cases)(
    "evaluates resource log $name",
    (vector) => {
      const boundaryCatalog = catalogWithResourceLogTarget(vector.target, false);
      expect(
        canonicalString(
          subProgressValue(
            boundaryCatalog,
            {
              balances: { "company.cash": vector.value },
              generatorCounts: { "generator.beige_tower": 0 },
            },
            0,
          ),
        ),
      ).toBe(vector.expect);
    },
  );

  it("defensively rejects a collapsed denominator after parsing", () => {
    const forgedCatalog = {
      progressCoordinates: [
        {
          tier: 0,
          kind: "resource_log",
          resourceId: "company.cash",
          target: "4e-15",
        },
      ],
    } as unknown as EconomyCatalog;
    expect(() =>
      subProgressValue(
        forgedCatalog,
        {
          balances: { "company.cash": "1e0" },
          generatorCounts: { "generator.beige_tower": 0 },
        },
        0,
      )).toThrow(RangeError);
  });
});

function catalogWithResourceLogTarget(target: string, composite: boolean): EconomyCatalog {
  const root = structuredClone(phase0Json) as Record<string, unknown>;
  const coordinates = root.progress_coordinates as Array<Record<string, unknown>>;
  if (composite) {
    const terms = coordinates[1].terms as Array<Record<string, unknown>>;
    const term = terms.find((candidate) => candidate.kind === "resource_log");
    if (!term) throw new Error("phase0 composite resource_log term is missing");
    term.target = target;
  } else {
    coordinates[0].target = target;
  }
  return parseCatalog(root);
}

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
      root.schema_version = 4;
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

function mutateMultiplierCatalog(name: string): unknown {
  const root = structuredClone(phase0Json) as Record<string, unknown>;
  const sources: Array<Record<string, unknown>> = [
    { id: "upgrade.a", slot: "upgrades", target: "all", provider: "upgrade.a" },
    {
      id: "upgrade.b",
      slot: "upgrades",
      target: "generator.beige_tower",
      provider: "upgrade.b",
    },
  ];
  switch (name) {
    case "valid-multiple-upgrades":
      break;
    case "duplicate-multiplier-source":
      sources[1].id = "upgrade.a";
      break;
    case "second-commons-provider":
      sources[0].slot = "commons";
      sources[1].slot = "commons";
      break;
    case "second-trust-provider":
      sources[0].slot = "trust";
      sources[1].slot = "trust";
      break;
    case "unknown-multiplier-slot":
      sources[0].slot = "dark_magic";
      break;
    case "unknown-multiplier-target":
      sources[0].target = "generator.missing";
      break;
    case "malformed-multiplier-provider":
      sources[0].provider = "Upgrade A";
      break;
    default:
      throw new Error(`unimplemented multiplier catalog case: ${name}`);
  }
  root.multiplier_sources = sources;
  return root;
}

// Keep the type imported and checked in browser builds; the class is the read-only client seam.
const _catalogTypeCheck: EconomyCatalog = parseCatalog(fixture.catalog);
void _catalogTypeCheck;
