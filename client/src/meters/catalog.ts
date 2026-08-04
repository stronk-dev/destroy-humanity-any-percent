export const METER_CATALOG_SCHEMA_VERSION = 1 as const;
export const METER_MIN_VALUE = 0 as const;
export const METER_MAX_VALUE = 100 as const;

export type MeterInput =
  | Readonly<{ kind: "ledger_fact"; factKind: string; delta: number }>
  | Readonly<{ kind: "contribution_slot"; slot: MeterSlot; sourceId: string; deltaPerAttendedHour: number }>;

export type MeterSlot = "upgrades" | "milestones" | "faction" | "doctrine" | "commons" | "trust" | "event_buffs" | "prestige";

export interface MeterBand { readonly id: string; readonly floorValue: number }
export interface MeterDecay { readonly towardValue: number; readonly ratePerHour: number }
export interface MeterDefinition {
  readonly id: string;
  readonly initialValue: number;
  readonly bands: readonly MeterBand[];
  readonly inputs: readonly MeterInput[];
  readonly decay: MeterDecay | null;
}
export interface TrustReseed {
  readonly baseValue: number;
  readonly notorietyNumerator: number;
  readonly notorietyDenominator: number;
  readonly floorValue: number;
  readonly ceilingValue: number;
}
export interface MeterCatalog {
  readonly schemaVersion: 1;
  readonly trustReseed: TrustReseed;
  readonly meters: readonly MeterDefinition[];
  readonly byId: ReadonlyMap<string, MeterDefinition>;
}

const mechanicalId = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;
const slots = new Set<MeterSlot>(["upgrades", "milestones", "faction", "doctrine", "commons", "trust", "event_buffs", "prestige"]);
export const REQUIRED_METER_IDS = Object.freeze([
  "doom.probability",
  "trust.employees.grievance", "trust.employees.standing",
  "trust.investors.grievance", "trust.investors.standing",
  "trust.press.grievance", "trust.press.standing",
  "trust.regulators.grievance", "trust.regulators.standing",
  "trust.users.grievance", "trust.users.standing",
] as const);

function syntax(message: string): never { throw new SyntaxError(message); }

function exactObject(value: unknown, keys: readonly string[], label: string): Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value)) syntax(`${label} must be an object`);
  const actual = Object.keys(value).sort();
  const expected = [...keys].sort();
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) syntax(`${label} has invalid keys`);
  return value as Record<string, unknown>;
}

function integer(value: unknown, minimum: number, maximum: number, label: string): number {
  if (!Number.isSafeInteger(value) || (value as number) < minimum || (value as number) > maximum) syntax(`${label} is invalid`);
  return value as number;
}

function identifier(value: unknown, label: string): string {
  if (typeof value !== "string" || !mechanicalId.test(value)) syntax(`${label} is invalid`);
  return value;
}

function parseInput(value: unknown, label: string): { readonly input: MeterInput; readonly uniqueness: string } {
  const discriminator = value !== null && typeof value === "object" && !Array.isArray(value) ? (value as Record<string, unknown>).kind : undefined;
  if (discriminator === "ledger_fact") {
    const row = exactObject(value, ["kind", "fact_kind", "delta"], label);
    const delta = integer(row.delta, -METER_MAX_VALUE, METER_MAX_VALUE, `${label}.delta`);
    if (delta === 0) syntax(`${label}.delta must be nonzero`);
    const factKind = identifier(row.fact_kind, `${label}.fact_kind`);
    return { input: Object.freeze({ kind: "ledger_fact", factKind, delta }), uniqueness: `fact\0${factKind}` };
  }
  if (discriminator === "contribution_slot") {
    const row = exactObject(value, ["kind", "slot", "source_id", "delta_per_attended_hour"], label);
    if (typeof row.slot !== "string" || !slots.has(row.slot as MeterSlot)) syntax(`${label}.slot is invalid`);
    const sourceId = identifier(row.source_id, `${label}.source_id`);
    const deltaPerAttendedHour = integer(row.delta_per_attended_hour, -METER_MAX_VALUE, METER_MAX_VALUE, `${label}.delta_per_attended_hour`);
    if (deltaPerAttendedHour === 0) syntax(`${label}.delta_per_attended_hour must be nonzero`);
    return { input: Object.freeze({ kind: "contribution_slot", slot: row.slot as MeterSlot, sourceId, deltaPerAttendedHour }), uniqueness: `slot\0${row.slot}\0${sourceId}` };
  }
  return syntax(`${label}.kind is invalid`);
}

