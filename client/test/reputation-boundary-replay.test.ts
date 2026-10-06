import { describe, expect, it } from "vitest";

import corpus from "../../testdata/replay/reputation-exit-boundary-v1.json";
import {
  applyFounderLogged, applyLoggedExit, canonicalJSONString, encodeFounderReplayState, encodeReplayState,
  loadReplayCatalogBundle, restoreFounderReplayState, restoreReplayState, withNextReplayCatalogBundle,
  type ReplayArtifacts, type ReplayCatalogBundle,
} from "../src/replay";

// R8 diagnostic replay only: no SQL transaction, live producer or natural pacing claim.
const commands = ["wind_down", "acquihire", "acquisition"];
const profiles = ["activate-plan", "activate-absent", "activate-empty", "next-inactive", "unknown-prefix", "owned-prefix", "requires-prefix", "unaffordable-prefix", "inactive-no-plan"];
const refusals: Record<string, readonly [string, string]> = {
  "next-inactive": ["not_eligible", "reputation_plan.tree_inactive"],
  "unknown-prefix": ["unknown_id", "reputation_plan.unknown_id"],
  "owned-prefix": ["not_eligible", "reputation_plan.owned"],
  "requires-prefix": ["not_eligible", "reputation_plan.requires"],
  "unaffordable-prefix": ["unaffordable", "reputation_plan.reputation"],
};
const sources = corpus.bundles as Record<string, { constants_hash: string; artifacts: Record<string, string> }>;
const bundles = new Map(Object.entries(sources).map(([name, source]) => [name, loadReplayCatalogBundle(source.constants_hash, source.artifacts as unknown as ReplayArtifacts)]));

async function catalogs(row: typeof corpus.cases[number]) {
  const current = await bundles.get(row.current)!;
  const next = await bundles.get(row.next)!;
  return { current, next, linked: row.current === row.next ? current : withNextReplayCatalogBundle(current, next) };
}
function restoreCompany(state: unknown, bundle: ReplayCatalogBundle) {
  if (!bundle.meters || !bundle.achievements) throw new Error("boundary population requires pinned foundations");
  return restoreReplayState(state, 18, bundle.economy, { meters: bundle.meters, achievements: bundle.achievements,
    doctrines: bundle.doctrines, opportunities: bundle.opportunities });
}

