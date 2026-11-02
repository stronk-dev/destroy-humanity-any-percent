import { flushSync, mount, tick, unmount, type ComponentProps } from "svelte";
import { beforeEach, describe, expect, it, type TestContext } from "vitest";

import views from "../../testdata/garden/view-fixtures-v1.json";
import type { GardenCurrentResponse } from "../src/api/generated/types";
import GardenSurface from "../src/game-ui/garden/GardenSurface.svelte";
import type { GardenPort } from "../src/game-ui/garden/garden-port";
import { t } from "../src/copy";
import { installTheme, UI_THEMES } from "../src/ui/themes";
import GardenSurfaceHarness from "./GardenSurfaceHarness.svelte";
import { assertNativeInputDrained, lifetimeKeyboard } from "./native-keyboard-lifetime";

const browser = typeof document !== "undefined";
beforeEach(assertNativeInputDrained);
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

function mounted(testContext: Pick<TestContext, "signal" | "onTestFinished">, port: GardenPort) {
  testContext.signal.throwIfAborted();
  const keyboard = lifetimeKeyboard(testContext.signal, async keys => {
    const { userEvent } = await import("vitest/browser");
    testContext.signal.throwIfAborted();
    await userEvent.keyboard(keys);
  });
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
  const adjacent: HTMLElement[] = [];
  let disposal: Promise<void> | undefined;
  const dispose = () => disposal ??= (async () => {
    try { await keyboard.closeAndDrain(); }
    finally { try { await unmount(app); } finally { target.remove(); for (const node of adjacent) node.remove(); } }
  })();
  testContext.onTestFinished(dispose);
  return { target, app, calls, press: keyboard.press, dispose,
    own<T extends HTMLElement>(node: T): T {
      testContext.signal.throwIfAborted();
      if (disposal) throw new Error("Garden fixture is closed");
      adjacent.push(node); return node;
    },
    holdCommands(after?: () => void) { commandStarted = () => { app.update({ pending: true }); after?.(); }; },
  };
}

describe("Garden_fixture_deadline", () => {
  const expiryName = "expires_after_the_real_native_fixture_is_ready";
  let signal: AbortSignal;
  let host: ReturnType<typeof mounted>;
  let adjacent: HTMLButtonElement;
  beforeEach(async testContext => {
    if (testContext.task.name !== expiryName) return;
    signal = testContext.signal;
    host = mounted(testContext, { current: async () => active() });
    adjacent = host.own(document.createElement("button")); adjacent.textContent = "Fixture neighbor";
    host.target.before(adjacent);
    await settle();
    await import("vitest/browser");
    host.target.querySelector<HTMLButtonElement>("button.cell")!.focus();
    await host.press("{Enter}"); await settle();
    expect(host.target.querySelectorAll(".menu button")).toHaveLength(3);
  });
  it.skipIf(!browser).fails(expiryName, async testContext => {
    await new Promise<void>(resolve => testContext.signal.addEventListener("abort", () => resolve(), { once: true }));
    testContext.signal.throwIfAborted();
  }, 50);
  it.skipIf(!browser)("successor_starts_without_the_expired_Garden", async () => {
    try {
      expect(signal.aborted).toBe(true);
      expect(signal.reason.message).toContain("Test timed out in 50ms");
      expect(host.target.isConnected, "framework deadline must dispose the old fixture before its successor").toBe(false);
      expect(adjacent.isConnected, "owned neighboring controls must not remain as successor Tab stops").toBe(false);
    } finally {
      // Keep the negative control from contaminating later tests on the old fixture.
      if (host.target.isConnected) { await host.dispose(); }
      adjacent.remove();
    }
  });
});

