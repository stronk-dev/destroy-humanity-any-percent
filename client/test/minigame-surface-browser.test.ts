import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { describe, expect, it } from "vitest";

import type { MinigameCommandRequest, MinigameCurrentResponse, MinigameResolutionReceipt, MinigameSessionResponse, MinigameSessionResponseActive, MinigameSessionResponseTerminal, PitchSnapshot } from "../src/api/generated/types";
import MinigameSessionSurface from "../src/game-ui/minigame/MinigameSessionSurface.svelte";
import { loadPitchContent } from "../src/game-ui/minigame/pitch-content";
import { MinigameAPIError, MinigameTransportError, type MinigameSessionPort } from "../src/game-ui/minigame/session-port";
import { installTheme, UI_THEMES } from "../src/ui/themes";

const browser = typeof document !== "undefined";
const SESSION = "01986666-0000-7000-8000-000000000001";
const descriptor = { constants_hash: `sha256:${"a".repeat(64)}`, engine_ref: "pitch", engine_version: "1.0.0", minigame_id: "pitch", mode: "solo" as const, session_id: SESSION };

async function pitchSnapshot(overrides: Partial<PitchSnapshot> = {}): Promise<PitchSnapshot> {
  const content = await loadPitchContent();
  const hand = content.catalog.metric_cards.slice(0, 7).map((card) => `${card.card_id}#1`).sort();
  return { deck_count: 17, funding_target: content.catalog.funding_curve[0]!.funding_target, hand, hands_remaining: 3, phase: "playing", pitch_content_hash: content.hash,
    pitch_schema_version: 1, revision: 1, round: 1, round_best_valuation: "0", run_currency: 4, shop_offers: [], slotted_hacks: [], ...overrides };
}

function active(snapshot: PitchSnapshot, revision: number): MinigameSessionResponseActive { return { ...descriptor, revision, snapshot: { ...snapshot, revision }, status: "active" }; }

const receipt: MinigameResolutionReceipt = { cap_reason_key: "cap.minigame_faucet", certified_result_hash: `sha256:${"c".repeat(64)}`, company_revision: 9, configured_cap_forfeit_units: 0,
  credited_delta: "2.5e2", credited_resource_id: "company.cash", founder_revision: 4, intent_id: "01986666-0000-7000-8000-00000000000f", minigame_id: "pitch", outcome: "applied",
  quality_change: { new: { decay_remainder_ppm: 0, grade_ppm: 1, last_founder_attended_ms: 1 }, old: { decay_remainder_ppm: 0, grade_ppm: 0, last_founder_attended_ms: 0 } },
  rating_change: { games_after: 1, games_before: 0, new_elo: 1000, old_elo: 1000, rated: false, season_member: "none" }, session_id: SESSION };

class FakePort implements MinigameSessionPort {
  currentValue: MinigameCurrentResponse = { kind: "none" };
  creates: string[] = [];
  commands: MinigameCommandRequest[] = [];
  replies: ((request: MinigameCommandRequest) => MinigameSessionResponse)[] = [];
  createReply: () => MinigameSessionResponseActive = () => { throw new Error("no create reply"); };
  async current(): Promise<MinigameCurrentResponse> { return this.currentValue; }
  async create(_minigameID: string, key: string): Promise<MinigameSessionResponseActive> { this.creates.push(key); return this.createReply(); }
  async command(_sessionID: string, request: MinigameCommandRequest): Promise<MinigameSessionResponse> {
    this.commands.push(request);
    const next = this.replies.shift();
    if (!next) throw new Error("unexpected command");
    return next(request);
  }
  async resolve(): Promise<MinigameSessionResponseTerminal> { throw new Error("resolve is not used by these fixtures"); }
}

async function settle(): Promise<void> { for (let index = 0; index < 6; index += 1) { await new Promise((resolve) => setTimeout(resolve, 0)); await tick(); flushSync(); } }

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

function buttonNamed(target: HTMLElement, text: string): HTMLButtonElement {
  const button = [...target.querySelectorAll("button")].find((candidate) => candidate.textContent?.trim().startsWith(text));
  if (!button) throw new Error(`missing button ${text}: ${target.textContent}`);
  return button;
}

let ids = 0;
function nextID(): string { ids += 1; return `01986666-0000-7000-8000-${ids.toString(16).padStart(12, "0")}`; }

