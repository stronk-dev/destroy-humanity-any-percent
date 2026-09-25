// Server Garden SG1 (rfc/minigame-server-garden.md): the `server_garden`
// artifact grammar, byte-twin of server/garden/catalog.go. Any violation
// fails the whole load.
export const GARDEN_SCHEMA_VERSION = 1 as const;
export const GARDEN_PAYOUT_SCORE_FACT_ID = "garden.harvest_units";
export const GARDEN_PPM = 1_000_000;
export const GARDEN_MAX_GRID_SIDE = 6;
export const GARDEN_MAX_AGE_TICKS = 1_000_000;
export const GARDEN_MAX_EFFECT_PPM = 10_000_000;
const MIN_TICK_MS = 60_000, MAX_TICK_MS = 86_400_000, MAX_HARVEST_UNITS = 1_000_000_000;
const MAX_CATCHUP_CAP_MS = 604_800_000, MAX_LOCKOUT_MS = 86_400_000, EVALUATION_BUDGET = 2_000_000, MAX_ROWS = 256;

export interface GardenDimensionRow { readonly min_level: number; readonly width: number; readonly height: number }
export interface GardenSubstrate { readonly substrate_id: string; readonly tick_ms: number; readonly effect_ppm: number; readonly chance_factor_ppm: number; readonly name_copy_key: string; readonly tooltip_copy_key: string }
export interface GardenSpecies { readonly species_id: string; readonly maturation_ticks: number; readonly harvest_units: number; readonly starter: boolean; readonly name_copy_key: string; readonly description_copy_key: string }
export interface GardenParent { readonly species_id: string; readonly min_mature_neighbours: number }
export interface GardenRecipe { readonly recipe_id: string; readonly child_species_id: string; readonly parents: readonly GardenParent[]; readonly chance_ppm: number }
export interface GardenPayout { readonly credited_resource_id: string; readonly sends_per_day: number; readonly per_send_cap: number; readonly conversion_ppm: number; readonly payout_score_fact_id: string; readonly cap_reason_key: string }
export interface GardenCatalog {
  readonly unlockId: string; readonly hostGeneratorId: string; readonly soulGate: "unrelated" | "human_hobby";
  readonly maxWidth: number; readonly maxHeight: number; readonly dimensionRows: readonly GardenDimensionRow[]; readonly gridHardcapReasonKey: string;
  readonly catchupCapMs: number; readonly catchupReasonKey: string; readonly substrateLockoutMs: number; readonly lockoutReasonKey: string;
  readonly defaultSubstrateId: string; readonly substrates: readonly GardenSubstrate[]; readonly species: readonly GardenSpecies[]; readonly recipes: readonly GardenRecipe[];
  readonly payout: GardenPayout;
}
export interface GardenDeclarations {
  readonly copyKeys: ReadonlySet<string>; readonly resourceIds: ReadonlySet<string>;
  readonly fiscalUnlockIds: ReadonlySet<string>; readonly fiscalGeneratorIds: ReadonlySet<string>;
}

const idPattern = /^[a-z][a-z0-9_]{0,47}$/u;
const mechanical = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/u;

