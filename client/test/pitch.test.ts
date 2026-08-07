import { describe, expect, it } from "vitest";
import fixture from "../../balance/testdata/pitch-v1.json";
import { parsePitchCatalog, pitchContentHash } from "../src/pitch/catalog";
import { applyPitch, createPitch } from "../src/pitch/engine";

const content = JSON.stringify(fixture);
const scaling = { "minigame.pitch": 1 } as const;

describe("The Pitch content and engine", () => {
  it("loads the exact launch catalog and rejects missing nested keys", () => {
    const catalog = parsePitchCatalog(fixture);
    expect(catalog.metric_cards).toHaveLength(12);
    expect(catalog.growth_hacks).toHaveLength(8);
    expect(catalog.growth_hacks.at(-1)?.effect).toEqual({ kind: "chain_factor", partner_hack_id: "ab_test", factor: "2e0" });
    const missing = structuredClone(fixture) as Record<string, any>;
    delete missing.growth_hacks[0].effect.factor;
    expect(() => parsePitchCatalog(missing)).toThrow();
  });

  it("deals deterministically and terminates after three empty hands", async () => {
    const contentHash = await pitchContentHash(content);
    const identity = { content, content_hash: contentHash, content_schema_version: 1 } as const;
    const first = await createPitch({ ...identity, seed: 7n, mode: "solo", scaling_inputs: scaling });
    const second = await createPitch({ ...identity, seed: 7n, mode: "solo", scaling_inputs: scaling });
    expect(second).toEqual(first);
    let snapshot = first;
    for (let revision = 1; revision <= 3; revision++) {
      const output = await applyPitch({ ...identity, seed: 7n, mode: "solo", scaling_inputs: scaling,
        revision, snapshot, command: { kind: "play_hand", card_ids: [] } });
      snapshot = output.snapshot;
      if (revision < 3) expect(output.result).toBeNull();
      else expect(output.result).toEqual({ outcome: "funding_failed", rating_delta: null,
        score_facts: [{ kind: "pitch.best_hand_exponent", value: 0 }, { kind: "pitch.final_round", value: 1 }] });
    }
  });

  it("rejects content identity drift and duplicate selections", async () => {
    const contentHash = await pitchContentHash(content);
    await expect(createPitch({ content, content_hash: `sha256:${"0".repeat(64)}`, content_schema_version: 1,
      seed: 1n, mode: "solo", scaling_inputs: scaling })).rejects.toThrow("identity");
    const snapshot = await createPitch({ content, content_hash: contentHash, content_schema_version: 1,
      seed: 1n, mode: "solo", scaling_inputs: scaling });
    const card = snapshot.hand[0]!;
    await expect(applyPitch({ content, content_hash: contentHash, content_schema_version: 1,
      seed: 1n, mode: "solo", scaling_inputs: scaling, revision: 1, snapshot,
      command: { kind: "play_hand", card_ids: [card, card] } })).rejects.toMatchObject({ code: "duplicate_card" });
  });

  it("rejects undeclared copy and snapshot state across the runtime boundary", async () => {
    const undeclared = structuredClone(fixture);
    undeclared.metric_cards[0]!.card_id = "unpublished_metric";
    undeclared.metric_cards[0]!.copy_key = "pitch.card.unpublished_metric";
    const undeclaredContent = JSON.stringify(undeclared);
    await expect(createPitch({ content: undeclaredContent, content_hash: await pitchContentHash(undeclaredContent),
      content_schema_version: 1, seed: 1n, mode: "solo", scaling_inputs: scaling })).rejects.toThrow();

    const contentHash = await pitchContentHash(content);
    const snapshot = await createPitch({ content, content_hash: contentHash, content_schema_version: 1,
      seed: 1n, mode: "solo", scaling_inputs: scaling });
    await expect(applyPitch({ content, content_hash: contentHash, content_schema_version: 1,
      seed: 1n, mode: "solo", scaling_inputs: scaling, revision: 1,
      snapshot: { ...snapshot, round_best_valuation: "-1e0" }, command: { kind: "play_hand", card_ids: [] } })).rejects.toThrow("Decimal domain");
    await expect(applyPitch({ content, content_hash: contentHash, content_schema_version: 1,
      seed: 1n, mode: "solo", scaling_inputs: scaling, revision: 1,
      snapshot: { ...snapshot, slotted_hacks: ["unpublished_hack"] }, command: { kind: "play_hand", card_ids: [] } })).rejects.toThrow("unknown Pitch hack");
  });
});
