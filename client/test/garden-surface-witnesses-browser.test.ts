import { flushSync, mount, tick, unmount, type ComponentProps } from "svelte";
import { expect, it } from "vitest";

import views from "../../testdata/garden/view-fixtures-v1.json";
import type { GardenCurrentResponse } from "../src/api/generated/types";
import GardenSurface from "../src/game-ui/garden/GardenSurface.svelte";
import type { GardenPort } from "../src/game-ui/garden/garden-port";
import { t } from "../src/copy";
import { installTheme, UI_THEMES } from "../src/ui/themes";
import GardenSurfaceHarness from "./GardenSurfaceHarness.svelte";

const browser = typeof document !== "undefined";
type Active = Extract<GardenCurrentResponse, { kind: "active" }>;

function active(substrate = "bare_metal", revision = 3): Active {
  const view = structuredClone(views.active) as Active;
  view.founder_revision = revision;
  view.server_ms += (revision - 3) * 1_000;
  view.garden.substrate_id = substrate;
  // These cases isolate ordering, not next-tick scheduling.
  view.garden.next_tick_wall_ms = null;
  return view;
}

async function settle(): Promise<void> {
  for (let index = 0; index < 5; index += 1) {
    await new Promise((resolve) => setTimeout(resolve, 0));
    await tick();
    flushSync();
  }
}

function mounted(port: GardenPort) {
  const calls: string[] = [];
  let commandStarted = () => {};
  const record = (value: string) => { calls.push(value); commandStarted(); };
  const target = document.createElement("main");
  document.body.append(target);
  installTheme(target, UI_THEMES.era_1995, false);
  const initial: ComponentProps<typeof GardenSurface> = {
    port, era: "era_1995", pending: false, controlsEnabled: true, refreshKey: 0, rejection: null, visible: () => true,
    onPlant: (row: number, col: number, species: string) => record(`plant ${row},${col} ${species}`),
    onUproot: (row: number, col: number) => record(`uproot ${row},${col}`),
    onHarvest: (plots: readonly { row: number; col: number }[]) => record(`harvest ${plots.map(({ row, col }) => `${row},${col}`).join(" ")}`),
    onSetSubstrate: (substrate: string) => record(`substrate ${substrate}`),
  };
  const app = mount(GardenSurfaceHarness, { target, props: { initial } });
  return { target, app, calls,
    holdCommands(after?: () => void) { commandStarted = () => { app.update({ pending: true }); after?.(); }; },
  };
}

for (const menu of ["empty", "mature"] as const) {
  it.skipIf(!browser)(`Garden unavailable ${menu} menu guards every callback and restores only explicit input`, async () => {
    const { userEvent } = await import("vitest/browser");
    const host = mounted({ current: async () => active() });
    try {
      await settle();
      const plot = host.target.querySelectorAll<HTMLButtonElement>("button.cell")[menu === "empty" ? 6 : 0]!;
      plot.click(); await settle();
      const trigger = host.target.querySelector<HTMLButtonElement>(".menu button")!;
      trigger.focus();
      host.app.update({ controlsEnabled: false }); await settle();
      expect(document.activeElement).toBe(host.target.querySelector("#garden-heading"));
      const controls = [...host.target.querySelectorAll<HTMLButtonElement>("button")].filter((node) => node.textContent?.trim() !== "Close");
      for (const control of controls) {
        expect(control.disabled).toBe(true);
        expect(control.getAttribute("aria-describedby")?.split(" ")).toContain("garden-stale");
        // Explicit dispatch bypasses HTMLButtonElement.click's native disabled
        // suppression: the component must guard its callbacks too.
        control.dispatchEvent(new MouseEvent("click", { bubbles: true }));
      }
      await settle();
      expect(host.calls).toEqual([]);
      expect(host.target.querySelector(".menu")).not.toBeNull();
      host.app.update({ controlsEnabled: true }); await settle();
      expect(host.calls).toEqual([]);
      expect(trigger.disabled).toBe(false);
      trigger.focus(); await userEvent.keyboard("{Enter}"); await settle();
      expect(host.calls).toEqual([menu === "empty" ? "plant 1,0 strain_a" : "harvest 0,0"]);
      expect(document.activeElement).toBe(plot);
    } finally { await unmount(host.app); host.target.remove(); }
  });
}

