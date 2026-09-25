import { describe, expect, it } from "vitest";

import epochSeed from "../../balance/epochs/phase0.json";
import candidateHash from "../../balance/testdata/t2/candidate-bundle-hash.txt?raw";
import categoriesCandidate from "../../balance/testdata/t2/categories-candidate-v1.json?raw";
import economyCandidate from "../../balance/testdata/t2/economy-candidate-v1.json?raw";
import routesCandidate from "../../balance/testdata/t2/routes-candidate-v1.json?raw";
import presentationCandidate from "../../balance/testdata/t2/presentation-candidate-v3.json";
import { applicationCopyCatalog } from "../src/copy";
import { parseCatalog } from "../src/economy-kernel";
import { loadReplayCatalogBundle, type ReplayArtifacts } from "../src/replay";
import { parseRoutesCatalog } from "../src/routes";
import raw_commons from "../../balance/commons/phase0.json?raw";
import raw_economy from "../../balance/catalogs/phase0.json?raw";
import raw_prestige from "../../balance/prestige/phase0.json?raw";
import raw_routes from "../../balance/routes/phase0.json?raw";
import raw_factions from "../../balance/factions/phase0.json?raw";
import raw_guilds from "../../balance/guilds/phase0.json?raw";
import raw_categories from "../../balance/categories/phase0.json?raw";
import raw_achievements from "../../balance/achievements/first-content.json?raw";
import raw_doctrines from "../../balance/doctrines/first-content.json?raw";
import raw_fiscal from "../../balance/fiscal/first-content.json?raw";
import raw_meters from "../../balance/meters/first-content.json?raw";
import raw_minigame_api from "../../balance/minigame-api/first-content.json?raw";
import raw_minigames from "../../balance/minigames/first-content.json?raw";
import raw_pets from "../../balance/pets/first-content.json?raw";
import raw_pitch from "../../balance/pitch.json?raw";
import raw_soul from "../../balance/soul/first-content.json?raw";
import raw_opportunities from "../../balance/opportunities/t0-t1.json?raw";
import raw_relevance from "../../balance/relevance/t0-t1.json?raw";
import raw_curriculum from "../../balance/curriculum/t0-t1.json?raw";

const EPOCH_EIGHT_RAW: Record<string, string> = {
  "balance/commons/phase0.json": raw_commons,
  "balance/catalogs/phase0.json": raw_economy,
  "balance/prestige/phase0.json": raw_prestige,
  "balance/routes/phase0.json": raw_routes,
  "balance/factions/phase0.json": raw_factions,
  "balance/guilds/phase0.json": raw_guilds,
  "balance/categories/phase0.json": raw_categories,
  "balance/achievements/first-content.json": raw_achievements,
  "balance/doctrines/first-content.json": raw_doctrines,
  "balance/fiscal/first-content.json": raw_fiscal,
  "balance/meters/first-content.json": raw_meters,
  "balance/minigame-api/first-content.json": raw_minigame_api,
  "balance/minigames/first-content.json": raw_minigames,
  "balance/pets/first-content.json": raw_pets,
  "balance/pitch.json": raw_pitch,
  "balance/soul/first-content.json": raw_soul,
  "balance/opportunities/t0-t1.json": raw_opportunities,
  "balance/relevance/t0-t1.json": raw_relevance,
  "balance/curriculum/t0-t1.json": raw_curriculum,
};

function epochEightArtifacts(): Record<string, string> {
  const artifacts: Record<string, string> = {};
  for (const row of epochSeed.artifacts) {
    const bytes = EPOCH_EIGHT_RAW[row.path];
    if (bytes === undefined) throw new Error(`missing artifact ${row.path}`);
    artifacts[row.name] = bytes;
  }
  return artifacts;
}

