import { describe, expect, it } from "vitest";
import { parseFeatures, parseGameUISnapshot } from "../src/game-ui/contracts";

const arm = {
  available: 3, bonus_factor_next_run: "1.002e0", bonus_factor_this_run: "1e0" as string | null,
  level: 4, spent: 1, per_level_ppm: 10_000, unlock_ppm: 50_000,
  nodes: [{ body_key: "reputation_tree.node.unlock_p05.body", cost: 1, kind: "bonus_unlock", node_id: "reputation.unlock.p05",
    requires: [], state: "owned", title_key: "reputation_tree.node.unlock_p05.title" }],
};
const features = { achievements: null, active_play: null, fiscal: null, meters: null, minigames: null, pets: null, reputation: arm };
const snapshot = {
  constants_hash: `sha256:${"a".repeat(64)}`, evaluated_through_ms: 1_800_000_000_000,
  facts: [{ fact_id: "feature.reputation_tree", value: true }], founder_revision: 1,
  generators: [], manual_action: { action_id: "manual.click", bucket_cap_milli: 50_000, refill_milli_per_ms: 25, refilled_at_ms: 1_800_000_000_000, tokens_milli: 50_000 },
  progress: [], resources: [], revision: 1,
  run: { category: "any_percent", exit_count: 0, founder_id: "01985555-1111-7111-8111-111111111111", run_seq: 1, run_started_at_ms: 1_799_999_000_000, tier: 0 },
  schema_version: 4, server_now_ms: 1_800_000_000_000,
  transitions: { cross_gate: null, wind_down: { eligible: false } }, upgrades: [], features,
};
const readers = {
  features: (value: unknown) => parseFeatures({ ...features, reputation: value }),
  snapshot: (value: unknown) => parseGameUISnapshot({ ...snapshot, features: { ...features, reputation: value } }),
};
const badStrings = ["", "NaN", "Infinity", "-Infinity", "1", "1e+0", "01e0", "1.0e0", "0", "0e0", "-1e0", "5e-1"];
const fields = ["bonus_factor_next_run", "bonus_factor_this_run"] as const;

describe("Reputation projection factor admission (R3/R9)", () => {
  it("pins the complete declared reader/field/value population", () => {
    expect(Object.keys(readers)).toEqual(["features", "snapshot"]);
    expect(fields).toHaveLength(2); expect(new Set(badStrings).size).toBe(12);
    expect(Object.keys(readers).length * fields.length * badStrings.length).toBe(48);
  });
  for (const [name, read] of Object.entries(readers)) {
    for (const profile of [
      { next: "1e0", current: null }, { next: "1.002e0", current: "1e0" }, { next: "6.52e0", current: "1.002e0" },
    ]) it(`${name}: preserves ${profile.next}/${profile.current}`, () => {
      const value = { ...arm, bonus_factor_next_run: profile.next, bonus_factor_this_run: profile.current };
      expect(() => read(value)).not.toThrow();
      expect(value.bonus_factor_next_run).toBe(profile.next); expect(value.bonus_factor_this_run).toBe(profile.current);
    });
    for (const field of fields) {
      for (const value of badStrings) it(`${name}: refuses ${field}=${JSON.stringify(value)}`, () => {
        expect(() => read({ ...arm, [field]: value })).toThrow();
      });
      it(`${name}: already refuses numeric ${field}`, () => { expect(() => read({ ...arm, [field]: 1 })).toThrow(); });
    }
  }
});
