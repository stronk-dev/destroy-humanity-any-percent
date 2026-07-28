import Decimal from "break_infinity.js";

import { sumGeometricSeries } from "./economy";
import {
  canonicalString,
  isStateValue,
  MAX_EXACT_INTEGER,
  parseCanonical,
} from "./numeric";

export const CATALOG_SCHEMA_VERSION = 3;

export type Scope = "company" | "founder" | "world" | "guild";

export interface HardcapDefinition {
  readonly amount: string;
  readonly reasonKey: string;
}

export interface ResourceDefinition {
  readonly id: string;
  readonly scope: Scope;
  readonly numericKind: "decimal";
  readonly initial: string;
  readonly minimum: string;
  readonly hardcap: HardcapDefinition | null;
}

export type CostCurve =
  | { readonly kind: "constant" }
  | { readonly kind: "linear"; readonly step: string }
  | { readonly kind: "geometric"; readonly ratio: string };

export interface PriceDefinition {
  readonly resourceId: string;
  readonly base: string;
  readonly curve: CostCurve;
}

export interface ProductionDefinition {
  readonly resourceId: string;
  readonly baseRate: string;
}

export interface GeneratorClassDefinition {
  readonly id: string;
  readonly price: PriceDefinition;
  readonly production: ProductionDefinition | null;
}

export const MULTIPLIER_SLOT_ORDER = [
  "upgrades", "milestones", "faction", "doctrine", "commons", "trust", "event_buffs", "prestige",
] as const;
export type MultiplierSlot = (typeof MULTIPLIER_SLOT_ORDER)[number];
export type ProgressKind = "resource_log" | "count_fraction" | "composite";

export interface ManualActionDefinition {
  readonly id: string;
  readonly output: { readonly resourceId: string; readonly amountPerAction: string };
}

export interface MultiplierSourceDefinition {
  readonly id: string;
  readonly slot: MultiplierSlot;
  readonly target: string;
  readonly provider: string;
}

export interface ProgressTerm {
  readonly weight: string;
  readonly kind: Exclude<ProgressKind, "composite">;
  readonly resourceId?: string;
  readonly target?: string;
  readonly countKey?: "generators.total_owned";
  readonly required?: number;
}

export interface ProgressCoordinateDefinition {
  readonly tier: number;
  readonly kind: ProgressKind;
  readonly resourceId?: string;
  readonly target?: string;
  readonly countKey?: "generators.total_owned";
  readonly required?: number;
  readonly terms?: readonly ProgressTerm[];
}

export interface ManualPolicy {
  readonly refillMilliPerMs: number;
  readonly bucketCapMilli: number;
}

export interface OfflinePolicy {
  readonly efficiency: string;
  readonly accrualCapMs: number;
  readonly bankRatioNumerator: number;
  readonly bankRatioDenominator: number;
  readonly bankCapMs: number;
  readonly burstSpeed: string;
  readonly burstMaxDurationMs: number;
}

const idPattern = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;

export class EconomyCatalog {
  readonly resources: readonly ResourceDefinition[];
  readonly generatorClasses: readonly GeneratorClassDefinition[];
  readonly manualActions: readonly ManualActionDefinition[];
  readonly multiplierSources: readonly MultiplierSourceDefinition[];
  readonly progressCoordinates: readonly ProgressCoordinateDefinition[];
  readonly manualPolicy: ManualPolicy | null;
  readonly offlinePolicy: OfflinePolicy | null;
  readonly #resourceById: ReadonlyMap<string, ResourceDefinition>;
  readonly #generatorById: ReadonlyMap<string, GeneratorClassDefinition>;

