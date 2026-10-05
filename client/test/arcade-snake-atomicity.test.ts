import { describe, expect, it } from "vitest";
import content from "../../testdata/arcade/snake-5x5-fixture-v1.json?raw";
import corpus from "../../testdata/arcade/snake-5x5-gate-v1.json";
import shared from "../../testdata/arcade/snake-rejection-atomicity-v1.json";
import { ArcadeRejection } from "../src/arcade/common";
import { applySnake, createSnake, decodeSnakeSnapshot, snakeTransition, SNAKE_SCALING_DESTINATION, type SnakeCommand } from "../src/arcade/snake";

describe("Snake rejected advances are atomic (AR4.5, AC5/AC6)", () => {
  it("pins the shared six-case population", () => {
    expect(shared.version).toBe(1);
    expect(shared.cases).toHaveLength(6);
    expect(new Set(shared.cases.map((row) => row.name)).size).toBe(6);
  });

  for (const moved of [false, true]) {
    it.each(shared.cases)(`${moved ? "moved" : "genesis"}: refuses $name without changing real state`, async (row) => {
      const identity = { content, content_hash: corpus.arcade_content_hash, content_schema_version: 1,
        seed: 455n, mode: "solo" as const, scaling_inputs: { [SNAKE_SCALING_DESTINATION]: 1 } };
      let snapshot = await createSnake(identity), revision = 1;
      if (moved) { snapshot = (await applySnake({ ...identity, snapshot, revision, command: '{"kind":"advance","through_tick":1,"turns":[]}' })).snapshot; revision++; }
      const state = decodeSnakeSnapshot(snapshot);
      const times: Record<string, number> = { current_tick: state.tick, next_tick: state.tick + 1,
        second_tick: state.tick + 2, after_death: state.tick + state.width - state.body[0]! % state.width + 1 };
      const resolve = (key: string): number => {
        if (!Object.hasOwn(times, key)) throw new Error(`unknown time marker ${key}`);
        return times[key]!;
      };
      const command = { kind: "advance", through_tick: resolve(row.through),
        turns: row.turns.map((turn) => ({ tick: resolve(turn.tick), direction: turn.direction })) } as SnakeCommand;
      await expect(applySnake({ ...identity, snapshot, revision, command: JSON.stringify(command) })).rejects.toMatchObject({ code: row.expected });
      const before = JSON.stringify(state);
      try {
        snakeTransition(state, command, JSON.parse(content).snake);
        throw new Error("direct transition admitted rejected advance");
      } catch (error) {
        expect(error).toBeInstanceOf(ArcadeRejection);
        expect((error as ArcadeRejection).code).toBe(row.expected);
      }
      expect(JSON.stringify(state), "direct transition preserves its actual caller-owned state").toBe(before);
      const output = await applySnake({ ...identity, snapshot, revision,
        command: JSON.stringify({ kind: "advance", through_tick: state.tick + 1, turns: [] }) });
      const next = decodeSnakeSnapshot(output.snapshot);
      expect(next.tick).toBe(state.tick + 1);
      expect(next.body[0]).toBe(state.body[0]! + 1);
      expect(next.revision).toBe(revision + 1);
      expect(output.result).toBeNull();
      const quit = await applySnake({ ...identity, snapshot: output.snapshot, revision: revision + 1, command: '{"kind":"quit"}' });
      expect(quit.result?.outcome).toBe("quit");
      await expect(applySnake({ ...identity, snapshot: quit.snapshot, revision: revision + 2,
        command: '{"kind":"quit"}' })).rejects.toMatchObject({ code: "illegal_phase" });
    });
  }
});
