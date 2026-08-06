import Decimal from "break_infinity.js";

import type { EconomyCatalog } from "./economy-kernel";
import { substream } from "./combat/rng";
import { isStateValue, MAX_EXACT_INTEGER, parseCanonical } from "./numeric";

export const ACTIVE_PLAY_SCHEMA_VERSION = 1;
export const ACTIVE_PLAY_SAMPLER_VERSION = "gamma6_exp.v1";
export const ACTIVE_PLAY_SPAWN_SUBSTREAM = "active_play.spawn.v1";
export const ACTIVE_PLAY_OPPORTUNITY_ID_SUBSTREAM = "active_play.opportunity_id.v1";
export const ACTIVE_PLAY_BUFF_ID_SUBSTREAM = "active_play.buff_id.v1";
const idPattern = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;

export type ActivePlayEffect =
  | { readonly effectRowId: string; readonly kind: "production_frenzy"; readonly weight: number; readonly factor: string; readonly durationMs: number; readonly targets: readonly ["generator_production"] }
  | { readonly effectRowId: string; readonly kind: "click_frenzy"; readonly weight: number; readonly factor: string; readonly durationMs: number; readonly actionIds: readonly string[] }
  | { readonly effectRowId: string; readonly kind: "building_special"; readonly weight: number; readonly perOwnedPpm: number; readonly durationMs: number; readonly eligibleGeneratorIds: readonly string[] }
  | { readonly effectRowId: string; readonly kind: "lucky_payout"; readonly weight: number; readonly luckyBankFrac: string; readonly luckyRateCap: string; readonly epsilon: string; readonly resourceId: string; readonly hardcapReasonKey: string };

export interface ActivePlayCatalog {
  readonly schedule: { readonly samplerVersion: string; readonly substreamLabel: string; readonly minimumIntervalMs: number; readonly scaleMs: number; readonly lifetimeMs: number; readonly maxDueTransitions: number };
  readonly effects: readonly ActivePlayEffect[];
  readonly combo: { readonly cap: string; readonly hardcapReasonKey: string };
}

