import { MAX_EXACT_INTEGER, parseCanonical, quantize, canonicalString } from "./numeric";
import type { EconomyCatalog } from "./economy-kernel";

export const ROUTES_SCHEMA_VERSION = 1;
export const ROUTE_CONTEXT_VERSION = 1;

export type RouteCondition =
  | { readonly kind: "resource_at_least" | "resource_at_most"; readonly resourceId: string; readonly value: string }
  | { readonly kind: "meter_band"; readonly meterId: string; readonly min: number; readonly max: number }
  | { readonly kind: "doctrine_is" | "doctrine_is_not"; readonly transition: string; readonly doctrineId: string }
  | { readonly kind: "structure_is"; readonly structureId: string }
  | { readonly kind: "ledger_fact_present"; readonly factKind: string }
  | { readonly kind: "region_trait"; readonly traitId: string };

export type RouteEffect =
  | { readonly kind: "discount"; readonly fraction: string }
  | { readonly kind: "substitute" };

export interface RouteAlternative {
  readonly routeId: string;
  readonly houseName: string;
  readonly active: boolean;
  readonly requiresContextVersion: number;
  readonly exclusionSlot: string;
  readonly exclusionValue: string;
  readonly predicate: readonly RouteCondition[];
  readonly effect: RouteEffect;
}

export interface GateDefinition {
  readonly gateId: string;
  readonly requirement: readonly { readonly resourceId: string; readonly amount: string }[];
  readonly routes: readonly RouteAlternative[];
}

export interface RouteContext {
  readonly contextVersion: number;
  readonly resources: Readonly<Record<string, string>>;
  readonly doctrinesByTransition: Readonly<Record<string, string>>;
  readonly structureId: string;
  readonly ledgerFactKinds: ReadonlySet<string>;
  readonly meterBands: Readonly<Record<string, number>>;
  readonly regionTraits: ReadonlySet<string>;
}

export class RoutesCatalog {
  readonly contextVersion: number;
  readonly depletionDistinctRoutesRequired: number;
  readonly knowledge: Readonly<{ registryFirstBonus: number; founderFirstGrant: number; repeatGrant: number; hintCost: number }>;
  readonly gates: readonly GateDefinition[];
  readonly #gateById: ReadonlyMap<string, GateDefinition>;
  readonly #routeById: ReadonlyMap<string, RouteAlternative>;

  constructor(contextVersion: number, depletion: number, knowledge: RoutesCatalog["knowledge"], gates: readonly GateDefinition[]) {
    this.contextVersion = contextVersion;
    this.depletionDistinctRoutesRequired = depletion;
    this.knowledge = Object.freeze({ ...knowledge });
    this.gates = Object.freeze([...gates]);
    this.#gateById = new Map(gates.map((gate) => [gate.gateId, gate]));
    this.#routeById = new Map(gates.flatMap((gate) => gate.routes.map((route) => [route.routeId, route] as const)));
  }

