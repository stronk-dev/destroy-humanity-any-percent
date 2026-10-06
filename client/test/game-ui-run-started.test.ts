import { describe, expect, it } from "vitest";

import { decodeGameUIEvent } from "../src/game-ui/events";
import { parseGameUISnapshot } from "../src/game-ui/contracts";
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
type Publication = Readonly<{ offset: number; data: unknown }>;
type Handshake = Readonly<{ publications?: readonly Publication[]; offset?: number; recovered?: boolean; epoch?: string }>;
async function boundary(options: Readonly<{ saved?: boolean; initial?: ReturnType<typeof snapshot>; handshake?: Handshake; deferHandshake?: boolean }> = {}) {
  const storage = new MemoryStorage();
  storage.setItem("cloud-clicker.credentials.v1", JSON.stringify({ accessToken: "access", refreshToken: "refresh", accountID: "account", recoveryCode: "recovery" }));
  if (options.saved) storage.setItem(`cloud-clicker.transport.v1.${founderID}`, JSON.stringify({
    [`player:${founderID}`]: { epoch: "player", offset: 1 }, world: { epoch: "world", offset: 0 },
  }));
  const sockets: ControlledSocket[] = [];
  const socketWaiters: ((socket: ControlledSocket) => void)[] = [];
  const recoveryWaiters: (() => void)[] = [];
  let response = options.initial ?? snapshot(1, 1);
  let reads = 0;
  const runtime = createBrowserGameUIRuntime(storage, async () => { reads += 1; return Response.json(response); }, crypto,
    () => {
      const socket = new ControlledSocket(); sockets.push(socket); socketWaiters.shift()?.(socket);
      return socket as unknown as WebSocket;
    }, { protocol: "http:", host: "controlled.invalid" });
  await runtime.snapshot();
  const received: unknown[] = [];
  let resolveSnapshot!: (value: unknown) => void;
  const nextSnapshot = new Promise<unknown>((resolve) => { resolveSnapshot = resolve; });
  const stop = runtime.subscribe(founderID, (message) => {
    received.push(message);
    if (message.kind === "snapshot") resolveSnapshot(message);
    if (message.kind === "transport_recovered") recoveryWaiters.shift()?.();
  });
  const socket = sockets[0]!;
  const connect = (connection: ControlledSocket, handshake: Handshake = {}) => {
    connection.emit("open"); connection.reply({ id: 1, connect: {} });
    const publications = handshake.publications ?? [];
    connection.reply({ id: 2, subscribe: { recoverable: true, positioned: true, recovered: handshake.recovered ?? false,
      epoch: handshake.epoch ?? "player", offset: handshake.offset ?? publications.at(-1)?.offset ?? 0, publications } });
    connection.reply({ id: 3, subscribe: { recoverable: true, positioned: true, recovered: handshake.recovered ?? false,
      epoch: "world", offset: 0, publications: [] } });
  };
  if (!options.deferHandshake) connect(socket, options.handshake ?? (options.saved ? { recovered: true, offset: 1 } : {}));
  return { runtime, socket, sockets, storage, received, nextSnapshot, stop, connect, get reads() { return reads; },
    nextSocket() { return new Promise<ControlledSocket>((resolve) => socketWaiters.push(resolve)); },
    nextRecovery() { return new Promise<void>((resolve) => recoveryWaiters.push(resolve)); },
    setSnapshot(value: ReturnType<typeof snapshot>) { response = value; },
    publish(value: Record<string, unknown>, offset = 1, revision = 2, connection = socket) {
      connection.reply({ push: { channel: `player:${founderID}`, pub: { offset, data: wire(value, revision) } } });
    } };
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

// Company revisions reflect an actual Exit pair: run2 can have current revision4,
// ends at5, and run3 starts at6. These are controlled bytes, not a SQL career.
const run2 = payload();
const run3 = { ...payload(), run_id: { company_stream_id: companyID, run_seq: 3 }, started_at_ms: now + 1000 };
const endedRun2 = {
  assisted: { advisor: false, commons: false }, attended_ms: 500, ended_at_ms: now + 1000,
  executed_routes: [], exit_type: "collapse", faction: null, founder_id: founderID,
  gates_crossed: ["gate.t0_to_t1"], generators_purchased_total: 1, ledger_fact_kinds: [], lifetime_value: "1e3",
  payout: { clout_reach_note: "clout.reach.preserved", network_slot_unlocks: [], reputation_delta: 2, route_knowledge: 25 },
  pre_timer: false, rta_ms: 1000, run_id: { company_stream_id: companyID, run_seq: 2 },
  started_at_ms: now, terminal_seq: 5, tier: 1,
};
function endWire(): Record<string, unknown> {
  const envelope = wire({}, 5);
  return { ...envelope, payload: { ...(envelope.payload as Record<string, unknown>), kind: "run_ended", payload: endedRun2 } };
}
function startMessage(value: Record<string, unknown>, revision: number) {
  return { kind: "event", revision, scope: "company", value: { cursor: revision, kind: "run_started", occurred_at_ms: now, payload: value } };
}
const endMessage = { kind: "event", revision: 5, scope: "company",
  value: { cursor: 5, kind: "run_ended", occurred_at_ms: now, payload: endedRun2 } };
function positions(storage: RuntimeStorage) {
  return JSON.parse(storage.getItem(`cloud-clicker.transport.v1.${founderID}`)!);
}

describe("R7 replay/reconnect boundary (controlled bytes, not real server)", () => {
  it.each(["absent", "null", "object"])("recovers the %s tree arm at the sampled run, before reporting recovery", async (arm) => {
    const value = payload(arm === "null" ? null : summary);
    if (arm === "absent") delete value.reputation_tree;
    const b = await boundary({ saved: true, initial: snapshot(4, 2), handshake: {
      recovered: true, offset: 2, publications: [{ offset: 2, data: wire(value, 3) }],
    } });
    try {
      expect(b.received).toEqual([startMessage(value, 3), { kind: "transport_recovered" }]);
      expect(b.socket.sent.map((command) => JSON.parse(command))).toEqual([
        { id: 1, connect: { token: "access" } },
        { id: 2, subscribe: { channel: `player:${founderID}`, recover: true, epoch: "player", offset: 1 } },
        { id: 3, subscribe: { channel: "world", recover: true, epoch: "world", offset: 0 } },
      ]);
      expect(positions(b.storage)).toEqual({ [`player:${founderID}`]: { epoch: "player", offset: 2 }, world: { epoch: "world", offset: 0 } });
      b.publish(value, 3, 3);
      expect(b.received).toEqual([startMessage(value, 3), { kind: "transport_recovered" }]);
      expect(b.reads).toBe(1); expect(b.socket.closes).toBe(0);
    } finally { b.stop(); }
  });

  it.each(["wrong-founder", "old-run", "future-run", "wrong-start-time"])("advances recovery position without reviving %s", async (cohort) => {
    const value = payload();
    if (cohort === "wrong-founder") value.founder_id = "01985555-3333-7333-8333-333333333333";
    else if (cohort === "wrong-start-time") value.started_at_ms = now - 1;
    else value.run_id = { company_stream_id: companyID, run_seq: cohort === "old-run" ? 1 : 3 };
    const b = await boundary({ saved: true, initial: snapshot(4, 2), handshake: {
      recovered: true, offset: 2, publications: [{ offset: 2, data: wire(value, 3) }],
    } });
    try {
      expect(b.received).toEqual([{ kind: "transport_recovered" }]);
      expect(positions(b.storage)[`player:${founderID}`]).toEqual({ epoch: "player", offset: 2 });
      expect(b.reads).toBe(1); expect(b.socket.closes).toBe(0);
    } finally { b.stop(); }
  });

  it.each(["live", "recovered-batch"])("never revives an older start after a newer start, %s", async (mode) => {
    const prefix: Publication[] = [{ offset: 2, data: wire(run2, 3) }, { offset: 3, data: endWire() }, { offset: 4, data: wire(run3, 6) }];
    const b = await boundary({ saved: true, initial: snapshot(4, 2), handshake: mode === "live"
      ? { recovered: true, offset: 1 }
      : { recovered: true, offset: 5, publications: [...prefix, { offset: 5, data: wire(run2, 3) }] } });
    try {
      const expectedPrefix = [startMessage(run2, 3), endMessage, startMessage(run3, 6)];
      if (mode === "live") {
        for (const publication of prefix) b.socket.reply({ push: { channel: `player:${founderID}`, pub: publication } });
        expect(b.received).toEqual([{ kind: "transport_recovered" }, ...expectedPrefix]);
        b.publish(run2, 5, 3);
        expect(b.received).toEqual([{ kind: "transport_recovered" }, ...expectedPrefix]);
      } else expect(b.received).toEqual([...expectedPrefix, { kind: "transport_recovered" }]);
      expect(positions(b.storage)[`player:${founderID}`]).toEqual({ epoch: "player", offset: 5 });
      expect(b.reads).toBe(1); expect(b.socket.closes).toBe(0);
    } finally { b.stop(); }
  });

  it("retains delivery identity across an actual1006 reconnect and ignores the old socket", async () => {
    const b = await boundary({ initial: snapshot(4, 2) });
    try {
      b.publish(run2, 1, 3);
      const newSocket = b.nextSocket(); const recovered = b.nextRecovery();
      b.socket.emit("close", { code: 1006 });
      const connection = await newSocket;
      expect(b.sockets).toHaveLength(2);
      b.connect(connection, { recovered: true, offset: 2, publications: [
        { offset: 1, data: wire(run2, 3) }, { offset: 2, data: wire(run2, 3) },
      ] });
      await recovered;
      expect(connection.sent.map((command) => JSON.parse(command))).toEqual([
        { id: 1, connect: { token: "access" } },
        { id: 2, subscribe: { channel: `player:${founderID}`, recover: true, epoch: "player", offset: 1 } },
        { id: 3, subscribe: { channel: "world", recover: true, epoch: "world", offset: 0 } },
      ]);
      expect(b.received).toEqual([{ kind: "transport_recovered" }, startMessage(run2, 3), { kind: "transport_recovered" }]);
      b.socket.reply({ push: { channel: `player:${founderID}`, pub: { offset: 3, data: endWire() } } });
      b.publish(run3, 4, 6); // Old connection cannot deliver or induce resync.
      expect(b.reads).toBe(1);
      expect(b.received).toHaveLength(3);
      connection.reply({ push: { channel: `player:${founderID}`, pub: { offset: 3, data: endWire() } } });
      b.publish(run3, 4, 6, connection);
      expect(b.received).toEqual([{ kind: "transport_recovered" }, startMessage(run2, 3),
        { kind: "transport_recovered" }, endMessage, startMessage(run3, 6)]);
      expect(positions(b.storage)[`player:${founderID}`]).toEqual({ epoch: "player", offset: 4 });
      expect(b.reads).toBe(1); expect(connection.closes).toBe(0);
    } finally { b.stop(); }
  });

  it("retains the newer-start high-water across reconnect while HTTP still describes the older run", async () => {
    const b = await boundary({ initial: snapshot(4, 2) });
    try {
      b.publish(run2, 1, 3);
      b.socket.reply({ push: { channel: `player:${founderID}`, pub: { offset: 2, data: endWire() } } });
      b.publish(run3, 3, 6);
      const before = [...b.received];
      const newSocket = b.nextSocket(); const recovered = b.nextRecovery();
      b.socket.emit("close", { code: 1006 });
      const connection = await newSocket;
      b.connect(connection, { recovered: true, offset: 5, publications: [
        { offset: 4, data: wire(run2, 3) }, { offset: 5, data: wire(run3, 6) },
      ] });
      await recovered;
      expect(b.received).toEqual([...before, { kind: "transport_recovered" }]);
      expect(positions(b.storage)[`player:${founderID}`]).toEqual({ epoch: "player", offset: 5 });
      expect(b.reads).toBe(1); expect(connection.closes).toBe(0);
    } finally { b.stop(); }
  });

  it("retains already-delivered identity through full-sync and fresh live resubscription", async () => {
    const b = await boundary({ initial: snapshot(4, 2) });
    try {
      b.publish(run2, 1, 3);
      const before = [...b.received];
      const newSocket = b.nextSocket();
      b.socket.emit("close", { code: 4000 });
      expect(await b.nextSnapshot).toMatchObject({ kind: "snapshot", value: { revision: 4, run: { run_seq: 2 } } });
      const connection = await newSocket;
      b.connect(connection);
      expect(b.received).toEqual([...before, { kind: "system", value: { kind: "resync_required" } },
        expect.objectContaining({ kind: "snapshot" }), { kind: "transport_recovered" }]);
      b.publish(run2, 1, 3, connection);
      expect(b.received.filter((message) => (message as { kind: string }).kind === "event")).toEqual([startMessage(run2, 3)]);
      expect(b.reads).toBe(2); expect(connection.closes).toBe(0);
    } finally { b.stop(); }
  });

  it.each(["unrecovered", "changed-epoch", "malformed-summary", "revision-gap"])("fetches fresh state then resubscribes without accepting %s history", async (fault) => {
    const b = await boundary({ saved: true, initial: snapshot(4, 2), deferHandshake: true });
    try {
      const fresh = { ...snapshot(6, 3), evaluated_through_ms: now + 1000, server_now_ms: now + 1000,
        run: { ...snapshot(6, 3).run, run_started_at_ms: now + 1000 } };
      expect(parseGameUISnapshot(fresh)).toMatchObject(fresh);
      b.setSnapshot(fresh);
      const newSocket = b.nextSocket();
      b.connect(b.socket, { recovered: fault !== "unrecovered", epoch: fault === "changed-epoch" ? "new-player" : "player",
        offset: 2, publications: [{ offset: 2, data: wire(fault === "malformed-summary"
          ? { ...run2, reputation_tree: { ...summary, bonus_factor: "9e-1" } }
          : fault === "revision-gap" ? run3 : run2, fault === "revision-gap" ? 6 : 3) }] });
      expect(b.received).toEqual([{ kind: "system", value: { kind: "resync_required" } }]);
      expect(b.socket.closes).toBe(1); expect(b.reads).toBe(2);
      expect(await b.nextSnapshot).toMatchObject({ kind: "snapshot", value: fresh });
      const connection = await newSocket;
      b.connect(connection);
      expect(connection.sent.map((command) => JSON.parse(command)).slice(1)).toEqual([
        { id: 2, subscribe: { channel: `player:${founderID}` } }, { id: 3, subscribe: { channel: "world" } },
      ]);
      expect(b.received.filter((message) => (message as { kind: string }).kind === "event")).toEqual([]);
      b.publish(run2, 1, 3, connection); // Fresh run3 state cannot revive old run2.
      b.publish(run3, 2, 6, connection);
      expect(b.received.filter((message) => (message as { kind: string }).kind === "event")).toEqual([startMessage(run3, 6)]);
      expect(positions(b.storage)[`player:${founderID}`]).toEqual({ epoch: "player", offset: 2 });
      expect(b.reads).toBe(2); expect(connection.closes).toBe(0);
    } finally { b.stop(); }
  });
});
