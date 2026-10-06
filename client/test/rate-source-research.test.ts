import Decimal from "break_infinity.js";
import { describe, expect, it } from "vitest";
import corpus from "../../testdata/axis-stack/rate-source-research-v1.json";
import instrumentGo from "../../server/production/axis_rate_source_research_test.go?raw";
import instrumentTS from "./rate-source-research.test.ts?raw";
import timingGo from "../../server/production/axis_timing_test.go?raw";
import fixtureGo from "../../server/production/axis_stack_test.go?raw";
import carryGo from "../../server/production/axis_partition_research_test.go?raw";
import fixtureJSON from "../../balance/testdata/axis-stack/economy-v5-fixture.json?raw";
import engineGo from "../../server/production/engine.go?raw";
import contentGo from "../../server/production/content.go?raw";
import accrualGo from "../../server/production/accrual.go?raw";
import decimalGo from "../../server/decimal/decimal.go?raw";
import canonicalGo from "../../server/decimal/canonical.go?raw";
import saveGo from "../../server/save/state.go?raw";
import numericTS from "../src/numeric.ts?raw";
import productionTS from "../src/production.ts?raw";
import { canonicalString, parseCanonical, quantize, sumDeterministic, MAX_EXACT_INTEGER } from "../src/numeric";
import { accrueConstant } from "../src/production";

// Raw diagnostic reconstruction only, never an accepted state/wire producer.
describe("R-012 actual producer-rate serialization diagnostics",()=>{
  it("binds source identities and the full declared producer census",async()=>{
    const inputs:Record<string,string>={
      "server/production/axis_rate_source_research_test.go":instrumentGo,"client/test/rate-source-research.test.ts":instrumentTS,
      "server/production/axis_timing_test.go":timingGo,"server/production/axis_stack_test.go":fixtureGo,
      "server/production/axis_partition_research_test.go":carryGo,"balance/testdata/axis-stack/economy-v5-fixture.json":fixtureJSON,
      "server/production/engine.go":engineGo,"server/production/content.go":contentGo,"server/production/accrual.go":accrualGo,
      "server/decimal/decimal.go":decimalGo,"server/decimal/canonical.go":canonicalGo,"server/save/state.go":saveGo,
      "client/src/numeric.ts":numericTS,"client/src/production.ts":productionTS,
    };
    expect(Object.keys(corpus.source_sha256).sort()).toEqual(Object.keys(inputs).sort());
    for(const [path,source] of Object.entries(inputs)) {
      const digest=new Uint8Array(await crypto.subtle.digest("SHA-256",new TextEncoder().encode(source)));
      expect(Array.from(digest,b=>b.toString(16).padStart(2,"0")).join(""),path).toBe(corpus.source_sha256[path as keyof typeof corpus.source_sha256]);
    }
    expect(corpus.version).toBe(1);expect(corpus.acceptance_status).toBe("NOT_PROVEN: production AC6 remains red");
    expect(corpus.rows).toHaveLength(64);expect(new Set(corpus.rows.map(x=>x.id)).size).toBe(64);
    const expected=[1,2,9,99,12345,123456789,1234567890123,MAX_EXACT_INTEGER].flatMap(count=>["1e0","9e-1"].flatMap(eff=>[2000,3114,2045,60000].map(ms=>`producer/${count}/${eff}/${ms}`)));
    expect(corpus.rows.map(x=>x.id)).toEqual(expected);
    expect(corpus.raw_bits_changed_profiles).toBe(corpus.rows.filter(x=>!x.raw_wire_equal).length);
    expect(corpus.delta_changed_profiles).toBe(corpus.rows.filter(x=>x.raw_delta!==x.rounded_delta).length);
  });
  for(const row of corpus.rows) it(row.id,()=>{
    const raw=row.rates.map(source=>{
      const mantissa=Number(source.mantissa_diagnostic),exponent=Number(source.exponent_diagnostic);
      expect(Number.isSafeInteger(exponent)).toBe(true);
      const bits=new DataView(new ArrayBuffer(8));bits.setFloat64(0,mantissa);
      expect(bits.getBigUint64(0).toString(16).padStart(16,"0")).toBe(source.mantissa_ieee_hex);
      const result=Decimal.fromMantissaExponent(mantissa,exponent);
      expect(canonicalString(result)).toBe(source.canonical);return result;
    });
    const rawDelta=quantize(sumDeterministic(raw).mul(new Decimal(row.elapsed_ms).div(1000)).mul(parseCanonical(row.efficiency)));
    expect(canonicalString(rawDelta)).toBe(row.raw_delta);
    expect(canonicalString(accrueConstant(row.rates.map(x=>x.canonical),row.elapsed_ms,row.efficiency))).toBe(row.rounded_delta);
    expect(row.actual_engine_cash).toBe(row.raw_delta);expect(row.restored_context_delta).toBe(row.raw_delta);
    expect(raw.every((value,i)=>value.eq(parseCanonical(row.rates[i]!.canonical)))).toBe(row.raw_wire_equal);
  });
});
