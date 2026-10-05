import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import fixture from "../../testdata/gameui/cosmetics-arm-v1.json";
import type { GameUICosmeticsArm } from "../src/api/generated/types";
import CosmeticOverlay from "../src/game-ui/cosmetics/CosmeticOverlay.svelte";
import { COSMETIC_SHOP_PRESENTATION } from "../src/game-ui/cosmetics/presentation";
import { installTheme, UI_THEMES } from "../src/ui/themes";
import CosmeticShelfHarness from "./CosmeticShelfHarness.svelte";

const browser = typeof document !== "undefined";
const arms = fixture.cases as unknown as Record<string, GameUICosmeticsArm>;
const CAT = "01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa";
const CURTAINS = ["curtain-horse_armor-paid_cosmetic_dlc", "curtain-horse_armor-reference_price_anchor", "curtain-shop-checkout_flow"];

async function settle(): Promise<void> { for (let index = 0; index < 4; index += 1) { await new Promise((resolve) => setTimeout(resolve, 0)); await tick(); flushSync(); } }

async function assertAxe(target: HTMLElement, label: string): Promise<void> {
  const result = await axe.run(target, { runOnly: { type: "tag", values: ["wcag2a", "wcag2aa", "wcag21aa", "wcag22aa"] } });
  expect(result.violations.filter((violation) => violation.impact === "serious" || violation.impact === "critical"), label).toEqual([]);
}

function host(width?: number): HTMLElement {
  const target = document.createElement("main");
  if (width) target.style.inlineSize = `${width}px`;
  document.body.append(target);
  installTheme(target, UI_THEMES.era_2000, false);
  return target;
}

// AC12 render half: every bound curtain is present, visible without hover,
// in this state.
function assertCurtainsVisible(target: HTMLElement, label: string): void {
  for (const id of CURTAINS) {
    const node = target.querySelector<HTMLElement>(`#${id}`);
    expect(node, `${label} ${id}`).not.toBeNull();
    const style = getComputedStyle(node!);
    expect(style.display, `${label} ${id}`).not.toBe("none");
    expect(style.visibility, `${label} ${id}`).not.toBe("hidden");
    expect(node!.getBoundingClientRect().height, `${label} ${id}`).toBeGreaterThan(0);
    expect(node!.textContent!.trim().length, `${label} ${id}`).toBeGreaterThan(0);
  }
  for (const control of target.querySelectorAll<HTMLButtonElement>("button")) {
    expect(control.getAttribute("aria-describedby")?.split(/\s+/u).sort(), `${label} ${control.textContent} curtain references`)
      .toEqual([...CURTAINS].sort());
  }
}

