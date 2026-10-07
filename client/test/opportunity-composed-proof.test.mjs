import { expect, test } from "vitest";
import { assertOpportunityClaimEffect } from "../tools/opportunity-claim-proof.mjs";

function population(effect = "active.lucky", credited = "2e1") {
  const isLucky = effect === "active.lucky";
  const result = {
    intent: { kind: "claim_opportunity", intent_id: "claim-1", expected_revision: 11, opportunity_id: "opportunity-1" },
    body: { outcome: "applied", intent_id: "claim-1", new_revision: 12,
      receipt: { opportunity: { opportunity_id: "opportunity-1", effect_row_id: effect,
        actual_credited_delta: isLucky ? credited : null, buff_instance_id: isLucky ? null : "buff-1",
        requested_delta: isLucky ? "2e1" : null, saturated: isLucky ? credited === "0" : null,
        cap_reason_key: isLucky && credited === "0" ? "cap.cash" : null } },
      snapshot: { run_seq: 1, balances: { "company.cash": "1.2e2" } },
    },
  };
  const after = { schema_version: 4, revision: 12, run: { run_seq: 1 },
    resources: [{ resource_id: "company.cash", amount: "1.2e2" }],
    features: { opportunity: { buffs: isLucky ? [] : [{ buff_instance_id: "buff-1", effect_row_id: effect }] } },
  };
  return { result, after };
}

for (const [effect, credited] of [["active.lucky", "2e1"], ["active.lucky", "0"],
  ["active.production", null], ["active.click", null], ["active.building", null]]) {
  test(`admits matching post-command ${effect} evidence (${credited ?? "buff"})`, () => {
    const { result, after } = population(effect, credited);
    expect(assertOpportunityClaimEffect(result, after)).toBe(effect === "active.lucky" ? "lucky" : "buff");
  });
}

for (const [name, corrupt] of [
  ["null Lucky successor", (p) => { p.after = null; }],
  ["absent cash row", (p) => { p.after.resources = []; }],
  ["duplicate cash row", (p) => { p.after.resources.push({ ...p.after.resources[0] }); }],
  ["old uncredited cash", (p) => { p.after.resources[0].amount = "1e2"; }],
  ["missing receipt cash", (p) => { delete p.result.body.snapshot.balances["company.cash"]; }],
  ["noncanonical receipt cash", (p) => { p.result.body.snapshot.balances["company.cash"] = "120"; p.after.resources[0].amount = "120"; }],
  ["NaN credited delta", (p) => { p.result.body.receipt.opportunity.actual_credited_delta = "NaN"; }],
  ["negative credited delta", (p) => { p.result.body.receipt.opportunity.actual_credited_delta = "-2e1"; }],
  ["numeric credited delta", (p) => { p.result.body.receipt.opportunity.actual_credited_delta = 20; }],
  ["stale successor revision", (p) => { p.after.revision = 11; }],
  ["unrelated newer successor", (p) => { p.after.revision = 13; }],
  ["non-incrementing receipt", (p) => { p.result.body.new_revision = 11; p.after.revision = 11; }],
  ["unsafe expected revision", (p) => { p.result.intent.expected_revision = Number.MAX_SAFE_INTEGER + 1; }],
  ["wrong intent binding", (p) => { p.result.body.intent_id = "claim-other"; }],
  ["wrong opportunity binding", (p) => { p.result.body.receipt.opportunity.opportunity_id = "opportunity-other"; }],
  ["rejected receipt", (p) => { p.result.body.outcome = "rejected"; }],
  ["wrong run binding", (p) => { p.after.run.run_seq = 2; }],
  ["wrong snapshot generation", (p) => { p.after.schema_version = 3; }],
  ["missing arm", (p) => { p.after.features.opportunity = null; }],
]) {
  test(`refuses ${name} instead of counting receipt-only Lucky success`, () => {
    const p = population(); corrupt(p);
    expect(() => assertOpportunityClaimEffect(p.result, p.after)).toThrow();
  });
}

for (const [name, corrupt] of [
  ["missing buff", (p) => { p.after.features.opportunity.buffs = []; }],
  ["unrelated buff ID", (p) => { p.after.features.opportunity.buffs[0].buff_instance_id = "buff-other"; }],
  ["buff with wrong effect", (p) => { p.after.features.opportunity.buffs[0].effect_row_id = "active.click"; }],
  ["duplicate buff ID", (p) => { p.after.features.opportunity.buffs.push({ ...p.after.features.opportunity.buffs[0] }); }],
  ["unknown effect", (p) => { p.result.body.receipt.opportunity.effect_row_id = "active.unknown"; p.after.features.opportunity.buffs[0].effect_row_id = "active.unknown"; }],
]) {
  test(`refuses ${name} in the buff arm`, () => {
    const p = population("active.production"); corrupt(p);
    expect(() => assertOpportunityClaimEffect(p.result, p.after)).toThrow();
  });
}
