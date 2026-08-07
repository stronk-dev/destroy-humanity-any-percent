import { describe, expect, it } from "vitest";
import { humanContentLocked, parseSoulCatalog, soulBand } from "../src/soul/catalog";
import invalidMaxExact from "../../testdata/soul/catalog-invalid-max-exact-v1.json";

const keys = new Set(["category.low_percent", "category.ethical_percent", "category.hundred_percent", "category.any_percent", "category.valuation"]);
const declarations = { copyKeys: keys, epochSeeded: false, catchupCeilingMs: 5000 } as const;
const fixture = () => ({ schema_version: 1, policy: { soul_floor: 0, soul_initial: 100, soul_max: 100, recovery_beat_ceiling_ms: 5000, max_session_wall_ms: 86400000 }, bands: [
  { band_member: "near_zero", min_inclusive: 0, max_inclusive: 9, human_content_locked: true, reason_key: "category.low_percent" },
  { band_member: "hollow", min_inclusive: 10, max_inclusive: 39, human_content_locked: false, reason_key: "category.ethical_percent" },
  { band_member: "dimming", min_inclusive: 40, max_inclusive: 74, human_content_locked: false, reason_key: "category.hundred_percent" },
  { band_member: "whole", min_inclusive: 75, max_inclusive: 100, human_content_locked: false, reason_key: "category.any_percent" },
], debit_sources: [{ source_id: "soul.fixture", owner_kind: "fixture", amount: 20, may_exhaust: true, single_use: true, curtain_copy_key: "category.valuation" }],
recovery_activities: [{ activity_id: "touch_grass.fixture", duration_attended_ms: 5000, recovery_amount: 15, reason_key: "category.any_percent" }],
ending_policy: { whole_variant: "earnest_ascension", depleted_variant: "training_data" } });

describe("Soul catalog", () => {
  it("matches every inclusive band boundary", () => {
    const catalog = parseSoulCatalog(fixture(), declarations);
    expect([0, 9, 10, 39, 40, 74, 75, 100].map((value) => soulBand(catalog, value).band_member)).toEqual(["near_zero", "near_zero", "hollow", "hollow", "dimming", "dimming", "whole", "whole"]);
    expect(humanContentLocked(catalog, 0)).toBe(true);
    expect(humanContentLocked(catalog, 10)).toBe(false);
  });
  it("fails closed on missing rows, copy drift, and epoch fixture content", () => {
    const missing = fixture(); delete (missing.policy as Partial<typeof missing.policy>).soul_initial;
    expect(() => parseSoulCatalog(missing, declarations)).toThrow();
    const unknown = fixture(); unknown.bands[0]!.reason_key = "soul.unknown";
    expect(() => parseSoulCatalog(unknown, declarations)).toThrow();
    expect(() => parseSoulCatalog(fixture(), { ...declarations, epochSeeded: true })).toThrow();
  });
  it("rejects the shared MaxExactInteger nonterminal-band mutation", () => {
    expect(() => parseSoulCatalog(invalidMaxExact, declarations)).toThrow();
  });
  it("rejects a heartbeat ceiling above the global catch-up ceiling", () => {
    const invalid = fixture(); (invalid.policy as { recovery_beat_ceiling_ms: number }).recovery_beat_ceiling_ms = 5001;
    expect(() => parseSoulCatalog(invalid, declarations)).toThrow();
  });
});
