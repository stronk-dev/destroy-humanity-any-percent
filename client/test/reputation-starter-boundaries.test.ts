import { describe, expect, it } from "vitest";

import corpus from "../../testdata/replay/reputation-tree-v1.json";
import effects from "../../testdata/reputation/starter-effects-v1.json";
import boundaries from "../../testdata/reputation/starter-boundaries-v1.json";
import { applyLoggedExit, canonicalJSONString, encodeReplayState, loadReplayCatalogBundle, restoreReplayState, withNextReplayCatalogBundle, type ReplayArtifacts, type ReplayCatalogBundle } from "../src/replay";
import { constantsHashArtifacts } from "./pet-fixture-bundle";

type Boundary = typeof boundaries.cases[number];

function artifactsFor(profile: string): ReplayArtifacts {
  const artifacts = { ...corpus.exit.artifacts } as ReplayArtifacts;
  const tree = JSON.parse(artifacts.reputation_tree!);
  if (profile === "retire") tree.nodes = tree.nodes.filter((node: { node_id: string }) => !["reputation.starter.generated_beige_tower", "reputation.starter.upgrade_continuous_feed_paper"].includes(node.node_id));
  else if (profile === "idempotent") {
    const content = JSON.parse(artifacts.curriculum!);
    content.first_failure.branches[2].starter_package.upgrade_id = "upgrade.continuous_feed_paper";
    return { ...artifacts, curriculum: JSON.stringify(content), reputation_tree: JSON.stringify(tree) };
  } else if (profile === "generated_cap") tree.nodes.find((node: { node_id: string }) => node.node_id === "reputation.starter.generated_beige_tower").starter.count = Number.MAX_SAFE_INTEGER - 10;
  else if (profile === "resource_cap") tree.nodes.find((node: { node_id: string }) => node.node_id === "reputation.starter.cash_small").starter = { kind: "resource_grant", resource_id: "company.permits", amount: "2.4e1" };
  else throw new Error(`unknown boundary profile ${profile}`);
  return { ...artifacts, reputation_tree: JSON.stringify(tree) };
}

async function setup(row: Boundary) {
  const current = await loadReplayCatalogBundle(corpus.exit.constants_hash, corpus.exit.artifacts as unknown as ReplayArtifacts);
  const artifacts = artifactsFor(row.profile);
  const next = await loadReplayCatalogBundle(await constantsHashArtifacts(artifacts), artifacts);
  expect(next.constantsHash).not.toBe(current.constantsHash);
  const inputs = structuredClone(corpus.exit.case.replay_inputs);
  const carry = inputs.resolved.founder_carry;
  carry.reputation_level = 552;
  carry.founder_extensions.reputation_spent = 552;
  carry.founder_extensions.reputation_unlock_ppm = 1_000_000;
  carry.founder_extensions.reputation_nodes_owned = [...effects.cases[0]!.owned];
  inputs.resolved.next_constants_hash = next.constantsHash;
  inputs.resolved.selected_branch = row.branch;
  const state = restoreReplayState(corpus.exit.case.pre_state, 18, current.economy, { meters: current.meters!, achievements: current.achievements!, doctrines: current.doctrines, opportunities: current.opportunities });
  state.balances["company.cash"] = row.pre_cash;
  return { current, next, inputs, state };
}