for (const abortDuringNative of [false, true]) {
  it.skipIf(!browser)(`Garden_fixture_lifetime ${abortDuringNative ? "drains trusted input on cancellation" : "keeps healthy native input unchanged"}`, async testContext => {
    const controller = new AbortController();
    const reason = new Error("controlled Garden cancellation");
    const host = mounted({ signal: controller.signal, onTestFinished: testContext.onTestFinished }, { current: async () => active() });
    const successor = document.createElement("input"); document.body.append(successor);
    // This is the next live subject, deliberately not owned by the expired host.
    testContext.onTestFinished(async () => { try { await host.dispose(); } finally { successor.remove(); } });
    let closing: Promise<void> | undefined;
    let connectedAtClose = false;
    const native: { type: string; key: string; trusted: boolean; connected: boolean }[] = [];
    const observe = (event: KeyboardEvent) => {
      native.push({ type: event.type, key: event.key, trusted: event.isTrusted, connected: host.target.isConnected });
      if (abortDuringNative && event.type === "keydown" && event.key === "Enter" && event.isTrusted) {
        controller.abort(reason);
        closing = host.dispose();
        connectedAtClose = host.target.isConnected;
      }
    };
    document.addEventListener("keydown", observe, true); document.addEventListener("keyup", observe, true);
    try {
      await settle(); host.target.querySelector<HTMLButtonElement>("button.cell")!.focus();
      if (abortDuringNative) {
        // Cancel while the actual native RPC is outstanding, never a held double.
        // Do not prevent its default action; teardown must wait for keyup.
        await expect(host.press("{Enter}")).rejects.toBe(reason);
        expect(connectedAtClose).toBe(true);
        expect(closing).toBeDefined(); await closing;
      } else {
        await host.press("{Enter}"); await settle();
        expect(host.target.querySelectorAll(".menu button")).toHaveLength(3);
        closing = host.dispose(); await closing;
      }
      expect(native).toEqual([
        { type: "keydown", key: "Enter", trusted: true, connected: true },
        { type: "keyup", key: "Enter", trusted: true, connected: true },
      ]);
      expect(host.dispose()).toBe(closing);
      expect(host.target.isConnected).toBe(false); assertNativeInputDrained();
      successor.focus();
      if (abortDuringNative) await expect(host.press(" ")).rejects.toBe(reason);
      else await expect(host.press(" ")).rejects.toThrow("fixture is closed");
      expect(successor.value).toBe(""); expect(document.activeElement).toBe(successor);
      expect(native).toHaveLength(2); expect(host.calls).toEqual([]);
    } finally {
      document.removeEventListener("keydown", observe, true); document.removeEventListener("keyup", observe, true);
      try { await host.dispose(); } finally { successor.remove(); }
    }
  });
}

for (const menu of ["empty", "mature"] as const) {
  it.skipIf(!browser)(`Garden unavailable ${menu} menu guards every callback and restores only explicit input`, async testContext => {
    await import("vitest/browser");
    const host = mounted(testContext, { current: async () => active() });
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
      trigger.focus(); await host.press("{Enter}"); await settle();
      expect(host.calls).toEqual([menu === "empty" ? "plant 1,0 strain_a" : "harvest 0,0"]);
      expect(document.activeElement).toBe(plot);
    } finally { await host.dispose(); }
  });
}

for (const key of ["{Enter}", " "]) {
  for (const action of ["menu-harvest", "harvest-all", "substrate"] as const) {
    it.skipIf(!browser)(`Garden pending ${action} retains native ${JSON.stringify(key)} focus and refuses repeats`, async testContext => {
      await import("vitest/browser");
      const host = mounted(testContext, { current: async () => active() });
      try {
        await settle();
        host.holdCommands();
        const origin = host.target.querySelector<HTMLButtonElement>("button.cell")!;
        origin.focus();
        let expectedFocus: HTMLButtonElement = origin;
        if (action === "menu-harvest") {
          await host.press("{Enter}{Tab}");
          expect(document.activeElement?.textContent?.trim()).toBe("Harvest");
        } else {
          await host.press(action === "harvest-all" ? "{Tab}" : "{Tab}{Tab}{Tab}{Tab}{Tab}");
          expectedFocus = document.activeElement as HTMLButtonElement;
          expect(expectedFocus.textContent?.trim()).toBe(action === "harvest-all" ? "Harvest all mature" : "Mainframe");
        }
        await host.press(key);
        await settle();
        const expected = action === "menu-harvest" ? "harvest 0,0" : action === "harvest-all" ? "harvest 0,0 0,1 1,1" : "substrate mainframe";
        expect(host.calls).toEqual([expected]);
        expect(document.activeElement).toBe(expectedFocus);
        expect(expectedFocus.disabled).toBe(false);
        expect(expectedFocus.getAttribute("aria-disabled")).toBe("true");
        expect(host.target.querySelector(".garden")?.getAttribute("aria-busy")).toBe("true");
        expect(host.target.querySelector(".live")?.textContent).toBe(t("common.pending", {}, "era_1995"));
        await host.press("{Enter} ");
        expectedFocus.click();
        await settle();
        expect(host.calls).toEqual([expected]);
        expect(host.target.querySelector(".menu")).toBeNull();
        host.app.update({ pending: false });
        await settle();
        expect(document.activeElement).toBe(expectedFocus);
        expect(expectedFocus.getAttribute("aria-disabled")).not.toBe("true");
        expect(host.target.querySelector(".live")?.textContent).toBe("");
        if (action === "menu-harvest") await host.press("{Enter}{Tab}");
        await host.press(key);
        await settle();
        expect(host.calls).toEqual([expected, expected]);
      } finally { await host.dispose(); }
    });
  }
}

