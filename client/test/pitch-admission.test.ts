import { beforeAll, describe, expect, it } from "vitest";
import fixture from "../../balance/testdata/pitch-v1.json";
import { pitchContentHash } from "../src/pitch/catalog";
import { applyPitch, createPitch, type PitchSnapshot } from "../src/pitch/engine";

const content = JSON.stringify(fixture);
const common = { content, content_schema_version: 1, seed: 2n, mode: "solo", scaling_inputs: { "minigame.pitch": 1 } } as const;
let identity: typeof common & { content_hash: string };
let playing: PitchSnapshot;
let shop: PitchSnapshot;

beforeAll(async () => {
  identity = { ...common, content_hash: await pitchContentHash(content) };
  playing = await createPitch(identity);
  shop = (await applyPitch({ ...identity, revision: 1, snapshot: playing,
    command: { kind: "play_hand", card_ids: ["demo_day#1", "patch_release#1", "testimonial#2", "uptime_nines#2"] } })).snapshot;
  expect(shop.phase).toBe("shop");
  expect(shop.shop_offers.length).toBeGreaterThan(0);
});

const variants = ["missing", "null", "wrong_case", "case_collision"] as const;
type Variant = typeof variants[number];
function mutate(source: object, key: string, variant: Variant): Record<string, unknown> {
  const fields = structuredClone(source) as Record<string, unknown>;
  if (!Object.hasOwn(fields, key)) throw new Error(`mutation has no canonical ${key}`);
  if (variant === "missing") delete fields[key];
  else if (variant === "null") fields[key] = null;
  else {
    fields[key.toUpperCase()] = fields[key];
    if (variant === "wrong_case") delete fields[key];
  }
  return fields;
}

describe("Pitch exact tenant JSON admission", () => {
  for (const key of ["phase", "round", "hands_remaining", "deck_count", "hand", "slotted_hacks",
    "run_currency", "shop_offers", "funding_target", "round_best_valuation", "revision", "pitch_content_hash", "pitch_schema_version"]) {
    for (const variant of variants) {
      it(`refuses snapshot/${key}/${variant}`, async () => {
        const snapshot = mutate(playing, key, variant) as unknown as PitchSnapshot;
        await expect(applyPitch({ ...identity, revision: 1, snapshot, command: { kind: "play_hand", card_ids: [] } })).rejects.toThrow();
      });
    }
  }
  for (const key of ["hack_id", "offer_id", "price"]) {
    for (const variant of variants) {
      it(`refuses offer/${key}/${variant}`, async () => {
        const snapshot = { ...shop, shop_offers: [mutate(shop.shop_offers[0]!, key, variant), ...shop.shop_offers.slice(1)] } as unknown as PitchSnapshot;
        await expect(applyPitch({ ...identity, revision: 2, snapshot, command: { kind: "end_shop" } })).rejects.toThrow();
      });
    }
  }
  for (const [kind, keys] of [["play_hand", ["kind", "card_ids"]], ["buy_hack", ["kind", "offer_id"]], ["end_shop", ["kind"]]] as const) {
    for (const key of keys) {
      for (const variant of variants) {
        it(`refuses command/${kind}/${key}/${variant}`, async () => {
          const command = kind === "play_hand" ? { kind, card_ids: [] } : kind === "buy_hack" ? { kind, offer_id: shop.shop_offers[0]!.offer_id } : { kind };
          const snapshot = kind === "play_hand" ? playing : shop;
          await expect(applyPitch({ ...identity, revision: snapshot.revision, snapshot,
            command: mutate(command, key, variant) })).rejects.toThrow();
        });
      }
    }
    it(`admits canonical command/${kind}`, async () => {
      const command = kind === "play_hand" ? { kind, card_ids: [] } : kind === "buy_hack" ? { kind, offer_id: shop.shop_offers[0]!.offer_id } : { kind };
      const snapshot = kind === "play_hand" ? playing : shop;
      const output = await applyPitch({ ...identity, revision: snapshot.revision, snapshot, command });
      expect(output.snapshot.revision).toBe(snapshot.revision + 1);
    });
  }
  for (const char of ["A", "z", "/"]) {
    it(`refuses noncanonical content hash/${char}`, async () => {
      await expect(applyPitch({ ...identity, revision: 1,
        snapshot: { ...playing, pitch_content_hash: `sha256:${char.repeat(64)}` },
        command: { kind: "play_hand", card_ids: [] } })).rejects.toThrow();
    });
  }
  it("preserves escaped canonical keys and explicitly present zero", async () => {
    const snapshot = JSON.parse(JSON.stringify(playing).replace('"phase"', '"\\u0070hase"')) as PitchSnapshot;
    const escaped = await applyPitch({ ...identity, revision: 1, snapshot, command: { kind: "play_hand", card_ids: [] } });
    const plain = await applyPitch({ ...identity, revision: 1, snapshot: playing, command: { kind: "play_hand", card_ids: [] } });
    expect(escaped).toEqual(plain);
    const zero = await applyPitch({ ...identity, revision: 1, snapshot: { ...playing, run_currency: 0 }, command: { kind: "play_hand", card_ids: [] } });
    expect(zero.snapshot.run_currency).toBe(0);
    expect(zero.snapshot.revision).toBe(2);
  });
});