export function loadGardenCatalog(bytes: string, declarations: GardenDeclarations): GardenCatalog {
  if (hasDuplicateKeys(bytes)) throw new SyntaxError("server_garden has duplicate keys");
  const root = exactObject(JSON.parse(bytes), ["schema_version", "unlock_id", "host_generator_id", "soul_gate", "grid", "clock", "default_substrate_id", "substrates", "species", "recipes", "payout"], "server_garden");
  if (root.schema_version !== GARDEN_SCHEMA_VERSION) throw new SyntaxError("server_garden schema_version");
  const unlockId = mechanicalString(root.unlock_id, "unlock_id"), hostGeneratorId = mechanicalString(root.host_generator_id, "host_generator_id");
  if (!declarations.fiscalUnlockIds.has(unlockId)) throw new SyntaxError("unlock_id is not a Fiscal unlock row");
  if (!declarations.fiscalGeneratorIds.has(hostGeneratorId)) throw new SyntaxError("host_generator_id is not a Fiscal generator level row");
  if (root.soul_gate !== "unrelated" && root.soul_gate !== "human_hobby") throw new SyntaxError("unknown soul_gate");
  const grid = exactObject(root.grid, ["max_width", "max_height", "dimension_rows", "hardcap_reason_key"], "grid");
  const maxWidth = integer(grid.max_width, 1, GARDEN_MAX_GRID_SIDE, "max_width"), maxHeight = integer(grid.max_height, 1, GARDEN_MAX_GRID_SIDE, "max_height");
  const gridHardcapReasonKey = copyKey(grid.hardcap_reason_key, declarations);
  const rawRows = boundedArray(grid.dimension_rows, 1, "dimension_rows");
  const dimensionRows: GardenDimensionRow[] = [];
  rawRows.forEach((source, index) => {
    const row = exactObject(source, ["min_level", "width", "height"], "dimension row");
    const minLevel = integer(row.min_level, 0, Number.MAX_SAFE_INTEGER, "min_level"), width = integer(row.width, 1, maxWidth, "width"), height = integer(row.height, 1, maxHeight, "height");
    const prior = dimensionRows[index - 1];
    if (index === 0 && minLevel !== 0 || prior && minLevel <= prior.min_level) throw new SyntaxError("dimension min_level must start at 0 and strictly ascend");
    if (prior && (width < prior.width || height < prior.height)) throw new SyntaxError("dimension rows must be non-decreasing");
    dimensionRows.push(Object.freeze({ min_level: minLevel, width, height }));
  });
  const last = dimensionRows.at(-1)!;
  if (last.width !== maxWidth || last.height !== maxHeight) throw new SyntaxError("the last dimension row must equal the max grid");
  const clock = exactObject(root.clock, ["catchup_cap_ms", "catchup_reason_key", "substrate_lockout_ms", "lockout_reason_key"], "clock");
  const catchupCapMs = integer(clock.catchup_cap_ms, 1, MAX_CATCHUP_CAP_MS, "catchup_cap_ms"), substrateLockoutMs = integer(clock.substrate_lockout_ms, 0, MAX_LOCKOUT_MS, "substrate_lockout_ms");
  const catchupReasonKey = copyKey(clock.catchup_reason_key, declarations), lockoutReasonKey = copyKey(clock.lockout_reason_key, declarations);
  const substrates = sortedRows(root.substrates, 1, "substrate_id", (source) => {
    const row = exactObject(source, ["substrate_id", "tick_ms", "effect_ppm", "chance_factor_ppm", "name_copy_key", "tooltip_copy_key"], "substrate");
    return { substrate_id: id(row.substrate_id), tick_ms: integer(row.tick_ms, MIN_TICK_MS, MAX_TICK_MS, "tick_ms"),
      effect_ppm: integer(row.effect_ppm, 0, GARDEN_MAX_EFFECT_PPM, "effect_ppm"), chance_factor_ppm: integer(row.chance_factor_ppm, 0, GARDEN_MAX_EFFECT_PPM, "chance_factor_ppm"),
      name_copy_key: copyKey(row.name_copy_key, declarations), tooltip_copy_key: copyKey(row.tooltip_copy_key, declarations) };
  });
  const defaultSubstrateId = typeof root.default_substrate_id === "string" ? root.default_substrate_id : "";
  if (!substrates.some((row) => row.substrate_id === defaultSubstrateId)) throw new SyntaxError("default_substrate_id does not exist");
  const species = sortedRows(root.species, 1, "species_id", (source) => {
    const row = exactObject(source, ["species_id", "maturation_ticks", "harvest_units", "starter", "name_copy_key", "description_copy_key"], "species");
    if (typeof row.starter !== "boolean") throw new SyntaxError("species starter must be boolean");
    return { species_id: id(row.species_id), maturation_ticks: integer(row.maturation_ticks, 1, GARDEN_MAX_AGE_TICKS, "maturation_ticks"),
      harvest_units: integer(row.harvest_units, 0, MAX_HARVEST_UNITS, "harvest_units"), starter: row.starter,
      name_copy_key: copyKey(row.name_copy_key, declarations), description_copy_key: copyKey(row.description_copy_key, declarations) };
  });
  if (!species.some((row) => row.starter)) throw new SyntaxError("at least one species must be a starter");
  const speciesIds = new Set(species.map((row) => row.species_id));
  const recipes = sortedRows(root.recipes, 0, "recipe_id", (source) => {
    const row = exactObject(source, ["recipe_id", "child_species_id", "parents", "chance_ppm"], "recipe");
    const recipeId = id(row.recipe_id);
    if (typeof row.child_species_id !== "string" || !speciesIds.has(row.child_species_id)) throw new SyntaxError(`recipe ${recipeId} child does not exist`);
    const chance = integer(row.chance_ppm, 1, GARDEN_PPM, "chance_ppm");
    if (!Array.isArray(row.parents) || row.parents.length < 1 || row.parents.length > 2) throw new SyntaxError(`recipe ${recipeId} needs 1-2 parents`);
    const parents: GardenParent[] = [];
    let sum = 0;
    for (const item of row.parents) {
      const parent = exactObject(item, ["species_id", "min_mature_neighbours"], "parent");
      if (typeof parent.species_id !== "string" || !speciesIds.has(parent.species_id)) throw new SyntaxError(`recipe ${recipeId} parent does not exist`);
      const prior = parents.at(-1);
      if (prior && byteCompare(prior.species_id, parent.species_id) >= 0) throw new SyntaxError(`recipe ${recipeId} parents must be distinct and sorted`);
      const minimum = integer(parent.min_mature_neighbours, 1, 8, "min_mature_neighbours");
      sum += minimum;
      parents.push(Object.freeze({ species_id: parent.species_id, min_mature_neighbours: minimum }));
    }
    if (sum > 8) throw new SyntaxError(`recipe ${recipeId} parent counts exceed 8`);
    return { recipe_id: recipeId, child_species_id: row.child_species_id, parents: Object.freeze(parents), chance_ppm: chance };
  });
  // Rule 5: reachability fixpoint from the starters.
  const reachable = new Set(species.filter((row) => row.starter).map((row) => row.species_id));
  for (let changed = true; changed;) {
    changed = false;
    for (const recipe of recipes) if (!reachable.has(recipe.child_species_id) && recipe.parents.every((parent) => reachable.has(parent.species_id))) { reachable.add(recipe.child_species_id); changed = true; }
  }
  for (const row of species) if (!reachable.has(row.species_id)) throw new SyntaxError(`species ${row.species_id} is unreachable`);
  // Rule 6: geometric satisfiability.
  const capacity = mooreCapacity(maxWidth, maxHeight);
  for (const recipe of recipes) if (recipe.parents.reduce((total, parent) => total + parent.min_mature_neighbours, 0) > capacity) throw new SyntaxError(`recipe ${recipe.recipe_id} cannot be satisfied`);
  // Rule 8: evaluation budget.
  const minimumTick = Math.min(...substrates.map((row) => row.tick_ms));
  if (Math.ceil(catchupCapMs / minimumTick) * maxWidth * maxHeight > EVALUATION_BUDGET) throw new SyntaxError("clock evaluation budget exceeded");
  const payoutRow = exactObject(root.payout, ["credited_resource_id", "sends_per_day", "per_send_cap", "conversion_ppm", "payout_score_fact_id", "cap_reason_key"], "payout");
  const payout: GardenPayout = Object.freeze({ credited_resource_id: mechanicalString(payoutRow.credited_resource_id, "credited_resource_id"),
    sends_per_day: integer(payoutRow.sends_per_day, 0, Number.MAX_SAFE_INTEGER, "sends_per_day"), per_send_cap: integer(payoutRow.per_send_cap, 0, Number.MAX_SAFE_INTEGER, "per_send_cap"),
    conversion_ppm: integer(payoutRow.conversion_ppm, 0, GARDEN_PPM, "conversion_ppm"), payout_score_fact_id: mechanicalString(payoutRow.payout_score_fact_id, "payout_score_fact_id"),
    cap_reason_key: copyKey(payoutRow.cap_reason_key, declarations) });
  if (!declarations.resourceIds.has(payout.credited_resource_id) || payout.payout_score_fact_id !== GARDEN_PAYOUT_SCORE_FACT_ID) throw new SyntaxError("garden payout policy");
  return Object.freeze({ unlockId, hostGeneratorId, soulGate: root.soul_gate, maxWidth, maxHeight, dimensionRows: Object.freeze(dimensionRows), gridHardcapReasonKey,
    catchupCapMs, catchupReasonKey, substrateLockoutMs, lockoutReasonKey, defaultSubstrateId,
    substrates: Object.freeze(substrates), species: Object.freeze(species), recipes: Object.freeze(recipes), payout });
}

