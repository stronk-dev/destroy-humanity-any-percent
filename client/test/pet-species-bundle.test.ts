import { describe, expect, it } from "vitest";

import { loadReplayCatalogBundle, type ReplayArtifacts } from "../src/replay";
import { constantsHashArtifacts, liveArtifacts, petSpeciesArtifacts, reputationArtifacts } from "./pet-fixture-bundle";

const load = async (artifacts: ReplayArtifacts) => loadReplayCatalogBundle(await constantsHashArtifacts(artifacts), artifacts);

describe("pet_species bundle wiring (P3)", () => {
  it("loads pet_species on the Reputation chain and joins constants identity", async () => {
    const bundle = await load(petSpeciesArtifacts);
    expect(bundle.petSpecies?.max_pets_per_founder).toBe(1);
    expect(await constantsHashArtifacts(petSpeciesArtifacts)).not.toBe(await constantsHashArtifacts(reputationArtifacts));
    expect((await load(reputationArtifacts)).petSpecies).toBeUndefined();
  });

  it("rejects pet_species without reputation_tree or pets, and an invalid artifact", async () => {
    const { pet_species: species } = petSpeciesArtifacts;
    await expect(load({ ...liveArtifacts, pet_species: species })).rejects.toThrow(/artifact set/u);
    const { pets: _pets, ...withoutPets } = petSpeciesArtifacts;
    await expect(load(withoutPets as ReplayArtifacts)).rejects.toThrow();
    await expect(load({ ...petSpeciesArtifacts, pet_species: `{"schema_version":1,"max_pets_per_founder":0,"species":[]}` })).rejects.toThrow();
  });
});
