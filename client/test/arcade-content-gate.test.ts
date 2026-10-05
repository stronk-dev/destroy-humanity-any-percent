import { describe, expect, it } from "vitest";
import candidate from "../../balance/testdata/arcade-v1.json?raw";
import fixture from "../../testdata/arcade/corpus-fixture-v1.json?raw";
import corpusSource from "../../testdata/arcade/content-gate-v2.json";
import { COPY_KEYS } from "../src/copy";
import { activeArcadeStage, arcadeContentHash, parseArcadeCatalog } from "../src/arcade/catalog";
import { ArcadeRejection, type ArcadeResult } from "../src/arcade/common";
import { applyMineGrid, createMineGrid, decodeMineGridSnapshot, MINE_GRID_SCALING_DESTINATION, mineGridMineCells, mineGridNeighbors, type MineGridSnapshot } from "../src/arcade/mine-grid";
import { applySnake, createSnake, SNAKE_SCALING_DESTINATION } from "../src/arcade/snake";

interface Corpus {
  readonly version: number;
  readonly arcade_content_hash: string;
  readonly transition_budget: number;
  readonly scenarios: readonly { readonly name: string; readonly engine: "mine_grid" | "snake"; readonly seed: string;
    readonly expected_genesis: unknown;
    readonly genesis_bytes: string;
    readonly steps: readonly { readonly command: unknown; readonly expect: string;
      readonly expected_snapshot: unknown; readonly expected_result: ArcadeResult | null;
      readonly snapshot_bytes: string; readonly result_bytes: string }[];
    readonly expected_terminal: unknown; readonly expected_result: ArcadeResult }[];
}

const corpus = corpusSource as unknown as Corpus;
const declared = new Set<string>(COPY_KEYS);

