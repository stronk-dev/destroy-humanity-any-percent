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
  it("requires exact member names across the root, failure, branches and starters", () => {
    const objects = (root: any): Array<Record<string, unknown>> => [root, root.first_failure, ...root.first_failure.branches.flatMap((row: any) => [row, row.starter_package])];
    for (const [index, object] of objects(curriculumJSON).entries()) {
      for (const key of Object.keys(object)) {
        const source = structuredClone(curriculumJSON);
        const changed = objects(source)[index];
        changed[key.toUpperCase()] = changed[key];
        delete changed[key];
        expect(() => parseCurriculumCatalog(source, economy, copyKeys, gateIDs), `${index}.${key}`).toThrow(SyntaxError);
      }
    }
  });
  it("requires explicit zero fields and the ruled branch/starter shapes", () => {
    const edits: Array<(source: any) => void> = [
      (v) => { delete v.first_failure.founder_exit_count; }, (v) => { v.first_failure.founder_exit_count = null; },
      (v) => { delete v.first_failure.branches[0].route_knowledge_bonus; }, (v) => { v.first_failure.branches[0].route_knowledge_bonus = null; },
      (v) => { v.first_failure.branches[1].minimum_owned_upgrades = 0; }, (v) => { v.first_failure.branches[0].starter_package.count = 0; },
    ];
    for (let index = 0; index < 3; index++) edits.push((v) => { v.first_failure.branches[index].starter_package = v.first_failure.branches[(index + 1) % 3].starter_package; });
    for (const edit of edits) {
      const source = structuredClone(curriculumJSON);
      edit(source);
      expect(() => parseCurriculumCatalog(source, economy, copyKeys, gateIDs)).toThrow(SyntaxError);
    }
    expect(parseCurriculumCatalog(curriculumJSON, economy, copyKeys, gateIDs).first_failure.founder_exit_count).toBe(0);
  });
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
