import { MAX_EXACT_INTEGER } from "./numeric";

const mechanical = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;

export interface RelevanceCatalogGenerator {
  readonly id: string;
  readonly tier: number;
  readonly category: string;
}

export interface RelevanceCatalogView {
  readonly generators: readonly RelevanceCatalogGenerator[];
  readonly upgradeIds: readonly string[];
}

export interface RelevanceWindow {
  readonly from_gate: string;
  readonly to_gate: string | null;
}

export interface RelevancePolicyItem {
  readonly purchasable_id: string;
  readonly availability_window: RelevanceWindow;
  readonly epsilon_ms: number;
  readonly trap_exempt: boolean;
  readonly justification_key: string | null;
  readonly group_ids: readonly string[];
}

export interface RelevancePolicyGroup {
  readonly group_id: string;
  readonly axis: "tier" | "category" | "declared";
  readonly member_ids: readonly string[];
  readonly epsilon_ms: number;
}

export interface RelevancePolicy {
  readonly schema_version: 1;
  readonly items: readonly RelevancePolicyItem[];
  readonly groups: readonly RelevancePolicyGroup[];
}

export function parseRelevancePolicy(source: unknown, catalog: RelevanceCatalogView, gateIds: readonly string[]): RelevancePolicy {
  const root = exactObject(source, ["schema_version", "items", "groups"], "relevance policy");
  if (root.schema_version !== 1 || !Array.isArray(root.items) || !Array.isArray(root.groups)) throw new SyntaxError("invalid relevance policy envelope");
  const gates = new Map(gateIds.map((id, index) => [mechanicalString(id, "gate id"), index]));
  if (gates.size !== gateIds.length) throw new SyntaxError("duplicate gate ids");
  const purchasables = new Set([...catalog.generators.map((row) => row.id), ...catalog.upgradeIds]);
  let prior = "";
  const items = root.items.map((raw): RelevancePolicyItem => {
    const row = exactObject(raw, ["purchasable_id", "availability_window", "epsilon_ms", "trap_exempt", "justification_key", "group_ids"], "relevance item");
    const id = mechanicalString(row.purchasable_id, "purchasable id");
    if (!purchasables.has(id) || prior !== "" && byteCompare(prior, id) >= 0) throw new SyntaxError("incomplete or unsorted relevance items");
    prior = id;
    const window = exactObject(row.availability_window, ["from_gate", "to_gate"], "availability window");
    const from = mechanicalString(window.from_gate, "from gate");
    const to = window.to_gate === null ? null : mechanicalString(window.to_gate, "to gate");
    if (!gates.has(from) || to !== null && (!gates.has(to) || (gates.get(from) as number) >= (gates.get(to) as number))) throw new SyntaxError("invalid relevance window");
    if (typeof row.trap_exempt !== "boolean" || row.trap_exempt !== (row.justification_key !== null)) throw new SyntaxError("invalid trap exemption");
    const justification = row.justification_key === null ? null : mechanicalString(row.justification_key, "justification key");
    return Object.freeze({ purchasable_id: id, availability_window: Object.freeze({ from_gate: from, to_gate: to }),
      epsilon_ms: safePositive(row.epsilon_ms, "item epsilon"), trap_exempt: row.trap_exempt,
      justification_key: justification, group_ids: sortedMechanical(row.group_ids, "item groups") });
  });
  if (items.length !== purchasables.size || items.some((item) => !purchasables.has(item.purchasable_id))) throw new SyntaxError("incomplete relevance item set");
  const itemById = new Map(items.map((item) => [item.purchasable_id, item]));
  prior = "";
  const groups = root.groups.map((raw): RelevancePolicyGroup => {
    const row = exactObject(raw, ["group_id", "axis", "member_ids", "epsilon_ms"], "relevance group");
    const id = mechanicalString(row.group_id, "group id");
    if (prior !== "" && byteCompare(prior, id) >= 0 || !new Set(["tier", "category", "declared"]).has(String(row.axis))) throw new SyntaxError("invalid relevance group");
    prior = id;
    const members = sortedMechanical(row.member_ids, "group members");
    if (members.length === 0 || members.some((member) => !itemById.get(member)?.group_ids.includes(id))) throw new SyntaxError("dangling/asymmetric group member");
    return Object.freeze({ group_id: id, axis: row.axis as RelevancePolicyGroup["axis"], member_ids: members, epsilon_ms: safePositive(row.epsilon_ms, "group epsilon") });
  });
  const groupById = new Map(groups.map((group) => [group.group_id, group]));
  const axes = new Map<string, Set<string>>();
  for (const item of items) {
    for (const groupId of item.group_ids) {
      const group = groupById.get(groupId);
      if (!group || !group.member_ids.includes(item.purchasable_id)) throw new SyntaxError("dangling/asymmetric item group");
      const seen = axes.get(item.purchasable_id) ?? new Set<string>();
      if (seen.has(group.axis)) throw new SyntaxError("multiple groups on one axis");
      seen.add(group.axis); axes.set(item.purchasable_id, seen);
    }
  }
  for (const group of groups) {
    if (group.axis === "declared") continue;
    const first = catalog.generators.find((row) => row.id === group.member_ids[0]);
    if (!first) throw new SyntaxError("derived group contains non-generator");
    const expected = catalog.generators.filter((row) => group.axis === "tier" ? row.tier === first.tier : row.category === first.category).map((row) => row.id).sort(byteCompare);
    if (!sameStrings(expected, group.member_ids)) throw new SyntaxError("incomplete derived group");
  }
  return Object.freeze({ schema_version: 1, items: Object.freeze(items), groups: Object.freeze(groups) });
}

function exactObject(source: unknown, keys: readonly string[], label: string): Record<string, unknown> {
  if (typeof source !== "object" || source === null || Array.isArray(source)) throw new SyntaxError(`${label} must be an object`);
  const value = source as Record<string, unknown>;
  const actual = Object.keys(value).sort(byteCompare); const expected = [...keys].sort(byteCompare);
  if (!sameStrings(actual, expected)) throw new SyntaxError(`${label} fields are not exact`);
  return value;
}

function mechanicalString(source: unknown, label: string): string {
  if (typeof source !== "string" || !mechanical.test(source)) throw new SyntaxError(`invalid ${label}`);
  return source;
}

function sortedMechanical(source: unknown, label: string): readonly string[] {
  if (!Array.isArray(source)) throw new SyntaxError(`${label} must be an array`);
  let prior = "";
  return Object.freeze(source.map((value) => {
    const id = mechanicalString(value, label);
    if (prior !== "" && byteCompare(prior, id) >= 0) throw new SyntaxError(`${label} must be byte sorted unique`);
    prior = id; return id;
  }));
}

function safePositive(source: unknown, label: string): number {
  if (typeof source !== "number" || !Number.isSafeInteger(source) || source < 1 || source > MAX_EXACT_INTEGER) throw new SyntaxError(`invalid ${label}`);
  return source;
}

function byteCompare(left: string, right: string): number {
  const a = new TextEncoder().encode(left); const b = new TextEncoder().encode(right);
  const length = Math.min(a.length, b.length);
  for (let index = 0; index < length; index++) if (a[index] !== b[index]) return (a[index] as number) - (b[index] as number);
  return a.length - b.length;
}

function sameStrings(left: readonly string[], right: readonly string[]): boolean {
  return left.length === right.length && left.every((value, index) => value === right[index]);
}
