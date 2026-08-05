import { MAX_EXACT_INTEGER } from "../numeric";
import { integrateFixedGrid } from "../fixed-grid";
import type { PetCatalog } from "./catalog";
import {
  PET_BEHAVIOR_QUEUE_HARDCAP,
  PET_STAT_IDS,
  PET_STATUS_BANDS,
  type PetBehaviorEvent,
  type PetBehaviorState,
  type PetCareRejectionDetail,
  type PetMood,
  type PetStatID,
  type PetStatusBand,
} from "./grammar";
import type { PetBehaviorQueueEntry, PetCareState } from "./state";

export interface PetCareTransitionInput {
  readonly action_id: string;
  readonly attended_before_ms: number;
  readonly attended_after_ms: number;
}

export interface PetCareTransitionResult {
  readonly state: PetCareState;
  readonly applied: boolean;
  readonly rejection_detail: PetCareRejectionDetail | "";
  readonly stat_id: PetStatID | "";
  readonly before_ppm: number;
  readonly applied_ppm: number;
  readonly after_ppm: number;
  readonly trust_before_ppm: number;
  readonly trust_after_ppm: number;
  readonly mood: PetMood;
  readonly status_band: PetStatusBand;
  readonly status_changed: boolean;
  readonly next_eligible_attended_ms: number;
  readonly eligible_action_ids: readonly string[];
}

type MutableCareState = {
  stats_ppm: Record<PetStatID, number>;
  stat_decay_remainders_ppm: Record<PetStatID, number>;
  cooldown_until_attended_ms: Record<string, number>;
  trust_ppm: number;
  trust_decay_remainder_ppm: number;
  evaluated_through_attended_ms: number;
  behavior_state: PetBehaviorState;
  behavior_entered_at_attended_ms: number;
  behavior_queue: PetBehaviorQueueEntry[];
  behavior_prng_cursor: number;
};

export function applyPetCareTransition(source: PetCareState, catalog: PetCatalog, input: PetCareTransitionInput): PetCareTransitionResult {
  if (input.attended_before_ms !== source.evaluated_through_attended_ms || !safe(input.attended_after_ms) || input.attended_after_ms < input.attended_before_ms) {
    throw new RangeError("stale pet care transition");
  }
  validateCatalogState(source, catalog);
  const state = cloneState(source);
  const statusBefore = careStatus(state, catalog).status_band;
  applyDecay(state, catalog, input.attended_after_ms - input.attended_before_ms);
  advanceBehavior(state, catalog, input.attended_before_ms, input.attended_after_ms);
  state.evaluated_through_attended_ms = input.attended_after_ms;
  const action = catalog.actions.find((row) => row.action_id === input.action_id);
  const base = {
    state, applied: false, rejection_detail: "unknown_action" as PetCareRejectionDetail | "",
    stat_id: "" as PetStatID | "", before_ppm: 0, applied_ppm: 0, after_ppm: 0,
    trust_before_ppm: state.trust_ppm, trust_after_ppm: state.trust_ppm,
    next_eligible_attended_ms: 0,
  };
  if (!action) return finish(base, statusBefore, catalog, "care_rejected");
  base.stat_id = action.stat_id;
  base.before_ppm = state.stats_ppm[action.stat_id];
  base.after_ppm = base.before_ppm;
  base.next_eligible_attended_ms = state.cooldown_until_attended_ms[action.action_id] ?? 0;
  if (base.next_eligible_attended_ms > input.attended_after_ms) {
    base.rejection_detail = "cooldown";
    return finish(base, statusBefore, catalog, "care_rejected");
  }
  if (base.before_ppm < action.min_eligible_ppm) {
    base.rejection_detail = "ineligible";
    return finish(base, statusBefore, catalog, "care_rejected");
  }
  const effective = base.before_ppm >= catalog.stat_policy.diminishing_threshold_ppm
    ? Number(BigInt(action.delta_ppm) * BigInt(catalog.stat_policy.diminishing_factor_ppm) / 1_000_000n)
    : action.delta_ppm;
  base.applied_ppm = Math.min(effective, 1_000_000 - base.before_ppm);
  if (base.applied_ppm === 0) {
    base.rejection_detail = "saturated";
    return finish(base, statusBefore, catalog, "care_rejected");
  }
  if (input.attended_after_ms > MAX_EXACT_INTEGER - action.cooldown_attended_ms) throw new RangeError("pet cooldown overflow");
  base.applied = true;
  base.rejection_detail = "";
  base.after_ppm = base.before_ppm + base.applied_ppm;
  state.stats_ppm[action.stat_id] = base.after_ppm;
  state.cooldown_until_attended_ms[action.action_id] = input.attended_after_ms + action.cooldown_attended_ms;
  base.next_eligible_attended_ms = state.cooldown_until_attended_ms[action.action_id]!;
  state.trust_ppm += Math.min(catalog.trust_policy.gain_ppm_per_effective_action, catalog.trust_policy.cap_ppm - state.trust_ppm);
  base.trust_after_ppm = state.trust_ppm;
  return finish(base, statusBefore, catalog, "care_applied");
}