// Failure-only observation, armed before the first asynchronous setup operation.
// A framework abort fires before finally removes the fixture from a screenshot.
function observePitchDeadline(signal: AbortSignal, read: () => Record<string, unknown>,
  report: (value: Record<string, unknown>) => void) {
  const began = performance.now();
  let stage = "setup", stageBegan = began;
  const completed: { stage: string; elapsed_ms: number }[] = [];
  const capture = () => {
    const now = performance.now();
    report({ stage, elapsed_ms: now - began, stage_elapsed_ms: now - stageBegan,
      completed_stages: completed.map(value => ({ ...value })), context: read() });
  };
  signal.addEventListener("abort", capture, { once: true });
  return {
    enter(next: string) {
      const now = performance.now();
      completed.push({ stage, elapsed_ms: now - stageBegan });
      stage = next; stageBegan = now;
    },
    dispose() { signal.removeEventListener("abort", capture); },
  };
}

describe("Pitch journey deadline observation", () => {
  let signal: AbortSignal;
  let release: () => void;
  let continuation: Promise<void>;
  let target: HTMLElement;
  const observations: Record<string, unknown>[] = [];

  it.skipIf(!browser).fails("fires a real deadline while the observed stage is waiting", async context => {
    signal = context.signal;
    target = host();
    const observation = observePitchDeadline(signal,
      () => ({ fixture_connected: target.isConnected }), value => observations.push(value));
    context.onTestFinished(() => { observation.dispose(); target.remove(); });
    observation.enter("waiting-controlled-gate");
    continuation = new Promise<void>(resolve => { release = resolve; });
    await continuation;
  }, 50);

  it.skipIf(!browser)("captures the stage before framework cleanup rather than a blank successor", async () => {
    try {
      expect(signal.aborted).toBe(true);
      expect(signal.reason.message).toContain("Test timed out in 50ms");
      expect(target.isConnected).toBe(false);
      expect(observations).toHaveLength(1);
      expect(observations[0]).toMatchObject({ stage: "waiting-controlled-gate", context: { fixture_connected: true },
        completed_stages: [{ stage: "setup", elapsed_ms: expect.any(Number) }] });
      for (const key of ["elapsed_ms", "stage_elapsed_ms"] as const) {
        expect(Number.isFinite(observations[0]![key])).toBe(true);
        expect(observations[0]![key]).toBeGreaterThanOrEqual(0);
      }
    } finally { release(); await continuation; }
  });

  it.skipIf(!browser)("keeps healthy observations quiet and removes the abort listener", () => {
    const controller = new AbortController();
    const emitted: Record<string, unknown>[] = [];
    const observation = observePitchDeadline(controller.signal, () => ({ fixture_connected: false }), value => emitted.push(value));
    observation.enter("load-content");
    observation.enter("mount");
    expect(emitted).toEqual([]);
    observation.dispose();
    controller.abort(new Error("after disposal"));
    expect(emitted).toEqual([]);
  });
});

