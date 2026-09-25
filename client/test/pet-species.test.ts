import { describe, expect, it } from "vitest";

import fixtureArtifact from "../../balance/testdata/pet-species/fixture-v1.json";
import drawVectors from "../../testdata/pet/adoption-draw-vectors-v1.json";
import fixtures from "../../testdata/pet/species-fixtures-v1.json";
import { applicationCopyCatalog, COPY_KEYS } from "../src/copy";
import { drawPetAdoption, parsePetSpeciesCatalog, petSpeciesRow, PET_TEMPERAMENT_ORDER } from "../src/pet/species";

const declarations = {
  copyKeys: new Set<string>(COPY_KEYS),
  companionKeys: new Set(applicationCopyCatalog.entries.filter((entry) => entry.tone === "companion").map((entry) => entry.key)),
};

describe("pet_species catalog (AC1)", () => {
  it("agrees with the shared Go fixture corpus", () => {
    expect(fixtures.cases.length).toBeGreaterThanOrEqual(17);
    for (const testCase of fixtures.cases) {
      let valid = true;
      try { parsePetSpeciesCatalog(testCase.artifact, declarations); } catch { valid = false; }
      expect(valid, testCase.name).toBe(testCase.valid);
    }
  });

  it("loads the fixture artifact", () => {
    const catalog = parsePetSpeciesCatalog(fixtureArtifact, declarations);
    const row = petSpeciesRow(catalog, "pet_species.server_room_cat")!;
    expect([catalog.max_pets_per_founder, row.name_keys.length, row.palette_ids.length]).toEqual([1, 12, 10]);
    expect(row.allowed_temperaments).toEqual(PET_TEMPERAMENT_ORDER);
  });
});

describe("adoption draws (AC3)", () => {
  it("matches the independent draw vectors, and intent_id never changes a draw", async () => {
    const temperaments = new Set<string>(), palettes = new Set<string>();
    for (const vector of drawVectors.vectors) {
      const draws = await drawPetAdoption(vector.founder_id, vector.nonce, vector.server_time_ms, drawVectors.row as never);
      expect(draws).toEqual({ pet_id: vector.pet_id, temperament: vector.temperament, palette_id: vector.palette_id });
      temperaments.add(draws.temperament); palettes.add(draws.palette_id);
    }
    expect([temperaments.size, palettes.size]).toEqual([6, 10]);
    const restricted = drawVectors.restricted_row;
    const draws = await drawPetAdoption(restricted.vector.founder_id, restricted.vector.nonce, restricted.vector.server_time_ms, restricted as never);
    expect(draws).toEqual({ pet_id: restricted.vector.pet_id, temperament: restricted.vector.temperament, palette_id: restricted.vector.palette_id });
  });

  it("rejects malformed nonces and founder ids", async () => {
    const row = { allowed_temperaments: ["lazy"] as const, palette_ids: ["pet_palette.fur_00"] };
    await expect(drawPetAdoption("01986666-1111-7111-8111-111111111111", "0123456789ABCDEF0123456789ABCDEF", 1, row)).rejects.toThrow();
    await expect(drawPetAdoption("01986666-1111-7111-8111-11111111111", "0123456789abcdef0123456789abcdef", 1, row)).rejects.toThrow();
  });
});