function applyDecay(state: MutableCareState, catalog: PetCatalog, elapsed: number): void {
  for (const row of catalog.stat_policy.stats) {
    const integrated = integrateFixedGrid(elapsed, row.decay_ppm_per_grid, state.stat_decay_remainders_ppm[row.stat_id], catalog.stat_policy.grid_ms);
    const distance = state.stats_ppm[row.stat_id] - row.floor_ppm;
    if (integrated.whole >= BigInt(distance)) {
      state.stats_ppm[row.stat_id] = row.floor_ppm;
      state.stat_decay_remainders_ppm[row.stat_id] = 0;
    } else {
      state.stats_ppm[row.stat_id] -= Number(integrated.whole);
      state.stat_decay_remainders_ppm[row.stat_id] = integrated.remainder;
    }
  }
  const integrated = integrateFixedGrid(elapsed, catalog.trust_policy.decay_ppm_per_grid, state.trust_decay_remainder_ppm, catalog.stat_policy.grid_ms);
  const distance = Math.abs(state.trust_ppm - catalog.trust_policy.neutral_ppm);
  if (integrated.whole >= BigInt(distance)) {
    state.trust_ppm = catalog.trust_policy.neutral_ppm;
    state.trust_decay_remainder_ppm = 0;
  } else if (state.trust_ppm > catalog.trust_policy.neutral_ppm) {
    state.trust_ppm -= Number(integrated.whole);
    state.trust_decay_remainder_ppm = integrated.remainder;
  } else if (state.trust_ppm < catalog.trust_policy.neutral_ppm) {
    state.trust_ppm += Number(integrated.whole);
    state.trust_decay_remainder_ppm = integrated.remainder;
  } else state.trust_decay_remainder_ppm = 0;
}

function finish(base: Omit<PetCareTransitionResult, "mood" | "status_band" | "status_changed" | "eligible_action_ids">,
  prior: PetStatusBand, catalog: PetCatalog, event: PetBehaviorEvent): PetCareTransitionResult {
  const state = base.state as MutableCareState;
  applyBehaviorEvent(state, catalog, event, state.evaluated_through_attended_ms);
  const status = careStatus(state, catalog);
  return { ...base, ...status, status_changed: status.status_band !== prior, eligible_action_ids: eligibleCareActions(state, catalog, state.evaluated_through_attended_ms) };
}

export function careStatus(state: PetCareState, catalog: PetCatalog): { readonly status_band: PetStatusBand; readonly mood: PetMood } {
  const scalar = Math.min(...PET_STAT_IDS.map((id) => state.stats_ppm[id]));
  let selected = 0;
  for (let index = 0; index < catalog.mood_policy.length; index++) if (catalog.mood_policy[index]!.floor_ppm <= scalar) selected = index;
  return { status_band: PET_STATUS_BANDS[selected]!, mood: catalog.mood_policy[selected]!.mood_member };
}

export function eligibleCareActions(state: PetCareState, catalog: PetCatalog, attendedMs: number): readonly string[] {
  return catalog.actions.filter((action) => (state.cooldown_until_attended_ms[action.action_id] ?? 0) <= attendedMs &&
    state.stats_ppm[action.stat_id] >= action.min_eligible_ppm && state.stats_ppm[action.stat_id] < 1_000_000).map((action) => action.action_id);
}

function advanceBehavior(state: MutableCareState, catalog: PetCatalog, before: number, after: number): void {
  drainDue(state, before);
  let nextGrid = nextGridBoundary(before, catalog.stat_policy.grid_ms);
  const seen = new Map<string, { readonly at: number; readonly entered: number }>();
  while (nextGrid !== null && nextGrid <= after) {
    if (!hasCandidate(catalog, state.behavior_state, "grid_tick")) {
      const due = state.behavior_queue[0]?.due_attended_ms;
      if (due === undefined || due > after) break;
      nextGrid = gridAtOrAfter(due, catalog.stat_policy.grid_ms);
      if (nextGrid === null || nextGrid > after) break;
    }
    drainDue(state, nextGrid);
    applyBehaviorEvent(state, catalog, "grid_tick", nextGrid);
    const signature = cycleSignature(state, nextGrid), prior = seen.get(signature);
    if (prior !== undefined) {
      const cycle = nextGrid - prior.at;
      if (cycle > 0) {
        let cycles = Math.floor((after - nextGrid) / cycle);
        const shiftEntered = state.behavior_entered_at_attended_ms - prior.entered === cycle;
        if (state.behavior_entered_at_attended_ms !== prior.entered && !shiftEntered) throw new RangeError("noncanonical pet behavior cycle");
        for (const entry of state.behavior_queue) cycles = Math.min(cycles, Math.floor((MAX_EXACT_INTEGER - entry.due_attended_ms) / cycle));
        if (shiftEntered) cycles = Math.min(cycles, Math.floor((MAX_EXACT_INTEGER - state.behavior_entered_at_attended_ms) / cycle));
        if (cycles > 0) {
          const shift = cycles * cycle;
          if (shiftEntered) state.behavior_entered_at_attended_ms += shift;
          state.behavior_queue = state.behavior_queue.map((entry) => ({ ...entry, due_attended_ms: entry.due_attended_ms + shift }));
          nextGrid += shift;
        }
      }
    } else seen.set(signature, { at: nextGrid, entered: state.behavior_entered_at_attended_ms });
    nextGrid = nextGrid > MAX_EXACT_INTEGER - catalog.stat_policy.grid_ms ? null : nextGrid + catalog.stat_policy.grid_ms;
  }
  drainDue(state, after);
}

