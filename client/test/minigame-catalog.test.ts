import { describe, expect, it } from "vitest";
import fixture from "../../testdata/minigame/catalog-v2.json";
import pitchFixture from "../../testdata/minigame/pitch-v3.json";
import { parseMinigameCatalog } from "../src/minigame/catalog";

describe("pinned minigame catalog", () => {
  it("loads the same complete policy row as Go", () => {
    const catalog = parseMinigameCatalog(fixture);
    expect(catalog.minigameIds).toEqual(["fixture.counter"]);
    expect(catalog.minigames[0]?.payout.payout_score_fact_id).toBe("score.total");
    expect(catalog.minigames[0]?.rating_policy.elo_ceiling).toBe(3000);
    expect(catalog.minigames[0]?.offline_quality.neutral_floor_ppm).toBe(500_000);
  });

  it("rejects missing policy bytes and noncanonical ordering", () => {
    expect(() => parseMinigameCatalog({ schema_version: 2, rating_seasons: [], minigames: [{ minigame_id: "partial" }] })).toThrow();
    expect(() => parseMinigameCatalog({ ...fixture, rating_seasons: ["ranked", "preseason"] })).toThrow();
  });

  it("loads the closed Fiscal unlock arm for The Pitch", () => {
    const definition = parseMinigameCatalog(pitchFixture).minigames[0]!;
    expect(definition.unlock_condition).toEqual({ kind: "fiscal_unlock", unlock_id: "minigame.pitch" });
    expect(definition.payout.credited_resource_id).toBe("company.cash");
  });

  it("loads the tier_at_least arm and rejects out-of-domain rows like Go (TT-PA2)", () => {
    const withUnlock = (unlock: unknown) => ({ ...pitchFixture, minigames: [{ ...pitchFixture.minigames[0]!, unlock_condition: unlock }] });
    for (const accepted of [{ kind: "tier_at_least", tier: 1 }, { exit_history_at_least: 1, kind: "tier_at_least", tier: 1 }, { kind: "tier_at_least", tier: 9 }]) {
      expect(parseMinigameCatalog(withUnlock(accepted)).minigames[0]!.unlock_condition).toEqual(accepted);
    }
    for (const rejected of [{ kind: "tier_at_least" }, { kind: "tier_at_least", tier: 10 }, { kind: "tier_at_least", tier: -1 }, { kind: "tier_at_least", tier: 1.5 },
      { extra: 1, kind: "tier_at_least", tier: 1 }, { exit_history_at_least: -1, kind: "tier_at_least", tier: 1 }, { kind: "tier_at_least", tier: "1" }]) {
      expect(() => parseMinigameCatalog(withUnlock(rejected)), JSON.stringify(rejected)).toThrow();
    }
  });
});
