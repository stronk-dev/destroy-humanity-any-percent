import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import type { SoulRecoveryProgressResponse, SoulRecoveryStartResponse, SoulRecoveryTerminalResponse } from "../src/api/generated/types";
import { t, type CopyKey } from "../src/copy";
import { MinigameAPIError, MinigameTransportError } from "../src/game-ui/minigame/session-port";
import type { RecoveryVisibility } from "../src/soul/recovery-scheduler";
import { createBrowserSoulRecoveryPort, loadSoulRecoveryContent, type SoulRecoveryContent, type SoulRecoveryPort } from "../src/game-ui/soul/recovery-surface";
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
  startReplies: (() => SoulRecoveryStartResponse | Promise<SoulRecoveryStartResponse>)[] = [];
  progressReplies: (() => SoulRecoveryProgressResponse)[] = [];
  terminalCalls: ("resolve" | "cancel")[] = [];
  terminalReply: () => SoulRecoveryTerminalResponse | Promise<SoulRecoveryTerminalResponse> = () => terminal("resolve");
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
  async resolve(): Promise<SoulRecoveryTerminalResponse> { this.terminalCalls.push("resolve"); return this.terminalReply(); }
  async cancel(): Promise<SoulRecoveryTerminalResponse> { this.terminalCalls.push("cancel"); return this.terminalReply(); }
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

function mountSurface(port: FakePort, visibility: FakeVisibility, clock: { now: number }, terminals: SoulRecoveryTerminalResponse[] = [], onExitToHost = () => {}) {
  const target = host();
  const app = mount(SoulRecoverySurface, { target, props: { port, content: content(), era: "era_1995", visibility, now: () => clock.now, toySeed: 7,
    onExitToHost, onTerminal: (response: SoulRecoveryTerminalResponse) => terminals.push(response) } });
  return { target, app };
}

