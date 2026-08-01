export type TransportKind = "receipt" | "snapshot" | "event" | "presence" | "system";
export interface TransportEnvelope {
  readonly v: 2;
  readonly ch: string;
  readonly kind: TransportKind;
  readonly rev: number;
  readonly constants_hash: string;
  readonly ts: string;
  readonly payload: Readonly<Record<string, unknown>>;
}

const hashPattern = /^sha256:[0-9a-f]{64}$/;
const idPattern = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;
const kinds = new Set<TransportKind>(["receipt", "snapshot", "event", "presence", "system"]);

export function decodeTransportEnvelope(source: unknown): TransportEnvelope | undefined {
  const root = object(source, "transport envelope");
  const kind = root.kind;
  if (typeof kind !== "string" || !kinds.has(kind as TransportKind)) return undefined;
  exact(root, ["v", "ch", "kind", "rev", "constants_hash", "ts", "payload"], "transport envelope");
  if (root.v !== 2 || typeof root.ch !== "string" || root.ch.length === 0 || !Number.isSafeInteger(root.rev) || (root.rev as number) < 0 ||
      typeof root.constants_hash !== "string" || !hashPattern.test(root.constants_hash) || typeof root.ts !== "string" || !Number.isFinite(Date.parse(root.ts)) ||
      !channelAllowsKind(root.ch, kind as TransportKind)) {
    throw new SyntaxError("invalid transport envelope");
  }
  const payload = object(root.payload, `${kind} payload`);
  validatePayload(kind as TransportKind, payload, root.ch, root.rev as number);
  return root as unknown as TransportEnvelope;
}

function validatePayload(kind: TransportKind, payload: Record<string, unknown>, channel: string, envelopeRevision: number): void {
  if (kind === "receipt") return; // Production C1 owns and validates this object unchanged.
  if (kind === "snapshot") {
    exact(payload, ["scope", "rev", "state"], "snapshot payload");
    if (!["company", "world", "guild", "cohort"].includes(String(payload.scope)) || !Number.isSafeInteger(payload.rev) || (payload.rev as number) < 0 ||
        payload.rev !== envelopeRevision || typeof payload.state !== "object" || payload.state === null || Array.isArray(payload.state) ||
        !scopeMatchesChannel(String(payload.scope), channel)) throw new SyntaxError("invalid snapshot payload");
    return;
  }
  if (kind === "event") {
    exact(payload, ["event_id", "kind", "scope", "rev", "cursor_effect", "payload"], "event payload");
    if (typeof payload.event_id !== "string" || payload.event_id.length === 0 || typeof payload.kind !== "string" || !idPattern.test(payload.kind) ||
        !["company", "founder"].includes(String(payload.scope)) ||
        !Number.isSafeInteger(payload.rev) || (payload.rev as number) < 1 || payload.rev !== envelopeRevision ||
        !["advance", "historical"].includes(String(payload.cursor_effect)) ||
        (payload.kind === "compensation") !== (payload.cursor_effect === "historical") ||
        typeof payload.payload !== "object" || payload.payload === null || Array.isArray(payload.payload)) throw new SyntaxError("invalid event payload");
    return;
  }
  if (kind === "presence") {
    exact(payload, ["joined", "left", "count"], "presence payload");
    if (!stringArray(payload.joined) || !stringArray(payload.left) || !Number.isSafeInteger(payload.count) || (payload.count as number) < 0) throw new SyntaxError("invalid presence payload");
    return;
  }
  exact(payload, payload.code === "server_restarting" ? ["code", "resume_after_ms"] : ["code"], "system payload");
  if (!["server_restarting", "history_expired", "resync_required"].includes(String(payload.code)) || payload.code === "server_restarting" && (!Number.isSafeInteger(payload.resume_after_ms) || (payload.resume_after_ms as number) < 0)) throw new SyntaxError("invalid system payload");
}

function object(source: unknown, name: string): Record<string, unknown> {
  if (typeof source !== "object" || source === null || Array.isArray(source)) throw new SyntaxError(`${name} must be an object`);
  return source as Record<string, unknown>;
}

function exact(source: object, fields: readonly string[], name: string): void {
  const actual = Object.keys(source).sort(); const expected = [...fields].sort();
  if (actual.length !== expected.length || actual.some((field, index) => field !== expected[index])) throw new SyntaxError(`${name} fields are not exact`);
}

function stringArray(value: unknown): value is string[] {
  return Array.isArray(value) && value.every((item) => typeof item === "string" && item.length > 0);
}

function channelAllowsKind(channel: string, kind: TransportKind): boolean {
  if (channel.startsWith("player:")) return validChannelID(channel.slice("player:".length));
  if (channel === "world") return kind === "snapshot" || kind === "presence" || kind === "system";
  if (channel === "feed") return kind === "event" || kind === "presence" || kind === "system";
  for (const prefix of ["guild:", "cohort:", "match:"]) {
    if (channel.startsWith(prefix)) return validChannelID(channel.slice(prefix.length)) && kind !== "receipt";
  }
  return false;
}

function validChannelID(value: string): boolean { return value.length > 0 && !value.includes(":"); }

function scopeMatchesChannel(scope: string, channel: string): boolean {
  if (scope === "company") return channel.startsWith("player:");
  if (scope === "world") return channel === "world";
  if (scope === "guild") return channel.startsWith("guild:");
  if (scope === "cohort") return channel.startsWith("cohort:");
  return false;
}

export type CursorResult = "deliver" | "duplicate" | "resync_required";

interface ScopeCursor { revision: number; seenAtRevision: Set<string> }

// PlayerRevisionCursor is the client-shell gap authority for private events.
// Historical compensation is visible audit output but never rewinds or advances
// authoritative state. Forward events may share a revision, so event IDs dedupe
// within the current revision instead of treating the second event as stale.
export class PlayerRevisionCursor {
  readonly #scopes = new Map<"company" | "founder", ScopeCursor>();

  reset(scope: "company" | "founder", revision: number): void {
    if (!Number.isSafeInteger(revision) || revision < 0) throw new RangeError("invalid stream revision");
    this.#scopes.set(scope, { revision, seenAtRevision: new Set() });
  }

  event(envelope: TransportEnvelope): CursorResult {
    if (envelope.kind !== "event") throw new TypeError("cursor accepts event envelopes only");
    const payload = envelope.payload;
    const scope = payload.scope as "company" | "founder";
    const effect = payload.cursor_effect as "advance" | "historical";
    const eventID = payload.event_id as string;
    if (effect === "historical") return "deliver";
    const current = this.#scopes.get(scope) ?? { revision: 0, seenAtRevision: new Set<string>() };
    if (envelope.rev < current.revision) return "duplicate";
    if (envelope.rev > current.revision + 1) return "resync_required";
    if (envelope.rev === current.revision + 1) {
      current.revision = envelope.rev;
      current.seenAtRevision.clear();
    }
    if (current.seenAtRevision.has(eventID)) return "duplicate";
    current.seenAtRevision.add(eventID);
    this.#scopes.set(scope, current);
    return "deliver";
  }

  revision(scope: "company" | "founder"): number { return this.#scopes.get(scope)?.revision ?? 0; }
}
