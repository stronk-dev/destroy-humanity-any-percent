import { describe, expect, it } from "vitest";

import { decodeGameUIEvent } from "../src/game-ui/events";
import { createBrowserGameUIRuntime, type RuntimeStorage } from "../src/game-ui/runtime";
import { decodeTransportEnvelope } from "../src/transport";

const founderID = "01985555-1111-7111-8111-111111111111";
const companyID = "01985555-2222-7222-8222-222222222222";
const constantsHash = `sha256:${"a".repeat(64)}`;
const now = 1_800_000_000_000;
const summary = { bonus_factor: "1.003e0", applied_starter_node_ids: ["reputation.starter.z", "reputation.starter.a"] };
function payload(tree: unknown = summary): Record<string, unknown> {
  return { founder_id: founderID, run_id: { company_stream_id: companyID, run_seq: 2 },
    started_at_ms: now, assisted: { advisor: false, commons: false }, reputation_tree: tree };
}
function wire(value: Record<string, unknown>, revision = 2): Record<string, unknown> {
  return { v: 2, ch: `player:${founderID}`, kind: "event", rev: revision,
    constants_hash: constantsHash, ts: new Date(now).toISOString(),
    payload: { event_id: `started-${revision}`, kind: "run_started", scope: "company", rev: revision,
      cursor_effect: "advance", payload: value } };
}
function decode(value: Record<string, unknown>) {
  return decodeGameUIEvent(decodeTransportEnvelope(wire(value))!);
}

describe("R7 next-run event reader (not Run End UI acceptance)", () => {
  it("retains the v1 absent-tree arm without inventing a tree", () => {
    const { reputation_tree: _tree, ...legacy } = payload();
    expect(decode(legacy)).toEqual({ cursor: 2, kind: "run_started", occurred_at_ms: now, payload: legacy });
    expect(decode(legacy)?.payload).not.toHaveProperty("reputation_tree");
  });
  it("retains the explicit v2 null arm", () => {
    expect(decode(payload(null))).toEqual({ cursor: 2, kind: "run_started", occurred_at_ms: now, payload: payload(null) });
  });
  it("retains the frozen factor and producer's non-lexical starter order", () => {
    expect(decode(payload())).toEqual({ cursor: 2, kind: "run_started", occurred_at_ms: now, payload: payload() });
  });
  it("accepts a unit factor with no applied starters", () => {
    const value = payload({ bonus_factor: "1e0", applied_starter_node_ids: [] });
    expect(decode(value)?.payload).toEqual(value);
  });
  it("preserves both boolean assistance flags", () => {
    const value = { ...payload(), assisted: { advisor: true, commons: true } };
    expect(decode(value)?.payload).toEqual(value);
  });
  it("does not confuse resource hardcaps with the Decimal state limit", () => {
    const value = payload({ ...summary, bonus_factor: "1e1001" });
    expect(decode(value)?.payload).toEqual(value);
  });

  const invalid: readonly [string, Record<string, unknown>][] = [
    ["extra base field", { ...payload(), extra: true }],
    ["missing founder", (() => { const { founder_id: _id, ...rest } = payload(); return rest; })()],
    ["invalid founder", { ...payload(), founder_id: "not-a-uuid" }],
    ["invalid company", { ...payload(), run_id: { company_stream_id: "not-a-uuid", run_seq: 2 } }],
    ["zero run sequence", { ...payload(), run_id: { company_stream_id: companyID, run_seq: 0 } }],
    ["unsafe run sequence", { ...payload(), run_id: { company_stream_id: companyID, run_seq: Number.MAX_SAFE_INTEGER + 1 } }],
    ["extra run field", { ...payload(), run_id: { company_stream_id: companyID, run_seq: 2, extra: true } }],
    ["zero start", { ...payload(), started_at_ms: 0 }],
    ["unsafe start", { ...payload(), started_at_ms: Number.MAX_SAFE_INTEGER + 1 }],
    ["fractional start", { ...payload(), started_at_ms: 0.5 }],
    ["string start", { ...payload(), started_at_ms: String(now) }],
    ["missing assisted field", { ...payload(), assisted: { advisor: false } }],
    ["extra assisted field", { ...payload(), assisted: { advisor: false, commons: false, extra: true } }],
    ["nonboolean assisted", { ...payload(), assisted: { advisor: 0, commons: false } }],
    ["array assistance", { ...payload(), assisted: [] }],
    ["scalar tree", payload(1)],
    ["array tree", payload([])],
    ["extra tree field", payload({ ...summary, extra: true })],
    ["missing starter IDs", payload({ bonus_factor: "1e0" })],
    ["missing factor", payload({ applied_starter_node_ids: [] })],
    ["numeric factor", payload({ ...summary, bonus_factor: 1 })],
    ["noncanonical factor", payload({ ...summary, bonus_factor: "1.0" })],
    ["below-one factor", payload({ ...summary, bonus_factor: "9e-1" })],
    ["out-of-state factor", payload({ ...summary, bonus_factor: "1e9000000000000000" })],
    ["NaN factor", payload({ ...summary, bonus_factor: "NaN" })],
    ["null starters", payload({ ...summary, applied_starter_node_ids: null })],
    ["scalar starters", payload({ ...summary, applied_starter_node_ids: "reputation.starter.a" })],
    ["duplicate starters", payload({ ...summary, applied_starter_node_ids: ["reputation.starter.a", "reputation.starter.a"] })],
    ["invalid starter", payload({ ...summary, applied_starter_node_ids: ["Bad ID"] })],
    ["numeric starter", payload({ ...summary, applied_starter_node_ids: [1] })],
  ];
  it.each(invalid)("refuses %s", (_label, value) => { expect(() => decode(value)).toThrow(); });
});

