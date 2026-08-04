import { parseCanonical } from "../numeric";

export const ACHIEVEMENT_CATALOG_SCHEMA_VERSION = 1 as const;
export const MAX_EXACT_INTEGER = 9_007_199_254_740_991 as const;

export type AchievementScope = "run" | "career";
export type AchievementCondition =
  | Readonly<{ kind: "fact_present"; factKind: string }>
  | Readonly<{ kind: "counter_at_least"; counter: string; minimum: number }>
  | Readonly<{ kind: "exit_count_at_least"; count: number }>
  | Readonly<{ kind: "owns_generator_at_least"; generatorId: string; count: number }>
  | Readonly<{ kind: "all_of"; conditions: readonly AchievementCondition[] }>;

export type AchievementProof =
  | Readonly<{ kind: "provenance"; eventKinds: readonly string[] }>
  | Readonly<{ kind: "burn"; eventKind: string; resourceId: string; minimum: string }>
  | Readonly<{ kind: "possession"; justificationCopyKey: string }>;

export interface AchievementDefinition {
  readonly id: string;
  readonly conditionScope: AchievementScope;
  readonly condition: AchievementCondition;
  readonly proof: AchievementProof;
  readonly scoreGrant: number;
  readonly copyKey: string;
}

export interface AchievementCatalog {
  readonly schemaVersion: 1;
  readonly definitions: readonly AchievementDefinition[];
  readonly byId: ReadonlyMap<string, AchievementDefinition>;
}

export interface AchievementRegistry {
  readonly copyKeys: ReadonlySet<string>;
  readonly generatorIds: ReadonlySet<string>;
  readonly eventKinds: ReadonlySet<string>;
  readonly resourceIds: ReadonlySet<string>;
  readonly runCounters: ReadonlySet<string>;
  readonly careerCounters: ReadonlySet<string>;
  readonly provenanceSources: ReadonlyMap<string, readonly string[]>;
}

const mechanicalId = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;
const maximumConditionDepth = 4;
const maximumConditionNodes = 64;

export function loadAchievementCatalog(bytes: string | Uint8Array, registry: AchievementRegistry): AchievementCatalog {
  validateRegistry(registry);
  const source = typeof bytes === "string" ? bytes : new TextDecoder("utf-8", { fatal: true }).decode(bytes);
  let parsed: unknown;
  try { parsed = JSON.parse(source); } catch (error) { syntax(`achievement catalog is invalid JSON: ${String(error)}`); }
  const root = exactObject(parsed, ["schema_version", "achievements"], "achievement catalog");
  if (root.schema_version !== ACHIEVEMENT_CATALOG_SCHEMA_VERSION || !Array.isArray(root.achievements)) syntax("achievement catalog has invalid root");
  let lastId = "";
  const definitions = root.achievements.map((raw, index): AchievementDefinition => {
    const label = `achievement catalog.achievements[${index}]`;
    const row = exactObject(raw, ["id", "condition_scope", "condition", "proof", "score_grant", "copy_key"], label);
    const id = identifier(row.id, `${label}.id`);
    if (id <= lastId || row.condition_scope !== "run" && row.condition_scope !== "career" || !registry.copyKeys.has(identifier(row.copy_key, `${label}.copy_key`))) syntax(`${label} is invalid`);
    lastId = id;
    const scope = row.condition_scope;
    const parsedCondition = parseCondition(row.condition, scope, registry, 1);
    if (parsedCondition.nodes > maximumConditionNodes) syntax(`${label}.condition is too large`);
    return Object.freeze({ id, conditionScope: scope, condition: parsedCondition.condition, proof: parseProof(row.proof, scope, parsedCondition.condition, registry), scoreGrant: integer(row.score_grant, 1, MAX_EXACT_INTEGER, `${label}.score_grant`), copyKey: row.copy_key as string });
  });
  return Object.freeze({ schemaVersion: 1, definitions: Object.freeze(definitions), byId: new Map(definitions.map((value) => [value.id, value])) });
}

function parseCondition(value: unknown, scope: AchievementScope, registry: AchievementRegistry, depth: number): { condition: AchievementCondition; nodes: number } {
  if (depth > maximumConditionDepth || value === null || typeof value !== "object" || Array.isArray(value)) syntax("achievement condition is invalid");
  const kind = (value as Record<string, unknown>).kind;
  if (kind === "fact_present") {
    const row = exactObject(value, ["kind", "fact_kind"], "fact_present");
    return { condition: Object.freeze({ kind, factKind: identifier(row.fact_kind, "fact_present.fact_kind") }), nodes: 1 };
  }
  if (kind === "counter_at_least") {
    const row = exactObject(value, ["kind", "counter", "minimum"], "counter_at_least");
    const counter = identifier(row.counter, "counter_at_least.counter");
    if (!(scope === "run" ? registry.runCounters : registry.careerCounters).has(counter)) syntax("counter is not registered for its scope");
    return { condition: Object.freeze({ kind, counter, minimum: integer(row.minimum, 0, MAX_EXACT_INTEGER, "counter_at_least.minimum") }), nodes: 1 };
  }
  if (kind === "exit_count_at_least") {
    const row = exactObject(value, ["kind", "count"], "exit_count_at_least");
    if (scope !== "career") syntax("exit_count_at_least requires career scope");
    return { condition: Object.freeze({ kind, count: integer(row.count, 1, MAX_EXACT_INTEGER, "exit_count_at_least.count") }), nodes: 1 };
  }
  if (kind === "owns_generator_at_least") {
    const row = exactObject(value, ["kind", "generator_id", "count"], "owns_generator_at_least");
    const generatorId = identifier(row.generator_id, "owns_generator_at_least.generator_id");
    if (scope !== "run" || !registry.generatorIds.has(generatorId)) syntax("owns_generator_at_least requires a registered run generator");
    return { condition: Object.freeze({ kind, generatorId, count: integer(row.count, 1, MAX_EXACT_INTEGER, "owns_generator_at_least.count") }), nodes: 1 };
  }
  if (kind === "all_of") {
    const row = exactObject(value, ["kind", "conditions"], "all_of");
    if (!Array.isArray(row.conditions) || row.conditions.length < 2 || row.conditions.length > 16) syntax("all_of conditions are invalid");
    let nodes = 1;
    const conditions = row.conditions.map((child) => { const result = parseCondition(child, scope, registry, depth + 1); nodes += result.nodes; return result.condition; });
    return { condition: Object.freeze({ kind, conditions: Object.freeze(conditions) }), nodes };
  }
  return syntax("achievement condition kind is invalid");
}