it.skipIf(!browser)("activates Pitch launcher → table → shop → terminal → Return by native keyboard through the port only", async context => {
  let target: HTMLElement | undefined;
  let port: FakePort | undefined;
  const observation = observePitchDeadline(context.signal, () => ({
    fixture_ready: target !== undefined, fixture_connected: target?.isConnected ?? false,
    creates: port?.creates.length ?? 0, commands: port?.commands.map(value => value.command.kind) ?? [],
    visibility: document.visibilityState, document_has_focus: document.hasFocus(),
    active_element: document.activeElement ? { tag: document.activeElement.tagName, id: document.activeElement.id } : null,
  }), value => console.error("pitch-journey-deadline", JSON.stringify(value)));
  context.onTestFinished(() => observation.dispose());
  observation.enter("import-native-driver");
  const { userEvent } = await import("vitest/browser");
  observation.enter("load-content");
  const snapshot = await pitchSnapshot();
  port = new FakePort();
  port.createReply = () => active(snapshot, 1);
  const terminals: MinigameResolutionReceipt[] = [];
  let exits = 0;
  observation.enter("mount");
  target = host();
  const app = mount(MinigameSessionSurface, { target, props: { port, minigameID: "pitch", era: "era_1995", newCommandID: nextID, onExitToHost: () => { exits += 1; }, onTerminal: (value: MinigameResolutionReceipt) => terminals.push(value) } });
  try {
    observation.enter("settle-launcher");
    await settle();
    observation.enter("axe-launcher");
    await assertAxe(target, "launcher");
    target.querySelector<HTMLElement>("#minigame-heading")!.focus();
    observation.enter("native-launcher-tab");
    await userEvent.keyboard("{Tab}");
    expect(document.activeElement).toBe(buttonNamed(target, "Start a pitch"));
    observation.enter("native-launcher-enter");
    await userEvent.keyboard("{Enter}");
    observation.enter("settle-active-table");
    await settle();
    expect(port.creates).toHaveLength(1);
    observation.enter("axe-active-table");
    await assertAxe(target, "active table");
    // No mechanical instance IDs leak into presentation.
    for (const instance of snapshot.hand) expect(target.textContent).not.toContain(instance);

    const boxes = [...target.querySelectorAll<HTMLInputElement>("fieldset input[type=checkbox]")];
    expect(boxes).toHaveLength(7);
    // Keyboard selection: focus + Space on native checkboxes; the fifth is refused visibly.
    boxes[4]!.focus();
    for (let index = 4; index >= 0; index -= 1) {
      expect(document.activeElement).toBe(boxes[index]);
      observation.enter(`native-card-${index}-space`);
      await userEvent.keyboard(" ");
      if (index > 0) {
        observation.enter(`native-card-${index}-previous`);
        await userEvent.keyboard("{Shift>}{Tab}{/Shift}");
      }
    }
    observation.enter("settle-selected-cards");
    await settle();
    expect(boxes.filter((box) => box.checked)).toHaveLength(4);
    expect(boxes[0]!.getAttribute("aria-disabled")).toBe("true");
    expect(target.querySelector("fieldset [role=status]")).not.toBeNull();

    port.replies.push(() => active({ ...snapshot, phase: "shop", hand: [], shop_offers: [{ hack_id: "buzzword", offer_id: "offer.1", price: 2 }] }, 2));
    buttonNamed(target, "Pitch these cards").focus();
    observation.enter("native-pitch-enter");
    await userEvent.keyboard("{Enter}");
    observation.enter("settle-shop");
    await settle();
    const played = port.commands[0]!;
    expect(played.expected_revision).toBe(1);
    expect(played.command).toEqual({ kind: "play_hand", card_ids: [...snapshot.hand.slice(1, 5)].sort() });
    observation.enter("axe-shop");
    await assertAxe(target, "shop");

    port.replies.push(() => active({ ...snapshot, phase: "shop", hand: [], run_currency: 2, shop_offers: [], slotted_hacks: ["buzzword"] }, 3));
    buttonNamed(target, "Buy for 2").focus();
    observation.enter("native-buy-space");
    await userEvent.keyboard(" ");
    observation.enter("settle-purchase");
    await settle();
    expect(port.commands[1]).toMatchObject({ expected_revision: 2, command: { kind: "buy_hack", offer_id: "offer.1" } });

    port.replies.push(() => ({ ...descriptor, revision: 4, snapshot: { ...snapshot, phase: "terminal", hand: [], revision: 4 }, status: "resolved", resolution_receipt: receipt }));
    buttonNamed(target, "Close the shop").focus();
    observation.enter("native-close-shop-enter");
    await userEvent.keyboard("{Enter}");
    observation.enter("settle-terminal");
    await settle();
    expect(port.commands[2]).toMatchObject({ expected_revision: 3, command: { kind: "end_shop" } });
    expect(terminals).toEqual([receipt]);
    expect(document.activeElement?.id).toBe("minigame-heading");
    expect(target.textContent).toContain("Credited to the company");
    observation.enter("axe-terminal");
    await assertAxe(target, "terminal");
    expect(new Set(port.commands.map((command) => command.command_id)).size).toBe(3);
    observation.enter("native-return-tab");
    await userEvent.keyboard("{Tab}");
    expect(document.activeElement).toBe(buttonNamed(target, "Back to the desk"));
    observation.enter("native-return-space");
    await userEvent.keyboard(" ");
    expect(exits).toBe(1);
    expect(port.commands).toHaveLength(3);
  } finally { unmount(app); target.remove(); }
});

