import { describe, expect, it } from "vitest";

import achievements from "../../balance/testdata/first-content/achievements-v1.json?raw";
import categories from "../../balance/testdata/first-content/categories-v1.json?raw";
import commons from "../../balance/commons/phase0.json?raw";
import doctrines from "../../balance/testdata/first-content/doctrines-v1.json?raw";
import economy from "../../balance/testdata/first-content/economy-v3.json?raw";
import factions from "../../balance/factions/phase0.json?raw";
import fiscal from "../../balance/testdata/first-content/fiscal-v1.json?raw";
import guilds from "../../balance/guilds/phase0.json?raw";
import meters from "../../balance/testdata/first-content/meters-v1.json?raw";
import minigameAPI from "../../balance/testdata/minigame-api-candidate-v1.json?raw";
import minigames from "../../testdata/minigame/pitch-v3.json?raw";
import pets from "../../balance/testdata/first-content/pets-v2.json?raw";
import pitch from "../../balance/testdata/pitch-v1.json?raw";
import prestige from "../../balance/prestige/phase0.json?raw";
import routes from "../../balance/testdata/permits-t3-gate-candidate-v1.json?raw";
import soul from "../../balance/testdata/first-content/soul-v1.json?raw";
import manifest from "../../planning/first-content-epoch/promotion-manifest.candidate.v1.json";
import { loadReplayCatalogBundle, type ReplayArtifacts } from "../src/replay";

const artifacts: ReplayArtifacts = {
  achievements,
  categories,
  commons,
  doctrines,
  economy,
  factions,
  fiscal,
  guilds,
  meters,
  minigame_api: minigameAPI,
  minigames,
  pets,
  pitch,
  prestige,
  routes,
  soul,
};

describe("first-content candidate artifacts", () => {
  it("loads the same complete sixteen-artifact bundle as the Go runtime", async () => {
    const hash = await constantsHashArtifacts(artifacts);
    expect(hash).toBe(manifest.constants_hash);
    const bundle = await loadReplayCatalogBundle(hash, artifacts);
    expect(bundle.constantsHash).toBe(hash);
    expect(bundle.meters?.meters).toHaveLength(11);
    expect(bundle.achievements?.definitions).toHaveLength(12);
    expect(bundle.doctrines?.transitions).toHaveLength(1);
    expect(bundle.minigames?.minigames).toHaveLength(1);
    expect(bundle.pets?.schema_version).toBe(2);
    expect(bundle.fiscal?.unlockRows.map((row) => row.unlockId)).toEqual(["minigame.pitch", "unlock.arcade"]);
    expect(bundle.soul?.recovery_activities).toHaveLength(3);
    expect(bundle.pitch?.metric_cards).toHaveLength(12);
    expect(bundle.minigameAPI?.tenants).toHaveLength(1);
  });

  it("rejects the pre-permits leaderboard gate set", async () => {
    const staleCategories = JSON.stringify({ ...JSON.parse(categories), full_gate_set: ["gate.t2_to_t3", "gate.t4_to_t5", "gate.t7_to_t8"] });
    const stale = { ...artifacts, categories: staleCategories };
    await expect(loadReplayCatalogBundle(await constantsHashArtifacts(stale), stale)).rejects.toThrow();
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
