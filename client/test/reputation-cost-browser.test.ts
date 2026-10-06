import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import type { GameUIReputationArm } from "../src/api/generated/types";
import { t } from "../src/copy";
import ReputationTreeSurface from "../src/game-ui/ReputationTreeSurface.svelte";
import { installTheme, UI_THEMES } from "../src/ui/themes";
import ReputationFocusHarness from "./fixtures/ReputationFocusHarness.svelte";

// R9 rendering diagnostic: controlled coherent props, not a host/SQL/mint proof.
// Large costs deliberately exercise published notation, not adopted balance data.
const browser = typeof document !== "undefined";
const populations = [
  { name: "small", costs: [1, 2, 3, 5], formatted: ["1", "2", "3", "5"], available: 3, nextBonus: "1.002e0" },
  { name: "notation boundaries", costs: [999, 1000, 12345, 999950], formatted: ["999", "1.00 K", "12.3 K", "1.00 M"], available: 5000, nextBonus: "3.9995e0" },
] as const;
type Population = typeof populations[number];
function fixture(population: Population): GameUIReputationArm {
  const node = (id: string, suffix: string, kind: "bonus_unlock" | "starter", index: number, requires: string[], state: "owned" | "available" | "locked" | "unaffordable") =>
    ({ node_id: id, title_key: `reputation_tree.node.${suffix}.title`, body_key: `reputation_tree.node.${suffix}.body`, kind, cost: population.costs[index]!, requires, state });
  return {
    available: population.available, level: population.available + population.costs[0], spent: population.costs[0],
    bonus_factor_next_run: population.nextBonus, bonus_factor_this_run: "1e0", per_level_ppm: 10_000, unlock_ppm: 50_000,
    nodes: [
      node("reputation.unlock.p05", "unlock_p05", "bonus_unlock", 0, [], "owned"),
      node("reputation.starter.cash_small", "starter_cash_small", "starter", 1, [], "available"),
      node("reputation.starter.generated_beige_tower", "starter_generated_beige_tower", "starter", 2, ["reputation.starter.cash_small"], "locked"),
      node("reputation.unlock.p25", "unlock_p25", "bonus_unlock", 3, ["reputation.unlock.p05"], "unaffordable"),
    ],
  };
}
async function settle(): Promise<void> { for (let index = 0; index < 4; index += 1) { await tick(); flushSync(); } }
function rowCosts(target: HTMLElement, population: Population): HTMLLIElement[] {
  const rows = [...target.querySelectorAll<HTMLLIElement>(".reputation li")];
  expect(rows).toHaveLength(4);
  for (const [index, row] of rows.entries()) {
    // Exactly one actual Amount output per row, regardless of state/control.
    const amounts = row.querySelectorAll<HTMLOutputElement>(".cc-amount > output");
    expect(amounts, `persistent Amount cost missing or duplicated on row ${index}`).toHaveLength(1);
    const amount = amounts[0]!;
    expect(amount.textContent).toBe(population.formatted[index]);
    expect(amount.closest("[hidden], [aria-hidden=true]")).toBeNull();
    const style = getComputedStyle(amount);
    expect(style.display).not.toBe("none"); expect(style.visibility).toBe("visible");
    expect(amount.getBoundingClientRect().width).toBeGreaterThan(0);
    expect(amount.getBoundingClientRect().height).toBeGreaterThan(0);
    // R9 reading order: title/body/cost/requirements/state/control.
    const body = row.querySelector("p")!;
    const state = row.querySelector(".state")!;
    expect(body.compareDocumentPosition(amount) & Node.DOCUMENT_POSITION_FOLLOWING).not.toBe(0);
    expect(amount.compareDocumentPosition(state) & Node.DOCUMENT_POSITION_FOLLOWING).not.toBe(0);
    if (index >= 2) {
      const requirement = [...row.querySelectorAll("p")].find((element) => !element.classList.contains("state") && element !== body && element.getAttribute("role") !== "status")!;
      expect(requirement, "requirement title remains a separate row element").toBeTruthy();
      expect(amount.compareDocumentPosition(requirement) & Node.DOCUMENT_POSITION_FOLLOWING).not.toBe(0);
    }
  }
  return rows;
}
function noMechanicalIDs(target: HTMLElement, arm: GameUIReputationArm): void {
  for (const node of arm.nodes) expect(target.textContent).not.toContain(node.node_id);
}

