import pinnedPrestigeBytes from "../../../../balance/prestige/phase0.json?raw";
import pinnedSoulBytes from "../../../../balance/soul/first-content.json?raw";

import { operations, type SoulRecoveryProgressResponse, type SoulRecoveryStartResponse, type SoulRecoveryTerminalResponse } from "../../api/generated/types";
import { COPY_KEYS, type CopyKey } from "../../copy";
import { createOperationCall, MinigameAPIError, MinigameTransportError, type Fetcher } from "../minigame/session-port";
import { parsePrestigePolicy } from "../../prestige";
import { parseSoulCatalog, type SoulCatalog, type SoulRecoveryActivity } from "../../soul/catalog";

// SoulRecoveryPort is the recovery surface's only transport seam (SR-C3,
// MA-C9): the generated MA2 operations typed by the generated DTOs.
export interface SoulRecoveryPort {
  start(activityID: string): Promise<SoulRecoveryStartResponse>;
  progress(sessionID: string, progressToken: string): Promise<SoulRecoveryProgressResponse>;
  resolve(sessionID: string): Promise<SoulRecoveryTerminalResponse>;
  cancel(sessionID: string): Promise<SoulRecoveryTerminalResponse>;
}

export function createBrowserSoulRecoveryPort(accessToken: () => string, fetcher: Fetcher = fetch): SoulRecoveryPort {
  const call = createOperationCall(accessToken, fetcher);
  return {
    start: (activityID) => call(operations.start_soul_recovery.method, operations.start_soul_recovery.path, { activity_id: activityID }),
    progress: (sessionID, progressToken) => call(operations.progress_soul_recovery.method, operations.progress_soul_recovery.path, { progress_token: progressToken, session_id: sessionID }),
    resolve: (sessionID) => call(operations.resolve_soul_recovery.method, operations.resolve_soul_recovery.path, { session_id: sessionID }),
    cancel: (sessionID) => call(operations.cancel_soul_recovery.method, operations.cancel_soul_recovery.path, { session_id: sessionID }),
  };
}

export interface SoulRecoveryContent {
  readonly catalog: SoulCatalog;
  // SR-C6 ruling: client cadence is recovery_beat_ceiling_ms / 3.
  readonly beatIntervalMS: number;
}

let pinnedContent: SoulRecoveryContent | undefined;

export function loadSoulRecoveryContent(soulBytes: string = pinnedSoulBytes, prestigeBytes: string = pinnedPrestigeBytes): SoulRecoveryContent {
  if (soulBytes === pinnedSoulBytes && prestigeBytes === pinnedPrestigeBytes) return pinnedContent ??= parseSoulRecoveryContent(soulBytes, prestigeBytes);
  return parseSoulRecoveryContent(soulBytes, prestigeBytes);
}

function parseSoulRecoveryContent(soulBytes: string, prestigeBytes: string): SoulRecoveryContent {
  const prestige = parsePrestigePolicy(JSON.parse(prestigeBytes));
  const catalog = parseSoulCatalog(JSON.parse(soulBytes), { copyKeys: new Set(COPY_KEYS), epochSeeded: true, catchupCeilingMs: prestige.catchupCeilingMs });
  return { catalog, beatIntervalMS: Math.max(1, Math.floor(catalog.policy.recovery_beat_ceiling_ms / 3)) };
}

export function recoveryActivity(content: SoulRecoveryContent, activityID: string): SoulRecoveryActivity {
  const row = content.catalog.recovery_activities.find((candidate) => candidate.activity_id === activityID);
  if (!row) throw new RangeError(`unknown Soul recovery activity ${activityID}`);
  return row;
}

// What a failed recovery request does to the surface.
export type RecoveryRejection =
  | Readonly<{ effect: "reconnect" }>
  | Readonly<{ effect: "gone" }>
  | Readonly<{ effect: "notice"; notice: CopyKey }>
  | Readonly<{ effect: "error"; notice: CopyKey }>;

// Exact server pairs from account/api.go's recovery handlers.
export const RECOVERY_REJECTIONS: ReadonlyMap<string, RecoveryRejection> = new Map<string, RecoveryRejection>([
  ["400 not_eligible/recovery_token", { effect: "reconnect" }],
  ["400 not_eligible/soul_recovery_not_ready", { effect: "notice", notice: "soul.recovery_surface.not_ready" }],
  ["404 unknown_id/recovery_session", { effect: "gone" }],
  ["409 not_eligible/exclusive_activity", { effect: "notice", notice: "minigame.rejection.exclusive_activity" }],
  ["409 conflict/recovery_session", { effect: "reconnect" }],
  ["409 idempotency_conflict/recovery_session", { effect: "error", notice: "minigame.error.generic" }],
  ["429 rate_limited/recovery_progress", { effect: "notice", notice: "minigame.rate_limited" }],
  ["503 not_configured/soul_recovery", { effect: "error", notice: "minigame.error.unavailable" }],
]);

export function recoveryRejectionFor(error: unknown): RecoveryRejection {
  if (error instanceof MinigameTransportError) return { effect: "reconnect" };
  if (!(error instanceof MinigameAPIError)) return { effect: "error", notice: "minigame.error.generic" };
  if (error.category === "internal_invariant") return { effect: "error", notice: "minigame.error.unavailable" };
  return RECOVERY_REJECTIONS.get(`${error.status} ${error.category}/${error.detail}`) ?? { effect: "error", notice: "minigame.error.generic" };
}

// A client-local, non-authoritative toy seed generated on mount (MA-C9): it
// only arranges decorative toy cells and never reaches the server.
export function localToySeed(cryptoSource: Crypto = crypto): number {
  const value = new Uint32Array(1);
  cryptoSource.getRandomValues(value);
  return value[0]!;
}

// Deterministic decorative cell order for a toy seed (xorshift32 Fisher-Yates).
export function toyCellOrder(seed: number, cells: number): readonly number[] {
  if (!Number.isSafeInteger(cells) || cells < 1) throw new RangeError("invalid toy cell count");
  let state = (seed >>> 0) || 0x9e3779b9;
  const next = (): number => { state ^= state << 13; state >>>= 0; state ^= state >>> 17; state ^= state << 5; state >>>= 0; return state; };
  const order = Array.from({ length: cells }, (_, index) => index);
  for (let index = cells - 1; index > 0; index -= 1) {
    const swap = next() % (index + 1);
    [order[index], order[swap]] = [order[swap]!, order[index]!];
  }
  return order;
}