for (const action of ["plant", "uproot", "harvest"] as const) {
  it.skipIf(!browser)(`Garden open-menu ${action} stays focusable but cannot dispatch while pending`, async testContext => {
    await import("vitest/browser");
    const host = mounted(testContext, { current: async () => active() });
    let phase = "open menu";
    const element = (node: EventTarget | null) => node instanceof HTMLElement ? {
      tag: node.tagName, id: node.id, text: node instanceof HTMLButtonElement ? node.textContent?.trim() : null,
      connected: node.isConnected,
      disabled: node instanceof HTMLButtonElement ? node.disabled : null,
      ariaDisabled: node.getAttribute("aria-disabled"),
    } : null;
    const input: unknown[] = [];
    const observe = (event: Event) => input.push({
      phase, type: event.type, key: event instanceof KeyboardEvent ? event.key : null,
      target: element(event.target), focused: element(document.activeElement),
    });
    const events = ["keydown", "keyup", "click", "focusin", "focusout"] as const;
    for (const event of events) document.addEventListener(event, observe, true);
    try {
      await settle();
      const cells = [...host.target.querySelectorAll<HTMLButtonElement>("button.cell")];
      cells[0]!.focus();
      if (action === "plant") await host.press("{ArrowDown}");
      await host.press("{Enter}{Tab}");
      if (action === "uproot") await host.press("{Tab}");
      const trigger = document.activeElement as HTMLButtonElement;
      expect(trigger.textContent?.trim()).toBe(action === "plant" ? "Plant Strain A (PENDING OWNER NAME)" : action === "uproot" ? "Uproot" : "Harvest");
      phase = "pending activation refused";
      host.app.update({ pending: true });
      await settle();
      expect(document.activeElement).toBe(trigger);
      expect(trigger.disabled).toBe(false);
      expect(trigger.getAttribute("aria-disabled")).toBe("true");
      await host.press("{Enter} ");
      trigger.click();
      await settle();
      expect(host.calls).toEqual([]);
      expect(host.target.querySelector(".menu")).not.toBeNull();
      phase = "reenable";
      host.app.update({ pending: false });
      await settle();
      expect(document.activeElement).toBe(trigger);
      expect(trigger.disabled).toBe(false);
      expect(trigger.getAttribute("aria-disabled")).not.toBe("true");
      phase = "fresh Enter";
      await host.press("{Enter}");
      await settle();
      expect(host.calls).toEqual([action === "plant" ? "plant 1,0 strain_a" : action === "uproot" ? "uproot 0,0" : "harvest 0,0"]);
      expect(host.target.querySelector(".menu")).toBeNull();
      expect(document.activeElement).toBe(cells[action === "plant" ? 6 : 0]);
    } catch (cause) {
      throw new Error(`Garden ${action} input failure: ${JSON.stringify({
        phase, focused: element(document.activeElement), calls: host.calls,
        menu: host.target.querySelector(".menu")?.textContent?.trim() ?? null, input,
      })}`, { cause });
    } finally {
      for (const event of events) document.removeEventListener(event, observe, true);
      await host.dispose();
    }
  });
}

for (const key of ["{Enter}", " "]) for (const removed of [false, true]) {
  it.skipIf(!browser)(`Garden menu completion preserves a newer ${removed ? "removed" : "retained"} focus choice after native ${JSON.stringify(key)}`, async testContext => {
    await import("vitest/browser");
    const host = mounted(testContext, { current: async () => active() });
    const newerChoice = host.own(document.createElement("button"));
    newerChoice.textContent = "Another control";
    document.body.append(newerChoice);
    try {
      await settle();
      host.target.querySelector<HTMLButtonElement>("button.cell")!.focus();
      await host.press("{Enter}{Tab}");
      expect(document.activeElement?.textContent?.trim()).toBe("Harvest");
      // A focus choice made after submission but before the menu's async
      // render/handoff completes must win. No real server response is mocked.
      host.holdCommands(() => {
        newerChoice.focus();
        expect(document.activeElement, "the newer control actually receives focus").toBe(newerChoice);
        if (removed) newerChoice.remove();
      });
      await host.press(key);
      await settle();
      expect(host.calls).toEqual(["harvest 0,0"]);
      expect(host.target.querySelector(".menu")).toBeNull();
      const expectedFocus = removed ? document.body : newerChoice;
      expect(document.activeElement).toBe(expectedFocus);
      host.app.update({ pending: false });
      await settle();
      expect(document.activeElement).toBe(expectedFocus);
    } finally { await host.dispose(); newerChoice.remove(); }
  });
}

