import { substream } from "../combat/rng";
import type { ArcadeSnakeContent } from "./catalog";
import { ArcadeRejection, encodeCanonical, keysOf, parseCommandObject, requireLiteralScaling, resolveArcadeCatalog, safeInteger,
  type ArcadeApplyInput, type ArcadeCreateInput, type ArcadeResult } from "./common";

// `snake` 1.0.0 (AR4), the TS mirror of server/arcade/snake.go. The client
// also runs `snakeStep` locally at presentation pace (AR6.3); the server
// snapshot stays the only truth.
export const SNAKE_ENGINE_REF = "snake" as const;
export const SNAKE_SCALING_DESTINATION = "minigame.arcade.snake" as const;
const RUN_SUBSTREAM = "snake.run.v1";
const FOOD_SEED_SUBSTREAM = "snake.food_seed.v1";
const FOOD_SUBSTREAM = "snake.food.v1";

export type SnakeDirection = "up" | "down" | "left" | "right";
export type SnakePhase = "playing" | "terminal";
export interface SnakeSnapshot {
  readonly arcade_content_hash: string; readonly arcade_schema_version: number; readonly phase: SnakePhase; readonly tick: number;
  readonly direction: SnakeDirection; readonly body: readonly number[]; readonly pending_growth: number; readonly food_cell: number;
  readonly food_index: number; readonly food_seed: string; readonly score: number; readonly width: number; readonly height: number; readonly revision: number;
}
export interface SnakeTurn { readonly tick: number; readonly direction: SnakeDirection }
export type SnakeCommand = Readonly<{ kind: "advance"; through_tick: number; turns: readonly SnakeTurn[] }> | Readonly<{ kind: "quit" }>;
export type SnakeMutable = { -readonly [K in keyof SnakeSnapshot]: SnakeSnapshot[K] };

const DELTAS: Readonly<Record<SnakeDirection, readonly [number, number]>> = { up: [0, -1], down: [0, 1], left: [-1, 0], right: [1, 0] };
const OPPOSITES: Readonly<Record<SnakeDirection, SnakeDirection>> = { up: "down", down: "up", left: "right", right: "left" };

export function isSnakeDirection(value: unknown): value is SnakeDirection { return value === "up" || value === "down" || value === "left" || value === "right"; }
export function snakeOpposite(direction: SnakeDirection): SnakeDirection { return OPPOSITES[direction]; }

export function snakeFoodSeed(seed: bigint): bigint {
  const runSeed = substream(seed, RUN_SUBSTREAM).next();
  return substream(runSeed, FOOD_SEED_SUBSTREAM).next();
}

/** Food k on the ascending list of empty cells, or -1 when the board is full. */
export function snakePlaceFood(body: readonly number[], cells: number, foodSeed: bigint, k: number): number {
  const occupied = new Set(body), empty: number[] = [];
  for (let cell = 0; cell < cells; cell++) if (!occupied.has(cell)) empty.push(cell);
  if (empty.length === 0) return -1;
  return empty[Number(substream(foodSeed ^ BigInt(k), FOOD_SUBSTREAM).bound(BigInt(empty.length)))]!;
}

export async function createSnake(input: ArcadeCreateInput): Promise<string> {
  const catalog = await resolveArcadeCatalog(input);
  if (input.mode !== "solo") throw new SyntaxError("snake is solo only");
  requireLiteralScaling(input.scaling_inputs, SNAKE_SCALING_DESTINATION);
  return encodeCanonical({ ...snakeGenesis(catalog.snake, input.seed), arcade_content_hash: input.content_hash, arcade_schema_version: input.content_schema_version });
}

export function snakeGenesis(content: ArcadeSnakeContent, seed: bigint): Omit<SnakeSnapshot, "arcade_content_hash" | "arcade_schema_version"> {
  const head = Math.floor(content.height / 2) * content.width + Math.floor(content.width / 2);
  const body = Array.from({ length: content.start_length }, (_, offset) => head - offset);
  const foodSeed = snakeFoodSeed(seed);
  return { phase: "playing", tick: 0, direction: "right", body, pending_growth: 0, food_cell: snakePlaceFood(body, content.width * content.height, foodSeed, 0),
    food_index: 1, food_seed: foodSeed.toString(), score: 0, width: content.width, height: content.height, revision: 1 };
}

