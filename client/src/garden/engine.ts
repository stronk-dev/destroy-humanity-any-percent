// Server Garden SG2–SG6, byte-twin of server/garden/{state,engine,commands}.go.
// Objects are built in Go's struct field order so JSON encodings match bytes.
import { substream } from "../combat/rng";
import { byteCompare, GARDEN_MAX_AGE_TICKS, GARDEN_MAX_EFFECT_PPM, GARDEN_MAX_GRID_SIDE, GARDEN_PPM, gardenDimension, gardenSpecies, gardenStarters, gardenSubstrate, type GardenCatalog, type GardenDimensionRow, type GardenRecipe, type GardenSubstrate } from "./catalog";

export interface GardenPlot { row: number; col: number; species_id: string; age_ticks: number; matured_effect_ppm: number | null }
export interface GardenState {
  salt_hex: string | null; tick_anchor_wall_ms: number | null; tick_seq: number; substrate_id: string;
  substrate_set_wall_ms: number | null; plots: GardenPlot[]; seed_collection: string[];
}
export interface GardenPlotRef { row: number; col: number; species_id: string }
export interface GardenSpawn { row: number; col: number; species_id: string; recipe_id: string }
export interface GardenAdvance {
  ticks_applied: number; tick_seq_after: number; catchup_forfeited_ms: number; catchup_reason_key: string | null;
  matured: GardenPlotRef[]; spawned: GardenSpawn[];
}
export interface GardenAdvanceInput { readonly serverMs: number; readonly hostLevel: number; readonly unlocked: boolean; readonly salt: string }
export interface GardenHarvestTarget { row: number; col: number }
export interface GardenHarvestedPlot { col: number; row: number; species_id: string; units: number }
export interface GardenHarvest { plots: GardenHarvestedPlot[]; total_units: number; seeds_discovered: string[]; harvest_hash: string }

export class GardenRejection extends Error {
  constructor(readonly category: string, readonly detail: string) { super(`garden rejection ${category}/${detail}`); }
}

const saltPattern = /^[0-9a-f]{16}$/u;
const idPattern = /^[a-z][a-z0-9_]{0,47}$/u;

export function newGardenState(catalog: GardenCatalog): GardenState {
  return { salt_hex: null, tick_anchor_wall_ms: null, tick_seq: 0, substrate_id: catalog.defaultSubstrateId, substrate_set_wall_ms: null, plots: [], seed_collection: gardenStarters(catalog) };
}

export function cloneGardenState(state: GardenState): GardenState {
  return { salt_hex: state.salt_hex, tick_anchor_wall_ms: state.tick_anchor_wall_ms, tick_seq: state.tick_seq, substrate_id: state.substrate_id,
    substrate_set_wall_ms: state.substrate_set_wall_ms, plots: state.plots.map((plot) => ({ ...plot })), seed_collection: [...state.seed_collection] };
}

/** Canonical encoding in Go field order. */
export function encodeGardenState(state: GardenState): GardenState {
  return { salt_hex: state.salt_hex, tick_anchor_wall_ms: state.tick_anchor_wall_ms, tick_seq: state.tick_seq, substrate_id: state.substrate_id,
    substrate_set_wall_ms: state.substrate_set_wall_ms,
    plots: state.plots.map((plot) => ({ row: plot.row, col: plot.col, species_id: plot.species_id, age_ticks: plot.age_ticks, matured_effect_ppm: plot.matured_effect_ppm })),
    seed_collection: [...state.seed_collection] };
}