for (const replacement of [false, true]) {
  for (const lateReply of ["watchdog", "resolve", "gone"] as const) {
    it.skipIf(!browser)(`retired Recovery heartbeat cannot affect replacement (${replacement}, ${lateReply})`, async () => {
      const { userEvent } = await import("vitest/browser");
      const clock = { now: 1_000 };
      const visibility = new FakeVisibility();
      const recoveryContent = content();
      const nextSession = replacement ? OTHER_SESSION : SESSION;
      const activity = recoveryContent.catalog.recovery_activities.find((row) => row.activity_id === (replacement ? "repot" : "defrag"))!;
      const requests: { operation: string; body: Record<string, unknown> }[] = [];
      const terminals: SoulRecoveryTerminalResponse[] = [];
      let starts = 0;
      let release!: (response: Response) => void;
      const pendingBeat = new Promise<Response>((resolve) => { release = resolve; });
      const response = terminal(lateReply === "resolve" ? "resolve" : "cancel", lateReply === "watchdog" ? "watchdog" : undefined);
      const port = createBrowserSoulRecoveryPort(() => "fixture-access-token", async (input, init) => {
        expect(init?.method).toBe("POST");
        const operation = String(input).split("/").at(-1)!;
        const body = JSON.parse(String(init?.body)) as Record<string, unknown>;
        requests.push({ operation, body });
        if (operation === "start") {
          starts += 1;
          return Response.json(starts === 1 ? started() : {
            ...started(nextSession, TOKEN_2, 1_234), activity_id: activity.activity_id,
            required_duration_attended_ms: activity.duration_attended_ms,
          });
        }
        expect(operation).toBe("progress");
        if (body.progress_token === TOKEN_1) {
          expect(body.session_id).toBe(SESSION);
          return pendingBeat;
        }
        expect(body).toEqual({ session_id: nextSession, progress_token: TOKEN_2 });
        return Response.json({ ...progressed(1_234, nextSession), required_duration_attended_ms: activity.duration_attended_ms });
      });
      const target = host();
      const app = mount(SoulRecoverySurface, { target, props: { port, content: recoveryContent, era: "era_1995", visibility,
        now: () => clock.now, toySeed: 7, onExitToHost: () => {}, onTerminal: (receipt: SoulRecoveryTerminalResponse) => terminals.push(receipt) } });
      try {
        await settle();
        target.querySelector<HTMLElement>("#recovery-heading")!.focus();
        await userEvent.keyboard("{Tab}{Enter}");
        await expect.poll(() => requests.filter((request) => request.operation === "progress").length).toBe(1);
        // Sleep past the unchanged ceiling while the old request is outstanding.
        // Foreground requires an explicit reconnect, not a fabricated network error.
        visibility.emit(false);
        clock.now += 100;
        visibility.emit(true);
        await settle();
        button(target, "Reconnect").focus();
        await userEvent.keyboard("{Enter}");
        await expect.poll(() => starts).toBe(2);
        await expect.poll(() => target.querySelector<HTMLProgressElement>("progress")?.value).toBe(1_234);
        visibility.emit(false);

        release(lateReply === "gone"
          ? Response.json({ category: "unknown_id", detail: "recovery_session" }, { status: 404 })
          : Response.json(response));
        await settle();
        if (!replacement) {
          // Rotation retains the same scheduler/session: an authoritative late
          // terminal or gone reply must still end that session, not be discarded.
          if (lateReply === "gone") {
            await expect.poll(() => target.querySelector("[role=alert]")?.textContent).toBe(t("soul.recovery_surface.gone", {}, "era_1995"));
            expect(terminals).toEqual([]);
          } else {
            await expect.poll(() => terminals).toEqual([response]);
          }
          expect(target.querySelector("progress")).toBeNull();
          expect(visibility.listener).toBeUndefined();
          return;
        }

        expect(terminals).toEqual([]);
        expect(target.querySelector<HTMLProgressElement>("progress")?.value).toBe(1_234);
        expect(target.querySelector("h2")?.textContent).toBe(t(activity.title_copy_key as CopyKey, {}, "era_1995"));
        // A retired 404 must not poison the new scheduler's later pause reason.
        clock.now += 100;
        visibility.emit(true);
        await settle();
        expect(button(target, "Reconnect").disabled).toBe(false);
        button(target, "Reconnect").focus();
        await userEvent.keyboard("{Enter}");
        await expect.poll(() => requests.some((request) => request.operation === "progress" && request.body.progress_token === TOKEN_2)).toBe(true);
        expect(terminals).toEqual([]);
        expect(target.querySelector<HTMLProgressElement>("progress")?.value).toBe(1_234);
        expect(requests.filter((request) => request.operation === "progress" && request.body.progress_token === TOKEN_1)).toHaveLength(1);
      } finally {
        await unmount(app); target.remove();
      }
    });
  }
}

for (const mounted of [true, false]) {
  it.skipIf(!browser)(`consumes watchdog heartbeat through the generated port only while mounted (${mounted})`, async () => {
    const { userEvent } = await import("vitest/browser");
    const requests: string[] = [];
    const response = terminal("cancel", "watchdog");
    let release!: () => void;
    const received = new Promise<void>((resolve) => { release = resolve; });
    const terminals: SoulRecoveryTerminalResponse[] = [];
    const visibility = new FakeVisibility();
    const port = createBrowserSoulRecoveryPort(() => "fixture-access-token", async (input, init) => {
      const route = String(input);
      requests.push(route);
      expect(init?.method).toBe("POST");
      if (route.endsWith("/start")) return Response.json(started());
      expect(route).toBe("/api/v1/soul-recovery/progress");
      expect(JSON.parse(String(init?.body))).toEqual({ session_id: SESSION, progress_token: TOKEN_1 });
      await received;
      return Response.json(response);
    });
    const target = host();
    const app = mount(SoulRecoverySurface, { target, props: { port, content: content(), era: "era_1995", visibility,
      now: () => 1_000, toySeed: 7, onExitToHost: () => {}, onTerminal: (receipt: SoulRecoveryTerminalResponse) => terminals.push(receipt) } });
    const newer = document.createElement("button");
    newer.textContent = "Another surface";
    document.body.append(newer);
    let unmounted = false;
    try {
      await settle();
      target.querySelector<HTMLButtonElement>("li button")!.focus();
      await userEvent.keyboard("{Enter}");
      await expect.poll(() => requests.length).toBe(2);
      if (!mounted) { await unmount(app); unmounted = true; newer.focus(); }
      release();
      if (mounted) {
        await expect.poll(() => terminals).toEqual([response]);
        expect(target.querySelector("[role=status]")?.textContent).toContain("ran too long");
        expect(target.querySelector("progress")).toBeNull();
        expect(target.querySelector("[role=alert]")).toBeNull();
        expect([...target.querySelectorAll("button")].map((control) => control.textContent?.trim())).toEqual(["Back"]);
        expect(document.activeElement?.id).toBe("recovery-heading");
        await assertAxe(target, "heartbeat watchdog terminal");
      }
      await settle(100);
      expect(requests).toEqual(["/api/v1/soul-recovery/start", "/api/v1/soul-recovery/progress"]);
      expect(visibility.listener).toBeUndefined();
      expect(terminals).toEqual(mounted ? [response] : []);
      if (!mounted) expect(document.activeElement).toBe(newer);
    } finally {
      release();
      if (!unmounted) await unmount(app);
      target.remove(); newer.remove();
    }
  });
}