for (const key of ["{Enter}", " "]) {
  for (const action of ["menu-harvest", "harvest-all", "substrate"] as const) {
    it.skipIf(!browser)(`Garden pending ${action} retains native ${JSON.stringify(key)} focus and refuses repeats`, async () => {
      const { userEvent } = await import("vitest/browser");
      const host = mounted({ current: async () => active() });
      try {
        await settle();
        host.holdCommands();
        const origin = host.target.querySelector<HTMLButtonElement>("button.cell")!;
        origin.focus();
        let expectedFocus: HTMLButtonElement = origin;
        if (action === "menu-harvest") {
          await userEvent.keyboard("{Enter}{Tab}");
          expect(document.activeElement?.textContent?.trim()).toBe("Harvest");
        } else {
          await userEvent.keyboard(action === "harvest-all" ? "{Tab}" : "{Tab}{Tab}{Tab}{Tab}{Tab}");
          expectedFocus = document.activeElement as HTMLButtonElement;
          expect(expectedFocus.textContent?.trim()).toBe(action === "harvest-all" ? "Harvest all mature" : "Mainframe");
        }
        await userEvent.keyboard(key);
        await settle();
        const expected = action === "menu-harvest" ? "harvest 0,0" : action === "harvest-all" ? "harvest 0,0 0,1 1,1" : "substrate mainframe";
        expect(host.calls).toEqual([expected]);
        expect(document.activeElement).toBe(expectedFocus);
        expect(expectedFocus.disabled).toBe(false);
        expect(expectedFocus.getAttribute("aria-disabled")).toBe("true");
        expect(host.target.querySelector(".garden")?.getAttribute("aria-busy")).toBe("true");
        expect(host.target.querySelector(".live")?.textContent).toBe(t("common.pending", {}, "era_1995"));
        await userEvent.keyboard("{Enter} ");
        expectedFocus.click();
        await settle();
        expect(host.calls).toEqual([expected]);
        expect(host.target.querySelector(".menu")).toBeNull();
        host.app.update({ pending: false });
        await settle();
        expect(document.activeElement).toBe(expectedFocus);
        expect(expectedFocus.getAttribute("aria-disabled")).not.toBe("true");
        expect(host.target.querySelector(".live")?.textContent).toBe("");
        if (action === "menu-harvest") await userEvent.keyboard("{Enter}{Tab}");
        await userEvent.keyboard(key);
        await settle();
        expect(host.calls).toEqual([expected, expected]);
      } finally { await unmount(host.app); host.target.remove(); }
    });
  }
}

for (const action of ["plant", "uproot", "harvest"] as const) {
  it.skipIf(!browser)(`Garden open-menu ${action} stays focusable but cannot dispatch while pending`, async () => {
    const { userEvent } = await import("vitest/browser");
    const host = mounted({ current: async () => active() });
    try {
      await settle();
      const cells = [...host.target.querySelectorAll<HTMLButtonElement>("button.cell")];
      cells[0]!.focus();
      if (action === "plant") await userEvent.keyboard("{ArrowDown}");
      await userEvent.keyboard("{Enter}{Tab}");
      if (action === "uproot") await userEvent.keyboard("{Tab}");
      const trigger = document.activeElement as HTMLButtonElement;
      expect(trigger.textContent?.trim()).toBe(action === "plant" ? "Plant Strain A (PENDING OWNER NAME)" : action === "uproot" ? "Uproot" : "Harvest");
      host.app.update({ pending: true });
      await settle();
      expect(document.activeElement).toBe(trigger);
      expect(trigger.disabled).toBe(false);
      expect(trigger.getAttribute("aria-disabled")).toBe("true");
      await userEvent.keyboard("{Enter} ");
      trigger.click();
      await settle();
      expect(host.calls).toEqual([]);
      expect(host.target.querySelector(".menu")).not.toBeNull();
      host.app.update({ pending: false });
      await settle();
      await userEvent.keyboard("{Enter}");
      await settle();
      expect(host.calls).toEqual([action === "plant" ? "plant 1,0 strain_a" : action === "uproot" ? "uproot 0,0" : "harvest 0,0"]);
      expect(host.target.querySelector(".menu")).toBeNull();
      expect(document.activeElement).toBe(cells[action === "plant" ? 6 : 0]);
    } finally { await unmount(host.app); host.target.remove(); }
  });
}