for (const removed of [false, true]) {
  it.skipIf(!browser)(`Garden Close handoff preserves a newer ${removed ? "removed" : "retained"} focus choice`, async testContext => {
    const host = mounted(testContext, { current: async () => active() });
    const newerChoice = host.own(document.createElement("button")); document.body.append(newerChoice);
    try {
      await settle();
      host.target.querySelector<HTMLButtonElement>("button.cell")!.click(); await settle();
      const close = [...host.target.querySelectorAll<HTMLButtonElement>(".menu button")]
        .find((candidate) => candidate.textContent?.trim() === "Close")!;
      close.focus(); close.click();
      // Controlled post-handler/pre-render lifecycle race, not native traversal.
      newerChoice.focus(); expect(document.activeElement).toBe(newerChoice);
      if (removed) newerChoice.remove();
      await settle();
      expect(document.activeElement).toBe(removed ? document.body : newerChoice);
      expect(host.target.querySelector(".menu")).toBeNull();
      expect(host.calls).toEqual([]);
    } finally { await host.dispose(); newerChoice.remove(); }
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

async function nativeWalk(target: HTMLElement, keys: readonly string[], expectedFocus: readonly number[], press: (keys: string) => Promise<void>): Promise<HTMLButtonElement[]> {
  await import("vitest/browser");
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
    await press(keys.join(""));
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
  it.skipIf(!browser)(`Garden native navigation: ${path.name}`, async testContext => {
    const started = performance.now();
    const { target, app, calls, press, dispose } = mounted(testContext, { current: async () => active() });
    try {
      await settle();
      await nativeWalk(target, path.keys, path.focus, press);
      expect(calls).toEqual([]);
      expect(target.querySelector(".menu")).toBeNull();
    } finally {
      await dispose();
      console.info("garden-native-navigation", JSON.stringify({ path: path.name, user_agent: navigator.userAgent, elapsed_ms: performance.now() - started }));
    }
  });
}

for (const key of ["{Enter}", " "]) {
  it.skipIf(!browser)(`Garden native ${JSON.stringify(key)} navigates and dispatches once, with pending refusal`, async testContext => {
    const started = performance.now();
    const observed = (stage: string) => console.info("garden-native-stage", JSON.stringify({ key, stage, user_agent: navigator.userAgent, elapsed_ms: performance.now() - started }));
    observed("entry");
    await import("vitest/browser");
    observed("helper-import");
    const { target, app, calls, press, dispose } = mounted(testContext, { current: async () => active() });
    try {
      await settle();
      observed("initial-mount");
      const cells = await nativeWalk(target, mixedKeys, [1, 7, 6, 0], press);
      observed("navigation");
      await press(key);
      await settle();
      observed("menu-open");
      expect(target.querySelector(".menu")).not.toBeNull();
      expect(calls).toEqual([]);
      await press("{Tab}");
      observed("tab");
      expect(document.activeElement?.textContent?.trim()).toBe("Harvest");
      await press(key);
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
      await press(key);
      await settle();
      expect(calls).toEqual(["harvest 0,0"]);
      observed("pending-refusal");
    } finally { await dispose(); observed("cleanup"); }
  });
}

for (const empty of [false, true]) {
  it.skipIf(!browser)(`Garden native Tab and Shift-Tab traverse the ${empty ? "seed" : "mature"} menu and substrates`, async testContext => {
    await import("vitest/browser");
    const { target, app, calls, press, dispose } = mounted(testContext, { current: async () => active() });
    try {
      await settle();
      const cells = [...target.querySelectorAll<HTMLButtonElement>("button.cell")];
      cells[0]!.focus();
      if (empty) { await press("{ArrowDown}"); await settle(); }
      const origin = cells[empty ? 6 : 0]!;
      expect(document.activeElement).toBe(origin);
      await press("{Enter}"); await settle();
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
        await press("{Tab}");
        expect(document.activeElement).toBe(control);
      }
      for (const control of [...path.slice(0, -1)].reverse().concat(origin)) {
        await press("{Shift>}{Tab}{/Shift}");
        expect(document.activeElement).toBe(control);
      }
      expect(calls).toEqual([]);
      expect(cells.filter((cell) => cell.tabIndex === 0)).toEqual([origin]);
    } finally { await dispose(); }
  });
}

for (const key of ["{Enter}", " "]) {
  for (const [empty, at, expected] of [
    [false, 1, "uproot 0,0"], [false, 2, null],
    [true, 0, "plant 1,0 strain_a"], [true, 1, "plant 1,0 strain_b"],
  ] as const) {
    it.skipIf(!browser)(`Garden native ${JSON.stringify(key)} reaches ${empty ? "seed" : "mature"} menu action ${at}`, async testContext => {
      await import("vitest/browser");
      const { target, app, calls, press, dispose } = mounted(testContext, { current: async () => active() });
      try {
        await settle();
        const cells = [...target.querySelectorAll<HTMLButtonElement>("button.cell")];
        cells[0]!.focus();
        if (empty) { await press("{ArrowDown}"); await settle(); }
        const origin = cells[empty ? 6 : 0]!;
        await press(key); await settle();
        const action = target.querySelectorAll<HTMLButtonElement>(".menu button")[at]!;
        for (let index = 0; index <= at; index++) await press("{Tab}");
        expect(document.activeElement).toBe(action);
        expect(calls).toEqual([]);
        await press(key); await settle();
        expect(calls).toEqual(expected === null ? [] : [expected]);
        expect(target.querySelector(".menu")).toBeNull();
        expect(document.activeElement).toBe(origin);
      } finally { await dispose(); }
    });
  }
}

for (const key of ["{Enter}", " "]) {
  it.skipIf(!browser)(`Garden native ${JSON.stringify(key)} reaches Harvest all and every substrate`, async testContext => {
    await import("vitest/browser");
    const { target, app, calls, press, dispose } = mounted(testContext, { current: async () => active() });
    try {
      await settle();
      target.querySelector<HTMLButtonElement>("button.cell")!.focus();
      await press("{Tab}");
      expect(document.activeElement?.textContent?.trim()).toBe("Harvest all mature");
      await press(key); await settle();
      expect(calls).toEqual(["harvest 0,0 0,1 1,1"]);
      const substrates = [...target.querySelectorAll<HTMLButtonElement>(".substrates button")];
      for (const [index, control] of substrates.entries()) {
        await press("{Tab}");
        expect(document.activeElement).toBe(control);
        await press(key); await settle();
        // The read still owns Bare metal; selecting it is a no-op. Other
        // selections dispatch exactly once, without an optimistic view patch.
        expect(calls).toEqual(["harvest 0,0 0,1 1,1", ...["chaos_monkey", "containerized", "mainframe"].slice(0, index).map((id) => `substrate ${id}`)]);
        expect(pressed(target)).toBe("Bare metal");
      }
    } finally { await dispose(); }
  });

  it.skipIf(!browser)(`Garden native ${JSON.stringify(key)} reaches Refresh after a failed read`, async testContext => {
    await import("vitest/browser");
    let reads = 0;
    const { target, app, calls, press, dispose, own } = mounted(testContext, { current: async () => {
      if (++reads === 1) throw new Error("controlled read failure");
      return active();
    } });
    const before = own(document.createElement("input")); target.before(before);
    try {
      await settle();
      expect(reads).toBe(1);
      before.focus(); await press("{Tab}");
      expect(document.activeElement).toBe(target.querySelector(".garden-help summary"));
      await press("{Tab}");
      expect(document.activeElement?.textContent?.trim()).toBe("Refresh");
      await press(key); await settle();
      expect(reads).toBe(2);
      expect(target.querySelectorAll("button.cell")).toHaveLength(36);
      expect(calls).toEqual([]);
    } finally { await dispose(); before.remove(); }
  });
}

it.skipIf(!browser)("Garden receipt refresh sequential control displays the newer read", async testContext => {
  const port = new OrderedPort();
  const { target, app, calls, press, dispose } = mounted(testContext, port);
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
  } finally { await dispose(); }
});

for (const outcome of ["active", "locked", "error", "old-error"] as const) {
  it.skipIf(!browser)(`Garden receipt refresh retains newest ${outcome} after late older completion`, async testContext => {
    const port = new OrderedPort();
    const { target, app, calls, press, dispose } = mounted(testContext, port);
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
    } finally { await dispose(); }
  });
}

it.skipIf(!browser)("Garden ignores late read completion after unmount", async testContext => {
  const port = new OrderedPort();
  const { target, app, calls, press, dispose } = mounted(testContext, port);
  let destroyed = false;
  try {
    await settle();
    expect(port.requests).toHaveLength(1);
    await dispose();
    destroyed = true;
    port.requests[0]!.resolve(active());
    await settle();
    expect(target.childElementCount).toBe(0);
    expect(calls).toEqual([]);
  } finally {
    if (!destroyed) await dispose();
    target.remove();
  }
});
