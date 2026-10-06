import { describe, expect, it } from "vitest";
import artifact from "../../testdata/axis-stack/anchor-research-v1.json";
import previous from "../../testdata/axis-stack/partition-research-v1.json";
import previousRaw from "../../testdata/axis-stack/partition-research-v1.json?raw";
import goldensRaw from "../../testdata/production-accrual.json?raw";
import instrumentGo from "../../server/production/axis_anchor_research_test.go?raw";
import carryInstrumentGo from "../../server/production/axis_partition_research_test.go?raw";
import instrumentTS from "./anchor-research.test.ts?raw";
import accrualGo from "../../server/production/accrual.go?raw";
import decimalGo from "../../server/decimal/decimal.go?raw";
import canonicalGo from "../../server/decimal/canonical.go?raw";
import numericTS from "../src/numeric.ts?raw";
import productionTS from "../src/production.ts?raw";
import { canonicalString, isStateValue, MAX_EXACT_INTEGER, parseCanonical, quantize } from "../src/numeric";
import { accrueConstant } from "../src/production";

// Independent test-only anchor/codec. No served client/save implementation.
interface Snapshot {
  version: number; initial: string; rates: string[]; efficiency: string; cap: string; elapsed_ms: number; wire: string;
}
interface Observation {
  id: string; initial: string; rates: string[]; efficiency: string; cap: string; end_ms: number; cut_ms: number;
  expected_error: boolean; reference_wire: string; one: Snapshot|null; split: Snapshot|null; rebased_wire: string;
}
const report=artifact as { version:number; acceptance_status:string; source_sha256:Record<string,string>; cases:Observation[];
  boundaries:{id:string;before:Snapshot;after:Snapshot;prior_reference:string;prior_agrees:boolean;retroactive_before:string}[];
  near_cap_after_debit:Snapshot[];restore_negatives:Record<string,string>;
  go_only_carry_diagnostics:{id:string;wire:string;residue:string;admitted:boolean}[];
  ordinary_rebase_hits:number;domain_refused:number;golden_refused:number;prior_boundary_differences:number;max_snapshot_bytes:number };

function value(raw:string) {
  if(typeof raw!=="string") throw new Error("invalid numeric field type");
  const result=parseCanonical(raw);
  if(!isStateValue(result)||result.lt(0)) throw new Error("negative/invalid research value");
  return result;
}
function build(initial:string,rates:string[],efficiency:string,cap:string,elapsed:number):Snapshot {
  if(!Array.isArray(rates)||rates.length>3||!rates.every(x=>typeof x==="string")) throw new Error("unsupported research sources");
  if(!Number.isSafeInteger(elapsed)||elapsed<0||elapsed>MAX_EXACT_INTEGER) throw new Error("invalid research elapsed");
  const base=value(initial),ceiling=value(cap);value(efficiency);rates.forEach(value);
  if(base.gt(ceiling)) throw new Error("invalid initial/cap");
  const result=quantize(base.add(accrueConstant(rates,elapsed,efficiency)));
  if(!isStateValue(result)) throw new Error("invalid sum before cap");
  return {version:1,initial,rates:[...rates],efficiency,cap,elapsed_ms:elapsed,wire:canonicalString(result.min(ceiling))};
}
function restore(encoded:string):Snapshot {
  if(new TextEncoder().encode(encoded).length>512) throw new Error("unsupported research JSON size");
  const raw:unknown=JSON.parse(encoded);
  if(typeof raw!=="object"||raw===null||Array.isArray(raw)) throw new Error("invalid research JSON object");
  const snapshot=raw as Snapshot;
  if(snapshot.version!==1) throw new Error("invalid research version");
  const reconstructed=build(snapshot.initial,snapshot.rates,snapshot.efficiency,snapshot.cap,snapshot.elapsed_ms);
  if(reconstructed.wire!==snapshot.wire||JSON.stringify(reconstructed)!==encoded) throw new Error("invalid/noncanonical reconstruction");
  return reconstructed;
}
function run(initial:string,rates:string[],efficiency:string,cap:string,cuts:number[],rebase=false):Snapshot {
  let snapshot=build(initial,rates,efficiency,cap,0),previous=0;
  for(const at of cuts) {
    snapshot=restore(JSON.stringify(snapshot));
    if(at<previous||!Number.isSafeInteger(at)||at>MAX_EXACT_INTEGER) throw new Error("invalid research interval");
    snapshot=build(rebase?snapshot.wire:snapshot.initial,rates,efficiency,cap,rebase?at-previous:at);
    previous=at;
  }
  return restore(JSON.stringify(snapshot));
}