for (const era of ["era_1995", "era_2000"] as const) {
  for (const population of populations) {
    it.skipIf(!browser)(`R9 persistent Amount costs in all four states: ${era}, ${population.name}`, async () => {
      const arm = fixture(population);
      const target = document.createElement("main"); document.body.append(target);
      installTheme(target, UI_THEMES[era], false);
      const purchases: string[] = [];
      const app = mount(ReputationTreeSurface, { target, props: {
        arm, era, pending: false, controlsEnabled: true,
        onPurchase: (id: string) => { purchases.push(id); },
      } });
      try {
        await settle();
        const rows = rowCosts(target, population);
        expect(rows.map((row) => row.dataset.state)).toEqual(["owned", "available", "locked", "unaffordable"]);
        expect(rows.map((row) => row.querySelectorAll("button").length)).toEqual([0, 1, 0, 0]);
        expect(rows[1]!.querySelector("button")!.textContent).toBe(t("reputation_tree.action.buy", { cost: population.costs[1] }, era));
        expect(purchases).toEqual([]); noMechanicalIDs(target, arm);
      } finally { await unmount(app); target.remove(); }
    });
    for (const key of ["{Enter}", " "]) {
      for (const outcome of ["owned", "available"] as const) {
        it.skipIf(!browser)(`R9 Amount costs persist through native confirmation/purchase: ${era}, ${population.name}, ${key === " " ? "Space" : "Enter"}, ${outcome}`, async () => {
          const { userEvent } = await import("vitest/browser");
          const initial = fixture(population);
          const target = document.createElement("main"); document.body.append(target);
          installTheme(target, UI_THEMES[era], false);
          const purchases: string[] = [];
          let finish: (() => void) | undefined;
          const app = mount(ReputationFocusHarness, { target, props: {
            initialArm: initial, era,
            onPurchase: (id: string) => {
              purchases.push(id); app.setPending(true);
              return new Promise<void>((resolve) => { finish = resolve; });
            },
          } });
          try {
            await settle();
            const row = rowCosts(target, population)[1]!;
            const buy = row.querySelector<HTMLButtonElement>("button")!;
            buy.focus(); await userEvent.keyboard(key); await settle();
            rowCosts(target, population);
            expect(row.querySelectorAll("button")).toHaveLength(2);
            expect(document.activeElement).toBe(row.querySelector("button"));
            expect(purchases).toEqual([]);
            await userEvent.keyboard("{Escape}"); await settle();
            rowCosts(target, population); expect(document.activeElement).toBe(row.querySelector("button"));
            row.querySelector<HTMLButtonElement>("button")!.focus();
            await userEvent.keyboard(key); await settle(); await userEvent.keyboard(key); await settle();
            rowCosts(target, population);
            expect(purchases).toEqual(["reputation.starter.cash_small"]);
            expect(row.getAttribute("aria-busy")).toBe("true");
            expect(row.dataset.state).toBe("available"); // No optimistic ownership.
            expect(document.activeElement).toBe(row);
            expect(row.querySelector<HTMLButtonElement>("button")!.disabled).toBe(true);
            const next = structuredClone(initial);
            next.nodes[1]!.state = outcome;
            if (outcome === "owned") {
              next.spent += population.costs[1]; next.available -= population.costs[1];
              next.nodes[2]!.state = "unaffordable";
            }
            app.deliverArm(next); await settle();
            expect(rowCosts(target, population)[1]).toBe(row);
            expect(row.dataset.state).toBe(outcome);
            expect(row.querySelectorAll("button")).toHaveLength(outcome === "owned" ? 0 : 1);
            expect(row.getAttribute("aria-busy")).toBe("true");
            expect(typeof finish).toBe("function"); finish!(); await settle();
            rowCosts(target, population);
            expect(row.hasAttribute("aria-busy")).toBe(false);
            expect(document.activeElement).toBe(row);
            expect(purchases).toHaveLength(1); noMechanicalIDs(target, initial);
          } finally { finish?.(); await settle(); await unmount(app); target.remove(); }
        });
      }
    }
  }
}
