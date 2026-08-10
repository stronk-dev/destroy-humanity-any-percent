import type { GameUISnapshot } from "../api/generated/types";
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

export function parseGameUISnapshot(source: unknown): GameUISnapshot {
  const root = object(source, "game UI snapshot");
  exact(root, ["constants_hash", "evaluated_through_ms", "facts", "generators", "manual_action", "progress", "resources", "revision", "run", "schema_version", "server_now_ms", "upgrades"], "game UI snapshot");
  if (root.schema_version !== 1 || typeof root.constants_hash !== "string" || !hash.test(root.constants_hash)) throw new SyntaxError("invalid game UI envelope");
  const revision = integer(root.revision, 1);
  const evaluatedThrough = integer(root.evaluated_through_ms, 1);
  const serverNow = integer(root.server_now_ms, evaluatedThrough);

  const facts = sortedRows(root.facts, "fact_id", "game UI facts");
  for (const row of facts) {
    exact(row, ["fact_id", "value"], "game UI fact");
    if (typeof row.value !== "boolean" && typeof row.value !== "string" && !Number.isSafeInteger(row.value)) throw new SyntaxError("invalid game UI fact value");
  }
  const generators = sortedRows(root.generators, "generator_id", "game UI generators");
  for (const row of generators) {
    exact(row, ["generator_id", "max_affordable", "next_cost", "next_cost_resource_id", "owned", "provisioned", "rate_contribution"], "game UI generator");
    integer(row.max_affordable, 0); integer(row.owned, 0); integer(row.provisioned, 0);
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

  const upgrades = sortedRows(root.upgrades, "upgrade_id", "game UI upgrades");
  for (const row of upgrades) {
    exact(row, ["cost_amount", "cost_resource_id", "eligible", "owned", "upgrade_id"], "game UI upgrade");
    decimal(row.cost_amount); identifier(row.cost_resource_id);
    if (typeof row.eligible !== "boolean" || typeof row.owned !== "boolean" || row.owned && row.eligible) throw new SyntaxError("invalid upgrade state");
  }
  return { ...(root as unknown as GameUISnapshot), revision, evaluated_through_ms: evaluatedThrough, server_now_ms: serverNow };
}

export function toShellSnapshot(snapshot: GameUISnapshot): AuthoritativeSnapshot {
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

export function eraForSnapshot(snapshot: GameUISnapshot): "era_1995" | "era_2000" {
  if (snapshot.run.tier === 0) return "era_1995";
  if (snapshot.run.tier === 1) return "era_2000";
  throw new RangeError(`tier ${snapshot.run.tier} has no shipped UI era`);
}
