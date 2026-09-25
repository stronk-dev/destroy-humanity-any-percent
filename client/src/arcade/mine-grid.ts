import { substream } from "../combat/rng";
import { arcadePreset, type ArcadeCatalog } from "./catalog";
import { ArcadeRejection, encodeCanonical, keysOf, parseCommandObject, requireLiteralScaling, resolveArcadeCatalog, safeInteger,
  type ArcadeApplyInput, type ArcadeCreateInput, type ArcadeResult } from "./common";

// `mine_grid` 1.0.0 (AR3), the TS mirror of server/arcade/mine_grid.go.
export const MINE_GRID_ENGINE_REF = "mine_grid" as const;
export const MINE_GRID_SCALING_DESTINATION = "minigame.arcade.mine_grid" as const;
const RUN_SUBSTREAM = "mine_grid.run.v1";
const MINES_SUBSTREAM = "mine_grid.mines.v1";

export type MineGridPhase = "setup" | "playing" | "terminal";
export interface MineGridRevealed { readonly adjacent: number; readonly cell: number }
export interface MineGridSnapshot {
  readonly arcade_content_hash: string; readonly arcade_schema_version: number; readonly phase: MineGridPhase; readonly preset_id: string | null;
  readonly width: number; readonly height: number; readonly mines: number; readonly first_cell: number;
  readonly revealed: readonly MineGridRevealed[]; readonly flags: readonly number[]; readonly exploded_cell: number;
  readonly mine_cells: readonly number[]; readonly revision: number;
}
export type MineGridCommand =
  | Readonly<{ kind: "choose_board"; preset_id: string }>
  | Readonly<{ kind: "reveal" | "toggle_flag" | "chord"; cell: number }>
  | Readonly<{ kind: "quit" }>;

type Mutable = { -readonly [K in keyof MineGridSnapshot]: MineGridSnapshot[K] };

export async function createMineGrid(input: ArcadeCreateInput): Promise<string> {
  await resolveArcadeCatalog(input);
  if (input.mode !== "solo") throw new SyntaxError("mine_grid is solo only");
  requireLiteralScaling(input.scaling_inputs, MINE_GRID_SCALING_DESTINATION);
  return encodeCanonical({ arcade_content_hash: input.content_hash, arcade_schema_version: input.content_schema_version, phase: "setup", preset_id: null,
    width: 0, height: 0, mines: 0, first_cell: -1, revealed: [], flags: [], exploded_cell: -1, mine_cells: [], revision: 1 } satisfies MineGridSnapshot);
}

export async function applyMineGrid(input: ArcadeApplyInput): Promise<{ readonly snapshot: string; readonly result: ArcadeResult | null }> {
  const catalog = await resolveArcadeCatalog(input);
  requireLiteralScaling(input.scaling_inputs, MINE_GRID_SCALING_DESTINATION);
  const snapshot = decodeMineGridSnapshot(input.snapshot);
  if (snapshot.revision !== input.revision || snapshot.arcade_content_hash !== input.content_hash || snapshot.arcade_schema_version !== input.content_schema_version) {
    throw new SyntaxError("mine_grid snapshot diverges from its identity");
  }
  validateAgainstCatalog(snapshot, catalog, input.seed);
  const command = decodeMineGridCommand(input.command);
  const result = transition(snapshot, command, catalog, input.seed);
  snapshot.revision = input.revision + 1;
  return { snapshot: encodeCanonical(snapshot), result };
}