for (const key of ["{Enter}", " "]) {
  it.skipIf(!browser)(`Garden menu completion preserves a newer focus choice after native ${JSON.stringify(key)}`, async () => {
    const { userEvent } = await import("vitest/browser");
    const host = mounted({ current: async () => active() });
    const newerChoice = document.createElement("button");
    newerChoice.textContent = "Another control";
    document.body.append(newerChoice);
    try {
      await settle();
      host.target.querySelector<HTMLButtonElement>("button.cell")!.focus();
      await userEvent.keyboard("{Enter}{Tab}");
      expect(document.activeElement?.textContent?.trim()).toBe("Harvest");
      // A focus choice made after submission but before the menu's async
      // render/handoff completes must win. No real server response is mocked.
      host.holdCommands(() => newerChoice.focus());
      await userEvent.keyboard(key);
      await settle();
      expect(host.calls).toEqual(["harvest 0,0"]);
      expect(host.target.querySelector(".menu")).toBeNull();
      expect(document.activeElement).toBe(newerChoice);
      host.app.update({ pending: false });
      await settle();
      expect(document.activeElement).toBe(newerChoice);
    } finally { await unmount(host.app); host.target.remove(); newerChoice.remove(); }
  });
}

class OrderedPort implements GardenPort {
  readonly requests: {
    resolve(value: GardenCurrentResponse): void;
    reject(error: Error): void;
  }[] = [];
  current(): Promise<GardenCurrentResponse> {
    return new Promise((resolve, reject) => this.requests.push({ resolve, reject }));
  }
}

function pressed(target: HTMLElement): string | null {
  return target.querySelector(".substrates button[aria-pressed=true]")?.textContent?.trim() ?? null;
}

const mixedKeys = ["{ArrowUp}", "{ArrowLeft}", "{ArrowRight}", "{ArrowDown}", "{ArrowLeft}", "{ArrowUp}"];

async function nativeWalk(target: HTMLElement, keys: readonly string[], expectedFocus: readonly number[]): Promise<HTMLButtonElement[]> {
  const { userEvent } = await import("vitest/browser");
  const cells = [...target.querySelectorAll<HTMLButtonElement>("button.cell")];
  expect(cells).toHaveLength(36);
  cells[0]!.focus();
  const focus: { cell: number; tabStops: number[] }[] = [];
  const recordFocus = (event: FocusEvent) => {
    const index = cells.indexOf(event.target as HTMLButtonElement);
    if (index >= 0) focus.push({ cell: index, tabStops: cells.flatMap((cell, at) => cell.tabIndex === 0 ? [at] : []) });
  };
  target.addEventListener("focusin", recordFocus);
  try {
    // Native key-down/up, with every intermediate focus and roving tabstop
    // observed; edge clamps must not introduce an extra focus transition.
    await userEvent.keyboard(keys.join(""));
    await settle();
  } finally { target.removeEventListener("focusin", recordFocus); }
  expect(focus).toEqual(expectedFocus.map((cell) => ({ cell, tabStops: [cell] })));
  const last = cells[expectedFocus[expectedFocus.length - 1]!]!;
  expect(document.activeElement).toBe(last);
  expect(cells.filter((cell) => cell.tabIndex === 0)).toEqual([last]);
  return cells;
}

