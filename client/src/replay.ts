import Decimal from "break_infinity.js";

import { loadAchievementCatalog, type AchievementCatalog, type AchievementRegistry } from "./achievements/catalog";
import { achievementScore, newlyEarned, type AchievementObservation } from "./achievements/evaluate";
import { enclosureIndex, parseCommonsCatalog, type CommonsCatalog } from "./commons";
import { COPY_KEYS } from "./copy";
import { ladderSourceId, manualRoleSourceId, parseCatalog, subProgressValue, validateCatalogGateReferences, type EconomyCatalog, type MultiplierSlot, MULTIPLIER_SLOT_ORDER } from "./economy-kernel";
import { parseFactionCatalog, type FactionCatalog } from "./faction";
import { parseGuildCatalog, type GuildCatalog } from "./guild";
import { loadMeterCatalog, validateMeterResourceSeparation, type MeterCatalog } from "./meters/catalog";
import { advanceMeters, contributionKey as meterContributionKey, newRunMeterState, validateMeterState } from "./meters/transition";
import { canonicalString, isStateValue, MAX_EXACT_INTEGER, parseCanonical, quantize, sumDeterministic } from "./numeric";
import { parsePrestigePolicy, type PrestigePolicy } from "./prestige";
import { parseMinigameCatalog, type MinigameCatalog } from "./minigame/catalog";
import { parsePetCatalog, type PetCatalog } from "./pet/catalog";
import { parsePetCareStates, validatePetCareStatesForCatalog, type PetCareState } from "./pet/state";
import { discountedRequirements, evaluatePredicate, parseRoutesCatalog, type RouteContext, type RoutesCatalog } from "./routes";

const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;
const uuidV7 = /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;
const hashPattern = /^sha256:[0-9a-f]{64}$/;
const mechanical = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;

export interface ReplayArtifacts { readonly categories: string; readonly economy: string; readonly routes: string; readonly commons: string; readonly prestige: string; readonly factions: string; readonly guilds: string; readonly meters?: string; readonly achievements?: string; readonly minigames?: string; readonly pets?: string }
export interface ReplayCatalogBundle {
  readonly constantsHash: string; readonly artifacts: ReplayArtifacts; readonly economy: EconomyCatalog; readonly routes: RoutesCatalog;
  readonly commons: CommonsCatalog; readonly prestige: PrestigePolicy; readonly factions: FactionCatalog; readonly guilds: GuildCatalog;
  readonly meters?: MeterCatalog; readonly achievements?: AchievementCatalog; readonly minigames?: MinigameCatalog; readonly pets?: PetCatalog;
  readonly next?: ReplayCatalogBundle;
}
export interface ReplayContribution { readonly slot: MultiplierSlot; readonly source_id: string; readonly target: string; readonly factor: string }
export interface ReplayEvent { readonly kind: string; readonly schema_version: number; readonly intent_id: string; readonly payload: unknown }
export interface ReplayChange { readonly resource_id: string; readonly before: string; readonly delta: string; readonly after: string }
export interface ReplayInvariant { readonly kind: "afford_fallback" | "residual_clamp" | "residual_abort"; readonly intent_id: string; readonly detail: string }
export interface ReplayState {
  wireVersion: 14 | 16;
  balances: Record<string, string>; generators: Record<string, number>; generatorPurchasedTotal: number;
  upgradesOwned: Set<string>; generatorsProvisioned: Record<string, number>; provisionRemaindersPpm: Record<string, number>; stockRateRemainderPpm: number;
  evaluatedThroughMs: number; computeCreditMs: number;
  manualTokenMilli: number; manualTokenRefilledAtMs: number; gatesCrossed: Record<string, boolean>; runSeq: number;
  doctrinesByTransition: Record<string, string>; structureId: string; ledgerFactKinds: Set<string>; meterBands: Record<string, number>;
  regionTraits: Set<string>; routeKnowledgeBalance: number; hintsUnlocked: Set<string>; compactMember: boolean; compactTithePpm: number;
  compactSolidarityPpm: number; compactSamples: { hourStartMs: number; compliancePpm: number; coveredMs: number }[]; tier: number;
  lifetimeValue: string; offerState: null | { offerId: string; exitType: string; terms: unknown; spawnedAtMs: number; expiresAtMs: number };
  runStartedAtMs: number; runPreTimer: boolean; offlineSpans: { fromMs: number; toMs: number }[]; collapsedOfflineMs: number;
  factionId: string; incorporatedAtMs: number | null; factionStockResource: string; stockUnits: number; stockProgressMs: number;
  consumedStockUnits: number; guildTitheCarryPpm: number; guildBoundaryGuildId: string; guildBoundarySeq: number; guildConsumedWindow: number;
  meterValues: Record<string, number>; meterDecayRemainders: Record<string, number>; meterInputRemainders: Record<string, number>;
  achievementsEarnedRun: Set<string>; achievementScoreRun: number;
}
export interface LoggedTransition { readonly state: ReplayState; readonly outcome: "applied" | "rejected"; readonly receipt: unknown; readonly events: readonly ReplayEvent[]; readonly invariants: readonly ReplayInvariant[] }
export interface LoggedExitTransition {
  readonly founder: FounderCarry; readonly finalCompany: ReplayState; readonly newCompany: ReplayState | null;
  readonly outcome: "applied" | "rejected"; readonly receipt: unknown;
  readonly founderEvents: readonly ReplayEvent[]; readonly companyEndedEvents: readonly ReplayEvent[]; readonly companyStartedEvents: readonly ReplayEvent[];
}
export interface FounderReplayState {
  wireVersion: 14 | 15 | 16 | 17 | 18;
  balances: Record<string, string>; generators: Record<string, number>; generatorPurchasedTotal: number;
  upgradesOwned: Set<string>; generatorsProvisioned: Record<string, number>; provisionRemaindersPpm: Record<string, number>;
  stockRateRemainderPpm: number; evaluatedThroughMs: number; computeCreditMs: number; manualTokenMilli: number;
  manualTokenRefilledAtMs: number; routeKnowledgeBalance: number; hintsUnlocked: Set<string>;
  ledgerFactKinds: Set<string>; reputationLevel: number; networkSlots: NetworkSlot[]; cloutLifetime: number; soul: number;
  ageMs: number; notoriety: number; advisorMode: boolean; exitHistory: FounderExitRecord[];
  achievementsEarnedLifetime: Set<string>; achievementScoreLifetime: number;
  minigameRatings: Record<string, { elo: number; season_member: string; games_counted: number }>;
  minigameOfflineQuality: Record<string, { grade_ppm: number; last_founder_attended_ms: number; decay_remainder_ppm: number }>;
  pets: Record<string, PetCareState>;
}
export interface FounderLoggedTransition {
  readonly state: FounderReplayState; readonly outcome: "applied" | "rejected"; readonly receipt: unknown;
  readonly events: readonly ReplayEvent[]; readonly resultConstantsHash: string;
}
export interface FounderReplayLogEntry {
  readonly seq: number; readonly intentId: string; readonly constantsHash: string; readonly canonicalPayload: string;
  readonly replayInputs: unknown; readonly receiptJSON: string; readonly eventsJSON: string; readonly appliedRevision: number | null;
  readonly serverTSMS: number; readonly source: null | { readonly companyStreamId: string; readonly runSeq: number; readonly runLogSeq: number };
}
export interface FounderReplayHead { readonly revision: number; readonly version: 14 | 15 | 16 | 17 | 18; readonly constantsHash: string; readonly state: unknown }
export interface FounderAttendanceSample {
  readonly companyStreamId: string; readonly runSeq: number; readonly companyRevision: number; readonly companyConstantsHash: string;
  readonly completedAttendedMs: number; readonly currentRunPartialAttendedMs: number; readonly effectiveFounderAttendedMs: number;
}
export type ReplayVerdict = "verified" | "log_gap" | "state_divergence" | "constants_mismatch" | "clock_violation" | "engine_mismatch";
export interface ReplayVerificationResult { readonly verdict: ReplayVerdict; readonly finalState: ReplayState | null }
export class ReplayClockViolation extends RangeError { constructor() { super("replay clock violation"); } }
export interface ReplayLogEntry {
  readonly seq: number; readonly canonicalPayload: string; readonly replayInputs: unknown;
  readonly receiptJSON: string; readonly eventsJSON: string; readonly terminal: boolean;
}

interface ReplayCommand { intent_id: string; company_stream_id: string; founder_id: string; revision: number; run_seq: number; run_log_seq: number }
interface FounderReplayCommand { intent_id: string; founder_stream_id: string; founder_id: string; revision: number; founder_log_seq: number; server_ts_ms: number }
interface FounderReplayWire { v: 1; command: FounderReplayCommand; evaluated_at_ms: number; resolved: Record<string, unknown> }
interface FounderExitRecord { readonly run_id: number; readonly exit_type: string; readonly occurred_at_ms: number; readonly reputation_delta: number }
interface ReplayGuildSettlement { boundary_seq: number; debit_units: number; credit_units: number }
interface ReplayGuildSettlementBatch { guild_id: string; base_seq: number; settlements: ReplayGuildSettlement[] }
interface ReplayAccrual { contributions: ReplayContribution[]; commons_weight_ppm: number | null; guild_settlement_batch: ReplayGuildSettlementBatch; route_context_version: number }
interface ReplayWire { v: 2 | 3 | 4; command: ReplayCommand; evaluated_at_ms: number; evaluation_mode: "online" | "offline"; resolved: Record<string, unknown> }
interface NetworkSlot { readonly slot: string; readonly carried_ref: string }
interface FounderCarry {
  founder_revision: number; founder_constants_hash: string; reputation_level: number; route_knowledge_balance: number;
  age_ms: number; notoriety: number; advisor_mode: boolean; network_slots: NetworkSlot[]; ledger_fact_kinds: string[]; exit_history_count: number;
  achievements_earned_lifetime: string[]; achievement_score_lifetime: number;
}
interface ExitTerms { reputation_delta: number; network_slot_unlocks: NetworkSlot[]; route_knowledge: number; clout_reach_note: string }

export async function loadReplayCatalogBundle(constantsHash: string, artifacts: ReplayArtifacts): Promise<ReplayCatalogBundle> {
  const names = Object.keys(artifacts).sort(byteCompare).join("\0");
  const legacy = "categories\0commons\0economy\0factions\0guilds\0prestige\0routes";
  const active = "achievements\0categories\0commons\0economy\0factions\0guilds\0meters\0prestige\0routes";
  const minigamesActive = `${active}\0minigames`.split("\0").sort(byteCompare).join("\0");
  const petsActive = `${minigamesActive}\0pets`.split("\0").sort(byteCompare).join("\0");
  if (!hashPattern.test(constantsHash) || ![legacy, active, minigamesActive, petsActive].includes(names)) throw new SyntaxError("invalid replay artifact set");
  const computed = await constantsHashArtifacts(artifacts);
  if (computed !== constantsHash) throw new SyntaxError("replay artifact label mismatch");
  const economy = parseCatalog(parseJSON(artifacts.economy)); const routes = parseRoutesCatalog(parseJSON(artifacts.routes));
  const gateIds = routes.gates.map((gate) => gate.gateId);
  validateCategoryCatalog(parseJSON(artifacts.categories), gateIds);
  validateCatalogGateReferences(economy, gateIds);
  const commons = parseCommonsCatalog(parseJSON(artifacts.commons)); const prestige = parsePrestigePolicy(parseJSON(artifacts.prestige));
  const factions = parseFactionCatalog(parseJSON(artifacts.factions), commons.minimumTithePpm, commons.defaultTithePpm, commons.maximumTithePpm);
  const guilds = parseGuildCatalog(parseJSON(artifacts.guilds));
  if (names === legacy) return Object.freeze({ constantsHash, artifacts: Object.freeze({ ...artifacts }), economy, routes, commons, prestige, factions, guilds });
  const meters = loadMeterCatalog(artifacts.meters!);
  validateMeterResourceSeparation(meters, economy.resources.map((value) => value.id));
  const achievements = loadAchievementCatalog(artifacts.achievements!, foundationAchievementRegistry(economy));
  if (names === active) return Object.freeze({ constantsHash, artifacts: Object.freeze({ ...artifacts }), economy, routes, commons, prestige, factions, guilds, meters, achievements });
  const minigames = parseMinigameCatalog(parseJSON(artifacts.minigames!));
  if (names === minigamesActive) return Object.freeze({ constantsHash, artifacts: Object.freeze({ ...artifacts }), economy, routes, commons, prestige, factions, guilds, meters, achievements, minigames });
  const pets = parsePetCatalog(parseJSON(artifacts.pets!));
  return Object.freeze({ constantsHash, artifacts: Object.freeze({ ...artifacts }), economy, routes, commons, prestige, factions, guilds, meters, achievements, minigames, pets });
}

const REPLAY_EVENT_KINDS = Object.freeze([
	"achievement_earned.v1",
  "compact_cascade_started", "compact_health_band_changed", "compact_left", "compact_recovered", "compact_recruitment_offered", "compact_sampled", "compact_signed", "compact_tithe_raised", "compensation",
  "exit_offer_declined", "exit_offer_expired", "exit_offer_spawned", "faction_stock_saturated", "founder_advanced", "gate_crossed", "generator_purchased", "guild_activity_evaluated", "guild_tithe_accrued",
  "incorporated", "invariant_reported", "meter_band_changed.v1", "route_executed", "route_hint_purchased", "route_knowledge_granted", "run_ended", "run_started", "upgrade_purchased",
] as const);

function foundationAchievementRegistry(catalog: EconomyCatalog): AchievementRegistry {
  return {
    copyKeys: new Set(COPY_KEYS), generatorIds: new Set(catalog.generatorClasses.map((value) => value.id)), eventKinds: new Set(REPLAY_EVENT_KINDS), resourceIds: new Set(catalog.resources.map((value) => value.id)),
    runCounters: new Set(["generators_purchased_total", "tier"]), careerCounters: new Set(["age_ms", "notoriety"]),
    provenanceSources: new Map([
      ["counter:run:generators_purchased_total", ["generator_purchased"]], ["counter:run:tier", ["gate_crossed"]],
      ["counter:career:age_ms", ["founder_advanced"]], ["counter:career:notoriety", ["founder_advanced"]],
      ["exit_count", ["founder_advanced", "run_ended"]],
    ]),
  };
}

function foundationsActive(bundle: ReplayCatalogBundle): bundle is ReplayCatalogBundle & { readonly meters: MeterCatalog; readonly achievements: AchievementCatalog } {
  return bundle.meters !== undefined && bundle.achievements !== undefined;
}

function validateCategoryCatalog(source: unknown, routeGateIds: readonly string[]): void {
  const root = exactObject(source, ["schema_version", "full_gate_set", "fact_sets", "categories"], "category catalog");
  if (root.schema_version !== 1) throw new SyntaxError("invalid category schema");
  const gates = sortedUniqueCategoryMechanical(array(root.full_gate_set, "full gate set"));
  if (!same(gates, [...routeGateIds].sort(byteCompare))) throw new SyntaxError("category gate set differs from routes");
  const rawSets = exactObject(root.fact_sets, ["completion_set", "forbidden_set"], "category fact sets");
  const completion = sortedFactSet(rawSets.completion_set, false); const forbidden = sortedFactSet(rawSets.forbidden_set, true);
  if (!same(forbidden, ["darkpattern.", "externality."])) throw new SyntaxError("invalid forbidden fact set");
  const sets = new Set(["completion_set", "forbidden_set"]);
  const rows = new Map<string, { timer: string; predicate: CategoryPredicate }>();
  for (const item of array(root.categories, "categories")) {
    const row = exactObject(item, ["id", "name_key", "timer", "predicate"], "category");
    const id = mechanicalString(row.id); const nameKey = mechanicalString(row.name_key); const timer = string(row.timer);
    if (rows.has(id) || nameKey !== `category.${id}` || !["rta", "attended", "none"].includes(timer)) throw new SyntaxError("invalid category row");
    rows.set(id, { timer, predicate: parseCategoryPredicate(row.predicate, sets, 0) });
  }
  if (!same([...rows.keys()].sort(byteCompare), ["any_percent", "ethical_percent", "hundred_percent", "low_percent", "valuation"])) throw new SyntaxError("invalid canonical categories");
  const any = rows.get("any_percent")!; const ethical = rows.get("ethical_percent")!; const hundred = rows.get("hundred_percent")!;
  const low = rows.get("low_percent")!; const valuation = rows.get("valuation")!;
  if (completion.length !== 0 || any.timer !== "rta" || any.predicate.kind !== "any" || ethical.timer !== "attended" || ethical.predicate.kind !== "facts_disjoint" || ethical.predicate.setRef !== "forbidden_set" ||
      hundred.timer !== "rta" || hundred.predicate.kind !== "all_of" || hundred.predicate.all?.length !== 2 || hundred.predicate.all[0]?.kind !== "all_gates" || hundred.predicate.all[1]?.kind !== "facts_superset" || hundred.predicate.all[1]?.setRef !== "completion_set" ||
      low.timer !== "rta" || low.predicate.kind !== "count_at_most" || low.predicate.field !== "generators_purchased_total" || low.predicate.literal !== 40 || valuation.timer !== "none" || valuation.predicate.kind !== "any") {
    throw new SyntaxError("invalid canonical category shapes");
  }
}

interface CategoryPredicate { kind: string; setRef?: string; field?: string; literal?: number; all?: CategoryPredicate[] }
function parseCategoryPredicate(source: unknown, sets: ReadonlySet<string>, depth: number): CategoryPredicate {
  if (depth > 4) throw new SyntaxError("category predicate depth");
  if (!isRecord(source) || typeof source.kind !== "string") throw new SyntaxError("invalid category predicate");
  const kind = source.kind;
  if (kind === "any" || kind === "all_gates") { onlyKeys(source, ["kind"], "category predicate"); return { kind }; }
  if (kind === "facts_superset" || kind === "facts_disjoint") { onlyKeys(source, ["kind", "set_ref"], "category predicate"); const setRef = string(source.set_ref); if (!sets.has(setRef)) throw new SyntaxError("unknown category fact set"); return { kind, setRef }; }
  if (kind === "count_at_most") { onlyKeys(source, ["kind", "field", "literal"], "category predicate"); if (source.field !== "generators_purchased_total") throw new SyntaxError("invalid category count field"); return { kind, field: source.field, literal: safeInteger(source.literal, 0, MAX_EXACT_INTEGER) }; }
  if (kind === "all_of") { onlyKeys(source, ["kind", "all"], "category predicate"); const all = array(source.all, "category children"); if (all.length < 2 || all.length > 8) throw new SyntaxError("invalid category children"); return { kind, all: all.map((child) => parseCategoryPredicate(child, sets, depth + 1)) }; }
  throw new SyntaxError("invalid category predicate kind");
}

function sortedUniqueCategoryMechanical(source: unknown[]): string[] { let last = ""; return source.map((item) => { const value = mechanicalString(item); if (byteCompare(value, last) <= 0) throw new SyntaxError("values must be sorted and unique"); last = value; return value; }); }
function sortedFactSet(source: unknown, allowPrefixes: boolean): string[] { let last = ""; return array(source, "fact set").map((item) => { const value = string(item); const prefix = value.endsWith("."); const exact = prefix ? value.slice(0, -1) : value; const namespace = exact.split(".")[0]; if (byteCompare(value, last) <= 0 || prefix && !allowPrefixes || !mechanical.test(exact) || !["darkpattern", "exit", "externality"].includes(namespace ?? "")) throw new SyntaxError("invalid fact set"); last = value; return value; }); }

export function withNextReplayCatalogBundle(current: ReplayCatalogBundle, next: ReplayCatalogBundle): ReplayCatalogBundle {
  return Object.freeze({ ...current, next });
}

export async function verifyReplayRun(
  genesis: unknown,
  catalogs: ReplayCatalogBundle,
  entries: readonly ReplayLogEntry[],
  identity: { readonly constantsHash: string; readonly engineMismatch?: boolean; readonly genesisVersion?: number },
): Promise<ReplayVerdict> {
  return (await verifyReplayRunDetailed(genesis, catalogs, entries, identity)).verdict;
}