it.skipIf(!browser)("resends the same command_id after a transport failure and clears selection on a revision conflict", async () => {
  const snapshot = await pitchSnapshot();
  const port = new FakePort();
  port.currentValue = { kind: "active", session: { ...descriptor, revision: 1, status: "active" }, snapshot };
  const target = host();
  const app = mount(MinigameSessionSurface, { target, props: { port, minigameID: "pitch", era: "era_1995", newCommandID: nextID, onExitToHost: () => {}, onTerminal: () => {}, retryDelayMS: 60_000 } });
  try {
    await settle();
    target.querySelector<HTMLInputElement>("fieldset input")!.click();
    await settle();
    port.replies.push(() => { throw new MinigameTransportError("offline"); });
    buttonNamed(target, "Pitch these cards").click();
    await settle();
    expect(target.querySelector("[role=alert]")?.textContent).toContain("Connection lost");
    await assertAxe(target, "paused_reconnect");
    port.replies.push(() => active(snapshot, 2));
    buttonNamed(target, "Retry").click();
    await settle();
    expect(port.commands).toHaveLength(2);
    expect(port.commands[1]).toEqual(port.commands[0]);

    target.querySelector<HTMLInputElement>("fieldset input")!.click();
    await settle();
    port.currentValue = { kind: "active", session: { ...descriptor, revision: 5, status: "active" }, snapshot: { ...snapshot, revision: 5 } };
    port.replies.push(() => { throw new MinigameAPIError(409, "conflict", "minigame_revision"); });
    buttonNamed(target, "Pitch these cards").click();
    await settle();
    expect(target.querySelector("[role=status]")?.textContent).toContain("moved on");
    expect([...target.querySelectorAll<HTMLInputElement>("fieldset input")].some((box) => box.checked)).toBe(false);
  } finally { unmount(app); target.remove(); }
});

it.skipIf(!browser)("fails closed on an unregistered tenant arm and on content that is not the bundled identity", async () => {
  const snapshot = await pitchSnapshot();
  for (const current of [
    { kind: "active", session: { ...descriptor, engine_version: "9.9.9", revision: 1, status: "active" }, snapshot },
    { kind: "active", session: { ...descriptor, revision: 1, status: "active" }, snapshot: { ...snapshot, pitch_content_hash: `sha256:${"0".repeat(64)}` } },
  ] as const) {
    const port = new FakePort();
    port.currentValue = current;
    const target = host();
    const app = mount(MinigameSessionSurface, { target, props: { port, minigameID: "pitch", era: "era_1995", newCommandID: nextID, onExitToHost: () => {}, onTerminal: () => {} } });
    try {
      await settle();
      expect(target.querySelector("fieldset")).toBeNull();
      expect(target.querySelector("[role=alert]")).not.toBeNull();
      await assertAxe(target, "error");
    } finally { unmount(app); target.remove(); }
  }
});

it.skipIf(!browser)("returns to the launcher with the typed reason when the minigame is locked", async () => {
  const port = new FakePort();
  port.createReply = () => { throw new MinigameAPIError(409, "not_eligible", "fiscal_unlock_required"); };
  const target = host();
  const app = mount(MinigameSessionSurface, { target, props: { port, minigameID: "pitch", era: "era_1995", newCommandID: nextID, onExitToHost: () => {}, onTerminal: () => {} } });
  try {
    await settle();
    buttonNamed(target, "Start a pitch").click();
    await settle();
    expect(target.textContent).toContain("Fiscal credit");
    buttonNamed(target, "Start a pitch").click();
    await settle();
    // The create key is held until success, so the retry is idempotent.
    expect(port.creates).toHaveLength(2);
    expect(port.creates[1]).toBe(port.creates[0]);
  } finally { unmount(app); target.remove(); }
});

