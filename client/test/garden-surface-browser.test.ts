import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import views from "../../testdata/garden/view-fixtures-v1.json";
import type { GardenCurrentResponse } from "../src/api/generated/types";
import GardenSurface from "../src/game-ui/garden/GardenSurface.svelte";
import type { GardenPort } from "../src/game-ui/garden/garden-port";
import { installTheme, UI_THEMES } from "../src/ui/themes";

const browser = typeof document !== "undefined";

class FakePort implements GardenPort {
  reads = 0;
  constructor(public responses: GardenCurrentResponse[]) {}
  async current(): Promise<GardenCurrentResponse> { this.reads += 1; return this.responses[Math.min(this.reads - 1, this.responses.length - 1)]!; }
}

async function settle(): Promise<void> { for (let index = 0; index < 5; index += 1) { await new Promise((resolve) => setTimeout(resolve, 0)); await tick(); flushSync(); } }

async function assertAxe(target: HTMLElement, label: string): Promise<void> {
  const result = await axe.run(target, { runOnly: { type: "tag", values: ["wcag2a", "wcag2aa", "wcag21aa", "wcag22aa"] } });
  expect(result.violations.filter((violation) => violation.impact === "serious" || violation.impact === "critical"), label).toEqual([]);
}

function host(width?: number): HTMLElement {
  const target = document.createElement("main");
  if (width) target.style.width = `${width}px`;
  document.body.append(target);
  installTheme(target, UI_THEMES.era_1995, false);
  return target;
}

function mountSurface(port: GardenPort, calls: string[], width?: number) {
  const target = host(width);
  const app = mount(GardenSurface, { target, props: { port, era: "era_1995", pending: false, refreshKey: 0, rejection: null, visible: () => true,
    onPlant: (row: number, col: number, species: string) => calls.push(`plant ${row},${col} ${species}`),
    onUproot: (row: number, col: number) => calls.push(`uproot ${row},${col}`),
    onHarvest: (plots: readonly { row: number; col: number }[]) => calls.push(`harvest ${plots.map((plot) => `${plot.row},${plot.col}`).join(" ")}`),
    onSetSubstrate: (substrate: string) => calls.push(`substrate ${substrate}`) } });
  return { target, app };
}

const active = views.active as GardenCurrentResponse;

it.skipIf(!browser)("renders the active garden as an accessible grid with non-colour stages and dispatches through callbacks only", async () => {
  const calls: string[] = [];
  const { target, app } = mountSurface(new FakePort([active]), calls);
  try {
    await settle();
    await assertAxe(target, "active");
    const grid = target.querySelector("[role=grid]")!;
    expect(grid.querySelectorAll("[role=row]")).toHaveLength(6);
    const cells = [...grid.querySelectorAll<HTMLButtonElement>("button.cell")];
    expect(cells).toHaveLength(36);
    // Roving tabindex: exactly one cell is in the tab order.
    expect(cells.filter((cell) => cell.tabIndex === 0)).toHaveLength(1);
    // Stage is text (visible glyph word + accessible name), never colour alone.
    const first = cells[0]!;
    expect(first.dataset.stage).toBe("mature");
    expect(first.textContent?.trim()).toBe("Bloom");
    expect(first.getAttribute("aria-label")).toContain("Mature");
    const dormant = cells[35]!;
    expect(dormant.dataset.stage).toBe("dormant");
    expect(dormant.getAttribute("aria-label")).toContain("Dormant");
    first.focus();
    first.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowRight", bubbles: true }));
    await settle();
    expect(document.activeElement).toBe(cells[1]);
    expect(cells[1]!.tabIndex).toBe(0);
    cells[0]!.click();
    await settle();
    const harvest = [...target.querySelectorAll("button")].find((button) => button.textContent?.trim() === "Harvest")!;
    harvest.click();
    await settle();
    expect(calls).toEqual(["harvest 0,0"]);
    expect(document.activeElement).toBe(cells[0]);
    const all = [...target.querySelectorAll("button")].find((button) => button.textContent?.trim() === "Harvest all mature")!;
    all.click();
    expect(calls.at(-1)).toMatch(/^harvest 0,0 0,1 1,1$/u);
    // Planting an empty active cell offers only collected seeds.
    const empty = cells.find((cell) => cell.dataset.stage === "empty")!;
    empty.click();
    await settle();
    await assertAxe(target, "plant menu");
    const plants = [...target.querySelectorAll(".menu button")].filter((button) => button.textContent?.startsWith("Plant"));
    expect(plants).toHaveLength(2);
    (plants[0] as HTMLButtonElement).click();
    expect(calls.at(-1)).toMatch(/^plant \d,\d strain_a$/u);
    const mainframe = [...target.querySelectorAll(".substrates button")].find((button) => button.textContent?.trim() === "Mainframe") as HTMLButtonElement;
    mainframe.click();
    expect(calls.at(-1)).toBe("substrate mainframe");
    expect(target.textContent).not.toContain("0123456789abcdef");
  } finally { unmount(app); target.remove(); }
});

it.skipIf(!browser)("announces matured and spawned plants once per refresh", async () => {
  const before = structuredClone(active) as Extract<GardenCurrentResponse, { kind: "active" }>;
  before.garden.plots = before.garden.plots.filter((plot) => !(plot.row === 0 && plot.col === 1))
    .map((plot) => plot.row === 1 && plot.col === 1 ? { ...plot, stage: "growing" as const } : plot);
  const port = new FakePort([before, active]);
  const { target, app } = mountSurface(port, []);
  try {
    await settle();
    expect(target.querySelector(".live")?.textContent).toBe("");
    document.dispatchEvent(new Event("visibilitychange"));
    await settle();
    expect(port.reads).toBe(2);
    expect(target.querySelector(".live")?.textContent).toBe("1 matured, 1 new seedlings");
    expect(target.querySelector(".live")?.getAttribute("aria-live")).toBe("polite");
  } finally { unmount(app); target.remove(); }
});

it.skipIf(!browser)("renders locked and error states accessibly", async () => {
  for (const [label, responses, text] of [["locked", [views.locked], "Locked"], ["error", [views.inactive], "could not be loaded"]] as const) {
    const { target, app } = mountSurface(new FakePort(responses as unknown as GardenCurrentResponse[]), []);
    try {
      await settle();
      expect(target.textContent).toContain(text);
      await assertAxe(target, label);
    } finally { unmount(app); target.remove(); }
  }
});

it.skipIf(!browser)("reflows a 6x6 grid at 320 CSS px without horizontal overflow and keeps 24px targets", async () => {
  const { target, app } = mountSurface(new FakePort([active]), [], 320);
  try {
    await settle();
    const grid = target.querySelector<HTMLElement>("[role=grid]")!;
    expect(grid.scrollWidth).toBeLessThanOrEqual(target.clientWidth);
    for (const cell of target.querySelectorAll<HTMLElement>("button.cell")) {
      const box = cell.getBoundingClientRect();
      expect(box.width).toBeGreaterThanOrEqual(24);
      expect(box.height).toBeGreaterThanOrEqual(24);
    }
  } finally { unmount(app); target.remove(); }
});