export function loadActivePlayCatalog(source: unknown, economy: EconomyCatalog): ActivePlayCatalog {
  const root = exactObject(source, ["schema_version", "schedule_policy", "effects", "combo_policy"], "opportunities catalog");
  if (root.schema_version !== ACTIVE_PLAY_SCHEMA_VERSION) throw new SyntaxError("invalid opportunities schema version");
  const scheduleRaw = exactObject(root.schedule_policy, ["sampler_version", "substream_label", "minimum_interval_ms", "scale_ms", "lifetime_ms", "max_due_transitions"], "schedule_policy");
  if (scheduleRaw.sampler_version !== ACTIVE_PLAY_SAMPLER_VERSION || scheduleRaw.substream_label !== ACTIVE_PLAY_SPAWN_SUBSTREAM) throw new SyntaxError("invalid opportunity sampler");
  const schedule = Object.freeze({ samplerVersion: scheduleRaw.sampler_version, substreamLabel: scheduleRaw.substream_label,
    minimumIntervalMs: positiveSafe(scheduleRaw.minimum_interval_ms), scaleMs: positiveSafe(scheduleRaw.scale_ms),
    lifetimeMs: positiveSafe(scheduleRaw.lifetime_ms), maxDueTransitions: positiveSafe(scheduleRaw.max_due_transitions) });
  if (!Array.isArray(root.effects) || root.effects.length === 0) throw new SyntaxError("effects must be non-empty");
  let previous = "";
  const effects = root.effects.map((source, index): ActivePlayEffect => {
    if (!isRecord(source)) throw new SyntaxError(`effects[${index}] must be an object`);
    const id = mechanical(source.effect_row_id), weight = positiveSafe(source.weight), kind = source.kind;
    if (previous !== "" && byteCompare(previous, id) >= 0) throw new SyntaxError("effects must be byte-sorted"); previous = id;
    if (kind === "production_frenzy") {
      const raw = exactObject(source, ["effect_row_id", "kind", "weight", "factor", "duration_ms", "targets"], `effects[${index}]`);
      const factor = positiveFactor(raw.factor); if (!Array.isArray(raw.targets) || raw.targets.length !== 1 || raw.targets[0] !== "generator_production" || !declaration(economy, id, "all")) throw new SyntaxError("invalid production frenzy");
      return Object.freeze({ effectRowId:id, kind, weight, factor, durationMs:positiveSafe(raw.duration_ms), targets:Object.freeze(["generator_production"] as const) });
    }
    if (kind === "click_frenzy") {
      const raw = exactObject(source, ["effect_row_id", "kind", "weight", "factor", "duration_ms", "action_ids"], `effects[${index}]`);
      const actionIds = ids(raw.action_ids); for (const action of actionIds) if (!economy.manualActions.some((row) => row.id === action) || !declaration(economy, id, action)) throw new SyntaxError("invalid click binding");
      return Object.freeze({ effectRowId:id, kind, weight, factor:positiveFactor(raw.factor), durationMs:positiveSafe(raw.duration_ms), actionIds:Object.freeze(actionIds) });
    }
    if (kind === "building_special") {
      const raw = exactObject(source, ["effect_row_id", "kind", "weight", "per_owned_ppm", "duration_ms", "eligible_generator_ids"], `effects[${index}]`);
      const generators = ids(raw.eligible_generator_ids); for (const generator of generators) if (!economy.generatorClass(generator) || !declaration(economy, `${id}.${generator}`, generator)) throw new SyntaxError("invalid building binding");
      return Object.freeze({ effectRowId:id, kind, weight, perOwnedPpm:ppm(raw.per_owned_ppm), durationMs:positiveSafe(raw.duration_ms), eligibleGeneratorIds:Object.freeze(generators) });
    }
    if (kind === "lucky_payout") {
      const raw = exactObject(source, ["effect_row_id", "kind", "weight", "lucky_bank_frac", "lucky_rate_cap", "epsilon", "resource_id", "hardcap_reason_key"], `effects[${index}]`);
      const frac=canonical(raw.lucky_bank_frac), cap=canonical(raw.lucky_rate_cap), epsilon=canonical(raw.epsilon), resource=mechanical(raw.resource_id);
      if (!new Decimal(frac).gt(0) || !new Decimal(frac).lt(1) || !new Decimal(cap).gt(0) || new Decimal(epsilon).lt(0) || economy.resource(resource)?.scope !== "company") throw new SyntaxError("invalid lucky policy");
      return Object.freeze({ effectRowId:id, kind, weight, luckyBankFrac:frac, luckyRateCap:cap, epsilon, resourceId:resource, hardcapReasonKey:mechanical(raw.hardcap_reason_key) });
    }
    throw new SyntaxError("invalid effect kind");
  });
  const comboRaw=exactObject(root.combo_policy,["combo_cap","hardcap_reason_key"],"combo_policy"), comboCap=canonical(comboRaw.combo_cap);
  if (!new Decimal(comboCap).gt(1)) throw new SyntaxError("invalid combo cap");
  return Object.freeze({ schedule, effects:Object.freeze(effects), combo:Object.freeze({cap:comboCap,hardcapReasonKey:mechanical(comboRaw.hardcap_reason_key)}) });
}

// TS verifies the integer selection evidence but deliberately does not
// approximate the server-only floating interval sampler.
export function selectActivePlayEffect(catalog: ActivePlayCatalog, baseSeed: bigint, sequence: number): { effectRowId:string; effectDraw:bigint; generatorDraw:bigint|null; selectedGenerator:string|null } {
  if (!Number.isSafeInteger(sequence) || sequence < 0) throw new RangeError("invalid spawn sequence");
  const random=substream(baseSeed ^ BigInt(sequence),ACTIVE_PLAY_SPAWN_SUBSTREAM); for(let i=0;i<6;i++) random.next();
  const total=catalog.effects.reduce((sum,row)=>sum+BigInt(row.weight),0n), draw=random.bound(total); let remaining=draw, selected:ActivePlayEffect|undefined;
  for(const row of catalog.effects){if(remaining<BigInt(row.weight)){selected=row;break;}remaining-=BigInt(row.weight);} if(!selected)throw new RangeError("effect draw");
  if(selected.kind!=="building_special")return{effectRowId:selected.effectRowId,effectDraw:draw,generatorDraw:null,selectedGenerator:null};
  const generatorDraw=random.bound(BigInt(selected.eligibleGeneratorIds.length)); return{effectRowId:selected.effectRowId,effectDraw:draw,generatorDraw,selectedGenerator:selected.eligibleGeneratorIds[Number(generatorDraw)]!};
}

