import Decimal from "break_infinity.js";
import { SplitMix64, substream } from "../combat/rng";
import { COPY_KEYS } from "../copy";
import { canonicalString, isStateValue, parseCanonical, quantize, sumDeterministic } from "../numeric";
import { byteCompare, parsePitchCatalog, pitchContentHash, PITCH_SCHEMA_VERSION, type PitchCatalog, type PitchHack } from "./catalog";

export const PITCH_ENGINE_VERSION = "1.0.0" as const;
export type PitchCommand = Readonly<{ kind: "play_hand"; card_ids: readonly string[] }> | Readonly<{ kind: "buy_hack"; offer_id: string }> | Readonly<{ kind: "end_shop" }>;
export interface PitchOffer { readonly offer_id: string; readonly hack_id: string; readonly price: number }
export interface PitchSnapshot {
  readonly phase: "playing" | "shop" | "terminal"; readonly round: number; readonly hands_remaining: number;
  readonly deck_count: number; readonly hand: readonly string[]; readonly slotted_hacks: readonly string[];
  readonly run_currency: number; readonly shop_offers: readonly PitchOffer[]; readonly funding_target: string;
  readonly round_best_valuation: string; readonly revision: number; readonly pitch_content_hash: string;
  readonly pitch_schema_version: 1;
}
export interface PitchResult { readonly outcome: "funded" | "funding_failed"; readonly rating_delta: null; readonly score_facts: readonly [{ readonly kind: "pitch.best_hand_exponent"; readonly value: number }, { readonly kind: "pitch.final_round"; readonly value: number }] }
export interface PitchContentInput { readonly content: string; readonly content_hash: string; readonly content_schema_version: number }

export class PitchRejection extends Error { constructor(readonly code: "duplicate_card" | "hack_slots_full" | "hand_too_large" | "illegal_phase" | "insufficient_currency" | "unknown_card" | "unknown_offer", readonly detail: string) { super(`${code}: ${detail}`); } }

export async function createPitch(input: PitchContentInput & { readonly seed: bigint; readonly mode: "solo"; readonly scaling_inputs: Readonly<Record<string, number>> }): Promise<PitchSnapshot> {
  const catalog = await resolveCatalog(input);
  if (input.mode !== "solo" || !validScaling(input.scaling_inputs)) throw new SyntaxError("invalid Pitch tenant input");
  const [hand, deckCount] = deal(catalog, input.seed, 1, 1);
  return freezeSnapshot({ phase: "playing", round: 1, hands_remaining: catalog.policy.hands_per_round,
    deck_count: deckCount, hand, slotted_hacks: [], run_currency: catalog.policy.start_currency, shop_offers: [],
    funding_target: catalog.funding_curve[0]!.funding_target, round_best_valuation: "0", revision: 1,
    pitch_content_hash: input.content_hash, pitch_schema_version: PITCH_SCHEMA_VERSION });
}

export async function applyPitch(input: PitchContentInput & { readonly seed: bigint; readonly mode: "solo"; readonly scaling_inputs: Readonly<Record<string, number>>; readonly revision: number; readonly snapshot: PitchSnapshot; readonly command: unknown }): Promise<{ readonly snapshot: PitchSnapshot; readonly result: PitchResult | null }> {
  const catalog = await resolveCatalog(input);
  if (input.mode !== "solo" || !validScaling(input.scaling_inputs)) throw new SyntaxError("invalid Pitch tenant input");
  const snapshot = parseSnapshot(structuredClone(input.snapshot), catalog);
  if (snapshot.revision !== input.revision || snapshot.pitch_content_hash !== input.content_hash || snapshot.pitch_schema_version !== input.content_schema_version) throw new SyntaxError("Pitch snapshot identity divergence");
  const command = parseCommand(input.command);
  let result: PitchResult | null = null;
  switch (command.kind) {
    case "play_hand": result = playHand(snapshot, command.card_ids, catalog, input.seed); break;
    case "buy_hack": buyHack(snapshot, command.offer_id, catalog); break;
    case "end_shop": endShop(snapshot, catalog, input.seed); break;
  }
  snapshot.revision = input.revision + 1;
  return Object.freeze({ snapshot: freezeSnapshot(snapshot), result });
}

type MutableSnapshot = { -readonly [K in keyof PitchSnapshot]: PitchSnapshot[K] extends readonly (infer V)[] ? V[] : PitchSnapshot[K] };

