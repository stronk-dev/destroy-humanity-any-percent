import { describe, expect, it } from "vitest";
import artifact from "../../testdata/axis-stack/activation-research-v1.json";
import { applyLogged, applyLoggedExit, canonicalJSONString, encodeReplayState, loadReplayCatalogBundle,
  restoreReplayState, withNextReplayCatalogBundle, type ReplayArtifacts, type ReplayCatalogBundle } from "../src/replay";

interface Action { name: string; pre_state: unknown; canonical_payload: unknown; replay_inputs: unknown;
  outcome: string; receipt_json: string; events_json: string; post_state_json: string }
interface Exit { name: string; pre_state: unknown; canonical_payload: unknown; replay_inputs: unknown; outcome: string;
  receipt_json: string; founder_output_json: string; final_company_json: string; new_company_json: string;
  founder_events_json: string; company_ended_events_json: string; company_started_events_json: string }
const corpus = artifact as unknown as { version: number; acceptance_status: string;
  current: { constants_hash: string; artifacts: ReplayArtifacts }; next: { constants_hash: string; artifacts: ReplayArtifacts };
  rows: { id: string; old: Action; exit: Exit; next: Action[] }[] };
const currentPromise = loadReplayCatalogBundle(corpus.current.constants_hash, corpus.current.artifacts);
const nextPromise = loadReplayCatalogBundle(corpus.next.constants_hash, corpus.next.artifacts);
function restore(raw: unknown, version: 18 | 19, bundle: ReplayCatalogBundle) {
  if (!bundle.meters || !bundle.achievements || !bundle.opportunities) throw new Error("activation foundation missing");
  return restoreReplayState(raw, version, bundle.economy, { meters: bundle.meters, achievements: bundle.achievements,
    doctrines: bundle.doctrines, opportunities: bundle.opportunities });
}
async function step(state: ReturnType<typeof restore>, row: Action, bundle: ReplayCatalogBundle, version: 18 | 19) {
  expect(canonicalJSONString(encodeReplayState(state))).toBe(canonicalJSONString(row.pre_state));
  const result = await applyLogged(state, canonicalJSONString(row.canonical_payload), bundle, row.replay_inputs);
  expect(result.outcome).toBe("applied"); expect(result.outcome).toBe(row.outcome);
  expect(canonicalJSONString(result.receipt)).toBe(row.receipt_json);
  expect(canonicalJSONString(result.events)).toBe(row.events_json);
  expect(canonicalJSONString(encodeReplayState(result.state))).toBe(row.post_state_json);
  const restored = restore(encodeReplayState(result.state), version, bundle);
  expect(canonicalJSONString(encodeReplayState(restored))).toBe(row.post_state_json);
  return restored;
}
describe("CV4/AC7 pinned activation, not full release acceptance", () => {
  it("pins complete six populations and unchanged served epoch identity", () => {
    expect(corpus.version).toBe(1);
    expect(corpus.acceptance_status).toBe("NOT_PROVEN: bounded pinned activation; production AC6 remains red");
    expect(corpus.current.constants_hash).toBe("sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6");
    expect(corpus.next.constants_hash).not.toBe(corpus.current.constants_hash);
    const ids = ["fresh", "veteran"].flatMap(profile => [0, 3114, 90000000].map(gap => `${profile}/online/${gap}`));
    expect(corpus.rows.map(row => row.id)).toEqual(ids);
    expect(corpus.rows.map(row => row.old.name)).toEqual(ids.map(id => `${id}/old-manual`));
    expect(corpus.rows.map(row => row.exit.name)).toEqual(ids);
    for (const row of corpus.rows) expect(row.next.map(action => action.name)).toEqual([`${row.id}/next-manual`, `${row.id}/next-purchase`]);
  });
  it.each(corpus.rows)("compares all four complete transitions continuously for $id", async row => {
    const current = await currentPromise, next = await nextPromise;
    let company = await step(restore(row.old.pre_state, 18, current), row.old, current, 18);
    expect(company.wireVersion).toBe(18); expect(company.achievementsAttainedRun).toBeNull(); expect(company.attainmentScoreRun).toBe(0);
    expect(canonicalJSONString(encodeReplayState(company))).toBe(canonicalJSONString(row.exit.pre_state));
    const result = await applyLoggedExit(company, canonicalJSONString(row.exit.canonical_payload), withNextReplayCatalogBundle(current, next), row.exit.replay_inputs);
    expect(result.outcome).toBe("applied"); expect(result.outcome).toBe(row.exit.outcome);
    expect(canonicalJSONString(result.receipt)).toBe(row.exit.receipt_json);
    expect(canonicalJSONString(result.founder)).toBe(row.exit.founder_output_json);
    expect(canonicalJSONString(encodeReplayState(result.finalCompany))).toBe(row.exit.final_company_json);
    expect(canonicalJSONString(encodeReplayState(restore(encodeReplayState(result.finalCompany), 18, current)))).toBe(row.exit.final_company_json);
    expect(canonicalJSONString(encodeReplayState(result.newCompany!))).toBe(row.exit.new_company_json);
    expect(canonicalJSONString(result.founderEvents)).toBe(row.exit.founder_events_json);
    expect(canonicalJSONString(result.companyEndedEvents)).toBe(row.exit.company_ended_events_json);
    expect(canonicalJSONString(result.companyStartedEvents)).toBe(row.exit.company_started_events_json);
    company = restore(encodeReplayState(result.newCompany!), 19, next);
    expect(company.wireVersion).toBe(19); expect(company.runSeq).toBe(3);
    expect(company.achievementsAttainedRun?.size).toBe(0); expect(company.attainmentScoreRun).toBe(0);
    expect(company.achievementScoreRun).toBe(0); expect(company.achievementsEarnedRun.size).toBe(0);
    for (const action of row.next) company = await step(company, action, next, 19);
    expect(company.generatorPurchasedTotal).toBe(1); expect(company.attainmentScoreRun).toBe(2);
    expect([...company.achievementsAttainedRun!]).toEqual(["achievement.generators_purchased_1"]);
    expect(company.achievementScoreRun).toBe(0); expect(company.achievementsEarnedRun.size).toBe(0);
  });
});