it.skipIf(!browser)("buys by keyboard, never shows owned before the receipt, then equips; curtains persist in every state", async () => {
  const { userEvent } = await import("vitest/browser");
  const calls: string[] = [];
  const initial = { arm: arms["acquirable-at-tier-1"]!, presentation: COSMETIC_SHOP_PRESENTATION, price: "$0.00", era: "era_2000" as const, pending: false,
    controlsEnabled: true, petName: () => "Mittens", receipt: null, rejection: null,
    onAcquire: (id: string) => { calls.push(`acquire:${id}`); }, onEquip: (id: string, pet: string) => { calls.push(`equip:${id}:${pet}`); },
    onUnequip: (pet: string) => { calls.push(`unequip:${pet}`); } };
  const target = host();
  const app = mount(CosmeticShelfHarness, { target, props: { initial } }) as unknown as { update(patch: Record<string, unknown>): void };
  try {
    await settle();
    await assertAxe(target, "unowned");
    assertCurtainsVisible(target, "unowned");
    const buy = target.querySelector<HTMLButtonElement>("button")!;
    expect(buy.getAttribute("aria-describedby")!.split(" ").sort()).toEqual([...CURTAINS].sort());
    expect(target.querySelector(".anchor")!.getAttribute("aria-describedby")).toBe("curtain-horse_armor-reference_price_anchor");
    buy.focus();
    expect(document.activeElement).toBe(buy);
    await userEvent.keyboard("{Enter}");
    expect(calls).toEqual(["acquire:horse_armor"]);
    // Pending: the button stays labelled and disabled; ownership is NOT rendered (I3).
    app.update({ pending: true });
    await settle();
    expect(buy.disabled).toBe(true);
    expect(target.querySelector("[data-state=owned]")).toBeNull();
    expect(target.textContent).not.toContain("Owned.");
    assertCurtainsVisible(target, "pending");
    // Applied receipt + authoritative snapshot: owned state, receipt line, focus moves.
    app.update({ pending: false, arm: arms["owned-not-worn"]!, receipt: { cosmeticId: "horse_armor", orderNumber: 1 } });
    await settle();
    expect(target.querySelector("[role=status].receipt")!.textContent).toContain("#1");
    expect(document.activeElement?.classList.contains("owned")).toBe(true);
    assertCurtainsVisible(target, "owned");
    await assertAxe(target, "owned");
    app.update({ pending: true });
    await settle();
    assertCurtainsVisible(target, "owned-pending");
    expect([...target.querySelectorAll<HTMLButtonElement>("button")].every((control) => control.disabled)).toBe(true);
    app.update({ pending: false });
    await settle();
    [...target.querySelectorAll<HTMLButtonElement>("button")].find((node) => node.textContent?.includes("Mittens"))!.click();
    expect(calls.at(-1)).toBe(`equip:horse_armor:${CAT}`);
    app.update({ arm: arms["owned-worn"]! });
    await settle();
    assertCurtainsVisible(target, "equipped");
    await assertAxe(target, "equipped");
    app.update({ pending: true });
    await settle();
    assertCurtainsVisible(target, "equipped-pending");
    expect([...target.querySelectorAll<HTMLButtonElement>("button")].every((control) => control.disabled)).toBe(true);
    for (const token of ["$2", "limited", "Limited", "countdown", "hurry"]) expect(target.textContent, token).not.toContain(token);
  } finally { unmount(app); target.remove(); }
});

it.skipIf(!browser)("activates Buy once with Space and waits for authoritative ownership", async () => {
  const { userEvent } = await import("vitest/browser");
  const calls: string[] = [];
  const target = host();
  const app = mount(CosmeticShelfHarness, { target, props: { initial: {
    arm: arms["acquirable-at-tier-1"]!, presentation: COSMETIC_SHOP_PRESENTATION, price: "$0.00", era: "era_2000" as const,
    pending: false, controlsEnabled: true, petName: () => "", receipt: null, rejection: null,
    onAcquire: (id: string) => { calls.push(id); }, onEquip: () => {}, onUnequip: () => {},
  } } }) as unknown as { update(patch: Record<string, unknown>): void };
  try {
    await settle();
    const buy = target.querySelector<HTMLButtonElement>("button")!;
    buy.focus();
    expect(document.activeElement).toBe(buy);
    await userEvent.keyboard(" ");
    expect(calls).toEqual(["horse_armor"]);
    app.update({ pending: true });
    await settle();
    expect(buy.disabled).toBe(true);
    expect(target.querySelector("[data-state=owned]")).toBeNull();
  } finally { unmount(app); target.remove(); }
});

it.skipIf(!browser)("shows the lock reason, the no-wearer line, and reflows at 320 px", async () => {
  const base = { presentation: COSMETIC_SHOP_PRESENTATION, price: "$0.00", era: "era_2000" as const, pending: false, controlsEnabled: true,
    petName: () => "", receipt: null, rejection: null, onAcquire: () => {}, onEquip: () => {}, onUnequip: () => {} };
  const target = host(320);
  const app = mount(CosmeticShelfHarness, { target, props: { initial: { ...base, arm: arms["locked-at-tier-0"]! } } }) as unknown as { update(patch: Record<string, unknown>): void };
  try {
    await settle();
    expect(target.textContent).toContain("Tier 1");
    expect(target.querySelector<HTMLButtonElement>("button")!.disabled).toBe(true);
    assertCurtainsVisible(target, "locked");
    expect(target.scrollWidth).toBeLessThanOrEqual(target.clientWidth + 1);
    app.update({ arm: arms["owned-no-wearer"]! });
    await settle();
    expect(target.textContent).toContain("Nobody to wear it yet.");
    await assertAxe(target, "no-wearer");
  } finally { unmount(app); target.remove(); }
});

