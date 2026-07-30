import { describe, expect, it } from "vitest";
import { decodeTransportEnvelope } from "../src/transport";
import wireVectors from "../../testdata/transport/wire-vectors.json";

const base = { v: 1, ch: "player:01985555-0000-7000-8000-000000000001", rev: 3, constants_hash: `sha256:${"a".repeat(64)}`, ts: "2026-07-29T12:00:00.000Z" };

describe("transport wire", () => {
  it("passes receipt payloads through unchanged", () => {
    const payload = { intent_id: "01985555-0001-7000-8000-000000000001", outcome: "applied", new_revision: 3, nested: { exact: true } };
    const decoded = decodeTransportEnvelope({ ...base, kind: "receipt", payload });
    expect(decoded?.payload).toBe(payload);
  });

  it("ignores an unknown kind before interpreting its payload", () => {
    expect(decodeTransportEnvelope({ ...base, kind: "future", payload: { anything: true }, future_field: 1 })).toBeUndefined();
  });

  it("rejects unknown fields inside known kinds", () => {
    expect(() => decodeTransportEnvelope({ ...base, kind: "system", payload: { code: "resync_required", surprise: true } })).toThrow(SyntaxError);
  });

  it("accepts the closed recovery signal", () => {
    expect(decodeTransportEnvelope({ ...base, kind: "system", payload: { code: "resync_required" } })?.kind).toBe("system");
  });

  it("rejects private receipt kinds on public channels", () => {
    expect(() => decodeTransportEnvelope({ ...base, ch: "world", kind: "receipt", payload: { outcome: "applied" } })).toThrow(SyntaxError);
  });

  it("binds snapshot scope and payload revision to the envelope", () => {
    expect(decodeTransportEnvelope({ ...base, ch: "world", rev: 7, kind: "snapshot", payload: { scope: "world", rev: 7, state: {} } })?.kind).toBe("snapshot");
    expect(() => decodeTransportEnvelope({ ...base, ch: "world", rev: 7, kind: "snapshot", payload: { scope: "company", rev: 7, state: {} } })).toThrow(SyntaxError);
    expect(() => decodeTransportEnvelope({ ...base, ch: "world", rev: 7, kind: "snapshot", payload: { scope: "world", rev: 8, state: {} } })).toThrow(SyntaxError);
  });

  it("binds event revision and object payloads", () => {
    const event = { ...base, ch: "feed", rev: 4, kind: "event", payload: { event_id: "event-4", kind: "run.ended", scope: "company", rev: 4, payload: {} } };
    expect(decodeTransportEnvelope(event)?.kind).toBe("event");
    expect(() => decodeTransportEnvelope({ ...event, payload: { ...event.payload, rev: 5 } })).toThrow(SyntaxError);
    expect(() => decodeTransportEnvelope({ ...event, payload: { ...event.payload, payload: [] } })).toThrow(SyntaxError);
  });

  it("matches the shared Go wire corpus", () => {
    expect(wireVectors.length).toBeGreaterThanOrEqual(10);
    for (const vector of wireVectors) {
      if (vector.valid) expect(decodeTransportEnvelope(vector.envelope)).toBeDefined();
      else expect(() => decodeTransportEnvelope(vector.envelope)).toThrow(SyntaxError);
    }
  });
});
