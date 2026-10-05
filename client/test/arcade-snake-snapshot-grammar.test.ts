import { describe, expect, it } from "vitest";
import content from "../../testdata/arcade/corpus-fixture-v1.json?raw";
import corpus from "../../testdata/arcade/content-gate-v2.json";
import grammar from "../../testdata/arcade/snake-snapshot-negatives-v1.json";
import { ArcadeRejection } from "../src/arcade/common";
import { applySnake, createSnake, decodeSnakeSnapshot, SNAKE_SCALING_DESTINATION } from "../src/arcade/snake";

async function population(stage: string) {
  const source = grammar.populations[stage as keyof typeof grammar.populations];
  expect(source, `known population ${stage}`).toBeDefined();
  const scenario = corpus.scenarios.find((row) => row.name === source.scenario);
  if (!scenario || scenario.engine !== "snake") throw new Error("missing actual Snake scenario");
  expect(source.through_step).toBeLessThanOrEqual(scenario.steps.length);
  const identity = { content, content_hash: corpus.arcade_content_hash, content_schema_version: 1,
    seed: BigInt(scenario.seed), mode: "solo" as const, scaling_inputs: { [SNAKE_SCALING_DESTINATION]: 1 } };
  let snapshot = await createSnake(identity), revision = 1;
  expect(snapshot).toBe(scenario.genesis_bytes);
  for (const step of scenario.steps.slice(0, source.through_step)) {
    try {
      const output = await applySnake({ ...identity, snapshot, revision, command: JSON.stringify(step.command) });
      expect(step.expect).toBe("applied");
      snapshot = output.snapshot;
      revision++;
    } catch (error) {
      if (!(error instanceof ArcadeRejection)) throw error;
      expect(error.code).toBe(step.expect);
    }
    expect(snapshot).toBe(step.snapshot_bytes);
  }
  expect(() => decodeSnakeSnapshot(snapshot)).not.toThrow();
  return { ...identity, snapshot, revision, command: '{"kind":"quit"}' };
}

function replaceExactlyOnce(snapshot: string, before: string, after: string): string {
  if (before === "" || snapshot.split(before).length !== 2) throw new Error("raw mutation must match exactly once");
  return snapshot.replace(before, after);
}

function mutate(snapshot: string, edits: readonly { before: string; after: string }[]): string {
  return edits.reduce((value, edit) => replaceExactlyOnce(value, edit.before, edit.after), snapshot);
}

describe("Snake shared snapshot grammar (AR4.3/AR7)", () => {
  it("pins the predeclared shared population and refuses broken mutation instruments", () => {
    expect(grammar.version).toBe(1);
    expect(grammar.cases).toHaveLength(45);
    expect(grammar.positive_cases).toHaveLength(7);
    expect(() => replaceExactlyOnce("{}", "missing", "x")).toThrow("match exactly once");
    expect(() => replaceExactlyOnce("xx", "x", "y")).toThrow("match exactly once");
    expect(() => replaceExactlyOnce("{}", "", "x")).toThrow("match exactly once");
  });

  it.each(grammar.cases)("decoder refuses $name", async (row) => {
    const input = await population(row.stage);
    expect(() => decodeSnakeSnapshot(mutate(input.snapshot, row.edits))).toThrow(SyntaxError);
    expect(() => decodeSnakeSnapshot(input.snapshot)).not.toThrow();
  });

  it.each(grammar.cases)("direct apply refuses $name", async (row) => {
    const input = await population(row.stage);
    await expect(applySnake({ ...input, snapshot: mutate(input.snapshot, row.edits) })).rejects.toThrow(SyntaxError);
    if (row.stage === "terminal") await expect(applySnake(input)).rejects.toMatchObject({ code: "illegal_phase" });
    else expect(await applySnake(input)).toEqual(await applySnake(input));
  });

  it.each(grammar.positive_cases)("accepts legal control $name", async (row) => {
    const input = await population(row.stage);
    const legal = mutate(input.snapshot, row.edits);
    expect(() => decodeSnakeSnapshot(legal)).not.toThrow();
    const output = await applySnake({ ...input, snapshot: legal });
    if (row.name === "tick_safe_max") {
      expect(decodeSnakeSnapshot(output.snapshot).tick).toBe(Number.MAX_SAFE_INTEGER);
      expect(output.result?.score_facts.find((fact) => fact.kind === "snake.ticks_survived")?.value).toBe(Number.MAX_SAFE_INTEGER);
    } else {
      const baseline = await applySnake(input);
      expect(output.snapshot).toBe(baseline.snapshot);
      expect(JSON.stringify(output.result)).toBe(JSON.stringify(baseline.result));
    }
  });
});
