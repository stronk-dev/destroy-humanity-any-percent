import axe from "axe-core";
import { flushSync, mount, tick, unmount } from "svelte";
import { expect, it, vi } from "vitest";

import candidateRaw from "../../balance/testdata/arcade-v1.json?raw";
import fixtureRaw from "../../testdata/arcade/corpus-fixture-v1.json?raw";
import corpus from "../../testdata/arcade/content-gate-v2.json";
import { arcadeContentHash, parseArcadeCatalog } from "../src/arcade/catalog";
import { applyMineGrid, createMineGrid, decodeMineGridSnapshot, MINE_GRID_SCALING_DESTINATION, type MineGridCommand, type MineGridSnapshot } from "../src/arcade/mine-grid";
import { applySnake, createSnake, decodeSnakeSnapshot, SNAKE_SCALING_DESTINATION, type SnakeCommand, type SnakeSnapshot } from "../src/arcade/snake";
import { COPY_KEYS, t } from "../src/copy";
import MineGridBoard from "../src/game-ui/minigame/MineGridBoard.svelte";
import SnakeBoard from "../src/game-ui/minigame/SnakeBoard.svelte";
import { installTheme, UI_THEMES } from "../src/ui/themes";
import MineGridBoardHarness from "./MineGridBoardHarness.svelte";

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