  constructor(
    resources: readonly ResourceDefinition[],
    generatorClasses: readonly GeneratorClassDefinition[],
    manualActions: readonly ManualActionDefinition[] = [],
    multiplierSources: readonly MultiplierSourceDefinition[] = [],
    progressCoordinates: readonly ProgressCoordinateDefinition[] = [],
    manualPolicy: ManualPolicy | null = null,
    offlinePolicy: OfflinePolicy | null = null,
  ) {
    this.resources = Object.freeze([...resources]);
    this.generatorClasses = Object.freeze([...generatorClasses]);
    this.manualActions = Object.freeze([...manualActions]);
    this.multiplierSources = Object.freeze([...multiplierSources]);
    this.progressCoordinates = Object.freeze([...progressCoordinates]);
    this.manualPolicy = manualPolicy;
    this.offlinePolicy = offlinePolicy;
    this.#resourceById = new Map(resources.map((definition) => [definition.id, definition]));
    this.#generatorById = new Map(
      generatorClasses.map((definition) => [definition.id, definition]),
    );
  }

  resource(id: string): ResourceDefinition | undefined {
    return this.#resourceById.get(id);
  }

  generatorClass(id: string): GeneratorClassDefinition | undefined {
    return this.#generatorById.get(id);
  }

  bulkCost(generatorId: string, owned: number, count: number): Decimal {
    const generator = this.generatorClass(generatorId);
    if (!generator) throw new RangeError(`unknown generator class: ${generatorId}`);
    return bulkCost(generator.price, owned, count);
  }

  maxAffordable(generatorId: string, cashSource: string | Decimal, owned: number): number {
    const generator = this.generatorClass(generatorId);
    if (!generator) throw new RangeError(`unknown generator class: ${generatorId}`);
    return maxAffordable(generator.price, cashSource, owned);
  }
}

export function parseCatalog(source: unknown): EconomyCatalog {
  if (!isRecord(source)) throw new SyntaxError("catalog must be an object");
  const version = source.schema_version;
  if (version !== 1 && version !== 2 && version !== CATALOG_SCHEMA_VERSION) {
    throw new SyntaxError(`unsupported catalog schema_version: ${String(version)}`);
  }
  const root = exactObject(
    source,
    version === 3
      ? ["schema_version", "resources", "generator_classes", "manual_actions", "multiplier_sources", "progress_coordinates", "manual_policy", "offline_policy"]
      : ["schema_version", "resources", "generator_classes"],
    "catalog",
  );
  if (!Array.isArray(root.resources)) throw new SyntaxError("catalog.resources must be an array");
  if (!Array.isArray(root.generator_classes)) {
    throw new SyntaxError("catalog.generator_classes must be an array");
  }

  const resources = root.resources.map((value, index) => parseResource(value, index));
  ensureUnique(resources, "resource");
  const resourceById = new Map(resources.map((definition) => [definition.id, definition]));

  const generatorClasses = root.generator_classes.map((value, index) =>
    parseGeneratorClass(value, index, version),
  );
  ensureUnique(generatorClasses, "generator class");
  for (const generator of generatorClasses) {
    const priceResource = resourceById.get(generator.price.resourceId);
    if (!priceResource) {
      throw new SyntaxError(
        `generator class ${generator.id} references unknown resource ${generator.price.resourceId}`,
      );
    }
    if (generator.production) {
      const outputResource = resourceById.get(generator.production.resourceId);
      if (!outputResource) {
        throw new SyntaxError(
          `generator class ${generator.id} references unknown production resource ${generator.production.resourceId}`,
        );
      }
      if (priceResource.scope !== outputResource.scope) {
        throw new SyntaxError(
          `generator class ${generator.id} crosses scopes from ${priceResource.scope} to ${outputResource.scope}`,
        );
      }
    }
  }

  if (version < 3) return new EconomyCatalog(resources, generatorClasses);

  if (!Array.isArray(root.manual_actions) || !Array.isArray(root.multiplier_sources) || !Array.isArray(root.progress_coordinates)) {
    throw new SyntaxError("catalog production-engine collections must be arrays");
  }
  const manualActions = root.manual_actions.map((value, index) => parseManualAction(value, index, resourceById));
  ensureUnique(manualActions, "manual action");
  const multiplierSources = root.multiplier_sources.map((value, index) => parseMultiplierSource(value, index, generatorClasses));
  ensureUnique(multiplierSources, "multiplier source");
  const progressCoordinates = root.progress_coordinates.map((value, index) => parseProgressCoordinate(value, index, resourceById));
  const tiers = new Set(progressCoordinates.map((definition) => definition.tier));
  if (tiers.size !== progressCoordinates.length || [0, 1, 2, 3].some((tier) => !tiers.has(tier))) {
    throw new SyntaxError("progress coordinates must contain each tier 0..3 exactly once");
  }
  const manualPolicy = parseManualPolicy(root.manual_policy);
  const offlinePolicy = parseOfflinePolicy(root.offline_policy);
  return new EconomyCatalog(resources, generatorClasses, manualActions, multiplierSources, progressCoordinates, manualPolicy, offlinePolicy);
}