for (const path of [
  {
    name: "mixed directions and outbound bottom/right edges",
    keys: [...mixedKeys, "{ArrowDown}", "{ArrowDown}", "{ArrowDown}", "{ArrowDown}", "{ArrowDown}",
      "{ArrowRight}", "{ArrowRight}", "{ArrowRight}", "{ArrowRight}", "{ArrowRight}", "{ArrowRight}", "{ArrowDown}"],
    focus: [1, 7, 6, 0, 6, 12, 18, 24, 30, 31, 32, 33, 34, 35],
  },
  {
    name: "native arrival and complete left/up return",
    keys: ["{ArrowRight}", "{ArrowRight}", "{ArrowRight}", "{ArrowRight}", "{ArrowRight}",
      "{ArrowDown}", "{ArrowDown}", "{ArrowDown}", "{ArrowDown}", "{ArrowDown}",
      "{ArrowLeft}", "{ArrowLeft}", "{ArrowLeft}", "{ArrowLeft}", "{ArrowLeft}",
      "{ArrowUp}", "{ArrowUp}", "{ArrowUp}", "{ArrowUp}", "{ArrowUp}"],
    focus: [1, 2, 3, 4, 5, 11, 17, 23, 29, 35, 34, 33, 32, 31, 30, 24, 18, 12, 6, 0],
  },
]) {
  it.skipIf(!browser)(`Garden native navigation: ${path.name}`, async () => {
    const started = performance.now();
    const { target, app, calls } = mounted({ current: async () => active() });
    try {
      await settle();
      await nativeWalk(target, path.keys, path.focus);
      expect(calls).toEqual([]);
      expect(target.querySelector(".menu")).toBeNull();
    } finally {
      await unmount(app);
      target.remove();
      console.info("garden-native-navigation", JSON.stringify({ path: path.name, user_agent: navigator.userAgent, elapsed_ms: performance.now() - started }));
    }
  });
}

for (const key of ["{Enter}", " "]) {
  it.skipIf(!browser)(`Garden native ${JSON.stringify(key)} navigates and dispatches once, with pending refusal`, async () => {
    const started = performance.now();
    const observed = (stage: string) => console.info("garden-native-stage", JSON.stringify({ key, stage, user_agent: navigator.userAgent, elapsed_ms: performance.now() - started }));
    observed("entry");
    const { userEvent } = await import("vitest/browser");
    observed("helper-import");
    const { target, app, calls } = mounted({ current: async () => active() });
    try {
      await settle();
      observed("initial-mount");
      const cells = await nativeWalk(target, mixedKeys, [1, 7, 6, 0]);
      observed("navigation");
      await userEvent.keyboard(key);
      await settle();
      observed("menu-open");
      expect(target.querySelector(".menu")).not.toBeNull();
      expect(calls).toEqual([]);
      await userEvent.keyboard("{Tab}");
      observed("tab");
      expect(document.activeElement?.textContent?.trim()).toBe("Harvest");
      await userEvent.keyboard(key);
      await settle();
      observed("command");
      expect(calls).toEqual(["harvest 0,0"]);
      expect(target.querySelector(".menu")).toBeNull();
      expect(document.activeElement).toBe(cells[0]);
      app.update({ pending: true });
      await settle();
      expect(cells.every((cell) => !cell.disabled && cell.getAttribute("aria-disabled") === "true")).toBe(true);
      expect(document.activeElement).toBe(cells[0]);
      expect([...target.querySelectorAll<HTMLButtonElement>(".substrates button")].every((button) => !button.matches(":disabled") && button.getAttribute("aria-disabled") === "true")).toBe(true);
      await userEvent.keyboard(key);
      await settle();
      expect(calls).toEqual(["harvest 0,0"]);
      observed("pending-refusal");
    } finally { await unmount(app); target.remove(); observed("cleanup"); }
  });
}

