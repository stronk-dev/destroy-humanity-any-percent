import { describe, expect, it } from "vitest";
import fixture from "../../balance/testdata/active-play-foundation-v1.json";
import phase0 from "../../balance/catalogs/phase0.json";
import { activePlayLuckyRequested, loadActivePlayCatalog } from "../src/active-play";
import { parseCatalog } from "../src/economy-kernel";

function economyWithSources() {
  return parseCatalog(phase0);
}

function economyWithoutActivePlaySources() {
  const source = structuredClone(phase0) as any;
  source.multiplier_sources = source.multiplier_sources.filter(
    (row: { provider: string }) => row.provider !== "active_play",
  );
  return parseCatalog(source);
}

describe("active-play catalog",()=>{
  it("requires exact member names at every catalog and effect boundary", () => {
    const objects = (root: any): Array<Record<string, unknown>> => [root, root.schedule_policy, root.combo_policy, ...root.effects];
    for (const [index, object] of objects(fixture.baseline).entries()) {
      for (const key of Object.keys(object)) {
        const candidate = structuredClone(fixture.baseline);
        const row = objects(candidate)[index];
        row[key.toUpperCase()] = row[key];
        delete row[key];
        expect(() => loadActivePlayCatalog(candidate, economyWithSources()), `${index}.${key}`).toThrow(SyntaxError);
      }
    }
  });
  it("rejects null fields from another effect arm and unsafe durations", () => {
    for (const [index, key] of ["factor", "targets", "duration_ms", "action_ids"].entries()) {
      const candidate = structuredClone(fixture.baseline) as any;
      candidate.effects[index][key] = null;
      expect(() => loadActivePlayCatalog(candidate, economyWithSources()), `${index}.${key}`).toThrow(SyntaxError);
    }
    for (const index of [0, 1, 3]) {
      const candidate = structuredClone(fixture.baseline);
      candidate.effects[index].duration_ms = 9007199254740992;
      expect(() => loadActivePlayCatalog(candidate, economyWithSources()), `${index}.duration_ms`).toThrow(SyntaxError);
    }
  });
  it("preserves the admitted duration boundaries and canonical escaped names", () => {
    for (const duration of [1, 9007199254740991]) {
      const candidate = structuredClone(fixture.baseline);
      for (const index of [0, 1, 3]) candidate.effects[index].duration_ms = duration;
      const catalog = loadActivePlayCatalog(candidate, economyWithSources());
      expect(catalog.effects.filter((effect) => "durationMs" in effect).map((effect) => effect.durationMs)).toEqual([duration, duration, duration]);
    }
    const source = JSON.stringify(fixture.baseline);
    const escaped = source.replace('"schema_version"', '"schema_versi\\u006fn"');
    expect(escaped).not.toBe(source);
    expect(loadActivePlayCatalog(JSON.parse(escaped), economyWithSources()).effects).toHaveLength(4);
    const duplicated = `${source.slice(0, -1)},"schema_version":2,"schema_version":1}`;
    expect(loadActivePlayCatalog(JSON.parse(duplicated), economyWithSources()).effects).toHaveLength(4);
  });
  it("loads the shared strict fixture",()=>{const catalog=loadActivePlayCatalog(fixture.baseline,economyWithSources());expect(catalog.effects.map((row)=>row.kind)).toEqual(["building_special","click_frenzy","lucky_payout","production_frenzy"]);});
  it("rejects missing and cross-artifact bindings",()=>{const missing=structuredClone(fixture.baseline) as any;delete missing.schedule_policy.scale_ms;expect(()=>loadActivePlayCatalog(missing,economyWithSources())).toThrow();expect(()=>loadActivePlayCatalog(fixture.baseline,economyWithoutActivePlaySources())).toThrow();});
  it("matches the five shared Lucky boundary vectors",()=>{for(const vector of fixture.lucky_vectors)expect(activePlayLuckyRequested(vector.bank,vector.rate,vector.fraction,vector.rate_cap,vector.epsilon),vector.name).toBe(vector.requested);});
});
