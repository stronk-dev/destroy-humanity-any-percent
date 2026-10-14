import { describe, expect, it } from "vitest";
import raw from "../../balance/commons/phase0.json";
import { parseCommonsCatalog } from "../src/commons";

function candidate() { return { ...structuredClone(raw), source_weights: [{ source_id: "source.example", slot: "upgrades", weight_ppm: 1000000, forsworn: false }] }; }

describe("Commons literal catalog admission", () => {
  for (const path of ["root", "weight"] as const) {
    for (const key of Object.keys(path === "root" ? candidate() : candidate().source_weights[0]!)) {
      it(`refuses alias/${path}/${key}`, () => {
        const source = candidate();
        const row = (path === "root" ? source : source.source_weights[0]!) as unknown as Record<string, unknown>;
        row[key.toUpperCase()] = row[key]; delete row[key];
        expect(() => parseCommonsCatalog(source)).toThrow();
      });
    }
  }
  it("preserves explicit false", () => expect(parseCommonsCatalog(candidate()).sourceWeights[0]!.forsworn).toBe(false));
  for (const variant of ["missing", "null"] as const) it(`requires forsworn/${variant}`, () => {
    const source = candidate();
    if (variant === "missing") Reflect.deleteProperty(source.source_weights[0]!, "forsworn");
    else Object.assign(source.source_weights[0]!, { forsworn: null });
    expect(() => parseCommonsCatalog(source)).toThrow();
  });
  for (const key of ["solidarity_window_ms", "cohort_target_size", "cohort_merge_floor", "npc_population_floor"] as const) {
    for (const maximum of [true, false]) it(`${key}/${maximum ? "maximum" : "first invalid"}`, () => {
      const value = Number.MAX_SAFE_INTEGER + (maximum ? 0 : 1), source = candidate(); source[key] = value;
      if (key === "cohort_merge_floor") source.cohort_target_size = source.npc_population_floor = value;
      if (maximum) expect(() => parseCommonsCatalog(source)).not.toThrow();
      else expect(() => parseCommonsCatalog(source)).toThrow();
    });
  }
  it("accepts the registered axis_stack slot", () => {
    const source = candidate(); source.source_weights[0]!.slot = "axis_stack";
    expect(parseCommonsCatalog(source).sourceWeights[0]!.slot).toBe("axis_stack");
  });
});