for (const empty of [false, true]) {
  it.skipIf(!browser)(`Garden native Tab and Shift-Tab traverse the ${empty ? "seed" : "mature"} menu and substrates`, async () => {
    const { userEvent } = await import("vitest/browser");
    const { target, app, calls } = mounted({ current: async () => active() });
    try {
      await settle();
      const cells = [...target.querySelectorAll<HTMLButtonElement>("button.cell")];
      cells[0]!.focus();
      if (empty) { await userEvent.keyboard("{ArrowDown}"); await settle(); }
      const origin = cells[empty ? 6 : 0]!;
      expect(document.activeElement).toBe(origin);
      await userEvent.keyboard("{Enter}"); await settle();
      const menu = [...target.querySelectorAll<HTMLButtonElement>(".menu button")];
      expect(menu).toHaveLength(3);
      expect(menu.map((row) => row.textContent?.trim())).toEqual(empty
        ? ["Plant Strain A (PENDING OWNER NAME)", "Plant Strain B (PENDING OWNER NAME)", "Close"]
        : ["Harvest", "Uproot", "Close"]);
      const all = [...target.querySelectorAll<HTMLButtonElement>("button")].find((row) => row.textContent?.trim() === "Harvest all mature")!;
      const substrates = [...target.querySelectorAll<HTMLButtonElement>(".substrates button")];
      expect(substrates.map((row) => row.textContent?.trim())).toEqual(["Bare metal", "Chaos Monkey", "Containerized", "Mainframe"]);
      const path = [...menu, all, ...substrates];
      for (const control of path) {
        await userEvent.keyboard("{Tab}");
        expect(document.activeElement).toBe(control);
      }
      for (const control of [...path.slice(0, -1)].reverse().concat(origin)) {
        await userEvent.keyboard("{Shift>}{Tab}{/Shift}");
        expect(document.activeElement).toBe(control);
      }
      expect(calls).toEqual([]);
      expect(cells.filter((cell) => cell.tabIndex === 0)).toEqual([origin]);
    } finally { await unmount(app); target.remove(); }
  });
}

for (const key of ["{Enter}", " "]) {
  for (const [empty, at, expected] of [
    [false, 1, "uproot 0,0"], [false, 2, null],
    [true, 0, "plant 1,0 strain_a"], [true, 1, "plant 1,0 strain_b"],
  ] as const) {
    it.skipIf(!browser)(`Garden native ${JSON.stringify(key)} reaches ${empty ? "seed" : "mature"} menu action ${at}`, async () => {
      const { userEvent } = await import("vitest/browser");
      const { target, app, calls } = mounted({ current: async () => active() });
      try {
        await settle();
        const cells = [...target.querySelectorAll<HTMLButtonElement>("button.cell")];
        cells[0]!.focus();
        if (empty) { await userEvent.keyboard("{ArrowDown}"); await settle(); }
        const origin = cells[empty ? 6 : 0]!;
        await userEvent.keyboard(key); await settle();
        const action = target.querySelectorAll<HTMLButtonElement>(".menu button")[at]!;
        for (let index = 0; index <= at; index++) await userEvent.keyboard("{Tab}");
        expect(document.activeElement).toBe(action);
        expect(calls).toEqual([]);
        await userEvent.keyboard(key); await settle();
        expect(calls).toEqual(expected === null ? [] : [expected]);
        expect(target.querySelector(".menu")).toBeNull();
        expect(document.activeElement).toBe(origin);
      } finally { await unmount(app); target.remove(); }
    });
  }
}

