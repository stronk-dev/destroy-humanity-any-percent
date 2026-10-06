import { describe, expect, it } from "vitest";

import corpus from "../../testdata/replay/reputation-earlier-activation-v1.json";
import companyCorpus from "../../testdata/replay/reputation-tree-v1.json";
import { applyFounderLogged, applyLoggedExit, canonicalJSONString, encodeFounderReplayState, loadReplayCatalogBundle, restoreFounderReplayState, restoreReplayState, withNextReplayCatalogBundle, type ReplayArtifacts } from "../src/replay";
import { reputationAvailable } from "../src/reputation";
import { constantsHashArtifacts } from "./pet-fixture-bundle";

const rows = [{ name: "unchanged", unlock: 50_000 }, { name: "higher", unlock: 60_000 }, { name: "lower", unlock: 40_000 }];

describe("R1/OD-7 owned effects at a new epoch", () => {
  it.each(rows)("rebinds the $name mirror without repricing history", async (row) => {
    const source = corpus.bundles["target-v22-epoch8"];
    const current = await loadReplayCatalogBundle(source.constants_hash, source.artifacts as unknown as ReplayArtifacts);
    let next = current;
    if (row.unlock !== 50_000) {
      const tree = JSON.parse(current.artifacts.reputation_tree!) as { nodes: Array<{ node_id: string; unlock_ppm?: number; cost: number }> };
      const ownedNode = tree.nodes.find((node) => node.node_id === "reputation.unlock.p05")!;
      ownedNode.unlock_ppm = row.unlock;
      ownedNode.cost = 3;
      const artifacts = { ...current.artifacts, reputation_tree: JSON.stringify(tree) };
      next = await loadReplayCatalogBundle(await constantsHashArtifacts(artifacts), artifacts);
      expect(next.constantsHash).not.toBe(current.constantsHash);
    }
    const fixture = corpus.cases.find((entry) => entry.case.state_version === 21)!.case;
    const preState = { ...structuredClone(fixture.post_state), reputation_spent: 4, reputation_unlock_ppm: 50_000,
      reputation_nodes_owned: ["reputation.retired.unknown", "reputation.unlock.p05"] };
    const state = restoreFounderReplayState(preState, 22, current);
    expect(canonicalJSONString(encodeFounderReplayState(state))).toBe(canonicalJSONString(preState));
    const inputs = structuredClone(fixture.replay_inputs);
    inputs.resolved.run_seq = 2;
    inputs.resolved.exit_record.run_id = 2;
    inputs.resolved.result_constants_hash = next.constantsHash;
    const transition = await applyFounderLogged(state, canonicalJSONString(fixture.canonical_payload), withNextReplayCatalogBundle(current, next), inputs);
    expect(transition.outcome).toBe("applied");
    expect(transition.resultConstantsHash).toBe(next.constantsHash);
    expect([state.reputationLevel, state.reputationSpent, state.reputationUnlockPpm, state.ageMs, state.routeKnowledgeBalance]).toEqual([11, 4, row.unlock, 12_345, 9]);
    expect(state.reputationNodesOwned).toEqual(preState.reputation_nodes_owned);
    expect(reputationAvailable(state.reputationLevel, state.reputationSpent)).toBe(7);
    expect(state.exitHistory.map((exit) => exit.run_id)).toEqual([1, 2]);
    expect(restoreFounderReplayState(encodeFounderReplayState(state), 22, next).reputationUnlockPpm).toBe(row.unlock);
  });

  it.each(rows)("assembles the $name Company run from the next effect", async (row) => {
    const fixture = companyCorpus.exit;
    const current = await loadReplayCatalogBundle(fixture.constants_hash, fixture.artifacts as unknown as ReplayArtifacts);
    let artifacts = fixture.next_artifacts as unknown as ReplayArtifacts;
    if (row.unlock !== 50_000) {
      const tree = JSON.parse(artifacts.reputation_tree!) as { nodes: Array<{ node_id: string; unlock_ppm?: number; cost: number }> };
      const ownedNode = tree.nodes.find((node) => node.node_id === "reputation.unlock.p05")!;
      ownedNode.unlock_ppm = row.unlock;
      ownedNode.cost = 3;
      artifacts = { ...artifacts, reputation_tree: JSON.stringify(tree) };
    }
    const next = await loadReplayCatalogBundle(await constantsHashArtifacts(artifacts), artifacts);
    expect(next.reputationTree!.nodes.map((node) => node.node_id)).toEqual(current.reputationTree!.nodes.map((node) => node.node_id));
    const inputs = structuredClone(fixture.case.replay_inputs);
    const carry = inputs.resolved.founder_carry;
    carry.reputation_level = 11;
    carry.founder_extensions.reputation_spent = 4;
    carry.founder_extensions.reputation_unlock_ppm = 50_000;
    carry.founder_extensions.reputation_nodes_owned = ["reputation.retired.unknown", "reputation.unlock.p05"];
    inputs.resolved.next_constants_hash = next.constantsHash;
    const state = restoreReplayState(fixture.case.pre_state, 18, current.economy, { meters: current.meters!, achievements: current.achievements!, doctrines: current.doctrines, opportunities: current.opportunities });
    const result = await applyLoggedExit(state, canonicalJSONString(fixture.case.canonical_payload), withNextReplayCatalogBundle(current, next), inputs);
    expect(result.outcome).toBe("applied");
    expect(result.founder.reputation_level).toBe(11);
    const extensions = result.founder.founder_extensions!;
    expect([extensions.reputation_spent, extensions.reputation_unlock_ppm, extensions.reputation_nodes_owned]).toEqual([4, row.unlock, carry.founder_extensions.reputation_nodes_owned]);
    expect(reputationAvailable(result.founder.reputation_level, extensions.reputation_spent!)).toBe(7);
    const factor = row.unlock === 60_000 ? "1.0066e0" : row.unlock === 40_000 ? "1.0044e0" : "1.0055e0";
    expect(result.companyStartedEvents[0]).toMatchObject({ kind: "run_started", schema_version: 2, payload: { reputation_tree: { bonus_factor: factor, applied_starter_node_ids: [] } } });
    expect(result.newCompany).not.toBeNull();
  });
});