function playHand(snapshot: MutableSnapshot, selected: readonly string[], catalog: PitchCatalog, seed: bigint): PitchResult | null {
  if (snapshot.phase !== "playing") throw reject("illegal_phase", "play_hand requires playing phase");
  if (selected.length > catalog.policy.play_size) throw reject("hand_too_large", "played hand exceeds play_size");
  const hand = new Set(snapshot.hand);
  for (const id of selected) if (!hand.has(id)) throw reject("unknown_card", "selected card is not in hand");
  const valuation = score(selected, snapshot.slotted_hacks, catalog);
  const best = Decimal.max(valuation, parseCanonical(snapshot.round_best_valuation));
  snapshot.round_best_valuation = canonicalString(best);
  snapshot.hands_remaining--;
  if (best.gte(parseCanonical(snapshot.funding_target))) {
    if (snapshot.round === catalog.funding_curve.length) {
      snapshot.phase = "terminal"; snapshot.shop_offers = [];
      return terminalResult(snapshot, catalog, "funded");
    }
    snapshot.run_currency += catalog.policy.round_clear_currency;
    snapshot.phase = "shop"; snapshot.hand = []; snapshot.shop_offers = shopOffers(catalog, seed, snapshot.round, snapshot.slotted_hacks);
    return null;
  }
  if (snapshot.hands_remaining === 0) {
    snapshot.phase = "terminal"; snapshot.shop_offers = [];
    return terminalResult(snapshot, catalog, "funding_failed");
  }
  const handNumber = catalog.policy.hands_per_round - snapshot.hands_remaining + 1;
  [snapshot.hand, snapshot.deck_count] = deal(catalog, seed, snapshot.round, handNumber);
  return null;
}

function buyHack(snapshot: MutableSnapshot, offerID: string, catalog: PitchCatalog): void {
  if (snapshot.phase !== "shop") throw reject("illegal_phase", "buy_hack requires shop phase");
  if (snapshot.slotted_hacks.length >= catalog.policy.hack_slots) throw reject("hack_slots_full", catalog.policy.hack_slots_reason_key);
  const index = snapshot.shop_offers.findIndex((offer) => offer.offer_id === offerID);
  if (index < 0) throw reject("unknown_offer", "offer is not active");
  const offer = snapshot.shop_offers[index]!;
  if (snapshot.run_currency < offer.price) throw reject("insufficient_currency", "run currency is below offer price");
  snapshot.run_currency -= offer.price;
  snapshot.slotted_hacks = [...snapshot.slotted_hacks, offer.hack_id].sort(byteCompare);
  snapshot.shop_offers = snapshot.shop_offers.filter((_, at) => at !== index);
}

function endShop(snapshot: MutableSnapshot, catalog: PitchCatalog, seed: bigint): void {
  if (snapshot.phase !== "shop" || snapshot.round >= catalog.funding_curve.length) throw reject("illegal_phase", "end_shop requires non-final shop");
  snapshot.round++; snapshot.phase = "playing"; snapshot.hands_remaining = catalog.policy.hands_per_round;
  [snapshot.hand, snapshot.deck_count] = deal(catalog, seed, snapshot.round, 1);
  snapshot.shop_offers = []; snapshot.round_best_valuation = "0"; snapshot.funding_target = catalog.funding_curve[snapshot.round - 1]!.funding_target;
}

function score(selected: readonly string[], hacks: readonly string[], catalog: PitchCatalog): Decimal {
  let flat = new Decimal(0); const cardFactors: Decimal[] = []; const counts = new Map<string, number>();
  for (const id of hacks) { const hack = hackByID(catalog, id); if (hack.effect.kind === "flat_add") flat = flat.add(parseCanonical(hack.effect.amount)); else if (hack.effect.kind === "card_factor") cardFactors.push(parseCanonical(hack.effect.factor)); }
  const terms = selected.map((instance) => {
    const base = baseCardID(instance), card = catalog.metric_cards.find((row) => row.card_id === base);
    if (!card) throw new SyntaxError("unknown Pitch card"); counts.set(base, (counts.get(base) ?? 0) + 1);
    let value = parseCanonical(card.base_metric).add(flat); for (const factor of cardFactors) value = value.mul(factor); return value;
  });
  let total = terms.length === 0 ? new Decimal(0) : sumDeterministic(terms); const owned = new Set(hacks); const pair = [...counts.values()].some((count) => count >= 2);
  for (const id of hacks) { const effect = hackByID(catalog, id).effect; if (effect.kind === "shape_factor" && (effect.shape === "pair" && pair || effect.shape === "full_hand" && selected.length === catalog.policy.play_size)) total = total.mul(parseCanonical(effect.factor)); else if (effect.kind === "chain_factor" && owned.has(effect.partner_hack_id)) total = total.mul(parseCanonical(effect.factor)); }
  total = quantize(total); if (!isStateValue(total) || total.lt(0)) throw new RangeError("Pitch valuation outside Decimal domain"); return total;
}

