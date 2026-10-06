import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import declaration from "../../balance/testdata/reputation-tree/fixture-v1.json";
import type { GameUIReputationArm } from "../src/api/generated/types";
import { t, type CopyEra } from "../src/copy";
import { installTheme, UI_THEMES } from "../src/ui/themes";
import ReputationPlanHarness from "./fixtures/ReputationPlanHarness.svelte";

// Controlled public props, not a server payout, adopted epoch or host/SQL proof.
// Invalid-selection cases characterize RP-292; they do NOT approve its policy.
const browser = typeof document !== "undefined";
const [unlock, cash, , unlock25] = declaration.nodes.map((node) => node.node_id);
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
function fixture(era: CopyEra, available: number, preview = 0) {
  const target = document.createElement("main"); document.body.append(target);
  installTheme(target, UI_THEMES[era], false);
  let arm = profile(available);
  const plans: string[][] = [];
  const app = mount(ReputationPlanHarness, { target, props: {
    initialArm: arm, initialPreview: preview, era,
    onChange: (value: readonly string[]) => { plans.push([...value]); },
  } });
  const boxes = () => [...target.querySelectorAll<HTMLInputElement>("input[type=checkbox]")];
  const box = (id: string) => {
    const index = arm.nodes.filter((node) => node.state !== "owned").findIndex((node) => node.node_id === id);
    expect(index).toBeGreaterThanOrEqual(0); return boxes()[index]!;
  };
  return {
    target, plans, boxes, box,
    async deliverArm(value: GameUIReputationArm) { arm = value; app.deliverArm(value); await settle(); },
    async deliverPreview(value: number) { app.deliverPreview(value); await settle(); },
    async open() {
      await settle(); expect(plans).toEqual([[]]);
      const details = target.querySelector<HTMLDetailsElement>("details")!;
      expect(details.open).toBe(false);
      const { userEvent } = await import("vitest/browser");
      details.querySelector<HTMLElement>("summary")!.focus(); await userEvent.keyboard("{Enter}");
      await settle(); expect(details.open).toBe(true);
    },
    async choose(id: string) {
      expect(box(id).disabled).toBe(false);
      const { userEvent } = await import("vitest/browser");
      box(id).focus(); await userEvent.keyboard(" "); await settle();
    },
    async clear() {
      const button = target.querySelector<HTMLButtonElement>("button")!;
      expect(button.disabled).toBe(false);
      const { userEvent } = await import("vitest/browser");
      button.focus(); await userEvent.keyboard("{Enter}"); await settle();
    },
    async dispose() { await unmount(app); target.remove(); },
  };
}
function projected(b: ReturnType<typeof fixture>, era: CopyEra, amount: number): void {
  expect(b.target.querySelector("[role=status]")!.textContent).toBe(t("reputation_tree.plan.projected", { amount }, era));
  for (const node of declaration.nodes) expect(b.target.textContent).not.toContain(node.node_id);
}

for (const era of ["era_1995", "era_2000"] as const) {
  it.skipIf(!browser)(`R9 mounted plan separately reacts to available and preview replacement: ${era}`, async () => {
    const b = fixture(era, 0);
    try {
      await b.open(); projected(b, era, 0); expect(b.boxes().every((box) => box.disabled)).toBe(true);
      await b.deliverPreview(3); projected(b, era, 3);
      expect(b.box(unlock!).disabled).toBe(false); expect(b.box(cash!).disabled).toBe(false);
      await b.deliverPreview(0); projected(b, era, 0); expect(b.boxes().every((box) => box.disabled)).toBe(true);
      await b.deliverArm(profile(1)); projected(b, era, 1);
      expect(b.box(unlock!).disabled).toBe(false); expect(b.box(cash!).disabled).toBe(true);
      await b.deliverArm(profile(3)); projected(b, era, 3);
      expect(b.box(cash!).disabled).toBe(false); expect(b.plans).toEqual([[]]);
      await b.choose(cash!); projected(b, era, 1); expect(b.plans).toEqual([[], [cash!]]);
    } finally { await b.dispose(); }
  });

  it.skipIf(!browser)(`R9 mounted plan uses newly owned unselected prerequisite: ${era}`, async () => {
    const b = fixture(era, 10);
    try {
      await b.open(); projected(b, era, 10);
      expect(b.boxes()).toHaveLength(9); expect(b.box(unlock25!).disabled).toBe(true);
      await b.deliverArm(profile(9, true)); projected(b, era, 9);
      expect(b.boxes()).toHaveLength(8); expect(b.box(unlock25!).disabled).toBe(false);
      expect(b.plans).toEqual([[]]);
      await b.choose(unlock25!); projected(b, era, 4); expect(b.plans).toEqual([[], [unlock25!]]);
      await b.deliverArm(profile(12, true)); projected(b, era, 7);
      expect(b.box(unlock25!).checked).toBe(true); expect(b.plans).toEqual([[], [unlock25!]]);
    } finally { await b.dispose(); }
  });

  it.skipIf(!browser)(`R9 mounted valid selection stays reflected after budget increase: ${era}`, async () => {
    const b = fixture(era, 3);
    try {
      await b.open(); await b.choose(cash!); projected(b, era, 1);
      await b.deliverArm(profile(10)); projected(b, era, 8);
      await b.deliverPreview(1); projected(b, era, 9);
      expect(b.box(cash!).checked).toBe(true); expect(b.plans).toEqual([[], [cash!]]);
      await b.clear(); projected(b, era, 11); expect(b.plans).toEqual([[], [cash!], []]);
      expect(b.boxes().every((box) => !box.checked)).toBe(true);
    } finally { await b.dispose(); }
  });

  it.skipIf(!browser)(`RP-292 characterization ONLY: newly owned selection stays hidden in callback until native Clear: ${era}`, async () => {
    const b = fixture(era, 3);
    try {
      await b.open(); await b.choose(unlock!); projected(b, era, 2);
      await b.deliverArm(profile(2, true));
      expect(b.boxes()).toHaveLength(8); expect(b.boxes().every((box) => !box.checked)).toBe(true);
      expect(b.plans).toEqual([[], [unlock!]]);
      // Current formula still subtracts the hidden selected owned node.
      // This is measured existing behavior, NOT the desired reset/prune rule.
      projected(b, era, 1);
      await b.clear(); projected(b, era, 2); expect(b.plans).toEqual([[], [unlock!], []]);
    } finally { await b.dispose(); }
  });

  it.skipIf(!browser)(`RP-292 characterization ONLY: budget loss retains over-budget checked plan until native Clear: ${era}`, async () => {
    const b = fixture(era, 3);
    try {
      await b.open(); await b.choose(cash!); await b.choose(unlock!); projected(b, era, 0);
      await b.deliverArm(profile(0)); projected(b, era, -3);
      expect(b.box(unlock!).checked).toBe(true); expect(b.box(cash!).checked).toBe(true);
      expect(b.box(unlock!).disabled).toBe(false); expect(b.box(cash!).disabled).toBe(false);
      expect(b.plans).toEqual([[], [cash!], [unlock!, cash!]]);
      await b.clear(); projected(b, era, 0); expect(b.plans.at(-1)).toEqual([]);
      expect(b.boxes().every((box) => box.disabled && !box.checked)).toBe(true);
    } finally { await b.dispose(); }
  });
}
