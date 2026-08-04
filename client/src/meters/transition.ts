import type { MeterCatalog, MeterDefinition } from "./catalog";
import { METER_MAX_VALUE, METER_MIN_VALUE } from "./catalog";

export const MILLIS_PER_HOUR = 3_600_000 as const;
export const MAXIMUM_ATTENDED_STEP = 86_400_000 as const;

export interface MeterState {
  readonly values: Record<string, number>;
  readonly decayRemainders: Record<string, number>;
  readonly inputRemainders: Record<string, number>;
}

export interface MeterAdvanceContext {
  readonly attendedMs: number;
  readonly newFactKinds: ReadonlySet<string>;
  readonly activeContributions: ReadonlySet<string>;
}

export interface MeterBandChange {
  readonly meterId: string;
  readonly fromBand: string;
  readonly toBand: string;
  readonly direction: "up" | "down";
  readonly valueBefore: number;
  readonly valueAfter: number;
}

export function inputRemainderKey(meterId: string, inputIndex: number): string {
  return `${meterId}:${inputIndex}`;
}

export function contributionKey(slot: string, sourceId: string): string {
  return `${slot}\0${sourceId}`;
}

export function newMeterState(catalog: MeterCatalog): MeterState {
  const state: MeterState = { values: {}, decayRemainders: {}, inputRemainders: {} };
  for (const meter of catalog.meters) {
    state.values[meter.id] = meter.initialValue;
    state.decayRemainders[meter.id] = 0;
    meter.inputs.forEach((input, index) => {
      if (input.kind === "contribution_slot") state.inputRemainders[inputRemainderKey(meter.id, index)] = 0;
    });
  }
  return state;
}

export function validateMeterState(catalog: MeterCatalog, state: MeterState): void {
  const expectedValues = catalog.meters.map((meter) => meter.id).sort();
  if (!sameStrings(Object.keys(state.values).sort(), expectedValues) || !sameStrings(Object.keys(state.decayRemainders).sort(), expectedValues)) invalidState();
  const expectedInputs: string[] = [];
  for (const meter of catalog.meters) {
    if (!integerInRange(state.values[meter.id], METER_MIN_VALUE, METER_MAX_VALUE) || !integerInRange(state.decayRemainders[meter.id], 0, MILLIS_PER_HOUR - 1)) invalidState();
    meter.inputs.forEach((input, index) => {
      if (input.kind !== "contribution_slot") return;
      const key = inputRemainderKey(meter.id, index);
      expectedInputs.push(key);
      if (!integerInRange(state.inputRemainders[key], 0, MILLIS_PER_HOUR - 1)) invalidState();
    });
  }
  if (!sameStrings(Object.keys(state.inputRemainders).sort(), expectedInputs.sort())) invalidState();
}

export function advanceMeters(catalog: MeterCatalog, state: MeterState, context: MeterAdvanceContext): readonly MeterBandChange[] {
  if (!integerInRange(context.attendedMs, 0, MAXIMUM_ATTENDED_STEP) || !(context.newFactKinds instanceof Set) || !(context.activeContributions instanceof Set)) invalidState();
  validateMeterState(catalog, state);
  const changes: MeterBandChange[] = [];
  for (const meter of catalog.meters) {
    const before = state.values[meter.id]!;
    let value = before;
    if (meter.decay !== null && value === meter.decay.towardValue) {
      state.decayRemainders[meter.id] = 0;
    } else if (meter.decay !== null && context.attendedMs > 0) {
      const [steps, remainder] = wholeSteps(meter.decay.ratePerHour, context.attendedMs, state.decayRemainders[meter.id]!);
      if (value < meter.decay.towardValue) {
        value += steps;
        if (value >= meter.decay.towardValue) {
          value = meter.decay.towardValue;
          state.decayRemainders[meter.id] = 0;
        } else state.decayRemainders[meter.id] = remainder;
      } else {
        value -= steps;
        if (value <= meter.decay.towardValue) {
          value = meter.decay.towardValue;
          state.decayRemainders[meter.id] = 0;
        } else state.decayRemainders[meter.id] = remainder;
      }
    }

    let factDelta = 0;
    let rateDelta = 0;
    meter.inputs.forEach((input, inputIndex) => {
      if (input.kind === "ledger_fact") {
        if (context.newFactKinds.has(input.factKind)) factDelta += input.delta;
        return;
      }
      if (!context.activeContributions.has(contributionKey(input.slot, input.sourceId)) || context.attendedMs === 0) return;
      const key = inputRemainderKey(meter.id, inputIndex);
      const sign = input.deltaPerAttendedHour < 0 ? -1 : 1;
      const [steps, remainder] = wholeSteps(Math.abs(input.deltaPerAttendedHour), context.attendedMs, state.inputRemainders[key]!);
      state.inputRemainders[key] = remainder;
      rateDelta += sign * steps;
    });
    value = clampValue(value + factDelta + rateDelta);
    state.values[meter.id] = value;
    const fromBand = bandFor(meter, before);
    const toBand = bandFor(meter, value);
    if (fromBand !== toBand) changes.push(Object.freeze({ meterId: meter.id, fromBand, toBand, direction: value > before ? "up" : "down", valueBefore: before, valueAfter: value }));
  }
  return Object.freeze(changes);
}

export function bandFor(meter: MeterDefinition, value: number): string {
  let band = meter.bands[0]!.id;
  for (const candidate of meter.bands.slice(1)) {
    if (value < candidate.floorValue) break;
    band = candidate.id;
  }
  return band;
}

function wholeSteps(rate: number, elapsedMs: number, remainder: number): readonly [number, number] {
  const numerator = rate * elapsedMs + remainder;
  if (!Number.isSafeInteger(numerator)) invalidState();
  return [Math.floor(numerator / MILLIS_PER_HOUR), numerator % MILLIS_PER_HOUR];
}

function clampValue(value: number): number {
  return Math.max(METER_MIN_VALUE, Math.min(METER_MAX_VALUE, value));
}

function integerInRange(value: unknown, minimum: number, maximum: number): value is number {
  return Number.isSafeInteger(value) && (value as number) >= minimum && (value as number) <= maximum;
}

function sameStrings(left: readonly string[], right: readonly string[]): boolean {
  return left.length === right.length && left.every((value, index) => value === right[index]);
}

function invalidState(): never {
  throw new RangeError("invalid meter state");
}
