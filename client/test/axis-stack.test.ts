import corpusJSON from "../../testdata/axis-stack/loader-corpus-v1.json";
import fixtureJSON from "../../balance/testdata/axis-stack/economy-v5-fixture.json";
import routesJSON from "../../balance/routes/phase0.json";
import epoch8EconomyJSON from "../../balance/catalogs/phase0.json";
import { describe, expect, it } from "vitest";

import { MULTIPLIER_SLOT_ORDER, parseCatalog, validateAxisInputs } from "../src/economy-kernel";
import { parseRoutesCatalog } from "../src/routes";

type PathElement = string | number | { id: string };
interface Mutation { op: "set" | "append" | "delete"; path: PathElement[]; value?: unknown }
interface Row { name: string; mutations: Mutation[]; expect: "accept" | "reject"; achievements_pinned?: boolean; maximum_attainment_run?: number; maximum_score_run?: number }
const corpus = corpusJSON as unknown as { economy_cases: Row[]; cross_cases: Row[]; route_cases: Row[] };

function mutate(base: unknown, mutations: readonly Mutation[]): unknown {
  const document = structuredClone(base) as Record<string, unknown>;
  const apply = (node: unknown, path: readonly PathElement[], mutation: Mutation): unknown => {
    if (path.length === 0) {
      if (mutation.op === "set") return structuredClone(mutation.value);
      if (mutation.op === "append") { (node as unknown[]).push(structuredClone(mutation.value)); return node; }
      throw new Error(`unsupported terminal op ${mutation.op}`);
    }
    const [head, ...rest] = path;
    if (Array.isArray(node)) {
      const index = typeof head === "number" ? head : node.findIndex((row) => (row as { id?: string }).id === (head as { id: string }).id);
      if (index < 0 || index >= node.length) throw new Error(`path element ${JSON.stringify(head)} selects nothing`);
      if (rest.length === 0 && mutation.op === "delete") { node.splice(index, 1); return node; }
      node[index] = apply(node[index], rest, mutation);
      return node;
    }
    const object = node as Record<string, unknown>;
    if (rest.length === 0 && mutation.op === "delete") { delete object[head as string]; return object; }
    object[head as string] = apply(object[head as string], rest, mutation);
    return object;
  };
  let result: unknown = document;
  for (const mutation of mutations) result = apply(result, mutation.path, mutation);
  return result;
}

const outcome = (run: () => unknown): "accept" | "reject" => { try { run(); return "accept"; } catch { return "reject"; } };

describe("axis stack economy v5 loader corpus (CV1, AC1)", () => {
  it("covers the shared corpus", () => expect(corpus.economy_cases.length).toBeGreaterThanOrEqual(20));
  for (const row of corpus.economy_cases) {
    it(row.name, () => expect(outcome(() => parseCatalog(mutate(fixtureJSON, row.mutations)))).toBe(row.expect));
  }
  for (const row of corpus.cross_cases) {
    it(`cross/${row.name}`, () => {
      const catalog = parseCatalog(mutate(fixtureJSON, row.mutations));
      expect(outcome(() => validateAxisInputs(catalog, row.achievements_pinned!, row.maximum_attainment_run!, row.maximum_score_run!))).toBe(row.expect);
    });
  }
  for (const row of corpus.route_cases) {
    it(`routes/${row.name}`, () => expect(outcome(() => parseRoutesCatalog(mutate(routesJSON, row.mutations)))).toBe(row.expect));
  }
  it("keeps the fixture shape, the epoch-8 catalog, and the OD-6 slot order", () => {
    const catalog = parseCatalog(fixtureJSON);
    expect(catalog.axisStack).toEqual({ input: "achievement_attainment_run", inputCap: 44, capReasonKey: "cap.axis_stack_input" });
    const intern = catalog.upgrade("upgrade.pr_intern_1")!;
    expect(intern.axisMinimum).toBe(6);
    expect(intern.requires).toEqual([]);
    expect(intern.effects).toEqual([{ sourceId: "upgrade.pr_intern_1.axis", slot: "axis_stack", target: "all", factorPpm: 25_000 }]);
    expect(parseCatalog(epoch8EconomyJSON).axisStack).toBeNull();
    expect([...MULTIPLIER_SLOT_ORDER]).toEqual(["upgrades", "milestones", "axis_stack", "faction", "doctrine", "commons", "trust", "event_buffs", "prestige"]);
  });
});
