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
import petSpecies from "../../balance/testdata/pet-species/fixture-v1.json?raw";
import type { ReplayArtifacts } from "../src/replay";

// Pet Adoption v1 fixture chain: the epoch-8 live artifacts, the Reputation
// tree fixture with its economy declaration, and the pet_species fixture.
export const liveArtifacts: ReplayArtifacts = { achievements, categories, commons, curriculum, doctrines, economy, factions, fiscal, guilds, meters,
  minigame_api: minigameAPI, minigames, opportunities, pets, pitch, prestige, relevance, routes, soul };
export const declaredEconomy = (() => {
  const root = JSON.parse(economy) as { multiplier_sources: unknown[] };
  root.multiplier_sources.push({ id: "reputation.founder_bonus", slot: "prestige", target: "all", provider: "reputation_tree" });
  return JSON.stringify(root);
})();
export const reputationArtifacts: ReplayArtifacts = { ...liveArtifacts, economy: declaredEconomy, reputation_tree: tree };
export const petSpeciesArtifacts: ReplayArtifacts = { ...reputationArtifacts, pet_species: petSpecies };

export async function constantsHashArtifacts(values: ReplayArtifacts): Promise<string> {
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