function terminalResult(snapshot: MutableSnapshot, catalog: PitchCatalog, outcome: "funded" | "funding_failed"): PitchResult {
  const exponent = pitchBestExponent(snapshot.round_best_valuation, catalog.policy.best_exponent_hardcap);
  return Object.freeze({ outcome, rating_delta: null, score_facts: Object.freeze([
    Object.freeze({ kind: "pitch.best_hand_exponent" as const, value: exponent }),
    Object.freeze({ kind: "pitch.final_round" as const, value: snapshot.round }),
  ]) as PitchResult["score_facts"] });
}

export function pitchBestExponent(valuation: string, hardcap = 1_000_000): number {
  const exponent = valuation === "0" ? 0 : parseCanonical(valuation).exponent;
  return Math.min(exponent, hardcap);
}

function deal(catalog: PitchCatalog, seed: bigint, round: number, handNumber: number): [string[], number] {
  const deck = catalog.metric_cards.flatMap((card) => [`${card.card_id}#1`, `${card.card_id}#2`]);
  const random = coordinateRandom(seed, "pitch.deck.v1", round);
  for (let index = deck.length - 1; index > 0; index--) { const swap = Number(random.bound(BigInt(index + 1))); [deck[index], deck[swap]] = [deck[swap]!, deck[index]!]; }
  const start = (handNumber - 1) * catalog.policy.hand_size, end = start + catalog.policy.hand_size;
  return [deck.slice(start, end).sort(byteCompare), deck.length - end];
}

function shopOffers(catalog: PitchCatalog, seed: bigint, round: number, ownedIDs: readonly string[]): PitchOffer[] {
  const owned = new Set(ownedIDs), pool = catalog.growth_hacks.filter((row) => !owned.has(row.hack_id));
  const random = coordinateRandom(seed, "pitch.shop.v1", round), result: PitchOffer[] = [];
  for (let slot = 1; slot <= Math.min(catalog.policy.shop_size, pool.length); slot++) {
    const total = pool.reduce((sum, row) => sum + BigInt(row.draft_weight), 0n); let draw = random.bound(total), chosen = 0;
    for (let index = 0; index < pool.length; index++) { if (draw < BigInt(pool[index]!.draft_weight)) { chosen = index; break; } draw -= BigInt(pool[index]!.draft_weight); }
    const hack = pool.splice(chosen, 1)[0]!; result.push({ hack_id: hack.hack_id, offer_id: `pitch.offer.${round}.${slot}.${hack.hack_id}`, price: hack.price });
  }
  return result.sort((left, right) => byteCompare(left.offer_id, right.offer_id));
}

