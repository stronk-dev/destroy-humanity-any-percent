import { MAX_EXACT_INTEGER } from "../numeric";
import { parseOfflineQualityPolicy, type OfflineQualityPolicy } from "./offline-quality";

const mechanical = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;
const version = /^[0-9]+\.[0-9]+\.[0-9]+$/;

export interface MinigameRatingPolicy {
  readonly starting_elo: number;
  readonly elo_floor: number;
  readonly elo_ceiling: number;
  readonly provisional_games: number;
  readonly season_member: string;
}

export interface MinigameDefinition {
  readonly minigame_id: string;
  readonly engine_ref: string;
  readonly engine_version: string;
  readonly modes: readonly ("async_snapshot" | "solo")[];
  readonly result_score_fact_ids: readonly string[];
  readonly scaling: Readonly<Record<string, unknown>>;
  readonly payout: Readonly<Record<string, unknown>>;
  readonly fallback: Readonly<Record<string, unknown>>;
  readonly offline_quality: OfflineQualityPolicy;
  readonly rating_policy: MinigameRatingPolicy;
  readonly unlock_condition: Readonly<Record<string, unknown>>;
	readonly soul_gate: "" | "human_hobby" | "unrelated";
}

export interface MinigameCatalog {
	readonly schemaVersion: 2 | 3;
  readonly minigameIds: readonly string[];
  readonly ratingSeasons: readonly string[];
  readonly minigames: readonly MinigameDefinition[];
}

export function parseMinigameCatalog(source: unknown): MinigameCatalog {
  const root = exactObject(source, ["schema_version", "rating_seasons", "minigames"], "minigame catalog");
	if (root.schema_version !== 2 && root.schema_version !== 3) throw new SyntaxError("invalid minigame catalog version");
	const schemaVersion = root.schema_version;
  const ratingSeasons = sortedMechanical(root.rating_seasons, "rating seasons");
  if (!Array.isArray(root.minigames)) throw new SyntaxError("minigames must be an array");
  let prior = "";
  const minigames = root.minigames.map((raw) => {
		const row = parseDefinition(raw, ratingSeasons, schemaVersion);
    if (prior !== "" && byteCompare(prior, row.minigame_id) >= 0) throw new SyntaxError("minigames are not byte sorted");
    prior = row.minigame_id;
    return row;
  });
	return Object.freeze({ schemaVersion, minigameIds: Object.freeze(minigames.map((row) => row.minigame_id)), ratingSeasons, minigames: Object.freeze(minigames) });
}

function parseDefinition(source: unknown, ratingSeasons: readonly string[], schemaVersion: 2 | 3): MinigameDefinition {
	const keys = ["minigame_id", "engine_ref", "engine_version", "modes", "result_score_fact_ids", "scaling", "payout", "fallback", "offline_quality", "rating_policy", "unlock_condition"];
	if (schemaVersion === 3) keys.push("soul_gate");
	const row = exactObject(source, keys, "minigame");
  const minigameId = mechanicalString(row.minigame_id, "minigame id");
  const engineRef = mechanicalString(row.engine_ref, "engine ref");
  if (typeof row.engine_version !== "string" || !version.test(row.engine_version)) throw new SyntaxError("invalid engine version");
  const modes = sortedStrings(row.modes, "modes", new Set(["async_snapshot", "solo"])) as readonly ("async_snapshot" | "solo")[];
  if (modes.length === 0) throw new SyntaxError("minigame needs a mode");
  const factIds = sortedMechanical(row.result_score_fact_ids, "score fact ids");
  const facts = new Set(factIds);
  const scaling = parseScaling(row.scaling);
  const payout = parsePayout(row.payout, facts);
  const fallback = parseFallback(row.fallback);
  const quality = parseOfflineQualityPolicy(row.offline_quality, {
    score_fact_ids: facts,
    automation_destinations: new Set([mechanicalString((row.offline_quality as Record<string, unknown>)?.automation_destination, "automation destination")]),
  });
  const rating = parseRating(row.rating_policy, ratingSeasons);
  const unlock = parseUnlock(row.unlock_condition);
	const soulGate = schemaVersion === 2 ? "" : row.soul_gate === "human_hobby" || row.soul_gate === "unrelated" ? row.soul_gate : (() => { throw new SyntaxError("invalid minigame Soul gate"); })();
  return Object.freeze({ minigame_id: minigameId, engine_ref: engineRef, engine_version: row.engine_version,
    modes: Object.freeze([...modes]), result_score_fact_ids: factIds, scaling, payout, fallback,
		offline_quality: quality, rating_policy: rating, unlock_condition: unlock, soul_gate: soulGate });
}

