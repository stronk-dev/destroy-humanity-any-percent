import type { GameUIFeatures, GameUISnapshot, GameUISnapshotV1, GameUISnapshotV2, GameUISnapshotV3 } from "../api/generated/types";
import { parseCanonical } from "../numeric";
import type { AuthoritativeSnapshot, DiscreteFact } from "../shell/contracts";

const mechanicalID = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;
const hash = /^sha256:[0-9a-f]{64}$/;
const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;

function object(value: unknown, label: string): Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value)) throw new SyntaxError(`${label} must be an object`);
  return value as Record<string, unknown>;
}

function exact(value: Record<string, unknown>, keys: readonly string[], label: string): void {
  const actual = Object.keys(value).sort();
  const expected = [...keys].sort();
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError(`${label} fields are not exact`);
}

function integer(value: unknown, minimum: number, maximum = Number.MAX_SAFE_INTEGER): number {
  if (!Number.isSafeInteger(value) || (value as number) < minimum || (value as number) > maximum) throw new SyntaxError("invalid safe integer");
  return value as number;
}

function identifier(value: unknown): string {
  if (typeof value !== "string" || !mechanicalID.test(value)) throw new SyntaxError("invalid mechanical ID");
  return value;
}

function decimal(value: unknown, positive = false): string {
  if (typeof value !== "string") throw new SyntaxError("invalid Decimal string");
  const parsed = parseCanonical(value);
  if (positive ? parsed.lte(0) : parsed.lt(0)) throw new SyntaxError("invalid Decimal domain");
  return value;
}

function sortedRows(values: unknown, id: string, label: string): Record<string, unknown>[] {
  if (!Array.isArray(values)) throw new SyntaxError(`${label} must be an array`);
  const rows = values.map((value, index) => object(value, `${label}[${index}]`));
  for (let index = 0; index < rows.length; index += 1) {
    identifier(rows[index][id]);
    if (index > 0 && String(rows[index - 1][id]) >= String(rows[index][id])) throw new SyntaxError(`${label} must be byte-sorted and unique`);
  }
  return rows;
}

export type ParsedGameUISnapshot = GameUISnapshot | GameUISnapshotV1 | GameUISnapshotV2 | GameUISnapshotV3;

// Live sync requires the current v4 shape (GS0.1 rule 4); v1–v3 remain
// decodable only for stored bootstrap receipts.
export function isLiveSnapshot(value: ParsedGameUISnapshot): value is GameUISnapshot { return value.schema_version === 4 && "features" in value; }