describe("Reputation Exit boundary Go/TS corpus", () => {
  it("pins all27 paired profiles and12 applied/15 refused observations", () => {
    expect(corpus.version).toBe(1);
    expect(corpus.cases.map((row) => row.name)).toEqual(commands.flatMap((command) => profiles.map((profile) => `${command}-${profile}`)));
    expect(corpus.cases.filter((row) => row.company.outcome === "applied")).toHaveLength(12);
    expect(corpus.cases.filter((row) => row.company.outcome === "rejected")).toHaveLength(15);
    expect(corpus.cases.filter((row) => row.current !== row.next)).toHaveLength(9);
    expect(corpus.cases.filter((row) => row.profile === "inactive-no-plan")).toHaveLength(3);
    expect(sources.live!.constants_hash).not.toBe(sources.tree!.constants_hash);
  });

  it.each(corpus.cases)("byte-compares complete Company/Founder replay for $name", async (row) => {
    const { current, next, linked } = await catalogs(row);
    const companyCase = row.company, founderCase = row.founder;
    const company = restoreCompany(companyCase.pre_state, current);
    const result = await applyLoggedExit(company, canonicalJSONString(companyCase.canonical_payload), linked, companyCase.replay_inputs);
    expect(result.outcome).toBe(companyCase.outcome);
    expect(canonicalJSONString(result.receipt)).toBe(companyCase.receipt_json);
    expect(canonicalJSONString(result.founder)).toBe(companyCase.founder_output_json);
    expect(canonicalJSONString(encodeReplayState(result.finalCompany))).toBe(companyCase.final_company_json);
    expect(canonicalJSONString(result.newCompany === null ? null : encodeReplayState(result.newCompany))).toBe(companyCase.new_company_json);
    expect(canonicalJSONString(result.founderEvents)).toBe(companyCase.founder_events_json);
    expect(canonicalJSONString(result.companyEndedEvents)).toBe(companyCase.company_ended_events_json);
    expect(canonicalJSONString(result.companyStartedEvents)).toBe(companyCase.company_started_events_json);
    const founder = restoreFounderReplayState(founderCase.pre_state, founderCase.state_version, current);
    expect(founder.reputationLevel).toBe(6);
    const transition = await applyFounderLogged(founder, canonicalJSONString(founderCase.canonical_payload), linked, founderCase.replay_inputs);
    expect(transition.outcome).toBe(founderCase.outcome);
    expect(canonicalJSONString(transition.receipt)).toBe(founderCase.receipt_json);
    expect(canonicalJSONString(transition.events)).toBe(founderCase.events_json);
    expect(canonicalJSONString(encodeFounderReplayState(transition.state))).toBe(founderCase.post_state_json);
    const failure = refusals[row.profile];
    if (failure) {
      expect(result.outcome).toBe("rejected");
      expect(result.receipt).toMatchObject({ rejection: { category: failure[0], detail: failure[1] } });
      expect(transition.receipt).toMatchObject({ outcome: "rejected", rejection: { category: failure[0], detail: failure[1] } });
      expect(transition.resultConstantsHash).toBe(current.constantsHash);
      expect(canonicalJSONString(encodeReplayState(company))).toBe(canonicalJSONString(companyCase.pre_state));
      expect(canonicalJSONString(encodeFounderReplayState(transition.state))).toBe(canonicalJSONString(founderCase.pre_state));
      expect(result.newCompany).toBeNull();
      expect(result.founderEvents).toEqual([]);
      expect(result.companyEndedEvents).toEqual([]);
      expect(result.companyStartedEvents).toEqual([]);
      expect(transition.events).toEqual([]);
      return;
    }
    const planned = row.profile === "activate-plan", activated = row.current !== row.next;
    expect(result.outcome).toBe("applied");
    expect(transition.resultConstantsHash).toBe(next.constantsHash);
    expect(transition.state.wireVersion).toBe(activated ? 22 : 21);
    expect(transition.state.reputationLevel).toBe(6);
    expect(transition.state.reputationSpent).toBe(planned ? 6 : 0);
    expect(transition.state.reputationUnlockPpm).toBe(planned ? 50000 : 0);
    expect(transition.state.reputationNodesOwned).toEqual(planned ? ["reputation.starter.cash_small", "reputation.starter.generated_beige_tower", "reputation.unlock.p05"] : []);
    expect(result.newCompany!.runSeq).toBe(3);
    expect(result.newCompany!.offerState).toBeNull();
    expect(result.companyStartedEvents).toHaveLength(1);
    if (activated) {
      expect(result.companyStartedEvents[0]).toMatchObject({ schema_version: 2, payload: { reputation_tree: {
        bonus_factor: planned ? "1.003e0" : "1e0",
        applied_starter_node_ids: planned ? ["reputation.starter.cash_small", "reputation.starter.generated_beige_tower"] : [],
      } } });
    } else {
      expect((result.companyStartedEvents[0]!.payload as Record<string, unknown>).reputation_tree ?? null).toBeNull();
    }
    if (planned) {
      expect(result.newCompany!.balances["company.cash"]).toBe("1e3");
      expect(result.newCompany!.generatorsProvisioned["generator.beige_tower"]).toBe(5);
      expect(result.newCompany!.generators["generator.beige_tower"]).toBe(0);
      expect(result.newCompany!.generatorPurchasedTotal).toBe(0);
      expect(result.founderEvents.map((event) => event.kind)).toEqual(["founder_advanced", "reputation_node_purchased.v1", "reputation_node_purchased.v1", "reputation_node_purchased.v1"]);
      expect(result.founderEvents.slice(1).map((event) => event.payload)).toMatchObject([
        { node_id: "reputation.unlock.p05", cost: 1, source: "exit_plan" },
        { node_id: "reputation.starter.cash_small", cost: 2, source: "exit_plan" },
        { node_id: "reputation.starter.generated_beige_tower", cost: 3, source: "exit_plan" },
      ]);
    }
  });

  it.each(commands)("preserves absent/empty %s identity without changing semantics", (command) => {
    const absent = corpus.cases.find((row) => row.name === `${command}-activate-absent`)!.company;
    const empty = corpus.cases.find((row) => row.name === `${command}-activate-empty`)!.company;
    expect(canonicalJSONString(absent.canonical_payload)).not.toBe(canonicalJSONString(empty.canonical_payload));
    for (const key of ["receipt_json", "founder_output_json", "final_company_json", "new_company_json", "founder_events_json", "company_ended_events_json", "company_started_events_json"] as const) expect(absent[key], key).toBe(empty[key]);
  });

  it.each(corpus.cases.filter((row) => row.company.outcome === "rejected" || row.current !== row.next))("refuses copied Founder evidence for $name", async (row) => {
    const { current, linked } = await catalogs(row);
    const source = row.founder;
    const inputs = structuredClone(source.replay_inputs) as { resolved: { reputation_delta: number; result_constants_hash: string; reputation_purchases?: { resolved_cost: number }[] } };
    if (row.company.outcome === "rejected") inputs.resolved.reputation_delta = 1;
    else inputs.resolved.result_constants_hash = "0".repeat(64);
    const founder = restoreFounderReplayState(source.pre_state, source.state_version, current);
    await expect(applyFounderLogged(founder, canonicalJSONString(source.canonical_payload), linked, inputs)).rejects.toThrow();
    if (row.current !== row.next) {
      const pinInputs = structuredClone(source.replay_inputs) as typeof inputs;
      pinInputs.resolved.result_constants_hash = current.constantsHash;
      const fresh = restoreFounderReplayState(source.pre_state, source.state_version, current);
      await expect(applyFounderLogged(fresh, canonicalJSONString(source.canonical_payload), linked, pinInputs)).rejects.toThrow();
    }
    if (row.profile === "activate-plan") {
      const costInputs = structuredClone(source.replay_inputs) as typeof inputs;
      costInputs.resolved.reputation_purchases![0]!.resolved_cost = 2;
      const fresh = restoreFounderReplayState(source.pre_state, source.state_version, current);
      await expect(applyFounderLogged(fresh, canonicalJSONString(source.canonical_payload), linked, costInputs)).rejects.toThrow();
    }
  });
});