for (const first of ["progress", "cancel"] as const) {
  it.skipIf(!browser)(`notifies once when watchdog heartbeat and Stop race (${first} replies first)`, async () => {
    const { userEvent } = await import("vitest/browser");
    const response = terminal("cancel", "watchdog");
    const releases = new Map<string, () => void>();
    const requests: string[] = [];
    const terminals: SoulRecoveryTerminalResponse[] = [];
    const visibility = new FakeVisibility();
    const port = createBrowserSoulRecoveryPort(() => "fixture-access-token", async (input) => {
      const operation = String(input).split("/").at(-1)!;
      requests.push(operation);
      if (operation === "start") return Response.json(started());
      expect(["progress", "cancel"]).toContain(operation);
      await new Promise<void>((resolve) => { releases.set(operation, resolve); });
      return Response.json(response);
    });
    const target = host();
    const app = mount(SoulRecoverySurface, { target, props: { port, content: content(), era: "era_1995", visibility,
      now: () => 1_000, toySeed: 7, onExitToHost: () => {}, onTerminal: (receipt: SoulRecoveryTerminalResponse) => terminals.push(receipt) } });
    try {
      await settle();
      target.querySelector<HTMLButtonElement>("li button")!.focus();
      await userEvent.keyboard("{Enter}");
      await expect.poll(() => releases.has("progress")).toBe(true);
      button(target, "Stop early").focus();
      await userEvent.keyboard("{Enter}");
      await expect.poll(() => releases.has("cancel")).toBe(true);
      releases.get(first)!();
      await expect.poll(() => terminals.length).toBe(1);
      releases.get(first === "progress" ? "cancel" : "progress")!();
      await settle(100);
      expect(terminals).toEqual([response]);
      expect(requests).toEqual(["start", "progress", "cancel"]);
      expect(visibility.listener).toBeUndefined();
      expect(target.querySelector("[role=status]")?.textContent).toContain("ran too long");
      expect(document.activeElement?.id).toBe("recovery-heading");
    } finally {
      for (const release of releases.values()) release();
      await unmount(app); target.remove();
    }
  });
}

