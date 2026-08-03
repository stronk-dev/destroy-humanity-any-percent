import Decimal from "break_infinity.js";

import { affordGeometricSeriesDetailed, sumGeometricSeries } from "./economy";
import {
  canonicalString,
  isStateValue,
  MAX_EXACT_INTEGER,
  parseCanonical,
  quantize,
  sumDeterministic,
} from "./numeric";
import { parseRoutePredicate, type RouteCondition } from "./routes";

export const CATALOG_SCHEMA_VERSION = 4;

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
  readonly tier: number;
  readonly category: string;
  readonly price: PriceDefinition;
  readonly production: ProductionDefinition | null;
  readonly provision: ProvisionDefinition | null;
  readonly provisionedHardcap: ProvisionedHardcap | null;
  readonly ladder: readonly LadderRung[];
  readonly roles: readonly GeneratorRole[];
}

export interface ProvisionDefinition { readonly generatorId: string; readonly ratePpm: number }
export interface ProvisionedHardcap { readonly count: number; readonly reasonKey: string }
export interface LadderRung { readonly purchasedAt: number; readonly multiplierPpm: number }
export type GeneratorRole =
  | { readonly kind: "provision"; readonly generatorId: string }
  | { readonly kind: "synergy_feed"; readonly poolId: string }
  | { readonly kind: "manual_output"; readonly actionId: string; readonly perPurchasedPpm: number }
  | { readonly kind: "stock_rate"; readonly perPurchasedPpm: number };

export interface UpgradeDefinition {
  readonly id: string;
  readonly cost: { readonly resourceId: string; readonly amount: string };
  readonly window: { readonly fromGate: string | null; readonly toGate: string | null };
  readonly requires: readonly RouteCondition[];
  readonly effects: readonly { readonly sourceId: string; readonly slot: "upgrades"; readonly target: string; readonly factor: string }[];
  readonly roles: readonly GeneratorRole["kind"][];
  readonly copyKey: string;
}

export interface SynergyPoolDefinition {
  readonly id: string;
  readonly sources: readonly { readonly kind: "generator" | "upgrade"; readonly id: string; readonly perCountPpm: number }[];
  readonly slot: MultiplierSlot;
  readonly curve: "linear" | "log";
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

export interface ProgressState {
  readonly balances: Readonly<Record<string, string>>;
  readonly generatorCounts: Readonly<Record<string, number>>;
}

const idPattern = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;
const minimumResourceLogTarget = new Decimal("5e-15");

export class EconomyCatalog {
  readonly resources: readonly ResourceDefinition[];
  readonly generatorClasses: readonly GeneratorClassDefinition[];
  readonly upgrades: readonly UpgradeDefinition[];
  readonly synergyPools: readonly SynergyPoolDefinition[];
  readonly provisionTickMs: number;
  readonly manualActions: readonly ManualActionDefinition[];
  readonly multiplierSources: readonly MultiplierSourceDefinition[];
  readonly progressCoordinates: readonly ProgressCoordinateDefinition[];
  readonly manualPolicy: ManualPolicy | null;
  readonly offlinePolicy: OfflinePolicy | null;
  readonly #resourceById: ReadonlyMap<string, ResourceDefinition>;
  readonly #generatorById: ReadonlyMap<string, GeneratorClassDefinition>;
  readonly #upgradeById: ReadonlyMap<string, UpgradeDefinition>;
  readonly #synergyById: ReadonlyMap<string, SynergyPoolDefinition>;