function transition(snapshot: Mutable, command: MineGridCommand, catalog: ArcadeCatalog, seed: bigint): ArcadeResult | null {
  if (command.kind === "quit") {
    if (snapshot.phase === "terminal") throw new ArcadeRejection("illegal_phase", "quit requires a non-terminal phase");
    return terminate(snapshot, seed, "quit");
  }
  if (command.kind === "choose_board") {
    if (snapshot.phase !== "setup") throw new ArcadeRejection("illegal_phase", "choose_board requires setup phase");
    const preset = arcadePreset(catalog, command.preset_id);
    if (!preset) throw new ArcadeRejection("unknown_preset", "preset is not in the pinned catalog");
    snapshot.preset_id = preset.preset_id; snapshot.width = preset.width; snapshot.height = preset.height; snapshot.mines = preset.mines;
    snapshot.phase = "playing";
    return null;
  }
  if (snapshot.phase !== "playing") throw new ArcadeRejection("illegal_phase", `${command.kind} requires playing phase`);
  if (command.cell < 0 || command.cell >= snapshot.width * snapshot.height) throw new ArcadeRejection("cell_out_of_range", "cell is outside the board");
  const revealed = new Map(snapshot.revealed.map((row) => [row.cell, row.adjacent]));
  const flags = new Set(snapshot.flags);
  if (command.kind === "reveal") {
    if (flags.has(command.cell)) throw new ArcadeRejection("cell_flagged", "flagged cells cannot be revealed");
    if (revealed.has(command.cell)) throw new ArcadeRejection("cell_revealed", "cell is already revealed");
    if (snapshot.first_cell === -1) snapshot.first_cell = command.cell;
    const mines = new Set(mineGridMineCells(snapshot.width, snapshot.height, snapshot.mines, snapshot.first_cell, seed));
    if (mines.has(command.cell)) { snapshot.exploded_cell = command.cell; return terminate(snapshot, seed, "detonated"); }
    floodReveal(snapshot, revealed, flags, mines, command.cell);
  } else if (command.kind === "toggle_flag") {
    if (revealed.has(command.cell)) throw new ArcadeRejection("cell_revealed", "revealed cells cannot be flagged");
    if (flags.has(command.cell)) flags.delete(command.cell); else flags.add(command.cell);
    snapshot.flags = [...flags].sort((left, right) => left - right);
    return null;
  } else {
    const adjacent = revealed.get(command.cell);
    if (adjacent === undefined || adjacent === 0) throw new ArcadeRejection("chord_unsatisfied", "chord requires a revealed numbered cell");
    const neighbors = mineGridNeighbors(snapshot.width, snapshot.height, command.cell);
    if (neighbors.filter((neighbor) => flags.has(neighbor)).length !== adjacent) throw new ArcadeRejection("chord_unsatisfied", "flagged neighbors do not match the count");
    const mines = new Set(mineGridMineCells(snapshot.width, snapshot.height, snapshot.mines, snapshot.first_cell, seed));
    for (const neighbor of neighbors) {
      if (!revealed.has(neighbor) && !flags.has(neighbor) && mines.has(neighbor)) { snapshot.exploded_cell = neighbor; return terminate(snapshot, seed, "detonated"); }
    }
    for (const neighbor of neighbors) if (!revealed.has(neighbor) && !flags.has(neighbor)) floodReveal(snapshot, revealed, flags, mines, neighbor);
  }
  if (snapshot.revealed.length === snapshot.width * snapshot.height - snapshot.mines) return terminate(snapshot, seed, "cleared");
  return null;
}

function floodReveal(snapshot: Mutable, revealed: Map<number, number>, flags: ReadonlySet<number>, mines: ReadonlySet<number>, start: number): void {
  const stack = [start];
  while (stack.length > 0) {
    const cell = stack.pop()!;
    if (revealed.has(cell) || flags.has(cell) || mines.has(cell)) continue;
    const neighbors = mineGridNeighbors(snapshot.width, snapshot.height, cell);
    const count = neighbors.filter((neighbor) => mines.has(neighbor)).length;
    revealed.set(cell, count);
    if (count === 0) stack.push(...neighbors);
  }
  snapshot.revealed = [...revealed.entries()].sort(([left], [right]) => left - right).map(([cell, adjacent]) => ({ adjacent, cell }));
}

function terminate(snapshot: Mutable, seed: bigint, outcome: "cleared" | "detonated" | "quit"): ArcadeResult {
  snapshot.phase = "terminal";
  snapshot.mine_cells = snapshot.first_cell >= 0 ? mineGridMineCells(snapshot.width, snapshot.height, snapshot.mines, snapshot.first_cell, seed) : [];
  return { outcome, rating_delta: null, score_facts: [
    { kind: "mine_grid.cells_revealed", value: snapshot.revealed.length },
    { kind: "mine_grid.cleared", value: outcome === "cleared" ? 1 : 0 },
  ] };
}

/** The up-to-8 Chebyshev-distance-1 cells, ascending (AR3.2). */
export function mineGridNeighbors(width: number, height: number, cell: number): number[] {
  const x = cell % width, y = Math.floor(cell / width), result: number[] = [];
  for (let dy = -1; dy <= 1; dy++) for (let dx = -1; dx <= 1; dx++) {
    const nx = x + dx, ny = y + dy;
    if (dx === 0 && dy === 0 || nx < 0 || ny < 0 || nx >= width || ny >= height) continue;
    result.push(ny * width + nx);
  }
  return result;
}

/** AR3.3: a pure function of (seed, board, first_cell); never stored. */
export function mineGridMineCells(width: number, height: number, mines: number, firstCell: number, seed: bigint): number[] {
  const excluded = new Set([...mineGridNeighbors(width, height, firstCell), firstCell]);
  const eligible: number[] = [];
  for (let cell = 0; cell < width * height; cell++) if (!excluded.has(cell)) eligible.push(cell);
  const runSeed = substream(seed, RUN_SUBSTREAM).next();
  const random = substream(runSeed, MINES_SUBSTREAM);
  for (let index = eligible.length - 1; index > 0; index--) {
    const swap = Number(random.bound(BigInt(index + 1)));
    [eligible[index], eligible[swap]] = [eligible[swap]!, eligible[index]!];
  }
  return eligible.slice(0, mines).sort((left, right) => left - right);
}

