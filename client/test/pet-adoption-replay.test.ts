import { describe, expect, it } from "vitest";

import corpus from "../../testdata/replay/pet-adoption-v1.json";
import { applyFounderLogged, canonicalJSONString, encodeFounderReplayState, loadReplayCatalogBundle, restoreFounderReplayState, type ReplayArtifacts, type ReplayCatalogBundle } from "../src/replay";

// AC4/AC7: the TS Founder replay byte-matches the Go-authored adoption corpus
// (testdata/replay/pet-adoption-v1.json, regenerated only from Go).
const bundles = new Map<string, Promise<ReplayCatalogBundle>>();
function bundle(name: string): Promise<ReplayCatalogBundle> {
  const source = (corpus.bundles as Record<string, { constants_hash: string; artifacts: Record<string, string> }>)[name]!;
  if (!bundles.has(name)) bundles.set(name, loadReplayCatalogBundle(source.constants_hash, source.artifacts as unknown as ReplayArtifacts));
  return bundles.get(name)!;
}

describe("pet adoption cross-runtime corpus", () => {
  it("pins every reachable PA4.2 row, the cap ordering, and care on the adopted pet", () => {
    expect(corpus.cases.map((row) => row.name)).toEqual(expect.arrayContaining(["rejects-inactive-adoption", "rejects-unknown-species",
      "rejects-unknown-name", "applies-starter-adoption", "rejects-second-adoption-at-cap", "rejects-unknown-species-at-cap", "care-applies-on-adopted-pet"]));
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

  it("re-derives draws from the nonce: a tampered nonce cannot reproduce the receipt", async () => {
    const applied = corpus.cases.find((row) => row.name === "applies-starter-adoption")!;
    const catalogs = await bundle(applied.bundle);
    const inputs = structuredClone(applied.replay_inputs) as { resolved: Record<string, unknown> };
    inputs.resolved.adoption_nonce = "deadbeefcafebabe0011223344556677";
    const state = restoreFounderReplayState(applied.pre_state, applied.state_version, catalogs);
    const transition = await applyFounderLogged(state, canonicalJSONString(applied.canonical_payload), catalogs, inputs);
    expect(canonicalJSONString(transition.receipt)).not.toBe(applied.receipt_json);
    const stripped = structuredClone(applied.replay_inputs) as { resolved: Record<string, unknown> };
    delete stripped.resolved.adoption_nonce;
    await expect(applyFounderLogged(restoreFounderReplayState(applied.pre_state, applied.state_version, catalogs), canonicalJSONString(applied.canonical_payload), catalogs, stripped)).rejects.toThrow();
  });
});
