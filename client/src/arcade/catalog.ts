// Demo Disc Arcade pinned artifact (AR1.2/AR3.1/AR4.1), the TS mirror of
// server/arcade/catalog.go. Cross-artifact checks belong to replay loading.
export const ARCADE_SCHEMA_VERSION = 1 as const;
export const ARCADE_ENGINE_VERSION = "1.0.0" as const;
const STAGE_TIER_MAX = 9;

export interface ArcadeStage { readonly stage_id: string; readonly min_tier: number; readonly title_copy_key: string; readonly toys: readonly string[] }
export interface ArcadePreset { readonly preset_id: string; readonly width: number; readonly height: number; readonly mines: number; readonly copy_key: string }
export interface ArcadeSnakeContent {
  readonly width: number; readonly height: number; readonly start_length: number; readonly growth_per_food: number;
  readonly max_ticks_per_advance: number; readonly presentation_tick_ms: number;
}
export interface ArcadeCatalog {
  readonly schema_version: 1;
  readonly container: { readonly stages: readonly ArcadeStage[] };
  readonly mine_grid: { readonly presets: readonly ArcadePreset[] };
  readonly snake: ArcadeSnakeContent;
}

const mechanical = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;

export function parseArcadeCatalog(source: unknown, declaredCopyKeys: ReadonlySet<string>): ArcadeCatalog {
  if (declaredCopyKeys.size === 0) throw new SyntaxError("arcade copy registry is empty");
  const root = exactObject(source, ["schema_version", "container", "mine_grid", "snake"], "arcade catalog");
  if (root.schema_version !== ARCADE_SCHEMA_VERSION) throw new SyntaxError("invalid arcade schema version");
  const container = exactObject(root.container, ["stages"], "arcade container");
  if (!Array.isArray(container.stages) || container.stages.length === 0) throw new SyntaxError("arcade stages must be a non-empty list");
  const seenStages = new Set<string>();
  let priorTier = -1;
  const stages = container.stages.map((item): ArcadeStage => {
    const row = exactObject(item, ["stage_id", "min_tier", "title_copy_key", "toys"], "arcade stage");
    const stageID = identifier(row.stage_id, "arcade stage id");
    const minTier = integer(row.min_tier, 0, STAGE_TIER_MAX, "arcade stage min_tier");
    if (seenStages.has(stageID) || minTier <= priorTier || !Array.isArray(row.toys) || row.toys.length === 0) throw new SyntaxError("invalid arcade stage");
    seenStages.add(stageID);
    priorTier = minTier;
    let priorToy = "";
    const toys = row.toys.map((toy) => {
      const id = identifier(toy, "arcade toy");
      if (priorToy !== "" && byteCompare(priorToy, id) >= 0) throw new SyntaxError("arcade toys must be byte-sorted and unique");
      priorToy = id;
      return id;
    });
    return Object.freeze({ stage_id: stageID, min_tier: minTier, title_copy_key: copyKey(row.title_copy_key, declaredCopyKeys), toys: Object.freeze(toys) });
  });
  const mineGrid = exactObject(root.mine_grid, ["presets"], "arcade mine_grid");
  if (!Array.isArray(mineGrid.presets) || mineGrid.presets.length === 0) throw new SyntaxError("mine_grid presets must be a non-empty list");
  let priorPreset = "";
  const presets = mineGrid.presets.map((item): ArcadePreset => {
    const row = exactObject(item, ["preset_id", "width", "height", "mines", "copy_key"], "mine_grid preset");
    const presetID = identifier(row.preset_id, "mine_grid preset id");
    if (priorPreset !== "" && byteCompare(priorPreset, presetID) >= 0) throw new SyntaxError("mine_grid presets must be byte-sorted");
    priorPreset = presetID;
    const width = integer(row.width, 5, 30, "mine_grid width"), height = integer(row.height, 5, 30, "mine_grid height");
    const mines = integer(row.mines, 1, width * height - 9, "mine_grid mines");
    if (row.copy_key !== `arcade.mine_grid.preset.${presetID}`) throw new SyntaxError("mine_grid preset copy key must follow its id");
    return Object.freeze({ preset_id: presetID, width, height, mines, copy_key: copyKey(row.copy_key, declaredCopyKeys) });
  });
  const snakeRow = exactObject(root.snake, ["width", "height", "start_length", "growth_per_food", "max_ticks_per_advance", "presentation_tick_ms"], "arcade snake");
  const width = integer(snakeRow.width, 5, 30, "snake width");
  const snake: ArcadeSnakeContent = Object.freeze({
    width, height: integer(snakeRow.height, 5, 30, "snake height"),
    start_length: integer(snakeRow.start_length, 2, Math.floor(width / 2), "snake start_length"),
    growth_per_food: integer(snakeRow.growth_per_food, 1, 8, "snake growth_per_food"),
    max_ticks_per_advance: integer(snakeRow.max_ticks_per_advance, 1, 256, "snake max_ticks_per_advance"),
    presentation_tick_ms: integer(snakeRow.presentation_tick_ms, 50, 1000, "snake presentation_tick_ms"),
  });
  return Object.freeze({ schema_version: 1, container: Object.freeze({ stages: Object.freeze(stages) }), mine_grid: Object.freeze({ presets: Object.freeze(presets) }), snake });
}

export function arcadePreset(catalog: ArcadeCatalog, presetID: string): ArcadePreset | undefined {
  return catalog.mine_grid.presets.find((row) => row.preset_id === presetID);
}

/** The stage with the greatest min_tier ≤ the authoritative tier (AR1.2). */
export function activeArcadeStage(catalog: ArcadeCatalog, tier: number): ArcadeStage | undefined {
  return [...catalog.container.stages].reverse().find((stage) => stage.min_tier <= tier);
}

export async function arcadeContentHash(bytes: string | Uint8Array): Promise<string> {
  const data: Uint8Array<ArrayBuffer> = typeof bytes === "string" ? new TextEncoder().encode(bytes) : new Uint8Array(bytes);
  const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", data));
  return `sha256:${[...digest].map((value) => value.toString(16).padStart(2, "0")).join("")}`;
}

function copyKey(value: unknown, declared: ReadonlySet<string>): string {
  const key = identifier(value, "arcade copy key");
  if (!declared.has(key)) throw new SyntaxError("unknown arcade copy key");
  return key;
}

function identifier(value: unknown, label: string): string {
  if (typeof value !== "string" || !mechanical.test(value)) throw new SyntaxError(`invalid ${label}`);
  return value;
}

function integer(value: unknown, minimum: number, maximum: number, label: string): number {
  if (typeof value !== "number" || !Number.isSafeInteger(value) || value < minimum || value > maximum) throw new SyntaxError(`invalid ${label}`);
  return value;
}

function exactObject(source: unknown, keys: readonly string[], label: string): Record<string, unknown> {
  if (source === null || typeof source !== "object" || Array.isArray(source)) throw new SyntaxError(`${label} must be an object`);
  const row = source as Record<string, unknown>, actual = Object.keys(row).sort(byteCompare), expected = [...keys].sort(byteCompare);
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError(`${label} fields are not exact`);
  return row;
}

export function byteCompare(left: string, right: string): number {
  const a = new TextEncoder().encode(left), b = new TextEncoder().encode(right);
  for (let index = 0; index < Math.min(a.length, b.length); index++) if (a[index] !== b[index]) return a[index]! - b[index]!;
  return a.length - b.length;
}
