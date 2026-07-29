import Decimal from "break_infinity.js";
import { parseCanonical } from "../numeric";
import type { DiscreteFact, IntentReceipt } from "./contracts";
import type { ClientShellPolicy } from "./policy";

export type ReconciliationDecision =
  | { readonly mode: "none" }
  | { readonly mode: "interpolate"; readonly durationMs: number }
  | { readonly mode: "rebase"; readonly explanation?: string };

export function reconcileContinuous(predicted: Decimal, authoritativeSource: string, policy: ClientShellPolicy, receipt?: IntentReceipt): ReconciliationDecision {
  const authoritative = parseCanonical(authoritativeSource);
  if (receipt?.status === "rejected") return { mode: "rebase", explanation: receipt.rejectionCode ?? "rejected" };
  if (predicted.eq(authoritative)) return { mode: "none" };
  const denominator = Decimal.max(authoritative.abs(), 1);
  const differencePPM = predicted.sub(authoritative).abs().div(denominator).mul(1_000_000);
  return differencePPM.lt(policy.epsilonLerpPpm) ? { mode: "interpolate", durationMs: policy.lerpDurationMs } : { mode: "rebase" };
}

export function reconcileDiscrete(predicted: DiscreteFact | undefined, authoritative: DiscreteFact, receipt?: IntentReceipt): ReconciliationDecision {
  if (receipt?.status === "rejected") return { mode: "rebase", explanation: receipt.rejectionCode ?? "rejected" };
  return predicted === authoritative ? { mode: "none" } : { mode: "rebase" };
}