async function mineIdentity(content = candidateRaw, seed = 7n) {
  return { content, content_hash: await arcadeContentHash(content), content_schema_version: 1, seed, mode: "solo" as const,
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

it.skipIf(!browser)("mine_grid: native keyboard completes the corpus game in one mount (AR6.2, AC12)", async () => {
  const { userEvent } = await import("vitest/browser");
  const scenario = corpus.scenarios.find((row) => row.name === "mine_grid_clear_small")!;
  expect(scenario.engine).toBe("mine_grid");
  expect(scenario.steps.map((row) => row.command)).toEqual([
    { kind: "choose_board", preset_id: "small" }, { kind: "reveal", cell: 12 },
  ]);
  const fixture = parseArcadeCatalog(JSON.parse(fixtureRaw), new Set(COPY_KEYS));
  const game = await new MineHost(await mineIdentity(fixtureRaw, BigInt(scenario.seed))).init();
  const target = host();
  let response: Promise<void> | undefined;
  const app = mount(MineGridBoardHarness, { target, props: {
    initial: game.parsed(), presets: fixture.mine_grid.presets,
    onCommand: (command: MineGridCommand) => {
      // This is the only gameplay entry: the mounted child's host callback.
      response = game.apply(command).then(() => flushSync(() => app.advance(game.parsed())));
    },
  } }) as unknown as { advance(next: MineGridSnapshot): void };
  try {
    await settle();
    button(target, "Small board").focus();
    await userEvent.keyboard("{Enter}");
    expect(game.commands).toEqual([{ kind: "choose_board", preset_id: "small" }]);
    await response; await settle();
    expect(game.parsed().mine_cells).toEqual([]);
    expect(target.querySelectorAll("[data-state=mine]")).toHaveLength(0);
    const cells = [...target.querySelectorAll<HTMLButtonElement>("[role=gridcell]")];
    expect(cells).toHaveLength(25);
    cells[0]!.focus();
    await userEvent.keyboard("{ArrowRight}{ArrowRight}{ArrowDown}{ArrowDown}");
    expect(document.activeElement).toBe(cells[12]);
    expect(cells.filter((cell) => cell.tabIndex === 0)).toEqual([cells[12]]);
    await userEvent.keyboard(" ");
    expect(game.commands).toEqual(scenario.steps.map((row) => row.command));
    await response; await settle();
    expect(game.parsed()).toEqual(scenario.expected_terminal);
    expect(target.querySelector("[role=status]")?.textContent).toContain("Board cleared.");
    expect(target.querySelectorAll("[data-state=revealed]")).toHaveLength(24);
    expect(target.querySelectorAll("[data-state=mine]")).toHaveLength(1);
    await assertAxe(target, "native keyboard cleared");
  } finally { await unmount(app as never); target.remove(); }
});

class SnakeServer {
  snapshot = "";
  revision = 1;
  submitted: SnakeCommand[] = [];
  currentCalls = 0;
  failNext = false;
  block = false;
  lastResponse: Promise<SnakeSnapshot> | undefined;
  constructor(private identity: Awaited<ReturnType<typeof snakeIdentity>>) {}
  async init(): Promise<this> { this.snapshot = await createSnake(this.identity); return this; }
  parsed(): SnakeSnapshot { return decodeSnakeSnapshot(this.snapshot); }
  submit = (command: SnakeCommand): Promise<SnakeSnapshot> => {
    this.lastResponse = this.apply(command);
    return this.lastResponse;
  };
  private async apply(command: SnakeCommand): Promise<SnakeSnapshot> {
    this.submitted.push(command);
    if (this.block) return new Promise(() => {});
    if (this.failNext) { this.failNext = false; throw new Error("rejected"); }
    const output = await applySnake({ ...this.identity, revision: this.revision, snapshot: this.snapshot, command: JSON.stringify(command) });
    this.snapshot = output.snapshot; this.revision++;
    return this.parsed();
  }
  current = async (): Promise<SnakeSnapshot> => { this.currentCalls++; return this.parsed(); };
}

function mountSnake(target: HTMLElement, server: SnakeServer, content: ReturnType<typeof parseArcadeCatalog>["snake"], extra: Record<string, unknown> = {}) {
  return mount(SnakeBoard, { target, props: { initial: server.parsed(), content, era: "era_1995", submit: server.submit, current: server.current, tickMS: 20, ...extra } });
}

function snakeHead(target: HTMLElement): number {
  return [...target.querySelectorAll(".cell")].findIndex((cell) => cell.getAttribute("data-state") === "head");
}

async function deliverSnakeTimers(ms: number): Promise<void> {
  await vi.advanceTimersByTimeAsync(ms);
  await tick();
  flushSync();
}

// Delay acknowledgement, not engine behavior: every delivered command still
// executes through SnakeServer's real shared engine and revision boundary.
function delayedSnakeHost(server: SnakeServer) {
  let release!: () => void;
  const gate = new Promise<void>((resolve) => { release = resolve; });
  let active = 0, peak = 0;
  const submitted: SnakeCommand[] = [], responses: Promise<SnakeSnapshot>[] = [];
  return {
    submitted, responses, release,
    get peak() { return peak; },
    submit(command: SnakeCommand): Promise<SnakeSnapshot> {
      const index = submitted.push(command) - 1;
      active++; peak = Math.max(peak, active);
      const response = (index === 0 ? gate : Promise.resolve()).then(() => server.submit(command)).finally(() => { active--; });
      responses.push(response);
      // Cleanup can release a deliberately broken ordering after an assertion
      // fires; the response is still explicitly awaited by the witness.
      void response.catch(() => {});
      return response;
    },
  };
}

it.skipIf(!browser)("snake: Quit waits for the actual pending advance acknowledgement (AR6.3, MA-C8)", async () => {
  const server = await new SnakeServer(await snakeIdentity(candidateRaw)).init();
  const delayed = delayedSnakeHost(server), target = host();
  vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
  const app = mountSnake(target, server, catalog.snake, { submit: delayed.submit, flushEvery: 4 });
  try {
    await tick(); flushSync();
    button(target, "Play").click();
    await deliverSnakeTimers(80);
    expect(delayed.submitted).toEqual([{ kind: "advance", through_tick: 4, turns: [] }]);
    expect(server.parsed().tick).toBe(0);
    button(target, "Quit").click();
    await tick(); flushSync();
    expect(delayed.submitted, "Quit must not overlap an unacknowledged advance").toHaveLength(1);
    expect(delayed.peak).toBe(1);
    delayed.release();
    await delayed.responses[0]; await tick(); flushSync();
    // Acknowledge the engine, then observe the actual next host call; one Svelte
    // tick alone cannot certify completion of the joined async request chain.
    await vi.waitFor(() => expect(delayed.submitted).toEqual([{ kind: "advance", through_tick: 4, turns: [] }, { kind: "quit" }]));
    await delayed.responses[1]; await tick(); flushSync();
    expect(delayed.peak).toBe(1);
    expect(server.parsed().phase).toBe("terminal");
    expect(server.parsed().tick).toBe(4);
    expect(server.revision).toBe(3);
    expect(target.querySelector("[role=status]")?.textContent).toContain("Game ended.");
  } finally { await unmount(app); delayed.release(); await Promise.allSettled(delayed.responses); target.remove(); vi.useRealTimers(); }
});

it.skipIf(!browser)("snake: a terminal suffix drains after a delayed earlier acknowledgement (AR6.3)", async () => {
  const content = parseArcadeCatalog(JSON.parse(fixtureRaw), new Set(COPY_KEYS)).snake;
  const server = await new SnakeServer(await snakeIdentity(fixtureRaw)).init();
  const delayed = delayedSnakeHost(server), target = host();
  vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
  const app = mountSnake(target, server, content, { submit: delayed.submit, flushEvery: 1 });
  try {
    await tick(); flushSync();
    button(target, "Play").click();
    await deliverSnakeTimers(20);
    expect(delayed.submitted).toEqual([{ kind: "advance", through_tick: 1, turns: [] }]);
    await deliverSnakeTimers(40);
    expect(target.querySelector("[role=status]")?.textContent).toContain("Game over");
    expect(server.parsed().phase).toBe("playing");
    expect(server.parsed().tick).toBe(0);
    expect(delayed.submitted).toHaveLength(1);
    delayed.release();
    await delayed.responses[0]; await tick(); flushSync();
    await vi.waitFor(() => expect(delayed.submitted, "terminal suffix must reach the real engine without another player action").toEqual([
      { kind: "advance", through_tick: 1, turns: [] }, { kind: "advance", through_tick: 3, turns: [] },
    ]));
    await delayed.responses[1]; await tick(); flushSync();
    expect(delayed.peak).toBe(1);
    expect(server.parsed().phase).toBe("terminal");
    expect(server.parsed().tick).toBe(3);
    expect(server.revision).toBe(3);
  } finally { await unmount(app); delayed.release(); await Promise.allSettled(delayed.responses); target.remove(); vi.useRealTimers(); }
});

it.skipIf(!browser)("snake: repeated Quit drains a paused partial lead without overlapping commands", async () => {
  const server = await new SnakeServer(await snakeIdentity(candidateRaw)).init();
  const delayed = delayedSnakeHost(server), target = host();
  vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
  const app = mountSnake(target, server, catalog.snake, { submit: delayed.submit, flushEvery: 4 });
  try {
    await tick(); flushSync();
    button(target, "Play").click();
    await deliverSnakeTimers(120);
    expect(delayed.submitted).toEqual([{ kind: "advance", through_tick: 4, turns: [] }]);
    const localHead = snakeHead(target);
    button(target, "Quit").click(); button(target, "Quit").click();
    await tick(); flushSync();
    expect(delayed.submitted).toHaveLength(1);
    await deliverSnakeTimers(200);
    expect(snakeHead(target)).toBe(localHead);
    delayed.release();
    await vi.waitFor(() => expect(delayed.submitted).toEqual([
      { kind: "advance", through_tick: 4, turns: [] }, { kind: "advance", through_tick: 6, turns: [] }, { kind: "quit" },
    ]));
    await delayed.responses[2]; await tick(); flushSync();
    expect(delayed.peak).toBe(1);
    expect(server.parsed().phase).toBe("terminal");
    expect(server.parsed().tick).toBe(6);
    expect(server.revision).toBe(4);
  } finally { await unmount(app); delayed.release(); await Promise.allSettled(delayed.responses); target.remove(); vi.useRealTimers(); }
});

it.skipIf(!browser)("snake: failed advance and failed resync stop Quit without retrying blindly", async () => {
  const server = await new SnakeServer(await snakeIdentity(candidateRaw)).init();
  server.failNext = true;
  const delayed = delayedSnakeHost(server), target = host();
  vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
  const app = mountSnake(target, server, catalog.snake, { submit: delayed.submit, flushEvery: 4,
    current: async () => { server.currentCalls++; throw new Error("current unavailable"); } });
  try {
    await tick(); flushSync();
    button(target, "Play").click();
    await deliverSnakeTimers(80);
    button(target, "Quit").click();
    delayed.release();
    await expect(delayed.responses[0]).rejects.toThrow("rejected");
    await vi.waitFor(() => expect(target.textContent).toContain(t("arcade.error.rejected", {}, "era_1995")));
    expect(server.currentCalls).toBe(1);
    await deliverSnakeTimers(400);
    expect(delayed.submitted).toEqual([{ kind: "advance", through_tick: 4, turns: [] }]);
    expect(delayed.peak).toBe(1);
    expect(server.parsed().tick).toBe(0);
    expect(server.revision).toBe(1);
  } finally { await unmount(app); delayed.release(); await Promise.allSettled(delayed.responses); target.remove(); vi.useRealTimers(); }
});

for (const rejected of [false, true]) {
  it.skipIf(!browser)(`snake: ${rejected ? "rejected" : "accepted"} delayed response after unmount cannot drain or resync`, async () => {
    const content = parseArcadeCatalog(JSON.parse(fixtureRaw), new Set(COPY_KEYS)).snake;
    const server = await new SnakeServer(await snakeIdentity(fixtureRaw)).init();
    server.failNext = rejected;
    const delayed = delayedSnakeHost(server), target = host();
    vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
    const app = mountSnake(target, server, content, { submit: delayed.submit, flushEvery: 1 });
    let mounted = true;
    try {
      await tick(); flushSync();
      button(target, "Play").click();
      await deliverSnakeTimers(60);
      expect(delayed.submitted).toEqual([{ kind: "advance", through_tick: 1, turns: [] }]);
      expect(target.querySelector("[role=status]")?.textContent).toContain("Game over");
      await unmount(app);
      mounted = false;
      delayed.release();
      if (rejected) await expect(delayed.responses[0]).rejects.toThrow("rejected");
      else await delayed.responses[0];
      await tick(); flushSync(); await deliverSnakeTimers(200);
      expect(delayed.submitted).toEqual([{ kind: "advance", through_tick: 1, turns: [] }]);
      expect(server.currentCalls).toBe(0);
      expect(server.parsed().phase).toBe("playing");
      expect(server.parsed().tick).toBe(rejected ? 0 : 1);
    } finally { if (mounted) await unmount(app); delayed.release(); await Promise.allSettled(delayed.responses); target.remove(); vi.useRealTimers(); }
  });
}

for (const destination of ["tab-control", "pace-control", "outside-toy"] as const) {
  it.skipIf(!browser)(`snake: leaving the board for ${destination} pauses until explicit Resume (AR6.3)`, async () => {
    const { userEvent } = await import("vitest/browser");
    const server = await new SnakeServer(await snakeIdentity(candidateRaw)).init();
    const target = host();
    const outside = document.createElement("button");
    outside.textContent = "Outside test control";
    document.body.append(outside);
    vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
    const app = mountSnake(target, server, catalog.snake, { flushEvery: 1 });
    try {
      await tick(); flushSync();
      const board = target.querySelector<HTMLElement>("[role=application]")!;
      board.focus();
      button(target, "Play").click();
      const initialHead = snakeHead(target);
      await deliverSnakeTimers(20); await server.lastResponse; await tick(); flushSync();
      expect(server.parsed().tick).toBe(1);
      expect(snakeHead(target)).toBe(initialHead + 1);
      // Remaining on the same board must not pause: a real second step commits.
      board.focus();
      await deliverSnakeTimers(20); await server.lastResponse; await tick(); flushSync();
      expect(server.parsed().tick).toBe(2);
      expect(snakeHead(target)).toBe(initialHead + 2);
      if (destination === "tab-control") {
        await userEvent.keyboard("{Tab}");
        // Native WebKit/macOS may tab to the select rather than a button. Both
        // are outside the board and inside the toy; do not simulate that focus.
        expect(document.activeElement).not.toBe(board);
        expect(target.querySelector(".controls")?.contains(document.activeElement)).toBe(true);
      } else if (destination === "pace-control") {
        const pace = target.querySelector<HTMLSelectElement>("select")!;
        pace.focus();
        expect(document.activeElement).toBe(pace);
      } else {
        outside.focus();
        expect(document.activeElement).toBe(outside);
      }
      await tick(); flushSync();
      await deliverSnakeTimers(100);
      expect(snakeHead(target), "focus leaving the board must stop delivered steps even inside the toy").toBe(initialHead + 2);
      expect(server.parsed().tick).toBe(2);
      expect(button(target, "Resume")).toBeTruthy();
      expect(server.submitted).toEqual([
        { kind: "advance", through_tick: 1, turns: [] }, { kind: "advance", through_tick: 2, turns: [] },
      ]);
      board.focus();
      button(target, "Resume").click();
      await deliverSnakeTimers(20); await server.lastResponse; await tick(); flushSync();
      expect(snakeHead(target)).toBe(initialHead + 3);
      expect(server.parsed().tick).toBe(3);
      expect(server.revision).toBe(4);
    } finally { await unmount(app); outside.remove(); target.remove(); vi.useRealTimers(); }
  });
}

for (const key of ["p", "{Escape}"] as const) {
  it.skipIf(!browser)(`snake: native ${key} pauses and resumes delivered engine steps (AR6.3)`, async () => {
    const { userEvent } = await import("vitest/browser");
    const server = await new SnakeServer(await snakeIdentity(candidateRaw)).init();
    const target = host();
    vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
    const app = mountSnake(target, server, catalog.snake, { flushEvery: 1 });
    try {
      await tick(); flushSync();
      const board = target.querySelector<HTMLElement>("[role=application]")!;
      board.focus(); button(target, "Play").click();
      const initialHead = snakeHead(target);
      await deliverSnakeTimers(20); await server.lastResponse; await tick(); flushSync();
      expect(server.parsed().tick).toBe(1);
      await userEvent.keyboard(key); await tick(); flushSync();
      expect(button(target, "Resume")).toBeTruthy();
      await deliverSnakeTimers(100);
      expect(snakeHead(target)).toBe(initialHead + 1);
      expect(server.submitted).toEqual([{ kind: "advance", through_tick: 1, turns: [] }]);
      await userEvent.keyboard(key);
      await deliverSnakeTimers(20); await server.lastResponse; await tick(); flushSync();
      expect(snakeHead(target)).toBe(initialHead + 2);
      expect(server.parsed().tick).toBe(2);
      expect(server.revision).toBe(3);
    } finally { await unmount(app); target.remove(); vi.useRealTimers(); }
  });
}

it.skipIf(!browser)("snake: controlled hidden event pauses; visible alone does not resume (AR6.3)", async () => {
  const server = await new SnakeServer(await snakeIdentity(candidateRaw)).init();
  const target = host();
  const original = Object.getOwnPropertyDescriptor(document, "visibilityState");
  let visible = true;
  Object.defineProperty(document, "visibilityState", { configurable: true, get: () => visible ? "visible" : "hidden" });
  vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
  const app = mountSnake(target, server, catalog.snake, { flushEvery: 1 });
  try {
    await tick(); flushSync();
    button(target, "Play").click();
    const initialHead = snakeHead(target);
    document.dispatchEvent(new Event("visibilitychange"));
    await deliverSnakeTimers(20); await server.lastResponse; await tick(); flushSync();
    expect(server.parsed().tick).toBe(1);
    expect(snakeHead(target)).toBe(initialHead + 1);
    visible = false;
    document.dispatchEvent(new Event("visibilitychange"));
    await tick(); flushSync();
    expect(button(target, "Resume")).toBeTruthy();
    await deliverSnakeTimers(100);
    expect(snakeHead(target)).toBe(initialHead + 1);
    visible = true;
    document.dispatchEvent(new Event("visibilitychange"));
    await deliverSnakeTimers(100);
    expect(snakeHead(target)).toBe(initialHead + 1);
    expect(server.submitted).toEqual([{ kind: "advance", through_tick: 1, turns: [] }]);
    button(target, "Resume").click();
    await deliverSnakeTimers(20); await server.lastResponse; await tick(); flushSync();
    expect(snakeHead(target)).toBe(initialHead + 2);
    expect(server.parsed().tick).toBe(2);
    expect(server.revision).toBe(3);
  } finally {
    await unmount(app); target.remove(); vi.useRealTimers();
    if (original) Object.defineProperty(document, "visibilityState", original);
    else Reflect.deleteProperty(document, "visibilityState");
  }
});

it.skipIf(!browser)("snake: native keyboard quit reaches a real terminal (AR6.3, AC12)", async () => {
  const { userEvent } = await import("vitest/browser");
  const server = await new SnakeServer(await snakeIdentity(candidateRaw)).init();
  const target = host();
  const app = mountSnake(target, server, catalog.snake);
  try {
    await settle();
    expect(target.textContent).toContain("Paused");
    button(target, "Quit").focus();
    await userEvent.keyboard("{Enter}");
    expect(server.submitted).toEqual([{ kind: "quit" }]);
    await server.lastResponse; await settle();
    expect(server.parsed().phase).toBe("terminal");
    expect(server.parsed().tick).toBe(0);
    expect(server.revision).toBe(2);
    expect(target.querySelector("[role=status]")?.textContent).toContain("Game ended.");
    expect(target.querySelectorAll("button")).toHaveLength(0);
    await assertAxe(target, "native keyboard quit");
  } finally { await unmount(app); target.remove(); }
});

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
  vi.useFakeTimers({ toFake: ["Date", "setTimeout", "clearTimeout"] });
  const app = mountSnake(target, server, content, { flushEvery: 4, tickMS: 40 });
  try {
    await tick(); flushSync();
    button(target, "Play").click();
    const start = Date.now();
    const initialHead = snakeHead(target);
    // Elapsed wall time cannot certify callback delivery under CI load. Moving
    // the wall clock alone leaves this real child and its real engine at tick 0.
    vi.setSystemTime(start + 190);
    await tick(); flushSync();
    expect(Date.now() - start).toBe(190);
    expect(server.submitted).toEqual([]);
    expect(snakeHead(target)).toBe(initialHead);
    for (let count = 1; count <= 3; count += 1) {
      await deliverSnakeTimers(39);
      expect(snakeHead(target)).toBe(initialHead + count - 1);
      await deliverSnakeTimers(1);
      expect(snakeHead(target)).toBe(initialHead + count);
      expect(server.submitted).toEqual([]);
    }
    await deliverSnakeTimers(39);
    expect(server.submitted).toEqual([]);
    await deliverSnakeTimers(1);
    expect(server.submitted).toEqual([{ kind: "advance", through_tick: 4, turns: [] }]);
    await server.lastResponse;
    await tick(); flushSync();
    expect(server.parsed().tick).toBe(4);
    expect(snakeHead(target)).toBe(server.parsed().body[0]);
    window.dispatchEvent(new Event("blur"));
    await tick(); flushSync();
    expect(target.textContent).toContain("Paused");
    await deliverSnakeTimers(200);
    expect(server.submitted).toHaveLength(1);
    expect(snakeHead(target)).toBe(initialHead + 4);
    server.failNext = true;
    button(target, "Resume").click();
    await deliverSnakeTimers(160);
    await expect(server.lastResponse).rejects.toThrow("rejected");
    await tick(); flushSync();
    expect(server.submitted).toEqual([
      { kind: "advance", through_tick: 4, turns: [] },
      { kind: "advance", through_tick: 8, turns: [] },
    ]);
    expect(server.currentCalls).toBe(1);
    expect(server.parsed().tick).toBe(4);
    expect(snakeHead(target)).toBe(server.parsed().body[0]);
    expect(target.textContent).toContain("The game caught up with the server.");
  } finally { await unmount(app); target.remove(); vi.useRealTimers(); }
});

it.skipIf(!browser)("snake: freezes at max_ticks_per_advance while a flush is unacknowledged", async () => {
  const content = { ...catalog.snake, max_ticks_per_advance: 5 };
  const server = await new SnakeServer(await snakeIdentity(candidateRaw)).init();
  server.block = true;
  const target = host();
  vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
  const app = mountSnake(target, server, content, { flushEvery: 100 });
  try {
    await tick(); flushSync();
    button(target, "Play").click();
    const initialHead = snakeHead(target);
    await deliverSnakeTimers(80);
    expect(snakeHead(target)).toBe(initialHead + 4);
    expect(server.submitted).toEqual([]);
    await deliverSnakeTimers(20);
    const first = snakeHead(target);
    expect(first).toBe(initialHead + 5);
    await deliverSnakeTimers(200);
    expect(snakeHead(target)).toBe(first);
    expect(server.parsed().tick).toBe(0);
    // Only the one frozen flush is outstanding: at most one command in flight.
    expect(server.submitted).toEqual([{ kind: "advance", through_tick: 5, turns: [] }]);
  } finally { await unmount(app); target.remove(); vi.useRealTimers(); }
});
