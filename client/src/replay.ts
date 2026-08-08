import Decimal from "break_infinity.js";

import { loadAchievementCatalog, type AchievementCatalog, type AchievementRegistry } from "./achievements/catalog";
import { activePlayBuffId, activePlayLuckyRequested, activePlayOpportunityId, loadActivePlayCatalog, selectActivePlayEffect, type ActivePlayCatalog, type ActivePlayEffect } from "./active-play";
import { achievementScore, newlyEarned, type AchievementObservation } from "./achievements/evaluate";
import { enclosureIndex, parseCommonsCatalog, type CommonsCatalog } from "./commons";
import { COPY_KEYS } from "./copy";
import { loadDoctrineCatalog, validateDoctrineRoutes, type DoctrineCatalog } from "./doctrines";
import { ladderSourceId, manualRoleSourceId, parseCatalog, subProgressValue, validateCatalogGateReferences, type EconomyCatalog, type MultiplierSlot, MULTIPLIER_SLOT_ORDER } from "./economy-kernel";
import { parseFactionCatalog, type FactionCatalog } from "./faction";
import { fiscalResolvedCost, harvestFiscal, loadFiscalCatalog, spendFiscal, sweepFiscal, type FiscalCatalog, type FiscalSpendTarget, type FiscalState, type FiscalSweep } from "./fiscal";
import { parseGuildCatalog, type GuildCatalog } from "./guild";
import { loadMeterCatalog, validateMeterResourceSeparation, type MeterCatalog } from "./meters/catalog";
import { advanceMeters, contributionKey as meterContributionKey, newRunMeterState, validateMeterState } from "./meters/transition";
import { canonicalString, isStateValue, MAX_EXACT_INTEGER, parseCanonical, quantize, sumDeterministic } from "./numeric";
import { parsePrestigePolicy, type PrestigePolicy } from "./prestige";
import { minigameCatalogSupportsSoul, parseMinigameCatalog, type MinigameCatalog } from "./minigame/catalog";
import { applyFounderMinigameResolution, type CertifiedMinigameResult, type MinigameRatingState } from "./minigame/resolution";
import { parsePetCatalog, petCatalogSupportsSoul, type PetCatalog } from "./pet/catalog";
import { parsePitchCatalog, type PitchCatalog } from "./pitch/catalog";
import { parsePetCareStates, validatePetCareStatesForCatalog, type PetCareState } from "./pet/state";
import { applyPetCareTransition, careStatus } from "./pet/transition";
import { discountedRequirements, evaluatePredicate, parseRoutesCatalog, type RouteContext, type RoutesCatalog } from "./routes";
import { parseSoulCatalog, soulBand, type SoulCatalog } from "./soul/catalog";
import { substream } from "./combat/rng";

const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;
const uuidV7 = /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;
const hashPattern = /^sha256:[0-9a-f]{64}$/;
const mechanical = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;

