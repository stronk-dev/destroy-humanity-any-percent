import { describe, expect, it } from "vitest";

import corpus from "../../testdata/replay/reputation-offer-plan-v1.json";
import {
  applyFounderLogged, applyLoggedExit, canonicalJSONString,
  encodeFounderReplayState, encodeReplayState, loadReplayCatalogBundle,
  restoreFounderReplayState, restoreReplayState, type ReplayArtifacts, type ReplayCatalogBundle,
} from "../src/replay";

// R8 supplement, Go-authored only. Stored diagnostic offers are NOT live
// generation, SQL transaction/retry, browser or naturally earned pacing proof.
const catalogs = loadReplayCatalogBundle(corpus.bundle.constants_hash, corpus.bundle.artifacts as unknown as ReplayArtifacts);
const planCases = corpus.cases.filter((row) => row.name.endsWith("payout-funded") || row.name.endsWith("promise-floor"));

function restoreCompany(state: unknown, bundle: ReplayCatalogBundle) {
  if (!bundle.meters || !bundle.achievements) throw new Error("offered replay population requires pinned foundation catalogs");
  return restoreReplayState(state, 18, bundle.economy, { meters: bundle.meters, achievements: bundle.achievements,
    doctrines: bundle.doctrines, opportunities: bundle.opportunities });
}

