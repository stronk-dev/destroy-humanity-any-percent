import axisArm from "../../testdata/axis-stack/game-ui-arm-v1.json";
import { describe, expect, it } from "vitest";

import { parseAxisStackArm } from "../src/game-ui/contracts";

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
});
