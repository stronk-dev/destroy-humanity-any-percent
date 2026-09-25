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
    exact(transitions, ["cross_gate", "wind_down"], "game UI transitions");
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
  exact(features, ["achievements", "active_play", "fiscal", "meters", "minigames", "pets"], "game UI features");
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
  return features as unknown as GameUIFeatures;
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

export function eraForSnapshot(snapshot: ParsedGameUISnapshot): "era_1995" | "era_2000" {
  if (snapshot.run.tier === 0) return "era_1995";
  if (snapshot.run.tier === 1) return "era_2000";
  throw new RangeError(`tier ${snapshot.run.tier} has no shipped UI era`);
}