for (const failure of ["transport", "conflict", "not-ready"] as const) {
  for (const first of ["progress", "resolve"] as const) {
    it.skipIf(!browser)(`keeps the Recovery terminal when pending Finish fails (${failure}, ${first} first)`, async () => {
      const { userEvent } = await import("vitest/browser");
      const response = terminal("cancel", "watchdog");
      const releases = new Map<string, () => void>();
      const requests: string[] = [];
      const terminals: SoulRecoveryTerminalResponse[] = [];
      const visibility = new FakeVisibility();
      const port = createBrowserSoulRecoveryPort(() => "fixture-access-token", async (input) => {
        const operation = String(input).split("/").at(-1)!;
        requests.push(operation);
        if (operation === "start") return Response.json(started(SESSION, TOKEN_1, REQUIRED));
        expect(["progress", "resolve"]).toContain(operation);
        await new Promise<void>((resolve) => { releases.set(operation, resolve); });
        if (operation === "progress") return Response.json(response);
        if (failure === "transport") throw new TypeError("Finish reply lost");
        if (failure === "not-ready") return Response.json({ category: "not_eligible", detail: "soul_recovery_not_ready" }, { status: 400 });
        return Response.json({ category: "idempotency_conflict", detail: "recovery_session" }, { status: 409 });
      });
      const target = host();
      const app = mount(SoulRecoverySurface, { target, props: { port, content: content(), era: "era_1995", visibility,
        now: () => 1_000, toySeed: 7, onExitToHost: () => {}, onTerminal: (receipt: SoulRecoveryTerminalResponse) => terminals.push(receipt) } });
      try {
        await settle();
        target.querySelector<HTMLButtonElement>("li button")!.focus();
        await userEvent.keyboard("{Enter}");
        await expect.poll(() => releases.has("progress")).toBe(true);
        button(target, "Finish").focus();
        await userEvent.keyboard("{Enter}");
        await expect.poll(() => releases.has("resolve")).toBe(true);
        releases.get(first)!();
        if (first === "progress") {
          await expect.poll(() => terminals).toEqual([response]);
        } else {
          // Without a terminal receipt, the same errors must still be surfaced.
          const selector = failure === "not-ready" ? "[role=status]" : "[role=alert]";
          const message = failure === "transport" ? "Connection lost" : failure === "not-ready" ? "Not quite done yet." : "cannot explain";
          await expect.poll(() => target.querySelector(selector)?.textContent).toContain(message);
          expect(terminals).toEqual([]);
        }
        releases.get(first === "progress" ? "resolve" : "progress")!();
        await settle(100);
        expect(terminals).toEqual([response]);
        expect(requests).toEqual(["start", "progress", "resolve"]);
        expect(visibility.listener).toBeUndefined();
        expect(target.querySelector("progress")).toBeNull();
        expect(target.querySelector("[role=alert]")).toBeNull();
        expect([...target.querySelectorAll("[role=status]")].map((node) => node.textContent))
          .toEqual(["This session ran too long and was closed. Nothing was restored."]);
        expect([...target.querySelectorAll("button")].map((control) => control.textContent?.trim())).toEqual(["Back"]);
        expect(document.activeElement?.id).toBe("recovery-heading");
      } finally {
        for (const release of releases.values()) release();
        await unmount(app); target.remove();
      }
    });
  }
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

it.skipIf(!browser)("sends no heartbeat when start completes in an already hidden document", async () => {
  const originalVisibility = Object.getOwnPropertyDescriptor(document, "visibilityState");
  let visible = false;
  Object.defineProperty(document, "visibilityState", { configurable: true, get: () => visible ? "visible" : "hidden" });
  const port = new FakePort();
  port.startReplies.push(() => started());
  port.steady = progressed(1_000);
  const target = host();
  const app = mount(SoulRecoverySurface, { target, props: { port, content: content(), era: "era_1995", now: () => 1_000, toySeed: 7,
    onExitToHost: () => {}, onTerminal: () => {} } });
  try {
    await settle();
    target.querySelector<HTMLButtonElement>("li button")!.click();
    await settle(45);
    expect(port.tokens).toEqual([]);
    expect(target.textContent).toContain("Paused while this tab is in the background.");
    visible = true;
    document.dispatchEvent(new Event("visibilitychange"));
    await settle(45);
    expect(port.tokens.length).toBeGreaterThan(0);
  } finally {
    unmount(app);
    target.remove();
    if (originalVisibility) Object.defineProperty(document, "visibilityState", originalVisibility);
    else Reflect.deleteProperty(document, "visibilityState");
  }
});

it.skipIf(!browser)("does not start heartbeats when a pending start resolves after unmount", async () => {
  const port = new FakePort();
  const visibility = new FakeVisibility();
  let releaseStart: (response: SoulRecoveryStartResponse) => void = () => { throw new Error("start was not pending"); };
  port.startReplies.push(() => new Promise((resolve) => { releaseStart = resolve; }));
  port.steady = progressed(1_000);
  const { target, app } = mountSurface(port, visibility, { now: 1_000 });
  let unmounted = false;
  try {
    await settle();
    target.querySelector<HTMLButtonElement>("li button")!.click();
    await settle();
    expect(port.starts).toEqual(["defrag"]);
    unmount(app);
    unmounted = true;
    releaseStart(started());
    await settle(80);
    expect(port.tokens).toEqual([]);
  } finally { if (!unmounted) unmount(app); target.remove(); }
});

it.skipIf(!browser)("does not restart heartbeats when a pending reconnect resolves after unmount", async () => {
  const port = new FakePort();
  const visibility = new FakeVisibility();
  port.startReplies.push(() => started());
  port.progressReplies.push(() => { throw new MinigameTransportError("offline"); });
  const { target, app } = mountSurface(port, visibility, { now: 1_000 });
  let unmounted = false;
  try {
    await settle();
    target.querySelector<HTMLButtonElement>("li button")!.click();
    await settle(80);
    expect(target.querySelector("[role=alert]")?.textContent).toContain("Connection lost");
    let releaseStart: (response: SoulRecoveryStartResponse) => void = () => { throw new Error("reconnect was not pending"); };
    port.startReplies.push(() => new Promise((resolve) => { releaseStart = resolve; }));
    button(target, "Reconnect").click();
    await settle();
    expect(port.starts).toEqual(["defrag", "defrag"]);
    const beatsBeforeUnmount = port.tokens.length;
    unmount(app);
    unmounted = true;
    releaseStart(started(SESSION, TOKEN_2, 1_000));
    await settle(80);
    expect(port.tokens).toHaveLength(beatsBeforeUnmount);
  } finally { if (!unmounted) unmount(app); target.remove(); }
});

it.skipIf(!browser)("native Tab reaches Soul recovery choices, Stop and terminal Return", async () => {
  const { userEvent } = await import("vitest/browser");
  const port = new FakePort();
  const visibility = new FakeVisibility();
  const terminals: SoulRecoveryTerminalResponse[] = [];
  port.startReplies.push(() => started());
  port.steady = progressed(1_000);
  port.terminalReply = () => terminal("cancel", "player");
  let exits = 0;
  const { target, app } = mountSurface(port, visibility, { now: 1_000 }, terminals, () => { exits += 1; });
  try {
    await settle();
    target.querySelector<HTMLElement>("#recovery-heading")!.focus();
    const choices = [...target.querySelectorAll<HTMLButtonElement>("li button")];
    expect(choices).toHaveLength(3);
    for (const choice of [...choices, button(target, "Back")]) {
      await userEvent.keyboard("{Tab}");
      expect(document.activeElement).toBe(choice);
    }
    for (const choice of [...choices].reverse()) {
      await userEvent.keyboard("{Shift>}{Tab}{/Shift}");
      expect(document.activeElement).toBe(choice);
    }
    await userEvent.keyboard("{Enter}");
    await settle();
    expect(port.starts).toEqual(["defrag"]);
    await assertAxe(target, "keyboard-active recovery");
    target.querySelector<HTMLElement>("#recovery-heading")!.focus();
    await userEvent.keyboard("{Tab}");
    expect(document.activeElement).toBe(button(target, "Stop early"));
    await userEvent.keyboard("{Enter}");
    await settle();
    expect(terminals).toEqual([terminal("cancel", "player")]);
    expect(document.activeElement?.id).toBe("recovery-heading");
    await assertAxe(target, "keyboard-terminal recovery");
    await userEvent.keyboard("{Tab}");
    expect(document.activeElement).toBe(button(target, "Back"));
    await userEvent.keyboard(" ");
    expect(exits).toBe(1);
    expect(port.terminalCalls).toEqual(["cancel"]);
  } finally { unmount(app); target.remove(); }
});

it.skipIf(!browser)("native Tab reaches Soul recovery Reconnect and Finish", async () => {
  const { userEvent } = await import("vitest/browser");
  const port = new FakePort();
  const visibility = new FakeVisibility();
  const terminals: SoulRecoveryTerminalResponse[] = [];
  port.startReplies.push(() => started(), () => started(SESSION, TOKEN_2, REQUIRED));
  port.progressReplies.push(() => { throw new MinigameTransportError("offline"); }, () => progressed(REQUIRED));
  let exits = 0;
  const { target, app } = mountSurface(port, visibility, { now: 1_000 }, terminals, () => { exits += 1; });
  try {
    await settle();
    target.querySelector<HTMLElement>("#recovery-heading")!.focus();
    await userEvent.keyboard("{Tab}");
    expect(document.activeElement).toBe(target.querySelector("li button"));
    await userEvent.keyboard("{Enter}");
    await expect.poll(() => target.querySelector("[role=alert]")?.textContent).toContain("Connection lost");
    target.querySelector<HTMLElement>("#recovery-heading")!.focus();
    await userEvent.keyboard("{Tab}");
    expect(document.activeElement).toBe(button(target, "Reconnect"));
    await userEvent.keyboard("{Tab}");
    expect(document.activeElement).toBe(button(target, "Stop early"));
    await userEvent.keyboard("{Shift>}{Tab}{/Shift}");
    expect(document.activeElement).toBe(button(target, "Reconnect"));
    await userEvent.keyboard("{Enter}");
    await settle();
    expect(port.starts).toEqual(["defrag", "defrag"]);
    await expect.poll(() => port.tokens.includes(TOKEN_2)).toBe(true);
    expect(target.querySelector("[role=alert]")).toBeNull();
    target.querySelector<HTMLElement>("#recovery-heading")!.focus();
    await userEvent.keyboard("{Tab}");
    expect(document.activeElement).toBe(button(target, "Finish"));
    await userEvent.keyboard("{Tab}");
    expect(document.activeElement).toBe(button(target, "Stop early"));
    await userEvent.keyboard("{Shift>}{Tab}{/Shift}");
    expect(document.activeElement).toBe(button(target, "Finish"));
    await userEvent.keyboard(" ");
    await settle();
    expect(port.terminalCalls).toEqual(["resolve"]);
    expect(terminals).toEqual([terminal("resolve")]);
    expect(document.activeElement?.id).toBe("recovery-heading");
    await userEvent.keyboard("{Tab}");
    expect(document.activeElement).toBe(button(target, "Back"));
    await userEvent.keyboard("{Enter}");
    expect(exits).toBe(1);
  } finally { unmount(app); target.remove(); }
});

for (const effect of ["gone", "error"] as const) {
  it.skipIf(!browser)(`native Tab reaches Soul recovery Return after ${effect}`, async () => {
    const { userEvent } = await import("vitest/browser");
    const port = new FakePort();
    port.startReplies.push(() => {
      if (effect === "gone") throw new MinigameAPIError(404, "unknown_id", "recovery_session");
      throw new MinigameAPIError(503, "not_configured", "soul_recovery");
    });
    let exits = 0;
    const { target, app } = mountSurface(port, new FakeVisibility(), { now: 1_000 }, [], () => { exits += 1; });
    try {
      await settle();
      target.querySelector<HTMLElement>("#recovery-heading")!.focus();
      await userEvent.keyboard("{Tab}");
      expect(document.activeElement).toBe(target.querySelector("li button"));
      await userEvent.keyboard("{Enter}");
      await settle();
      expect(port.starts).toEqual(["defrag"]);
      expect(target.querySelector("[role=alert]")).not.toBeNull();
      target.querySelector<HTMLElement>("#recovery-heading")!.focus();
      await userEvent.keyboard("{Tab}");
      expect(document.activeElement).toBe(button(target, "Back"));
      await userEvent.keyboard("{Enter}");
      expect(exits).toBe(1);
      expect(port.terminalCalls).toEqual([]);
      expect(port.tokens).toEqual([]);
    } finally { unmount(app); target.remove(); }
  });
}

for (const action of ["resolve", "cancel"] as const) {
  for (const mounted of [true, false]) {
    it.skipIf(!browser)(`delayed Recovery ${action} notifies only a still-mounted host (${mounted})`, async () => {
      const { userEvent } = await import("vitest/browser");
      const port = new FakePort();
      const visibility = new FakeVisibility();
      const terminals: SoulRecoveryTerminalResponse[] = [];
      const response = terminal(action, action === "cancel" ? "player" : undefined);
      port.startReplies.push(() => started(SESSION, TOKEN_1, REQUIRED));
      port.steady = progressed(REQUIRED);
      let releaseTerminal!: (value: SoulRecoveryTerminalResponse) => void;
      port.terminalReply = () => new Promise((resolve) => { releaseTerminal = resolve; });
      const { target, app } = mountSurface(port, visibility, { now: 1_000 }, terminals);
      const newer = document.createElement("button");
      newer.textContent = "Another surface";
      document.body.append(newer);
      let unmounted = false;
      try {
        await settle();
        target.querySelector<HTMLButtonElement>("li button")!.focus();
        await userEvent.keyboard("{Enter}");
        await settle();
        button(target, action === "resolve" ? "Finish" : "Stop early").focus();
        await userEvent.keyboard("{Enter}");
        await settle();
        expect(port.terminalCalls).toEqual([action]);
        expect(terminals).toEqual([]);
        if (!mounted) {
          await unmount(app);
          unmounted = true;
          newer.focus();
        }
        releaseTerminal(response);
        await settle(80);
        expect(port.terminalCalls).toEqual([action]);
        expect(terminals).toEqual(mounted ? [response] : []);
        if (mounted) {
          expect(target.querySelector("[role=status]")?.textContent).toBe(action === "resolve" ? "Done. Soul went from 40 to 52." : "Stopped. Nothing was restored.");
          expect(document.activeElement?.id).toBe("recovery-heading");
        } else {
          expect(document.activeElement).toBe(newer);
          expect(target.textContent).toBe("");
          expect(visibility.listener).toBeUndefined();
          const beats = port.tokens.length;
          await settle(80);
          expect(port.tokens).toHaveLength(beats);
        }
      } finally {
        if (!unmounted) await unmount(app);
        target.remove(); newer.remove();
      }
    });
  }
}
for (const action of ["resolve", "cancel"] as const) {
  for (const newerFocus of ["pending", "host-callback"] as const) {
    it.skipIf(!browser)(`Recovery terminal preserves newer focus (${action}, ${newerFocus})`, async () => {
      const { userEvent } = await import("vitest/browser");
      const port = new FakePort();
      const visibility = new FakeVisibility();
      const terminals: SoulRecoveryTerminalResponse[] = [];
      const response = terminal(action, action === "cancel" ? "player" : undefined);
      port.startReplies.push(() => started(SESSION, TOKEN_1, REQUIRED));
      port.steady = progressed(REQUIRED);
      let releaseTerminal!: (value: SoulRecoveryTerminalResponse) => void;
      port.terminalReply = () => new Promise((resolve) => { releaseTerminal = resolve; });
      const newer = document.createElement("button");
      newer.type = "button"; newer.tabIndex = 0; newer.textContent = "Another host control";
      const target = host();
      const app = mount(SoulRecoverySurface, { target, props: { port, content: content(), era: "era_1995", visibility,
        now: () => 1_000, toySeed: 7, onExitToHost: () => {}, onTerminal: (receipt: SoulRecoveryTerminalResponse) => {
          terminals.push(receipt);
          if (newerFocus === "host-callback") newer.focus();
        } } });
      document.body.append(newer);
      try {
        await settle();
        target.querySelector<HTMLElement>("#recovery-heading")!.focus();
        await userEvent.keyboard("{Tab}");
        expect(document.activeElement).toBe(target.querySelector("li button"));
        await userEvent.keyboard("{Enter}");
        await expect.poll(() => target.querySelector("progress") !== null).toBe(true);
        const origin = button(target, action === "resolve" ? "Finish" : "Stop early");
        origin.focus();
        await userEvent.keyboard("{Enter}");
        await expect.poll(() => port.terminalCalls).toEqual([action]);
        await expect.poll(() => origin.disabled).toBe(true);
        if (newerFocus === "pending") {
          await userEvent.keyboard("{Tab}");
          expect(document.activeElement).toBe(newer);
        }
        releaseTerminal(response);
        await expect.poll(() => terminals).toEqual([response]);
        await expect.poll(() => target.querySelector("progress")).toBeNull();
        await settle();
        expect(document.activeElement).toBe(newer);
        expect(target.querySelector("[role=status]")?.textContent).toBe(action === "resolve" ? "Done. Soul went from 40 to 52." : "Stopped. Nothing was restored.");
        expect(port.terminalCalls).toEqual([action]);
        expect(visibility.listener).toBeUndefined();
      } finally {
        releaseTerminal?.(response);
        await unmount(app); target.remove(); newer.remove();
      }
    });
  }
}
for (const returnedActivity of ["repot", "server_room", "unknown_activity"] as const) {
  it.skipIf(!browser)(`Recovery reconnect binds returned activity (${returnedActivity})`, async () => {
    const { userEvent } = await import("vitest/browser");
    const recoveryContent = content();
    const activity = recoveryContent.catalog.recovery_activities.find((row) => row.activity_id === returnedActivity);
    const required = activity?.duration_attended_ms ?? REQUIRED;
    const requests: { operation: string; body: Record<string, unknown> }[] = [];
    const terminals: SoulRecoveryTerminalResponse[] = [];
    const visibility = new FakeVisibility();
    let starts = 0;
    const port = createBrowserSoulRecoveryPort(() => "fixture-access-token", async (input, init) => {
      expect(init?.method).toBe("POST");
      const operation = String(input).split("/").at(-1)!;
      const body = JSON.parse(String(init?.body)) as Record<string, unknown>;
      requests.push({ operation, body });
      if (operation === "start") {
        starts += 1;
        expect(body).toEqual({ activity_id: "defrag" });
        return Response.json(starts === 1 ? started() : {
          ...started(OTHER_SESSION, TOKEN_2, 1_234), activity_id: returnedActivity,
          required_duration_attended_ms: required,
        });
      }
      expect(operation).toBe("progress");
      if (body.session_id === SESSION) {
        expect(body.progress_token).toBe(TOKEN_1);
        return Response.json({ category: "not_eligible", detail: "recovery_token" }, { status: 400 });
      }
      expect(body).toEqual({ session_id: OTHER_SESSION, progress_token: TOKEN_2 });
      return Response.json({ ...progressed(1_234, OTHER_SESSION), required_duration_attended_ms: required, eligible: false });
    });
    const target = host();
    const app = mount(SoulRecoverySurface, { target, props: { port, content: recoveryContent, era: "era_1995", visibility,
      now: () => 1_000, toySeed: 7, onExitToHost: () => {}, onTerminal: (response: SoulRecoveryTerminalResponse) => terminals.push(response) } });
    try {
      await settle();
      target.querySelector<HTMLElement>("#recovery-heading")!.focus();
      await userEvent.keyboard("{Tab}{Enter}");
      await expect.poll(() => [...target.querySelectorAll("button")].some((control) => control.textContent?.trim() === "Reconnect")).toBe(true);
      button(target, "Reconnect").focus();
      await userEvent.keyboard("{Enter}");
      await expect.poll(() => starts).toBe(2);
      if (activity) {
        await expect.poll(() => requests.some((request) => request.operation === "progress" && request.body.session_id === OTHER_SESSION)).toBe(true);
        expect(target.querySelector("h2")?.textContent).toBe(t(activity.title_copy_key as CopyKey, {}, "era_1995"));
        expect(target.querySelector("[data-toy]")?.getAttribute("data-toy")).toBe(activity.toy_kind);
        expect(target.querySelector<HTMLProgressElement>("progress")?.max).toBe(required);
        expect(target.querySelector<HTMLProgressElement>("progress")?.value).toBe(1_234);
        expect(target.textContent).toContain(t(activity.disclosure_copy_key as CopyKey, {}, "era_1995"));
      } else {
        await expect.poll(() => target.querySelector("[role=alert]")?.textContent).toBe(t("minigame.error.generic", {}, "era_1995"));
        expect(target.querySelector("progress")).toBeNull();
        expect(target.querySelector("[data-toy]")).toBeNull();
        expect(visibility.listener).toBeUndefined();
        expect(requests.some((request) => request.operation === "progress" && request.body.session_id === OTHER_SESSION)).toBe(false);
      }
      expect(requests.filter((request) => request.operation === "progress" && request.body.session_id === SESSION)).toHaveLength(1);
      expect(terminals).toEqual([]);
    } finally {
      await unmount(app); target.remove();
    }
  });
}
