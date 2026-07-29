import Decimal from "break_infinity.js";

import { canonicalString, isStateValue, parseCanonical, quantize } from "./numeric";

export const COMMONS_SCHEMA_VERSION = 1;
export const PPM = 1_000_000;

export interface CommonsSourceWeight { readonly sourceId: string; readonly slot: string; readonly weightPpm: number; readonly forsworn: boolean }
export interface CommonsCatalog {
  readonly sourceWeights: readonly CommonsSourceWeight[];
  readonly defaultTithePpm: number; readonly minimumTithePpm: number; readonly maximumTithePpm: number;
  readonly guildHealthWeightPpm: number; readonly cohortHealthWeightPpm: number; readonly serverHealthWeightPpm: number;
  readonly collectiveWeightPpm: number; readonly collapseHealthPpm: number; readonly healthyHealthPpm: number; readonly maximumBonus: string;
  readonly healthRecoveryPpmPerHour: number; readonly healthDecayPpmPerHour: number; readonly solidarityWindowMs: number;
  readonly cohortTargetSize: number; readonly cohortMergeFloor: number; readonly npcPopulationFloor: number;
  readonly npcWeightPpm: number; readonly npcCompliancePpm: number; readonly populationTolerancePpm: number;
}

const idPattern = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;
const slots = new Set(["upgrades", "milestones", "faction", "doctrine", "commons", "trust", "event_buffs", "prestige"]);
const keys = ["schema_version", "source_weights", "default_tithe_ppm", "minimum_tithe_ppm", "maximum_tithe_ppm", "guild_health_weight_ppm", "cohort_health_weight_ppm", "server_health_weight_ppm", "collective_weight_ppm", "collapse_health_ppm", "healthy_health_ppm", "maximum_bonus", "health_recovery_ppm_per_hour", "health_decay_ppm_per_hour", "solidarity_window_ms", "cohort_target_size", "cohort_merge_floor", "npc_population_floor", "npc_weight_ppm", "npc_compliance_ppm", "population_tolerance_ppm"] as const;

export function parseCommonsCatalog(source: unknown): CommonsCatalog {
  const raw = exactObject(source, keys, "commons catalog");
  if (raw.schema_version !== COMMONS_SCHEMA_VERSION || !Array.isArray(raw.source_weights)) throw new SyntaxError("invalid commons schema");
  const ratio = (key: string): number => boundedInteger(raw[key], key, 0, PPM);
  const positive = (key: string): number => boundedInteger(raw[key], key, 1, Number.MAX_SAFE_INTEGER);
  const weights = raw.source_weights.map((item, index) => {
    const value = exactObject(item, ["source_id", "slot", "weight_ppm", "forsworn"], `source_weights[${index}]`);
    if (typeof value.source_id !== "string" || !idPattern.test(value.source_id) || typeof value.slot !== "string" || !slots.has(value.slot) || value.slot === "commons" || typeof value.forsworn !== "boolean") throw new SyntaxError("invalid commons source weight");
    return Object.freeze({ sourceId: value.source_id, slot: value.slot, weightPpm: boundedInteger(value.weight_ppm, "weight_ppm", 1, PPM), forsworn: value.forsworn });
  });
  if (new Set(weights.map((item) => item.sourceId)).size !== weights.length) throw new SyntaxError("duplicate commons source weight");
  if (typeof raw.maximum_bonus !== "string" || !parseCanonical(raw.maximum_bonus).gt(0)) throw new SyntaxError("invalid maximum_bonus");
  const catalog: CommonsCatalog = Object.freeze({
    sourceWeights: Object.freeze(weights), defaultTithePpm: ratio("default_tithe_ppm"), minimumTithePpm: ratio("minimum_tithe_ppm"), maximumTithePpm: ratio("maximum_tithe_ppm"),
    guildHealthWeightPpm: ratio("guild_health_weight_ppm"), cohortHealthWeightPpm: ratio("cohort_health_weight_ppm"), serverHealthWeightPpm: ratio("server_health_weight_ppm"),
    collectiveWeightPpm: ratio("collective_weight_ppm"), collapseHealthPpm: ratio("collapse_health_ppm"), healthyHealthPpm: ratio("healthy_health_ppm"), maximumBonus: raw.maximum_bonus,
    healthRecoveryPpmPerHour: ratio("health_recovery_ppm_per_hour"), healthDecayPpmPerHour: ratio("health_decay_ppm_per_hour"), solidarityWindowMs: positive("solidarity_window_ms"),
    cohortTargetSize: positive("cohort_target_size"), cohortMergeFloor: positive("cohort_merge_floor"), npcPopulationFloor: positive("npc_population_floor"),
    npcWeightPpm: ratio("npc_weight_ppm"), npcCompliancePpm: ratio("npc_compliance_ppm"), populationTolerancePpm: ratio("population_tolerance_ppm"),
  });
  if (catalog.minimumTithePpm > catalog.defaultTithePpm || catalog.defaultTithePpm > catalog.maximumTithePpm || catalog.guildHealthWeightPpm + catalog.cohortHealthWeightPpm + catalog.serverHealthWeightPpm !== PPM || catalog.healthyHealthPpm <= catalog.collapseHealthPpm || catalog.healthRecoveryPpmPerHour <= catalog.healthDecayPpmPerHour || catalog.cohortMergeFloor > catalog.cohortTargetSize || catalog.npcPopulationFloor < catalog.cohortMergeFloor) throw new SyntaxError("invalid commons policy relationship");
  return catalog;
}

