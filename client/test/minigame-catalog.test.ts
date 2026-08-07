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
});
