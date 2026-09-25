import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

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

it.skipIf(!browser)("mounts The Pitch as the tenant child and plays launcher → table → shop → terminal through the port only", async () => {
  const snapshot = await pitchSnapshot();
  const port = new FakePort();
  port.createReply = () => active(snapshot, 1);
  const terminals: MinigameResolutionReceipt[] = [];
  const target = host();
  const app = mount(MinigameSessionSurface, { target, props: { port, minigameID: "pitch", era: "era_1995", newCommandID: nextID, onExitToHost: () => {}, onTerminal: (value: MinigameResolutionReceipt) => terminals.push(value) } });
  try {
    await settle();
    await assertAxe(target, "launcher");
    buttonNamed(target, "Start a pitch").click();
    await settle();
    expect(port.creates).toHaveLength(1);
    await assertAxe(target, "active table");
    // No mechanical instance IDs leak into presentation.
    for (const instance of snapshot.hand) expect(target.textContent).not.toContain(instance);

    const boxes = [...target.querySelectorAll<HTMLInputElement>("fieldset input[type=checkbox]")];
    expect(boxes).toHaveLength(7);
    // Keyboard selection: focus + Space on native checkboxes; the fifth is refused visibly.
    for (const box of boxes.slice(0, 5).reverse()) { box.focus(); box.click(); }
    await settle();
    expect(boxes.filter((box) => box.checked)).toHaveLength(4);
    expect(boxes[0]!.getAttribute("aria-disabled")).toBe("true");
    expect(target.querySelector("fieldset [role=status]")).not.toBeNull();

    port.replies.push(() => active({ ...snapshot, phase: "shop", hand: [], shop_offers: [{ hack_id: "buzzword", offer_id: "offer.1", price: 2 }] }, 2));
    buttonNamed(target, "Pitch these cards").click();
    await settle();
    const played = port.commands[0]!;
    expect(played.expected_revision).toBe(1);
    expect(played.command).toEqual({ kind: "play_hand", card_ids: [...snapshot.hand.slice(1, 5)].sort() });
    await assertAxe(target, "shop");

    port.replies.push(() => active({ ...snapshot, phase: "shop", hand: [], run_currency: 2, shop_offers: [], slotted_hacks: ["buzzword"] }, 3));
    buttonNamed(target, "Buy for 2").click();
    await settle();
    expect(port.commands[1]).toMatchObject({ expected_revision: 2, command: { kind: "buy_hack", offer_id: "offer.1" } });

    port.replies.push(() => ({ ...descriptor, revision: 4, snapshot: { ...snapshot, phase: "terminal", hand: [], revision: 4 }, status: "resolved", resolution_receipt: receipt }));
    buttonNamed(target, "Close the shop").click();
    await settle();
    expect(port.commands[2]).toMatchObject({ expected_revision: 3, command: { kind: "end_shop" } });
    expect(terminals).toEqual([receipt]);
    expect(document.activeElement?.id).toBe("minigame-heading");
    expect(target.textContent).toContain("Credited to the company");
    await assertAxe(target, "terminal");
    expect(new Set(port.commands.map((command) => command.command_id)).size).toBe(3);
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
