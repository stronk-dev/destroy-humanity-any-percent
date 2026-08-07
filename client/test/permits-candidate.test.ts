import { describe, expect, it } from "vitest";

import doctrineCorpus from "../../balance/testdata/doctrines-catalog-parity-v1.json";
import economyCandidate from "../../balance/testdata/valid/permits-economy-candidate-v1.json";
import routesCandidate from "../../balance/testdata/permits-t3-gate-candidate-v1.json";
import hardcapVector from "../../testdata/permits-hardcap-v1.json";
import { loadDoctrineCatalog, validateDoctrineRoutes } from "../src/doctrines";
import { parseCatalog } from "../src/economy-kernel";
import { parseRoutesCatalog, validateRouteCatalogResources } from "../src/routes";
import { PredictionMachine } from "../src/shell/prediction";
import { parseClientShellPolicy } from "../src/shell/policy";
import policySource from "../../balance/client-shell/phase0.json";

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

  it("clamps the shared near-cap vector on the client prediction path", () => {
    const economy = parseCatalog(economyCandidate);
    const resource = economy.resource(hardcapVector.resource_id);
    expect(resource?.hardcap).toEqual({ amount: hardcapVector.hardcap, reasonKey: hardcapVector.hardcap_reason_key });
    expect(economy.generatorClass("generator.legal_dept")?.production?.baseRate).toBe(hardcapVector.rate_per_second);

    const machine = new PredictionMachine(parseClientShellPolicy(policySource));
    machine.initialize({
      revision: 1,
      evaluatedThroughMs: 0,
      constantsHash: `sha256:${"a".repeat(64)}`,
      resources: {
        [hardcapVector.resource_id]: {
          amount: hardcapVector.start,
          ratePerSecond: hardcapVector.rate_per_second,
          cap: { amount: hardcapVector.hardcap, reasonKey: hardcapVector.hardcap_reason_key },
        },
      },
      discrete: {},
      progress: [],
    }, 0);
    machine.pulse(hardcapVector.elapsed_ms);
    expect(machine.snapshot(hardcapVector.elapsed_ms).resources[hardcapVector.resource_id]).toEqual({ mantissa: 2.4, exponent: 1 });
  });
});