export function parseGameUISnapshot(source: unknown): ParsedGameUISnapshot {
  const root = object(source, "game UI snapshot");
  if (root.schema_version !== 1 && root.schema_version !== 2 && root.schema_version !== 3 && root.schema_version !== 4) throw new SyntaxError("invalid game UI envelope");
  const fields = ["constants_hash", "evaluated_through_ms", "facts", "generators", "manual_action", "progress", "resources", "revision", "run", "schema_version", "server_now_ms", "upgrades"];
  if (root.schema_version >= 2) fields.push("founder_revision");
  if (root.schema_version >= 3) fields.push("transitions");
  if (root.schema_version === 4) fields.push("features");
  exact(root, fields, "game UI snapshot");
  if (typeof root.constants_hash !== "string" || !hash.test(root.constants_hash)) throw new SyntaxError("invalid game UI envelope");
  const revision = integer(root.revision, 1);
  if (root.schema_version >= 2) integer(root.founder_revision, 1);
  const evaluatedThrough = integer(root.evaluated_through_ms, 1);
  const serverNow = integer(root.server_now_ms, evaluatedThrough);

  const facts = sortedRows(root.facts, "fact_id", "game UI facts");
  for (const row of facts) {
    exact(row, ["fact_id", "value"], "game UI fact");
    if (typeof row.value !== "boolean" && typeof row.value !== "string" && !Number.isSafeInteger(row.value)) throw new SyntaxError("invalid game UI fact value");
  }
  const generators = sortedRows(root.generators, "generator_id", "game UI generators");
  for (const row of generators) {
    exact(row, ["generator_id", "max_affordable", "next_cost", "next_cost_resource_id", "owned", "provisioned", "rate_contribution", ...(root.schema_version === 4 ? ["provision_cap"] : [])], "game UI generator");
    integer(row.max_affordable, 0); integer(row.owned, 0);
    const provisioned = integer(row.provisioned, 0);
    if (root.schema_version === 4 && row.provision_cap !== null) {
      const cap = parseIntCap(row.provision_cap, "game UI provision cap");
      if (provisioned > cap.amount) throw new SyntaxError("provisioned count exceeds its visible cap");
    }
    decimal(row.next_cost); decimal(row.rate_contribution); identifier(row.next_cost_resource_id);
  }
  const manual = object(root.manual_action, "game UI manual action");
  exact(manual, ["action_id", "bucket_cap_milli", "refill_milli_per_ms", "refilled_at_ms", "tokens_milli"], "game UI manual action");
  identifier(manual.action_id);
  const bucket = integer(manual.bucket_cap_milli, 1);
  integer(manual.refill_milli_per_ms, 1); integer(manual.refilled_at_ms, 1, serverNow); integer(manual.tokens_milli, 0, bucket);

  const progress = sortedRows(root.progress, "stage_id", "game UI progress");
  for (const row of progress) {
    exact(row, ["current", "stage_id", "target"], "game UI progress row");
    decimal(row.current); decimal(row.target, true);
  }
  const resources = sortedRows(root.resources, "resource_id", "game UI resources");
  for (const row of resources) {
    exact(row, ["amount", "cap", "rate_per_second", "resource_id"], "game UI resource");
    const amount = parseCanonical(decimal(row.amount)); decimal(row.rate_per_second);
    if (row.cap !== null) {
      const cap = object(row.cap, "game UI resource cap"); exact(cap, ["amount", "reason_key"], "game UI resource cap");
      const maximum = parseCanonical(decimal(cap.amount)); identifier(cap.reason_key);
      if (amount.gt(maximum)) throw new SyntaxError("resource exceeds its visible cap");
    }
  }
  const run = object(root.run, "game UI run");
  exact(run, ["category", "exit_count", "founder_id", "run_seq", "run_started_at_ms", "tier"], "game UI run");
  identifier(run.category); integer(run.exit_count, 0); integer(run.run_seq, 1); integer(run.run_started_at_ms, 1, serverNow); integer(run.tier, 0, 9);
  if (typeof run.founder_id !== "string" || !uuid.test(run.founder_id)) throw new SyntaxError("invalid Founder ID");

  if (root.schema_version === 4) parseFeatures(root.features);
  if (root.schema_version >= 3) {
    const transitions = object(root.transitions, "game UI transitions");
    // Tier 2 content §E2: `incorporate` is an optional control, present only when incorporate can apply.
    const offersIncorporate = Object.hasOwn(transitions, "incorporate");
    exact(transitions, offersIncorporate ? ["cross_gate", "incorporate", "wind_down"] : ["cross_gate", "wind_down"], "game UI transitions");
    if (offersIncorporate) {
      if ((run.tier as number) < 2) throw new SyntaxError("incorporate control below Tier 2");
      const incorporate = object(transitions.incorporate, "game UI incorporate transition");
      exact(incorporate, ["factions"], "game UI incorporate transition");
      if (!Array.isArray(incorporate.factions) || incorporate.factions.length === 0) throw new SyntaxError("invalid game UI incorporate factions");
      let prior = "";
      for (const source of incorporate.factions) {
        const row = object(source, "game UI incorporate faction");
        exact(row, ["copy_key", "faction_id"], "game UI incorporate faction");
        const factionID = identifier(row.faction_id);
        if (factionID <= prior || row.copy_key !== `incorporate.${factionID}`) throw new SyntaxError("invalid game UI incorporate faction");
        prior = factionID;
      }
    }
    if (transitions.cross_gate !== null) {
      const crossGate = object(transitions.cross_gate, "game UI cross-gate transition");
      exact(crossGate, ["eligible", "gate_id", "route_id"], "game UI cross-gate transition");
      if (typeof crossGate.eligible !== "boolean" || crossGate.route_id !== null) throw new SyntaxError("invalid game UI cross-gate transition");
      identifier(crossGate.gate_id);
    }
    const windDown = object(transitions.wind_down, "game UI wind-down transition");
    exact(windDown, ["eligible"], "game UI wind-down transition");
    if (typeof windDown.eligible !== "boolean") throw new SyntaxError("invalid game UI wind-down transition");
  }

  const upgrades = sortedRows(root.upgrades, "upgrade_id", "game UI upgrades");
  for (const row of upgrades) {
    exact(row, ["cost_amount", "cost_resource_id", "eligible", "owned", "upgrade_id"], "game UI upgrade");
    decimal(row.cost_amount); identifier(row.cost_resource_id);
    if (typeof row.eligible !== "boolean" || typeof row.owned !== "boolean" || row.owned && row.eligible) throw new SyntaxError("invalid upgrade state");
  }
  return { ...(root as unknown as ParsedGameUISnapshot), revision, evaluated_through_ms: evaluatedThrough, server_now_ms: serverNow };
}

