import { describe, expect, it } from "vitest";

import { parseFeaturesPresentation, FEATURES_PRESENTATION } from "../src/game-ui/features-presentation";
import source from "../src/game-ui/features-presentation.json";
import { fiscalPhase } from "../src/game-ui/fiscal-phase";

describe("GS1 Fiscal display phase", () => {
  it("switches exactly at the pinned window edges", () => {
    const period = { early_ms: 100, guaranteed_ms: 200 };
    expect([99, 100, 199, 200].map((elapsed) => fiscalPhase(elapsed, period))).toEqual(["ripening", "early", "early", "guaranteed"]);
  });
});

describe("features presentation", () => {
  it("binds every live meter and the Pitch unlock, and withholds unlock.arcade", () => {
    expect(FEATURES_PRESENTATION.trustMeters.size).toBe(10);
    expect(FEATURES_PRESENTATION.fiscalUnlocks.has("minigame.pitch")).toBe(true);
    expect(FEATURES_PRESENTATION.fiscalUnlocks.has("unlock.arcade")).toBe(false);
  });
  it("fails closed on unknown copy and unsorted rows", () => {
    expect(() => parseFeaturesPresentation({ ...source, meter_bands: [{ band_id: "high", title_key: "meters.band.nope" }] })).toThrow(/unknown copy/);
    expect(() => parseFeaturesPresentation({ ...source, meter_bands: [...source.meter_bands].reverse() })).toThrow(/sorted/);
  });
});