export function enclosureIndex(catalog: CommonsCatalog, contributions: readonly { readonly sourceId: string; readonly slot: string; readonly factor: string }[]): string {
  const bySource = new Map(catalog.sourceWeights.map((item) => [item.sourceId, item]));
  const seen = new Set<string>(); let all = new Decimal(1); let clean = new Decimal(1);
  for (const contribution of contributions) {
    const weight = bySource.get(contribution.sourceId); if (!weight) continue;
    if (seen.has(contribution.sourceId) || contribution.slot !== weight.slot) throw new RangeError("invalid commons contribution");
    const factor = parseCanonical(contribution.factor); if (!isStateValue(factor) || !factor.gt(0)) throw new RangeError("invalid commons factor");
    seen.add(contribution.sourceId); const weighted = factor.pow(weight.weightPpm / PPM); all = all.mul(weighted); if (!weight.forsworn) clean = clean.mul(weighted);
  }
  let value = new Decimal(1).sub(clean.div(all)); if (value.lt(0)) value = new Decimal(0); if (value.gt(1)) value = new Decimal(1);
  return canonicalString(quantize(value));
}

export function commonsModifier(catalog: CommonsCatalog, healthPpm: number, solidarityPpm: number): string {
  boundedInteger(healthPpm, "health_ppm", 0, PPM); boundedInteger(solidarityPpm, "solidarity_ppm", 0, PPM);
  let collective = new Decimal(0);
  if (healthPpm > catalog.collapseHealthPpm) { const x = new Decimal(healthPpm - catalog.collapseHealthPpm).div(PPM - catalog.collapseHealthPpm); collective = x.mul(x.pow(0.5)); }
  const inside = new Decimal(catalog.collectiveWeightPpm).div(PPM).mul(collective).add(new Decimal(PPM - catalog.collectiveWeightPpm).div(PPM).mul(new Decimal(solidarityPpm).div(PPM)));
  return canonicalString(quantize(new Decimal(1).add(parseCanonical(catalog.maximumBonus).mul(inside))));
}

function exactObject(source: unknown, expected: readonly string[], label: string): Record<string, unknown> { if (typeof source !== "object" || source === null || Array.isArray(source)) throw new SyntaxError(`${label} must be an object`); const value = source as Record<string, unknown>; const actual = Object.keys(value).sort(); const keys = [...expected].sort(); if (actual.length !== keys.length || actual.some((key, index) => key !== keys[index])) throw new SyntaxError(`${label} fields are not exact`); return value; }
function boundedInteger(value: unknown, label: string, minimum: number, maximum: number): number { if (typeof value !== "number" || !Number.isSafeInteger(value) || value < minimum || value > maximum) throw new SyntaxError(`${label} outside integer domain`); return value; }
