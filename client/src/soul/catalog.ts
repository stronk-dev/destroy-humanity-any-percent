import { MAX_EXACT_INTEGER } from "../numeric";

export const SOUL_BANDS = ["near_zero", "hollow", "dimming", "whole"] as const;
export type SoulBand = (typeof SOUL_BANDS)[number];
export type SoulOwnerKind = "event" | "longevity" | "contract" | "fixture";
export type SoulEndingVariant = "earnest_ascension" | "training_data";

export interface SoulPolicy { readonly soul_floor: number; readonly soul_initial: number; readonly soul_max: number }
export interface SoulBandRow { readonly band_member: SoulBand; readonly min_inclusive: number; readonly max_inclusive: number; readonly human_content_locked: boolean; readonly reason_key: string }
export interface SoulDebitSource { readonly source_id: string; readonly owner_kind: SoulOwnerKind; readonly amount: number; readonly may_exhaust: boolean; readonly single_use: boolean; readonly curtain_copy_key: string }
export interface SoulRecoveryActivity { readonly activity_id: string; readonly duration_attended_ms: number; readonly recovery_amount: number; readonly reason_key: string }
export interface SoulCatalog {
  readonly schema_version: 1;
  readonly policy: SoulPolicy;
  readonly bands: readonly SoulBandRow[];
  readonly debit_sources: readonly SoulDebitSource[];
  readonly recovery_activities: readonly SoulRecoveryActivity[];
  readonly ending_policy: { readonly whole_variant: "earnest_ascension"; readonly depleted_variant: "training_data" };
}

export interface SoulDeclarations { readonly copyKeys: ReadonlySet<string>; readonly epochSeeded: boolean }

const mechanical = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;
const owners = new Set<SoulOwnerKind>(["event", "longevity", "contract", "fixture"]);

export function parseSoulCatalog(source: unknown, declarations: SoulDeclarations): SoulCatalog {
  if (declarations.copyKeys.size === 0) throw new SyntaxError("Soul copy registry is empty");
  const root = exactObject(source, ["schema_version", "policy", "bands", "debit_sources", "recovery_activities", "ending_policy"], "Soul catalog");
  if (root.schema_version !== 1) throw new SyntaxError("invalid Soul catalog version");
  const policyRow = exactObject(root.policy, ["soul_floor", "soul_initial", "soul_max"], "Soul policy");
  const floor = safeInteger(policyRow.soul_floor, 0, MAX_EXACT_INTEGER), maximum = safeInteger(policyRow.soul_max, floor, MAX_EXACT_INTEGER);
  const policy: SoulPolicy = Object.freeze({ soul_floor: floor, soul_initial: safeInteger(policyRow.soul_initial, floor, maximum), soul_max: maximum });
  if (!Array.isArray(root.bands) || root.bands.length !== SOUL_BANDS.length) throw new SyntaxError("Soul bands must be complete");
  let next = floor;
  const bands = Object.freeze(root.bands.map((item, index) => {
    const row = exactObject(item, ["band_member", "min_inclusive", "max_inclusive", "human_content_locked", "reason_key"], "Soul band");
    if (row.band_member !== SOUL_BANDS[index] || row.min_inclusive !== next || typeof row.human_content_locked !== "boolean" || row.human_content_locked !== (row.band_member === "near_zero")) throw new SyntaxError("invalid Soul band partition");
    const max = safeInteger(row.max_inclusive, next, maximum), reason = copyKey(row.reason_key, declarations);
    next = max + 1;
    return Object.freeze({ band_member: row.band_member as SoulBand, min_inclusive: row.min_inclusive as number, max_inclusive: max, human_content_locked: row.human_content_locked, reason_key: reason });
  }));
  if (bands.at(-1)?.max_inclusive !== maximum) throw new SyntaxError("Soul bands do not cover policy domain");
  if (!Array.isArray(root.debit_sources) || !Array.isArray(root.recovery_activities)) throw new SyntaxError("Soul content rows must be arrays");
  let prior = "";
  const debitSources = Object.freeze(root.debit_sources.map((item) => {
    const row = exactObject(item, ["source_id", "owner_kind", "amount", "may_exhaust", "single_use", "curtain_copy_key"], "Soul debit source");
    const sourceId = identifier(row.source_id, "Soul source");
    if (byteCompare(prior, sourceId) >= 0 || typeof row.owner_kind !== "string" || !owners.has(row.owner_kind as SoulOwnerKind) || typeof row.may_exhaust !== "boolean" || row.may_exhaust !== row.single_use || declarations.epochSeeded && row.owner_kind === "fixture") throw new SyntaxError("invalid Soul debit source");
    prior = sourceId;
    return Object.freeze({ source_id: sourceId, owner_kind: row.owner_kind as SoulOwnerKind, amount: safeInteger(row.amount, 1, MAX_EXACT_INTEGER), may_exhaust: row.may_exhaust, single_use: row.single_use as boolean, curtain_copy_key: copyKey(row.curtain_copy_key, declarations) });
  }));
  prior = "";
  const activities = Object.freeze(root.recovery_activities.map((item) => {
    const row = exactObject(item, ["activity_id", "duration_attended_ms", "recovery_amount", "reason_key"], "Soul recovery activity");
    const activityId = identifier(row.activity_id, "Soul activity");
    if (byteCompare(prior, activityId) >= 0) throw new SyntaxError("Soul activities are not byte sorted");
    prior = activityId;
    return Object.freeze({ activity_id: activityId, duration_attended_ms: safeInteger(row.duration_attended_ms, 1, MAX_EXACT_INTEGER), recovery_amount: safeInteger(row.recovery_amount, 1, MAX_EXACT_INTEGER), reason_key: copyKey(row.reason_key, declarations) });
  }));
  if (!declarations.epochSeeded && (debitSources.length === 0 || activities.length === 0)) throw new SyntaxError("fixture Soul catalog needs source and activity rows");
  const ending = exactObject(root.ending_policy, ["whole_variant", "depleted_variant"], "Soul ending policy");
  if (ending.whole_variant !== "earnest_ascension" || ending.depleted_variant !== "training_data") throw new SyntaxError("invalid Soul ending policy");
  return Object.freeze({ schema_version: 1, policy, bands, debit_sources: debitSources, recovery_activities: activities,
    ending_policy: Object.freeze({ whole_variant: "earnest_ascension", depleted_variant: "training_data" }) });
}

