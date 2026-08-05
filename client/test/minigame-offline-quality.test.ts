import { describe, expect, it } from "vitest";
import fixture from "../../testdata/minigame/offline-quality-v1.json";
import { offlineGradeForScore, parseOfflineQualityPolicy, parseOfflineQualityState } from "../src/minigame/offline-quality";

const declarations = {
  score_fact_ids: new Set(fixture.declarations.score_fact_ids),
  automation_destinations: new Set(fixture.declarations.automation_destinations),
};

describe("minigame offline quality", () => {
  it("matches the shared C34 score thresholds", () => {
    const policy = parseOfflineQualityPolicy(fixture.policy, declarations);
    for (const row of fixture.grade_cases) expect(offlineGradeForScore(policy, row.score)).toBe(row.grade_ppm);
    expect(parseOfflineQualityState(fixture.state)).toEqual(fixture.state);
  });

  it("rejects ambiguous policy and state rows", () => {
    const base = structuredClone(fixture.policy) as Record<string, any>;
    const cases = [
      { ...base, score_fact: "score.other" },
      { ...base, grade_curve: [] },
      { ...base, grade_curve: [base.grade_curve[1], base.grade_curve[0]] },
      { ...base, grade_curve: [{ ...base.grade_curve[0], grade_ppm: 500000 }] },
      { ...base, grade_curve: [{ ...base.grade_curve[0], label: "low" }] },
      { ...base, timezone: "utc" },
    ];
    for (const value of cases) expect(() => parseOfflineQualityPolicy(value, declarations)).toThrow();
    expect(() => parseOfflineQualityState({ ...fixture.state, decay_remainder_ppm: 1_000_000 })).toThrow();
  });
});
