import { afterEach, describe, expect, it, vi } from "vitest";

import { createBrowserGameUIRuntime, newIntentID, type RuntimeStorage } from "../src/game-ui/runtime";

const snapshot = {
  constants_hash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  evaluated_through_ms: 1_800_000_000_000,
  facts: [{ fact_id: "bootstrap.needed", value: false }],
  generators: [],
  manual_action: { action_id: "manual.click", bucket_cap_milli: 50_000, refill_milli_per_ms: 25, refilled_at_ms: 1_800_000_000_000, tokens_milli: 50_000 },
  progress: [], resources: [], revision: 1,
  run: { category: "any_percent", exit_count: 0, founder_id: "01985555-1111-7111-8111-111111111111", run_seq: 1, run_started_at_ms: 1_799_999_000_000, tier: 0 },
  schema_version: 1, server_now_ms: 1_800_000_000_000, upgrades: [],
};
const currentSnapshot = { ...snapshot, founder_revision: 1, schema_version: 3,
  transitions: { cross_gate: { eligible: false, gate_id: "gate.t0_to_t1", route_id: null }, wind_down: { eligible: false } } };

class MemoryStorage implements RuntimeStorage {
  readonly values = new Map<string, string>();
  readonly writes: string[] = [];
  getItem(key: string): string | null { return this.values.get(key) ?? null; }
  setItem(key: string, value: string): void { this.values.set(key, value); this.writes.push(key); }
  removeItem(key: string): void { this.values.delete(key); }
}

type FakeSocketListener = (event: { data?: string; code?: number }) => void;

class FakeSocket {
  readonly listeners = new Map<string, FakeSocketListener[]>();
  readonly sent: string[] = [];
  closeCount = 0;
  addEventListener(kind: string, listener: FakeSocketListener): void { this.listeners.set(kind, [...(this.listeners.get(kind) ?? []), listener]); }
  send(value: string): void { this.sent.push(value); }
  close(): void { this.closeCount += 1; }
  emit(kind: string, event: { data?: string; code?: number } = {}): void { for (const listener of this.listeners.get(kind) ?? []) listener(event); }
  reply(value: unknown): void { this.emit("message", { data: JSON.stringify(value) }); }
}

function openAndConnect(socket: FakeSocket): void {
  socket.emit("open");
  socket.reply({ id: 1, connect: {} });
}

function subscribeReplies(socket: FakeSocket, options: { playerOffset?: number; worldOffset?: number; recovered?: boolean; publications?: unknown[] } = {}): void {
  const recovered = options.recovered ?? false;
  socket.reply({ id: 2, subscribe: { recoverable: true, positioned: true, recovered, epoch: "player-epoch", offset: options.playerOffset ?? 0, publications: options.publications ?? [] } });
  socket.reply({ id: 3, subscribe: { recoverable: true, positioned: true, recovered, epoch: "world-epoch", offset: options.worldOffset ?? 0, publications: [] } });
}

function eventEnvelope(revision: number, eventID = `event-${revision}`, cursorEffect: "advance" | "historical" = "advance"): unknown {
  const historical = cursorEffect === "historical";
  return {
    v: 2, ch: `player:${snapshot.run.founder_id}`, kind: "event", rev: revision, constants_hash: snapshot.constants_hash, ts: "2026-08-11T12:00:00Z",
    payload: {
      event_id: eventID, kind: historical ? "compensation" : "gate_crossed", scope: "company", rev: revision, cursor_effect: cursorEffect,
      payload: historical ? {} : { founder_id: snapshot.run.founder_id, gate_id: "gate.t0_to_t1", route_id: null, run_id: { company_stream_id: "01985555-2222-7222-8222-222222222222", run_seq: 1 } },
    },
  };
}

function publication(socket: FakeSocket, channel: string, offset: number, data: unknown): void {
  socket.reply({ push: { channel, pub: { offset, data } } });
}

afterEach(() => { vi.useRealTimers(); });

