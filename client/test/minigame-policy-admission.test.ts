import { describe, expect, it } from "vitest";
import fixture from "../../testdata/minigame/catalog-v2.json";
import { parseMinigameCatalog } from "../src/minigame/catalog";

describe("minigame catalog required policy values", () => {
  for (const [section, keys] of [["payout", ["sends_per_day", "per_send_cap", "conversion_ppm"]],
    ["rating_policy", ["starting_elo", "elo_floor", "provisional_games"]], ["offline_quality", ["decay_ppm_per_grid"]]] as const) {
    for (const key of keys) it(`refuses ${section}/${key}`, () => {
      const source = structuredClone(fixture);
      Object.assign(source.minigames[0]![section], { [key]: null });
      expect(() => parseMinigameCatalog(source)).toThrow();
    });
  }
  for (const key of ["score_threshold", "grade_ppm"] as const) {
    for (const variant of ["alias", "null", "missing"] as const) it(`refuses curve/${key}/${variant}`, () => {
      const source = structuredClone(fixture), quality = source.minigames[0]!.offline_quality;
      quality.neutral_floor_ppm = quality.grade_curve[0]!.grade_ppm = 0;
      const row = quality.grade_curve[0]! as Record<string, unknown>;
      if (variant === "alias") { row[key.toUpperCase()] = row[key]; delete row[key]; }
      else if (variant === "missing") delete row[key]; else row[key] = null;
      expect(() => parseMinigameCatalog(source)).toThrow();
    });
  }
  for (const [kind, idKey, refKey] of [["bot", "policy_id", "bot_ref"], ["npc_partner", "profile_id", "npc_profile"]] as const) {
    for (const variant of ["id alias", "version alias", "null reduction"] as const) it(`refuses ${kind}/${variant}`, () => {
      const source = structuredClone(fixture), identity: Record<string, unknown> = { [idKey]: "fixture.policy", version: "1.0.0" };
      const fallback: Record<string, unknown> = { kind, [refKey]: identity, rate_reduction_ppm: 0 };
      if (variant === "id alias") { identity[idKey.toUpperCase()] = identity[idKey]; delete identity[idKey]; }
      else if (variant === "version alias") { identity.VERSION = identity.version; delete identity.version; }
      else fallback.rate_reduction_ppm = null;
      Object.assign(source.minigames[0]!, { fallback });
      expect(() => parseMinigameCatalog(source)).toThrow();
    });
  }
  it("preserves explicitly present legal zeros", () => {
    const source = structuredClone(fixture), row = source.minigames[0]!;
    row.payout.sends_per_day = row.payout.per_send_cap = row.payout.conversion_ppm = 0;
    row.rating_policy.starting_elo = row.rating_policy.elo_floor = row.rating_policy.provisional_games = 0;
    row.offline_quality.decay_ppm_per_grid = row.offline_quality.neutral_floor_ppm = row.offline_quality.grade_curve[0]!.grade_ppm = 0;
    expect(() => parseMinigameCatalog(source)).not.toThrow();
  });
});
