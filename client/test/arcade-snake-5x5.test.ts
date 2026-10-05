import { describe, expect, it } from "vitest";
import fixture from "../../testdata/arcade/snake-5x5-fixture-v1.json?raw";
import historicalFixture from "../../testdata/arcade/corpus-fixture-v1.json?raw";
import corpus from "../../testdata/arcade/snake-5x5-gate-v1.json";
import historical from "../../testdata/arcade/content-gate-v2.json";
import { arcadeContentHash } from "../src/arcade/catalog";
import { ArcadeRejection, type ArcadeResult } from "../src/arcade/common";
import { applySnake, createSnake, SNAKE_SCALING_DESTINATION, type SnakeSnapshot } from "../src/arcade/snake";

function validClearing(genesis: Pick<SnakeSnapshot, "width" | "height" | "body">,
  terminal: Pick<SnakeSnapshot, "width" | "height" | "body" | "phase" | "food_cell" | "score" | "tick">,
  result: ArcadeResult | null): boolean {
  return genesis.width === 5 && genesis.height === 5 && JSON.stringify(genesis.body) === "[12,11]" &&
    terminal.width === 5 && terminal.height === 5 && terminal.phase === "terminal" && terminal.food_cell === -1 &&
    terminal.body.length === 25 && new Set(terminal.body).size === 25 &&
    terminal.body.every((cell) => Number.isInteger(cell) && cell >= 0 && cell < 25) && terminal.score === 24 &&
    result?.outcome === "cleared" && result.rating_delta === null && result.score_facts.length === 2 &&
    result.score_facts[0]?.kind === "snake.food_eaten" && result.score_facts[0]?.value === terminal.score &&
    result.score_facts[1]?.kind === "snake.ticks_survived" && result.score_facts[1]?.value === terminal.tick;
}

describe("exact 5×5 Snake clearing witness (AR7, RP-196)", () => {
  it("changes only the test fixture width and accounts for the entire Go research population", async () => {
    const old = JSON.parse(historicalFixture) as { snake: { width: number } };
    old.snake.width = 5;
    expect(JSON.parse(fixture)).toEqual(old);
    expect(await arcadeContentHash(fixture)).toBe(corpus.arcade_content_hash);
    expect(corpus.version).toBe(1);
    expect(corpus.seed_population).toBe(1024);
    expect(corpus.observations.map((row) => row.seed)).toEqual(Array.from({ length: 1024 }, (_, i) => i + 1));
    const counts: Record<string, number> = { cleared: 0, crashed: 0, excluded_food: 0 };
    for (const row of corpus.observations) {
      expect(Object.hasOwn(counts, row.outcome), `classified seed ${row.seed}`).toBe(true);
      counts[row.outcome] = counts[row.outcome]! + 1;
    }
    expect(counts).toEqual(corpus.outcomes);
    expect(corpus.largest_observed_tick).toBe(Math.max(...corpus.observations.map((row) => row.tick)));
    expect(corpus.scenario.seed).toBe(String(corpus.observations.find((row) => row.outcome === "cleared")?.seed));
    expect(corpus.transition_budget).toBe(corpus.scenario.steps.length);
  });

  it("byte-replays every actual attempted output, including terminal overshoot and phase refusal", async () => {
    const identity = { content: fixture, content_hash: corpus.arcade_content_hash, content_schema_version: 1,
      seed: BigInt(corpus.scenario.seed), mode: "solo" as const, scaling_inputs: { [SNAKE_SCALING_DESTINATION]: 1 } };
    let snapshot = await createSnake(identity), revision = 1;
    let result: ArcadeResult | null = null;
    expect(snapshot, "literal Go genesis").toBe(corpus.scenario.genesis_bytes);
    const genesis = JSON.parse(snapshot) as SnakeSnapshot;
    let attempts = 0, sawEating = false, sawGrowth = false, observedStates = 1;
    const refusals = new Set<string>();
    for (const step of corpus.scenario.steps) {
      const before = snapshot, beforeResult = JSON.stringify(result), beforeRevision = revision;
      try {
        const output = await applySnake({ ...identity, snapshot, revision, command: JSON.stringify(step.command) });
        expect(step.expect, "actual command admission").toBe("applied");
        snapshot = output.snapshot;
        result = output.result;
        revision++;
      } catch (error) {
        if (!(error instanceof ArcadeRejection)) throw error;
        expect(error.code).toBe(step.expect);
        refusals.add(error.code);
        expect(snapshot, "rejected attempt leaves actual snapshot unchanged").toBe(before);
        expect(JSON.stringify(result), "rejected attempt leaves result unchanged").toBe(beforeResult);
        expect(revision, "rejected attempt leaves revision unchanged").toBe(beforeRevision);
      }
      expect(snapshot, `attempt ${attempts + 1}: literal Go snapshot`).toBe(step.snapshot_bytes);
      expect(JSON.stringify(result), `attempt ${attempts + 1}: literal Go result`).toBe(step.result_bytes);
      const actual = JSON.parse(snapshot) as SnakeSnapshot;
      const prior = JSON.parse(before) as SnakeSnapshot;
      expect(actual.width).toBe(5);
      expect(actual.height).toBe(5);
      expect(actual.revision).toBe(revision);
      if (actual.phase === "playing") {
        sawEating ||= actual.score > prior.score;
        sawGrowth ||= actual.body.length > prior.body.length;
      }
      attempts++;
      observedStates++;
    }
    expect(attempts).toBe(corpus.transition_budget);
    expect(observedStates).toBe(corpus.transition_budget + 1);
    expect(sawEating, "actual intermediate food consumption").toBe(true);
    expect(sawGrowth, "actual intermediate growth").toBe(true);
    expect([...refusals].sort()).toEqual(["advance_past_terminal", "illegal_phase"]);
    expect(snapshot).toBe(JSON.stringify(corpus.scenario.expected_terminal));
    expect(JSON.stringify(result)).toBe(JSON.stringify(corpus.scenario.expected_result));
    expect(validClearing(genesis, JSON.parse(snapshot) as SnakeSnapshot, result), "actual exact 5x5 cleared output").toBe(true);
  });

  it("refuses the 6×5 substitute, incomplete body and fabricated outcome/facts", () => {
    const source = corpus.scenario;
    const genesis = JSON.parse(source.genesis_bytes) as SnakeSnapshot;
    const terminal = JSON.parse(source.steps.at(-1)!.snapshot_bytes) as SnakeSnapshot;
    const result = JSON.parse(source.steps.at(-1)!.result_bytes) as ArcadeResult;
    expect(validClearing(genesis, terminal, result)).toBe(true);
    const old = historical.scenarios.find((row) => row.name === "snake_cleared");
    expect(old, "negative substitute exists").toBeDefined();
    expect(validClearing(JSON.parse(old!.genesis_bytes) as SnakeSnapshot,
      JSON.parse(old!.steps.at(-1)!.snapshot_bytes) as SnakeSnapshot, old!.expected_result as ArcadeResult)).toBe(false);
    expect(validClearing(genesis, { ...terminal, body: terminal.body.slice(0, 24) }, result)).toBe(false);
    expect(validClearing(genesis, { ...terminal, body: [...terminal.body.slice(0, 24), terminal.body[0]!] }, result)).toBe(false);
    expect(validClearing(genesis, terminal, { ...result, outcome: "quit" })).toBe(false);
    for (const index of [0, 1]) {
      const facts = result.score_facts.map((fact, i) => ({ ...fact, value: fact.value + (i === index ? 1 : 0) }));
      expect(validClearing(genesis, terminal, { ...result, score_facts: facts })).toBe(false);
    }
  });
});
