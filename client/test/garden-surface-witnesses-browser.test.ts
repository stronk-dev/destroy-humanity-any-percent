import { flushSync, mount, tick, unmount, type ComponentProps } from "svelte";
import { expect, it } from "vitest";

import views from "../../testdata/garden/view-fixtures-v1.json";
import type { GardenCurrentResponse } from "../src/api/generated/types";
import GardenSurface from "../src/game-ui/garden/GardenSurface.svelte";
import type { GardenPort } from "../src/game-ui/garden/garden-port";
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
  const target = document.createElement("main");
  document.body.append(target);
  installTheme(target, UI_THEMES.era_1995, false);
  const initial: ComponentProps<typeof GardenSurface> = {
    port, era: "era_1995", pending: false, refreshKey: 0, rejection: null, visible: () => true,
    onPlant: (row: number, col: number, species: string) => calls.push(`plant ${row},${col} ${species}`),
    onUproot: (row: number, col: number) => calls.push(`uproot ${row},${col}`),
    onHarvest: (plots: readonly { row: number; col: number }[]) => calls.push(`harvest ${plots.map(({ row, col }) => `${row},${col}`).join(" ")}`),
    onSetSubstrate: (substrate: string) => calls.push(`substrate ${substrate}`),
  };
  const app = mount(GardenSurfaceHarness, { target, props: { initial } });
  return { target, app, calls };
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

for (const key of ["{Enter}", " "]) {
  it.skipIf(!browser)(`Garden native ${JSON.stringify(key)} navigates and dispatches once, with pending refusal`, async () => {
    const { userEvent } = await import("vitest/browser");
    const { target, app, calls } = mounted({ current: async () => active() });
    try {
      await settle();
      const cells = [...target.querySelectorAll<HTMLButtonElement>("button.cell")];
      expect(cells).toHaveLength(36);
      cells[0]!.focus();
      for (const [arrow, index] of [["{ArrowUp}", 0], ["{ArrowLeft}", 0], ["{ArrowRight}", 1],
        ["{ArrowDown}", 7], ["{ArrowLeft}", 6], ["{ArrowUp}", 0]] as const) {
        await userEvent.keyboard(arrow);
        await settle();
        expect(document.activeElement, arrow).toBe(cells[index]);
        expect(cells.filter((cell) => cell.tabIndex === 0), arrow).toEqual([cells[index]]);
      }
      await userEvent.keyboard(key);
      await settle();
      expect(target.querySelector(".menu")).not.toBeNull();
      expect(calls).toEqual([]);
      await userEvent.keyboard("{Tab}");
      expect(document.activeElement?.textContent?.trim()).toBe("Harvest");
      await userEvent.keyboard(key);
      await settle();
      expect(calls).toEqual(["harvest 0,0"]);
      expect(target.querySelector(".menu")).toBeNull();
      expect(document.activeElement).toBe(cells[0]);
      app.update({ pending: true });
      await settle();
      expect(cells.every((cell) => cell.disabled)).toBe(true);
      expect([...target.querySelectorAll<HTMLButtonElement>(".substrates button")].every((button) => button.matches(":disabled"))).toBe(true);
      await userEvent.keyboard(key);
      await settle();
      expect(calls).toEqual(["harvest 0,0"]);
    } finally { await unmount(app); target.remove(); }
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
  await settle();
  expect(port.requests).toHaveLength(1);
  await unmount(app);
  port.requests[0]!.resolve(active());
  await settle();
  expect(target.childElementCount).toBe(0);
  expect(calls).toEqual([]);
  target.remove();
});
