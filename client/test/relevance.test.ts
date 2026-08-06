import { describe, expect, it } from "vitest";
import economyFixture from "../../testdata/economy-foundation-v4.json";
import mutationFixture from "../../testdata/harness/relevance/policy-mutations-v1.json";
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

  it("applies the shared Go/TypeScript mutation corpus", () => {
    expect(mutationFixture.schema_version).toBe(1);
    for (const test of mutationFixture.cases) {
      const value = clone() as unknown as {
        schema_version: number;
        items: Array<Record<string, unknown> & { availability_window: Record<string, unknown>; group_ids: string[] }>;
        groups: Array<Record<string, unknown> & { member_ids: string[] }>;
      };
      switch (test.id) {
        case "missing_item_epsilon": delete value.items[0]!.epsilon_ms; break;
        case "unsorted_items": [value.items[0], value.items[1]] = [value.items[1]!, value.items[0]!]; break;
        case "dangling_group": value.items[0]!.group_ids = ["group.missing"]; break;
        case "exemption_mismatch": value.items[2]!.trap_exempt = false; break;
        case "incomplete_derived_group": value.groups[0]!.member_ids = ["generator.high"]; break;
        case "unknown_gate": value.items[0]!.availability_window.from_gate = "gate.unknown"; break;
        case "integral_decimal_epsilon": value.items[0]!.epsilon_ms = 1000.0; break;
        case "integral_decimal_schema": value.schema_version = 1.0; break;
        default: throw new Error(`unknown relevance mutation ${test.id}`);
      }
      if (test.accepted) expect(() => parseRelevancePolicy(value, catalog, gates), test.id).not.toThrow();
      else expect(() => parseRelevancePolicy(value, catalog, gates), test.id).toThrow();
    }
  });
});
