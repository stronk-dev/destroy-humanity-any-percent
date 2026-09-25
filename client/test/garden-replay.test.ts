import { describe, expect, it } from "vitest";

import corpus from "../../testdata/replay/garden-v1.json";
import { applyFounderLogged, applyLoggedExit, canonicalJSONString, encodeFounderReplayState, encodeReplayState, loadReplayCatalogBundle, restoreFounderReplayState, restoreReplayState, withNextReplayCatalogBundle, type ReplayArtifacts, type ReplayCatalogBundle } from "../src/replay";

// Server Garden AC7/AC9 (Founder half) and the v25 Exit-activation witness:
// the TS Founder replay byte-matches the Go-authored testdata/replay/garden-v1.json.
const bundles = new Map<string, Promise<ReplayCatalogBundle>>();
function bundle(name: string): Promise<ReplayCatalogBundle> {
  const source = (corpus.bundles as Record<string, { constants_hash: string; artifacts: Record<string, string> }>)[name]!;
  if (!bundles.has(name)) bundles.set(name, loadReplayCatalogBundle(source.constants_hash, source.artifacts as unknown as ReplayArtifacts));
  return bundles.get(name)!;
}

describe("garden intents cross-runtime corpus", () => {
  it("pins every SG5 detail and the advance pre-step", () => {
    const names = corpus.cases.map((row) => row.name);
    for (const required of ["rejects-inactive-below-v25", "rejects-client-tick-seq", "rejects-locked", "buys-the-unlock", "plants-and-anchors", "rejects-occupied",
      "rejects-dormant", "rejects-unknown-species", "rejects-uncollected-seed", "rejects-uproot-empty", "rejects-unknown-substrate", "rejects-unchanged-substrate",
      "switches-substrate-after-growth", "rejects-substrate-lockout", "buys-a-host-level", "uproots", "catches-up-past-the-cap"]) {
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

  it("refuses a recorded advance that no longer matches the recomputation", async () => {
    const applied = corpus.cases.find((row) => row.name === "switches-substrate-after-growth")!;
    const catalogs = await bundle(applied.bundle);
    const inputs = structuredClone(applied.replay_inputs) as { resolved: { advance: { ticks_applied: number } } };
    inputs.resolved.advance.ticks_applied = 1;
    const state = restoreFounderReplayState(applied.pre_state, applied.state_version, catalogs);
    await expect(applyFounderLogged(state, canonicalJSONString(applied.canonical_payload), catalogs, inputs)).rejects.toThrow(/advance/u);
  });
});

describe("Founder v25 activation at Exit (SG2)", () => {
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
    expect(post.server_garden).toEqual({ salt_hex: null, tick_anchor_wall_ms: null, tick_seq: 0, substrate_id: "bare_metal", substrate_set_wall_ms: null, plots: [], seed_collection: ["strain_a", "strain_b"] });
  });
});