describe("Reputation starter next-bundle boundaries", () => {
  it("pins three legal profiles and the separately forbidden retirement", () => {
    expect(boundaries.version).toBe(2);
    expect(boundaries.cases.map((row) => row.profile)).toEqual(["idempotent", "generated_cap", "resource_cap"]);
    expect(boundaries.forbidden_transitions).toEqual([{ profile: "retire", removed_node_ids: ["reputation.starter.generated_beige_tower", "reputation.starter.upgrade_continuous_feed_paper"] }]);
  });

  it.each(boundaries.forbidden_transitions)("refuses $profile instead of erasing its population", async (row) => {
    const current = await loadReplayCatalogBundle(corpus.exit.constants_hash, corpus.exit.artifacts as unknown as ReplayArtifacts);
    const artifacts = artifactsFor(row.profile);
    const next = await loadReplayCatalogBundle(await constantsHashArtifacts(artifacts), artifacts);
    expect(current.reputationTree!.nodes.filter((node) => !next.reputationTree!.nodes.some((candidate) => candidate.node_id === node.node_id)).map((node) => node.node_id)).toEqual(row.removed_node_ids);
    expect(() => withNextReplayCatalogBundle(current, next)).toThrow(/Reputation node removed/u);
    const state = restoreReplayState(corpus.exit.case.pre_state, 18, current.economy, { meters: current.meters!, achievements: current.achievements!, doctrines: current.doctrines, opportunities: current.opportunities });
    const before = canonicalJSONString(encodeReplayState(state));
    const inputs = structuredClone(corpus.exit.case.replay_inputs);
    inputs.resolved.next_constants_hash = next.constantsHash;
    await expect(applyLoggedExit(state, canonicalJSONString(corpus.exit.case.canonical_payload), { ...current, next }, inputs)).rejects.toThrow(/Reputation node removed/u);
    expect(canonicalJSONString(encodeReplayState(state))).toBe(before);
  });

  it.each(boundaries.cases)("applies $profile on the actual next bundle", async (row) => {
    const { current, next, inputs, state } = await setup(row);
    const transition = await applyLoggedExit(state, canonicalJSONString(corpus.exit.case.canonical_payload), withNextReplayCatalogBundle(current, next), inputs);
    expect(transition.outcome).toBe("applied");
    expect(transition.newCompany!.balances["company.cash"]).toBe(row.cash);
    expect(transition.newCompany!.balances["company.permits"]).toBe(row.permits);
    expect(transition.newCompany!.generatorsProvisioned["generator.beige_tower"]).toBe(row.provisioned);
    expect(transition.newCompany!.generators["generator.beige_tower"]).toBe(0);
    expect(transition.newCompany!.generatorPurchasedTotal).toBe(0);
    expect([...transition.newCompany!.upgradesOwned]).toEqual(row.upgrades);
    expect(transition.companyStartedEvents[0]).toMatchObject({ schema_version: 2, payload: { reputation_tree: { bonus_factor: "6.52e0", applied_starter_node_ids: row.applied } } });
    expect(transition.founder.reputation_level).toBe(552);
    expect(transition.founder.founder_extensions!.reputation_spent).toBe(552);
    expect(transition.founder.founder_extensions!.reputation_nodes_owned).toEqual(effects.cases[0]!.owned);
  });

  it.each(["generated_cap", "resource_cap"])("refuses the over-cap raw %s artifact at rule7", async (profile) => {
    const artifacts = artifactsFor(profile);
    const tree = JSON.parse(artifacts.reputation_tree!);
    if (profile === "generated_cap") tree.nodes.find((node: { node_id: string }) => node.node_id === "reputation.starter.generated_beige_tower").starter.count++;
    else tree.nodes.find((node: { node_id: string }) => node.node_id === "reputation.starter.cash_small").starter.amount = "2.5e1";
    const invalid = { ...artifacts, reputation_tree: JSON.stringify(tree) };
    await expect(loadReplayCatalogBundle(await constantsHashArtifacts(invalid), invalid)).rejects.toThrow(/rule 7/u);
  });

  it.each(["generated_cap", "resource_cap"])("defensively refuses a parsed %s grant fault without clamping", async (profile) => {
    const row = boundaries.cases.find((candidate) => candidate.profile === profile)!;
    const { current, next, inputs, state } = await setup(row);
    // Deliberate internal fault after strict loading: NOT an admitted artifact,
    // a production pin or persistence proof. Original parsed objects stay frozen.
    const tree = next.reputationTree!;
    const nodes = tree.nodes.map((node) => {
      if (node.kind !== "starter") return node;
      if (profile === "generated_cap" && node.starter.kind === "generated_generators") return { ...node, starter: { ...node.starter, count: node.starter.count + 1 } };
      if (profile === "resource_cap" && node.node_id === "reputation.starter.cash_small" && node.starter.kind === "resource_grant") return { ...node, starter: { ...node.starter, amount: "2.5e1" } };
      return node;
    });
    const fault: ReplayCatalogBundle = { ...next, reputationTree: { ...tree, nodes } };
    await expect(applyLoggedExit(state, canonicalJSONString(corpus.exit.case.canonical_payload), withNextReplayCatalogBundle(current, fault), inputs)).rejects.toThrow(profile === "generated_cap" ? "Reputation starter exceeds the provisioned hardcap" : "above_hardcap");
    expect(Object.isFrozen(next.reputationTree)).toBe(true);
  });
});