describe("offered Reputation-plan Go/TS replay parity", () => {
  it("pins both kinds and all ten Company/ten Founder arms", () => {
    expect(corpus.version).toBe(1);
    expect(corpus.cases.map((row) => row.name)).toEqual([
      "acquihire-payout-funded", "acquihire-last-unaffordable", "acquihire-absent", "acquihire-empty", "acquihire-promise-floor",
      "acquisition-payout-funded", "acquisition-last-unaffordable", "acquisition-absent", "acquisition-empty", "acquisition-promise-floor",
    ]);
    expect(corpus.cases.filter((row) => row.company.outcome === "applied")).toHaveLength(8);
    expect(corpus.cases.filter((row) => row.founder !== null)).toHaveLength(10);
    expect(planCases).toHaveLength(4);
  });

  it.each(corpus.cases)("byte-compares both replay arms for $name", async (row) => {
    const bundle = await catalogs;
    const testCase = row.company;
    const state = restoreCompany(testCase.pre_state, bundle);
    const transition = await applyLoggedExit(state, canonicalJSONString(testCase.canonical_payload), bundle, testCase.replay_inputs);
    expect(transition.outcome).toBe(testCase.outcome);
    expect(canonicalJSONString(transition.receipt)).toBe(testCase.receipt_json);
    expect(canonicalJSONString(transition.founder)).toBe(testCase.founder_output_json);
    expect(canonicalJSONString(encodeReplayState(transition.finalCompany))).toBe(testCase.final_company_json);
    expect(canonicalJSONString(transition.newCompany === null ? null : encodeReplayState(transition.newCompany))).toBe(testCase.new_company_json);
    expect(canonicalJSONString(transition.founderEvents)).toBe(testCase.founder_events_json);
    expect(canonicalJSONString(transition.companyEndedEvents)).toBe(testCase.company_ended_events_json);
    expect(canonicalJSONString(transition.companyStartedEvents)).toBe(testCase.company_started_events_json);
    const founderCase = row.founder!;
    const founder = restoreFounderReplayState(founderCase.pre_state, founderCase.state_version, bundle);
    expect(founder.reputationLevel).toBe(0);
    expect(founder.reputationSpent).toBe(0);
    const result = await applyFounderLogged(founder, canonicalJSONString(founderCase.canonical_payload), bundle, founderCase.replay_inputs);
    expect(result.outcome).toBe(founderCase.outcome);
    expect(result.resultConstantsHash).toBe(corpus.bundle.constants_hash);
    expect(canonicalJSONString(result.receipt)).toBe(founderCase.receipt_json);
    expect(canonicalJSONString(result.events)).toBe(founderCase.events_json);
    expect(canonicalJSONString(encodeFounderReplayState(result.state))).toBe(founderCase.post_state_json);
    if (testCase.outcome === "rejected") {
      expect(transition.receipt).toMatchObject({ outcome: "rejected", rejection: { category: "unaffordable", detail: "reputation_plan.reputation" } });
      expect(canonicalJSONString(encodeReplayState(state))).toBe(canonicalJSONString(testCase.pre_state));
      expect(transition.newCompany).toBeNull();
      expect(transition.founderEvents).toEqual([]);
      expect(transition.companyEndedEvents).toEqual([]);
      expect(transition.companyStartedEvents).toEqual([]);
      expect(result.outcome).toBe("rejected");
      expect(result.events).toEqual([]);
      expect(canonicalJSONString(encodeFounderReplayState(result.state))).toBe(canonicalJSONString(founderCase.pre_state));
      return;
    }
    expect(transition.founder.reputation_level).toBe(row.kind === "acquihire" ? 18 : 20);
    expect(transition.newCompany!.runSeq).toBe(3);
    expect(transition.newCompany!.offerState).toBeNull();
    const funded = row.name.endsWith("payout-funded") || row.name.endsWith("promise-floor");
    expect(transition.founder.founder_extensions!.reputation_spent).toBe(funded ? 6 : 0);
    expect(transition.founder.founder_extensions!.reputation_nodes_owned).toEqual(funded ? [
      "reputation.starter.cash_small", "reputation.starter.generated_beige_tower", "reputation.unlock.p05",
    ] : []);
    expect(transition.companyStartedEvents[0]).toMatchObject({ kind: "run_started", schema_version: 2, payload: {
      reputation_tree: { bonus_factor: funded ? (row.kind === "acquihire" ? "1.009e0" : "1.01e0") : "1e0",
        applied_starter_node_ids: funded ? ["reputation.starter.cash_small", "reputation.starter.generated_beige_tower"] : [] },
    } });
    if (funded) {
      expect(transition.newCompany!.balances["company.cash"]).toBe("1e3");
      expect(transition.newCompany!.generatorsProvisioned["generator.beige_tower"]).toBe(5);
      expect(transition.newCompany!.generators["generator.beige_tower"]).toBe(0);
      expect(transition.newCompany!.generatorPurchasedTotal).toBe(0);
      expect(transition.founderEvents.map((event) => event.kind)).toEqual([
        "founder_advanced", "reputation_node_purchased.v1", "reputation_node_purchased.v1", "reputation_node_purchased.v1",
      ]);
      expect(transition.founderEvents.slice(1).map((event) => event.payload)).toMatchObject([
        { node_id: "reputation.unlock.p05", cost: 1, source: "exit_plan" },
        { node_id: "reputation.starter.cash_small", cost: 2, source: "exit_plan" },
        { node_id: "reputation.starter.generated_beige_tower", cost: 3, source: "exit_plan" },
      ]);
    }
    const resolvedIndex = transition.companyEndedEvents.findIndex((event) => event.kind === "exit_offer_resolved");
    const endedIndex = transition.companyEndedEvents.findIndex((event) => event.kind === "run_ended");
    expect(resolvedIndex).toBeGreaterThanOrEqual(0);
    expect(endedIndex).toBeGreaterThan(resolvedIndex);
    expect(transition.companyEndedEvents[resolvedIndex]!.payload).toEqual({ offer_id: "01986666-d001-7000-8000-000000000001", resolution: "accepted" });
    expect(result.outcome).toBe("applied");
  });

  it.each(["acquihire", "acquisition"])("keeps absent/empty %s plan outputs equal but requests distinct", (kind) => {
    const absent = corpus.cases.find((row) => row.name === `${kind}-absent`)!.company;
    const empty = corpus.cases.find((row) => row.name === `${kind}-empty`)!.company;
    expect(canonicalJSONString(absent.canonical_payload)).not.toBe(canonicalJSONString(empty.canonical_payload));
    for (const key of ["receipt_json", "founder_output_json", "final_company_json", "new_company_json", "founder_events_json", "company_ended_events_json", "company_started_events_json"] as const) {
      expect(absent[key], key).toBe(empty[key]);
    }
  });

  it.each(planCases)("refuses copied cost, promise and prerequisite-order corruption for $name", async (row) => {
    const bundle = await catalogs;
    const founderCase = row.founder!;
    const founderInputs = structuredClone(founderCase.replay_inputs) as { resolved: { reputation_purchases: { resolved_cost: number }[] } };
    founderInputs.resolved.reputation_purchases[0]!.resolved_cost = 2;
    const founder = restoreFounderReplayState(founderCase.pre_state, 22, bundle);
    await expect(applyFounderLogged(founder, canonicalJSONString(founderCase.canonical_payload), bundle, founderInputs)).rejects.toThrow();
    const companyInputs = structuredClone(row.company.replay_inputs) as { resolved: { selected_terms: { payout_preview: { reputation_delta: number } } } };
    companyInputs.resolved.selected_terms.payout_preview.reputation_delta = row.level + 1;
    const company = restoreCompany(row.company.pre_state, bundle);
    await expect(applyLoggedExit(company, canonicalJSONString(row.company.canonical_payload), bundle, companyInputs)).rejects.toThrow();
    expect(canonicalJSONString(encodeReplayState(company))).toBe(canonicalJSONString(row.company.pre_state));
    const payload = structuredClone(row.company.canonical_payload) as Record<string, unknown> & { reputation_plan: string[] };
    [payload.reputation_plan[1], payload.reputation_plan[2]] = [payload.reputation_plan[2]!, payload.reputation_plan[1]!];
    const reordered = await applyLoggedExit(company, canonicalJSONString(payload), bundle, row.company.replay_inputs);
    expect(reordered.receipt).toMatchObject({ outcome: "rejected", rejection: { category: "not_eligible", detail: "reputation_plan.requires" } });
    expect(canonicalJSONString(encodeReplayState(company))).toBe(canonicalJSONString(row.company.pre_state));
    expect(reordered.founderEvents).toEqual([]);
    expect(reordered.companyEndedEvents).toEqual([]);
    expect(reordered.companyStartedEvents).toEqual([]);
    expect(reordered.newCompany).toBeNull();
  });

  it.each(corpus.cases.filter((row) => row.company.outcome === "rejected"))("refuses copied credited delta on rejected Founder arm for $name", async (row) => {
    const bundle = await catalogs;
    const founderCase = row.founder!;
    const inputs = structuredClone(founderCase.replay_inputs) as { resolved: { reputation_delta: number } };
    inputs.resolved.reputation_delta = 1;
    const founder = restoreFounderReplayState(founderCase.pre_state, 22, bundle);
    await expect(applyFounderLogged(founder, canonicalJSONString(founderCase.canonical_payload), bundle, inputs)).rejects.toThrow();
  });
});
