import Decimal from "break_infinity.js";

import type { CurriculumCatalog, CurriculumStarter } from "./curriculum";
import type { EconomyCatalog } from "./economy-kernel";
import { canonicalString, isStateValue, MAX_EXACT_INTEGER, parseCanonical, quantize } from "./numeric";

// Reputation tree artifact loader and pure accounting/bonus arithmetic
// (rfc/reputation-tree-v1.md R1–R3), byte-parallel to server/reputation.

export const REPUTATION_TREE_SCHEMA_VERSION = 1 as const;
export const REPUTATION_MAX_NODES = 64;
export const REPUTATION_PROVIDER = "reputation_tree";
const PPM = 1_000_000;
const idPattern = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;

export type ReputationNode =
  | Readonly<{ node_id: string; kind: "bonus_unlock"; cost: number; requires: readonly string[]; unlock_ppm: number; title_key: string; body_key: string }>
  | Readonly<{ node_id: string; kind: "starter"; cost: number; requires: readonly string[]; starter: CurriculumStarter; title_key: string; body_key: string }>;

export interface ReputationTree {
  readonly bonus: Readonly<{ per_level_ppm: number; source_id: string; slot: "prestige"; target: "all" }>;
  readonly nodes: readonly ReputationNode[];
}

export interface ReputationDeclarations {
  readonly economy: EconomyCatalog;
  readonly curriculum?: CurriculumCatalog;
  readonly copyKeys: ReadonlySet<string>;
}

function reject(rule: number, message: string): never { throw new SyntaxError(`invalid reputation tree: rule ${rule}: ${message}`); }

export function loadReputationTree(source: unknown, declarations: ReputationDeclarations): ReputationTree {
  if (declarations.copyKeys.size === 0) reject(0, "missing declarations");
  const root = exact(source, ["schema_version", "bonus", "nodes"], 0);
  if (root.schema_version !== REPUTATION_TREE_SCHEMA_VERSION) reject(1, "schema");
  const bonusRaw = exact(root.bonus, ["per_level_ppm", "source_id", "slot", "target"], 1);
  const perLevel = bonusRaw.per_level_ppm;
  if (!Number.isSafeInteger(perLevel) || (perLevel as number) < 1 || (perLevel as number) > PPM || typeof bonusRaw.source_id !== "string" || !idPattern.test(bonusRaw.source_id) ||
      bonusRaw.slot !== "prestige" || bonusRaw.target !== "all" ||
      !declarations.economy.multiplierSources.some((row) => row.id === bonusRaw.source_id && row.slot === "prestige" && row.target === "all" && row.provider === REPUTATION_PROVIDER)) reject(1, "bonus declaration");
  if (!Array.isArray(root.nodes) || root.nodes.length < 1 || root.nodes.length > REPUTATION_MAX_NODES) reject(2, "node count");
  const nodes: ReputationNode[] = [];
  const seen = new Set<string>();
  let previousBonus: Extract<ReputationNode, { kind: "bonus_unlock" }> | undefined;
  for (const [index, value] of (root.nodes as unknown[]).entries()) {
    if (!isRecord(value)) reject(0, `nodes[${index}]`);
    const kind = value.kind;
    const keys = kind === "bonus_unlock" ? ["node_id", "kind", "cost", "requires", "unlock_ppm", "title_key", "body_key"]
      : kind === "starter" ? ["node_id", "kind", "cost", "requires", "starter", "title_key", "body_key"] : reject(0, `nodes[${index}] kind`);
    const raw = exact(value, keys, 0);
    const id = raw.node_id;
    if (typeof id !== "string" || !idPattern.test(id) || !id.startsWith("reputation.")) reject(3, `nodes[${index}] id`);
    if (seen.has(id)) reject(3, `nodes[${index}] duplicate id`);
    if (!Number.isSafeInteger(raw.cost) || (raw.cost as number) < 1 || (raw.cost as number) > MAX_EXACT_INTEGER) reject(4, `nodes[${index}] cost`);
    if (!Array.isArray(raw.requires) || raw.requires.some((entry) => typeof entry !== "string")) reject(0, `nodes[${index}] requires`);
    const requires = raw.requires as string[];
    for (const [position, requirement] of requires.entries()) {
      if (position > 0 && byteCompare(requires[position - 1]!, requirement) >= 0) reject(5, `nodes[${index}] requires unsorted`);
      if (!seen.has(requirement) || requirement === id) reject(5, `nodes[${index}] requires a later or unknown row`);
    }
    let node: ReputationNode;
    if (kind === "bonus_unlock") {
      const unlock = raw.unlock_ppm;
      if (!Number.isSafeInteger(unlock) || (unlock as number) < 1 || (unlock as number) > PPM) reject(6, `nodes[${index}] unlock ppm`);
      if (previousBonus && ((unlock as number) <= previousBonus.unlock_ppm || !requires.includes(previousBonus.node_id))) reject(6, `nodes[${index}] unlock ladder`);
      node = { node_id: id, kind, cost: raw.cost as number, requires: Object.freeze([...requires]), unlock_ppm: unlock as number, title_key: raw.title_key as string, body_key: raw.body_key as string };
      previousBonus = node;
    } else {
      node = { node_id: id, kind: "starter", cost: raw.cost as number, requires: Object.freeze([...requires]), starter: parseStarter(raw.starter, declarations.economy, index), title_key: raw.title_key as string, body_key: raw.body_key as string };
    }
    if (typeof raw.title_key !== "string" || typeof raw.body_key !== "string" || !idPattern.test(raw.title_key) || !idPattern.test(raw.body_key) ||
        !declarations.copyKeys.has(raw.title_key) || !declarations.copyKeys.has(raw.body_key)) reject(8, `nodes[${index}] copy key`);
    seen.add(id);
    nodes.push(Object.freeze(node));
  }
  if (!previousBonus || previousBonus.unlock_ppm !== PPM) reject(6, "unlock ladder must end at 1000000");
  validateHeadroom(nodes, declarations);
  return Object.freeze({ bonus: Object.freeze({ per_level_ppm: perLevel as number, source_id: bonusRaw.source_id as string, slot: "prestige", target: "all" }), nodes: Object.freeze(nodes) });
}

