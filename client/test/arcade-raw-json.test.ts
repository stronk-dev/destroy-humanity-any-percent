import { describe, expect, it } from "vitest";
import candidate from "../../balance/testdata/arcade-v1.json?raw";
import vectors from "../../testdata/arcade/raw-json-v1.json";
import { arcadeContentHash } from "../src/arcade/catalog";
import { createMineGrid, applyMineGrid, MINE_GRID_SCALING_DESTINATION } from "../src/arcade/mine-grid";
import { createSnake, applySnake, SNAKE_SCALING_DESTINATION } from "../src/arcade/snake";

const content = JSON.stringify(JSON.parse(candidate));

describe("Arcade raw catalog and command parity (AR1/AR3/AR4)", () => {
  it.each(vectors.catalog)("catalog: $name", async (test) => {
    expect(content.split(test.find)).toHaveLength(2);
    const changed = content.replace(test.find, test.replace);
    for (const snake of [false, true]) {
      const input = { content: changed, content_hash: await arcadeContentHash(changed), content_schema_version: 1,
        seed: 7n, mode: "solo" as const, scaling_inputs: { [snake ? SNAKE_SCALING_DESTINATION : MINE_GRID_SCALING_DESTINATION]: 1 } };
      const created = snake ? createSnake(input) : createMineGrid(input);
      if (test.valid) expect(JSON.parse(await created).revision).toBe(1);
      else await expect(created).rejects.toThrow(SyntaxError);
    }
  });

  it.each(vectors.commands)("command: $name", async (test) => {
    const snake = test.engine === "snake";
    const identity = { content, content_hash: await arcadeContentHash(content), content_schema_version: 1,
      seed: 7n, mode: "solo" as const, scaling_inputs: { [snake ? SNAKE_SCALING_DESTINATION : MINE_GRID_SCALING_DESTINATION]: 1 } };
    let snapshot = snake ? await createSnake(identity) : await createMineGrid(identity);
    let revision = 1;
    if (!snake) {
      snapshot = (await applyMineGrid({ ...identity, revision, snapshot, command: '{"kind":"choose_board","preset_id":"small"}' })).snapshot;
      revision++;
    }
    const input = { ...identity, revision, snapshot, command: test.raw };
    const applied = snake ? applySnake(input) : applyMineGrid(input);
    if (test.expect === "applied") expect(JSON.parse((await applied).snapshot).revision).toBe(revision + 1);
    else await expect(applied).rejects.toMatchObject({ code: test.expect });
    expect(input.snapshot).toBe(snapshot);
    expect(input.command).toBe(test.raw);
  });
});
