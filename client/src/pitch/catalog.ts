import { MAX_EXACT_INTEGER, parseCanonical } from "../numeric";

export const PITCH_SCHEMA_VERSION = 1 as const;

export type PitchEffect =
  | Readonly<{ kind: "flat_add"; amount: string }>
  | Readonly<{ kind: "card_factor"; factor: string }>
  | Readonly<{ kind: "shape_factor"; shape: "pair" | "full_hand"; factor: string }>
  | Readonly<{ kind: "chain_factor"; partner_hack_id: string; factor: string }>;

export interface PitchCard { readonly card_id: string; readonly base_metric: string; readonly copies: 2; readonly copy_key: string }
export interface PitchHack { readonly hack_id: string; readonly price: number; readonly draft_weight: number; readonly effect: PitchEffect; readonly copy_key: string }
export interface PitchFundingRow { readonly round: number; readonly funding_target: string }
export interface PitchPolicy {
  readonly hand_size: 7; readonly play_size: 4; readonly hands_per_round: 3; readonly hack_slots: 4;
  readonly start_currency: 4; readonly shop_size: 3; readonly round_clear_currency: 3;
  readonly hand_size_reason_key: string; readonly play_size_reason_key: string;
  readonly hands_per_round_reason_key: string; readonly hack_slots_reason_key: string;
  readonly rounds_reason_key: string; readonly exponent_reason_key: "cap.pitch_exponent";
  readonly best_exponent_hardcap: 1_000_000;
}
export interface PitchCatalog {
  readonly schema_version: 1;
  readonly policy: PitchPolicy;
  readonly metric_cards: readonly PitchCard[];
  readonly growth_hacks: readonly PitchHack[];
  readonly funding_curve: readonly PitchFundingRow[];
}

const mechanical = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;

export function parsePitchCatalog(source: unknown, declaredCopyKeys?: ReadonlySet<string>): PitchCatalog {
  const root = exactObject(source, ["schema_version", "policy", "metric_cards", "growth_hacks", "funding_curve"], "Pitch catalog");
  if (root.schema_version !== PITCH_SCHEMA_VERSION) throw new SyntaxError("invalid Pitch schema version");
  const policy = parsePolicy(root.policy, declaredCopyKeys);
  if (!Array.isArray(root.metric_cards) || root.metric_cards.length !== 12 || !Array.isArray(root.growth_hacks) || root.growth_hacks.length !== 8 ||
    !Array.isArray(root.funding_curve) || root.funding_curve.length !== 8) throw new SyntaxError("invalid Pitch row cardinality");
  let prior = "";
  const cards = root.metric_cards.map((source) => {
    const row = exactObject(source, ["card_id", "base_metric", "copies", "copy_key"], "Pitch card");
    const id = mechanicalString(row.card_id, "card id");
    if (prior !== "" && byteCompare(prior, id) >= 0 || row.copies !== 2 || row.copy_key !== `pitch.card.${id}` ||
      declaredCopyKeys && !declaredCopyKeys.has(row.copy_key)) throw new SyntaxError("invalid Pitch card");
    const metric = canonicalNonnegative(row.base_metric, "base metric");
    prior = id;
    return Object.freeze({ card_id: id, base_metric: metric, copies: 2 as const, copy_key: row.copy_key });
  });
  prior = "";
  const hacks = root.growth_hacks.map((source) => {
    const row = exactObject(source, ["hack_id", "price", "draft_weight", "effect", "copy_key"], "Pitch hack");
    const id = mechanicalString(row.hack_id, "hack id");
    if (prior !== "" && byteCompare(prior, id) >= 0 || row.copy_key !== `pitch.hack.${id}` || declaredCopyKeys && !declaredCopyKeys.has(row.copy_key)) {
      throw new SyntaxError("invalid Pitch hack");
    }
    prior = id;
    return Object.freeze({ hack_id: id, price: safeInteger(row.price, 0, MAX_EXACT_INTEGER, "hack price"),
      draft_weight: safeInteger(row.draft_weight, 1, MAX_EXACT_INTEGER, "draft weight"), effect: parseEffect(row.effect), copy_key: row.copy_key });
  });
  const hackIDs = new Set(hacks.map((row) => row.hack_id));
  for (const row of hacks) if (row.effect.kind === "chain_factor" && (!hackIDs.has(row.effect.partner_hack_id) || row.effect.partner_hack_id === row.hack_id)) {
    throw new SyntaxError("invalid Pitch chain partner");
  }
  let priorTarget = parseCanonical("0");
  const funding = root.funding_curve.map((source, index) => {
    const row = exactObject(source, ["round", "funding_target"], "Pitch funding row");
    const target = canonicalPositive(row.funding_target, "funding target");
    if (row.round !== index + 1 || !target.gt(priorTarget)) throw new SyntaxError("invalid Pitch funding curve");
    priorTarget = target;
    return Object.freeze({ round: index + 1, funding_target: row.funding_target as string });
  });
  return Object.freeze({ schema_version: 1 as const, policy, metric_cards: Object.freeze(cards),
    growth_hacks: Object.freeze(hacks), funding_curve: Object.freeze(funding) });
}