export function bulkCost(price: PriceDefinition, owned: number, count: number): Decimal {
  validateCount(owned, "owned");
  validateCount(count, "count");
  if (count > MAX_EXACT_INTEGER - owned) {
    throw new RangeError("owned + count exceeds the exact-integer domain");
  }
  const base = parsePositive(price.base, "price.base");
  if (count === 0) return new Decimal(0);

  const countValue = new Decimal(count);
  let cost: Decimal;
  switch (price.curve.kind) {
    case "constant":
      cost = base.mul(countValue);
      break;
    case "linear": {
      const step = parseNonNegative(price.curve.step, "price.curve.step");
      const first = base.add(step.mul(owned));
      const triangle = countValue.mul(count - 1).div(2);
      cost = first.mul(countValue).add(step.mul(triangle));
      break;
    }
    case "geometric": {
      const ratio = parseCanonical(price.curve.ratio);
      if (ratio.lt(1)) throw new RangeError("geometric ratio must be at least one");
      cost = sumGeometricSeries(count, base, ratio, owned);
      break;
    }
  }
  if (!isStateValue(cost)) throw new RangeError("cost is outside the finite Decimal domain");
  return cost;
}

export function maxAffordable(
  price: PriceDefinition,
  cashSource: string | Decimal,
  owned: number,
): number {
  validateCount(owned, "owned");
  const cash = cashSource instanceof Decimal ? new Decimal(cashSource) : parseCanonical(cashSource);
  if (!isStateValue(cash) || cash.lt(0)) throw new RangeError("cash must be non-negative state");

  let low = 0;
  let high = MAX_EXACT_INTEGER - owned;
  while (low < high) {
    const middle = low + Math.floor((high - low + 1) / 2);
    let affordable = false;
    try {
      affordable = bulkCost(price, owned, middle).lte(cash);
    } catch {
      affordable = false;
    }
    if (affordable) low = middle;
    else high = middle - 1;
  }
  return low;
}

function parseResource(source: unknown, index: number): ResourceDefinition {
  const path = `catalog.resources[${index}]`;
  const value = exactObject(
    source,
    ["id", "scope", "numeric_kind", "initial", "minimum", "hardcap"],
    path,
  );
  const id = parseId(value.id, `${path}.id`);
  if (!isScope(value.scope)) throw new SyntaxError(`${path}.scope is unsupported`);
  if (value.numeric_kind !== "decimal") {
    throw new SyntaxError(`${path}.numeric_kind is unsupported`);
  }
  const initial = parseCanonicalField(value.initial, `${path}.initial`);
  const minimum = parseCanonicalField(value.minimum, `${path}.minimum`);
  if (parseCanonical(initial).lt(parseCanonical(minimum))) {
    throw new SyntaxError(`${path}.initial is below minimum`);
  }

  let hardcap: HardcapDefinition | null = null;
  if (value.hardcap !== null) {
    const capValue = exactObject(value.hardcap, ["amount", "reason_key"], `${path}.hardcap`);
    const amount = parseCanonicalField(capValue.amount, `${path}.hardcap.amount`);
    const reasonKey = parseId(capValue.reason_key, `${path}.hardcap.reason_key`);
    if (parseCanonical(amount).lt(parseCanonical(minimum))) {
      throw new SyntaxError(`${path}.hardcap is below minimum`);
    }
    if (parseCanonical(initial).gt(parseCanonical(amount))) {
      throw new SyntaxError(`${path}.initial exceeds hardcap`);
    }
    hardcap = Object.freeze({ amount, reasonKey });
  }

  return Object.freeze({
    id,
    scope: value.scope,
    numericKind: "decimal",
    initial,
    minimum,
    hardcap,
  });
}

