import Decimal from "break_infinity.js";

import { MULTIPLIER_SLOT_ORDER, type EconomyCatalog, type MultiplierSlot } from "./economy-kernel";
import { substream } from "./combat/rng";
import { canonicalString, MAX_EXACT_INTEGER, quantize } from "./numeric";

export const FISCAL_SCHEMA_VERSION = 1;
export const FISCAL_EARLY_SUBSTREAM = "fiscal.early_harvest.v1";
const PPM = 1_000_000;
const idPattern = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;

export interface FiscalCatalog {
  readonly clock: { readonly earlyMs: number; readonly guaranteedMs: number; readonly autoMs: number; readonly earlySuccessPpm: number };
  readonly credit: { readonly creditPerPeriod: number; readonly hardcap: number; readonly hardcapReasonKey: string };
  readonly hoard: { readonly ppmPerCredit: number; readonly capCredits: number; readonly slot: MultiplierSlot; readonly sourceId: string; readonly target: "all" };
  readonly generatorLevelRows: readonly { readonly generatorId: string; readonly ppmPerLevel: number; readonly levelHardcap: number; readonly hardcapReasonKey: string; readonly slot: MultiplierSlot; readonly sourceId: string }[];
  readonly unlockRows: readonly { readonly unlockId: string; readonly cost: number }[];
}

export function loadFiscalCatalog(source: unknown, economy: EconomyCatalog): FiscalCatalog {
  const root = exactObject(source, ["schema_version", "clock_policy", "credit_policy", "hoard_policy", "generator_level_rows", "unlock_rows"], "fiscal catalog");
  if (root.schema_version !== FISCAL_SCHEMA_VERSION) throw new SyntaxError("invalid fiscal schema version");
  const clock = exactObject(root.clock_policy, ["early_ms", "guaranteed_ms", "auto_ms", "early_success_ppm"], "clock_policy");
  const earlyMs = positiveSafe(clock.early_ms); const guaranteedMs = positiveSafe(clock.guaranteed_ms); const autoMs = positiveSafe(clock.auto_ms);
  const earlySuccessPpm = ppm(clock.early_success_ppm, true);
  if (earlyMs > guaranteedMs || guaranteedMs > autoMs) throw new SyntaxError("fiscal clock thresholds are not ordered");
  const credit = exactObject(root.credit_policy, ["credit_per_period", "hardcap", "hardcap_reason_key"], "credit_policy");
  const creditPerPeriod = positiveSafe(credit.credit_per_period); const hardcap = positiveSafe(credit.hardcap); const hardcapReasonKey = mechanical(credit.hardcap_reason_key);
  if (creditPerPeriod > hardcap) throw new SyntaxError("fiscal mint exceeds hardcap");
  const hoardRaw = exactObject(root.hoard_policy, ["ppm_per_credit", "cap_credits", "slot", "source_id", "target"], "hoard_policy");
  const hoard = Object.freeze({ ppmPerCredit: ppm(hoardRaw.ppm_per_credit, false), capCredits: positiveSafe(hoardRaw.cap_credits), slot: slot(hoardRaw.slot), sourceId: mechanical(hoardRaw.source_id), target: literalAll(hoardRaw.target) });
  if (hoard.capCredits > hardcap || !validDeclaration(economy, hoard.sourceId, hoard.slot, hoard.target)) throw new SyntaxError("invalid fiscal hoard declaration");
  if (!Array.isArray(root.generator_level_rows) || root.generator_level_rows.length === 0 || !Array.isArray(root.unlock_rows) || root.unlock_rows.length === 0) throw new SyntaxError("fiscal rows must be non-empty arrays");
  let previous = "";
  const generatorLevelRows = root.generator_level_rows.map((value, index) => {
    const raw = exactObject(value, ["generator_id", "ppm_per_level", "level_hardcap", "hardcap_reason_key", "slot", "source_id"], `generator_level_rows[${index}]`);
    const row = Object.freeze({ generatorId: mechanical(raw.generator_id), ppmPerLevel: ppm(raw.ppm_per_level, false), levelHardcap: positiveSafe(raw.level_hardcap), hardcapReasonKey: mechanical(raw.hardcap_reason_key), slot: slot(raw.slot), sourceId: mechanical(raw.source_id) });
    if (previous !== "" && byteCompare(previous, row.generatorId) >= 0) throw new SyntaxError("generator rows must be byte-sorted and unique"); previous = row.generatorId;
    if (!economy.generatorClasses.some((candidate) => candidate.id === row.generatorId) || !validDeclaration(economy, row.sourceId, row.slot, row.generatorId)) throw new SyntaxError("invalid fiscal generator declaration");
    return row;
  });
  previous = "";
  const unlockRows = root.unlock_rows.map((value, index) => {
    const raw = exactObject(value, ["unlock_id", "cost"], `unlock_rows[${index}]`);
    const row = Object.freeze({ unlockId: mechanical(raw.unlock_id), cost: positiveSafe(raw.cost) });
    if (row.cost > hardcap || previous !== "" && byteCompare(previous, row.unlockId) >= 0) throw new SyntaxError("unlock rows must be byte-sorted, unique, and affordable within the hardcap"); previous = row.unlockId;
    return row;
  });
  return Object.freeze({
    clock: Object.freeze({ earlyMs, guaranteedMs, autoMs, earlySuccessPpm }),
    credit: Object.freeze({ creditPerPeriod, hardcap, hardcapReasonKey }), hoard,
    generatorLevelRows: Object.freeze(generatorLevelRows), unlockRows: Object.freeze(unlockRows),
  });
}