export async function verifyReplayRunDetailed(
  genesis: unknown,
  catalogs: ReplayCatalogBundle,
  entries: readonly ReplayLogEntry[],
  identity: { readonly constantsHash: string; readonly engineMismatch?: boolean; readonly genesisVersion?: number },
): Promise<ReplayVerificationResult> {
  if (identity.engineMismatch) return { verdict: "engine_mismatch", finalState: null };
  if (identity.constantsHash !== catalogs.constantsHash) return { verdict: "constants_mismatch", finalState: null };
  let state: ReplayState;
  try { state = restoreReplayState(genesis, identity.genesisVersion ?? 14, catalogs.economy, foundationsActive(catalogs) ? catalogs : undefined); }
  catch { return { verdict: "constants_mismatch", finalState: null }; }
  let terminal = false;
  for (let index = 0; index < entries.length; index++) {
    const entry = entries[index]!;
    if (entry.seq !== index + 1 || terminal) return { verdict: "log_gap", finalState: null };
    try {
      if (!isRecord(entry.replayInputs) || !isRecord(entry.replayInputs.command) || !isRecord(entry.replayInputs.resolved)) return { verdict: "state_divergence", finalState: null };
      if (entry.replayInputs.command.run_log_seq !== entry.seq) return { verdict: "log_gap", finalState: null };
      if (entry.replayInputs.resolved.kind === "exit") {
        const transition = await applyLoggedExit(state, entry.canonicalPayload, catalogs, entry.replayInputs);
        const events = [...transition.founderEvents, ...transition.companyEndedEvents, ...transition.companyStartedEvents];
        if (canonicalJSONString(transition.receipt) !== entry.receiptJSON || canonicalJSONString(events) !== entry.eventsJSON) return { verdict: "state_divergence", finalState: null };
        state = transition.finalCompany;
        terminal = transition.outcome === "applied";
      } else {
        const transition = await applyLogged(state, entry.canonicalPayload, catalogs, entry.replayInputs);
        if (canonicalJSONString(transition.receipt) !== entry.receiptJSON || canonicalJSONString(transition.events) !== entry.eventsJSON) return { verdict: "state_divergence", finalState: null };
        state = transition.state;
      }
    } catch (error) {
      return { verdict: error instanceof ReplayClockViolation ? "clock_violation" : "state_divergence", finalState: null };
    }
  }
  return terminal ? { verdict: "verified", finalState: state } : { verdict: "log_gap", finalState: null };
}

const saveV14Keys = ["balances", "generators", "generators_purchased_total", "upgrades_owned", "generators_provisioned", "provision_remainders_ppm", "stock_rate_remainder_ppm", "evaluated_through", "compute_credit_ms", "manual_token_milli", "manual_token_refilled_at", "gates_crossed", "run_seq", "doctrines_by_transition", "structure_id", "ledger_fact_kinds", "meter_bands", "region_traits", "route_knowledge_balance", "hints_unlocked", "compact_member", "compact_tithe_ppm", "compact_solidarity_ppm", "compact_solidarity_samples", "tier", "lifetime_value", "offer_state", "run_started_at_ms", "reputation_level", "reputation_unlock_ppm", "network_slots", "clout_lifetime", "soul", "age_ms", "notoriety", "advisor_mode", "exit_history", "run_pre_timer", "offline_spans", "collapsed_offline_ms", "faction_id", "incorporated_at_ms", "stock_units", "stock_progress_ms", "consumed_stock_units", "guild_tithe_carry_ppm", "guild_boundary_seq", "guild_consumed_window_units", "guild_boundary_guild_id"] as const;
const foundationSaveKeys = ["meter_values", "meter_decay_remainders", "meter_input_remainders", "achievements_earned_run", "achievement_score_run", "achievements_earned_lifetime", "achievement_score_lifetime"] as const;
const founderMinigameSaveKeys = ["minigame_ratings", "minigame_offline_quality"] as const;
const founderPetSaveKeys = ["pets"] as const;

export function restoreReplayState(source: unknown, version: number, catalog: EconomyCatalog, foundationCatalogs?: { readonly meters: MeterCatalog; readonly achievements: AchievementCatalog }): ReplayState {
  const requestedVersion = version;
  let foundationRaw: Record<string, unknown> | null = null;
  if (version === 16) {
    const activeKeys = [...saveV14Keys.filter((key) => key !== "meter_bands"), ...foundationSaveKeys];
    foundationRaw = exactObject(source, activeKeys, "save v16");
    source = { ...foundationRaw, meter_bands: {} };
    for (const key of foundationSaveKeys) delete (source as Record<string, unknown>)[key];
    version = 14;
  }
  if (version === 12) {
    if (!isRecord(source) || Object.hasOwn(source, "generators_purchased_total")) throw new SyntaxError("invalid save v12 envelope");
    const generators = integerRecord(source.generators, 0, MAX_EXACT_INTEGER, "generators");
    let purchased = 0;
    for (const count of Object.values(generators)) {
      if (count > MAX_EXACT_INTEGER - purchased) { purchased = MAX_EXACT_INTEGER; break; }
      purchased += count;
    }
    source = { ...source, generators_purchased_total: purchased };
  } else if (version !== 13 && version !== 14) throw new SyntaxError("unsupported replay save version");
  if (version < 14) {
    if (!isRecord(source)) throw new SyntaxError("invalid legacy replay save");
    const generators = integerRecord(source.generators, 0, MAX_EXACT_INTEGER, "generators");
    const provisionRemainders = Object.fromEntries(catalog.generatorClasses.filter((value) => value.provision !== null).map((value) => [value.provision!.generatorId, 0]));
    source = { ...source, upgrades_owned: [], generators_provisioned: Object.fromEntries(Object.keys(generators).map((id) => [id, 0])), provision_remainders_ppm: provisionRemainders, stock_rate_remainder_ppm: 0 };
  }
  const raw = exactObject(source, saveV14Keys, "save v14");
  const balances = stringRecord(raw.balances, "balances"); const expectedResources = catalog.resources.filter((value) => value.scope === "company").map((value) => value.id).sort(byteCompare);
  if (!same(Object.keys(balances).sort(byteCompare), expectedResources)) throw new SyntaxError("company balance set differs from catalog");
  for (const definition of catalog.resources.filter((value) => value.scope === "company")) validateBalance(balances[definition.id]!, definition.minimum, definition.hardcap?.amount);
  const generators = integerRecord(raw.generators, 0, MAX_EXACT_INTEGER, "generators"); const expectedGenerators = catalog.generatorClasses.filter((value) => value.production !== null).map((value) => value.id).sort(byteCompare);
  if (!same(Object.keys(generators).sort(byteCompare), expectedGenerators)) throw new SyntaxError("generator set differs from catalog");
  const upgradesOwned = mechanicalSet(raw.upgrades_owned);
  for (const id of upgradesOwned) if (!catalog.upgrade(id)) throw new SyntaxError(`unknown owned upgrade ${id}`);
  const generatorsProvisioned = integerRecord(raw.generators_provisioned, 0, MAX_EXACT_INTEGER, "generators_provisioned");
  if (!same(Object.keys(generatorsProvisioned).sort(byteCompare), expectedGenerators)) throw new SyntaxError("provisioned generator set differs from catalog");
  for (const generator of catalog.generatorClasses.filter((value) => value.production !== null)) {
    const count = generatorsProvisioned[generator.id]!;
    if (generator.provisionedHardcap === null ? count !== 0 : count > generator.provisionedHardcap.count) throw new SyntaxError(`provisioned generator ${generator.id} exceeds its catalog cap`);
  }
  const expectedRemainders = catalog.generatorClasses.filter((value) => value.provision !== null).map((value) => value.provision!.generatorId).sort(byteCompare);
  const provisionRemaindersPpm = integerRecord(raw.provision_remainders_ppm, 0, 999_999, "provision_remainders_ppm");
  if (!same(Object.keys(provisionRemaindersPpm).sort(byteCompare), expectedRemainders)) throw new SyntaxError("provision remainder set differs from catalog");
  const evaluatedThroughMs = cursor(raw.evaluated_through, "evaluated_through"); const manualTokenRefilledAtMs = cursor(raw.manual_token_refilled_at, "manual_token_refilled_at");
  if (manualTokenRefilledAtMs > evaluatedThroughMs) throw new SyntaxError("manual cursor exceeds evaluation cursor");
  const compactSamples = array(raw.compact_solidarity_samples, "compact samples").map((item) => { const value = exactObject(item, ["hour_start", "compliance_ppm", "covered_ms"], "compact sample"); return { hourStartMs: cursor(value.hour_start, "hour_start"), compliancePpm: safeInteger(value.compliance_ppm, 0, 1_000_000), coveredMs: safeInteger(value.covered_ms, 1, 3_600_000) }; });
  const offlineSpans = array(raw.offline_spans, "offline_spans").map((item) => { const value = exactObject(item, ["from_ms", "to_ms"], "offline span"); const fromMs = safeInteger(value.from_ms, 1, MAX_EXACT_INTEGER); const toMs = safeInteger(value.to_ms, 1, MAX_EXACT_INTEGER); if (toMs <= fromMs) throw new SyntaxError("invalid offline span"); return { fromMs, toMs }; });
  let offerState: ReplayState["offerState"] = null;
  if (raw.offer_state !== null) { const value = exactObject(raw.offer_state, ["offer_id", "exit_type", "terms_json", "spawned_at_ms", "expires_at_ms"], "offer state"); if (typeof value.offer_id !== "string" || !uuidV7.test(value.offer_id) || typeof value.exit_type !== "string") throw new SyntaxError("invalid offer"); offerState = { offerId: value.offer_id, exitType: value.exit_type, terms: value.terms_json, spawnedAtMs: safeInteger(value.spawned_at_ms, 1, MAX_EXACT_INTEGER), expiresAtMs: safeInteger(value.expires_at_ms, 1, MAX_EXACT_INTEGER) }; }
  const factionId = nullableMechanical(raw.faction_id); const incorporatedAtMs = raw.incorporated_at_ms === null ? null : safeInteger(raw.incorporated_at_ms, 1, MAX_EXACT_INTEGER);
  const state: ReplayState = { wireVersion: requestedVersion === 16 ? 16 : 14, balances, generators, generatorPurchasedTotal: safeInteger(raw.generators_purchased_total, 0, MAX_EXACT_INTEGER), upgradesOwned, generatorsProvisioned, provisionRemaindersPpm, stockRateRemainderPpm: safeInteger(raw.stock_rate_remainder_ppm, 0, 999_999), evaluatedThroughMs, computeCreditMs: safeInteger(raw.compute_credit_ms, 0, MAX_EXACT_INTEGER), manualTokenMilli: safeInteger(raw.manual_token_milli, 0, MAX_EXACT_INTEGER), manualTokenRefilledAtMs,
    gatesCrossed: booleanRecord(raw.gates_crossed), runSeq: safeInteger(raw.run_seq, 1, MAX_EXACT_INTEGER), doctrinesByTransition: mechanicalRecord(raw.doctrines_by_transition), structureId: nullableMechanical(raw.structure_id, false),
    ledgerFactKinds: mechanicalSet(raw.ledger_fact_kinds), meterBands: integerRecord(raw.meter_bands, 0, 100, "meter_bands"), regionTraits: mechanicalSet(raw.region_traits), routeKnowledgeBalance: safeInteger(raw.route_knowledge_balance, 0, MAX_EXACT_INTEGER), hintsUnlocked: mechanicalSet(raw.hints_unlocked),
    compactMember: boolean(raw.compact_member), compactTithePpm: safeInteger(raw.compact_tithe_ppm, 0, 1_000_000), compactSolidarityPpm: safeInteger(raw.compact_solidarity_ppm, 0, 1_000_000), compactSamples,
    tier: safeInteger(raw.tier, 0, 9), lifetimeValue: canonical(raw.lifetime_value), offerState, runStartedAtMs: safeInteger(raw.run_started_at_ms, 0, MAX_EXACT_INTEGER), runPreTimer: boolean(raw.run_pre_timer), offlineSpans, collapsedOfflineMs: safeInteger(raw.collapsed_offline_ms, 0, MAX_EXACT_INTEGER),
    factionId, incorporatedAtMs, factionStockResource: "", stockUnits: safeInteger(raw.stock_units, 0, MAX_EXACT_INTEGER), stockProgressMs: safeInteger(raw.stock_progress_ms, 0, MAX_EXACT_INTEGER), consumedStockUnits: safeInteger(raw.consumed_stock_units, 0, MAX_EXACT_INTEGER),
    guildTitheCarryPpm: safeInteger(raw.guild_tithe_carry_ppm, 0, 999_999), guildBoundaryGuildId: nullableUUID(raw.guild_boundary_guild_id), guildBoundarySeq: safeInteger(raw.guild_boundary_seq, 0, MAX_EXACT_INTEGER), guildConsumedWindow: safeInteger(raw.guild_consumed_window_units, 0, MAX_EXACT_INTEGER),
    meterValues: {}, meterDecayRemainders: {}, meterInputRemainders: {}, achievementsEarnedRun: new Set(), achievementScoreRun: 0 };
  if (safeInteger(raw.reputation_level, 0, MAX_EXACT_INTEGER) !== 0 || safeInteger(raw.reputation_unlock_ppm, 0, 1_000_000) !== 0 || array(raw.network_slots, "network_slots").length !== 0 || safeInteger(raw.clout_lifetime, 0, MAX_EXACT_INTEGER) !== 0 || safeInteger(raw.soul, -MAX_EXACT_INTEGER, MAX_EXACT_INTEGER) !== 0 || safeInteger(raw.age_ms, 0, MAX_EXACT_INTEGER) !== 0 || safeInteger(raw.notoriety, 0, MAX_EXACT_INTEGER) !== 0 || boolean(raw.advisor_mode) || array(raw.exit_history, "exit_history").length !== 0) throw new SyntaxError("founder prestige state leaked into company scope");
  if (state.routeKnowledgeBalance !== 0 || state.hintsUnlocked.size !== 0 || Object.values(state.gatesCrossed).some((value) => !value)) throw new SyntaxError("invalid company route state");
  if (parseCanonical(state.lifetimeValue).lt(0)) throw new SyntaxError("negative lifetime value");
  if (state.computeCreditMs > (catalog.offlinePolicy?.bankCapMs ?? 0) || state.manualTokenMilli > (catalog.manualPolicy?.bucketCapMilli ?? 0)) throw new SyntaxError("production counter exceeds catalog cap");
  if (!state.compactMember && (state.compactTithePpm !== 0 || state.compactSolidarityPpm !== 0 || state.compactSamples.length !== 0)) throw new SyntaxError("non-member has compact state");
  let lastSample = -1;
  for (const sample of state.compactSamples) { if (sample.hourStartMs % 3_600_000 !== 0 || sample.hourStartMs <= lastSample) throw new SyntaxError("invalid compact sample order"); lastSample = sample.hourStartMs; }
  if (state.offerState && (state.offerState.expiresAtMs <= state.offerState.spawnedAtMs || !["acquihire", "acquisition", "ipo", "collapse", "scripted_first"].includes(state.offerState.exitType))) throw new SyntaxError("invalid offer state");
  if (state.runPreTimer && state.runStartedAtMs === 0 || state.runStartedAtMs > state.evaluatedThroughMs) throw new SyntaxError("invalid run start");
  let lastOfflineEnd = -1;
  let totalOffline = state.collapsedOfflineMs;
  for (const span of state.offlineSpans) { if (span.fromMs < state.runStartedAtMs || span.toMs > state.evaluatedThroughMs || span.fromMs < lastOfflineEnd) throw new SyntaxError("invalid offline span order"); totalOffline += span.toMs - span.fromMs; if (!Number.isSafeInteger(totalOffline) || totalOffline > MAX_EXACT_INTEGER) throw new SyntaxError("offline total exceeds exact domain"); lastOfflineEnd = span.toMs; }
  if (state.runStartedAtMs === 0 && totalOffline !== 0 || state.runStartedAtMs !== 0 && totalOffline > state.evaluatedThroughMs - state.runStartedAtMs) throw new SyntaxError("offline total exceeds run duration");
  if (state.factionId === "" && (state.incorporatedAtMs !== null || state.stockUnits !== 0 || state.stockProgressMs !== 0 || state.consumedStockUnits !== 0)) throw new SyntaxError("orphan faction stock state");
  if (state.factionId !== "" && (state.incorporatedAtMs === null || state.incorporatedAtMs > state.evaluatedThroughMs)) throw new SyntaxError("invalid faction incorporation");
  if (state.guildBoundaryGuildId === "" && state.guildBoundarySeq !== 0) throw new SyntaxError("invalid guild watermark");
  if (requestedVersion === 16) {
    if (!foundationCatalogs || !foundationRaw) throw new SyntaxError("save v16 requires pinned foundation catalogs");
    state.meterValues = integerRecord(foundationRaw.meter_values, 0, 100, "meter_values");
    state.meterDecayRemainders = integerRecord(foundationRaw.meter_decay_remainders, 0, 3_599_999, "meter_decay_remainders");
    state.meterInputRemainders = plainIntegerRecord(foundationRaw.meter_input_remainders, 0, 3_599_999, "meter_input_remainders");
    validateMeterState(foundationCatalogs.meters, { values: state.meterValues, decayRemainders: state.meterDecayRemainders, inputRemainders: state.meterInputRemainders });
    state.achievementsEarnedRun = new Set(sortedUniqueMechanical(array(foundationRaw.achievements_earned_run, "run achievements")));
    state.achievementScoreRun = safeInteger(foundationRaw.achievement_score_run, 0, MAX_EXACT_INTEGER);
    if (array(foundationRaw.achievements_earned_lifetime, "company lifetime achievements").length !== 0 || safeInteger(foundationRaw.achievement_score_lifetime, 0, MAX_EXACT_INTEGER) !== 0 || achievementScore(foundationCatalogs.achievements, state.achievementsEarnedRun) !== state.achievementScoreRun) throw new SyntaxError("invalid company achievement state");
  }
  return state;
}

export function encodeReplayStateV14(state: ReplayState): unknown {
  return { balances: sortedRecord(state.balances), generators: sortedRecord(state.generators), generators_purchased_total: state.generatorPurchasedTotal,
    upgrades_owned: [...state.upgradesOwned].sort(byteCompare), generators_provisioned: sortedRecord(state.generatorsProvisioned), provision_remainders_ppm: sortedRecord(state.provisionRemaindersPpm), stock_rate_remainder_ppm: state.stockRateRemainderPpm,
    evaluated_through: rfc3339(state.evaluatedThroughMs),
    compute_credit_ms: state.computeCreditMs, manual_token_milli: state.manualTokenMilli, manual_token_refilled_at: rfc3339(state.manualTokenRefilledAtMs),
    gates_crossed: sortedRecord(state.gatesCrossed), run_seq: state.runSeq, doctrines_by_transition: sortedRecord(state.doctrinesByTransition), structure_id: state.structureId,
    ledger_fact_kinds: [...state.ledgerFactKinds].sort(byteCompare), meter_bands: sortedRecord(state.meterBands), region_traits: [...state.regionTraits].sort(byteCompare),
    route_knowledge_balance: state.routeKnowledgeBalance, hints_unlocked: [...state.hintsUnlocked].sort(byteCompare), compact_member: state.compactMember,
    compact_tithe_ppm: state.compactTithePpm, compact_solidarity_ppm: state.compactSolidarityPpm,
    compact_solidarity_samples: state.compactSamples.map((value) => ({ hour_start: rfc3339(value.hourStartMs), compliance_ppm: value.compliancePpm, covered_ms: value.coveredMs })),
    tier: state.tier, lifetime_value: state.lifetimeValue,
    offer_state: state.offerState ? { offer_id: state.offerState.offerId, exit_type: state.offerState.exitType, terms_json: state.offerState.terms, spawned_at_ms: state.offerState.spawnedAtMs, expires_at_ms: state.offerState.expiresAtMs } : null,
    run_started_at_ms: state.runStartedAtMs, reputation_level: 0, reputation_unlock_ppm: 0, network_slots: [], clout_lifetime: 0, soul: 0, age_ms: 0,
    notoriety: 0, advisor_mode: false, exit_history: [], run_pre_timer: state.runPreTimer,
    offline_spans: state.offlineSpans.map((value) => ({ from_ms: value.fromMs, to_ms: value.toMs })), collapsed_offline_ms: state.collapsedOfflineMs,
    faction_id: state.factionId || null, incorporated_at_ms: state.incorporatedAtMs, stock_units: state.stockUnits, stock_progress_ms: state.stockProgressMs,
    consumed_stock_units: state.consumedStockUnits, guild_tithe_carry_ppm: state.guildTitheCarryPpm, guild_boundary_seq: state.guildBoundarySeq,
    guild_consumed_window_units: state.guildConsumedWindow, guild_boundary_guild_id: state.guildBoundaryGuildId || null };
}

export function encodeReplayState(state: ReplayState): unknown {
  const base = encodeReplayStateV14(state) as Record<string, unknown>;
  if (state.wireVersion !== 16) return base;
  delete base.meter_bands;
  return { ...base, meter_values: sortedRecord(state.meterValues), meter_decay_remainders: sortedRecord(state.meterDecayRemainders), meter_input_remainders: sortedRecord(state.meterInputRemainders), achievements_earned_run: [...state.achievementsEarnedRun].sort(byteCompare), achievement_score_run: state.achievementScoreRun, achievements_earned_lifetime: [], achievement_score_lifetime: 0 };
}

