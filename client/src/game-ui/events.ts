import { parseCanonical } from "../numeric";
import type { TransportEnvelope } from "../transport";

type RunID = Readonly<{ company_stream_id: string; run_seq: number }>;
export type PrestigeTerms = Readonly<{
  clout_reach_note: string;
  network_slot_unlocks: readonly Readonly<{ carried_ref: string; slot: string }>[];
  reputation_delta: number;
  route_knowledge: number;
}>;

export type GateCrossedEvent = Readonly<{ cursor: number; kind: "gate_crossed"; occurred_at_ms: number; payload: Readonly<{ founder_id: string; gate_id: string; route_id: string | null; run_id: RunID }> }>;
export type ExitOfferSpawnedEvent = Readonly<{ cursor: number; kind: "exit_offer_spawned"; occurred_at_ms: number; payload: Readonly<{ exit_type: string; expires_at_ms: number; offer_id: string; payout_preview: PrestigeTerms }> }>;
export type ExitOfferResolvedEvent = Readonly<{ cursor: number; kind: "exit_offer_resolved"; occurred_at_ms: number; payload: Readonly<{ offer_id: string; resolution: "accepted" | "declined" | "expired"; run_seq: number | null }> }>;
export type RunEndedEvent = Readonly<{ cursor: number; kind: "run_ended"; occurred_at_ms: number; payload: Readonly<{
  assisted: Readonly<{ advisor: boolean; commons: boolean }>;
  attended_ms: number;
  ended_at_ms: number;
  executed_routes: readonly string[];
  exit_type: string;
  faction: string | null;
  founder_id: string;
  gates_crossed: readonly string[];
  generators_purchased_total: number;
  ledger_fact_kinds: readonly string[];
  lifetime_value: string;
  payout: PrestigeTerms;
  pre_timer: boolean;
  rta_ms: number;
  run_id: RunID;
  started_at_ms: number;
  terminal_seq: number;
  tier: number;
}> }>;
export type RunEndSurfaceProps = Readonly<{ ended: RunEndedEvent }>;
export type GameUILifecycleEvent = GateCrossedEvent | ExitOfferSpawnedEvent | ExitOfferResolvedEvent | RunEndedEvent;
export type GameUISystemEvent = Readonly<{ kind: "resync_required" }> | Readonly<{ kind: "server_restarting"; resume_after_ms: number }>;

const mechanicalID = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;
const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;

function object(value: unknown, label: string): Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value)) throw new SyntaxError(`${label} must be an object`);
  return value as Record<string, unknown>;
}
function exact(value: Record<string, unknown>, keys: readonly string[], label: string): void {
  const actual = Object.keys(value).sort(), expected = [...keys].sort();
  if (actual.length !== expected.length || actual.some((key, index) => key !== expected[index])) throw new SyntaxError(`${label} fields are not exact`);
}
function safe(value: unknown, minimum = 0): number {
  if (!Number.isSafeInteger(value) || (value as number) < minimum) throw new SyntaxError("invalid event integer");
  return value as number;
}
function id(value: unknown): string {
  if (typeof value !== "string" || !mechanicalID.test(value)) throw new SyntaxError("invalid event ID");
  return value;
}
function uuidString(value: unknown): string {
  if (typeof value !== "string" || !uuid.test(value)) throw new SyntaxError("invalid event UUID");
  return value;
}
function sortedIDs(value: unknown): readonly string[] {
  if (!Array.isArray(value)) throw new SyntaxError("invalid ID set");
  const result = value.map(id);
  if (result.some((item, index) => index > 0 && result[index - 1] >= item)) throw new SyntaxError("event IDs are not sorted and unique");
  return result;
}
function runID(value: unknown): RunID {
  const row = object(value, "run ID"); exact(row, ["company_stream_id", "run_seq"], "run ID");
  return { company_stream_id: uuidString(row.company_stream_id), run_seq: safe(row.run_seq, 1) };
}
function prestigeTerms(value: unknown): PrestigeTerms {
  const row = object(value, "prestige terms"); exact(row, ["clout_reach_note", "network_slot_unlocks", "reputation_delta", "route_knowledge"], "prestige terms");
  if (typeof row.clout_reach_note !== "string" || row.clout_reach_note.length === 0 || !Array.isArray(row.network_slot_unlocks)) throw new SyntaxError("invalid prestige terms");
  const slots = row.network_slot_unlocks.map((value) => {
    const slot = object(value, "network slot"); exact(slot, ["carried_ref", "slot"], "network slot");
    return { carried_ref: id(slot.carried_ref), slot: id(slot.slot) };
  });
  if (slots.some((slot, index) => index > 0 && slots[index - 1].slot >= slot.slot)) throw new SyntaxError("network slots are not sorted");
  return { clout_reach_note: row.clout_reach_note, network_slot_unlocks: slots, reputation_delta: safe(row.reputation_delta), route_knowledge: safe(row.route_knowledge) };
}

