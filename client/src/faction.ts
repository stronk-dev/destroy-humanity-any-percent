import { MAX_EXACT_INTEGER } from "./numeric";

export interface FactionDefinition {
  readonly id: string;
  readonly produces: string;
  readonly consumes: string;
  readonly compact: { readonly autoSign: true; readonly tithePpm: number } | null;
}

export interface FactionCatalog {
  readonly stockCap: number;
  readonly stockIntervalMs: number;
  readonly factions: readonly FactionDefinition[];
  readonly byId: ReadonlyMap<string, FactionDefinition>;
}

const mechanical = /^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$/;
const phase0Ids = ["bootstrapper", "enterprise", "open_source", "vc_funded"];
const phase0Resources = ["compliance", "hype", "libraries", "revenue"];

export function parseFactionCatalog(source: unknown, minimumTithePpm: number, defaultTithePpm: number, maximumTithePpm: number): FactionCatalog {
  const root = exactObject(source, ["schema_version", "stock_cap", "stock_interval_ms", "factions"]);
  if (root.schema_version !== 1 || !safe(root.stock_cap, 1, MAX_EXACT_INTEGER) || !safe(root.stock_interval_ms, 1, 86_400_000) ||
      !safe(minimumTithePpm, 0, 1_000_000) || !safe(defaultTithePpm, minimumTithePpm, maximumTithePpm) ||
      !safe(maximumTithePpm, defaultTithePpm, 1_000_000) || !Array.isArray(root.factions) || root.factions.length !== 4) {
    throw new SyntaxError("invalid faction catalog");
  }
  const factions = root.factions.map((item) => {
    const raw = exactObject(item, ["id", "produces", "consumes", "compact", "modifier_slots", "incorporation_copy_key"]);
    if (typeof raw.id !== "string" || !mechanical.test(raw.id) || typeof raw.produces !== "string" || !mechanical.test(raw.produces) ||
        typeof raw.consumes !== "string" || !mechanical.test(raw.consumes) || raw.produces === raw.consumes ||
        !Array.isArray(raw.modifier_slots) || raw.modifier_slots.length !== 0 || raw.incorporation_copy_key !== `incorporate.${raw.id}`) {
      throw new SyntaxError("invalid faction definition");
    }
    let compact: FactionDefinition["compact"] = null;
    if (raw.compact !== null) {
      const binding = exactObject(raw.compact, ["auto_sign", "tithe_ppm"]);
      if (raw.id !== "open_source" || binding.auto_sign !== true || !safe(binding.tithe_ppm, minimumTithePpm, maximumTithePpm) || binding.tithe_ppm <= defaultTithePpm) {
        throw new SyntaxError("invalid faction compact binding");
      }
      compact = Object.freeze({ autoSign: true as const, tithePpm: binding.tithe_ppm });
    } else if (raw.id === "open_source") {
      throw new SyntaxError("open_source requires compact binding");
    }
    return Object.freeze({ id: raw.id, produces: raw.produces, consumes: raw.consumes, compact });
  }).sort((a, b) => byteCompare(a.id, b.id));
  if (!same(factions.map((value) => value.id), phase0Ids) || !same(factions.map((value) => value.produces).sort(byteCompare), phase0Resources) ||
      !same(factions.map((value) => value.consumes).sort(byteCompare), phase0Resources) || new Set(factions.map((value) => value.produces)).size !== 4 ||
      new Set(factions.map((value) => value.consumes)).size !== 4 || !singleCycle(factions)) {
    throw new SyntaxError("invalid faction cycle");
  }
  const frozen = Object.freeze(factions);
  return Object.freeze({ stockCap: root.stock_cap, stockIntervalMs: root.stock_interval_ms, factions: frozen, byId: new Map(frozen.map((value) => [value.id, value])) });
}

function singleCycle(factions: readonly FactionDefinition[]): boolean {
  const byId = new Map(factions.map((value) => [value.id, value]));
  const consumer = new Map(factions.map((value) => [value.consumes, value.id]));
  const seen = new Set<string>(); let current = factions[0]!.id;
  for (let index = 0; index < factions.length; index++) { if (seen.has(current)) return false; seen.add(current); current = consumer.get(byId.get(current)!.produces)!; }
  return current === factions[0]!.id && seen.size === factions.length;
}

function exactObject(source: unknown, keys: readonly string[]): Record<string, unknown> {
  if (typeof source !== "object" || source === null || Array.isArray(source)) throw new SyntaxError("expected object");
  const value = source as Record<string, unknown>; const actual = Object.keys(value).sort(byteCompare); const expected = [...keys].sort(byteCompare);
  if (!same(actual, expected)) throw new SyntaxError("object fields are not exact"); return value;
}
function safe(value: unknown, min: number, max: number): value is number { return typeof value === "number" && Number.isSafeInteger(value) && value >= min && value <= max; }
function same(left: readonly string[], right: readonly string[]): boolean { return left.length === right.length && left.every((value, index) => value === right[index]); }
function byteCompare(left: string, right: string): number { return left < right ? -1 : left > right ? 1 : 0; }
