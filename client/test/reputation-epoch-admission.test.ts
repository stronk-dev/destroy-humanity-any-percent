import { describe, expect, it } from "vitest";

import population from "../../testdata/reputation/epoch-transitions-v1.json";
import founderCorpus from "../../testdata/replay/reputation-earlier-activation-v1.json";
import companyCorpus from "../../testdata/replay/reputation-tree-v1.json";
import { applyFounderLogged, applyLoggedExit, canonicalJSONString, encodeFounderReplayState, encodeReplayState, loadReplayCatalogBundle, restoreFounderReplayState, restoreReplayState, withNextReplayCatalogBundle, type ReplayArtifacts } from "../src/replay";
import { constantsHashArtifacts } from "./pet-fixture-bundle";

type FixtureNode = { node_id: string; kind: string; requires: string[]; unlock_ppm?: number; [key: string]: unknown };

// Repair only fixture dependencies/terminal ladder; the candidate must load
// independently so rejection specifically concerns losing a previous ID.
function candidateArtifacts(source: ReplayArtifacts, removedID: string, appendNode = false): ReplayArtifacts {
  const tree = JSON.parse(source.reputation_tree!) as { nodes: FixtureNode[] };
  tree.nodes = tree.nodes.filter((node) => node.node_id !== removedID);
  let previous = "";
  let lastUnlock: FixtureNode | undefined;
  for (const node of tree.nodes) {
    node.requires = node.requires.filter((id) => id !== removedID);
    if (node.kind === "bonus_unlock") {
      if (previous && !node.requires.includes(previous)) node.requires.push(previous);
      previous = node.node_id;
      lastUnlock = node;
    }
    node.requires.sort();
  }
  lastUnlock!.unlock_ppm = 1_000_000;
  if (appendNode) tree.nodes.push({ ...structuredClone(tree.nodes.find((node) => node.node_id === "reputation.starter.cash_small")!), node_id: "reputation.starter.appended_cash" });
  return { ...source, reputation_tree: JSON.stringify(tree) };
}

const source = founderCorpus.bundles["target-v22-epoch8"];
const inactive = founderCorpus.bundles["source-v21"];
const founderFixture = founderCorpus.cases.find((row) => row.case.state_version === 21)!.case;

async function requireRemovalRefusal(operation: () => Promise<unknown>): Promise<void> {
  let refusal: unknown;
  try { await operation(); } catch (error) { refusal = error; }
  expect(refusal).toBeInstanceOf(RangeError);
  expect(String(refusal)).toMatch(/Reputation node removed/u);
}