export function mooreCapacity(width: number, height: number): number { return Math.min(width, 3) * Math.min(height, 3) - 1; }

export function gardenDimension(catalog: GardenCatalog, level: number): GardenDimensionRow {
  let active = catalog.dimensionRows[0]!;
  for (const row of catalog.dimensionRows) if (row.min_level <= level) active = row;
  return active;
}

export function gardenStarters(catalog: GardenCatalog): string[] { return catalog.species.filter((row) => row.starter).map((row) => row.species_id); }

export function gardenSubstrate(catalog: GardenCatalog, id: string): GardenSubstrate | undefined { return catalog.substrates.find((row) => row.substrate_id === id); }
export function gardenSpecies(catalog: GardenCatalog, id: string): GardenSpecies | undefined { return catalog.species.find((row) => row.species_id === id); }

/** SG1 rule 11: species, substrate, and recipe ids are append-only across epochs. */
export function validateGardenTransition(current: GardenCatalog | undefined, next: GardenCatalog | undefined): void {
  if (!current) return;
  if (!next) throw new RangeError("server_garden cannot disappear between epochs");
  for (const row of current.species) if (!gardenSpecies(next, row.species_id)) throw new RangeError(`species ${row.species_id} was removed`);
  for (const row of current.substrates) if (!gardenSubstrate(next, row.substrate_id)) throw new RangeError(`substrate ${row.substrate_id} was removed`);
  for (const row of current.recipes) if (!next.recipes.some((candidate) => candidate.recipe_id === row.recipe_id)) throw new RangeError(`recipe ${row.recipe_id} was removed`);
}