/** SG2 codec rules plus the pinned-catalog binding. */
export function parseGardenState(source: unknown, catalog: GardenCatalog): GardenState {
  const raw = exact(source, ["salt_hex", "tick_anchor_wall_ms", "tick_seq", "substrate_id", "substrate_set_wall_ms", "plots", "seed_collection"], "server_garden");
  if (raw.salt_hex !== null && (typeof raw.salt_hex !== "string" || !saltPattern.test(raw.salt_hex))) throw new SyntaxError("salt_hex grammar");
  const optional = (value: unknown, label: string): number | null => value === null ? null : int(value, 0, Number.MAX_SAFE_INTEGER, label);
  if (typeof raw.substrate_id !== "string" || !gardenSubstrate(catalog, raw.substrate_id)) throw new SyntaxError("unknown garden substrate");
  if (!Array.isArray(raw.plots) || !Array.isArray(raw.seed_collection)) throw new SyntaxError("garden collections must be arrays");
  const plots: GardenPlot[] = raw.plots.map((item) => {
    const plot = exact(item, ["row", "col", "species_id", "age_ticks", "matured_effect_ppm"], "garden plot");
    if (typeof plot.species_id !== "string" || !gardenSpecies(catalog, plot.species_id)) throw new SyntaxError("unknown plot species");
    return { row: int(plot.row, 0, GARDEN_MAX_GRID_SIDE - 1, "row"), col: int(plot.col, 0, GARDEN_MAX_GRID_SIDE - 1, "col"), species_id: plot.species_id,
      age_ticks: int(plot.age_ticks, 0, GARDEN_MAX_AGE_TICKS, "age_ticks"), matured_effect_ppm: plot.matured_effect_ppm === null ? null : int(plot.matured_effect_ppm, 0, GARDEN_MAX_EFFECT_PPM, "matured_effect_ppm") };
  });
  for (let index = 1; index < plots.length; index += 1) if (!plotLess(plots[index - 1]!, plots[index]!)) throw new SyntaxError("plots must be sorted and unique");
  const seeds: string[] = [];
  for (const species of raw.seed_collection) {
    if (typeof species !== "string" || !idPattern.test(species) || !gardenSpecies(catalog, species) || seeds.length > 0 && byteCompare(seeds.at(-1)!, species) >= 0) throw new SyntaxError("seed_collection must be sorted, unique, and pinned");
    seeds.push(species);
  }
  for (const starter of gardenStarters(catalog)) if (!seeds.includes(starter)) throw new SyntaxError("seed_collection must include every starter");
  return { salt_hex: raw.salt_hex as string | null, tick_anchor_wall_ms: optional(raw.tick_anchor_wall_ms, "tick_anchor_wall_ms"), tick_seq: int(raw.tick_seq, 0, Number.MAX_SAFE_INTEGER, "tick_seq"),
    substrate_id: raw.substrate_id, substrate_set_wall_ms: optional(raw.substrate_set_wall_ms, "substrate_set_wall_ms"), plots, seed_collection: seeds };
}

export function gardenNeedsSalt(state: GardenState, unlocked: boolean): boolean { return unlocked && state.salt_hex === null; }

function plotLess(left: { row: number; col: number }, right: { row: number; col: number }): boolean { return left.row < right.row || left.row === right.row && left.col < right.col; }
function findPlot(state: GardenState, row: number, col: number): number { return state.plots.findIndex((plot) => plot.row === row && plot.col === col); }
function insertPlot(state: GardenState, plot: GardenPlot): void {
  let index = 0;
  while (index < state.plots.length && plotLess(state.plots[index]!, plot)) index += 1;
  state.plots.splice(index, 0, plot);
}
function collect(state: GardenState, species: string): boolean {
  if (state.seed_collection.includes(species)) return false;
  state.seed_collection.push(species);
  state.seed_collection.sort(byteCompare);
  return true;
}
function active(dimension: GardenDimensionRow, row: number, col: number): boolean { return row < dimension.height && col < dimension.width; }

