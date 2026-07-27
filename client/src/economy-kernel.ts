import Decimal from "break_infinity.js";

import { sumGeometricSeries } from "./economy";
import {
  canonicalString,
  isStateValue,
  MAX_EXACT_INTEGER,
  parseCanonical,
} from "./numeric";

export const CATALOG_SCHEMA_VERSION = 1;

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

export interface GeneratorClassDefinition {
  readonly id: string;
  readonly price: PriceDefinition;
}

const idPattern = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;

export class EconomyCatalog {
  readonly resources: readonly ResourceDefinition[];
  readonly generatorClasses: readonly GeneratorClassDefinition[];
  readonly #resourceById: ReadonlyMap<string, ResourceDefinition>;
  readonly #generatorById: ReadonlyMap<string, GeneratorClassDefinition>;

  constructor(
    resources: readonly ResourceDefinition[],
    generatorClasses: readonly GeneratorClassDefinition[],
  ) {
    this.resources = Object.freeze([...resources]);
    this.generatorClasses = Object.freeze([...generatorClasses]);
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
  const root = exactObject(
    source,
    ["schema_version", "resources", "generator_classes"],
    "catalog",
  );
  if (root.schema_version !== CATALOG_SCHEMA_VERSION) {
    throw new SyntaxError(`unsupported catalog schema_version: ${String(root.schema_version)}`);
  }
  if (!Array.isArray(root.resources)) throw new SyntaxError("catalog.resources must be an array");
  if (!Array.isArray(root.generator_classes)) {
    throw new SyntaxError("catalog.generator_classes must be an array");
  }

  const resources = root.resources.map((value, index) => parseResource(value, index));
  ensureUnique(resources, "resource");
  const resourceIds = new Set(resources.map((definition) => definition.id));

  const generatorClasses = root.generator_classes.map((value, index) =>
    parseGeneratorClass(value, index),
  );
  ensureUnique(generatorClasses, "generator class");
  for (const generator of generatorClasses) {
    if (!resourceIds.has(generator.price.resourceId)) {
      throw new SyntaxError(
        `generator class ${generator.id} references unknown resource ${generator.price.resourceId}`,
      );
    }
  }

  return new EconomyCatalog(resources, generatorClasses);
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

function parseGeneratorClass(source: unknown, index: number): GeneratorClassDefinition {
  const path = `catalog.generator_classes[${index}]`;
  const value = exactObject(source, ["id", "price"], path);
  const id = parseId(value.id, `${path}.id`);
  const priceValue = exactObject(value.price, ["resource_id", "base", "curve"], `${path}.price`);
  const resourceId = parseId(priceValue.resource_id, `${path}.price.resource_id`);
  const base = parseCanonicalField(priceValue.base, `${path}.price.base`);
  if (!parseCanonical(base).gt(0)) throw new SyntaxError(`${path}.price.base must be positive`);
  const curve = parseCurve(priceValue.curve, `${path}.price.curve`);
  const price = Object.freeze({ resourceId, base, curve });
  return Object.freeze({ id, price });
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
