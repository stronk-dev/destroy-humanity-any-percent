import { describe, expect, it } from "vitest";
import fixture from "../../testdata/arcade/corpus-fixture-v1.json?raw";
import negatives from "../../testdata/arcade/snapshot-value-negatives-v1.json";
import rawGrammar from "../../testdata/arcade/snapshot-raw-negatives-v1.json";
import { arcadeContentHash } from "../src/arcade/catalog";
import { applyMineGrid, createMineGrid, decodeMineGridSnapshot, MINE_GRID_SCALING_DESTINATION } from "../src/arcade/mine-grid";

async function population(stage: string) {
  const identity = { content: fixture, content_hash: await arcadeContentHash(fixture), content_schema_version: 1,
    seed: BigInt(negatives.seed), mode: "solo" as const, scaling_inputs: { [MINE_GRID_SCALING_DESTINATION]: 1 } };
  let snapshot = await createMineGrid(identity);
  let revision = 1;
  const commands = stage === "genesis" ? [] : [{ kind: "choose_board", preset_id: negatives.preset_id },
    ...(stage === "unplaced" ? [] : [{ kind: "reveal", cell: negatives.first_cell }]),
    ...(stage === "terminal" ? [{ kind: "quit" }] : [])];
  for (const command of commands) {
    snapshot = (await applyMineGrid({ ...identity, snapshot, revision, command: JSON.stringify(command) })).snapshot;
    revision++;
  }
  expect(decodeMineGridSnapshot(snapshot).phase).toBe(stage === "genesis" ? "setup" : stage === "terminal" ? "terminal" : "playing");
  if (stage === "placed" || stage === "terminal") expect(decodeMineGridSnapshot(snapshot).revealed.length).toBeGreaterThan(0);
  return { ...identity, snapshot, revision, command: '{"kind":"quit"}' };
}

function forge(snapshot: string, path: readonly (string | number)[], value: unknown): string {
  const root: unknown = JSON.parse(snapshot);
  let target = root as Record<string | number, unknown>;
  for (const key of path.slice(0, -1)) target = target[key] as Record<string | number, unknown>;
  target[path[path.length - 1]!] = value;
  return JSON.stringify(root);
}

describe("Mine Grid shared snapshot value grammar (AR3.2/AR3.4/AR7)", () => {
  it.each(negatives.cases)("decoder refuses $name", async (row) => {
    const input = await population(row.stage);
    const clean = decodeMineGridSnapshot(input.snapshot);
    expect(() => decodeMineGridSnapshot(forge(input.snapshot, row.path, row.value))).toThrow(SyntaxError);
    expect(decodeMineGridSnapshot(input.snapshot)).toEqual(clean);
  });

  it.each(negatives.cases)("apply refuses $name without corrupting the clean control", async (row) => {
    const input = await population(row.stage);
    await expect(applyMineGrid({ ...input, snapshot: forge(input.snapshot, row.path, row.value) })).rejects.toThrow(SyntaxError);
    if (row.stage !== "terminal") {
      const baseline = await applyMineGrid(input);
      expect(await applyMineGrid(input)).toEqual(baseline);
    } else {
      // Clean terminals are valid snapshots but cannot be quit a second time.
      await expect(applyMineGrid(input)).rejects.toMatchObject({ code: "illegal_phase" });
    }
  });
});

function replaceExactlyOnce(snapshot: string, before: string, after: string): string {
  expect(snapshot.split(before), "raw mutation must match exactly once").toHaveLength(2);
  return snapshot.replace(before, after);
}

describe("Mine Grid shared raw snapshot grammar (AR3.4/AR7)", () => {
  it.each(rawGrammar.cases)("decoder refuses $name", async (row) => {
    const input = await population(row.stage);
    const forged = replaceExactlyOnce(input.snapshot, row.before, row.after);
    expect(() => decodeMineGridSnapshot(forged)).toThrow(SyntaxError);
    expect(() => decodeMineGridSnapshot(input.snapshot)).not.toThrow();
  });

  it.each(rawGrammar.cases)("direct apply refuses $name", async (row) => {
    const input = await population(row.stage);
    const forged = replaceExactlyOnce(input.snapshot, row.before, row.after);
    await expect(applyMineGrid({ ...input, snapshot: forged })).rejects.toThrow(SyntaxError);
    if (row.stage === "terminal") {
      await expect(applyMineGrid(input)).rejects.toMatchObject({ code: "illegal_phase" });
    } else {
      const baseline = await applyMineGrid(input);
      expect(await applyMineGrid(input)).toEqual(baseline);
    }
  });

  it.each(rawGrammar.positive_cases)("accepts valid spelling $name", async (row) => {
    const input = await population(row.stage);
    const legal = replaceExactlyOnce(input.snapshot, row.before, row.after);
    expect(() => decodeMineGridSnapshot(legal)).not.toThrow();
    const output = await applyMineGrid({ ...input, snapshot: legal });
    // A flagged board is intentionally different; every other spelling normalizes identically.
    if (row.name !== "flag_coordinates_at_bounds") expect(output).toEqual(await applyMineGrid(input));
    else expect(decodeMineGridSnapshot(output.snapshot).flags).toEqual([0, 80]);
  });

  it("rejects an unmatched or ambiguous mutation instrument", () => {
    expect(() => replaceExactlyOnce("{}", "missing", "x")).toThrow();
    expect(() => replaceExactlyOnce("xx", "x", "y")).toThrow();
  });

  it("refuses malformed nesting without exhausting the scanner stack", () => {
    const nested = '{"flags":' + "[".repeat(10_000) + "0" + "]".repeat(10_000) + "}";
    expect(() => decodeMineGridSnapshot(nested)).toThrow(SyntaxError);
  });
});
