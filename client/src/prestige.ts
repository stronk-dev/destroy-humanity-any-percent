import Decimal from "break_infinity.js";

import { MAX_EXACT_INTEGER, parseCanonical } from "./numeric";

export interface PrestigePolicy {
  readonly valueResourceId: string; readonly threshold: string; readonly exitModifiersPpm: Readonly<Record<string, number>>;
  readonly collapseRouteKnowledge: number; readonly catchupCeilingMs: number; readonly offerDurationMs: number;
  readonly declineDriftPpm: number; readonly spawnGatePpm: readonly number[]; readonly advisorPerRunPpm: number; readonly advisorCapPpm: number;
}

export function parsePrestigePolicy(source: unknown): PrestigePolicy {
  if (typeof source !== "object" || source === null || Array.isArray(source)) throw new SyntaxError("prestige policy must be an object");
  const raw = source as Record<string, unknown>;
  const keys = ["schema_version", "value_resource_id", "threshold", "exit_modifiers_ppm", "collapse_route_knowledge", "catchup_ceiling_ms", "offer_duration_ms", "decline_drift_ppm", "spawn_gate_ppm", "advisor_per_run_ppm", "advisor_cap_ppm"].sort();
  if (Object.keys(raw).sort().join("\0") !== keys.join("\0") || raw.schema_version !== 1 || typeof raw.value_resource_id !== "string" ||
      !/^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/.test(raw.value_resource_id) || typeof raw.threshold !== "string" || !parseCanonical(raw.threshold).gt(0) ||
      !safe(raw.collapse_route_knowledge, 0, MAX_EXACT_INTEGER) || !safe(raw.catchup_ceiling_ms, 1, 86_400_000) ||
      !safe(raw.offer_duration_ms, 1, MAX_EXACT_INTEGER) || !safe(raw.decline_drift_ppm, 0, 1_000_000) ||
      !safe(raw.advisor_per_run_ppm, 0, 1_000_000) || !safe(raw.advisor_cap_ppm, 0, 1_000_000) || !Array.isArray(raw.spawn_gate_ppm) ||
      raw.spawn_gate_ppm.length !== 10 || raw.spawn_gate_ppm.some((value) => !safe(value, 0, 1_000_000)) ||
      typeof raw.exit_modifiers_ppm !== "object" || raw.exit_modifiers_ppm === null || Array.isArray(raw.exit_modifiers_ppm)) throw new SyntaxError("invalid prestige policy");
  const modifiers = raw.exit_modifiers_ppm as Record<string, unknown>; const kinds = ["acquihire", "acquisition", "collapse", "ipo", "scripted_first"];
  if (Object.keys(modifiers).sort().join("\0") !== kinds.join("\0") || kinds.some((kind) => !safe(modifiers[kind], 0, MAX_EXACT_INTEGER))) throw new SyntaxError("invalid exit modifiers");
  return Object.freeze({ valueResourceId: raw.value_resource_id, threshold: raw.threshold, exitModifiersPpm: Object.freeze(modifiers as Record<string, number>),
    collapseRouteKnowledge: raw.collapse_route_knowledge, catchupCeilingMs: raw.catchup_ceiling_ms, offerDurationMs: raw.offer_duration_ms,
    declineDriftPpm: raw.decline_drift_ppm, spawnGatePpm: Object.freeze(raw.spawn_gate_ppm as number[]), advisorPerRunPpm: raw.advisor_per_run_ppm, advisorCapPpm: raw.advisor_cap_ppm });
}

function safe(value: unknown, min: number, max: number): value is number { return typeof value === "number" && Number.isSafeInteger(value) && value >= min && value <= max; }

export function reputationLevel(lifetimeValueSource: string, thresholdSource: string): number {
  const lifetimeValue = parseCanonical(lifetimeValueSource);
  const threshold = parseCanonical(thresholdSource);
  if (lifetimeValue.lt(0) || !threshold.gt(0)) throw new RangeError("invalid prestige values");
  const ratio = Decimal.fromMantissaExponent(
    lifetimeValue.mantissa / threshold.mantissa,
    lifetimeValue.exponent - threshold.exponent,
  ).normalize();
  if (ratio.lt(1)) return 0;
  let low = 1;
  let high = MAX_EXACT_INTEGER;
  while (low < high) {
    const mid = low + Math.floor((high - low + 1) / 2);
    const candidate = new Decimal(mid);
    const cube = candidate.mul(candidate).mul(candidate);
    if (cube.lte(ratio)) low = mid;
    else high = mid - 1;
  }
  return low;
}

export function reputationDelta(
  lifetimeValue: string,
  threshold: string,
  currentLevel: number,
  modifierPPM: number,
): number {
  if (
    !Number.isSafeInteger(currentLevel) ||
    currentLevel < 0 ||
    !Number.isSafeInteger(modifierPPM) ||
    modifierPPM < 0
  ) {
    throw new RangeError("invalid prestige integers");
  }
  const delta = Math.max(0, reputationLevel(lifetimeValue, threshold) - currentLevel);
  const exact = (BigInt(delta) * BigInt(modifierPPM)) / 1_000_000n;
  const cap = BigInt(MAX_EXACT_INTEGER);
  if (exact >= cap) return MAX_EXACT_INTEGER;
  return Number(exact);
}

export function moralReseed(notoriety: number): number {
  if (!Number.isSafeInteger(notoriety) || notoriety < 0) throw new RangeError("invalid notoriety");
  if (notoriety >= 100) return 55;
  return Math.max(55, Math.min(90, 90 - Math.floor((notoriety * 35) / 100)));
}
