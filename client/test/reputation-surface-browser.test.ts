import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import type { GameUIReputationArm } from "../src/api/generated/types";
import ReputationTreeSurface from "../src/game-ui/ReputationTreeSurface.svelte";
import { installTheme, UI_THEMES } from "../src/ui/themes";

const browser = typeof document !== "undefined";
const node = (id: string, suffix: string, kind: "bonus_unlock" | "starter", cost: number, requires: string[], state: string) =>
  ({ body_key: `reputation_tree.node.${suffix}.body`, cost, kind, node_id: id, requires, state, title_key: `reputation_tree.node.${suffix}.title` });
const arm = {
  available: 3, bonus_factor_next_run: "1.002e0", bonus_factor_this_run: "1e0", level: 4, per_level_ppm: 10_000, spent: 1, unlock_ppm: 50_000,
  nodes: [
    node("reputation.unlock.p05", "unlock_p05", "bonus_unlock", 1, [], "owned"),
    node("reputation.starter.cash_small", "starter_cash_small", "starter", 2, [], "available"),
    node("reputation.starter.generated_beige_tower", "starter_generated_beige_tower", "starter", 3, ["reputation.starter.cash_small"], "locked"),
    node("reputation.unlock.p25", "unlock_p25", "bonus_unlock", 5, ["reputation.unlock.p05"], "unaffordable"),
  ],
} as GameUIReputationArm;

async function settle(): Promise<void> { for (let index = 0; index < 4; index += 1) { await tick(); flushSync(); } }

it.skipIf(!browser)("renders server-derived states and buys only through an explicit confirm", async () => {
  const target = document.createElement("main");
  document.body.append(target);
  installTheme(target, UI_THEMES.era_1995, false);
  const purchases: string[] = [];
  const app = mount(ReputationTreeSurface, { target, props: { arm, era: "era_1995", pending: false, controlsEnabled: true, onPurchase: (id: string) => purchases.push(id) } });
  try {
    await settle();
    const axeResult = await axe.run(target, { runOnly: { type: "tag", values: ["wcag2a", "wcag2aa", "wcag21aa", "wcag22aa"] } });
    expect(axeResult.violations.filter((row) => row.impact === "serious" || row.impact === "critical")).toEqual([]);
    for (const id of arm.nodes.map((row) => row.node_id)) expect(target.textContent).not.toContain(id);
    const cards = [...target.querySelectorAll("li")];
    expect(cards.map((card) => card.dataset.state)).toEqual(["owned", "available", "locked", "unaffordable"]);
    // Only the available node has a control; owned/locked/unaffordable state is text.
    expect(cards.map((card) => card.querySelectorAll("button").length)).toEqual([0, 1, 0, 0]);
    const buy = cards[1]!.querySelector("button")!;
    buy.click();
    await settle();
    expect(purchases).toEqual([]);
    const confirm = cards[1]!.querySelector("button")!;
    expect(document.activeElement).toBe(confirm);
    confirm.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    await settle();
    expect(purchases).toEqual([]);
    expect(cards[1]!.querySelectorAll("button")).toHaveLength(1);
    expect(cards[1]!.querySelector("[role=group]")).toBeNull();
    expect(document.activeElement).toBe(cards[1]!.querySelector("button"));
    cards[1]!.querySelector("button")!.click();
    await settle();
    cards[1]!.querySelector("button")!.click();
    await settle();
    expect(purchases).toEqual(["reputation.starter.cash_small"]);
  } finally { unmount(app); target.remove(); }
});

it.skipIf(!browser)("disables buying while pending or when Founder controls are unavailable", async () => {
  const target = document.createElement("main");
  document.body.append(target);
  const app = mount(ReputationTreeSurface, { target, props: { arm, era: "era_1995", pending: false, controlsEnabled: false, onPurchase: () => {} } });
  try {
    await settle();
    expect(target.querySelector("li[data-state=available] button")!.hasAttribute("disabled")).toBe(true);
  } finally { unmount(app); target.remove(); }
});

it.skipIf(!browser)("plans in tree order, gates prerequisites and budget, and cascades deselection", async () => {
  const { default: ReputationPlanPanel } = await import("../src/game-ui/ReputationPlanPanel.svelte");
  const target = document.createElement("main");
  document.body.append(target);
  const state = { selected: [] as string[] };
  const app = mount(ReputationPlanPanel, { target, props: { arm, era: "era_1995", previewDelta: 2, onChange: (plan: readonly string[]) => { state.selected = [...plan]; } } });
  try {
    await settle();
    target.querySelector("details")!.open = true;
    const boxes = () => [...target.querySelectorAll<HTMLInputElement>("input[type=checkbox]")];
    // Unowned nodes only: cash_small, generated_beige_tower, p25. Budget 3 + 2 = 5.
    expect(boxes().map((box) => box.disabled)).toEqual([false, true, false]);
    boxes()[2]!.click(); await settle();
    expect(state.selected).toEqual(["reputation.unlock.p25"]);
    expect(boxes().map((box) => box.disabled)).toEqual([true, true, false]);
    boxes()[2]!.click(); await settle();
    boxes()[0]!.click(); await settle();
    expect(boxes()[1]!.disabled).toBe(false);
    boxes()[1]!.click(); await settle();
    expect(state.selected).toEqual(["reputation.starter.cash_small", "reputation.starter.generated_beige_tower"]);
    boxes()[0]!.click(); await settle();
    expect(state.selected).toEqual([]);
    const result = await axe.run(target, { runOnly: { type: "tag", values: ["wcag2a", "wcag2aa", "wcag21aa", "wcag22aa"] } });
    expect(result.violations.filter((row) => row.impact === "serious" || row.impact === "critical")).toEqual([]);
  } finally { unmount(app); target.remove(); }
});
