import { describe, expect, it } from "vitest";
import source from "../../testdata/axis-stack/external-replay-research-v1.json";
import { applyLogged, canonicalJSONString, encodeReplayState, loadReplayCatalogBundle,
  restoreReplayState, type ReplayArtifacts, type ReplayCatalogBundle } from "../src/replay";

interface Action {
  name: string; pre_state_json: string; canonical_payload: unknown; replay_inputs: unknown;
  outcome: string; receipt_json: string; events_json: string; post_state_json: string;
}
interface Control { name: string; omitted: Action; duplicate_pre_state_json: string; rejected: Action }
const corpus = source as unknown as {
  version: number; acceptance_status: string;
  bundle: { constants_hash: string; artifacts: ReplayArtifacts }; actions: Action[]; controls: Control[];
};
const catalogs = loadReplayCatalogBundle(corpus.bundle.constants_hash, corpus.bundle.artifacts);
function restore(raw: string, bundle: ReplayCatalogBundle) {
  if (!bundle.meters || !bundle.achievements || !bundle.opportunities) throw new Error("external replay foundation missing");
  return restoreReplayState(JSON.parse(raw), 19, bundle.economy, { meters: bundle.meters,
    achievements: bundle.achievements, doctrines: bundle.doctrines, opportunities: bundle.opportunities });
}
async function compare(action: Action) {
  const bundle = await catalogs;
  const state = restore(action.pre_state_json, bundle);
  expect(canonicalJSONString(encodeReplayState(state))).toBe(action.pre_state_json);
  const result = await applyLogged(state, canonicalJSONString(action.canonical_payload), bundle, action.replay_inputs);
  expect(result.outcome).toBe(action.outcome);
  expect(canonicalJSONString(result.receipt)).toBe(action.receipt_json);
  expect(canonicalJSONString(result.events)).toBe(action.events_json);
  expect(canonicalJSONString(encodeReplayState(result.state))).toBe(action.post_state_json);
  expect(canonicalJSONString(encodeReplayState(restore(action.post_state_json, bundle)))).toBe(action.post_state_json);
  return result.state;
}

describe("external/Guild replay compatibility, not conservation acceptance", () => {
  it("requires the complete twelve primary cells and six control sets", () => {
    const names = ["commons", "fiscal", "commons_fiscal_guild", "single", "ordered", "nonzero_base"]
      .flatMap(profile => ["online", "offline"].map(mode => `${profile}/${mode}`));
    expect(corpus.version).toBe(1);
    expect(corpus.acceptance_status).toBe("NOT_PROVEN: bounded external/Guild replay compatibility; production AC6 remains red");
    expect(corpus.actions.map(action => action.name)).toEqual(names);
    expect(corpus.controls.map(control => control.name)).toEqual(names.slice(6));
    expect(corpus.actions.every(action => action.outcome === "applied")).toBe(true);
  });
  it.each(corpus.actions)("reproduces complete Go state/receipt/events for $name", async action => { await compare(action); });
  for (const field of ["receipt_json", "events_json", "post_state_json"] as const) {
    it(`rejects a forged expected ${field}`, async () => {
      const action = corpus.actions[0]!;
      const wrong = field === "post_state_json" ? action.pre_state_json : field === "events_json" ? "[]" : "{}";
      expect(wrong).not.toBe(action[field]);
      await expect(compare({ ...action, [field]: wrong })).rejects.toThrow();
    });
  }
  it.each(corpus.controls)("omitted clearing changes state despite equal cash for $name", async control => {
    const state = await compare(control.omitted);
    const primary = corpus.actions.find(action => action.name === control.name)!;
    const ordinary = restore(primary.post_state_json, await catalogs);
    expect(state.balances["company.cash"]).toBe(ordinary.balances["company.cash"]);
    expect(canonicalJSONString(encodeReplayState(state))).not.toBe(primary.post_state_json);
    expect(state.guildBoundarySeq).not.toBe(ordinary.guildBoundarySeq);
  });
  it.each(corpus.controls)("duplicate clearing refuses unchanged for $name", async control => {
    const bundle = await catalogs;
    const primary = corpus.actions.find(action => action.name === control.name)!;
    const state = restore(control.duplicate_pre_state_json, bundle);
    const before = canonicalJSONString(encodeReplayState(state));
    await expect(applyLogged(state, canonicalJSONString(primary.canonical_payload), bundle, primary.replay_inputs)).rejects.toThrow();
    expect(canonicalJSONString(encodeReplayState(state))).toBe(before);
  });
  it.each(corpus.controls)("late unaffordable purchase rolls back for $name", async control => {
    expect(control.rejected.outcome).toBe("rejected");
    expect(control.rejected.post_state_json).toBe(control.rejected.pre_state_json);
    expect(JSON.parse(control.rejected.events_json)).toEqual([]);
    await compare(control.rejected);
  });
});
