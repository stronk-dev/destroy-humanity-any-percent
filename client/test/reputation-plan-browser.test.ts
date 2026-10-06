import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import declaration from "../../balance/testdata/reputation-tree/fixture-v1.json";
import type { GameUIReputationArm } from "../src/api/generated/types";
import { t, type CopyEra, type CopyKey } from "../src/copy";
import ReputationPlanPanel from "../src/game-ui/ReputationPlanPanel.svelte";
import { installTheme, UI_THEMES } from "../src/ui/themes";

// Full declared proposal, not an adopted epoch or a live-server payout. The
// component receives an explicit authoritative-preview diagnostic input.
const browser = typeof document !== "undefined";
const ids = declaration.nodes.map((node) => node.node_id);
const [unlock, cash, tower, unlock25, upgrade, large] = ids;
function profile(available: number, ownsFirst = false): GameUIReputationArm {
  return {
    available, level: available + (ownsFirst ? 1 : 0), spent: ownsFirst ? 1 : 0,
    bonus_factor_this_run: null, bonus_factor_next_run: ownsFirst ? "1.0065e0" : "1e0",
    per_level_ppm: declaration.bonus.per_level_ppm, unlock_ppm: ownsFirst ? 50_000 : 0,
    nodes: declaration.nodes.map((node, index) => ({
      node_id: node.node_id, kind: node.kind, cost: node.cost, requires: [...node.requires],
      title_key: node.title_key, body_key: node.body_key,
      state: ownsFirst && index === 0 ? "owned" : node.requires.some((id) => !(ownsFirst && id === unlock))
        ? "locked" : node.cost > available ? "unaffordable" : "available",
    })) as GameUIReputationArm["nodes"],
  };
}
async function settle(): Promise<void> { await tick(); flushSync(); }
function fixture(era: CopyEra, available: number, preview: number, ownsFirst = false) {
  const wrapper = document.createElement("main"); document.body.append(wrapper);
  const before = document.createElement("button"); before.textContent = "Before plan diagnostic"; before.tabIndex = 0;
  const target = document.createElement("div");
  const after = document.createElement("button"); after.textContent = "After plan diagnostic"; after.tabIndex = 0;
  wrapper.append(before, target, after); installTheme(wrapper, UI_THEMES[era], false);
  const arm = profile(available, ownsFirst);
  const plans: string[][] = [];
  const app = mount(ReputationPlanPanel, { target, props: {
    arm, era, previewDelta: preview, onChange: (value: readonly string[]) => { plans.push([...value]); },
  } });
  const details = () => target.querySelector<HTMLDetailsElement>("details.plan")!;
  const boxes = () => [...target.querySelectorAll<HTMLInputElement>("input[type=checkbox]")];
  const visibleIDs = arm.nodes.filter((node) => node.state !== "owned").map((node) => node.node_id);
  const box = (id: string) => {
    const index = visibleIDs.indexOf(id); expect(index).toBeGreaterThanOrEqual(0); return boxes()[index]!;
  };
  return { wrapper, target, before, after, arm, plans, details, boxes, box,
    summary: () => details().querySelector<HTMLElement>("summary")!,
    clear: () => target.querySelector<HTMLButtonElement>("button")!,
    async dispose() { await unmount(app); wrapper.remove(); },
  };
}
function state(b: ReturnType<typeof fixture>, era: CopyEra, expected: readonly string[], budget: number) {
  expect(b.plans.at(-1)).toEqual(expected);
  const unowned = b.arm.nodes.filter((node) => node.state !== "owned");
  expect(b.boxes().map((box) => box.checked)).toEqual(unowned.map((node) => expected.includes(node.node_id)));
  expect(b.target.querySelector("[role=status]")!.textContent).toBe(t("reputation_tree.plan.projected", { amount: budget }, era));
  expect(b.clear().disabled).toBe(expected.length === 0);
  for (const id of ids) expect(b.target.textContent).not.toContain(id);
}
async function choose(b: ReturnType<typeof fixture>, id: string) {
  const { userEvent } = await import("vitest/browser");
  expect(b.box(id).disabled, `declared node ${id} is selectable`).toBe(false);
  b.box(id).focus(); await userEvent.keyboard(" "); await settle();
}
async function open(b: ReturnType<typeof fixture>, key = "{Enter}") {
  const { userEvent } = await import("vitest/browser");
  expect(b.details().open).toBe(false);
  b.summary().focus(); await userEvent.keyboard(key); await settle(); expect(b.details().open).toBe(true);
}