it.skipIf(!browser).each(["revision conflict", "unknown session", "terminal receipt"] as const)("ignores a late %s after leaving the Pitch table", async (outcome) => {
  const { userEvent } = await import("vitest/browser");
  const snapshot = await pitchSnapshot();
  const port = new FakePort();
  const currentValue: MinigameCurrentResponse = { kind: "active", session: { ...descriptor, revision: 1, status: "active" }, snapshot };
  let reads = 0;
  port.current = async () => { reads += 1; return currentValue; };
  let complete!: (value: MinigameSessionResponse) => void;
  let fail!: (error: unknown) => void;
  const reply = new Promise<MinigameSessionResponse>((resolve, reject) => { complete = resolve; fail = reject; });
  port.command = async (_sessionID, request) => { port.commands.push(request); return reply; };
  const terminals: MinigameResolutionReceipt[] = [];
  const target = host();
  let left = false;
  const app = mount(MinigameSessionSurface, { target, props: {
    port, minigameID: "pitch", era: "era_1995", newCommandID: nextID,
    onExitToHost: () => { left = true; void unmount(app); },
    onTerminal: (value: MinigameResolutionReceipt) => terminals.push(value),
  } });
  try {
    await settle();
    target.querySelector<HTMLInputElement>("fieldset input")!.click();
    await settle();
    buttonNamed(target, "Pitch these cards").click();
    await settle();
    expect(port.commands).toHaveLength(1);
    expect(buttonNamed(target, "Pitch these cards").disabled).toBe(true);
    // Leaving is permitted while the server owns the in-flight command. It
    // closes only this presentation; it must not cancel or repeat gameplay.
    buttonNamed(target, "Leave the table").focus();
    await userEvent.keyboard(" ");
    await settle();
    expect(left).toBe(true);
    expect(target.querySelector(".minigame-session")).toBeNull();
    if (outcome === "terminal receipt") complete({ ...descriptor, revision: 2, snapshot: { ...snapshot, phase: "terminal", hand: [], revision: 2 }, status: "resolved", resolution_receipt: receipt });
    else fail(outcome === "revision conflict"
      ? new MinigameAPIError(409, "conflict", "minigame_revision")
      : new MinigameAPIError(404, "unknown_id", "minigame_session"));
    await settle();
    expect(reads, "a detached surface must not refetch after its command settles").toBe(1);
    expect(port.commands).toHaveLength(1);
    expect(terminals, "a detached surface must not refresh its former host").toEqual([]);
    expect(target.childElementCount).toBe(0);
  } finally { if (!left) await unmount(app); target.remove(); }
});
for (const newerFocus of ["pending", "host-callback"] as const) {
  it.skipIf(!browser)(`Pitch terminal preserves newer focus (${newerFocus})`, async () => {
    const { userEvent } = await import("vitest/browser");
    const snapshot = await pitchSnapshot({ phase: "shop", hand: [], shop_offers: [] });
    const port = new FakePort();
    port.currentValue = { kind: "active", session: { ...descriptor, revision: 1, status: "active" }, snapshot };
    let complete!: (value: MinigameSessionResponse) => void;
    const reply = new Promise<MinigameSessionResponse>((resolve) => { complete = resolve; });
    port.command = async (_sessionID, request) => { port.commands.push(request); return reply; };
    const response: MinigameSessionResponseTerminal = { ...descriptor, revision: 2,
      snapshot: { ...snapshot, phase: "terminal", hand: [], revision: 2 }, status: "resolved", resolution_receipt: receipt };
    const terminals: MinigameResolutionReceipt[] = [];
    const newer = document.createElement("button");
    newer.type = "button"; newer.tabIndex = 0; newer.textContent = "Another host control";
    const target = host();
    const app = mount(MinigameSessionSurface, { target, props: { port, minigameID: "pitch", era: "era_1995", newCommandID: nextID,
      onExitToHost: () => {}, onTerminal: (value: MinigameResolutionReceipt) => {
        terminals.push(value);
        if (newerFocus === "host-callback") newer.focus();
      } } });
    document.body.append(newer);
    try {
      await expect.poll(() => target.querySelector(".pitch") !== null).toBe(true);
      const origin = buttonNamed(target, "Close the shop");
      origin.focus();
      await userEvent.keyboard("{Enter}");
      await expect.poll(() => port.commands.length).toBe(1);
      await expect.poll(() => origin.disabled).toBe(true);
      expect(port.commands[0]).toMatchObject({ expected_revision: 1, command: { kind: "end_shop" } });
      if (newerFocus === "pending") {
        await userEvent.keyboard("{Tab}");
        expect(document.activeElement).toBe(buttonNamed(target, "Leave the table"));
        await userEvent.keyboard("{Tab}");
        expect(document.activeElement).toBe(newer);
      }
      complete(response);
      await expect.poll(() => terminals).toEqual([receipt]);
      await expect.poll(() => target.querySelector(".pitch")).toBeNull();
      await settle();
      expect(document.activeElement).toBe(newer);
      expect(target.textContent).toContain("Credited to the company");
      expect(port.commands).toHaveLength(1);
    } finally {
      complete(response);
      await unmount(app); target.remove(); newer.remove();
    }
  });
}
