import { expect, it } from "vitest";
import { activateCosmeticBuy, assertNativeCosmeticBuyTrace, installCosmeticBuyFocusTrace } from "../tools/activate-cosmetic-buy.mjs";

for (const key of ["Enter", " "]) {
  const label = "Buy for $0.00";
  const event = { trusted: true, focused: true, button: label, cosmetic: "horse_armor", state: "unowned",
    connected: true, disabled: false, surface: "desk", busy: "false" };
  const trace = () => ({ dropped_events: 0, events: key === "Enter"
    ? [{ ...event, type: "keydown", key }, { ...event, type: "click" }, { type: "keyup", key, trusted: true }]
    : [{ ...event, type: "keydown", key }, { type: "keyup", key, trusted: true }, { ...event, type: "click" }] });
  it(`native Cosmetic ${JSON.stringify(key)} trace admits the browser's single activation ordering`, () => {
    expect(() => assertNativeCosmeticBuyTrace(trace(), key, label)).not.toThrow();
  });
  for (const wrong of ["untrusted-click", "wrong-key", "unfocused", "duplicate-click", "missing-keyup", "truncated", "pending"]) {
    it(`native Cosmetic ${JSON.stringify(key)} trace rejects ${wrong} instead of accepting purchase alone`, () => {
      const value = trace();
      const click = value.events.find((row) => row.type === "click");
      switch (wrong) {
        case "untrusted-click": click.trusted = false; break;
        case "wrong-key": value.events[0].key = "Escape"; break;
        case "unfocused": value.events[0].focused = false; break;
        case "duplicate-click": value.events.push({ ...click }); break;
        case "missing-keyup": value.events = value.events.filter((row) => row.type !== "keyup"); break;
        case "truncated": value.dropped_events = 1; break;
        case "pending": click.busy = "true"; break;
      }
      expect(() => assertNativeCosmeticBuyTrace(value, key, label)).toThrow("lacks one native");
    });
  }
}

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

for (const state of ["focused", "other-control", "replaced", "disabled"]) {
  it.skipIf(typeof document === "undefined")(`Cosmetic focus observation distinguishes ${state} without activating or changing focus`, () => {
    const root = document.createElement("main");
    root.innerHTML = '<button>Other action</button><section data-testid="cosmetic-shelf"><article data-cosmetic="horse_armor"><button>Buy for $0.00</button></article></section>';
    document.body.append(root);
    const dispose = installCosmeticBuyFocusTrace();
    const read = globalThis.__cosmeticBuyFocusBoundary;
    const other = root.querySelector("button"), buy = root.querySelector("article button");
    let clicks = 0;
    root.addEventListener("click", () => { clicks += 1; });
    try {
      buy.focus();
      const before = read();
      expect(before.active_is_control).toBe(true);
      expect(before.active.node).toBe(before.control.node);
      if (state === "other-control") other.focus();
      if (state === "disabled") { other.focus(); buy.disabled = true; buy.focus(); }
      if (state === "replaced") buy.replaceWith(buy.cloneNode(true));
      const active = document.activeElement, observed = read();
      expect(document.activeElement).toBe(active);
      expect(clicks).toBe(0);
      expect(observed.active_is_control).toBe(state === "focused");
      expect(observed.control.disabled).toBe(state === "disabled");
      expect(observed.control.node === before.control.node).toBe(state !== "replaced");
      expect(observed.events.some((event) => event.type === "focusin" && event.target.node === before.control.node)).toBe(true);
      if (state === "other-control" || state === "disabled") {
        expect(observed.active.button).toBe("Other action");
        expect(observed.events.some((event) => event.type === "focusout" && event.target.node === before.control.node && event.related.node === observed.active.node)).toBe(true);
      }
    } finally { root.remove(); dispose(); }
  });
}

it.skipIf(typeof document === "undefined")("Cosmetic focus observation exposes truncation and never copies input values", () => {
  const root = document.createElement("section");
  root.innerHTML = '<button>First</button><input type="password" value="not-diagnostic-data">';
  document.body.append(root);
  const dispose = installCosmeticBuyFocusTrace();
  try {
    const [button, input] = root.children;
    for (let index = 0; index < 20; index += 1) { button.focus(); input.focus(); }
    const observed = globalThis.__cosmeticBuyFocusBoundary();
    expect(observed.events).toHaveLength(32);
    expect(observed.dropped_events).toBeGreaterThan(0);
    expect(observed.active.tag).toBe("INPUT");
    expect(observed.control).toBe(null);
    expect(observed.active_is_control).toBe(false);
    expect(JSON.stringify(observed)).not.toContain("not-diagnostic-data");
  } finally { root.remove(); dispose(); }
});
