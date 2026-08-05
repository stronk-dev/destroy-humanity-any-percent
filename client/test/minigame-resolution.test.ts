import { describe, expect, it } from "vitest";
import catalogFixture from "../../testdata/minigame/catalog-v2.json";
import fixture from "../../testdata/minigame/resolution-v1.json";
import { parseMinigameCatalog } from "../src/minigame/catalog";
import { applyFounderMinigameResolution } from "../src/minigame/resolution";

describe("minigame resolution arithmetic", () => {
  const definition = parseMinigameCatalog(catalogFixture).minigames[0]!;
  for (const row of fixture.cases) it(row.name, () => {
    expect(applyFounderMinigameResolution(row.rating, row.quality, row.result, definition.rating_policy,
      definition.offline_quality, row.founder_attended_ms)).toEqual({
      rating_before: row.rating, rating_after: row.expected_rating,
      quality_before: row.quality, quality_after: row.expected_quality,
    });
  });
});
