import { describe, expect, it } from "vitest";
import economyFixture from "../../testdata/economy-foundation-v4.json";
import policyFixture from "../../testdata/harness/relevance/policy-v1.json";
import routesFixture from "../../balance/routes/phase0.json";
import { parseRelevancePolicy } from "../src/relevance";

const catalog = {
  generators: economyFixture.generator_classes.map((row) => ({ id: row.id, tier: row.tier, category: row.category })),
  upgradeIds: economyFixture.upgrades.map((row) => row.id),
};
const gates = routesFixture.gates.map((row) => row.gate_id);

function clone(): typeof policyFixture {
  return structuredClone(policyFixture);
}

describe("relevance policy parity", () => {
  it("loads the shared complete policy", () => {
    const policy = parseRelevancePolicy(policyFixture, catalog, gates);
    expect(policy.items.map((row) => row.purchasable_id)).toEqual(["generator.high", "generator.low", "upgrade.click"]);
    expect(policy.groups).toHaveLength(4);
  });

  it("rejects the same missing, order, dangling, and biconditional mutations", () => {
    const missing = clone() as unknown as { items: Array<Record<string, unknown>> }; delete missing.items[0]?.epsilon_ms;
    const unsorted = clone(); [unsorted.items[0], unsorted.items[1]] = [unsorted.items[1]!, unsorted.items[0]!];
    const dangling = clone(); dangling.items[0]!.group_ids = ["group.missing"];
    const mismatch = clone(); mismatch.items[2]!.trap_exempt = false;
    for (const value of [missing, unsorted, dangling, mismatch]) expect(() => parseRelevancePolicy(value, catalog, gates)).toThrow();
  });
});