function parseStarter(source: unknown, economy: EconomyCatalog, index: number): CurriculumStarter {
  if (!isRecord(source)) reject(7, `nodes[${index}] starter`);
  if (source.kind === "resource_grant") {
    const item = exact(source, ["kind", "resource_id", "amount"], 7);
    if (typeof item.resource_id !== "string" || typeof item.amount !== "string") reject(7, `nodes[${index}] starter`);
    let amount: Decimal;
    try { amount = parseCanonical(item.amount); } catch { reject(7, `nodes[${index}] starter amount`); }
    if (economy.resource(item.resource_id)?.scope !== "company" || !amount.gt(0) || !isStateValue(amount)) reject(7, `nodes[${index}] starter`);
    return Object.freeze({ kind: "resource_grant", resource_id: item.resource_id, amount: item.amount });
  }
  if (source.kind === "generated_generators") {
    const item = exact(source, ["kind", "generator_id", "count"], 7);
    const generator = typeof item.generator_id === "string" ? economy.generatorClass(item.generator_id) : undefined;
    if (!generator || economy.resource(generator.price.resourceId)?.scope !== "company" || !Number.isSafeInteger(item.count) || (item.count as number) < 1) reject(7, `nodes[${index}] starter`);
    return Object.freeze({ kind: "generated_generators", generator_id: item.generator_id as string, count: item.count as number });
  }
  if (source.kind === "preowned_upgrade") {
    const item = exact(source, ["kind", "upgrade_id"], 7);
    const upgrade = typeof item.upgrade_id === "string" ? economy.upgrade(item.upgrade_id) : undefined;
    if (!upgrade || economy.resource(upgrade.cost.resourceId)?.scope !== "company") reject(7, `nodes[${index}] starter`);
    return Object.freeze({ kind: "preowned_upgrade", upgrade_id: item.upgrade_id as string });
  }
  return reject(7, `nodes[${index}] starter kind`);
}

