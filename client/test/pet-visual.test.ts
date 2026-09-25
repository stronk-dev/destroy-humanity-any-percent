import { describe, expect, it } from "vitest";

import speciesFixture from "../../balance/testdata/pet-species/fixture-v1.json";
import { contrastRatio, parsePetPalette, petVisualSpec, PET_SWATCHES } from "../src/game-ui/pet/visual";
import { UI_ERAS, UI_THEMES } from "../src/ui/themes";

describe("pet visual contract (PA8.5)", () => {
  it("maps bands to poses and never animates under reduced motion", () => {
    const identity = { palette_id: "pet_palette.fur_03" };
    expect(["high", "normal", "low", "floor"].map((band) => petVisualSpec(identity, band as never, false).pose)).toEqual(["content", "content", "low", "withdrawn"]);
    expect(petVisualSpec(identity, "high", true)).toEqual({ family: "cat", palette_id: "pet_palette.fur_03", pose: "content", animate: false });
    expect(petVisualSpec(identity, "high", false).animate).toBe(true);
    expect(() => petVisualSpec({ palette_id: "pet_palette.fur_99" }, "high", false)).toThrow();
  });

  it("has a swatch for every fixture palette id whose outline meets 3:1 non-text contrast on every era surface", () => {
    const ids = speciesFixture.species.flatMap((row) => row.palette_ids);
    expect([...PET_SWATCHES.keys()].sort()).toEqual([...ids].sort());
    for (const era of UI_ERAS) {
      for (const background of [UI_THEMES[era].color.surface, UI_THEMES[era].color.bg]) {
        for (const [id, swatch] of PET_SWATCHES) expect(contrastRatio(swatch.outline, background), `${era} ${id} on ${background}`).toBeGreaterThanOrEqual(3);
      }
    }
  });

  it("fails the contrast gate for a pale outline", () => {
    const pale = parsePetPalette({ schema_version: 1, swatches: [{ palette_id: "pet_palette.fur_00", fur: "#ffffff", outline: "#dddddd" }] });
    expect(contrastRatio(pale.get("pet_palette.fur_00")!.outline, UI_THEMES.era_2000.color.bg)).toBeLessThan(3);
  });
});
