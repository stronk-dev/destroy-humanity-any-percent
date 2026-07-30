export type TransportKind = "receipt" | "snapshot" | "event" | "presence" | "system";
export interface TransportEnvelope {
  readonly v: 1;
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
  if (root.v !== 1 || typeof root.ch !== "string" || root.ch.length === 0 || !Number.isSafeInteger(root.rev) || (root.rev as number) < 0 ||
      typeof root.constants_hash !== "string" || !hashPattern.test(root.constants_hash) || typeof root.ts !== "string" || !Number.isFinite(Date.parse(root.ts))) {
    throw new SyntaxError("invalid transport envelope");
  }
  const payload = object(root.payload, `${kind} payload`);
  validatePayload(kind as TransportKind, payload);
  return root as unknown as TransportEnvelope;
}

function validatePayload(kind: TransportKind, payload: Record<string, unknown>): void {
  if (kind === "receipt") return; // Production C1 owns and validates this object unchanged.
  if (kind === "snapshot") {
    exact(payload, ["scope", "rev", "state"], "snapshot payload");
    if (!["company", "world", "guild", "cohort"].includes(String(payload.scope)) || !Number.isSafeInteger(payload.rev) || (payload.rev as number) < 0 || typeof payload.state !== "object" || payload.state === null) throw new SyntaxError("invalid snapshot payload");
    return;
  }
  if (kind === "event") {
    exact(payload, ["event_id", "kind", "rev", "payload"], "event payload");
    if (typeof payload.event_id !== "string" || typeof payload.kind !== "string" || !idPattern.test(payload.kind) || !Number.isSafeInteger(payload.rev) || (payload.rev as number) < 1 || typeof payload.payload !== "object" || payload.payload === null) throw new SyntaxError("invalid event payload");
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
  return Array.isArray(value) && value.every((item) => typeof item === "string");
}
