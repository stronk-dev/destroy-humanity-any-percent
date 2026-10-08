import { describe, expect, it } from "vitest";

import cosmetics from "../../testdata/replay/cosmetic-v1.json";
import reputation from "../../testdata/replay/reputation-tree-v1.json";
import { applyFounderLogged, applyLoggedExit, canonicalJSONString, encodeFounderReplayState, encodeReplayState,
  loadReplayCatalogBundle, restoreFounderReplayState, restoreReplayState, withNextReplayCatalogBundle,
  type ReplayArtifacts, type ReplayCatalogBundle } from "../src/replay";
import { constantsHashArtifacts, petSpeciesArtifacts } from "./pet-fixture-bundle";

function companyState(source: unknown, catalogs: ReplayCatalogBundle) {
  return restoreReplayState(source, 18, catalogs.economy, { meters: catalogs.meters!, achievements: catalogs.achievements!,
    doctrines: catalogs.doctrines, opportunities: catalogs.opportunities });
}

describe("pet Exit boundaries (PA3.5/PA6, AC6/AC8/AC14)", () => {
  it.each([false, true])("uses the recorded next-run pin; species activation=%s", async (activate) => {
    const row = reputation.exit_cases.find((entry) => entry.name === "plan-applies-with-in-plan-prerequisite")!;
    const source = row.company;
    expect(row.founder!.state_version).toBe(22);
    expect(source.constants_hash).toBe(source.next_constants_hash);
    const current = await loadReplayCatalogBundle(source.constants_hash, source.artifacts as unknown as ReplayArtifacts);
    const speciesHash = await constantsHashArtifacts(petSpeciesArtifacts);
    const species = await loadReplayCatalogBundle(speciesHash, petSpeciesArtifacts);
    // A newer bundle is available in BOTH arms; availability must not activate it.
    const joined = withNextReplayCatalogBundle(current, species);
    const hash = activate ? speciesHash : source.constants_hash;
    const version = activate ? 23 : 22;
    const inputs = structuredClone(source.case.replay_inputs);
    inputs.resolved.next_constants_hash = hash;
    const transition = await applyLoggedExit(companyState(source.case.pre_state, joined), canonicalJSONString(source.case.canonical_payload), joined, inputs);
    expect(transition.outcome).toBe("applied");
    expect(canonicalJSONString(transition.receipt)).toBe(source.case.receipt_json);
    expect(canonicalJSONString(encodeReplayState(transition.newCompany!))).toBe(source.case.new_company_json);
    const wantCarry = JSON.parse(source.case.founder_output_json);
    const wantFounder = JSON.parse(row.founder!.post_state_json);
    if (activate) {
      wantCarry.founder_extensions.pet_identities = {};
      wantFounder.pet_identities = {};
    }
    expect(canonicalJSONString(transition.founder)).toBe(canonicalJSONString(wantCarry));

    const founderCase = row.founder!;
    const founderInputs = structuredClone(founderCase.replay_inputs);
    founderInputs.resolved.result_constants_hash = hash;
    founderInputs.resolved.result_founder_wire_version = version;
    const state = restoreFounderReplayState(founderCase.pre_state, founderCase.state_version, joined);
    const applied = await applyFounderLogged(state, canonicalJSONString(founderCase.canonical_payload), joined, founderInputs);
    expect(applied.outcome).toBe("applied");
    expect(applied.resultConstantsHash).toBe(hash);
    expect(applied.state.wireVersion).toBe(version);
    expect(canonicalJSONString(encodeFounderReplayState(applied.state))).toBe(canonicalJSONString(wantFounder));

    if (!activate) return;
    const missing = companyState(source.case.pre_state, current);
    const before = canonicalJSONString(encodeReplayState(missing));
    await expect(applyLoggedExit(missing, canonicalJSONString(source.case.canonical_payload), current, inputs)).rejects.toThrow("next catalog bundle mismatch");
    expect(canonicalJSONString(encodeReplayState(missing))).toBe(before);
    founderInputs.resolved.result_founder_wire_version = 22;
    const wrongVersion = restoreFounderReplayState(founderCase.pre_state, founderCase.state_version, joined);
    const founderBefore = canonicalJSONString(encodeFounderReplayState(wrongVersion));
    await expect(applyFounderLogged(wrongVersion, canonicalJSONString(founderCase.canonical_payload), joined, founderInputs)).rejects.toThrow();
    expect(canonicalJSONString(encodeFounderReplayState(wrongVersion))).toBe(founderBefore);
  });

  const exits = cosmetics.exit_cases.filter((row) => row.name === "exit-wind-down-preserves-owned-equipped" || row.name === "exit-accept-offer-preserves-owned-equipped");
  it("includes both nonempty pet carry paths", () => expect(exits).toHaveLength(2));
  it.each(exits)("preserves identity and care across $name, refusing an omitted carry", async (row) => {
    const source = row.company;
    const catalogs = await loadReplayCatalogBundle(source.constants_hash, source.artifacts as unknown as ReplayArtifacts);
    expect(source.next_constants_hash).toBe(source.constants_hash);
    const before = row.founder!.pre_state as Record<string, unknown>;
    expect(Object.keys(before.pet_identities as object)).toHaveLength(1);
    expect(Object.keys(before.pets as object)).toHaveLength(1);
    const company = await applyLoggedExit(companyState(source.case.pre_state, catalogs), canonicalJSONString(source.case.canonical_payload), catalogs, source.case.replay_inputs);
    const founderCase = row.founder!;
    const founder = await applyFounderLogged(restoreFounderReplayState(founderCase.pre_state, founderCase.state_version, catalogs), canonicalJSONString(founderCase.canonical_payload), catalogs, founderCase.replay_inputs);
    expect(company.outcome).toBe("applied");
    expect(founder.outcome).toBe("applied");
    const post = encodeFounderReplayState(founder.state) as Record<string, unknown>;
    for (const field of ["pets", "pet_identities"] as const) {
      expect(canonicalJSONString(post[field])).toBe(canonicalJSONString(before[field]));
      expect(canonicalJSONString(company.founder!.founder_extensions![field])).toBe(canonicalJSONString(before[field]));
    }
    const inputs = structuredClone(source.case.replay_inputs);
    const extensions = inputs.resolved.founder_carry.founder_extensions as Record<string, unknown>;
    delete extensions.pet_identities;
    const state = companyState(source.case.pre_state, catalogs);
    const encodedBefore = canonicalJSONString(encodeReplayState(state));
    await expect(applyLoggedExit(state, canonicalJSONString(source.case.canonical_payload), catalogs, inputs)).rejects.toThrow();
    expect(canonicalJSONString(encodeReplayState(state))).toBe(encodedBefore);
  });

  it.each(exits)("requires every recorded identity field across $name, not invented zeroes", async (row) => {
    const source = row.company;
    const catalogs = await loadReplayCatalogBundle(source.constants_hash, source.artifacts as unknown as ReplayArtifacts);
    for (const field of ["species_id", "temperament", "palette_id", "name_key", "adopted_at_ms", "adopted_at_attended_ms"]) {
      for (const nullValue of [false, true]) {
        const inputs = structuredClone(source.case.replay_inputs);
        const identities = inputs.resolved.founder_carry.founder_extensions.pet_identities as Record<string, Record<string, unknown>>;
        expect(Object.keys(identities)).toHaveLength(1);
        for (const identity of Object.values(identities)) {
          if (nullValue) identity[field] = null;
          else delete identity[field];
        }
        const state = companyState(source.case.pre_state, catalogs);
        const before = canonicalJSONString(encodeReplayState(state));
        await expect(applyLoggedExit(state, canonicalJSONString(source.case.canonical_payload), catalogs, inputs), `${field}/${nullValue ? "null" : "omitted"}`).rejects.toThrow();
        expect(canonicalJSONString(encodeReplayState(state))).toBe(before);
      }
    }
    // Zero is a valid explicit coordinate, not a missing-value sentinel.
    const inputs = structuredClone(source.case.replay_inputs);
    const identities = inputs.resolved.founder_carry.founder_extensions.pet_identities as Record<string, Record<string, unknown>>;
    for (const identity of Object.values(identities)) {
      identity.adopted_at_ms = 0;
      identity.adopted_at_attended_ms = 0;
    }
    const applied = await applyLoggedExit(companyState(source.case.pre_state, catalogs), canonicalJSONString(source.case.canonical_payload), catalogs, inputs);
    expect(applied.outcome).toBe("applied");
    expect(canonicalJSONString(applied.founder!.founder_extensions!.pet_identities)).toBe(canonicalJSONString(identities));
    expect(canonicalJSONString(applied.receipt)).toBe(source.case.receipt_json);
  });
});
