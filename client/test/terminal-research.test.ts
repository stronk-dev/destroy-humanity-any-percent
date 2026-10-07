import { describe, expect, it } from "vitest";
import artifact from "../../testdata/axis-stack/terminal-research-v1.json";
import oldArtifact from "../../testdata/replay/apply-logged-v1.json";
import { applyLogged, applyLoggedExit, canonicalJSONString, encodeReplayState,
  loadReplayCatalogBundle, restoreReplayState, withNextReplayCatalogBundle, type ReplayArtifacts, type ReplayCatalogBundle } from "../src/replay";

interface TerminalCase {
  name: string; pre_state: unknown; canonical_payload: unknown; replay_inputs: unknown; outcome: string;
  receipt_json: string; founder_output_json: string; final_company_json: string; new_company_json: string;
  founder_events_json: string; company_ended_events_json: string; company_started_events_json: string;
}
interface Row {
  profile: { id: string };
  exit: TerminalCase;
  next: { name: string; pre_state: unknown; canonical_payload: unknown; replay_inputs: unknown; outcome: string;
    receipt_json: string; events_json: string; post_state_json: string };
}
const corpus = artifact as unknown as { version: number; acceptance_status: string; negative_cases: number;
  bundle: { constants_hash: string; artifacts: ReplayArtifacts }; rows: Row[] };
const catalogs = loadReplayCatalogBundle(corpus.bundle.constants_hash, corpus.bundle.artifacts);
function restore(state: unknown, bundle: ReplayCatalogBundle) {
  if (!bundle.meters || !bundle.achievements || !bundle.opportunities) throw new Error("terminal foundation catalogs missing");
  return restoreReplayState(state, 19, bundle.economy, { meters: bundle.meters, achievements: bundle.achievements,
    doctrines: bundle.doctrines, opportunities: bundle.opportunities });
}
async function checkNext(state: ReturnType<typeof restore>, row: Row, bundle: ReplayCatalogBundle) {
  expect(canonicalJSONString(encodeReplayState(state))).toBe(canonicalJSONString(row.next.pre_state));
  const result = await applyLogged(state, canonicalJSONString(row.next.canonical_payload), bundle, row.next.replay_inputs);
  expect(result.outcome).toBe("applied");
  expect(result.outcome).toBe(row.next.outcome);
  expect(canonicalJSONString(result.receipt)).toBe(row.next.receipt_json);
  expect(canonicalJSONString(result.events)).toBe(row.next.events_json);
  expect(canonicalJSONString(encodeReplayState(result.state))).toBe(row.next.post_state_json);
  expect(canonicalJSONString(encodeReplayState(restore(encodeReplayState(result.state), bundle)))).toBe(row.next.post_state_json);
}