describe("browser Game UI runtime", () => {
  it("mints the server-required UUIDv7 intent identity", () => {
    const cryptoSource = { getRandomValues(value: Uint8Array) { value.fill(0xff); return value; } } as Crypto;
    expect(newIntentID(cryptoSource, 1_800_000_000_000)).toBe("01a3185c-5000-7fff-bfff-ffffffffffff");
    expect(newIntentID(cryptoSource, 1_800_000_000_000)).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/u);
    expect(() => newIntentID(cryptoSource, -1)).toThrow(/timestamp/);
  });

  it("persists the retry key before bootstrap and credentials before returning the snapshot", async () => {
    const storage = new MemoryStorage();
    const requests: Array<{ input: string; init?: RequestInit }> = [];
    const fetcher = async (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
      requests.push({ input: String(input), init });
      return new Response(JSON.stringify({
        account: { account_id: "account", created_at: "2026-08-10T12:00:00.000Z", recovery_code: "recover" },
        session: { access_token: "access", refresh_token: "refresh" }, game_ui_snapshot: snapshot,
      }), { status: 201, headers: { "Content-Type": "application/json" } });
    };
    const cryptoSource = { getRandomValues(value: Uint8Array) { value.fill(0xab); return value; }, randomUUID: () => "01985555-1111-7111-8111-111111111119" } as unknown as Crypto;
    const runtime = createBrowserGameUIRuntime(storage, fetcher as typeof fetch, cryptoSource);
    expect(runtime.hasCredentials()).toBe(false);
    expect((await runtime.bootstrap()).revision).toBe(1);
    expect(requests[0]).toMatchObject({ input: "/api/v1/bootstrap", init: { method: "POST" } });
    expect(requests[0].init?.body).toBe(JSON.stringify({ idempotency_key: "ab".repeat(32) }));
    expect(storage.writes).toEqual(["cloud-clicker.bootstrap-key.v1", "cloud-clicker.credentials.v1"]);
    expect(runtime.hasCredentials()).toBe(true);
  });

  it("authenticates snapshots and intents from the one persisted credential document", async () => {
    const storage = new MemoryStorage();
    storage.setItem("cloud-clicker.credentials.v1", JSON.stringify({ accessToken: "access", refreshToken: "refresh", accountID: "account", recoveryCode: "recover" }));
    const requests: RequestInit[] = [];
    const fetcher = async (_input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
      requests.push(init ?? {});
      return new Response(JSON.stringify(requests.length === 1 ? currentSnapshot : { outcome: "applied", new_revision: 2 }), { status: 200, headers: { "Content-Type": "application/json" } });
    };
    const runtime = createBrowserGameUIRuntime(storage, fetcher as typeof fetch);
    expect((await runtime.snapshot()).revision).toBe(1);
    await runtime.intent({ kind: "perform_manual_batch" });
    expect(requests.map((request) => (request.headers as Record<string, string>).Authorization)).toEqual(["Bearer access", "Bearer access"]);
  });

  it("rejects a legacy receipt snapshot on the live sync operation", async () => {
    const storage = new MemoryStorage();
    storage.setItem("cloud-clicker.credentials.v1", JSON.stringify({ accessToken: "access", refreshToken: "refresh", accountID: "account", recoveryCode: "recover" }));
    const runtime = createBrowserGameUIRuntime(storage, async () => new Response(JSON.stringify(snapshot), { status: 200 }));
    await expect(runtime.snapshot()).rejects.toThrow(/schema v3/);
  });

  it("records initial stream positions and decodes raw publications inside the runtime boundary", () => {
    const storage = new MemoryStorage();
    storage.setItem("cloud-clicker.credentials.v1", JSON.stringify({ accessToken: "access", refreshToken: "refresh", accountID: "account", recoveryCode: "recover" }));
    const socket = new FakeSocket();
    const runtime = createBrowserGameUIRuntime(storage, fetch, crypto, () => socket as unknown as WebSocket, { protocol: "http:", host: "localhost" });
    const received: unknown[] = [];
    runtime.subscribe(snapshot.run.founder_id, (message) => received.push(message));
    openAndConnect(socket);
    expect(socket.sent).toEqual([
      JSON.stringify({ id: 1, connect: { token: "access" } }),
      JSON.stringify({ id: 2, subscribe: { channel: `player:${snapshot.run.founder_id}` } }),
      JSON.stringify({ id: 3, subscribe: { channel: "world" } }),
    ]);
    subscribeReplies(socket);
    const envelope = { v: 2, ch: "world", kind: "presence", rev: 1, constants_hash: snapshot.constants_hash, ts: "2026-08-11T12:00:00Z", payload: { joined: [], left: [], count: 7 } };
    publication(socket, "world", 1, envelope);
    const world = { v: 2, ch: "world", kind: "snapshot", rev: 2, constants_hash: snapshot.constants_hash, ts: "2026-08-11T12:00:01Z", payload: { scope: "world", rev: 2, state: { v: 1, world_rev: 2, planet: { depletion_ppm: 0, health_ppm: 0 }, commons: { server_health_ppm: 0, active_founders: 0, compact_members: 0 }, population: { online: 3, founders_total: 1 }, milestones: { active_id: null, progress_ppm: 0 }, epoch: { epoch_id: 6, name: "First Content" } } } };
    publication(socket, "world", 2, world);
    publication(socket, "world", 3, { kind: "future_kind" });
    expect(received).toEqual([
      { kind: "transport_recovered" },
      { kind: "presence", count: 7 },
      { kind: "presence", count: 3 },
    ]);
    expect(JSON.parse(storage.getItem(`cloud-clicker.transport.v1.${snapshot.run.founder_id}`)!)).toEqual({
      [`player:${snapshot.run.founder_id}`]: { epoch: "player-epoch", offset: 0 },
      world: { epoch: "world-epoch", offset: 3 },
    });
  });

  it("recovers saved positions, suppresses duplicates, and does not let historical events move the scope cursor", async () => {
    vi.useFakeTimers();
    const storage = new MemoryStorage();
    storage.setItem("cloud-clicker.credentials.v1", JSON.stringify({ accessToken: "access", refreshToken: "refresh", accountID: "account", recoveryCode: "recover" }));
    const sockets: FakeSocket[] = [];
    const runtime = createBrowserGameUIRuntime(storage, async () => new Response(JSON.stringify(currentSnapshot), { status: 200 }), crypto, () => {
      const socket = new FakeSocket(); sockets.push(socket); return socket as unknown as WebSocket;
    }, { protocol: "http:", host: "localhost" });
    await runtime.snapshot();
    const received: unknown[] = [];
    runtime.subscribe(snapshot.run.founder_id, (message) => received.push(message));
    openAndConnect(sockets[0]); subscribeReplies(sockets[0]);
    publication(sockets[0], `player:${snapshot.run.founder_id}`, 1, eventEnvelope(2));
    publication(sockets[0], `player:${snapshot.run.founder_id}`, 2, eventEnvelope(2));
    publication(sockets[0], `player:${snapshot.run.founder_id}`, 3, eventEnvelope(99, "historical-99", "historical"));
    sockets[0].emit("close", { code: 1006 });
    await vi.runAllTimersAsync();
    openAndConnect(sockets[1]);
    expect(JSON.parse(sockets[1].sent[1])).toEqual({ id: 2, subscribe: { channel: `player:${snapshot.run.founder_id}`, recover: true, epoch: "player-epoch", offset: 3 } });
    expect(JSON.parse(sockets[1].sent[2])).toEqual({ id: 3, subscribe: { channel: "world", recover: true, epoch: "world-epoch", offset: 0 } });
    subscribeReplies(sockets[1], { recovered: true, playerOffset: 4, publications: [{ offset: 4, data: eventEnvelope(3, "event-3") }] });
    expect(received.filter((value) => (value as { kind?: string }).kind === "event")).toHaveLength(2);
    expect(received).toContainEqual({ kind: "historical_event", revision: 99, scope: "company", eventID: "historical-99", eventKind: "compensation", value: {} });
    expect(received).toContainEqual({ kind: "transport_recovered" });
    expect(JSON.parse(storage.getItem(`cloud-clicker.transport.v1.${snapshot.run.founder_id}`)!)[`player:${snapshot.run.founder_id}`]).toEqual({ epoch: "player-epoch", offset: 4 });
  });

  it("full-syncs and resubscribes live for a revision gap or unrecoverable history", async () => {
    vi.useFakeTimers();
    const storage = new MemoryStorage();
    storage.setItem("cloud-clicker.credentials.v1", JSON.stringify({ accessToken: "access", refreshToken: "refresh", accountID: "account", recoveryCode: "recover" }));
    const requests: string[] = [];
    const sockets: FakeSocket[] = [];
    const runtime = createBrowserGameUIRuntime(storage, async (input) => {
      requests.push(String(input));
      return new Response(JSON.stringify(requests.length === 1 ? currentSnapshot : { ...currentSnapshot, revision: 3 }), { status: 200 });
    }, crypto, () => {
      const socket = new FakeSocket(); sockets.push(socket); return socket as unknown as WebSocket;
    }, { protocol: "http:", host: "localhost" });
    await runtime.snapshot();
    const received: unknown[] = [];
    runtime.subscribe(snapshot.run.founder_id, (message) => received.push(message));
    openAndConnect(sockets[0]); subscribeReplies(sockets[0]);
    publication(sockets[0], `player:${snapshot.run.founder_id}`, 1, eventEnvelope(3));
    await vi.runAllTimersAsync();
    await vi.waitFor(() => expect(received).toContainEqual({ kind: "snapshot", value: { ...currentSnapshot, revision: 3 } }));
    await vi.runAllTimersAsync();
    expect(requests).toEqual(["/api/v1/founder/state", "/api/v1/founder/state"]);
    await vi.waitFor(() => expect(sockets).toHaveLength(2));
    openAndConnect(sockets[1]);
    expect(JSON.parse(sockets[1].sent[1])).toEqual({ id: 2, subscribe: { channel: `player:${snapshot.run.founder_id}` } });
    subscribeReplies(sockets[1]);
    sockets[1].emit("close", { code: 1006 });
    await vi.runAllTimersAsync();
    openAndConnect(sockets[2]);
    sockets[2].reply({ id: 2, subscribe: { recoverable: true, positioned: true, recovered: false, epoch: "new-player", offset: 0, publications: [] } });
    await vi.runAllTimersAsync();
    await vi.waitFor(() => expect(requests).toHaveLength(3));
    await vi.runAllTimersAsync();
    expect(requests).toHaveLength(3);
    expect(sockets[2].closeCount).toBe(1);
    await vi.waitFor(() => expect(sockets).toHaveLength(4));
    openAndConnect(sockets[3]);
    expect(JSON.parse(sockets[3].sent[1]).subscribe).toEqual({ channel: `player:${snapshot.run.founder_id}` });
  });

  it("full-syncs overflow and invalid frames, delays drain recovery, and fails credential expiry visibly", async () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-08-21T12:00:00Z"));
    const storage = new MemoryStorage();
    storage.setItem("cloud-clicker.credentials.v1", JSON.stringify({ accessToken: "access", refreshToken: "refresh", accountID: "account", recoveryCode: "recover" }));
    const sockets: FakeSocket[] = [];
    let snapshots = 0;
    const runtime = createBrowserGameUIRuntime(storage, async () => { snapshots += 1; return new Response(JSON.stringify(currentSnapshot), { status: 200 }); }, crypto, () => {
      const socket = new FakeSocket(); sockets.push(socket); return socket as unknown as WebSocket;
    }, { protocol: "http:", host: "localhost" });
    const received: unknown[] = [];
    runtime.subscribe(snapshot.run.founder_id, (message) => received.push(message));
    openAndConnect(sockets[0]); subscribeReplies(sockets[0]);
    sockets[0].emit("close", { code: 4000 });
    await vi.runAllTimersAsync();
    await vi.waitFor(() => expect(snapshots).toBe(1));
    await vi.runAllTimersAsync();
    await vi.waitFor(() => expect(sockets).toHaveLength(2));
    openAndConnect(sockets[1]); subscribeReplies(sockets[1]);
    sockets[1].emit("close", { code: 4004 });
    await vi.runAllTimersAsync();
    await vi.waitFor(() => expect(snapshots).toBe(2));
    await vi.runAllTimersAsync();
    await vi.waitFor(() => expect(sockets).toHaveLength(3));
    openAndConnect(sockets[2]); subscribeReplies(sockets[2]);
    publication(sockets[2], "world", 0, { v: 2, ch: "world", kind: "system", rev: 0, constants_hash: snapshot.constants_hash, ts: "2026-08-21T12:00:00Z", payload: { code: "server_restarting", resume_after_ms: 5000 } });
    sockets[2].emit("close", { code: 4003 });
    await vi.advanceTimersByTimeAsync(4999);
    expect(sockets).toHaveLength(3);
    await vi.advanceTimersByTimeAsync(1);
    expect(sockets).toHaveLength(4);
    openAndConnect(sockets[3]); subscribeReplies(sockets[3], { recovered: true });
    sockets[3].emit("close", { code: 4001 });
    await vi.runAllTimersAsync();
    expect(sockets).toHaveLength(4);
    expect(received.at(-1)).toEqual({ kind: "transport_closed" });
  });

  it("stops a replaced connection without retrying or refreshing credentials", async () => {
    vi.useFakeTimers();
    const storage = new MemoryStorage();
    storage.setItem("cloud-clicker.credentials.v1", JSON.stringify({ accessToken: "access", refreshToken: "refresh", accountID: "account", recoveryCode: "recover" }));
    const sockets: FakeSocket[] = [];
    let requests = 0;
    const runtime = createBrowserGameUIRuntime(storage, async () => { requests += 1; return new Response("{}"); }, crypto, () => {
      const socket = new FakeSocket(); sockets.push(socket); return socket as unknown as WebSocket;
    }, { protocol: "http:", host: "localhost" });
    const received: unknown[] = [];
    runtime.subscribe(snapshot.run.founder_id, (message) => received.push(message));
    openAndConnect(sockets[0]); subscribeReplies(sockets[0]);
    sockets[0].emit("close", { code: 4002 });
    await vi.runAllTimersAsync();
    expect(sockets).toHaveLength(1);
    expect(requests).toBe(0);
    expect(received.at(-1)).toEqual({ kind: "transport_closed" });
  });
});
