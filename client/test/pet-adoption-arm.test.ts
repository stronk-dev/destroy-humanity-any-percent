import { describe, expect, it } from "vitest";

import { parseFeatures } from "../src/game-ui/contracts";

const base = { achievements: null, active_play: null, fiscal: null, meters: null, minigames: null, pets: null };
const row = { eligible_action_ids: ["care.feed"], name_key: "pet.name.server_room_cat.n04", palette_id: "pet_palette.fur_02",
  pet_id: "01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa", species_id: "pet_species.server_room_cat", status_band: "high", temperament: "sassy" };
const arm = { pet_adoption: { cap: 1, count: 1, name_keys: ["pet.name.server_room_cat.n04"], starter_species_id: "pet_species.server_room_cat" }, pets: [row] };

describe("pet_adoption snapshot arm (AC11)", () => {
  it("accepts the exact arm, absent or null", () => {
    expect(() => parseFeatures({ ...base, pet_adoption: arm })).not.toThrow();
    expect(() => parseFeatures({ ...base, pet_adoption: null })).not.toThrow();
    expect(() => parseFeatures(base)).not.toThrow();
  });

  it("rejects any raw care internal and inconsistent counts", () => {
    for (const leak of ["stats_ppm", "trust_ppm", "mood", "behavior_queue"]) {
      expect(() => parseFeatures({ ...base, pet_adoption: { ...arm, pets: [{ ...row, [leak]: 1 }] } }), leak).toThrow();
    }
    expect(() => parseFeatures({ ...base, pet_adoption: { ...arm, pets: [] } })).toThrow();
    expect(() => parseFeatures({ ...base, pet_adoption: { ...arm, pet_adoption: { ...arm.pet_adoption, count: 2 } } })).toThrow();
  });
});
