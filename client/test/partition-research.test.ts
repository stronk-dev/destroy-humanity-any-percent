import { describe, expect, it } from "vitest";
import corpus from "../../testdata/axis-stack/partition-research-v1.json";
import fixtureRaw from "../../balance/testdata/axis-stack/economy-v5-fixture.json?raw";
import goldensRaw from "../../testdata/decimal-vectors.json?raw";
import instrumentGo from "../../server/production/axis_partition_research_test.go?raw";
import timingGo from "../../server/production/axis_timing_test.go?raw";
import engineGo from "../../server/production/engine.go?raw";
import contentGo from "../../server/production/content.go?raw";
import accrualGo from "../../server/production/accrual.go?raw";
import ledgerGo from "../../server/economy/ledger.go?raw";
import decimalGo from "../../server/decimal/decimal.go?raw";
import canonicalGo from "../../server/decimal/canonical.go?raw";
import numericTS from "../src/numeric.ts?raw";
import productionTS from "../src/production.ts?raw";
import instrumentTS from "./partition-research.test.ts?raw";
import { canonicalString, parseCanonical } from "../src/numeric";
import { accrueConstant } from "../src/production";

// Independent bounded BigInt research model, NOT a client/save implementation.
interface Rational { n: bigint; d: bigint }
interface Projection { wire: string; residue: string }
interface Frozen {
  id: string; rate: string; initial: string; efficiency: string; cap: string; end_ms: number; cut_ms: number;
  current_one: string; current_split: string; current_states_equal: boolean;
  reference: Projection; prototype_one: Projection; prototype_split: Projection; discarded: Projection;
}
interface Boundary {
  id: string; initial: string; rate_before: string; rate_after: string; efficiency: string; cap: string;
  end_before: number; cut_before: number; end_after: number; cut_after: number; debit: string;
  reference_before: Projection; before: Projection; after_debit: string; reference_after: Projection; after: Projection; retroactive_before: Projection;
}
const report = corpus as { schema_version: number; seed: number; measurement_valid: boolean; acceptance_status: string;
  source_sha256: Record<string,string>; frozen: Frozen[]; boundaries: Boundary[]; projection_controls: {source:string;expected:string}[];
  current_partition_divergences: number; current_one_reference_differences: number; discarded_residue_hits: number;
  retroactive_hits: number; retained_cap_excess_hits: number; invalid_controls_rejected: number };
