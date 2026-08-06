import { describe, expect, it } from "vitest";
import fixture from "../../balance/testdata/active-play-foundation-v1.json";
import phase0 from "../../balance/catalogs/phase0.json";
import { loadActivePlayCatalog } from "../src/active-play";
import { parseCatalog } from "../src/economy-kernel";

function economyWithSources() {
  const source=structuredClone(phase0) as any;
  source.multiplier_sources.push(
    {id:"active.building.generator.beige_tower",slot:"event_buffs",target:"generator.beige_tower",provider:"active_play"},
    {id:"active.click",slot:"event_buffs",target:"manual.click",provider:"active_play"},
    {id:"active.production",slot:"event_buffs",target:"all",provider:"active_play"},
  );
  return parseCatalog(source);
}

describe("active-play catalog",()=>{
  it("loads the shared strict fixture",()=>{const catalog=loadActivePlayCatalog(fixture.baseline,economyWithSources());expect(catalog.effects.map((row)=>row.kind)).toEqual(["building_special","click_frenzy","lucky_payout","production_frenzy"]);});
  it("rejects missing and cross-artifact bindings",()=>{const missing=structuredClone(fixture.baseline) as any;delete missing.schedule_policy.scale_ms;expect(()=>loadActivePlayCatalog(missing,economyWithSources())).toThrow();expect(()=>loadActivePlayCatalog(fixture.baseline,parseCatalog(phase0))).toThrow();});
});