function parseGeneratorClass(
  source: unknown,
  index: number,
  schemaVersion: unknown,
): GeneratorClassDefinition {
  const path = `catalog.generator_classes[${index}]`;
  const keys = schemaVersion === 1 ? ["id", "price"] : ["id", "price", "production"];
  const value = exactObject(source, keys, path);
  const id = parseId(value.id, `${path}.id`);
  const priceValue = exactObject(value.price, ["resource_id", "base", "curve"], `${path}.price`);
  const resourceId = parseId(priceValue.resource_id, `${path}.price.resource_id`);
  const base = parseCanonicalField(priceValue.base, `${path}.price.base`);
  if (!parseCanonical(base).gt(0)) throw new SyntaxError(`${path}.price.base must be positive`);
  const curve = parseCurve(priceValue.curve, `${path}.price.curve`);
  const price = Object.freeze({ resourceId, base, curve });
  if (schemaVersion === 1) return Object.freeze({ id, price, production: null });

  const productionValue = exactObject(
    value.production,
    ["resource_id", "base_rate"],
    `${path}.production`,
  );
  const productionResourceId = parseId(
    productionValue.resource_id,
    `${path}.production.resource_id`,
  );
  const baseRate = parseCanonicalField(
    productionValue.base_rate,
    `${path}.production.base_rate`,
  );
  if (!parseCanonical(baseRate).gt(0)) {
    throw new SyntaxError(`${path}.production.base_rate must be positive`);
  }
  const production = Object.freeze({ resourceId: productionResourceId, baseRate });
  return Object.freeze({ id, price, production });
}

function parseManualAction(
  source: unknown,
  index: number,
  resources: ReadonlyMap<string, ResourceDefinition>,
): ManualActionDefinition {
  const path = `catalog.manual_actions[${index}]`;
  const value = exactObject(source, ["id", "output"], path);
  const id = parseId(value.id, `${path}.id`);
  const output = exactObject(value.output, ["resource_id", "amount_per_action"], `${path}.output`);
  const resourceId = parseId(output.resource_id, `${path}.output.resource_id`);
  if (resources.get(resourceId)?.scope !== "company") {
    throw new SyntaxError(`${path}.output.resource_id must reference a company resource`);
  }
  const amountPerAction = parseCanonicalField(output.amount_per_action, `${path}.output.amount_per_action`);
  if (!parseCanonical(amountPerAction).gt(0)) throw new SyntaxError(`${path}.output amount must be positive`);
  return Object.freeze({ id, output: Object.freeze({ resourceId, amountPerAction }) });
}

function parseMultiplierSource(
  source: unknown,
  index: number,
  generators: readonly GeneratorClassDefinition[],
): MultiplierSourceDefinition {
  const path = `catalog.multiplier_sources[${index}]`;
  const value = exactObject(source, ["id", "slot", "target", "provider"], path);
  const id = parseId(value.id, `${path}.id`);
  if (typeof value.slot !== "string" || !MULTIPLIER_SLOT_ORDER.includes(value.slot as MultiplierSlot)) {
    throw new SyntaxError(`${path}.slot is unsupported`);
  }
  const target = value.target === "all" ? "all" : parseId(value.target, `${path}.target`);
  if (target !== "all" && !generators.some((generator) => generator.id === target)) {
    throw new SyntaxError(`${path}.target references an unknown generator`);
  }
  const provider = parseId(value.provider, `${path}.provider`);
  return Object.freeze({ id, slot: value.slot as MultiplierSlot, target, provider });
}

function parseProgressCoordinate(
  source: unknown,
  index: number,
  resources: ReadonlyMap<string, ResourceDefinition>,
): ProgressCoordinateDefinition {
  const path = `catalog.progress_coordinates[${index}]`;
  if (!isRecord(source) || !Number.isSafeInteger(source.tier) || (source.tier as number) < 0 || (source.tier as number) > 3 || typeof source.kind !== "string") {
    throw new SyntaxError(`${path} requires tier 0..3 and a supported kind`);
  }
  const tier = source.tier as number;
  switch (source.kind) {
    case "resource_log": {
      const value = exactObject(source, ["tier", "kind", "resource", "target"], path);
      const term = parseProgressTerm({ kind: value.kind, resource: value.resource, target: value.target }, `${path}`, resources, false);
      return Object.freeze({ tier, kind: "resource_log", resourceId: term.resourceId, target: term.target });
    }
    case "count_fraction": {
      const value = exactObject(source, ["tier", "kind", "count", "required"], path);
      const term = parseProgressTerm({ kind: value.kind, count: value.count, required: value.required }, `${path}`, resources, false);
      return Object.freeze({ tier, kind: "count_fraction", countKey: term.countKey, required: term.required });
    }
    case "composite": {
      const value = exactObject(source, ["tier", "kind", "terms"], path);
      if (!Array.isArray(value.terms) || value.terms.length === 0) throw new SyntaxError(`${path}.terms must be non-empty`);
      const terms = value.terms.map((term, termIndex) => parseProgressTerm(term, `${path}.terms[${termIndex}]`, resources, true));
      const sum = terms.reduce((total, term) => total.add(parseCanonical(term.weight)), new Decimal(0));
      if (!sum.eq(1)) throw new SyntaxError(`${path}.terms weights must sum to 1e0`);
      return Object.freeze({ tier, kind: "composite", terms: Object.freeze(terms) });
    }
    default:
      throw new SyntaxError(`${path}.kind is unsupported`);
  }
}