class MemoryStorage implements RuntimeStorage {
  readonly values = new Map<string, string>();
  getItem(key: string): string | null { return this.values.get(key) ?? null; }
  setItem(key: string, value: string): void { this.values.set(key, value); }
  removeItem(key: string): void { this.values.delete(key); }
}
type Listener = (event: { data?: string; code?: number }) => void;
class ControlledSocket {
  readonly listeners = new Map<string, Listener[]>();
  readonly sent: string[] = [];
  closes = 0;
  addEventListener(kind: string, listener: Listener): void { this.listeners.set(kind, [...(this.listeners.get(kind) ?? []), listener]); }
  send(value: string): void { this.sent.push(value); }
  close(): void { this.closes += 1; }
  emit(kind: string, value: { data?: string; code?: number } = {}): void { for (const listener of this.listeners.get(kind) ?? []) listener(value); }
  reply(value: unknown): void { this.emit("message", { data: JSON.stringify(value) }); }
}
function snapshot(revision: number, runSeq: number) {
  return { constants_hash: constantsHash, evaluated_through_ms: now, server_now_ms: now,
    facts: [{ fact_id: "bootstrap.needed", value: false }], generators: [],
    manual_action: { action_id: "manual.click", bucket_cap_milli: 50_000, refill_milli_per_ms: 25, refilled_at_ms: now, tokens_milli: 50_000 },
    progress: [], resources: [], revision, founder_revision: 1,
    run: { category: "any_percent", exit_count: runSeq - 1, founder_id: founderID, run_seq: runSeq, run_started_at_ms: now, tier: 0 },
    schema_version: 4, features: { achievements: null, active_play: null, fiscal: null, meters: null, minigames: null, pets: null },
    transitions: { cross_gate: { eligible: false, gate_id: "gate.t0_to_t1", route_id: null }, wind_down: { eligible: false } }, upgrades: [] };
}
async function boundary() {
  const storage = new MemoryStorage();
  storage.setItem("cloud-clicker.credentials.v1", JSON.stringify({ accessToken: "access", refreshToken: "refresh", accountID: "account", recoveryCode: "recovery" }));
  const socket = new ControlledSocket();
  let response = snapshot(1, 1);
  let reads = 0;
  const runtime = createBrowserGameUIRuntime(storage, async () => { reads += 1; return Response.json(response); }, crypto,
    () => socket as unknown as WebSocket, { protocol: "http:", host: "controlled.invalid" });
  await runtime.snapshot();
  const received: unknown[] = [];
  let resolveSnapshot!: (value: unknown) => void;
  const nextSnapshot = new Promise<unknown>((resolve) => { resolveSnapshot = resolve; });
  const stop = runtime.subscribe(founderID, (message) => {
    received.push(message);
    if (message.kind === "snapshot") resolveSnapshot(message);
  });
  socket.emit("open"); socket.reply({ id: 1, connect: {} });
  socket.reply({ id: 2, subscribe: { recoverable: true, positioned: true, recovered: false, epoch: "player", offset: 0, publications: [] } });
  socket.reply({ id: 3, subscribe: { recoverable: true, positioned: true, recovered: false, epoch: "world", offset: 0, publications: [] } });
  return { runtime, socket, received, nextSnapshot, stop, get reads() { return reads; },
    setSnapshot(value: ReturnType<typeof snapshot>) { response = value; },
    publish(value: Record<string, unknown>, offset = 1, revision = 2) { socket.reply({ push: { channel: `player:${founderID}`, pub: { offset, data: wire(value, revision) } } }); } };
}
const delivered = { kind: "event", revision: 2, scope: "company",
  value: { cursor: 2, kind: "run_started", occurred_at_ms: now, payload: payload() } };