function bool(value: unknown, label: string): boolean {
  if (typeof value !== "boolean") throw new SyntaxError(`${label} must be a boolean`);
  return value;
}

function oneOf<T extends string>(value: unknown, values: readonly T[], label: string): T {
  if (typeof value !== "string" || !values.includes(value as T)) throw new SyntaxError(`invalid ${label}`);
  return value as T;
}

function parseIntCap(value: unknown, label: string): { amount: number; reason_key: string } {
  const cap = object(value, label);
  exact(cap, ["amount", "reason_key"], label);
  return { amount: integer(cap.amount, 0), reason_key: identifier(cap.reason_key) };
}

// GS0.1 arms: exact keys, byte-sorted rows, exact safe integers. Unknown or
// malformed arms fail closed; arms that are not produced yet must be null.
export function parseFeatures(source: unknown): GameUIFeatures {
  const features = object(source, "game UI features");
  // Reputation Tree v1 R9: `reputation` is an additive optional v4 arm.
  // Pet Adoption v1 PA7: `pet_adoption` is likewise an additive optional arm.
  // Cosmetic Shop v1 §7.1: `cosmetics` is likewise an additive optional arm.
  // Clout v1 CV9: `axis_stack` is likewise an additive optional arm.
  // Garage Player Surfaces GS5: `opportunity` is likewise an additive optional arm.
  exact(features, ["achievements", "active_play", ...("axis_stack" in features ? ["axis_stack"] : []), ...("cosmetics" in features ? ["cosmetics"] : []), "fiscal", "meters", "minigames", ...("opportunity" in features ? ["opportunity"] : []), ...("pet_adoption" in features ? ["pet_adoption"] : []), "pets", ...("reputation" in features ? ["reputation"] : [])], "game UI features");
  if (features.active_play !== null || features.pets !== null) throw new SyntaxError("unproduced game UI arm must be null");
  if (features.achievements !== null) {
    const arm = object(features.achievements, "achievements arm");
    exact(arm, ["rows", "score"], "achievements arm");
    const score = object(arm.score, "achievements score");
    exact(score, ["lifetime", "run"], "achievements score");
    integer(score.lifetime, 0); integer(score.run, 0);
    for (const row of sortedRows(arm.rows, "achievement_id", "achievement rows")) {
      exact(row, ["achievement_id", "condition_scope", "copy_key", "earned", "proof_kind", "score_grant"], "achievement row");
      oneOf(row.condition_scope, ["career", "run"], "achievement scope"); identifier(row.copy_key);
      if (row.earned !== null) oneOf(row.earned, ["lifetime", "run"], "achievement earned state");
      oneOf(row.proof_kind, ["burn", "possession", "provenance"], "achievement proof"); integer(row.score_grant, 0);
    }
  }
  if (features.meters !== null) {
    const arm = object(features.meters, "meters arm");
    exact(arm, ["meters"], "meters arm");
    for (const row of sortedRows(arm.meters, "meter_id", "meter rows")) {
      exact(row, ["band_id", "bands", "max", "meter_id", "min", "value"], "meter row");
      const minimum = integer(row.min, 0), maximum = integer(row.max, minimum);
      integer(row.value, minimum, maximum);
      const bands = Array.isArray(row.bands) ? row.bands.map((band, index) => object(band, `meter band ${index}`)) : (() => { throw new SyntaxError("meter bands must be an array"); })();
      let floor = -1;
      for (const band of bands) {
        exact(band, ["band_id", "floor_value"], "meter band");
        identifier(band.band_id);
        const next = integer(band.floor_value, minimum, maximum);
        if (next <= floor) throw new SyntaxError("meter bands must ascend");
        floor = next;
      }
      if (!bands.some((band) => band.band_id === identifier(row.band_id))) throw new SyntaxError("meter band is not declared");
    }
  }
  if (features.fiscal !== null) {
    const arm = object(features.fiscal, "fiscal arm");
    exact(arm, ["credit", "credit_cap", "credit_per_period", "generator_levels", "hoard", "period", "sweep_preview", "unlocks"], "fiscal arm");
    const cap = parseIntCap(arm.credit_cap, "fiscal credit cap");
    integer(arm.credit, 0, cap.amount); integer(arm.credit_per_period, 0);
    const hoard = object(arm.hoard, "fiscal hoard");
    exact(hoard, ["cap_credits", "preview_ppm", "reason_note"], "fiscal hoard");
    integer(hoard.cap_credits, 0); integer(hoard.preview_ppm, 0); oneOf(hoard.reason_note, ["next_run"], "hoard note");
    const period = object(arm.period, "fiscal period");
    exact(period, ["auto_ms", "early_ms", "early_success_ppm", "guaranteed_ms", "opened_wall_ms", "seq"], "fiscal period");
    const early = integer(period.early_ms, 0), guaranteed = integer(period.guaranteed_ms, early);
    integer(period.auto_ms, guaranteed); integer(period.early_success_ppm, 0, 1_000_000); integer(period.opened_wall_ms, 0); integer(period.seq, 0);
    const sweep = object(arm.sweep_preview, "fiscal sweep preview");
    exact(sweep, ["credit_after", "credited", "periods", "saturated"], "fiscal sweep preview");
    integer(sweep.credit_after, 0, cap.amount); integer(sweep.credited, 0); integer(sweep.periods, 0); bool(sweep.saturated, "sweep saturated");
    for (const row of sortedRows(arm.generator_levels, "generator_id", "fiscal levels")) {
      exact(row, ["generator_id", "level", "level_cap", "next_level_cost", "ppm_per_level"], "fiscal level");
      const levelCap = parseIntCap(row.level_cap, "fiscal level cap");
      const level = integer(row.level, 0, levelCap.amount); integer(row.ppm_per_level, 0);
      if (row.next_level_cost === null ? level < levelCap.amount : level >= levelCap.amount || integer(row.next_level_cost, 0) < 0) throw new SyntaxError("fiscal level cost disagrees with its cap");
    }
    for (const row of sortedRows(arm.unlocks, "unlock_id", "fiscal unlocks")) {
      exact(row, ["cost", "owned", "unlock_id"], "fiscal unlock");
      integer(row.cost, 0); bool(row.owned, "unlock owned");
    }
  }
  if (features.minigames !== null) {
    const arm = object(features.minigames, "minigames arm");
    exact(arm, ["rows"], "minigames arm");
    for (const row of sortedRows(arm.rows, "minigame_id", "minigame rows")) {
      exact(row, ["active_session", "human_content_locked", "minigame_id", "unlocked"], "minigame row");
      bool(row.active_session, "active session"); bool(row.human_content_locked, "human content lock"); bool(row.unlocked, "unlocked");
    }
  }
  if (features.reputation !== undefined && features.reputation !== null) parseReputationArm(features.reputation);
  if (features.pet_adoption !== undefined && features.pet_adoption !== null) parsePetAdoptionArm(features.pet_adoption);
  if (features.cosmetics !== undefined && features.cosmetics !== null) parseCosmeticsArm(features.cosmetics);
  if (features.axis_stack !== undefined && features.axis_stack !== null) parseAxisStackArm(features.axis_stack);
  if (features.opportunity !== undefined && features.opportunity !== null) parseOpportunityArm(features.opportunity);
  return features as unknown as GameUIFeatures;
}