export function decodeGameUIEvent(envelope: TransportEnvelope): GameUILifecycleEvent | undefined {
  if (envelope.kind !== "event") return undefined;
  const wrapper = envelope.payload;
  const kind = wrapper.kind;
  const payload = object(wrapper.payload, "Game UI event payload");
  const occurredAtMS = Date.parse(envelope.ts);
  if (!Number.isSafeInteger(envelope.rev) || !Number.isFinite(occurredAtMS)) throw new SyntaxError("invalid event coordinate");
  if (kind === "gate_crossed") {
    exact(payload, ["founder_id", "gate_id", "route_id", "run_id"], "gate_crossed");
    if (payload.route_id !== null && typeof payload.route_id !== "string") throw new SyntaxError("invalid route ID");
    return { cursor: envelope.rev, kind, occurred_at_ms: occurredAtMS, payload: { founder_id: uuidString(payload.founder_id), gate_id: id(payload.gate_id), route_id: payload.route_id === null ? null : id(payload.route_id), run_id: runID(payload.run_id) } };
  }
  if (kind === "exit_offer_spawned") {
    exact(payload, ["exit_type", "expires_at_ms", "offer_id", "payout_preview"], "exit_offer_spawned");
    return { cursor: envelope.rev, kind, occurred_at_ms: occurredAtMS, payload: { exit_type: id(payload.exit_type), expires_at_ms: safe(payload.expires_at_ms, 1), offer_id: uuidString(payload.offer_id), payout_preview: prestigeTerms(payload.payout_preview) } };
  }
  if (kind === "exit_offer_declined" || kind === "exit_offer_expired") {
    exact(payload, kind === "exit_offer_declined" ? ["offer_id", "run_seq"] : ["offer_id"], kind);
    return { cursor: envelope.rev, kind: "exit_offer_resolved", occurred_at_ms: occurredAtMS, payload: { offer_id: uuidString(payload.offer_id), resolution: kind === "exit_offer_declined" ? "declined" : "expired", run_seq: kind === "exit_offer_declined" ? safe(payload.run_seq, 1) : null } };
  }
  if (kind === "exit_offer_resolved") {
    exact(payload, ["offer_id", "resolution"], kind);
    if (payload.resolution !== "accepted") throw new SyntaxError("invalid accepted offer resolution");
    return { cursor: envelope.rev, kind, occurred_at_ms: occurredAtMS, payload: { offer_id: uuidString(payload.offer_id), resolution: payload.resolution, run_seq: null } };
  }
  if (kind !== "run_ended") return undefined;
  exact(payload, ["assisted", "attended_ms", "ended_at_ms", "executed_routes", "exit_type", "faction", "founder_id", "gates_crossed", "generators_purchased_total", "ledger_fact_kinds", "lifetime_value", "payout", "pre_timer", "rta_ms", "run_id", "started_at_ms", "terminal_seq", "tier"], "run_ended");
  const assisted = object(payload.assisted, "run_ended.assisted"); exact(assisted, ["advisor", "commons"], "run_ended.assisted");
  if (typeof assisted.advisor !== "boolean" || typeof assisted.commons !== "boolean" || typeof payload.pre_timer !== "boolean" ||
      payload.faction !== null && typeof payload.faction !== "string" || typeof payload.exit_type !== "string") throw new SyntaxError("invalid run-ended flags");
  const started = safe(payload.started_at_ms, 1), ended = safe(payload.ended_at_ms, started), rta = safe(payload.rta_ms);
  if (rta !== ended - started || typeof payload.lifetime_value !== "string" || parseCanonical(payload.lifetime_value).lt(0)) throw new SyntaxError("invalid run-ended timing");
  return { cursor: envelope.rev, kind, occurred_at_ms: occurredAtMS, payload: {
    assisted: { advisor: assisted.advisor, commons: assisted.commons }, attended_ms: safe(payload.attended_ms), ended_at_ms: ended,
    executed_routes: sortedIDs(payload.executed_routes), exit_type: id(payload.exit_type), faction: payload.faction === null ? null : id(payload.faction),
    founder_id: uuidString(payload.founder_id), gates_crossed: sortedIDs(payload.gates_crossed), generators_purchased_total: safe(payload.generators_purchased_total),
    ledger_fact_kinds: sortedIDs(payload.ledger_fact_kinds), lifetime_value: payload.lifetime_value, payout: prestigeTerms(payload.payout),
    pre_timer: payload.pre_timer, rta_ms: rta, run_id: runID(payload.run_id), started_at_ms: started, terminal_seq: safe(payload.terminal_seq, 1), tier: safe(payload.tier),
  } };
}

export function decodeGameUISystemEvent(envelope: TransportEnvelope): GameUISystemEvent | undefined {
  if (envelope.kind !== "system") return undefined;
  if (envelope.payload.code === "resync_required") return { kind: "resync_required" };
  if (envelope.payload.code === "server_restarting") return { kind: "server_restarting", resume_after_ms: safe(envelope.payload.resume_after_ms) };
  return undefined;
}