export async function applySnake(input: ArcadeApplyInput): Promise<{ readonly snapshot: string; readonly result: ArcadeResult | null }> {
  const catalog = await resolveArcadeCatalog(input);
  requireLiteralScaling(input.scaling_inputs, SNAKE_SCALING_DESTINATION);
  const snapshot = decodeSnakeSnapshot(input.snapshot);
  if (snapshot.revision !== input.revision || snapshot.arcade_content_hash !== input.content_hash || snapshot.arcade_schema_version !== input.content_schema_version ||
    snapshot.width !== catalog.snake.width || snapshot.height !== catalog.snake.height || snapshot.food_seed !== snakeFoodSeed(input.seed).toString()) {
    throw new SyntaxError("snake snapshot diverges from its identity");
  }
  const command = decodeSnakeCommand(input.command);
  const { snapshot: next, result } = snakeTransition(snapshot, command, catalog.snake);
  next.revision = input.revision + 1;
  return { snapshot: encodeCanonical(next), result };
}

/** Pure transition on a copy: a rejected command mutates nothing (AR4.5). */
export function snakeTransition(snapshot: SnakeSnapshot, command: SnakeCommand, content: ArcadeSnakeContent): { snapshot: SnakeMutable; result: ArcadeResult | null } {
  if (snapshot.phase !== "playing") throw new ArcadeRejection("illegal_phase", `${command.kind} requires playing phase`);
  const work: SnakeMutable = { ...snapshot, body: [...snapshot.body] };
  if (command.kind === "quit") return { snapshot: work, result: terminate(work, "quit") };
  if (command.through_tick <= snapshot.tick || command.through_tick > snapshot.tick + content.max_ticks_per_advance) throw new ArcadeRejection("advance_window", "through_tick is outside the advance window");
  command.turns.forEach((turn, index) => {
    if (turn.tick <= snapshot.tick || turn.tick > command.through_tick) throw new ArcadeRejection("advance_window", "turn tick is outside the advance window");
    if (index > 0 && command.turns[index - 1]!.tick >= turn.tick) throw new ArcadeRejection("turns_not_ascending", "turn ticks must be strictly ascending");
  });
  const foodSeed = BigInt(work.food_seed);
  let turnIndex = 0;
  for (let t = work.tick + 1; t <= command.through_tick; t++) {
    const turn = command.turns[turnIndex];
    if (turn && turn.tick === t) {
      if (turn.direction === work.direction || turn.direction === OPPOSITES[work.direction]) throw new ArcadeRejection("invalid_turn", "a turn must be a 90 degree change");
      work.direction = turn.direction;
      turnIndex++;
    }
    const outcome = snakeStep(work, t, content.growth_per_food, foodSeed);
    if (outcome !== null) {
      if (t !== command.through_tick) throw new ArcadeRejection("advance_past_terminal", "the game ended before through_tick");
      return { snapshot: work, result: terminate(work, outcome) };
    }
  }
  return { snapshot: work, result: null };
}

/** AR4.4 in its exact order; returns a terminal outcome or null. */
export function snakeStep(snapshot: SnakeMutable, t: number, growthPerFood: number, foodSeed: bigint): "cleared" | "crashed" | null {
  const [dx, dy] = DELTAS[snapshot.direction];
  const head = snapshot.body[0]!;
  const x = head % snapshot.width + dx, y = Math.floor(head / snapshot.width) + dy;
  if (x < 0 || y < 0 || x >= snapshot.width || y >= snapshot.height) { snapshot.tick = t; return "crashed"; }
  const next = y * snapshot.width + x;
  const growing = snapshot.pending_growth > 0;
  const kept = growing ? snapshot.body : snapshot.body.slice(0, -1);
  if (kept.includes(next)) { snapshot.tick = t; return "crashed"; }
  snapshot.body = [next, ...kept];
  if (growing) snapshot.pending_growth--;
  if (next === snapshot.food_cell) {
    snapshot.score++;
    snapshot.pending_growth += growthPerFood;
    snapshot.food_cell = snakePlaceFood(snapshot.body, snapshot.width * snapshot.height, foodSeed, snapshot.food_index);
    snapshot.food_index++;
    if (snapshot.food_cell === -1) { snapshot.tick = t; return "cleared"; }
  }
  snapshot.tick = t;
  return null;
}