export function restoreFounderReplayState(source: unknown, version: number, catalogs: ReplayCatalogBundle): FounderReplayState {
  const requestedVersion = version;
  let foundationRaw: Record<string, unknown> | null = null;
  if (version >= 15 && version <= 18) {
    const activeKeys = [...saveV14Keys.filter((key) => key !== "meter_bands"), ...foundationSaveKeys.slice(0, version === 15 ? 3 : foundationSaveKeys.length), ...(version >= 17 ? founderMinigameSaveKeys : []), ...(version >= 18 ? founderPetSaveKeys : [])];
    foundationRaw = exactObject(source, activeKeys, "Founder save v16");
    source = { ...foundationRaw, meter_bands: {} };
    for (const key of [...foundationSaveKeys, ...founderMinigameSaveKeys, ...founderPetSaveKeys]) delete (source as Record<string, unknown>)[key];
    version = 14;
  }
  if (version !== 14) throw new SyntaxError("unsupported Founder replay save version");
  const raw = exactObject(source, saveV14Keys, "Founder save v14");
  const founderResources = catalogs.economy.resources.filter((value) => value.scope === "founder");
  const balances = stringRecord(raw.balances, "Founder balances");
  if (!same(Object.keys(balances).sort(byteCompare), founderResources.map((value) => value.id).sort(byteCompare))) throw new SyntaxError("Founder balance set differs from catalog");
  for (const resource of founderResources) validateBalance(balances[resource.id]!, resource.minimum, resource.hardcap?.amount);
  const founderGenerators = catalogs.economy.generatorClasses.filter((value) => catalogs.economy.resource(value.price.resourceId)?.scope === "founder");
  const generatorIds = founderGenerators.map((value) => value.id).sort(byteCompare);
  const generators = integerRecord(raw.generators, 0, MAX_EXACT_INTEGER, "Founder generators");
  if (!same(Object.keys(generators).sort(byteCompare), generatorIds)) throw new SyntaxError("Founder generator set differs from catalog");
  const generatorsProvisioned = integerRecord(raw.generators_provisioned, 0, MAX_EXACT_INTEGER, "Founder provisioned generators");
  if (!same(Object.keys(generatorsProvisioned).sort(byteCompare), generatorIds)) throw new SyntaxError("Founder provisioned set differs from catalog");
  const upgradesOwned = mechanicalSet(raw.upgrades_owned);
  for (const id of upgradesOwned) if (catalogs.economy.resource(catalogs.economy.upgrade(id)?.cost.resourceId ?? "")?.scope !== "founder") throw new SyntaxError("non-Founder upgrade in Founder state");
  const provisionRemaindersPpm = integerRecord(raw.provision_remainders_ppm, 0, 999_999, "Founder provision remainders");
  const expectedRemainders = founderGenerators.filter((value) => value.provision !== null).map((value) => value.provision!.generatorId).sort(byteCompare);
  if (!same(Object.keys(provisionRemaindersPpm).sort(byteCompare), expectedRemainders)) throw new SyntaxError("Founder provision remainder set differs from catalog");
  const evaluatedThroughMs = cursor(raw.evaluated_through, "Founder evaluated_through");
  const manualTokenRefilledAtMs = cursor(raw.manual_token_refilled_at, "Founder manual_token_refilled_at");
  if (manualTokenRefilledAtMs > evaluatedThroughMs || safeInteger(raw.compute_credit_ms, 0, MAX_EXACT_INTEGER) !== 0 || safeInteger(raw.manual_token_milli, 0, MAX_EXACT_INTEGER) !== 0) throw new SyntaxError("Company production state leaked into Founder scope");
  if (Object.keys(booleanRecord(raw.gates_crossed)).length !== 0 || safeInteger(raw.run_seq, 0, MAX_EXACT_INTEGER) !== 0 || Object.keys(mechanicalRecord(raw.doctrines_by_transition)).length !== 0 || nullableMechanical(raw.structure_id, false) !== "" || Object.keys(integerRecord(raw.meter_bands, 0, 100, "Founder meter bands")).length !== 0 || mechanicalSet(raw.region_traits).size !== 0) throw new SyntaxError("Company route context leaked into Founder scope");
  if (boolean(raw.compact_member) || safeInteger(raw.compact_tithe_ppm, 0, 1_000_000) !== 0 || safeInteger(raw.compact_solidarity_ppm, 0, 1_000_000) !== 0 || array(raw.compact_solidarity_samples, "Founder compact samples").length !== 0) throw new SyntaxError("Company compact state leaked into Founder scope");
  if (safeInteger(raw.tier, 0, 9) !== 0 || canonical(raw.lifetime_value) !== "0" || raw.offer_state !== null || safeInteger(raw.run_started_at_ms, 0, MAX_EXACT_INTEGER) !== 0 || boolean(raw.run_pre_timer) || array(raw.offline_spans, "Founder offline spans").length !== 0 || safeInteger(raw.collapsed_offline_ms, 0, MAX_EXACT_INTEGER) !== 0) throw new SyntaxError("Company prestige state leaked into Founder scope");
  if (raw.faction_id !== null || raw.incorporated_at_ms !== null || safeInteger(raw.stock_units, 0, MAX_EXACT_INTEGER) !== 0 || safeInteger(raw.stock_progress_ms, 0, MAX_EXACT_INTEGER) !== 0 || safeInteger(raw.consumed_stock_units, 0, MAX_EXACT_INTEGER) !== 0 || safeInteger(raw.stock_rate_remainder_ppm, 0, 999_999) !== 0) throw new SyntaxError("Company faction state leaked into Founder scope");
  if (safeInteger(raw.guild_tithe_carry_ppm, 0, 999_999) !== 0 || safeInteger(raw.guild_boundary_seq, 0, MAX_EXACT_INTEGER) !== 0 || safeInteger(raw.guild_consumed_window_units, 0, MAX_EXACT_INTEGER) !== 0 || raw.guild_boundary_guild_id !== null) throw new SyntaxError("Company guild state leaked into Founder scope");
  const networkSlots = parseFounderNetworkSlots(raw.network_slots);
  const exitHistory = parseFounderExitHistory(raw.exit_history);
  const earnedLifetime = requestedVersion >= 16 ? new Set(sortedUniqueMechanical(array(foundationRaw!.achievements_earned_lifetime, "Founder lifetime achievements"))) : new Set<string>();
  const lifetimeScore = requestedVersion >= 16 ? safeInteger(foundationRaw!.achievement_score_lifetime, 0, MAX_EXACT_INTEGER) : 0;
  if (requestedVersion >= 15) {
    if (!catalogs.meters || Object.keys(integerRecord(foundationRaw!.meter_values, 0, 100, "Founder meter values")).length !== 0 || Object.keys(integerRecord(foundationRaw!.meter_decay_remainders, 0, 3_599_999, "Founder meter decay")).length !== 0 || Object.keys(plainIntegerRecord(foundationRaw!.meter_input_remainders, 0, 3_599_999, "Founder meter inputs")).length !== 0) throw new SyntaxError("invalid Founder meter state");
  }
  if (requestedVersion >= 16) {
    if (!catalogs.achievements || array(foundationRaw!.achievements_earned_run, "Founder run achievements").length !== 0 || safeInteger(foundationRaw!.achievement_score_run, 0, MAX_EXACT_INTEGER) !== 0 || achievementScore(catalogs.achievements, earnedLifetime) !== lifetimeScore) throw new SyntaxError("invalid active Founder achievement state");
  }
  let minigameRatings: FounderReplayState["minigameRatings"] = {}, minigameOfflineQuality: FounderReplayState["minigameOfflineQuality"] = {}, pets: Record<string, PetCareState> = {};
  if (requestedVersion >= 17) {
    if (!catalogs.minigames) throw new SyntaxError("Founder v17 requires minigames artifact");
    const ratingsRaw = exactRecord(foundationRaw!.minigame_ratings, catalogs.minigames.minigameIds, "minigame ratings");
    const qualityRaw = exactRecord(foundationRaw!.minigame_offline_quality, catalogs.minigames.minigameIds, "minigame offline quality");
    minigameRatings = Object.fromEntries(catalogs.minigames.minigameIds.map((id) => { const row = exactObject(ratingsRaw[id], ["elo", "season_member", "games_counted"], "minigame rating"); const season = mechanicalString(row.season_member); if (!catalogs.minigames!.ratingSeasons.includes(season)) throw new SyntaxError("unknown rating season"); return [id, { elo: safeInteger(row.elo, -MAX_EXACT_INTEGER, MAX_EXACT_INTEGER), season_member: season, games_counted: safeInteger(row.games_counted, 0, MAX_EXACT_INTEGER) }]; }));
    minigameOfflineQuality = Object.fromEntries(catalogs.minigames.minigameIds.map((id) => { const row = exactObject(qualityRaw[id], ["grade_ppm", "last_founder_attended_ms", "decay_remainder_ppm"], "minigame offline quality"); return [id, { grade_ppm: safeInteger(row.grade_ppm, 0, 1_000_000), last_founder_attended_ms: safeInteger(row.last_founder_attended_ms, 0, MAX_EXACT_INTEGER), decay_remainder_ppm: safeInteger(row.decay_remainder_ppm, 0, 999_999) }]; }));
  } else if (catalogs.minigames) throw new SyntaxError("minigames artifact requires Founder v17+");
  if (requestedVersion >= 18) {
    if (!catalogs.pets) throw new SyntaxError("Founder v18 requires pets artifact");
    pets = parsePetCareStates(foundationRaw!.pets, { action_ids: catalogs.pets.actions.map((row) => row.action_id), behavior_ids: ["active", "care_response", "idle", "resting"] });
    validatePetCareStatesForCatalog(pets, catalogs.pets);
  } else if (catalogs.pets) throw new SyntaxError("pets artifact requires Founder v18");
  return { wireVersion: requestedVersion as 14 | 15 | 16 | 17 | 18, balances, generators, generatorPurchasedTotal: safeInteger(raw.generators_purchased_total, 0, MAX_EXACT_INTEGER), upgradesOwned, generatorsProvisioned, provisionRemaindersPpm,
    stockRateRemainderPpm: 0, evaluatedThroughMs, computeCreditMs: 0, manualTokenMilli: 0, manualTokenRefilledAtMs,
    routeKnowledgeBalance: safeInteger(raw.route_knowledge_balance, 0, MAX_EXACT_INTEGER), hintsUnlocked: mechanicalSet(raw.hints_unlocked), ledgerFactKinds: mechanicalSet(raw.ledger_fact_kinds),
    reputationLevel: safeInteger(raw.reputation_level, 0, MAX_EXACT_INTEGER), networkSlots, cloutLifetime: safeInteger(raw.clout_lifetime, 0, MAX_EXACT_INTEGER), soul: safeInteger(raw.soul, -MAX_EXACT_INTEGER, MAX_EXACT_INTEGER),
    ageMs: safeInteger(raw.age_ms, 0, MAX_EXACT_INTEGER), notoriety: safeInteger(raw.notoriety, 0, MAX_EXACT_INTEGER), advisorMode: boolean(raw.advisor_mode), exitHistory,
    achievementsEarnedLifetime: earnedLifetime, achievementScoreLifetime: lifetimeScore, minigameRatings, minigameOfflineQuality, pets };
}

export function encodeFounderReplayState(state: FounderReplayState): unknown {
  const base: Record<string, unknown> = {
    balances: sortedRecord(state.balances), generators: sortedRecord(state.generators), generators_purchased_total: state.generatorPurchasedTotal,
    upgrades_owned: [...state.upgradesOwned].sort(byteCompare), generators_provisioned: sortedRecord(state.generatorsProvisioned), provision_remainders_ppm: sortedRecord(state.provisionRemaindersPpm), stock_rate_remainder_ppm: state.stockRateRemainderPpm,
    evaluated_through: rfc3339(state.evaluatedThroughMs), compute_credit_ms: 0, manual_token_milli: 0, manual_token_refilled_at: rfc3339(state.manualTokenRefilledAtMs),
    gates_crossed: {}, run_seq: 0, doctrines_by_transition: {}, structure_id: "", ledger_fact_kinds: [...state.ledgerFactKinds].sort(byteCompare), meter_bands: {}, region_traits: [],
    route_knowledge_balance: state.routeKnowledgeBalance, hints_unlocked: [...state.hintsUnlocked].sort(byteCompare), compact_member: false, compact_tithe_ppm: 0, compact_solidarity_ppm: 0, compact_solidarity_samples: [],
    tier: 0, lifetime_value: "0", offer_state: null, run_started_at_ms: 0, reputation_level: state.reputationLevel, reputation_unlock_ppm: 0, network_slots: [...state.networkSlots],
    clout_lifetime: state.cloutLifetime, soul: state.soul, age_ms: state.ageMs, notoriety: state.notoriety, advisor_mode: state.advisorMode, exit_history: [...state.exitHistory], run_pre_timer: false,
    offline_spans: [], collapsed_offline_ms: 0, faction_id: null, incorporated_at_ms: null, stock_units: 0, stock_progress_ms: 0, consumed_stock_units: 0,
    guild_tithe_carry_ppm: 0, guild_boundary_seq: 0, guild_consumed_window_units: 0, guild_boundary_guild_id: null,
  };
  if (state.wireVersion === 14) return base;
  delete base.meter_bands;
  if (state.wireVersion === 15) return { ...base, meter_values: {}, meter_decay_remainders: {}, meter_input_remainders: {} };
  const active: Record<string, unknown> = { ...base, meter_values: {}, meter_decay_remainders: {}, meter_input_remainders: {}, achievements_earned_run: [], achievement_score_run: 0,
    achievements_earned_lifetime: [...state.achievementsEarnedLifetime].sort(byteCompare), achievement_score_lifetime: state.achievementScoreLifetime };
  if (state.wireVersion >= 17) Object.assign(active, { minigame_ratings: sortedRecord(state.minigameRatings), minigame_offline_quality: sortedRecord(state.minigameOfflineQuality) });
  if (state.wireVersion >= 18) Object.assign(active, { pets: sortedRecord(state.pets) });
  return active;
}

export function applyFounderLogged(state: FounderReplayState, canonicalPayload: string, catalogs: ReplayCatalogBundle, replayInputs: unknown): FounderLoggedTransition {
  if (!hashPattern.test(catalogs.constantsHash)) throw new SyntaxError("invalid Founder catalog bundle");
  const wire = parseFounderReplayWire(replayInputs); const request = parseIntent(canonicalPayload, wire.command.intent_id);
  const before = restoreFounderReplayState(encodeFounderReplayState(state), state.wireVersion, catalogs);
  const rollback = (): void => { Object.assign(state, before); };
  try {
    const kind = string(wire.resolved.kind);
    if (kind === "invalid") {
      onlyKeys(wire.resolved, ["kind", "detail"], "invalid Founder inputs");
      if (request.invalid === undefined || string(wire.resolved.detail) !== request.invalid) throw new RangeError("invalid Founder arm mismatch");
      return founderRejected(state, wire.command.intent_id, wire.command.revision, "invalid", request.invalid, catalogs.constantsHash);
    }
    if (kind === "buy_route_hint") {
      onlyKeys(wire.resolved, ["kind", "route_context_version", "route_knowledge_balance"], "route hint inputs");
      if (request.kind !== "buy_route_hint" || request.invalid !== undefined || request.expected_revision !== wire.command.revision || safeInteger(wire.resolved.route_context_version, 1, MAX_EXACT_INTEGER) !== catalogs.routes.contextVersion) throw new RangeError("route hint command mismatch");
      state.routeKnowledgeBalance = safeInteger(wire.resolved.route_knowledge_balance, 0, MAX_EXACT_INTEGER);
      const routeId = string(request.route_id); const route = catalogs.routes.route(routeId);
      if (!route) return founderRejected(state, request.intent_id, wire.command.revision, "unknown_id", routeId, catalogs.constantsHash, rollback);
      if (state.hintsUnlocked.has(routeId)) return founderRejected(state, request.intent_id, wire.command.revision, "already_unlocked", routeId, catalogs.constantsHash, rollback);
      if (state.routeKnowledgeBalance < catalogs.routes.knowledge.hintCost) return founderRejected(state, request.intent_id, wire.command.revision, "insufficient_route_knowledge", routeId, catalogs.constantsHash, rollback);
      state.routeKnowledgeBalance -= catalogs.routes.knowledge.hintCost; state.hintsUnlocked.add(routeId);
      const eventValue = event("route_hint_purchased", request.intent_id, { route_id: routeId, cost: catalogs.routes.knowledge.hintCost });
      const receipt = { applied_count: 1, evaluated_at: rfc3339(state.evaluatedThroughMs), intent_id: request.intent_id, new_revision: wire.command.revision + 1, outcome: "applied", receipt: { changes: [] }, snapshot: founderWireSnapshot(state) };
      return { state, outcome: "applied", receipt, events: [eventValue], resultConstantsHash: catalogs.constantsHash };
    }
    if (kind === "exit.v1") return applyFounderExit(state, request, wire, catalogs);
    throw new RangeError("unknown Founder replay arm");
  } catch (error) { rollback(); throw error; }
}

export function verifyFounderReplayHistory(genesis: unknown, genesisRevision: number, genesisVersion: 14 | 15 | 16 | 17 | 18, genesisHash: string,
  founderStreamId: string, founderId: string, entries: readonly FounderReplayLogEntry[], head: FounderReplayHead,
  bundles: readonly ReplayCatalogBundle[]): ReplayVerdict {
  const catalogs = new Map(bundles.map((value) => [value.constantsHash, value]));
  if (entries.length === 0 || !uuid.test(founderStreamId) || !uuid.test(founderId)) return "log_gap";
  let bundle = catalogs.get(genesisHash); if (!bundle) return "constants_mismatch";
  let state: FounderReplayState;
  try { state = restoreFounderReplayState(genesis, genesisVersion, bundle); } catch { return "state_divergence"; }
  let revision: number;
  try { revision = safeInteger(genesisRevision, 1, MAX_EXACT_INTEGER); } catch { return "log_gap"; }
  let hash = genesisHash;
  for (let index = 0; index < entries.length; index++) {
    const entry = entries[index]!;
    if (entry.seq !== index + 1 || entry.constantsHash !== hash) return "log_gap";
    bundle = catalogs.get(hash)!; if (!bundle) return "constants_mismatch";
    try {
      const wire = parseFounderReplayWire(entry.replayInputs);
      if (wire.command.intent_id !== entry.intentId || wire.command.founder_stream_id !== founderStreamId || wire.command.founder_id !== founderId || wire.command.revision !== revision || wire.command.founder_log_seq !== entry.seq || wire.command.server_ts_ms !== entry.serverTSMS) return "log_gap";
      const exitArm = wire.resolved.kind === "exit.v1";
      if (exitArm !== (entry.source !== null)) return "state_divergence";
      if (entry.source) {
        if (wire.resolved.company_stream_id !== entry.source.companyStreamId || wire.resolved.run_seq !== entry.source.runSeq || wire.resolved.run_log_seq !== entry.source.runLogSeq) return "state_divergence";
      }
      let executionBundle = bundle;
      if (exitArm && wire.resolved.result_constants_hash !== hash) {
        const next = catalogs.get(string(wire.resolved.result_constants_hash)); if (!next) return "constants_mismatch";
        executionBundle = withNextReplayCatalogBundle(bundle, next);
      }
      const transition = applyFounderLogged(state, entry.canonicalPayload, executionBundle, entry.replayInputs);
      if (canonicalJSONString(transition.receipt) !== entry.receiptJSON || canonicalJSONString(transition.events) !== entry.eventsJSON) return "state_divergence";
      if (transition.outcome === "applied") { if (entry.appliedRevision !== revision + 1) return "log_gap"; revision++; }
      else if (entry.appliedRevision !== null) return "log_gap";
      hash = transition.resultConstantsHash;
    } catch { return "state_divergence"; }
  }
  if (revision !== head.revision || hash !== head.constantsHash || state.wireVersion !== head.version || canonicalJSONString(encodeFounderReplayState(state)) !== canonicalJSONString(head.state)) return "state_divergence";
  return "verified";
}

export function completedFounderAttendedMS(state: FounderReplayState): number {
  return safeInteger(state.ageMs, 0, MAX_EXACT_INTEGER);
}

export function effectiveFounderAttendedMS(completed: number, partial: number): number {
  const base = safeInteger(completed, 0, MAX_EXACT_INTEGER); const current = safeInteger(partial, 0, MAX_EXACT_INTEGER);
  if (current > MAX_EXACT_INTEGER - base) throw new RangeError("Founder attendance overflow");
  return base + current;
}

