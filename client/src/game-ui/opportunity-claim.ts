import type { CopyKey } from "../copy";

// GS5: what the Desk region shows about the last applied claim. Read from the
// applied receipt's `receipt.opportunity` evidence; nothing is inferred.
export interface LastClaim { readonly effectRowID: string; readonly credited: string | null; readonly saturated: boolean; readonly capReasonKey: string | null }

export const OPPORTUNITY_REJECTIONS: ReadonlyMap<string, CopyKey> = new Map<string, CopyKey>([
  ["not_eligible/opportunity_expired", "desk.opportunity.rejection.expired"],
  ["not_eligible/opportunity_not_pending", "desk.opportunity.rejection.not_pending"],
  ["unknown_id/opportunity_id", "desk.opportunity.rejection.not_pending"],
]);

export function lastClaimFromReceipt(receipt: Readonly<Record<string, unknown>>): LastClaim {
  const inner = receipt.receipt;
  const claim = inner !== null && typeof inner === "object" && !Array.isArray(inner) ? (inner as Record<string, unknown>).opportunity : undefined;
  if (claim === null || typeof claim !== "object" || Array.isArray(claim)) throw new SyntaxError("applied claim receipt has no opportunity evidence");
  const row = claim as Record<string, unknown>;
  if (typeof row.effect_row_id !== "string") throw new SyntaxError("claim evidence has no effect row");
  const credited = typeof row.actual_credited_delta === "string" ? row.actual_credited_delta : null;
  const capReasonKey = typeof row.cap_reason_key === "string" ? row.cap_reason_key : null;
  return Object.freeze({ effectRowID: row.effect_row_id, credited, saturated: row.saturated === true, capReasonKey });
}