export function loadMeterCatalog(bytes: string | Uint8Array): MeterCatalog {
  const source = typeof bytes === "string" ? bytes : new TextDecoder("utf-8", { fatal: true }).decode(bytes);
  let value: unknown;
  try { value = JSON.parse(source); } catch (error) { syntax(`meter catalog is invalid JSON: ${String(error)}`); }
  const root = exactObject(value, ["schema_version", "trust_reseed", "meters"], "meter catalog");
  if (root.schema_version !== METER_CATALOG_SCHEMA_VERSION || !Array.isArray(root.meters) || root.meters.length !== REQUIRED_METER_IDS.length) syntax("meter catalog has invalid version or meter count");
  const reseedRow = exactObject(root.trust_reseed, ["base_value", "notoriety_numerator", "notoriety_denominator", "floor_value", "ceiling_value"], "meter catalog.trust_reseed");
  const trustReseed = Object.freeze({
    baseValue: integer(reseedRow.base_value, 0, 100, "trust_reseed.base_value"),
    notorietyNumerator: integer(reseedRow.notoriety_numerator, 0, Number.MAX_SAFE_INTEGER, "trust_reseed.notoriety_numerator"),
    notorietyDenominator: integer(reseedRow.notoriety_denominator, 1, Number.MAX_SAFE_INTEGER, "trust_reseed.notoriety_denominator"),
    floorValue: integer(reseedRow.floor_value, 0, 100, "trust_reseed.floor_value"),
    ceilingValue: integer(reseedRow.ceiling_value, 0, 100, "trust_reseed.ceiling_value"),
  });
  if (trustReseed.floorValue > trustReseed.baseValue || trustReseed.baseValue > trustReseed.ceilingValue) syntax("trust_reseed bounds are invalid");
  const meters = root.meters.map((raw, meterIndex): MeterDefinition => {
    const label = `meter catalog.meters[${meterIndex}]`;
    const row = exactObject(raw, ["id", "scope", "min_value", "max_value", "initial_value", "bands", "inputs", "decay"], label);
    if (row.id !== REQUIRED_METER_IDS[meterIndex] || row.scope !== "company" || row.min_value !== METER_MIN_VALUE || row.max_value !== METER_MAX_VALUE || !Array.isArray(row.bands) || row.bands.length === 0 || !Array.isArray(row.inputs)) syntax(`${label} is invalid`);
    const bands: MeterBand[] = [];
    const seenBands = new Set<string>();
    let priorFloor = -1;
    for (let bandIndex = 0; bandIndex < row.bands.length; bandIndex += 1) {
      const bandRow = exactObject(row.bands[bandIndex], ["id", "floor_value"], `${label}.bands[${bandIndex}]`);
      const id = identifier(bandRow.id, `${label}.bands[${bandIndex}].id`);
      const floorValue = integer(bandRow.floor_value, 0, 100, `${label}.bands[${bandIndex}].floor_value`);
      if (seenBands.has(id) || floorValue <= priorFloor || bandIndex === 0 && floorValue !== 0) syntax(`${label}.bands are invalid`);
      seenBands.add(id); priorFloor = floorValue; bands.push(Object.freeze({ id, floorValue }));
    }
    const inputs: MeterInput[] = [];
    const seenInputs = new Set<string>();
    row.inputs.forEach((rawInput, inputIndex) => {
      const parsed = parseInput(rawInput, `${label}.inputs[${inputIndex}]`);
      if (seenInputs.has(parsed.uniqueness)) syntax(`${label}.inputs contain a duplicate source`);
      seenInputs.add(parsed.uniqueness); inputs.push(parsed.input);
    });
    let decay: MeterDecay | null = null;
    if (row.decay !== null) {
      const decayRow = exactObject(row.decay, ["toward_value", "rate_per_attended_hour"], `${label}.decay`);
      decay = Object.freeze({ towardValue: integer(decayRow.toward_value, 0, 100, `${label}.decay.toward_value`), ratePerHour: integer(decayRow.rate_per_attended_hour, 1, 100, `${label}.decay.rate_per_attended_hour`) });
    }
    return Object.freeze({ id: row.id as string, initialValue: integer(row.initial_value, 0, 100, `${label}.initial_value`), bands: Object.freeze(bands), inputs: Object.freeze(inputs), decay });
  });
  return Object.freeze({ schemaVersion: 1, trustReseed, meters: Object.freeze(meters), byId: new Map(meters.map((meter) => [meter.id, meter])) });
}

export function validateMeterResourceSeparation(catalog: MeterCatalog, resourceIds: readonly string[]): void {
  const meters = new Set(catalog.meters.map((meter) => meter.id));
  const collision = resourceIds.find((id) => meters.has(id));
  if (collision !== undefined) syntax(`meter/economy ID collision ${collision}`);
}
