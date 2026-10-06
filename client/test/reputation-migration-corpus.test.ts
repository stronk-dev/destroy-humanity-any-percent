import { describe, expect, it } from "vitest";

import migrations from "../../testdata/save-migrations.json";
import baseline from "../../testdata/save-migrations-baseline.json";
import source from "../../testdata/replay/reputation-tree-v1.json";
import sourceRaw from "../../testdata/replay/reputation-tree-v1.json?raw";
import { applyFounderLogged, applyLoggedExit, canonicalJSONString, encodeFounderReplayState, encodeReplayState, loadReplayCatalogBundle, restoreFounderReplayState, restoreReplayState, withNextReplayCatalogBundle, type FounderReplayState, type ReplayArtifacts } from "../src/replay";

describe("R7 shared Founder migration corpus", () => {
  it("requires the complete census, ratchet and byte-pinned source", async () => {
    expect(migrations.corpus_version).toBe(9);
    expect(migrations.cases).toHaveLength(11);
    expect(migrations.founder_cases).toHaveLength(4);
    expect(baseline.minimum_case_count).toBe(15);
    expect(baseline.schema_version).toBe(1);
    expect([...migrations.cases, ...migrations.founder_cases].map((row) => row.name).sort()).toEqual([...baseline.required_case_names].sort());
    expect(migrations.founder_cases.map((row) => row.name).sort()).toEqual(["founder-v21-nonzero-unlock", "founder-v21-to-v22", "founder-v22-spent-over-level", "founder-v22-unlock-mismatch"]);
    expect(migrations.founder_source.path).toBe("replay/reputation-tree-v1.json");
    const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", new TextEncoder().encode(sourceRaw)));
    expect([...digest].map((byte) => byte.toString(16).padStart(2, "0")).join("")).toBe(migrations.founder_source.sha256);
  });

  it.each(migrations.founder_cases)("executes $name at its declared admission/activation boundary", async (row) => {
    expect(Object.keys(row).sort()).toEqual(["error_stage", "expected_reputation", "expected_version", "from_version", "input_patch", "name", "operation", "scope", "source_case"]);
    expect(row.scope).toBe("founder");
    if (row.operation === "decode") expect(row.expected_version).toBe(row.from_version);
    expect(Object.keys(row.input_patch).every((key) => ["reputation_level", "reputation_spent", "reputation_nodes_owned", "reputation_unlock_ppm"].includes(key))).toBe(true);
    const matches = source.exit_cases.filter((entry) => entry.name === row.source_case);
    expect(matches).toHaveLength(1);
    const selected = matches[0]!;
    const founderCase = selected.founder;
    expect(founderCase).not.toBeNull();
    const company = selected.company;
    const current = await loadReplayCatalogBundle(company.constants_hash, company.artifacts as unknown as ReplayArtifacts);
    const next = await loadReplayCatalogBundle(company.next_constants_hash, company.next_artifacts as unknown as ReplayArtifacts);
    const bundle = withNextReplayCatalogBundle(current, next);
    const patched = { ...structuredClone(founderCase!.pre_state), ...row.input_patch };
    const inputBundle = row.from_version === 22 ? next : current;
    const restore = () => restoreFounderReplayState(patched, row.from_version, inputBundle);
    if (row.error_stage !== "none") {
      expect(row.operation).toBe("decode");
      expect(row.expected_reputation).toBeNull();
      if (row.error_stage === "pinned_mirror") expect(restore).toThrow(/does not mirror/u);
      else if (row.from_version === 21) expect(restore).toThrow(/before Founder v22/u);
      else expect(restore).toThrow(/invalid reputation state/u);
      return;
    }
    expect(row.operation).toBe("new_run_activation");
    expect(row.from_version).toBe(founderCase!.state_version);
    const state = restore();
    expect(state.wireVersion).toBe(21); // Loading must not activate the tree.
    const transition = await applyFounderLogged(state, canonicalJSONString(founderCase!.canonical_payload), bundle, founderCase!.replay_inputs);
    expect(transition.outcome).toBe("applied");
    expect(transition.state.wireVersion).toBe(row.expected_version);
    expect(accounting(transition.state)).toEqual(row.expected_reputation);
    expect(canonicalJSONString(encodeFounderReplayState(transition.state))).toBe(founderCase!.post_state_json);
    expect(canonicalJSONString(transition.receipt)).toBe(founderCase!.receipt_json);

    const previousCompany = restoreReplayState(company.case.pre_state, 18, current.economy, { meters: current.meters!, achievements: current.achievements!, doctrines: current.doctrines, opportunities: current.opportunities });
    const exit = await applyLoggedExit(previousCompany, canonicalJSONString(company.case.canonical_payload), bundle, company.case.replay_inputs);
    expect(exit.outcome).toBe("applied");
    const extensions = exit.founder.founder_extensions!;
    expect({ level: exit.founder.reputation_level, spent: extensions.reputation_spent, owned: extensions.reputation_nodes_owned, unlock_ppm: extensions.reputation_unlock_ppm }).toEqual(row.expected_reputation);
    expect(canonicalJSONString(exit.founder)).toBe(company.case.founder_output_json);
    expect(canonicalJSONString(encodeReplayState(exit.newCompany!))).toBe(company.case.new_company_json);
    expect(canonicalJSONString(exit.receipt)).toBe(company.case.receipt_json);
  });
});

function accounting(state: FounderReplayState) {
  return { level: state.reputationLevel, spent: state.reputationSpent, owned: state.reputationNodesOwned, unlock_ppm: state.reputationUnlockPpm };
}
