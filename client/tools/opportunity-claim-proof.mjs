// GS5-A4 tooling oracle. The composed driver and its counterexample tests
// consume this one predicate; this is not product or payout implementation.
export function assertOpportunityClaimEffect(result, after) {
  const claim = result.body?.receipt?.opportunity;
  if (result.intent?.kind !== "claim_opportunity" || !claim?.effect_row_id) throw new Error(`GS5 claim receipt carried no opportunity evidence: ${JSON.stringify(result.body)}`);
  const arm = after?.features?.opportunity;
  const lucky = claim.effect_row_id === "active.lucky" && typeof claim.actual_credited_delta === "string";
  const buffed = typeof claim.buff_instance_id === "string" && arm?.buffs?.some((buff) => buff.buff_instance_id === claim.buff_instance_id);
  if (!lucky && !buffed) throw new Error(`GS5 claim effect is absent from the next snapshot: ${JSON.stringify({ claim, arm })}`);
  return claim.effect_row_id === "active.lucky" ? "lucky" : "buff";
}
