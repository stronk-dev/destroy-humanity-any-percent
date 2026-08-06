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

type MutationCase = (typeof mutationFixture.cases)[number];

function mutationParent(root: unknown, path: readonly string[]): { parent: Record<string, unknown> | unknown[]; key: string } {
  if (path.length === 0) throw new Error("empty relevance mutation path");
  let value: unknown = root;
  for (const component of path.slice(0, -1)) {
    if (Array.isArray(value)) value = value[Number(component)];
    else if (value !== null && typeof value === "object") value = (value as Record<string, unknown>)[component];
    else throw new Error(`invalid relevance mutation path ${path.join(".")}`);
  }
  if (!Array.isArray(value) && (value === null || typeof value !== "object")) throw new Error(`invalid relevance mutation parent ${path.join(".")}`);
  return { parent: value as Record<string, unknown> | unknown[], key: path[path.length - 1] as string };
}

function applyMutation(test: MutationCase): string {
  const root = clone() as unknown;
  const { parent, key } = mutationParent(root, test.path);
  const index = Array.isArray(parent) ? Number(key) : key;
  if (test.operation === "delete") {
    if (Array.isArray(parent)) parent.splice(index as number, 1); else delete parent[index as string];
  } else if (test.operation === "swap") {
    if (!Array.isArray(parent) || test.swap_index === null) throw new Error(`invalid swap ${test.id}`);
    [parent[index as number], parent[test.swap_index]] = [parent[test.swap_index], parent[index as number]];
  } else if (test.operation === "replace" || test.operation === "replace_number") {
    if (test.value_json === null) throw new Error(`missing replacement ${test.id}`);
    const replacement = test.operation === "replace_number" ? `__RAW_RELEVANCE_NUMBER_${test.id}__` : JSON.parse(test.value_json) as unknown;
    if (Array.isArray(parent)) parent[index as number] = replacement;
    else parent[index as string] = replacement;
    const encoded = JSON.stringify(root);
    return test.operation === "replace_number" ? encoded.replace(`"${replacement}"`, test.value_json) : encoded;
  } else throw new Error(`unknown relevance mutation ${test.operation}`);
  return JSON.stringify(root);
}

describe("relevance policy parity", () => {
  it("loads the shared complete policy", () => {
    const policy = parseRelevancePolicy(JSON.stringify(policyFixture), catalog, gates);
    expect(policy.items.map((row) => row.purchasable_id)).toEqual(["generator.high", "generator.low", "upgrade.click"]);
    expect(policy.groups).toHaveLength(4);
  });

  it("applies the shared Go/TypeScript mutation corpus", () => {
    expect(mutationFixture.schema_version).toBe(1);
    for (const test of mutationFixture.cases) {
      const value = applyMutation(test);
      if (test.accepted) expect(() => parseRelevancePolicy(value, catalog, gates), test.id).not.toThrow();
      else expect(() => parseRelevancePolicy(value, catalog, gates), test.id).toThrow();
    }
  });
});
