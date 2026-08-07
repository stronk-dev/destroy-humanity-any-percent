import { describe, expect, it } from "vitest";

import doctrineCorpus from "../../balance/testdata/doctrines-catalog-parity-v1.json";
import economyCandidate from "../../balance/testdata/valid/permits-economy-candidate-v1.json";
import routesCandidate from "../../balance/testdata/permits-t3-gate-candidate-v1.json";
import { loadDoctrineCatalog, validateDoctrineRoutes } from "../src/doctrines";
import { parseCatalog } from "../src/economy-kernel";
import { parseRoutesCatalog, validateRouteCatalogResources } from "../src/routes";

describe("Permits and T3 gate candidate artifacts", () => {
  it("loads byte-identically across the economy, routes, and doctrine boundaries", () => {
    const economy = parseCatalog(economyCandidate);
    const routes = parseRoutesCatalog(routesCandidate);
    const doctrines = loadDoctrineCatalog(doctrineCorpus.baseline);

    expect(economy.resource("company.permits")?.hardcap).toEqual({
      amount: "2.4e1",
      reasonKey: "resource.company_permits.cap.phase0",
    });
    expect(economy.generatorClass("generator.legal_dept")?.production).toEqual({
      resourceId: "company.permits",
      baseRate: "1e-3",
    });
    expect(routes.gates.map((gate) => gate.gateId)).toEqual([
      "gate.t2_to_t3",
      "gate.t3_to_t4",
      "gate.t4_to_t5",
      "gate.t7_to_t8",
    ]);
    expect(routes.gates[1]?.requirement).toEqual([
      { resourceId: "company.cash", amount: "1e12" },
      { resourceId: "company.permits", amount: "1.2e1" },
    ]);
    expect(() => validateRouteCatalogResources(routes, economy)).not.toThrow();
    expect(() => validateDoctrineRoutes(doctrines, routes)).not.toThrow();
  });
});