function applyBehaviorEvent(state: MutableCareState, catalog: PetCatalog, event: PetBehaviorEvent, at: number): void {
  const row = catalog.behavior_policy.find((candidate) => candidate.from_state === state.behavior_state && candidate.event === event);
  if (!row) return;
  const dueWide = BigInt(at) + BigInt(row.duration_grid_ticks) * BigInt(catalog.stat_policy.grid_ms);
  if (dueWide > BigInt(MAX_EXACT_INTEGER)) throw new RangeError("pet behavior due overflow");
  const queue = state.behavior_queue.filter((entry) => entry.behavior_id !== row.to_state);
  if (queue.length >= PET_BEHAVIOR_QUEUE_HARDCAP) throw new RangeError("pet behavior queue full");
  queue.push({ behavior_id: row.to_state, due_attended_ms: Number(dueWide) });
  queue.sort(compareQueue);
  state.behavior_queue = queue;
}

function drainDue(state: MutableCareState, through: number): void {
  state.behavior_queue.sort(compareQueue);
  let index = 0;
  while (index < state.behavior_queue.length && state.behavior_queue[index]!.due_attended_ms <= through) {
    const entry = state.behavior_queue[index]!;
    state.behavior_state = entry.behavior_id as PetBehaviorState;
    state.behavior_entered_at_attended_ms = entry.due_attended_ms;
    index++;
  }
  state.behavior_queue = state.behavior_queue.slice(index);
}

function compareQueue(left: PetBehaviorQueueEntry, right: PetBehaviorQueueEntry): number {
  return left.due_attended_ms - right.due_attended_ms || byteCompare(left.behavior_id, right.behavior_id);
}

function hasCandidate(catalog: PetCatalog, state: PetBehaviorState, event: PetBehaviorEvent): boolean {
  return catalog.behavior_policy.some((row) => row.from_state === state && row.event === event);
}

function nextGridBoundary(value: number, grid: number): number | null {
  const result = (Math.floor(value / grid) + 1) * grid;
  return Number.isSafeInteger(result) && result <= MAX_EXACT_INTEGER ? result : null;
}

function gridAtOrAfter(value: number, grid: number): number | null {
  const result = Math.ceil(value / grid) * grid;
  return Number.isSafeInteger(result) && result <= MAX_EXACT_INTEGER ? result : null;
}

function cycleSignature(state: MutableCareState, at: number): string {
  return `${state.behavior_state}${state.behavior_queue.map((entry) => `|${entry.behavior_id}:${entry.due_attended_ms - at}`).join("")}`;
}

function cloneState(state: PetCareState): MutableCareState {
  return { ...state, stats_ppm: { ...state.stats_ppm }, stat_decay_remainders_ppm: { ...state.stat_decay_remainders_ppm },
    cooldown_until_attended_ms: { ...state.cooldown_until_attended_ms }, behavior_queue: state.behavior_queue.map((entry) => ({ ...entry })) };
}

function validateCatalogState(state: PetCareState, catalog: PetCatalog): void {
  if (state.behavior_prng_cursor !== 0 || state.trust_decay_remainder_ppm >= catalog.stat_policy.grid_ms ||
    PET_STAT_IDS.some((id) => state.stat_decay_remainders_ppm[id] >= catalog.stat_policy.grid_ms)) throw new RangeError("invalid pet state for catalog");
}

function safe(value: number): boolean {
  return Number.isSafeInteger(value) && value >= 0 && value <= MAX_EXACT_INTEGER;
}

function byteCompare(left: string, right: string): number {
  const a = new TextEncoder().encode(left), b = new TextEncoder().encode(right);
  for (let index = 0; index < Math.min(a.length, b.length); index++) if (a[index] !== b[index]) return a[index]! - b[index]!;
  return a.length - b.length;
}