describe("OD-7 append-only Reputation epoch admission", () => {
  it("names every source ID and all five legal transition controls", async () => {
    expect(Object.keys(population).sort()).toEqual(["removed_node_ids", "valid_profiles", "version"]);
    expect(population.version).toBe(1);
    expect(population.removed_node_ids).toHaveLength(9);
    expect(population.valid_profiles).toEqual(["unchanged", "retune", "append", "activation", "inactive"]);
    for (const entry of [source, companyCorpus.exit]) {
      const current = await loadReplayCatalogBundle(entry.constants_hash, entry.artifacts as unknown as ReplayArtifacts);
      expect(population.removed_node_ids).toEqual(current.reputationTree!.nodes.map((node) => node.node_id));
    }
  });

  it.each(population.removed_node_ids)("refuses $0 at linked loading", async (id) => {
    const current = await loadReplayCatalogBundle(source.constants_hash, source.artifacts as unknown as ReplayArtifacts);
    const artifacts = candidateArtifacts(current.artifacts, id);
    const next = await loadReplayCatalogBundle(await constantsHashArtifacts(artifacts), artifacts);
    expect(next.reputationTree!.nodes).toHaveLength(8);
    expect(next.reputationTree!.nodes.some((node) => node.node_id === id)).toBe(false);
    expect(() => withNextReplayCatalogBundle(current, next)).toThrow(/Reputation node removed/u);
  });

  it.each(population.removed_node_ids)("refuses $0 in direct-linked Founder replay", async (id) => {
    const current = await loadReplayCatalogBundle(source.constants_hash, source.artifacts as unknown as ReplayArtifacts);
    const artifacts = candidateArtifacts(current.artifacts, id);
    const next = await loadReplayCatalogBundle(await constantsHashArtifacts(artifacts), artifacts);
    const state = restoreFounderReplayState({ ...structuredClone(founderFixture.post_state), reputation_spent: 4, reputation_unlock_ppm: 0,
      reputation_nodes_owned: ["reputation.retired.unknown"] }, 22, current);
    const before = canonicalJSONString(encodeFounderReplayState(state));
    const inputs = structuredClone(founderFixture.replay_inputs);
    inputs.resolved.run_seq = 2;
    inputs.resolved.exit_record.run_id = 2;
    inputs.resolved.result_constants_hash = next.constantsHash;
    // Deliberately bypass convenience linking: the replay entry must defend itself.
    await requireRemovalRefusal(() => applyFounderLogged(state, canonicalJSONString(founderFixture.canonical_payload), { ...current, next }, inputs));
    expect(canonicalJSONString(encodeFounderReplayState(state))).toBe(before);
  });

  it.each(population.removed_node_ids)("refuses $0 in direct-linked Company replay", async (id) => {
    const fixture = companyCorpus.exit;
    const current = await loadReplayCatalogBundle(fixture.constants_hash, fixture.artifacts as unknown as ReplayArtifacts);
    const artifacts = candidateArtifacts(fixture.next_artifacts as unknown as ReplayArtifacts, id);
    const next = await loadReplayCatalogBundle(await constantsHashArtifacts(artifacts), artifacts);
    const inputs = structuredClone(fixture.case.replay_inputs);
    const carry = inputs.resolved.founder_carry;
    carry.reputation_level = 11;
    carry.founder_extensions.reputation_spent = 4;
    carry.founder_extensions.reputation_unlock_ppm = 0;
    carry.founder_extensions.reputation_nodes_owned = ["reputation.retired.unknown"];
    inputs.resolved.next_constants_hash = next.constantsHash;
    const state = restoreReplayState(fixture.case.pre_state, 18, current.economy, { meters: current.meters!, achievements: current.achievements!, doctrines: current.doctrines, opportunities: current.opportunities });
    const before = canonicalJSONString(encodeReplayState(state));
    await requireRemovalRefusal(() => applyLoggedExit(state, canonicalJSONString(fixture.case.canonical_payload), { ...current, next }, inputs));
    expect(canonicalJSONString(encodeReplayState(state))).toBe(before);
  });

  it("refuses whole-tree withdrawal at linked loading", async () => {
    const current = await loadReplayCatalogBundle(source.constants_hash, source.artifacts as unknown as ReplayArtifacts);
    const next = await loadReplayCatalogBundle(inactive.constants_hash, inactive.artifacts as unknown as ReplayArtifacts);
    expect(() => withNextReplayCatalogBundle(current, next)).toThrow(/Reputation tree removed/u);
  });

  it.each(population.valid_profiles)("preserves legal $0 linking and standalone history", async (profile) => {
    let current = await loadReplayCatalogBundle(source.constants_hash, source.artifacts as unknown as ReplayArtifacts);
    let next = current;
    if (profile === "retune") {
      const tree = JSON.parse(current.artifacts.reputation_tree!);
      const node = tree.nodes.find((candidate: FixtureNode) => candidate.node_id === "reputation.unlock.p05");
      node.unlock_ppm = 60_000;
      node.cost = 3;
      const artifacts = { ...current.artifacts, reputation_tree: JSON.stringify(tree) };
      next = await loadReplayCatalogBundle(await constantsHashArtifacts(artifacts), artifacts);
    } else if (profile === "append") {
      const artifacts = candidateArtifacts(current.artifacts, "", true);
      next = await loadReplayCatalogBundle(await constantsHashArtifacts(artifacts), artifacts);
    } else if (profile === "activation") current = await loadReplayCatalogBundle(inactive.constants_hash, inactive.artifacts as unknown as ReplayArtifacts);
    else if (profile === "inactive") { current = await loadReplayCatalogBundle(inactive.constants_hash, inactive.artifacts as unknown as ReplayArtifacts); next = current; }
    else expect(profile).toBe("unchanged");
    expect(withNextReplayCatalogBundle(current, next).next).toBe(next);
  });
});
