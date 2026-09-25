import type { MinigameCommandRequest, MinigameCurrentResponse, MinigameSessionDescriptor, MinigameSessionResponse, MinigameSessionResponseTerminal, PitchSnapshot } from "../../api/generated/types";
import type { CopyKey } from "../../copy";
import { MinigameAPIError, MinigameTransportError } from "./session-port";

// The closed MA-C9 surface states. `none` is not a surface state: a
// {kind:"none"} current answer leaves the surface `active`-less and the host
// renders the launcher (GS7/OD-17 alternative that keeps the ruled five).
export type MinigameSurfaceState =
  | Readonly<{ kind: "loading" }>
  | Readonly<{ kind: "active"; session: MinigameSessionDescriptor; snapshot: PitchSnapshot; inFlight: PendingCommand | null }>
  | Readonly<{ kind: "paused_reconnect"; last: Readonly<{ session: MinigameSessionDescriptor; snapshot: PitchSnapshot }> | null; retry: PendingCommand | null }>
  | Readonly<{ kind: "required_terminal"; response: MinigameSessionResponseTerminal }>
  | Readonly<{ kind: "error"; message: CopyKey }>;

export type MinigameLauncherState = Readonly<{ kind: "launcher"; notice: CopyKey | null }>;
export type MinigameHostState = MinigameSurfaceState | MinigameLauncherState;

export interface PendingCommand { readonly commandID: string; readonly sessionID: string; readonly request: MinigameCommandRequest }

// What the surface does after a typed rejection. `refetch` asks current();
// `stay` keeps the table and shows the notice; `launcher` returns to the
// launcher with the notice; `pause` enters paused_reconnect; `error` is terminal.
export type RejectionEffect = "error" | "launcher" | "pause" | "refetch" | "stay";

export interface RejectionRow { readonly effect: RejectionEffect; readonly notice: CopyKey | null; readonly clearSelection: boolean }

const row = (effect: RejectionEffect, notice: CopyKey | null, clearSelection = false): RejectionRow => Object.freeze({ effect, notice, clearSelection });

// Exact server pairs from minigameErrorJSON. An unlisted pair is an error.
export const MINIGAME_REJECTIONS: ReadonlyMap<string, RejectionRow> = new Map([
  ["409 not_eligible/fiscal_unlock_required", row("launcher", "minigame.rejection.fiscal_unlock_required")],
  ["409 not_eligible/human_content_locked", row("launcher", "minigame.rejection.human_content_locked")],
  ["409 not_eligible/exclusive_activity", row("launcher", "minigame.rejection.exclusive_activity")],
  ["409 not_eligible/illegal_phase", row("refetch", "pitch.rejection.illegal_phase", true)],
  ["409 not_eligible/hand_too_large", row("stay", "pitch.rejection.hand_too_large")],
  ["409 not_eligible/duplicate_card", row("stay", "pitch.rejection.bad_selection", true)],
  ["409 not_eligible/unknown_card", row("stay", "pitch.rejection.bad_selection", true)],
  ["409 not_eligible/insufficient_currency", row("stay", "pitch.rejection.insufficient_currency")],
  ["409 not_eligible/hack_slots_full", row("stay", "pitch.rejection.hack_slots_full")],
  ["409 not_eligible/unknown_offer", row("refetch", "pitch.rejection.unknown_offer")],
  ["409 conflict/minigame_revision", row("refetch", "minigame.conflict.revision", true)],
  ["409 conflict/minigame_session", row("pause", null)],
  ["409 idempotency_conflict/minigame_command", row("error", "minigame.error.generic")],
  ["409 idempotency_conflict/minigame_session", row("error", "minigame.error.generic")],
  ["404 unknown_id/minigame_session", row("refetch", null, true)],
  ["404 unknown_id/minigame_tenant", row("error", "minigame.error.generic")],
  ["404 unknown_id/founder", row("error", "minigame.error.generic")],
  ["429 rate_limited/ip", row("stay", "minigame.rate_limited")],
  ["429 rate_limited/account", row("stay", "minigame.rate_limited")],
  ["503 not_configured/minigame_api", row("error", "minigame.error.unavailable")],
]);

export function rejectionFor(error: unknown): RejectionRow {
  if (error instanceof MinigameTransportError) return row("pause", null);
  if (!(error instanceof MinigameAPIError)) return row("error", "minigame.error.generic");
  if (error.category === "internal_invariant") return row("error", "minigame.error.unavailable");
  return MINIGAME_REJECTIONS.get(`${error.status} ${error.category}/${error.detail}`) ?? row("error", "minigame.error.generic");
}

export function stateFromCurrent(value: MinigameCurrentResponse): MinigameHostState {
  if (value.kind === "none") return { kind: "launcher", notice: null };
  if (value.session.status === "active") return { kind: "active", session: value.session, snapshot: value.snapshot, inFlight: null };
  // claimed: another request holds the session; wait and re-read.
  return { kind: "paused_reconnect", last: { session: value.session, snapshot: value.snapshot }, retry: null };
}

export function stateFromSession(value: MinigameSessionResponse): MinigameSurfaceState {
  if (value.status === "resolved") return { kind: "required_terminal", response: value };
  const { snapshot, ...session } = value;
  return { kind: "active", session: { ...session, status: "active" }, snapshot, inFlight: null };
}

export function commandRequest(state: Extract<MinigameSurfaceState, { kind: "active" }>, commandID: string, command: MinigameCommandRequest["command"]): PendingCommand {
  return { commandID, sessionID: state.session.session_id, request: { command_id: commandID, expected_revision: state.session.revision, command } };
}
