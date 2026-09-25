import { describe, expect, it } from "vitest";

import { initialPetCareState } from "../src/pet/identity";
import { encodeFounderReplayState, loadReplayCatalogBundle, restoreFounderReplayState, type FounderReplayState, type ReplayArtifacts, type ReplayCatalogBundle } from "../src/replay";
import { constantsHashArtifacts, petSpeciesArtifacts, reputationArtifacts } from "./pet-fixture-bundle";

const PET = "01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa";
const identity = { species_id: "pet_species.server_room_cat", temperament: "curious" as const, palette_id: "pet_palette.fur_03",
  name_key: "pet.name.server_room_cat.n07", adopted_at_ms: 1_790_000_000_000, adopted_at_attended_ms: 5_400_000 };

const load = async (artifacts: ReplayArtifacts) => loadReplayCatalogBundle(await constantsHashArtifacts(artifacts), artifacts);

function founderState(bundle: ReplayCatalogBundle, overrides: Partial<FounderReplayState>): FounderReplayState {
  const founderResources = bundle.economy.resources.filter((row) => row.scope === "founder");
  const founderGenerators = bundle.economy.generatorClasses.filter((row) => bundle.economy.resource(row.price.resourceId)?.scope === "founder");
  return {
    wireVersion: 23, balances: Object.fromEntries(founderResources.map((row) => [row.id, row.initial])),
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
    petIdentities: { [PET]: identity }, cosmetics: { owned: [], equipped: {} }, serverGarden: null,
    ...overrides,
  };
}

describe("Founder v23 pet identities (AC2)", () => {
  it("round-trips an adopted pet and rejects every identity invariant violation", async () => {
    const bundle = await load(petSpeciesArtifacts);
    const encoded = encodeFounderReplayState(founderState(bundle, {})) as Record<string, unknown>;
    expect(Object.keys(encoded.pet_identities as object)).toEqual([PET]);
    const restored = restoreFounderReplayState(encoded, 23, bundle);
    expect(restored.petIdentities[PET]).toEqual(identity);
    expect(encodeFounderReplayState(restored)).toEqual(encoded);
    const cases: [string, Record<string, unknown>][] = [
      ["identity without care", { pets: {} }],
      ["care without identity", { pet_identities: {} }],
      ["missing key", { pet_identities: { [PET]: { ...identity, adopted_at_ms: undefined } } }],
      ["extra key", { pet_identities: { [PET]: { ...identity, extra: 1 } } }],
      ["unknown species", { pet_identities: { [PET]: { ...identity, species_id: "pet_species.robot_vacuum" } } }],
      ["unknown palette", { pet_identities: { [PET]: { ...identity, palette_id: "pet_palette.fur_99" } } }],
      ["unknown name", { pet_identities: { [PET]: { ...identity, name_key: "pet.name.server_room_cat.n99" } } }],
      ["adopted after the care watermark", { pet_identities: { [PET]: { ...identity, adopted_at_attended_ms: 5_400_001 } } }],
    ];
    for (const [name, patch] of cases) {
      const candidate = JSON.parse(JSON.stringify({ ...encoded, ...patch })) as Record<string, unknown>;
      expect(() => restoreFounderReplayState(candidate, 23, bundle), name).toThrow();
    }
    const nonV7 = "01986666-aaaa-4aaa-8aaa-aaaaaaaaaaaa";
    const v4 = { ...encoded, pets: { [nonV7]: (encoded.pets as Record<string, unknown>)[PET] }, pet_identities: { [nonV7]: identity } };
    expect(() => restoreFounderReplayState(v4, 23, bundle), "non-v7 pet id").toThrow();
    const { pet_identities: _omitted, ...missing } = encoded;
    expect(() => restoreFounderReplayState(missing, 23, bundle), "v23 without pet_identities").toThrow();
  });

  it("binds v23 to the pinned pet_species artifact in both directions", async () => {
    const species = await load(petSpeciesArtifacts);
    const tree = await load(reputationArtifacts);
    const v23 = encodeFounderReplayState(founderState(species, {})) as Record<string, unknown>;
    expect(() => restoreFounderReplayState(v23, 23, tree)).toThrow(/requires pet_species/u);
    const v22 = encodeFounderReplayState(founderState(tree, { wireVersion: 22, pets: {}, petIdentities: {} })) as Record<string, unknown>;
    expect(() => restoreFounderReplayState(v22, 22, tree)).not.toThrow();
    expect(() => restoreFounderReplayState(v22, 22, species)).toThrow(/requires Founder v23/u);
    expect(() => restoreFounderReplayState({ ...v22, pet_identities: {} }, 22, tree)).toThrow();
  });
});
