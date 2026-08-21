import type { BootstrapResponse } from "../api/generated/types";
import { decodeTransportEnvelope, decodeWorldSnapshot, PlayerRevisionCursor } from "../transport";
import { parseGameUISnapshot, type ParsedGameUISnapshot } from "./contracts";
import { decodeGameUIEvent, decodeGameUISystemEvent, type GameUILifecycleEvent, type GameUISystemEvent } from "./events";

export interface GameUICredentials {
  readonly accessToken: string;
  readonly refreshToken: string;
  readonly accountID: string;
  readonly recoveryCode: string;
}

export type GameUIRuntimeMessage =
  | Readonly<{ kind: "event"; revision: number; scope: "company" | "founder"; value: GameUILifecycleEvent }>
  | Readonly<{ kind: "historical_event"; revision: number; scope: "company" | "founder"; eventID: string; eventKind: string; value: Readonly<Record<string, unknown>> }>
  | Readonly<{ kind: "presence"; count: number }>
  | Readonly<{ kind: "receipt" }>
  | Readonly<{ kind: "snapshot"; value: ParsedGameUISnapshot }>
  | Readonly<{ kind: "system"; value: GameUISystemEvent }>
  | Readonly<{ kind: "transport_recovered" }>
  | Readonly<{ kind: "transport_closed" }>;