  constructor(
    resources: readonly ResourceDefinition[],
    generatorClasses: readonly GeneratorClassDefinition[],
    upgrades: readonly UpgradeDefinition[] = [],
    synergyPools: readonly SynergyPoolDefinition[] = [],
    provisionTickMs = 0,
    manualActions: readonly ManualActionDefinition[] = [],
    multiplierSources: readonly MultiplierSourceDefinition[] = [],
    progressCoordinates: readonly ProgressCoordinateDefinition[] = [],
    manualPolicy: ManualPolicy | null = null,
    offlinePolicy: OfflinePolicy | null = null,
  ) {
    this.resources = Object.freeze([...resources]);
    this.generatorClasses = Object.freeze([...generatorClasses]);
    this.upgrades = Object.freeze([...upgrades]);
    this.synergyPools = Object.freeze([...synergyPools]);
    this.provisionTickMs = provisionTickMs;
    this.manualActions = Object.freeze([...manualActions]);
    this.multiplierSources = Object.freeze([...multiplierSources]);
    this.progressCoordinates = Object.freeze([...progressCoordinates]);
    this.manualPolicy = manualPolicy;
    this.offlinePolicy = offlinePolicy;
    this.#resourceById = new Map(resources.map((definition) => [definition.id, definition]));
    this.#generatorById = new Map(
      generatorClasses.map((definition) => [definition.id, definition]),
    );
    this.#upgradeById = new Map(upgrades.map((definition) => [definition.id, definition]));
    this.#synergyById = new Map(synergyPools.map((definition) => [definition.id, definition]));
  }

  resource(id: string): ResourceDefinition | undefined {
    return this.#resourceById.get(id);
  }

  generatorClass(id: string): GeneratorClassDefinition | undefined {
    return this.#generatorById.get(id);
  }

