import { describe, expect, it } from "vitest";

import { loadReplayCatalogBundle, type ReplayArtifacts } from "../src/replay";
import { constantsHashArtifacts, cosmeticsArtifacts, liveArtifacts, petSpeciesArtifacts } from "./cosmetic-fixture-bundle";

const load = async (artifacts: ReplayArtifacts) => loadReplayCatalogBundle(await constantsHashArtifacts(artifacts), artifacts);

describe("cosmetics bundle wiring (Cosmetic Shop v1 C2, OD-10)", () => {
  it("loads cosmetics on the pet_species chain and joins constants identity", async () => {
    const bundle = await load(cosmeticsArtifacts);
    expect(bundle.cosmetics?.items.map((item) => item.cosmetic_id)).toEqual(["horse_armor"]);
    expect(await constantsHashArtifacts(cosmeticsArtifacts)).not.toBe(await constantsHashArtifacts(petSpeciesArtifacts));
    expect((await load(petSpeciesArtifacts)).cosmetics).toBeUndefined();
  });

  it("rejects cosmetics without pet_species, and a priced artifact", async () => {
    const { cosmetics } = cosmeticsArtifacts;
    await expect(load({ ...liveArtifacts, cosmetics })).rejects.toThrow(/artifact set/u);
    const priced = cosmetics!.replace("\"slot\": \"pet\",", "\"slot\": \"pet\", \"price\": 0,");
    expect(priced).not.toBe(cosmetics);
    await expect(load({ ...cosmeticsArtifacts, cosmetics: priced })).rejects.toThrow();
  });
});
