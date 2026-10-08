import { describe, expect, it } from "vitest";
import candidate from "../../balance/testdata/t2/presentation-candidate-v3.json";
import { GAME_UI_PRESENTATION, parseGameUIPresentation } from "../src/game-ui/presentation";
import { gatePresentation, generatorPresentation, tier2UpgradePresentation } from "../src/game-ui/tier2-presentation";
import { upgradePresentation } from "../src/game-ui/axis-presentation";

describe("Tier-2 candidate presentation consumption (E4)", () => {
  it("binds the actual three generators, three upgrades and gate", () => {
    for (const id of ["generator.open_plan_floor", "generator.managed_services_contract", "generator.hot_desk_program"]) {
      expect(generatorPresentation(id)).toEqual(candidate.generators.find((row) => row.id === id));
    }
    for (const id of ["upgrade.ping_pong_table", "upgrade.move_fast_break_things", "upgrade.nap_pod"]) {
      expect(upgradePresentation(id)).toEqual(candidate.upgrades.find((row) => row.id === id));
    }
    expect(gatePresentation("gate.t1_to_t2")).toEqual(candidate.gates.find((row) => row.id === "gate.t1_to_t2"));
  });

  it("preserves every pinned binding, including the older Garage cap binding", () => {
    for (const [id, row] of GAME_UI_PRESENTATION.generators) expect(generatorPresentation(id)).toBe(row);
    for (const [id, row] of GAME_UI_PRESENTATION.upgrades) expect(tier2UpgradePresentation(id)).toBe(row);
    for (const [id, row] of GAME_UI_PRESENTATION.gates) expect(gatePresentation(id)).toBe(row);
    expect(upgradePresentation("upgrade.pr_intern_1").id).toBe("upgrade.pr_intern_1");
  });

  it("refuses unknown content and malformed candidate bindings instead of inventing labels", () => {
    for (const resolve of [generatorPresentation, upgradePresentation, gatePresentation]) expect(() => resolve("unknown.id")).toThrow(/missing presentation binding/u);
    const duplicate = structuredClone(candidate); duplicate.generators.splice(1, 0, { ...duplicate.generators[0]! });
    expect(() => parseGameUIPresentation(duplicate)).toThrow(/sorted and unique/u);
    const extra = structuredClone(candidate); Object.assign(extra.generators[0]!, { unruled: true });
    expect(() => parseGameUIPresentation(extra)).toThrow(/fields are not exact/u);
    const version = structuredClone(candidate); version.schema_version = 2;
    expect(() => parseGameUIPresentation(version)).toThrow(/schema v3/u);
  });
});
