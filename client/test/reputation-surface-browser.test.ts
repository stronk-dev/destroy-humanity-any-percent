import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import type { GameUIReputationArm } from "../src/api/generated/types";
import { t, type CopyKey } from "../src/copy";
import ReputationTreeSurface from "../src/game-ui/ReputationTreeSurface.svelte";
import ReputationFocusHarness from "./fixtures/ReputationFocusHarness.svelte";
import { installTheme, UI_THEMES } from "../src/ui/themes";

const browser = typeof document !== "undefined";
const node = (id: string, suffix: string, kind: "bonus_unlock" | "starter", cost: number, requires: string[], state: string) =>
  ({ body_key: `reputation_tree.node.${suffix}.body`, cost, kind, node_id: id, requires, state, title_key: `reputation_tree.node.${suffix}.title` });
const arm = {
  tree_active: true,
  available: 3, bonus_factor_next_run: "1.002e0", bonus_factor_this_run: "1e0", level: 4, per_level_ppm: 10_000, spent: 1, unlock_ppm: 50_000,
  nodes: [
    node("reputation.unlock.p05", "unlock_p05", "bonus_unlock", 1, [], "owned"),
    node("reputation.starter.cash_small", "starter_cash_small", "starter", 2, [], "available"),
    node("reputation.starter.generated_beige_tower", "starter_generated_beige_tower", "starter", 3, ["reputation.starter.cash_small"], "locked"),
    node("reputation.unlock.p25", "unlock_p25", "bonus_unlock", 5, ["reputation.unlock.p05"], "unaffordable"),
  ],
} as GameUIReputationArm;

for (const era of ["era_1995", "era_2000"] as const) {
  for (const key of ["{Enter}", " "]) {
    for (const outcome of ["owned", "available"] as const) {
      it.skipIf(!browser)(`retains Reputation purchase row focus after native ${key === " " ? "Space" : "Enter"} and ${outcome} in ${era}`, async () => {
        const { userEvent } = await import("vitest/browser");
        const target = document.createElement("main");
        document.body.append(target);
        installTheme(target, UI_THEMES[era], false);
        const purchases: string[] = [];
        const app = mount(ReputationFocusHarness, { target, props: {
          initialArm: structuredClone(arm), era,
          onPurchase: (id: string) => { purchases.push(id); app.setPending(true); },
        } });
        try {
          await settle();
          const row = target.querySelector<HTMLLIElement>("li[data-state=available]")!;
          const buy = row.querySelector<HTMLButtonElement>("button")!;
          buy.focus();
          await userEvent.keyboard(key);
          await settle();
          expect(purchases).toEqual([]);
          expect(document.activeElement).toBe(row.querySelector("button"));
          await userEvent.keyboard(key);
          await settle();
          expect(purchases).toEqual(["reputation.starter.cash_small"]);
          const heldButtons = [...target.querySelectorAll<HTMLButtonElement>("li button")];
          expect(heldButtons).toHaveLength(1);
          expect(heldButtons.every((button) => button.disabled)).toBe(true);
          expect(document.activeElement, "confirmed control must not strand focus on body while pending").toBe(row);
          const authoritative = structuredClone(arm);
          authoritative.nodes[1]!.state = outcome;
          if (outcome === "owned") {
            authoritative.spent += 2; authoritative.available -= 2;
            authoritative.nodes[2]!.state = "unaffordable";
          }
          app.deliverArm(authoritative);
          await settle();
          expect(target.querySelectorAll("li")[1]).toBe(row);
          expect(row.dataset.state).toBe(outcome);
          expect(row.querySelectorAll("button")).toHaveLength(outcome === "owned" ? 0 : 1);
          expect(document.activeElement, "focus remains on the exact row after authoritative replacement").toBe(row);
          expect(purchases).toHaveLength(1);
          for (const id of arm.nodes.map((node) => node.node_id)) expect(target.textContent).not.toContain(id);
        } finally { unmount(app); target.remove(); }
      });
    }
  }
  it.skipIf(!browser)(`cancels Reputation confirmation through native Escape in ${era}`, async () => {
    const { userEvent } = await import("vitest/browser");
    const target = document.createElement("main"); document.body.append(target);
    const purchases: string[] = [];
    const app = mount(ReputationTreeSurface, { target, props: { arm: structuredClone(arm), era, pending: false, controlsEnabled: true, onPurchase: (id: string) => { purchases.push(id); } } });
    try {
      await settle();
      const row = target.querySelector("li[data-state=available]")!;
      const buy = row.querySelector<HTMLButtonElement>("button")!;
      buy.focus(); await userEvent.keyboard("{Enter}"); await settle();
      expect(document.activeElement).toBe(row.querySelector("button"));
      await userEvent.keyboard("{Escape}"); await settle();
      expect(row.querySelectorAll("button")).toHaveLength(1);
      expect(row.querySelector("[role=group]")).toBeNull();
      expect(document.activeElement).toBe(row.querySelector("button"));
      expect(purchases).toEqual([]);
    } finally { unmount(app); target.remove(); }
  });
}