for (const era of ["era_1995", "era_2000"] as const) {
  it.skipIf(!browser)(`R9 full-tree plan defaults empty, renders registered rows and passes axe: ${era}`, async () => {
    const b = fixture(era, 3, 0);
    try {
      await settle();
      expect(declaration.nodes.map((node) => node.cost)).toEqual([1, 2, 3, 5, 4, 12, 25, 100, 400]);
      expect(b.details().open).toBe(false); expect(b.plans).toEqual([[]]); state(b, era, [], 3);
      expect(b.boxes()).toHaveLength(9);
      expect([...b.target.querySelectorAll("label")].map((label) => label.textContent?.trim())).toEqual(
        declaration.nodes.map((node) => `${t(node.title_key as CopyKey, {}, era)} ${t("reputation_tree.action.buy", { cost: node.cost }, era)}`));
      expect(b.boxes().map((box) => box.disabled)).toEqual([false, false, true, true, true, true, true, true, true]);
      await open(b);
      // Sentinels exist solely for the separate keyboard diagnostic. Axe
      // measures the actual panel, not our unstyled test-only buttons.
      b.before.remove(); b.after.remove();
      const result = await axe.run(b.wrapper, { runOnly: { type: "tag", values: ["wcag2a", "wcag2aa", "wcag21aa", "wcag22aa"] } });
      expect(result.violations).toEqual([]);
    } finally { await b.dispose(); }
  });

  for (const preview of [0, 2, 3]) {
    it.skipIf(!browser)(`R9 plan uses the supplied preview budget without enabling an over-budget node: ${era}, preview${preview}`, async () => {
      const b = fixture(era, 0, preview);
      try {
        await settle(); await open(b); state(b, era, [], preview);
        expect(b.boxes().map((box) => box.disabled)).toEqual([preview < 1, preview < 2, true, true, true, true, true, true, true]);
        if (preview === 0) { expect(b.plans).toEqual([[]]); return; }
        // Choose the independent rows in reverse order; the callback's order
        // must still be the declared topological order, not click order.
        await choose(b, cash!); state(b, era, [cash!], preview - 2);
        expect(b.box(unlock!).disabled).toBe(preview < 3);
        expect(b.box(tower!).disabled).toBe(true);
        if (preview === 3) { await choose(b, unlock!); state(b, era, [unlock!, cash!], 0); }
        const clear = b.clear(); clear.focus();
        const { userEvent } = await import("vitest/browser");
        await userEvent.keyboard("{Enter}"); await settle(); state(b, era, [], preview);
      } finally { await b.dispose(); }
    });
  }

  for (const removed of ["cash-root", "unlock-root"] as const) {
    it.skipIf(!browser)(`R9 full-tree plan drops transitive dependants, not the independent branch: ${era}, ${removed}`, async () => {
      const b = fixture(era, 12, 2, true);
      try {
        await settle(); await open(b); state(b, era, [], 14);
        expect(b.boxes()).toHaveLength(8); expect(b.target.textContent).not.toContain(t("reputation_tree.node.unlock_p05.title", {}, era));
        await choose(b, cash!); state(b, era, [cash!], 12);
        expect(b.box(tower!).disabled).toBe(false); expect(b.box(large!).disabled).toBe(true);
        await choose(b, tower!); state(b, era, [cash!, tower!], 9);
        await choose(b, upgrade!); state(b, era, [cash!, tower!, upgrade!], 5);
        await choose(b, unlock25!); state(b, era, [cash!, tower!, unlock25!, upgrade!], 0);
        expect(b.box(large!).disabled).toBe(true); expect(b.box(ids[6]!).disabled).toBe(true);
        await choose(b, removed === "cash-root" ? cash! : unlock25!);
        state(b, era, removed === "cash-root" ? [unlock25!] : [cash!, tower!, upgrade!], removed === "cash-root" ? 9 : 5);
        expect(b.box(large!).disabled).toBe(true);
        if (removed === "cash-root") {
          expect(b.box(tower!).disabled).toBe(true); expect(b.box(upgrade!).disabled).toBe(true);
        } else { expect(b.box(cash!).checked).toBe(true); expect(b.box(tower!).checked).toBe(true); expect(b.box(upgrade!).checked).toBe(true); }
      } finally { await b.dispose(); }
    });
  }

  it.skipIf(!browser)(`R9 all nine plan nodes fit the exact supplied budget and cascade across both prerequisite branches: ${era}`, async () => {
    const b = fixture(era, 1, 551);
    try {
      await settle(); await open(b); state(b, era, [], 552);
      const budgets = [551, 549, 546, 541, 537, 525, 500, 400, 0];
      for (const [index, id] of ids.entries()) { await choose(b, id); state(b, era, ids.slice(0, index + 1), budgets[index]!); }
      expect(b.plans).toEqual([[], ...ids.map((_id, index) => ids.slice(0, index + 1))]);
      await choose(b, unlock!); state(b, era, [cash!, tower!, upgrade!], 543);
      expect(b.boxes().map((box) => box.disabled)).toEqual([false, false, false, true, false, true, true, true, true]);
      const { userEvent } = await import("vitest/browser");
      b.clear().focus(); await userEvent.keyboard(" "); await settle(); state(b, era, [], 552);
      expect(b.boxes().map((box) => box.disabled)).toEqual([false, false, true, true, true, true, true, true, true]);
    } finally { await b.dispose(); }
  });

  for (const key of ["{Enter}", " "]) {
    it.skipIf(!browser)(`R9 native sequential plan path reaches disclosure, enabled boxes and Clear: ${era}, ${key === " " ? "Space" : "Enter"}`, async () => {
      const { userEvent } = await import("vitest/browser");
      const b = fixture(era, 3, 0);
      try {
        await settle(); state(b, era, [], 3);
        // Only sentinel gets direct focus. Browser Tab/Space must reach and
        // operate every plan control; forced focus would mask missing stops.
        b.before.focus(); expect(document.activeElement).toBe(b.before);
        await userEvent.keyboard("{Tab}"); await settle(); expect(document.activeElement).toBe(b.summary());
        await userEvent.keyboard(key); await settle(); expect(b.details().open).toBe(true);
        await userEvent.keyboard("{Tab}"); await settle(); expect(document.activeElement).toBe(b.box(unlock!));
        await userEvent.keyboard(" "); await settle(); state(b, era, [unlock!], 2);
        await userEvent.keyboard("{Tab}"); await settle(); expect(document.activeElement).toBe(b.box(cash!));
        await userEvent.keyboard(" "); await settle(); state(b, era, [unlock!, cash!], 0);
        await userEvent.keyboard("{Tab}"); await settle(); expect(document.activeElement).toBe(b.clear());
        await userEvent.keyboard("{Tab}"); await settle(); expect(document.activeElement).toBe(b.after);
        for (const expected of [b.clear(), b.box(cash!), b.box(unlock!), b.summary(), b.before]) {
          await userEvent.keyboard("{Shift>}{Tab}{/Shift}"); await settle(); expect(document.activeElement).toBe(expected);
        }
        for (const expected of [b.summary(), b.box(unlock!), b.box(cash!), b.clear()]) {
          await userEvent.keyboard("{Tab}"); await settle(); expect(document.activeElement).toBe(expected);
        }
        await userEvent.keyboard(key); await settle(); state(b, era, [], 3);
        expect(b.plans).toEqual([[], [unlock!], [unlock!, cash!], []]);
      } finally { await b.dispose(); }
    });
  }
}
