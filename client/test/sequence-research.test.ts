import { describe, expect, it } from "vitest";
import artifact from "../../testdata/axis-stack/sequence-research-v1.json";
import { applyLogged, canonicalJSONString, encodeReplayState, loadReplayCatalogBundle,
  restoreReplayState, type ReplayArtifacts, type ReplayCatalogBundle } from "../src/replay";

interface Action { name: string; pre_state_json: string; canonical_payload: unknown; replay_inputs: unknown;
  outcome: string; receipt_json: string; events_json: string; post_state_json: string }
interface Row { id: string; effect: string; founder_id: string; gap_ms: number; modes: string[];
  first_spawn_attended_ms: number; prelude_count: number; actions: Action[] }
const corpus = artifact as unknown as { version: number; acceptance_status: string; negative_cases: number;
  seed_candidates_enumerated: number; selected_first_founders: Record<string, string>;
  provision_tick_ms: number; catchup_ceiling_ms: number; command_attempts: number;
  bundle: { constants_hash: string; artifacts: ReplayArtifacts }; rows: Row[] };
const catalogs = loadReplayCatalogBundle(corpus.bundle.constants_hash, corpus.bundle.artifacts);
const effects = ["active.building", "active.click", "active.lucky", "active.production"];
const modePaths = [["online", "offline", "online"], ["offline", "online", "offline"]];
const kinds = ["buy_generator", "claim_opportunity", "buy_upgrade", "spend_compute_credit", "buy_generator",
  "perform_manual_batch", "perform_manual_batch", "buy_upgrade"];