export function parseFounderAttendanceSample(value: unknown): FounderAttendanceSample {
  const raw = exactObject(value, ["company_stream_id", "run_seq", "company_revision", "company_constants_hash", "completed_attended_ms", "current_run_partial_attended_ms", "effective_founder_attended_ms"], "Founder attendance sample");
  const companyStreamId = string(raw.company_stream_id); const companyConstantsHash = string(raw.company_constants_hash);
  if (!uuid.test(companyStreamId) || !/^sha256:[0-9a-f]{64}$/.test(companyConstantsHash)) throw new SyntaxError("invalid Founder attendance identity");
  return { companyStreamId, runSeq: safeInteger(raw.run_seq, 1, MAX_EXACT_INTEGER), companyRevision: safeInteger(raw.company_revision, 1, MAX_EXACT_INTEGER), companyConstantsHash,
    completedAttendedMs: safeInteger(raw.completed_attended_ms, 0, MAX_EXACT_INTEGER), currentRunPartialAttendedMs: safeInteger(raw.current_run_partial_attended_ms, 0, MAX_EXACT_INTEGER),
    effectiveFounderAttendedMs: safeInteger(raw.effective_founder_attended_ms, 0, MAX_EXACT_INTEGER) };
}

export function validateFounderAttendanceSample(state: FounderReplayState, actualFounderRevision: number, expectedFounderRevision: number, sample: FounderAttendanceSample): number {
  const actual = safeInteger(actualFounderRevision, 1, MAX_EXACT_INTEGER); const expected = safeInteger(expectedFounderRevision, 1, MAX_EXACT_INTEGER);
  if (actual !== expected || sample.completedAttendedMs !== completedFounderAttendedMS(state)) throw new RangeError("stale Founder attendance sample");
  const effective = effectiveFounderAttendedMS(sample.completedAttendedMs, sample.currentRunPartialAttendedMs);
  if (effective !== sample.effectiveFounderAttendedMs) throw new RangeError("invalid Founder attendance total");
  return effective;
}

export async function applyLogged(state: ReplayState, canonicalPayload: string, catalogs: ReplayCatalogBundle, replayInputs: unknown): Promise<LoggedTransition> {
  const wire = parseReplayWire(replayInputs, state, catalogs); const request = parseIntent(canonicalPayload, wire.command.intent_id);
  if (request.expected_revision !== wire.command.revision || wire.command.run_seq !== state.runSeq) throw new RangeError("command/state mismatch");
  if (request.kind === "buy_route_hint") throw new RangeError("founder-scope intent is not replayable");
  deriveFactionStockResource(state, catalogs.factions);
  const stateBefore = cloneReplayState(state, catalogs);
  const revision = wire.command.revision; const before = { ...state.balances };
  if (request.invalid !== undefined) return rejected(state, request.intent_id, revision, "invalid", request.invalid);
  if (wire.evaluated_at_ms < state.evaluatedThroughMs) throw new ReplayClockViolation();
  const resolved = wire.resolved; const kind = string(resolved.kind); if (string(resolved.intent_kind) !== request.kind) throw new RangeError("resolved intent mismatch");
  let accrual: ReplayAccrual; let founderCarry: FounderCarry | null = null; let declined = 0;
  if (kind === "cross_gate" && request.kind === "cross_gate") { onlyKeys(resolved, ["kind", "intent_kind", "accrual", "declined_exit_offer_count", "founder_carry"], "cross gate inputs"); accrual = parseAccrual(resolved.accrual, catalogs); declined = safeInteger(resolved.declined_exit_offer_count ?? 0, 0, MAX_EXACT_INTEGER); founderCarry = resolved.founder_carry === undefined || resolved.founder_carry === null ? null : parseFounderCarry(resolved.founder_carry, catalogs, wire.v); }
  else { if (kind !== "accrual") throw new RangeError("resolved union mismatch"); const hasCarry = "founder_carry" in resolved; onlyKeys(resolved, hasCarry ? ["kind", "intent_kind", "accrual", "founder_carry"] : ["kind", "intent_kind", "accrual"], "accrual inputs"); accrual = parseAccrual(resolved.accrual, catalogs); founderCarry = !hasCarry || resolved.founder_carry === null ? null : parseFounderCarry(resolved.founder_carry, catalogs, wire.v); }
  if (foundationsActive(catalogs) && wire.v >= 4 && founderCarry === null) throw new RangeError("missing active Founder carry");
  if (state.compactMember !== (accrual.commons_weight_ppm !== null)) throw new RangeError("commons weight presence mismatch");
  const rejectState = (category: string, detail: string): LoggedTransition => { restoreReplaySnapshot(state, stateBefore); return rejected(state, request.intent_id, revision, category, detail); };
  try {
  applyGuildSettlements(state, accrual.guild_settlement_batch, catalogs.factions.stockCap);
  const contributions = assembleContributions(state, catalogs.economy, accrual.contributions);
  const effectiveAccrual = { ...accrual, contributions };
  const preflight = preflightRejection(state, catalogs, request, wire.evaluated_at_ms);
  if (preflight !== null) return rejectState(preflight[0], preflight[1]);
  const evaluation = evaluate(state, catalogs.economy, wire.evaluated_at_ms, wire.evaluation_mode, contributions);
  const events = runHooks(state, catalogs, wire.command, evaluation, effectiveAccrual);
  const postAccrualBalances = { ...state.balances };
  const invariants: ReplayInvariant[] = [];
  let appliedCount = 0;
  switch (request.kind) {
    case "buy_generator": { const result = buyGenerator(state, catalogs.economy, request, invariants); if (result.rejection) return rejectState(result.rejection[0], result.rejection[1]); appliedCount = result.count; events.push(event("generator_purchased", request.intent_id, { generator_id: request.generator_id, count: appliedCount, cost_resource_id: request.costResource, cost: request.cost })); break; }
    case "buy_upgrade": { const result = buyUpgrade(state, catalogs.economy, catalogs.routes, request); if (result.rejection) return rejectState(result.rejection[0], result.rejection[1]); events.push(event("upgrade_purchased", request.intent_id, { upgrade_id: request.upgrade_id, cost_resource_id: request.costResource, cost: request.cost })); appliedCount = 1; break; }
    case "perform_manual_batch": { const result = manualBatch(state, catalogs.economy, request, wire.evaluated_at_ms, contributions); if (result.rejection) return rejectState(result.rejection[0], result.rejection[1]); appliedCount = result.count; break; }
    case "cross_gate": { const result = crossGate(state, catalogs.routes, request, wire.command); if (result.rejection) return rejectState(result.rejection[0], result.rejection[1]); events.push(event("gate_crossed", request.intent_id, { founder_id: wire.command.founder_id, gate_id: request.gate_id, route_id: request.route_id, run_id: { company_stream_id: wire.command.company_stream_id, run_seq: state.runSeq } })); if (request.route_id !== null) events.push(event("route_executed", request.intent_id, { founder_id: wire.command.founder_id, gate_id: request.gate_id, route_id: request.route_id, run_id: { company_stream_id: wire.command.company_stream_id, run_seq: state.runSeq } })); appliedCount = 1; break; }
    case "sign_compact": state.compactMember = true; state.compactTithePpm = request.tithe_ppm; state.compactSolidarityPpm = 0; state.compactSamples = []; events.push(event("compact_signed", request.intent_id, compactMembershipPayload(wire.command, state.runSeq, request.tithe_ppm, false, true))); appliedCount = 1; break;
    case "leave_compact": { const prior = state.compactTithePpm; state.compactMember = false; state.compactTithePpm = 0; state.compactSolidarityPpm = 0; state.compactSamples = []; events.push(event("compact_left", request.intent_id, compactMembershipPayload(wire.command, state.runSeq, prior, true, false))); appliedCount = 1; break; }
    case "incorporate": { const member = catalogs.factions.byId.get(request.faction_id)!; state.factionId = member.id; state.factionStockResource = member.produces; state.incorporatedAtMs = state.evaluatedThroughMs; events.push(event("incorporated", request.intent_id, { compact_auto_signed: member.compact !== null, faction_id: member.id, founder_id: wire.command.founder_id, incorporated_at_ms: state.incorporatedAtMs, run_id: { company_stream_id: wire.command.company_stream_id, run_seq: state.runSeq }, stock_resource: member.produces })); if (member.compact) { if (state.compactMember) { const prior = state.compactTithePpm; state.compactTithePpm = Math.max(prior, member.compact.tithePpm); events.push(event("compact_tithe_raised", request.intent_id, { founder_id: wire.command.founder_id, new_tithe_ppm: state.compactTithePpm, prior_tithe_ppm: prior, run_id: { company_stream_id: wire.command.company_stream_id, run_seq: state.runSeq } })); } else { state.compactMember = true; state.compactTithePpm = member.compact.tithePpm; state.compactSolidarityPpm = 0; state.compactSamples = []; events.push(event("compact_signed", request.intent_id, compactMembershipPayload(wire.command, state.runSeq, member.compact.tithePpm, false, true))); } } appliedCount = 1; break; }
    case "decline_exit_offer": state.offerState = null; events.push(event("exit_offer_declined", request.intent_id, { offer_id: request.offer_id, run_seq: state.runSeq })); appliedCount = 1; break;
    default: return rejectState("invalid", request.kind);
  }
  for (const report of invariants) {
    if (report.kind === "residual_abort") continue;
    events.push(event("invariant_reported", request.intent_id, { detail: report.detail, invariant_kind: report.kind }));
  }
  const actionDebits = resourceDebits(postAccrualBalances, state.balances);
  await afterPrestigeTransition(state, catalogs.prestige, request, wire.command, wire.evaluated_at_ms, founderCarry, declined, events);
  if (foundationsActive(catalogs) && wire.v >= 4) applyFoundationTransition(catalogs, stateBefore, state, founderCarry!, wire.command, request, wire.evaluated_at_ms, contributions, actionDebits, false, events);
  return applied(state, catalogs.economy, request.intent_id, revision + 1, appliedCount, before, events, invariants);
  } catch (error) { restoreReplaySnapshot(state, stateBefore); throw error; }
}

export async function applyLoggedExit(company: ReplayState, canonicalPayload: string, catalogs: ReplayCatalogBundle, replayInputs: unknown): Promise<LoggedExitTransition> {
  const wire = parseReplayWire(replayInputs, company, catalogs);
  const request = parseIntent(canonicalPayload, wire.command.intent_id);
  if (request.expected_revision !== wire.command.revision || wire.command.run_seq !== company.runSeq) throw new RangeError("terminal command/state mismatch");
  const resolved = wire.resolved;
  onlyKeys(resolved, ["kind", "intent_kind", "accrual", "founder_carry", "executed_route_ids", "selected_exit_type", "selected_terms", "next_constants_hash"], "terminal resolved inputs");
  if (resolved.kind !== "exit" || resolved.intent_kind !== request.kind || typeof resolved.selected_exit_type !== "string" || !hashPattern.test(string(resolved.next_constants_hash))) throw new RangeError("terminal resolved union mismatch");
  const selectedTerms = exactObject(resolved.selected_terms, Object.keys(resolved.selected_terms as object), "selected terms");
  const nextHash = string(resolved.next_constants_hash);
  const next = nextHash === catalogs.constantsHash ? catalogs : catalogs.next;
  if (!next || next.constantsHash !== nextHash) throw new RangeError("next catalog bundle mismatch");
  if (foundationsActive(next) && wire.v < 3) throw new SyntaxError("foundation activation requires replay inputs v3+");
  const accrual = parseAccrual(resolved.accrual, catalogs);
  if (company.compactMember !== (accrual.commons_weight_ppm !== null)) throw new RangeError("commons weight presence mismatch");
  const founder = parseFounderCarry(resolved.founder_carry, catalogs, wire.v);
  const executedRoutes = sortedUniqueMechanical(array(resolved.executed_route_ids, "executed route ids"));
  deriveFactionStockResource(company, catalogs.factions);
  const companyBefore = cloneReplayState(company, catalogs);
  const revision = wire.command.revision;
  if (wire.evaluated_at_ms < company.evaluatedThroughMs) throw new ReplayClockViolation();
  let prefix: ReplayEvent[] = [];
  let actionDebits: Record<string, string> = {};
  let exitType: string;
  let terms: ExitTerms;
  const rejectState = (category: string, detail: string): LoggedExitTransition => { restoreReplaySnapshot(company, companyBefore); return rejectedExit(company, founder, request.intent_id, revision, category, detail); };
  try {
  applyGuildSettlements(company, accrual.guild_settlement_batch, catalogs.factions.stockCap);
  const contributions = assembleContributions(company, catalogs.economy, accrual.contributions);
  const effectiveAccrual = { ...accrual, contributions };

  if (request.kind === "cross_gate") {
    const preflight = preflightRejection(company, catalogs, request, wire.evaluated_at_ms);
    if (preflight !== null) return rejectState(preflight[0], preflight[1]);
    const evaluation = evaluate(company, catalogs.economy, wire.evaluated_at_ms, wire.evaluation_mode, contributions);
    prefix = runHooks(company, catalogs, wire.command, evaluation, effectiveAccrual);
    const postAccrualBalances = { ...company.balances };
    const result = crossGate(company, catalogs.routes, request, wire.command);
    if (result.rejection) return rejectState(result.rejection[0], result.rejection[1]);
    prefix.push(event("gate_crossed", request.intent_id, { founder_id: wire.command.founder_id, gate_id: request.gate_id, route_id: request.route_id, run_id: { company_stream_id: wire.command.company_stream_id, run_seq: company.runSeq } }));
    if (request.route_id !== null) prefix.push(event("route_executed", request.intent_id, { founder_id: wire.command.founder_id, gate_id: request.gate_id, route_id: request.route_id, run_id: { company_stream_id: wire.command.company_stream_id, run_seq: company.runSeq } }));
    actionDebits = resourceDebits(postAccrualBalances, company.balances);
    if (attendedMS(company, wire.evaluated_at_ms) < 900_000 || founder.exit_history_count !== 0) throw new RangeError("invalid scripted exit state");
    exitType = "scripted_first";
    terms = computeExitTerms(company, founder, catalogs.prestige, exitType);
  } else {
    if (request.kind === "file_ipo") return rejectState("not_eligible", "ipo_chain");
    if (request.kind === "wind_down" && company.tier < 1) return rejectState("not_eligible", "tier");
    exitType = request.kind === "wind_down" && founder.exit_history_count === 0 ? "scripted_first" : "collapse";
    let promised: { payout_preview: ExitTerms; market_modifier_ppm: number } | null = null;
    if (request.kind === "accept_exit_offer") {
      if (!company.offerState || company.offerState.offerId !== request.offer_id) return rejectState("not_eligible", "exit_offer");
      if (company.offerState.expiresAtMs <= wire.evaluated_at_ms) return rejectState("offer_expired", request.offer_id);
      if (canonicalJSONString(company.offerState.terms) !== canonicalJSONString(selectedTerms)) throw new RangeError("selected offer terms mismatch");
      promised = decodeStoredOfferTerms(selectedTerms);
      exitType = company.offerState.exitType;
    } else if (request.kind !== "wind_down") throw new RangeError("non-terminal intent at terminal boundary");
    const evaluation = evaluate(company, catalogs.economy, wire.evaluated_at_ms, wire.evaluation_mode, contributions);
    prefix = runHooks(company, catalogs, wire.command, evaluation, effectiveAccrual);
    terms = computeExitTerms(company, founder, catalogs.prestige, exitType);
    if (promised) terms = promiseTerms(promised.payout_preview, applyTermsModifier(terms, promised.market_modifier_ppm));
  }
  if (resolved.selected_exit_type !== exitType) throw new RangeError("selected exit type mismatch");
  if (foundationsActive(catalogs) && wire.v >= 4) applyFoundationTransition(catalogs, companyBefore, company, founder, wire.command, request, wire.evaluated_at_ms, contributions, actionDebits, true, prefix);
  return finishLoggedExit(company, founder, request.intent_id, wire.command, wire.evaluated_at_ms, exitType, terms, prefix, executedRoutes, next);
  } catch (error) { restoreReplaySnapshot(company, companyBefore); throw error; }
}

function cloneReplayState(state: ReplayState, catalogs: ReplayCatalogBundle): ReplayState {
  const cloned = restoreReplayState(encodeReplayState(state), state.wireVersion, catalogs.economy, foundationsActive(catalogs) ? catalogs : undefined);
  cloned.factionStockResource = state.factionStockResource;
  return cloned;
}

function restoreReplaySnapshot(target: ReplayState, snapshot: ReplayState): void {
  Object.assign(target, snapshot);
}

type Intent = Record<string, any> & { intent_id: string; kind: string; expected_revision: number; invalid?: string; cost?: string; costResource?: string };
interface Evaluation { changes: ReplayChange[]; elapsedMs: number; productionMs: number; bankedMs: number; progressDeltaPpm: number }

function preflightRejection(state: ReplayState, catalogs: ReplayCatalogBundle, request: Intent, nowMs: number): [string, string] | null {
  switch (request.kind) {
    case "buy_generator": {
      const generator = catalogs.economy.generatorClass(request.generator_id);
      return generator?.production ? null : ["unknown_id", request.generator_id];
    }
    case "buy_upgrade": {
      const upgrade = catalogs.economy.upgrades.find((value) => value.id === request.upgrade_id);
      if (!upgrade) return ["unknown_id", request.upgrade_id];
      return state.upgradesOwned.has(request.upgrade_id) ? ["not_eligible", "owned"] : null;
    }
    case "perform_manual_batch":
      return catalogs.economy.manualActions.some((value) => value.id === request.action_id) ? null : ["unknown_id", request.action_id];
    case "cross_gate":
      if (!catalogs.routes.gate(request.gate_id)) return ["unknown_id", request.gate_id];
      if (state.gatesCrossed[request.gate_id]) return ["gate_already_crossed", request.gate_id];
      if (request.route_id !== null && !catalogs.routes.route(request.route_id)) return ["unknown_id", request.route_id];
      return null;
    case "sign_compact":
      if (state.compactMember) return ["already_member", "compact"];
      return request.tithe_ppm < catalogs.commons.minimumTithePpm || request.tithe_ppm > catalogs.commons.maximumTithePpm ? ["invalid", "tithe_ppm"] : null;
    case "leave_compact": {
      const faction = state.factionId === "" ? undefined : catalogs.factions.byId.get(state.factionId);
      if (faction?.compact?.autoSign) return ["faction_bound", state.factionId];
      return state.compactMember ? null : ["not_member", "compact"];
    }
    case "incorporate":
      if (!catalogs.factions.byId.has(request.faction_id)) return ["unknown_id", request.faction_id];
      if (state.tier < 2) return ["not_eligible", "tier"];
      return state.factionId === "" ? null : ["already_incorporated", state.factionId];
    case "decline_exit_offer":
      if (!state.offerState || state.offerState.offerId !== request.offer_id) return ["not_eligible", "exit_offer"];
      return state.offerState.expiresAtMs <= nowMs ? ["offer_expired", request.offer_id] : null;
    default:
      return ["invalid", request.kind];
  }
}

function assembleContributions(state: ReplayState, catalog: EconomyCatalog, external: readonly ReplayContribution[]): ReplayContribution[] {
  const combined = [...external, ...contentContributions(state, catalog)].sort((a, b) => byteCompare(contributionKey(a), contributionKey(b)));
  validateContributionSet(catalog, combined);
  return combined;
}

function contentContributions(state: ReplayState, catalog: EconomyCatalog): ReplayContribution[] {
  const result: ReplayContribution[] = [];
  for (const upgrade of catalog.upgrades) if (state.upgradesOwned.has(upgrade.id)) for (const effect of upgrade.effects) result.push({ slot: effect.slot, source_id: effect.sourceId, target: effect.target, factor: effect.factor });
  for (const generator of catalog.generatorClasses) {
    const purchased = state.generators[generator.id]; if (!Number.isSafeInteger(purchased) || purchased! < 0) throw new RangeError("invalid purchased generator count");
    for (const rung of generator.ladder) { if (purchased! < rung.purchasedAt) break; result.push({ slot: "milestones", source_id: ladderSourceId(generator.id, rung.purchasedAt), target: generator.id, factor: canonicalString(quantize(new Decimal(rung.multiplierPpm).div(1_000_000))) }); }
    for (const role of generator.roles) if (role.kind === "manual_output" && purchased! > 0) result.push({ slot: "upgrades", source_id: manualRoleSourceId(generator.id, role.actionId), target: role.actionId, factor: countPpmFactor(purchased!, role.perPurchasedPpm) });
  }
  for (const pool of catalog.synergyPools) {
    let total = 0n;
    for (const source of pool.sources) { const count = source.kind === "generator" ? state.generators[source.id] : state.upgradesOwned.has(source.id) ? 1 : 0; if (!Number.isSafeInteger(count) || count! < 0) throw new RangeError("invalid synergy source count"); total += BigInt(count!) * BigInt(source.perCountPpm); }
    if (total === 0n) continue;
    const declaration = catalog.multiplierSources.find((value) => value.id === pool.id); if (!declaration) throw new RangeError("missing synergy declaration");
    result.push({ slot: pool.slot, source_id: pool.id, target: declaration.target, factor: synergyFactor(pool.curve, total) });
  }
  return result.sort((a, b) => byteCompare(contributionKey(a), contributionKey(b)));
}