export function activePlayBuffId(baseSeed: bigint, sequence: number, attendedMs: number): string {
	return activePlayDeterministicId(baseSeed, sequence, attendedMs, ACTIVE_PLAY_BUFF_ID_SUBSTREAM);
}

export function activePlayOpportunityId(baseSeed: bigint, sequence: number, attendedMs: number): string {
	return activePlayDeterministicId(baseSeed, sequence, attendedMs, ACTIVE_PLAY_OPPORTUNITY_ID_SUBSTREAM);
}

function activePlayDeterministicId(baseSeed: bigint, sequence: number, attendedMs: number, label: string): string {
  if (!Number.isSafeInteger(sequence) || sequence < 0 || !Number.isSafeInteger(attendedMs) || attendedMs < 0) throw new RangeError("invalid buff identity coordinate");
  const random = substream(baseSeed ^ BigInt(sequence) ^ BigInt(attendedMs), label);
  const bytes = new Uint8Array(16); let first=random.next(),second=random.next();
  for(let index=7;index>=0;index--){bytes[index]=Number(first&255n);first>>=8n;} for(let index=15;index>=8;index--){bytes[index]=Number(second&255n);second>>=8n;}
  let timestamp=BigInt(attendedMs)&((1n<<48n)-1n);for(let index=5;index>=0;index--){bytes[index]=Number(timestamp&255n);timestamp>>=8n;}
  bytes[6]=(bytes[6]!&0x0f)|0x70;bytes[8]=(bytes[8]!&0x3f)|0x80;const hex=[...bytes].map((value)=>value.toString(16).padStart(2,"0")).join("");
  return `${hex.slice(0,8)}-${hex.slice(8,12)}-${hex.slice(12,16)}-${hex.slice(16,20)}-${hex.slice(20)}`;
}

function declaration(economy:EconomyCatalog,id:string,target:string):boolean{return economy.multiplierSources.some((row)=>row.id===id&&row.slot==="event_buffs"&&row.target===target&&row.provider==="active_play");}
function positiveSafe(value:unknown):number{if(typeof value!=="number"||!Number.isSafeInteger(value)||value<=0||value>MAX_EXACT_INTEGER)throw new SyntaxError("expected positive safe integer");return value;}
function ppm(value:unknown):number{const result=positiveSafe(value);if(result>1_000_000)throw new SyntaxError("invalid ppm");return result;}
function mechanical(value:unknown):string{if(typeof value!=="string"||!idPattern.test(value))throw new SyntaxError("invalid mechanical id");return value;}
function canonical(value:unknown):string{if(typeof value!=="string")throw new SyntaxError("expected decimal string");const parsed=parseCanonical(value);if(!isStateValue(parsed))throw new SyntaxError("invalid decimal");return value;}
function positiveFactor(value:unknown):string{const result=canonical(value);if(!new Decimal(result).gt(1))throw new SyntaxError("factor must exceed one");return result;}
function ids(value:unknown):string[]{if(!Array.isArray(value)||value.length===0)throw new SyntaxError("expected ids");const result=value.map(mechanical);if(result.some((item,index)=>index>0&&byteCompare(result[index-1]!,item)>=0))throw new SyntaxError("ids must be sorted");return result;}
function isRecord(source:unknown):source is Record<string,unknown>{return typeof source==="object"&&source!==null&&!Array.isArray(source);}
function exactObject(source:unknown,keys:readonly string[],label:string):Record<string,unknown>{if(!isRecord(source))throw new SyntaxError(`${label} must be an object`);const actual=Object.keys(source).sort(byteCompare),expected=[...keys].sort(byteCompare);if(actual.length!==expected.length||actual.some((value,index)=>value!==expected[index]))throw new SyntaxError(`${label} fields are not exact`);return source;}
function byteCompare(left:string,right:string):number{const a=new TextEncoder().encode(left),b=new TextEncoder().encode(right);for(let i=0;i<Math.min(a.length,b.length);i++)if(a[i]!==b[i])return a[i]!-b[i]!;return a.length-b.length;}
