import { describe, expect, it } from "vitest";

import corpus from "../../testdata/replay/reputation-earlier-activation-v1.json";
import {
  applyFounderLogged, canonicalJSONString, encodeFounderReplayState,
  loadReplayCatalogBundle, restoreFounderReplayState, withNextReplayCatalogBundle,
  type ReplayArtifacts,
} from "../src/replay";
import { reputationAvailable } from "../src/reputation";

describe("R7/R8 shared earlier-Founder activation corpus", () => {
  it("contains every registered source and both pinned fixture targets", () => {
    expect(Object.keys(corpus).sort()).toEqual(["bundles", "cases", "schema_version", "source_versions"]);
    expect(corpus.schema_version).toBe(1);
    expect(corpus.source_versions).toEqual([14, 16, 17, 18, 19, 20, 21]);
    expect(corpus.cases).toHaveLength(7);
    expect(corpus.cases.map((row) => row.case.state_version)).toEqual(corpus.source_versions);
    expect(corpus.cases.map((row) => row.case.name)).toEqual(corpus.source_versions.map((version) => `founder-v${version}-to-v22`));
    expect(Object.keys(corpus.bundles).sort()).toEqual([
      ...corpus.source_versions.map((version) => `source-v${version}`), "target-v22-epoch8", "target-v22-legacy",
    ].sort());
    for (const bundle of Object.values(corpus.bundles)) {
      expect(Object.keys(bundle).sort()).toEqual(["artifacts", "constants_hash"]);
    }
  });

  it.each(corpus.cases)("replays $case.name to the complete Go result", async (row) => {
    expect(Object.keys(row).sort()).toEqual(["case", "next_bundle", "source_bundle"]);
    const testCase = row.case;
    expect(Object.keys(testCase).sort()).toEqual([
      "canonical_payload", "events", "events_json", "name", "outcome", "post_state", "post_state_json",
      "pre_state", "receipt", "receipt_json", "replay_inputs", "result_constants_hash", "state_version",
    ]);
    expect(row.source_bundle).toBe(`source-v${testCase.state_version}`);
    expect(row.next_bundle).toBe(testCase.state_version === 21 ? "target-v22-epoch8" : "target-v22-legacy");
    const source = corpus.bundles[row.source_bundle as keyof typeof corpus.bundles];
    const target = corpus.bundles[row.next_bundle as keyof typeof corpus.bundles];
    const current = await loadReplayCatalogBundle(source.constants_hash, source.artifacts as unknown as ReplayArtifacts);
    const next = await loadReplayCatalogBundle(target.constants_hash, target.artifacts as unknown as ReplayArtifacts);
    const bundle = withNextReplayCatalogBundle(current, next);
    const state = restoreFounderReplayState(testCase.pre_state, testCase.state_version, current);
    expect(state.wireVersion).toBe(testCase.state_version);
    expect([state.reputationLevel, state.reputationSpent, state.reputationUnlockPpm]).toEqual([11, 0, 0]);
    expect(state.reputationNodesOwned).toEqual([]);
    expect(canonicalJSONString(encodeFounderReplayState(state))).toBe(canonicalJSONString(testCase.pre_state));

    const transition = await applyFounderLogged(state, canonicalJSONString(testCase.canonical_payload), bundle, testCase.replay_inputs);
    expect(transition.outcome).toBe("applied");
    expect(transition.outcome).toBe(testCase.outcome);
    expect(transition.resultConstantsHash).toBe(next.constantsHash);
    expect(transition.resultConstantsHash).toBe(testCase.result_constants_hash);
    expect(canonicalJSONString(transition.receipt)).toBe(testCase.receipt_json);
    expect(canonicalJSONString(transition.events)).toBe(testCase.events_json);
    const encoded = encodeFounderReplayState(transition.state);
    expect(canonicalJSONString(encoded)).toBe(testCase.post_state_json);
    // The readable and canonical representations are bound too; corrupting
    // either half must not create an unused, reassuring expectation.
    expect(canonicalJSONString(testCase.receipt)).toBe(testCase.receipt_json);
    expect(canonicalJSONString(testCase.events)).toBe(testCase.events_json);
    expect(canonicalJSONString(testCase.post_state)).toBe(testCase.post_state_json);
    const restored = restoreFounderReplayState(encoded, 22, next);
    expect(restored.wireVersion).toBe(22);
    expect([restored.reputationLevel, restored.reputationSpent, restored.reputationUnlockPpm]).toEqual([11, 0, 0]);
    expect(restored.reputationNodesOwned).toEqual([]);
    expect(reputationAvailable(restored.reputationLevel, restored.reputationSpent)).toBe(11);
    expect([restored.ageMs, restored.routeKnowledgeBalance]).toEqual([12_345, 9]);
  });
});
