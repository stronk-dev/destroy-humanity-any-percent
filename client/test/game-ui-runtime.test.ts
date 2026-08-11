import { describe, expect, it } from "vitest";

import { createBrowserGameUIRuntime, type RuntimeStorage } from "../src/game-ui/runtime";

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

class MemoryStorage implements RuntimeStorage {
  readonly values = new Map<string, string>();
  readonly writes: string[] = [];
  getItem(key: string): string | null { return this.values.get(key) ?? null; }
  setItem(key: string, value: string): void { this.values.set(key, value); this.writes.push(key); }
  removeItem(key: string): void { this.values.delete(key); }
}

describe("browser Game UI runtime", () => {
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
      return new Response(JSON.stringify(requests.length === 1 ? snapshot : { outcome: "applied", new_revision: 2 }), { status: 200, headers: { "Content-Type": "application/json" } });
    };
    const runtime = createBrowserGameUIRuntime(storage, fetcher as typeof fetch);
    expect((await runtime.snapshot()).revision).toBe(1);
    await runtime.intent({ kind: "perform_manual_batch" });
    expect(requests.map((request) => (request.headers as Record<string, string>).Authorization)).toEqual(["Bearer access", "Bearer access"]);
  });

  it("decodes raw socket publications inside the runtime boundary", () => {
    const storage = new MemoryStorage();
    storage.setItem("cloud-clicker.credentials.v1", JSON.stringify({ accessToken: "access", refreshToken: "refresh", accountID: "account", recoveryCode: "recover" }));
    const listeners = new Map<string, ((event: { data?: string }) => void)[]>();
    const sent: string[] = [];
    const socket = {
      addEventListener(kind: string, listener: (event: { data?: string }) => void) { listeners.set(kind, [...(listeners.get(kind) ?? []), listener]); },
      send(value: string) { sent.push(value); },
      close() {},
    } as unknown as WebSocket;
    const runtime = createBrowserGameUIRuntime(storage, fetch, crypto, () => socket, { protocol: "http:", host: "localhost" });
    const received: unknown[] = [];
    runtime.subscribe(snapshot.run.founder_id, (message) => received.push(message));
    for (const listener of listeners.get("open") ?? []) listener({});
    for (const listener of listeners.get("message") ?? []) listener({ data: JSON.stringify({ id: 1, connect: {} }) });
    expect(sent).toEqual([
      JSON.stringify({ id: 1, connect: { token: "access" } }),
      JSON.stringify({ id: 2, subscribe: { channel: `player:${snapshot.run.founder_id}` } }),
      JSON.stringify({ id: 3, subscribe: { channel: "world" } }),
    ]);
    const envelope = { v: 2, ch: "world", kind: "presence", rev: 1, constants_hash: snapshot.constants_hash, ts: "2026-08-11T12:00:00Z", payload: { joined: [], left: [], count: 7 } };
    for (const listener of listeners.get("message") ?? []) listener({ data: JSON.stringify({ push: { pub: { data: envelope } } }) });
    for (const listener of listeners.get("message") ?? []) listener({ data: JSON.stringify({ push: { pub: { data: {} } } }) });
    for (const listener of listeners.get("message") ?? []) listener({ data: "{" });
    expect(received).toEqual([
      { kind: "presence", count: 7 },
      { kind: "system", value: { kind: "resync_required" } },
      { kind: "system", value: { kind: "resync_required" } },
    ]);
  });
});