it.skipIf(!browser)("renders the overlay layer and annoyed pose with no text node, static under actual reduced motion (AC15)", async () => {
  const { commands } = await import("vitest/browser");
  const motion = commands as typeof commands & {
    setReducedMotionPreference(preference: "reduce" | "no-preference"): Promise<void>;
  };
  await motion.setReducedMotionPreference("no-preference");
  try {
    const target = host();
    const sprite = document.createElement("div");
    sprite.style.position = "relative";
    sprite.style.inlineSize = "120px";
    sprite.style.blockSize = "120px";
    target.append(sprite);
    const animated = mount(CosmeticOverlay, { target: sprite, props: { renderKey: "horse_armor", reaction: "annoyed", animate: true } });
    try {
      await settle();
      const overlay = sprite.querySelector<HTMLElement>("[data-render=horse_armor]")!;
      const flick = overlay.querySelector<HTMLElement>(".flick")!;
      const assertPose = (label: string): void => {
        expect(overlay.getAttribute("aria-hidden"), label).toBe("true");
        expect(overlay.dataset.reaction, label).toBe("annoyed");
        expect(overlay.textContent!.trim(), `${label} overlay caption`).toBe("");
        const walker = document.createTreeWalker(overlay, NodeFilter.SHOW_TEXT);
        while (walker.nextNode()) expect(walker.currentNode.textContent!.trim(), `${label} overlay text node`).toBe("");
        expect(overlay.getBoundingClientRect().width, label).toBeGreaterThan(0);
        expect(overlay.getBoundingClientRect().height, label).toBeGreaterThan(0);
        for (const part of [".chest", ".crest", ".ears-back", ".flick"]) {
          expect(getComputedStyle(overlay.querySelector(part)!).display, `${label} ${part}`).not.toBe("none");
        }
      };
      expect(window.matchMedia("(prefers-reduced-motion: reduce)").matches).toBe(false);
      assertPose("ordinary motion");
      expect(getComputedStyle(flick).animationName, "ordinary preference animates the fixture pose").not.toBe("none");

      // Keep animate:true. Only the real browser preference changes, so the
      // static arm cannot pass merely because the host disabled animation.
      await motion.setReducedMotionPreference("reduce");
      await settle();
      expect(window.matchMedia("(prefers-reduced-motion: reduce)").matches).toBe(true);
      assertPose("reduced motion");
      expect(getComputedStyle(flick).animationName, "actual reduced-motion preference must keep the pose static").toBe("none");

      await motion.setReducedMotionPreference("no-preference");
      await settle();
      expect(window.matchMedia("(prefers-reduced-motion: reduce)").matches).toBe(false);
      assertPose("restored ordinary motion");
      expect(getComputedStyle(flick).animationName, "removing the preference restores ordinary animation").not.toBe("none");
    } finally { unmount(animated); target.remove(); }

    // The explicit host flag is a separate defense, not a substitute for the
    // browser-preference population above.
    const still = host();
    const staticOverlay = mount(CosmeticOverlay, { target: still, props: { renderKey: "horse_armor", reaction: "annoyed", animate: false } });
    try {
      await settle();
      expect(getComputedStyle(still.querySelector(".flick")!).animationName).toBe("none");
    } finally { unmount(staticOverlay); still.remove(); }
  } finally { await motion.setReducedMotionPreference("no-preference"); }
});
