import { describe, expect, it } from "vitest";

import achievements from "../../balance/achievements/first-content.json?raw";
import categories from "../../balance/categories/phase0.json?raw";
import commons from "../../balance/commons/phase0.json?raw";
import curriculum from "../../balance/curriculum/t0-t1.json?raw";
import doctrines from "../../balance/doctrines/first-content.json?raw";
import economy from "../../balance/catalogs/phase0.json?raw";
import factions from "../../balance/factions/phase0.json?raw";
import fiscal from "../../balance/fiscal/first-content.json?raw";
import guilds from "../../balance/guilds/phase0.json?raw";
import meters from "../../balance/meters/first-content.json?raw";
import minigameAPI from "../../balance/minigame-api/first-content.json?raw";
import minigames from "../../balance/minigames/first-content.json?raw";
import opportunities from "../../balance/opportunities/t0-t1.json?raw";
import pets from "../../balance/pets/first-content.json?raw";
import pitch from "../../balance/pitch.json?raw";
import prestige from "../../balance/prestige/phase0.json?raw";
import relevance from "../../balance/relevance/t0-t1.json?raw";
import routes from "../../balance/routes/phase0.json?raw";
import soul from "../../balance/soul/first-content.json?raw";
import tree from "../../balance/testdata/reputation-tree/fixture-v1.json?raw";
import { encodeFounderReplayState, loadReplayCatalogBundle, restoreFounderReplayState, type FounderReplayState, type ReplayArtifacts, type ReplayCatalogBundle } from "../src/replay";

// The epoch-8 live artifact set (balance/epochs/phase0.json) plus the fixture tree.
const live: ReplayArtifacts = { achievements, categories, commons, curriculum, doctrines, economy, factions, fiscal, guilds, meters,
  minigame_api: minigameAPI, minigames, opportunities, pets, pitch, prestige, relevance, routes, soul };
const declaredEconomy = (() => {
  const root = JSON.parse(economy) as { multiplier_sources: unknown[] };
  root.multiplier_sources.push({ id: "reputation.founder_bonus", slot: "prestige", target: "all", provider: "reputation_tree" });
  return JSON.stringify(root);
})();


async function treeBundle(): Promise<ReplayCatalogBundle> {
  const complete = { ...live, economy: declaredEconomy, reputation_tree: tree };
  return loadReplayCatalogBundle(await constantsHashArtifacts(complete), complete);
}

function founderState(bundle: ReplayCatalogBundle, overrides: Partial<FounderReplayState>): FounderReplayState {
  const founderResources = bundle.economy.resources.filter((row) => row.scope === "founder");
  const founderGenerators = bundle.economy.generatorClasses.filter((row) => bundle.economy.resource(row.price.resourceId)?.scope === "founder");
  return {
    wireVersion: 22, balances: Object.fromEntries(founderResources.map((row) => [row.id, row.initial])),
    generators: Object.fromEntries(founderGenerators.map((row) => [row.id, 0])), generatorPurchasedTotal: 0, upgradesOwned: new Set(),
    generatorsProvisioned: Object.fromEntries(founderGenerators.map((row) => [row.id, 0])),
    provisionRemaindersPpm: Object.fromEntries(founderGenerators.filter((row) => row.provision).map((row) => [row.provision!.generatorId, 0])),
    stockRateRemainderPpm: 0, evaluatedThroughMs: 1_786_000_000_000, computeCreditMs: 0, manualTokenMilli: 0, manualTokenRefilledAtMs: 1_786_000_000_000,
    routeKnowledgeBalance: 0, hintsUnlocked: new Set(), ledgerFactKinds: new Set(), reputationLevel: 9, networkSlots: [], cloutLifetime: 0,
    soul: bundle.soul!.policy.soul_initial, ageMs: 0, notoriety: 0, advisorMode: false, exitHistory: [],
    achievementsEarnedLifetime: new Set(), achievementScoreLifetime: 0,
    minigameRatings: Object.fromEntries(bundle.minigames!.minigames.map((row) => [row.minigame_id, { elo: row.rating_policy.starting_elo, season_member: row.rating_policy.season_member, games_counted: 0 }])),
    minigameOfflineQuality: Object.fromEntries(bundle.minigames!.minigames.map((row) => [row.minigame_id, { grade_ppm: row.offline_quality.neutral_floor_ppm, last_founder_attended_ms: 0, decay_remainder_ppm: 0 }])),
    pets: {}, fiscalCredit: 0, fiscalPeriodOpenedWallMs: 1_786_000_000_000, fiscalPeriodSequence: 0,
    fiscalGeneratorLevels: Object.fromEntries(bundle.fiscal!.generatorLevelRows.map((row) => [row.generatorId, 0])), fiscalUnlocks: new Set(),
    soulExhaustedSourceIds: new Set(), minigameSessionSeq: 0,
    reputationUnlockPpm: 250_000, reputationSpent: 6, reputationNodesOwned: ["reputation.unlock.p05", "reputation.unlock.p25"], petIdentities: {}, cosmetics: { owned: [], equipped: {} }, serverGarden: null,
    ...overrides,
  };
}

