import { expect, it } from "vitest";
import fixture from "../../balance/testdata/pitch-v1.json";
import vector from "../../testdata/pitch/aggregation-compatibility-research-v1.json";
import { parsePitchCatalog, pitchContentHash } from "../src/pitch/catalog";
import { applyPitch, createPitch, PITCH_ENGINE_VERSION } from "../src/pitch/engine";

it("retains Pitch 1.0.0's admitted aggregation-dependent valuation and funding decision", async () => {
  expect(vector.version).toBe(1);
  expect(vector.engine_version).toBe(PITCH_ENGINE_VERSION);
  const synthetic = structuredClone(fixture);
  for (const row of vector.cards) {
    const card = synthetic.metric_cards.find((entry) => entry.card_id === row.card_id);
    if (!card) throw new Error(`unknown research card ${row.card_id}`);
    card.base_metric = row.base_metric;
  }
  const hack = synthetic.growth_hacks.find((entry) => entry.hack_id === vector.hack_id);
  if (!hack || hack.effect.kind !== "card_factor") throw new Error("invalid research card factor");
  hack.effect.factor = vector.card_factor;
  for (const [index, row] of synthetic.funding_curve.entries()) row.funding_target = `1e${index + 4}`;
  synthetic.funding_curve[0]!.funding_target = vector.funding_target;
  expect(() => parsePitchCatalog(synthetic)).not.toThrow();
  const content = JSON.stringify(synthetic);
  const identity = { content, content_hash: await pitchContentHash(content), content_schema_version: 1,
    seed: BigInt(vector.seed), mode: "solo", scaling_inputs: { "minigame.pitch": 1 } } as const;
  const genesis = await createPitch(identity);
  // This keeps the seeded hand, but hack ownership is synthetic/admitted;
  // no acquisition/reachable-player/SQL or published-catalog claim follows.
  const output = await applyPitch({ ...identity, revision: 1,
    snapshot: { ...genesis, slotted_hacks: [vector.hack_id] },
    command: { kind: "play_hand", card_ids: vector.cards.map((row) => row.instance) } });
  expect(output.snapshot.round_best_valuation).toBe(vector.current_valuation);
  expect(output.snapshot.phase).toBe(vector.current_phase);
  expect(output.result).toBeNull();
});