function parseProgressTerm(
  source: unknown,
  path: string,
  resources: ReadonlyMap<string, ResourceDefinition>,
  weighted: boolean,
): ProgressTerm {
  if (!isRecord(source) || typeof source.kind !== "string") throw new SyntaxError(`${path} must be a progress term`);
  const weightKeys = weighted ? ["weight"] : [];
  let weight = "1e0";
  if (weighted) {
    weight = parseCanonicalField(source.weight, `${path}.weight`);
    const parsed = parseCanonical(weight);
    if (!parsed.gt(0) || parsed.gt(1)) throw new SyntaxError(`${path}.weight must be in (0,1]`);
  }
  if (source.kind === "resource_log") {
    const value = exactObject(source, [...weightKeys, "kind", "resource", "target"], path);
    const resourceId = parseId(value.resource, `${path}.resource`);
    if (resources.get(resourceId)?.scope !== "company") throw new SyntaxError(`${path}.resource must reference company state`);
    const target = parseCanonicalField(value.target, `${path}.target`);
    if (!parseCanonical(target).gt(0)) throw new SyntaxError(`${path}.target must be positive`);
    return Object.freeze({ weight, kind: "resource_log", resourceId, target });
  }
  if (source.kind === "count_fraction") {
    const value = exactObject(source, [...weightKeys, "kind", "count", "required"], path);
    if (value.count !== "generators.total_owned" || !Number.isSafeInteger(value.required) || (value.required as number) <= 0) {
      throw new SyntaxError(`${path} requires generators.total_owned and a positive safe integer`);
    }
    return Object.freeze({ weight, kind: "count_fraction", countKey: "generators.total_owned", required: value.required as number });
  }
  throw new SyntaxError(`${path}.kind is unsupported`);
}

function parseManualPolicy(source: unknown): ManualPolicy {
  const value = exactObject(source, ["refill_milli_per_ms", "bucket_cap_milli"], "catalog.manual_policy");
  const refillMilliPerMs = parsePositiveSafeInteger(value.refill_milli_per_ms, "catalog.manual_policy.refill_milli_per_ms");
  const bucketCapMilli = parsePositiveSafeInteger(value.bucket_cap_milli, "catalog.manual_policy.bucket_cap_milli");
  return Object.freeze({ refillMilliPerMs, bucketCapMilli });
}

function parseOfflinePolicy(source: unknown): OfflinePolicy {
  const path = "catalog.offline_policy";
  const value = exactObject(source, ["efficiency", "accrual_cap_ms", "bank_ratio_numerator", "bank_ratio_denominator", "bank_cap_ms", "burst_speed", "burst_max_duration_ms"], path);
  const efficiency = parseCanonicalField(value.efficiency, `${path}.efficiency`);
  if (parseCanonical(efficiency).lt(0) || parseCanonical(efficiency).gt(1)) throw new SyntaxError(`${path}.efficiency must be in [0,1]`);
  const burstSpeed = parseCanonicalField(value.burst_speed, `${path}.burst_speed`);
  if (!parseCanonical(burstSpeed).gt(0)) throw new SyntaxError(`${path}.burst_speed must be positive`);
  const accrualCapMs = parsePositiveSafeInteger(value.accrual_cap_ms, `${path}.accrual_cap_ms`);
  const bankRatioNumerator = parsePositiveSafeInteger(value.bank_ratio_numerator, `${path}.bank_ratio_numerator`);
  const bankRatioDenominator = parsePositiveSafeInteger(value.bank_ratio_denominator, `${path}.bank_ratio_denominator`);
  if (bankRatioNumerator > bankRatioDenominator) throw new SyntaxError(`${path} bank ratio may not exceed one`);
  return Object.freeze({
    efficiency,
    accrualCapMs,
    bankRatioNumerator,
    bankRatioDenominator,
    bankCapMs: parsePositiveSafeInteger(value.bank_cap_ms, `${path}.bank_cap_ms`),
    burstSpeed,
    burstMaxDurationMs: parsePositiveSafeInteger(value.burst_max_duration_ms, `${path}.burst_max_duration_ms`),
  });
}

