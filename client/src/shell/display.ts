import Decimal from "break_infinity.js";
import { isStateValue, parseCanonical } from "../numeric";
import type { IntentReceipt, ResourceState } from "./contracts";
import type { ClientShellPolicy } from "./policy";
import type { DecimalTuple } from "./prediction";
import { reconcileContinuous, type ReconciliationDecision } from "./reconciliation";

export interface CounterView { readonly value: string; readonly activityPpm: number; readonly capReasonKey?: string; readonly explanation?: string; readonly pulse: boolean }

export class DisplayCounter {
  readonly #policy: ClientShellPolicy; readonly #reducedMotion: boolean;
  #resource: ResourceState; #value: Decimal; #from: Decimal; #to: Decimal; #startedMs = 0; #durationMs = 0;
  #activityPpm = 0; #explanation: string | undefined; #pulse = false; #lastReducedMs = -Infinity; #reducedValue = "0";
  constructor(resource: ResourceState, policy: ClientShellPolicy, reducedMotion = false) {
    this.#resource = resource; this.#policy = policy; this.#reducedMotion = reducedMotion;
    this.#value = parseCanonical(resource.amount); this.#from = this.#value; this.#to = this.#value; this.#reducedValue = this.#value.toString();
  }
  applyPrediction(predicted: DecimalTuple): void {
    const next = Decimal.fromMantissaExponent(predicted.mantissa, predicted.exponent);
    if (!isStateValue(next) || next.lt(0)) throw new RangeError("invalid predicted counter");
    if (parseCanonical(this.#resource.ratePerSecond).gt(0)) this.#activityPpm = Math.min(1_000_000, this.#activityPpm + 1);
    if (this.#durationMs > 0) this.#to = next;
    else { this.#value = next; this.#from = next; this.#to = next; }
  }
  applyAuthoritative(resource: ResourceState, nowMs: number, receipt?: IntentReceipt): ReconciliationDecision {
    const decision = reconcileContinuous(this.#value, resource.amount, this.#policy, receipt);
    this.#resource = resource; this.#explanation = decision.mode === "rebase" ? decision.explanation : undefined;
    if (decision.mode === "interpolate" && !this.#reducedMotion) { this.#from = this.#value; this.#to = parseCanonical(resource.amount); this.#startedMs = nowMs; this.#durationMs = decision.durationMs; }
    else if (decision.mode === "interpolate") { this.#value = parseCanonical(resource.amount); this.#from = this.#value; this.#to = this.#value; this.#durationMs = 0; }
    else if (decision.mode === "rebase") { this.#value = parseCanonical(resource.amount); this.#from = this.#value; this.#to = this.#value; this.#durationMs = 0; this.#pulse = !this.#reducedMotion; }
    return decision;
  }
  replaceAuthoritative(resource: ResourceState): void {
    this.#resource = resource; this.#value = parseCanonical(resource.amount); this.#from = this.#value; this.#to = this.#value;
    this.#durationMs = 0; this.#explanation = undefined; this.#pulse = false; this.#reducedValue = this.#value.toString();
  }
  view(nowMs: number): CounterView {
    if (this.#durationMs > 0) {
      const progress = Math.max(0, Math.min(1, (nowMs - this.#startedMs) / this.#durationMs));
      this.#value = this.#from.add(this.#to.sub(this.#from).mul(progress));
      if (progress === 1) this.#durationMs = 0;
    }
    let value = this.#value.toString();
    if (this.#reducedMotion) {
      if (nowMs - this.#lastReducedMs >= this.#policy.reducedMotionRenderMs) { this.#lastReducedMs = nowMs; this.#reducedValue = value; }
      value = this.#reducedValue;
    }
    const capReasonKey = this.#resource.cap && this.#value.gte(parseCanonical(this.#resource.cap.amount)) ? this.#resource.cap.reasonKey : undefined;
    const result = { value, activityPpm: this.#activityPpm, ...(capReasonKey ? { capReasonKey } : {}), ...(this.#explanation ? { explanation: this.#explanation } : {}), pulse: this.#pulse };
    this.#pulse = false; return result;
  }
}
