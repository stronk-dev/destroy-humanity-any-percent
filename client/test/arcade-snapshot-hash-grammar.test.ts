import { describe, expect, it } from "vitest";
import content from "../../testdata/arcade/corpus-fixture-v1.json?raw";
import grammar from "../../testdata/arcade/snapshot-hash-grammar-v1.json";
import { arcadeContentHash } from "../src/arcade/catalog";
import { createMineGrid, decodeMineGridSnapshot, MINE_GRID_SCALING_DESTINATION } from "../src/arcade/mine-grid";
import { createSnake, decodeSnakeSnapshot, SNAKE_SCALING_DESTINATION } from "../src/arcade/snake";

describe.each(["mine_grid", "snake"] as const)("%s snapshot hash grammar", (engine) => {
  it.each(grammar.cases)("$name", async (row) => {
    const snake = engine === "snake";
    const identity = { content, content_hash: await arcadeContentHash(content), content_schema_version: 1,
      seed: 7n, mode: "solo" as const, scaling_inputs: { [snake ? SNAKE_SCALING_DESTINATION : MINE_GRID_SCALING_DESTINATION]: 1 } };
    const snapshot = snake ? await createSnake(identity) : await createMineGrid(identity);
    const decode = snake ? decodeSnakeSnapshot : decodeMineGridSnapshot;
    expect(() => decode(snapshot)).not.toThrow();
    const candidate = JSON.stringify({ ...JSON.parse(snapshot), arcade_content_hash: row.hash });
    // This tests hash spelling only, not binding a synthetic digest to content.
    if (row.valid) expect(decode(candidate).arcade_content_hash).toBe(row.hash);
    else expect(() => decode(candidate)).toThrow(SyntaxError);
  });
});