const abs = (n: bigint): bigint => n < 0n ? -n : n;
function rational(n: bigint,d = 1n): Rational {
  if (d <= 0n) throw new Error("invalid denominator");
  let a = abs(n), b = d; while (b !== 0n) { const r = a % b; a = b; b = r; }
  return {n:n/a,d:d/a};
}
function parse(raw:string):Rational {
  if (raw.length>160) throw new Error("research literal too long");
  const ratio=/^(-?\d+)\/(\d+)$/.exec(raw);
  if (ratio) return rational(BigInt(ratio[1]!),BigInt(ratio[2]!));
  const decimal=/^(-?)(\d+)(?:\.(\d+))?(?:[eE](-?\d+))?$/.exec(raw);
  if (!decimal) throw new Error("invalid research literal");
  const exponent=Number(decimal[4]??0);
  if (!Number.isSafeInteger(exponent)||Math.abs(exponent)>32) throw new Error("research exponent outside -32..32");
  const fraction=decimal[3]??"",shift=exponent-fraction.length;
  if (Math.abs(shift)>160) throw new Error("research literal scale outside domain");
  const n=BigInt((decimal[1]??"")+decimal[2]+fraction);
  return shift>=0?rational(n*10n**BigInt(shift)):rational(n,10n**BigInt(-shift));
}
const add=(a:Rational,b:Rational):Rational=>rational(a.n*b.d+b.n*a.d,a.d*b.d);
const sub=(a:Rational,b:Rational):Rational=>add(a,{n:-b.n,d:b.d});
const mul=(a:Rational,b:Rational):Rational=>rational(a.n*b.n,a.d*b.d);
const cmp=(a:Rational,b:Rational):number=>a.n*b.d<b.n*a.d?-1:a.n*b.d>b.n*a.d?1:0;
const ratioString=(a:Rational):string=>a.d===1n?String(a.n):`${a.n}/${a.d}`;
function project(value:Rational):string {
  let denominator=value.d;
  const powers=[0,0];
  for (let i=0;i<2;i++) { const factor=i===0?2n:5n; while (denominator%factor===0n) {
    denominator/=factor;powers[i] = powers[i]!+1; if (powers[i]!>48) throw new Error("research scale exceeds48");
  }}
  if (denominator!==1n) throw new Error("nonterminating research rational");
  const scale=Math.max(...powers),units=abs(value.n)*(10n**BigInt(scale)/value.d);
  if (units===0n) return "0";
  const digits=String(units),exponent=digits.length-scale-1;
  if (Math.abs(exponent)>32) throw new Error("research magnitude outside -32..32");
  const significant=digits.replace(/0+$/,"");
  const coefficient=significant.length===1?significant:`${significant[0]}.${significant.slice(1)}`;
  return canonicalString(`${value.n<0n?"-":""}${coefficient}e${exponent}`);
}
function commit(value:Rational,cap:Rational):Projection {
  if (value.n<0n) throw new Error("negative research balance");
  if (cmp(value,cap)>=0) return {wire:project(cap),residue:"0"};
  const wire=project(value); return {wire,residue:ratioString(sub(value,parse(wire)))};
}
function restore(raw:unknown,cap:Rational):Rational {
  if (typeof raw!=="object"||raw===null||Object.keys(raw).sort().join(",")!=="residue,wire") throw new Error("invalid research fields");
  const fields=raw as Projection; parseCanonical(fields.wire);
  const visible=parse(fields.wire),residue=parse(fields.residue);
  if (ratioString(residue)!==fields.residue) throw new Error("noncanonical residue");
  const value=add(visible,residue);
  if (value.n<0n||cmp(value,cap)>0||cmp(visible,cap)===0&&residue.n!==0n) throw new Error("invalid reconstructed balance");
  project(value);return value;
}
function delta(rate:string,efficiency:string,elapsed:number):Rational {
  if (!Number.isSafeInteger(elapsed)||elapsed<0||elapsed>60000) throw new Error("unsupported research interval");
  const r=parse(rate),e=parse(efficiency);if (r.n<0n||e.n<0n) throw new Error("negative research input");
  return mul(mul(r,e),rational(BigInt(elapsed),1000n));
}
function accumulate(initial:string,rate:string,efficiency:string,cuts:number[],cap:Rational,discard=false):Projection {
  let value=parse(initial),previous=0,result:Projection={wire:initial,residue:"0"};
  for (const cut of cuts) {
    if (cut<=previous) throw new Error("unordered cuts");
    result=commit(add(value,delta(rate,efficiency,cut-previous)),cap);
    if (discard) result.residue="0";
    value=restore(JSON.parse(JSON.stringify(result)),cap);previous=cut;
  }return result;
}
function current(initial:string,rate:string,efficiency:string,cuts:number[],cap:string):string {
  let value=parseCanonical(initial),previous=0;
  for (const cut of cuts) {value=parseCanonical(canonicalString(value.add(accrueConstant([rate],cut-previous,efficiency)).min(parseCanonical(cap))));previous=cut;}
  return canonicalString(value);
}