const canonicalDecimal = { test: (value: string): boolean => { try { parseCanonical(value); return true; } catch { return false; } } };

// Clout v1 CV9: the decoder rejects contradictions — a saturation flag that
// disagrees with input/cap, a contribution for an unowned intern or with a
// factor differing from the intern's, and unsorted rows.
export function parseAxisStackArm(source: unknown): void {
  const arm = object(source, "axis stack arm");
  exact(arm, ["attained", "cap_reason_key", "contributions", "input_cap", "input_kind", "input_value", "interns", "product", "saturated"], "axis stack arm");
  if (arm.input_kind !== "achievement_attainment_run" && arm.input_kind !== "achievement_score_run") throw new SyntaxError("invalid axis input kind");
  const cap = integer(arm.input_cap, 1), value = integer(arm.input_value, 0);
  identifier(arm.cap_reason_key);
  if (typeof arm.saturated !== "boolean" || arm.saturated !== value > cap) throw new SyntaxError("axis saturation contradicts input and cap");
  if (typeof arm.product !== "string" || !canonicalDecimal.test(arm.product)) throw new SyntaxError("axis product must be canonical");
  if (!Array.isArray(arm.interns) || !Array.isArray(arm.contributions) || !Array.isArray(arm.attained)) throw new SyntaxError("axis rows must be arrays");
  const interns = new Map<string, { owned: boolean; factor: string }>();
  let prior = "";
  for (const [index, value] of arm.interns.entries()) {
    const row = object(value, `axis intern ${index}`);
    exact(row, ["factor", "factor_ppm", "minimum", "owned", "upgrade_id"], "axis intern");
    const id = identifier(row.upgrade_id);
    if (id <= prior || typeof row.factor !== "string" || !canonicalDecimal.test(row.factor)) throw new SyntaxError("axis interns must be sorted with canonical factors");
    integer(row.factor_ppm, 1, 1_000_000); integer(row.minimum, 0); bool(row.owned, "axis intern owned");
    interns.set(id, { owned: row.owned as boolean, factor: row.factor });
    prior = id;
  }
  prior = "";
  for (const [index, value] of arm.contributions.entries()) {
    const row = object(value, `axis contribution ${index}`);
    exact(row, ["factor", "source_id", "upgrade_id"], "axis contribution");
    const source = identifier(row.source_id), intern = interns.get(identifier(row.upgrade_id));
    if (source <= prior || !intern || !intern.owned || row.factor !== intern.factor) throw new SyntaxError("axis contribution contradicts its intern");
    prior = source;
  }
  if (arm.contributions.length !== [...interns.values()].filter((row) => row.owned).length) throw new SyntaxError("every owned intern contributes exactly once");
  prior = "";
  for (const [index, value] of arm.attained.entries()) {
    const row = object(value, `axis attained ${index}`);
    exact(row, ["achievement_id", "earned_this_run"], "axis attained");
    const id = identifier(row.achievement_id);
    if (id <= prior) throw new SyntaxError("axis attained rows must be sorted");
    bool(row.earned_this_run, "earned this run");
    prior = id;
  }
}

