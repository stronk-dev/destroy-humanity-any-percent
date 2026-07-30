import { idiv } from "../integer";

export type Temperament = "lazy" | "playful" | "curious" | "sassy" | "shy" | "chaotic";
export type ChartResult = -1 | 0 | 1;

const order: readonly Temperament[] = Object.freeze(["lazy", "playful", "curious", "sassy", "shy", "chaotic"]);
const INT32_MAX = 2_147_483_647n;
const INT32_MIN = -2_147_483_648n;

export function chart(attacker: Temperament, defender: Temperament): ChartResult {
  const attackerIndex = order.indexOf(attacker);
  const defenderIndex = order.indexOf(defender);
  if (attackerIndex < 0 || defenderIndex < 0) throw new RangeError("unknown temperament");
  const delta = (defenderIndex - attackerIndex + order.length) % order.length;
  if (delta === 1 || delta === 2) return 1;
  if (delta === 4 || delta === 5) return -1;
  return 0;
}

export function damage(basePower: number, attackerAtk: number, result: ChartResult, critical: boolean): number {
  int32(basePower, "base power"); int32(attackerAtk, "attacker atk");
  if (basePower < 0 || attackerAtk < 0 || (result !== -1 && result !== 0 && result !== 1)) throw new RangeError("invalid damage input");
  if (basePower === 0) return 0;
  let value = idiv(BigInt(basePower) * BigInt(attackerAtk), 64n);
  if (result === 1) value = idiv(value * 13n, 10n);
  if (result === -1) value = idiv(value * 10n, 13n);
  if (critical) value = idiv(value * 3n, 2n);
  if (value < 1n) value = 1n;
  return saturateInt32(value);
}

export function saturateInt32(value: bigint): number {
  if (value > INT32_MAX) return Number(INT32_MAX);
  if (value < INT32_MIN) return Number(INT32_MIN);
  return Number(value);
}

export function clamp(value: bigint, minimum: number, maximum: number): number {
  int32(minimum, "minimum"); int32(maximum, "maximum");
  if (minimum < 0 || maximum < minimum) throw new RangeError("invalid clamp bounds");
  if (value < BigInt(minimum)) value = BigInt(minimum);
  if (value > BigInt(maximum)) value = BigInt(maximum);
  return Number(value);
}

function int32(value: number, label: string): void {
  if (!Number.isInteger(value) || value < Number(INT32_MIN) || value > Number(INT32_MAX)) throw new RangeError(`${label} outside int32`);
}
