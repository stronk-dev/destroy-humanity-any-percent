import pinnedCatalog from "../../balance/testdata/cosmetics/fixture-v1.json?raw";
import { describe, expect, it } from "vitest";

import { loadCosmeticCatalog } from "../src/cosmetic/catalog";
import source from "../src/game-ui/cosmetics/presentation.json";
import { COSMETIC_SHOP_PRESENTATION, parseCosmeticShopPresentation } from "../src/game-ui/cosmetics/presentation";

// Cosmetic Shop v1 AC12 (loader half) and §7.2's build-time ID equality.
describe("cosmetic presentation curtain contract", () => {
  it("binds the pinned catalog ids exactly and exports the full curtain list", () => {
    expect([...COSMETIC_SHOP_PRESENTATION.cosmetics.keys()]).toEqual(loadCosmeticCatalog(pinnedCatalog).items.map((item) => item.cosmetic_id));
    expect(COSMETIC_SHOP_PRESENTATION.curtainList.map((row) => `${row.element}:${row.pattern}:${row.key}`)).toEqual([
      "cosmetic:horse_armor:paid_cosmetic_dlc:cosmetic.horse_armor_free.disclosure",
      "cosmetic:horse_armor:reference_price_anchor:cosmetic.horse_armor_free.anchor_disclosure",
      "cosmetic_shop:checkout_flow:shop.cosmetics.checkout_disclosure",
    ]);
  });

  it("rejects every unbound, duplicated, misplaced, or unknown curtain", () => {
    const mutate = (edit: (value: typeof source) => void) => { const copy = structuredClone(source); edit(copy); return () => parseCosmeticShopPresentation(copy); };
    expect(mutate((value) => { value.cosmetics[0]!.curtains = value.cosmetics[0]!.curtains.filter((row) => row.pattern !== "paid_cosmetic_dlc"); })).toThrow(/paid_cosmetic_dlc/u);
    expect(mutate((value) => { value.cosmetics[0]!.curtains = value.cosmetics[0]!.curtains.filter((row) => row.pattern !== "reference_price_anchor"); })).toThrow(/anchor/u);
    expect(mutate((value) => { (value.cosmetics[0] as { anchor_key: string | null }).anchor_key = null; })).toThrow(/anchor/u);
    expect(mutate((value) => { value.cosmetic_shop.curtains = []; })).toThrow(/checkout_flow/u);
    expect(mutate((value) => { value.cosmetics[0]!.curtains.push({ ...value.cosmetics[0]!.curtains[0]! }); })).toThrow(/twice/u);
    expect(mutate((value) => { value.cosmetics[0]!.curtains[0]!.pattern = "fake_scarcity"; })).toThrow(/unknown pattern/u);
    expect(mutate((value) => { value.cosmetics[0]!.curtains.push({ pattern: "re_release", key: "cosmetic.horse_armor_free.disclosure" }); })).toThrow(/re_release/u);
    expect(mutate((value) => { value.cosmetics[0]!.curtains[0]!.key = "cosmetic.undeclared"; })).toThrow(/copy key/u);
    expect(mutate((value) => { (value.cosmetics[0] as Record<string, unknown>).price = 0; })).toThrow(/not exact/u);
  });
});
