import type { BootstrapResponse, GameUISnapshot } from "../api/generated/types";
import { decodeTransportEnvelope } from "../transport";
import { parseGameUISnapshot } from "./contracts";
import { decodeGameUIEvent, decodeGameUISystemEvent, type GameUILifecycleEvent, type GameUISystemEvent } from "./events";

export interface GameUICredentials {
  readonly accessToken: string;
  readonly refreshToken: string;
  readonly accountID: string;
  readonly recoveryCode: string;
}

export type GameUIRuntimeMessage =
  | Readonly<{ kind: "event"; revision: number; scope: "company" | "founder"; value: GameUILifecycleEvent }>
  | Readonly<{ kind: "presence"; count: number }>
  | Readonly<{ kind: "receipt" }>
  | Readonly<{ kind: "system"; value: GameUISystemEvent }>
  | Readonly<{ kind: "transport_closed" }>;

export interface GameUIRuntime {
  hasCredentials(): boolean;
  bootstrap(): Promise<GameUISnapshot>;
  snapshot(): Promise<GameUISnapshot>;
  intent(body: Readonly<Record<string, unknown>>): Promise<void>;
  subscribe(founderID: string, listener: (message: GameUIRuntimeMessage) => void): () => void;
}

export interface RuntimeStorage {
  getItem(key: string): string | null;
  setItem(key: string, value: string): void;
  removeItem(key: string): void;
}

const credentialKey = "cloud-clicker.credentials.v1";
const bootstrapKey = "cloud-clicker.bootstrap-key.v1";

function randomHex(bytes: number, cryptoSource: Crypto): string {
  const value = new Uint8Array(bytes);
  cryptoSource.getRandomValues(value);
  return [...value].map((part) => part.toString(16).padStart(2, "0")).join("");
}

function credentials(storage: RuntimeStorage): GameUICredentials | undefined {
  try {
    const raw = storage.getItem(credentialKey);
    if (!raw) return undefined;
    const value = JSON.parse(raw) as Record<string, unknown>;
    if (typeof value.accessToken !== "string" || typeof value.refreshToken !== "string" || typeof value.accountID !== "string" || typeof value.recoveryCode !== "string") return undefined;
    return { accessToken: value.accessToken, refreshToken: value.refreshToken, accountID: value.accountID, recoveryCode: value.recoveryCode };
  } catch { return undefined; }
}

async function responseJSON(response: Response): Promise<unknown> {
  const value = await response.json();
  if (!response.ok) throw new Error(`game UI request failed (${response.status})`);
  return value;
}

export function createBrowserGameUIRuntime(
  storage: RuntimeStorage = localStorage,
  fetcher: typeof fetch = fetch,
  cryptoSource: Crypto = crypto,
  socketFactory: (url: string) => WebSocket = (url) => new WebSocket(url),
  locationSource?: Readonly<{ protocol: string; host: string }>,
): GameUIRuntime {
  const authHeaders = (): HeadersInit => {
    const current = credentials(storage);
    if (!current) throw new Error("missing game UI credentials");
    return { Authorization: `Bearer ${current.accessToken}`, "Content-Type": "application/json" };
  };
  return {
    hasCredentials: () => credentials(storage) !== undefined,
    async bootstrap() {
      let pending = storage.getItem(bootstrapKey);
      if (!pending) { pending = randomHex(32, cryptoSource); storage.setItem(bootstrapKey, pending); }
      const value = await responseJSON(await fetcher("/api/v1/bootstrap", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ idempotency_key: pending }) })) as BootstrapResponse;
      const parsed = parseGameUISnapshot(value.game_ui_snapshot);
      storage.setItem(credentialKey, JSON.stringify({ accessToken: value.session.access_token, refreshToken: value.session.refresh_token, accountID: value.account.account_id, recoveryCode: value.account.recovery_code } satisfies GameUICredentials));
      storage.removeItem(bootstrapKey);
      return parsed;
    },
    async snapshot() {
      return parseGameUISnapshot(await responseJSON(await fetcher("/api/v1/founder/state", { headers: authHeaders() })));
    },
    async intent(body) {
      await responseJSON(await fetcher("/api/v1/intents", { method: "POST", headers: authHeaders(), body: JSON.stringify(body) }));
    },
    subscribe(founderID, listener) {
      const current = credentials(storage);
      if (!current) return () => {};
      const socketLocation = locationSource ?? location;
      const scheme = socketLocation.protocol === "https:" ? "wss:" : "ws:";
      const socket = socketFactory(`${scheme}//${socketLocation.host}/connection/websocket`);
      let closed = false;
      socket.addEventListener("open", () => socket.send(JSON.stringify({ id: 1, connect: { token: current.accessToken } })));
      socket.addEventListener("message", (event) => {
        const decoder = typeof event.data === "string" ? event.data : "";
        for (const line of decoder.split("\n").filter(Boolean)) {
          try {
            const value = JSON.parse(line) as Record<string, unknown>;
            if (value.id === 1 && value.connect) {
              socket.send(JSON.stringify({ id: 2, subscribe: { channel: `player:${founderID}` } }));
              socket.send(JSON.stringify({ id: 3, subscribe: { channel: "world" } }));
            }
            const push = value.push as { pub?: { data?: unknown } } | undefined;
            if (push?.pub?.data === undefined) continue;
            let raw = push.pub.data;
            if (typeof raw === "string") raw = JSON.parse(raw);
            const envelope = decodeTransportEnvelope(raw);
            if (!envelope) {
              listener({ kind: "system", value: { kind: "resync_required" } });
              continue;
            }
            if (envelope.kind === "receipt") {
              listener({ kind: "receipt" });
              continue;
            }
            if (envelope.kind === "presence" && envelope.ch === "world") {
              listener({ kind: "presence", count: envelope.payload.count as number });
              continue;
            }
            if (envelope.kind === "system") {
              const system = decodeGameUISystemEvent(envelope);
              if (system) listener({ kind: "system", value: system });
              continue;
            }
            if (envelope.kind === "event") {
              const event = decodeGameUIEvent(envelope);
              if (event) listener({ kind: "event", revision: envelope.rev, scope: envelope.payload.scope as "company" | "founder", value: event });
            }
          } catch {
            listener({ kind: "system", value: { kind: "resync_required" } });
          }
        }
      });
      socket.addEventListener("close", () => { if (!closed) listener({ kind: "transport_closed" }); });
      return () => { closed = true; socket.close(); };
    },
  };
}

export function newIntentID(cryptoSource: Crypto = crypto): string { return cryptoSource.randomUUID(); }
