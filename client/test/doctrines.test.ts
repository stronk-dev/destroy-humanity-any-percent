import { describe, expect, it } from "vitest";
import corpus from "../../balance/testdata/doctrines-catalog-parity-v1.json";
import { loadDoctrineCatalog, validateDoctrineRoutes } from "../src/doctrines";
import { parseRoutesCatalog } from "../src/routes";

function mutate(root: any, operation: { op: string; path: Array<string | number>; value?: unknown }): void {
  let current = root;
  for (const component of operation.path.slice(0, -1)) current = current[component];
  const last = operation.path.at(-1)!;
  if (operation.op === "delete") delete current[last]; else current[last] = operation.value;
}

describe("doctrine catalog", () => {
  it("matches the shared Go/TypeScript parity corpus", () => {
    expect(corpus.version).toBe(1);
    for (const vector of corpus.cases) {
      const candidate = structuredClone(corpus.baseline);
      for (const operation of vector.operations) mutate(candidate, operation);
      const load = () => loadDoctrineCatalog(candidate);
      if (vector.valid) expect(load, vector.name).not.toThrow(); else expect(load, vector.name).toThrow();
    }
  });

  it("closes doctrine references across the routes artifact", () => {
    const doctrines = loadDoctrineCatalog(corpus.baseline);
    const routes = parseRoutesCatalog({ schema_version: 1, context_version: 1, depletion_distinct_routes_required: 2, knowledge: { registry_first_bonus: 100, founder_first_grant: 25, repeat_grant: 5, hint_cost: 50 }, gates: [
      { gate_id: "gate.t3_to_t4", requirement: [{ resource_id: "company.cash", amount: "1e12" }], routes: [] },
      { gate_id: "gate.t4_to_t5", requirement: [{ resource_id: "company.cash", amount: "1e15" }], routes: [{ route_id: "route.capture", house_name: "Capture", active: true, requires_context_version: 1, exclusion_slot: "doctrine:transition.t3_to_t4", exclusion_value: "doctrine.capture", predicate: [{ kind: "doctrine_is", transition: "transition.t3_to_t4", doctrine_id: "doctrine.capture" }], effect: { kind: "discount", fraction: "5e-1" } }] },
    ] });
    expect(() => validateDoctrineRoutes(doctrines, routes)).not.toThrow();
    const missing = loadDoctrineCatalog({ schema_version: 1, transitions: [{ transition_id: "transition.t3_to_t4", source_tier: 3, gate_id: "gate.t3_to_t4", doctrine_ids: ["doctrine.ethical", "doctrine.other"] }] });
    expect(() => validateDoctrineRoutes(missing, routes)).toThrow(/undeclared/);
  });
});