describe("Founder v22 Reputation tree state", () => {
  it("round-trips and enforces accounting and the unlock mirror against the pinned tree", async () => {
    const bundle = await treeBundle();
    const state = founderState(bundle, {});
    const encoded = encodeFounderReplayState(state) as Record<string, unknown>;
    expect(encoded.reputation_spent).toBe(6);
    expect(encoded.reputation_unlock_ppm).toBe(250_000);
    const restored = restoreFounderReplayState(encoded, 22, bundle);
    expect(restored.reputationSpent).toBe(6);
    expect(restored.reputationNodesOwned).toEqual(["reputation.unlock.p05", "reputation.unlock.p25"]);
    expect(encodeFounderReplayState(restored)).toEqual(encoded);
    for (const [name, patch] of [
      ["spent over level", { reputation_spent: 10 }],
      ["mirror mismatch", { reputation_unlock_ppm: 50_000 }],
      ["unsorted owned", { reputation_nodes_owned: ["reputation.unlock.p25", "reputation.unlock.p05"] }],
      ["missing spent", { reputation_spent: undefined }],
    ] as const) {
      const candidate: Record<string, unknown> = { ...encoded, ...patch };
      if ("reputation_spent" in patch && patch.reputation_spent === undefined) delete candidate.reputation_spent;
      expect(() => restoreFounderReplayState(candidate, 22, bundle), name).toThrow();
    }
  });

  it("rejects a pre-v22 unlock mirror and v22 without the tree artifact", async () => {
    const bundle = await treeBundle();
    const plain = await loadReplayCatalogBundle(await constantsHashArtifacts(live), live);
    const v21 = encodeFounderReplayState(founderState(plain, { wireVersion: 21, reputationUnlockPpm: 0, reputationSpent: 0, reputationNodesOwned: [] })) as Record<string, unknown>;
    expect(() => restoreFounderReplayState(v21, 21, plain)).not.toThrow();
    expect(() => restoreFounderReplayState({ ...v21, reputation_unlock_ppm: 50_000 }, 21, plain)).toThrow(/before Founder v22/u);
    expect(() => restoreFounderReplayState(encodeFounderReplayState(founderState(bundle, {})), 22, plain)).toThrow();
    expect(() => restoreFounderReplayState(v21, 21, bundle)).toThrow(/requires Founder v22/u);
  });
});

async function constantsHashArtifacts(values: ReplayArtifacts): Promise<string> {
  const encoder = new TextEncoder();
  const chunks: Uint8Array[] = [];
  for (const name of Object.keys(values).sort() as Array<keyof ReplayArtifacts>) {
    const nameBytes = encoder.encode(name);
    const data = encoder.encode(values[name]);
    chunks.push(frame(nameBytes.length), nameBytes, frame(data.length), data);
  }
  const input = new Uint8Array(chunks.reduce((sum, value) => sum + value.length, 0));
  let offset = 0;
  for (const chunk of chunks) {
    input.set(chunk, offset);
    offset += chunk.length;
  }
  const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", input));
  return `sha256:${[...digest].map((value) => value.toString(16).padStart(2, "0")).join("")}`;
}

function frame(value: number): Uint8Array {
  const result = new Uint8Array(8);
  new DataView(result.buffer).setBigUint64(0, BigInt(value), false);
  return result;
}
