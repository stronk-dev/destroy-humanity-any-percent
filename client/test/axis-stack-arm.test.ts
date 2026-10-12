import axisArm from "../../testdata/axis-stack/game-ui-arm-v1.json";
import { describe, expect, it } from "vitest";

import { parseAxisStackArm, parseFeatures } from "../src/game-ui/contracts";

// Clout v1 CV9: the decoder accepts the Go producer's golden arm and rejects
// every internal contradiction.
describe("axis stack arm decoder", () => {
  const clone = () => structuredClone(axisArm) as Record<string, any>;
  it("accepts the Go-projected golden arm", () => expect(() => parseAxisStackArm(axisArm)).not.toThrow());
  const cases: Record<string, (arm: Record<string, any>) => void> = {
    saturation_contradicts_input: (arm) => { arm.saturated = true; },
    contribution_for_unowned_intern: (arm) => { arm.contributions.push({ factor: arm.interns[1].factor, source_id: "upgrade.pr_intern_2.axis", upgrade_id: "upgrade.pr_intern_2" }); },
    contribution_factor_differs: (arm) => { arm.contributions[0].factor = "9e0"; },
    owned_intern_without_contribution: (arm) => { arm.contributions = []; },
    unsorted_interns: (arm) => { arm.interns.reverse(); },
    non_canonical_product: (arm) => { arm.product = "1.20"; },
    unknown_input_kind: (arm) => { arm.input_kind = "clout_run"; },
    extra_field: (arm) => { arm.local_factor = "1e0"; },
  };
  for (const [name, mutate] of Object.entries(cases)) it(`rejects ${name}`, () => { const arm = clone(); mutate(arm); expect(() => parseAxisStackArm(arm)).toThrow(); });

  const bothOwned = () => {
    const arm = clone();
    arm.input_value = 12;
    arm.product = "1.612e0";
    arm.interns[0].factor = "1.3e0";
    arm.interns[1].factor = "1.24e0";
    arm.interns[1].owned = true;
    arm.attained.splice(1, 0, { achievement_id: "achievement.generators_owned_100", earned_this_run: false });
    arm.contributions = arm.interns.map((row: { upgrade_id: string; factor: string }) => ({
      source_id: `${row.upgrade_id}.axis`, upgrade_id: row.upgrade_id, factor: row.factor,
    }));
    return arm;
  };
  const features = (axis_stack: unknown) => ({
    achievements: null, active_play: null, fiscal: null, meters: null, minigames: null, pets: null, axis_stack,
  });

  it("accepts one distinct contribution for each owned intern without rewriting supplied values", () => {
    const arm = bothOwned(), before = structuredClone(arm);
    expect(() => parseAxisStackArm(arm)).not.toThrow();
    expect(parseFeatures(features(arm)).axis_stack).toBe(arm);
    expect(arm).toEqual(before);
  });

  it.each([0, 1])("rejects duplicating intern %i while omitting the other owned intern", (repeated) => {
    const arm = bothOwned(), omitted = 1 - repeated;
    // The source IDs stay sorted and unique, both factors match their named
    // intern, and the row count still equals the owned count. Counts alone do
    // not prove that every owned intern contributes exactly once.
    arm.contributions[omitted].upgrade_id = arm.interns[repeated].upgrade_id;
    arm.contributions[omitted].factor = arm.interns[repeated].factor;
    const before = structuredClone(arm);
    expect(() => parseAxisStackArm(arm)).toThrow(/exactly once/);
    expect(() => parseFeatures(features(arm))).toThrow(/exactly once/);
    expect(arm).toEqual(before);
  });

  it("accepts distinct owned interns with equal supplied factors", () => {
    const arm = bothOwned();
    arm.product = "1.69e0";
    arm.interns[1].factor_ppm = arm.interns[0].factor_ppm;
    for (const row of [...arm.interns, ...arm.contributions]) row.factor = "1.3e0";
    expect(() => parseAxisStackArm(arm)).not.toThrow();
  });
});