const uuidV7 = /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u;

// Cosmetic Shop v1 §7.1: the decoder rejects internal contradictions, e.g. an
// owned item that is still acquirable, or a worn value absent from worn_by.
export function parseCosmeticsArm(source: unknown): void {
  const arm = object(source, "cosmetics arm");
  exact(arm, ["active", "items", "wearers"], "cosmetics arm");
  if (typeof arm.active !== "boolean" || !Array.isArray(arm.items) || !Array.isArray(arm.wearers)) throw new SyntaxError("invalid cosmetics arm");
  if (!arm.active && (arm.items.length !== 0 || arm.wearers.length !== 0)) throw new SyntaxError("inactive cosmetics must be empty");
  const wornBy = new Map<string, string[]>();
  let prior = "";
  for (const [index, value] of arm.items.entries()) {
    const row = object(value, `cosmetic ${index}`);
    exact(row, ["acquirable", "cosmetic_id", "lock", "owned", "worn_by"], "cosmetic item");
    identifier(row.cosmetic_id);
    if (typeof row.owned !== "boolean" || typeof row.acquirable !== "boolean") throw new SyntaxError("cosmetic flags must be boolean");
    if (row.owned && row.acquirable) throw new SyntaxError("an owned cosmetic cannot be acquirable");
    if (row.lock !== null) {
      const lock = object(row.lock, "cosmetic lock");
      exact(lock, ["kind", "tier"], "cosmetic lock");
      oneOf(lock.kind, ["active_company_tier_at_least"] as const, "cosmetic lock kind"); integer(lock.tier, 0);
      if (row.acquirable || row.owned) throw new SyntaxError("a locked cosmetic cannot be owned or acquirable");
    }
    if (!Array.isArray(row.worn_by) || row.worn_by.some((pet) => typeof pet !== "string" || !uuidV7.test(pet))) throw new SyntaxError("cosmetic worn_by must be pet ids");
    if (row.worn_by.length !== 0 && !row.owned) throw new SyntaxError("an unowned cosmetic cannot be worn");
    if ((row.cosmetic_id as string) <= prior) throw new SyntaxError("cosmetic items must follow catalog order");
    prior = row.cosmetic_id as string;
    wornBy.set(row.cosmetic_id as string, row.worn_by as string[]);
  }
  let priorPet = "";
  for (const [index, value] of arm.wearers.entries()) {
    const row = object(value, `cosmetic wearer ${index}`);
    exact(row, ["pet_id", "worn"], "cosmetic wearer");
    if (typeof row.pet_id !== "string" || !uuidV7.test(row.pet_id) || row.pet_id <= priorPet) throw new SyntaxError("cosmetic wearers must be sorted pet ids");
    priorPet = row.pet_id;
    if (row.worn !== null && !(wornBy.get(identifier(row.worn)) ?? []).includes(row.pet_id)) throw new SyntaxError("a worn cosmetic is absent from its worn_by");
  }
  for (const [id, pets] of wornBy) for (const pet of pets) {
    const wearer = (arm.wearers as { pet_id: string; worn: string | null }[]).find((row) => row.pet_id === pet);
    if (!wearer || wearer.worn !== id) throw new SyntaxError("worn_by names a pet that does not wear the item");
  }
}