function coordinateRandom(seed: bigint, label: string, coordinate: number): SplitMix64 { const runSeed = substream(seed, "pitch.run.v1").next(); return substream(runSeed ^ BigInt(coordinate), label); }
function hackByID(catalog: PitchCatalog, id: string): PitchHack { const row = catalog.growth_hacks.find((candidate) => candidate.hack_id === id); if (!row) throw new SyntaxError("unknown Pitch hack"); return row; }
function baseCardID(instance: string): string { const match = /^([a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*)#([1-9][0-9]*)$/.exec(instance); if (!match) throw new SyntaxError("invalid Pitch card instance"); return match[1]!; }
function validScaling(values: Readonly<Record<string, number>>): boolean { return Object.keys(values).length === 1 && values["minigame.pitch"] === 1; }
function reject(code: PitchRejection["code"], detail: string): PitchRejection { return new PitchRejection(code, detail); }

function parseCommand(source: unknown): PitchCommand {
  const probe = exactRecord(source, "Pitch command");
  if (probe.kind === "play_hand") { const row = exactObject(source, ["kind", "card_ids"], "play_hand"); if (!Array.isArray(row.card_ids) || row.card_ids.some((id) => typeof id !== "string")) throw reject("unknown_card", "play_hand schema mismatch"); if (row.card_ids.length > 4) throw reject("hand_too_large", "played hand exceeds play_size"); const ids = [...row.card_ids] as string[]; if (!strictlySorted(ids)) throw reject("duplicate_card", "card_ids must be unique and byte-sorted"); return { kind: "play_hand", card_ids: ids }; }
  if (probe.kind === "buy_hack") { const row = exactObject(source, ["kind", "offer_id"], "buy_hack"); if (typeof row.offer_id !== "string" || row.offer_id === "") throw reject("unknown_offer", "buy_hack schema mismatch"); return { kind: "buy_hack", offer_id: row.offer_id }; }
  if (probe.kind === "end_shop") { exactObject(source, ["kind"], "end_shop"); return { kind: "end_shop" }; }
  throw reject("illegal_phase", "unknown command kind");
}

function parseSnapshot(source: unknown, catalog: PitchCatalog): MutableSnapshot {
  const row = exactObject(source, ["phase", "round", "hands_remaining", "deck_count", "hand", "slotted_hacks", "run_currency", "shop_offers", "funding_target", "round_best_valuation", "revision", "pitch_content_hash", "pitch_schema_version"], "Pitch snapshot");
  if (row.phase !== "playing" && row.phase !== "shop" && row.phase !== "terminal" || !Array.isArray(row.hand) || !Array.isArray(row.slotted_hacks) || !Array.isArray(row.shop_offers)) throw new SyntaxError("invalid Pitch snapshot");
  const result = row as unknown as MutableSnapshot;
  for (const key of ["round", "hands_remaining", "deck_count", "run_currency", "revision"] as const) if (!Number.isSafeInteger(result[key])) throw new SyntaxError("invalid Pitch snapshot integer");
  if (result.round < 1 || result.round > 8 || result.hands_remaining < 0 || result.hands_remaining > 3 || result.deck_count < 0 || result.deck_count > 24 || result.run_currency < 0 || result.revision < 1 ||
    result.pitch_schema_version !== 1 || !/^sha256:[0-9a-f]{64}$/.test(result.pitch_content_hash) || !strictlySorted(result.hand as string[]) || !strictlySorted(result.slotted_hacks as string[])) throw new SyntaxError("invalid Pitch snapshot domain");
  if (parseCanonical(result.funding_target).lt(0) || parseCanonical(result.round_best_valuation).lt(0)) throw new SyntaxError("invalid Pitch Decimal domain");
  if (result.funding_target !== catalog.funding_curve[result.round - 1]?.funding_target || result.hand.length > 7 || result.slotted_hacks.length > 4) throw new SyntaxError("Pitch snapshot/catalog divergence");
  for (const instance of result.hand) { const base = baseCardID(instance); const card = catalog.metric_cards.find((candidate) => candidate.card_id === base); const ordinal = Number(instance.slice(instance.lastIndexOf("#") + 1)); if (!card || ordinal < 1 || ordinal > card.copies) throw new SyntaxError("unknown Pitch card instance"); }
  for (const hackID of result.slotted_hacks) hackByID(catalog, hackID);
  let prior = ""; result.shop_offers = result.shop_offers.map((source) => { const offer = exactObject(source, ["offer_id", "hack_id", "price"], "Pitch offer") as unknown as PitchOffer; const hack = catalog.growth_hacks.find((candidate) => candidate.hack_id === offer.hack_id); if (typeof offer.offer_id !== "string" || prior !== "" && byteCompare(prior, offer.offer_id) >= 0 || !hack || offer.price !== hack.price) throw new SyntaxError("invalid Pitch offer"); prior = offer.offer_id; return offer; });
  return result;
}

async function resolveCatalog(input: PitchContentInput): Promise<PitchCatalog> { if (input.content_schema_version !== PITCH_SCHEMA_VERSION || input.content_hash !== await pitchContentHash(input.content)) throw new SyntaxError("Pitch content identity mismatch"); return parsePitchCatalog(JSON.parse(input.content), new Set(COPY_KEYS)); }
function freezeSnapshot(source: MutableSnapshot): PitchSnapshot {
  return Object.freeze({
    deck_count: source.deck_count,
    funding_target: source.funding_target,
    hand: Object.freeze([...source.hand]),
    hands_remaining: source.hands_remaining,
    phase: source.phase,
    pitch_content_hash: source.pitch_content_hash,
    pitch_schema_version: source.pitch_schema_version,
    revision: source.revision,
    round: source.round,
    round_best_valuation: source.round_best_valuation,
    run_currency: source.run_currency,
    shop_offers: Object.freeze(source.shop_offers.map((row) => Object.freeze({ hack_id: row.hack_id, offer_id: row.offer_id, price: row.price }))),
    slotted_hacks: Object.freeze([...source.slotted_hacks]),
  });
}
function strictlySorted(values: readonly string[]): boolean { return values.every((value, index) => value !== "" && (index === 0 || byteCompare(values[index - 1]!, value) < 0)); }
function exactRecord(source: unknown, label: string): Record<string, unknown> { if (typeof source !== "object" || source === null || Array.isArray(source)) throw new SyntaxError(`${label} must be an object`); return source as Record<string, unknown>; }
function exactObject(source: unknown, keys: readonly string[], label: string): Record<string, unknown> { const row = exactRecord(source, label), actual = Object.keys(row).sort(byteCompare), expected = [...keys].sort(byteCompare); if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError(`${label} fields are not exact`); return row; }
