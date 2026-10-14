import { describe, expect, it } from "vitest";
import fixture from "../../testdata/economy-kernel.json";
import { parseCatalog } from "../src/economy-kernel";

describe("resource catalog literal keys", () => {
  for (const scope of ["resource", "hardcap"] as const) {
    const resource = fixture.catalog.resources[0]!;
    for (const key of Object.keys(scope === "resource" ? resource : resource.hardcap!)) {
      it(`refuses ${scope}/${key}`, () => {
        const source = structuredClone(fixture.catalog), resource = source.resources[0]!;
        const row = (scope === "resource" ? resource : resource.hardcap!) as unknown as Record<string, unknown>;
        row[key.toUpperCase()] = row[key]; delete row[key];
        expect(() => parseCatalog(source)).toThrow();
      });
    }
  }
  it("preserves both explicit cap and explicit null", () => {
    expect(() => parseCatalog(fixture.catalog)).not.toThrow();
    const source = structuredClone(fixture.catalog); source.resources[0]!.hardcap = null;
    expect(() => parseCatalog(source)).not.toThrow();
  });
});
