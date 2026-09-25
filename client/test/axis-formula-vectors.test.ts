import vectorsJSON from "../../testdata/axis-stack/formula-vectors-v1.json";
import fixtureJSON from "../../balance/testdata/axis-stack/economy-v5-fixture.json";
import { describe, expect, it } from "vitest";

import { parseCatalog } from "../src/economy-kernel";
import { canonicalString } from "../src/numeric";
import { axisInput, contentContributions, contributionFactorForTarget, type ReplayState } from "../src/replay";

interface Vector { name: string; input: string; attainment_score_run: number; achievement_score_run: number; owned: string[] | null; extra_intern_ppm: number;
  x: number; saturated: boolean; contributions: string[]; product: string }
const vectors = (vectorsJSON as unknown as { vectors: Vector[] }).vectors;

function catalogFor(vector: Vector) {
  const root = structuredClone(fixtureJSON) as Record<string, unknown> & { axis_stack: Record<string, unknown>; upgrades: unknown[]; multiplier_sources: unknown[] };
  root.axis_stack.input = vector.input;
  if (vector.extra_intern_ppm > 0) {
    root.upgrades.push({ id: "upgrade.pr_intern_9", cost: { resource: "company.cash", amount: "1e9" }, window: { from_gate: "gate.t0_to_t1", to_gate: null },
      requires: [{ kind: "axis_at_least", minimum: 1 }], effects: [{ source_id: "upgrade.pr_intern_9.axis", slot: "axis_stack", target: "all", factor_ppm: vector.extra_intern_ppm }],
      roles: ["synergy_feed"], copy_key: "upgrade.pr_intern_9" });
    root.multiplier_sources.push({ id: "upgrade.pr_intern_9.axis", slot: "axis_stack", target: "all", provider: "axis_stack" });
  }
  return parseCatalog(root);
}

describe("Clout v1 axis stack formula (CV3, AC2) reproduces the Go-authored vectors", () => {
  it("covers the corpus", () => expect(vectors.length).toBeGreaterThanOrEqual(8));
  for (const vector of vectors) {
    it(vector.name, () => {
      const catalog = catalogFor(vector);
      const generators = Object.fromEntries(catalog.generatorClasses.filter((value) => value.production !== null).map((value) => [value.id, 0]));
      const state = { generators, upgradesOwned: new Set(vector.owned ?? []), achievementsAttainedRun: new Set(), attainmentScoreRun: vector.attainment_score_run,
        achievementScoreRun: vector.achievement_score_run } as unknown as ReplayState;
      expect(axisInput(state, catalog)).toEqual({ x: vector.x, saturated: vector.saturated });
      const contributions = contentContributions(state, catalog);
      expect(contributions.filter((value) => value.slot === "axis_stack").map((value) => `${value.source_id}=${value.factor}`)).toEqual(vector.contributions);
      expect(canonicalString(contributionFactorForTarget(catalog, "all", contributions))).toBe(vector.product);
    });
  }
});
