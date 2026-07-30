import { describe, expect, it } from "vitest";
import { decodeTransportEnvelope } from "../src/transport";

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
});