function countPpmFactor(count: number, perCountPpm: number): string { return canonicalString(quantize(new Decimal((BigInt(count) * BigInt(perCountPpm)).toString()).div(1_000_000).add(1))); }
function synergyFactor(curve: "linear" | "log", totalPpm: bigint): string { const base = new Decimal(totalPpm.toString()).div(1_000_000).add(1); return canonicalString(quantize(curve === "linear" ? base : new Decimal(base.log10()).add(1))); }
function contributionKey(value: ReplayContribution): string { return `${value.slot}\0${value.source_id}\0${value.target}`; }
function validateContributionSet(catalog: EconomyCatalog, values: readonly ReplayContribution[]): void { const declared = new Map(catalog.multiplierSources.map((value) => [value.id, value])); const seen = new Set<string>(); for (const value of values) { const source = declared.get(value.source_id); const factor = parseCanonical(value.factor); if (!source || source.slot !== value.slot || source.target !== value.target || seen.has(value.source_id) || !isStateValue(factor) || !factor.gt(0)) throw new RangeError("invalid multiplier contribution"); seen.add(value.source_id); } }
function contributionFactorForTarget(catalog: EconomyCatalog, target: string, values: readonly ReplayContribution[]): Decimal { validateContributionSet(catalog, values); let factor = new Decimal(1); for (const slot of MULTIPLIER_SLOT_ORDER) for (const value of values.filter((item) => item.slot === slot && item.target === target).sort((a, b) => byteCompare(a.source_id, b.source_id))) factor = factor.mul(parseCanonical(value.factor)); const result = quantize(factor); if (!isStateValue(result) || !result.gt(0)) throw new RangeError("invalid target contribution factor"); return result; }

function evaluate(state: ReplayState, catalog: EconomyCatalog, nowMs: number, mode: "online" | "offline", contributions: readonly ReplayContribution[]): Evaluation {
  if (nowMs < state.evaluatedThroughMs) throw new ReplayClockViolation(); if (nowMs === state.evaluatedThroughMs) return { changes: [], elapsedMs: 0, productionMs: 0, bankedMs: 0, progressDeltaPpm: 0 };
  const elapsedMs = nowMs - state.evaluatedThroughMs; let productionMs = elapsedMs; let efficiency = new Decimal(1); let bankedMs = 0; const beforeProgress = progressPpm(catalog, state);
  if (mode === "offline") { const policy = catalog.offlinePolicy!; efficiency = parseCanonical(policy.efficiency); productionMs = Math.min(productionMs, policy.accrualCapMs); bankedMs = Number((BigInt(elapsedMs - productionMs) * BigInt(policy.bankRatioNumerator)) / BigInt(policy.bankRatioDenominator)); bankedMs = Math.min(bankedMs, policy.bankCapMs - state.computeCreditMs); }
  const accrued = accrueContent(state, catalog, productionMs, efficiency, contributions);
  const changes = applyLedger(state, catalog, accrued.entries, true); state.computeCreditMs += bankedMs; state.generatorsProvisioned = accrued.provisioned; state.provisionRemaindersPpm = accrued.remainders; state.evaluatedThroughMs = nowMs; return { changes, elapsedMs, productionMs, bankedMs, progressDeltaPpm: Math.max(0, progressPpm(catalog, state) - beforeProgress) };
}

function accrueContent(state: ReplayState, catalog: EconomyCatalog, productionMs: number, efficiency: Decimal, contributions: readonly ReplayContribution[]): { entries: { resource: string; delta: Decimal }[]; provisioned: Record<string, number>; remainders: Record<string, number> } {
  const provisioned = { ...state.generatorsProvisioned }; const remainders = { ...state.provisionRemaindersPpm }; const deltas = new Map<string, Decimal[]>();
  const accrueSegment = (segmentMs: number): void => {
    if (segmentMs <= 0) return;
    const rates = productionRates(catalog, state.generators, provisioned, contributions);
    for (const [resource, values] of rates) { const delta = quantize(sumDeterministic(values).mul(segmentMs).div(1000).mul(efficiency)); if (delta.eq(0)) continue; const prior = deltas.get(resource) ?? []; prior.push(delta); deltas.set(resource, prior); }
  };
  if (catalog.provisionTickMs === 0 || productionMs === 0) accrueSegment(productionMs);
  else {
    if (state.runStartedAtMs <= 0 || state.evaluatedThroughMs < state.runStartedAtMs || Math.floor(productionMs / catalog.provisionTickMs) > Math.floor(catalog.offlinePolicy!.accrualCapMs / catalog.provisionTickMs) + 1) throw new RangeError("invalid provision bucket horizon");
    let cursorMs = state.evaluatedThroughMs; const endMs = cursorMs + productionMs; let nextBoundary = state.runStartedAtMs + (Math.floor((cursorMs - state.runStartedAtMs) / catalog.provisionTickMs) + 1) * catalog.provisionTickMs;
    if (!Number.isSafeInteger(endMs) || !Number.isSafeInteger(nextBoundary)) throw new RangeError("provision bucket exceeds exact domain");
    while (cursorMs < endMs) { const segmentEnd = Math.min(endMs, nextBoundary); accrueSegment(segmentEnd - cursorMs); cursorMs = segmentEnd; if (cursorMs === nextBoundary) { materializeProvisionBoundary(catalog, state.generators, provisioned, remainders); nextBoundary += catalog.provisionTickMs; } }
  }
  const entries = [...deltas.entries()].sort(([a], [b]) => byteCompare(a, b)).map(([resource, values]) => ({ resource, delta: quantize(sumDeterministic(values)) })).filter((entry) => !entry.delta.eq(0));
  return { entries, provisioned, remainders };
}

function materializeProvisionBoundary(catalog: EconomyCatalog, purchased: Record<string, number>, provisioned: Record<string, number>, remainders: Record<string, number>): void {
  const staged: Record<string, number> = {}; const stagedRemainders: Record<string, number> = {};
  for (const source of catalog.generatorClasses) {
    if (!source.provision) continue;
    const target = catalog.generatorClass(source.provision.generatorId); const sourcePurchased = purchased[source.id]; const sourceProvisioned = provisioned[source.id]; const priorRemainder = remainders[source.provision.generatorId]; const currentTarget = provisioned[source.provision.generatorId];
    if (!target?.provisionedHardcap || !Number.isSafeInteger(sourcePurchased) || sourcePurchased! < 0 || !Number.isSafeInteger(sourceProvisioned) || sourceProvisioned! < 0 || !Number.isSafeInteger(priorRemainder) || priorRemainder! < 0 || priorRemainder! >= 1_000_000 || !Number.isSafeInteger(currentTarget) || currentTarget! < 0) throw new RangeError("invalid provision state");
    const numerator = (BigInt(sourcePurchased!) + BigInt(sourceProvisioned!)) * BigInt(source.provision.ratePpm) + BigInt(priorRemainder!); const quotient = numerator / 1_000_000n; const remainder = numerator % 1_000_000n; const headroom = BigInt(target.provisionedHardcap.count - currentTarget!);
    staged[source.provision.generatorId] = Number(quotient > headroom ? headroom : quotient); stagedRemainders[source.provision.generatorId] = Number(remainder);
  }
  for (const target of Object.keys(stagedRemainders)) { remainders[target] = stagedRemainders[target]!; provisioned[target] = provisioned[target]! + staged[target]!; }
}

function productionRates(catalog: EconomyCatalog, counts: Record<string, number>, provisioned: Record<string, number>, contributions: readonly ReplayContribution[]): Map<string, Decimal[]> {
  const declared = new Map(catalog.multiplierSources.map((value) => [value.id, value])); const bySource = new Map<string, ReplayContribution>();
  for (const contribution of contributions) { const source = declared.get(contribution.source_id); if (!source || source.slot !== contribution.slot || source.target !== contribution.target || bySource.has(contribution.source_id)) throw new RangeError("invalid multiplier contribution"); bySource.set(contribution.source_id, contribution); }
  const result = new Map<string, Decimal[]>();
  for (const generator of catalog.generatorClasses.filter((value) => value.production !== null)) { const count = counts[generator.id]; const generated = provisioned[generator.id]; if (!Number.isSafeInteger(count) || count! < 0 || !Number.isSafeInteger(generated) || generated! < 0) throw new RangeError("invalid generator count"); if (count === 0 && generated === 0) continue; let rate = parseCanonical(generator.production!.baseRate).mul(new Decimal((BigInt(count!) + BigInt(generated!)).toString())); for (const slot of MULTIPLIER_SLOT_ORDER) { const sources = [...bySource.values()].filter((value) => value.slot === slot && (value.target === "all" || value.target === generator.id)).sort((a, b) => byteCompare(a.source_id, b.source_id)); for (const source of sources) rate = rate.mul(parseCanonical(source.factor)); } const values = result.get(generator.production!.resourceId) ?? []; values.push(rate); result.set(generator.production!.resourceId, values); }
  return result;
}

function applyLedger(state: ReplayState, catalog: EconomyCatalog, entries: readonly { resource: string; delta: Decimal }[], saturating: boolean): ReplayChange[] {
  const grouped = new Map<string, Decimal[]>(); for (const entry of entries) { const values = grouped.get(entry.resource) ?? []; values.push(entry.delta); grouped.set(entry.resource, values); }
  const changes: ReplayChange[] = []; for (const resource of [...grouped.keys()].sort(byteCompare)) { const definition = catalog.resource(resource); if (!definition || definition.scope !== "company") throw new LedgerError("unknown_resource"); const before = parseCanonical(state.balances[resource]!); const net = sumDeterministic(grouped.get(resource)!); let after = quantize(before.add(net)); if (after.lt(parseCanonical(definition.minimum))) throw new LedgerError("below_minimum"); let preferred = net; if (definition.hardcap && after.gt(parseCanonical(definition.hardcap.amount))) { if (!saturating) throw new LedgerError("above_hardcap"); after = parseCanonical(definition.hardcap.amount); preferred = quantize(after.sub(before)); } if (after.eq(before)) continue; const delta = saturating ? reproducibleDelta(before, after, preferred) : quantize(after.sub(before)); state.balances[resource] = canonicalString(after); changes.push({ resource_id: resource, before: canonicalString(before), delta: canonicalString(delta), after: canonicalString(after) }); }
  return changes;
}

class LedgerError extends RangeError { constructor(readonly code: "unknown_resource" | "below_minimum" | "above_hardcap") { super(code); } }

function reproducibleDelta(before: Decimal, after: Decimal, preferred: Decimal): Decimal {
  const nominal = quantize(preferred);
  if (!isStateValue(nominal)) throw new LedgerError("above_hardcap");
  if (nominal.eq(0)) {
    if (quantize(before.add(nominal)).eq(after)) return nominal;
    throw new LedgerError("above_hardcap");
  }
  for (const offset of [0, -1, 1, -2, 2]) {
    const candidate = quantize(Decimal.fromMantissaExponent(nominal.mantissa + offset * 1e-11, nominal.exponent));
    if (!isStateValue(candidate) || candidate.lt(0)) continue;
    if (quantize(before.add(candidate)).eq(after)) return candidate;
  }
  throw new LedgerError("above_hardcap");
}

function runHooks(state: ReplayState, catalogs: ReplayCatalogBundle, command: ReplayCommand, result: Evaluation, accrual: ReplayAccrual): ReplayEvent[] {
  if (result.elapsedMs === 0) return [];
  for (const change of result.changes) if (change.resource_id === catalogs.prestige.valueResourceId) state.lifetimeValue = canonicalString(quantize(parseCanonical(state.lifetimeValue).add(parseCanonical(change.delta))));
  recordOfflineSpan(state, state.evaluatedThroughMs - result.elapsedMs, state.evaluatedThroughMs, catalogs.prestige.catchupCeilingMs);
  const events: ReplayEvent[] = [];
  if (state.factionId !== "" && result.elapsedMs <= catalogs.prestige.catchupCeilingMs) { let factorPpm = 1_000_000n; for (const generator of catalogs.economy.generatorClasses) { const purchased = state.generators[generator.id]; if (!Number.isSafeInteger(purchased) || purchased! < 0) throw new RangeError("invalid stock-rate generator count"); for (const role of generator.roles) if (role.kind === "stock_rate") factorPpm += BigInt(purchased!) * BigInt(role.perPurchasedPpm); } const numerator = BigInt(result.elapsedMs) * factorPpm + BigInt(state.stockRateRemainderPpm); const scaledMs = numerator / 1_000_000n; state.stockRateRemainderPpm = Number(numerator % 1_000_000n); if (scaledMs > BigInt(MAX_EXACT_INTEGER - state.stockProgressMs)) throw new RangeError("stock-rate accumulator overflow"); const total = state.stockProgressMs + Number(scaledMs); let earned = Math.floor(total / catalogs.factions.stockIntervalMs); state.stockProgressMs = total % catalogs.factions.stockIntervalMs; const before = state.stockUnits; earned = Math.min(earned, catalogs.factions.stockCap - state.stockUnits); state.stockUnits += earned; if (before !== catalogs.factions.stockCap && state.stockUnits === catalogs.factions.stockCap) events.push(event("faction_stock_saturated", "", { faction_id: state.factionId, stock_cap: catalogs.factions.stockCap, stock_resource: state.factionStockResource })); }
  if (accrual.contributions.some((value) => value.source_id === "guild.stock_consumption")) { const numerator = BigInt(result.progressDeltaPpm) * BigInt(catalogs.guilds.guildTithePpm) + BigInt(state.guildTitheCarryPpm); const xp = Number(numerator / 1_000_000n); state.guildTitheCarryPpm = Number(numerator % 1_000_000n); events.push(event(xp === 0 ? "guild_activity_evaluated" : "guild_tithe_accrued", "", { founder_id: command.founder_id, progress_delta_ppm: result.progressDeltaPpm, run_id: { company_stream_id: command.company_stream_id, run_seq: state.runSeq }, xp_delta: xp })); }
  if (state.compactMember && result.productionMs > 0) { if (accrual.commons_weight_ppm === null) throw new RangeError("missing commons weight"); const enclosure = enclosureIndex(catalogs.commons, accrual.contributions.map((value) => ({ sourceId: value.source_id, slot: value.slot, factor: value.factor }))); const compliance = compliancePpm(state.compactTithePpm, catalogs.commons.defaultTithePpm, enclosure); appendCompactSamples(state, state.evaluatedThroughMs - result.productionMs, state.evaluatedThroughMs, compliance, catalogs.commons.solidarityWindowMs); let capacity = new Decimal(0); for (const change of result.changes) capacity = capacity.add(parseCanonical(change.delta).mul(state.compactTithePpm).div(1_000_000)); events.push(event("compact_sampled", "", { capacity: canonicalString(quantize(capacity)), compliance_ppm: compliance, enclosure, founder_id: command.founder_id, run_id: { company_stream_id: command.company_stream_id, run_seq: state.runSeq }, sampled_ms: result.productionMs, solidarity_ppm: state.compactSolidarityPpm, weight_ppm: accrual.commons_weight_ppm })); }
  return events;
}

function applyFoundationTransition(catalogs: ReplayCatalogBundle, before: ReplayState, state: ReplayState, founder: FounderCarry, command: ReplayCommand, request: Intent, nowMs: number, contributions: readonly ReplayContribution[], actionDebits: Readonly<Record<string, string>>, terminal: boolean, events: ReplayEvent[]): void {
  if (!foundationsActive(catalogs) || state.wireVersion !== 16) throw new RangeError("invalid foundation hook inputs");
  const attendedMs = attendedMS(state, state.evaluatedThroughMs) - attendedMS(before, before.evaluatedThroughMs);
  if (!Number.isSafeInteger(attendedMs) || attendedMs < 0 || attendedMs > 86_400_000) throw new RangeError("invalid foundation attended time");
  const newFacts = new Set([...state.ledgerFactKinds].filter((fact) => !before.ledgerFactKinds.has(fact)));
  const activeContributions = new Set(contributions.filter((value) => !parseCanonical(value.factor).eq(1)).map((value) => meterContributionKey(value.slot, value.source_id)));
  const changes = advanceMeters(catalogs.meters, { values: state.meterValues, decayRemainders: state.meterDecayRemainders, inputRemainders: state.meterInputRemainders }, { attendedMs, newFactKinds: newFacts, activeContributions });
  const runId = { company_stream_id: command.company_stream_id, run_seq: state.runSeq };
  for (const change of changes) events.push(event("meter_band_changed.v1", request.intent_id, { run_id: runId, meter_id: change.meterId, from_band: change.fromBand, to_band: change.toBand, direction: change.direction, value_before: change.valueBefore, value_after: change.valueAfter }));

  const run: AchievementObservation = { facts: state.ledgerFactKinds, counters: { generators_purchased_total: state.generatorPurchasedTotal, tier: state.tier }, exitCount: 0, generators: state.generators };
  const careerFacts = new Set(founder.ledger_fact_kinds);
  let ageMs = founder.age_ms; let exitCount = founder.exit_history_count;
  if (terminal) {
    const attended = attendedMS(state, nowMs);
    if (ageMs > MAX_EXACT_INTEGER - attended) throw new RangeError("Founder age overflow");
    ageMs += attended; exitCount++;
    for (const fact of state.ledgerFactKinds) careerFacts.add(fact);
  }
  const career: AchievementObservation = { facts: careerFacts, counters: { age_ms: ageMs, notoriety: founder.notoriety }, exitCount, generators: {} };
  const earned = newlyEarned(catalogs.achievements, state.achievementsEarnedRun, new Set(founder.achievements_earned_lifetime), run, career)
    .filter((definition) => achievementProofSatisfied(definition, actionDebits, events));
  for (const definition of earned) state.achievementsEarnedRun.add(definition.id);
  state.achievementScoreRun = achievementScore(catalogs.achievements, state.achievementsEarnedRun);
  for (const definition of earned) events.push(event("achievement_earned.v1", request.intent_id, { run_id: runId, achievement_id: definition.id, condition_scope: definition.conditionScope, score_grant: definition.scoreGrant }));
}

function achievementProofSatisfied(definition: AchievementCatalog["definitions"][number], actionDebits: Readonly<Record<string, string>>, events: readonly ReplayEvent[]): boolean {
  const proof = definition.proof;
  if (proof.kind !== "burn") return true;
  if (!events.some((value) => value.kind === proof.eventKind)) return false;
  const debit = actionDebits[proof.resourceId];
  return debit !== undefined && parseCanonical(debit).gte(parseCanonical(proof.minimum));
}

function resourceDebits(postAccrual: Readonly<Record<string, string>>, after: Readonly<Record<string, string>>): Record<string, string> {
  if (Object.keys(postAccrual).length !== Object.keys(after).length) throw new RangeError("action debit key mismatch");
  const debits: Record<string, string> = {};
  for (const [resourceId, beforeRaw] of Object.entries(postAccrual)) {
    const afterRaw = after[resourceId];
    if (afterRaw === undefined) throw new RangeError("action debit key mismatch");
    const beforeValue = parseCanonical(beforeRaw); const afterValue = parseCanonical(afterRaw);
    if (beforeValue.gt(afterValue)) debits[resourceId] = canonicalString(quantize(beforeValue.sub(afterValue)));
  }
  return debits;
}