describe("arcade shared content gate (AR7)", () => {
  it("budgets every attempted command, including rejected attempts", () => {
    const attempts = corpus.scenarios.reduce((count, scenario) => count + scenario.steps.length, 0);
    expect(corpus.transition_budget).toBe(attempts);
  });

  it("byte-replays every Go-generated scenario, rejections included", async () => {
    expect(corpus.version).toBe(2);
    expect(await arcadeContentHash(fixture)).toBe(corpus.arcade_content_hash);
    let transitions = 0;
    let observations = 0;
    for (const scenario of corpus.scenarios) {
      const snake = scenario.engine === "snake";
      const identity = { content: fixture, content_hash: corpus.arcade_content_hash, content_schema_version: 1, seed: BigInt(scenario.seed), mode: "solo" as const,
        scaling_inputs: { [snake ? SNAKE_SCALING_DESTINATION : MINE_GRID_SCALING_DESTINATION]: 1 } };
      let snapshot = snake ? await createSnake(identity) : await createMineGrid(identity);
      let result: ArcadeResult | null = null;
      let revision = 1;
      expect(snapshot, `${scenario.name}: literal Go genesis bytes`).toBe(scenario.genesis_bytes);
      observations++;
      for (const step of scenario.steps) {
        transitions++;
        const before = snapshot;
        const beforeResult = JSON.stringify(result);
        const beforeRevision = revision;
        const label = `${scenario.name} attempt ${transitions}: ${JSON.stringify(step.command)}`;
        try {
          const input = { ...identity, revision, snapshot, command: JSON.stringify(step.command) };
          const output = snake ? await applySnake(input) : await applyMineGrid(input);
          expect("applied", label).toBe(step.expect);
          snapshot = output.snapshot; result = output.result; revision++;
        } catch (error) {
          if (!(error instanceof ArcadeRejection)) throw error;
          expect(error.code, label).toBe(step.expect);
          expect(snapshot, "a rejection mutates nothing").toBe(before);
          expect(JSON.stringify(result), "a rejection preserves the previous result").toBe(beforeResult);
          expect(revision, "a rejection does not advance revision").toBe(beforeRevision);
        }
        expect(snapshot, `${label}: literal Go snapshot bytes`).toBe(step.snapshot_bytes);
        expect(JSON.stringify(result), `${label}: literal Go result bytes`).toBe(step.result_bytes);
        expect(revision, `${label}: revision`).toBe((step.expected_snapshot as { revision: number }).revision);
        observations++;
      }
      expect(snapshot, scenario.name).toBe(JSON.stringify(scenario.expected_terminal));
      expect(JSON.stringify(result), scenario.name).toBe(JSON.stringify(scenario.expected_result));
    }
    expect(transitions).toBe(corpus.transition_budget);
    expect(observations).toBe(corpus.scenarios.length + corpus.transition_budget);
  });

  it("covers every outcome and rejection code of both engines", () => {
    const outcomes = new Set(corpus.scenarios.map((row) => `${row.engine}:${row.expected_result.outcome}`));
    expect([...outcomes].sort()).toEqual(["mine_grid:cleared", "mine_grid:detonated", "mine_grid:quit", "snake:cleared", "snake:crashed", "snake:quit"]);
    const codes = new Set(corpus.scenarios.flatMap((row) => row.steps.map((step) => `${row.engine}:${step.expect}`)));
    for (const code of ["cell_flagged", "cell_out_of_range", "cell_revealed", "chord_unsatisfied", "illegal_phase", "unknown_preset"]) expect(codes.has(`mine_grid:${code}`), code).toBe(true);
    for (const code of ["advance_past_terminal", "advance_window", "illegal_phase", "invalid_turn", "turns_not_ascending"]) expect(codes.has(`snake:${code}`), code).toBe(true);
  });

  it("never exposes mines in actual nonterminal outputs and refuses forged hidden state (AC4)", async () => {
    const scenarios = corpus.scenarios.filter((row) => row.engine === "mine_grid");
    const populations = new Set<string>();
    let observations = 0;
    let refusalPopulations = 0;
    for (const scenario of scenarios) {
      const identity = { content: fixture, content_hash: corpus.arcade_content_hash, content_schema_version: 1, seed: BigInt(scenario.seed), mode: "solo" as const,
        scaling_inputs: { [MINE_GRID_SCALING_DESTINATION]: 1 } };
      let snapshot = await createMineGrid(identity);
      let revision = 1;
      const inspect = async (): Promise<void> => {
        // Do not ask the decoder under test to certify its own output. Observe
        // the actual engine bytes before testing decoder/apply refusal separately.
        const raw = JSON.parse(snapshot) as MineGridSnapshot;
        const placed = raw.first_cell >= 0;
        populations.add(`${raw.phase}:${placed ? "placed" : "unplaced"}`);
        observations++;
        if (raw.phase !== "terminal") {
          expect(raw.mine_cells, `${scenario.name} revision ${revision}: hidden mines`).toEqual([]);
          expect(raw.exploded_cell, `${scenario.name} revision ${revision}: hidden explosion`).toBe(-1);
        } else {
          const mines = placed ? mineGridMineCells(raw.width, raw.height, raw.mines, raw.first_cell, identity.seed) : [];
          expect(raw.mine_cells, `${scenario.name}: terminal mine disclosure`).toEqual(mines);
        }
        expect(decodeMineGridSnapshot(snapshot), `${scenario.name}: clean decoder positive control`).toEqual(raw);
        if (raw.phase !== "terminal" && placed) {
          const mines = mineGridMineCells(raw.width, raw.height, raw.mines, raw.first_cell, identity.seed);
          expect(mines).toHaveLength(raw.mines);
          for (const [label, forged] of [
            ["mine list", { ...raw, mine_cells: mines }],
            ["exploded cell", { ...raw, exploded_cell: mines[0]! }],
          ] as const) {
            const leaked = JSON.stringify(forged);
            expect(() => decodeMineGridSnapshot(leaked), `${scenario.name}: ${label} decoder refusal`).toThrow(SyntaxError);
            await expect(applyMineGrid({ ...identity, revision, snapshot: leaked, command: '{"kind":"quit"}' }),
              `${scenario.name}: ${label} engine refusal`).rejects.toThrow(SyntaxError);
            refusalPopulations++;
          }
        }
      };
      await inspect();
      for (const step of scenario.steps) {
        try {
          const output = await applyMineGrid({ ...identity, revision, snapshot, command: JSON.stringify(step.command) });
          expect(step.expect, scenario.name).toBe("applied");
          snapshot = output.snapshot;
          revision++;
        } catch (error) {
          if (!(error instanceof ArcadeRejection)) throw error;
          expect(error.code, scenario.name).toBe(step.expect);
        }
        await inspect();
      }
      expect(snapshot, `${scenario.name}: unchanged corpus terminal`).toBe(JSON.stringify(scenario.expected_terminal));
    }
    expect(observations).toBe(scenarios.reduce((count, scenario) => count + 1 + scenario.steps.length, 0));
    expect([...populations].sort()).toEqual(["playing:placed", "playing:unplaced", "setup:unplaced", "terminal:placed", "terminal:unplaced"]);
    expect(refusalPopulations).toBeGreaterThan(0);
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
      "wrong schema": (v) => { v.schema_version = 2; },
      "unknown container key": (v) => { v.container.extra = 1; },
      "unknown stage key": (v) => { v.container.stages[0].extra = 1; },
      "empty stages": (v) => { v.container.stages = []; },
      "unsorted toys": (v) => { v.container.stages[0].toys = ["arcade.snake", "arcade.mine_grid"]; },
      "unknown title key": (v) => { v.container.stages[0].title_copy_key = "arcade.stage.missing.title"; },
      "stage tier above max": (v) => { v.container.stages[0].min_tier = 10; },
      "equal stage tiers": (v) => { v.container.stages.push({ ...v.container.stages[0], stage_id: "fixture_next" }); },
      "descending stage tiers": (v) => { v.container.stages.push({ ...v.container.stages[0], stage_id: "fixture_next" }); v.container.stages[0].min_tier = 1; },
      "unknown preset key": (v) => { v.mine_grid.presets[0].extra = 1; },
      "unsorted presets": (v) => { [v.mine_grid.presets[0], v.mine_grid.presets[1]] = [v.mine_grid.presets[1], v.mine_grid.presets[0]]; },
      "narrow board": (v) => { v.mine_grid.presets[2].width = 4; },
      "tall board": (v) => { v.mine_grid.presets[2].height = 31; },
      "too many mines": (v) => { v.mine_grid.presets[2].mines = 9 * 9 - 8; },
      "no mines": (v) => { v.mine_grid.presets[2].mines = 0; },
      "preset copy drift": (v) => { v.mine_grid.presets[2].copy_key = "arcade.mine_grid.preset.large"; },
      "unknown snake key": (v) => { v.snake.extra = 1; },
      "snake start too long": (v) => { v.snake.start_length = 11; },
      "snake growth zero": (v) => { v.snake.growth_per_food = 0; },
      "snake advance too large": (v) => { v.snake.max_ticks_per_advance = 257; },
      "snake tick too fast": (v) => { v.snake.presentation_tick_ms = 49; },
    };
    expect(Object.keys(cases)).toHaveLength(22);
    for (const [name, change] of Object.entries(cases)) expect(() => parseArcadeCatalog(mutate(change), declared), name).toThrow(SyntaxError);
  });

  it("loads inclusive bounds and structurally valid ascending test-only stages", () => {
    const value = JSON.parse(candidate);
    value.mine_grid.presets[2].mines = 9 * 9 - 9;
    value.snake.start_length = 10;
    value.snake.presentation_tick_ms = 50;
    value.container.stages.push({ ...value.container.stages[0], stage_id: "fixture_next", min_tier: 9 });
    const catalog = parseArcadeCatalog(value, declared);
    expect(catalog.container.stages.map((stage) => stage.min_tier)).toEqual([0, 9]);
    expect(catalog.snake.start_length).toBe(10);
    expect(catalog.mine_grid.presets[2]?.mines).toBe(72);
    expect(activeArcadeStage(catalog, 8)?.stage_id).toBe("cover_disc");
    expect(activeArcadeStage(catalog, 9)?.stage_id).toBe("fixture_next");
  });

  it("matches Go's mine placement shape: first-reveal exclusion and sorted output", () => {
    const mines = mineGridMineCells(9, 9, 10, 40, 7n);
    expect(mines).toHaveLength(10);
    const excluded = new Set([...mineGridNeighbors(9, 9, 40), 40]);
    expect(mines.every((cell, index) => !excluded.has(cell) && (index === 0 || mines[index - 1]! < cell))).toBe(true);
  });
});