for (const key of ["{Enter}", " "]) {
  it.skipIf(!browser)(`Garden native ${JSON.stringify(key)} reaches Harvest all and every substrate`, async () => {
    const { userEvent } = await import("vitest/browser");
    const { target, app, calls } = mounted({ current: async () => active() });
    try {
      await settle();
      target.querySelector<HTMLButtonElement>("button.cell")!.focus();
      await userEvent.keyboard("{Tab}");
      expect(document.activeElement?.textContent?.trim()).toBe("Harvest all mature");
      await userEvent.keyboard(key); await settle();
      expect(calls).toEqual(["harvest 0,0 0,1 1,1"]);
      const substrates = [...target.querySelectorAll<HTMLButtonElement>(".substrates button")];
      for (const [index, control] of substrates.entries()) {
        await userEvent.keyboard("{Tab}");
        expect(document.activeElement).toBe(control);
        await userEvent.keyboard(key); await settle();
        // The read still owns Bare metal; selecting it is a no-op. Other
        // selections dispatch exactly once, without an optimistic view patch.
        expect(calls).toEqual(["harvest 0,0 0,1 1,1", ...["chaos_monkey", "containerized", "mainframe"].slice(0, index).map((id) => `substrate ${id}`)]);
        expect(pressed(target)).toBe("Bare metal");
      }
    } finally { await unmount(app); target.remove(); }
  });

  it.skipIf(!browser)(`Garden native ${JSON.stringify(key)} reaches Refresh after a failed read`, async () => {
    const { userEvent } = await import("vitest/browser");
    let reads = 0;
    const { target, app, calls } = mounted({ current: async () => {
      if (++reads === 1) throw new Error("controlled read failure");
      return active();
    } });
    const before = document.createElement("input"); target.before(before);
    try {
      await settle();
      expect(reads).toBe(1);
      before.focus(); await userEvent.keyboard("{Tab}");
      expect(document.activeElement?.textContent?.trim()).toBe("Refresh");
      await userEvent.keyboard(key); await settle();
      expect(reads).toBe(2);
      expect(target.querySelectorAll("button.cell")).toHaveLength(36);
      expect(calls).toEqual([]);
    } finally { await unmount(app); target.remove(); before.remove(); }
  });
}

it.skipIf(!browser)("Garden receipt refresh sequential control displays the newer read", async () => {
  const port = new OrderedPort();
  const { target, app, calls } = mounted(port);
  try {
    await settle();
    expect(port.requests).toHaveLength(1);
    port.requests[0]!.resolve(active());
    await settle();
    expect(pressed(target)).toBe("Bare metal");
    app.update({ refreshKey: 1 });
    await settle();
    expect(port.requests).toHaveLength(2);
    port.requests[1]!.resolve(active("mainframe", 4));
    await settle();
    expect(pressed(target)).toBe("Mainframe");
    expect(calls).toEqual([]);
  } finally { await unmount(app); target.remove(); }
});

for (const outcome of ["active", "locked", "error", "old-error"] as const) {
  it.skipIf(!browser)(`Garden receipt refresh retains newest ${outcome} after late older completion`, async () => {
    const port = new OrderedPort();
    const { target, app, calls } = mounted(port);
    try {
      await settle();
      expect(port.requests).toHaveLength(1);
      port.requests[0]!.resolve(active());
      await settle();
      app.update({ refreshKey: 1 });
      await settle();
      expect(port.requests).toHaveLength(2);
      app.update({ refreshKey: 2 });
      await settle();
      expect(port.requests).toHaveLength(3);
      if (outcome === "locked") port.requests[2]!.resolve(views.locked as Extract<GardenCurrentResponse, { kind: "locked" }>);
      else if (outcome === "error") port.requests[2]!.reject(new Error("newest read failed"));
      else port.requests[2]!.resolve(active("mainframe", 5));
      await settle();
      const newestDOM = target.innerHTML;
      if (outcome === "locked") {
        expect(target.textContent).toContain("Locked");
        expect(target.querySelector(".grid")).toBeNull();
      } else if (outcome === "error") {
        expect(pressed(target)).toBe("Bare metal");
        expect(target.textContent).toContain("This view may be out of date. Refreshing.");
      } else expect(pressed(target)).toBe("Mainframe");
      if (outcome === "old-error") port.requests[1]!.reject(new Error("older read failed"));
      else port.requests[1]!.resolve(active("bare_metal", 4));
      await settle();
      expect(target.innerHTML === newestDOM, "an older request must not replace the newer receipt refresh").toBe(true);
      expect(calls).toEqual([]);
    } finally { await unmount(app); target.remove(); }
  });
}

it.skipIf(!browser)("Garden ignores late read completion after unmount", async () => {
  const port = new OrderedPort();
  const { target, app, calls } = mounted(port);
  let destroyed = false;
  try {
    await settle();
    expect(port.requests).toHaveLength(1);
    await unmount(app);
    destroyed = true;
    port.requests[0]!.resolve(active());
    await settle();
    expect(target.childElementCount).toBe(0);
    expect(calls).toEqual([]);
  } finally {
    if (!destroyed) await unmount(app);
    target.remove();
  }
});
