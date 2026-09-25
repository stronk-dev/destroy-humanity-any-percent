import { describe, expect, it } from "vitest";

import { initialPetCareState } from "../src/pet/identity";
import { encodeFounderReplayState, loadReplayCatalogBundle, restoreFounderReplayState, type FounderReplayState, type ReplayArtifacts, type ReplayCatalogBundle } from "../src/replay";
import { constantsHashArtifacts, cosmeticsArtifacts, petSpeciesArtifacts } from "./cosmetic-fixture-bundle";

const PET = "01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa";
const identity = { species_id: "pet_species.server_room_cat", temperament: "curious" as const, palette_id: "pet_palette.fur_03",
  name_key: "pet.name.server_room_cat.n07", adopted_at_ms: 1_790_000_000_000, adopted_at_attended_ms: 5_400_000 };

const load = async (artifacts: ReplayArtifacts) => loadReplayCatalogBundle(await constantsHashArtifacts(artifacts), artifacts);

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
    petIdentities: { [PET]: identity }, cosmetics: { owned: ["horse_armor"], equipped: { [PET]: "horse_armor" } },
    ...overrides,
  };
}

describe("Founder v24 cosmetics (Cosmetic Shop v1 AC3)", () => {
  it("round-trips owned + equipped canonically and rejects every shape violation", async () => {
    const bundle = await load(cosmeticsArtifacts);
    const encoded = encodeFounderReplayState(founderState(bundle, {})) as Record<string, unknown>;
    expect(encoded.cosmetics).toEqual({ owned: ["horse_armor"], equipped: { [PET]: "horse_armor" } });
    const restored = restoreFounderReplayState(encoded, 24, bundle);
    expect(restored.cosmetics).toEqual({ owned: ["horse_armor"], equipped: { [PET]: "horse_armor" } });
    expect(encodeFounderReplayState(restored)).toEqual(encoded);
    const cases: [string, unknown][] = [
      ["null owned", { owned: null, equipped: {} }],
      ["null equipped", { owned: [], equipped: null }],
      ["unsorted owned", { owned: ["zebra_armor", "horse_armor"], equipped: {} }],
      ["duplicate owned", { owned: ["horse_armor", "horse_armor"], equipped: {} }],
      ["unknown id", { owned: ["zebra_armor"], equipped: {} }],
      ["equipped by a non-pet", { owned: ["horse_armor"], equipped: { "01986666-bbbb-7bbb-8bbb-bbbbbbbbbbbb": "horse_armor" } }],
      ["equipped not owned", { owned: [], equipped: { [PET]: "horse_armor" } }],
      ["extra key", { owned: [], equipped: {}, price: 0 }],
      ["missing equipped", { owned: [] }],
    ];
    for (const [name, cosmetics] of cases) {
      expect(() => restoreFounderReplayState({ ...encoded, cosmetics }, 24, bundle), name).toThrow();
    }
    const { cosmetics: _omitted, ...missing } = encoded;
    expect(() => restoreFounderReplayState(missing, 24, bundle), "v24 without cosmetics").toThrow();
  });

  it("binds v24 to the pinned cosmetics artifact in both directions", async () => {
    const shop = await load(cosmeticsArtifacts);
    const species = await load(petSpeciesArtifacts);
    const v24 = encodeFounderReplayState(founderState(shop, {})) as Record<string, unknown>;
    expect(() => restoreFounderReplayState(v24, 24, species)).toThrow(/requires cosmetics/u);
    const v23 = encodeFounderReplayState(founderState(species, { wireVersion: 23 })) as Record<string, unknown>;
    expect(v23.cosmetics).toBeUndefined();
    expect(() => restoreFounderReplayState(v23, 23, species)).not.toThrow();
    expect(() => restoreFounderReplayState(v23, 23, shop)).toThrow(/requires Founder v24/u);
    expect(() => restoreFounderReplayState({ ...v23, cosmetics: { owned: [], equipped: {} } }, 23, species)).toThrow();
  });
});
