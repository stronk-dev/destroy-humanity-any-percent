import { describe, expect, it } from "vitest";

import fixture from "../../testdata/gameui/cosmetics-arm-v1.json";
import { parseCosmeticsArm } from "../src/game-ui/contracts";

// Cosmetic Shop v1 AC10 (TS half): the decoder accepts every Go-projected arm
// and rejects the two named contradictions plus a price-shaped field.
describe("game UI cosmetics arm", () => {
  it("accepts every shared Go-projected case", () => {
    expect(Object.keys(fixture.cases).sort()).toEqual(["acquirable-at-tier-1", "locked-at-tier-0", "owned-no-wearer", "owned-not-worn", "owned-worn"]);
    for (const [name, arm] of Object.entries(fixture.cases)) expect(() => parseCosmeticsArm(arm), name).not.toThrow();
  });

  it("rejects owned+acquirable, a worn value absent from worn_by, and unknown fields", () => {
    const acquirable = structuredClone(fixture.cases["acquirable-at-tier-1"]);
    acquirable.items[0]!.owned = true;
    expect(() => parseCosmeticsArm(acquirable)).toThrow(/cannot be acquirable/u);
    const worn = structuredClone(fixture.cases["owned-worn"]) as { items: { worn_by: string[] }[] };
    worn.items[0]!.worn_by = [];
    expect(() => parseCosmeticsArm(worn)).toThrow(/absent from its worn_by/u);
    const priced = structuredClone(fixture.cases["acquirable-at-tier-1"]) as Record<string, unknown> & { items: Record<string, unknown>[] };
    priced.items[0]!.price = 0;
    expect(() => parseCosmeticsArm(priced)).toThrow(/not exact|keys/u);
    const inactive = { ...structuredClone(fixture.cases["owned-worn"]), active: false };
    expect(() => parseCosmeticsArm(inactive)).toThrow(/inactive/u);
  });

  function twoPetArm(): { items: { worn_by: string[] }[]; wearers: { pet_id: string; worn: string | null }[] } {
    const twoPets = structuredClone(fixture.cases["owned-worn"]) as {
      items: { worn_by: string[] }[];
      wearers: { pet_id: string; worn: string | null }[];
    };
    const secondPet = "01986666-bbbb-7bbb-8bbb-bbbbbbbbbbbb";
    twoPets.items[0]!.worn_by.push(secondPet);
    twoPets.wearers.push({ pet_id: secondPet, worn: "horse_armor" });
    expect(() => parseCosmeticsArm(twoPets)).not.toThrow();
    return twoPets;
  }

  it("rejects reversed worn_by pet ids", () => {
    const reversed = twoPetArm();
    reversed.items[0]!.worn_by.reverse();
    expect(() => parseCosmeticsArm(reversed)).toThrow(/worn_by.*sorted/u);
  });

  it("rejects duplicate worn_by pet ids", () => {
    const duplicate = twoPetArm();
    duplicate.items[0]!.worn_by.push(duplicate.wearers[1]!.pet_id);
    expect(() => parseCosmeticsArm(duplicate)).toThrow(/worn_by.*sorted/u);
  });

  it("rejects a lock tier outside the catalog's 0..8 domain", () => {
    const outOfDomain = structuredClone(fixture.cases["locked-at-tier-0"]) as { items: { lock: { tier: number } | null }[] };
    outOfDomain.items[0]!.lock!.tier = 9;
    expect(() => parseCosmeticsArm(outOfDomain)).toThrow(/cosmetic lock tier/u);
  });
});
