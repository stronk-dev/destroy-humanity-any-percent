import validFixture from "../../balance/testdata/server-garden/fixture-v1.json?raw";
import fiscalFixture from "../../balance/testdata/server-garden/fiscal-fixture-v1.json?raw";
import productionFiscal from "../../balance/fiscal/first-content.json?raw";
import catalogCorpus from "../../testdata/garden/catalog-fixtures-v1.json";
import engineCorpus from "../../testdata/garden/engine-corpus-v1.json";
import { describe, expect, it } from "vitest";

import { COPY_KEYS } from "../src/copy";
import { loadGardenCatalog, validateGardenTransition, type GardenCatalog, type GardenDeclarations } from "../src/garden/catalog";
import { advanceGarden, GardenRejection, harvestGarden, parseGardenState, plantGarden, setGardenSubstrate, uprootGarden, type GardenState } from "../src/garden/engine";

interface Op { op: "set" | "delete"; path: (string | number)[]; value?: unknown }
interface Case { name: string; ops?: Op[]; duplicate_key?: string }

function declarations(fiscalBytes: string): GardenDeclarations {
  const fiscal = JSON.parse(fiscalBytes) as { generator_level_rows: { generator_id: string }[]; unlock_rows: { unlock_id: string }[] };
  return { copyKeys: new Set(COPY_KEYS), resourceIds: new Set(catalogCorpus.resource_ids),
    fiscalUnlockIds: new Set(fiscal.unlock_rows.map((row) => row.unlock_id)), fiscalGeneratorIds: new Set(fiscal.generator_level_rows.map((row) => row.generator_id)) };
}

function applyCase(test: Case): string {
  if (test.duplicate_key) {
    const index = validFixture.indexOf("{");
    return `${validFixture.slice(0, index + 1)}${JSON.stringify(test.duplicate_key)}: "unrelated",${validFixture.slice(index + 1)}`;
  }
  let document = JSON.parse(validFixture) as unknown;
  for (const op of test.ops ?? []) document = mutate(document, op.path, op);
  return JSON.stringify(document);
}

function mutate(node: unknown, path: readonly (string | number)[], op: Op): unknown {
  if (path.length === 0) return structuredClone(op.value);
  const [head, ...rest] = path;
  if (Array.isArray(node)) { node[head as number] = mutate(node[head as number], rest, op); return node; }
  const record = node as Record<string, unknown>;
  if (rest.length === 0 && op.op === "delete") { delete record[head as string]; return record; }
  record[head as string] = mutate(record[head as string], rest, op);
  return record;
}

const fixtureDeclarations = declarations(fiscalFixture);

describe("server_garden loader (SG1, AC1)", () => {
  it("loads the SG12 fixture", () => {
    const catalog = loadGardenCatalog(validFixture, fixtureDeclarations);
    expect([catalog.species.length, catalog.substrates.length, catalog.recipes.length, catalog.unlockId]).toEqual([5, 4, 5, "minigame.server_garden"]);
  });

  it("rejects every shared fixture case", () => {
    expect(catalogCorpus.cases.length).toBeGreaterThanOrEqual(40);
    for (const test of catalogCorpus.cases as Case[]) expect(() => loadGardenCatalog(applyCase(test), fixtureDeclarations), test.name).toThrow();
  });

  it("rejects the fixture garden under the production Fiscal artifact (no garden unlock row)", () => {
    expect(() => loadGardenCatalog(validFixture, declarations(productionFiscal))).toThrow(/unlock_id/u);
  });

  it("keeps ids append-only across epochs", () => {
    const current = loadGardenCatalog(validFixture, fixtureDeclarations);
    expect(() => validateGardenTransition(current, current)).not.toThrow();
    const renamed = validFixture.replace('"recipe_id": "r_e_cluster"', '"recipe_id": "r_e_cluster2"');
    expect(() => validateGardenTransition(current, loadGardenCatalog(renamed, fixtureDeclarations))).toThrow(/r_e_cluster/u);
  });
});

function scenarioCatalog(ops: readonly Op[] | null): GardenCatalog {
  let document = JSON.parse(validFixture) as unknown;
  for (const op of ops ?? []) document = mutate(document, op.path, op);
  return loadGardenCatalog(JSON.stringify(document), fixtureDeclarations);
}