/** SG3's lazy wall-clock advance: exact, keyed by the absolute tick sequence. */
export function advanceGarden(catalog: GardenCatalog, state: GardenState, input: GardenAdvanceInput): GardenAdvance {
  const summary: GardenAdvance = { ticks_applied: 0, tick_seq_after: state.tick_seq, catchup_forfeited_ms: 0, catchup_reason_key: null, matured: [], spawned: [] };
  if (!Number.isSafeInteger(input.serverMs) || input.serverMs < 0 || !Number.isSafeInteger(input.hostLevel) || input.hostLevel < 0) throw new RangeError("invalid garden advance");
  if ((input.salt !== "") !== gardenNeedsSalt(state, input.unlocked)) throw new RangeError("salt must be supplied exactly when it is initialized");
  if (!input.unlocked) return summary;
  if (state.salt_hex === null) {
    if (!saltPattern.test(input.salt)) throw new RangeError("salt grammar");
    state.salt_hex = input.salt;
  }
  if (state.tick_anchor_wall_ms === null) { state.tick_anchor_wall_ms = input.serverMs; return summary; }
  let anchor = state.tick_anchor_wall_ms;
  if (input.serverMs <= anchor) return summary;
  let elapsed = input.serverMs - anchor;
  if (elapsed > catalog.catchupCapMs) {
    summary.catchup_forfeited_ms = elapsed - catalog.catchupCapMs;
    anchor += summary.catchup_forfeited_ms;
    elapsed = catalog.catchupCapMs;
    summary.catchup_reason_key = catalog.catchupReasonKey;
  }
  const substrate = gardenSubstrate(catalog, state.substrate_id);
  if (!substrate) throw new RangeError("unknown substrate");
  const ticks = Math.floor(elapsed / substrate.tick_ms);
  const dimension = gardenDimension(catalog, input.hostLevel);
  const base = substream(BigInt(`0x${state.salt_hex}`), "garden.founder.v1").next();
  for (let step = 1; step <= ticks; step += 1) {
    if (fixedPoint(catalog, state, dimension)) break;
    tick(catalog, state, dimension, substrate, base, state.tick_seq + step, summary);
  }
  state.tick_seq += ticks;
  state.tick_anchor_wall_ms = anchor + ticks * substrate.tick_ms;
  summary.ticks_applied = ticks;
  summary.tick_seq_after = state.tick_seq;
  return summary;
}

/** SG8: `garden_advanced.v1` is emitted iff this is true. */
export function gardenAdvanceVisible(advance: GardenAdvance): boolean {
  return advance.ticks_applied > 0 && (advance.matured.length > 0 || advance.spawned.length > 0) || advance.catchup_forfeited_ms > 0;
}

function census(state: GardenState, dimension: GardenDimensionRow, row: number, col: number): Map<string, number> {
  const counts = new Map<string, number>();
  for (let dr = -1; dr <= 1; dr += 1) for (let dc = -1; dc <= 1; dc += 1) {
    if (dr === 0 && dc === 0) continue;
    const r = row + dr, c = col + dc;
    if (r < 0 || c < 0 || !active(dimension, r, c)) continue;
    const index = findPlot(state, r, c);
    if (index >= 0 && state.plots[index]!.matured_effect_ppm !== null) counts.set(state.plots[index]!.species_id, (counts.get(state.plots[index]!.species_id) ?? 0) + 1);
  }
  return counts;
}

function eligible(catalog: GardenCatalog, counts: Map<string, number>): GardenRecipe[] {
  return catalog.recipes.filter((recipe) => recipe.parents.every((parent) => (counts.get(parent.species_id) ?? 0) >= parent.min_mature_neighbours));
}

function fixedPoint(catalog: GardenCatalog, state: GardenState, dimension: GardenDimensionRow): boolean {
  if (state.plots.some((plot) => active(dimension, plot.row, plot.col) && plot.matured_effect_ppm === null)) return false;
  for (let row = 0; row < dimension.height; row += 1) for (let col = 0; col < dimension.width; col += 1) {
    if (findPlot(state, row, col) >= 0) continue;
    if (eligible(catalog, census(state, dimension, row, col)).length > 0) return false;
  }
  return true;
}