function parsePolicy(source: unknown, declared?: ReadonlySet<string>): PitchPolicy {
  const row = exactObject(source, ["hand_size", "play_size", "hands_per_round", "hack_slots", "start_currency", "shop_size",
    "round_clear_currency", "hand_size_reason_key", "play_size_reason_key", "hands_per_round_reason_key",
    "hack_slots_reason_key", "rounds_reason_key", "exponent_reason_key", "best_exponent_hardcap"], "Pitch policy");
  if (row.hand_size !== 7 || row.play_size !== 4 || row.hands_per_round !== 3 || row.hack_slots !== 4 || row.start_currency !== 4 ||
    row.shop_size !== 3 || row.round_clear_currency !== 3 || row.best_exponent_hardcap !== 1_000_000 || row.exponent_reason_key !== "cap.pitch_exponent") {
    throw new SyntaxError("invalid Pitch policy literals");
  }
  for (const key of [row.hand_size_reason_key, row.play_size_reason_key, row.hands_per_round_reason_key,
    row.hack_slots_reason_key, row.rounds_reason_key, row.exponent_reason_key]) {
    if (typeof key !== "string" || !mechanical.test(key) || declared && !declared.has(key)) throw new SyntaxError("invalid Pitch reason key");
  }
  return Object.freeze(row as unknown as PitchPolicy);
}

function parseEffect(source: unknown): PitchEffect {
  const probe = exactRecord(source, "Pitch effect");
  switch (probe.kind) {
    case "flat_add": { const row = exactObject(source, ["kind", "amount"], "flat add"); canonicalPositive(row.amount, "flat add"); return Object.freeze(row as unknown as PitchEffect); }
    case "card_factor": { const row = exactObject(source, ["kind", "factor"], "card factor"); canonicalPositive(row.factor, "card factor"); return Object.freeze(row as unknown as PitchEffect); }
    case "shape_factor": { const row = exactObject(source, ["kind", "shape", "factor"], "shape factor"); if (row.shape !== "pair" && row.shape !== "full_hand") throw new SyntaxError("invalid Pitch shape"); canonicalPositive(row.factor, "shape factor"); return Object.freeze(row as unknown as PitchEffect); }
    case "chain_factor": { const row = exactObject(source, ["kind", "partner_hack_id", "factor"], "chain factor"); mechanicalString(row.partner_hack_id, "chain partner"); canonicalPositive(row.factor, "chain factor"); return Object.freeze(row as unknown as PitchEffect); }
    default: throw new SyntaxError("invalid Pitch effect kind");
  }
}

function canonicalPositive(source: unknown, label: string) { const value = canonicalNonnegative(source, label); if (!parseCanonical(value).gt(0)) throw new SyntaxError(`invalid ${label}`); return parseCanonical(value); }
function canonicalNonnegative(source: unknown, label: string): string { if (typeof source !== "string") throw new SyntaxError(`invalid ${label}`); const value = parseCanonical(source); if (value.lt(0)) throw new SyntaxError(`invalid ${label}`); return source; }
function mechanicalString(source: unknown, label: string): string { if (typeof source !== "string" || !mechanical.test(source)) throw new SyntaxError(`invalid ${label}`); return source; }
function exactRecord(source: unknown, label: string): Record<string, unknown> { if (typeof source !== "object" || source === null || Array.isArray(source)) throw new SyntaxError(`${label} must be an object`); return source as Record<string, unknown>; }
function exactObject(source: unknown, keys: readonly string[], label: string): Record<string, unknown> { const row = exactRecord(source, label); const actual = Object.keys(row).sort(byteCompare), expected = [...keys].sort(byteCompare); if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError(`${label} fields are not exact`); return row; }
function safeInteger(source: unknown, minimum: number, maximum: number, label: string): number { if (typeof source !== "number" || !Number.isSafeInteger(source) || source < minimum || source > maximum) throw new SyntaxError(`invalid ${label}`); return source; }
export function byteCompare(left: string, right: string): number { const a = new TextEncoder().encode(left), b = new TextEncoder().encode(right); for (let index = 0; index < Math.min(a.length, b.length); index++) if (a[index] !== b[index]) return a[index]! - b[index]!; return a.length - b.length; }

export async function pitchContentHash(bytes: string | Uint8Array): Promise<string> {
  const data: Uint8Array<ArrayBuffer> = typeof bytes === "string" ? new TextEncoder().encode(bytes) : new Uint8Array(bytes);
  const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", data));
  return `sha256:${[...digest].map((value) => value.toString(16).padStart(2, "0")).join("")}`;
}
