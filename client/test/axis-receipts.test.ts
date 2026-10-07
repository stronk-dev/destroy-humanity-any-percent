import { describe, expect, it } from "vitest";
import ordinary from "../../testdata/axis-stack/logged-policy-research-v1.json";
import activation from "../../testdata/axis-stack/activation-research-v1.json";
import { applyLogged, applyLoggedExit, canonicalJSONString, loadReplayCatalogBundle,
  restoreReplayState, withNextReplayCatalogBundle, type ReplayArtifacts, type ReplayCatalogBundle } from "../src/replay";

interface Bundle { constants_hash: string; artifacts: ReplayArtifacts }
interface Action { pre_state: unknown; canonical_payload: unknown; replay_inputs: unknown }
const policies = ordinary as unknown as { bundle: Bundle; rows: { profile: { id: string }; case: Action }[] };
const runs = activation as unknown as { current: Bundle; next: Bundle; rows: { id: string; old: Action; exit: Action }[] };
function restore(raw: unknown, version: 18 | 19, bundle: ReplayCatalogBundle) {
  if (!bundle.meters || !bundle.achievements || !bundle.opportunities) throw new Error("receipt foundation missing");
  return restoreReplayState(raw, version, bundle.economy, { meters: bundle.meters, achievements: bundle.achievements,
    doctrines: bundle.doctrines, opportunities: bundle.opportunities });
}
function snapshot(receipt: unknown): Record<string, unknown> {
  return (receipt as { snapshot: Record<string, unknown> }).snapshot;
}
const empty = { input_kind: "achievement_attainment_run", input_value: 0, input_cap: 44,
  cap_reason_key: "cap.axis_stack_input", saturated: false, contributions: [], product: "1e0" };
const owned8 = { ...empty, input_value: 8, contributions: [{ source_id: "upgrade.pr_intern_1.axis",
  upgrade_id: "upgrade.pr_intern_1", factor: "1.2e0" }], product: "1.2e0" };

describe("CV5 actual applied receipt projection, independent literal oracles", () => {
  it("requires the exact ordinary/activation populations", () => {
    expect(policies.rows).toHaveLength(24); expect(runs.rows).toHaveLength(6);
  });
  it.each(policies.rows)("ordinary $profile.id exposes owned factor and product", async row => {
    const bundle = await loadReplayCatalogBundle(policies.bundle.constants_hash, policies.bundle.artifacts);
    const result = await applyLogged(restore(row.case.pre_state, 19, bundle), canonicalJSONString(row.case.canonical_payload), bundle, row.case.replay_inputs);
    expect(result.outcome).toBe("applied");
    expect(snapshot(result.receipt).axis_stack).toEqual(owned8);
  });
  it.each(runs.rows)("old receipt remains legacy and actual Exit $id resets projection", async row => {
    const current = await loadReplayCatalogBundle(runs.current.constants_hash, runs.current.artifacts);
    const next = await loadReplayCatalogBundle(runs.next.constants_hash, runs.next.artifacts);
    const old = await applyLogged(restore(row.old.pre_state, 18, current), canonicalJSONString(row.old.canonical_payload), current, row.old.replay_inputs);
    expect(old.outcome).toBe("applied"); expect(snapshot(old.receipt)).not.toHaveProperty("axis_stack");
    const result = await applyLoggedExit(restore(row.exit.pre_state, 18, current), canonicalJSONString(row.exit.canonical_payload), withNextReplayCatalogBundle(current, next), row.exit.replay_inputs);
    expect(result.outcome).toBe("applied"); expect(snapshot(result.receipt).axis_stack).toEqual(empty);
  });
});