export function minigameCatalogSupportsSoul(catalog: MinigameCatalog): boolean {
	return catalog.schemaVersion === 3 && catalog.minigames.every((row) => row.soul_gate === "human_hobby" || row.soul_gate === "unrelated");
}

function parseScaling(source: unknown): Readonly<Record<string, unknown>> {
  const root = exactObject(source, ["schema_version", "scaling_inputs"], "scaling policy");
  if (root.schema_version !== 1 || !Array.isArray(root.scaling_inputs) || root.scaling_inputs.length === 0) throw new SyntaxError("invalid scaling policy");
  const destinations = new Set<string>();
  for (const raw of root.scaling_inputs) {
    const row = exactObject(raw, ["destination", "destination_class", "source_kind", "source_ref", "op", "operand", "clamp_min", "clamp_max"], "scaling row");
    const destination = mechanicalString(row.destination, "scaling destination");
    if (destinations.has(destination)) throw new SyntaxError("duplicate scaling destination");
    destinations.add(destination);
    if (!new Set(["breadth", "presentation"]).has(String(row.destination_class))) throw new SyntaxError("ranked power scaling is forbidden");
    if (!new Set(["literal", "tier", "purchased_generator_count", "founder_carry_counter", "attended_quality_grade"]).has(String(row.source_kind)) ||
      !new Set(["identity", "add", "mul", "floordiv"]).has(String(row.op))) throw new SyntaxError("invalid scaling grammar");
    if (typeof row.source_ref !== "string" || row.source_ref.length === 0) throw new SyntaxError("invalid scaling source");
    const operand = safeInteger(row.operand, -MAX_EXACT_INTEGER, MAX_EXACT_INTEGER, "scaling operand");
    const minimum = safeInteger(row.clamp_min, -MAX_EXACT_INTEGER, MAX_EXACT_INTEGER, "scaling minimum");
    const maximum = safeInteger(row.clamp_max, -MAX_EXACT_INTEGER, MAX_EXACT_INTEGER, "scaling maximum");
    if (minimum > maximum || row.op === "identity" && operand !== 0 || row.op === "floordiv" && operand <= 0) throw new SyntaxError("invalid scaling bounds");
  }
  return Object.freeze(root);
}

function parsePayout(source: unknown, facts: ReadonlySet<string>): Readonly<Record<string, unknown>> {
  const row = exactObject(source, ["credited_resource_id", "sends_per_day", "per_send_cap", "conversion_ppm", "payout_score_fact_id", "cap_reason_key"], "payout policy");
  mechanicalString(row.credited_resource_id, "credited resource");
  const fact = mechanicalString(row.payout_score_fact_id, "payout score fact");
  if (!facts.has(fact)) throw new SyntaxError("unknown payout score fact");
  mechanicalString(row.cap_reason_key, "cap reason");
  safeInteger(row.sends_per_day, 0, MAX_EXACT_INTEGER, "sends per day");
  safeInteger(row.per_send_cap, 0, MAX_EXACT_INTEGER, "per send cap");
  safeInteger(row.conversion_ppm, 0, 1_000_000, "conversion ppm");
  return Object.freeze(row);
}

function parseFallback(source: unknown): Readonly<Record<string, unknown>> {
  const probe = exactRecord(source, "fallback");
  switch (probe.kind) {
    case "solo": return Object.freeze(exactObject(source, ["kind"], "solo fallback"));
    case "bot": {
      const row = exactObject(source, ["kind", "bot_ref", "rate_reduction_ppm"], "bot fallback");
      parseIdentity(row.bot_ref, "policy_id"); safeInteger(row.rate_reduction_ppm, 0, 1_000_000, "fallback reduction"); return Object.freeze(row);
    }
    case "npc_partner": {
      const row = exactObject(source, ["kind", "npc_profile", "rate_reduction_ppm"], "npc fallback");
      parseIdentity(row.npc_profile, "profile_id"); safeInteger(row.rate_reduction_ppm, 0, 1_000_000, "fallback reduction"); return Object.freeze(row);
    }
    default: throw new SyntaxError("invalid fallback kind");
  }
}