  gate(id: string): GateDefinition | undefined { return this.#gateById.get(id); }
  route(id: string): RouteAlternative | undefined { return this.#routeById.get(id); }

  maxRoutesPerRun(): number {
    const valuesBySlot = new Map<string, Set<string>>();
    for (const route of this.#routeById.values()) {
      const values = valuesBySlot.get(route.exclusionSlot) ?? new Set<string>();
      values.add(route.exclusionValue);
      valuesBySlot.set(route.exclusionSlot, values);
    }
    const slots = [...valuesBySlot.keys()].sort(byteCompare);
    let maximum = 0;
    const assignment = new Map<string, string>();
    const search = (index: number): void => {
      if (index === slots.length) {
        let count = 0;
        for (const route of this.#routeById.values()) {
          if (assignment.get(route.exclusionSlot) === route.exclusionValue) count++;
        }
        maximum = Math.max(maximum, count);
        return;
      }
      const slot = slots[index]!;
      for (const value of [...valuesBySlot.get(slot)!].sort(byteCompare)) {
        assignment.set(slot, value);
        search(index + 1);
      }
    };
    search(0);
    return maximum;
  }
}

const idPattern = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;
const exclusionPattern = /^(?:structure|doctrine:[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*)$/;

export function parseRoutesCatalog(source: unknown): RoutesCatalog {
  const root = exactObject(source, ["schema_version", "context_version", "depletion_distinct_routes_required", "knowledge", "gates"], "routes catalog");
  if (root.schema_version !== ROUTES_SCHEMA_VERSION || root.context_version !== ROUTE_CONTEXT_VERSION) throw new SyntaxError("unsupported routes schema/context version");
  const depletion = safePositiveInteger(root.depletion_distinct_routes_required, "depletion_distinct_routes_required");
  const knowledgeRaw = exactObject(root.knowledge, ["registry_first_bonus", "founder_first_grant", "repeat_grant", "hint_cost"], "knowledge");
  const knowledge = {
    registryFirstBonus: safePositiveInteger(knowledgeRaw.registry_first_bonus, "registry_first_bonus"),
    founderFirstGrant: safePositiveInteger(knowledgeRaw.founder_first_grant, "founder_first_grant"),
    repeatGrant: safePositiveInteger(knowledgeRaw.repeat_grant, "repeat_grant"),
    hintCost: safePositiveInteger(knowledgeRaw.hint_cost, "hint_cost"),
  };
  if (!Array.isArray(root.gates) || root.gates.length === 0) throw new SyntaxError("gates must be a non-empty array");
  const gateIds = new Set<string>();
  const routeIds = new Set<string>();
  const gates = root.gates.map((source, index) => {
    const raw = exactObject(source, ["gate_id", "requirement", "routes"], `gates[${index}]`);
    const gateId = mechanicalId(raw.gate_id, `gates[${index}].gate_id`);
    if (gateIds.has(gateId)) throw new SyntaxError(`duplicate gate ${gateId}`);
    gateIds.add(gateId);
    if (!Array.isArray(raw.requirement) || raw.requirement.length === 0 || !Array.isArray(raw.routes)) throw new SyntaxError(`gates[${index}] collections must be arrays and requirement must not be empty`);
    const requirementIds = new Set<string>();
    const requirement = raw.requirement.map((item, requirementIndex) => {
      const value = exactObject(item, ["resource_id", "amount"], `requirement[${requirementIndex}]`);
      const resourceId = mechanicalId(value.resource_id, "requirement.resource_id");
      if (requirementIds.has(resourceId)) throw new SyntaxError(`duplicate gate requirement ${resourceId}`);
      requirementIds.add(resourceId);
      const amount = canonicalDecimal(value.amount, "requirement.amount");
      if (!parseCanonical(amount).gt(0)) throw new SyntaxError("requirement amount must be positive");
      return Object.freeze({ resourceId, amount });
    });
    const routes = raw.routes.map((item, routeIndex) => {
      const value = exactObject(item, ["route_id", "house_name", "active", "requires_context_version", "exclusion_slot", "exclusion_value", "predicate", "effect"], `routes[${routeIndex}]`);
      const routeId = mechanicalId(value.route_id, "route_id");
      if (routeIds.has(routeId)) throw new SyntaxError(`duplicate route ${routeId}`);
      routeIds.add(routeId);
      if (typeof value.house_name !== "string" || value.house_name.length === 0 || typeof value.active !== "boolean") throw new SyntaxError("route house_name/active invalid");
      const requiresContextVersion = safePositiveInteger(value.requires_context_version, "requires_context_version");
      if (value.active && requiresContextVersion > ROUTE_CONTEXT_VERSION) throw new SyntaxError(`active route ${routeId} requires unavailable context version`);
      if (typeof value.exclusion_slot !== "string" || !exclusionPattern.test(value.exclusion_slot)) throw new SyntaxError("invalid exclusion_slot");
      const exclusionValue = mechanicalId(value.exclusion_value, "exclusion_value");
      const predicate = parseRoutePredicate(value.predicate);
      if (predicate.some((condition) => condition.kind === "meter_band" || condition.kind === "region_trait") && requiresContextVersion < 2) throw new SyntaxError("meter and region conditions require context version 2");
      const exclusionMatched = value.exclusion_slot === "structure"
        ? predicate.some((condition) => condition.kind === "structure_is" && condition.structureId === exclusionValue)
        : predicate.some((condition) => condition.kind === "doctrine_is" && `doctrine:${condition.transition}` === value.exclusion_slot && condition.doctrineId === exclusionValue);
      if (!exclusionMatched) throw new SyntaxError("exclusion slot/value must match an explicit predicate condition");
      validateRouteChronology(gateId, predicate);
      const effect = parseEffect(value.effect);
      return Object.freeze({ routeId, houseName: value.house_name, active: value.active, requiresContextVersion, exclusionSlot: value.exclusion_slot, exclusionValue, predicate, effect });
    });
    return Object.freeze({ gateId, requirement: Object.freeze(requirement), routes: Object.freeze(routes) });
  });
  const catalog = new RoutesCatalog(ROUTE_CONTEXT_VERSION, depletion, knowledge, gates);
  if (routeIds.size === 0 || catalog.maxRoutesPerRun() >= depletion) throw new SyntaxError("depletion is reachable in one run");
  return catalog;
}

function validateRouteChronology(gateId: string, predicate: readonly RouteCondition[]): void {
  const transitions = predicate
    .filter((condition): condition is Extract<RouteCondition, { kind: "doctrine_is" | "doctrine_is_not" }> => condition.kind === "doctrine_is" || condition.kind === "doctrine_is_not")
    .map((condition) => condition.transition);
  if (transitions.length === 0) return;
  const gateTier = adjacentBoundaryStart(gateId, "gate");
  if (gateTier === undefined) throw new SyntaxError("doctrine-bearing route requires a canonical adjacent tier gate");
  for (const transition of transitions) {
    const transitionTier = adjacentBoundaryStart(transition, "transition");
    if (transitionTier === undefined || gateTier < transitionTier) throw new SyntaxError(`doctrine transition ${transition} occurs after gate ${gateId}`);
  }
}

function adjacentBoundaryStart(value: string, prefix: "gate" | "transition"): number | undefined {
  const match = new RegExp(`^${prefix}\\.t([0-9]+)_to_t([0-9]+)$`).exec(value);
  if (!match) return undefined;
  const from = Number(match[1]);
  const to = Number(match[2]);
  return Number.isSafeInteger(from) && Number.isSafeInteger(to) && to === from + 1 ? from : undefined;
}

export function parseRoutePredicate(source: unknown): readonly RouteCondition[] {
  const values = Array.isArray(source) ? source : [source];
  if (values.length === 0) throw new SyntaxError("predicate must not be empty");
  return Object.freeze(values.map((value, index) => parseCondition(value, index)));
}

function parseCondition(source: unknown, index: number): RouteCondition {
  if (!isRecord(source) || typeof source.kind !== "string") throw new SyntaxError(`predicate[${index}] must have a kind`);
  switch (source.kind) {
    case "resource_at_least":
    case "resource_at_most": {
      const raw = exactObject(source, ["kind", "resource_id", "value"], `predicate[${index}]`);
      return Object.freeze({ kind: source.kind, resourceId: mechanicalId(raw.resource_id, "resource_id"), value: canonicalDecimal(raw.value, "value") });
    }
    case "meter_band": {
      const raw = exactObject(source, ["kind", "meter_id", "min", "max"], `predicate[${index}]`);
      const min = boundedMeter(raw.min, "min");
      const max = boundedMeter(raw.max, "max");
      if (min > max) throw new SyntaxError("meter min exceeds max");
      return Object.freeze({ kind: source.kind, meterId: mechanicalId(raw.meter_id, "meter_id"), min, max });
    }
    case "doctrine_is":
    case "doctrine_is_not": {
      const raw = exactObject(source, ["kind", "transition", "doctrine_id"], `predicate[${index}]`);
      return Object.freeze({ kind: source.kind, transition: mechanicalId(raw.transition, "transition"), doctrineId: mechanicalId(raw.doctrine_id, "doctrine_id") });
    }
    case "structure_is": {
      const raw = exactObject(source, ["kind", "structure_id"], `predicate[${index}]`);
      return Object.freeze({ kind: source.kind, structureId: mechanicalId(raw.structure_id, "structure_id") });
    }
    case "ledger_fact_present": {
      const raw = exactObject(source, ["kind", "fact_kind"], `predicate[${index}]`);
      return Object.freeze({ kind: source.kind, factKind: mechanicalId(raw.fact_kind, "fact_kind") });
    }
    case "region_trait": {
      const raw = exactObject(source, ["kind", "trait_id"], `predicate[${index}]`);
      return Object.freeze({ kind: source.kind, traitId: mechanicalId(raw.trait_id, "trait_id") });
    }
    default: throw new SyntaxError(`unknown route condition ${source.kind}`);
  }
}

function parseEffect(source: unknown): RouteEffect {
  if (!isRecord(source) || source.kind !== "discount" && source.kind !== "substitute") throw new SyntaxError("unknown route effect");
  if (source.kind === "substitute") { exactObject(source, ["kind"], "effect"); return Object.freeze({ kind: "substitute" }); }
  const raw = exactObject(source, ["kind", "fraction"], "effect");
  const fraction = canonicalDecimal(raw.fraction, "fraction");
  const value = parseCanonical(fraction);
  if (!value.gt(0) || !value.lt(1)) throw new SyntaxError("discount fraction must be in (0,1)");
  return Object.freeze({ kind: "discount", fraction });
}

export function evaluatePredicate(predicate: readonly RouteCondition[], context: RouteContext): boolean {
  if (!Number.isSafeInteger(context.contextVersion) || context.contextVersion < 1) throw new RangeError("invalid context version");
  return predicate.every((condition) => {
    switch (condition.kind) {
      case "resource_at_least": { const value = context.resources[condition.resourceId]; return value !== undefined && parseCanonical(value).gte(parseCanonical(condition.value)); }
      case "resource_at_most": { const value = context.resources[condition.resourceId]; return value !== undefined && parseCanonical(value).lte(parseCanonical(condition.value)); }
      case "meter_band": { const value = context.meterBands[condition.meterId]; return value !== undefined && value >= condition.min && value <= condition.max; }
      case "doctrine_is": return context.doctrinesByTransition[condition.transition] === condition.doctrineId;
      case "doctrine_is_not": { const value = context.doctrinesByTransition[condition.transition]; return value !== undefined && value !== condition.doctrineId; }
      case "structure_is": return context.structureId === condition.structureId;
      case "ledger_fact_present": return context.ledgerFactKinds.has(condition.factKind);
      case "region_trait": return context.regionTraits.has(condition.traitId);
    }
  });
}

export function discountedRequirements(gate: GateDefinition, route: RouteAlternative): readonly { resourceId: string; amount: string }[] {
  if (route.effect.kind === "substitute") return Object.freeze([]);
  const fraction = route.effect.fraction;
  return Object.freeze(gate.requirement.map((requirement) => Object.freeze({
    resourceId: requirement.resourceId,
    amount: canonicalString(quantize(parseCanonical(requirement.amount).mul(parseCanonical(fraction)))),
  })));
}

export function validateRouteCatalogResources(catalog: RoutesCatalog, economy: EconomyCatalog): void {
  const validate = (id: string): void => {
    const resource = economy.resource(id);
    if (!resource || resource.scope !== "company") throw new SyntaxError(`routes catalog references unknown company resource ${id}`);
  };
  for (const gate of catalog.gates) {
    for (const requirement of gate.requirement) validate(requirement.resourceId);
    for (const route of gate.routes) {
      for (const condition of route.predicate) {
        if (condition.kind === "resource_at_least" || condition.kind === "resource_at_most") validate(condition.resourceId);
      }
    }
  }
}

function exactObject(source: unknown, keys: readonly string[], label: string): Record<string, unknown> {
  if (!isRecord(source)) throw new SyntaxError(`${label} must be an object`);
  const actual = Object.keys(source).sort(byteCompare);
  const expected = [...keys].sort(byteCompare);
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError(`${label} fields are not exact`);
  return source;
}
function isRecord(source: unknown): source is Record<string, unknown> { return typeof source === "object" && source !== null && !Array.isArray(source); }
function mechanicalId(value: unknown, label: string): string { if (typeof value !== "string" || !idPattern.test(value)) throw new SyntaxError(`${label} must be a mechanical id`); return value; }
function canonicalDecimal(value: unknown, label: string): string { if (typeof value !== "string") throw new SyntaxError(`${label} must be a canonical Decimal string`); parseCanonical(value); return value; }
function safePositiveInteger(value: unknown, label: string): number { if (typeof value !== "number" || !Number.isSafeInteger(value) || value <= 0 || value > MAX_EXACT_INTEGER) throw new SyntaxError(`${label} must be a positive safe integer`); return value; }
function boundedMeter(value: unknown, label: string): number { if (typeof value !== "number" || !Number.isInteger(value) || value < 0 || value > 100) throw new SyntaxError(`${label} must be an integer in 0..100`); return value; }
function byteCompare(left: string, right: string): number { const a = new TextEncoder().encode(left); const b = new TextEncoder().encode(right); for (let i = 0; i < Math.min(a.length, b.length); i++) { if (a[i] !== b[i]) return a[i]! - b[i]!; } return a.length - b.length; }
