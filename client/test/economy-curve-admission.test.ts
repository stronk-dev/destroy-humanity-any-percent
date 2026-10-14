import { describe, expect, it } from "vitest";
import fixture from "../../testdata/economy-kernel.json";
import { parseCatalog } from "../src/economy-kernel";

function candidate(curve: Record<string, unknown>) {
  const root = structuredClone(fixture.catalog);
  Object.assign(root.generator_classes[0]!.price, { curve });
  return root;
}

describe("closed cost-curve arms", () => {
  const curves: Record<string, unknown>[] = [{ kind: "constant" }, { kind: "linear", step: "0" }, { kind: "geometric", ratio: "1e0" }];
  for (const curve of curves) {
    it(`preserves ${curve.kind} including legal zero/unit controls`, () => expect(() => parseCatalog(candidate(curve))).not.toThrow());
    for (const key of Object.keys(curve)) it(`refuses ${curve.kind}/alias/${key}`, () => {
      const bad = { ...curve, [key.toUpperCase()]: curve[key] }; delete bad[key];
      expect(() => parseCatalog(candidate(bad))).toThrow();
    });
    for (const wrong of ["step", "ratio"]) {
      if (wrong in curve) continue;
      it(`refuses ${curve.kind}/wrong-arm-null/${wrong}`, () => expect(() => parseCatalog(candidate({ ...curve, [wrong]: null }))).toThrow());
    }
  }
});