function parseIdentity(source: unknown, idKey: "policy_id" | "profile_id"): void {
  const row = exactObject(source, [idKey, "version"], "fallback identity");
  mechanicalString(row[idKey], "fallback id");
  if (typeof row.version !== "string" || !version.test(row.version)) throw new SyntaxError("invalid fallback version");
}

function parseRating(source: unknown, seasons: readonly string[]): MinigameRatingPolicy {
  const row = exactObject(source, ["starting_elo", "elo_floor", "elo_ceiling", "provisional_games", "season_member"], "rating policy");
  const floor = safeInteger(row.elo_floor, -MAX_EXACT_INTEGER, MAX_EXACT_INTEGER, "elo floor");
  const ceiling = safeInteger(row.elo_ceiling, -MAX_EXACT_INTEGER, MAX_EXACT_INTEGER, "elo ceiling");
  const start = safeInteger(row.starting_elo, floor, ceiling, "starting elo");
  const season = mechanicalString(row.season_member, "rating season");
  if (!seasons.includes(season) || floor > ceiling) throw new SyntaxError("invalid rating policy");
  return Object.freeze({ starting_elo: start, elo_floor: floor, elo_ceiling: ceiling,
    provisional_games: safeInteger(row.provisional_games, 0, MAX_EXACT_INTEGER, "provisional games"), season_member: season });
}

function parseUnlock(source: unknown): Readonly<Record<string, unknown>> {
  const probe = exactRecord(source, "unlock condition");
  if (probe.kind === "always") return Object.freeze(exactObject(source, ["kind"], "always unlock"));
  if (probe.kind !== "fact_equals") throw new SyntaxError("invalid unlock condition");
  const row = exactObject(source, ["kind", "fact_id", "value"], "fact unlock");
  mechanicalString(row.fact_id, "unlock fact");
  if (!(typeof row.value === "boolean" || typeof row.value === "string" || Number.isSafeInteger(row.value))) throw new SyntaxError("invalid unlock value");
  return Object.freeze(row);
}

function sortedMechanical(source: unknown, label: string): readonly string[] {
  return sortedStrings(source, label, undefined, true);
}

function sortedStrings(source: unknown, label: string, allowed?: ReadonlySet<string>, requireMechanical = false): readonly string[] {
  if (!Array.isArray(source)) throw new SyntaxError(`${label} must be an array`);
  let prior = "";
  return Object.freeze(source.map((item) => {
    if (typeof item !== "string" || requireMechanical && !mechanical.test(item) || allowed && !allowed.has(item) || prior !== "" && byteCompare(prior, item) >= 0) throw new SyntaxError(`invalid ${label}`);
    prior = item; return item;
  }));
}

function mechanicalString(source: unknown, label: string): string {
  if (typeof source !== "string" || !mechanical.test(source)) throw new SyntaxError(`invalid ${label}`);
  return source;
}

function exactRecord(source: unknown, label: string): Record<string, unknown> {
  if (typeof source !== "object" || source === null || Array.isArray(source)) throw new SyntaxError(`${label} must be an object`);
  return source as Record<string, unknown>;
}

function exactObject(source: unknown, keys: readonly string[], label: string): Record<string, unknown> {
  const value = exactRecord(source, label);
  const actual = Object.keys(value).sort(byteCompare); const expected = [...keys].sort(byteCompare);
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError(`${label} fields are not exact`);
  return value;
}

function safeInteger(value: unknown, minimum: number, maximum: number, label: string): number {
  if (typeof value !== "number" || !Number.isSafeInteger(value) || value < minimum || value > maximum) throw new SyntaxError(`${label} outside integer domain`);
  return value;
}

function byteCompare(left: string, right: string): number {
  const a = new TextEncoder().encode(left); const b = new TextEncoder().encode(right);
  for (let index = 0; index < Math.min(a.length, b.length); index++) if (a[index] !== b[index]) return a[index]! - b[index]!;
  return a.length - b.length;
}