async function settle(): Promise<void> { for (let index = 0; index < 4; index += 1) { await tick(); flushSync(); } }

it.skipIf(!browser)("R9 native Tab diagnostic sentinel control", async () => {
  const { userEvent } = await import("vitest/browser");
  const wrapper = document.createElement("main"); document.body.append(wrapper);
  const before = document.createElement("button"); before.textContent = "Before diagnostic"; before.tabIndex = 0;
  const after = document.createElement("button"); after.textContent = "After diagnostic"; after.tabIndex = 0;
  wrapper.append(before, after);
  try {
    before.focus(); expect(document.activeElement).toBe(before);
    await userEvent.keyboard("{Tab}"); await settle(); expect(document.activeElement).toBe(after);
    await userEvent.keyboard("{Shift>}{Tab}{/Shift}"); await settle(); expect(document.activeElement).toBe(before);
  } finally { wrapper.remove(); }
});

function keyboardArm(population: "mixed" | "two-buyable" | "none-buyable"): GameUIReputationArm {
  // Exact declared arm fields: the legacy fixture's tree_active metadata is
  // not part of this component diagnostic's wire-shaped population.
  const value: GameUIReputationArm = {
    available: arm.available, bonus_factor_next_run: arm.bonus_factor_next_run,
    bonus_factor_this_run: arm.bonus_factor_this_run, level: arm.level,
    per_level_ppm: arm.per_level_ppm, spent: arm.spent, unlock_ppm: arm.unlock_ppm,
    nodes: structuredClone(arm.nodes),
  };
  if (population === "two-buyable") {
    value.available = 8; value.level = 9; value.bonus_factor_next_run = "1.0045e0";
    value.nodes[3]!.state = "available";
  } else if (population === "none-buyable") {
    value.available = 0; value.level = 1; value.bonus_factor_next_run = "1e0";
    value.nodes[1]!.state = "unaffordable";
  }
  return value;
}

for (const era of ["era_1995", "era_2000"] as const) {
  for (const population of ["mixed", "two-buyable", "none-buyable"] as const) {
    for (const controls of ["ready", "pending", "offline"] as const) {
      it.skipIf(!browser)(`R9 native Tab visits header and enabled rows in order: ${era}, ${population}, ${controls}`, async () => {
        const { userEvent } = await import("vitest/browser");
        const profile = keyboardArm(population);
        const wrapper = document.createElement("main"); document.body.append(wrapper);
        const before = document.createElement("button"); before.textContent = "Before diagnostic"; before.tabIndex = 0;
        const target = document.createElement("div");
        const after = document.createElement("button"); after.textContent = "After diagnostic"; after.tabIndex = 0;
        wrapper.append(before, target, after); installTheme(wrapper, UI_THEMES[era], false);
        const purchases: string[] = [];
        const app = mount(ReputationTreeSurface, { target, props: {
          arm: profile, era, pending: controls === "pending", controlsEnabled: controls !== "offline",
          onPurchase: (id: string) => { purchases.push(id); },
        } });
        try {
          await settle();
          const heading = target.querySelector<HTMLHeadingElement>("#reputation-heading")!;
          expect(heading).toBeTruthy();
          const rows = [...target.querySelectorAll<HTMLLIElement>(".reputation li")];
          expect(rows.map((row) => row.querySelector("h2")!.textContent)).toEqual(profile.nodes.map((node) => t(node.title_key as CopyKey, {}, era)));
          expect(rows.map((row) => row.dataset.state)).toEqual(profile.nodes.map((node) => node.state));
          expect(rows.every((row) => row.tabIndex === -1)).toBe(true);
          const enabled: HTMLButtonElement[] = [];
          for (const [index, node] of profile.nodes.entries()) {
            const buttons = [...rows[index]!.querySelectorAll<HTMLButtonElement>("button")];
            expect(buttons).toHaveLength(node.state === "available" ? 1 : 0);
            if (node.state === "available") {
              expect(buttons[0]!.disabled).toBe(controls !== "ready");
              if (controls === "ready") enabled.push(buttons[0]!);
            }
          }
          // The browser decides focus. Never programmatically focus the
          // header or row controls, which would mask a missing Tab stop.
          before.focus(); expect(document.activeElement).toBe(before);
          for (const expected of [heading, ...enabled, after]) {
            await userEvent.keyboard("{Tab}"); await settle();
            expect(document.activeElement, "R9 forward native Tab order").toBe(expected);
          }
          for (const expected of [...enabled].reverse()) {
            await userEvent.keyboard("{Shift>}{Tab}{/Shift}"); await settle();
            expect(document.activeElement, "R9 reverse native row order").toBe(expected);
          }
          await userEvent.keyboard("{Shift>}{Tab}{/Shift}"); await settle();
          expect(document.activeElement, "R9 reverse Tab must reach header").toBe(heading);
          await userEvent.keyboard("{Shift>}{Tab}{/Shift}"); await settle();
          expect(document.activeElement).toBe(before);
          expect(purchases).toEqual([]); expect(target.querySelector("[role=group]")).toBeNull();
        } finally { await unmount(app); wrapper.remove(); }
      });
    }
  }
}