export interface GameUIRuntime {
  hasCredentials(): boolean;
  bootstrap(): Promise<ParsedGameUISnapshot>;
  snapshot(): Promise<ParsedGameUISnapshot>;
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
const transportKeyPrefix = "cloud-clicker.transport.v1.";
const reconnectDelayMS = 1_000;

interface StreamPosition { readonly epoch: string; readonly offset: number }
type StreamPositions = Record<string, StreamPosition>;

function streamPositions(storage: RuntimeStorage, key: string, channels: ReadonlySet<string>): StreamPositions {
  try {
    const raw = storage.getItem(key);
    if (!raw) return {};
    const value = JSON.parse(raw) as Record<string, unknown>;
    const result: StreamPositions = {};
    for (const [channel, position] of Object.entries(value)) {
      if (!channels.has(channel) || typeof position !== "object" || position === null || Array.isArray(position)) return {};
      const row = position as Record<string, unknown>;
      if (Object.keys(row).sort().join("\0") !== "epoch\0offset" || typeof row.epoch !== "string" || row.epoch.length === 0 ||
          !Number.isSafeInteger(row.offset) || (row.offset as number) < 0) return {};
      result[channel] = { epoch: row.epoch, offset: row.offset as number };
    }
    return result;
  } catch { return {}; }
}

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
  let latestSnapshot: ParsedGameUISnapshot | undefined;
  let activeCursor: PlayerRevisionCursor | undefined;
  const authHeaders = (): HeadersInit => {
    const current = credentials(storage);
    if (!current) throw new Error("missing game UI credentials");
    return { Authorization: `Bearer ${current.accessToken}`, "Content-Type": "application/json" };
  };
  const rememberSnapshot = (value: ParsedGameUISnapshot): ParsedGameUISnapshot => {
    latestSnapshot = value;
    activeCursor?.reset("company", value.revision);
    activeCursor?.reset("founder", "founder_revision" in value ? value.founder_revision : 0);
    return value;
  };
  const loadSnapshot = async (): Promise<ParsedGameUISnapshot> => {
    const parsed = parseGameUISnapshot(await responseJSON(await fetcher("/api/v1/founder/state", { headers: authHeaders() })));
    if (parsed.schema_version !== 2 || !("founder_revision" in parsed)) throw new SyntaxError("live Game UI snapshot must use schema v2");
    return rememberSnapshot(parsed);
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
      return rememberSnapshot(parsed);
    },
    snapshot: loadSnapshot,
    async intent(body) {
      await responseJSON(await fetcher("/api/v1/intents", { method: "POST", headers: authHeaders(), body: JSON.stringify(body) }));
    },
    subscribe(founderID, listener) {
      const socketLocation = locationSource ?? location;
      const scheme = socketLocation.protocol === "https:" ? "wss:" : "ws:";
      const playerChannel = `player:${founderID}`;
      const channels = new Set([playerChannel, "world"]);
      const positionKey = `${transportKeyPrefix}${founderID}`;
      let positions = streamPositions(storage, positionKey, channels);
      const cursor = new PlayerRevisionCursor();
      activeCursor = cursor;
      if (latestSnapshot) {
        cursor.reset("company", latestSnapshot.revision);
        cursor.reset("founder", "founder_revision" in latestSnapshot ? latestSnapshot.founder_revision : 0);
      }
      let disposed = false;
      let reconnectTimer: ReturnType<typeof setTimeout> | undefined;
      let drainUntilMS = 0;
      let fullSync: Promise<void> | undefined;
      let active: { socket: WebSocket; intentional: boolean } | undefined;

      const persistPositions = (): void => storage.setItem(positionKey, JSON.stringify(positions));
      const clearPositions = (): void => { positions = {}; storage.removeItem(positionKey); };
      const closeActive = (): void => {
        if (!active) return;
        active.intentional = true;
        active.socket.close();
        active = undefined;
      };
      const stop = (): void => {
        if (reconnectTimer !== undefined) clearTimeout(reconnectTimer);
        reconnectTimer = undefined;
        closeActive();
      };
      const terminal = (): void => { stop(); if (!disposed) listener({ kind: "transport_closed" }); };
      const scheduleConnect = (delayMS: number): void => {
        if (disposed || reconnectTimer !== undefined || fullSync !== undefined) return;
        reconnectTimer = setTimeout(() => { reconnectTimer = undefined; connect(); }, Math.max(0, delayMS));
      };
      const authoritativeResync = (notify = true): void => {
        if (disposed || fullSync !== undefined) return;
        stop();
        clearPositions();
        if (notify) listener({ kind: "system", value: { kind: "resync_required" } });
        fullSync = loadSnapshot().then((value) => {
          if (!disposed) listener({ kind: "snapshot", value });
        }).catch(() => {
          if (!disposed) listener({ kind: "transport_closed" });
          disposed = true;
        }).finally(() => {
          fullSync = undefined;
          if (!disposed) scheduleConnect(0);
        });
      };
      const consumeEnvelope = (channel: string, publication: { data?: unknown; offset?: unknown }): boolean => {
        if (!channels.has(channel) || publication.data === undefined || !Number.isSafeInteger(publication.offset) || (publication.offset as number) < 0) return false;
        const offset = publication.offset as number;
        let raw = publication.data;
        if (typeof raw === "string") raw = JSON.parse(raw);
        const envelope = decodeTransportEnvelope(raw);
        if (!envelope) {
          if (offset === 0) return true;
          const existing = positions[channel];
          if (!existing || offset <= existing.offset) return false;
          positions[channel] = { epoch: existing.epoch, offset };
          persistPositions();
          return true;
        }
        if (envelope.ch !== channel || offset === 0 && !(envelope.kind === "system" && envelope.payload.code === "server_restarting")) return false;
        if (envelope.kind === "event") {
          const disposition = cursor.event(envelope);
          if (disposition === "resync_required") { authoritativeResync(); return true; }
          if (disposition === "deliver") {
            const scope = envelope.payload.scope as "company" | "founder";
            if (envelope.payload.cursor_effect === "historical") {
              listener({ kind: "historical_event", revision: envelope.rev, scope, eventID: envelope.payload.event_id as string,
                eventKind: envelope.payload.kind as string, value: envelope.payload.payload as Readonly<Record<string, unknown>> });
            } else {
              const event = decodeGameUIEvent(envelope);
              if (event) listener({ kind: "event", revision: envelope.rev, scope, value: event });
            }
          }
        } else if (envelope.kind === "receipt") {
          listener({ kind: "receipt" });
        } else if (envelope.kind === "presence" && envelope.ch === "world") {
          listener({ kind: "presence", count: envelope.payload.count as number });
        } else if (envelope.kind === "snapshot" && envelope.ch === "world") {
          const world = decodeWorldSnapshot(envelope.payload.state);
          listener({ kind: "presence", count: world.population.online });
        } else if (envelope.kind === "system") {
          const system = decodeGameUISystemEvent(envelope);
          if (system?.kind === "server_restarting") drainUntilMS = Math.max(drainUntilMS, Date.now() + system.resume_after_ms);
          if (system) listener({ kind: "system", value: system });
          if (system?.kind === "resync_required" || envelope.payload.code === "history_expired") { authoritativeResync(system === undefined); return true; }
        }
        if (disposed || fullSync !== undefined) return true;
        if (offset === 0) return true;
        const existing = positions[channel];
        if (!existing || offset <= existing.offset) return false;
        positions[channel] = { epoch: existing.epoch, offset };
        persistPositions();
        return true;
      };
      const connect = (): void => {
        if (disposed || active) return;
        const current = credentials(storage);
        if (!current) { terminal(); return; }
        const socket = socketFactory(`${scheme}//${socketLocation.host}/connection/websocket`);
        const connection = { socket, intentional: false };
        active = connection;
        const requestedRecovery = new Map<number, string>();
        const subscribed = new Set<string>();
        socket.addEventListener("open", () => socket.send(JSON.stringify({ id: 1, connect: { token: current.accessToken } })));
        socket.addEventListener("message", (event) => {
          if (disposed || active !== connection) return;
          const decoder = typeof event.data === "string" ? event.data : "";
          for (const line of decoder.split("\n").filter(Boolean)) {
            try {
              const value = JSON.parse(line) as Record<string, unknown>;
              if (value.error !== undefined) { terminal(); return; }
              if (value.id === 1 && value.connect) {
                for (const [id, channel] of [[2, playerChannel], [3, "world"]] as const) {
                  const position = positions[channel];
                  const subscribe = position ? { channel, recover: true, epoch: position.epoch, offset: position.offset } : { channel };
                  if (position) requestedRecovery.set(id, channel);
                  socket.send(JSON.stringify({ id, subscribe }));
                }
                continue;
              }
              if ((value.id === 2 || value.id === 3) && value.subscribe && typeof value.subscribe === "object") {
                const channel = value.id === 2 ? playerChannel : "world";
                const result = value.subscribe as Record<string, unknown>;
                const resultOffset = result.offset ?? 0;
                if (requestedRecovery.has(value.id as number) && result.recovered !== true) { authoritativeResync(); return; }
                if (result.recoverable !== true || result.positioned !== true || typeof result.epoch !== "string" || result.epoch.length === 0 ||
                    !Number.isSafeInteger(resultOffset) || (resultOffset as number) < 0 || !Array.isArray(result.publications ?? [])) {
                  authoritativeResync(); return;
                }
                const prior = positions[channel];
                if (requestedRecovery.has(value.id as number) && prior?.epoch !== result.epoch) { authoritativeResync(); return; }
                positions[channel] = { epoch: result.epoch, offset: prior?.offset ?? 0 };
                for (const publication of (result.publications ?? []) as Array<{ data?: unknown; offset?: unknown }>) {
                  if (!consumeEnvelope(channel, publication)) { authoritativeResync(); return; }
                  if (fullSync !== undefined) return;
                }
                if ((resultOffset as number) < positions[channel].offset) { authoritativeResync(); return; }
                positions[channel] = { epoch: result.epoch, offset: resultOffset as number };
                persistPositions();
                subscribed.add(channel);
                if (subscribed.size === channels.size) listener({ kind: "transport_recovered" });
                continue;
              }
              const push = value.push as { channel?: unknown; pub?: { data?: unknown; offset?: unknown } } | undefined;
              if (push?.pub && typeof push.channel === "string" && !consumeEnvelope(push.channel, push.pub)) { authoritativeResync(); return; }
              if (fullSync !== undefined) return;
            } catch { authoritativeResync(); return; }
          }
        });
        socket.addEventListener("close", (event) => {
          if (disposed || connection.intentional || active !== connection) return;
          active = undefined;
          const code = (event as CloseEvent).code;
          if (code === 4000 || code === 4004) { authoritativeResync(); return; }
          if (code === 4001 || code === 4002) { terminal(); return; }
          scheduleConnect(Math.max(reconnectDelayMS, drainUntilMS - Date.now()));
        });
      };
      connect();
      return () => {
        disposed = true;
        stop();
        if (activeCursor === cursor) activeCursor = undefined;
      };
    },
  };
}

export function newIntentID(cryptoSource: Crypto = crypto): string { return cryptoSource.randomUUID(); }
