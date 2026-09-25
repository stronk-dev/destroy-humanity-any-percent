import { describe, expect, it } from "vitest";

import corpus from "../../testdata/replay/cosmetic-v1.json";
import { applyFounderLogged, applyLoggedExit, canonicalJSONString, encodeFounderReplayState, encodeReplayState, loadReplayCatalogBundle, restoreFounderReplayState, restoreReplayState, withNextReplayCatalogBundle, type ReplayArtifacts, type ReplayCatalogBundle } from "../src/replay";

// Cosmetic Shop v1 AC5/AC6 (and C3's Exit-activation witness): the TS Founder
// replay byte-matches the Go-authored corpus testdata/replay/cosmetic-v1.json.
const bundles = new Map<string, Promise<ReplayCatalogBundle>>();
function bundle(name: string): Promise<ReplayCatalogBundle> {
  const source = (corpus.bundles as Record<string, { constants_hash: string; artifacts: Record<string, string> }>)[name]!;
  if (!bundles.has(name)) bundles.set(name, loadReplayCatalogBundle(source.constants_hash, source.artifacts as unknown as ReplayArtifacts));
  return bundles.get(name)!;
}

describe("cosmetic intents cross-runtime corpus", () => {
  it("pins every §4 row", () => {
    const names = corpus.cases.map((row) => row.name);
    for (const required of ["rejects-inactive-below-v24", "rejects-price-field", "rejects-locked-at-tier-0", "applies-acquire-at-tier-1", "rejects-second-acquire",
      "rejects-equip-unknown-pet", "applies-equip", "rejects-already-equipped", "applies-unequip", "rejects-unequip-nothing-equipped", "pair-rejects-equip-not-owned", "pair-replaces"]) {
      expect(names, required).toContain(required);
    }
  });

  it.each(corpus.cases)("replays $name to the Go receipt, events, and state", async (testCase) => {
    const catalogs = await bundle(testCase.bundle);
    const state = restoreFounderReplayState(testCase.pre_state, testCase.state_version, catalogs);
    const transition = await applyFounderLogged(state, canonicalJSONString(testCase.canonical_payload), catalogs, testCase.replay_inputs);
    expect(transition.outcome).toBe(testCase.outcome);
    expect(canonicalJSONString(transition.receipt)).toBe(testCase.receipt_json);
    expect(canonicalJSONString(transition.events)).toBe(testCase.events_json);
    expect(canonicalJSONString(encodeFounderReplayState(transition.state))).toBe(testCase.post_state_json);
  });

  it("refuses a recorded acquisition whose frozen tier was lowered", async () => {
    const applied = corpus.cases.find((row) => row.name === "applies-acquire-at-tier-1")!;
    const catalogs = await bundle(applied.bundle);
    const inputs = structuredClone(applied.replay_inputs) as { resolved: { active_company: { tier: number } } };
    inputs.resolved.active_company.tier = 0;
    const state = restoreFounderReplayState(applied.pre_state, applied.state_version, catalogs);
    const transition = await applyFounderLogged(state, canonicalJSONString(applied.canonical_payload), catalogs, inputs);
    expect(canonicalJSONString(transition.receipt)).not.toBe(applied.receipt_json);
  });
});

describe("Founder v24 activation at Exit (AC4)", () => {
  it.each(corpus.exit_cases)("replays $name on the Company log and its Founder arm", async (exitCase) => {
    const company = exitCase.company;
    const current = await loadReplayCatalogBundle(company.constants_hash, company.artifacts as unknown as ReplayArtifacts);
    const next = await loadReplayCatalogBundle(company.next_constants_hash, company.next_artifacts as unknown as ReplayArtifacts);
    const joined = withNextReplayCatalogBundle(current, next);
    const testCase = company.case;
    const state = restoreReplayState(testCase.pre_state, 18, joined.economy, { meters: joined.meters!, achievements: joined.achievements!, doctrines: joined.doctrines, opportunities: joined.opportunities });
    const transition = await applyLoggedExit(state, canonicalJSONString(testCase.canonical_payload), joined, testCase.replay_inputs);
    expect(transition.outcome).toBe("applied");
    expect(canonicalJSONString(transition.receipt)).toBe(testCase.receipt_json);
    expect(canonicalJSONString(transition.founder)).toBe(testCase.founder_output_json);
    expect(canonicalJSONString(encodeReplayState(transition.newCompany!))).toBe(testCase.new_company_json);
    const founderCase = exitCase.founder!;
    const founderState = restoreFounderReplayState(founderCase.pre_state, founderCase.state_version, joined);
    const founderTransition = await applyFounderLogged(founderState, canonicalJSONString(founderCase.canonical_payload), joined, founderCase.replay_inputs);
    expect(founderTransition.outcome).toBe("applied");
    expect(canonicalJSONString(founderTransition.receipt)).toBe(founderCase.receipt_json);
    const post = encodeFounderReplayState(founderTransition.state) as Record<string, unknown>;
    expect(canonicalJSONString(post)).toBe(founderCase.post_state_json);
    expect(post.cosmetics).toEqual({ owned: [], equipped: {} });
  });
});
