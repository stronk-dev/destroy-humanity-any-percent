import { describe, expect, it } from "vitest";

import { newGardenState } from "../src/garden/engine";
import { initialPetCareState } from "../src/pet/identity";
import { encodeFounderReplayState, loadReplayCatalogBundle, restoreFounderReplayState, type FounderReplayState, type ReplayArtifacts, type ReplayCatalogBundle } from "../src/replay";
import { constantsHashArtifacts, cosmeticsArtifacts, gardenArtifacts } from "./garden-fixture-bundle";

const load = async (artifacts: ReplayArtifacts) => loadReplayCatalogBundle(await constantsHashArtifacts(artifacts), artifacts);

const PET = "01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa";
const identity = { species_id: "pet_species.server_room_cat", temperament: "curious" as const, palette_id: "pet_palette.fur_03",
  name_key: "pet.name.server_room_cat.n07", adopted_at_ms: 1_790_000_000_000, adopted_at_attended_ms: 5_400_000 };


function founderState(bundle: ReplayCatalogBundle, overrides: Partial<FounderReplayState>): FounderReplayState {
  const founderResources = bundle.economy.resources.filter((row) => row.scope === "founder");
  const founderGenerators = bundle.economy.generatorClasses.filter((row) => bundle.economy.resource(row.price.resourceId)?.scope === "founder");
  return {
    wireVersion: 24, balances: Object.fromEntries(founderResources.map((row) => [row.id, row.initial])),
    generators: Object.fromEntries(founderGenerators.map((row) => [row.id, 0])), generatorPurchasedTotal: 0, upgradesOwned: new Set(),
    generatorsProvisioned: Object.fromEntries(founderGenerators.map((row) => [row.id, 0])),
    provisionRemaindersPpm: Object.fromEntries(founderGenerators.filter((row) => row.provision).map((row) => [row.provision!.generatorId, 0])),
    stockRateRemainderPpm: 0, evaluatedThroughMs: 1_786_000_000_000, computeCreditMs: 0, manualTokenMilli: 0, manualTokenRefilledAtMs: 1_786_000_000_000,
    routeKnowledgeBalance: 0, hintsUnlocked: new Set(), ledgerFactKinds: new Set(), reputationLevel: 9, networkSlots: [], cloutLifetime: 0,
    soul: bundle.soul!.policy.soul_initial, ageMs: 0, notoriety: 0, advisorMode: false, exitHistory: [],
    achievementsEarnedLifetime: new Set(), achievementScoreLifetime: 0,
    minigameRatings: Object.fromEntries(bundle.minigames!.minigames.map((row) => [row.minigame_id, { elo: row.rating_policy.starting_elo, season_member: row.rating_policy.season_member, games_counted: 0 }])),
    minigameOfflineQuality: Object.fromEntries(bundle.minigames!.minigames.map((row) => [row.minigame_id, { grade_ppm: row.offline_quality.neutral_floor_ppm, last_founder_attended_ms: 0, decay_remainder_ppm: 0 }])),
    pets: { [PET]: initialPetCareState(bundle.pets!, 5_400_000) }, fiscalCredit: 0, fiscalPeriodOpenedWallMs: 1_786_000_000_000, fiscalPeriodSequence: 0,
    fiscalGeneratorLevels: Object.fromEntries(bundle.fiscal!.generatorLevelRows.map((row) => [row.generatorId, 0])), fiscalUnlocks: new Set(),
    soulExhaustedSourceIds: new Set(), minigameSessionSeq: 0, reputationUnlockPpm: 0, reputationSpent: 0, reputationNodesOwned: [],
    petIdentities: { [PET]: identity }, cosmetics: { owned: ["horse_armor"], equipped: { [PET]: "horse_armor" } }, serverGarden: null,
    ...overrides,
  };
}


describe("Founder v25 server_garden (Server Garden SG2, AC2)", () => {
  it("pins Founder v25 only on the full chain against the pinned Fiscal rows", async () => {
    const bundle = await load(gardenArtifacts);
    expect(bundle.garden?.species.length).toBe(5);
    const { cosmetics: _cosmetics, ...withoutCosmetics } = gardenArtifacts;
    await expect(load(withoutCosmetics as ReplayArtifacts)).rejects.toThrow(/artifact set/u);
    await expect(load({ ...gardenArtifacts, fiscal: cosmeticsArtifacts.fiscal! })).rejects.toThrow(/unlock_id/u);
  });

  it("round-trips the garden and rejects its absence at v25 and its presence below v25", async () => {
    const bundle = await load(gardenArtifacts);
    const shop = await load(cosmeticsArtifacts);
    // Build a v24 save through the real codec, then extend it to v25.
    const v24 = encodeFounderReplayState(founderState(shop, {})) as Record<string, unknown>;
    const garden = { ...newGardenState(bundle.garden!), salt_hex: "0123456789abcdef", tick_anchor_wall_ms: 1_000_000, tick_seq: 7,
      plots: [{ row: 0, col: 1, species_id: "strain_a", age_ticks: 3, matured_effect_ppm: 1_000_000 }] };
    const v25 = { ...v24, server_garden: garden };
    const restored = restoreFounderReplayState(v25, 25, bundle);
    expect(restored.serverGarden).toEqual(garden);
    expect(encodeFounderReplayState(restored)).toEqual(v25);
    expect(JSON.stringify((encodeFounderReplayState(restored) as Record<string, unknown>).server_garden))
      .toBe('{"salt_hex":"0123456789abcdef","tick_anchor_wall_ms":1000000,"tick_seq":7,"substrate_id":"bare_metal","substrate_set_wall_ms":null,"plots":[{"row":0,"col":1,"species_id":"strain_a","age_ticks":3,"matured_effect_ppm":1000000}],"seed_collection":["strain_a","strain_b"]}');
    expect(() => restoreFounderReplayState(v24, 25, bundle), "v25 without server_garden").toThrow();
    expect(() => restoreFounderReplayState(v25, 24, shop), "server_garden at v24").toThrow();
    const cases: [string, unknown][] = [
      ["missing tick_seq", (({ tick_seq: _t, ...rest }) => rest)(garden)],
      ["uppercase salt", { ...garden, salt_hex: "0123456789ABCDEF" }],
      ["unknown substrate", { ...garden, substrate_id: "quantum" }],
      ["unknown species", { ...garden, plots: [{ ...garden.plots[0]!, species_id: "strain_z" }] }],
      ["missing starter seed", { ...garden, seed_collection: ["strain_a"] }],
      ["extra key", { ...garden, cursor: 1 }],
    ];
    for (const [name, value] of cases) expect(() => restoreFounderReplayState({ ...v24, server_garden: value }, 25, bundle), name).toThrow();
  });
});

