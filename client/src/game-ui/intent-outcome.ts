import type { CopyKey } from "../copy";

// GS0.2: the typed result of POST /api/v1/intents. A rejected intent is an
// HTTP 200 whose body names the rejection; it is never an offline condition.
export type IntentOutcome =
  | Readonly<{ outcome: "applied"; receipt: Readonly<Record<string, unknown>> }>
  | Readonly<{ outcome: "rejected"; category: string; detail: string; currentRevision: number; sessionExpired: boolean }>;

// A non-2xx intent response with an exact {category, detail} API error body.
export class GameUIRequestError extends Error {
  constructor(readonly status: number, readonly category: string, readonly detail: string) { super(`${status} ${category}/${detail}`); }
}

const mechanical = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/u;

function exactKeys(value: Readonly<Record<string, unknown>>, required: readonly string[], optional: readonly string[] = []): boolean {
  const keys = Object.keys(value);
  return required.every((key) => keys.includes(key)) && keys.every((key) => required.includes(key) || optional.includes(key));
}

export function parseIntentOutcome(value: unknown): IntentOutcome {
  if (value === null || typeof value !== "object" || Array.isArray(value)) throw new SyntaxError("intent receipt must be an object");
  const row = value as Readonly<Record<string, unknown>>;
  if (row.outcome === "applied") return Object.freeze({ outcome: "applied", receipt: row });
  if (row.outcome !== "rejected" || !exactKeys(row, ["current_revision", "intent_id", "outcome", "rejection"]) ||
      !Number.isSafeInteger(row.current_revision) || (row.current_revision as number) < 0) throw new SyntaxError("intent receipt outcome is invalid");
  const rejection = row.rejection;
  if (rejection === null || typeof rejection !== "object" || Array.isArray(rejection)) throw new SyntaxError("intent rejection must be an object");
  const pair = rejection as Readonly<Record<string, unknown>>;
  if (!exactKeys(pair, ["category", "detail"], ["session_expired"]) || typeof pair.category !== "string" || !mechanical.test(pair.category) ||
      typeof pair.detail !== "string" || pair.detail === "" || pair.session_expired !== undefined && pair.session_expired !== true) {
    throw new SyntaxError("intent rejection must be exactly {category, detail}");
  }
  return Object.freeze({ outcome: "rejected", category: pair.category, detail: pair.detail, currentRevision: row.current_revision as number, sessionExpired: pair.session_expired === true });
}

export function parseIntentErrorBody(status: number, value: unknown): GameUIRequestError | undefined {
  if (value === null || typeof value !== "object" || Array.isArray(value)) return undefined;
  const row = value as Readonly<Record<string, unknown>>;
  if (!exactKeys(row, ["category", "detail"]) || typeof row.category !== "string" || !mechanical.test(row.category) || typeof row.detail !== "string" || row.detail === "") return undefined;
  return new GameUIRequestError(status, row.category, row.detail);
}

// What the host does with a finished intent.
export type IntentEffect = "none" | "refresh" | "offline";
export interface IntentNotice { readonly effect: IntentEffect; readonly notice: CopyKey | null; readonly invariant: boolean }

// Surface-specific exact pairs, checked before the shared rows.
export type SurfaceRejections = ReadonlyMap<string, CopyKey>;

export function noticeForOutcome(outcome: IntentOutcome, surface: SurfaceRejections = new Map()): IntentNotice {
  if (outcome.outcome === "applied") return { effect: "none", notice: null, invariant: false };
  const pair = `${outcome.category}/${outcome.detail}`;
  const exact = surface.get(pair) ?? surface.get(`${outcome.category}/*`);
  if (exact) return { effect: "none", notice: exact, invariant: false };
  switch (outcome.category) {
    case "revision_conflict": return { effect: "refresh", notice: "intent.conflict", invariant: false };
    case "unaffordable": return { effect: "none", notice: "intent.rejection.unaffordable", invariant: false };
    case "cap_exceeded": return { effect: "none", notice: "intent.rejection.cap_exceeded", invariant: false };
  }
  if (pair === "not_eligible/exclusive_activity") return { effect: "none", notice: "intent.rejection.exclusive_activity", invariant: false };
  if (pair === "not_eligible/minigame_session_active") return { effect: "none", notice: "intent.rejection.minigame_session_active", invariant: false };
  return { effect: "none", notice: "intent.rejection.unknown", invariant: true };
}

export function noticeForError(error: unknown): IntentNotice {
  if (!(error instanceof GameUIRequestError)) return { effect: "offline", notice: null, invariant: false };
  if (error.status === 409) return { effect: "refresh", notice: "intent.conflict", invariant: false };
  if (error.status === 429) return { effect: "none", notice: "intent.rate_limited", invariant: false };
  if (error.status === 400) return { effect: "none", notice: "intent.rejection.unknown", invariant: true };
  return { effect: "offline", notice: null, invariant: false };
}