function buyGenerator(state: ReplayState, catalog: EconomyCatalog, request: Intent, invariants: ReplayInvariant[]): { count: number; rejection?: [string, string] } {
  const generator = catalog.generatorClass(request.generator_id)!; const owned = state.generators[generator.id]; if (owned === undefined) throw new RangeError("missing generator count"); const balance = parseCanonical(state.balances[generator.price.resourceId]!); let count = request.count.value;
  if (request.count.mode === "max") { const affordability = catalog.maxAffordableDetailed(generator.id, balance, owned); count = affordability.count; if (affordability.usedFallback) invariants.push({ kind: "afford_fallback", intent_id: request.intent_id, detail: request.generator_id }); }
  if (count <= 0) return { count: 0, rejection: ["unaffordable", request.generator_id] };
  if (count > MAX_EXACT_INTEGER - owned) return { count: 0, rejection: ["cap_exceeded", request.generator_id] };
  if (count > MAX_EXACT_INTEGER - state.generatorPurchasedTotal) return { count: 0, rejection: ["cap_exceeded", "generators_purchased_total"] };
  let cost: Decimal;
  try { cost = catalog.bulkCost(generator.id, owned, count); } catch { return { count: 0, rejection: ["invalid", request.generator_id] }; }
  if (cost.gt(balance)) return { count: 0, rejection: ["unaffordable", request.generator_id] };
  const residual = quantize(balance.sub(cost));
  if (residual.lt(0)) {
    const unit = Decimal.fromMantissaExponent(1, balance.exponent - 11);
    if (isStateValue(unit) && residual.abs().lte(unit)) { cost = balance; invariants.push({ kind: "residual_clamp", intent_id: request.intent_id, detail: request.generator_id }); }
    else { invariants.push({ kind: "residual_abort", intent_id: request.intent_id, detail: request.generator_id }); throw new RangeError("generator residual cannot be reconciled"); }
  }
  applyLedger(state, catalog, [{ resource: generator.price.resourceId, delta: cost.neg() }], false); state.generators[generator.id] = owned + count; state.generatorPurchasedTotal += count; request.cost = canonicalString(cost); request.costResource = generator.price.resourceId; return { count };
}
function buyUpgrade(state: ReplayState, catalog: EconomyCatalog, routes: RoutesCatalog, request: Intent): { rejection?: [string, string] } {
  const upgrade = catalog.upgrades.find((value) => value.id === request.upgrade_id)!;
  if (upgrade.window.fromGate !== null && !state.gatesCrossed[upgrade.window.fromGate] || upgrade.window.toGate !== null && state.gatesCrossed[upgrade.window.toGate]) return { rejection: ["not_eligible", "window"] };
  const context: RouteContext = { contextVersion: routes.contextVersion, resources: state.balances, doctrinesByTransition: state.doctrinesByTransition, structureId: state.structureId, ledgerFactKinds: state.ledgerFactKinds, meterBands: state.wireVersion === 16 ? state.meterValues : state.meterBands, regionTraits: state.regionTraits };
  if (!evaluatePredicate(upgrade.requires, context)) return { rejection: ["not_eligible", "requires"] };
  const balance = parseCanonical(state.balances[upgrade.cost.resourceId]!); const cost = parseCanonical(upgrade.cost.amount);
  if (balance.lt(cost)) return { rejection: ["unaffordable", request.upgrade_id] };
  applyLedger(state, catalog, [{ resource: upgrade.cost.resourceId, delta: cost.neg() }], false); state.upgradesOwned.add(upgrade.id); request.cost = upgrade.cost.amount; request.costResource = upgrade.cost.resourceId; return {};
}
function manualBatch(state: ReplayState, catalog: EconomyCatalog, request: Intent, nowMs: number, contributions: readonly ReplayContribution[]): { count: number; rejection?: [string, string] } { const action = catalog.manualActions.find((value) => value.id === request.action_id)!; const policy = catalog.manualPolicy!; if (nowMs > state.manualTokenRefilledAtMs) { const elapsed = nowMs - state.manualTokenRefilledAtMs; state.manualTokenMilli = Math.min(policy.bucketCapMilli, state.manualTokenMilli + elapsed * policy.refillMilliPerMs); state.manualTokenRefilledAtMs = nowMs; } const applied = Math.min(request.count, Math.floor(state.manualTokenMilli / 1000)); state.manualTokenMilli -= applied * 1000; if (applied > 0) { try { const factor = contributionFactorForTarget(catalog, request.action_id, contributions); applyLedger(state, catalog, [{ resource: action.output.resourceId, delta: parseCanonical(action.output.amountPerAction).mul(applied).mul(factor) }], false); } catch (error) { if (error instanceof LedgerError && error.code === "above_hardcap") return { count: 0, rejection: ["cap_exceeded", request.action_id] }; throw error; } } return { count: applied }; }
function crossGate(state: ReplayState, routes: RoutesCatalog, request: Intent, command: ReplayCommand): { rejection?: [string, string] } { const gate = routes.gate(request.gate_id); if (!gate) return { rejection: ["unknown_id", request.gate_id] }; if (state.gatesCrossed[request.gate_id]) return { rejection: ["gate_already_crossed", request.gate_id] }; let requirements = gate.requirement; if (request.route_id !== null) { const route = routes.route(request.route_id); if (!route) return { rejection: ["unknown_id", request.route_id] }; const context: RouteContext = { contextVersion: routes.contextVersion, resources: state.balances, doctrinesByTransition: state.doctrinesByTransition, structureId: state.structureId, ledgerFactKinds: state.ledgerFactKinds, meterBands: state.wireVersion === 16 ? state.meterValues : state.meterBands, regionTraits: state.regionTraits }; if (!route.active || route.requiresContextVersion > context.contextVersion || !evaluatePredicate(route.predicate, context)) return { rejection: ["route_predicate_unmet", request.route_id] }; requirements = discountedRequirements(gate, route); }
  for (const requirement of requirements) if (parseCanonical(state.balances[requirement.resourceId]!).lt(parseCanonical(requirement.amount))) return { rejection: ["requirement_not_met", requirement.resourceId] }; for (const requirement of requirements) state.balances[requirement.resourceId] = canonicalString(quantize(parseCanonical(state.balances[requirement.resourceId]!).sub(parseCanonical(requirement.amount)))); state.gatesCrossed[request.gate_id] = true; const match = /^gate\.t([0-9]+)_to_t([0-9]+)$/.exec(request.gate_id); const from = Number(match?.[1]); const to = Number(match?.[2]); if (!match || !Number.isSafeInteger(from) || !Number.isSafeInteger(to) || to !== from + 1 || to > 9) throw new RangeError("gate tier mismatch"); state.tier = Math.max(state.tier, to); void command; return {}; }

async function afterPrestigeTransition(state: ReplayState, policy: PrestigePolicy, request: Intent, command: ReplayCommand, nowMs: number, carry: FounderCarry | null, declined: number, events: ReplayEvent[]): Promise<void> {
  if (state.offerState && state.offerState.expiresAtMs <= nowMs) { events.push(event("exit_offer_expired", request.intent_id, { offer_id: state.offerState.offerId })); state.offerState = null; }
  if (request.kind !== "cross_gate" || state.offerState) return;
  if (carry === null) throw new RangeError("missing founder carry");
  if (safeInteger(carry.exit_history_count, 0, MAX_EXACT_INTEGER) === 0) return;
  if (state.tier < 0 || state.tier >= policy.spawnGatePpm.length) throw new RangeError("tier outside offer policy");
  const seed = (await founderSeed(command.founder_id, state.runSeq)) ^ (BigInt(state.tier) << 32n) ^ BigInt(declined);
  const draws = new SplitMix64(seed); const spawn = draws.ppm(); const exitType = (draws.next() & 1n) === 0n ? "acquihire" : "acquisition"; const driftUp = (draws.next() & 1n) === 0n;
  if (policy.spawnGatePpm[state.tier]! <= spawn) return;
  const reputationLevelValue = safeInteger(carry.reputation_level, 0, MAX_EXACT_INTEGER);
  const baseDelta = reputationDeltaExact(state.lifetimeValue, policy.threshold, reputationLevelValue, policy.exitModifiersPpm[exitType]!);
  const factor = marketModifierPpm(declined, policy.declineDriftPpm, driftUp);
  const terms = { reputation_delta: ppmFloor(baseDelta, factor), network_slot_unlocks: [], route_knowledge: 0, clout_reach_note: "clout.reach.preserved" };
  const termsJSON = { market_modifier_ppm: factor, payout_preview: terms };
  const offerId = offerUUID(seed, nowMs); const expiresAtMs = nowMs + policy.offerDurationMs;
  state.offerState = { offerId, exitType, terms: termsJSON, spawnedAtMs: nowMs, expiresAtMs };
  events.push(event("exit_offer_spawned", request.intent_id, { exit_type: exitType, expires_at_ms: expiresAtMs, offer_id: offerId, payout_preview: terms }));
}

