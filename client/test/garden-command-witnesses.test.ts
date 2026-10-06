import { expect, it } from "vitest";
import fixture from "../../balance/testdata/server-garden/fixture-v1.json?raw";
import fiscal from "../../balance/testdata/server-garden/fiscal-fixture-v1.json";
import catalogCorpus from "../../testdata/garden/catalog-fixtures-v1.json";
import population from "../../testdata/garden/command-witnesses-v1.json";
import { COPY_KEYS } from "../src/copy";
import { loadGardenCatalog } from "../src/garden/catalog";
import { encodeGardenState, newGardenState, parseGardenState, plantGarden, uprootGarden, setGardenSubstrate, harvestGarden, gateGarden, GardenRejection, type GardenPlot } from "../src/garden/engine";

const declarations = { copyKeys: new Set(COPY_KEYS), resourceIds: new Set(catalogCorpus.resource_ids),
  fiscalUnlockIds: new Set(fiscal.unlock_rows.map((row) => row.unlock_id)), fiscalGeneratorIds: new Set(fiscal.generator_level_rows.map((row) => row.generator_id)) };
const catalog = loadGardenCatalog(fixture, declarations);
type Tuple = [number, number, string, number, number | null];
interface Case {
  name: string; anchor?: number; stamp?: number; initial: Tuple[]; post: Tuple[]; error: string | null; result: unknown;
  step: { op: string; host_level?: number; row?: number; col?: number; species_id?: string; server_ms?: number; substrate_id?: string; intent_id?: string; plots?: { row: number; col: number }[] };
}
const cases = population.cases as Case[];
const plots = (tuples: Tuple[]): GardenPlot[] => tuples.map(([row, col, species_id, age_ticks, matured_effect_ppm]) => ({ row, col, species_id, age_ticks, matured_effect_ppm }));

it("pins fifteen independently declared SG5 commands", () => {
  expect(population.version).toBe(1); expect(cases).toHaveLength(15);
  expect(new Set(cases.map((test) => test.name)).size).toBe(15);
});
it.each(cases)("matches direct command result and unchanged refusal state: $name", async (test) => {
  const initial = newGardenState(catalog);
  initial.salt_hex = "0123456789abcdef"; initial.tick_anchor_wall_ms = test.anchor ?? 1_000_000; initial.tick_seq = 17;
  initial.substrate_set_wall_ms = test.stamp ?? null; initial.plots = plots(test.initial);
  const state = parseGardenState(initial, catalog);
  const before = JSON.stringify(encodeGardenState(state));
  const post = structuredClone(state); post.plots = plots(test.post);
  const step = test.step;
  if (test.error === null && step.op === "set_substrate") {
    post.substrate_id = step.substrate_id!; post.tick_anchor_wall_ms = step.server_ms!; post.substrate_set_wall_ms = step.server_ms!;
  }
  let result: unknown; let failure: unknown;
  try {
    switch (step.op) {
      case "plant": result = plantGarden(catalog, state, step.host_level!, step.row!, step.col!, step.species_id!); break;
      case "uproot": result = uprootGarden(state, step.row!, step.col!); break;
      case "set_substrate": result = setGardenSubstrate(catalog, state, step.server_ms!, step.substrate_id!); break;
      case "harvest": result = await harvestGarden(catalog, state, step.intent_id!, step.plots!); break;
      default: throw new Error("unknown witness operation");
    }
  } catch (error) { failure = error; }
  if (test.error !== null) {
    expect(failure).toBeInstanceOf(GardenRejection);
    const rejected = failure as GardenRejection;
    expect(`${rejected.category}/${rejected.detail}`).toBe(test.error);
    expect(JSON.stringify(encodeGardenState(state))).toBe(before);
  } else {
    expect(failure).toBeUndefined(); expect(JSON.stringify(result)).toBe(JSON.stringify(test.result));
  }
  expect(JSON.stringify(encodeGardenState(parseGardenState(state, catalog)))).toBe(JSON.stringify(encodeGardenState(parseGardenState(post, catalog))));
});

const gates = ["unrelated", "human_hobby"].flatMap((mode) => [false, true].flatMap((unlocked) => [false, true].map((humanLocked) => ({ mode, unlocked, humanLocked }))));
it("pins eight SG5 gate combinations", () => { expect(gates).toHaveLength(8); });
it.each(gates)("checks gate priority $mode/$unlocked/$humanLocked", ({ mode, unlocked, humanLocked }) => {
  const raw = JSON.parse(fixture) as { soul_gate: string }; raw.soul_gate = mode;
  const ruled = loadGardenCatalog(JSON.stringify(raw), declarations);
  const expected = !unlocked ? "not_eligible/fiscal_unlock_required" : mode === "human_hobby" && humanLocked ? "not_eligible/human_content_locked" : null;
  let failure: unknown;
  try { gateGarden(ruled, unlocked, humanLocked); } catch (error) { failure = error; }
  if (expected === null) { expect(failure).toBeUndefined(); return; }
  expect(failure).toBeInstanceOf(GardenRejection);
  const rejected = failure as GardenRejection; expect(`${rejected.category}/${rejected.detail}`).toBe(expected);
});
