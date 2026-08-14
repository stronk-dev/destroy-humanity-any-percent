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
  it("loads the shared strict fixture",()=>{const catalog=loadActivePlayCatalog(fixture.baseline,economyWithSources());expect(catalog.effects.map((row)=>row.kind)).toEqual(["building_special","click_frenzy","lucky_payout","production_frenzy"]);});
  it("rejects missing and cross-artifact bindings",()=>{const missing=structuredClone(fixture.baseline) as any;delete missing.schedule_policy.scale_ms;expect(()=>loadActivePlayCatalog(missing,economyWithSources())).toThrow();expect(()=>loadActivePlayCatalog(fixture.baseline,economyWithoutActivePlaySources())).toThrow();});
  it("matches the five shared Lucky boundary vectors",()=>{for(const vector of fixture.lucky_vectors)expect(activePlayLuckyRequested(vector.bank,vector.rate,vector.fraction,vector.rate_cap,vector.epsilon),vector.name).toBe(vector.requested);});
});
