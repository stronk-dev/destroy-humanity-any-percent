import { describe, expect, it } from "vitest";
import fixtureSource from "../../balance/testdata/axis-stack/economy-v5-fixture.json?raw";
import corpus from "../../testdata/axis-stack/role-floor-v1.json";
import { parseCatalog } from "../src/economy-kernel";

describe("CV1 axis upgrade role floor", () => {
  it("pins the unchanged fixture and complete paired population", async () => {
    const hash = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(fixtureSource));
    expect(Array.from(new Uint8Array(hash), (byte) => byte.toString(16).padStart(2, "0")).join("")).toBe(corpus.fixture_sha256);
    expect(corpus.schema_version).toBe(1);
    expect(corpus.cases).toHaveLength(6);
    expect(corpus.cases.filter((row) => row.expect === "accept")).toHaveLength(3);
    expect(corpus.cases.filter((row) => row.expect === "reject")).toHaveLength(3);
  });
  for (const row of corpus.cases) it(row.name, () => {
    const document = JSON.parse(fixtureSource) as { upgrades: { id: string; roles: string[] }[] };
    for (const id of row.ids) {
      const selected = document.upgrades.filter((upgrade) => upgrade.id === id);
      expect(selected, `selected ${id}`).toHaveLength(1);
      selected[0]!.roles = [...row.roles];
    }
    if (row.expect === "accept") expect(() => parseCatalog(document)).not.toThrow();
    else {
      expect(row.expect).toBe("reject");
      expect(() => parseCatalog(document)).toThrow(/axis upgrade requires at least one role/u);
    }
  });
});