function sortedRows<T>(source: unknown, minimum: number, key: string, parse: (row: unknown) => T): T[] {
  const rows = boundedArray(source, minimum, key).map(parse);
  for (let index = 1; index < rows.length; index += 1) {
    if (byteCompare((rows[index - 1] as Record<string, string>)[key]!, (rows[index] as Record<string, string>)[key]!) >= 0) throw new SyntaxError(`${key} rows must be unique and byte-sorted`);
  }
  return rows.map((row) => Object.freeze(row));
}

function boundedArray(source: unknown, minimum: number, label: string): unknown[] {
  if (!Array.isArray(source) || source.length < minimum || source.length > MAX_ROWS) throw new SyntaxError(`${label} must be a bounded array`);
  return source;
}

function id(source: unknown): string {
  if (typeof source !== "string" || !idPattern.test(source)) throw new SyntaxError("garden id grammar");
  return source;
}

function mechanicalString(source: unknown, label: string): string {
  if (typeof source !== "string" || !mechanical.test(source)) throw new SyntaxError(`${label} is not mechanical`);
  return source;
}

function copyKey(source: unknown, declarations: GardenDeclarations): string {
  const key = mechanicalString(source, "copy key");
  if (!declarations.copyKeys.has(key)) throw new SyntaxError(`unknown copy key ${key}`);
  return key;
}

function integer(source: unknown, minimum: number, maximum: number, label: string): number {
  if (typeof source !== "number" || !Number.isSafeInteger(source) || source < minimum || source > maximum) throw new SyntaxError(`${label} outside its exact domain`);
  return source;
}

function exactObject(source: unknown, keys: readonly string[], label: string): Record<string, unknown> {
  if (source === null || typeof source !== "object" || Array.isArray(source)) throw new SyntaxError(`${label} must be an object`);
  const actual = Object.keys(source).sort(byteCompare), expected = [...keys].sort(byteCompare);
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError(`${label} keys are not exact`);
  return source as Record<string, unknown>;
}

/** Duplicate keys at any depth, which JSON.parse would silently collapse. */
export function hasDuplicateKeys(text: string): boolean {
  const scopes: { kind: "{" | "["; keys: Set<string>; expectKey: boolean }[] = [];
  for (let index = 0; index < text.length; index += 1) {
    const char = text[index]!, top = scopes.at(-1);
    if (char === "\"") {
      let end = index + 1;
      while (text[end] !== "\"") end += text[end] === "\\" ? 2 : 1;
      if (top?.kind === "{" && top.expectKey) {
        const key = JSON.parse(text.slice(index, end + 1)) as string;
        if (top.keys.has(key)) return true;
        top.keys.add(key); top.expectKey = false;
      }
      index = end; continue;
    }
    if (char === "{" || char === "[") scopes.push({ kind: char, keys: new Set(), expectKey: char === "{" });
    else if (char === "}" || char === "]") scopes.pop();
    else if (char === "," && top?.kind === "{") top.expectKey = true;
  }
  return false;
}

export function byteCompare(left: string, right: string): number {
  const a = new TextEncoder().encode(left), b = new TextEncoder().encode(right);
  for (let index = 0; index < Math.min(a.length, b.length); index += 1) if (a[index] !== b[index]) return a[index]! - b[index]!;
  return a.length - b.length;
}
