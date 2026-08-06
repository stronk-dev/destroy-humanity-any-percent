import { MAX_EXACT_INTEGER } from "../numeric";
import { PET_BEHAVIOR_EVENTS, PET_BEHAVIOR_STATES, PET_MOODS, PET_STAT_IDS, type PetBehaviorEvent, type PetBehaviorState, type PetMood, type PetStatID } from "./grammar";

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

export interface PetCatalog {
	readonly schema_version: 1 | 2;
  readonly stat_policy: { readonly grid_ms: number; readonly stats: readonly { readonly stat_id: PetStatID; readonly initial_ppm: number; readonly floor_ppm: number; readonly decay_ppm_per_grid: number }[]; readonly diminishing_threshold_ppm: number; readonly diminishing_factor_ppm: number };
	readonly actions: readonly { readonly action_id: string; readonly stat_id: PetStatID; readonly delta_ppm: number; readonly cooldown_attended_ms: number; readonly min_eligible_ppm: number; readonly soul_gate: "" | "essential" | "recovery" | "ordinary" }[];
  readonly trust_policy: { readonly initial_ppm: number; readonly neutral_ppm: number; readonly floor_ppm: number; readonly cap_ppm: number; readonly gain_ppm_per_effective_action: number; readonly decay_ppm_per_grid: number };
  readonly mood_policy: readonly PetMoodThreshold[];
  readonly behavior_policy: readonly PetBehaviorCandidate[];
}

const mechanical = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;

export function parsePetCatalog(source: unknown): PetCatalog {
  const root = exactObject(source, ["schema_version", "stat_policy", "actions", "trust_policy", "mood_policy", "behavior_policy"], "pet catalog");
	if (root.schema_version !== 1 && root.schema_version !== 2) throw new SyntaxError("invalid pet catalog version");
	const schemaVersion = root.schema_version;
  const stat = exactObject(root.stat_policy, ["grid_ms", "stats", "diminishing_threshold_ppm", "diminishing_factor_ppm"], "stat policy");
  if (!Array.isArray(stat.stats) || stat.stats.length !== PET_STAT_IDS.length) throw new SyntaxError("invalid pet stat policy");
  const seenStats = new Set<PetStatID>();
  const stats = stat.stats.map((item) => { const row = exactObject(item, ["stat_id", "initial_ppm", "floor_ppm", "decay_ppm_per_grid"], "stat row"); if (typeof row.stat_id !== "string" || !PET_STAT_IDS.includes(row.stat_id as PetStatID) || seenStats.has(row.stat_id as PetStatID)) throw new SyntaxError("invalid stat row"); const floor = safeInteger(row.floor_ppm, 0, 1_000_000); const initial = safeInteger(row.initial_ppm, floor, 1_000_000); seenStats.add(row.stat_id as PetStatID); return { stat_id: row.stat_id as PetStatID, initial_ppm: initial, floor_ppm: floor, decay_ppm_per_grid: safeInteger(row.decay_ppm_per_grid, 0, 1_000_000) }; });
  if (!Array.isArray(root.actions)) throw new SyntaxError("invalid pet actions");
  let priorAction = "";
	const actions = root.actions.map((item) => { const keys = ["action_id", "stat_id", "delta_ppm", "cooldown_attended_ms", "min_eligible_ppm"]; if (schemaVersion === 2) keys.push("soul_gate"); const row = exactObject(item, keys, "pet action"); if (typeof row.action_id !== "string" || !mechanical.test(row.action_id) || byteCompare(priorAction, row.action_id) >= 0 || typeof row.stat_id !== "string" || !PET_STAT_IDS.includes(row.stat_id as PetStatID)) throw new SyntaxError("invalid pet action"); const soulGate: "" | "essential" | "recovery" | "ordinary" = schemaVersion === 1 ? "" : row.soul_gate === "essential" || row.soul_gate === "recovery" || row.soul_gate === "ordinary" ? row.soul_gate : (() => { throw new SyntaxError("invalid pet Soul gate"); })(); priorAction = row.action_id; const statRow = stats.find((candidate) => candidate.stat_id === row.stat_id)!; return { action_id: row.action_id, stat_id: row.stat_id as PetStatID, delta_ppm: safeInteger(row.delta_ppm, 1, 1_000_000), cooldown_attended_ms: safeInteger(row.cooldown_attended_ms, 0, MAX_EXACT_INTEGER), min_eligible_ppm: safeInteger(row.min_eligible_ppm, 0, statRow.floor_ppm), soul_gate: soulGate }; });
  const trust = exactObject(root.trust_policy, ["initial_ppm", "neutral_ppm", "floor_ppm", "cap_ppm", "gain_ppm_per_effective_action", "decay_ppm_per_grid"], "trust policy");
  const floor = safeInteger(trust.floor_ppm, 0, 1_000_000), neutral = safeInteger(trust.neutral_ppm, floor, 1_000_000), initial = safeInteger(trust.initial_ppm, neutral, 1_000_000), cap = safeInteger(trust.cap_ppm, initial, 1_000_000);
  const grammar = parsePetCatalogGrammar({ mood_thresholds: root.mood_policy, behavior_candidates: root.behavior_policy });
  const transitionKeys = new Set<string>(); for (const row of grammar.behavior_candidates) { const key = `${row.from_state}\0${row.event}`; if (transitionKeys.has(key)) throw new SyntaxError("nondeterministic pet transition"); transitionKeys.add(key); }
	return { schema_version: schemaVersion, stat_policy: { grid_ms: safeInteger(stat.grid_ms, 1, MAX_EXACT_INTEGER), stats, diminishing_threshold_ppm: safeInteger(stat.diminishing_threshold_ppm, 0, 1_000_000), diminishing_factor_ppm: safeInteger(stat.diminishing_factor_ppm, 0, 1_000_000) }, actions, trust_policy: { initial_ppm: initial, neutral_ppm: neutral, floor_ppm: floor, cap_ppm: cap, gain_ppm_per_effective_action: safeInteger(trust.gain_ppm_per_effective_action, 0, 1_000_000), decay_ppm_per_grid: safeInteger(trust.decay_ppm_per_grid, 0, 1_000_000) }, mood_policy: grammar.mood_thresholds, behavior_policy: grammar.behavior_candidates };
}

export function petCatalogSupportsSoul(catalog: PetCatalog): boolean {
	return catalog.schema_version === 2 && catalog.actions.every((row) => row.soul_gate === "essential" || row.soul_gate === "ordinary" || row.soul_gate === "recovery");
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
    if (seenMoods.size === 0 && floor !== 0 || floor <= priorFloor) throw new SyntaxError("mood floors must ascend from zero");
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
