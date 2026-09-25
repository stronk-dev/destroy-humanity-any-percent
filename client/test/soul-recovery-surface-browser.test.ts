import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import type { SoulRecoveryProgressResponse, SoulRecoveryStartResponse, SoulRecoveryTerminalResponse } from "../src/api/generated/types";
import { MinigameAPIError, MinigameTransportError } from "../src/game-ui/minigame/session-port";
import type { RecoveryVisibility } from "../src/soul/recovery-scheduler";
import { loadSoulRecoveryContent, type SoulRecoveryContent, type SoulRecoveryPort } from "../src/game-ui/soul/recovery-surface";
import SoulRecoverySurface from "../src/game-ui/soul/SoulRecoverySurface.svelte";
import { installTheme, UI_THEMES } from "../src/ui/themes";

const browser = typeof document !== "undefined";
const SESSION = "01986666-0000-7000-8000-00000000000a";
const OTHER_SESSION = "01986666-0000-7000-8000-00000000000b";
const TOKEN_1 = "01986666-0000-4000-8000-000000000001";
const TOKEN_2 = "01986666-0000-4000-8000-000000000002";
const REQUIRED = 900_000;

function content(): SoulRecoveryContent { return { ...loadSoulRecoveryContent(), beatIntervalMS: 25 }; }

function started(sessionID = SESSION, token = TOKEN_1, attended = 0): SoulRecoveryStartResponse {
  return { activity_id: "defrag", attended_progress_ms: attended, last_progress_server_ms: 1, progress_token: token, required_duration_attended_ms: REQUIRED, session_id: sessionID, started_wall_ms: 1 };
}
function progressed(attended: number, sessionID = SESSION): SoulRecoveryProgressResponse {
  return { attended_progress_ms: attended, eligible: attended >= REQUIRED, last_progress_server_ms: 2, required_duration_attended_ms: REQUIRED, session_id: sessionID };
}
function terminal(action: "cancel" | "resolve", cancelledBy?: "player" | "watchdog"): SoulRecoveryTerminalResponse {
  return { action, activity_id: "defrag", band_after: "whole", band_before: "dimming", ...(cancelledBy ? { cancelled_by: cancelledBy } : {}), company_revision: 3, founder_revision: 4,
    intent_id: "01986666-0000-7000-8000-0000000000ff", outcome: "applied", session_id: SESSION, soul_after: action === "resolve" ? 52 : 40, soul_before: 40 };
}

class FakeVisibility implements RecoveryVisibility {
  listener: ((visible: boolean) => void) | undefined;
  subscribe(callback: (visible: boolean) => void): () => void { this.listener = callback; return () => { this.listener = undefined; }; }
  emit(visible: boolean): void { this.listener?.(visible); }
}

class FakePort implements SoulRecoveryPort {
  starts: string[] = [];
  tokens: string[] = [];
  steady: SoulRecoveryProgressResponse | undefined;
  startReplies: (() => SoulRecoveryStartResponse)[] = [];
  progressReplies: (() => SoulRecoveryProgressResponse)[] = [];
  terminalReply: () => SoulRecoveryTerminalResponse = () => terminal("resolve");
  async start(activityID: string): Promise<SoulRecoveryStartResponse> { this.starts.push(activityID); const next = this.startReplies.shift(); if (!next) throw new Error("unexpected start"); return next(); }
  async progress(_sessionID: string, token: string): Promise<SoulRecoveryProgressResponse> {
    this.tokens.push(token);
    const next = this.progressReplies.shift();
    // Unscripted beats repeat the last authoritative answer.
    if (!next) { if (!this.steady) throw new Error("unexpected progress"); return this.steady; }
    const value = next();
    this.steady = value;
    return value;
  }
  async resolve(): Promise<SoulRecoveryTerminalResponse> { return this.terminalReply(); }
  async cancel(): Promise<SoulRecoveryTerminalResponse> { return this.terminalReply(); }
}

async function settle(ms = 0): Promise<void> {
  if (ms > 0) await new Promise((resolve) => setTimeout(resolve, ms));
  for (let index = 0; index < 5; index += 1) { await new Promise((resolve) => setTimeout(resolve, 0)); await tick(); flushSync(); }
}

async function assertAxe(target: HTMLElement, label: string): Promise<void> {
  const result = await axe.run(target, { runOnly: { type: "tag", values: ["wcag2a", "wcag2aa", "wcag21aa", "wcag22aa"] } });
  expect(result.violations.filter((violation) => violation.impact === "serious" || violation.impact === "critical"), label).toEqual([]);
}

function host(): HTMLElement {
  const target = document.createElement("main");
  document.body.append(target);
  installTheme(target, UI_THEMES.era_1995, false);
  return target;
}

function button(target: HTMLElement, text: string): HTMLButtonElement {
  const found = [...target.querySelectorAll("button")].find((candidate) => candidate.textContent?.trim() === text);
  if (!found) throw new Error(`missing button ${text}: ${target.textContent}`);
  return found;
}

function mountSurface(port: FakePort, visibility: FakeVisibility, clock: { now: number }, terminals: SoulRecoveryTerminalResponse[] = []) {
  const target = host();
  const app = mount(SoulRecoverySurface, { target, props: { port, content: content(), era: "era_1995", visibility, now: () => clock.now, toySeed: 7,
    onExitToHost: () => {}, onTerminal: (response: SoulRecoveryTerminalResponse) => terminals.push(response) } });
  return { target, app };
}

