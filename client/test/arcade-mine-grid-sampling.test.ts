import { describe, expect, it } from "vitest";
import fixture from "../../testdata/arcade/corpus-fixture-v1.json?raw";
import witness from "../../testdata/arcade/mine-grid-sampling-v1.json";
import { substream } from "../src/combat/rng";
import { arcadeContentHash } from "../src/arcade/catalog";
import { ArcadeRejection, type ArcadeResult } from "../src/arcade/common";
import { applyMineGrid, createMineGrid, decodeMineGridSnapshot, mineGridMineCells, mineGridNeighbors, MINE_GRID_SCALING_DESTINATION } from "../src/arcade/mine-grid";

describe("Mine Grid consumes a real rejected shuffle draw (AR3.3, AC3)", () => {
  const seed = BigInt(witness.seed);

  it("executes the real two-substream seed and rejects zero at the actual board bound", () => {
    expect(witness.version).toBe(1);
    expect(witness.seed).toBe("15581846558861750132");
    const run = substream(seed, "mine_grid.run.v1").next();
    expect(run.toString()).toBe(witness.run_seed);
    const raw = substream(run, "mine_grid.mines.v1");
    const first = raw.next(), accepted = raw.next();
    const bound = BigInt(witness.bound), threshold = (1n << 64n) % bound;
    expect(witness.bound).toBe(9 * 9 - mineGridNeighbors(9, 9, 40).length - 1);
    expect(threshold).toBe(16n);
    expect(witness.threshold).toBe(Number(threshold));
    expect(first.toString()).toBe(witness.first_draw);
    expect(first).toBe(0n);
    expect(first < threshold).toBe(true);
    expect(accepted.toString()).toBe(witness.accepted_draw);
    expect(accepted >= threshold).toBe(true);
    const bounded = substream(run, "mine_grid.mines.v1");
    expect(bounded.bound(bound)).toBe(accepted % bound);
    expect(bounded.next(), "rejection must consume both initial draws").toBe(raw.next());
  });

  it("matches the literal real Go placement through Mine Grid's own shuffle call", () => {
    const mines = mineGridMineCells(9, 9, 10, 40, seed);
    expect(mines).toEqual(witness.mines);
    expect(mines).toHaveLength(10);
    const excluded = new Set([40, ...mineGridNeighbors(9, 9, 40)]);
    expect(mines.every((cell, index) => !excluded.has(cell) && (index === 0 || mines[index - 1]! < cell))).toBe(true);
  });

  it("byte-replays actual create/choose/reveal/quit and refused terminal retry", async () => {
    expect(await arcadeContentHash(fixture)).toBe(witness.arcade_content_hash);
    const scenario = witness.scenario;
    expect(scenario.engine).toBe("mine_grid");
    expect(scenario.seed).toBe(witness.seed);
    expect(scenario.steps).toHaveLength(4);
    expect(witness.transition_budget).toBe(scenario.steps.length);
    const identity = { content: fixture, content_hash: witness.arcade_content_hash, content_schema_version: 1,
      seed, mode: "solo" as const, scaling_inputs: { [MINE_GRID_SCALING_DESTINATION]: 1 } };
    let snapshot = await createMineGrid(identity), revision = 1, observations = 1;
    let result: ArcadeResult | null = null;
    expect(snapshot, "literal Go genesis").toBe(scenario.genesis_bytes);
    for (const [index, step] of scenario.steps.entries()) {
      const before = snapshot, beforeRevision = revision, beforeResult = JSON.stringify(result);
      try {
        const actual = await applyMineGrid({ ...identity, snapshot, revision, command: JSON.stringify(step.command) });
        expect(step.expect).toBe("applied");
        snapshot = actual.snapshot; result = actual.result; revision++;
      } catch (error) {
        if (!(error instanceof ArcadeRejection)) throw error;
        expect(error.code).toBe(step.expect);
        expect(snapshot).toBe(before);
        expect(revision).toBe(beforeRevision);
        expect(JSON.stringify(result)).toBe(beforeResult);
      }
      expect(snapshot, `attempt ${index + 1}: literal Go snapshot`).toBe(step.snapshot_bytes);
      expect(JSON.stringify(result), `attempt ${index + 1}: literal Go result`).toBe(step.result_bytes);
      expect(decodeMineGridSnapshot(snapshot).revision).toBe(revision);
      observations++;
    }
    expect(observations).toBe(5);
    expect(decodeMineGridSnapshot(snapshot).mine_cells).toEqual(witness.mines);
    expect(result?.outcome).toBe("quit");
    expect(snapshot).toBe(JSON.stringify(scenario.expected_terminal));
    expect(JSON.stringify(result)).toBe(JSON.stringify(scenario.expected_result));
  });
});
