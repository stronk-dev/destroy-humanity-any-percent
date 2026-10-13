import { describe, expect, it } from "vitest";
import source from "../../testdata/axis-stack/temporal-replay-research-v1.json";
import { parseCanonical } from "../src/numeric";
import { applyLogged, canonicalJSONString, encodeReplayState, loadReplayCatalogBundle,
  restoreReplayState, type ReplayArtifacts, type ReplayCatalogBundle } from "../src/replay";

interface Action {
  name: string; pre_state_json: string; canonical_payload: unknown; replay_inputs: unknown;
  outcome: string; receipt_json: string; events_json: string; post_state_json: string;
}
interface Row {
  id: string; profile: string; mode: string; offset_ms: number; attended_ms: number;
  frozen_changed: boolean; omitted_changed: boolean; action: Action;
}
const corpus = source as unknown as {
  version: number; acceptance_status: string; source_sha256: Record<string, string>; invalid_cases: number;
  bundle: { constants_hash: string; artifacts: ReplayArtifacts }; rows: Row[];
};
const catalogs = loadReplayCatalogBundle(corpus.bundle.constants_hash, corpus.bundle.artifacts);
const offsets = [999, 1000, 1001, 1499, 1500, 1501, 6000];
function restore(raw: unknown, bundle: ReplayCatalogBundle) {
  if (!bundle.meters || !bundle.achievements || !bundle.opportunities) throw new Error("temporal replay foundation missing");
  return restoreReplayState(raw, 19, bundle.economy, { meters: bundle.meters, achievements: bundle.achievements,
    doctrines: bundle.doctrines, opportunities: bundle.opportunities });
}

describe("CV3 temporal replay compatibility, not conservation acceptance", () => {
  it("requires all42 declared cells and their existing refusal controls", () => {
    expect(corpus.version).toBe(1);
    expect(corpus.acceptance_status).toBe("NOT_PROVEN: bounded temporal replay compatibility; production AC6 remains red");
    const ids = ["production", "building", "combined"].flatMap(profile => ["online", "offline"].flatMap(mode => offsets.map(offset => `${profile}/${mode}/${offset}`)));
    expect(corpus.rows.map(row => row.id)).toEqual(ids);
    expect(corpus.rows.filter(row => row.frozen_changed)).toHaveLength(24);
    expect(corpus.rows.filter(row => row.omitted_changed)).toHaveLength(24);
    expect(corpus.invalid_cases).toBe(84);
    for (const row of corpus.rows) {
      expect(row.id).toBe(`${row.profile}/${row.mode}/${row.offset_ms}`);
      expect(row.action.name).toBe(`${row.id}/perform_manual_batch`);
      expect(row.action.outcome).toBe("applied");
      expect(row.attended_ms).toBe(row.offset_ms > 5000 ? 4000 : 4000 + row.offset_ms);
      const inputs = row.action.replay_inputs as { evaluation_mode: string; evaluated_at_ms: number; resolved: { active_play: { attended_now_ms: number } } };
      expect(inputs.evaluation_mode).toBe(row.mode);
      expect(inputs.evaluated_at_ms).toBe(Date.parse("2026-10-12T12:00:00.000Z") + row.offset_ms);
      expect(inputs.resolved.active_play.attended_now_ms).toBe(row.attended_ms);
    }
  });

  it.each(corpus.rows)("reproduces whole Go state, receipt and ordered events for $id", async row => {
    const bundle = await catalogs;
    const state = restore(JSON.parse(row.action.pre_state_json), bundle);
    const before = canonicalJSONString(encodeReplayState(state));
    expect(before).toBe(row.action.pre_state_json);
    const originalLifetime = state.lifetimeValue;
    const originalBuffs = structuredClone(state.activeBuffs);
    const result = await applyLogged(state, canonicalJSONString(row.action.canonical_payload), bundle, row.action.replay_inputs);
    expect(result.outcome).toBe("applied");
    expect(canonicalJSONString(result.receipt)).toBe(row.action.receipt_json);
    expect(canonicalJSONString(result.events)).toBe(row.action.events_json);
    const encoded = encodeReplayState(result.state);
    expect(canonicalJSONString(encoded)).toBe(row.action.post_state_json);
    expect(canonicalJSONString(encodeReplayState(restore(encoded, bundle)))).toBe(row.action.post_state_json);
    expect(parseCanonical(result.state.lifetimeValue).gt(parseCanonical(originalLifetime))).toBe(true);
    const expectedBuffs = originalBuffs.filter(buff => buff.expiresAttendedMs > row.attended_ms);
    expect(result.state.activeBuffs).toEqual(expectedBuffs);
    expect(result.state.offlineSpans).toEqual(row.offset_ms > 5000
      ? [{ fromMs: Date.parse("2026-10-12T12:00:00.000Z"), toMs: Date.parse("2026-10-12T12:00:00.000Z") + row.offset_ms }]
      : []);
    expect(canonicalJSONString(encodeReplayState(state))).toBe(row.action.post_state_json);
  });

  for (const [index, fault] of ["forged attended coordinate", "missing schedule"].entries()) {
    it.each(corpus.rows)(`refuses ${fault} without mutation for $id`, async row => {
      const bundle = await catalogs;
      const state = restore(JSON.parse(row.action.pre_state_json), bundle);
      const before = canonicalJSONString(encodeReplayState(state));
      // Match Go's two actual codec variants without storing84 duplicate
      // copies of the otherwise identical replay context in the artifact.
      const inputs = structuredClone(row.action.replay_inputs) as { resolved: { active_play?: { attended_now_ms: number } } };
      if (index === 0) inputs.resolved.active_play!.attended_now_ms++;
      else delete inputs.resolved.active_play;
      await expect(applyLogged(state, canonicalJSONString(row.action.canonical_payload), bundle, inputs)).rejects.toThrow();
      expect(canonicalJSONString(encodeReplayState(state))).toBe(before);
    });
  }
});