function restore(raw: unknown, bundle: ReplayCatalogBundle) {
  if (!bundle.meters || !bundle.achievements || !bundle.opportunities) throw new Error("sequence foundation missing");
  return restoreReplayState(raw, 19, bundle.economy, { meters: bundle.meters, achievements: bundle.achievements,
    doctrines: bundle.doctrines, opportunities: bundle.opportunities });
}
describe("CV3/CV4 actual action/buff/mode sequences, not AC6 acceptance", () => {
  it("pins sixteen complete sequences and exact modes/kinds/refusal population", () => {
    expect(corpus.version).toBe(1);
    expect(corpus.acceptance_status).toBe("NOT_PROVEN: bounded action/buff/mode sequences; production AC6 remains red");
    expect(corpus.seed_candidates_enumerated).toBe(4096);
    expect(corpus.provision_tick_ms).toBe(60000);
    expect(corpus.catchup_ceiling_ms).toBe(5000);
    expect(Object.keys(corpus.selected_first_founders).sort()).toEqual(effects);
    expect(corpus.negative_cases).toBe(48);
    const ids = effects.flatMap(effect => [3114, 90000000].flatMap(gap => modePaths.map(modes => `${effect}/${gap}/${modes.join("-")}`)));
    expect(corpus.rows.map(row => row.id)).toEqual(ids);
    expect(corpus.command_attempts).toBe(corpus.rows.reduce((total, row) => total + row.actions.length, 0));
    for (const row of corpus.rows) {
      expect(row.founder_id).toBe(corpus.selected_first_founders[row.effect]);
      expect(row.prelude_count).toBe(Math.floor((row.first_spawn_attended_ms - 1) / corpus.catchup_ceiling_ms));
      const allKinds = [...Array<string>(row.prelude_count).fill("perform_manual_batch"), ...kinds];
      expect(row.actions.map(action => (action.canonical_payload as { kind: string }).kind)).toEqual(allKinds);
      expect(row.actions.map(action => action.name)).toEqual(allKinds.map((kind, index) => `${row.id}/step${index + 1}/${kind}`));
      expect(row.actions.map(action => (action.replay_inputs as { evaluation_mode: string }).evaluation_mode))
        .toEqual([...Array<string>(row.prelude_count).fill("online"), "online", "online", "online", "online", ...row.modes, "online"]);
      expect(row.actions.map(action => action.outcome)).toEqual([...Array<string>(row.prelude_count + 7).fill("applied"), "rejected"]);
      const first = Date.parse("2026-10-07T12:00:00.000Z") + row.first_spawn_attended_ms;
      const returned = first + 500 + row.gap_ms;
      expect(row.actions.map(action => (action.replay_inputs as { evaluated_at_ms: number }).evaluated_at_ms))
        .toEqual([...Array.from({ length: row.prelude_count }, (_, i) => Date.parse("2026-10-07T12:00:00.000Z") + (i + 1) * corpus.catchup_ceiling_ms),
          first, first, first + 1, first + 1, first + 500, returned, returned + corpus.catchup_ceiling_ms, returned + corpus.catchup_ceiling_ms]);
    }
  });
  it.each(corpus.rows)("compares every full continuous transition for $id", async row => {
    const bundle = await catalogs;
    let state = restore(JSON.parse(row.actions[0]!.pre_state_json), bundle);
    let originalBuff: string | undefined;
    for (const [actionIndex, action] of row.actions.entries()) {
      const index = actionIndex - row.prelude_count;
      const before = canonicalJSONString(encodeReplayState(state));
      expect(before).toBe(action.pre_state_json);
      const result = await applyLogged(state, canonicalJSONString(action.canonical_payload), bundle, action.replay_inputs);
      expect(result.outcome).toBe(action.outcome);
      expect(canonicalJSONString(result.receipt)).toBe(action.receipt_json);
      expect(canonicalJSONString(result.events)).toBe(action.events_json);
      expect(canonicalJSONString(encodeReplayState(result.state))).toBe(action.post_state_json);
      state = restore(encodeReplayState(result.state), bundle);
      expect(canonicalJSONString(encodeReplayState(state))).toBe(action.post_state_json);
      if (index === 0) { expect(state.attainmentScoreRun).toBe(12); expect(state.generators["generator.beige_tower"]).toBe(100); expect(state.pendingOpportunity?.effectRowId).toBe(row.effect); }
      if (index === 1) {
        expect(state.pendingOpportunity).toBeNull();
        expect(state.activeBuffs).toHaveLength(row.effect === "active.lucky" ? 0 : 1);
        originalBuff = state.activeBuffs[0]?.buffInstanceId;
      }
      if (index === 2) expect(state.upgradesOwned.has("upgrade.pr_intern_2")).toBe(true);
      if (index === 3) { expect(state.computeCreditMs).toBe(2000); expect(state.computeBurstRemainingMs).toBe(3000); }
      if (index === 6 && originalBuff) expect(state.activeBuffs.some(buff => buff.buffInstanceId === originalBuff)).toBe(false);
      if (index === 7) {
        expect(action.post_state_json).toBe(before); expect(state.generators["generator.beige_tower"]).toBe(101); expect(state.achievementScoreRun).toBe(0);
        if (row.gap_ms === 3114) expect(state.generatorsProvisioned["generator.beige_tower"]).toBe(0);
        else expect(state.generatorsProvisioned["generator.beige_tower"]).toBeGreaterThan(0);
      }
    }
  });
  for (const fault of ["missing", "effect", "next"] as const) {
    it.each(corpus.rows)(`refuses ${fault} claim evidence unchanged for $id`, async row => {
      const bundle = await catalogs, action = row.actions[row.prelude_count + 1]!;
      const state = restore(JSON.parse(action.pre_state_json), bundle);
      const before = canonicalJSONString(encodeReplayState(state));
      const inputs = structuredClone(action.replay_inputs) as { resolved: { active_play: { claim: null | { effect_row_id: string; next_opportunity_attended_ms: number } } } };
      expect(inputs.resolved.active_play.claim).not.toBeNull();
      if (fault === "missing") inputs.resolved.active_play.claim = null;
      else if (fault === "effect") inputs.resolved.active_play.claim!.effect_row_id = "active.invalid";
      else inputs.resolved.active_play.claim!.next_opportunity_attended_ms++;
      await expect(applyLogged(state, canonicalJSONString(action.canonical_payload), bundle, inputs)).rejects.toThrow();
      expect(canonicalJSONString(encodeReplayState(state))).toBe(before);
    });
  }
});
