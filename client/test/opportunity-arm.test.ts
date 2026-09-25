import { describe, expect, it } from "vitest";

import { parseFeatures } from "../src/game-ui/contracts";

const base = { achievements: null, active_play: null, fiscal: null, meters: null, minigames: null, pets: null };
const pending = { effect_row_id: "active.building", expires_attended_ms: 5_000, opportunity_id: "01986666-0000-7000-8000-000000000001", selected_generator_id: "generator.beige_tower" };
const buff = { buff_instance_id: "01986666-0000-7000-8000-00000000000b", effect_row_id: "active.production", expires_attended_ms: 4_000, selected_target: null };
const arm = { attended_now_ms: 2_000, buffs: [buff], combo: { cap: "1e4", reason_key: "cap.active_combo", saturated: false }, pending };

describe("opportunity snapshot arm (GS5)", () => {
  it("accepts the exact arm, absent or null", () => {
    expect(() => parseFeatures({ ...base, opportunity: arm })).not.toThrow();
    expect(() => parseFeatures({ ...base, opportunity: { ...arm, pending: null, buffs: [] } })).not.toThrow();
    expect(() => parseFeatures({ ...base, opportunity: null })).not.toThrow();
    expect(() => parseFeatures(base)).not.toThrow();
  });

  it("fails closed on expired rows, unsorted buffs, extra keys, and a non-null active_play", () => {
    expect(() => parseFeatures({ ...base, opportunity: { ...arm, pending: { ...pending, expires_attended_ms: 2_000 } } })).toThrow(/expired opportunity/);
    expect(() => parseFeatures({ ...base, opportunity: { ...arm, buffs: [{ ...buff, expires_attended_ms: 2_000 }] } })).toThrow(/expired buff/);
    const later = { ...buff, buff_instance_id: "01986666-0000-7000-8000-00000000000c" };
    expect(() => parseFeatures({ ...base, opportunity: { ...arm, buffs: [later, buff] } })).toThrow(/sorted/);
    expect(() => parseFeatures({ ...base, opportunity: { ...arm, combo: { ...arm.combo, cap: "0" } } })).toThrow();
    expect(() => parseFeatures({ ...base, opportunity: { ...arm, extra: 1 } })).toThrow(/exact/);
    expect(() => parseFeatures({ ...base, active_play: arm })).toThrow(/must be null/);
  });
});