export interface ReplayArtifacts { readonly categories: string; readonly economy: string; readonly routes: string; readonly commons: string; readonly prestige: string; readonly factions: string; readonly guilds: string; readonly meters?: string; readonly achievements?: string; readonly doctrines?: string; readonly minigames?: string; readonly pets?: string; readonly fiscal?: string; readonly opportunities?: string; readonly soul?: string; readonly pitch?: string; readonly minigame_api?: string }
export interface MinigameAPIPolicy { readonly schemaVersion: 1; readonly tenants: readonly { readonly engineRef: string; readonly engineVersion: string; readonly minigameId: string }[] }
export interface ReplayCatalogBundle {
  readonly constantsHash: string; readonly artifacts: ReplayArtifacts; readonly economy: EconomyCatalog; readonly routes: RoutesCatalog;
  readonly commons: CommonsCatalog; readonly prestige: PrestigePolicy; readonly factions: FactionCatalog; readonly guilds: GuildCatalog;
  readonly meters?: MeterCatalog; readonly achievements?: AchievementCatalog; readonly doctrines?: DoctrineCatalog; readonly minigames?: MinigameCatalog; readonly pets?: PetCatalog; readonly fiscal?: FiscalCatalog; readonly opportunities?: ActivePlayCatalog; readonly soul?: SoulCatalog; readonly pitch?: PitchCatalog; readonly minigameAPI?: MinigameAPIPolicy;
  readonly next?: ReplayCatalogBundle;
}
export interface ReplayContribution { readonly slot: MultiplierSlot; readonly source_id: string; readonly target: string; readonly factor: string }
export interface ReplayEvent { readonly kind: string; readonly schema_version: number; readonly intent_id: string; readonly payload: unknown }
export interface ReplayChange { readonly resource_id: string; readonly before: string; readonly delta: string; readonly after: string }
export interface ReplayInvariant { readonly kind: "afford_fallback" | "residual_clamp" | "residual_abort"; readonly intent_id: string; readonly detail: string }
export interface ReplayPendingOpportunity { opportunityId: string; spawnedAttendedMs: number; expiresAttendedMs: number; effectRowId: string; selectedGeneratorId: string | null }
export interface ReplayActiveBuff { buffInstanceId: string; effectRowId: string; selectedTarget: string | null; activatedAttendedMs: number; expiresAttendedMs: number }
export interface ReplayState {
  wireVersion: 14 | 16 | 17 | 18;
  balances: Record<string, string>; generators: Record<string, number>; generatorPurchasedTotal: number;
  upgradesOwned: Set<string>; generatorsProvisioned: Record<string, number>; provisionRemaindersPpm: Record<string, number>; stockRateRemainderPpm: number;
  evaluatedThroughMs: number; computeCreditMs: number; computeBurstRemainingMs: number;
  opportunitySpawnSeq: number; nextOpportunityAttendedMs: number; pendingOpportunity: ReplayPendingOpportunity | null; activeBuffs: ReplayActiveBuff[];
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
	wireVersion: 14 | 15 | 16 | 17 | 18 | 19 | 20 | 21;
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
  fiscalCredit: number; fiscalPeriodOpenedWallMs: number; fiscalPeriodSequence: number;
  fiscalGeneratorLevels: Record<string, number>; fiscalUnlocks: Set<string>;
	soulExhaustedSourceIds: Set<string>;
	minigameSessionSeq: number;
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
export interface FounderReplayHead { readonly revision: number; readonly version: 14 | 15 | 16 | 17 | 18 | 19 | 20 | 21; readonly constantsHash: string; readonly state: unknown }
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
interface ActiveSpawnEvidence { sequence:number; sampled_interval_ms:number; effect_draw:string; generator_draw:string|null; effect_row_id:string; selected_generator_id:string|null; opportunity_id:string; spawned_attended_ms:number; expires_attended_ms:number }
interface ActiveClaimEvidence { opportunity_id:string; effect_row_id:string; selected_target:string|null; buff_instance_id:string|null; requested_delta:string|null; actual_credited_delta:string|null; saturated:boolean|null; cap_reason_key:string|null; next_sampled_interval_ms:number; next_opportunity_attended_ms:number }
interface ActiveScheduleEvidence { attended_now_ms:number; before_sequence:number; before_next_opportunity_attended_ms:number; after_sequence:number; after_next_opportunity_attended_ms:number; expired_buffs:{buff_instance_id:string}[]; missed_opportunity_id:string|null; spawned:ActiveSpawnEvidence|null; claim:ActiveClaimEvidence|null }
interface ReplayWire { v: 2 | 3 | 4 | 5; command: ReplayCommand; evaluated_at_ms: number; evaluation_mode: "online" | "offline"; resolved: Record<string, unknown> }
interface NetworkSlot { readonly slot: string; readonly carried_ref: string }
interface FounderCarry {
  founder_revision: number; founder_constants_hash: string; reputation_level: number; route_knowledge_balance: number;
  age_ms: number; notoriety: number; advisor_mode: boolean; network_slots: NetworkSlot[]; ledger_fact_kinds: string[]; exit_history_count: number;
  achievements_earned_lifetime: string[]; achievement_score_lifetime: number;
}
interface ExitTerms { reputation_delta: number; network_slot_unlocks: NetworkSlot[]; route_knowledge: number; clout_reach_note: string }

export async function loadReplayCatalogBundle(constantsHash: string, artifacts: ReplayArtifacts): Promise<ReplayCatalogBundle> {
  const names = Object.keys(artifacts).sort(byteCompare);
  const required = ["categories", "commons", "economy", "factions", "guilds", "prestige", "routes"];
  const allowed = new Set([...required, "achievements", "doctrines", "fiscal", "meters", "minigame_api", "minigames", "opportunities", "pets", "pitch", "soul"]);
  const foundations = artifacts.meters !== undefined || artifacts.achievements !== undefined;
  if (!hashPattern.test(constantsHash) || names.some((name) => !allowed.has(name)) || required.some((name) => !names.includes(name)) ||
      (artifacts.meters === undefined) !== (artifacts.achievements === undefined) || artifacts.doctrines !== undefined && !foundations ||
      artifacts.minigames !== undefined && !foundations || artifacts.pets !== undefined && artifacts.minigames === undefined ||
      artifacts.fiscal !== undefined && artifacts.pets === undefined || artifacts.soul !== undefined && artifacts.fiscal === undefined || artifacts.pitch !== undefined && artifacts.soul === undefined ||
      artifacts.minigame_api !== undefined && artifacts.pitch === undefined || artifacts.opportunities !== undefined && artifacts.doctrines === undefined) throw new SyntaxError("invalid replay artifact set");
  const computed = await constantsHashArtifacts(artifacts);
  if (computed !== constantsHash) throw new SyntaxError("replay artifact label mismatch");
  const economy = parseCatalog(parseJSON(artifacts.economy)); const routes = parseRoutesCatalog(parseJSON(artifacts.routes));
  const gateIds = routes.gates.map((gate) => gate.gateId);
  validateCategoryCatalog(parseJSON(artifacts.categories), gateIds);
  validateCatalogGateReferences(economy, gateIds);
  const commons = parseCommonsCatalog(parseJSON(artifacts.commons)); const prestige = parsePrestigePolicy(parseJSON(artifacts.prestige));
  const factions = parseFactionCatalog(parseJSON(artifacts.factions), commons.minimumTithePpm, commons.defaultTithePpm, commons.maximumTithePpm);
  const guilds = parseGuildCatalog(parseJSON(artifacts.guilds));
  if (!foundations) return Object.freeze({ constantsHash, artifacts: Object.freeze({ ...artifacts }), economy, routes, commons, prestige, factions, guilds });
  const meters = loadMeterCatalog(artifacts.meters!);
  validateMeterResourceSeparation(meters, economy.resources.map((value) => value.id));
  const achievements = loadAchievementCatalog(artifacts.achievements!, foundationAchievementRegistry(economy));
  const doctrines = artifacts.doctrines === undefined ? undefined : loadDoctrineCatalog(artifacts.doctrines);
  if (doctrines) validateDoctrineRoutes(doctrines, routes);
  const minigames = artifacts.minigames === undefined ? undefined : parseMinigameCatalog(parseJSON(artifacts.minigames));
  const pets = artifacts.pets === undefined ? undefined : parsePetCatalog(parseJSON(artifacts.pets));
  const fiscal = artifacts.fiscal === undefined ? undefined : loadFiscalCatalog(parseJSON(artifacts.fiscal), economy);
	const soul = artifacts.soul === undefined ? undefined : parseSoulCatalog(parseJSON(artifacts.soul), { copyKeys: new Set(COPY_KEYS), epochSeeded: true, catchupCeilingMs: prestige.catchupCeilingMs });
	if (soul && (!minigames || !minigameCatalogSupportsSoul(minigames) || !pets || !petCatalogSupportsSoul(pets))) throw new SyntaxError("Soul requires bumped minigame and pet artifacts");
  const pitch = artifacts.pitch === undefined ? undefined : parsePitchCatalog(parseJSON(artifacts.pitch), new Set(COPY_KEYS));
  const minigameAPI = artifacts.minigame_api === undefined ? undefined : parseMinigameAPIPolicy(parseJSON(artifacts.minigame_api));
  if (minigameAPI && (!pitch || !minigames || !minigameAPI.tenants.some((row) => row.minigameId === "pitch" && row.engineRef === "pitch" && row.engineVersion === "1.0.0"))) throw new SyntaxError("minigame API requires the Pitch content chain");
  const opportunities = artifacts.opportunities === undefined ? undefined : loadActivePlayCatalog(parseJSON(artifacts.opportunities), economy);
  if (opportunities && opportunities.schedule.minimumIntervalMs + opportunities.schedule.lifetimeMs <= prestige.catchupCeilingMs) throw new SyntaxError("opportunity schedule exceeds one pending transition per online horizon");
	return Object.freeze({ constantsHash, artifacts: Object.freeze({ ...artifacts }), economy, routes, commons, prestige, factions, guilds, meters, achievements, doctrines, minigames, pets, fiscal, opportunities, soul, pitch, minigameAPI });
}

function parseMinigameAPIPolicy(source: unknown): MinigameAPIPolicy {
  const root = exactObject(source, ["operations", "schema_version", "tenants"], "minigame API policy");
  if (root.schema_version !== 1) throw new SyntaxError("invalid minigame API schema version");
  const required = ["create_minigame_session", "get_current_minigame_session", "play_minigame_command", "resolve_minigame_session"];
  const operations = array(root.operations, "minigame API operations").map((item) => {
    const row = exactObject(item, ["operation_id", "version"], "minigame API operation");
    return { id: mechanicalString(row.operation_id), version: safeInteger(row.version, 1, 1) };
  });
  if (!same(operations.map((row) => row.id), required) || operations.some((row) => row.version !== 1)) throw new SyntaxError("invalid minigame API operations");
  let prior = "";
  const tenants = array(root.tenants, "minigame API tenants").map((item) => {
    const row = exactObject(item, ["engine_ref", "engine_version", "minigame_id"], "minigame API tenant");
    const tenant = { engineRef: mechanicalString(row.engine_ref), engineVersion: string(row.engine_version), minigameId: mechanicalString(row.minigame_id) };
    const key = `${tenant.engineRef}\0${tenant.engineVersion}\0${tenant.minigameId}`;
    if (!/^[1-9][0-9]*\.[0-9]+\.[0-9]+$/.test(tenant.engineVersion) || byteCompare(prior, key) >= 0) throw new SyntaxError("invalid minigame API tenant order");
    prior = key;
    return tenant;
  });
  if (tenants.length === 0) throw new SyntaxError("empty minigame API tenant set");
  return Object.freeze({ schemaVersion: 1, tenants: Object.freeze(tenants) });
}

const REPLAY_EVENT_KINDS = Object.freeze([
	"achievement_earned.v1",
  "compact_cascade_started", "compact_health_band_changed", "compact_left", "compact_recovered", "compact_recruitment_offered", "compact_sampled", "compact_signed", "compact_tithe_raised", "compensation",
	"compute_credit_spent", "doctrine_picked",
  "exit_offer_declined", "exit_offer_expired", "exit_offer_spawned", "faction_stock_saturated", "founder_advanced", "gate_crossed", "generator_purchased", "guild_activity_evaluated", "guild_tithe_accrued",
	"incorporated", "invariant_reported", "meter_band_changed.v1", "route_executed", "route_hint_purchased", "route_knowledge_granted", "run_ended", "run_started", "upgrade_purchased",
	"pet_care_applied.v1", "pet_status_changed.v1",
	"minigame_rating_changed.v1", "minigame_resolved.v1",
	"soul_price_paid.v1", "soul_band_changed.v1", "soul_depleted.v1",
	"soul_recovery_started.v1", "soul_recovery_cancelled.v1", "soul_recovered.v1",
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
const companyDoctrineSaveKeys = ["compute_burst_remaining_ms"] as const;
const companyActivePlaySaveKeys = ["opportunity_spawn_seq", "next_opportunity_attended_ms", "pending_opportunity", "active_buffs"] as const;
const founderMinigameSaveKeys = ["minigame_ratings", "minigame_offline_quality"] as const;
const founderPetSaveKeys = ["pets"] as const;
const founderFiscalSaveKeys = ["fiscal_credit", "fiscal_period_opened_wall_ms", "fiscal_period_seq", "fiscal_generator_levels", "fiscal_unlocks"] as const;
const founderSoulSaveKeys = ["soul_exhausted_source_ids"] as const;
const founderMinigameAPISaveKeys = ["minigame_session_seq"] as const;

export function restoreReplayState(source: unknown, version: number, catalog: EconomyCatalog, foundationCatalogs?: { readonly meters: MeterCatalog; readonly achievements: AchievementCatalog; readonly doctrines?: DoctrineCatalog; readonly opportunities?: ActivePlayCatalog }): ReplayState {
  const requestedVersion = version;
  let foundationRaw: Record<string, unknown> | null = null;
  if (version === 16 || version === 17 || version === 18) {
    const activeKeys = [...saveV14Keys.filter((key) => key !== "meter_bands"), ...foundationSaveKeys, ...(version >= 17 ? companyDoctrineSaveKeys : []), ...(version >= 18 ? companyActivePlaySaveKeys : [])];
    foundationRaw = exactObject(source, activeKeys, `save v${version}`);
    source = { ...foundationRaw, meter_bands: {} };
    for (const key of [...foundationSaveKeys, ...companyDoctrineSaveKeys, ...companyActivePlaySaveKeys]) delete (source as Record<string, unknown>)[key];
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
  const state: ReplayState = { wireVersion: requestedVersion === 18 ? 18 : requestedVersion === 17 ? 17 : requestedVersion === 16 ? 16 : 14, balances, generators, generatorPurchasedTotal: safeInteger(raw.generators_purchased_total, 0, MAX_EXACT_INTEGER), upgradesOwned, generatorsProvisioned, provisionRemaindersPpm, stockRateRemainderPpm: safeInteger(raw.stock_rate_remainder_ppm, 0, 999_999), evaluatedThroughMs, computeCreditMs: safeInteger(raw.compute_credit_ms, 0, MAX_EXACT_INTEGER), computeBurstRemainingMs: 0, opportunitySpawnSeq: 0, nextOpportunityAttendedMs: 0, pendingOpportunity: null, activeBuffs: [], manualTokenMilli: safeInteger(raw.manual_token_milli, 0, MAX_EXACT_INTEGER), manualTokenRefilledAtMs,
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
  if (requestedVersion === 16 || requestedVersion === 17 || requestedVersion === 18) {
    if (!foundationCatalogs || !foundationRaw) throw new SyntaxError("save v16 requires pinned foundation catalogs");
    state.meterValues = integerRecord(foundationRaw.meter_values, 0, 100, "meter_values");
    state.meterDecayRemainders = integerRecord(foundationRaw.meter_decay_remainders, 0, 3_599_999, "meter_decay_remainders");
    state.meterInputRemainders = plainIntegerRecord(foundationRaw.meter_input_remainders, 0, 3_599_999, "meter_input_remainders");
    validateMeterState(foundationCatalogs.meters, { values: state.meterValues, decayRemainders: state.meterDecayRemainders, inputRemainders: state.meterInputRemainders });
    state.achievementsEarnedRun = new Set(sortedUniqueMechanical(array(foundationRaw.achievements_earned_run, "run achievements")));
    state.achievementScoreRun = safeInteger(foundationRaw.achievement_score_run, 0, MAX_EXACT_INTEGER);
    if (array(foundationRaw.achievements_earned_lifetime, "company lifetime achievements").length !== 0 || safeInteger(foundationRaw.achievement_score_lifetime, 0, MAX_EXACT_INTEGER) !== 0 || achievementScore(foundationCatalogs.achievements, state.achievementsEarnedRun) !== state.achievementScoreRun) throw new SyntaxError("invalid company achievement state");
    if (requestedVersion >= 17) {
      if (!foundationCatalogs.doctrines) throw new SyntaxError("save v17 requires a pinned doctrines catalog");
      state.computeBurstRemainingMs = safeInteger(foundationRaw.compute_burst_remaining_ms, 0, catalog.offlinePolicy?.burstMaxDurationMs ?? 0);
    }
    if (requestedVersion >= 18) {
      if (!foundationCatalogs.opportunities) throw new SyntaxError("save v18 requires a pinned opportunities catalog");
      state.opportunitySpawnSeq = safeInteger(foundationRaw.opportunity_spawn_seq, 0, MAX_EXACT_INTEGER);
      state.nextOpportunityAttendedMs = safeInteger(foundationRaw.next_opportunity_attended_ms, 0, MAX_EXACT_INTEGER);
      state.pendingOpportunity = parseReplayPendingOpportunity(foundationRaw.pending_opportunity, foundationCatalogs.opportunities);
      state.activeBuffs = parseReplayActiveBuffs(foundationRaw.active_buffs, foundationCatalogs.opportunities);
    }
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
  if (state.wireVersion === 14) {
    if (state.computeBurstRemainingMs !== 0) throw new RangeError("inactive compute burst state");
    return base;
  }
  delete base.meter_bands;
  const active = { ...base, meter_values: sortedRecord(state.meterValues), meter_decay_remainders: sortedRecord(state.meterDecayRemainders), meter_input_remainders: sortedRecord(state.meterInputRemainders), achievements_earned_run: [...state.achievementsEarnedRun].sort(byteCompare), achievement_score_run: state.achievementScoreRun, achievements_earned_lifetime: [], achievement_score_lifetime: 0 };
  if (state.wireVersion === 16) {
    if (state.computeBurstRemainingMs !== 0) throw new RangeError("inactive compute burst state");
    return active;
  }
  const doctrinal = { ...active, compute_burst_remaining_ms: state.computeBurstRemainingMs };
  if (state.wireVersion === 17) {
    if (state.opportunitySpawnSeq !== 0 || state.nextOpportunityAttendedMs !== 0 || state.pendingOpportunity !== null || state.activeBuffs.length !== 0) throw new RangeError("inactive active-play state");
    return doctrinal;
  }
  return { ...doctrinal, opportunity_spawn_seq: state.opportunitySpawnSeq, next_opportunity_attended_ms: state.nextOpportunityAttendedMs,
    pending_opportunity: state.pendingOpportunity === null ? null : { opportunity_id: state.pendingOpportunity.opportunityId, spawned_attended_ms: state.pendingOpportunity.spawnedAttendedMs,
      expires_attended_ms: state.pendingOpportunity.expiresAttendedMs, effect_row_id: state.pendingOpportunity.effectRowId, selected_generator_id: state.pendingOpportunity.selectedGeneratorId },
    active_buffs: state.activeBuffs.map((value) => ({ buff_instance_id: value.buffInstanceId, effect_row_id: value.effectRowId, selected_target: value.selectedTarget,
      activated_attended_ms: value.activatedAttendedMs, expires_attended_ms: value.expiresAttendedMs })) };
}

function activePlayEffect(catalog: ActivePlayCatalog, id: string): ActivePlayCatalog["effects"][number] {
  const value = catalog.effects.find((row) => row.effectRowId === id);
  if (!value) throw new SyntaxError(`unknown active-play effect ${id}`);
  return value;
}

function parseReplayPendingOpportunity(source: unknown, catalog: ActivePlayCatalog): ReplayPendingOpportunity | null {
  if (source === null) return null;
  const raw = exactObject(source, ["opportunity_id", "spawned_attended_ms", "expires_attended_ms", "effect_row_id", "selected_generator_id"], "pending opportunity");
  const opportunityId = uuidV7String(raw.opportunity_id); const effectRowId = mechanicalString(raw.effect_row_id); const effect = activePlayEffect(catalog, effectRowId);
  const spawnedAttendedMs = safeInteger(raw.spawned_attended_ms, 0, MAX_EXACT_INTEGER); const expiresAttendedMs = safeInteger(raw.expires_attended_ms, 1, MAX_EXACT_INTEGER);
  const selectedGeneratorId = raw.selected_generator_id === null ? null : mechanicalString(raw.selected_generator_id);
  if (expiresAttendedMs <= spawnedAttendedMs || (effect.kind === "building_special") !== (selectedGeneratorId !== null) ||
      selectedGeneratorId !== null && (effect.kind !== "building_special" || !effect.eligibleGeneratorIds.includes(selectedGeneratorId))) throw new SyntaxError("invalid pending opportunity");
  return { opportunityId, spawnedAttendedMs, expiresAttendedMs, effectRowId, selectedGeneratorId };
}

function parseReplayActiveBuffs(source: unknown, catalog: ActivePlayCatalog): ReplayActiveBuff[] {
  let previous = "";
  return array(source, "active buffs").map((item) => {
    const raw = exactObject(item, ["buff_instance_id", "effect_row_id", "selected_target", "activated_attended_ms", "expires_attended_ms"], "active buff");
    const buffInstanceId = uuidV7String(raw.buff_instance_id); const effectRowId = mechanicalString(raw.effect_row_id); const effect = activePlayEffect(catalog, effectRowId);
    if (byteCompare(previous, buffInstanceId) >= 0) throw new SyntaxError("active buffs must be byte-sorted"); previous = buffInstanceId;
    const activatedAttendedMs = safeInteger(raw.activated_attended_ms, 0, MAX_EXACT_INTEGER); const expiresAttendedMs = safeInteger(raw.expires_attended_ms, 1, MAX_EXACT_INTEGER);
    const selectedTarget = raw.selected_target === null ? null : mechanicalString(raw.selected_target);
    const expectedTarget = effect.kind === "building_special";
    if (expiresAttendedMs <= activatedAttendedMs || expectedTarget !== (selectedTarget !== null) || selectedTarget !== null && (!expectedTarget || !effect.eligibleGeneratorIds.includes(selectedTarget))) throw new SyntaxError("invalid active buff");
    if (effect.kind === "lucky_payout") throw new SyntaxError("lucky payout cannot persist as a buff");
    return { buffInstanceId, effectRowId, selectedTarget, activatedAttendedMs, expiresAttendedMs };
  });
}

export function restoreFounderReplayState(source: unknown, version: number, catalogs: ReplayCatalogBundle): FounderReplayState {
  const requestedVersion = version;
  let foundationRaw: Record<string, unknown> | null = null;
	if (version >= 15 && version <= 21) {
		const activeKeys = [...saveV14Keys.filter((key) => key !== "meter_bands"), ...foundationSaveKeys.slice(0, version === 15 ? 3 : foundationSaveKeys.length), ...(version >= 17 ? founderMinigameSaveKeys : []), ...(version >= 18 ? founderPetSaveKeys : []), ...(version >= 19 ? founderFiscalSaveKeys : []), ...(version >= 20 ? founderSoulSaveKeys : []), ...(version >= 21 ? founderMinigameAPISaveKeys : [])];
    foundationRaw = exactObject(source, activeKeys, "Founder save v16");
    source = { ...foundationRaw, meter_bands: {} };
		for (const key of [...foundationSaveKeys, ...founderMinigameSaveKeys, ...founderPetSaveKeys, ...founderFiscalSaveKeys, ...founderSoulSaveKeys, ...founderMinigameAPISaveKeys]) delete (source as Record<string, unknown>)[key];
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
	let fiscalCredit = 0, fiscalPeriodOpenedWallMs = 0, fiscalPeriodSequence = 0, fiscalGeneratorLevels: Record<string, number> = {}, fiscalUnlocks = new Set<string>();
	let soulExhaustedSourceIds = new Set<string>();
	let minigameSessionSeq = 0;
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
  if (requestedVersion >= 19) {
    if (!catalogs.fiscal) throw new SyntaxError("Founder v19 requires fiscal artifact");
    fiscalCredit = safeInteger(foundationRaw!.fiscal_credit, 0, catalogs.fiscal.credit.hardcap);
    fiscalPeriodOpenedWallMs = safeInteger(foundationRaw!.fiscal_period_opened_wall_ms, 1, MAX_EXACT_INTEGER);
    fiscalPeriodSequence = safeInteger(foundationRaw!.fiscal_period_seq, 0, MAX_EXACT_INTEGER);
    const ids = catalogs.fiscal.generatorLevelRows.map((row) => row.generatorId);
    const rawLevels = exactRecord(foundationRaw!.fiscal_generator_levels, ids, "fiscal generator levels");
    fiscalGeneratorLevels = Object.fromEntries(catalogs.fiscal.generatorLevelRows.map((row) => [row.generatorId, safeInteger(rawLevels[row.generatorId], 0, row.levelHardcap)]));
    fiscalUnlocks = new Set(sortedUniqueMechanical(array(foundationRaw!.fiscal_unlocks, "fiscal unlocks")));
    for (const id of fiscalUnlocks) if (!catalogs.fiscal.unlockRows.some((row) => row.unlockId === id)) throw new SyntaxError("unknown fiscal unlock");
	} else if (catalogs.fiscal) throw new SyntaxError("fiscal artifact requires Founder v19");
	if (requestedVersion >= 20) {
		if (!catalogs.soul) throw new SyntaxError("Founder v20 requires Soul artifact");
		safeInteger(raw.soul, catalogs.soul.policy.soul_floor, catalogs.soul.policy.soul_max);
		soulExhaustedSourceIds = new Set(sortedUniqueMechanical(array(foundationRaw!.soul_exhausted_source_ids, "Soul exhausted source ids")));
		for (const id of soulExhaustedSourceIds) {
			const source = catalogs.soul.debit_sources.find((row) => row.source_id === id);
			if (!source?.may_exhaust) throw new SyntaxError("unknown exhausted Soul source");
		}
	} else {
		if (safeInteger(raw.soul, -MAX_EXACT_INTEGER, MAX_EXACT_INTEGER) !== 0) throw new SyntaxError("inactive Soul state before v20");
		if (catalogs.soul) throw new SyntaxError("Soul artifact requires Founder v20");
	}
	if (requestedVersion >= 21) {
		if (!catalogs.minigameAPI) throw new SyntaxError("Founder v21 requires minigame API artifact");
		minigameSessionSeq = safeInteger(foundationRaw!.minigame_session_seq, 0, MAX_EXACT_INTEGER);
	} else if (catalogs.minigameAPI) throw new SyntaxError("minigame API artifact requires Founder v21");
	return { wireVersion: requestedVersion as 14 | 15 | 16 | 17 | 18 | 19 | 20 | 21, balances, generators, generatorPurchasedTotal: safeInteger(raw.generators_purchased_total, 0, MAX_EXACT_INTEGER), upgradesOwned, generatorsProvisioned, provisionRemaindersPpm,
    stockRateRemainderPpm: 0, evaluatedThroughMs, computeCreditMs: 0, manualTokenMilli: 0, manualTokenRefilledAtMs,
    routeKnowledgeBalance: safeInteger(raw.route_knowledge_balance, 0, MAX_EXACT_INTEGER), hintsUnlocked: mechanicalSet(raw.hints_unlocked), ledgerFactKinds: mechanicalSet(raw.ledger_fact_kinds),
		reputationLevel: safeInteger(raw.reputation_level, 0, MAX_EXACT_INTEGER), networkSlots, cloutLifetime: safeInteger(raw.clout_lifetime, 0, MAX_EXACT_INTEGER), soul: requestedVersion >= 20 ? safeInteger(raw.soul, catalogs.soul!.policy.soul_floor, catalogs.soul!.policy.soul_max) : 0,
    ageMs: safeInteger(raw.age_ms, 0, MAX_EXACT_INTEGER), notoriety: safeInteger(raw.notoriety, 0, MAX_EXACT_INTEGER), advisorMode: boolean(raw.advisor_mode), exitHistory,
    achievementsEarnedLifetime: earnedLifetime, achievementScoreLifetime: lifetimeScore, minigameRatings, minigameOfflineQuality, pets,
		fiscalCredit, fiscalPeriodOpenedWallMs, fiscalPeriodSequence, fiscalGeneratorLevels, fiscalUnlocks, soulExhaustedSourceIds, minigameSessionSeq };
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
	if (state.wireVersion >= 19) Object.assign(active, { fiscal_credit: state.fiscalCredit, fiscal_period_opened_wall_ms: state.fiscalPeriodOpenedWallMs,
		fiscal_period_seq: state.fiscalPeriodSequence, fiscal_generator_levels: sortedRecord(state.fiscalGeneratorLevels), fiscal_unlocks: [...state.fiscalUnlocks].sort(byteCompare) });
	if (state.wireVersion >= 20) Object.assign(active, { soul_exhausted_source_ids: [...state.soulExhaustedSourceIds].sort(byteCompare) });
	if (state.wireVersion >= 21) Object.assign(active, { minigame_session_seq: state.minigameSessionSeq });
  return active;
}

export async function applyFounderLogged(state: FounderReplayState, canonicalPayload: string, catalogs: ReplayCatalogBundle, replayInputs: unknown): Promise<FounderLoggedTransition> {
  if (!hashPattern.test(catalogs.constantsHash)) throw new SyntaxError("invalid Founder catalog bundle");
  const wire = parseFounderReplayWire(replayInputs);
  const before = restoreFounderReplayState(encodeFounderReplayState(state), state.wireVersion, catalogs);
  const rollback = (): void => { Object.assign(state, before); };
  let fiscalSweep: FiscalSweep | null = null;
  if (catalogs.fiscal) {
    if (state.wireVersion < 19) throw new RangeError("inactive Fiscal state");
    const fiscalState = founderFiscalState(state); fiscalSweep = sweepFiscal(catalogs.fiscal, fiscalState, wire.command.server_ts_ms); applyFounderFiscalState(state, fiscalState);
  }
  const finish = (transition: FounderLoggedTransition): FounderLoggedTransition => {
    if (transition.outcome !== "applied") { rollback(); return transition; }
    if (!fiscalSweep) return transition;
    const receipt = exactObject(transition.receipt, Object.keys(transition.receipt as Record<string, unknown>), "Founder receipt");
    const fiscal_sweep = fiscalSweepWire(fiscalSweep);
    return { ...transition, receipt: { ...receipt, fiscal_sweep }, events: [event("fiscal_period_harvested.v1", wire.command.intent_id, { source: "automatic", ...fiscal_sweep }), ...transition.events] };
  };
  try {
    const kind = string(wire.resolved.kind);
    if (kind === "resolve_minigame_session") return finish(await applyFounderMinigameLogged(state, canonicalPayload, catalogs, wire));
		if (kind === "soul_recovery") return finish(applyFounderSoulRecovery(state, canonicalPayload, catalogs, wire));
		if (kind === "start_minigame_session") return finish(await applyFounderStartMinigameSession(state, canonicalPayload, catalogs, wire));
    const request = parseIntent(canonicalPayload, wire.command.intent_id);
    if (kind === "invalid") {
      onlyKeys(wire.resolved, ["kind", "detail"], "invalid Founder inputs");
      if (request.invalid === undefined || string(wire.resolved.detail) !== request.invalid) throw new RangeError("invalid Founder arm mismatch");
      return finish(founderRejected(state, wire.command.intent_id, wire.command.revision, "invalid", request.invalid, catalogs.constantsHash));
    }
    if (kind === "buy_route_hint") {
      onlyKeys(wire.resolved, ["kind", "route_context_version", "route_knowledge_balance"], "route hint inputs");
      if (request.kind !== "buy_route_hint" || request.invalid !== undefined || request.expected_revision !== wire.command.revision || safeInteger(wire.resolved.route_context_version, 1, MAX_EXACT_INTEGER) !== catalogs.routes.contextVersion) throw new RangeError("route hint command mismatch");
      state.routeKnowledgeBalance = safeInteger(wire.resolved.route_knowledge_balance, 0, MAX_EXACT_INTEGER);
      const routeId = string(request.route_id); const route = catalogs.routes.route(routeId);
      if (!route) return finish(founderRejected(state, request.intent_id, wire.command.revision, "unknown_id", routeId, catalogs.constantsHash, rollback));
      if (state.hintsUnlocked.has(routeId)) return finish(founderRejected(state, request.intent_id, wire.command.revision, "already_unlocked", routeId, catalogs.constantsHash, rollback));
      if (state.routeKnowledgeBalance < catalogs.routes.knowledge.hintCost) return finish(founderRejected(state, request.intent_id, wire.command.revision, "insufficient_route_knowledge", routeId, catalogs.constantsHash, rollback));
      state.routeKnowledgeBalance -= catalogs.routes.knowledge.hintCost; state.hintsUnlocked.add(routeId);
      const eventValue = event("route_hint_purchased", request.intent_id, { route_id: routeId, cost: catalogs.routes.knowledge.hintCost });
      const receipt = { applied_count: 1, evaluated_at: rfc3339(state.evaluatedThroughMs), intent_id: request.intent_id, new_revision: wire.command.revision + 1, outcome: "applied", receipt: { changes: [] }, snapshot: founderWireSnapshot(state) };
      return finish({ state, outcome: "applied", receipt, events: [eventValue], resultConstantsHash: catalogs.constantsHash });
    }
    if (kind === "care_action") {
      exactKeys(wire.resolved, ["kind", "attendance", "pet_attended_before_ms"], "pet care inputs");
      if (request.kind !== "care_action" || request.invalid !== undefined || request.expected_revision !== wire.command.revision || !catalogs.pets) throw new RangeError("pet care command mismatch");
      const petId = uuidV7String(request.pet_id); const actionId = mechanicalString(request.action_id);
      const sample = parseFounderAttendanceSample(wire.resolved.attendance);
      if (sample.companyConstantsHash !== catalogs.constantsHash) throw new RangeError("pet care catalog context mismatch");
      const attended = validateFounderAttendanceSample(state, wire.command.revision, request.expected_revision, sample);
      const beforeCursor = safeInteger(wire.resolved.pet_attended_before_ms, 0, MAX_EXACT_INTEGER);
      const care = state.pets[petId];
      if (care === undefined) {
        if (beforeCursor !== 0) throw new RangeError("unknown pet has a care cursor");
        return finish(founderRejected(state, request.intent_id, wire.command.revision, "unknown_id", "unknown_pet", catalogs.constantsHash, rollback));
      }
      if (care.evaluated_through_attended_ms !== beforeCursor) throw new RangeError("stale pet care cursor");
			if (catalogs.soul) {
				const action = catalogs.pets.actions.find((row) => row.action_id === actionId);
				if (!action) throw new RangeError("unknown pet Soul gate");
				if (action.soul_gate === "ordinary" && soulBand(catalogs.soul, state.soul).human_content_locked)
					return finish(founderRejected(state, request.intent_id, wire.command.revision, "not_eligible", "human_content_locked", catalogs.constantsHash, rollback));
			}
      const priorBand = careStatus(care, catalogs.pets).status_band;
      const careResult = applyPetCareTransition(care, catalogs.pets, { action_id: actionId, attended_before_ms: beforeCursor, attended_after_ms: attended });
      if (!careResult.applied) {
        const category = careResult.rejection_detail === "unknown_action" ? "unknown_id" : "not_eligible";
        return finish(founderRejected(state, request.intent_id, wire.command.revision, category, careResult.rejection_detail, catalogs.constantsHash, rollback));
      }
      state.pets[petId] = careResult.state;
      const receipt = { intent_id: request.intent_id, outcome: "applied", founder_revision: wire.command.revision + 1,
        pet_id: petId, action_id: actionId, stat_id: careResult.stat_id, before_ppm: careResult.before_ppm,
        applied_ppm: careResult.applied_ppm, after_ppm: careResult.after_ppm, trust_before_ppm: careResult.trust_before_ppm,
        trust_after_ppm: careResult.trust_after_ppm, mood: careResult.mood, status_band: careResult.status_band,
        next_eligible_attended_ms: careResult.next_eligible_attended_ms };
      const events = [event("pet_care_applied.v1", request.intent_id, { pet_id: petId, action_id: actionId,
        stat_id: careResult.stat_id, before_ppm: careResult.before_ppm, applied_ppm: careResult.applied_ppm,
        after_ppm: careResult.after_ppm, trust_before_ppm: careResult.trust_before_ppm,
        trust_after_ppm: careResult.trust_after_ppm, mood: careResult.mood, status_band: careResult.status_band,
        next_eligible_attended_ms: careResult.next_eligible_attended_ms })];
      if (careResult.status_changed) events.push(event("pet_status_changed.v1", request.intent_id,
        { pet_id: petId, from_status_band: priorBand, to_status_band: careResult.status_band }));
      return finish({ state, outcome: "applied", receipt, events, resultConstantsHash: catalogs.constantsHash });
    }
    if (kind === "harvest_fiscal_period") return finish(await applyFounderFiscalHarvest(state, request, wire, catalogs));
    if (kind === "spend_fiscal_credit") return finish(applyFounderFiscalSpend(state, request, wire, catalogs));
    if (kind === "exit.v1") return finish(applyFounderExit(state, request, wire, catalogs));
    throw new RangeError("unknown Founder replay arm");
  } catch (error) { rollback(); throw error; }
}

async function applyFounderStartMinigameSession(state: FounderReplayState, canonicalPayload: string,
	catalogs: ReplayCatalogBundle, wire: FounderReplayWire): Promise<FounderLoggedTransition> {
	if (state.wireVersion < 21 || !catalogs.minigameAPI || !catalogs.minigames) throw new RangeError("inactive minigame API state");
	const payload = exactObject(parseJSON(canonicalPayload), ["kind", "session_id", "minigame_id"], "minigame start payload");
	const sessionId = uuidV7String(payload.session_id), minigameId = mechanicalString(payload.minigame_id);
	if (payload.kind !== "start_minigame_session" || canonicalJSONString(payload) !== canonicalPayload) throw new RangeError("minigame start command mismatch");
	const resolved = exactObject(wire.resolved, ["kind", "company_stream_id", "run_seq", "sequence_before", "sequence_after", "seed"], "minigame start inputs");
	const companyStreamId = uuidString(resolved.company_stream_id), runSeq = safeInteger(resolved.run_seq, 1, MAX_EXACT_INTEGER);
	const before = safeInteger(resolved.sequence_before, 0, MAX_EXACT_INTEGER - 1), after = safeInteger(resolved.sequence_after, 1, MAX_EXACT_INTEGER);
	if (resolved.kind !== "start_minigame_session" || after !== before + 1 || state.minigameSessionSeq !== before) throw new RangeError("minigame start sequence mismatch");
	const definition = catalogs.minigames.minigames.find((row) => row.minigame_id === minigameId);
	const tenant = catalogs.minigameAPI.tenants.find((row) => row.minigameId === minigameId);
	if (!definition || !tenant || tenant.engineRef !== definition.engine_ref || tenant.engineVersion !== definition.engine_version) throw new RangeError("minigame start tenant mismatch");
	const seed = substream((await founderSeed(wire.command.founder_id, runSeq)) ^ BigInt(after), "minigame.session.v1").next().toString();
	if (string(resolved.seed) !== seed) throw new RangeError("minigame start seed mismatch");
	state.minigameSessionSeq = after;
	const receipt = { founder_revision: wire.command.revision + 1, intent_id: wire.command.intent_id, minigame_id: minigameId,
		outcome: "applied", seed, sequence_after: after, sequence_before: before, session_id: sessionId };
	void companyStreamId;
	return { state, outcome: "applied", receipt, events: [], resultConstantsHash: catalogs.constantsHash };
}

function applyFounderSoulRecovery(state: FounderReplayState, canonicalPayload: string, catalogs: ReplayCatalogBundle, wire: FounderReplayWire): FounderLoggedTransition {
	if (!catalogs.soul || state.wireVersion < 20) throw new RangeError("inactive Soul state");
	const payload = exactObject(parseJSON(canonicalPayload), ["kind", "session_id"], "Soul recovery payload");
	const action = string(payload.kind);
	const sessionId = uuidV7String(payload.session_id);
	if ((action !== "resolve_soul_recovery" && action !== "cancel_soul_recovery") || sessionId !== wire.command.intent_id || canonicalJSONString(payload) !== canonicalPayload) throw new RangeError("Soul recovery command mismatch");
	const resolved = exactObject(wire.resolved, ["kind", "action", "session_id", "activity_id", "company_stream_id", "run_seq", "founder_attended_start_ms", "founder_attended_end_ms", "recovery_amount", "soul_before", "soul_after", "band_before", "band_after", "reason_key"], "Founder Soul recovery");
	if (resolved.kind !== "soul_recovery" || resolved.action !== action || uuidV7String(resolved.session_id) !== sessionId) throw new RangeError("Soul recovery resolved mismatch");
	const activityId = mechanicalString(resolved.activity_id), companyStreamId = uuidString(resolved.company_stream_id);
	const runSeq = safeInteger(resolved.run_seq, 1, MAX_EXACT_INTEGER), start = safeInteger(resolved.founder_attended_start_ms, 0, MAX_EXACT_INTEGER), end = safeInteger(resolved.founder_attended_end_ms, start, MAX_EXACT_INTEGER);
	const activity = catalogs.soul.recovery_activities.find((row) => row.activity_id === activityId);
	if (!activity || resolved.reason_key !== activity.reason_key || resolved.recovery_amount !== activity.recovery_amount) throw new RangeError("Soul recovery activity mismatch");
	const soulBefore = safeInteger(resolved.soul_before, catalogs.soul.policy.soul_floor, catalogs.soul.policy.soul_max), beforeBand = soulBand(catalogs.soul, soulBefore);
	if (state.soul !== soulBefore || resolved.band_before !== beforeBand.band_member) throw new RangeError("Soul recovery before mismatch");
	let soulAfter = soulBefore;
	if (action === "resolve_soul_recovery") {
		if (end - start < activity.duration_attended_ms) throw new RangeError("Soul recovery duration");
		soulAfter = Math.min(catalogs.soul.policy.soul_max, soulBefore + activity.recovery_amount);
	}
	const afterBand = soulBand(catalogs.soul, soulAfter);
	if (safeInteger(resolved.soul_after, catalogs.soul.policy.soul_floor, catalogs.soul.policy.soul_max) !== soulAfter || resolved.band_after !== afterBand.band_member) throw new RangeError("Soul recovery after mismatch");
	state.soul = soulAfter;
	const receipt = { intent_id: sessionId, outcome: "applied", founder_revision: wire.command.revision + 1, session_id: sessionId, activity_id: activityId,
		action, soul_before: soulBefore, soul_after: soulAfter, band_before: beforeBand.band_member, band_after: afterBand.band_member };
	const payloadBase: Record<string, unknown> = { session_id: sessionId, activity_id: activityId, company_stream_id: companyStreamId, run_seq: runSeq,
		founder_attended_start_ms: start, founder_attended_end_ms: end, soul_before: soulBefore, soul_after: soulAfter };
	const events: ReplayEvent[] = action === "resolve_soul_recovery"
		? [event("soul_recovered.v1", sessionId, { ...payloadBase, recovery_amount: activity.recovery_amount, band_before: beforeBand.band_member, band_after: afterBand.band_member, reason_key: activity.reason_key })]
		: [event("soul_recovery_cancelled.v1", sessionId, payloadBase)];
	if (beforeBand.band_member !== afterBand.band_member) events.push(event("soul_band_changed.v1", sessionId, { soul_before: soulBefore, soul_after: soulAfter, band_before: beforeBand.band_member, band_after: afterBand.band_member, reason_key: afterBand.reason_key }));
	return { state, outcome: "applied", receipt, events, resultConstantsHash: catalogs.constantsHash };
}

function founderFiscalState(state: FounderReplayState): FiscalState { return { credit: state.fiscalCredit, periodOpenedWallMs: state.fiscalPeriodOpenedWallMs, periodSequence: state.fiscalPeriodSequence, generatorLevels: { ...state.fiscalGeneratorLevels }, unlocks: [...state.fiscalUnlocks].sort(byteCompare) }; }
function applyFounderFiscalState(state: FounderReplayState, value: FiscalState): void { state.fiscalCredit = value.credit; state.fiscalPeriodOpenedWallMs = value.periodOpenedWallMs; state.fiscalPeriodSequence = value.periodSequence; state.fiscalGeneratorLevels = { ...value.generatorLevels }; state.fiscalUnlocks = new Set(value.unlocks); }
function fiscalSweepWire(value: FiscalSweep): Record<string, unknown> { return { periods: value.periods, credit_before: value.creditBefore, credited: value.credited, credit_after: value.creditAfter, opened_before_ms: value.openedBeforeMs, opened_after_ms: value.openedAfterMs, seq_before: value.sequenceBefore, seq_after: value.sequenceAfter, saturated: value.saturated, hardcap_reason_key: value.hardcapReasonKey }; }
function parseFiscalTarget(source: unknown): FiscalSpendTarget { if (!isRecord(source)) throw new RangeError("Fiscal target must be an object"); const raw = source; const kind = string(raw.kind); if (kind === "generator_level") { exactKeys(raw, ["kind", "generator_id", "levels"], "Fiscal generator target"); return { kind, generatorId: mechanicalString(raw.generator_id), levels: safeInteger(raw.levels, 1, MAX_EXACT_INTEGER) }; } if (kind === "unlock") { exactKeys(raw, ["kind", "unlock_id"], "Fiscal unlock target"); return { kind, unlockId: mechanicalString(raw.unlock_id) }; } throw new RangeError("unknown Fiscal target"); }
function fiscalTargetWire(target: FiscalSpendTarget): Record<string, unknown> { return target.kind === "generator_level" ? { kind: target.kind, generator_id: target.generatorId, levels: target.levels } : { kind: target.kind, unlock_id: target.unlockId }; }

async function applyFounderFiscalHarvest(state: FounderReplayState, request: Intent, wire: FounderReplayWire, catalogs: ReplayCatalogBundle): Promise<FounderLoggedTransition> {
  if (request.kind !== "harvest_fiscal_period" || request.invalid !== undefined || request.expected_revision !== wire.command.revision || !catalogs.fiscal || state.wireVersion < 19) throw new RangeError("Fiscal harvest command mismatch");
  exactKeys(wire.resolved, ["kind", "now_wall_ms", "period_opened_wall_ms_before", "periods_swept", "seq_before", "draw_ppm", "outcome"], "Fiscal harvest inputs");
  const now = safeInteger(wire.resolved.now_wall_ms, 1, MAX_EXACT_INTEGER), opened = safeInteger(wire.resolved.period_opened_wall_ms_before, 1, MAX_EXACT_INTEGER), periods = safeInteger(wire.resolved.periods_swept, 0, MAX_EXACT_INTEGER), sequence = safeInteger(wire.resolved.seq_before, 0, MAX_EXACT_INTEGER);
  const draw = wire.resolved.draw_ppm === null ? null : safeInteger(wire.resolved.draw_ppm, 0, 999_999), outcome = string(wire.resolved.outcome);
  if (now !== wire.command.server_ts_ms) throw new RangeError("Fiscal harvest timestamp mismatch"); const fiscalState = founderFiscalState(state), creditBefore = fiscalState.credit; let saturated = false;
  if (outcome === "consumed_by_auto") { if (periods < 1 || draw !== null || fiscalState.periodSequence !== sequence + periods || fiscalState.periodOpenedWallMs !== opened + periods * catalogs.fiscal.clock.autoMs) throw new RangeError("consumed Fiscal harvest mismatch"); }
  else {
    let applied: Awaited<ReturnType<typeof harvestFiscal>>;
    try { applied = await harvestFiscal(catalogs.fiscal, fiscalState, wire.command.founder_id, now); }
    catch (error) { if (outcome === "rejected" && draw === null && periods === 0 && error instanceof RangeError && error.message === "fiscal period not ripe") return founderRejected(state, request.intent_id, wire.command.revision, "not_eligible", "period_not_ripe", catalogs.constantsHash); throw error; }
    if (applied.periodOpenedBeforeWallMs !== opened || applied.periodsSwept !== periods || applied.sequenceBefore !== sequence || applied.drawPpm !== draw || applied.outcome !== outcome) throw new RangeError("Fiscal harvest resolution mismatch"); saturated = applied.saturated; applyFounderFiscalState(state, fiscalState);
  }
  const receipt = { intent_id: request.intent_id, outcome: "applied", founder_revision: wire.command.revision + 1, fiscal_sweep: null, source: "manual", fiscal_credit_before: creditBefore, fiscal_credit_after: state.fiscalCredit, period_opened_wall_ms: state.fiscalPeriodOpenedWallMs, periods_swept: periods, seq_before: sequence, seq_after: state.fiscalPeriodSequence, draw_ppm: draw, harvest_outcome: outcome, saturated };
  const events = outcome === "consumed_by_auto" ? [] : [event("fiscal_period_harvested.v1", request.intent_id, { source: "manual", outcome, credit_before: creditBefore, credit_after: state.fiscalCredit, period_opened_wall_ms_before: opened, period_opened_wall_ms_after: state.fiscalPeriodOpenedWallMs, seq_before: sequence, seq_after: state.fiscalPeriodSequence, draw_ppm: draw, saturated })];
  return { state, outcome: "applied", receipt, events, resultConstantsHash: catalogs.constantsHash };
}

function applyFounderFiscalSpend(state: FounderReplayState, request: Intent, wire: FounderReplayWire, catalogs: ReplayCatalogBundle): FounderLoggedTransition {
  if (request.kind !== "spend_fiscal_credit" || request.invalid !== undefined || request.expected_revision !== wire.command.revision || !catalogs.fiscal || state.wireVersion < 19) throw new RangeError("Fiscal spend command mismatch");
  exactKeys(wire.resolved, ["kind", "target", "resolved_cost"], "Fiscal spend inputs"); const target = parseFiscalTarget(request.target), resolvedTarget = parseFiscalTarget(wire.resolved.target);
  if (canonicalJSONString(fiscalTargetWire(target)) !== canonicalJSONString(fiscalTargetWire(resolvedTarget))) throw new RangeError("Fiscal target mismatch");
  const fiscalState = founderFiscalState(state); let expectedCost = 0; try { expectedCost = fiscalResolvedCost(catalogs.fiscal, fiscalState, target); } catch { expectedCost = 0; }
  const resolvedCost = safeInteger(wire.resolved.resolved_cost, 0, MAX_EXACT_INTEGER); if (resolvedCost !== expectedCost) throw new RangeError("Fiscal cost mismatch"); const creditBefore = fiscalState.credit;
  let applied: ReturnType<typeof spendFiscal>;
  try { applied = spendFiscal(catalogs.fiscal, fiscalState, wire.command.server_ts_ms, target); }
  catch (error) { const message = error instanceof Error ? error.message : ""; const [category, detail] = message.includes("already unlocked") ? ["not_eligible", "already_unlocked"] : message.includes("insufficient") ? ["unaffordable", "fiscal_credit"] : message.includes("level purchase") ? ["cap_exceeded", target.kind === "generator_level" ? target.generatorId : target.unlockId] : ["unknown_id", target.kind === "generator_level" ? target.generatorId : target.unlockId]; return founderRejected(state, request.intent_id, wire.command.revision, category, detail, catalogs.constantsHash); }
  if (applied.resolvedCost !== resolvedCost) throw new RangeError("Fiscal spend result mismatch"); applyFounderFiscalState(state, fiscalState); const wireTarget = fiscalTargetWire(target);
  return { state, outcome: "applied", receipt: { intent_id: request.intent_id, outcome: "applied", founder_revision: wire.command.revision + 1, fiscal_sweep: null, target: wireTarget, resolved_cost: resolvedCost, fiscal_credit_before: creditBefore, fiscal_credit_after: state.fiscalCredit }, events: [event("fiscal_credit_spent.v1", request.intent_id, { target: wireTarget, resolved_cost: resolvedCost, fiscal_credit_before: creditBefore, fiscal_credit_after: state.fiscalCredit })], resultConstantsHash: catalogs.constantsHash };
}

export async function verifyFounderReplayHistory(genesis: unknown, genesisRevision: number, genesisVersion: 14 | 15 | 16 | 17 | 18 | 19 | 20 | 21, genesisHash: string,
  founderStreamId: string, founderId: string, entries: readonly FounderReplayLogEntry[], head: FounderReplayHead,
  bundles: readonly ReplayCatalogBundle[]): Promise<ReplayVerdict> {
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
      const exitArm = wire.resolved.kind === "exit.v1"; const minigameArm = wire.resolved.kind === "resolve_minigame_session"; const soulRecoveryArm = wire.resolved.kind === "soul_recovery";
      if ((exitArm || minigameArm || soulRecoveryArm) !== (entry.source !== null)) return "state_divergence";
      if (entry.source) {
        if (exitArm && (wire.resolved.company_stream_id !== entry.source.companyStreamId || wire.resolved.run_seq !== entry.source.runSeq || wire.resolved.run_log_seq !== entry.source.runLogSeq)) return "state_divergence";
        if (soulRecoveryArm && (wire.resolved.company_stream_id !== entry.source.companyStreamId || wire.resolved.run_seq !== entry.source.runSeq)) return "state_divergence";
      }
      let executionBundle = bundle;
      if (exitArm && wire.resolved.result_constants_hash !== hash) {
        const next = catalogs.get(string(wire.resolved.result_constants_hash)); if (!next) return "constants_mismatch";
        executionBundle = withNextReplayCatalogBundle(bundle, next);
      }
      const transition = await applyFounderLogged(state, entry.canonicalPayload, executionBundle, entry.replayInputs);
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

interface MinigameResolutionPayload { readonly kind: "resolve_minigame_session"; readonly session_id: string; readonly result: CertifiedMinigameResult }

async function parseMinigameResolutionPayload(canonicalPayload: string, intentId: string): Promise<MinigameResolutionPayload & { readonly resultHash: string }> {
  const parsed = JSON.parse(canonicalPayload) as unknown;
  if (canonicalJSONString(parsed) !== canonicalPayload) throw new SyntaxError("minigame payload is not canonical");
  const root = exactObject(parsed, ["kind", "session_id", "result"], "minigame resolution payload");
  if (root.kind !== "resolve_minigame_session" || uuidV7String(root.session_id) !== intentId) throw new RangeError("minigame command mismatch");
  const resultRaw = exactObject(root.result, ["outcome", "rating_delta", "score_facts"], "certified minigame result");
  const result: CertifiedMinigameResult = { outcome: mechanicalString(resultRaw.outcome), rating_delta: resultRaw.rating_delta === null ? null : safeInteger(resultRaw.rating_delta, -MAX_EXACT_INTEGER, MAX_EXACT_INTEGER),
    score_facts: array(resultRaw.score_facts, "certified score facts").map((item) => { const row = exactObject(item, ["kind", "value"], "certified score fact"); return { kind: mechanicalString(row.kind), value: safeInteger(row.value, -MAX_EXACT_INTEGER, MAX_EXACT_INTEGER) }; }) };
  const resultHash = await sha256Canonical(root.result);
  return { kind: "resolve_minigame_session", session_id: intentId, result, resultHash };
}

async function applyCompanyMinigameLogged(state: ReplayState, canonicalPayload: string, catalogs: ReplayCatalogBundle, wire: ReplayWire): Promise<LoggedTransition> {
  if (!catalogs.minigames) throw new RangeError("missing pinned minigame policy");
  const payload = await parseMinigameResolutionPayload(canonicalPayload, wire.command.intent_id);
  const resolved = exactObject(wire.resolved, ["kind", "session_id", "minigame_id", "certified_result_hash", "payout_policy", "selected_score", "faucet", "credited_delta", "founder_log", "company_revision", "founder_revision", "rating_change", "quality_change"], "company minigame resolution");
  if (resolved.kind !== "resolve_minigame_session" || resolved.session_id !== payload.session_id || resolved.certified_result_hash !== payload.resultHash) throw new RangeError("certified result mismatch");
  const minigameId = mechanicalString(resolved.minigame_id); const definition = catalogs.minigames.minigames.find((row) => row.minigame_id === minigameId);
  if (!definition || canonicalJSONString(definition.payout) !== canonicalJSONString(resolved.payout_policy)) throw new RangeError("pinned payout mismatch");
  const payout = exactObject(definition.payout, ["credited_resource_id", "sends_per_day", "per_send_cap", "conversion_ppm", "payout_score_fact_id", "cap_reason_key"], "payout policy");
  const selectedScore = uniqueCertifiedScore(payload.result, string(payout.payout_score_fact_id));
  if (safeInteger(resolved.selected_score, 0, MAX_EXACT_INTEGER) !== selectedScore) throw new RangeError("payout score mismatch");
  const faucet = parseMinigameFaucet(resolved.faucet);
  validateMinigameFaucet(faucet, definition.fallback, payout, selectedScore);
  const companyRevision = safeInteger(resolved.company_revision, 2, MAX_EXACT_INTEGER); const founderRevision = safeInteger(resolved.founder_revision, 2, MAX_EXACT_INTEGER);
  const founderLog = exactObject(resolved.founder_log, ["stream_id", "revision", "sequence"], "Founder log coordinate");
  if (companyRevision !== wire.command.revision + 1 || uuidString(founderLog.stream_id) === "" || safeInteger(founderLog.revision, 2, MAX_EXACT_INTEGER) !== founderRevision || safeInteger(founderLog.sequence, 1, MAX_EXACT_INTEGER) < 1) throw new RangeError("resolution coordinate mismatch");
  const ratingChange = parseRatingChange(resolved.rating_change); const qualityChange = parseQualityChange(resolved.quality_change);
  const creditedDelta = canonical(resolved.credited_delta); if (parseCanonical(creditedDelta).lt(0)) throw new RangeError("negative payout");
  const changes = applyLedger(state, catalogs.economy, [{ resource: string(payout.credited_resource_id), delta: parseCanonical(creditedDelta) }], true);
  if (changes.length !== 1 || changes[0]!.delta !== creditedDelta) throw new RangeError("payout ledger divergence");
  const receipt = { intent_id: payload.session_id, outcome: "applied", session_id: payload.session_id, minigame_id: minigameId,
    certified_result_hash: payload.resultHash, company_revision: companyRevision, founder_revision: founderRevision,
    credited_resource_id: string(payout.credited_resource_id), credited_delta: creditedDelta,
    configured_cap_forfeit_units: faucet.forfeited_units, cap_reason_key: faucet.cap_reason_key,
    rating_change: ratingChange, quality_change: qualityChange };
  const eventValue = event("minigame_resolved.v1", payload.session_id, { session_id: payload.session_id, minigame_id: minigameId,
    certified_result_hash: payload.resultHash, credited_resource_id: string(payout.credited_resource_id), credited_delta: creditedDelta,
    configured_cap_forfeit_units: faucet.forfeited_units, cap_reason_key: faucet.cap_reason_key, founder_revision: founderRevision });
  return { state, outcome: "applied", receipt, events: [eventValue], invariants: [] };
}

async function applyFounderMinigameLogged(state: FounderReplayState, canonicalPayload: string, catalogs: ReplayCatalogBundle, wire: FounderReplayWire): Promise<FounderLoggedTransition> {
  if (!catalogs.minigames || state.wireVersion < 17) throw new RangeError("missing Founder minigame state");
  const payload = await parseMinigameResolutionPayload(canonicalPayload, wire.command.intent_id);
  const resolved = exactObject(wire.resolved, ["kind", "session_id", "minigame_id", "certified_result_hash", "rating_before", "rating_after", "quality_before", "quality_after", "attendance"], "Founder minigame resolution");
  if (resolved.kind !== "resolve_minigame_session" || resolved.session_id !== payload.session_id || resolved.certified_result_hash !== payload.resultHash) throw new RangeError("Founder certified result mismatch");
  const minigameId = mechanicalString(resolved.minigame_id); const definition = catalogs.minigames.minigames.find((row) => row.minigame_id === minigameId);
  const beforeRating = parseMinigameRating(resolved.rating_before); const afterRating = parseMinigameRating(resolved.rating_after);
  const beforeQuality = parseMinigameQuality(resolved.quality_before); const afterQuality = parseMinigameQuality(resolved.quality_after);
  const currentRating = state.minigameRatings[minigameId]; const currentQuality = state.minigameOfflineQuality[minigameId];
  if (!definition || canonicalJSONString(currentRating) !== canonicalJSONString(beforeRating) || canonicalJSONString(currentQuality) !== canonicalJSONString(beforeQuality)) throw new RangeError("Founder minigame state mismatch");
  const attendance = parseFounderAttendanceSample(resolved.attendance);
  if (attendance.companyConstantsHash !== catalogs.constantsHash) throw new RangeError("Founder minigame catalog mismatch");
  const attended = validateFounderAttendanceSample(state, wire.command.revision, wire.command.revision, attendance);
  const transition = applyFounderMinigameResolution(beforeRating, beforeQuality, payload.result, definition.rating_policy, definition.offline_quality, attended);
  if (canonicalJSONString(transition.rating_after) !== canonicalJSONString(afterRating) || canonicalJSONString(transition.quality_after) !== canonicalJSONString(afterQuality)) throw new RangeError("Founder minigame arithmetic mismatch");
  state.minigameRatings[minigameId] = afterRating; state.minigameOfflineQuality[minigameId] = afterQuality;
  const ratingChange = { rated: payload.result.rating_delta !== null, old_elo: beforeRating.elo, new_elo: afterRating.elo,
    season_member: afterRating.season_member, games_before: beforeRating.games_counted, games_after: afterRating.games_counted };
  const qualityChange = { old: beforeQuality, new: afterQuality };
  const receipt = { intent_id: payload.session_id, outcome: "applied", founder_revision: wire.command.revision + 1,
    session_id: payload.session_id, certified_result_hash: payload.resultHash, rating_change: ratingChange, quality_change: qualityChange };
  const eventValue = event("minigame_rating_changed.v1", payload.session_id, { session_id: payload.session_id, minigame_id: minigameId,
    certified_result_hash: payload.resultHash, old_elo: beforeRating.elo, new_elo: afterRating.elo, season_member: afterRating.season_member,
    old_quality: beforeQuality, new_quality: afterQuality });
  return { state, outcome: "applied", receipt, events: [eventValue], resultConstantsHash: catalogs.constantsHash };
}

interface MinigameFaucetReplay { attended_day: number; quota_before: number; quota_after: number; remainder_before_ppm: number; remainder_after_ppm: number; reduced_score: number; converted_units: number; credited_units: number; forfeited_units: number; cap_reason_key: string }
function parseMinigameFaucet(source: unknown): MinigameFaucetReplay { const row = exactObject(source, ["attended_day", "quota_before", "quota_after", "remainder_before_ppm", "remainder_after_ppm", "reduced_score", "converted_units", "credited_units", "forfeited_units", "cap_reason_key"], "minigame faucet"); return {
  attended_day: safeInteger(row.attended_day, 0, MAX_EXACT_INTEGER), quota_before: safeInteger(row.quota_before, 0, MAX_EXACT_INTEGER), quota_after: safeInteger(row.quota_after, 0, MAX_EXACT_INTEGER),
  remainder_before_ppm: safeInteger(row.remainder_before_ppm, 0, 999_999), remainder_after_ppm: safeInteger(row.remainder_after_ppm, 0, 999_999),
  reduced_score: safeInteger(row.reduced_score, 0, MAX_EXACT_INTEGER), converted_units: safeInteger(row.converted_units, 0, MAX_EXACT_INTEGER),
  credited_units: safeInteger(row.credited_units, 0, MAX_EXACT_INTEGER), forfeited_units: safeInteger(row.forfeited_units, 0, MAX_EXACT_INTEGER), cap_reason_key: string(row.cap_reason_key) }; }
function validateMinigameFaucet(value: MinigameFaucetReplay, fallbackSource: unknown, payout: Record<string, unknown>, score: number): void {
  const fallback = fallbackSource as Record<string, unknown>; const reduction = fallback.kind === "solo" ? 0 : safeInteger(fallback.rate_reduction_ppm, 0, 1_000_000);
  const reduced = integratePPM(score, 1_000_000 - reduction, 0); const converted = integratePPM(reduced.whole, safeInteger(payout.conversion_ppm, 0, 1_000_000), value.remainder_before_ppm);
  const canCredit = value.quota_before < safeInteger(payout.sends_per_day, 0, MAX_EXACT_INTEGER); const expectedCredit = canCredit ? Math.min(converted.whole, safeInteger(payout.per_send_cap, 0, MAX_EXACT_INTEGER)) : 0;
  const expectedForfeit = converted.whole - expectedCredit; const expectedReason = expectedForfeit > 0 ? string(payout.cap_reason_key) : "";
  if (value.reduced_score !== reduced.whole || value.converted_units !== converted.whole || value.remainder_after_ppm !== converted.remainder || value.credited_units !== expectedCredit ||
      value.forfeited_units !== expectedForfeit || value.converted_units !== value.credited_units + value.forfeited_units || value.quota_after !== value.quota_before + (canCredit ? 1 : 0) || value.cap_reason_key !== expectedReason) throw new RangeError("faucet replay mismatch");
}
function integratePPM(value: number, ppm: number, remainder: number): { whole: number; remainder: number } { const total = BigInt(value) * BigInt(ppm) + BigInt(remainder); const whole = total / 1_000_000n; const rest = total % 1_000_000n; if (whole > BigInt(MAX_EXACT_INTEGER)) throw new RangeError("minigame conversion overflow"); return { whole: Number(whole), remainder: Number(rest) }; }
function uniqueCertifiedScore(result: CertifiedMinigameResult, kind: string): number { const rows = result.score_facts.filter((row) => row.kind === kind); if (rows.length !== 1 || rows[0]!.value < 0) throw new RangeError("invalid payout score"); return rows[0]!.value; }
function parseMinigameRating(source: unknown): MinigameRatingState { const row = exactObject(source, ["elo", "season_member", "games_counted"], "minigame rating"); return { elo: safeInteger(row.elo, -MAX_EXACT_INTEGER, MAX_EXACT_INTEGER), season_member: mechanicalString(row.season_member), games_counted: safeInteger(row.games_counted, 0, MAX_EXACT_INTEGER) }; }
function parseMinigameQuality(source: unknown): { grade_ppm: number; last_founder_attended_ms: number; decay_remainder_ppm: number } { const row = exactObject(source, ["grade_ppm", "last_founder_attended_ms", "decay_remainder_ppm"], "minigame quality"); return { grade_ppm: safeInteger(row.grade_ppm, 0, 1_000_000), last_founder_attended_ms: safeInteger(row.last_founder_attended_ms, 0, MAX_EXACT_INTEGER), decay_remainder_ppm: safeInteger(row.decay_remainder_ppm, 0, 999_999) }; }
function parseRatingChange(source: unknown): unknown { const row = exactObject(source, ["rated", "old_elo", "new_elo", "season_member", "games_before", "games_after"], "rating change"); return { rated: boolean(row.rated), old_elo: safeInteger(row.old_elo, -MAX_EXACT_INTEGER, MAX_EXACT_INTEGER), new_elo: safeInteger(row.new_elo, -MAX_EXACT_INTEGER, MAX_EXACT_INTEGER), season_member: mechanicalString(row.season_member), games_before: safeInteger(row.games_before, 0, MAX_EXACT_INTEGER), games_after: safeInteger(row.games_after, 0, MAX_EXACT_INTEGER) }; }
function parseQualityChange(source: unknown): unknown { const row = exactObject(source, ["old", "new"], "quality change"); return { old: parseMinigameQuality(row.old), new: parseMinigameQuality(row.new) }; }
async function sha256Canonical(value: unknown): Promise<string> { const digest = new Uint8Array(await crypto.subtle.digest("SHA-256", new TextEncoder().encode(canonicalJSONString(value)))); return `sha256:${[...digest].map((part) => part.toString(16).padStart(2, "0")).join("")}`; }

export async function applyLogged(state: ReplayState, canonicalPayload: string, catalogs: ReplayCatalogBundle, replayInputs: unknown): Promise<LoggedTransition> {
  const wire = parseReplayWire(replayInputs, state, catalogs);
	if (wire.resolved.kind === "soul_recovery_suppression") return applySuppressedLogged(state, canonicalPayload, catalogs, wire);
  if (wire.resolved.kind === "resolve_minigame_session") return applyCompanyMinigameLogged(state, canonicalPayload, catalogs, wire);
  const request = parseIntent(canonicalPayload, wire.command.intent_id);
  if (request.expected_revision !== wire.command.revision || wire.command.run_seq !== state.runSeq) throw new RangeError("command/state mismatch");
  if (request.kind === "buy_route_hint") throw new RangeError("founder-scope intent is not replayable");
  deriveFactionStockResource(state, catalogs.factions);
  const stateBefore = cloneReplayState(state, catalogs);
  const revision = wire.command.revision; const before = { ...state.balances };
  if (request.invalid !== undefined) return rejected(state, request.intent_id, revision, "invalid", request.invalid);
  if (wire.evaluated_at_ms < state.evaluatedThroughMs) throw new ReplayClockViolation();
  const resolved = wire.resolved; const kind = string(resolved.kind); if (string(resolved.intent_kind) !== request.kind) throw new RangeError("resolved intent mismatch");
  let accrual: ReplayAccrual; let founderCarry: FounderCarry | null = null; let declined = 0; const hasActive="active_play" in resolved; const activeEvidence=hasActive?parseActiveSchedule(resolved.active_play):null;
  if (kind === "cross_gate" && request.kind === "cross_gate") { onlyKeys(resolved, hasActive?["kind", "intent_kind", "accrual", "declined_exit_offer_count", "founder_carry","active_play"]:["kind", "intent_kind", "accrual", "declined_exit_offer_count", "founder_carry"], "cross gate inputs"); accrual = parseAccrual(resolved.accrual, catalogs); declined = safeInteger(resolved.declined_exit_offer_count ?? 0, 0, MAX_EXACT_INTEGER); founderCarry = resolved.founder_carry === undefined || resolved.founder_carry === null ? null : parseFounderCarry(resolved.founder_carry, catalogs, wire.v); }
  else { if (kind !== "accrual") throw new RangeError("resolved union mismatch"); const hasCarry = "founder_carry" in resolved; const keys=["kind","intent_kind","accrual",...(hasCarry?["founder_carry"]:[]),...(hasActive?["active_play"]:[])]; onlyKeys(resolved,keys,"accrual inputs"); accrual = parseAccrual(resolved.accrual, catalogs); founderCarry = !hasCarry || resolved.founder_carry === null ? null : parseFounderCarry(resolved.founder_carry, catalogs, wire.v); }
  if (foundationsActive(catalogs) && wire.v >= 4 && founderCarry === null) throw new RangeError("missing active Founder carry");
  if((state.wireVersion===18)!==(activeEvidence!==null)||state.wireVersion===18&&wire.v<5)throw new RangeError("active-play resolved presence");
  if (state.compactMember !== (accrual.commons_weight_ppm !== null)) throw new RangeError("commons weight presence mismatch");
  const rejectState = (category: string, detail: string): LoggedTransition => { restoreReplaySnapshot(state, stateBefore); return rejected(state, request.intent_id, revision, category, detail); };
  try {
  const activeEvents=activeEvidence===null?[]:await applyActiveSchedule(state,catalogs,wire.command,wire.evaluated_at_ms,activeEvidence);
  applyGuildSettlements(state, accrual.guild_settlement_batch, catalogs.factions.stockCap);
  const activeValues=activeEvidence===null?[]:activeContributions(state,catalogs.opportunities!,activeEvidence.attended_now_ms);
  const contributions = assembleContributions(state, catalogs.economy, [...accrual.contributions,...activeValues]);
  const effectiveAccrual = { ...accrual, contributions };
  const preflight = preflightRejection(state, catalogs, request, wire.evaluated_at_ms);
  if (preflight !== null) return rejectState(preflight[0], preflight[1]);
  const evaluation = evaluate(state, catalogs.economy, wire.evaluated_at_ms, wire.evaluation_mode, contributions);
  const events = [...activeEvents,...runHooks(state, catalogs, wire.command, evaluation, effectiveAccrual)];
  const postAccrualBalances = { ...state.balances };
  const invariants: ReplayInvariant[] = [];
  let appliedCount = 0;
  let opportunityResult:ActiveClaimEvidence|null=null;
  switch (request.kind) {
    case "buy_generator": { const result = buyGenerator(state, catalogs.economy, request, invariants); if (result.rejection) return rejectState(result.rejection[0], result.rejection[1]); appliedCount = result.count; events.push(event("generator_purchased", request.intent_id, { generator_id: request.generator_id, count: appliedCount, cost_resource_id: request.costResource, cost: request.cost })); break; }
    case "buy_upgrade": { const result = buyUpgrade(state, catalogs.economy, catalogs.routes, request); if (result.rejection) return rejectState(result.rejection[0], result.rejection[1]); events.push(event("upgrade_purchased", request.intent_id, { upgrade_id: request.upgrade_id, cost_resource_id: request.costResource, cost: request.cost })); appliedCount = 1; break; }
    case "perform_manual_batch": { const result = manualBatch(state, catalogs.economy, request, wire.evaluated_at_ms, contributions); if (result.rejection) return rejectState(result.rejection[0], result.rejection[1]); appliedCount = result.count; break; }
    case "cross_gate": { const result = crossGate(state, catalogs.routes, request, wire.command); if (result.rejection) return rejectState(result.rejection[0], result.rejection[1]); events.push(event("gate_crossed", request.intent_id, { founder_id: wire.command.founder_id, gate_id: request.gate_id, route_id: request.route_id, run_id: { company_stream_id: wire.command.company_stream_id, run_seq: state.runSeq } })); if (request.route_id !== null) events.push(event("route_executed", request.intent_id, { founder_id: wire.command.founder_id, gate_id: request.gate_id, route_id: request.route_id, run_id: { company_stream_id: wire.command.company_stream_id, run_seq: state.runSeq } })); appliedCount = 1; break; }
    case "pick_doctrine": state.doctrinesByTransition[request.transition_id] = request.doctrine_id; events.push(event("doctrine_picked", request.intent_id, { founder_id: wire.command.founder_id, run_id: { company_stream_id: wire.command.company_stream_id, run_seq: state.runSeq }, transition_id: request.transition_id, doctrine_id: request.doctrine_id })); appliedCount = 1; break;
    case "spend_compute_credit": { if (state.computeBurstRemainingMs !== 0) return rejectState("not_eligible", "burst_active"); if (state.computeCreditMs < request.amount_ms) return rejectState("not_eligible", "compute_credit_balance"); const speed = catalogs.economy.offlinePolicy!.burstSpeed; if (!parseCanonical(speed).gt(1)) throw new RangeError("invalid burst speed"); state.computeCreditMs -= request.amount_ms; state.computeBurstRemainingMs = request.amount_ms; events.push(event("compute_credit_spent", request.intent_id, { founder_id: wire.command.founder_id, run_id: { company_stream_id: wire.command.company_stream_id, run_seq: state.runSeq }, amount_ms: request.amount_ms, target: request.target, burst_duration_ms: request.amount_ms, burst_speed: speed })); appliedCount = 1; break; }
    case "claim_opportunity": { if(activeEvidence===null)throw new RangeError("missing active schedule evidence");const result=await claimOpportunity(state,catalogs,wire.command,request,activeEvidence,contributions);if(result.rejection)return rejectState(result.rejection[0],result.rejection[1]);events.push(...result.events);opportunityResult=result.claim!;appliedCount=1;break; }
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
  const result=applied(state, catalogs.economy, request.intent_id, revision + 1, appliedCount, before, events, invariants);if(opportunityResult!==null){const receipt=(result.receipt as {receipt:Record<string,unknown>});receipt.receipt.opportunity=opportunityResult;}return result;
  } catch (error) { restoreReplaySnapshot(state, stateBefore); throw error; }
}

async function applySuppressedLogged(state: ReplayState, canonicalPayload: string, catalogs: ReplayCatalogBundle, wire: ReplayWire): Promise<LoggedTransition> {
	const payload = exactObject(parseJSON(canonicalPayload), ["kind", "session_id"], "Soul suppression payload");
	const intentKind = string(payload.kind), sessionId = uuidV7String(payload.session_id);
	if ((intentKind !== "resolve_soul_recovery" && intentKind !== "cancel_soul_recovery") || sessionId !== wire.command.intent_id || canonicalJSONString(payload) !== canonicalPayload) throw new RangeError("Soul suppression command mismatch");
	const resolved = exactObject(wire.resolved, ["kind", "intent_kind", "suppression", "accrual", "active_play"], "Soul suppression inputs");
	if (resolved.kind !== "soul_recovery_suppression" || resolved.intent_kind !== intentKind) throw new RangeError("Soul suppression resolved mismatch");
	const suppression = exactObject(resolved.suppression, ["from_evaluated_ms", "to_evaluated_ms", "founder_attended_start_ms", "founder_attended_end_ms", "session_id"], "Soul suppression coordinates");
	const from = safeInteger(suppression.from_evaluated_ms, 1, MAX_EXACT_INTEGER), to = safeInteger(suppression.to_evaluated_ms, from, MAX_EXACT_INTEGER);
	const attendedStart = safeInteger(suppression.founder_attended_start_ms, 0, MAX_EXACT_INTEGER), attendedEnd = safeInteger(suppression.founder_attended_end_ms, attendedStart, MAX_EXACT_INTEGER);
	if (from !== state.evaluatedThroughMs || to !== wire.evaluated_at_ms || suppression.session_id !== sessionId || wire.evaluation_mode !== "online" || wire.command.run_seq !== state.runSeq) throw new RangeError("Soul suppression coordinate drift");
	const accrual = parseAccrual(resolved.accrual, catalogs);
	if (accrual.guild_settlement_batch.settlements.length !== 0 || state.compactMember !== (accrual.commons_weight_ppm !== null)) throw new RangeError("Soul suppression accrual mismatch");
	const before = cloneReplayState(state, catalogs);
	const activeEvidence = resolved.active_play === null ? null : parseActiveSchedule(resolved.active_play);
	if ((state.wireVersion === 18) !== (activeEvidence !== null)) throw new RangeError("Soul suppression active-play evidence");
	const events = activeEvidence === null ? [] : (await applyActiveSchedule(state, catalogs, wire.command, to, activeEvidence))
		.map((item) => ({ ...item, intent_id: sessionId }));
	const evaluation = evaluate(state, catalogs.economy, to, "online", accrual.contributions);
	runHooks(state, catalogs, wire.command, evaluation, accrual);
	state.balances = { ...before.balances }; state.generatorsProvisioned = { ...before.generatorsProvisioned }; state.provisionRemaindersPpm = { ...before.provisionRemaindersPpm };
	state.stockUnits = before.stockUnits; state.stockProgressMs = before.stockProgressMs; state.stockRateRemainderPpm = before.stockRateRemainderPpm; state.consumedStockUnits = before.consumedStockUnits;
	state.guildTitheCarryPpm = before.guildTitheCarryPpm; state.guildBoundaryGuildId = before.guildBoundaryGuildId; state.guildBoundarySeq = before.guildBoundarySeq; state.guildConsumedWindow = before.guildConsumedWindow;
	state.meterValues = { ...before.meterValues }; state.meterDecayRemainders = { ...before.meterDecayRemainders }; state.meterInputRemainders = { ...before.meterInputRemainders };
	state.achievementsEarnedRun = new Set(before.achievementsEarnedRun); state.achievementScoreRun = before.achievementScoreRun; state.lifetimeValue = before.lifetimeValue;
	if (to > state.manualTokenRefilledAtMs) { const elapsed = to - state.manualTokenRefilledAtMs, policy = catalogs.economy.manualPolicy!; state.manualTokenMilli = Math.min(policy.bucketCapMilli, state.manualTokenMilli + elapsed * policy.refillMilliPerMs); state.manualTokenRefilledAtMs = to; }
	return { state, outcome: "applied", receipt: { intent_id: sessionId, outcome: "applied", revision: wire.command.revision + 1, session_id: sessionId,
		from_evaluated_ms: from, to_evaluated_ms: to, suppressed_output: true }, events, invariants: [] };
}

export async function applyLoggedExit(company: ReplayState, canonicalPayload: string, catalogs: ReplayCatalogBundle, replayInputs: unknown): Promise<LoggedExitTransition> {
  const wire = parseReplayWire(replayInputs, company, catalogs);
  const request = parseIntent(canonicalPayload, wire.command.intent_id);
  if (request.expected_revision !== wire.command.revision || wire.command.run_seq !== company.runSeq) throw new RangeError("terminal command/state mismatch");
  const resolved = wire.resolved;
  const hasActive="active_play" in resolved,hasNextActive="next_active_play" in resolved,hasMinigameActivity="minigame_session_active" in resolved;onlyKeys(resolved, ["kind", "intent_kind", "accrual", "founder_carry", "executed_route_ids", "selected_exit_type", "selected_terms", "next_constants_hash",...(hasActive?["active_play"]:[]),...(hasNextActive?["next_active_play"]:[]),...(hasMinigameActivity?["minigame_session_active"]:[])], "terminal resolved inputs");
  if (resolved.kind !== "exit" || resolved.intent_kind !== request.kind || typeof resolved.selected_exit_type !== "string" || !hashPattern.test(string(resolved.next_constants_hash))) throw new RangeError("terminal resolved union mismatch");
  const selectedTerms = exactObject(resolved.selected_terms, Object.keys(resolved.selected_terms as object), "selected terms");
  const nextHash = string(resolved.next_constants_hash);
  const next = nextHash === catalogs.constantsHash ? catalogs : catalogs.next;
  if (!next || next.constantsHash !== nextHash) throw new RangeError("next catalog bundle mismatch");
  const activeEvidence=hasActive?parseActiveSchedule(resolved.active_play):null,nextActive=hasNextActive?parseActiveSpawn(resolved.next_active_play):null;if((company.wireVersion===18)!==(activeEvidence!==null)||company.wireVersion===18&&wire.v<5||(next.opportunities!==undefined)!==(nextActive!==null)||activeEvidence?.claim!==null&&activeEvidence!==null)throw new RangeError("terminal active-play evidence mismatch");
  if ((catalogs.minigameAPI !== undefined) !== hasMinigameActivity) throw new RangeError("terminal minigame activity evidence mismatch");
  const minigameSessionActive = hasMinigameActivity ? boolean(resolved.minigame_session_active) : false;
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
  if (minigameSessionActive) return rejectState("not_eligible", "minigame_session_active");
  try {
  const activeEvents=activeEvidence===null?[]:await applyActiveSchedule(company,catalogs,wire.command,wire.evaluated_at_ms,activeEvidence);
  applyGuildSettlements(company, accrual.guild_settlement_batch, catalogs.factions.stockCap);
  const activeValues=activeEvidence===null?[]:activeContributions(company,catalogs.opportunities!,activeEvidence.attended_now_ms);const contributions = assembleContributions(company, catalogs.economy, [...accrual.contributions,...activeValues]);
  const effectiveAccrual = { ...accrual, contributions };

  if (request.kind === "cross_gate") {
    const preflight = preflightRejection(company, catalogs, request, wire.evaluated_at_ms);
    if (preflight !== null) return rejectState(preflight[0], preflight[1]);
    const evaluation = evaluate(company, catalogs.economy, wire.evaluated_at_ms, wire.evaluation_mode, contributions);
    prefix = [...activeEvents,...runHooks(company, catalogs, wire.command, evaluation, effectiveAccrual)];
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
    prefix = [...activeEvents,...runHooks(company, catalogs, wire.command, evaluation, effectiveAccrual)];
    terms = computeExitTerms(company, founder, catalogs.prestige, exitType);
    if (promised) terms = promiseTerms(promised.payout_preview, applyTermsModifier(terms, promised.market_modifier_ppm));
  }
  if (resolved.selected_exit_type !== exitType) throw new RangeError("selected exit type mismatch");
  if (foundationsActive(catalogs) && wire.v >= 4) applyFoundationTransition(catalogs, companyBefore, company, founder, wire.command, request, wire.evaluated_at_ms, contributions, actionDebits, true, prefix);
  return await finishLoggedExit(company, founder, request.intent_id, wire.command, wire.evaluated_at_ms, exitType, terms, prefix, executedRoutes, next,nextActive);
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
      if (catalogs.doctrines?.transitions.some((value) => value.gateId === request.gate_id && value.sourceTier === state.tier && !state.doctrinesByTransition[value.transitionId])) return ["not_eligible", "doctrine_required"];
      return null;
    case "pick_doctrine": { const transition = catalogs.doctrines?.transition(request.transition_id); if (!transition) return ["unknown_id", request.transition_id]; if (!catalogs.doctrines!.allows(request.transition_id, request.doctrine_id)) return ["unknown_id", request.doctrine_id]; if (state.tier !== transition.sourceTier) return ["not_eligible", "tier"]; if (state.gatesCrossed[transition.gateId]) return ["not_eligible", "gate_crossed"]; return state.doctrinesByTransition[request.transition_id] ? ["not_eligible", "doctrine_already_picked"] : null; }
    case "spend_compute_credit":
      return request.amount_ms > catalogs.economy.offlinePolicy!.burstMaxDurationMs ? ["not_eligible", "burst_duration"] : null;
    case "claim_opportunity":
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

function parseActiveSchedule(source: unknown): ActiveScheduleEvidence {
  const raw=exactObject(source,["attended_now_ms","before_sequence","before_next_opportunity_attended_ms","after_sequence","after_next_opportunity_attended_ms","expired_buffs","missed_opportunity_id","spawned","claim"],"active-play resolved");
  const expired=array(raw.expired_buffs,"expired buffs").map((item)=>{const row=exactObject(item,["buff_instance_id"],"expired buff");return{buff_instance_id:uuidV7String(row.buff_instance_id)};});
  for(let index=1;index<expired.length;index++)if(byteCompare(expired[index-1]!.buff_instance_id,expired[index]!.buff_instance_id)>=0)throw new SyntaxError("expired buffs not sorted");
  const spawned=raw.spawned===null?null:parseActiveSpawn(raw.spawned); const claim=raw.claim===null?null:parseActiveClaim(raw.claim);
  return{attended_now_ms:safeInteger(raw.attended_now_ms,0,MAX_EXACT_INTEGER),before_sequence:safeInteger(raw.before_sequence,0,MAX_EXACT_INTEGER),before_next_opportunity_attended_ms:safeInteger(raw.before_next_opportunity_attended_ms,0,MAX_EXACT_INTEGER),after_sequence:safeInteger(raw.after_sequence,0,MAX_EXACT_INTEGER),after_next_opportunity_attended_ms:safeInteger(raw.after_next_opportunity_attended_ms,0,MAX_EXACT_INTEGER),expired_buffs:expired,missed_opportunity_id:raw.missed_opportunity_id===null?null:uuidV7String(raw.missed_opportunity_id),spawned,claim};
}
function parseActiveSpawn(source:unknown):ActiveSpawnEvidence{const row=exactObject(source,["sequence","sampled_interval_ms","effect_draw","generator_draw","effect_row_id","selected_generator_id","opportunity_id","spawned_attended_ms","expires_attended_ms"],"active spawn");return{sequence:safeInteger(row.sequence,0,MAX_EXACT_INTEGER),sampled_interval_ms:safeInteger(row.sampled_interval_ms,1,MAX_EXACT_INTEGER),effect_draw:uint64String(row.effect_draw),generator_draw:row.generator_draw===null?null:uint64String(row.generator_draw),effect_row_id:mechanicalString(row.effect_row_id),selected_generator_id:row.selected_generator_id===null?null:mechanicalString(row.selected_generator_id),opportunity_id:uuidV7String(row.opportunity_id),spawned_attended_ms:safeInteger(row.spawned_attended_ms,0,MAX_EXACT_INTEGER),expires_attended_ms:safeInteger(row.expires_attended_ms,1,MAX_EXACT_INTEGER)};}
function parseActiveClaim(source:unknown):ActiveClaimEvidence{const row=exactObject(source,["opportunity_id","effect_row_id","selected_target","buff_instance_id","requested_delta","actual_credited_delta","saturated","cap_reason_key","next_sampled_interval_ms","next_opportunity_attended_ms"],"active claim");return{opportunity_id:uuidV7String(row.opportunity_id),effect_row_id:mechanicalString(row.effect_row_id),selected_target:row.selected_target===null?null:mechanicalString(row.selected_target),buff_instance_id:row.buff_instance_id===null?null:uuidV7String(row.buff_instance_id),requested_delta:row.requested_delta===null?null:canonicalDecimalString(row.requested_delta),actual_credited_delta:row.actual_credited_delta===null?null:canonicalDecimalString(row.actual_credited_delta),saturated:row.saturated===null?null:boolean(row.saturated),cap_reason_key:row.cap_reason_key===null?null:mechanicalString(row.cap_reason_key),next_sampled_interval_ms:safeInteger(row.next_sampled_interval_ms,1,MAX_EXACT_INTEGER),next_opportunity_attended_ms:safeInteger(row.next_opportunity_attended_ms,1,MAX_EXACT_INTEGER)};}
function uint64String(value:unknown):string{const result=string(value);if(!/^(0|[1-9][0-9]*)$/.test(result)){throw new SyntaxError("invalid uint64");}const parsed=BigInt(result);if(parsed<0n||parsed>((1n<<64n)-1n))throw new SyntaxError("invalid uint64");return result;}
function canonicalDecimalString(value:unknown):string{const result=string(value);parseCanonical(result);return result;}

async function applyActiveSchedule(state:ReplayState,catalogs:ReplayCatalogBundle,command:ReplayCommand,nowMs:number,evidence:ActiveScheduleEvidence):Promise<ReplayEvent[]>{
  const catalog=catalogs.opportunities;if(state.wireVersion!==18||!catalog||evidence.before_sequence!==state.opportunitySpawnSeq||evidence.before_next_opportunity_attended_ms!==state.nextOpportunityAttendedMs)throw new RangeError("active schedule state mismatch");
  const clock=cloneReplayState(state,catalogs);if(nowMs>clock.evaluatedThroughMs)recordOfflineSpan(clock,clock.evaluatedThroughMs,nowMs,catalogs.prestige.catchupCeilingMs);const attended=attendedMS(clock,nowMs);if(attended!==evidence.attended_now_ms)throw new RangeError("active attended mismatch");
  const expired=state.activeBuffs.filter((row)=>row.expiresAttendedMs<=attended).map((row)=>row.buffInstanceId).sort(byteCompare);if(canonicalJSONString(expired)!==canonicalJSONString(evidence.expired_buffs.map((row)=>row.buff_instance_id)))throw new RangeError("active expiry mismatch");state.activeBuffs=state.activeBuffs.filter((row)=>row.expiresAttendedMs>attended);
  let missed:string|null=null,spawnApplied=false;for(let transitions=0;;transitions++){if(transitions>=catalog.schedule.maxDueTransitions)throw new RangeError("active due-transition ceiling");if(state.pendingOpportunity!==null){if(state.pendingOpportunity.expiresAttendedMs>attended)break;const expiredAt=state.pendingOpportunity.expiresAttendedMs;missed=state.pendingOpportunity.opportunityId;state.pendingOpportunity=null;state.nextOpportunityAttendedMs=!spawnApplied&&evidence.spawned!==null?expiredAt+evidence.spawned.sampled_interval_ms:evidence.after_next_opportunity_attended_ms;if(!Number.isSafeInteger(state.nextOpportunityAttendedMs)||state.nextOpportunityAttendedMs<0)throw new RangeError("active post-miss schedule overflow");continue;}if(state.nextOpportunityAttendedMs===0||state.nextOpportunityAttendedMs>attended)break;if(spawnApplied)throw new RangeError("multiple active spawns are not representable");const spawn=evidence.spawned;if(!spawn||spawn.sequence!==state.opportunitySpawnSeq||spawn.spawned_attended_ms!==state.nextOpportunityAttendedMs||spawn.expires_attended_ms-spawn.spawned_attended_ms!==catalog.schedule.lifetimeMs)throw new RangeError("active spawn coordinate mismatch");const seed=await founderSeed(command.founder_id,state.runSeq),selection=selectActivePlayEffect(catalog,seed,spawn.sequence);if(selection.effectRowId!==spawn.effect_row_id||selection.effectDraw.toString()!==spawn.effect_draw||(selection.generatorDraw?.toString()??null)!==spawn.generator_draw||selection.selectedGenerator!==spawn.selected_generator_id||activePlayOpportunityId(seed,spawn.sequence,spawn.spawned_attended_ms)!==spawn.opportunity_id)throw new RangeError("active integer draw mismatch");state.pendingOpportunity={opportunityId:spawn.opportunity_id,spawnedAttendedMs:spawn.spawned_attended_ms,expiresAttendedMs:spawn.expires_attended_ms,effectRowId:spawn.effect_row_id,selectedGeneratorId:spawn.selected_generator_id};state.opportunitySpawnSeq++;state.nextOpportunityAttendedMs=0;spawnApplied=true;}
  if(missed!==evidence.missed_opportunity_id||(evidence.spawned!==null)!==spawnApplied)throw new RangeError("active compound schedule mismatch");
  if(state.opportunitySpawnSeq!==evidence.after_sequence||state.nextOpportunityAttendedMs!==evidence.after_next_opportunity_attended_ms)throw new RangeError("active scheduler result mismatch");
  const events:ReplayEvent[]=[];for(const id of expired)events.push(event("buff_expired.v1","",{buff_instance_id:id,attended_ms:attended}));if(missed)events.push(event("opportunity_expired.v1","",{opportunity_id:missed,attended_ms:attended}));if(evidence.spawned)events.push(event("opportunity_spawned.v1","",{opportunity_id:evidence.spawned.opportunity_id,spawned_attended_ms:evidence.spawned.spawned_attended_ms,expires_attended_ms:evidence.spawned.expires_attended_ms,effect_row_id:evidence.spawned.effect_row_id,selected_generator_id:evidence.spawned.selected_generator_id}));return events;
}

function activeContributions(state:ReplayState,catalog:ActivePlayCatalog,attended:number):ReplayContribution[]{return activeContributionSet(state,catalog,attended).values;}
function activeContributionSet(state:ReplayState,catalog:ActivePlayCatalog,attended:number):{values:ReplayContribution[];saturated:boolean}{const result:ReplayContribution[]=[];for(const buff of state.activeBuffs){if(buff.expiresAttendedMs<=attended)continue;const effect=catalog.effects.find((row)=>row.effectRowId===buff.effectRowId);if(!effect)throw new RangeError("unknown active buff");const source=`active_play.${effect.effectRowId}.${buff.buffInstanceId}`;if(effect.kind==="production_frenzy")result.push({slot:"event_buffs",source_id:source,target:"all",factor:effect.factor});else if(effect.kind==="click_frenzy")for(const action of effect.actionIds)result.push({slot:"event_buffs",source_id:source,target:action,factor:effect.factor});else if(effect.kind==="building_special"){if(buff.selectedTarget===null)throw new RangeError("missing active target");const owned=state.generators[buff.selectedTarget];if(!Number.isSafeInteger(owned)||owned!<0)throw new RangeError("invalid active owned count");result.push({slot:"event_buffs",source_id:source,target:buff.selectedTarget,factor:countPpmFactor(owned!,effect.perOwnedPpm)});}else throw new RangeError("instant effect cannot persist");}result.sort((a,b)=>byteCompare(contributionKey(a),contributionKey(b)));return clampActiveContributions(result,catalog);}
function clampActiveContributions(values:ReplayContribution[],catalog:ActivePlayCatalog):{values:ReplayContribution[];saturated:boolean}{const result=values.map((row)=>({...row})),cap=parseCanonical(catalog.combo.cap),byTarget=new Map<string,number[]>();for(let index=0;index<result.length;index++){const target=result[index]!.target;byTarget.set(target,[...(byTarget.get(target)??[]),index]);}const all=clampActiveContributionGroup(result,byTarget.get("all")??[],cap);let saturated=all.saturated;for(const target of [...byTarget.keys()].filter((value)=>value!=="all").sort(byteCompare)){const indexes=byTarget.get(target)!;const combines=indexes.some((index)=>{const declaration=activeDeclarationId(result[index]!.source_id),effect=catalog.effects.find((row)=>row.effectRowId===declaration);return effect?.kind==="building_special";});const limit=combines?safeActiveCapFactor(cap,all.product):cap,group=clampActiveContributionGroup(result,indexes,limit);saturated=saturated||group.saturated;}return{values:result,saturated};}
function clampActiveContributionGroup(values:ReplayContribution[],indexes:number[],cap:Decimal):{product:Decimal;saturated:boolean}{let product=new Decimal(1);for(let at=0;at<indexes.length;at++){const index=indexes[at]!,candidate=quantize(product.mul(parseCanonical(values[index]!.factor)));if(candidate.gt(cap)){const factor=safeActiveCapFactor(cap,product);values[index]={...values[index]!,factor:canonicalString(factor)};for(const later of indexes.slice(at+1))values[later]={...values[later]!,factor:"1e0"};return{product:quantize(product.mul(factor)),saturated:true};}product=candidate;}return{product,saturated:false};}
function safeActiveCapFactor(cap:Decimal,product:Decimal):Decimal{const nominal=quantize(cap.div(product));for(const offset of [0,-1,-2]){const candidate=quantize(Decimal.fromMantissaExponent(nominal.mantissa+offset*1e-11,nominal.exponent));if(isStateValue(candidate)&&candidate.gt(0)&&quantize(product.mul(candidate)).lte(cap))return candidate;}return new Decimal(1);}

async function claimOpportunity(state:ReplayState,catalogs:ReplayCatalogBundle,command:ReplayCommand,request:Intent,schedule:ActiveScheduleEvidence,contributions:readonly ReplayContribution[]):Promise<{rejection?:[string,string];claim?:ActiveClaimEvidence;events:ReplayEvent[]}>{
  const pending=state.pendingOpportunity;if(!pending){if(schedule.missed_opportunity_id===request.opportunity_id)return{rejection:["not_eligible","opportunity_expired"],events:[]};return{rejection:["not_eligible","opportunity_not_pending"],events:[]};}if(pending.opportunityId!==request.opportunity_id)return{rejection:["unknown_id","opportunity_id"],events:[]};if(pending.expiresAttendedMs<=schedule.attended_now_ms)throw new RangeError("expired pending opportunity");
  const effect=catalogs.opportunities!.effects.find((row)=>row.effectRowId===pending.effectRowId);if(!effect)throw new RangeError("unknown opportunity effect");const expected=schedule.claim;if(expected===null)throw new RangeError("missing applied active claim evidence");const nextCoordinate=schedule.attended_now_ms+expected.next_sampled_interval_ms;if(!Number.isSafeInteger(nextCoordinate)||nextCoordinate!==expected.next_opportunity_attended_ms)throw new RangeError("next opportunity schedule mismatch");state.pendingOpportunity=null;state.nextOpportunityAttendedMs=nextCoordinate;
  const claim:ActiveClaimEvidence={opportunity_id:pending.opportunityId,effect_row_id:pending.effectRowId,selected_target:pending.selectedGeneratorId,buff_instance_id:null,requested_delta:null,actual_credited_delta:null,saturated:null,cap_reason_key:null,next_sampled_interval_ms:expected.next_sampled_interval_ms,next_opportunity_attended_ms:expected.next_opportunity_attended_ms};const events:ReplayEvent[]=[];let claimPayload:Record<string,unknown>={opportunity_id:pending.opportunityId,effect_row_id:pending.effectRowId,selected_target:pending.selectedGeneratorId};
  if(effect.kind==="lucky_payout"){const bank=parseCanonical(state.balances[effect.resourceId]!);const rates=productionRates(catalogs.economy,state.generators,state.generatorsProvisioned,contributions);const rate=sumDeterministic(rates.get(effect.resourceId)??[]),requested=parseCanonical(activePlayLuckyRequested(canonicalString(bank),canonicalString(rate),effect.luckyBankFrac,effect.luckyRateCap,effect.epsilon));const changes=applyLedger(state,catalogs.economy,[{resource:effect.resourceId,delta:requested}],true);const actual=changes.find((row)=>row.resource_id===effect.resourceId)?.delta??"0";const saturated=!parseCanonical(actual).eq(requested);claim.requested_delta=canonicalString(requested);claim.actual_credited_delta=actual;claim.saturated=saturated;claim.cap_reason_key=saturated?effect.hardcapReasonKey:null;claimPayload={...claimPayload,requested_delta:claim.requested_delta,actual_credited_delta:actual,saturated,cap_reason_key:claim.cap_reason_key};}
  else{const buffId=activePlayBuffId(await founderSeed(command.founder_id,state.runSeq),state.opportunitySpawnSeq-1,schedule.attended_now_ms);claim.buff_instance_id=buffId;const selected=effect.kind==="building_special"?pending.selectedGeneratorId:null;if(effect.kind==="building_special"&&selected===null)throw new RangeError("missing building target");const buff={buffInstanceId:buffId,effectRowId:effect.effectRowId,selectedTarget:selected,activatedAttendedMs:schedule.attended_now_ms,expiresAttendedMs:schedule.attended_now_ms+effect.durationMs};state.activeBuffs.push(buff);state.activeBuffs.sort((a,b)=>byteCompare(a.buffInstanceId,b.buffInstanceId));const activeCatalog=catalogs.opportunities!,combo=activeContributionSet(state,activeCatalog,schedule.attended_now_ms);claim.cap_reason_key=combo.saturated?activeCatalog.combo.hardcapReasonKey:null;claimPayload={...claimPayload,buff_instance_id:buffId,cap_reason_key:claim.cap_reason_key};events.push(event("buff_started.v1",request.intent_id,{buff_instance_id:buffId,effect_row_id:effect.effectRowId,selected_target:selected,activated_attended_ms:buff.activatedAttendedMs,expires_attended_ms:buff.expiresAttendedMs,hardcap_reason_key:claim.cap_reason_key},2));}
  if(canonicalJSONString(claim)!==canonicalJSONString(expected))throw new RangeError("active claim evidence mismatch");events.unshift(event("opportunity_claimed.v1",request.intent_id,claimPayload,effect.kind==="lucky_payout"?1:2));return{claim,events};
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
function validateContributionSet(catalog: EconomyCatalog, values: readonly ReplayContribution[]): void { const declared = new Map(catalog.multiplierSources.map((value) => [value.id, value])); const seen = new Set<string>(); for (const value of values) { const declarationId=activeDeclarationId(value.source_id);const source=declared.get(declarationId)??(declarationId===value.source_id?undefined:declared.get(`${declarationId}.${value.target}`)); const factor = parseCanonical(value.factor);const identity=`${value.source_id}\0${value.target}`; if (!source || source.slot !== value.slot || source.target !== value.target || declarationId!==value.source_id&&source.provider!=="active_play" || seen.has(identity) || !isStateValue(factor) || !factor.gt(0)) throw new RangeError("invalid multiplier contribution"); seen.add(identity); } }
function activeDeclarationId(sourceId:string):string{if(!sourceId.startsWith("active_play."))return sourceId;const parts=sourceId.slice("active_play.".length).split(".");const index=parts.findIndex((part)=>part.includes("-"));return index<1?sourceId:parts.slice(0,index).join(".");}
function contributionFactorForTarget(catalog: EconomyCatalog, target: string, values: readonly ReplayContribution[]): Decimal { validateContributionSet(catalog, values); let factor = new Decimal(1); for (const slot of MULTIPLIER_SLOT_ORDER) for (const value of values.filter((item) => item.slot === slot && item.target === target).sort((a, b) => byteCompare(a.source_id, b.source_id))) factor = factor.mul(parseCanonical(value.factor)); const result = quantize(factor); if (!isStateValue(result) || !result.gt(0)) throw new RangeError("invalid target contribution factor"); return result; }

function evaluate(state: ReplayState, catalog: EconomyCatalog, nowMs: number, mode: "online" | "offline", contributions: readonly ReplayContribution[]): Evaluation {
  if (nowMs < state.evaluatedThroughMs) throw new ReplayClockViolation(); if (nowMs === state.evaluatedThroughMs) return { changes: [], elapsedMs: 0, productionMs: 0, bankedMs: 0, progressDeltaPpm: 0 };
  const elapsedMs = nowMs - state.evaluatedThroughMs; let productionMs = elapsedMs; let efficiency = new Decimal(1); let bankedMs = 0; const beforeProgress = progressPpm(catalog, state);
  if (state.computeBurstRemainingMs < 0 || state.computeBurstRemainingMs > 0 && state.wireVersion < 17) throw new RangeError("invalid compute burst state");
  if (mode === "offline") { const policy = catalog.offlinePolicy!; efficiency = parseCanonical(policy.efficiency); productionMs = Math.min(productionMs, policy.accrualCapMs); bankedMs = Number((BigInt(elapsedMs - productionMs) * BigInt(policy.bankRatioNumerator)) / BigInt(policy.bankRatioDenominator)); bankedMs = Math.min(bankedMs, policy.bankCapMs - state.computeCreditMs); }
  const boostedWallMs = Math.min(elapsedMs, state.computeBurstRemainingMs); const boostedProductionMs = Math.min(productionMs, boostedWallMs);
  const bonusFactor = boostedProductionMs === 0 ? new Decimal(0) : parseCanonical(catalog.offlinePolicy!.burstSpeed).sub(1);
  if (boostedProductionMs > 0 && (!isStateValue(bonusFactor) || !bonusFactor.gt(0))) throw new RangeError("invalid compute burst speed");
  const accrued = accrueContent(state, catalog, productionMs, boostedProductionMs, efficiency, bonusFactor, contributions);
  const changes = applyLedger(state, catalog, accrued.entries, true); state.computeCreditMs += bankedMs; state.computeBurstRemainingMs -= boostedWallMs; state.generatorsProvisioned = accrued.provisioned; state.provisionRemaindersPpm = accrued.remainders; state.evaluatedThroughMs = nowMs; return { changes, elapsedMs, productionMs, bankedMs, progressDeltaPpm: Math.max(0, progressPpm(catalog, state) - beforeProgress) };
}

function accrueContent(state: ReplayState, catalog: EconomyCatalog, productionMs: number, boostedProductionMs: number, efficiency: Decimal, bonusFactor: Decimal, contributions: readonly ReplayContribution[]): { entries: { resource: string; delta: Decimal }[]; provisioned: Record<string, number>; remainders: Record<string, number> } {
  const provisioned = { ...state.generatorsProvisioned }; const remainders = { ...state.provisionRemaindersPpm }; const deltas = new Map<string, Decimal[]>();
  let remainingBoostedMs = boostedProductionMs;
  const accrueSegment = (segmentMs: number): void => {
    if (segmentMs <= 0) return;
    const rates = productionRates(catalog, state.generators, provisioned, contributions);
    const bonusMs = Math.min(segmentMs, remainingBoostedMs);
    for (const [resource, values] of rates) { const prior = deltas.get(resource) ?? []; const rate = sumDeterministic(values); const effectiveSeconds = new Decimal(segmentMs).add(new Decimal(bonusMs).mul(bonusFactor)).div(1000); const delta = quantize(rate.mul(effectiveSeconds).mul(efficiency)); if (!delta.eq(0)) prior.push(delta); if (prior.length > 0) deltas.set(resource, prior); }
    remainingBoostedMs -= bonusMs;
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
  validateContributionSet(catalog,contributions);const bySource = new Map<string, ReplayContribution>();
  for (const contribution of contributions) { const identity=`${contribution.source_id}\0${contribution.target}`;bySource.set(identity, contribution); }
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
  if (!foundationsActive(catalogs) || state.wireVersion < 16) throw new RangeError("invalid foundation hook inputs");
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
  const context: RouteContext = { contextVersion: routes.contextVersion, resources: state.balances, doctrinesByTransition: state.doctrinesByTransition, structureId: state.structureId, ledgerFactKinds: state.ledgerFactKinds, meterBands: state.wireVersion >= 16 ? state.meterValues : state.meterBands, regionTraits: state.regionTraits };
  if (!evaluatePredicate(upgrade.requires, context)) return { rejection: ["not_eligible", "requires"] };
  const balance = parseCanonical(state.balances[upgrade.cost.resourceId]!); const cost = parseCanonical(upgrade.cost.amount);
  if (balance.lt(cost)) return { rejection: ["unaffordable", request.upgrade_id] };
  applyLedger(state, catalog, [{ resource: upgrade.cost.resourceId, delta: cost.neg() }], false); state.upgradesOwned.add(upgrade.id); request.cost = upgrade.cost.amount; request.costResource = upgrade.cost.resourceId; return {};
}
function manualBatch(state: ReplayState, catalog: EconomyCatalog, request: Intent, nowMs: number, contributions: readonly ReplayContribution[]): { count: number; rejection?: [string, string] } { const action = catalog.manualActions.find((value) => value.id === request.action_id)!; const policy = catalog.manualPolicy!; if (nowMs > state.manualTokenRefilledAtMs) { const elapsed = nowMs - state.manualTokenRefilledAtMs; state.manualTokenMilli = Math.min(policy.bucketCapMilli, state.manualTokenMilli + elapsed * policy.refillMilliPerMs); state.manualTokenRefilledAtMs = nowMs; } const applied = Math.min(request.count, Math.floor(state.manualTokenMilli / 1000)); state.manualTokenMilli -= applied * 1000; if (applied > 0) { try { const factor = contributionFactorForTarget(catalog, request.action_id, contributions); applyLedger(state, catalog, [{ resource: action.output.resourceId, delta: parseCanonical(action.output.amountPerAction).mul(applied).mul(factor) }], false); } catch (error) { if (error instanceof LedgerError && error.code === "above_hardcap") return { count: 0, rejection: ["cap_exceeded", request.action_id] }; throw error; } } return { count: applied }; }
function crossGate(state: ReplayState, routes: RoutesCatalog, request: Intent, command: ReplayCommand): { rejection?: [string, string] } { const gate = routes.gate(request.gate_id); if (!gate) return { rejection: ["unknown_id", request.gate_id] }; if (state.gatesCrossed[request.gate_id]) return { rejection: ["gate_already_crossed", request.gate_id] }; let requirements = gate.requirement; if (request.route_id !== null) { const route = routes.route(request.route_id); if (!route) return { rejection: ["unknown_id", request.route_id] }; const context: RouteContext = { contextVersion: routes.contextVersion, resources: state.balances, doctrinesByTransition: state.doctrinesByTransition, structureId: state.structureId, ledgerFactKinds: state.ledgerFactKinds, meterBands: state.wireVersion >= 16 ? state.meterValues : state.meterBands, regionTraits: state.regionTraits }; if (!route.active || route.requiresContextVersion > context.contextVersion || !evaluatePredicate(route.predicate, context)) return { rejection: ["route_predicate_unmet", request.route_id] }; requirements = discountedRequirements(gate, route); }
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

async function finishLoggedExit(company: ReplayState, founder: FounderCarry, intentId: string, command: ReplayCommand, nowMs: number, exitType: string, inputTerms: ExitTerms, prefix: ReplayEvent[], executedRoutes: string[], next: ReplayCatalogBundle,nextActive:ActiveSpawnEvidence|null): Promise<LoggedExitTransition> {
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
  company.computeBurstRemainingMs = 0;
  bindEventIntent(prefix, intentId);
  const currentActive = company.wireVersion >= 16;
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
  if(next.opportunities){if(nextActive===null||nextActive.sequence!==0||nextActive.spawned_attended_ms!==nextActive.sampled_interval_ms||nextActive.expires_attended_ms-nextActive.spawned_attended_ms!==next.opportunities.schedule.lifetimeMs)throw new RangeError("missing next active schedule");const seed=await founderSeed(command.founder_id,newCompany.runSeq),selection=selectActivePlayEffect(next.opportunities,seed,0);if(selection.effectRowId!==nextActive.effect_row_id||selection.effectDraw.toString()!==nextActive.effect_draw||(selection.generatorDraw?.toString()??null)!==nextActive.generator_draw||selection.selectedGenerator!==nextActive.selected_generator_id||activePlayOpportunityId(seed,0,nextActive.spawned_attended_ms)!==nextActive.opportunity_id)throw new RangeError("next active selection mismatch");newCompany.nextOpportunityAttendedMs=nextActive.spawned_attended_ms;}else if(nextActive!==null)throw new RangeError("unexpected next active schedule");
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
  return { wireVersion: bundle.opportunities ? 18 : bundle.doctrines ? 17 : active ? 16 : 14, balances, generators, generatorPurchasedTotal: 0, upgradesOwned: new Set(), generatorsProvisioned: Object.fromEntries(Object.keys(generators).map((id) => [id, 0])), provisionRemaindersPpm: Object.fromEntries(catalog.generatorClasses.filter((value) => value.provision !== null).map((value) => [value.provision!.generatorId, 0])), stockRateRemainderPpm: 0, evaluatedThroughMs: nowMs, computeCreditMs: 0, computeBurstRemainingMs: 0, opportunitySpawnSeq: 0, nextOpportunityAttendedMs: 0, pendingOpportunity: null, activeBuffs: [], manualTokenMilli: catalog.manualPolicy!.bucketCapMilli, manualTokenRefilledAtMs: nowMs, gatesCrossed: {}, runSeq: prior.runSeq + 1, doctrinesByTransition: {}, structureId: "", ledgerFactKinds: new Set(), meterBands: active ? {} : { "trust.regulators.standing": reseed, "trust.regulators.grievance": 100 - reseed }, regionTraits: new Set(), routeKnowledgeBalance: 0, hintsUnlocked: new Set(), compactMember: false, compactTithePpm: 0, compactSolidarityPpm: 0, compactSamples: [], tier: 0, lifetimeValue: "0", offerState: null, runStartedAtMs: nowMs, runPreTimer: false, offlineSpans: [], collapsedOfflineMs: 0, factionId: "", incorporatedAtMs: null, factionStockResource: "", stockUnits: 0, stockProgressMs: 0, consumedStockUnits: 0, guildTitheCarryPpm: 0, guildBoundaryGuildId: prior.guildBoundaryGuildId, guildBoundarySeq: prior.guildBoundarySeq, guildConsumedWindow: 0, meterValues: meterState?.values ?? {}, meterDecayRemainders: meterState?.decayRemainders ?? {}, meterInputRemainders: meterState?.inputRemainders ?? {}, achievementsEarnedRun: new Set(), achievementScoreRun: 0 };
}

function rejectedExit(company: ReplayState, founder: FounderCarry, intentId: string, revision: number, category: string, detail: string): LoggedExitTransition {
  return { founder, finalCompany: company, newCompany: null, outcome: "rejected", receipt: { current_revision: revision, intent_id: intentId, outcome: "rejected", rejection: { category, detail } }, founderEvents: [], companyEndedEvents: [], companyStartedEvents: [] };
}

function sortedUniqueMechanical(source: unknown[]): string[] {
  let last = "";
  return source.map((item) => { const value = mechanicalString(item); if (byteCompare(value, last) <= 0) throw new SyntaxError("values must be sorted and unique"); last = value; return value; });
}

function parseReplayWire(source: unknown, state: ReplayState, catalogs: ReplayCatalogBundle): ReplayWire { const root = exactObject(source, ["v", "command", "evaluated_at_ms", "evaluation_mode", "resolved"], "replay inputs"); if (root.v !== 2 && root.v !== 3 && root.v !== 4 && root.v !== 5 || foundationsActive(catalogs) && root.v < 3 || root.evaluation_mode !== "online" && root.evaluation_mode !== "offline") throw new SyntaxError("invalid replay envelope"); const command = objectWithOnlyKeys(root.command, ["intent_id", "company_stream_id", "founder_id", "revision", "run_seq", "run_log_seq"], "command"); const parsed: ReplayCommand = { intent_id: uuidV7String(command.intent_id), company_stream_id: command.company_stream_id === undefined ? "" : string(command.company_stream_id), founder_id: command.founder_id === undefined ? "" : string(command.founder_id), revision: safeInteger(command.revision, 1, MAX_EXACT_INTEGER), run_seq: safeInteger(command.run_seq, 1, MAX_EXACT_INTEGER), run_log_seq: safeInteger(command.run_log_seq, 1, MAX_EXACT_INTEGER) }; if (parsed.run_seq !== state.runSeq || !hashPattern.test(catalogs.constantsHash)) throw new RangeError("replay command mismatch"); return { v: root.v as 2|3|4|5, command: parsed, evaluated_at_ms: safeInteger(root.evaluated_at_ms, 1, MAX_EXACT_INTEGER), evaluation_mode: root.evaluation_mode, resolved: objectWithOnlyKeys(root.resolved, Object.keys(root.resolved as object), "resolved") }; }
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
	const probe = wire.resolved;
	const probedVersion = safeInteger(probe.result_founder_wire_version, 1, 21);
	const activatesSoul = probedVersion >= 20 && state.wireVersion < 20;
  const keys = ["kind", "outcome", "company_stream_id", "run_seq", "run_log_seq", "result_constants_hash", "reputation_delta", "route_knowledge_delta", "attended_ms", "age_ms_before", "age_ms_after", "achievement_score_delta", "added_network_slots", "added_ledger_fact_kinds", "added_lifetime_achievements", "exit_record", "result_founder_wire_version", "rejection", ...(activatesSoul ? ["next_soul"] : [])];
	const raw = exactObject(wire.resolved, keys, "Founder Exit inputs");
  if (request.kind !== "cross_gate" && request.expected_founder_revision !== wire.command.revision) throw new RangeError("Founder Exit revision mismatch");
  const companyStreamId = uuidString(raw.company_stream_id); const runSeq = safeInteger(raw.run_seq, 1, MAX_EXACT_INTEGER); safeInteger(raw.run_log_seq, 1, MAX_EXACT_INTEGER);
  const resultHash = string(raw.result_constants_hash); if (!hashPattern.test(resultHash)) throw new SyntaxError("invalid Founder result hash");
  const ageBefore = safeInteger(raw.age_ms_before, 0, MAX_EXACT_INTEGER); const ageAfter = safeInteger(raw.age_ms_after, 0, MAX_EXACT_INTEGER); const attended = safeInteger(raw.attended_ms, 0, MAX_EXACT_INTEGER);
  if (ageBefore !== state.ageMs || ageAfter < ageBefore || attended !== ageAfter - ageBefore) throw new RangeError("invalid Founder attendance facts");
	const resultVersion = safeInteger(raw.result_founder_wire_version, 1, 21);
	if (![14, 15, 16, 17, 18, 19, 20, 21].includes(resultVersion)) throw new RangeError("unsupported Founder result version");
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
		if (!resultCatalogs.minigames) throw new RangeError("missing minigame activation artifact");
		state.minigameRatings = Object.fromEntries(resultCatalogs.minigames.minigames.map((definition) => [definition.minigame_id, {
			elo: definition.rating_policy.starting_elo, season_member: definition.rating_policy.season_member, games_counted: 0,
		}]));
		state.minigameOfflineQuality = Object.fromEntries(resultCatalogs.minigames.minigames.map((definition) => [definition.minigame_id, {
			grade_ppm: definition.offline_quality.neutral_floor_ppm, last_founder_attended_ms: state.ageMs, decay_remainder_ppm: 0,
		}]));
  }
  if (resultVersion >= 18 && state.wireVersion < 18) {
    if (!resultCatalogs.pets) throw new RangeError("missing pet activation artifact"); state.pets = {};
  }
	if (resultVersion >= 19 && state.wireVersion < 19) {
    if (!resultCatalogs.fiscal) throw new RangeError("missing fiscal activation artifact");
    state.fiscalCredit = 0; state.fiscalPeriodOpenedWallMs = safeInteger(wire.command.server_ts_ms, 1, MAX_EXACT_INTEGER); state.fiscalPeriodSequence = 0;
    state.fiscalGeneratorLevels = Object.fromEntries(resultCatalogs.fiscal.generatorLevelRows.map((row) => [row.generatorId, 0])); state.fiscalUnlocks = new Set();
	}
	if (resultVersion >= 20 && state.wireVersion < 20) {
		if (!resultCatalogs.soul) throw new RangeError("missing Soul activation artifact");
		const evidence = exactObject(raw.next_soul, ["soul_initial", "band_member"], "next Soul activation");
		const initial = safeInteger(evidence.soul_initial, resultCatalogs.soul.policy.soul_floor, resultCatalogs.soul.policy.soul_max);
		const band = soulBand(resultCatalogs.soul, initial);
		if (initial !== resultCatalogs.soul.policy.soul_initial || evidence.band_member !== band.band_member) throw new RangeError("Soul activation evidence mismatch");
		state.soul = initial; state.soulExhaustedSourceIds = new Set();
	}
	if (resultVersion >= 21 && state.wireVersion < 21) {
		if (!resultCatalogs.minigameAPI) throw new RangeError("missing minigame API activation artifact");
		state.minigameSessionSeq = 0;
	}
	if (resultVersion >= 21) state.minigameSessionSeq = 0;
	state.wireVersion = resultVersion as 14 | 15 | 16 | 17 | 18 | 19 | 20 | 21;
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

function parseFounderCarry(source: unknown, catalogs: ReplayCatalogBundle, wireVersion: 2 | 3 | 4 | 5): FounderCarry {
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
    case "pick_doctrine":
      if (!hasExactKeys(raw, ["kind", "expected_revision", "transition_id", "doctrine_id"])) { base.invalid = "pick_doctrine.fields"; return base; }
      if (!isMechanical(raw.transition_id)) base.invalid = "transition_id"; else base.transition_id = raw.transition_id;
      if (!isMechanical(raw.doctrine_id)) base.invalid = "doctrine_id"; else base.doctrine_id = raw.doctrine_id;
      return base;
    case "spend_compute_credit":
      if (!hasExactKeys(raw, ["kind", "expected_revision", "amount_ms", "target"])) { base.invalid = "spend_compute_credit.fields"; return base; }
      if (!isPositiveSafeInteger(raw.amount_ms)) base.invalid = "amount_ms"; else base.amount_ms = raw.amount_ms;
      if (raw.target !== "accelerate") base.invalid = "target"; else base.target = raw.target;
      return base;
    case "harvest_fiscal_period":
      if (!hasExactKeys(raw, ["kind", "expected_revision"])) base.invalid = "harvest_fiscal_period.fields";
      return base;
    case "spend_fiscal_credit":
      if (!hasExactKeys(raw, ["kind", "expected_revision", "target"])) { base.invalid = "spend_fiscal_credit.fields"; return base; }
      try { base.target = fiscalTargetWire(parseFiscalTarget(raw.target)); } catch { base.invalid = "target"; }
      return base;
    case "claim_opportunity":
      if (!hasExactKeys(raw, ["kind", "expected_revision", "opportunity_id"])) { base.invalid = "claim_opportunity.fields"; return base; }
      if (!isUUIDV7(raw.opportunity_id)) base.invalid = "opportunity_id"; else base.opportunity_id = raw.opportunity_id;
      return base;
    case "buy_route_hint":
      if (!hasExactKeys(raw, ["kind", "expected_revision", "route_id"])) { base.invalid = "buy_route_hint.fields"; return base; }
      if (!isMechanical(raw.route_id)) base.invalid = "route_id"; else base.route_id = raw.route_id;
      return base;
    case "care_action":
      if (!hasExactKeys(raw, ["kind", "expected_revision", "pet_id", "action_id"])) { base.invalid = "care_action.fields"; return base; }
      if (!isUUIDV7(raw.pet_id)) base.invalid = "pet_id"; else base.pet_id = raw.pet_id;
      if (!isMechanical(raw.action_id)) base.invalid = "action_id"; else base.action_id = raw.action_id;
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
  if (state.wireVersion >= 16) Object.assign(snapshot, { meter_values: sortedRecord(state.meterValues), meter_decay_remainders: sortedRecord(state.meterDecayRemainders), meter_input_remainders: sortedRecord(state.meterInputRemainders), achievements_earned_run: [...state.achievementsEarnedRun].sort(byteCompare), achievement_score_run: state.achievementScoreRun });
  else snapshot.meter_bands = sortedRecord(state.meterBands);
  if (state.wireVersion >= 17) snapshot.compute_burst_remaining_ms = state.computeBurstRemainingMs;
  if (state.wireVersion >= 18) Object.assign(snapshot, { opportunity_spawn_seq: state.opportunitySpawnSeq, next_opportunity_attended_ms: state.nextOpportunityAttendedMs,
    pending_opportunity: state.pendingOpportunity === null ? null : { opportunity_id: state.pendingOpportunity.opportunityId, spawned_attended_ms: state.pendingOpportunity.spawnedAttendedMs,
      expires_attended_ms: state.pendingOpportunity.expiresAttendedMs, effect_row_id: state.pendingOpportunity.effectRowId, selected_generator_id: state.pendingOpportunity.selectedGeneratorId },
    active_buffs: state.activeBuffs.map((value) => ({ buff_instance_id: value.buffInstanceId, effect_row_id: value.effectRowId, selected_target: value.selectedTarget,
      activated_attended_ms: value.activatedAttendedMs, expires_attended_ms: value.expiresAttendedMs })) });
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