describe("CV4 actual terminal/next-run research, NOT AC6 acceptance", () => {
  it("pins exactly six pairs/three refusals and non-acceptance", () => {
    const expected = [0, 3114, 120001].flatMap(elapsed => ["quiet", "combined"].map(profile => `${profile}/online/${elapsed}`));
    expect(corpus.version).toBe(1);
    expect(corpus.acceptance_status).toBe("NOT_PROVEN: bounded terminal research; production AC6 remains red");
    expect(corpus.rows.map(row => row.profile.id)).toEqual(expected);
    expect(corpus.rows.map(row => row.exit.name)).toEqual(expected);
    expect(corpus.rows.map(row => row.next.name)).toEqual(expected.map(id => `${id}/next-manual`));
    expect(corpus.negative_cases).toBe(3);
  });

  it.each(corpus.rows)("compares complete terminal outputs and continues for $profile.id", async row => {
    const bundle = await catalogs;
    const state = restore(row.exit.pre_state, bundle);
    expect(canonicalJSONString(encodeReplayState(state))).toBe(canonicalJSONString(row.exit.pre_state));
    const result = await applyLoggedExit(state, canonicalJSONString(row.exit.canonical_payload), bundle, row.exit.replay_inputs);
    expect(result.outcome).toBe("applied");
    expect(result.outcome).toBe(row.exit.outcome);
    expect(canonicalJSONString(result.receipt)).toBe(row.exit.receipt_json);
    expect(canonicalJSONString(result.founder)).toBe(row.exit.founder_output_json);
    expect(canonicalJSONString(encodeReplayState(result.finalCompany))).toBe(row.exit.final_company_json);
    expect(canonicalJSONString(encodeReplayState(restore(encodeReplayState(result.finalCompany), bundle)))).toBe(row.exit.final_company_json);
    expect(canonicalJSONString(encodeReplayState(result.newCompany!))).toBe(row.exit.new_company_json);
    expect(canonicalJSONString(result.founderEvents)).toBe(row.exit.founder_events_json);
    expect(canonicalJSONString(result.companyEndedEvents)).toBe(row.exit.company_ended_events_json);
    expect(canonicalJSONString(result.companyStartedEvents)).toBe(row.exit.company_started_events_json);
    const next = result.newCompany!;
    expect(next.wireVersion).toBe(19);
    expect(next.runSeq).toBe(3);
    expect(next.achievementsAttainedRun?.size).toBe(0);
    expect(next.attainmentScoreRun).toBe(0);
    expect(next.pendingOpportunity).toBeNull();
    expect(next.activeBuffs).toEqual([]);
    expect(next.opportunitySpawnSeq).toBe(0);
    expect(next.nextOpportunityAttendedMs).toBeGreaterThan(0);
    await checkNext(restore(encodeReplayState(next), bundle), row, bundle);
  });

  // This direct arm remains useful when TS Exit fails. It does NOT establish
  // that the browser produced this new run: the input is the Go-produced state.
  it.each(corpus.rows)("compares direct restored Go next-run outputs for $profile.id", async row => {
    const bundle = await catalogs;
    await checkNext(restore(row.next.pre_state, bundle), row, bundle);
  });

  it.each(["current", "next", "sequence"])("refuses copied terminal $0 unchanged", async fault => {
    const row = corpus.rows[0]!;
    expect(row.profile.id).toBe("quiet/online/0");
    const bundle = await catalogs;
    const state = restore(row.exit.pre_state, bundle);
    const before = canonicalJSONString(encodeReplayState(state));
    const inputs = structuredClone(row.exit.replay_inputs) as { resolved: {
      active_play?: unknown; next_active_play?: { sequence: number } } };
    expect(inputs.resolved.active_play).toBeDefined();
    expect(inputs.resolved.next_active_play).toBeDefined();
    if (fault === "current") delete inputs.resolved.active_play;
    else if (fault === "next") delete inputs.resolved.next_active_play;
    else inputs.resolved.next_active_play!.sequence++;
    await expect(applyLoggedExit(state, canonicalJSONString(row.exit.canonical_payload), bundle, inputs))
      .rejects.toThrow(fault === "sequence" ? /^missing next active schedule$/ : /^terminal active-play evidence mismatch$/);
    expect(canonicalJSONString(encodeReplayState(state))).toBe(before);
  });
});

interface OldExit {
  constants_hash: string; artifacts: ReplayArtifacts; next_constants_hash: string; next_artifacts: ReplayArtifacts; case: TerminalCase;
}
const oldCorpus = oldArtifact as unknown as { active_play_exit: OldExit; active_foundation_exit: OldExit };
const oldActive = oldCorpus.active_play_exit;
async function oldBundle(fixture: OldExit) {
  return withNextReplayCatalogBundle(await loadReplayCatalogBundle(fixture.constants_hash, fixture.artifacts),
    await loadReplayCatalogBundle(fixture.next_constants_hash, fixture.next_artifacts));
}
function restoreOld(state: unknown, version: 16 | 18, bundle: ReplayCatalogBundle) {
  if (!bundle.meters || !bundle.achievements) throw new Error("old terminal foundation catalogs missing");
  return restoreReplayState(state, version, bundle.economy, { meters: bundle.meters, achievements: bundle.achievements,
    doctrines: bundle.doctrines, opportunities: bundle.opportunities });
}

