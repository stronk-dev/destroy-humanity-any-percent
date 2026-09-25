import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it } from "vitest";

import candidateRaw from "../../balance/testdata/arcade-v1.json?raw";
import fixtureRaw from "../../testdata/arcade/corpus-fixture-v1.json?raw";
import { arcadeContentHash, parseArcadeCatalog } from "../src/arcade/catalog";
import { applyMineGrid, createMineGrid, decodeMineGridSnapshot, MINE_GRID_SCALING_DESTINATION, type MineGridCommand, type MineGridSnapshot } from "../src/arcade/mine-grid";
import { applySnake, createSnake, decodeSnakeSnapshot, SNAKE_SCALING_DESTINATION, type SnakeCommand, type SnakeSnapshot } from "../src/arcade/snake";
import { COPY_KEYS } from "../src/copy";
import MineGridBoard from "../src/game-ui/minigame/MineGridBoard.svelte";
import SnakeBoard from "../src/game-ui/minigame/SnakeBoard.svelte";
import { installTheme, UI_THEMES } from "../src/ui/themes";

const browser = typeof document !== "undefined";
const catalog = parseArcadeCatalog(JSON.parse(candidateRaw), new Set(COPY_KEYS));

async function settle(ms = 0): Promise<void> {
  if (ms > 0) await new Promise((resolve) => setTimeout(resolve, ms));
  for (let index = 0; index < 5; index += 1) { await new Promise((resolve) => setTimeout(resolve, 0)); await tick(); flushSync(); }
}

async function assertAxe(target: HTMLElement, label: string): Promise<void> {
  const result = await axe.run(target, { runOnly: { type: "tag", values: ["wcag2a", "wcag2aa", "wcag21aa", "wcag22aa"] } });
  expect(result.violations.filter((violation) => violation.impact === "serious" || violation.impact === "critical"), label).toEqual([]);
}

function host(): HTMLElement {
  const target = document.createElement("main");
  document.body.append(target);
  installTheme(target, UI_THEMES.era_1995, false);
  return target;
}

async function mineIdentity(content = candidateRaw) {
  return { content, content_hash: await arcadeContentHash(content), content_schema_version: 1, seed: 7n, mode: "solo" as const,
    scaling_inputs: { [MINE_GRID_SCALING_DESTINATION]: 1 } };
}

// A fake host: applies each command through the real TS engine.
class MineHost {
  snapshot = "";
  revision = 1;
  commands: MineGridCommand[] = [];
  constructor(private identity: Awaited<ReturnType<typeof mineIdentity>>) {}
  async init(): Promise<this> { this.snapshot = await createMineGrid(this.identity); return this; }
  async apply(command: MineGridCommand): Promise<void> {
    this.commands.push(command);
    const output = await applyMineGrid({ ...this.identity, revision: this.revision, snapshot: this.snapshot, command: JSON.stringify(command) });
    this.snapshot = output.snapshot; this.revision++;
  }
  parsed(): MineGridSnapshot { return decodeMineGridSnapshot(this.snapshot); }
}

function mountMine(target: HTMLElement, host: MineHost, onCommand: (command: MineGridCommand) => void = (command) => { host.commands.push(command); }) {
  return mount(MineGridBoard, { target, props: { snapshot: host.parsed(), presets: catalog.mine_grid.presets, era: "era_1995", pending: false, onCommand } });
}

function button(target: HTMLElement, text: string): HTMLButtonElement {
  const found = [...target.querySelectorAll("button")].find((candidate) => candidate.textContent?.trim() === text);
  if (!found) throw new Error(`missing button ${text}: ${target.textContent}`);
  return found;
}

it.skipIf(!browser)("mine_grid: setup lists presets, the grid is a roving-tabindex grid, and keys act without a pointer (AR6.2)", async () => {
  const target = host();
  const game = await new MineHost(await mineIdentity()).init();
  let app = mountMine(target, game);
  try {
    await settle();
    await assertAxe(target, "setup");
    button(target, "Small board").click();
    expect(game.commands).toEqual([{ kind: "choose_board", preset_id: "small" }]);
    unmount(app);
    await game.apply({ kind: "choose_board", preset_id: "small" });
    await game.apply({ kind: "reveal", cell: 40 });
    const sent: MineGridCommand[] = [];
    app = mountMine(target, game, (command) => sent.push(command));
    await settle();
    const cells = [...target.querySelectorAll<HTMLButtonElement>("[role=gridcell]")];
    expect(cells).toHaveLength(81);
    expect(target.querySelector("[role=grid]")).not.toBeNull();
    expect(cells.filter((cell) => cell.tabIndex === 0)).toHaveLength(1);
    // Revealed cells announce their count; nothing names a mine before terminal.
    expect(cells[40]!.getAttribute("aria-label")).toMatch(/Row 5, column 5: \d+ nearby/u);
    expect(target.querySelectorAll("[data-state=mine]")).toHaveLength(0);
    expect(target.querySelector("[role=status]")?.textContent).toMatch(/cells open/u);
    await assertAxe(target, "playing");
    const hidden = cells.findIndex((cell) => cell.getAttribute("aria-label")?.endsWith("hidden"));
    cells[hidden]!.focus();
    cells[hidden]!.dispatchEvent(new KeyboardEvent("keydown", { key: "f", bubbles: true }));
    cells[hidden]!.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowRight", bubbles: true }));
    await settle();
    expect(document.activeElement).not.toBe(cells[hidden]);
    cells[40]!.dispatchEvent(new KeyboardEvent("keydown", { key: "c", bubbles: true }));
    button(target, "Flag").click();
    await settle();
    cells[hidden]!.click();
    expect(sent).toEqual([{ kind: "toggle_flag", cell: hidden }, { kind: "chord", cell: 40 }, { kind: "toggle_flag", cell: hidden }]);
    // Quit asks first, then shows every mine with a non-color glyph.
    button(target, "Quit").click();
    await settle();
    expect(target.textContent).toContain("Nothing is lost.");
    unmount(app);
    await game.apply({ kind: "quit" });
    app = mountMine(target, game);
    await settle();
    const mines = target.querySelectorAll("[data-state=mine]");
    expect(mines).toHaveLength(10);
    expect(mines[0]!.textContent?.trim()).toBe("X");
    expect(target.querySelector("[role=status]")?.textContent).toContain("Game ended.");
    await assertAxe(target, "terminal");
  } finally { unmount(app); target.remove(); }
});