describe("R-012 frozen-anchor research, not production AC6 acceptance",()=>{
  it("binds all declared populations and selected source identities",async()=>{
    const inputs:Record<string,string>={
      "server/production/axis_anchor_research_test.go":instrumentGo,"client/test/anchor-research.test.ts":instrumentTS,
      "server/production/axis_partition_research_test.go":carryInstrumentGo,
      "testdata/axis-stack/partition-research-v1.json":previousRaw,"testdata/production-accrual.json":goldensRaw,
      "server/production/accrual.go":accrualGo,"server/decimal/decimal.go":decimalGo,"server/decimal/canonical.go":canonicalGo,
      "client/src/production.ts":productionTS,"client/src/numeric.ts":numericTS,
    };
    expect(Object.keys(report.source_sha256).sort()).toEqual(Object.keys(inputs).sort());
    for(const [path,source] of Object.entries(inputs)) {
      const bytes=new Uint8Array(await crypto.subtle.digest("SHA-256",new TextEncoder().encode(source)));
      expect(Array.from(bytes,b=>b.toString(16).padStart(2,"0")).join(""),path).toBe(report.source_sha256[path]);
    }
    expect(report.version).toBe(1);expect(report.acceptance_status).toBe("NOT_PROVEN: production AC6 remains red");
    expect(report.cases).toHaveLength(603);expect(report.boundaries).toHaveLength(12);
    expect(report.near_cap_after_debit).toHaveLength(3);expect(Object.keys(report.restore_negatives)).toHaveLength(16);
    expect(report.go_only_carry_diagnostics).toHaveLength(3);
    expect(new Set([...report.cases,...report.boundaries].map(row=>row.id)).size).toBe(615);
    expect(report.cases.filter(row=>row.id.startsWith("ordinary/")||row.id.startsWith("cap/"))).toHaveLength(520);
    const domain=report.cases.filter(row=>row.id.startsWith("domain/")),golden=report.cases.filter(row=>row.id.startsWith("golden/"));
    expect(domain).toHaveLength(64);expect(golden).toHaveLength(16);
    expect(domain.filter(row=>row.expected_error)).toHaveLength(4);expect(golden.filter(row=>row.expected_error)).toHaveLength(5);
    expect(report.domain_refused).toBe(4);expect(report.golden_refused).toBe(5);
    expect(report.ordinary_rebase_hits).toBe(report.cases.slice(0,520).filter(row=>row.rebased_wire!==row.reference_wire).length);
    expect(report.ordinary_rebase_hits).toBeGreaterThanOrEqual(45);
    expect(report.prior_boundary_differences).toBe(report.boundaries.filter(row=>!row.prior_agrees).length);
    const snapshots=report.cases.flatMap(row=>[row.one,row.split]).filter(s=>s!==null)
      .concat(report.boundaries.flatMap(row=>[row.before,row.after]),report.near_cap_after_debit);
    const sizes=snapshots.map(s=>new TextEncoder().encode(JSON.stringify(s)).length);
    expect(Math.max(...sizes)).toBe(report.max_snapshot_bytes);expect(report.max_snapshot_bytes).toBeLessThanOrEqual(512);
  });
  for(const row of report.cases) it(row.id,()=>{
    const one=()=>run(row.initial,row.rates,row.efficiency,row.cap,[row.end_ms]);
    const split=()=>run(row.initial,row.rates,row.efficiency,row.cap,[row.cut_ms,row.end_ms]);
    if(row.expected_error) {
      expect(one).toThrow();expect(split).toThrow();expect(()=>accrueConstant(row.rates,row.end_ms,row.efficiency)).toThrow();
      expect(row.one).toBeNull();expect(row.split).toBeNull();return;
    }
    expect(one()).toEqual(row.one);expect(split()).toEqual(row.split);expect(split()).toEqual(one());
    expect(one().wire).toBe(row.reference_wire);
    expect(run(row.initial,row.rates,row.efficiency,row.cap,[row.cut_ms,row.end_ms],true).wire).toBe(row.rebased_wire);
    const old=previous.frozen.find(x=>x.id===row.id);
    if(old) {expect(row.reference_wire).toBe(old.current_one);expect(row.rebased_wire).toBe(old.current_split);}
    if(row.id.startsWith("golden/")) {
      const goldens=JSON.parse(goldensRaw) as {vectors:{name:string;expect?:string}[]};
      expect(row.reference_wire).toBe(goldens.vectors.find(x=>`golden/${x.name}`===row.id)!.expect);
    }
  });
  for(const row of report.boundaries) it(row.id,()=>{
    const input=previous.boundaries.find(x=>x.id===row.id)!;
    const before=run(input.initial,[input.rate_before],input.efficiency,input.cap,[input.cut_before,input.end_before]);
    expect(before).toEqual(row.before);
    const debit=canonicalString(parseCanonical(before.wire).sub(parseCanonical(input.debit)));
    const after=run(debit,[input.rate_after],input.efficiency,input.cap,[input.cut_after,input.end_after]);
    expect(after).toEqual(row.after);
    expect(after).toEqual(run(debit,[input.rate_after],input.efficiency,input.cap,[input.end_after]));
    expect(row.prior_reference).toBe(input.reference_after.wire);expect(row.prior_agrees).toBe(after.wire===row.prior_reference);
    const retroactive=build(input.initial,[input.rate_after],input.efficiency,input.cap,input.end_before);
    expect(retroactive.wire).toBe(row.retroactive_before);
    if(row.id.startsWith("rate-debit/")) expect(retroactive.wire).not.toBe(before.wire);
    if(row.id.startsWith("cap-debit/")) {
      expect(before.wire).toBe(input.cap);expect(after.initial).toBe(debit);expect(after.initial).not.toBe(before.initial);
      const stale=build(before.initial,before.rates,before.efficiency,before.cap,input.end_before+input.end_after);
      expect(canonicalString(parseCanonical(stale.wire).sub(parseCanonical(input.debit)))).not.toBe(after.wire);
    }
  });
  for(const encoded of Object.values(report.restore_negatives)) it(`restore/${encoded}`,()=>expect(()=>restore(encoded)).toThrow());
  for(const [i,row] of report.cases.filter(x=>x.id.startsWith("near-cap/")).entries()) it(`post-debit/${row.id}`,()=>{
    const initial=canonicalString(parseCanonical(row.one!.wire).sub(parseCanonical("1.13e1")));
    const after=run(initial,row.rates,row.efficiency,row.cap,[500,1000]);
    expect(after).toEqual(report.near_cap_after_debit[i]);expect(after.initial).not.toBe(row.initial);
  });
});
