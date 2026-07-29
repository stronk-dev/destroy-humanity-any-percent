import Decimal from "break_infinity.js";
import { isStateValue, parseCanonical } from "../numeric";
import type { AuthoritativeSnapshot } from "./contracts";
import type { ClientShellPolicy } from "./policy";

export interface DecimalTuple { readonly mantissa: number; readonly exponent: number }
export interface PredictedSnapshot { readonly revision: number; readonly atMonotonicMs: number; readonly resources: Readonly<Record<string, DecimalTuple>> }
export type PredictionOutput = { readonly kind: "predicted_snapshot"; readonly snapshot: PredictedSnapshot } | { readonly kind: "offline_required"; readonly gapMs: number } | { readonly kind: "worker_metric"; readonly name: "tick_overrun"; readonly value: number };

function tuple(value: Decimal): DecimalTuple { return { mantissa: value.mantissa, exponent: value.exponent }; }

export class PredictionMachine {
  readonly #policy: ClientShellPolicy;
  #revision = 0; #lastPulseMs: number | undefined; #accumulatorMs = 0; #sinceSnapshotMs = 0;
  #resources = new Map<string, { amount: Decimal; rate: Decimal; cap?: Decimal }>();
  constructor(policy: ClientShellPolicy) { this.#policy = policy; }
  initialize(snapshot: AuthoritativeSnapshot, monotonicMs: number): void { this.applyAuthoritative(snapshot); this.#lastPulseMs = monotonicMs; this.#accumulatorMs = 0; this.#sinceSnapshotMs = 0; }
  applyAuthoritative(snapshot: AuthoritativeSnapshot): void {
    if (snapshot.revision < this.#revision) return;
    this.#revision = snapshot.revision; this.#resources.clear();
    for (const [id, value] of Object.entries(snapshot.resources)) this.#resources.set(id, { amount: parseCanonical(value.amount), rate: parseCanonical(value.ratePerSecond), cap: value.cap ? parseCanonical(value.cap.amount) : undefined });
  }
  pulse(monotonicMs: number): PredictionOutput[] {
    if (!Number.isFinite(monotonicMs) || this.#lastPulseMs === undefined || monotonicMs < this.#lastPulseMs) throw new RangeError("non-monotonic prediction clock");
    const gapMs = monotonicMs - this.#lastPulseMs; this.#lastPulseMs = monotonicMs;
    if (gapMs > this.#policy.catchupCeilingMs) { this.#accumulatorMs = 0; this.#sinceSnapshotMs = 0; return [{ kind: "offline_required", gapMs }]; }
    this.#accumulatorMs += gapMs; const outputs: PredictionOutput[] = []; let steps = 0;
    while (this.#accumulatorMs >= this.#policy.tickMs && steps < 100) {
      const seconds = new Decimal(this.#policy.tickMs).div(1000);
      for (const value of this.#resources.values()) {
        let next = value.amount.add(value.rate.mul(seconds));
        if (value.cap && next.gt(value.cap)) next = value.cap;
        if (!isStateValue(next)) throw new RangeError("prediction outside Decimal domain");
        value.amount = next;
      }
      this.#accumulatorMs -= this.#policy.tickMs; this.#sinceSnapshotMs += this.#policy.tickMs; steps++;
    }
    if (this.#accumulatorMs >= this.#policy.tickMs) outputs.push({ kind: "worker_metric", name: "tick_overrun", value: this.#accumulatorMs });
    if (this.#sinceSnapshotMs >= this.#policy.snapshotMs) { this.#sinceSnapshotMs %= this.#policy.snapshotMs; outputs.push({ kind: "predicted_snapshot", snapshot: this.snapshot(monotonicMs) }); }
    return outputs;
  }
  snapshot(atMonotonicMs: number): PredictedSnapshot { return { revision: this.#revision, atMonotonicMs, resources: Object.fromEntries([...this.#resources].map(([id, value]) => [id, tuple(value.amount)])) }; }
}
