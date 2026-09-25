import { describe, expect, it } from "vitest";

import corpus from "../../testdata/replay/reputation-tree-v1.json";
import { applyFounderLogged, applyLoggedExit, canonicalJSONString, encodeFounderReplayState, encodeReplayState, loadReplayCatalogBundle, restoreFounderReplayState, restoreReplayState, withNextReplayCatalogBundle, type ReplayArtifacts, type ReplayCatalogBundle } from "../src/replay";

// AC4: the TS Founder replay byte-matches the Go-authored purchase corpus
// (testdata/replay/reputation-tree-v1.json, regenerated only from Go).
const bundles = new Map<string, Promise<ReplayCatalogBundle>>();
function bundle(name: string): Promise<ReplayCatalogBundle> {
  const source = (corpus.bundles as Record<string, { constants_hash: string; artifacts: Record<string, string> }>)[name]!;
  if (!bundles.has(name)) bundles.set(name, loadReplayCatalogBundle(source.constants_hash, source.artifacts as unknown as ReplayArtifacts));
  return bundles.get(name)!;
}

describe("Reputation purchase cross-runtime corpus", () => {
  it("pins every R5 rejection row and the full chain", () => {
    expect(corpus.version).toBe(1);
    expect(corpus.cases.map((row) => row.name)).toContain("rejects-inactive-tree");
    expect(corpus.cases.filter((row) => row.name.startsWith("chain-"))).toHaveLength(9);
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

  it("rejects tampered resolved inputs instead of trusting them", async () => {
    const applied = corpus.cases.find((row) => row.name === "applies-first-unlock")!;
    const catalogs = await bundle(applied.bundle);
    for (const patch of [{ resolved_cost: 2 }, { reputation_level: 4 }, { owned_before: ["reputation.unlock.p05"] }, { reputation_spent_before: 1 }]) {
      const inputs = structuredClone(applied.replay_inputs) as { resolved: Record<string, unknown> };
      Object.assign(inputs.resolved, patch);
      const state = restoreFounderReplayState(applied.pre_state, applied.state_version, catalogs);
      await expect(applyFounderLogged(state, canonicalJSONString(applied.canonical_payload), catalogs, inputs), JSON.stringify(patch)).rejects.toThrow();
    }
  });
});

describe("Reputation starters at new-run assembly", () => {
  it("replays the burnout Exit with owned starters and run_started v2 byte-identically (AC8)", async () => {
    const exit = corpus.exit;
    const current = await loadReplayCatalogBundle(exit.constants_hash, exit.artifacts as unknown as ReplayArtifacts);
    const next = await loadReplayCatalogBundle(exit.next_constants_hash, exit.next_artifacts as unknown as ReplayArtifacts);
    const bundle = withNextReplayCatalogBundle(current, next);
    const testCase = exit.case;
    const state = restoreReplayState(testCase.pre_state, 18, bundle.economy, { meters: bundle.meters!, achievements: bundle.achievements!, doctrines: bundle.doctrines, opportunities: bundle.opportunities });
    const transition = await applyLoggedExit(state, canonicalJSONString(testCase.canonical_payload), bundle, testCase.replay_inputs);
    expect(transition.outcome).toBe("applied");
    expect(canonicalJSONString(transition.receipt)).toBe(testCase.receipt_json);
    expect(canonicalJSONString(transition.founder)).toBe(testCase.founder_output_json);
    expect(canonicalJSONString(encodeReplayState(transition.newCompany!))).toBe(testCase.new_company_json);
    expect(canonicalJSONString(transition.companyStartedEvents)).toBe(testCase.company_started_events_json);
    expect(transition.newCompany!.generatorsProvisioned["generator.beige_tower"]).toBe(15);
    expect(transition.newCompany!.generators["generator.beige_tower"]).toBe(0);
    expect(transition.companyStartedEvents[0]).toMatchObject({ kind: "run_started", schema_version: 2 });
  });
});