describe("R-012 bounded conserved-state research — not production AC6 acceptance",()=>{
  it("binds source identities and the complete predeclared population",async()=>{
    const inputs:Record<string,string>={
      "balance/testdata/axis-stack/economy-v5-fixture.json":fixtureRaw,"testdata/decimal-vectors.json":goldensRaw,
      "server/production/axis_partition_research_test.go":instrumentGo,"server/production/axis_timing_test.go":timingGo,
      "server/production/engine.go":engineGo,"server/production/content.go":contentGo,"server/production/accrual.go":accrualGo,
      "server/economy/ledger.go":ledgerGo,"server/decimal/decimal.go":decimalGo,"server/decimal/canonical.go":canonicalGo,
      "client/src/numeric.ts":numericTS,"client/src/production.ts":productionTS,"client/test/partition-research.test.ts":instrumentTS,
    };
    expect(Object.keys(report.source_sha256).sort()).toEqual(Object.keys(inputs).sort());
    for (const [path,source] of Object.entries(inputs)) {
      const bytes=new Uint8Array(await crypto.subtle.digest("SHA-256",new TextEncoder().encode(source)));
      expect(Array.from(bytes,(b)=>b.toString(16).padStart(2,"0")).join(""),path).toBe(report.source_sha256[path]);
    }
    expect(report.schema_version).toBe(1);expect(report.seed).toBe(120307);expect(report.measurement_valid).toBe(true);
    expect(report.acceptance_status).toBe("NOT_PROVEN: production AC6 remains red");
    expect(report.frozen).toHaveLength(520);expect(report.boundaries).toHaveLength(12);expect(report.projection_controls).toHaveLength(10);
    expect(new Set([...report.frozen,...report.boundaries].map(row=>row.id)).size).toBe(532);
    expect(report.current_partition_divergences).toBe(report.frozen.filter(row=>!row.current_states_equal).length);
    expect(report.current_one_reference_differences).toBe(report.frozen.filter(row=>row.current_one!==row.reference.wire).length);
    expect(report.discarded_residue_hits).toBe(report.frozen.filter(row=>row.discarded.wire!==row.reference.wire).length);
    expect(report.current_partition_divergences).toBeGreaterThan(0);expect(report.discarded_residue_hits).toBeGreaterThan(0);
    expect(report.retroactive_hits).toBe(8);expect(report.retained_cap_excess_hits).toBe(4);expect(report.invalid_controls_rejected).toBe(8);
  });
  for(const row of report.frozen) it(row.id,()=>{
    const cap=parse(row.cap),reference=commit(add(parse(row.initial),delta(row.rate,row.efficiency,row.end_ms)),cap);
    expect(reference).toEqual(row.reference);
    expect(accumulate(row.initial,row.rate,row.efficiency,[row.end_ms],cap)).toEqual(row.prototype_one);
    expect(accumulate(row.initial,row.rate,row.efficiency,[row.cut_ms,row.end_ms],cap)).toEqual(row.prototype_split);
    expect(row.prototype_split).toEqual(reference);
    expect(accumulate(row.initial,row.rate,row.efficiency,[row.cut_ms,row.end_ms],cap,true)).toEqual(row.discarded);
    // Scalar primitive parity only: not served ApplyLogged/state parity.
    expect(current(row.initial,row.rate,row.efficiency,[row.end_ms],row.cap)).toBe(row.current_one);
    expect(current(row.initial,row.rate,row.efficiency,[row.cut_ms,row.end_ms],row.cap)).toBe(row.current_split);
  });
  for(const row of report.boundaries) it(row.id,()=>{
    const cap=parse(row.cap),before=accumulate(row.initial,row.rate_before,row.efficiency,[row.cut_before,row.end_before],cap);
    expect(before).toEqual(row.before);expect(before).toEqual(row.reference_before);
    const afterDebit=canonicalString(parseCanonical(before.wire).sub(parseCanonical(row.debit)));
    expect(afterDebit).toBe(row.after_debit);
    expect(accumulate(afterDebit,row.rate_after,row.efficiency,[row.cut_after,row.end_after],cap)).toEqual(row.after);
    expect(row.after).toEqual(row.reference_after);
    const retroactive=commit(add(parse(row.initial),delta(row.rate_after,row.efficiency,row.end_before)),cap);
    expect(retroactive).toEqual(row.retroactive_before);
    if (row.id.startsWith("rate-debit/")) expect(retroactive.wire).not.toBe(row.reference_before.wire);
    if (row.id.startsWith("cap-debit/")) {
      expect(before).toEqual({wire:"1e4",residue:"0"});
      const excess=sub(add(parse(row.initial),delta(row.rate_before,row.efficiency,row.end_before)),cap);
      expect(excess.n>0n).toBe(true);
      expect(()=>restore({wire:"1e4",residue:ratioString(excess)},cap)).toThrow();
    }
  });
  for(const row of report.projection_controls) it(`projection/${row.source}`,()=>expect(project(parse(row.source))).toBe(row.expected));
  const invalid:Record<string,()=>unknown>={
    "large exponent":()=>parse("1e33"),"small exponent":()=>parse("1e-33"),"nonterminating":()=>project(rational(1n,3n)),
    "large scale":()=>project(rational(1n,2n**49n)),"missing residue":()=>restore({wire:"1e4"},parse("1e12")),
    "negative reconstruction":()=>restore({wire:"0",residue:"-1"},parse("1e12")),
    "noncanonical wire":()=>restore({wire:"10000",residue:"0"},parse("1e12")),"elapsed outside domain":()=>delta("1e0","1e0",60001),
  };
  for(const [name,run] of Object.entries(invalid)) it(`refusal/${name}`,()=>expect(run).toThrow());
});