  upgrade(id: string): UpgradeDefinition | undefined { return this.#upgradeById.get(id); }
  synergyPool(id: string): SynergyPoolDefinition | undefined { return this.#synergyById.get(id); }

  bulkCost(generatorId: string, owned: number, count: number): Decimal {
    const generator = this.generatorClass(generatorId);
    if (!generator) throw new RangeError(`unknown generator class: ${generatorId}`);
    return bulkCost(generator.price, owned, count);
  }

  maxAffordable(generatorId: string, cashSource: string | Decimal, owned: number): number {
    return this.maxAffordableDetailed(generatorId, cashSource, owned).count;
  }

  maxAffordableDetailed(generatorId: string, cashSource: string | Decimal, owned: number): { count: number; usedFallback: boolean } {
    const generator = this.generatorClass(generatorId);
    if (!generator) throw new RangeError(`unknown generator class: ${generatorId}`);
    const cash = cashSource instanceof Decimal ? new Decimal(cashSource) : parseCanonical(cashSource);
    if (generator.price.curve.kind === "geometric") {
      const result = affordGeometricSeriesDetailed(
        cash,
        parseCanonical(generator.price.base),
        parseCanonical(generator.price.curve.ratio),
        owned,
      );
      return {
        count: Math.min(result.count, MAX_EXACT_INTEGER - owned),
        usedFallback: result.usedFallback,
      };
    }
    return { count: maxAffordable(generator.price, cash, owned), usedFallback: false };
  }
}

export function parseCatalog(source: unknown): EconomyCatalog {
  if (!isRecord(source)) throw new SyntaxError("catalog must be an object");
  const version = source.schema_version;
  if (version !== 1 && version !== 2 && version !== 3 && version !== CATALOG_SCHEMA_VERSION) {
    throw new SyntaxError(`unsupported catalog schema_version: ${String(version)}`);
  }
  const root = exactObject(
    source,
    version === 4
      ? ["schema_version", "resources", "generator_classes", "upgrades", "synergy_pools", "provision_tick_ms", "manual_actions", "multiplier_sources", "progress_coordinates", "manual_policy", "offline_policy"]
      : version === 3
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
  for (const slot of ["commons", "trust"] as const) {
    if (multiplierSources.filter((definition) => definition.slot === slot).length > 1) {
      throw new SyntaxError(`multiplier slot ${slot} is single-provider`);
    }
  }
  const progressCoordinates = root.progress_coordinates.map((value, index) => parseProgressCoordinate(value, index, resourceById));
  const tiers = new Set(progressCoordinates.map((definition) => definition.tier));
  if (tiers.size !== progressCoordinates.length || [0, 1, 2, 3].some((tier) => !tiers.has(tier))) {
    throw new SyntaxError("progress coordinates must contain each tier 0..3 exactly once");
  }
  const manualPolicy = parseManualPolicy(root.manual_policy);
  const offlinePolicy = parseOfflinePolicy(root.offline_policy);
  if (version === 3) return new EconomyCatalog(resources, generatorClasses, [], [], 0, manualActions, multiplierSources, progressCoordinates, manualPolicy, offlinePolicy);

  if (!Array.isArray(root.upgrades) || !Array.isArray(root.synergy_pools)) throw new SyntaxError("catalog purchasable-content collections must be arrays");
  const provisionTickMs = parsePositiveSafeInteger(root.provision_tick_ms, "catalog.provision_tick_ms");
  const upgrades = root.upgrades.map((value, index) => parseUpgrade(value, index, resourceById));
  ensureUnique(upgrades, "upgrade");
  const upgradeById = new Map(upgrades.map((definition) => [definition.id, definition]));
  const allMultiplierSources = [...multiplierSources];
  const multiplierIds = new Set(multiplierSources.map((definition) => definition.id));
  for (const upgrade of upgrades) {
    for (const effect of upgrade.effects) {
      if (multiplierIds.has(effect.sourceId)) throw new SyntaxError(`duplicate multiplier source id: ${effect.sourceId}`);
      if (!generatorClasses.some((generator) => generator.id === effect.target) && !manualActions.some((action) => action.id === effect.target)) throw new SyntaxError(`upgrade ${upgrade.id} references unknown effect target ${effect.target}`);
      multiplierIds.add(effect.sourceId);
      allMultiplierSources.push(Object.freeze({ id: effect.sourceId, slot: effect.slot, target: effect.target, provider: upgrade.id }));
    }
  }
  const synergyPools = root.synergy_pools.map((value, index) => parseSynergyPool(value, index, generatorClasses, upgradeById));
  ensureUnique(synergyPools, "synergy pool");
  for (const pool of synergyPools) {
    const declaration = allMultiplierSources.find((value) => value.id === pool.id);
    if (!declaration || declaration.slot !== pool.slot || declaration.provider !== pool.id) throw new SyntaxError(`synergy pool ${pool.id} must map to one matching multiplier source`);
  }
  validateGeneratorContent(generatorClasses, synergyPools, manualActions);
  for (const generator of generatorClasses) {
    for (const rung of generator.ladder) {
      const id = ladderSourceId(generator.id, rung.purchasedAt);
      if (multiplierIds.has(id)) throw new SyntaxError(`duplicate ladder contribution source ${id}`);
      multiplierIds.add(id);
      allMultiplierSources.push(Object.freeze({ id, slot: "milestones", target: generator.id, provider: generator.id }));
    }
    for (const role of generator.roles) {
      if (role.kind !== "manual_output") continue;
      const id = manualRoleSourceId(generator.id, role.actionId);
      if (multiplierIds.has(id)) throw new SyntaxError(`duplicate manual role contribution source ${id}`);
      multiplierIds.add(id);
      allMultiplierSources.push(Object.freeze({ id, slot: "upgrades", target: role.actionId, provider: generator.id }));
    }
  }
  return new EconomyCatalog(resources, generatorClasses, upgrades, synergyPools, provisionTickMs, manualActions, allMultiplierSources, progressCoordinates, manualPolicy, offlinePolicy);
}

export function ladderSourceId(generatorId: string, purchasedAt: number): string { return `${generatorId}.ladder.purchased_${purchasedAt}`; }
export function manualRoleSourceId(generatorId: string, actionId: string): string { return `${generatorId}.role.manual_output.${actionId}`; }

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

export function subProgressValue(
  catalog: EconomyCatalog,
  state: ProgressState,
  tier: number,
): Decimal {
  const definition = catalog.progressCoordinates.find((coordinate) => coordinate.tier === tier);
  if (!definition) throw new RangeError(`missing progress coordinate tier ${tier}`);
  return evaluateProgressDefinition(definition, state);
}

function evaluateProgressDefinition(
  definition: ProgressCoordinateDefinition,
  state: ProgressState,
): Decimal {
  if (definition.kind === "resource_log") {
    return resourceLogProgress(state, definition.resourceId!, definition.target!);
  }
  if (definition.kind === "count_fraction") {
    return countFractionProgress(state, definition.countKey!, definition.required!);
  }
  return clampProgress(
    sumDeterministic(
      definition.terms!.map((term) => {
        const value = term.kind === "resource_log"
          ? resourceLogProgress(state, term.resourceId!, term.target!)
          : countFractionProgress(state, term.countKey!, term.required!);
        return value.mul(parseCanonical(term.weight));
      }),
    ),
  );
}

function resourceLogProgress(state: ProgressState, resourceId: string, target: string): Decimal {
  const encoded = state.balances[resourceId];
  if (encoded === undefined) throw new RangeError(`missing progress resource ${resourceId}`);
  const value = parseCanonical(encoded);
  const targetValue = parseCanonical(target);
  if (value.lt(0) || !targetValue.gt(0)) throw new RangeError("invalid progress resource state");
  const numerator = new Decimal(Decimal.add(1, value).log10());
  const denominator = new Decimal(Decimal.add(1, targetValue).log10());
  if (!isStateValue(denominator) || !denominator.gt(0)) {
    throw new RangeError("invalid resource_log denominator");
  }
  return clampProgress(numerator.div(denominator));
}

function countFractionProgress(
  state: ProgressState,
  countKey: "generators.total_owned",
  required: number,
): Decimal {
  if (countKey !== "generators.total_owned" || !Number.isSafeInteger(required) || required <= 0) {
    throw new RangeError("invalid progress count definition");
  }
  let total = 0;
  for (const count of Object.values(state.generatorCounts)) {
    validateCount(count, "generator count");
    if (count > MAX_EXACT_INTEGER - total) throw new RangeError("generator total exceeds exact domain");
    total += count;
  }
  return clampProgress(new Decimal(total).div(required));
}

function clampProgress(value: Decimal): Decimal {
  if (!isStateValue(value)) throw new RangeError("progress value is outside state domain");
  return quantize(Decimal.max(0, Decimal.min(1, value)));
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
  const value = schemaVersion === 4
    ? exactObjectWithOptional(source, ["id", "tier", "category", "price", "production", "ladder", "roles"], ["provisions", "provisioned_hardcap"], path)
    : exactObject(source, schemaVersion === 1 ? ["id", "price"] : ["id", "price", "production"], path);
  const id = parseId(value.id, `${path}.id`);
  const priceValue = exactObject(value.price, ["resource_id", "base", "curve"], `${path}.price`);
  const resourceId = parseId(priceValue.resource_id, `${path}.price.resource_id`);
  const base = parseCanonicalField(priceValue.base, `${path}.price.base`);
  if (!parseCanonical(base).gt(0)) throw new SyntaxError(`${path}.price.base must be positive`);
  const curve = parseCurve(priceValue.curve, `${path}.price.curve`);
  const price = Object.freeze({ resourceId, base, curve });
  if (schemaVersion === 1) return Object.freeze({ id, tier: 0, category: "", price, production: null, provision: null, provisionedHardcap: null, ladder: Object.freeze([]), roles: Object.freeze([]) });

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
  if (schemaVersion !== 4) return Object.freeze({ id, tier: 0, category: "", price, production, provision: null, provisionedHardcap: null, ladder: Object.freeze([]), roles: Object.freeze([]) });

  const tier = safeInteger(value.tier, 0, 3);
  const category = parseId(value.category, `${path}.category`);
  let provision: ProvisionDefinition | null = null;
  if (value.provisions !== undefined) {
    const raw = exactObject(value.provisions, ["generator_id", "rate_ppm"], `${path}.provisions`);
    provision = Object.freeze({ generatorId: parseId(raw.generator_id, `${path}.provisions.generator_id`), ratePpm: parsePositiveSafeInteger(raw.rate_ppm, `${path}.provisions.rate_ppm`) });
  }
  let provisionedHardcap: ProvisionedHardcap | null = null;
  if (value.provisioned_hardcap !== undefined) {
    const raw = exactObject(value.provisioned_hardcap, ["count", "reason_key"], `${path}.provisioned_hardcap`);
    provisionedHardcap = Object.freeze({ count: parsePositiveSafeInteger(raw.count, `${path}.provisioned_hardcap.count`), reasonKey: parseId(raw.reason_key, `${path}.provisioned_hardcap.reason_key`) });
  }
  if (!Array.isArray(value.ladder) || !Array.isArray(value.roles) || value.roles.length === 0) throw new SyntaxError(`${path} requires ladder and at least one role`);
  let prior = 0;
  const ladder = value.ladder.map((source, rungIndex) => {
    const raw = exactObject(source, ["purchased_at", "multiplier_ppm"], `${path}.ladder[${rungIndex}]`);
    const purchasedAt = parsePositiveSafeInteger(raw.purchased_at, `${path}.ladder[${rungIndex}].purchased_at`);
    const multiplierPpm = parsePositiveSafeInteger(raw.multiplier_ppm, `${path}.ladder[${rungIndex}].multiplier_ppm`);
    if (multiplierPpm === 1_000_000) throw new Error(`${path}.ladder[${rungIndex}].multiplier_ppm must be non-neutral`);
    if (purchasedAt <= prior) throw new SyntaxError(`${path}.ladder thresholds must increase`);
    prior = purchasedAt;
    return Object.freeze({ purchasedAt, multiplierPpm });
  });
  const seenRoles = new Set<string>();
  const roles = value.roles.map((source, roleIndex) => {
    const role = parseGeneratorRole(source, `${path}.roles[${roleIndex}]`);
    const target = role.kind === "provision" ? role.generatorId : role.kind === "synergy_feed" ? role.poolId : role.kind === "manual_output" ? role.actionId : "";
    const key = `${role.kind}:${target}`;
    if (seenRoles.has(key)) throw new SyntaxError(`${path} has duplicate role binding ${key}`);
    seenRoles.add(key);
    return role;
  });
  return Object.freeze({ id, tier, category, price, production, provision, provisionedHardcap, ladder: Object.freeze(ladder), roles: Object.freeze(roles) });
}

function parseGeneratorRole(source: unknown, path: string): GeneratorRole {
  if (!isRecord(source) || typeof source.kind !== "string") throw new SyntaxError(`${path} requires a role kind`);
  switch (source.kind) {
    case "provision": { const raw = exactObject(source, ["kind", "generator_id"], path); return Object.freeze({ kind: "provision", generatorId: parseId(raw.generator_id, `${path}.generator_id`) }); }
    case "synergy_feed": { const raw = exactObject(source, ["kind", "pool_id"], path); return Object.freeze({ kind: "synergy_feed", poolId: parseId(raw.pool_id, `${path}.pool_id`) }); }
    case "manual_output": { const raw = exactObject(source, ["kind", "action_id", "per_purchased_ppm"], path); return Object.freeze({ kind: "manual_output", actionId: parseId(raw.action_id, `${path}.action_id`), perPurchasedPpm: parsePositiveSafeInteger(raw.per_purchased_ppm, `${path}.per_purchased_ppm`) }); }
    case "stock_rate": { const raw = exactObject(source, ["kind", "per_purchased_ppm"], path); return Object.freeze({ kind: "stock_rate", perPurchasedPpm: parsePositiveSafeInteger(raw.per_purchased_ppm, `${path}.per_purchased_ppm`) }); }
    default: throw new SyntaxError(`${path}.kind is unsupported`);
  }
}

function parseUpgrade(source: unknown, index: number, resources: ReadonlyMap<string, ResourceDefinition>): UpgradeDefinition {
  const path = `catalog.upgrades[${index}]`;
  const value = exactObject(source, ["id", "cost", "window", "requires", "effects", "roles", "copy_key"], path);
  const id = parseId(value.id, `${path}.id`);
  const costRaw = exactObject(value.cost, ["resource", "amount"], `${path}.cost`);
  const resourceId = parseId(costRaw.resource, `${path}.cost.resource`);
  if (resources.get(resourceId)?.scope !== "company") throw new SyntaxError(`${path}.cost.resource must reference company state`);
  const amount = parseCanonicalField(costRaw.amount, `${path}.cost.amount`);
  if (!parseCanonical(amount).gt(0)) throw new SyntaxError(`${path}.cost.amount must be positive`);
  const windowRaw = exactObject(value.window, ["from_gate", "to_gate"], `${path}.window`);
  const fromGate = nullableId(windowRaw.from_gate, `${path}.window.from_gate`);
  const toGate = nullableId(windowRaw.to_gate, `${path}.window.to_gate`);
  const requires = parseRoutePredicate(value.requires);
  if (!Array.isArray(value.effects) || value.effects.length === 0 || !Array.isArray(value.roles)) throw new SyntaxError(`${path} requires non-empty effects and roles array`);
  const effectIds = new Set<string>();
  const effects = value.effects.map((source, effectIndex) => {
    const raw = exactObject(source, ["source_id", "slot", "target", "factor"], `${path}.effects[${effectIndex}]`);
    const sourceId = parseId(raw.source_id, `${path}.effects[${effectIndex}].source_id`);
    if (effectIds.has(sourceId) || raw.slot !== "upgrades") throw new SyntaxError(`${path}.effects must have unique sources in upgrades slot`);
    effectIds.add(sourceId);
    const factor = parseCanonicalField(raw.factor, `${path}.effects[${effectIndex}].factor`);
    if (!parseCanonical(factor).gt(0)) throw new SyntaxError(`${path}.effects factor must be positive`);
    return Object.freeze({ sourceId, slot: "upgrades" as const, target: parseId(raw.target, `${path}.effects[${effectIndex}].target`), factor });
  });
  const roleSet = new Set<string>();
  const roles = value.roles.map((role, roleIndex) => {
    if (typeof role !== "string" || !["provision", "synergy_feed", "manual_output", "stock_rate"].includes(role) || roleSet.has(role)) throw new SyntaxError(`${path}.roles[${roleIndex}] is invalid or duplicate`);
    roleSet.add(role);
    return role as GeneratorRole["kind"];
  });
  return Object.freeze({ id, cost: Object.freeze({ resourceId, amount }), window: Object.freeze({ fromGate, toGate }), requires, effects: Object.freeze(effects), roles: Object.freeze(roles), copyKey: parseId(value.copy_key, `${path}.copy_key`) });
}

function parseSynergyPool(source: unknown, index: number, generators: readonly GeneratorClassDefinition[], upgrades: ReadonlyMap<string, UpgradeDefinition>): SynergyPoolDefinition {
  const path = `catalog.synergy_pools[${index}]`;
  const value = exactObject(source, ["id", "sources", "slot", "curve"], path);
  const id = parseId(value.id, `${path}.id`);
  if (typeof value.slot !== "string" || !MULTIPLIER_SLOT_ORDER.includes(value.slot as MultiplierSlot) || value.curve !== "linear" && value.curve !== "log" || !Array.isArray(value.sources) || value.sources.length === 0) throw new SyntaxError(`${path} has invalid slot, curve, or sources`);
  const seen = new Set<string>();
  const sources = value.sources.map((source, sourceIndex) => {
    const raw = exactObject(source, ["kind", "id_or_class", "per_count_ppm"], `${path}.sources[${sourceIndex}]`);
    if (raw.kind !== "generator" && raw.kind !== "upgrade") throw new SyntaxError(`${path}.sources[${sourceIndex}].kind is unsupported`);
    const sourceId = parseId(raw.id_or_class, `${path}.sources[${sourceIndex}].id_or_class`);
    const key = `${raw.kind}:${sourceId}`;
    if (seen.has(key) || raw.kind === "generator" && !generators.some((generator) => generator.id === sourceId) || raw.kind === "upgrade" && !upgrades.has(sourceId)) throw new SyntaxError(`${path}.sources[${sourceIndex}] is duplicate or unknown`);
    seen.add(key);
    return Object.freeze({ kind: raw.kind, id: sourceId, perCountPpm: parsePositiveSafeInteger(raw.per_count_ppm, `${path}.sources[${sourceIndex}].per_count_ppm`) });
  });
  return Object.freeze({ id, sources: Object.freeze(sources), slot: value.slot as MultiplierSlot, curve: value.curve });
}

function validateGeneratorContent(generators: readonly GeneratorClassDefinition[], pools: readonly SynergyPoolDefinition[], manualActions: readonly ManualActionDefinition[]): void {
  const byId = new Map(generators.map((generator) => [generator.id, generator]));
  const providerByTarget = new Set<string>();
  for (const generator of generators) {
    if (generator.provision) {
      const target = byId.get(generator.provision.generatorId);
      if (!target || generator.tier !== target.tier + 1 || target.provisionedHardcap === null || providerByTarget.has(target.id)) throw new SyntaxError(`generator ${generator.id} has invalid provision topology`);
      providerByTarget.add(target.id);
    }
    for (const role of generator.roles) {
      if (role.kind === "provision" && generator.provision?.generatorId !== role.generatorId) throw new SyntaxError(`generator ${generator.id} provision role is unbound`);
      if (role.kind === "synergy_feed" && !pools.some((pool) => pool.id === role.poolId && pool.sources.some((source) => source.kind === "generator" && source.id === generator.id))) throw new SyntaxError(`generator ${generator.id} synergy role is unbound`);
      if (role.kind === "manual_output" && !manualActions.some((action) => action.id === role.actionId)) throw new SyntaxError(`generator ${generator.id} manual role is unbound`);
    }
  }
}

export function validateCatalogGateReferences(catalog: EconomyCatalog, gateIds: readonly string[]): void {
  const known = new Set(gateIds);
  if (known.size !== gateIds.length || gateIds.some((id) => !idPattern.test(id))) throw new SyntaxError("invalid route gate ids");
  for (const upgrade of catalog.upgrades) {
    if (upgrade.window.fromGate !== null && !known.has(upgrade.window.fromGate) || upgrade.window.toGate !== null && !known.has(upgrade.window.toGate)) throw new SyntaxError(`upgrade ${upgrade.id} references unknown availability gate`);
    for (const condition of upgrade.requires) {
      if ((condition.kind === "resource_at_least" || condition.kind === "resource_at_most") && catalog.resource(condition.resourceId)?.scope !== "company") throw new SyntaxError(`upgrade ${upgrade.id} references non-company resource`);
    }
  }
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
    const targetValue = parseCanonical(target);
    const denominator = new Decimal(Decimal.add(1, targetValue).log10());
    if (targetValue.lt(minimumResourceLogTarget) || !isStateValue(denominator) || !denominator.gt(0)) {
      throw new SyntaxError(`${path}.target must be at least 5e-15 with a finite positive logarithm`);
    }
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

function exactObjectWithOptional(source: unknown, requiredKeys: readonly string[], optionalKeys: readonly string[], path: string): Record<string, unknown> {
  if (!isRecord(source)) throw new SyntaxError(`${path} must be an object`);
  const allowed = new Set([...requiredKeys, ...optionalKeys]);
  if (requiredKeys.some((key) => !Object.hasOwn(source, key)) || Object.keys(source).some((key) => !allowed.has(key))) throw new SyntaxError(`${path} has missing or unknown fields`);
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

function nullableId(source: unknown, path: string): string | null {
  return source === null ? null : parseId(source, path);
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

function safeInteger(source: unknown, minimum: number, maximum: number): number {
  if (typeof source !== "number" || !Number.isSafeInteger(source) || source < minimum || source > maximum) throw new SyntaxError(`value must be a safe integer in ${minimum}..${maximum}`);
  return source;
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