describe("CV4 terminal legacy/refusal companions (real pinned histories)", () => {
  it("keeps actual old v18 terminal outputs exact", async () => {
    const bundle = await oldBundle(oldActive);
    const row = oldActive.case;
    const result = await applyLoggedExit(restoreOld(row.pre_state, 18, bundle), canonicalJSONString(row.canonical_payload), bundle, row.replay_inputs);
    expect(result.outcome).toBe("applied");
    expect(canonicalJSONString(result.receipt)).toBe(row.receipt_json);
    expect(canonicalJSONString(result.founder)).toBe(row.founder_output_json);
    expect(canonicalJSONString(encodeReplayState(result.finalCompany))).toBe(row.final_company_json);
    expect(canonicalJSONString(encodeReplayState(result.newCompany!))).toBe(row.new_company_json);
    expect(canonicalJSONString(result.founderEvents)).toBe(row.founder_events_json);
    expect(canonicalJSONString(result.companyEndedEvents)).toBe(row.company_ended_events_json);
    expect(canonicalJSONString(result.companyStartedEvents)).toBe(row.company_started_events_json);
  });

  it("refuses missing current evidence on actual old v18 unchanged", async () => {
    const bundle = await oldBundle(oldActive);
    const row = oldActive.case;
    const state = restoreOld(row.pre_state, 18, bundle);
    const before = canonicalJSONString(encodeReplayState(state));
    const inputs = structuredClone(row.replay_inputs) as { resolved: Record<string, unknown> };
    expect(inputs.resolved.active_play).toBeDefined();
    delete inputs.resolved.active_play;
    await expect(applyLoggedExit(state, canonicalJSONString(row.canonical_payload), bundle, inputs)).rejects.toThrow(/^terminal active-play evidence mismatch$/);
    expect(canonicalJSONString(encodeReplayState(state))).toBe(before);
  });

  it.each([18, 19] as const)("refuses pre-v5 inputs on actual v%s unchanged", async version => {
    const bundle = version === 18 ? await oldBundle(oldActive) : await catalogs;
    const row = version === 18 ? oldActive.case : corpus.rows[0]!.exit;
    const state = version === 18 ? restoreOld(row.pre_state, 18, bundle) : restore(row.pre_state, bundle);
    const before = canonicalJSONString(encodeReplayState(state));
    const inputs = structuredClone(row.replay_inputs) as { v: number };
    inputs.v = 4;
    await expect(applyLoggedExit(state, canonicalJSONString(row.canonical_payload), bundle, inputs)).rejects.toThrow(/^terminal active-play evidence mismatch$/);
    expect(canonicalJSONString(encodeReplayState(state))).toBe(before);
  });

  it("refuses unexpected current evidence on actual v16 unchanged", async () => {
    const fixture = oldCorpus.active_foundation_exit;
    const bundle = await oldBundle(fixture);
    const row = fixture.case;
    const state = restoreOld(row.pre_state, 16, bundle);
    const before = canonicalJSONString(encodeReplayState(state));
    const inputs = structuredClone(row.replay_inputs) as { resolved: Record<string, unknown> };
    const active = oldActive.case.replay_inputs as { resolved: { active_play: unknown } };
    expect(inputs.resolved.active_play).toBeUndefined();
    inputs.resolved.active_play = structuredClone(active.resolved.active_play);
    await expect(applyLoggedExit(state, canonicalJSONString(row.canonical_payload), bundle, inputs)).rejects.toThrow(/^terminal active-play evidence mismatch$/);
    expect(canonicalJSONString(encodeReplayState(state))).toBe(before);
  });
});
