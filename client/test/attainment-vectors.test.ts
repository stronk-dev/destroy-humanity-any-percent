import vectorsJSON from "../../testdata/axis-stack/attainment-vectors-v1.json";
import fixtureJSON from "../../balance/testdata/axis-stack/economy-v5-fixture.json";
import achievementsBytes from "../../balance/achievements/first-content.json?raw";
import { describe, expect, it } from "vitest";

import { loadAchievementCatalog } from "../src/achievements/catalog";
import { parseCatalog } from "../src/economy-kernel";
import { achievementProofSatisfied, attainRun, foundationAchievementRegistry, type ReplayEvent, type ReplayState } from "../src/replay";

interface Vector {
  name: string; attained_before: string[] | null; earned_now: string[] | null; counters: Record<string, number>; generators: Record<string, number> | null;
  event_kinds: string[] | null; cash_debit: string | null; attained_after: string[]; score_after: number; reattained_order: string[];
}
const vectors = (vectorsJSON as unknown as { vectors: Vector[] }).vectors;
const economy = parseCatalog(fixtureJSON);
const achievements = loadAchievementCatalog(achievementsBytes, foundationAchievementRegistry(economy));

describe("Clout v1 attainment second pass (CV2, AC5) replays the Go-authored vectors", () => {
  it("covers the corpus", () => expect(vectors.length).toBeGreaterThanOrEqual(9));
  for (const vector of vectors) {
    it(vector.name, () => {
      const attained = new Set(vector.attained_before ?? []);
      const state = { achievementsEarnedRun: new Set(vector.earned_now ?? []), achievementsAttainedRun: attained,
        attainmentScoreRun: [...attained].reduce((sum, id) => sum + achievements.byId.get(id)!.scoreGrant, 0) } as unknown as ReplayState;
      const events: ReplayEvent[] = (vector.event_kinds ?? []).map((kind) => ({ kind, schema_version: 1, intent_id: "x", payload: {} }) as ReplayEvent);
      const proofEvents = [...events];
      const debits: Record<string, string> = vector.cash_debit === null ? {} : { "company.cash": vector.cash_debit };
      attainRun(achievements, state, { facts: new Set(), counters: vector.counters, exitCount: 0, generators: vector.generators ?? {} },
        new Set(vector.earned_now ?? []), (definition) => achievementProofSatisfied(definition, debits, proofEvents), { company_stream_id: "s", run_seq: 2 }, "i", events);
      expect([...state.achievementsAttainedRun!].sort()).toEqual(vector.attained_after);
      expect(state.attainmentScoreRun).toBe(vector.score_after);
      expect(events.slice(proofEvents.length).map((value) => (value.payload as { achievement_id: string }).achievement_id)).toEqual(vector.reattained_order);
      expect(events.slice(proofEvents.length).every((value) => value.kind === "achievement_reattained.v1")).toBe(true);
    });
  }
});
