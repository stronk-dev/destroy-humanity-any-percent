import { describe, expect, it } from "vitest";
import fixture from "../../balance/factions/phase0.json";
import compact from "../../balance/commons/phase0.json";
import { parseFactionCatalog } from "../src/faction";

function parse(source: unknown) {
  return parseFactionCatalog(source, compact.minimum_tithe_ppm, compact.default_tithe_ppm, compact.maximum_tithe_ppm);
}

type Root = typeof fixture;
const targets: { path: string; select(root: Root): object }[] = [
  { path: "root", select: (root) => root },
  ...fixture.factions.map((row, index) => ({ path: row.id, select: (root: Root) => root.factions[index]! })),
  { path: "open_source/compact", select: (root) => root.factions.find((row) => row.id === "open_source")!.compact! },
];

describe("faction literal catalog keys", () => {
  it("preserves explicitly nullable compact bindings", () => {
    expect(parse(fixture).factions).toHaveLength(4);
  });
  for (const target of targets) {
    for (const key of Object.keys(target.select(fixture))) {
      it(`refuses alias/${target.path}/${key}`, () => {
        const candidate = structuredClone(fixture);
        const row = target.select(candidate) as Record<string, unknown>;
        row[key.toUpperCase()] = row[key];
        delete row[key];
        expect(() => parse(candidate)).toThrow();
      });
    }
  }
  for (const [index, row] of fixture.factions.entries()) {
    if (row.compact !== null) continue;
    it(`requires nullable compact/${row.id}`, () => {
      const candidate = structuredClone(fixture);
      Reflect.deleteProperty(candidate.factions[index]!, "compact");
      expect(() => parse(candidate)).toThrow();
    });
  }
});