export function fiscalHoardFactor(catalog: FiscalCatalog, credit: number): string {
  if (!Number.isSafeInteger(credit) || credit < 0 || credit > catalog.credit.hardcap) throw new RangeError("invalid fiscal credit");
  return ppmFactor(Math.min(credit, catalog.hoard.capCredits), catalog.hoard.ppmPerCredit);
}

export function fiscalGeneratorFactor(catalog: FiscalCatalog, generatorId: string, level: number): string {
  const row = catalog.generatorLevelRows.find((candidate) => candidate.generatorId === generatorId);
  if (!row || !Number.isSafeInteger(level) || level < 0 || level > row.levelHardcap) throw new RangeError("invalid fiscal generator level");
  return ppmFactor(level, row.ppmPerLevel);
}

export function fiscalGeneratorCost(catalog: FiscalCatalog, generatorId: string, current: number, levels: number): number {
  const row = catalog.generatorLevelRows.find((candidate) => candidate.generatorId === generatorId);
  if (!row || !Number.isSafeInteger(current) || !Number.isSafeInteger(levels) || current < 0 || levels <= 0 || current > row.levelHardcap || levels > row.levelHardcap - current) throw new RangeError("invalid fiscal level purchase");
  const result = BigInt(levels) * (2n * BigInt(current) + BigInt(levels) + 1n) / 2n;
  if (result > BigInt(MAX_EXACT_INTEGER)) throw new RangeError("fiscal level cost exceeds safe integer");
  return Number(result);
}

export async function fiscalEarlyHarvestDraw(founderId: string, sequence: number): Promise<number> {
  if (founderId.length === 0 || !Number.isSafeInteger(sequence) || sequence < 0) throw new RangeError("invalid fiscal draw identity");
  const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", new TextEncoder().encode(founderId)));
  let seed = 0n; for (let index = 0; index < 8; index++) seed = (seed << 8n) | BigInt(digest[index]!);
  return Number(substream(seed ^ BigInt(sequence), FISCAL_EARLY_SUBSTREAM).bound(BigInt(PPM)));
}

function ppmFactor(count: number, perCountPpm: number): string {
  const product = BigInt(count) * BigInt(perCountPpm);
  return canonicalString(quantize(new Decimal(product.toString()).div(PPM).add(1)));
}
function validDeclaration(economy: EconomyCatalog, sourceId: string, sourceSlot: MultiplierSlot, target: string): boolean { return economy.multiplierSources.some((value) => value.id === sourceId && value.slot === sourceSlot && value.target === target && value.provider === "fiscal"); }
function positiveSafe(value: unknown): number { if (typeof value !== "number" || !Number.isSafeInteger(value) || value <= 0) throw new SyntaxError("expected positive safe integer"); return value; }
function ppm(value: unknown, allowZero: boolean): number { if (typeof value !== "number" || !Number.isSafeInteger(value) || value > PPM || value < (allowZero ? 0 : 1)) throw new SyntaxError("invalid ppm"); return value; }
function mechanical(value: unknown): string { if (typeof value !== "string" || !idPattern.test(value)) throw new SyntaxError("invalid mechanical id"); return value; }
function slot(value: unknown): MultiplierSlot { if (typeof value !== "string" || !MULTIPLIER_SLOT_ORDER.includes(value as MultiplierSlot)) throw new SyntaxError("invalid multiplier slot"); return value as MultiplierSlot; }
function literalAll(value: unknown): "all" { if (value !== "all") throw new SyntaxError("target must be all"); return value; }
function isRecord(source: unknown): source is Record<string, unknown> { return typeof source === "object" && source !== null && !Array.isArray(source); }
function exactObject(source: unknown, keys: readonly string[], label: string): Record<string, unknown> { if (!isRecord(source)) throw new SyntaxError(`${label} must be an object`); const actual = Object.keys(source).sort(byteCompare); const expected = [...keys].sort(byteCompare); if (actual.length !== expected.length || actual.some((value, index) => value !== expected[index])) throw new SyntaxError(`${label} fields are not exact`); return source; }
function byteCompare(left: string, right: string): number { const a = new TextEncoder().encode(left); const b = new TextEncoder().encode(right); for (let index = 0; index < Math.min(a.length, b.length); index++) if (a[index] !== b[index]) return a[index]! - b[index]!; return a.length - b.length; }