// R2 rule 7: the curriculum's largest grant plus every tree grant never exceeds a cap.
function validateHeadroom(nodes: readonly ReputationNode[], declarations: ReputationDeclarations): void {
  const grants = new Map<string, Decimal>();
  const generated = new Map<string, bigint>();
  const upgrades = new Set<string>();
  for (const node of nodes) {
    if (node.kind !== "starter") continue;
    const starter = node.starter;
    if (starter.kind === "resource_grant") {
      const prior = grants.get(starter.resource_id);
      const amount = parseCanonical(starter.amount);
      grants.set(starter.resource_id, prior ? prior.add(amount) : amount);
    } else if (starter.kind === "generated_generators") {
      generated.set(starter.generator_id, (generated.get(starter.generator_id) ?? 0n) + BigInt(starter.count));
    } else {
      if (upgrades.has(starter.upgrade_id)) reject(7, `duplicate preowned upgrade ${starter.upgrade_id}`);
      upgrades.add(starter.upgrade_id);
    }
  }
  const curriculumGrant = new Map<string, Decimal>();
  const curriculumCount = new Map<string, number>();
  for (const branch of declarations.curriculum?.first_failure.branches ?? []) {
    const starter = branch.starter_package;
    if (starter.kind === "resource_grant") {
      const amount = parseCanonical(starter.amount); const prior = curriculumGrant.get(starter.resource_id);
      if (!prior || amount.gt(prior)) curriculumGrant.set(starter.resource_id, amount);
    } else if (starter.kind === "generated_generators" && starter.count > (curriculumCount.get(starter.generator_id) ?? 0)) curriculumCount.set(starter.generator_id, starter.count);
  }
  for (const resourceId of [...grants.keys()].sort(byteCompare)) {
    const resource = declarations.economy.resource(resourceId)!;
    let total = parseCanonical(resource.initial).add(grants.get(resourceId)!);
    const extra = curriculumGrant.get(resourceId);
    if (extra) total = total.add(extra);
    if (resource.hardcap && total.gt(parseCanonical(resource.hardcap.amount)) || !isStateValue(total)) reject(7, `resource ${resourceId} exceeds its hardcap in aggregate`);
  }
  for (const generatorId of [...generated.keys()].sort(byteCompare)) {
    const generator = declarations.economy.generatorClass(generatorId)!;
    const limit = BigInt(generator.provisionedHardcap?.count ?? MAX_EXACT_INTEGER);
    if (generated.get(generatorId)! + BigInt(curriculumCount.get(generatorId) ?? 0) > limit) reject(7, `generator ${generatorId} exceeds its provisioned hardcap in aggregate`);
  }
}

/** R1: available = level − spent; never persisted. */
export function reputationAvailable(level: number, spent: number): number {
  if (!Number.isSafeInteger(level) || level < 0 || !Number.isSafeInteger(spent) || spent < 0 || spent > level) throw new RangeError("invalid reputation state");
  return level - spent;
}

/** R1: the largest unlock_ppm among owned bonus_unlock nodes known to the tree, or 0 (OD-7). */
export function reputationUnlockPpm(tree: ReputationTree, owned: readonly string[]): number {
  if (!sortedUnique(owned)) throw new RangeError("invalid reputation state");
  let result = 0;
  for (const id of owned) {
    const node = tree.nodes.find((candidate) => candidate.node_id === id);
    if (node?.kind === "bonus_unlock" && node.unlock_ppm > result) result = node.unlock_ppm;
  }
  return result;
}

/** R3: 1 + level × per_level_ppm × unlock_ppm / 1e12, from the earned level (never available). */
export function reputationBonusFactor(level: number, spent: number, perLevelPpm: number, unlockPpm: number): string {
  reputationAvailable(level, spent);
  if (!Number.isSafeInteger(perLevelPpm) || perLevelPpm < 1 || perLevelPpm > PPM || !Number.isSafeInteger(unlockPpm) || unlockPpm < 0 || unlockPpm > PPM) throw new RangeError("invalid reputation state");
  const numerator = BigInt(level) * BigInt(perLevelPpm) * BigInt(unlockPpm);
  const value = quantize(new Decimal(numerator.toString()).div(1e12).add(1));
  if (!isStateValue(value) || value.lt(1)) throw new RangeError("invalid reputation bonus");
  return canonicalString(value);
}

export function sortedUnique(ids: readonly string[]): boolean {
  return ids.every((id, index) => idPattern.test(id) && (index === 0 || byteCompare(ids[index - 1]!, id) < 0));
}

function isRecord(source: unknown): source is Record<string, unknown> { return typeof source === "object" && source !== null && !Array.isArray(source); }
function exact(source: unknown, keys: readonly string[], rule: number): Record<string, unknown> {
  if (!isRecord(source)) reject(rule, "expected an object");
  const actual = Object.keys(source).sort(byteCompare); const expected = [...keys].sort(byteCompare);
  if (actual.length !== expected.length || actual.some((value, index) => value !== expected[index])) reject(rule, "fields are not exact");
  return source;
}
function byteCompare(left: string, right: string): number { const a = new TextEncoder().encode(left); const b = new TextEncoder().encode(right); for (let index = 0; index < Math.min(a.length, b.length); index++) if (a[index] !== b[index]) return a[index]! - b[index]!; return a.length - b.length; }