for (const era of ["era_1995", "era_2000"] as const) {
  for (const key of ["{Enter}", " "]) {
    for (const pendingStart of ["synchronous", "delayed"] as const) {
      for (const outcome of ["owned", "available"] as const) {
        it.skipIf(!browser)(`attributes busy to only the submitted Reputation row: ${era}, ${key === " " ? "Space" : "Enter"}, ${pendingStart}, ${outcome}`, async () => {
          const { userEvent } = await import("vitest/browser");
          const initial = structuredClone(arm);
          initial.level = 9; initial.available = 8; initial.bonus_factor_next_run = "1.0045e0";
          initial.nodes[3]!.state = "available";
          const target = document.createElement("main"); document.body.append(target);
          installTheme(target, UI_THEMES[era], false);
          const purchases: string[] = [];
          let finish: (() => void) | undefined;
          const app = mount(ReputationFocusHarness, { target, props: {
            initialArm: initial, era,
            onPurchase: (id: string) => {
              purchases.push(id);
              if (pendingStart === "synchronous") app.setPending(true);
              return new Promise<void>((resolve) => { finish = resolve; });
            },
          } });
          const busyRows = () => [...target.querySelectorAll("li[aria-busy=true]")];
          const buttons = () => [...target.querySelectorAll<HTMLButtonElement>("li button")];
          try {
            await settle();
            expect(buttons()).toHaveLength(2);
            expect(busyRows()).toEqual([]);
            // An unrelated host refresh disables controls but owns no purchase row.
            app.setPending(true); await settle();
            expect(busyRows()).toEqual([]);
            expect(buttons().every((button) => button.disabled)).toBe(true);
            app.setPending(false); await settle();
            const row = target.querySelectorAll<HTMLLIElement>("li")[1]!;
            row.querySelector<HTMLButtonElement>("button")!.focus();
            await userEvent.keyboard(key); await settle();
            expect(busyRows()).toEqual([]);
            expect(purchases).toEqual([]);
            await userEvent.keyboard(key); await settle();
            expect(purchases).toEqual(["reputation.starter.cash_small"]);
            expect(busyRows(), "the submitted row owns the held purchase task").toEqual([row]);
            expect(buttons()).toHaveLength(2);
            expect(buttons().every((button) => button.disabled)).toBe(true);
            expect(document.activeElement).toBe(row);
            if (pendingStart === "delayed") { app.setPending(true); await settle(); }
            expect(busyRows()).toEqual([row]);
            const next = structuredClone(initial);
            next.nodes[1]!.state = outcome;
            if (outcome === "owned") {
              next.spent += 2; next.available -= 2; next.nodes[2]!.state = "available";
            }
            app.deliverArm(next); await settle();
            // Even if parent pending clears first, the returned task is still held.
            expect(busyRows()).toEqual([row]);
            expect(buttons()).toHaveLength(2);
            expect(buttons().every((button) => button.disabled)).toBe(true);
            expect(typeof finish).toBe("function"); finish!(); await settle();
            expect(busyRows()).toEqual([]);
            expect(buttons().every((button) => !button.disabled)).toBe(true);
            expect(document.activeElement).toBe(row);
            app.setPending(true); await settle();
            expect(busyRows(), "unrelated pending must not revive the completed row").toEqual([]);
            app.setPending(false); await settle();
            // A later purchase must attribute to its own row, not the previous one.
            const second = target.querySelectorAll<HTMLLIElement>("li")[3]!;
            second.querySelector<HTMLButtonElement>("button")!.focus();
            await userEvent.keyboard(key); await settle();
            await userEvent.keyboard(key); await settle();
            expect(purchases).toEqual(["reputation.starter.cash_small", "reputation.unlock.p25"]);
            expect(busyRows()).toEqual([second]);
            expect(document.activeElement).toBe(second);
            app.setPending(true); finish!(); await settle();
            // Conversely, a settled task cannot clear the parent's held refresh.
            expect(busyRows()).toEqual([second]);
            expect(buttons()).toHaveLength(2);
            expect(buttons().every((button) => button.disabled)).toBe(true);
            app.deliverArm(structuredClone(next)); await settle();
            expect(busyRows()).toEqual([]);
            expect(purchases).toHaveLength(2);
          } finally { finish?.(); await unmount(app); target.remove(); }
        });
      }
    }
  }
}

it.skipIf(!browser)("renders server-derived states and buys only through an explicit confirm", async () => {
  const target = document.createElement("main");
  document.body.append(target);
  installTheme(target, UI_THEMES.era_1995, false);
  const purchases: string[] = [];
  const app = mount(ReputationTreeSurface, { target, props: { arm, era: "era_1995", pending: false, controlsEnabled: true, onPurchase: (id: string) => { purchases.push(id); } } });
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