describe("Tier 2 content candidates (rfc/tier2-content.md §A/§B, fixture-first)", () => {
  it("loads the candidate bundle under the identity Go pins", async () => {
    const artifacts = { ...epochEightArtifacts(), economy: economyCandidate, routes: routesCandidate, categories: categoriesCandidate } as unknown as ReplayArtifacts;
    const bundle = await loadReplayCatalogBundle(candidateHash.trim(), artifacts);
    expect(bundle).toBeDefined();
    const routes = parseRoutesCatalog(JSON.parse(routesCandidate));
    expect(routes.gates.map((gate) => gate.gateId)).toEqual(["gate.t0_to_t1", "gate.t1_to_t2", "gate.t2_to_t3", "gate.t3_to_t4", "gate.t4_to_t5", "gate.t7_to_t8"]);
    const economy = parseCatalog(JSON.parse(economyCandidate));
    for (const id of ["generator.open_plan_floor", "generator.managed_services_contract", "generator.hot_desk_program"]) {
      expect(economy.generatorClass(id)?.tier).toBe(2);
    }
  });

  it("refuses the candidate routes with the epoch-8 categories (gate set must include gate.t1_to_t2)", async () => {
    const artifacts = { ...epochEightArtifacts(), economy: economyCandidate, routes: routesCandidate } as unknown as ReplayArtifacts;
    // A correct identity label isolates the rejection to the category/route gate-set check.
    await expect(loadReplayCatalogBundle(await artifactHash(artifacts), artifacts)).rejects.toThrow();
    const accepted = { ...artifacts, categories: categoriesCandidate } as unknown as ReplayArtifacts;
    expect(await artifactHash(accepted)).toBe(candidateHash.trim());
  });
});

describe("Tier 2 presentation candidate (§E4)", () => {
  const economy = JSON.parse(economyCandidate) as { generator_classes: { id: string; provisioned_hardcap?: { reason_key: string } }[]; upgrades: { id: string }[] };
  type Row = { id: string; title_key: string; description_key?: string; cap_reason_key?: string | null };
  const sorted = (rows: readonly Row[]) => rows.every((row, index) => index === 0 || rows[index - 1]!.id < row.id);
  const declared = (key: string | null | undefined) => key === null || key === undefined || applicationCopyCatalog.byKey.has(key);

  it("binds every candidate generator, upgrade, and previewable gate to declared copy", () => {
    const generators = presentationCandidate.generators as Row[], upgrades = presentationCandidate.upgrades as Row[], gates = presentationCandidate.gates as Row[];
    expect(presentationCandidate.schema_version).toBe(3);
    expect(sorted(generators) && sorted(upgrades) && sorted(gates)).toBe(true);
    expect(generators.map((row) => row.id)).toEqual(economy.generator_classes.map((row) => row.id).sort());
    expect(upgrades.map((row) => row.id)).toEqual(economy.upgrades.map((row) => row.id).sort());
    expect(gates.map((row) => row.id)).toEqual(["gate.t0_to_t1", "gate.t1_to_t2"]);
    for (const row of [...generators, ...upgrades, ...gates]) {
      for (const key of [row.title_key, row.description_key, row.cap_reason_key]) expect(declared(key), `${row.id} -> ${key}`).toBe(true);
    }
    for (const generator of economy.generator_classes) {
      const binding = generators.find((row) => row.id === generator.id)!;
      expect(binding.cap_reason_key ?? null, generator.id).toBe(generator.provisioned_hardcap?.reason_key ?? null);
    }
  });
});

// Mirrors the replay loader's framed SHA-256 identity (name/bytes length-prefixed, byte-sorted names).
async function artifactHash(artifacts: ReplayArtifacts): Promise<string> {
  const encoder = new TextEncoder();
  const record = artifacts as unknown as Record<string, string>;
  const chunks: Uint8Array[] = [];
  const frame = (value: number) => { const result = new Uint8Array(8); new DataView(result.buffer).setBigUint64(0, BigInt(value)); return result; };
  for (const name of Object.keys(record).sort()) {
    const nameBytes = encoder.encode(name), data = encoder.encode(record[name]);
    chunks.push(frame(nameBytes.length), nameBytes, frame(data.length), data);
  }
  const input = new Uint8Array(chunks.reduce((sum, chunk) => sum + chunk.length, 0));
  let offset = 0;
  for (const chunk of chunks) { input.set(chunk, offset); offset += chunk.length; }
  const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", input));
  return `sha256:${[...digest].map((value) => value.toString(16).padStart(2, "0")).join("")}`;
}