it.skipIf(!browser)("runs pick → beats → hidden pause → network pause → reconnect rotation → finish through the port only", async () => {
  const port = new FakePort();
  const visibility = new FakeVisibility();
  const clock = { now: 1_000 };
  const terminals: SoulRecoveryTerminalResponse[] = [];
  const { target, app } = mountSurface(port, visibility, clock, terminals);
  try {
    await settle();
    await assertAxe(target, "picker");
    expect(target.querySelectorAll("li button")).toHaveLength(3);
    port.startReplies.push(() => started());
    port.progressReplies.push(() => progressed(450_000));
    target.querySelector<HTMLButtonElement>("li button")!.click();
    await settle(80);
    expect(port.starts).toEqual(["defrag"]);
    expect(port.tokens[0]).toBe(TOKEN_1);
    expect(target.querySelector("progress")!.value).toBe(450_000);
    await assertAxe(target, "active");

    visibility.emit(false);
    await settle();
    expect(target.textContent).toContain("Paused while this tab is in the background.");
    visibility.emit(true);
    await settle();
    expect(target.textContent).not.toContain("Paused while this tab is in the background.");

    port.progressReplies.push(() => { throw new MinigameTransportError("offline"); });
    await settle(80);
    expect(target.querySelector("[role=alert]")?.textContent).toContain("Connection lost");
    await assertAxe(target, "paused network");
    const beatsBeforeReconnect = port.tokens.length;
    await settle(120);
    // No beat is sent while reconnect is required, and none is replayed.
    expect(port.tokens.length).toBe(beatsBeforeReconnect);

    port.startReplies.push(() => started(SESSION, TOKEN_2, 450_000));
    port.progressReplies.push(() => progressed(REQUIRED));
    button(target, "Reconnect").click();
    await settle(80);
    expect(port.starts).toEqual(["defrag", "defrag"]);
    expect(port.tokens.at(-1)).toBe(TOKEN_2);
    button(target, "Finish").click();
    await settle();
    expect(terminals).toEqual([terminal("resolve")]);
    expect(target.textContent).toContain("Soul went from 40 to 52");
    expect(document.activeElement?.id).toBe("recovery-heading");
    await assertAxe(target, "terminal");
  } finally { unmount(app); target.remove(); }
});

it.skipIf(!browser)("treats an elapsed gap over the missed ceiling as a network pause on becoming visible", async () => {
  const port = new FakePort();
  const visibility = new FakeVisibility();
  const clock = { now: 1_000 };
  const { target, app } = mountSurface(port, visibility, clock);
  try {
    await settle();
    port.startReplies.push(() => started());
    port.progressReplies.push(() => progressed(1_000));
    target.querySelector<HTMLButtonElement>("li button")!.click();
    await settle();
    visibility.emit(false);
    clock.now += 10_000;
    visibility.emit(true);
    await settle();
    expect(target.querySelector("[role=alert]")?.textContent).toContain("Connection lost");
  } finally { unmount(app); target.remove(); }
});

it.skipIf(!browser)("ends without a reconnect-start when the server says the session is gone", async () => {
  const port = new FakePort();
  const visibility = new FakeVisibility();
  const { target, app } = mountSurface(port, visibility, { now: 1_000 });
  try {
    await settle();
    port.startReplies.push(() => started());
    port.progressReplies.push(() => { throw new MinigameAPIError(404, "unknown_id", "recovery_session"); });
    target.querySelector<HTMLButtonElement>("li button")!.click();
    await settle(80);
    expect(target.textContent).toContain("This session has already ended.");
    expect(port.starts).toEqual(["defrag"]);
    await assertAxe(target, "ended");
  } finally { unmount(app); target.remove(); }
});

it.skipIf(!browser)("renders a watchdog cancellation honestly and a not-ready finish as a notice", async () => {
  const port = new FakePort();
  const visibility = new FakeVisibility();
  const { target, app } = mountSurface(port, visibility, { now: 1_000 });
  try {
    await settle();
    port.startReplies.push(() => started(OTHER_SESSION, TOKEN_1, REQUIRED));
    target.querySelector<HTMLButtonElement>("li button")!.click();
    await settle();
    port.terminalReply = () => { throw new MinigameAPIError(400, "not_eligible", "soul_recovery_not_ready"); };
    button(target, "Finish").click();
    await settle();
    expect(target.querySelector("[role=status]")?.textContent).toContain("Not quite done yet.");
    port.terminalReply = () => terminal("cancel", "watchdog");
    button(target, "Stop early").click();
    await settle();
    expect(target.textContent).toContain("ran too long");
  } finally { unmount(app); target.remove(); }
});

it.skipIf(!browser)("keeps a required reconnect visible across a background/foreground cycle", async () => {
  const port = new FakePort();
  const visibility = new FakeVisibility();
  const { target, app } = mountSurface(port, visibility, { now: 1_000 });
  try {
    await settle();
    port.startReplies.push(() => started());
    port.progressReplies.push(() => { throw new MinigameTransportError("offline"); });
    target.querySelector<HTMLButtonElement>("li button")!.click();
    await settle(80);
    expect(target.querySelector("[role=alert]")?.textContent).toContain("Connection lost");
    visibility.emit(false);
    await settle();
    visibility.emit(true);
    await settle();
    expect(target.querySelector("[role=alert]")?.textContent).toContain("Connection lost");
    button(target, "Reconnect");
  } finally { unmount(app); target.remove(); }
});
