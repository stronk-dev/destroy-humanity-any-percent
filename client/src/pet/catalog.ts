import { MAX_EXACT_INTEGER } from "../numeric";
import { PET_BEHAVIOR_EVENTS, PET_BEHAVIOR_STATES, PET_MOODS, type PetBehaviorEvent, type PetBehaviorState, type PetMood } from "./grammar";

export interface PetMoodThreshold {
  readonly mood_member: PetMood;
  readonly floor_ppm: number;
}

export interface PetBehaviorCandidate {
  readonly from_state: PetBehaviorState;
  readonly event: PetBehaviorEvent;
  readonly to_state: PetBehaviorState;
  readonly duration_grid_ticks: number;
}

export interface PetCatalogGrammar {
  readonly mood_thresholds: readonly PetMoodThreshold[];
  readonly behavior_candidates: readonly PetBehaviorCandidate[];
}

export function parsePetCatalogGrammar(source: unknown): PetCatalogGrammar {
  const root = exactObject(source, ["mood_thresholds", "behavior_candidates"], "pet catalog grammar");
  if (!Array.isArray(root.mood_thresholds) || root.mood_thresholds.length !== PET_MOODS.length || !Array.isArray(root.behavior_candidates)) {
    throw new SyntaxError("invalid pet catalog rows");
  }
  const seenMoods = new Set<PetMood>();
  let priorFloor = -1;
  const thresholds = root.mood_thresholds.map((source) => {
    const row = exactObject(source, ["mood_member", "floor_ppm"], "mood threshold");
    if (typeof row.mood_member !== "string" || !PET_MOODS.includes(row.mood_member as PetMood) || seenMoods.has(row.mood_member as PetMood)) {
      throw new SyntaxError("invalid mood member");
    }
    const floor = safeInteger(row.floor_ppm, 0, 1_000_000);
    if (floor <= priorFloor) throw new SyntaxError("mood floors must ascend");
    priorFloor = floor;
    seenMoods.add(row.mood_member as PetMood);
    return { mood_member: row.mood_member as PetMood, floor_ppm: floor };
  });
  const seenCandidates = new Set<string>();
  const candidates = root.behavior_candidates.map((source) => {
    const row = exactObject(source, ["from_state", "event", "to_state", "duration_grid_ticks"], "behavior candidate");
    if (typeof row.from_state !== "string" || !PET_BEHAVIOR_STATES.includes(row.from_state as PetBehaviorState) ||
      typeof row.event !== "string" || !PET_BEHAVIOR_EVENTS.includes(row.event as PetBehaviorEvent) ||
      typeof row.to_state !== "string" || !PET_BEHAVIOR_STATES.includes(row.to_state as PetBehaviorState)) {
      throw new SyntaxError("invalid behavior candidate enum");
    }
    const key = `${row.from_state}\u0000${row.event}\u0000${row.to_state}`;
    if (seenCandidates.has(key)) throw new SyntaxError("duplicate behavior candidate");
    seenCandidates.add(key);
    return {
      from_state: row.from_state as PetBehaviorState,
      event: row.event as PetBehaviorEvent,
      to_state: row.to_state as PetBehaviorState,
      duration_grid_ticks: safeInteger(row.duration_grid_ticks, 1, MAX_EXACT_INTEGER),
    };
  });
  return { mood_thresholds: thresholds, behavior_candidates: candidates };
}

function exactObject(source: unknown, keys: readonly string[], label: string): Record<string, unknown> {
  if (typeof source !== "object" || source === null || Array.isArray(source)) throw new SyntaxError(`${label} must be an object`);
  const value = source as Record<string, unknown>;
  const actual = Object.keys(value).sort(byteCompare);
  const expected = [...keys].sort(byteCompare);
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError(`${label} fields are not exact`);
  return value;
}

function safeInteger(value: unknown, minimum: number, maximum: number): number {
  if (typeof value !== "number" || !Number.isSafeInteger(value) || value < minimum || value > maximum) throw new SyntaxError("integer outside pet catalog domain");
  return value;
}

function byteCompare(left: string, right: string): number {
  const a = new TextEncoder().encode(left); const b = new TextEncoder().encode(right);
  for (let index = 0; index < Math.min(a.length, b.length); index++) if (a[index] !== b[index]) return a[index]! - b[index]!;
  return a.length - b.length;
}