export function soulBand(catalog: SoulCatalog, value: number): SoulBandRow {
  safeInteger(value, catalog.policy.soul_floor, catalog.policy.soul_max);
  const band = catalog.bands.find((row) => value <= row.max_inclusive);
  if (!band) throw new RangeError("Soul value outside catalog bands");
  return band;
}

export function humanContentLocked(catalog: SoulCatalog, value: number): boolean { return soulBand(catalog, value).human_content_locked }

function copyKey(value: unknown, declarations: SoulDeclarations): string {
  const result = identifier(value, "Soul copy key");
  if (!declarations.copyKeys.has(result)) throw new SyntaxError("unknown Soul copy key");
  return result;
}

function identifier(value: unknown, label: string): string {
  if (typeof value !== "string" || !mechanical.test(value)) throw new SyntaxError(`invalid ${label}`);
  return value;
}

function safeInteger(value: unknown, minimum: number, maximum: number): number {
  if (typeof value !== "number" || !Number.isSafeInteger(value) || value < minimum || value > maximum) throw new SyntaxError("Soul integer outside exact domain");
  return value;
}

function exactObject(source: unknown, keys: readonly string[], label: string): Record<string, unknown> {
  if (source === null || typeof source !== "object" || Array.isArray(source)) throw new SyntaxError(`${label} must be an object`);
  const object = source as Record<string, unknown>, actual = Object.keys(object).sort(byteCompare), expected = [...keys].sort(byteCompare);
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError(`${label} fields are not exact`);
  return object;
}

function byteCompare(left: string, right: string): number {
  const a = new TextEncoder().encode(left), b = new TextEncoder().encode(right);
  for (let index = 0; index < Math.min(a.length, b.length); index++) if (a[index] !== b[index]) return a[index]! - b[index]!;
  return a.length - b.length;
}
