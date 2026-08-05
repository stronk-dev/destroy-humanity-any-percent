import { MAX_EXACT_INTEGER } from "../numeric";
import type { PetCatalog } from "./catalog";
import { PET_BEHAVIOR_QUEUE_HARDCAP, PET_BEHAVIOR_STATES, PET_STAT_IDS, type PetBehaviorState, type PetStatID } from "./grammar";

const petID = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;
const mechanicalID = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;

export interface PetBehaviorQueueEntry {
  readonly behavior_id: string;
  readonly due_attended_ms: number;
}

export interface PetCareState {
  readonly stats_ppm: Record<PetStatID, number>;
  readonly stat_decay_remainders_ppm: Record<PetStatID, number>;
  readonly cooldown_until_attended_ms: Record<string, number>;
  readonly trust_ppm: number;
  readonly trust_decay_remainder_ppm: number;
  readonly evaluated_through_attended_ms: number;
  readonly behavior_state: PetBehaviorState;
  readonly behavior_entered_at_attended_ms: number;
  readonly behavior_queue: readonly PetBehaviorQueueEntry[];
  readonly behavior_prng_cursor: number;
}

export interface PetStateDeclarations {
  readonly action_ids: readonly string[];
  readonly behavior_ids: readonly string[];
}

const careStateKeys = ["stats_ppm", "stat_decay_remainders_ppm", "cooldown_until_attended_ms", "trust_ppm",
  "trust_decay_remainder_ppm", "evaluated_through_attended_ms", "behavior_state", "behavior_entered_at_attended_ms", "behavior_queue", "behavior_prng_cursor"] as const;

export function parsePetCareStates(source: unknown, declarations: PetStateDeclarations): Record<string, PetCareState> {
  const actions = declaredIDs(declarations.action_ids, "action");
  const behaviors = declaredIDs(declarations.behavior_ids, "behavior");
  const rawStates = record(source, "pets");
  const result: Record<string, PetCareState> = {};
  for (const id of Object.keys(rawStates).sort(byteCompare)) {
    if (!petID.test(id)) throw new SyntaxError("invalid pet id");
    const raw = exactObject(rawStates[id], careStateKeys, "pet state");
    const stats = statRecord(raw.stats_ppm, 1_000_000, "stats_ppm");
    const remainders = statRecord(raw.stat_decay_remainders_ppm, MAX_EXACT_INTEGER, "stat_decay_remainders_ppm");
    const cooldownRaw = record(raw.cooldown_until_attended_ms, "cooldowns");
    const cooldowns: Record<string, number> = {};
    for (const actionID of Object.keys(cooldownRaw).sort(byteCompare)) {
      if (!actions.has(actionID)) throw new SyntaxError("unknown pet action cooldown");
      cooldowns[actionID] = safeInteger(cooldownRaw[actionID], 0, MAX_EXACT_INTEGER);
    }
    if (!Array.isArray(raw.behavior_queue) || raw.behavior_queue.length > PET_BEHAVIOR_QUEUE_HARDCAP) throw new SyntaxError("invalid behavior queue");
    const queue = raw.behavior_queue.map((entry) => {
      const value = exactObject(entry, ["behavior_id", "due_attended_ms"], "behavior queue entry");
      if (typeof value.behavior_id !== "string" || !behaviors.has(value.behavior_id)) throw new SyntaxError("unknown pet behavior");
      return { behavior_id: value.behavior_id, due_attended_ms: safeInteger(value.due_attended_ms, 0, MAX_EXACT_INTEGER) };
    });
    if (typeof raw.behavior_state !== "string" || !PET_BEHAVIOR_STATES.includes(raw.behavior_state as PetBehaviorState)) throw new SyntaxError("invalid behavior state");
    if (new Set(queue.map((entry) => entry.behavior_id)).size !== queue.length) throw new SyntaxError("duplicate queued pet behavior");
    for (let index = 1; index < queue.length; index++) {
      const prior = queue[index - 1]!, current = queue[index]!;
      if (prior.due_attended_ms > current.due_attended_ms || prior.due_attended_ms === current.due_attended_ms && byteCompare(prior.behavior_id, current.behavior_id) >= 0) throw new SyntaxError("noncanonical behavior queue");
    }
    const evaluatedThrough = safeInteger(raw.evaluated_through_attended_ms, 0, MAX_EXACT_INTEGER);
    const behaviorEnteredAt = safeInteger(raw.behavior_entered_at_attended_ms, 0, evaluatedThrough);
    result[id] = {
      stats_ppm: stats,
      stat_decay_remainders_ppm: remainders,
      cooldown_until_attended_ms: cooldowns,
      trust_ppm: safeInteger(raw.trust_ppm, 0, 1_000_000),
      trust_decay_remainder_ppm: safeInteger(raw.trust_decay_remainder_ppm, 0, MAX_EXACT_INTEGER),
      evaluated_through_attended_ms: evaluatedThrough,
      behavior_state: raw.behavior_state as PetBehaviorState,
      behavior_entered_at_attended_ms: behaviorEnteredAt,
      behavior_queue: queue,
      behavior_prng_cursor: safeInteger(raw.behavior_prng_cursor, 0, 0),
    };
  }
  return result;
}

export function validatePetCareStatesForCatalog(states: Readonly<Record<string, PetCareState>>, catalog: PetCatalog): void {
  for (const state of Object.values(states)) {
    if (state.trust_decay_remainder_ppm >= catalog.stat_policy.grid_ms ||
      PET_STAT_IDS.some((id) => state.stat_decay_remainders_ppm[id] >= catalog.stat_policy.grid_ms)) {
      throw new SyntaxError("pet decay remainder exceeds pinned grid");
    }
  }
}

function declaredIDs(values: readonly string[], label: string): ReadonlySet<string> {
  const result = new Set<string>();
  for (const value of values) {
    if (!mechanicalID.test(value) || result.has(value)) throw new SyntaxError(`invalid ${label} declarations`);
    result.add(value);
  }
  return result;
}

function statRecord(source: unknown, maximum: number, label: string): Record<PetStatID, number> {
  const value = exactObject(source, PET_STAT_IDS, label);
  return Object.fromEntries(PET_STAT_IDS.map((id) => [id, safeInteger(value[id], 0, maximum)])) as Record<PetStatID, number>;
}

function record(source: unknown, label: string): Record<string, unknown> {
  if (typeof source !== "object" || source === null || Array.isArray(source)) throw new SyntaxError(`${label} must be an object`);
  return source as Record<string, unknown>;
}

function exactObject(source: unknown, keys: readonly string[], label: string): Record<string, unknown> {
  const value = record(source, label);
  if (!same(Object.keys(value).sort(byteCompare), [...keys].sort(byteCompare))) throw new SyntaxError(`${label} fields are not exact`);
  return value;
}

function safeInteger(value: unknown, minimum: number, maximum: number): number {
  if (typeof value !== "number" || !Number.isSafeInteger(value) || value < minimum || value > maximum) throw new SyntaxError("integer outside pet state domain");
  return value;
}

function same(left: readonly string[], right: readonly string[]): boolean {
  return left.length === right.length && left.every((value, index) => value === right[index]);
}

function byteCompare(left: string, right: string): number {
  const a = new TextEncoder().encode(left); const b = new TextEncoder().encode(right);
  for (let index = 0; index < Math.min(a.length, b.length); index++) if (a[index] !== b[index]) return a[index]! - b[index]!;
  return a.length - b.length;
}