describe("R7 real runtime over controlled HTTP/socket bytes", () => {
  it.each(["event-first", "snapshot-first-same-revision", "snapshot-first-advanced-revision"])("delivers the current run's summary, %s", async (order) => {
    const b = await boundary();
    try {
      if (order !== "event-first") {
        b.setSnapshot(snapshot(order === "snapshot-first-same-revision" ? 2 : 3, 2));
        await b.runtime.snapshot();
      }
      b.publish(payload());
      expect(b.received).toEqual([{ kind: "transport_recovered" }, delivered]);
      b.publish(payload()); // Same channel offset is still at-most-once.
      expect(b.received).toEqual([{ kind: "transport_recovered" }, delivered]);
      b.publish(payload(), 2); // Same event at a new outbox offset is still a duplicate.
      expect(b.received).toEqual([{ kind: "transport_recovered" }, delivered]);
      b.setSnapshot(snapshot(3, 2)); await b.runtime.snapshot();
      b.publish(payload(), 3); // A later cursor reset must not erase delivery identity.
      expect(b.received).toEqual([{ kind: "transport_recovered" }, delivered]);
      expect(b.socket.closes).toBe(0);
    } finally { b.stop(); }
  });
  it("does not suppress the next distinct run's summary", async () => {
    const b = await boundary();
    try {
      b.publish(payload());
      const next = { ...payload(), run_id: { company_stream_id: companyID, run_seq: 3 }, started_at_ms: now + 1 };
      b.publish(next, 2, 3);
      expect(b.received).toEqual([{ kind: "transport_recovered" }, delivered,
        { kind: "event", revision: 3, scope: "company", value: { cursor: 3, kind: "run_started", occurred_at_ms: now, payload: next } }]);
      expect(b.socket.closes).toBe(0);
    } finally { b.stop(); }
  });
  it.each(["wrong-founder", "old-run", "future-run", "wrong-start-time"])("does not revive a duplicate from %s", async (kind) => {
    const b = await boundary();
    try {
      b.setSnapshot(snapshot(3, 2)); await b.runtime.snapshot();
      const value = payload();
      if (kind === "wrong-founder") value.founder_id = "01985555-3333-7333-8333-333333333333";
      else if (kind === "wrong-start-time") value.started_at_ms = now - 1;
      else value.run_id = { company_stream_id: companyID, run_seq: kind === "old-run" ? 1 : 3 };
      b.publish(value);
      expect(b.received).toEqual([{ kind: "transport_recovered" }]);
      expect(b.socket.closes).toBe(0);
    } finally { b.stop(); }
  });
  it("routes a malformed summary through the existing resync, not delivery", async () => {
    const b = await boundary();
    try {
      b.publish(payload({ ...summary, bonus_factor: "9e-1" }));
      expect(b.received).toEqual([{ kind: "transport_recovered" }, { kind: "system", value: { kind: "resync_required" } }]);
      expect(b.socket.closes).toBe(1);
      expect(b.reads).toBe(2);
      expect(await b.nextSnapshot).toMatchObject({ kind: "snapshot", value: {
        schema_version: 4, revision: 1, run: { founder_id: founderID, run_seq: 1 },
      } });
      expect(b.received.filter((value) => (value as { kind: string }).kind === "event")).toEqual([]);
    } finally { b.stop(); }
  });
});
