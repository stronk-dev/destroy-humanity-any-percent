import { describe, expect, it } from "vitest";

import { decodeGameUIAnnouncement } from "../src/game-ui/events";
import type { TransportEnvelope } from "../src/transport";

function envelope(kind: string, payload: Record<string, unknown>, rev = 5): TransportEnvelope {
  return { v: 2, ch: "player:x", kind: "event", rev, constants_hash: `sha256:${"a".repeat(64)}`, ts: "2027-01-01T00:00:00.000Z",
    payload: { event_id: `e-${rev}`, kind, scope: "founder", rev, cursor_effect: "advance", payload } };
}

// Shapes are exactly the producers in server/production/founder_replay.go and
// active_play.go, validated by server/save/intent.go.
const automatic = { source: "automatic", periods: 2, credit_before: 4, credited: 6, credit_after: 10, opened_before_ms: 1_000, opened_after_ms: 3_000,
  seq_before: 1, seq_after: 3, saturated: false, hardcap_reason_key: "cap.fiscal_credit" };
const manual = { source: "manual", outcome: "guaranteed", credit_before: 4, credit_after: 7, period_opened_wall_ms_before: 1_000,
  period_opened_wall_ms_after: 2_000, seq_before: 1, seq_after: 2, draw_ppm: null, saturated: false };
const buff = { buff_instance_id: "01986666-0000-7000-8000-00000000000b", effect_row_id: "active.production", selected_target: null,
  activated_attended_ms: 1_000, expires_attended_ms: 6_000, hardcap_reason_key: null };

describe("GS0.3 remainder decoders", () => {
  const reattained = { achievement_id: "achievement.first_gate", score_grant: 2,
    run_id: { company_stream_id: "01985555-2222-7222-8222-222222222222", run_seq: 1 } };
  const reattainmentEnvelope = (payload: Record<string, unknown>): TransportEnvelope => {
    const event = envelope("achievement_reattained.v1", payload);
    return { ...event, payload: { ...event.payload, scope: "company" } };
  };
  it("decodes exact Company re-attainment without adding lifetime ownership", () => {
    expect(decodeGameUIAnnouncement(reattainmentEnvelope(reattained))).toEqual({ cursor: 5, kind: "achievement_reattained", payload: reattained });
  });
  for (const changes of [
    { extra: true }, { condition_scope: "run" }, { achievement_id: null }, { achievement_id: "not an id" },
    { score_grant: null }, { score_grant: 0 }, { score_grant: -1 }, { score_grant: 1.5 }, { score_grant: Number.MAX_SAFE_INTEGER + 1 },
    { run_id: null }, { run_id: { company_stream_id: "invalid", run_seq: 1 } },
    { run_id: { company_stream_id: reattained.run_id.company_stream_id, run_seq: 0 } },
  ]) {
    it(`refuses malformed re-attainment ${JSON.stringify(changes)}`, () => {
      expect(() => decodeGameUIAnnouncement(reattainmentEnvelope({ ...reattained, ...changes }))).toThrow();
    });
  }
  it("refuses re-attainment on the Founder stream and missing required fields", () => {
    expect(() => decodeGameUIAnnouncement(envelope("achievement_reattained.v1", reattained))).toThrow();
    for (const key of Object.keys(reattained)) {
      const payload: Record<string, unknown> = { ...reattained };
      delete payload[key];
      expect(() => decodeGameUIAnnouncement(reattainmentEnvelope(payload))).toThrow();
    }
  });

  it("decodes both Fiscal harvest sources and a buff start", () => {
    expect(decodeGameUIAnnouncement(envelope("fiscal_period_harvested.v1", automatic))).toEqual({ cursor: 5, kind: "fiscal_period_harvested", payload: { source: "automatic", credit_after: 10 } });
    expect(decodeGameUIAnnouncement(envelope("fiscal_period_harvested.v1", manual))).toEqual({ cursor: 5, kind: "fiscal_period_harvested", payload: { source: "manual", credit_after: 7 } });
    expect(decodeGameUIAnnouncement(envelope("buff_started.v1", buff))).toEqual({ cursor: 5, kind: "buff_started",
      payload: { buff_instance_id: buff.buff_instance_id, effect_row_id: "active.production", expires_attended_ms: 6_000 } });
    const { hardcap_reason_key: _omitted, ...v1Buff } = buff;
    expect(decodeGameUIAnnouncement(envelope("buff_started.v1", v1Buff))?.kind).toBe("buff_started");
  });

  it("fails closed on contradictions and unknown keys", () => {
    expect(() => decodeGameUIAnnouncement(envelope("fiscal_period_harvested.v1", { ...automatic, credit_after: 11 }))).toThrow(/automatic/);
    expect(() => decodeGameUIAnnouncement(envelope("fiscal_period_harvested.v1", { ...manual, seq_after: 3 }))).toThrow(/manual/);
    expect(() => decodeGameUIAnnouncement(envelope("fiscal_period_harvested.v1", { ...manual, outcome: "consumed_by_auto" }))).toThrow(/outcome/);
    expect(() => decodeGameUIAnnouncement(envelope("fiscal_period_harvested.v1", { ...manual, source: "gift" }))).toThrow(/source/);
    expect(() => decodeGameUIAnnouncement(envelope("fiscal_period_harvested.v1", { ...automatic, extra: 1 }))).toThrow();
    expect(() => decodeGameUIAnnouncement(envelope("buff_started.v1", { ...buff, expires_attended_ms: 1_000 }))).toThrow(/expire/);
    expect(() => decodeGameUIAnnouncement(envelope("buff_started.v1", { ...buff, buff_instance_id: "not-a-uuid" }))).toThrow(/instance/);
    expect(() => decodeGameUIAnnouncement(envelope("buff_started.v1", { ...buff, mood: 1 }))).toThrow();
  });
});
