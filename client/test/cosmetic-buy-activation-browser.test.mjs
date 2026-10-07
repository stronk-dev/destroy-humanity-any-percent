import { expect, it } from "vitest";
import { activateCosmeticBuy } from "../tools/activate-cosmetic-buy.mjs";

// Test the exact function sent to Chromium, not a copy of its readiness logic.
// DOM fixtures prove its guard/one-activation boundary, not server acquisition
// or native pointer/keyboard accessibility (AC11 has its own population).
for (const blocked of ["busy", "disabled", "owned", "wrong-label", "wrong-surface", "covered"]) {
  it.skipIf(typeof document === "undefined")(`Cosmetic Buy waits while ${blocked}, then activates the ready DOM consumer exactly once`, async () => {
    const main = document.createElement("main");
    main.dataset.surface = "desk"; main.setAttribute("aria-busy", "false");
    main.style.cssText = "position:fixed;left:20px;top:20px;width:300px;z-index:10000";
    const shelf = document.createElement("section"); shelf.dataset.testid = "cosmetic-shelf";
    const item = document.createElement("article");
    item.dataset.cosmetic = "horse_armor"; item.dataset.state = "unowned"; item.style.position = "relative";
    const button = document.createElement("button");
    button.textContent = "Buy for $0.00"; button.style.cssText = "width:200px;height:40px";
    const cover = document.createElement("div"); cover.style.cssText = "position:absolute;inset:0;z-index:2";
    item.append(button); shelf.append(item); main.append(shelf); document.body.append(main);
    const clicks = [];
    button.addEventListener("click", (event) => clicks.push({ disabled: button.disabled, busy: main.getAttribute("aria-busy"),
      state: item.dataset.state, surface: main.dataset.surface, label: button.textContent, trusted: event.isTrusted }));
    switch (blocked) {
      case "busy": main.setAttribute("aria-busy", "true"); break;
      case "disabled": button.disabled = true; break;
      case "owned": item.dataset.state = "owned"; break;
      case "wrong-label": button.textContent = "Another action"; break;
      case "wrong-surface": main.dataset.surface = "offer_sheet"; break;
      case "covered": item.append(cover); break;
    }
    const unblock = () => {
      main.setAttribute("aria-busy", "false"); button.disabled = false; item.dataset.state = "unowned";
      button.textContent = "Buy for $0.00"; main.dataset.surface = "desk"; cover.remove();
    };
    let completed = false;
    const activation = activateCosmeticBuy("Buy for $0.00").then(() => { completed = true; });
    try {
      await new Promise(requestAnimationFrame); await new Promise(requestAnimationFrame);
      expect(clicks, "the blocked control must not dispatch").toEqual([]);
      expect(completed).toBe(false);
      unblock(); await activation;
      expect(clicks).toEqual([{ disabled: false, busy: "false", state: "unowned", surface: "desk", label: "Buy for $0.00", trusted: false }]);
      await new Promise(requestAnimationFrame);
      expect(clicks).toHaveLength(1);
    } finally {
      unblock(); await activation; main.remove();
    }
  });
}
