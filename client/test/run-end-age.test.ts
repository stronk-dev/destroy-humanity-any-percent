import { describe, expect, it } from "vitest";
import vectors from "../../testdata/replay/run-ended-age-v1.json";
import historical from "../../testdata/replay/apply-logged-v1.json";
import { applyLoggedExit, canonicalJSONString, loadReplayCatalogBundle, restoreReplayState } from "../src/replay";
import { decodeTransportEnvelope } from "../src/transport";
import { decodeGameUIEvent } from "../src/game-ui/events";

function decoded(payload: Record<string, unknown>) {
  const envelope = decodeTransportEnvelope({ v: 2, ch: `player:${String(payload.founder_id)}`, kind: "event", rev: 2,
    constants_hash: vectors.constants_hash, ts: new Date(payload.ended_at_ms as number).toISOString(),
    payload: { event_id: "run-end-age", kind: "run_ended", scope: "company", rev: 2, cursor_effect: "advance", payload } });
  if (!envelope) throw new Error("fixture transport refused");
  const value = decodeGameUIEvent(envelope);
  if (value?.kind !== "run_ended") throw new Error("fixture run-end refused");
  return value;
}

describe("immutable Founder age on run end", () => {
  it("retains the complete versioned population and original catalog binding", () => {
    expect(vectors.version).toBe(1);
    expect(vectors.artifact_source).toBe("apply-logged-v1.json:first_content_exit");
    expect(vectors.constants_hash).toBe(historical.first_content_exit.constants_hash);
    expect(vectors.cases).toHaveLength(5);
    expect(vectors.cases.map((row) => row.expected_schema_version)).toEqual([4, 2, 4, 4, 3]);
  });
  for (const row of vectors.cases) {
    it(`${row.name}: matches complete Go event and preserves private age`, async () => {
      const catalogs = await loadReplayCatalogBundle(vectors.constants_hash, historical.first_content_exit.artifacts);
      if (!catalogs.meters || !catalogs.achievements) throw new Error("missing fixture foundations");
      const foundations = { meters: catalogs.meters, achievements: catalogs.achievements, doctrines: catalogs.doctrines, opportunities: catalogs.opportunities };
      const restore = () => restoreReplayState(row.pre_state, 18, catalogs.economy, foundations);
      const result = await applyLoggedExit(restore(), canonicalJSONString(row.canonical_payload), catalogs, row.replay_inputs);
      expect(result.outcome).toBe("applied");
      expect(result.founder.age_ms).toBe(row.expected_founder_age_ms);
      const events = result.companyEndedEvents.filter((event) => event.kind === "run_ended");
      expect(events).toHaveLength(1);
      expect(canonicalJSONString(events[0])).toBe(row.ended_event_json);
      const payload = structuredClone(events[0]!.payload) as Record<string, unknown>;
      const view = decoded(payload);
      if (row.expected_schema_version === 4) {
        expect(view.payload.founder_age_ms).toBe(row.expected_founder_age_ms);
        payload.founder_age_ms = 0; // mutation of the source object cannot rewrite an ended run
        expect(view.payload.founder_age_ms).toBe(row.expected_founder_age_ms);
      } else {
        expect(Object.hasOwn(view.payload, "founder_age_ms")).toBe(false);
        expect(view.payload.founder_age_ms).toBeUndefined();
      }
      const legacy = structuredClone(row.replay_inputs); legacy.v = 12;
      const prior = await applyLoggedExit(restore(), canonicalJSONString(row.canonical_payload), catalogs, legacy);
      expect(result.founder).toEqual(prior.founder);
      expect(result.finalCompany).toEqual(prior.finalCompany);
      expect(result.newCompany).toEqual(prior.newCompany);
      expect(result.receipt).toEqual(prior.receipt);
    });
  }
  it("refuses malformed or extra age fields instead of inventing historical values", () => {
    const payload = JSON.parse(vectors.cases[0]!.ended_event_json).payload as Record<string, unknown>;
    for (const value of [null, true, "3601000", -1, 1.5, Number.MAX_SAFE_INTEGER + 1, 0, undefined]) {
      expect(() => decoded({ ...payload, founder_age_ms: value })).toThrow();
    }
    expect(decoded({ ...payload, attended_ms: 0, founder_age_ms: 0 }).payload.founder_age_ms).toBe(0);
    expect(decoded({ ...payload, attended_ms: 0, founder_age_ms: Number.MAX_SAFE_INTEGER }).payload.founder_age_ms).toBe(Number.MAX_SAFE_INTEGER);
    const { founder_age_ms: _age, ...legacy } = payload;
    expect(Object.hasOwn(decoded(legacy).payload, "founder_age_ms")).toBe(false);
    expect(() => decoded({ ...legacy, Founder_Age_MS: 3_601_000 })).toThrow(/exact/);
  });
});