function terminate(snapshot: SnakeMutable, outcome: "cleared" | "crashed" | "quit"): ArcadeResult {
  snapshot.phase = "terminal";
  return { outcome, rating_delta: null, score_facts: [{ kind: "snake.food_eaten", value: snapshot.score }, { kind: "snake.ticks_survived", value: snapshot.tick }] };
}

export function decodeSnakeCommand(source: string): SnakeCommand {
  const row = parseCommandObject(source, () => { throw new ArcadeRejection("illegal_phase", "command is not an object"); });
  if (row.kind === "advance") {
    if (keysOf(row) !== "kind\0through_tick\0turns" || !Array.isArray(row.turns)) throw new ArcadeRejection("advance_window", "advance schema mismatch");
    if (!safeInteger(row.through_tick)) throw new ArcadeRejection("advance_window", "through_tick must be a safe integer");
    const turns = row.turns.map((item): SnakeTurn => {
      if (item === null || typeof item !== "object" || Array.isArray(item) || keysOf(item as Record<string, unknown>) !== "direction\0tick") throw new ArcadeRejection("advance_window", "turn schema mismatch");
      const turn = item as Record<string, unknown>;
      if (!safeInteger(turn.tick) || typeof turn.direction !== "string") throw new ArcadeRejection("advance_window", "turn tick must be a safe integer and direction a string");
      if (!isSnakeDirection(turn.direction)) throw new ArcadeRejection("invalid_turn", "unknown direction");
      return { tick: turn.tick, direction: turn.direction };
    });
    return { kind: "advance", through_tick: row.through_tick, turns };
  }
  if (row.kind === "quit") {
    if (keysOf(row) !== "kind") throw new ArcadeRejection("illegal_phase", "quit schema mismatch");
    return { kind: "quit" };
  }
  throw new ArcadeRejection("illegal_phase", "unknown command kind");
}

export function decodeSnakeSnapshot(source: string): SnakeMutable {
  const value = JSON.parse(source) as SnakeMutable;
  const keys = ["arcade_content_hash", "arcade_schema_version", "body", "direction", "food_cell", "food_index", "food_seed", "height", "pending_growth", "phase",
    "revision", "score", "tick", "width"];
  if (value === null || typeof value !== "object" || Object.keys(value).sort().join("\0") !== keys.join("\0") || value.arcade_schema_version !== 1 ||
    !/^sha256:[0-9a-f]{64}$/.test(value.arcade_content_hash) || value.phase !== "playing" && value.phase !== "terminal" || value.revision < 1 || value.tick < 0 ||
    value.width < 5 || value.width > 30 || value.height < 5 || value.height > 30 || !Array.isArray(value.body) || value.body.length < 2 ||
    value.pending_growth < 0 || value.score < 0 || value.food_index < 1 || value.food_index !== value.score + 1 || !isSnakeDirection(value.direction) ||
    !/^(0|[1-9][0-9]*)$/.test(value.food_seed) || BigInt(value.food_seed) >= 1n << 64n) throw new SyntaxError("invalid snake snapshot");
  const cells = value.width * value.height, seen = new Set<number>();
  for (const cell of value.body) {
    if (!Number.isSafeInteger(cell) || cell < 0 || cell >= cells || seen.has(cell)) throw new SyntaxError("invalid snake body");
    seen.add(cell);
  }
  if (value.food_cell < -1 || value.food_cell >= cells || value.food_cell >= 0 && seen.has(value.food_cell) || value.food_cell === -1 && value.phase !== "terminal") {
    throw new SyntaxError("invalid snake food");
  }
  return value;
}
