import { describe, expect, it } from "vitest";
import achievementCatalogParityCorpus from "../../balance/testdata/achievements-catalog-parity-v1.json";

import { loadAchievementCatalog, type AchievementRegistry } from "../src/achievements/catalog";
import { achievementScore, newlyEarned } from "../src/achievements/evaluate";

const registry: AchievementRegistry = {
  copyKeys: new Set(["achievement.first_gate", "achievement.possession_warning"]),
  generatorIds: new Set(["generator.clickfarm"]),
  eventKinds: new Set(["gate_crossed", "generator_purchased"]),
  resourceIds: new Set(["company.cash"]),
  runCounters: new Set(["generators_purchased_total", "tier"]),
  careerCounters: new Set(["age_ms", "notoriety"]),
  provenanceSources: new Map([["fact:gate.tier_1", ["gate_crossed"]]]),
};

function validCatalog(): Record<string, unknown> {
  return { schema_version: 1, achievements: [
    { id: "achievement.first_gate", condition_scope: "run", condition: { kind: "fact_present", fact_kind: "gate.tier_1" }, proof: { kind: "provenance", event_kinds: ["gate_crossed"] }, score_grant: 4, copy_key: "achievement.first_gate" },
    { id: "achievement.generator_hoard", condition_scope: "run", condition: { kind: "owns_generator_at_least", generator_id: "generator.clickfarm", count: 300 }, proof: { kind: "possession", justification_copy_key: "achievement.possession_warning" }, score_grant: 8, copy_key: "achievement.first_gate" },
  ] };
}

function applyParityOperation(root: any, operation: { op: string; path: Array<string | number>; value?: unknown }): void {
  let current = root;
  for (const component of operation.path.slice(0, -1)) current = current[component];
  const last = operation.path.at(-1);
  if (last === undefined) throw new Error("empty parity operation path");
  if (operation.op === "delete") delete current[last];
  else if (operation.op === "set") current[last] = operation.value;
  else throw new Error(`invalid parity operation ${operation.op}`);
}

describe("achievement catalog", () => {
  it("loads the closed proof union, evaluates byte order, and derives score", () => {
    const catalog = loadAchievementCatalog(JSON.stringify(validCatalog()), registry);
    const empty = { facts: new Set<string>(), counters: {}, exitCount: 0, generators: {} };
    const run = { ...empty, facts: new Set(["gate.tier_1"]), generators: { "generator.clickfarm": 300 } };
    const earned = newlyEarned(catalog, new Set(), new Set(), run, empty);
    expect(earned.map((value) => value.id)).toEqual(["achievement.first_gate", "achievement.generator_hoard"]);
    expect(achievementScore(catalog, new Set(earned.map((value) => value.id)))).toBe(12);
    expect(newlyEarned(catalog, new Set([earned[0]!.id]), new Set([earned[1]!.id]), run, empty)).toEqual([]);
  });

  it("rejects bare possession, stale Clout fields, unknown copy, and scope drift", () => {
    const cases: Array<(value: any) => void> = [
      (value) => { value.achievements[1].proof.justification_copy_key = "missing.copy"; },
      (value) => { value.achievements[0].proof = { kind: "possession", justification_copy_key: "achievement.possession_warning" }; },
      (value) => { value.achievements[1].proof = { kind: "provenance", event_kinds: ["generator_purchased"] }; },
      (value) => { value.achievements[0].clout_grant_ppm = 4; },
      (value) => { value.achievements[1].condition_scope = "career"; },
      (value) => { value.achievements[0].proof.event_kinds = ["generator_purchased"]; },
    ];
    for (const mutate of cases) { const value = validCatalog(); mutate(value); expect(() => loadAchievementCatalog(JSON.stringify(value), registry)).toThrow(); }
  });

  it("matches the shared Go/TypeScript catalog parity corpus", () => {
    expect(achievementCatalogParityCorpus.version).toBe(1);
    for (const vector of achievementCatalogParityCorpus.cases) {
      const catalog = validCatalog();
      for (const operation of vector.operations) applyParityOperation(catalog, operation);
      const load = () => loadAchievementCatalog(JSON.stringify(catalog), registry);
      if (vector.valid) expect(load, vector.name).not.toThrow();
      else expect(load, vector.name).toThrow();
    }
  });
});