function nullableIdentifier(value: unknown): void { if (value !== null) identifier(value); }

// GS5: the pending opportunity and live buffs are projected only while
// unexpired at attended_now_ms; a row at or past its expiry is a projector
// contradiction and fails closed.
function parseOpportunityArm(source: unknown): void {
  const arm = object(source, "opportunity arm");
  exact(arm, ["attended_now_ms", "buffs", "combo", "pending"], "opportunity arm");
  const now = integer(arm.attended_now_ms, 0);
  const combo = object(arm.combo, "opportunity combo");
  exact(combo, ["cap", "reason_key", "saturated"], "opportunity combo");
  decimal(combo.cap, true); identifier(combo.reason_key); bool(combo.saturated, "combo saturated");
  if (arm.pending !== null) {
    const pending = object(arm.pending, "pending opportunity");
    exact(pending, ["effect_row_id", "expires_attended_ms", "opportunity_id", "selected_generator_id"], "pending opportunity");
    identifier(pending.effect_row_id); nullableIdentifier(pending.selected_generator_id);
    if (typeof pending.opportunity_id !== "string" || !uuidV7.test(pending.opportunity_id)) throw new SyntaxError("opportunity id must be UUIDv7");
    if (integer(pending.expires_attended_ms, 1) <= now) throw new SyntaxError("an expired opportunity must not be projected");
  }
  if (!Array.isArray(arm.buffs)) throw new SyntaxError("buffs must be an array");
  let prior = "";
  for (const [index, value] of arm.buffs.entries()) {
    const row = object(value, `buff ${index}`);
    exact(row, ["buff_instance_id", "effect_row_id", "expires_attended_ms", "selected_target"], "buff row");
    if (typeof row.buff_instance_id !== "string" || !uuidV7.test(row.buff_instance_id) || row.buff_instance_id <= prior) throw new SyntaxError("buffs must be UUIDv7, sorted and unique");
    prior = row.buff_instance_id;
    identifier(row.effect_row_id); nullableIdentifier(row.selected_target);
    if (integer(row.expires_attended_ms, 1) <= now) throw new SyntaxError("an expired buff must not be projected");
  }
}

const petStatusBands = ["floor", "high", "low", "normal"] as const;
const petTemperaments = ["chaotic", "curious", "lazy", "playful", "sassy", "shy"] as const;