function tick(catalog: GardenCatalog, state: GardenState, dimension: GardenDimensionRow, substrate: GardenSubstrate, base: bigint, s: number, summary: GardenAdvance): void {
  for (const plot of state.plots) {
    if (!active(dimension, plot.row, plot.col) || plot.matured_effect_ppm !== null) continue;
    plot.age_ticks = Math.min(plot.age_ticks + 1, GARDEN_MAX_AGE_TICKS);
    if (plot.age_ticks >= gardenSpecies(catalog, plot.species_id)!.maturation_ticks) {
      plot.matured_effect_ppm = substrate.effect_ppm;
      summary.matured.push({ row: plot.row, col: plot.col, species_id: plot.species_id });
    }
  }
  const candidates: { row: number; col: number; recipes: GardenRecipe[] }[] = [];
  for (let row = 0; row < dimension.height; row += 1) for (let col = 0; col < dimension.width; col += 1) {
    if (findPlot(state, row, col) >= 0) continue;
    const recipes = eligible(catalog, census(state, dimension, row, col));
    if (recipes.length > 0) candidates.push({ row, col, recipes });
  }
  if (candidates.length === 0) return;
  const random = substream(base ^ BigInt(s), "garden.tick.v1");
  for (const cell of candidates) {
    const draw = Number(random.bound(BigInt(GARDEN_PPM)));
    let cumulative = 0;
    for (const recipe of cell.recipes) {
      const effective = Math.min(GARDEN_PPM, Math.floor(recipe.chance_ppm * substrate.chance_factor_ppm / GARDEN_PPM));
      cumulative = Math.min(GARDEN_PPM, cumulative + effective);
      if (draw < cumulative) {
        insertPlot(state, { row: cell.row, col: cell.col, species_id: recipe.child_species_id, age_ticks: 0, matured_effect_ppm: null });
        summary.spawned.push({ row: cell.row, col: cell.col, species_id: recipe.child_species_id, recipe_id: recipe.recipe_id });
        break;
      }
    }
  }
}

/** SG5 steps 2–3. */
export function gateGarden(catalog: GardenCatalog, unlocked: boolean, humanContentLocked: boolean): void {
  if (!unlocked) throw new GardenRejection("not_eligible", "fiscal_unlock_required");
  if (catalog.soulGate === "human_hobby" && humanContentLocked) throw new GardenRejection("not_eligible", "human_content_locked");
}

export function plantGarden(catalog: GardenCatalog, state: GardenState, hostLevel: number, row: number, col: number, speciesId: string): { row: number; col: number; species_id: string } {
  if (!active(gardenDimension(catalog, hostLevel), row, col)) throw new GardenRejection("not_eligible", "plot_dormant");
  if (findPlot(state, row, col) >= 0) throw new GardenRejection("not_eligible", "plot_occupied");
  if (!gardenSpecies(catalog, speciesId)) throw new GardenRejection("unknown_id", "garden_species");
  if (!state.seed_collection.includes(speciesId)) throw new GardenRejection("not_eligible", "seed_not_collected");
  insertPlot(state, { row, col, species_id: speciesId, age_ticks: 0, matured_effect_ppm: null });
  return { row, col, species_id: speciesId };
}

export function uprootGarden(state: GardenState, row: number, col: number): { row: number; col: number; species_id: string; was_mature: boolean } {
  const index = findPlot(state, row, col);
  if (index < 0) throw new GardenRejection("unknown_id", "garden_plot");
  const [plot] = state.plots.splice(index, 1);
  return { row, col, species_id: plot!.species_id, was_mature: plot!.matured_effect_ppm !== null };
}

