import { describe, expect, it } from "vitest";
import { parseMinigameCatalog } from "../src/minigame/catalog";

describe("minigame activation catalog", () => {
  it("closes sorted minigame and season domains", () => {
    const catalog = parseMinigameCatalog({ schema_version: 1, minigame_ids: ["combat.duel"], rating_seasons: ["preseason", "ranked"] });
    expect(catalog.minigameIds).toEqual(["combat.duel"]); expect(catalog.ratingSeasons).toEqual(["preseason", "ranked"]);
    expect(() => parseMinigameCatalog({ schema_version: 1, minigame_ids: ["z", "a"], rating_seasons: [] })).toThrow();
  });
});