// PA7: identities plus band and eligible actions only; any raw care field
// (stats, Trust, remainders, cooldowns, behavior, mood) fails closed.
function parsePetAdoptionArm(source: unknown): void {
  const arm = object(source, "pet adoption arm");
  exact(arm, ["pet_adoption", "pets"], "pet adoption arm");
  const adoption = object(arm.pet_adoption, "pet adoption availability");
  exact(adoption, ["cap", "count", "name_keys", "starter_species_id"], "pet adoption availability");
  const cap = integer(adoption.cap, 1), count = integer(adoption.count, 0);
  if (count > cap) throw new SyntaxError("pet count exceeds its cap");
  if (!Array.isArray(adoption.name_keys) || adoption.name_keys.length === 0) throw new SyntaxError("pet name pool must be non-empty");
  for (const key of adoption.name_keys) identifier(key);
  identifier(adoption.starter_species_id);
  if (!Array.isArray(arm.pets) || arm.pets.length !== count) throw new SyntaxError("pet rows must match the pet count");
  for (const [index, value] of arm.pets.entries()) {
    const row = object(value, `pet ${index}`);
    exact(row, ["eligible_action_ids", "name_key", "palette_id", "pet_id", "species_id", "status_band", "temperament"], "pet row");
    oneOf(row.status_band, petStatusBands, "pet status band"); oneOf(row.temperament, petTemperaments, "pet temperament");
    if (typeof row.pet_id !== "string" || !/^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u.test(row.pet_id)) throw new SyntaxError("pet id must be UUIDv7");
    for (const key of [row.name_key, row.palette_id, row.species_id]) identifier(key);
    if (!Array.isArray(row.eligible_action_ids)) throw new SyntaxError("pet eligible actions must be an array");
    for (const action of row.eligible_action_ids) identifier(action);
  }
}

function parseReputationArm(source: unknown): void {
  const arm = object(source, "reputation arm");
  exact(arm, ["available", "bonus_factor_next_run", "bonus_factor_this_run", "level", "nodes", "per_level_ppm", "spent", "unlock_ppm"], "reputation arm");
  const level = integer(arm.level, 0), spent = integer(arm.spent, 0, level);
  if (integer(arm.available, 0) !== level - spent) throw new SyntaxError("reputation available must equal level - spent");
  integer(arm.per_level_ppm, 1, 1_000_000); integer(arm.unlock_ppm, 0, 1_000_000);
  if (typeof arm.bonus_factor_next_run !== "string" || arm.bonus_factor_this_run !== null && typeof arm.bonus_factor_this_run !== "string") throw new SyntaxError("reputation bonus factors must be canonical strings");
  if (!Array.isArray(arm.nodes)) throw new SyntaxError("reputation nodes must be an array");
  const seen = new Set<string>();
  for (const [index, value] of arm.nodes.entries()) {
    const row = object(value, `reputation node ${index}`);
    exact(row, ["body_key", "cost", "kind", "node_id", "requires", "state", "title_key"], "reputation node");
    const id = identifier(row.node_id); identifier(row.title_key); identifier(row.body_key); integer(row.cost, 1);
    oneOf(row.kind, ["bonus_unlock", "starter"], "reputation node kind"); oneOf(row.state, ["available", "locked", "owned", "unaffordable"], "reputation node state");
    if (!Array.isArray(row.requires) || row.requires.some((requirement) => typeof requirement !== "string" || !seen.has(requirement))) throw new SyntaxError("reputation requires must name earlier nodes");
    if (seen.has(id)) throw new SyntaxError("duplicate reputation node");
    seen.add(id);
  }
}

export function toShellSnapshot(snapshot: ParsedGameUISnapshot): AuthoritativeSnapshot {
  return {
    revision: snapshot.revision,
    evaluatedThroughMs: snapshot.evaluated_through_ms,
    constantsHash: snapshot.constants_hash,
    resources: Object.fromEntries(snapshot.resources.map((row) => [row.resource_id, {
      amount: row.amount,
      ratePerSecond: row.rate_per_second,
      ...(row.cap === null ? {} : { cap: { amount: row.cap.amount, reasonKey: row.cap.reason_key } }),
    }])),
    discrete: Object.fromEntries(snapshot.facts.map((row) => [row.fact_id, row.value as DiscreteFact])),
    progress: snapshot.progress.map((row) => ({ stageId: row.stage_id, current: row.current, target: row.target })),
  };
}

export function eraForSnapshot(snapshot: ParsedGameUISnapshot): "era_1995" | "era_2000" | "era_2010" {
  if (snapshot.run.tier === 0) return "era_1995";
  if (snapshot.run.tier === 1) return "era_2000";
  // Tier 2 is the IT Company era (rfc/tier2-content.md §E1); tier >= 3 stays fail-closed.
  if (snapshot.run.tier === 2) return "era_2010";
  throw new RangeError(`tier ${snapshot.run.tier} has no shipped UI era`);
}
