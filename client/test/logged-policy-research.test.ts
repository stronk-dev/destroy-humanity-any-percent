import { describe, expect, it } from "vitest";
import artifact from "../../testdata/axis-stack/logged-policy-research-v1.json";
import { parseCanonical } from "../src/numeric";
import { applyLogged, canonicalJSONString, encodeReplayState, loadReplayCatalogBundle,
  restoreReplayState, type ReplayArtifacts, type ReplayCatalogBundle } from "../src/replay";

interface Row {
  profile: { id: string; mode: string; elapsed_ms: number; initial_burst_ms: number; provider_count: number; legal_count: number };
  case: { name: string; pre_state: unknown; canonical_payload: unknown; replay_inputs: unknown; outcome: string;
    receipt_json: string; events_json: string; post_state_json: string; post_state: unknown };
}
const corpus = artifact as unknown as { version: number; acceptance_status: string; negative_cases: number;
  bundle: { constants_hash: string; artifacts: ReplayArtifacts }; rows: Row[] };
const catalogs = loadReplayCatalogBundle(corpus.bundle.constants_hash, corpus.bundle.artifacts);
function restore(state: unknown, bundle: ReplayCatalogBundle) {
  if (!bundle.meters || !bundle.achievements) throw new Error("missing logged policy foundation catalogs");
  return restoreReplayState(state, 19, bundle.economy, { meters: bundle.meters, achievements: bundle.achievements,
    doctrines: bundle.doctrines, opportunities: bundle.opportunities });
}

describe("R-012 actual logged policy replay (synthetic inventory, NOT AC6 acceptance)", () => {
  it("pins all24 declarations,9 negatives and non-acceptance", () => {
    const expected: string[] = [];
    for (const mode of ["online", "offline"]) for (const end of [3114, 59999, 120001, 90000000]) {
      for (const profile of ["quiet", "boosted", "combined"]) expected.push(`${profile}/${mode}/${end}`);
    }
    expect(corpus.version).toBe(1);
    expect(corpus.acceptance_status).toBe("NOT_PROVEN: bounded logged replay only; production AC6 remains red");
    expect(corpus.rows.map(row => row.profile.id)).toEqual(expected);
    expect(corpus.rows.map(row => row.case.name)).toEqual(expected);
    expect(corpus.negative_cases).toBe(9);
  });

  it.each(corpus.rows)("compares complete actual outputs for $profile.id", async row => {
    const bundle = await catalogs;
    const state = restore(row.case.pre_state, bundle);
    expect(canonicalJSONString(encodeReplayState(state))).toBe(canonicalJSONString(row.case.pre_state));
    const result = await applyLogged(state, canonicalJSONString(row.case.canonical_payload), bundle, row.case.replay_inputs);
    expect(result.outcome).toBe(row.case.outcome);
    expect(result.outcome).toBe("applied");
    expect(canonicalJSONString(result.receipt)).toBe(row.case.receipt_json);
    expect(canonicalJSONString(result.events)).toBe(row.case.events_json);
    expect(canonicalJSONString(encodeReplayState(result.state))).toBe(row.case.post_state_json);
    const restored = restore(encodeReplayState(result.state), bundle);
    expect(canonicalJSONString(encodeReplayState(restored))).toBe(row.case.post_state_json);
    expect(result.state.generators["generator.beige_tower"]).toBe(2);
    expect(result.state.generatorPurchasedTotal).toBe(2 + row.profile.provider_count + row.profile.legal_count);
    expect(result.state.computeBurstRemainingMs).toBe(0);
    expect(parseCanonical(result.state.balances["company.permits"]!).gt(0)).toBe(row.profile.legal_count > 0);
  });

  const negativeRows = corpus.rows.filter(row => row.profile.mode === "online" && row.profile.elapsed_ms === 90000000);
  it.each(negativeRows.flatMap(row => ["missing", "from", "to"].map(fault => ({ row, fault }))))(
    "refuses copied catchup $fault for $row.profile.id", async ({ row, fault }) => {
      const bundle = await catalogs;
      const state = restore(row.case.pre_state, bundle);
      const inputs = structuredClone(row.case.replay_inputs) as { offline_catchup?: { opened_at_ms: number; offline_span: { from_ms: number; to_ms: number } } };
      expect(inputs.offline_catchup).toBeDefined();
      if (fault === "missing") delete inputs.offline_catchup;
      else if (fault === "from") inputs.offline_catchup!.offline_span.from_ms++;
      else inputs.offline_catchup!.offline_span.to_ms--;
      await expect(applyLogged(state, canonicalJSONString(row.case.canonical_payload), bundle, inputs)).rejects.toThrow(/offline catchup/);
      expect(canonicalJSONString(encodeReplayState(state))).toBe(canonicalJSONString(row.case.pre_state));
    });
});
