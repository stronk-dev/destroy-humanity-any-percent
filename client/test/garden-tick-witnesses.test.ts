import { expect, it } from "vitest";
import fixture from "../../balance/testdata/server-garden/fixture-v1.json?raw";
import fiscal from "../../balance/testdata/server-garden/fiscal-fixture-v1.json";
import catalogCorpus from "../../testdata/garden/catalog-fixtures-v1.json";
import population from "../../testdata/garden/tick-witnesses-v1.json";
import { COPY_KEYS } from "../src/copy";
import { SplitMix64, substream } from "../src/combat/rng";
import { loadGardenCatalog, gardenSubstrate } from "../src/garden/catalog";
import { advanceGarden, encodeGardenState, newGardenState, parseGardenState, type GardenPlot } from "../src/garden/engine";

type PlotTuple = [number, number, string, number, number | null];
interface Witness {
  name: string; level: number; seq: number; ticks: number; substrate: string; chances: number[];
  maturation_a?: number; chance_factor?: number; initial: PlotTuple[]; post: PlotTuple[];
  matured: [number, number, string][]; spawned: [number, number, string, string][];
}
const cases = population.cases as Witness[];
const declarations = { copyKeys: new Set(COPY_KEYS), resourceIds: new Set(catalogCorpus.resource_ids),
  fiscalUnlockIds: new Set(fiscal.unlock_rows.map((row) => row.unlock_id)), fiscalGeneratorIds: new Set(fiscal.generator_level_rows.map((row) => row.generator_id)) };
function plots(tuples: PlotTuple[]): GardenPlot[] {
  return tuples.map(([row, col, species_id, age_ticks, matured_effect_ppm]) => ({ row, col, species_id, age_ticks, matured_effect_ppm }));
}

it("pins the independent SG4 population and literal PRNG draws", () => {
  expect(population.version).toBe(1);
  expect(cases).toHaveLength(16);
  expect(new Set(cases.map((test) => test.name)).size).toBe(16);
  expect(new SplitMix64(0n).next()).toBe(0xe220a8397b1dcdafn);
  const base = substream(BigInt(`0x${population.salt}`), "garden.founder.v1").next();
  expect(base.toString()).toBe(population.base);
  expect(population.draws).toHaveLength(2);
  for (const [index, draws] of population.draws.entries()) {
    expect(draws).toHaveLength(8);
    const rng = substream(base ^ BigInt(index + 1), "garden.tick.v1");
    for (const draw of draws) expect(Number(rng.bound(1_000_000n))).toBe(draw);
  }
});

// No tick/reference implementation and no Go-produced outcomes: full literal post-plots
// and event records adjudicate each actual engine transition independently.
it.each(cases)("matches independent literal SG4 outcome: $name", (test) => {
  expect(test.chances).toHaveLength(5);
  expect(test.ticks >= 1 && test.ticks <= 2).toBe(true);
  const raw = JSON.parse(fixture) as { recipes: { chance_ppm: number }[]; species: { maturation_ticks: number }[]; substrates: { chance_factor_ppm: number }[] };
  for (const [index, chance] of test.chances.entries()) raw.recipes[index]!.chance_ppm = chance;
  if (test.maturation_a !== undefined) raw.species[0]!.maturation_ticks = test.maturation_a;
  if (test.chance_factor !== undefined) raw.substrates[0]!.chance_factor_ppm = test.chance_factor;
  const catalog = loadGardenCatalog(JSON.stringify(raw), declarations);
  const initial = newGardenState(catalog);
  initial.salt_hex = population.salt; initial.tick_anchor_wall_ms = 1_000_000;
  initial.tick_seq = test.seq; initial.substrate_id = test.substrate; initial.plots = plots(test.initial);
  const state = parseGardenState(initial, catalog);
  const at = 1_000_000 + test.ticks * gardenSubstrate(catalog, test.substrate)!.tick_ms;
  const post = structuredClone(state);
  post.tick_anchor_wall_ms = at; post.tick_seq = test.seq + test.ticks; post.plots = plots(test.post);
  const expected = { ticks_applied: test.ticks, tick_seq_after: test.seq + test.ticks,
    catchup_forfeited_ms: 0, catchup_reason_key: null,
    matured: test.matured.map(([row, col, species_id]) => ({ row, col, species_id })),
    spawned: test.spawned.map(([row, col, species_id, recipe_id]) => ({ row, col, species_id, recipe_id })) };
  const actual = advanceGarden(catalog, state, { serverMs: at, hostLevel: test.level, unlocked: true, salt: "" });
  expect(JSON.stringify(actual)).toBe(JSON.stringify(expected));
  expect(JSON.stringify(encodeGardenState(parseGardenState(state, catalog)))).toBe(JSON.stringify(encodeGardenState(parseGardenState(post, catalog))));
});
