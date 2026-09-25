import { describe, expect, it } from "vitest";
import candidate from "../../balance/testdata/arcade-v1.json?raw";
import fixture from "../../testdata/arcade/corpus-fixture-v1.json?raw";
import corpusSource from "../../testdata/arcade/content-gate-v1.json";
import { COPY_KEYS } from "../src/copy";
import { activeArcadeStage, arcadeContentHash, parseArcadeCatalog } from "../src/arcade/catalog";
import { ArcadeRejection, type ArcadeResult } from "../src/arcade/common";
import { applyMineGrid, createMineGrid, MINE_GRID_SCALING_DESTINATION, mineGridMineCells, mineGridNeighbors } from "../src/arcade/mine-grid";
import { applySnake, createSnake, SNAKE_SCALING_DESTINATION } from "../src/arcade/snake";

interface Corpus {
  readonly version: number;
  readonly arcade_content_hash: string;
  readonly transition_budget: number;
  readonly scenarios: readonly { readonly name: string; readonly engine: "mine_grid" | "snake"; readonly seed: string;
    readonly steps: readonly { readonly command: unknown; readonly expect: string }[];
    readonly expected_terminal: unknown; readonly expected_result: ArcadeResult }[];
}

const corpus = corpusSource as unknown as Corpus;
const declared = new Set<string>(COPY_KEYS);

describe("arcade shared content gate (AR7)", () => {
  it("byte-replays every Go-generated scenario, rejections included", async () => {
    expect(await arcadeContentHash(fixture)).toBe(corpus.arcade_content_hash);
    let transitions = 0;
    for (const scenario of corpus.scenarios) {
      const snake = scenario.engine === "snake";
      const identity = { content: fixture, content_hash: corpus.arcade_content_hash, content_schema_version: 1, seed: BigInt(scenario.seed), mode: "solo" as const,
        scaling_inputs: { [snake ? SNAKE_SCALING_DESTINATION : MINE_GRID_SCALING_DESTINATION]: 1 } };
      let snapshot = snake ? await createSnake(identity) : await createMineGrid(identity);
      let result: ArcadeResult | null = null;
      let revision = 1;
      for (const step of scenario.steps) {
        const before = snapshot;
        try {
          const input = { ...identity, revision, snapshot, command: JSON.stringify(step.command) };
          const output = snake ? await applySnake(input) : await applyMineGrid(input);
          expect("applied", `${scenario.name} ${JSON.stringify(step.command)}`).toBe(step.expect);
          snapshot = output.snapshot; result = output.result; revision++; transitions++;
        } catch (error) {
          if (!(error instanceof ArcadeRejection)) throw error;
          expect(error.code, `${scenario.name} ${JSON.stringify(step.command)}`).toBe(step.expect);
          expect(snapshot, "a rejection mutates nothing").toBe(before);
        }
      }
      expect(snapshot, scenario.name).toBe(JSON.stringify(scenario.expected_terminal));
      expect(JSON.stringify(result), scenario.name).toBe(JSON.stringify(scenario.expected_result));
    }
    expect(transitions).toBe(corpus.transition_budget);
  });

  it("covers every outcome and rejection code of both engines", () => {
    const outcomes = new Set(corpus.scenarios.map((row) => `${row.engine}:${row.expected_result.outcome}`));
    expect([...outcomes].sort()).toEqual(["mine_grid:cleared", "mine_grid:detonated", "mine_grid:quit", "snake:cleared", "snake:crashed", "snake:quit"]);
    const codes = new Set(corpus.scenarios.flatMap((row) => row.steps.map((step) => `${row.engine}:${step.expect}`)));
    for (const code of ["cell_flagged", "cell_out_of_range", "cell_revealed", "chord_unsatisfied", "illegal_phase", "unknown_preset"]) expect(codes.has(`mine_grid:${code}`), code).toBe(true);
    for (const code of ["advance_past_terminal", "advance_window", "illegal_phase", "invalid_turn", "turns_not_ascending"]) expect(codes.has(`snake:${code}`), code).toBe(true);
  });

  it("never exposes a mine position in a non-terminal snapshot (AC4)", () => {
    for (const scenario of corpus.scenarios.filter((row) => row.engine === "mine_grid")) {
      const terminal = scenario.expected_terminal as { phase: string; mine_cells: number[]; mines: number; first_cell: number };
      expect(terminal.phase).toBe("terminal");
      if (terminal.first_cell >= 0) expect(terminal.mine_cells.length).toBe(terminal.mines);
    }
  });
});

describe("arcade catalog (AR1.2/AR3.1/AR4.1)", () => {
  it("loads the candidate and the corpus fixture and selects the active stage by tier", () => {
    const catalog = parseArcadeCatalog(JSON.parse(candidate), declared);
    expect(catalog.container.stages.map((row) => row.stage_id)).toEqual(["cover_disc"]);
    expect(activeArcadeStage(catalog, 0)?.toys).toEqual(["arcade.mine_grid", "arcade.snake"]);
    expect(activeArcadeStage(catalog, 1)?.stage_id).toBe("cover_disc");
    expect(() => parseArcadeCatalog(JSON.parse(fixture), declared)).not.toThrow();
  });

  it("rejects the same loader defects as Go", () => {
    const mutate = (change: (value: Record<string, any>) => void) => { const value = JSON.parse(candidate); change(value); return value; };
    const cases: Record<string, (value: Record<string, any>) => void> = {
      "unknown root key": (v) => { v.extra = 1; },
      "unsorted toys": (v) => { v.container.stages[0].toys = ["arcade.snake", "arcade.mine_grid"]; },
      "too many mines": (v) => { v.mine_grid.presets[2].mines = 9 * 9 - 8; },
      "preset copy drift": (v) => { v.mine_grid.presets[2].copy_key = "arcade.mine_grid.preset.large"; },
      "snake start too long": (v) => { v.snake.start_length = 11; },
      "snake tick too fast": (v) => { v.snake.presentation_tick_ms = 49; },
    };
    for (const [name, change] of Object.entries(cases)) expect(() => parseArcadeCatalog(mutate(change), declared), name).toThrow(SyntaxError);
  });

  it("matches Go's mine placement shape: first-reveal exclusion and sorted output", () => {
    const mines = mineGridMineCells(9, 9, 10, 40, 7n);
    expect(mines).toHaveLength(10);
    const excluded = new Set([...mineGridNeighbors(9, 9, 40), 40]);
    expect(mines.every((cell, index) => !excluded.has(cell) && (index === 0 || mines[index - 1]! < cell))).toBe(true);
  });
});
