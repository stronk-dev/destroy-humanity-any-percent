import economyJSON from "../../balance/catalogs/phase0.json";
import curriculumJSON from "../../balance/testdata/t0-t1/curriculum-v2.json";
import { describe, expect, it } from "vitest";

import { COPY_KEYS } from "../src/copy/generated/types";
import { parseCurriculumCatalog } from "../src/curriculum";
import { parseCatalog } from "../src/economy-kernel";

const economy = parseCatalog(economyJSON);
const copyKeys = new Set<string>(COPY_KEYS);
const gateIDs = ["gate.t0_to_t1"];

describe("curriculum catalog", () => {
  it("pins the ratified candidate tuple without compiling it into the loader", () => {
    const candidate = parseCurriculumCatalog(curriculumJSON, economy, copyKeys, gateIDs);
    expect(candidate.first_failure.branches).toMatchObject([
      {
        branch: "acquihire",
        minimum_purchased_generators: 200,
        route_knowledge_bonus: 0,
        starter_package: { kind: "resource_grant", amount: "1e4" },
      },
      {
        branch: "burnout",
        cheapest_price_factor: "2e0",
        route_knowledge_bonus: 50,
        starter_package: { kind: "generated_generators", count: 10 },
      },
      {
        branch: "pivot",
        route_knowledge_bonus: 25,
        starter_package: { kind: "preowned_upgrade" },
      },
    ]);

    const retuned = JSON.parse(JSON.stringify(curriculumJSON));
    retuned.first_failure.branches[0]!.minimum_purchased_generators = 201;
    retuned.first_failure.branches[0]!.route_knowledge_bonus = 1;
    retuned.first_failure.branches[0]!.starter_package.amount = "2e4";
    retuned.first_failure.branches[1]!.cheapest_price_factor = "3e0";
    retuned.first_failure.branches[1]!.starter_package.count = 11;
    expect(() => parseCurriculumCatalog(retuned, economy, copyKeys, gateIDs)).not.toThrow();
  });

  it("rejects unbound and out-of-domain starter values", () => {
    const unknown = JSON.parse(JSON.stringify(curriculumJSON));
    unknown.first_failure.branches[2]!.starter_package.upgrade_id = "upgrade.unknown";
    expect(() => parseCurriculumCatalog(unknown, economy, copyKeys, gateIDs)).toThrow(/starter/u);

    const invalid = JSON.parse(JSON.stringify(curriculumJSON));
    invalid.first_failure.branches[1]!.starter_package.count = 0;
    expect(() => parseCurriculumCatalog(invalid, economy, copyKeys, gateIDs)).toThrow(/starter/u);
  });
});