interface Step { op: string; server_ms?: number; host_level?: number; unlocked?: boolean; salt?: string; row?: number; col?: number; species_id?: string; substrate_id?: string; intent_id?: string; plots?: { row: number; col: number }[] }

async function runStep(catalog: GardenCatalog, state: GardenState, step: Step): Promise<unknown> {
  switch (step.op) {
    case "advance": return advanceGarden(catalog, state, { serverMs: step.server_ms ?? 0, hostLevel: step.host_level ?? 0, unlocked: step.unlocked ?? false, salt: step.salt ?? "" });
    case "plant": return plantGarden(catalog, state, step.host_level ?? 0, step.row ?? 0, step.col ?? 0, step.species_id ?? "");
    case "uproot": return uprootGarden(state, step.row ?? 0, step.col ?? 0);
    case "set_substrate": return setGardenSubstrate(catalog, state, step.server_ms ?? 0, step.substrate_id ?? "");
    case "harvest": return harvestGarden(catalog, state, step.intent_id ?? "", step.plots ?? []);
  }
  throw new Error(`unknown corpus op ${step.op}`);
}

describe("garden engine corpus (SG3–SG6, AC3/AC4/AC5 parity)", () => {
  it("reproduces every Go outcome byte-for-byte", async () => {
    expect(engineCorpus.scenarios.length).toBeGreaterThanOrEqual(10);
    for (const scenario of engineCorpus.scenarios) {
      const catalog = scenarioCatalog(scenario.catalog_ops as Op[] | null);
      let state = parseGardenState(scenario.initial, catalog);
      for (const [index, step] of scenario.steps.entries()) {
        const expected = scenario.outcomes[index]!;
        const working = structuredClone(state);
        let outcome: { result: unknown; rejection: unknown; state: unknown };
        try {
          const result = await runStep(catalog, working, step as Step);
          outcome = { result, rejection: null, state: working };
          state = working;
        } catch (error) {
          if (!(error instanceof GardenRejection)) throw error;
          outcome = { result: null, rejection: { category: error.category, detail: error.detail }, state };
        }
        expect(JSON.stringify(outcome), `${scenario.name} step ${index}`).toBe(JSON.stringify(expected));
      }
    }
  });

  it("is partition-invariant below the cap", () => {
    const catalog = scenarioCatalog([{ op: "set", path: ["recipes", 0, "chance_ppm"], value: 200000 }, { op: "set", path: ["recipes", 1, "chance_ppm"], value: 200000 }]);
    for (let trial = 0; trial < 30; trial += 1) {
      const start: GardenState = { salt_hex: (0x1234abcd00000000n + BigInt(trial * 7919)).toString(16).padStart(16, "0"), tick_anchor_wall_ms: 1_000_000, tick_seq: 0,
        substrate_id: ["bare_metal", "containerized", "mainframe", "chaos_monkey"][trial % 4]!, substrate_set_wall_ms: null,
        plots: [{ row: 0, col: 0, species_id: "strain_a", age_ticks: 0, matured_effect_ppm: null }, { row: 2, col: 2, species_id: "strain_b", age_ticks: 0, matured_effect_ppm: null }],
        seed_collection: ["strain_a", "strain_b"] };
      const end = 1_000_000 + 1 + (trial * 2_654_435) % (catalog.catchupCapMs - 1);
      const level = trial % 9;
      const whole = structuredClone(start);
      advanceGarden(catalog, whole, { serverMs: end, hostLevel: level, unlocked: true, salt: "" });
      for (const fraction of [0.1, 0.37, 0.5, 0.93]) {
        const parts = structuredClone(start);
        advanceGarden(catalog, parts, { serverMs: 1_000_000 + Math.floor((end - 1_000_000) * fraction), hostLevel: level, unlocked: true, salt: "" });
        advanceGarden(catalog, parts, { serverMs: end, hostLevel: level, unlocked: true, salt: "" });
        expect(JSON.stringify(parts), `trial ${trial} fraction ${fraction}`).toBe(JSON.stringify(whole));
      }
    }
  });
});
