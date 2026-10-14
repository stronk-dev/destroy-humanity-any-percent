import { describe, expect, it } from "vitest";
import foundation from "../../testdata/economy-foundation-v4.json";
import axis from "../../balance/testdata/axis-stack/economy-v5-fixture.json";
import legacy from "../../testdata/economy-kernel.json";
import { parseCatalog } from "../src/economy-kernel";
import { parseRoutePredicate } from "../src/routes";

type Path = (string | number)[];
type Row = Record<string, unknown>;
function select(root: unknown, path: Path): Row {
  let value = root;
  for (const key of path) value = (value as Record<string | number, unknown>)[key];
  return value as Row;
}
function objects(root: unknown, path: Path = []): Path[] {
  if (Array.isArray(root)) return root.flatMap((child, index) => objects(child, [...path, index]));
  if (root === null || typeof root !== "object") return [];
  return [path, ...Object.entries(root).flatMap(([key, child]) =>
    path.length === 0 && ["resources", "generator_classes", "manual_policy", "offline_policy"].includes(key)
      ? [] : objects(child, [...path, key]))];
}

describe("remaining Economy catalog literal member admission", () => {
  for (const [name, source] of Object.entries({ v4: foundation, v5: axis })) {
    it(`${name}/original`, () => expect(() => parseCatalog(source)).not.toThrow());
    for (const path of objects(source)) for (const key of Object.keys(select(source, path))) {
      it(`${name}/${path.join("/")}/alias/${key}`, () => {
        const candidate = structuredClone(source), row = select(candidate, path);
        row[key.toUpperCase()] = row[key]; delete row[key];
        expect(() => parseCatalog(candidate)).toThrow();
      });
    }
  }
  const progress = (v: unknown) => select(v, ["progress_coordinates", 0]);
  for (const [name, edit] of Object.entries({
    "missing tier": (v: unknown) => { delete progress(v).tier; },
    "null tier": (v: unknown) => { progress(v).tier = null; },
    "wrong count": (v: unknown) => { progress(v).count = null; },
    "wrong required": (v: unknown) => { progress(v).required = 0; },
    "wrong terms": (v: unknown) => { progress(v).terms = null; },
    "static wrong operand": (v: unknown) => { select(v, ["upgrades", 0, "effects", 0]).factor_ppm = null; },
    "v4 axis null": (v: unknown) => { select(v, []).axis_stack = null; },
  })) it(name, () => {
    const candidate = structuredClone(foundation); edit(candidate);
    expect(() => parseCatalog(candidate)).toThrow();
  });
  for (const key of ["manual_actions", "multiplier_sources", "progress_coordinates", "manual_policy", "offline_policy", "upgrades", "synergy_pools", "axis_stack"]) {
    it(`legacy wrong member/${key}`, () => {
      const candidate: Row = structuredClone(legacy.catalog); candidate[key] = null;
      expect(() => parseCatalog(candidate)).toThrow();
    });
  }
});

describe("shared Route literal active-arm admission", () => {
  const rows: Row[] = [
    { kind: "resource_at_least", resource_id: "company.cash", value: "0" },
    { kind: "resource_at_most", resource_id: "company.cash", value: "1e3" },
    { kind: "meter_band", meter_id: "meter.standing", min: 0, max: 100 },
    { kind: "doctrine_is", transition: "transition.t0_to_t1", doctrine_id: "doctrine.a" },
    { kind: "doctrine_is_not", transition: "transition.t0_to_t1", doctrine_id: "doctrine.a" },
    { kind: "ledger_fact_present", fact_kind: "fact.example" },
    { kind: "structure_is", structure_id: "structure.example" },
    { kind: "region_trait", trait_id: "trait.example" },
  ];
  for (const source of rows) {
    it(`${source.kind}/original`, () => expect(() => parseRoutePredicate([source])).not.toThrow());
    for (const key of Object.keys(source)) it(`${source.kind}/alias/${key}`, () => {
      const candidate = structuredClone(source); candidate[key.toUpperCase()] = candidate[key]; delete candidate[key];
      expect(() => parseRoutePredicate([candidate])).toThrow();
    });
    for (const key of ["resource_id", "value", "meter_id", "min", "max", "transition", "doctrine_id", "fact_kind", "structure_id", "trait_id"]) {
      if (key in source) continue;
      it(`${source.kind}/wrong-arm-null/${key}`, () => expect(() => parseRoutePredicate([{ ...source, [key]: null }])).toThrow());
    }
  }
});