async function snakeIdentity(content: string) {
  return { content, content_hash: await arcadeContentHash(content), content_schema_version: 1, seed: 505n, mode: "solo" as const,
    scaling_inputs: { [SNAKE_SCALING_DESTINATION]: 1 } };
}

class SnakeServer {
  snapshot = "";
  revision = 1;
  submitted: SnakeCommand[] = [];
  currentCalls = 0;
  failNext = false;
  block = false;
  constructor(private identity: Awaited<ReturnType<typeof snakeIdentity>>) {}
  async init(): Promise<this> { this.snapshot = await createSnake(this.identity); return this; }
  parsed(): SnakeSnapshot { return decodeSnakeSnapshot(this.snapshot); }
  submit = async (command: SnakeCommand): Promise<SnakeSnapshot> => {
    this.submitted.push(command);
    if (this.block) return new Promise(() => {});
    if (this.failNext) { this.failNext = false; throw new Error("rejected"); }
    const output = await applySnake({ ...this.identity, revision: this.revision, snapshot: this.snapshot, command: JSON.stringify(command) });
    this.snapshot = output.snapshot; this.revision++;
    return this.parsed();
  };
  current = async (): Promise<SnakeSnapshot> => { this.currentCalls++; return this.parsed(); };
}

function mountSnake(target: HTMLElement, server: SnakeServer, content: ReturnType<typeof parseArcadeCatalog>["snake"], extra: Record<string, unknown> = {}) {
  return mount(SnakeBoard, { target, props: { initial: server.parsed(), content, era: "era_1995", submit: server.submit, current: server.current, tickMS: 20, ...extra } });
}

it.skipIf(!browser)("snake: D-pad steering flushes one exact advance at the terminal tick, validated by the engine (AR6.3)", async () => {
  const content = parseArcadeCatalog(JSON.parse(fixtureRaw), new Set(COPY_KEYS)).snake;
  const server = await new SnakeServer(await snakeIdentity(fixtureRaw)).init();
  const target = host();
  const app = mountSnake(target, server, content);
  try {
    await settle();
    await assertAxe(target, "snake ready");
    expect(target.textContent).toContain("Paused");
    button(target, "Up").click();
    await settle(400);
    // The 6x5 head at (3,2) turns north at tick 1 and hits the wall at tick 3.
    expect(server.submitted).toEqual([{ kind: "advance", through_tick: 3, turns: [{ tick: 1, direction: "up" }] }]);
    expect(server.parsed().phase).toBe("terminal");
    expect(target.querySelector("[role=status]")?.textContent).toContain("Game over");
    await assertAxe(target, "snake terminal");
  } finally { unmount(app); target.remove(); }
});

it.skipIf(!browser)("snake: batches every flushEvery ticks, auto-pauses on blur, and resyncs on a rejected flush", async () => {
  const content = catalog.snake;
  const server = await new SnakeServer(await snakeIdentity(candidateRaw)).init();
  const target = host();
  const app = mountSnake(target, server, content, { flushEvery: 4, tickMS: 40 });
  try {
    await settle();
    button(target, "Play").click();
    await settle(190);
    expect(server.submitted[0]).toEqual({ kind: "advance", through_tick: 4, turns: [] });
    window.dispatchEvent(new Event("blur"));
    await settle();
    expect(target.textContent).toContain("Paused");
    const beforePause = server.submitted.length;
    await settle(200);
    expect(server.submitted.length).toBe(beforePause);
    server.failNext = true;
    button(target, "Resume").click();
    await settle(180);
    expect(server.currentCalls).toBeGreaterThan(0);
    expect(target.textContent).toContain("The game caught up with the server.");
  } finally { unmount(app); target.remove(); }
});

it.skipIf(!browser)("snake: freezes at max_ticks_per_advance while a flush is unacknowledged", async () => {
  const content = { ...catalog.snake, max_ticks_per_advance: 5 };
  const server = await new SnakeServer(await snakeIdentity(candidateRaw)).init();
  server.block = true;
  const target = host();
  const app = mountSnake(target, server, content, { flushEvery: 100 });
  try {
    await settle();
    button(target, "Play").click();
    await settle(300);
    const head = () => [...target.querySelectorAll(".cell")].findIndex((cell) => cell.getAttribute("data-state") === "head");
    const first = head();
    await settle(200);
    expect(head()).toBe(first);
    // Only the one frozen flush is outstanding: at most one command in flight.
    expect(server.submitted).toEqual([{ kind: "advance", through_tick: 5, turns: [] }]);
  } finally { unmount(app); target.remove(); }
});