export function setGardenSubstrate(catalog: GardenCatalog, state: GardenState, serverMs: number, substrateId: string): { from_substrate_id: string; to_substrate_id: string; partial_tick_forfeited_ms: number } {
  if (!gardenSubstrate(catalog, substrateId)) throw new GardenRejection("unknown_id", "garden_substrate");
  if (substrateId === state.substrate_id) throw new GardenRejection("not_eligible", "substrate_unchanged");
  if (state.substrate_set_wall_ms !== null && serverMs - state.substrate_set_wall_ms < catalog.substrateLockoutMs) throw new GardenRejection("not_eligible", "substrate_lockout");
  const forfeited = state.tick_anchor_wall_ms !== null && serverMs > state.tick_anchor_wall_ms ? serverMs - state.tick_anchor_wall_ms : 0;
  const result = { from_substrate_id: state.substrate_id, to_substrate_id: substrateId, partial_tick_forfeited_ms: forfeited };
  state.substrate_id = substrateId; state.tick_anchor_wall_ms = serverMs; state.substrate_set_wall_ms = serverMs;
  return result;
}

export function validGardenHarvestTargets(targets: readonly GardenHarvestTarget[]): boolean {
  if (targets.length < 1 || targets.length > GARDEN_MAX_GRID_SIDE * GARDEN_MAX_GRID_SIDE) return false;
  return targets.every((target, index) => Number.isSafeInteger(target.row) && Number.isSafeInteger(target.col) && target.row >= 0 && target.row < GARDEN_MAX_GRID_SIDE &&
    target.col >= 0 && target.col < GARDEN_MAX_GRID_SIDE && (index === 0 || plotLess(targets[index - 1]!, target)));
}

/** SG6's Founder side, including the canonical harvest hash. */
export async function harvestGarden(catalog: GardenCatalog, state: GardenState, intentId: string, targets: readonly GardenHarvestTarget[]): Promise<GardenHarvest> {
  if (!validGardenHarvestTargets(targets)) throw new RangeError("invalid garden harvest targets");
  for (const target of targets) if (findPlot(state, target.row, target.col) < 0) throw new GardenRejection("unknown_id", "garden_plot");
  for (const target of targets) if (state.plots[findPlot(state, target.row, target.col)]!.matured_effect_ppm === null) throw new GardenRejection("not_eligible", "plant_not_mature");
  const harvest: GardenHarvest = { plots: [], total_units: 0, seeds_discovered: [], harvest_hash: "" };
  for (const target of targets) {
    const index = findPlot(state, target.row, target.col);
    const [plot] = state.plots.splice(index, 1);
    const units = Number(BigInt(gardenSpecies(catalog, plot!.species_id)!.harvest_units) * BigInt(plot!.matured_effect_ppm!) / BigInt(GARDEN_PPM));
    harvest.plots.push({ col: plot!.col, row: plot!.row, species_id: plot!.species_id, units });
    harvest.total_units += units;
    if (collect(state, plot!.species_id)) { harvest.seeds_discovered.push(plot!.species_id); harvest.seeds_discovered.sort(byteCompare); }
  }
  harvest.harvest_hash = await gardenHarvestHash(intentId, harvest.plots, harvest.total_units);
  return harvest;
}

export async function gardenHarvestHash(intentId: string, plots: readonly GardenHarvestedPlot[], totalUnits: number): Promise<string> {
  const canonical = JSON.stringify({ intent_id: intentId, plots: plots.map((plot) => ({ col: plot.col, row: plot.row, species_id: plot.species_id, units: plot.units })), total_units: totalUnits });
  const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", new TextEncoder().encode(canonical)));
  return `sha256:${[...digest].map((value) => value.toString(16).padStart(2, "0")).join("")}`;
}

function exact(source: unknown, keys: readonly string[], label: string): Record<string, unknown> {
  if (source === null || typeof source !== "object" || Array.isArray(source)) throw new SyntaxError(`${label} must be an object`);
  const actual = Object.keys(source).sort(), expected = [...keys].sort();
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError(`${label} keys are not exact`);
  return source as Record<string, unknown>;
}

function int(source: unknown, minimum: number, maximum: number, label: string): number {
  if (typeof source !== "number" || !Number.isSafeInteger(source) || source < minimum || source > maximum) throw new SyntaxError(`${label} outside its exact domain`);
  return source;
}