function reputationDeltaExact(lifetimeValue: string, threshold: string, current: number, modifier: number): number {
  const ratio = parseCanonical(lifetimeValue).div(parseCanonical(threshold)); if (ratio.lt(1)) return 0; let low = 1; let high = MAX_EXACT_INTEGER;
  while (low < high) { const middle = low + Math.floor((high - low + 1) / 2); const candidate = new Decimal(middle); const cube = candidate.mul(candidate).mul(candidate); if (isStateValue(cube) && cube.lte(ratio)) low = middle; else high = middle - 1; }
  if (low <= current) return 0; return Number((BigInt(low - current) * BigInt(modifier)) / 1_000_000n);
}
function marketModifierPpm(declined: number, step: number, up: boolean): number { if (declined <= 0 || step <= 0) return 1_000_000; const delta = Math.min(declined, 10) * step; return up ? 1_000_000 + delta : delta < 1_000_000 ? 1_000_000 - delta : 0; }
function ppmFloor(value: number, factor: number): number { const result = (BigInt(value) * BigInt(factor)) / 1_000_000n; return result > BigInt(MAX_EXACT_INTEGER) ? MAX_EXACT_INTEGER : Number(result); }
async function founderSeed(founderId: string, runSeq: number): Promise<bigint> { const digest = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(founderId)); return new DataView(digest).getBigUint64(0) ^ BigInt(runSeq); }
class SplitMix64 { #state: bigint; constructor(seed: bigint) { this.#state = seed & ((1n << 64n) - 1n); } next(): bigint { const mask = (1n << 64n) - 1n; this.#state = (this.#state + 0x9e3779b97f4a7c15n) & mask; let value = this.#state; value = ((value ^ (value >> 30n)) * 0xbf58476d1ce4e5b9n) & mask; value = ((value ^ (value >> 27n)) * 0x94d049bb133111ebn) & mask; return (value ^ (value >> 31n)) & mask; } ppm(): number { const bound = 1_000_000n; const threshold = (1n << 64n) % bound; for (;;) { const draw = this.next(); if (draw >= threshold) return Number(draw % bound); } } }
function offerUUID(seed: bigint, nowMs: number): string { const random = new SplitMix64(seed); const bytes = new Uint8Array(16); let timestamp = BigInt(nowMs) & ((1n << 48n) - 1n); for (let index = 5; index >= 0; index--) { bytes[index] = Number(timestamp & 255n); timestamp >>= 8n; } let first = random.next(); for (let index = 13; index >= 6; index--) { bytes[index] = Number(first & 255n); first >>= 8n; } bytes[14] = Number((random.next() >> 56n) & 255n); bytes[15] = Number((random.next() >> 48n) & 255n); bytes[6] = (bytes[6]! & 0x0f) | 0x70; bytes[8] = (bytes[8]! & 0x3f) | 0x80; const hex = [...bytes].map((value) => value.toString(16).padStart(2, "0")).join(""); return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`; }

function attendedMS(state: ReplayState, endedAtMs: number): number {
  if (state.runStartedAtMs === 0 || endedAtMs < state.runStartedAtMs) throw new RangeError("invalid attended-time state");
  const rta = endedAtMs - state.runStartedAtMs;
  let offline = Math.min(state.collapsedOfflineMs, rta);
  for (const span of state.offlineSpans) {
    const from = Math.max(span.fromMs, state.runStartedAtMs);
    const to = Math.min(span.toMs, endedAtMs);
    if (to <= from) continue;
    const duration = to - from;
    if (duration > rta - offline) { offline = rta; break; }
    offline += duration;
  }
  return rta - offline;
}

function computeExitTerms(company: ReplayState, founder: FounderCarry, policy: PrestigePolicy, exitType: string): ExitTerms {
  const modifier = policy.exitModifiersPpm[exitType];
  if (modifier === undefined) throw new RangeError("unknown exit type");
  const reputation = Math.min(reputationDeltaExact(company.lifetimeValue, policy.threshold, founder.reputation_level, modifier), MAX_EXACT_INTEGER - founder.reputation_level);
  const knowledge = exitType === "collapse" || exitType === "scripted_first" ? policy.collapseRouteKnowledge : 0;
  return { reputation_delta: reputation, network_slot_unlocks: [], route_knowledge: knowledge, clout_reach_note: "clout.reach.preserved" };
}

function decodeStoredOfferTerms(source: Record<string, unknown>): { payout_preview: ExitTerms; market_modifier_ppm: number } {
  exactKeys(source, ["payout_preview", "market_modifier_ppm"], "stored offer terms");
  const preview = exactObject(source.payout_preview, ["reputation_delta", "network_slot_unlocks", "route_knowledge", "clout_reach_note"], "offer payout");
  const slots = parseNetworkSlots(preview.network_slot_unlocks);
  if (typeof preview.clout_reach_note !== "string" || preview.clout_reach_note === "") throw new SyntaxError("invalid clout note");
  return { market_modifier_ppm: safeInteger(source.market_modifier_ppm, 0, 2_000_000), payout_preview: { reputation_delta: safeInteger(preview.reputation_delta, 0, MAX_EXACT_INTEGER), network_slot_unlocks: slots, route_knowledge: safeInteger(preview.route_knowledge, 0, MAX_EXACT_INTEGER), clout_reach_note: preview.clout_reach_note } };
}

function parseNetworkSlots(source: unknown): NetworkSlot[] {
  let last = "";
  return array(source, "network slots").map((item) => {
    const raw = exactObject(item, ["slot", "carried_ref"], "network slot");
    const slot = mechanicalString(raw.slot); const carried_ref = mechanicalString(raw.carried_ref);
    if (byteCompare(slot, last) <= 0) throw new SyntaxError("network slots must be sorted and unique");
    last = slot;
    return { slot, carried_ref };
  });
}

function applyTermsModifier(terms: ExitTerms, factor: number): ExitTerms {
  return { ...terms, reputation_delta: ppmFloor(terms.reputation_delta, factor), route_knowledge: ppmFloor(terms.route_knowledge, factor) };
}

function promiseTerms(preview: ExitTerms, current: ExitTerms): ExitTerms {
  const slots = new Map<string, NetworkSlot>();
  for (const slot of current.network_slot_unlocks) slots.set(slot.slot, slot);
  for (const slot of preview.network_slot_unlocks) slots.set(slot.slot, slot);
  return { ...current, reputation_delta: Math.max(preview.reputation_delta, current.reputation_delta), route_knowledge: Math.max(preview.route_knowledge, current.route_knowledge), network_slot_unlocks: [...slots.values()].sort((left, right) => byteCompare(left.slot, right.slot)) };
}

function finishLoggedExit(company: ReplayState, founder: FounderCarry, intentId: string, command: ReplayCommand, nowMs: number, exitType: string, inputTerms: ExitTerms, prefix: ReplayEvent[], executedRoutes: string[], next: ReplayCatalogBundle): LoggedExitTransition {
  const attended = attendedMS(company, nowMs);
  const terms = { ...inputTerms, reputation_delta: Math.min(inputTerms.reputation_delta, MAX_EXACT_INTEGER - founder.reputation_level) };
  if (terms.route_knowledge > MAX_EXACT_INTEGER - founder.route_knowledge_balance || attended > MAX_EXACT_INTEGER - founder.age_ms) throw new RangeError("founder carry overflow");
  founder.reputation_level += terms.reputation_delta;
  founder.route_knowledge_balance += terms.route_knowledge;
  founder.age_ms += attended;
  founder.ledger_fact_kinds = [...new Set([...founder.ledger_fact_kinds, ...company.ledgerFactKinds])].sort(byteCompare);
  const slots = new Map<string, NetworkSlot>();
  for (const slot of founder.network_slots) slots.set(slot.slot, slot);
  for (const slot of terms.network_slot_unlocks) slots.set(slot.slot, slot);
  founder.network_slots = [...slots.values()].sort((left, right) => byteCompare(left.slot, right.slot));
  founder.exit_history_count++;
  bindEventIntent(prefix, intentId);
  const currentActive = company.wireVersion === 16;
  if (currentActive) {
    if (!foundationsActive(next)) throw new RangeError("foundation mechanics cannot disappear between epochs");
    const lifetime = new Set(founder.achievements_earned_lifetime);
    for (const achievementId of company.achievementsEarnedRun) if (lifetime.has(achievementId)) throw new RangeError("run achievement already owned for life");
    founder.achievements_earned_lifetime = [...new Set([...founder.achievements_earned_lifetime, ...company.achievementsEarnedRun])].sort(byteCompare);
    founder.achievement_score_lifetime = achievementScore(next.achievements, new Set(founder.achievements_earned_lifetime));
  } else if (foundationsActive(next)) {
    founder.achievements_earned_lifetime = [];
    founder.achievement_score_lifetime = 0;
  }
  const newCompany = newRunState(next, company, founder, nowMs);
  const runID = { company_stream_id: command.company_stream_id, run_seq: company.runSeq };
  const founderEvent = event("founder_advanced", intentId, { exit_type: exitType, founder_id: command.founder_id, occurred_at_ms: nowMs, reputation_delta: terms.reputation_delta, route_knowledge: terms.route_knowledge, run_id: runID });
  const endedEvent = event("run_ended", intentId, { assisted: { advisor: founder.advisor_mode, commons: company.compactMember }, attended_ms: attended, ended_at_ms: nowMs, executed_routes: executedRoutes, exit_type: exitType, faction: company.factionId || null, founder_id: command.founder_id, gates_crossed: Object.keys(company.gatesCrossed).filter((gate) => company.gatesCrossed[gate]).sort(byteCompare), generators_purchased_total: company.generatorPurchasedTotal, ledger_fact_kinds: [...company.ledgerFactKinds].sort(byteCompare), lifetime_value: company.lifetimeValue, payout: terms, pre_timer: company.runPreTimer, rta_ms: nowMs - company.runStartedAtMs, run_id: runID, started_at_ms: company.runStartedAtMs, terminal_seq: command.run_log_seq, tier: company.tier }, 2);
  const startedEvent = event("run_started", intentId, { assisted: { advisor: founder.advisor_mode, commons: false }, founder_id: command.founder_id, run_id: { company_stream_id: command.company_stream_id, run_seq: newCompany.runSeq }, started_at_ms: nowMs });
  const receipt = { applied_count: 1, evaluated_at: rfc3339(nowMs), founder_revision: founder.founder_revision + 1, intent_id: intentId, new_revision: command.revision + 2, outcome: "applied", receipt: { changes: [] }, snapshot: wireSnapshot(newCompany, next.economy) };
  return { founder, finalCompany: company, newCompany, outcome: "applied", receipt, founderEvents: [founderEvent], companyEndedEvents: [...prefix, endedEvent], companyStartedEvents: [startedEvent] };
}

function newRunState(bundle: ReplayCatalogBundle, prior: ReplayState, founder: FounderCarry, nowMs: number): ReplayState {
  const catalog = bundle.economy;
  if (prior.runSeq >= MAX_EXACT_INTEGER) throw new RangeError("run sequence exhausted");
  const balances = Object.fromEntries(catalog.resources.filter((value) => value.scope === "company").map((value) => [value.id, value.minimum]));
  const generators = Object.fromEntries(catalog.generatorClasses.filter((value) => value.production !== null).map((value) => [value.id, 0]));
  const active = foundationsActive(bundle);
  const meterState = active ? newRunMeterState(bundle.meters, founder.notoriety) : null;
  const reseed = founder.notoriety >= 100 ? 55 : Math.max(55, Math.min(90, 90 - Math.floor(founder.notoriety * 35 / 100)));
  return { wireVersion: active ? 16 : 14, balances, generators, generatorPurchasedTotal: 0, upgradesOwned: new Set(), generatorsProvisioned: Object.fromEntries(Object.keys(generators).map((id) => [id, 0])), provisionRemaindersPpm: Object.fromEntries(catalog.generatorClasses.filter((value) => value.provision !== null).map((value) => [value.provision!.generatorId, 0])), stockRateRemainderPpm: 0, evaluatedThroughMs: nowMs, computeCreditMs: 0, manualTokenMilli: catalog.manualPolicy!.bucketCapMilli, manualTokenRefilledAtMs: nowMs, gatesCrossed: {}, runSeq: prior.runSeq + 1, doctrinesByTransition: {}, structureId: "", ledgerFactKinds: new Set(), meterBands: active ? {} : { "trust.regulators.standing": reseed, "trust.regulators.grievance": 100 - reseed }, regionTraits: new Set(), routeKnowledgeBalance: 0, hintsUnlocked: new Set(), compactMember: false, compactTithePpm: 0, compactSolidarityPpm: 0, compactSamples: [], tier: 0, lifetimeValue: "0", offerState: null, runStartedAtMs: nowMs, runPreTimer: false, offlineSpans: [], collapsedOfflineMs: 0, factionId: "", incorporatedAtMs: null, factionStockResource: "", stockUnits: 0, stockProgressMs: 0, consumedStockUnits: 0, guildTitheCarryPpm: 0, guildBoundaryGuildId: prior.guildBoundaryGuildId, guildBoundarySeq: prior.guildBoundarySeq, guildConsumedWindow: 0, meterValues: meterState?.values ?? {}, meterDecayRemainders: meterState?.decayRemainders ?? {}, meterInputRemainders: meterState?.inputRemainders ?? {}, achievementsEarnedRun: new Set(), achievementScoreRun: 0 };
}

function rejectedExit(company: ReplayState, founder: FounderCarry, intentId: string, revision: number, category: string, detail: string): LoggedExitTransition {
  return { founder, finalCompany: company, newCompany: null, outcome: "rejected", receipt: { current_revision: revision, intent_id: intentId, outcome: "rejected", rejection: { category, detail } }, founderEvents: [], companyEndedEvents: [], companyStartedEvents: [] };
}

function sortedUniqueMechanical(source: unknown[]): string[] {
  let last = "";
  return source.map((item) => { const value = mechanicalString(item); if (byteCompare(value, last) <= 0) throw new SyntaxError("values must be sorted and unique"); last = value; return value; });
}

function parseReplayWire(source: unknown, state: ReplayState, catalogs: ReplayCatalogBundle): ReplayWire { const root = exactObject(source, ["v", "command", "evaluated_at_ms", "evaluation_mode", "resolved"], "replay inputs"); if (root.v !== 2 && root.v !== 3 && root.v !== 4 || foundationsActive(catalogs) && root.v < 3 || root.evaluation_mode !== "online" && root.evaluation_mode !== "offline") throw new SyntaxError("invalid replay envelope"); const command = objectWithOnlyKeys(root.command, ["intent_id", "company_stream_id", "founder_id", "revision", "run_seq", "run_log_seq"], "command"); const parsed: ReplayCommand = { intent_id: uuidV7String(command.intent_id), company_stream_id: command.company_stream_id === undefined ? "" : string(command.company_stream_id), founder_id: command.founder_id === undefined ? "" : string(command.founder_id), revision: safeInteger(command.revision, 1, MAX_EXACT_INTEGER), run_seq: safeInteger(command.run_seq, 1, MAX_EXACT_INTEGER), run_log_seq: safeInteger(command.run_log_seq, 1, MAX_EXACT_INTEGER) }; if (parsed.run_seq !== state.runSeq || !hashPattern.test(catalogs.constantsHash)) throw new RangeError("replay command mismatch"); return { v: root.v, command: parsed, evaluated_at_ms: safeInteger(root.evaluated_at_ms, 1, MAX_EXACT_INTEGER), evaluation_mode: root.evaluation_mode, resolved: objectWithOnlyKeys(root.resolved, Object.keys(root.resolved as object), "resolved") }; }
function parseFounderReplayWire(source: unknown): FounderReplayWire {
  const root = exactObject(source, ["v", "command", "evaluated_at_ms", "resolved"], "Founder replay inputs");
  if (root.v !== 1) throw new SyntaxError("invalid Founder replay version");
  const raw = exactObject(root.command, ["intent_id", "founder_stream_id", "founder_id", "revision", "founder_log_seq", "server_ts_ms"], "Founder command");
  const command: FounderReplayCommand = { intent_id: uuidV7String(raw.intent_id), founder_stream_id: uuidString(raw.founder_stream_id), founder_id: uuidString(raw.founder_id),
    revision: safeInteger(raw.revision, 1, MAX_EXACT_INTEGER), founder_log_seq: safeInteger(raw.founder_log_seq, 1, MAX_EXACT_INTEGER), server_ts_ms: safeInteger(raw.server_ts_ms, 1, MAX_EXACT_INTEGER) };
  const evaluated = safeInteger(root.evaluated_at_ms, 1, MAX_EXACT_INTEGER);
  if (evaluated !== command.server_ts_ms) throw new RangeError("Founder replay clock mismatch");
  return { v: 1, command, evaluated_at_ms: evaluated, resolved: objectWithOnlyKeys(root.resolved, Object.keys(root.resolved as object), "Founder resolved inputs") };
}

function parseFounderNetworkSlots(source: unknown): NetworkSlot[] {
  let last = "";
  return array(source, "Founder network slots").map((item) => { const raw = exactObject(item, ["slot", "carried_ref"], "Founder network slot"); const slot = mechanicalString(raw.slot); const carriedRef = mechanicalString(raw.carried_ref); if (byteCompare(slot, last) <= 0) throw new SyntaxError("Founder network slots must be sorted and unique"); last = slot; return { slot, carried_ref: carriedRef }; });
}

function parseFounderExitHistory(source: unknown): FounderExitRecord[] {
  let last = 0;
  return array(source, "Founder exit history").map((item) => { const raw = exactObject(item, ["run_id", "exit_type", "occurred_at_ms", "reputation_delta"], "Founder exit record"); const runId = safeInteger(raw.run_id, 1, MAX_EXACT_INTEGER); const exitType = string(raw.exit_type); if (runId <= last || !["acquihire", "acquisition", "ipo", "collapse", "scripted_first"].includes(exitType)) throw new SyntaxError("invalid Founder exit history"); last = runId; return { run_id: runId, exit_type: exitType, occurred_at_ms: safeInteger(raw.occurred_at_ms, 1, MAX_EXACT_INTEGER), reputation_delta: safeInteger(raw.reputation_delta, 0, MAX_EXACT_INTEGER) }; });
}

function founderRejected(state: FounderReplayState, intentId: string, revision: number, category: string, detail: string, resultHash: string, rollback?: () => void): FounderLoggedTransition {
  rollback?.();
  return { state, outcome: "rejected", receipt: { current_revision: revision, intent_id: intentId, outcome: "rejected", rejection: { category, detail } }, events: [], resultConstantsHash: resultHash };
}

function founderWireSnapshot(state: FounderReplayState): unknown {
  const snapshot: Record<string, unknown> = { balances: sortedRecord(state.balances), collapsed_offline_ms: 0, compact_member: false, compact_solidarity_ppm: 0, compact_tithe_ppm: 0,
    compute_credit_ms: 0, consumed_stock_units: 0, doctrines_by_transition: {}, evaluated_through: rfc3339(state.evaluatedThroughMs), faction_id: null, gates_crossed: {},
    generators: sortedRecord(state.generators), generators_provisioned: sortedRecord(state.generatorsProvisioned), generators_purchased_total: state.generatorPurchasedTotal,
    guild_boundary_guild_id: null, guild_boundary_seq: 0, guild_consumed_window_units: 0, guild_tithe_carry_ppm: 0, hints_unlocked: [...state.hintsUnlocked].sort(byteCompare),
    incorporated_at_ms: null, ledger_fact_kinds: [...state.ledgerFactKinds].sort(byteCompare), lifetime_value: "0", manual_token_milli: 0,
    manual_token_refilled_at: rfc3339(state.manualTokenRefilledAtMs), offer_state: null, offline_spans: [], provision_remainders_ppm: sortedRecord(state.provisionRemaindersPpm),
    provisioned_hardcaps: {}, region_traits: [], route_knowledge_balance: state.routeKnowledgeBalance, run_pre_timer: false, run_seq: 0, run_started_at_ms: 0,
    stock_progress_ms: 0, stock_rate_remainder_ppm: 0, stock_resource: null, stock_units: 0, structure_id: "", tier: 0, upgrades_owned: [...state.upgradesOwned].sort(byteCompare) };
  if (state.wireVersion >= 16) Object.assign(snapshot, { meter_values: {}, meter_decay_remainders: {}, meter_input_remainders: {}, achievements_earned_run: [], achievement_score_run: 0 });
  else snapshot.meter_bands = {};
  return snapshot;
}

function applyFounderExit(state: FounderReplayState, request: Intent, wire: FounderReplayWire, catalogs: ReplayCatalogBundle): FounderLoggedTransition {
	const inputHash = catalogs.constantsHash;
  const keys = ["kind", "outcome", "company_stream_id", "run_seq", "run_log_seq", "result_constants_hash", "reputation_delta", "route_knowledge_delta", "attended_ms", "age_ms_before", "age_ms_after", "achievement_score_delta", "added_network_slots", "added_ledger_fact_kinds", "added_lifetime_achievements", "exit_record", "result_founder_wire_version", "rejection"];
  const raw = exactObject(wire.resolved, keys, "Founder Exit inputs");
  if (request.kind !== "cross_gate" && request.expected_founder_revision !== wire.command.revision) throw new RangeError("Founder Exit revision mismatch");
  const companyStreamId = uuidString(raw.company_stream_id); const runSeq = safeInteger(raw.run_seq, 1, MAX_EXACT_INTEGER); safeInteger(raw.run_log_seq, 1, MAX_EXACT_INTEGER);
  const resultHash = string(raw.result_constants_hash); if (!hashPattern.test(resultHash)) throw new SyntaxError("invalid Founder result hash");
  const ageBefore = safeInteger(raw.age_ms_before, 0, MAX_EXACT_INTEGER); const ageAfter = safeInteger(raw.age_ms_after, 0, MAX_EXACT_INTEGER); const attended = safeInteger(raw.attended_ms, 0, MAX_EXACT_INTEGER);
  if (ageBefore !== state.ageMs || ageAfter < ageBefore || attended !== ageAfter - ageBefore) throw new RangeError("invalid Founder attendance facts");
  const resultVersion = safeInteger(raw.result_founder_wire_version, 1, 18);
  if (![14, 15, 16, 17, 18].includes(resultVersion)) throw new RangeError("unsupported Founder result version");
  const outcome = string(raw.outcome);
  if (outcome === "rejected") {
    const rejection = exactObject(raw.rejection, ["category", "detail"], "Founder Exit rejection"); const category = string(rejection.category); const detail = string(rejection.detail);
    if (resultHash !== inputHash || safeInteger(raw.reputation_delta, 0, MAX_EXACT_INTEGER) !== 0 || safeInteger(raw.route_knowledge_delta, 0, MAX_EXACT_INTEGER) !== 0 || attended !== 0 || safeInteger(raw.achievement_score_delta, 0, MAX_EXACT_INTEGER) !== 0 || array(raw.added_network_slots, "rejected Founder slots").length !== 0 || array(raw.added_ledger_fact_kinds, "rejected Founder facts").length !== 0 || array(raw.added_lifetime_achievements, "rejected Founder achievements").length !== 0 || raw.exit_record !== null || resultVersion !== state.wireVersion) throw new RangeError("non-neutral rejected Founder Exit");
    return founderRejected(state, wire.command.intent_id, wire.command.revision, category, detail, inputHash);
  }
  if (outcome !== "applied" || raw.rejection !== null) throw new RangeError("invalid Founder Exit outcome");
  const reputationDelta = safeInteger(raw.reputation_delta, 0, MAX_EXACT_INTEGER); const routeDelta = safeInteger(raw.route_knowledge_delta, 0, MAX_EXACT_INTEGER); const achievementDelta = safeInteger(raw.achievement_score_delta, 0, MAX_EXACT_INTEGER);
  if (reputationDelta > MAX_EXACT_INTEGER - state.reputationLevel || routeDelta > MAX_EXACT_INTEGER - state.routeKnowledgeBalance || achievementDelta > MAX_EXACT_INTEGER - state.achievementScoreLifetime) throw new RangeError("Founder Exit overflow");
  const addedSlots = parseFounderNetworkSlots(raw.added_network_slots); const addedFacts = sortedUniqueMechanical(array(raw.added_ledger_fact_kinds, "added Founder facts")); const addedAchievements = sortedUniqueMechanical(array(raw.added_lifetime_achievements, "added Founder achievements"));
  const exit = parseFounderExitHistory([raw.exit_record])[0]!;
  if (exit.run_id !== runSeq || exit.reputation_delta !== reputationDelta || state.exitHistory.some((value) => value.run_id >= exit.run_id)) throw new RangeError("invalid appended Founder Exit record");
  state.reputationLevel += reputationDelta; state.routeKnowledgeBalance += routeDelta; state.ageMs = ageAfter; state.achievementScoreLifetime += achievementDelta;
  const slotMap = new Map(state.networkSlots.map((value) => [value.slot, value])); for (const slot of addedSlots) slotMap.set(slot.slot, slot); state.networkSlots = [...slotMap.values()].sort((a, b) => byteCompare(a.slot, b.slot));
  for (const fact of addedFacts) state.ledgerFactKinds.add(fact); for (const id of addedAchievements) state.achievementsEarnedLifetime.add(id);
  state.exitHistory.push(exit);
  const resultCatalogs = resultHash === inputHash ? catalogs : catalogs.next;
  if (!resultCatalogs || resultCatalogs.constantsHash !== resultHash || resultVersion < state.wireVersion) throw new RangeError("missing Founder result catalogs");
  if (resultVersion >= 17 && state.wireVersion < 17) {
    if (!resultCatalogs.minigames || resultCatalogs.minigames.minigameIds.length !== 0) throw new RangeError("unsupported minigame activation content");
    state.minigameRatings = {}; state.minigameOfflineQuality = {};
  }
  if (resultVersion >= 18 && state.wireVersion < 18) {
    if (!resultCatalogs.pets) throw new RangeError("missing pet activation artifact"); state.pets = {};
  }
  state.wireVersion = resultVersion as 14 | 15 | 16 | 17 | 18;
  Object.assign(state, restoreFounderReplayState(encodeFounderReplayState(state), resultVersion, resultCatalogs));
  const receipt = { intent_id: wire.command.intent_id, outcome: "applied", founder_revision: wire.command.revision + 1, result_constants_hash: resultHash };
  const founderEvent = event("founder_advanced", wire.command.intent_id, { founder_id: wire.command.founder_id, run_id: { company_stream_id: companyStreamId, run_seq: runSeq }, exit_type: exit.exit_type, reputation_delta: reputationDelta, route_knowledge: routeDelta, occurred_at_ms: exit.occurred_at_ms });
  return { state, outcome: "applied", receipt, events: [founderEvent], resultConstantsHash: resultHash };
}
function parseAccrual(source: unknown, catalogs: ReplayCatalogBundle): ReplayAccrual {
  const raw = objectWithOnlyKeys(source, ["contributions", "commons_weight_ppm", "guild_settlement_batch", "route_context_version"], "accrual");
  const commonsWeight = raw.commons_weight_ppm ?? null;
  if (commonsWeight !== null) safeInteger(commonsWeight, 0, 1_000_000);
  const routeContextVersion = raw.route_context_version ?? 0;
  if (routeContextVersion !== catalogs.routes.contextVersion) throw new RangeError("route context mismatch");
  let lastContribution = "";
  const contributions = array(raw.contributions, "contributions").map((item) => {
    const value = exactObject(item, ["slot", "source_id", "target", "factor"], "contribution");
    if (typeof value.slot !== "string" || !(MULTIPLIER_SLOT_ORDER as readonly string[]).includes(value.slot) || typeof value.source_id !== "string" || !mechanical.test(value.source_id) || typeof value.target !== "string" || !mechanical.test(value.target) && value.target !== "all") throw new SyntaxError("invalid contribution");
    const factor = canonical(value.factor);
    if (!parseCanonical(factor).gt(0)) throw new SyntaxError("non-positive contribution");
    const key = `${value.slot}\0${value.source_id}\0${value.target}`;
    if (lastContribution !== "" && byteCompare(key, lastContribution) <= 0) throw new SyntaxError("contributions must be sorted and unique");
    lastContribution = key;
    return { slot: value.slot as MultiplierSlot, source_id: value.source_id, target: value.target, factor };
  });
  const rawBatch = objectWithOnlyKeys(raw.guild_settlement_batch, ["guild_id", "base_seq", "settlements"], "settlement batch");
  const guildId = string(rawBatch.guild_id); const baseSeq = safeInteger(rawBatch.base_seq, 0, MAX_EXACT_INTEGER);
  if (guildId !== "") uuidV7String(guildId);
  let lastSettlement = baseSeq;
  const settlements = array(rawBatch.settlements, "settlements").map((item) => {
    const value = objectWithOnlyKeys(item, ["boundary_seq", "debit_units", "credit_units"], "settlement");
    const settlement = { boundary_seq: safeInteger(value.boundary_seq, 1, MAX_EXACT_INTEGER), debit_units: safeInteger(value.debit_units, 0, MAX_EXACT_INTEGER), credit_units: safeInteger(value.credit_units, 0, MAX_EXACT_INTEGER) };
    if (settlement.boundary_seq <= lastSettlement) throw new SyntaxError("settlements must be sorted and unique");
    lastSettlement = settlement.boundary_seq; return settlement;
  });
  if (guildId === "" && (baseSeq !== 0 || settlements.length !== 0)) throw new SyntaxError("invalid empty settlement batch");
  return { contributions, commons_weight_ppm: commonsWeight as number | null, guild_settlement_batch: { guild_id: guildId, base_seq: baseSeq, settlements }, route_context_version: routeContextVersion as number };
}

function applyGuildSettlements(state: ReplayState, batch: ReplayGuildSettlementBatch, stockCap: number): void {
  if (batch.guild_id === "") return;
  if (state.guildBoundaryGuildId !== batch.guild_id) {
    if (state.guildBoundaryGuildId === "") { if (batch.base_seq < state.guildBoundarySeq || batch.base_seq > state.guildBoundarySeq && batch.settlements.length !== 0) throw new RangeError("guild settlement base mismatch"); }
    else if (batch.settlements.length !== 0) throw new RangeError("guild settlement switch carries results");
    state.guildBoundaryGuildId = batch.guild_id; state.guildBoundarySeq = batch.base_seq; state.guildConsumedWindow = 0;
  } else if (batch.base_seq !== state.guildBoundarySeq) throw new RangeError("guild settlement base mismatch");
  for (const settlement of batch.settlements) {
    if (settlement.boundary_seq <= state.guildBoundarySeq || settlement.debit_units > state.stockUnits || settlement.credit_units > stockCap - state.consumedStockUnits) throw new RangeError("invalid guild settlement");
    state.stockUnits -= settlement.debit_units; state.consumedStockUnits += settlement.credit_units;
    state.guildConsumedWindow = settlement.credit_units; state.guildBoundarySeq = settlement.boundary_seq;
  }
}

function parseFounderCarry(source: unknown, catalogs: ReplayCatalogBundle, wireVersion: 2 | 3 | 4): FounderCarry {
  const legacyKeys = ["founder_revision", "founder_constants_hash", "reputation_level", "route_knowledge_balance", "age_ms", "notoriety", "advisor_mode", "network_slots", "ledger_fact_kinds", "exit_history_count"];
  const carry = { ...objectWithOnlyKeys(source, wireVersion >= 3 ? [...legacyKeys, "achievements_earned_lifetime", "achievement_score_lifetime"] : legacyKeys, "founder carry") };
  safeInteger(carry.founder_revision, 1, MAX_EXACT_INTEGER);
  if (carry.founder_constants_hash !== catalogs.constantsHash) throw new RangeError("founder catalog mismatch");
  carry.reputation_level = safeInteger(carry.reputation_level ?? 0, 0, MAX_EXACT_INTEGER);
  carry.route_knowledge_balance = safeInteger(carry.route_knowledge_balance ?? 0, 0, MAX_EXACT_INTEGER);
  carry.age_ms = safeInteger(carry.age_ms ?? 0, 0, MAX_EXACT_INTEGER);
  carry.notoriety = safeInteger(carry.notoriety ?? 0, 0, MAX_EXACT_INTEGER);
  carry.advisor_mode = carry.advisor_mode ?? false; boolean(carry.advisor_mode);
  carry.exit_history_count = safeInteger(carry.exit_history_count ?? 0, 0, MAX_EXACT_INTEGER);
  const earnedLifetime = wireVersion >= 3 ? sortedUniqueMechanical(array(carry.achievements_earned_lifetime, "lifetime achievements")) : [];
  const lifetimeScore = wireVersion >= 3 ? safeInteger(carry.achievement_score_lifetime, 0, MAX_EXACT_INTEGER) : 0;
  carry.achievements_earned_lifetime = earnedLifetime;
  carry.achievement_score_lifetime = lifetimeScore;
  if (foundationsActive(catalogs)) {
    if (wireVersion < 3 || achievementScore(catalogs.achievements, new Set(earnedLifetime)) !== lifetimeScore) throw new RangeError("invalid active Founder achievement carry");
  } else if (earnedLifetime.length !== 0 || lifetimeScore !== 0) throw new RangeError("legacy Founder carry contains active foundation state");
  let lastFact = "";
  for (const item of array(carry.ledger_fact_kinds, "founder facts")) {
    const fact = string(item);
    if (fact === "" || byteCompare(fact, lastFact) <= 0) throw new SyntaxError("founder facts must be sorted and unique");
    lastFact = fact;
  }
  let lastSlot = "";
  for (const item of array(carry.network_slots, "network slots")) {
    const slot = exactObject(item, ["slot", "carried_ref"], "network slot");
    const slotID = string(slot.slot);
    if (slotID === "" || byteCompare(slotID, lastSlot) <= 0 || string(slot.carried_ref) === "") throw new SyntaxError("network slots must be sorted and complete");
    lastSlot = slotID;
  }
  return carry as unknown as FounderCarry;
}

function parseIntent(payload: string, intentId: string): Intent {
  const source = parseJSON(payload);
  if (typeof source !== "object" || source === null || Array.isArray(source) || "intent_id" in source) throw new SyntaxError("invalid canonical payload");
  const raw = source as Record<string, unknown>;
  const kind = string(raw.kind);
  const base: Intent = { intent_id: intentId, kind, expected_revision: safeInteger(raw.expected_revision, 1, MAX_EXACT_INTEGER) };
  switch (kind) {
    case "buy_generator": {
      if (!hasExactKeys(raw, ["kind", "expected_revision", "generator_id", "count"])) { base.invalid = "buy_generator.fields"; return base; }
      if (!isMechanical(raw.generator_id)) base.invalid = "generator_id"; else base.generator_id = raw.generator_id;
      if (raw.count !== null && !isRecord(raw.count)) { base.invalid = "count"; return base; }
      const count = (raw.count ?? {}) as Record<string, unknown>;
      base.count = count;
      if (count.mode === "max") {
        if (!hasExactKeys(count, ["mode"])) base.invalid = "count.max";
      } else if (count.mode === "exact") {
        if (!hasExactKeys(count, ["mode", "value"]) || !isPositiveSafeInteger(count.value)) base.invalid = "count.exact";
      } else base.invalid = "count.mode";
      return base;
    }
    case "buy_upgrade":
      if (!hasExactKeys(raw, ["kind", "expected_revision", "upgrade_id"])) { base.invalid = "buy_upgrade.fields"; return base; }
      if (!isMechanical(raw.upgrade_id)) base.invalid = "upgrade_id"; else base.upgrade_id = raw.upgrade_id;
      return base;
    case "perform_manual_batch":
      if (!hasExactKeys(raw, ["kind", "expected_revision", "action_id", "count", "window_ms"])) { base.invalid = "perform_manual_batch.fields"; return base; }
      if (!isMechanical(raw.action_id)) base.invalid = "action_id"; else base.action_id = raw.action_id;
      if (!isPositiveSafeInteger(raw.count)) base.invalid = "count"; else base.count = raw.count;
      if (!isPositiveSafeInteger(raw.window_ms)) base.invalid = "window_ms"; else base.window_ms = raw.window_ms;
      return base;
    case "cross_gate":
      if (!hasExactKeys(raw, ["kind", "expected_revision", "gate_id", "route_id"])) { base.invalid = "cross_gate.fields"; return base; }
      if (!isMechanical(raw.gate_id)) base.invalid = "gate_id"; else base.gate_id = raw.gate_id;
      if (raw.route_id === null) base.route_id = null;
      else if (!isMechanical(raw.route_id)) base.invalid = "route_id";
      else base.route_id = raw.route_id;
      return base;
    case "buy_route_hint":
      if (!hasExactKeys(raw, ["kind", "expected_revision", "route_id"])) { base.invalid = "buy_route_hint.fields"; return base; }
      if (!isMechanical(raw.route_id)) base.invalid = "route_id"; else base.route_id = raw.route_id;
      return base;
    case "sign_compact":
      if (!hasExactKeys(raw, ["kind", "expected_revision", "tithe_ppm"])) { base.invalid = "sign_compact.fields"; return base; }
      if (!isNonNegativeSafeInteger(raw.tithe_ppm) || raw.tithe_ppm > 1_000_000) base.invalid = "tithe_ppm"; else base.tithe_ppm = raw.tithe_ppm;
      return base;
    case "leave_compact":
      if (!hasExactKeys(raw, ["kind", "expected_revision"])) base.invalid = "leave_compact.fields";
      return base;
    case "incorporate":
      if (!hasExactKeys(raw, ["kind", "expected_revision", "faction_id"])) { base.invalid = "incorporate.fields"; return base; }
      if (!isMechanical(raw.faction_id)) base.invalid = "faction_id"; else base.faction_id = raw.faction_id;
      return base;
    case "accept_exit_offer":
      if (!hasExactKeys(raw, ["kind", "expected_revision", "expected_founder_revision", "offer_id"])) { base.invalid = "accept_exit_offer.fields"; return base; }
      if (!isPositiveSafeInteger(raw.expected_founder_revision)) base.invalid = "expected_founder_revision"; else base.expected_founder_revision = raw.expected_founder_revision;
      if (!isUUIDV7(raw.offer_id)) base.invalid = "offer_id"; else base.offer_id = raw.offer_id;
      return base;
    case "wind_down":
    case "file_ipo":
      if (!hasExactKeys(raw, ["kind", "expected_revision", "expected_founder_revision"])) { base.invalid = `${kind}.fields`; return base; }
      if (!isPositiveSafeInteger(raw.expected_founder_revision)) base.invalid = "expected_founder_revision"; else base.expected_founder_revision = raw.expected_founder_revision;
      return base;
    case "decline_exit_offer":
      if (!hasExactKeys(raw, ["kind", "expected_revision", "offer_id"])) { base.invalid = "decline_exit_offer.fields"; return base; }
      if (!isUUIDV7(raw.offer_id)) base.invalid = "offer_id"; else base.offer_id = raw.offer_id;
      return base;
    default:
      base.invalid = kind;
      return base;
  }
}

function applied(state: ReplayState, catalog: EconomyCatalog, intentId: string, revision: number, count: number, before: Record<string, string>, events: ReplayEvent[], invariants: ReplayInvariant[]): LoggedTransition { bindEventIntent(events, intentId); const changes = Object.keys(before).sort(byteCompare).flatMap((resource) => before[resource] === state.balances[resource] ? [] : [{ resource_id: resource, before: before[resource]!, delta: canonicalString(quantize(parseCanonical(state.balances[resource]!).sub(parseCanonical(before[resource]!)))), after: state.balances[resource]! }]); return { state, outcome: "applied", receipt: { applied_count: count, evaluated_at: rfc3339(state.evaluatedThroughMs), intent_id: intentId, new_revision: revision, outcome: "applied", receipt: { changes }, snapshot: wireSnapshot(state, catalog) }, events, invariants }; }
function bindEventIntent(events: ReplayEvent[], intentId: string): void { for (const value of events) { if (value.intent_id !== "" && value.intent_id !== intentId) throw new RangeError("event intent mismatch"); (value as { intent_id: string }).intent_id = intentId; } }
function rejected(state: ReplayState, intentId: string, revision: number, category: string, detail: string): LoggedTransition { return { state, outcome: "rejected", receipt: { current_revision: revision, intent_id: intentId, outcome: "rejected", rejection: { category, detail } }, events: [], invariants: [] }; }
function event(kind: string, intentId: string, payload: unknown, schemaVersion = 1): ReplayEvent { return { kind, schema_version: schemaVersion, intent_id: intentId, payload }; }
function compactMembershipPayload(command: ReplayCommand, runSeq: number, tithe: number, prior: boolean, next: boolean): unknown { return { founder_id: command.founder_id, new_member: next, prior_member: prior, run_id: { company_stream_id: command.company_stream_id, run_seq: runSeq }, tithe_ppm: tithe }; }
function deriveFactionStockResource(state: ReplayState, catalog: FactionCatalog): void { if (state.factionId === "") { if (state.factionStockResource !== "") throw new RangeError("orphan stock resource"); return; } const member = catalog.byId.get(state.factionId); if (!member) throw new RangeError("unknown faction state"); state.factionStockResource = member.produces; }
function progressPpm(catalog: EconomyCatalog, state: ReplayState): number { if (!catalog.progressCoordinates.some((value) => value.tier === state.tier)) return 0; return Math.floor(subProgressValue(catalog, { balances: state.balances, generatorCounts: state.generators }, state.tier).mul(1_000_000).toNumber()); }
function recordOfflineSpan(state: ReplayState, fromMs: number, toMs: number, ceiling: number): void { if (toMs - fromMs <= ceiling) return; const last = state.offlineSpans.at(-1); if (last && fromMs <= last.toMs) { last.toMs = Math.max(last.toMs, toMs); return; } if (state.offlineSpans.length === 256) { const removed = state.offlineSpans.shift()!; state.collapsedOfflineMs += removed.toMs - removed.fromMs; } state.offlineSpans.push({ fromMs, toMs }); }
function compliancePpm(tithe: number, defaultTithe: number, enclosure: string): number { if (defaultTithe <= 0) return 0; const titheRatio = Decimal.min(1, new Decimal(tithe).div(defaultTithe)); const clean = new Decimal(1).sub(parseCanonical(enclosure)); return Math.max(0, Math.min(1_000_000, Math.floor(titheRatio.mul(clean).mul(1_000_000).toNumber()))); }
function appendCompactSamples(state: ReplayState, start: number, end: number, compliance: number, windowMs: number): void { while (start < end) { const hour = Math.floor(start / 3_600_000) * 3_600_000; const boundary = Math.min(end, hour + 3_600_000); const covered = boundary - start; const last = state.compactSamples.at(-1); if (last?.hourStartMs === hour) { const numerator = BigInt(last.compliancePpm) * BigInt(last.coveredMs) + BigInt(compliance) * BigInt(covered); last.coveredMs += covered; last.compliancePpm = Number(numerator / BigInt(last.coveredMs)); } else state.compactSamples.push({ hourStartMs: hour, compliancePpm: compliance, coveredMs: covered }); start = boundary; } const cutoff = Math.floor((end - windowMs) / 3_600_000) * 3_600_000; state.compactSamples = state.compactSamples.filter((value) => value.hourStartMs >= cutoff); const numerator = state.compactSamples.reduce((sum, value) => sum + BigInt(value.compliancePpm) * BigInt(value.coveredMs), 0n); state.compactSolidarityPpm = Math.max(0, Math.min(1_000_000, Number(numerator / BigInt(windowMs)))); }

function wireSnapshot(state: ReplayState, catalog: EconomyCatalog): unknown {
  const snapshot: Record<string, unknown> = { balances: sortedRecord(state.balances), collapsed_offline_ms: state.collapsedOfflineMs, compact_member: state.compactMember, compact_solidarity_ppm: state.compactSolidarityPpm, compact_tithe_ppm: state.compactTithePpm, compute_credit_ms: state.computeCreditMs, consumed_stock_units: state.consumedStockUnits, doctrines_by_transition: sortedRecord(state.doctrinesByTransition), evaluated_through: rfc3339(state.evaluatedThroughMs), faction_id: state.factionId || null, gates_crossed: sortedRecord(state.gatesCrossed), generators: sortedRecord(state.generators), generators_provisioned: sortedRecord(state.generatorsProvisioned), generators_purchased_total: state.generatorPurchasedTotal, guild_boundary_guild_id: state.guildBoundaryGuildId || null, guild_boundary_seq: state.guildBoundarySeq, guild_consumed_window_units: state.guildConsumedWindow, guild_tithe_carry_ppm: state.guildTitheCarryPpm, hints_unlocked: [...state.hintsUnlocked].sort(byteCompare), incorporated_at_ms: state.incorporatedAtMs, ledger_fact_kinds: [...state.ledgerFactKinds].sort(byteCompare), lifetime_value: state.lifetimeValue, manual_token_milli: state.manualTokenMilli, manual_token_refilled_at: rfc3339(state.manualTokenRefilledAtMs), offer_state: state.offerState ? { expires_at_ms: state.offerState.expiresAtMs, exit_type: state.offerState.exitType, offer_id: state.offerState.offerId, spawned_at_ms: state.offerState.spawnedAtMs, terms_json: state.offerState.terms } : null, offline_spans: state.offlineSpans.map((value) => ({ from_ms: value.fromMs, to_ms: value.toMs })), provision_remainders_ppm: sortedRecord(state.provisionRemaindersPpm), provisioned_hardcaps: provisionedHardcaps(catalog), region_traits: [...state.regionTraits].sort(byteCompare), route_knowledge_balance: state.routeKnowledgeBalance, run_pre_timer: state.runPreTimer, run_seq: state.runSeq, run_started_at_ms: state.runStartedAtMs, stock_progress_ms: state.stockProgressMs, stock_rate_remainder_ppm: state.stockRateRemainderPpm, stock_resource: state.factionStockResource || null, stock_units: state.stockUnits, structure_id: state.structureId, tier: state.tier, upgrades_owned: [...state.upgradesOwned].sort(byteCompare) };
  if (state.wireVersion === 16) Object.assign(snapshot, { meter_values: sortedRecord(state.meterValues), meter_decay_remainders: sortedRecord(state.meterDecayRemainders), meter_input_remainders: sortedRecord(state.meterInputRemainders), achievements_earned_run: [...state.achievementsEarnedRun].sort(byteCompare), achievement_score_run: state.achievementScoreRun });
  else snapshot.meter_bands = sortedRecord(state.meterBands);
  return snapshot;
}
function provisionedHardcaps(catalog: EconomyCatalog): Record<string, { count: number; reason_key: string }> { return Object.fromEntries(catalog.generatorClasses.filter((value) => value.provisionedHardcap !== null).sort((a, b) => byteCompare(a.id, b.id)).map((value) => [value.id, { count: value.provisionedHardcap!.count, reason_key: value.provisionedHardcap!.reasonKey }])); }
export function canonicalJSONString(value: unknown): string { if (value === null || typeof value !== "object") return JSON.stringify(value); if (Array.isArray(value)) return `[${value.map(canonicalJSONString).join(",")}]`; const object = value as Record<string, unknown>; return `{${Object.keys(object).sort(byteCompare).map((key) => `${JSON.stringify(key)}:${canonicalJSONString(object[key])}`).join(",")}}`; }

async function constantsHashArtifacts(artifacts: ReplayArtifacts): Promise<string> { const encoder = new TextEncoder(); const chunks: Uint8Array[] = []; for (const name of Object.keys(artifacts).sort(byteCompare) as (keyof ReplayArtifacts)[]) { const nameBytes = encoder.encode(name); const data = encoder.encode(artifacts[name]); chunks.push(frame(nameBytes.length), nameBytes, frame(data.length), data); } const total = chunks.reduce((sum, value) => sum + value.length, 0); const input = new Uint8Array(total); let offset = 0; for (const chunk of chunks) { input.set(chunk, offset); offset += chunk.length; } const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", input)); return `sha256:${[...digest].map((value) => value.toString(16).padStart(2, "0")).join("")}`; }
function frame(value: number): Uint8Array { const result = new Uint8Array(8); new DataView(result.buffer).setBigUint64(0, BigInt(value)); return result; }
function validateBalance(value: string, minimum: string, hardcap?: string): void { const parsed = parseCanonical(value); if (!isStateValue(parsed) || parsed.lt(parseCanonical(minimum)) || hardcap !== undefined && parsed.gt(parseCanonical(hardcap))) throw new SyntaxError("invalid balance"); }
function rfc3339(ms: number): string { return new Date(ms).toISOString().replace(/\.000Z$/, "Z").replace(/(\.\d*?)0+Z$/, "$1Z"); }
function parseJSON(source: string): unknown { return JSON.parse(source) as unknown; }
function isRecord(value: unknown): value is Record<string, unknown> { return typeof value === "object" && value !== null && !Array.isArray(value); }
function hasExactKeys(value: Record<string, unknown>, keys: readonly string[]): boolean { return same(Object.keys(value).sort(byteCompare), [...keys].sort(byteCompare)); }
function isMechanical(value: unknown): value is string { return typeof value === "string" && mechanical.test(value); }
function isUUIDV7(value: unknown): value is string { return typeof value === "string" && uuidV7.test(value); }
function isPositiveSafeInteger(value: unknown): value is number { return typeof value === "number" && Number.isSafeInteger(value) && value >= 1 && value <= MAX_EXACT_INTEGER; }
function isNonNegativeSafeInteger(value: unknown): value is number { return typeof value === "number" && Number.isSafeInteger(value) && value >= 0 && value <= MAX_EXACT_INTEGER; }
function objectWithOnlyKeys(source: unknown, keys: readonly string[], label: string): Record<string, unknown> { if (!isRecord(source)) throw new SyntaxError(`${label} must be an object`); onlyKeys(source, keys, label); return source; }
function onlyKeys(value: Record<string, unknown>, keys: readonly string[], label: string): void { const allowed = new Set(keys); if (Object.keys(value).some((key) => !allowed.has(key))) throw new SyntaxError(`${label} has unknown fields`); }
function exactObject(source: unknown, keys: readonly string[], label: string): Record<string, unknown> { if (typeof source !== "object" || source === null || Array.isArray(source)) throw new SyntaxError(`${label} must be an object`); const value = source as Record<string, unknown>; exactKeys(value, keys, label); return value; }
function exactRecord(source: unknown, keys: readonly string[], label: string): Record<string, unknown> { const value = exactObject(source, keys, label); return value; }
function exactKeys(value: Record<string, unknown>, keys: readonly string[], label: string): void { if (!same(Object.keys(value).sort(byteCompare), [...keys].sort(byteCompare))) throw new SyntaxError(`${label} fields are not exact`); }
function array(value: unknown, label: string): unknown[] { if (!Array.isArray(value)) throw new SyntaxError(`${label} must be an array`); return value; }
function string(value: unknown): string { if (typeof value !== "string") throw new SyntaxError("expected string"); return value; }
function mechanicalString(value: unknown): string { const result = string(value); if (!mechanical.test(result)) throw new SyntaxError("expected mechanical id"); return result; }
function uuidString(value: unknown): string { const result = string(value); if (!uuid.test(result)) throw new SyntaxError("expected UUID"); return result; }
function uuidV7String(value: unknown): string { const result = string(value); if (!uuidV7.test(result)) throw new SyntaxError("expected UUIDv7"); return result; }
function nullableUUID(value: unknown): string { return value === null ? "" : uuidV7String(value); }
function nullableMechanical(value: unknown, nullAllowed = true): string { if (value === null && nullAllowed) return ""; if (value === "" && !nullAllowed) return ""; return mechanicalString(value); }
function canonical(value: unknown): string { const result = string(value); if (canonicalString(parseCanonical(result)) !== result) throw new SyntaxError("non-canonical Decimal"); return result; }
function safeInteger(value: unknown, min: number, max: number): number { if (typeof value !== "number" || !Number.isSafeInteger(value) || value < min || value > max) throw new SyntaxError("integer outside exact domain"); return value; }
function boolean(value: unknown): boolean { if (typeof value !== "boolean") throw new SyntaxError("expected boolean"); return value; }
function stringRecord(value: unknown, label: string): Record<string, string> { if (typeof value !== "object" || value === null || Array.isArray(value)) throw new SyntaxError(`${label} must be an object`); const result: Record<string, string> = {}; for (const [key, item] of Object.entries(value)) { mechanicalString(key); result[key] = canonical(item); } return result; }
function integerRecord(value: unknown, min: number, max: number, label: string): Record<string, number> { if (typeof value !== "object" || value === null || Array.isArray(value)) throw new SyntaxError(`${label} must be an object`); const result: Record<string, number> = {}; for (const [key, item] of Object.entries(value)) { mechanicalString(key); result[key] = safeInteger(item, min, max); } return result; }
function plainIntegerRecord(value: unknown, min: number, max: number, label: string): Record<string, number> { if (typeof value !== "object" || value === null || Array.isArray(value)) throw new SyntaxError(`${label} must be an object`); const result: Record<string, number> = {}; for (const [key, item] of Object.entries(value)) result[key] = safeInteger(item, min, max); return result; }
function booleanRecord(value: unknown): Record<string, boolean> { if (typeof value !== "object" || value === null || Array.isArray(value)) throw new SyntaxError("expected boolean object"); const result: Record<string, boolean> = {}; for (const [key, item] of Object.entries(value)) { mechanicalString(key); result[key] = boolean(item); } return result; }
function mechanicalRecord(value: unknown): Record<string, string> { if (typeof value !== "object" || value === null || Array.isArray(value)) throw new SyntaxError("expected mechanical object"); const result: Record<string, string> = {}; for (const [key, item] of Object.entries(value)) { mechanicalString(key); result[key] = mechanicalString(item); } return result; }
function mechanicalSet(value: unknown): Set<string> { const result = new Set<string>(); for (const item of array(value, "mechanical set")) { const key = mechanicalString(item); if (result.has(key)) throw new SyntaxError("duplicate mechanical id"); result.add(key); } return result; }
function cursor(value: unknown, label: string): number { const source = string(value); const parsed = Date.parse(source); if (!Number.isSafeInteger(parsed) || parsed <= 0 || rfc3339(parsed) !== source) throw new SyntaxError(`invalid ${label}`); return parsed; }
function sortedRecord<T>(value: Record<string, T>): Record<string, T> { return Object.fromEntries(Object.keys(value).sort(byteCompare).map((key) => [key, value[key]!])); }
function same(left: readonly string[], right: readonly string[]): boolean { return left.length === right.length && left.every((value, index) => value === right[index]); }
function byteCompare(left: string, right: string): number { const a = new TextEncoder().encode(left); const b = new TextEncoder().encode(right); for (let i = 0; i < Math.min(a.length, b.length); i++) if (a[i] !== b[i]) return a[i]! - b[i]!; return a.length - b.length; }
