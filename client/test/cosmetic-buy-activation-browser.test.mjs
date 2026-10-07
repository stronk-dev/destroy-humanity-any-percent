import { expect, it } from "vitest";
import { activateCosmeticBuy } from "../tools/activate-cosmetic-buy.mjs";

// Test the exact function sent to Chromium, not a copy of its readiness logic.
// DOM fixtures prove its guard/one-activation boundary, not server acquisition
// or native pointer/keyboard accessibility (AC11 has its own population).
for (const action of ["Buy", "Equip"]) for (const blocked of ["busy", "disabled", "wrong-state", "wrong-label", "wrong-surface", "covered"]) {
  it.skipIf(typeof document === "undefined")(`Cosmetic ${action} waits while ${blocked}, then activates the ready DOM consumer exactly once`, async () => {
    const state = action === "Buy" ? "unowned" : "owned";
    const label = action === "Buy" ? "Buy for $0.00" : "Equip on Mittens";
    const main = document.createElement("main");
    main.dataset.surface = "desk"; main.setAttribute("aria-busy", "false");
    main.style.cssText = "position:fixed;left:20px;top:20px;width:300px;z-index:10000";
    const shelf = document.createElement("section"); shelf.dataset.testid = "cosmetic-shelf";
    const item = document.createElement("article");
    item.dataset.cosmetic = "horse_armor"; item.dataset.state = state; item.style.position = "relative";
    const button = document.createElement("button");
    button.textContent = label; button.style.cssText = "width:200px;height:40px";
    const cover = document.createElement("div"); cover.style.cssText = "position:absolute;inset:0;z-index:2";
    item.append(button); shelf.append(item); main.append(shelf); document.body.append(main);
    const clicks = [];
    button.addEventListener("click", (event) => clicks.push({ disabled: button.disabled, busy: main.getAttribute("aria-busy"),
      state: item.dataset.state, surface: main.dataset.surface, label: button.textContent, trusted: event.isTrusted }));
    switch (blocked) {
      case "busy": main.setAttribute("aria-busy", "true"); break;
      case "disabled": button.disabled = true; break;
      case "wrong-state": item.dataset.state = state === "owned" ? "unowned" : "owned"; break;
      case "wrong-label": button.textContent = "Another action"; break;
      case "wrong-surface": main.dataset.surface = "offer_sheet"; break;
      case "covered": item.append(cover); break;
    }
    const unblock = () => {
      main.setAttribute("aria-busy", "false"); button.disabled = false; item.dataset.state = state;
      button.textContent = label; main.dataset.surface = "desk"; cover.remove();
    };
    let completed = false;
    const activation = activateCosmeticBuy(action === "Buy" ? label : { label, state, budgetMs: 30_000 }).then(() => { completed = true; });
    try {
      await new Promise(requestAnimationFrame); await new Promise(requestAnimationFrame);
      expect(clicks, "the blocked control must not dispatch").toEqual([]);
      expect(completed).toBe(false);
      unblock(); await activation;
      expect(clicks).toEqual([{ disabled: false, busy: "false", state, surface: "desk", label, trusted: false }]);
      await new Promise(requestAnimationFrame);
      expect(clicks).toHaveLength(1);
    } finally {
      unblock(); await activation; main.remove();
    }
  });
}

it.skipIf(typeof document === "undefined")("Cosmetic DOM activation refuses an ambiguous wearer label without clicking either", async () => {
  const main = document.createElement("main");
  main.innerHTML = '<section data-testid="cosmetic-shelf"><article data-cosmetic="horse_armor" data-state="owned"><button>Equip</button><button>Equip</button></article></section>';
  document.body.append(main);
  let clicks = 0;
  main.addEventListener("click", () => { clicks += 1; });
  try {
    await expect(activateCosmeticBuy({ label: "Equip", state: "owned", budgetMs: 30_000 })).rejects.toThrow("ambiguous");
    expect(clicks).toBe(0);
  } finally { main.remove(); }
});

it.skipIf(typeof document === "undefined")("Cosmetic DOM activation refuses invalid or expanded action budgets", async () => {
  for (const budgetMs of [0, -1, 30_001, Number.POSITIVE_INFINITY]) {
    await expect(activateCosmeticBuy({ label: "Equip", state: "owned", budgetMs })).rejects.toThrow("invalid cosmetic");
  }
  await expect(activateCosmeticBuy({ label: "Equip", state: "pending", budgetMs: 30_000 })).rejects.toThrow("invalid cosmetic");
});