export function decodeMineGridCommand(source: string): MineGridCommand {
  const row = parseCommandObject(source, () => { throw new ArcadeRejection("illegal_phase", "command is not an object"); });
  if (row.kind === "choose_board") {
    if (keysOf(row) !== "kind\0preset_id" || typeof row.preset_id !== "string") throw new ArcadeRejection("unknown_preset", "choose_board requires a preset_id string");
    return { kind: "choose_board", preset_id: row.preset_id };
  }
  if (row.kind === "reveal" || row.kind === "toggle_flag" || row.kind === "chord") {
    if (keysOf(row) !== "cell\0kind" || !safeInteger(row.cell)) throw new ArcadeRejection("cell_out_of_range", "cell command requires an integer cell");
    return { kind: row.kind, cell: row.cell };
  }
  if (row.kind === "quit") {
    if (keysOf(row) !== "kind") throw new ArcadeRejection("illegal_phase", "quit schema mismatch");
    return { kind: "quit" };
  }
  throw new ArcadeRejection("illegal_phase", "unknown command kind");
}

export function decodeMineGridSnapshot(source: string): Mutable {
  const value = JSON.parse(source) as Mutable;
  const keys = ["arcade_content_hash", "arcade_schema_version", "exploded_cell", "first_cell", "flags", "height", "mine_cells", "mines", "phase", "preset_id",
    "revealed", "revision", "width"];
  if (value === null || typeof value !== "object" || Object.keys(value).sort().join("\0") !== keys.join("\0") || value.arcade_schema_version !== 1 ||
    !/^sha256:[0-9a-f]{64}$/.test(value.arcade_content_hash) || !["setup", "playing", "terminal"].includes(value.phase) || value.revision < 1 ||
    !Number.isSafeInteger(value.first_cell) || value.first_cell < -1 || !Number.isSafeInteger(value.exploded_cell) || value.exploded_cell < -1 ||
    !Array.isArray(value.revealed) || !Array.isArray(value.flags) || !Array.isArray(value.mine_cells)) throw new SyntaxError("invalid mine_grid snapshot");
  const cells = value.width * value.height;
  if (value.preset_id === null ? value.width !== 0 || value.height !== 0 || value.mines !== 0 || value.first_cell !== -1 || value.revealed.length !== 0 ||
      value.flags.length !== 0 || value.exploded_cell !== -1 || value.phase === "playing"
    : value.phase === "setup" || value.width < 5 || value.width > 30 || value.height < 5 || value.height > 30 || value.mines < 1 || value.mines > cells - 9 ||
      value.first_cell >= cells || value.exploded_cell >= cells) throw new SyntaxError("invalid mine_grid snapshot board");
  // The hidden-information invariant: no mine position before terminal.
  if (value.phase !== "terminal" && (value.mine_cells.length !== 0 || value.exploded_cell !== -1)) throw new SyntaxError("mine_grid snapshot leaks mines");
  const sortedWithin = (list: readonly number[]) => list.every((cell, index) => cell >= 0 && cell < cells && (index === 0 || list[index - 1]! < cell));
  if (!sortedWithin(value.revealed.map((row) => row.cell)) || value.revealed.some((row) => row.adjacent < 0 || row.adjacent > 8) ||
    !sortedWithin(value.flags) || !sortedWithin(value.mine_cells)) throw new SyntaxError("invalid mine_grid snapshot cells");
  return value;
}

function validateAgainstCatalog(value: Mutable, catalog: ArcadeCatalog, seed: bigint): void {
  if (value.preset_id === null) return;
  const preset = arcadePreset(catalog, value.preset_id);
  if (!preset || preset.width !== value.width || preset.height !== value.height || preset.mines !== value.mines) throw new SyntaxError("mine_grid snapshot/catalog divergence");
  if (value.first_cell === -1) {
    if (value.revealed.length !== 0 || value.mine_cells.length !== 0) throw new SyntaxError("mine_grid snapshot/catalog divergence");
    return;
  }
  const mineList = mineGridMineCells(value.width, value.height, value.mines, value.first_cell, seed), mines = new Set(mineList);
  for (const row of value.revealed) {
    const count = mineGridNeighbors(value.width, value.height, row.cell).filter((neighbor) => mines.has(neighbor)).length;
    if (mines.has(row.cell) || count !== row.adjacent) throw new SyntaxError("mine_grid snapshot/catalog divergence");
  }
  if (value.phase === "terminal" && (value.mine_cells.length !== mineList.length || mineList.some((cell, index) => value.mine_cells[index] !== cell) ||
    value.exploded_cell !== -1 && !mines.has(value.exploded_cell))) throw new SyntaxError("mine_grid snapshot/catalog divergence");
}