function parseCurve(source: unknown, path: string): CostCurve {
  if (!isRecord(source) || typeof source.kind !== "string") {
    throw new SyntaxError(`${path} must contain a curve kind`);
  }
  switch (source.kind) {
    case "constant":
      exactObject(source, ["kind"], path);
      return Object.freeze({ kind: "constant" });
    case "linear": {
      const value = exactObject(source, ["kind", "step"], path);
      const step = parseCanonicalField(value.step, `${path}.step`);
      if (parseCanonical(step).lt(0)) throw new SyntaxError(`${path}.step must be non-negative`);
      return Object.freeze({ kind: "linear", step });
    }
    case "geometric": {
      const value = exactObject(source, ["kind", "ratio"], path);
      const ratio = parseCanonicalField(value.ratio, `${path}.ratio`);
      if (parseCanonical(ratio).lt(1)) throw new SyntaxError(`${path}.ratio must be at least one`);
      return Object.freeze({ kind: "geometric", ratio });
    }
    default:
      throw new SyntaxError(`${path}.kind is unsupported`);
  }
}

function exactObject(
  source: unknown,
  expectedKeys: readonly string[],
  path: string,
): Record<string, unknown> {
  if (!isRecord(source)) throw new SyntaxError(`${path} must be an object`);
  const actualKeys = Object.keys(source).sort();
  const wantedKeys = [...expectedKeys].sort();
  if (
    actualKeys.length !== wantedKeys.length ||
    actualKeys.some((key, index) => key !== wantedKeys[index])
  ) {
    throw new SyntaxError(`${path} must contain exactly: ${wantedKeys.join(", ")}`);
  }
  return source;
}

function isRecord(source: unknown): source is Record<string, unknown> {
  return typeof source === "object" && source !== null && !Array.isArray(source);
}

function isScope(source: unknown): source is Scope {
  return source === "company" || source === "founder" || source === "world" || source === "guild";
}

function parseId(source: unknown, path: string): string {
  if (typeof source !== "string" || !idPattern.test(source)) {
    throw new SyntaxError(`${path} is not a valid mechanical ID`);
  }
  return source;
}

function parseCanonicalField(source: unknown, path: string): string {
  if (typeof source !== "string") throw new SyntaxError(`${path} must be a canonical string`);
  try {
    parseCanonical(source);
  } catch {
    throw new SyntaxError(`${path} must be an RFC-0001 canonical state decimal`);
  }
  return source;
}

function ensureUnique(
  definitions: readonly { readonly id: string }[],
  kind: string,
): void {
  const ids = new Set<string>();
  for (const definition of definitions) {
    if (ids.has(definition.id)) throw new SyntaxError(`duplicate ${kind} id: ${definition.id}`);
    ids.add(definition.id);
  }
}

function validateCount(value: number, name: string): void {
  if (!Number.isSafeInteger(value) || value < 0 || value > MAX_EXACT_INTEGER) {
    throw new RangeError(`${name} must be a non-negative exact integer`);
  }
}

function parsePositiveSafeInteger(source: unknown, path: string): number {
  if (!Number.isSafeInteger(source) || (source as number) <= 0) {
    throw new SyntaxError(`${path} must be a positive safe integer`);
  }
  return source as number;
}

function parsePositive(source: string, path: string): Decimal {
  const value = parseCanonical(source);
  if (!value.gt(0)) throw new RangeError(`${path} must be positive`);
  return value;
}

function parseNonNegative(source: string, path: string): Decimal {
  const value = parseCanonical(source);
  if (value.lt(0)) throw new RangeError(`${path} must be non-negative`);
  return value;
}

export function canonicalBulkCost(
  catalog: EconomyCatalog,
  generatorId: string,
  owned: number,
  count: number,
): string {
  return canonicalString(catalog.bulkCost(generatorId, owned, count));
}
