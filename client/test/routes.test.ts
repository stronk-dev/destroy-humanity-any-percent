import { describe, expect, it } from "vitest";

import catalogJson from "../../balance/routes/phase0.json";
import economyJson from "../../balance/catalogs/phase0.json";
import validCatalogJson from "../../balance/routes-testdata/valid/minimal.json";
import reachableCatalogJson from "../../balance/routes-testdata/invalid/reachable-depletion.json";
import unknownFieldCatalogJson from "../../balance/routes-testdata/invalid/unknown-field.json";
import unboundExclusionCatalogJson from "../../balance/routes-testdata/invalid/unbound-exclusion.json";
import temporalImpossibilityCatalogJson from "../../balance/routes-testdata/invalid/temporal-impossibility.json";
import danglingResourceCatalogJson from "../../balance/routes-testdata/invalid/dangling-resource.json";
import vectorsJson from "../../testdata/routes/predicate-vectors.json";
import { parseCatalog } from "../src/economy-kernel";
import { evaluatePredicate, parseRoutePredicate, parseRoutesCatalog, validateRouteCatalogResources, type RouteContext } from "../src/routes";

interface FixtureContext {
  context_version: number;
  resources: Record<string, string>;
  doctrines_by_transition: Record<string, string>;
  structure_id: string;
  ledger_fact_kinds: string[];
  meter_bands: Record<string, number>;
  region_traits: string[];
}

const fixture = vectorsJson as unknown as { schema_version: number; vectors: Array<{ name: string; condition: unknown; context: FixtureContext; expected: boolean }> };

function context(source: FixtureContext): RouteContext {
  return {
    contextVersion: source.context_version,
    resources: source.resources,
    doctrinesByTransition: source.doctrines_by_transition,
    structureId: source.structure_id,
    ledgerFactKinds: new Set(source.ledger_fact_kinds),
    meterBands: source.meter_bands,
    regionTraits: new Set(source.region_traits),
  };
}

describe("routes catalog and predicate parity", () => {
  it("loads the shipped catalog and proves Depletion unreachable in one run", () => {
    const catalog = parseRoutesCatalog(catalogJson);
    expect(() => validateRouteCatalogResources(catalog, parseCatalog(economyJson))).not.toThrow();
    expect(catalog.gates).toHaveLength(3);
    expect(catalog.maxRoutesPerRun()).toBe(4);
    expect(catalog.depletionDistinctRoutesRequired).toBe(5);
    expect(catalog.route("route.nonprofit_wrapper_zip")?.effect).toEqual({ kind: "substitute" });
  });

  it.each(fixture.vectors)("evaluates $name", (vector) => {
    expect(evaluatePredicate(parseRoutePredicate(vector.condition), context(vector.context))).toBe(vector.expected);
  });

  it("rejects a single-run-reachable catalog and unavailable active context", () => {
    expect(parseRoutesCatalog(validCatalogJson).gates).toHaveLength(1);
    expect(() => parseRoutesCatalog(reachableCatalogJson)).toThrow(/reachable/);
    expect(() => parseRoutesCatalog(unknownFieldCatalogJson)).toThrow(/fields/);
    expect(() => parseRoutesCatalog(unboundExclusionCatalogJson)).toThrow(/exclusion/);
    expect(() => parseRoutesCatalog(temporalImpossibilityCatalogJson)).toThrow(/after gate/);
    expect(() => validateRouteCatalogResources(parseRoutesCatalog(danglingResourceCatalogJson), parseCatalog(economyJson))).toThrow(/unknown company resource/);
    const reachable = structuredClone(catalogJson) as typeof catalogJson;
    reachable.depletion_distinct_routes_required = 4;
    expect(() => parseRoutesCatalog(reachable)).toThrow(/reachable/);
    const unavailable = structuredClone(catalogJson) as typeof catalogJson;
    unavailable.gates[1]!.routes[0]!.active = true;
    expect(() => parseRoutesCatalog(unavailable)).toThrow(/unavailable context/);
    const lyingContext = structuredClone(catalogJson) as typeof catalogJson;
    lyingContext.gates[1]!.routes[0]!.requires_context_version = 1;
    expect(() => parseRoutesCatalog(lyingContext)).toThrow(/context version 2/);
    const sameBoundary = structuredClone(catalogJson) as typeof catalogJson;
    sameBoundary.gates[1]!.gate_id = "gate.t3_to_t4";
    expect(() => parseRoutesCatalog(sameBoundary)).not.toThrow();
  });

  it("rejects unknown predicate fields", () => {
    expect(() => parseRoutePredicate({ kind: "structure_is", structure_id: "structure.nonprofit", surprise: true })).toThrow(/fields/);
  });
});
