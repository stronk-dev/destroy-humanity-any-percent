import { describe, expect, it } from "vitest";
import fixture from "../../testdata/replay/permits-gate-v1.json";
import { applyLogged, canonicalJSONString, encodeReplayState, loadReplayCatalogBundle, restoreReplayState, type ReplayArtifacts } from "../src/replay";

describe("ratified two-resource Permits gate replay", () => {
  it.each(fixture.cases)("matches the Go live handler and replay: $name", async (row) => {
    const bundle = await loadReplayCatalogBundle(fixture.constants_hash, fixture.artifacts as ReplayArtifacts);
    const state = restoreReplayState(row.pre_state, fixture.state_version, bundle.economy);
    const result = await applyLogged(state, canonicalJSONString(row.canonical_payload), bundle, row.replay_inputs);

    expect(result.outcome).toBe(row.outcome);
    expect(canonicalJSONString(result.receipt)).toBe(row.receipt_json);
    expect(canonicalJSONString(result.events)).toBe(row.events_json);
    expect(canonicalJSONString(encodeReplayState(state))).toBe(row.post_state_json);

    if (row.outcome === "rejected") {
      expect(canonicalJSONString(encodeReplayState(state))).toBe(canonicalJSONString(row.pre_state));
      expect(result.events).toEqual([]);
      expect(result.receipt).toMatchObject({ rejection: {
        category: "requirement_not_met",
        detail: row.name === "permits-short" ? "company.permits" : "company.cash",
      } });
    } else {
      expect(state.balances).toEqual(row.name === "exact-requirements"
        ? { "company.cash": "0", "company.permits": "0" }
        : { "company.cash": "5e11", "company.permits": "1.2e1" });
      expect(state.tier).toBe(4);
      expect(state.gatesCrossed["gate.t3_to_t4"]).toBe(true);
      expect(result.events.map((event) => event.kind)).toEqual(["gate_crossed"]);
    }
  });
});
