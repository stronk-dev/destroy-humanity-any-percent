import { describe, expect, it } from "vitest";
import corpus from "../../testdata/replay/garden-v1.json";
import { applyFounderLogged, applyLoggedExit, canonicalJSONString, encodeFounderReplayState, loadReplayCatalogBundle, restoreFounderReplayState, restoreReplayState, type ReplayArtifacts } from "../src/replay";
import { constantsHashArtifacts, gardenArtifacts } from "./garden-fixture-bundle";
import { validateGardenTransition } from "../src/garden/catalog";
import { newGardenState, parseGardenState } from "../src/garden/engine";

const retained = { salt_hex: "0123456789abcdef", tick_anchor_wall_ms: 1000000, tick_seq: 7, substrate_id: "containerized", substrate_set_wall_ms: 999999,
  plots: [{ row: 0, col: 0, species_id: "strain_a", age_ticks: 1, matured_effect_ppm: null },
    { row: 1, col: 1, species_id: "strain_b", age_ticks: 4, matured_effect_ppm: 1234567 },
    { row: 5, col: 5, species_id: "strain_c", age_ticks: 6, matured_effect_ppm: 7654321 }], seed_collection: ["strain_a", "strain_b", "strain_c"] };
const required = ["meters", "achievements", "minigames", "pets", "fiscal", "soul", "pitch", "minigame_api", "reputation_tree", "pet_species", "cosmetics"];

describe("Garden G2/G3 actual binding and permanent Exit", () => {
  it.each(required)("refuses the transitive chain without %s", async (name) => {
    const artifacts = Object.fromEntries(Object.entries(gardenArtifacts).filter(([key]) => key !== name)) as unknown as ReplayArtifacts;
    await expect(loadReplayCatalogBundle(await constantsHashArtifacts(artifacts), artifacts)).rejects.toThrow();
  });
  it("admits complete/Garden-absent controls and rejects the other label", async () => {
    const complete = await loadReplayCatalogBundle(await constantsHashArtifacts(gardenArtifacts), gardenArtifacts);
    expect(complete.garden?.species).toHaveLength(5);
    const { server_garden: _garden, ...without } = gardenArtifacts;
    const reduced = await loadReplayCatalogBundle(await constantsHashArtifacts(without), without);
    expect(reduced.garden).toBeUndefined();
    expect(reduced.constantsHash).not.toBe(complete.constantsHash);
    await expect(loadReplayCatalogBundle(reduced.constantsHash, gardenArtifacts)).rejects.toThrow(/label mismatch/u);
  });

  it("carries literal nonempty Garden across both replay axes, not empty activation", async () => {
    const activation = corpus.exit_cases[0]!;
    const grown = await loadReplayCatalogBundle(activation.company.next_constants_hash, activation.company.next_artifacts as unknown as ReplayArtifacts);
    const row = activation.company.case;
    // Reuse the already admitted terminal's scheduling/routing evidence, but
    // enter on an existing v25 pin with independently chosen permanent bytes.
    const inputs = structuredClone(row.replay_inputs);
    inputs.resolved.founder_carry.founder_constants_hash = grown.constantsHash;
    const extensions = inputs.resolved.founder_carry.founder_extensions as unknown as Record<string, unknown>;
    extensions.server_garden = structuredClone(retained);
    const foundations = { meters: grown.meters!, achievements: grown.achievements!, doctrines: grown.doctrines, opportunities: grown.opportunities };
    const company = restoreReplayState(row.pre_state, 18, grown.economy, foundations);
    const transition = await applyLoggedExit(company, canonicalJSONString(row.canonical_payload), grown, inputs);
    expect(transition.outcome).toBe("applied");
    const carry = transition.founder as unknown as { founder_extensions: { server_garden: unknown } };
    expect(canonicalJSONString(carry.founder_extensions.server_garden)).toBe(canonicalJSONString(retained));
    expect(transition.newCompany?.runSeq).toBe(3);
    const founderRow = activation.founder!;
    const founder = restoreFounderReplayState({ ...founderRow.pre_state, server_garden: structuredClone(retained) }, 25, grown);
    const founderTransition = await applyFounderLogged(founder, canonicalJSONString(founderRow.canonical_payload), grown, founderRow.replay_inputs);
    expect(founderTransition.outcome).toBe("applied");
    expect(canonicalJSONString((encodeFounderReplayState(founderTransition.state) as Record<string, unknown>).server_garden)).toBe(canonicalJSONString(retained));
    const missing = structuredClone(inputs);
    delete (missing.resolved.founder_carry.founder_extensions as unknown as Record<string, unknown>).server_garden;
    await expect(applyLoggedExit(restoreReplayState(row.pre_state, 18, grown.economy, foundations), canonicalJSONString(row.canonical_payload), grown, missing)).rejects.toThrow();
    const oldWire = { ...inputs, v: 11 };
    await expect(applyLoggedExit(restoreReplayState(row.pre_state, 18, grown.economy, foundations), canonicalJSONString(row.canonical_payload), grown, oldWire)).rejects.toThrow();
  });

  it("records the admitted starter-retune/carry conflict without inventing grants", async () => {
    const current = await loadReplayCatalogBundle(await constantsHashArtifacts(gardenArtifacts), gardenArtifacts);
    for (const starter of [false, true]) {
      const raw = JSON.parse(gardenArtifacts.server_garden!) as { species: { starter: boolean; harvest_units: number }[] };
      raw.species[0]!.harvest_units = 6;
      if (starter) raw.species[2]!.starter = true;
      const artifacts = { ...gardenArtifacts, server_garden: JSON.stringify(raw) };
      const next = await loadReplayCatalogBundle(await constantsHashArtifacts(artifacts), artifacts);
      expect(() => validateGardenTransition(current.garden, next.garden)).not.toThrow();
      const untouched = newGardenState(current.garden!);
      if (starter) expect(() => parseGardenState(untouched, next.garden!)).toThrow(/starter/u);
      else expect(parseGardenState(untouched, next.garden!)).toEqual(untouched);
    }
  });
});