function parseProof(value: unknown, scope: AchievementScope, condition: AchievementCondition, registry: AchievementRegistry): AchievementProof {
  if (value === null || typeof value !== "object" || Array.isArray(value)) syntax("achievement proof is invalid");
  const kind = (value as Record<string, unknown>).kind;
  if (kind === "provenance") {
    const row = exactObject(value, ["kind", "event_kinds"], "provenance proof");
    if (!Array.isArray(row.event_kinds) || row.event_kinds.length === 0 || containsKind(condition, "owns_generator_at_least")) syntax("provenance proof is incompatible");
    const eventKinds = row.event_kinds.map((item) => identifier(item, "provenance.event_kinds"));
    if (eventKinds.some((item, index) => !registry.eventKinds.has(item) || index > 0 && eventKinds[index - 1]! >= item)) syntax("provenance event kinds are invalid");
    const declared = new Set(eventKinds);
    for (const source of provenanceSourceKeys(condition, scope)) {
      const required = registry.provenanceSources.get(source);
      if (!required || required.length === 0 || required.some((eventKind) => !declared.has(eventKind))) syntax("provenance proof does not derive its condition");
    }
    return Object.freeze({ kind, eventKinds: Object.freeze(eventKinds) });
  }
  if (kind === "burn") {
    const row = exactObject(value, ["kind", "event_kind", "resource_id", "minimum"], "burn proof");
    const eventKind = identifier(row.event_kind, "burn.event_kind");
    const resourceId = identifier(row.resource_id, "burn.resource_id");
    if (scope !== "run" || !registry.eventKinds.has(eventKind) || !registry.resourceIds.has(resourceId) || typeof row.minimum !== "string") syntax("burn proof is incompatible");
    const minimum = parseCanonical(row.minimum);
    if (!minimum.gt(0)) syntax("burn minimum must be positive");
    return Object.freeze({ kind, eventKind, resourceId, minimum: minimum.toString() });
  }
  if (kind === "possession") {
    const row = exactObject(value, ["kind", "justification_copy_key"], "possession proof");
    const justificationCopyKey = identifier(row.justification_copy_key, "possession.justification_copy_key");
    if (scope !== "run" || !containsKind(condition, "owns_generator_at_least") || !registry.copyKeys.has(justificationCopyKey)) syntax("possession proof is incompatible");
    return Object.freeze({ kind, justificationCopyKey });
  }
  return syntax("achievement proof kind is invalid");
}

function containsKind(condition: AchievementCondition, kind: AchievementCondition["kind"]): boolean {
  return condition.kind === kind || condition.kind === "all_of" && condition.conditions.some((child) => containsKind(child, kind));
}

function provenanceSourceKeys(condition: AchievementCondition, scope: AchievementScope): readonly string[] {
  switch (condition.kind) {
    case "fact_present": return [`fact:${condition.factKind}`];
    case "counter_at_least": return [`counter:${scope}:${condition.counter}`];
    case "exit_count_at_least": return ["exit_count"];
    case "owns_generator_at_least": return [];
    case "all_of": return condition.conditions.flatMap((child) => provenanceSourceKeys(child, scope)).sort();
  }
}

function validateRegistry(registry: AchievementRegistry): void {
  for (const values of [registry.copyKeys, registry.generatorIds, registry.eventKinds, registry.resourceIds, registry.runCounters, registry.careerCounters]) {
    for (const value of values) identifier(value, "achievement registry key");
  }
  for (const [source, kinds] of registry.provenanceSources) {
    if (!validProvenanceSource(source) || kinds.length === 0 || kinds.some((kind, index) => !registry.eventKinds.has(kind) || !mechanicalId.test(kind) || index > 0 && kinds[index - 1]! >= kind)) syntax("achievement provenance source registry is invalid");
  }
}

function validProvenanceSource(value: string): boolean {
  if (value === "exit_count") return true;
  const parts = value.split(":");
  if (parts.length === 2 && parts[0] === "fact") return mechanicalId.test(parts[1]!);
  return parts.length === 3 && parts[0] === "counter" && (parts[1] === "run" || parts[1] === "career") && mechanicalId.test(parts[2]!);
}

function exactObject(value: unknown, keys: readonly string[], label: string): Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value)) syntax(`${label} must be an object`);
  const actual = Object.keys(value).sort(); const expected = [...keys].sort();
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) syntax(`${label} has invalid keys`);
  return value as Record<string, unknown>;
}

function identifier(value: unknown, label: string): string {
  if (typeof value !== "string" || !mechanicalId.test(value)) syntax(`${label} is invalid`);
  return value;
}

function integer(value: unknown, minimum: number, maximum: number, label: string): number {
  if (!Number.isSafeInteger(value) || (value as number) < minimum || (value as number) > maximum) syntax(`${label} is invalid`);
  return value as number;
}

function syntax(message: string): never { throw new SyntaxError(message); }
