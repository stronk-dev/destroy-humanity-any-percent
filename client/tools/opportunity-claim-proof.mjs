// GS5-A4 tooling oracle. The composed driver and its counterexample tests
// consume this one predicate; this is not product or payout implementation.
import { isDeepStrictEqual } from "node:util";

export function assertQueuedOpportunityClaim({ purchase, read, result, requests, pendingID, history }) {
  const fail = (detail) => { throw new Error(`GS5 queued claim proof: ${detail}`); };
  if (requests?.length !== 2 || requests[0].phase !== "purchase-held" || requests[1].phase !== "released" ||
      !isDeepStrictEqual(requests[0].intent, purchase?.intent) || !isDeepStrictEqual(requests[1].intent, result?.intent)) fail("not one purchase then one released Claim");
  if (purchase.intent.kind !== "buy_generator" || purchase.intent.generator_id !== "generator.beige_tower" ||
      purchase.intent.count?.mode !== "exact" || purchase.intent.count.value !== 1 ||
      purchase.body?.outcome !== "applied" || purchase.body.intent_id !== purchase.intent.intent_id ||
      !Number.isSafeInteger(purchase.intent.expected_revision) || purchase.intent.expected_revision < 1 ||
      purchase.body.new_revision !== purchase.intent.expected_revision + 1) fail("missing exact applied purchase");
  if (result?.intent?.kind !== "claim_opportunity" || result.intent.opportunity_id !== pendingID || !pendingID ||
      result.intent.intent_id === purchase.intent.intent_id || result.body?.outcome !== "applied" ||
      result.body.intent_id !== result.intent.intent_id || result.body.new_revision !== purchase.body.new_revision + 1 ||
      result.intent.expected_revision !== purchase.body.new_revision || read?.schema_version !== 4 ||
      read.revision !== purchase.body.new_revision || read.run?.run_seq !== purchase.body.snapshot?.run_seq ||
      result.body.snapshot?.run_seq !== read.run.run_seq) fail("Claim did not bind the refreshed purchase revision/run/opportunity");
  if (!Array.isArray(history) || history.length !== 2) fail("missing exact committed history");
  for (const [index, command] of [purchase, result].entries()) {
    const row = history[index], { intent_id: intentID, ...payload } = command.intent;
    if (row?.intent_id !== intentID || row.applied_revision !== command.body.new_revision ||
        row.run_seq !== read.run.run_seq || row.founder_id !== read.run.founder_id ||
        !isDeepStrictEqual(row.payload, payload) || !isDeepStrictEqual(row.receipt, command.body) ||
        !isDeepStrictEqual(row.record_receipt, command.body) ||
        !isDeepStrictEqual(row.balances, command.body.snapshot.balances) ||
        !isDeepStrictEqual(row.generators, command.body.snapshot.generators)) fail("SQL payload, receipt or saved state differs from the submitted command");
  }
  if (!history[0].stream_id || history[0].stream_id !== history[1].stream_id ||
      !Number.isSafeInteger(history[0].seq) || history[0].seq < 1 || history[1].seq !== history[0].seq + 1) fail("not consecutive commands in the same Company stream");
}

export function assertOpportunityReadStatus(status) {
  if (status !== 200) throw new Error(`GS5 successor read failed (${status})`);
}

export function assertOpportunityClaimEffect(result, after, parseCanonical) {
  const fail = (detail) => { throw new Error(`GS5 claim effect proof: ${detail}`); };
  const intent = result?.intent;
  const receipt = result?.body;
  const claim = receipt?.receipt?.opportunity;
  if (intent?.kind !== "claim_opportunity" || receipt?.outcome !== "applied" || !claim?.effect_row_id) fail("missing applied claim evidence");
  if (typeof intent.intent_id !== "string" || !intent.intent_id || receipt.intent_id !== intent.intent_id ||
      typeof intent.opportunity_id !== "string" || !intent.opportunity_id || claim.opportunity_id !== intent.opportunity_id) fail("unbound intent or opportunity");
  if (!Number.isSafeInteger(intent.expected_revision) || intent.expected_revision < 0 ||
      !Number.isSafeInteger(receipt.new_revision) || receipt.new_revision !== intent.expected_revision + 1 ||
      after?.schema_version !== 4 || after.revision !== receipt.new_revision) fail("not the exact post-command revision");
  if (!Number.isSafeInteger(receipt.snapshot?.run_seq) || receipt.snapshot.run_seq < 1 ||
      after.run?.run_seq !== receipt.snapshot.run_seq) fail("not the same run");
  const arm = after?.features?.opportunity;
  if (!arm || !Array.isArray(arm.buffs)) fail("missing authoritative opportunity arm");
  if (claim.effect_row_id === "active.lucky") {
    const nonnegative = (value) => {
      if (typeof value !== "string" || typeof parseCanonical !== "function") fail("missing canonical amount/parser");
      try {
        const parsed = parseCanonical(value);
        if (parsed.lt(0)) fail("negative Lucky amount");
        return parsed;
      }
      catch { fail("invalid Lucky amount"); }
    };
    const credited = nonnegative(claim.actual_credited_delta);
    const cash = receipt.snapshot?.balances?.["company.cash"];
    const bank = nonnegative(cash);
    const rows = Array.isArray(after.resources) ? after.resources.filter((row) => row?.resource_id === "company.cash") : [];
    // The receipt snapshot includes lazy accrual as well as the Lucky credit.
    // Reads project this persisted balance, not a separately advanced bank.
    if (rows.length !== 1 || rows[0].amount !== cash) fail("Lucky cash is absent from the next snapshot");
    if (credited.eq(0)) {
      // A zero-credit exception needs the actual published cap, not merely a
      // receipt's saturation assertion. Do not make all zeros look successful.
      if (claim.saturated !== true || claim.cap_reason_key !== "cap.cash" || rows[0].cap?.amount !== cash) fail("zero Lucky credit is not at its hardcap");
    } else {
      const changes = Array.isArray(receipt.receipt.changes) ? receipt.receipt.changes.filter((row) => row?.resource_id === "company.cash") : [];
      if (changes.length !== 1 || changes[0].after !== cash) fail("Lucky credit has no matching cash change");
      if (nonnegative(changes[0].delta).lt(credited) || !bank.gt(nonnegative(changes[0].before))) fail("cash change cannot contain the Lucky credit");
    }
    return "lucky";
  }
  if (!["active.building", "active.click", "active.production"].includes(claim.effect_row_id)) fail("unknown effect");
  if (typeof claim.buff_instance_id !== "string" || !claim.buff_instance_id) fail("missing buff ID");
  const buffs = arm.buffs.filter((buff) => buff?.buff_instance_id === claim.buff_instance_id);
  if (buffs.length !== 1 || buffs[0].effect_row_id !== claim.effect_row_id) fail("claim buff/effect is absent from the next snapshot");
  return "buff";
}
